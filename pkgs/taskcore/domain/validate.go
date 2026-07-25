package domain

import "fmt"

// ValidStatus reports whether s is a writable Status enum.
//
//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func ValidStatus(s Status) bool {
	switch s {
	case StatusReady, StatusRunning, StatusBlocked, StatusReview, StatusDone, StatusFailed, StatusOnHold:
		return true
	default:
		return false
	}
}

// ValidClientWritableStatus reports whether a client may set s on create or PATCH.
//
//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func ValidClientWritableStatus(s Status) bool {
	return ValidStatus(s)
}

// ValidPriority reports whether p is a writable Priority enum.
//
//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func ValidPriority(p Priority) bool {
	switch p {
	case PriorityLow, PriorityMedium, PriorityHigh, PriorityCritical:
		return true
	default:
		return false
	}
}

// ValidateActor returns ErrInvalidInput when a is not a known actor enum.
//
//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func ValidateActor(a Actor) error {
	switch a {
	case ActorUser, ActorAgent:
		return nil
	default:
		return fmt.Errorf("%w: actor", ErrInvalidInput)
	}
}
