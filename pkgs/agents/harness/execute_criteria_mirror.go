package harness

import "github.com/AlexsanderHamir/Hamix/pkgs/tasks/calltrace"
import (
	"context"
	"errors"
	"log/slog"

	"github.com/AlexsanderHamir/Hamix/pkgs/agents/harness/internal/reports"
	cyclesdomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcycles/domain"
)

func (h *Harness) bestEffortMirrorExecuteCriteria(
	ctx context.Context,
	cycleID string,
	state *processState,
) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "agent.harness.bestEffortMirrorExecuteCriteria",
		"cycle_id", cycleID)
	if !state.verify.verifySnap.Enabled || len(state.verify.verifySnap.Criteria) == 0 {
		return
	}
	selfReport, err := reports.ParseCriteriaReportPartial(h.opts.ReportDir, cycleID)
	if err != nil {
		if !errors.Is(err, ErrCriteriaReportMissing) {
			slog.Warn("agent harness execute criteria mirror parse failed",
				"cmd", calltrace.LogCmd, "operation", "agent.harness.bestEffortMirrorExecuteCriteria.parse_err",
				"cycle_id", cycleID, "err", err)
		}
		return
	}
	if uerr := h.persistCriteriaReports(ctx, cycleID, cyclesdomain.ExecuteCriteriaReportAttemptSeq,
		state.verify.verifySnap.Criteria, state.verify.previouslyPassed, selfReport); uerr != nil {
		state.verify.mirrorDegraded = true
		slog.Warn("agent harness execute criteria mirror upsert failed",
			"cmd", calltrace.LogCmd, "operation", "agent.harness.bestEffortMirrorExecuteCriteria.upsert_err",
			"cycle_id", cycleID, "err", uerr)
	}
}
