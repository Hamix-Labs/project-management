package contract

// ChecklistItemView is one definition row plus completion for a subject task.
type ChecklistItemView struct {
	ID                string              `json:"id"`
	SortOrder         int                 `json:"sort_order"`
	Text              string              `json:"text"`
	VerifyCommands    []VerifyCommandView `json:"verify_commands,omitempty"`
	Done              bool                `json:"done"`
	Evidence          string              `json:"evidence,omitempty"`
	VerifiedBy        string              `json:"verified_by,omitempty"`
	VerifierReasoning string              `json:"verifier_reasoning,omitempty"`
	CycleID           string              `json:"cycle_id,omitempty"`
}

// CreateChecklistItemInput is one criterion seeded at task create.
type CreateChecklistItemInput struct {
	Text           string               `json:"text"`
	VerifyCommands []VerifyCommandInput `json:"verify_commands,omitempty"`
}

// VerifyCommandInput is a verify command on checklist create/update.
type VerifyCommandInput struct {
	Command         string `json:"command"`
	ExpectedOutcome string `json:"expected_outcome,omitempty"`
	// TimeoutSeconds caps this command's wall clock. Nil/omit = no timeout.
	// When set, must be > 0.
	TimeoutSeconds *int `json:"timeout_seconds,omitempty"`
}

// VerifyCommandView is a persisted command row on checklist API responses.
type VerifyCommandView struct {
	SortOrder       int    `json:"sort_order"`
	Command         string `json:"command"`
	ExpectedOutcome string `json:"expected_outcome,omitempty"`
	// TimeoutSeconds is omitted when null (unlimited).
	TimeoutSeconds *int `json:"timeout_seconds,omitempty"`
}
