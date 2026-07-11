package domain

import cyclesdomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcycles/domain"

type (
	// TaskCycle aliases taskcycles/domain until taskcore (#186).
	TaskCycle = cyclesdomain.TaskCycle
	// TaskCyclePhase aliases taskcycles/domain until taskcore (#186).
	TaskCyclePhase = cyclesdomain.TaskCyclePhase
	// TaskCycleStreamEvent aliases taskcycles/domain until taskcore (#186).
	TaskCycleStreamEvent = cyclesdomain.TaskCycleStreamEvent
	// TaskCycleCriteriaReport aliases taskcycles/domain until taskcore (#186).
	TaskCycleCriteriaReport = cyclesdomain.TaskCycleCriteriaReport
	// TaskCycleVerifyReport aliases taskcycles/domain until taskcore (#186).
	TaskCycleVerifyReport = cyclesdomain.TaskCycleVerifyReport
	// TaskCycleCommandRun aliases taskcycles/domain until taskcore (#186).
	TaskCycleCommandRun = cyclesdomain.TaskCycleCommandRun
	// TaskCycleCommit aliases taskcycles/domain until taskcore (#186).
	TaskCycleCommit = cyclesdomain.TaskCycleCommit
	// Phase aliases taskcycles/domain until taskcore (#186).
	Phase = cyclesdomain.Phase
	// CycleStatus aliases taskcycles/domain until taskcore (#186).
	CycleStatus = cyclesdomain.CycleStatus
	// PhaseStatus aliases taskcycles/domain until taskcore (#186).
	PhaseStatus = cyclesdomain.PhaseStatus
)

const (
	PhaseExecute = cyclesdomain.PhaseExecute
	PhaseVerify  = cyclesdomain.PhaseVerify

	CycleStatusRunning   = cyclesdomain.CycleStatusRunning
	CycleStatusSucceeded = cyclesdomain.CycleStatusSucceeded
	CycleStatusFailed    = cyclesdomain.CycleStatusFailed
	CycleStatusAborted   = cyclesdomain.CycleStatusAborted

	PhaseStatusRunning   = cyclesdomain.PhaseStatusRunning
	PhaseStatusSucceeded = cyclesdomain.PhaseStatusSucceeded
	PhaseStatusFailed    = cyclesdomain.PhaseStatusFailed
	PhaseStatusSkipped   = cyclesdomain.PhaseStatusSkipped

	ExecuteCriteriaReportAttemptSeq = cyclesdomain.ExecuteCriteriaReportAttemptSeq
	PhaseInterruptReason            = cyclesdomain.PhaseInterruptReason
	PhaseDetailsRunCorrelationID    = cyclesdomain.PhaseDetailsRunCorrelationID
	PhaseDetailsSessionID           = cyclesdomain.PhaseDetailsSessionID
)

// ValidPhaseTransition forwards to taskcycles/domain.
//
//funclogmeasure:skip category=hot-path reason="Pure forwarder to taskcycles/domain; operation trace is emitted by the BC chokepoint."
func ValidPhaseTransition(prev, next Phase) bool {
	return cyclesdomain.ValidPhaseTransition(prev, next)
}

// ValidInterruptResumeTransition forwards to taskcycles/domain.
//
//funclogmeasure:skip category=hot-path reason="Pure forwarder to taskcycles/domain; operation trace is emitted by the BC chokepoint."
func ValidInterruptResumeTransition(last *TaskCyclePhase, next Phase) bool {
	return cyclesdomain.ValidInterruptResumeTransition(last, next)
}

// ValidVerifyOnlyRetryTransition forwards to taskcycles/domain.
//
//funclogmeasure:skip category=hot-path reason="Pure forwarder to taskcycles/domain; operation trace is emitted by the BC chokepoint."
func ValidVerifyOnlyRetryTransition(last *TaskCyclePhase, next Phase) bool {
	return cyclesdomain.ValidVerifyOnlyRetryTransition(last, next)
}

// TerminalCycleStatus forwards to taskcycles/domain.
//
//funclogmeasure:skip category=hot-path reason="Pure forwarder to taskcycles/domain; operation trace is emitted by the BC chokepoint."
func TerminalCycleStatus(s CycleStatus) bool {
	return cyclesdomain.TerminalCycleStatus(s)
}

// TerminalPhaseStatus forwards to taskcycles/domain.
//
//funclogmeasure:skip category=hot-path reason="Pure forwarder to taskcycles/domain; operation trace is emitted by the BC chokepoint."
func TerminalPhaseStatus(s PhaseStatus) bool {
	return cyclesdomain.TerminalPhaseStatus(s)
}

// RunCorrelationIDFromDetailsJSON forwards to taskcycles/domain.
//
//funclogmeasure:skip category=hot-path reason="Pure forwarder to taskcycles/domain; operation trace is emitted by the BC chokepoint."
func RunCorrelationIDFromDetailsJSON(detailsJSON []byte) string {
	return cyclesdomain.RunCorrelationIDFromDetailsJSON(detailsJSON)
}

// SessionIDFromDetailsJSON forwards to taskcycles/domain.
//
//funclogmeasure:skip category=hot-path reason="Pure forwarder to taskcycles/domain; operation trace is emitted by the BC chokepoint."
func SessionIDFromDetailsJSON(detailsJSON []byte) string {
	return cyclesdomain.SessionIDFromDetailsJSON(detailsJSON)
}
