package execute

import (
	"github.com/AlexsanderHamir/Hamix/pkgs/agents/harness/internal/reports"
)

// ProbeCriteriaReport validates criteria-report.json for active checklist ids.
// Returns a parse-error string for Cursor recovery hints (empty when OK/skipped).
//
//funclogmeasure:skip category=hot-path reason="Lightweight probe; execute/verify chokepoints emit operation trace."
func ProbeCriteriaReport(reportDir, cycleID string, expectedActiveIDs map[string]struct{}) string {
	return ProbeCriteriaReportWithReceipt(reportDir, cycleID, expectedActiveIDs, false, "")
}

// ProbeCriteriaReportWithReceipt is like ProbeCriteriaReport but, when
// requireReceipt is true, also requires a matching MCP submit receipt before
// accepting the JSON report (tool-only path).
//
//funclogmeasure:skip category=hot-path reason="Lightweight probe; execute/verify chokepoints emit operation trace."
func ProbeCriteriaReportWithReceipt(reportDir, cycleID string, expectedActiveIDs map[string]struct{}, requireReceipt bool, nonce string) string {
	if len(expectedActiveIDs) == 0 {
		return ""
	}
	if requireReceipt {
		if err := reports.RequireCriteriaSubmitReceipt(reportDir, cycleID, nonce); err != nil {
			return err.Error()
		}
	}
	if _, err := reports.ParseCriteriaReport(reportDir, cycleID, expectedActiveIDs); err != nil {
		return err.Error()
	}
	return ""
}
