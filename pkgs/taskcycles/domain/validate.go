package domain

// ValidPhase reports whether p is a known Phase enum.
//
//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func ValidPhase(p Phase) bool {
	switch p {
	case PhaseExecute, PhaseVerify:
		return true
	default:
		return false
	}
}
