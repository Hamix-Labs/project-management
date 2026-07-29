package prompt

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/AlexsanderHamir/Hamix/pkgs/obs/calltrace"
)

// VerifyCriterionLine is one command-backed criterion for the verify-report contract.
type VerifyCriterionLine struct {
	ID       string
	Text     string
	Evidence string
}

// VerifyReportContract is the shared verify-report artifact instructions for
// fresh BuildVerifyPrompt and same-chat resume deltas (ADR-0085 / ADR-0090).
type VerifyReportContract struct {
	ReportPath             string
	LockedIDs              []string
	Criteria               []VerifyCriterionLine
	CommandEvidenceSection string
	GitContext             string
	DiffSection            string
	Feedback               string
	// ToolOnly requires hamix.submit_verify_report (default product path).
	// When false, legacy freeform Write instructions are used (kill-switch).
	ToolOnly bool
}

const verifyReportSchema = `Schema: {"criteria":[{"id":"...","verified":true|false,"reasoning":"..."}]}`

// FormatVerifyReportContract renders path, schema, command evidence, and diff.
// Judgment is expected_outcome Γåö shell output only (ADR-0090).
//
//funclogmeasure:skip category=hot-path reason="Pure prompt compose without I/O."
func FormatVerifyReportContract(c VerifyReportContract) string {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "prompt.FormatVerifyReportContract",
		"criteria", len(c.Criteria), "locked", len(c.LockedIDs), "tool_only", c.ToolOnly)
	var b strings.Builder
	if c.ToolOnly {
		b.WriteString("Call the MCP tool `hamix.submit_verify_report` with one entry per command-backed criterion below.\n")
		b.WriteString("For each criterion, set `verified` iff every attached command's `expected_outcome` is satisfied by that command's captured output.\n")
		b.WriteString("Put the outcomeΓåöoutput interpretation in `reasoning` (it becomes criterion evidence).\n")
		b.WriteString("Do **not** freeform-Write `verify-report.json` ΓÇö only the submit tool is accepted.\n")
		b.WriteString("Do **not** re-judge criterion text or execute evidence.\n\n")
	} else if strings.TrimSpace(c.ReportPath) != "" {
		fmt.Fprintf(&b, "Write `%s` only.\n\n", c.ReportPath)
		b.WriteString(verifyReportSchema)
		b.WriteString("\n")
		b.WriteString("For each criterion, set `verified` iff every attached command's `expected_outcome` is satisfied by that command's captured output.\n")
		b.WriteString("Put the outcomeΓåöoutput interpretation in `reasoning`. Do not re-judge criterion text or execute evidence.\n\n")
	} else {
		b.WriteString(verifyReportSchema)
		b.WriteString("\n\n")
	}
	if len(c.LockedIDs) > 0 {
		b.WriteString("## Locked passes (do not re-evaluate)\n\n")
		b.WriteString("These criteria were verified in earlier attempts. Do NOT include them in your report.\n\n")
		for _, id := range c.LockedIDs {
			fmt.Fprintf(&b, "- [%s]\n", id)
		}
		b.WriteString("\n")
	}
	if len(c.Criteria) > 0 {
		b.WriteString("## Command-backed criteria (report these ids)\n\n")
		for _, it := range c.Criteria {
			fmt.Fprintf(&b, "- [%s] %s\n", it.ID, it.Text)
		}
		b.WriteString("\n")
	}
	if sec := strings.TrimSpace(c.CommandEvidenceSection); sec != "" {
		b.WriteString(sec)
	}
	if git := strings.TrimSpace(c.GitContext); git != "" {
		b.WriteString(git)
	}
	if diff := strings.TrimSpace(c.DiffSection); diff != "" {
		b.WriteString("\nDiff:\n")
		b.WriteString(diff)
	}
	out := b.String()
	if fb := strings.TrimSpace(c.Feedback); fb != "" {
		out = AppendVerifyFeedback(out, fb)
	}
	return out
}
