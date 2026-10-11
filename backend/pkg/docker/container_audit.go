package docker

import (
	"context"
	"strings"
	"time"

	"pentagi/pkg/database"
	obs "pentagi/pkg/observability"

	"github.com/sirupsen/logrus"
)

type containerAuditContextKey struct{}

// WithContainerAuditContext carries only the flow and container type needed to
// connect low-level Docker events to a flow without passing container IDs into
// the audit record.
func WithContainerAuditContext(ctx context.Context, flowID int64, containerType database.ContainerType) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithValue(ctx, containerAuditContextKey{}, containerAuditContext{
		flowID:        flowID,
		containerType: containerType,
	})
}

type containerAuditContext struct {
	flowID        int64
	containerType database.ContainerType
}

func containerAuditContextFromContext(ctx context.Context) containerAuditContext {
	if ctx == nil {
		return containerAuditContext{}
	}
	metadata, _ := ctx.Value(containerAuditContextKey{}).(containerAuditContext)
	return metadata
}

func containerExitAuditFields(
	ctx context.Context,
	operation, reason, outcome string,
	duration time.Duration,
) logrus.Fields {
	metadata := containerAuditContextFromContext(ctx)
	fields := logrus.Fields{
		"action":      "container_exit",
		"operation":   safeContainerAuditToken(operation, "other", "stop", "remove", "startup"),
		"reason":      safeContainerAuditToken(reason, "other", "requested", "removed", "not_found", "runtime_error", "database_error", "stop_failed", "remove_error", "startup_failure"),
		"outcome":     safeContainerAuditToken(outcome, "error", "success", "error"),
		"duration_ms": safeAuditDuration(duration),
	}
	if metadata.flowID > 0 {
		fields["flow_id"] = metadata.flowID
	}
	if containerType, ok := safeContainerType(metadata.containerType); ok {
		fields["container_type"] = containerType
	}
	addMCPAuditCorrelation(fields, ctx)
	return fields
}

func containerCleanupAuditFields(
	ctx context.Context,
	outcome string,
	flowCount, attemptedCount, succeededCount, failedCount int,
	duration time.Duration,
) logrus.Fields {
	fields := logrus.Fields{
		"action":          "container_cleanup",
		"outcome":         safeContainerAuditToken(outcome, "error", "success", "error"),
		"flow_count":      nonNegativeAuditCount(flowCount),
		"attempted_count": nonNegativeAuditCount(attemptedCount),
		"succeeded_count": nonNegativeAuditCount(succeededCount),
		"failed_count":    nonNegativeAuditCount(failedCount),
		"duration_ms":     safeAuditDuration(duration),
	}
	addMCPAuditCorrelation(fields, ctx)
	return fields
}

func addMCPAuditCorrelation(fields logrus.Fields, ctx context.Context) {
	correlation := obs.AuditCorrelationFromContext(ctx)
	if correlation.Source != obs.AuditSourceMCP {
		return
	}
	for key, value := range correlation.Fields() {
		fields[key] = value
	}
}

func safeContainerType(containerType database.ContainerType) (database.ContainerType, bool) {
	switch containerType {
	case database.ContainerTypePrimary, database.ContainerTypeSecondary:
		return containerType, true
	default:
		return "", false
	}
}

func safeContainerAuditToken(value, fallback string, allowed ...string) string {
	value = strings.TrimSpace(value)
	for _, candidate := range allowed {
		if value == candidate {
			return value
		}
	}
	return fallback
}

func safeAuditDuration(duration time.Duration) int64 {
	if duration < 0 {
		return 0
	}
	return duration.Milliseconds()
}

func nonNegativeAuditCount(value int) int {
	if value < 0 {
		return 0
	}
	return value
}

func (dc *dockerClient) logContainerExitAudit(
	ctx context.Context,
	operation, reason string,
	startedAt time.Time,
	executionErr error,
) {
	outcome := "success"
	if executionErr != nil {
		outcome = "error"
	}
	fields := containerExitAuditFields(ctx, operation, reason, outcome, time.Since(startedAt))
	entry := logrus.WithContext(ctx).WithFields(fields)
	if executionErr != nil {
		entry.Warn("Container exit audit recorded")
		return
	}
	entry.Info("Container exit audit recorded")
}

func (dc *dockerClient) logContainerCleanupAudit(
	ctx context.Context,
	startedAt time.Time,
	flowCount, attemptedCount, succeededCount, failedCount int,
) {
	outcome := "success"
	if failedCount > 0 {
		outcome = "error"
	}
	logrus.WithContext(ctx).WithFields(containerCleanupAuditFields(
		ctx,
		outcome,
		flowCount,
		attemptedCount,
		succeededCount,
		failedCount,
		time.Since(startedAt),
	)).Info("Container cleanup audit recorded")
}
