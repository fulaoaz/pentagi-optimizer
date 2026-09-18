package tools

import (
	"context"
	"time"

	"pentagi/pkg/database"
	obs "pentagi/pkg/observability"

	"github.com/sirupsen/logrus"
)

func mcpContainerLifecycleAuditFields(
	ctx context.Context,
	flowID int64,
	operation, outcome string,
	duration time.Duration,
) (logrus.Fields, bool) {
	correlation := obs.AuditCorrelationFromContext(ctx)
	if correlation.Source != obs.AuditSourceMCP {
		return nil, false
	}

	fields := logrus.Fields{
		"action":         "mcp_container_lifecycle",
		"flow_id":        flowID,
		"container_type": database.ContainerTypePrimary,
		"operation":      operation,
		"outcome":        outcome,
		"duration_ms":    duration.Milliseconds(),
	}
	for key, value := range correlation.Fields() {
		fields[key] = value
	}
	return fields, true
}

func (fte *flowToolsExecutor) logMCPContainerLifecycle(
	ctx context.Context,
	operation string,
	startedAt time.Time,
	executionErr error,
) {
	outcome := "success"
	if executionErr != nil {
		outcome = "error"
	}

	fields, ok := mcpContainerLifecycleAuditFields(ctx, fte.flowID, operation, outcome, time.Since(startedAt))
	if !ok {
		return
	}
	logrus.WithContext(ctx).WithFields(fields).Info("MCP container lifecycle action completed")
}

func containerRecoveryAuditFields(
	ctx context.Context,
	flowID int64,
	recoveryReason, outcome string,
	duration time.Duration,
) logrus.Fields {
	fields := logrus.Fields{
		"action":          "container_recovery",
		"flow_id":         flowID,
		"container_type":  database.ContainerTypePrimary,
		"recovery_reason": recoveryReason,
		"outcome":         outcome,
		"duration_ms":     duration.Milliseconds(),
	}

	correlation := obs.AuditCorrelationFromContext(ctx)
	if correlation.Source == obs.AuditSourceMCP {
		for key, value := range correlation.Fields() {
			fields[key] = value
		}
	}
	return fields
}

func (fte *flowToolsExecutor) logContainerRecovery(
	ctx context.Context,
	recoveryReason string,
	startedAt time.Time,
	executionErr error,
) {
	outcome := "success"
	if executionErr != nil {
		outcome = "error"
	}

	entry := logrus.WithContext(ctx).WithFields(containerRecoveryAuditFields(
		ctx,
		fte.flowID,
		recoveryReason,
		outcome,
		time.Since(startedAt),
	))
	if executionErr != nil {
		entry.Warn("Container recovery failed")
		return
	}
	entry.Info("Container recovery completed")
}
