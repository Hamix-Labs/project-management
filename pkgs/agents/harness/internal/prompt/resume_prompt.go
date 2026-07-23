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

// PolishNoticeInput drives AppendPolishNotice sections.
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
	b.WriteString("## Operator retry — resume from failure\n\n")
	b.WriteString("This is a **new execution attempt** continuing work from a failed prior attempt ")
	b.WriteString(fmt.Sprintf("(new cycle_id=%s).\n\n", cycle.ID))
	b.WriteString("Before changing anything:\n")
	b.WriteString("1. Inspect the working tree you were given (`git status`, read relevant files).\n")
	b.WriteString("2. Continue from that state; do not revert work that satisfies locked criteria below.\n")
	if block := FormatKnownCommitsForResume(parentCommits); block != "" {
		b.WriteString("3. ")
		b.WriteString(strings.TrimSpace(block))
		b.WriteString("Those commits are already indexed for this task — list only **new** commits you create in `commits[]` on your criteria report.\n")
		b.WriteString("4. A clean tree does **not** mean the task succeeded — complete remaining criteria and write the criteria report.\n")
	} else {
		b.WriteString("3. A clean tree does **not** mean the task succeeded — complete remaining criteria and write the criteria report.\n")
	}
	b.WriteString("\n")
	return b.String() + prompt
}

// AppendPolishNotice is for human polish from review: resume the Cursor conversation
// with operator instructions. This is not a failure-resume path.
//
//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func AppendPolishNotice(prompt string, cycle *cyclesdomain.TaskCycle, in PolishNoticeInput) string {
	if cycle == nil {
		return prompt
	}
	var b strings.Builder
	b.WriteString("## Human polish — refine completed work\n\n")
	b.WriteString("This is a **new execution attempt** after the prior attempt succeeded and entered human review ")
	b.WriteString(fmt.Sprintf("(new cycle_id=%s).\n\n", cycle.ID))
	b.WriteString("You are continuing the same Cursor conversation. Treat this as intentional refinement of accepted work, ")
	b.WriteString("not recovery from a failure and not a full rewrite unless the instructions say so.\n\n")

	instructions := strings.TrimSpace(in.Instructions)
	b.WriteString("### Operator polish instructions (authoritative)\n\n")
	if instructions != "" {
		b.WriteString(instructions)
		b.WriteString("\n\n")
	} else {
		b.WriteString("(No freeform instructions were provided — follow the criterion sections below.)\n\n")
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
		b.WriteString("This polish has **no flagged or newly added criteria**. After execute finishes, the harness will **skip the verify agent**.\n")
		b.WriteString("Your execute-phase claim that the polish instructions are satisfied is sufficient for this attempt to return to human review.\n")
		b.WriteString("Still inspect the tree, change only what the instructions require, commit new work normally, and do not undo prior good work.\n\n")
	} else {
		b.WriteString("### Verification mode for this polish\n\n")
		b.WriteString("Criteria listed as already verified must remain satisfied — do not undo them.\n")
		b.WriteString("Active / flagged / new criteria will be re-verified by the independent verify agent after execute.\n")
		b.WriteString("Write a complete criteria report for active criteria only.\n\n")
	}

	b.WriteString("### Before changing anything\n\n")
	b.WriteString("1. Inspect the working tree (`git status`, read relevant files).\n")
	b.WriteString("2. Keep prior good work; change only what the polish instructions and flagged/new criteria require.\n")
	if block := FormatKnownCommitsForResume(in.KnownCommits); block != "" {
		b.WriteString("3. ")
		b.WriteString(strings.TrimSpace(block))
		b.WriteString("Those commits are already indexed — list only **new** commits you create in `commits[]` on your criteria report.\n")
		if in.SkipVerify {
			b.WriteString("4. Apply the polish instructions; your execute claim ends this attempt (no verify agent).\n")
		} else {
			b.WriteString("4. Re-satisfy active criteria after your changes and write the criteria report.\n")
		}
	} else if in.SkipVerify {
		b.WriteString("3. Apply the polish instructions; your execute claim ends this attempt (no verify agent).\n")
	} else {
		b.WriteString("3. Re-satisfy active criteria after your changes and write the criteria report.\n")
	}
	b.WriteString("\n")
	return b.String() + prompt
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
		b.WriteString("4. A clean tree does **not** mean the task succeeded — complete remaining criteria and write the criteria report.\n")
	} else {
		b.WriteString("3. A clean tree does **not** mean the task succeeded — complete remaining criteria and write the criteria report.\n")
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
	b.WriteString("Before you finish this execute phase, commit work that satisfies criteria you are claiming, and list new commits in `commits[]` on your criteria report.\n\n")
	if operatorResume {
		b.WriteString("Create **new** commits only in this attempt; prior attempt SHAs are already indexed.\n\n")
	}
	b.WriteString("Use normal descriptive commit messages only — do **not** embed task IDs, cycle IDs, or ID markers.\n")
	b.WriteString("Create **new commits only** — fix mistakes with a follow-up commit; never amend, rebase, squash, or delete history.\n")
	b.WriteString("You may commit incrementally during the run. Uncommitted local changes are allowed if you already committed the work you are claiming.\n")
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
		b.WriteString(fmt.Sprintf("- %s — %s\n", short, c.Message))
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
