package prompt

import "github.com/AlexsanderHamir/Hamix/pkgs/obs/calltrace"
import (
	"fmt"
	cyclesdomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcycles/domain"
	"log/slog"
	"sort"
	"strings"
)

// RecoveryKind selects the structured delta template for a resumed Cursor session.
type RecoveryKind string

const (
	RecoveryVerifyImplementation  RecoveryKind = "verify_implementation_fail"
	RecoveryCriteriaReportInvalid RecoveryKind = "criteria_report_invalid"
	RecoveryCriteriaReportMissing RecoveryKind = "criteria_report_missing"
	RecoveryProcessRestart        RecoveryKind = "process_restart"
	RecoveryOperatorRetryResume   RecoveryKind = "operator_retry_resume"
	RecoveryVerifyInfra           RecoveryKind = "verify_infra_retry"
	// RecoveryHumanPolish is Cursor --resume stdin for operator polish (not failure recovery).
	RecoveryHumanPolish RecoveryKind = "human_polish"
	// RecoveryHumanOpenPR is Cursor --resume stdin for approve-and-open-PR.
	RecoveryHumanOpenPR RecoveryKind = "human_open_pr"
)

// CriterionFailure is one failed criterion for structured verify recovery text.
type CriterionFailure struct {
	ID        string
	Reasoning string
	Verifier  string
}

// RecoveryContext carries harness state into delta-only stdin prompts (ADR-0031).
type RecoveryContext struct {
	Kind       RecoveryKind
	Phase      cyclesdomain.Phase
	CycleID    string
	AttemptSeq int64
	ReportPath string

	FailedCriteria   []CriterionFailure
	LockedCriteria   []string
	ReportParseErr   string
	ExpectedIDs      []string
	ScopeFiles       []string
	FailureClass     string
	FailureReason    string
	InterruptedPhase cyclesdomain.Phase
	GitPorcelain     string
	// ToolOnly selects MCP submit instructions (default) vs legacy Write.
	ToolOnly bool
	// Polish drives RecoveryHumanPolish deltas (ComposePolishDirective).
	Polish PolishNoticeInput
	// OpenPRKnownCommits drives RecoveryHumanOpenPR deltas.
	OpenPRKnownCommits []cyclesdomain.TaskCycleCommit
}

const (
	recoveryMaxTotalBytes       = 8 * 1024
	recoveryMaxReasoningRunes   = 2048
	recoverySectionContinuation = "## Continuation (Hamix attempt %d)"
)

// ComposeRecoveryDelta builds a self-contained follow-up prompt for --resume.
//
//funclogmeasure:skip category=hot-path reason="Pure prompt compose without I/O; caller logs byte metrics."
func ComposeRecoveryDelta(ctx RecoveryContext) string {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "prompt.ComposeRecoveryDelta",
		"kind", string(ctx.Kind), "phase", string(ctx.Phase))
	var b strings.Builder
	composeExecuteRecoveryDelta(&b, ctx)
	out := truncateRecoveryBytes(b.String(), recoveryMaxTotalBytes)
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "prompt.ComposeRecoveryDelta.done",
		"recovery_hint_kind", string(ctx.Kind), "recovery_hint_bytes", len(out))
	return out
}

//funclogmeasure:skip category=hot-path reason="Pure string builder; ComposeRecoveryDelta logs byte metrics."
func composeExecuteRecoveryDelta(b *strings.Builder, ctx RecoveryContext) {
	if ctx.Kind == RecoveryHumanPolish {
		composeHumanPolishRecoveryDelta(b, ctx)
		return
	}
	if ctx.Kind == RecoveryHumanOpenPR {
		b.WriteString(ComposeOpenPRDirective(&cyclesdomain.TaskCycle{ID: ctx.CycleID}, ctx.OpenPRKnownCommits))
		return
	}
	fmt.Fprintf(b, recoverySectionContinuation+"\n\n", ctx.AttemptSeq)
	b.WriteString("You are continuing the same Cursor session. Do not restart discovery or revert locked work.\n\n")
	b.WriteString("### What changed\n\n")
	switch ctx.Kind {
	case RecoveryVerifyImplementation:
		b.WriteString("Verification rejected the implementation. Address the failures below.\n\n")
		b.WriteString(FormatVerifyFailuresStructured(ctx.FailedCriteria))
	case RecoveryCriteriaReportInvalid:
		b.WriteString(ComposeCriteriaReportRecoveryDelta(ctx.ReportPath, ctx.ReportParseErr, ctx.ExpectedIDs, ctx.LockedCriteria, ctx.ToolOnly))
	case RecoveryCriteriaReportMissing:
		if ctx.ToolOnly {
			b.WriteString("The criteria self-report was not submitted via MCP.\n\n")
			b.WriteString("Call `hamix.submit_criteria_report` with every expected criterion ID. Do not freeform-Write the report file.\n\n")
		} else {
			b.WriteString("The criteria self-report file is missing.\n\n")
			if ctx.ReportPath != "" {
				fmt.Fprintf(b, "Write it at: `%s`\n\n", ctx.ReportPath)
			}
		}
		if len(ctx.ExpectedIDs) > 0 {
			b.WriteString("Expected criterion IDs: ")
			b.WriteString(strings.Join(ctx.ExpectedIDs, ", "))
			b.WriteString("\n\n")
		}
	case RecoveryProcessRestart:
		fmt.Fprintf(b, "The worker restarted during %s. Inspect the tree and continue from the last known good state.\n\n", ctx.InterruptedPhase)
		if ctx.FailureReason != "" {
			fmt.Fprintf(b, "Last failure: %s\n\n", ctx.FailureReason)
		}
	case RecoveryOperatorRetryResume:
		b.WriteString("Operator chose Resume from failure on a new task attempt.\n\n")
		if ctx.FailureClass != "" {
			fmt.Fprintf(b, "Failure class: %s\n", ctx.FailureClass)
		}
		if ctx.FailureReason != "" {
			fmt.Fprintf(b, "Failure reason: %s\n\n", ctx.FailureReason)
		}
		if len(ctx.ScopeFiles) > 0 {
			b.WriteString("Scope lock (do not modify outside this set):\n")
			for _, f := range ctx.ScopeFiles {
				fmt.Fprintf(b, "- %s\n", f)
			}
			b.WriteString("\n")
		}
	default:
		b.WriteString("Continue the in-progress execute work.\n\n")
	}
	b.WriteString("### Do this next\n\n")
	b.WriteString("1. Fix the issue described above.\n")
	if ctx.ReportPath != "" && ctx.Kind != RecoveryCriteriaReportInvalid && ctx.Kind != RecoveryCriteriaReportMissing {
		fmt.Fprintf(b, "2. Update `%s` for active criteria only.\n", ctx.ReportPath)
	}
	b.WriteString("\n### Do not\n\n")
	if len(ctx.LockedCriteria) > 0 {
		b.WriteString("- Re-do locked criteria: ")
		b.WriteString(strings.Join(ctx.LockedCriteria, ", "))
		b.WriteString("\n")
	}
	b.WriteString("- Amend, rebase, or squash commits from this cycle\n\n")
	if ctx.ReportPath != "" {
		b.WriteString("### Artifacts\n\n")
		fmt.Fprintf(b, "- criteria-report.json: `%s` (schema v1, claimed_done + evidence per active id)\n", ctx.ReportPath)
	}
}

//funclogmeasure:skip category=hot-path reason="Pure string builder; ComposeRecoveryDelta logs byte metrics."
func composeHumanPolishRecoveryDelta(b *strings.Builder, ctx RecoveryContext) {
	fmt.Fprintf(b, recoverySectionContinuation+"\n\n", ctx.AttemptSeq)
	b.WriteString("You are continuing the same Cursor session for **human polish**. ")
	b.WriteString("Do not restart discovery or revert locked work.\n\n")
	cycle := &cyclesdomain.TaskCycle{ID: ctx.CycleID, AttemptSeq: ctx.AttemptSeq}
	b.WriteString(ComposePolishDirective(cycle, ctx.Polish))
	b.WriteString("### Do this next\n\n")
	b.WriteString("1. Apply the operator polish instructions above.\n")
	if ctx.Polish.SkipVerify {
		b.WriteString("2. Do not re-claim or re-hunt locked/prior criteria; your execute claim ends this attempt.\n")
	} else if ctx.ReportPath != "" {
		fmt.Fprintf(b, "2. Update `%s` for active (flagged/new) criteria only.\n", ctx.ReportPath)
	}
	b.WriteString("\n### Do not\n\n")
	if len(ctx.LockedCriteria) > 0 {
		b.WriteString("- Re-do locked criteria: ")
		b.WriteString(strings.Join(ctx.LockedCriteria, ", "))
		b.WriteString("\n")
	}
	b.WriteString("- Amend, rebase, or squash commits from this cycle\n")
	b.WriteString("- Re-audit the original task as if it failed\n\n")
	if ctx.ReportPath != "" && !ctx.Polish.SkipVerify {
		b.WriteString("### Artifacts\n\n")
		fmt.Fprintf(b, "- criteria-report.json: `%s` (schema v1, claimed_done + evidence per active id)\n", ctx.ReportPath)
	}
}

// FormatVerifyFailuresStructured renders per-criterion failure blocks for resume deltas.
//
//funclogmeasure:skip category=hot-path reason="Pure string format without I/O."
func FormatVerifyFailuresStructured(failures []CriterionFailure) string {
	if len(failures) == 0 {
		return ""
	}
	var b strings.Builder
	b.WriteString("### Verification failures\n\n")
	for _, f := range failures {
		verifier := f.Verifier
		if verifier == "" {
			verifier = "verify"
		}
		fmt.Fprintf(&b, "- **[%s]** (%s)\n", f.ID, verifier)
		if r := strings.TrimSpace(f.Reasoning); r != "" {
			fmt.Fprintf(&b, "  Reasoning: %s\n", truncateRecoveryRunes(r, recoveryMaxReasoningRunes))
		}
		b.WriteString("  Required: address this before claiming done again.\n")
	}
	b.WriteString("\n")
	return b.String()
}

// ComposeCriteriaReportRecoveryDelta builds the invalid-report recovery section.
//
//funclogmeasure:skip category=hot-path reason="Pure string format without I/O."
func ComposeCriteriaReportRecoveryDelta(reportPath, parseErr string, expected, locked []string, toolOnly bool) string {
	var b strings.Builder
	b.WriteString("The criteria self-report JSON is invalid or incomplete.\n\n")
	if parseErr != "" {
		fmt.Fprintf(&b, "Parse error: %s\n\n", parseErr)
	}
	if toolOnly {
		b.WriteString("Re-submit via MCP tool `hamix.submit_criteria_report`. Do not freeform-Write the report file.\n\n")
	} else if reportPath != "" {
		fmt.Fprintf(&b, "Fix the file at: `%s`\n\n", reportPath)
	}
	if len(expected) > 0 {
		sort.Strings(expected)
		b.WriteString("Expected criterion IDs: ")
		b.WriteString(strings.Join(expected, ", "))
		b.WriteString("\n\n")
	}
	if !toolOnly {
		b.WriteString("Schema:\n```json\n{\"criteria\":[{\"id\":\"<id>\",\"claimed_done\":true,\"evidence\":\"...\"}]}\n```\n\n")
	}
	if len(locked) > 0 {
		b.WriteString("Locked criteria are already satisfied; omit them from the report.\n\n")
	}
	return b.String()
}

//funclogmeasure:skip category=hot-path reason="Pure truncation helper without I/O."
func truncateRecoveryRunes(s string, max int) string {
	if max <= 0 {
		return ""
	}
	runes := []rune(s)
	if len(runes) <= max {
		return s
	}
	return string(runes[:max]) + "ΓÇª"
}

//funclogmeasure:skip category=hot-path reason="Pure truncation helper without I/O."
func truncateRecoveryBytes(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "ΓÇª"
}
