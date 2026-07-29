package resume

import (
	"github.com/AlexsanderHamir/Hamix/pkgs/agents/harness/internal/verify"
	cyclesdomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcycles/domain"
)

// Entry selects which cycle-loop branch resumes after interruption or retry.
type Entry int

const (
	EntryExecute Entry = iota
	EntryVerifyOnly
	EntryAfterExecuteSuccess
)

// ContinuationFailureKind categorizes the parent attempt failure for continuation prompts.
type ContinuationFailureKind string

const (
	ContinuationFailureRunner         ContinuationFailureKind = "runner"
	ContinuationFailureExecuteGate    ContinuationFailureKind = "executeGate"
	ContinuationFailureVerify         ContinuationFailureKind = "verify"
	ContinuationFailureInfrastructure ContinuationFailureKind = "infrastructure"
	ContinuationFailureOperator       ContinuationFailureKind = "operator"
)

// CriterionVerdict is the shared locked-pass DTO (same shape as verify.Verdict).
type CriterionVerdict = verify.Verdict

// ContinuationBundle rehydrates cross-cycle resume context from a parent attempt.
type ContinuationBundle struct {
	Entry                  Entry
	LineageAttempt         int64
	ParentCycleID          string
	FailureClass           ContinuationFailureKind
	FailureReason          string
	FailurePhase           cyclesdomain.Phase
	ScopeFiles             []string
	Commits                []cyclesdomain.TaskCycleCommit
	CriteriaEvidence       []cyclesdomain.TaskCycleCriteriaReport
	LockedPasses           map[string]CriterionVerdict
	ExecuteFeedback        string
	CriteriaReportProbeErr string
	RunnerFeedback         string
	GitDiagnostics         string
	Warnings               []string
	Sufficient             bool
}

// Checkpoint is the in-cycle resume state reconstructed from the phase ledger.
type Checkpoint struct {
	Entry        Entry
	LockedPasses map[string]CriterionVerdict
	KnownCommits []cyclesdomain.TaskCycleCommit
	Continuation *ContinuationBundle
}
