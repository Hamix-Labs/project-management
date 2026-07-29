package verify

import "strings"

// ComposeCommandVerifyEvidence merges execute claim evidence with the verify
// agent's expected_outcomeΓåöoutput interpretation (ADR-0090).
func ComposeCommandVerifyEvidence(executeEvidence, verifyReasoning string) string {
	e := strings.TrimSpace(executeEvidence)
	r := strings.TrimSpace(verifyReasoning)
	switch {
	case e == "":
		return r
	case r == "":
		return e
	default:
		return e + "\n\n## Command verification\n\n" + r
	}
}
