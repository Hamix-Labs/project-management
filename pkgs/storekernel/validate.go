package storekernel

import "github.com/AlexsanderHamir/Hamix/pkgs/obs/calltrace"
import (
	"fmt"
	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
	cyclesdomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcycles/domain"
	"log/slog"
)

// ValidStatus reports whether s is a writable taskcoredomain.Status enum.
func ValidStatus(s taskcoredomain.Status) bool {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.kernel.ValidStatus")
	switch s {
	case taskcoredomain.StatusReady, taskcoredomain.StatusRunning, taskcoredomain.StatusBlocked, taskcoredomain.StatusReview, taskcoredomain.StatusDone, taskcoredomain.StatusFailed, taskcoredomain.StatusOnHold:
		return true
	default:
		return false
	}
}

// ValidClientWritableStatus reports whether a client may set s on create or PATCH.
func ValidClientWritableStatus(s taskcoredomain.Status) bool {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.kernel.ValidClientWritableStatus")
	return ValidStatus(s)
}

// ValidPriority reports whether p is a writable taskcoredomain.Priority enum.
func ValidPriority(p taskcoredomain.Priority) bool {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.kernel.ValidPriority")
	switch p {
	case taskcoredomain.PriorityLow, taskcoredomain.PriorityMedium, taskcoredomain.PriorityHigh, taskcoredomain.PriorityCritical:
		return true
	default:
		return false
	}
}

// ValidateActor returns taskcoredomain.ErrInvalidInput when a is not a known actor enum.
func ValidateActor(a taskcoredomain.Actor) error {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.kernel.ValidateActor")
	switch a {
	case taskcoredomain.ActorUser, taskcoredomain.ActorAgent:
		return nil
	default:
		return fmt.Errorf("%w: actor", taskcoredomain.ErrInvalidInput)
	}
}

// ValidPhase reports whether p is a known cyclesdomain.Phase enum.
func ValidPhase(p cyclesdomain.Phase) bool {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.kernel.ValidPhase")
	switch p {
	case cyclesdomain.PhaseExecute, cyclesdomain.PhaseVerify:
		return true
	default:
		return false
	}
}

// ValidTerminalCycleStatus reports whether s is a terminal CycleStatus.
func ValidTerminalCycleStatus(s cyclesdomain.CycleStatus) bool {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.kernel.ValidTerminalCycleStatus")
	return cyclesdomain.TerminalCycleStatus(s)
}

// ValidTerminalPhaseStatus reports whether s is a terminal PhaseStatus.
func ValidTerminalPhaseStatus(s cyclesdomain.PhaseStatus) bool {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.kernel.ValidTerminalPhaseStatus")
	return cyclesdomain.TerminalPhaseStatus(s)
}
