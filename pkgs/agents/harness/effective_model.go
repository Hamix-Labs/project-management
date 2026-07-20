package harness

import "github.com/AlexsanderHamir/Hamix/pkgs/obs/calltrace"
import (
	"encoding/json"
	"log/slog"

	"github.com/AlexsanderHamir/Hamix/pkgs/agents/runner"
	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
	cyclesdomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcycles/domain"
)

func effectiveModelFromCycleMeta(r runner.Runner, task *taskcoredomain.Task, cycle *cyclesdomain.TaskCycle) string {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "agent.harness.effectiveModelFromCycleMeta",
		"task_id", task.ID, "cycle_id", cycleIDOrEmpty(cycle))
	if cycle != nil && len(cycle.MetaJSON) > 0 {
		var meta map[string]any
		if json.Unmarshal(cycle.MetaJSON, &meta) == nil {
			if v, ok := meta["cursor_model_effective"].(string); ok && v != "" {
				return v
			}
		}
	}
	req := runner.Request{
		TaskID:      task.ID,
		Prompt:      task.InitialPrompt,
		CursorModel: task.CursorModel,
	}
	if attr, ok := r.(runner.Attributor); ok {
		return attr.MetricsLabels(req)["model"]
	}
	if ml, ok := r.(runner.MetricsLabeler); ok {
		return ml.MetricsLabels(req)["model"]
	}
	return r.EffectiveModel(req)
}
