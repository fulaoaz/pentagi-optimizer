package services

import (
	"context"
	"math"
	"strings"
	"time"

	"pentagi/pkg/database"
	obs "pentagi/pkg/observability"
	"pentagi/pkg/server/models"

	"github.com/sirupsen/logrus"
)

const containerArtifactAuditAction = "container_artifact_sync"

type containerArtifactAudit struct {
	ctx            context.Context
	flowID         uint64
	startedAt      time.Time
	operation      string
	stage          string
	outcome        string
	force          bool
	requestedCount int
	completedCount int
	fileCount      int
	byteCount      int64
}

func newContainerArtifactAudit(ctx context.Context, flowID uint64) *containerArtifactAudit {
	if ctx == nil {
		ctx = context.Background()
	}
	return &containerArtifactAudit{
		ctx:       ctx,
		flowID:    flowID,
		startedAt: time.Now(),
		operation: "pull",
		stage:     "flow_lookup",
		outcome:   "error",
	}
}

func (audit *containerArtifactAudit) setStage(stage string) {
	audit.stage = stage
}

func (audit *containerArtifactAudit) setRequest(requestedCount int, force bool) {
	audit.requestedCount = nonNegativeArtifactCount(requestedCount)
	audit.force = force
}

func (audit *containerArtifactAudit) recordCommitted(files []models.FlowFile) {
	audit.completedCount++
	for _, file := range files {
		if file.IsDir {
			continue
		}
		audit.fileCount++
		if file.Size > 0 {
			audit.byteCount = saturatingAuditBytes(audit.byteCount, file.Size)
		}
	}
}

func (audit *containerArtifactAudit) markSuccess() {
	audit.outcome = "success"
	audit.stage = "complete"
}

func (audit *containerArtifactAudit) emit(status int) {
	outcome := audit.outcome
	if status >= 400 {
		outcome = "error"
	}
	logrus.WithContext(audit.ctx).WithFields(containerArtifactAuditFields(
		audit.ctx,
		audit.flowID,
		audit.operation,
		audit.stage,
		outcome,
		audit.requestedCount,
		audit.completedCount,
		audit.fileCount,
		audit.byteCount,
		audit.force,
		time.Since(audit.startedAt),
	)).Info("Container artifact synchronization audit recorded")
}

func containerArtifactAuditFields(
	ctx context.Context,
	flowID uint64,
	operation, stage, outcome string,
	requestedCount, completedCount, fileCount int,
	byteCount int64,
	force bool,
	duration time.Duration,
) logrus.Fields {
	fields := logrus.Fields{
		"action":          containerArtifactAuditAction,
		"flow_id":         flowID,
		"container_type":  database.ContainerTypePrimary,
		"operation":       safeArtifactAuditToken(operation, "other", "pull"),
		"stage":           safeArtifactAuditToken(stage, "unknown", "flow_lookup", "authorization", "request", "validation", "cache_prepare", "runtime_check", "copy", "extract", "commit", "complete"),
		"outcome":         safeArtifactAuditToken(outcome, "error", "success", "error"),
		"requested_count": nonNegativeArtifactCount(requestedCount),
		"completed_count": nonNegativeArtifactCount(completedCount),
		"file_count":      nonNegativeArtifactCount(fileCount),
		"byte_count":      nonNegativeAuditBytes(byteCount),
		"force":           force,
		"duration_ms":     safeArtifactAuditDuration(duration),
	}
	correlation := obs.AuditCorrelationFromContext(ctx)
	if correlation.Source == obs.AuditSourceMCP {
		for key, value := range correlation.Fields() {
			fields[key] = value
		}
	}
	return fields
}

func safeArtifactAuditToken(value, fallback string, allowed ...string) string {
	value = strings.TrimSpace(value)
	for _, candidate := range allowed {
		if value == candidate {
			return value
		}
	}
	return fallback
}

func safeArtifactAuditDuration(duration time.Duration) int64 {
	if duration < 0 {
		return 0
	}
	return duration.Milliseconds()
}

func nonNegativeArtifactCount(value int) int {
	if value < 0 {
		return 0
	}
	return value
}

func nonNegativeAuditBytes(value int64) int64 {
	if value < 0 {
		return 0
	}
	return value
}

func saturatingAuditBytes(current, added int64) int64 {
	if current < 0 {
		current = 0
	}
	if added <= 0 || current >= math.MaxInt64-added {
		if added > 0 && current >= math.MaxInt64-added {
			return math.MaxInt64
		}
		return current
	}
	return current + added
}
