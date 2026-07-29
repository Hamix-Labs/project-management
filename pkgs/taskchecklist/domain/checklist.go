package domain

import (
	"time"
)

// TaskChecklistItem is a definition row owned by a task.
// Completion is recorded by the agent worker as execute_claim (no verify
// commands), execute_agent (command-backed verify), or agent_self (execute
// did not claim done ΓÇö failure-only).
type TaskChecklistItem struct {
	ID        string `json:"id"`
	TaskID    string `json:"task_id"`
	SortOrder int    `json:"sort_order"`
	Text      string `json:"text"`
}

// TaskChecklistItemCommand is an optional shell check attached to a
// checklist definition. The worker runs these during verify and writes
// stdout/stderr to temp files under the cycle report dir; the LLM
// verifier interprets the artifacts against ExpectedOutcome.
type TaskChecklistItemCommand struct {
	ID              string `json:"id"`
	ItemID          string `json:"item_id"`
	SortOrder       int    `json:"sort_order"`
	Command         string `json:"command"`
	ExpectedOutcome string `json:"expected_outcome"`
	// TimeoutSeconds is an optional per-command wall-clock cap. Nil means no
	// timeout (cancel only via the parent cycle context).
	TimeoutSeconds *int `json:"timeout_seconds,omitempty"`
}

// TaskChecklistCompletion records that subject TaskID satisfied checklist item ItemID.
// By holds tasks/domain.Actor wire values ("user", "agent") without importing tasks/domain.
type TaskChecklistCompletion struct {
	TaskID            string       `json:"task_id"`
	ItemID            string       `json:"item_id"`
	At                time.Time    `json:"at"`
	By                string       `json:"by"`
	Evidence          string       `json:"evidence"`
	VerifiedBy        VerifierKind `json:"verified_by"`
	VerifierReasoning string       `json:"verifier_reasoning"`
	CycleID           string       `json:"cycle_id"`
}
