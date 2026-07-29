package harness

import (
	"github.com/AlexsanderHamir/Hamix/pkgs/agents/harness/internal/execute"
)

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func expectedActiveCriterionIDs(state *processState) map[string]struct{} {
	expected := make(map[string]struct{}, len(state.verify.verifySnap.Criteria))
	for _, it := range state.verify.verifySnap.Criteria {
		if _, locked := state.verify.lockedPasses[it.ID]; locked {
			continue
		}
		expected[it.ID] = struct{}{}
	}
	return expected
}

// probeCriteriaReport validates criteria-report.json for active checklist ids and
// records any parse error on state for Cursor recovery hints (ADR-0031).
//
//funclogmeasure:skip category=hot-path reason="Lightweight probe; execute/verify chokepoints emit operation trace."
func (h *Harness) probeCriteriaReport(state *processState, cycleID string) {
	state.verify.reportParseErr = ""
	if !state.verify.verifySnap.Enabled || len(state.verify.verifySnap.Criteria) == 0 {
		return
	}
	requireReceipt := state.agentMCP.enabled
	state.verify.reportParseErr = execute.ProbeCriteriaReportWithReceipt(
		h.opts.ReportDir,
		cycleID,
		expectedActiveCriterionIDs(state),
		requireReceipt,
		state.agentMCP.nonce,
	)
}
