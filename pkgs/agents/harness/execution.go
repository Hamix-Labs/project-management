package harness

import (
	"context"

	"github.com/AlexsanderHamir/Hamix/pkgs/agents/harness/internal/execute"
	"github.com/AlexsanderHamir/Hamix/pkgs/agents/runner"
	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
	cyclesdomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcycles/domain"
)

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func (h *Harness) executeSvc() *execute.Service {
	if h.execute == nil {
		h.execute = execute.NewService(execute.Deps{
			Store:     h.store,
			Git:       h.gitSvc(),
			ReportDir: h.opts.ReportDir,
		})
	}
	h.execute.SetReportDir(h.opts.ReportDir)
	return h.execute
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func (h *Harness) executePhasePorts(state *processState, opts cycleLoopOpts) execute.PhasePorts {
	return execute.PhasePorts{
		StartPhase: func(ctx context.Context, cycle *cyclesdomain.TaskCycle) (*cyclesdomain.TaskCyclePhase, bool) {
			return h.startExecutePhase(ctx, cycle, state)
		},
		PlanRun: func(ctx context.Context, task *taskcoredomain.Task, cycle *cyclesdomain.TaskCycle) (execute.RunPlan, error) {
			dec, err := h.planExecuteRun(ctx, task, cycle, state, opts)
			if err != nil {
				return execute.RunPlan{}, err
			}
			return executeRunPlanFromDecision(dec), nil
		},
		PlanFallback: func(ctx context.Context, task *taskcoredomain.Task, cycle *cyclesdomain.TaskCycle) execute.RunPlan {
			return executeRunPlanFromDecision(h.planExecuteResumeFallback(ctx, task, cycle, state, opts))
		},
		Invoke: func(ctx context.Context, task *taskcoredomain.Task, cycle *cyclesdomain.TaskCycle, exec *cyclesdomain.TaskCyclePhase, plan execute.RunPlan) (runner.Result, error) {
			return h.invokeRunnerWithTask(ctx, task, cycle, exec, decisionFromExecuteRunPlan(plan))
		},
		ConsumeOperatorCancel: h.consumeOperatorCancel,
		Publish:               h.publish,
		WorkingDir:            h.opts.WorkingDir,
		ReportDir:             h.opts.ReportDir,
		RepoRoot:              h.repoRootForGit(context.Background()),
		IsFreshOrFallback: func(mode string) bool {
			return mode == string(CursorResumeFresh) || mode == string(CursorResumeFallback)
		},
	}
}

//funclogmeasure:skip category=hot-path reason="Pure DTO mapping without I/O; callers own operation traces."
func executeRunPlanFromDecision(dec CursorResumeDecision) execute.RunPlan {
	return execute.RunPlan{
		Mode:            string(dec.Mode),
		ResumeSessionID: dec.ResumeSessionID,
		Prompt:          dec.Prompt,
	}
}

//funclogmeasure:skip category=hot-path reason="Pure DTO mapping without I/O; callers own operation traces."
func decisionFromExecuteRunPlan(plan execute.RunPlan) CursorResumeDecision {
	return CursorResumeDecision{
		Mode:            CursorResumeMode(plan.Mode),
		ResumeSessionID: plan.ResumeSessionID,
		Prompt:          plan.Prompt,
	}
}
