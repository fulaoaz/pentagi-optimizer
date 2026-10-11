package tools

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"pentagi/pkg/config"
	"pentagi/pkg/database"
	"pentagi/pkg/docker"
	obs "pentagi/pkg/observability"

	"github.com/moby/moby/api/types/container"
	"github.com/sirupsen/logrus"
	logrustest "github.com/sirupsen/logrus/hooks/test"
)

func TestMCPContainerLifecycleAuditFieldsKeepOnlySafeMetadata(t *testing.T) {
	ctx := obs.WithAuditCorrelation(context.Background(), obs.AuditCorrelation{
		Source:        obs.AuditSourceMCP,
		RequestID:     "request-123",
		PrincipalHash: "principal-hash",
		SessionHash:   "session-hash",
		OriginHash:    "origin-hash",
		ClientIPHash:  "client-ip-hash",
	})

	fields, ok := mcpContainerLifecycleAuditFields(ctx, 11, "provision", "success", 125*time.Millisecond)
	if !ok {
		t.Fatal("MCP correlation did not produce container lifecycle audit fields")
	}

	for key, expected := range map[string]any{
		"action":         "mcp_container_lifecycle",
		"flow_id":        int64(11),
		"container_type": database.ContainerTypePrimary,
		"operation":      "provision",
		"outcome":        "success",
		"duration_ms":    int64(125),
		"audit_source":   obs.AuditSourceMCP,
		"request_id":     "request-123",
		"principal_hash": "principal-hash",
		"session_hash":   "session-hash",
		"origin_hash":    "origin-hash",
		"client_ip_hash": "client-ip-hash",
	} {
		if actual := fields[key]; actual != expected {
			t.Fatalf("field %q = %v, want %v", key, actual, expected)
		}
	}

	for _, forbidden := range []string{"container_id", "local_id", "container_name", "name", "image", "work_dir", "host_dir", "args", "result", "output"} {
		if _, found := fields[forbidden]; found {
			t.Fatalf("container lifecycle audit fields must not include %q", forbidden)
		}
	}

	if _, ok := mcpContainerLifecycleAuditFields(context.Background(), 11, "provision", "success", 0); ok {
		t.Fatal("non-MCP context unexpectedly produced container lifecycle audit fields")
	}
}

func TestContainerRecoveryAuditFieldsKeepSafeMetadata(t *testing.T) {
	mcpContext := obs.WithAuditCorrelation(context.Background(), obs.AuditCorrelation{
		Source:        obs.AuditSourceMCP,
		RequestID:     "request-456",
		PrincipalHash: "principal-hash",
		SessionHash:   "session-hash",
		OriginHash:    "origin-hash",
		ClientIPHash:  "client-ip-hash",
	})

	for _, test := range []struct {
		name            string
		ctx             context.Context
		recoveryReason  string
		outcome         string
		wantCorrelation bool
	}{
		{
			name:            "runtime unavailable with MCP correlation",
			ctx:             mcpContext,
			recoveryReason:  "runtime_unavailable",
			outcome:         "error",
			wantCorrelation: true,
		},
		{
			name:            "runtime stopped without MCP correlation",
			ctx:             context.Background(),
			recoveryReason:  "runtime_stopped",
			outcome:         "success",
			wantCorrelation: false,
		},
		{
			name:            "stale status without MCP correlation",
			ctx:             context.Background(),
			recoveryReason:  "stale_status",
			outcome:         "success",
			wantCorrelation: false,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			fields := containerRecoveryAuditFields(
				test.ctx,
				23,
				test.recoveryReason,
				test.outcome,
				375*time.Millisecond,
			)

			for key, expected := range map[string]any{
				"action":          "container_recovery",
				"flow_id":         int64(23),
				"container_type":  database.ContainerTypePrimary,
				"recovery_reason": test.recoveryReason,
				"outcome":         test.outcome,
				"duration_ms":     int64(375),
			} {
				if actual := fields[key]; actual != expected {
					t.Fatalf("field %q = %v, want %v", key, actual, expected)
				}
			}

			correlationKeys := []string{
				"audit_source",
				"request_id",
				"principal_hash",
				"session_hash",
				"origin_hash",
				"client_ip_hash",
			}
			for _, key := range correlationKeys {
				_, found := fields[key]
				if found != test.wantCorrelation {
					t.Fatalf("correlation field %q presence = %v, want %v", key, found, test.wantCorrelation)
				}
			}

			for _, forbidden := range []string{
				"container_id", "local_id", "container_name", "name", "image",
				"work_dir", "host_dir", "args", "result", "output", "error",
				"error_message",
			} {
				if _, found := fields[forbidden]; found {
					t.Fatalf("container recovery audit fields must not include %q", forbidden)
				}
			}
		})
	}
}

type recoveryQuerierMock struct {
	database.Querier
	primary    database.Container
	primaryErr error
}

func (m *recoveryQuerierMock) GetFlowPrimaryContainer(context.Context, int64) (database.Container, error) {
	return m.primary, m.primaryErr
}

type recoveryDockerMock struct {
	docker.DockerClient
	running        bool
	runningErr     error
	removeErr      error
	runResult      database.Container
	runErr         error
	inspectCalls   int
	removeCalls    int
	provisionCalls int
}

func (m *recoveryDockerMock) IsContainerRunning(context.Context, string) (bool, error) {
	m.inspectCalls++
	return m.running, m.runningErr
}

func (m *recoveryDockerMock) RemoveContainer(context.Context, string, int64) error {
	m.removeCalls++
	return m.removeErr
}

func (m *recoveryDockerMock) RunContainer(
	context.Context,
	string,
	database.ContainerType,
	int64,
	*container.Config,
	*container.HostConfig,
) (database.Container, error) {
	m.provisionCalls++
	return m.runResult, m.runErr
}

func TestPrepareEmitsContainerRecoveryAudit(t *testing.T) {
	ctx := obs.WithAuditCorrelation(context.Background(), obs.AuditCorrelation{
		Source:        obs.AuditSourceMCP,
		RequestID:     "request-recovery",
		PrincipalHash: "principal-recovery",
		SessionHash:   "session-recovery",
		OriginHash:    "origin-recovery",
		ClientIPHash:  "client-recovery",
	})

	for _, test := range []struct {
		name           string
		status         database.ContainerStatus
		running        bool
		runningErr     error
		removeErr      error
		runErr         error
		wantReason     string
		wantOutcome    string
		wantPrepareErr bool
		wantInspect    int
		wantRemove     int
		wantProvision  int
	}{
		{
			name:          "runtime unavailable recovers",
			status:        database.ContainerStatusRunning,
			runningErr:    errors.New("sensitive runtime failure"),
			wantReason:    "runtime_unavailable",
			wantOutcome:   "success",
			wantInspect:   1,
			wantRemove:    1,
			wantProvision: 1,
		},
		{
			name:          "stopped runtime recovers",
			status:        database.ContainerStatusRunning,
			wantReason:    "runtime_stopped",
			wantOutcome:   "success",
			wantInspect:   1,
			wantRemove:    1,
			wantProvision: 1,
		},
		{
			name:          "stale status recovers",
			status:        database.ContainerStatusStopped,
			wantReason:    "stale_status",
			wantOutcome:   "success",
			wantRemove:    1,
			wantProvision: 1,
		},
		{
			name:           "remove failure is audited",
			status:         database.ContainerStatusRunning,
			runningErr:     errors.New("sensitive runtime failure"),
			removeErr:      errors.New("sensitive remove failure"),
			wantReason:     "runtime_unavailable",
			wantOutcome:    "error",
			wantPrepareErr: true,
			wantInspect:    1,
			wantRemove:     1,
		},
		{
			name:           "provision failure is audited",
			status:         database.ContainerStatusRunning,
			runErr:         errors.New("sensitive provision failure"),
			wantReason:     "runtime_stopped",
			wantOutcome:    "error",
			wantPrepareErr: true,
			wantInspect:    1,
			wantRemove:     1,
			wantProvision:  1,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			hook := logrustest.NewGlobal()
			defer hook.Reset()

			primary := database.Container{
				ID:      71,
				Type:    database.ContainerTypePrimary,
				Name:    "sensitive-container-name",
				Image:   "sensitive-container-image",
				Status:  test.status,
				FlowID:  23,
				LocalID: sql.NullString{String: "sensitive-container-id", Valid: true},
			}
			dockerClient := &recoveryDockerMock{
				running:    test.running,
				runningErr: test.runningErr,
				removeErr:  test.removeErr,
				runResult: database.Container{
					ID:      72,
					Type:    database.ContainerTypePrimary,
					LocalID: sql.NullString{String: "new-sensitive-container-id", Valid: true},
				},
				runErr: test.runErr,
			}
			executor := &flowToolsExecutor{
				db:     &recoveryQuerierMock{primary: primary},
				docker: dockerClient,
				cfg:    &config.Config{DataDir: t.TempDir()},
				flowID: 23,
				image:  "sensitive-image",
			}

			err := executor.Prepare(ctx)
			if (err != nil) != test.wantPrepareErr {
				t.Fatalf("Prepare() error = %v, want error = %v", err, test.wantPrepareErr)
			}
			if dockerClient.inspectCalls != test.wantInspect {
				t.Fatalf("inspect calls = %d, want %d", dockerClient.inspectCalls, test.wantInspect)
			}
			if dockerClient.removeCalls != test.wantRemove {
				t.Fatalf("remove calls = %d, want %d", dockerClient.removeCalls, test.wantRemove)
			}
			if dockerClient.provisionCalls != test.wantProvision {
				t.Fatalf("provision calls = %d, want %d", dockerClient.provisionCalls, test.wantProvision)
			}

			var recoveryEntry *logrus.Entry
			for index := range hook.Entries {
				entry := &hook.Entries[index]
				if entry.Data["action"] == "container_recovery" {
					recoveryEntry = entry
					break
				}
			}
			if recoveryEntry == nil {
				t.Fatalf("container recovery audit entry not found: %#v", hook.Entries)
			}
			if got := recoveryEntry.Data["recovery_reason"]; got != test.wantReason {
				t.Fatalf("recovery reason = %v, want %q", got, test.wantReason)
			}
			if got := recoveryEntry.Data["outcome"]; got != test.wantOutcome {
				t.Fatalf("outcome = %v, want %q", got, test.wantOutcome)
			}
			for _, key := range []string{"request_id", "principal_hash", "session_hash", "origin_hash", "client_ip_hash"} {
				if _, found := recoveryEntry.Data[key]; !found {
					t.Fatalf("recovery entry is missing MCP correlation field %q", key)
				}
			}

			serialized := fmt.Sprint(recoveryEntry.Data)
			for _, sensitive := range []string{
				"sensitive-container-name", "sensitive-container-image",
				"sensitive-container-id", "new-sensitive-container-id",
				"sensitive-image", "sensitive runtime failure",
				"sensitive remove failure", "sensitive provision failure",
			} {
				if strings.Contains(serialized, sensitive) {
					t.Fatalf("recovery audit entry contains sensitive value %q: %s", sensitive, serialized)
				}
			}
		})
	}
}
