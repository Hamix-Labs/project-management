package orchestration

import (
	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
	cyclesdomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcycles/domain"
)

// TerminationReason is a stable cycle terminate_reason string persisted to the store.
type TerminationReason string

const (
	ReasonVerifyTampered             TerminationReason = "verify_tampered"
	ReasonRunnerTimeout              TerminationReason = "runner_timeout"
	ReasonStreamIdle                 TerminationReason = "stream_idle"
	ReasonRunnerNonZeroExit          TerminationReason = "runner_non_zero_exit"
	ReasonRunnerInvalidOutput        TerminationReason = "runner_invalid_output"
	ReasonRunnerError                TerminationReason = "runner_error"
	ReasonExecuteInvalidCommit       TerminationReason = "execute_invalid_commit"
	ReasonExecuteMissingCommits      TerminationReason = "execute_missing_commits"
	ReasonExecuteUnregisteredCommits TerminationReason = "execute_unregistered_commits"
	ReasonCancelledByOperator        TerminationReason = "cancelled_by_operator"
	ReasonChecklistCompletionFailed  TerminationReason = "checklist_completion_failed"
	ReasonCursorMissingSessionID     TerminationReason = "cursor_missing_session_id"
	ReasonCursorResumeSession        TerminationReason = "cursor_resume_session"
)

// VerifyEffects lists side effects the harness root applies after verify (ADR-0090 one-shot).
type VerifyEffects struct {
	TerminalFailure bool
	Tampered        bool
}

// ExecuteEffects lists side effects the harness root applies after DecideExecutePostRun.
type ExecuteEffects struct {
	ContinueToClaimAcceptance bool
	StopLoop         bool
	TerminateFailed  bool
	TransitionTask   taskcoredomain.Status
	Reason           TerminationReason
	ResultSummary    string
}

// FinalizeEffects lists side effects after DecideFinalizeSuccess.
type FinalizeEffects struct {
	CycleStatus cyclesdomain.CycleStatus
	TaskStatus  taskcoredomain.Status
	Reason      TerminationReason
}
