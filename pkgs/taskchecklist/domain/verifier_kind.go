package domain

// VerifierKind records how a checklist completion was proven.
type VerifierKind string

const (
	VerifierAgentSelf    VerifierKind = "agent_self"
	VerifierExecuteAgent VerifierKind = "execute_agent"
	// VerifierExecuteClaim is harness acceptance of an execute criteria
	// report for criteria with no verify_commands (ADR-0090).
	VerifierExecuteClaim       VerifierKind = "execute_claim"
	VerifierDeterministicCheck VerifierKind = "deterministic_check"
	VerifierHumanOverride      VerifierKind = "human_override"
	// VerifierLegacy is retained for historical completion rows and
	// verify-disabled checklist writes. New HTTP creates reject it;
	// retire after migrate clears verified_by='legacy' (see harness README).
	VerifierLegacy VerifierKind = "legacy"
)

// ValidVerifierKind reports whether k is allowed on completion rows.
//
//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func ValidVerifierKind(k VerifierKind) bool {
	switch k {
	case VerifierAgentSelf, VerifierExecuteAgent, VerifierExecuteClaim, VerifierDeterministicCheck, VerifierHumanOverride, VerifierLegacy:
		return true
	default:
		return false
	}
}
