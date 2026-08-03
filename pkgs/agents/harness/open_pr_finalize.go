package harness

import (
	"context"
	"encoding/json"
	"log/slog"
	"strings"

	"github.com/AlexsanderHamir/Hamix/pkgs/agents/harness/internal/orchestration"
	"github.com/AlexsanderHamir/Hamix/pkgs/agents/sidecar"
	"github.com/AlexsanderHamir/Hamix/pkgs/obs/calltrace"
	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
	cyclesdomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcycles/domain"
	taskeventsdomain "github.com/AlexsanderHamir/Hamix/pkgs/taskevents/domain"
)

func (h *Harness) runCycleLoopFinalizeOpenPR(
	parentCtx context.Context,
	task *taskcoredomain.Task,
	cycle *cyclesdomain.TaskCycle,
	state *processState,
) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "agent.harness.Harness.runCycleLoopFinalizeOpenPR",
		"task_id", task.ID, "cycle_id", cycle.ID)
	if state.agentMCP.enabled {
		if err := sidecar.RequirePullRequestSubmitReceipt(h.opts.ReportDir, cycle.ID, state.agentMCP.nonce); err != nil {
			slog.Warn("agent harness open_pr receipt missing", "cmd", calltrace.LogCmd,
				"operation", "agent.harness.Harness.runCycleLoopFinalizeOpenPR.receipt_err",
				"task_id", task.ID, "cycle_id", cycle.ID, "err", err)
			h.bestEffortTerminate(parentCtx, state, task.ID, cyclesdomain.CycleStatusFailed, "open_pr_receipt_missing")
			_ = h.transitionTask(parentCtx, task.ID, taskcoredomain.StatusFailed, "open_pr_receipt_missing")
			return
		}
	}
	rep, err := sidecar.ParsePullRequestReport(h.opts.ReportDir, cycle.ID)
	if err != nil {
		slog.Warn("agent harness open_pr report missing", "cmd", calltrace.LogCmd,
			"operation", "agent.harness.Harness.runCycleLoopFinalizeOpenPR.report_err",
			"task_id", task.ID, "cycle_id", cycle.ID, "err", err)
		h.bestEffortTerminate(parentCtx, state, task.ID, cyclesdomain.CycleStatusFailed, "open_pr_report_missing")
		_ = h.transitionTask(parentCtx, task.ID, taskcoredomain.StatusFailed, "open_pr_report_missing")
		return
	}
	effects := orchestration.FinalizeEffects{
		CycleStatus:    cyclesdomain.CycleStatusSucceeded,
		TaskStatus:     taskcoredomain.StatusPrReady,
		PullRequestURL: strings.TrimSpace(rep.URL),
	}
	if ok := h.applyFinalizeEffects(parentCtx, task, cycle, state, effects); !ok {
		slog.Error("agent harness open_pr finalize effects incomplete", "cmd", calltrace.LogCmd,
			"operation", "agent.harness.Harness.runCycleLoopFinalizeOpenPR.effects_incomplete",
			"task_id", task.ID, "cycle_id", cycle.ID)
		if state.cycle.cycleStarted {
			h.bestEffortTerminate(parentCtx, state, task.ID, cyclesdomain.CycleStatusFailed, "finalize_effects_failed")
		}
		return
	}
	h.stampStackLayerPRs(parentCtx, task, rep)
	h.emitPROpened(parentCtx, task.ID, cycle.ID, rep)
}

func (h *Harness) stampStackLayerPRs(ctx context.Context, task *taskcoredomain.Task, rep sidecar.PullRequestReport) {
	if h == nil || h.store == nil || task == nil || task.WorktreeID == nil {
		return
	}
	if len(rep.Layers) == 0 {
		return
	}
	urlByBranch := make(map[string]string, len(rep.Layers))
	for _, layer := range rep.Layers {
		head := strings.TrimSpace(layer.Head)
		url := strings.TrimSpace(layer.URL)
		if head == "" || url == "" {
			continue
		}
		urlByBranch[head] = url
	}
	if len(urlByBranch) == 0 {
		return
	}
	if err := h.store.ApplyStackPullRequestURLs(ctx, strings.TrimSpace(*task.WorktreeID), urlByBranch); err != nil {
		slog.Warn("agent harness stack PR URL stamp failed", "cmd", calltrace.LogCmd,
			"operation", "agent.harness.Harness.stampStackLayerPRs",
			"task_id", task.ID, "err", err)
	}
}

type prOpenedPayload struct {
	URL     string `json:"url"`
	Number  int    `json:"number,omitempty"`
	Title   string `json:"title,omitempty"`
	Base    string `json:"base,omitempty"`
	Head    string `json:"head,omitempty"`
	CycleID string `json:"cycle_id,omitempty"`
}

func (h *Harness) emitPROpened(ctx context.Context, taskID, cycleID string, rep sidecar.PullRequestReport) {
	if h == nil || h.store == nil {
		return
	}
	raw, err := json.Marshal(prOpenedPayload{
		URL:     rep.URL,
		Number:  rep.Number,
		Title:   rep.Title,
		Base:    rep.Base,
		Head:    rep.Head,
		CycleID: cycleID,
	})
	if err != nil {
		slog.Warn("pr_opened: marshal failed", "cmd", calltrace.LogCmd,
			"operation", "agent.harness.Harness.emitPROpened", "task_id", taskID, "err", err)
		return
	}
	if err := h.store.AppendTaskEvent(ctx, taskID, taskeventsdomain.EventPROpened, taskcoredomain.ActorAgent, raw); err != nil {
		slog.Warn("pr_opened: append event failed", "cmd", calltrace.LogCmd,
			"operation", "agent.harness.Harness.emitPROpened", "task_id", taskID, "err", err)
	}
}
