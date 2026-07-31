package verify

import (
	"context"
	"time"

	"github.com/AlexsanderHamir/Hamix/pkgs/agents/runner"
	checklistdomain "github.com/AlexsanderHamir/Hamix/pkgs/taskchecklist/domain"
)

// Hooks wires harness-owned side effects (SSE, metrics, progress persistence).
type Hooks struct {
	Publish         func(taskID, cycleID string)
	PersistProgress func(ctx context.Context, taskID, cycleID string, phaseSeq int64, ev runner.ProgressEvent)
	RecordVerdict   func(kind checklistdomain.VerifierKind, passed bool)
	ObserveDuration func(d time.Duration)
}
