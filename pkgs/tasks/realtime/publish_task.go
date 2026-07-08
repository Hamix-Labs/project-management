package realtime

import (
	"context"
	"fmt"
)

// TaskLoader loads the post-commit task row for SSE enrichment (ADR-0026 S2).
// Returns an opaque wire payload (typically *domain.Task from callers).
type TaskLoader func(ctx context.Context, taskID string) (any, error)

// PublishEnrichedTaskUpdated loads the task and publishes task_updated with Data.
// Call only after the store mutation succeeds (ADR-0026 S1). pub may be nil (no-op).
//
//funclogmeasure:skip category=hot-path reason="Shared publish helper; operation trace is emitted by handler and agentworker callers."
func PublishEnrichedTaskUpdated(ctx context.Context, pub Publisher, load TaskLoader, taskID string) error {
	if pub == nil || taskID == "" {
		return nil
	}
	task, err := load(ctx, taskID)
	if err != nil {
		return fmt.Errorf("publish enriched task_updated: %w", err)
	}
	pub.Publish(Event{
		Type: TaskUpdated,
		ID:   taskID,
		Data: task,
	})
	return nil
}
