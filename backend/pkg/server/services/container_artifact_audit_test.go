package services

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"pentagi/pkg/database"
	obs "pentagi/pkg/observability"
	"pentagi/pkg/server/models"

	"github.com/sirupsen/logrus"
	logrustest "github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestContainerArtifactAuditFieldsKeepOnlySafeMetadata(t *testing.T) {
	ctx := obs.WithAuditCorrelation(context.Background(), obs.AuditCorrelation{
		Source:        obs.AuditSourceMCP,
		RequestID:     "request-artifact",
		PrincipalHash: "principal-artifact",
		SessionHash:   "session-artifact",
		OriginHash:    "origin-artifact",
		ClientIPHash:  "client-artifact",
	})

	fields := containerArtifactAuditFields(
		ctx,
		17,
		"pull",
		"complete",
		"success",
		2,
		2,
		3,
		4096,
		true,
		1250*time.Millisecond,
	)
	for key, expected := range map[string]any{
		"action":          containerArtifactAuditAction,
		"flow_id":         uint64(17),
		"container_type":  database.ContainerTypePrimary,
		"operation":       "pull",
		"stage":           "complete",
		"outcome":         "success",
		"requested_count": 2,
		"completed_count": 2,
		"file_count":      3,
		"byte_count":      int64(4096),
		"force":           true,
		"duration_ms":     int64(1250),
		"audit_source":    obs.AuditSourceMCP,
		"request_id":      "request-artifact",
		"principal_hash":  "principal-artifact",
		"session_hash":    "session-artifact",
		"origin_hash":     "origin-artifact",
		"client_ip_hash":  "client-artifact",
	} {
		if actual := fields[key]; actual != expected {
			t.Fatalf("field %q = %v, want %v", key, actual, expected)
		}
	}

	for _, forbidden := range []string{
		"container_id", "local_id", "container_name", "name", "image", "container_path",
		"cache_path", "path", "file_name", "content", "output", "error", "error_message",
	} {
		if _, found := fields[forbidden]; found {
			t.Fatalf("artifact audit fields must not include %q", forbidden)
		}
	}
}

func TestContainerArtifactAuditRecordCommittedCountsRegularFiles(t *testing.T) {
	audit := newContainerArtifactAudit(context.Background(), 1)
	audit.recordCommitted([]models.FlowFile{
		{IsDir: true, Size: 999},
		{Size: 12},
		{Size: 34},
		{Size: -1},
	})

	assert.Equal(t, 1, audit.completedCount)
	assert.Equal(t, 3, audit.fileCount)
	assert.Equal(t, int64(46), audit.byteCount)
}

func TestPullFlowFilesEmitsSafeArtifactAuditOnSuccess(t *testing.T) {
	db := setupFlowFileServiceTestDB(t)
	dataDir := t.TempDir()
	seedFlow(t, db, 17, 1)
	fakeDocker := &fakeDockerClient{
		running: true,
		copyFromBody: buildContainerTar([]tarTestEntry{
			{name: "credential.txt", typeflag: 0, content: "SECRET_TOKEN=do-not-log"},
		}),
	}
	svc := NewFlowFileService(db, dataDir, "", fakeDocker, &flowFileCaptureSubscriptions{})
	c, w := newFlowFileTestContext(
		"POST",
		"/flows/17/files/pull",
		strings.NewReader(`{"path":"/var/lib/credential.txt"}`),
		[]string{"flow_files.upload", "containers.view"},
		1,
		17,
	)
	c.Request = c.Request.WithContext(obs.WithAuditCorrelation(c.Request.Context(), obs.AuditCorrelation{
		Source:        obs.AuditSourceMCP,
		RequestID:     "request-pull-success",
		PrincipalHash: "principal-pull-success",
		SessionHash:   "session-pull-success",
	}))

	hook := logrustest.NewGlobal()
	defer hook.Reset()
	svc.PullFlowFiles(c)

	require.Equal(t, 200, w.Code)
	auditEntry := findContainerArtifactAuditEntry(hook.AllEntries())
	require.NotNil(t, auditEntry)
	assert.Equal(t, "complete", auditEntry.Data["stage"])
	assert.Equal(t, "success", auditEntry.Data["outcome"])
	assert.Equal(t, 1, auditEntry.Data["requested_count"])
	assert.Equal(t, 1, auditEntry.Data["completed_count"])
	assert.Equal(t, 1, auditEntry.Data["file_count"])
	assert.Equal(t, int64(len("SECRET_TOKEN=do-not-log")), auditEntry.Data["byte_count"])
	assert.Equal(t, "request-pull-success", auditEntry.Data["request_id"])

	serialized := fmt.Sprint(auditEntry.Data)
	for _, sensitive := range []string{
		"/var/lib/credential.txt", "credential.txt", "SECRET_TOKEN=do-not-log", "container_id", "error_message",
	} {
		if strings.Contains(serialized, sensitive) {
			t.Fatalf("success artifact audit contains sensitive value %q: %s", sensitive, serialized)
		}
	}
}

func TestPullFlowFilesEmitsErrorArtifactAuditWithoutDockerDetails(t *testing.T) {
	db := setupFlowFileServiceTestDB(t)
	dataDir := t.TempDir()
	seedFlow(t, db, 18, 1)
	fakeDocker := &fakeDockerClient{
		running:     true,
		copyFromErr: errors.New("docker://daemon/secret-container failed for /etc/shadow"),
	}
	svc := NewFlowFileService(db, dataDir, "", fakeDocker, &flowFileCaptureSubscriptions{})
	c, w := newFlowFileTestContext(
		"POST",
		"/flows/18/files/pull",
		strings.NewReader(`{"path":"/etc/shadow"}`),
		[]string{"flow_files.upload", "containers.view"},
		1,
		18,
	)

	hook := logrustest.NewGlobal()
	defer hook.Reset()
	svc.PullFlowFiles(c)

	require.Equal(t, 500, w.Code)
	auditEntry := findContainerArtifactAuditEntry(hook.AllEntries())
	require.NotNil(t, auditEntry)
	assert.Equal(t, "copy", auditEntry.Data["stage"])
	assert.Equal(t, "error", auditEntry.Data["outcome"])
	assert.Equal(t, 0, auditEntry.Data["completed_count"])
	assert.Equal(t, 0, auditEntry.Data["file_count"])
	assert.Equal(t, int64(0), auditEntry.Data["byte_count"])

	serialized := fmt.Sprint(auditEntry.Data)
	for _, sensitive := range []string{
		"docker://daemon/secret-container", "/etc/shadow", "secret-container", "error_message",
	} {
		if strings.Contains(serialized, sensitive) {
			t.Fatalf("error artifact audit contains sensitive value %q: %s", sensitive, serialized)
		}
	}
}

func findContainerArtifactAuditEntry(entries []*logrus.Entry) *logrus.Entry {
	for _, entry := range entries {
		if entry.Data["action"] == containerArtifactAuditAction {
			return entry
		}
	}
	return nil
}
