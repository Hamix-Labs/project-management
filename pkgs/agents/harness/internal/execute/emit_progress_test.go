package execute

import (
	"context"
	"testing"

	"github.com/AlexsanderHamir/Hamix/pkgs/agents/runner"
	cyclesdomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcycles/domain"
)

func TestPhasePorts_emitProgress_nilSafe(t *testing.T) {
	t.Parallel()
	var ports PhasePorts
	ports.emitProgress(context.Background(), "t", "c", &cyclesdomain.TaskCyclePhase{PhaseSeq: 1},
		runner.SetupProgressEvent(runner.ProgressRunStateSetupGit, "Captured git snapshot…"))
}

func TestPhasePorts_emitProgress_invokesCallback(t *testing.T) {
	t.Parallel()
	var got runner.ProgressEvent
	phase := &cyclesdomain.TaskCyclePhase{PhaseSeq: 2}
	ports := PhasePorts{
		EmitProgress: func(ctx context.Context, taskID, cycleID string, p *cyclesdomain.TaskCyclePhase, ev runner.ProgressEvent) {
			if taskID != "task" || cycleID != "cycle" || p != phase {
				t.Fatalf("unexpected args task=%q cycle=%q phase=%v", taskID, cycleID, p)
			}
			got = ev
		},
	}
	ports.emitProgress(context.Background(), "task", "cycle", phase,
		runner.SetupProgressEvent(runner.ProgressRunStateSetupPlan, "Planned Cursor session…"))
	if got.Subtype != runner.ProgressRunStateSetupPlan {
		t.Fatalf("subtype: got %q", got.Subtype)
	}
	if got.Tool != runner.ProgressToolHarnessSetup {
		t.Fatalf("tool: got %q", got.Tool)
	}
}
