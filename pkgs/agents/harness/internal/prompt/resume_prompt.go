package prompt

import (
	"fmt"
	"strings"

	cyclesdomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcycles/domain"
)

// PolishCriterion is one checklist row referenced in a polish notice.
type PolishCriterion struct {
	ID   string
	Text string
}

// PolishNoticeInput drives ComposePolishDirective / AppendPolishNotice sections.
type PolishNoticeInput struct {
	Instructions string
	SkipVerify   bool
	Flagged      []PolishCriterion
	New          []PolishCriterion
	KnownCommits []cyclesdomain.TaskCycleCommit
}

// AppendOperatorRetryResumeNotice is for cross-cycle "Resume from failure" attempts.
// Unlike AppendResumeNotice (ADR-0006 in-process restart), this cycle is new while
// git work and indexed commits may carry over from the parent attempt.
//
//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func AppendOperatorRetryResumeNotice(prompt string, cycle *cyclesdomain.TaskCycle, parentCommits []cyclesdomain.TaskCycleCommit) string {
	if cycle == nil {
		return prompt
	}
	var b strings.Builder
	b.WriteString("## Operator retry ΓÇö resume from failure\n\n")
	b.WriteString("This is a **new execution attempt** continuing work from a failed prior attempt ")
	b.WriteString(fmt.Sprintf("(new cycle_id=%s).\n\n", cycle.ID))
	b.WriteString("Before changing anything:\n")
	b.WriteString("1. Inspect the working tree you were given (`git status`, read relevant files).\n")
	b.WriteString("2. Continue from that state; do not revert work that satisfies locked criteria below.\n")
	if block := FormatKnownCommitsForResume(parentCommits); block != "" {
		b.WriteString("3. ")
		b.WriteString(strings.TrimSpace(block))
		b.WriteString("Those commits are already indexed for this task — create only **new** commits via `hamix.commit` (do not use Shell `git commit`).\n")
		b.WriteString("4. A clean tree does **not** mean the task succeeded — complete remaining criteria and write the criteria report.\n")
	} else {
		b.WriteString("3. A clean tree does **not** mean the task succeeded ΓÇö complete remaining criteria and write the criteria report.\n")
	}
	b.WriteString("\n")
	return b.String() + prompt
}

// ComposePolishDirective builds the human-polish execute directive (shared by full
// prompts and Cursor --resume deltas). Optional Flagged / New / SkipVerify sections
// appear only when relevant ΓÇö one composer for all polish combos.
//
//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func ComposePolishDirective(cycle *cyclesdomain.TaskCycle, in PolishNoticeInput) string {
	if cycle == nil {
		return ""
	}
	var b strings.Builder
	b.WriteString("## Human polish ΓÇö refine completed work\n\n")
	b.WriteString("This is a **new execution attempt** after the prior attempt succeeded and entered human review ")
	b.WriteString(fmt.Sprintf("(new cycle_id=%s).\n\n", cycle.ID))
	b.WriteString("The prior attempt's checklist was accepted into review (independent verification had approved ")
	b.WriteString("those criteria where applicable). The human is now requesting **polishments**. ")
	b.WriteString("This is not failure recovery and not a worker restart.\n\n")
	b.WriteString("You are continuing the same Cursor conversation. Do not rediscover or re-audit the original task. ")
	b.WriteString("Change only what the polish instructions and any flagged/new criteria below require ΓÇö ")
	b.WriteString("not a full rewrite unless the instructions say so.\n\n")

	instructions := strings.TrimSpace(in.Instructions)
	b.WriteString("### Operator polish instructions (authoritative)\n\n")
	if instructions != "" {
		b.WriteString(instructions)
		b.WriteString("\n\n")
	} else {
		b.WriteString("(No freeform instructions were provided ΓÇö follow the criterion sections below.)\n\n")
	}

	if len(in.Flagged) > 0 {
		b.WriteString("### Human-flagged incorrect criteria\n\n")
		b.WriteString("The reviewer marked these criteria as **not done correctly**. Prior evidence and completions for them are invalid for this attempt.\n")
		b.WriteString("You must fix the underlying work, re-claim them in your criteria report, and expect independent verification.\n\n")
		for _, c := range in.Flagged {
			b.WriteString(fmt.Sprintf("- [%s] %s\n", c.ID, c.Text))
		}
		b.WriteString("\n")
	}

	if len(in.New) > 0 {
		b.WriteString("### Newly added criteria\n\n")
		b.WriteString("These criteria were added during polish. Implement them from scratch; they have never been verified.\n\n")
		for _, c := range in.New {
			b.WriteString(fmt.Sprintf("- [%s] %s\n", c.ID, c.Text))
		}
		b.WriteString("\n")
	}

	if in.SkipVerify {
		b.WriteString("### Verification mode for this polish\n\n")
		b.WriteString("This polish has **no flagged or newly added criteria**. After execute finishes, the harness will **skip the verify phase**.\n")
		b.WriteString("Prior criteria remain accepted ΓÇö do not re-claim or re-hunt them.\n")
		b.WriteString("Your execute-phase claim that the polish instructions are satisfied is sufficient for this attempt to return to human review.\n")
		b.WriteString("Still inspect the tree, change only what the instructions require, commit new work normally, and do not undo prior good work.\n\n")
	} else {
		b.WriteString("### Verification mode for this polish\n\n")
		b.WriteString("Criteria listed as already verified must remain satisfied — do not undo them.\n")
		b.WriteString("Apply the polish instructions; fix flagged criteria; implement newly added criteria.\n")
		b.WriteString("For active criteria with verify commands, run those commands and confirm expected_outcome before claiming done.\n")
		b.WriteString("Write a complete criteria report for active criteria only.\n\n")
	}

	b.WriteString("### Before changing anything\n\n")
	b.WriteString("1. Inspect the working tree (`git status`, read relevant files).\n")
	b.WriteString("2. Keep prior good work; change only what the polish instructions and flagged/new criteria require.\n")
	if block := FormatKnownCommitsForResume(in.KnownCommits); block != "" {
		b.WriteString("3. ")
		b.WriteString(strings.TrimSpace(block))
		b.WriteString("Those commits are already indexed — create only **new** commits via `hamix.commit` (do not use Shell `git commit`).\n")
		if in.SkipVerify {
			b.WriteString("4. Apply the polish instructions; your execute claim ends this attempt (no verify phase).\n")
		} else {
			b.WriteString("4. Re-satisfy active criteria after your changes and write the criteria report.\n")
		}
	} else if in.SkipVerify {
		b.WriteString("3. Apply the polish instructions; your execute claim ends this attempt (no verify phase).\n")
	} else {
		b.WriteString("3. Re-satisfy active criteria after your changes and write the criteria report.\n")
	}
	b.WriteString("\n")
	return b.String()
}

// AppendPolishNotice prepends ComposePolishDirective for full execute prompts.
// This is not a failure-resume path.
//
//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func AppendPolishNotice(prompt string, cycle *cyclesdomain.TaskCycle, in PolishNoticeInput) string {
	directive := ComposePolishDirective(cycle, in)
	if directive == "" {
		return prompt
	}
	return directive + prompt
}

// AppendResumeNotice prepends an in-process worker resume notice.
//
//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func AppendResumeNotice(prompt string, cycle *cyclesdomain.TaskCycle, interruptedPhase cyclesdomain.Phase, knownCommits []cyclesdomain.TaskCycleCommit) string {
	if cycle == nil {
		return prompt
	}
	var b strings.Builder
	b.WriteString("## Worker resume notice\n\n")
	b.WriteString("This is a **resume** of an in-flight cycle, not a new task. ")
	b.WriteString("The server restarted while this cycle was running ")
	b.WriteString(fmt.Sprintf("(cycle_id=%s, interrupted during %s).\n\n", cycle.ID, interruptedPhase))
	b.WriteString("Before changing anything:\n")
	b.WriteString("1. Inspect the working tree you were given (`git status`, read relevant files).\n")
	b.WriteString("2. Continue from that state; do not revert work that satisfies locked criteria below.\n")
	if block := FormatKnownCommitsForResume(knownCommits); block != "" {
		b.WriteString("3. ")
		b.WriteString(strings.TrimSpace(block))
		b.WriteString("4. A clean tree does **not** mean the task succeeded ΓÇö complete remaining criteria and write the criteria report.\n")
	} else {
		b.WriteString("3. A clean tree does **not** mean the task succeeded ΓÇö complete remaining criteria and write the criteria report.\n")
	}
	b.WriteString("\n")
	return b.String() + prompt
}

// AppendGitCommitPolicy appends execute-phase git commit instructions.
//
//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func AppendGitCommitPolicy(prompt string, operatorResume bool) string {
	var b strings.Builder
	b.WriteString("## Git commits (required)\n\n")
	b.WriteString("Before you finish this execute phase, stage work with Shell `git add`, then create commits **only** via the MCP tool `hamix.commit`.\n")
	b.WriteString("Do **not** use Shell `git commit`. Do not put commit SHAs on the criteria report — the harness records SHAs from `hamix.commit`.\n\n")
	if operatorResume {
		b.WriteString("Create **new** commits only in this attempt; prior attempt SHAs are already indexed.\n\n")
	}
	b.WriteString("Use normal descriptive commit messages only — do **not** embed task IDs, cycle IDs, or ID markers.\n")
	b.WriteString("Create **new commits only** — fix mistakes with a follow-up `hamix.commit`; never amend, rebase, squash, or delete history.\n")
	b.WriteString("You may commit incrementally during the run (multiple `hamix.commit` calls). Uncommitted local changes are allowed if you already committed the work you are claiming.\n")
	b.WriteString("Do not push.\n\n")
	return b.String() + prompt
}

// FormatKnownCommitsForResume lists commits already indexed for the task.
//
//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func FormatKnownCommitsForResume(commits []cyclesdomain.TaskCycleCommit) string {
	if len(commits) == 0 {
		return ""
	}
	var b strings.Builder
	b.WriteString("Known commits already indexed for this task (all prior attempts):\n")
	for _, c := range commits {
		short := c.SHA
		if len(short) > 12 {
			short = short[:12]
		}
		b.WriteString(fmt.Sprintf("- %s ΓÇö %s\n", short, c.Message))
	}
	b.WriteByte('\n')
	return b.String()
}

// FormatVerifyDiffSection renders the git diff block for verify prompts.
//
//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func FormatVerifyDiffSection(diff string, fetchErr error) string {
	if fetchErr != nil {
		return "(diff unavailable: " + fetchErr.Error() + ")"
	}
	return diff
}
