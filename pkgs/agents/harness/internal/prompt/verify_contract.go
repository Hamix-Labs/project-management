package prompt

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/AlexsanderHamir/Hamix/pkgs/obs/calltrace"
)

// VerifyCriterionLine is one active criterion for the verify-report contract.
type VerifyCriterionLine struct {
	ID       string
	Text     string
	Evidence string
}

// VerifyReportContract is the shared verify-report artifact instructions for
// fresh BuildVerifyPrompt and same-chat resume deltas (ADR-0085).
type VerifyReportContract struct {
	ReportPath             string
	LockedIDs              []string
	Criteria               []VerifyCriterionLine
	CommandEvidenceSection string
	GitContext             string
	DiffSection            string
	Feedback               string
}

const verifyReportSchema = `Schema: {"criteria":[{"id":"...","verified":true|false,"reasoning":"..."}]}`

// FormatVerifyReportContract renders path, schema, criteria, evidence, and diff.
//
//funclogmeasure:skip category=hot-path reason="Pure prompt compose without I/O."
func FormatVerifyReportContract(c VerifyReportContract) string {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "prompt.FormatVerifyReportContract",
		"criteria", len(c.Criteria), "locked", len(c.LockedIDs))
	var b strings.Builder
	if strings.TrimSpace(c.ReportPath) != "" {
		fmt.Fprintf(&b, "Write `%s` only.\n\n", c.ReportPath)
	}
	b.WriteString(verifyReportSchema)
	b.WriteString("\n\n")
	if len(c.LockedIDs) > 0 {
		b.WriteString("## Locked passes (do not re-evaluate)\n\n")
		b.WriteString("These criteria were verified in earlier attempts. Do NOT include them in your report.\n\n")
		for _, id := range c.LockedIDs {
			fmt.Fprintf(&b, "- [%s]\n", id)
		}
		b.WriteString("\n")
	}
	for _, it := range c.Criteria {
		fmt.Fprintf(&b, "- [%s] %s\n  execute claimed_done: true (assertion only)\n  execute evidence: %s\n",
			it.ID, it.Text, it.Evidence)
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
