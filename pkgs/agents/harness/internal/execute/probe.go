package execute

import (
	"github.com/AlexsanderHamir/Hamix/pkgs/agents/harness/internal/reports"
)

// ProbeCriteriaReport validates criteria-report.json for active checklist ids.
// Returns a parse-error string for Cursor recovery hints (empty when OK/skipped).
//
//funclogmeasure:skip category=hot-path reason="Lightweight probe; execute/verify chokepoints emit operation trace."
func ProbeCriteriaReport(reportDir, cycleID string, expectedActiveIDs map[string]struct{}) string {
	if len(expectedActiveIDs) == 0 {
		return ""
	}
	if _, err := reports.ParseCriteriaReport(reportDir, cycleID, expectedActiveIDs); err != nil {
		return err.Error()
	}
	return ""
}
