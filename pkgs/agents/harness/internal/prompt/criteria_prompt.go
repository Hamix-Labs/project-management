package prompt

import (
	"fmt"
	"strings"
)

// ChecklistItem is one Done-criteria row for execute prompt injection.
type ChecklistItem struct {
	ID   string
	Text string
}

// InjectCriteria prepends the Done-criteria block before the operator's
// initial prompt. alreadyVerified is the set of criterion IDs proven
// passed in earlier retry attempts; when non-empty, those items render
// under a separate "Already verified" header and are omitted from the
// active checklist.
//
// reportPath is the absolute path for this cycle's criteria-report.json
// (under Options.ReportDir). When toolOnly is true (default product path),
// the agent must call hamix.submit_criteria_report instead of freeform Write.
//
//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func InjectCriteria(prompt string, items []ChecklistItem, reportPath string, alreadyVerified map[string]struct{}, toolOnly bool) string {
	if len(items) == 0 {
		return prompt
	}
	active := make([]ChecklistItem, 0, len(items))
	locked := make([]ChecklistItem, 0, len(alreadyVerified))
	for _, it := range items {
		if _, ok := alreadyVerified[it.ID]; ok {
			locked = append(locked, it)
			continue
		}
		active = append(active, it)
	}

	var criteria strings.Builder

	if len(locked) > 0 {
		criteria.WriteString("\n\n## Already verified (do not re-do)\n\n")
		criteria.WriteString("These criteria were proven passed in an earlier attempt. Do not undo or modify the work that satisfied them; do not include them in your report.\n\n")
		for _, it := range locked {
			criteria.WriteString(fmt.Sprintf("- [%s] %s\n", it.ID, it.Text))
		}
	}

	if len(active) == 0 {
		criteria.WriteString("\n\n## Done criteria (required)\n\nAll criteria are already verified. Re-run is a no-op; the worker will exit successfully.\n")
		return strings.TrimPrefix(criteria.String(), "\n\n") + "\n\n" + prompt
	}

	criteria.WriteString("\n\n## Done criteria (required)\n\n")
	if toolOnly {
		criteria.WriteString("You must satisfy every criterion below. When finished, call the MCP tool `hamix.submit_criteria_report` with one entry per active criterion (and optional `commits` for commits created in this execute visit).\n")
		criteria.WriteString("Do **not** freeform-Write `criteria-report.json` ΓÇö only the submit tool is accepted.\n")
		criteria.WriteString("Git discipline: create **new commits only** ΓÇö never amend, rebase, squash, or delete history; fix mistakes with a follow-up commit.\n")
		criteria.WriteString("claimed_done is your assertion that you completed the work. Criteria without verify commands are accepted from this report. Criteria with verify commands are checked later by matching each command's expected_outcome to worker-captured output (do not treat the claim as final for those).\n")
	} else {
		criteria.WriteString("You must satisfy every criterion below. When finished, write a JSON report at:\n")
		criteria.WriteString(fmt.Sprintf("`%s`\n\n", reportPath))
		criteria.WriteString("Schema:\n```json\n{\"schema_version\":1,\"criteria\":[{\"id\":\"<id>\",\"claimed_done\":true,\"evidence\":\"...\"}],\"commits\":[{\"sha\":\"<full-or-abbrev>\",\"branch\":\"optional\"}]}\n```\n")
		criteria.WriteString("Use only `schema_version`, `criteria`, and `commits` top-level fields ΓÇö no extra keys; put metadata in `evidence`.\n")
		criteria.WriteString("List commits **created in this execute visit** under `commits` (incremental is fine ΓÇö the worker accumulates them).\n")
		criteria.WriteString("Git discipline: create **new commits only** ΓÇö never amend, rebase, squash, or delete history; fix mistakes with a follow-up commit.\n")
		criteria.WriteString("claimed_done is your assertion that you completed the work. Criteria without verify commands are accepted from this report. Criteria with verify commands are checked later by matching each command's expected_outcome to worker-captured output (do not treat the claim as final for those).\n")
	}
	if len(locked) > 0 {
		criteria.WriteString("(Report only the criteria below; do NOT include already-verified IDs.)\n")
	}
	criteria.WriteString("\n")
	for _, it := range active {
		criteria.WriteString(fmt.Sprintf("- [%s] %s\n", it.ID, it.Text))
	}
	return strings.TrimPrefix(criteria.String(), "\n\n") + "\n\n" + prompt
}

// AppendExecuteHarnessFeedback appends execute-phase harness feedback when non-empty.
//
//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func AppendExecuteHarnessFeedback(prompt string, feedback string) string {
	feedback = strings.TrimSpace(feedback)
	if feedback == "" {
		return prompt
	}
	return prompt + "\n\n## Execute harness feedback\n\n" + feedback + "\n"
}
