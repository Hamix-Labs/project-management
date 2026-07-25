package domain

import (
	"time"
)

// TaskChecklistItem is a definition row owned by a task.
// Completion is recorded only by the agent worker after verify (verified_by=execute_agent)
// or when execute did not claim done (verified_by=agent_self, failure-only).
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
