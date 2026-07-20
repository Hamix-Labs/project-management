package domain

// ValidPhase reports whether p is a known Phase enum.
func ValidPhase(p Phase) bool {
	switch p {
	case PhaseExecute, PhaseVerify:
		return true
	default:
		return false
	}
}
