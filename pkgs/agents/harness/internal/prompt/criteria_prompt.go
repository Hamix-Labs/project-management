package prompt

import (
	"fmt"
	"strings"
)

// ChecklistCommand is one operator-authored verify command for execute prompt injection.
type ChecklistCommand struct {
	Command         string
	ExpectedOutcome string
}

// ChecklistItem is one Done-criteria row for execute prompt injection.
type ChecklistItem struct {
	ID       string
	Text     string
	Commands []ChecklistCommand
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
		criteria.WriteString("You must satisfy every criterion below. When finished, call the MCP tool `hamix.submit_criteria_report` with one entry per active criterion.\n")
		criteria.WriteString("Do **not** freeform-Write `criteria-report.json` — only the submit tool is accepted.\n")
		criteria.WriteString("Git discipline: stage with Shell `git add`, then create **new commits only** via `hamix.commit` — never Shell `git commit`, amend, rebase, squash, or delete history; fix mistakes with a follow-up `hamix.commit`.\n")
		criteria.WriteString("For each criterion that lists verify commands: run those commands in the worktree and confirm the output matches each expected_outcome before claiming done. Put a short summary of the work and command results in `evidence`. Do not set claimed_done true if a required command fails that check.\n")
		criteria.WriteString("claimed_done is accepted by the harness as final (no separate verify phase). Only claim done when the criterion is actually satisfied.\n")
	} else {
		criteria.WriteString("You must satisfy every criterion below. When finished, write a JSON report at:\n")
		criteria.WriteString(fmt.Sprintf("`%s`\n\n", reportPath))
		criteria.WriteString("Schema:\n```json\n{\"schema_version\":1,\"criteria\":[{\"id\":\"<id>\",\"claimed_done\":true,\"evidence\":\"...\"}]}\n```\n")
		criteria.WriteString("Use only `schema_version` and `criteria` top-level fields — no extra keys; put metadata in `evidence`.\n")
		criteria.WriteString("Git discipline: stage with Shell `git add`, then create **new commits only** via `hamix.commit` — never Shell `git commit`, amend, rebase, squash, or delete history.\n")
		criteria.WriteString("For each criterion that lists verify commands: run those commands in the worktree and confirm the output matches each expected_outcome before claiming done. Put a short summary of the work and command results in `evidence`. Do not set claimed_done true if a required command fails that check.\n")
		criteria.WriteString("claimed_done is accepted by the harness as final (no separate verify phase). Only claim done when the criterion is actually satisfied.\n")
	}
	if len(locked) > 0 {
		criteria.WriteString("(Report only the criteria below; do NOT include already-verified IDs.)\n")
	}
	criteria.WriteString("\n")
	for _, it := range active {
		criteria.WriteString(fmt.Sprintf("- [%s] %s\n", it.ID, it.Text))
		writeCriterionCommands(&criteria, it.Commands)
	}
	return strings.TrimPrefix(criteria.String(), "\n\n") + "\n\n" + prompt
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func writeCriterionCommands(b *strings.Builder, cmds []ChecklistCommand) {
	if len(cmds) == 0 {
		return
	}
	b.WriteString("  Verify commands (run before claiming done):\n")
	for i, cmd := range cmds {
		b.WriteString(fmt.Sprintf("  %d. `%s`\n", i+1, cmd.Command))
		if outcome := strings.TrimSpace(cmd.ExpectedOutcome); outcome != "" {
			b.WriteString(fmt.Sprintf("     Expected outcome: %s\n", outcome))
		}
	}
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
