package handler

import (
	"encoding/json"
	"fmt"

	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/domain"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/store"
)

type taskComposePayloadJSON struct {
	Title                 string                           `json:"title"`
	InitialPrompt         string                           `json:"initial_prompt"`
	Status                domain.Status                    `json:"status"`
	Priority              domain.Priority                  `json:"priority"`
	ProjectID             *string                          `json:"project_id"`
	RepositoryID          *string                          `json:"repository_id,omitempty"`
	ProjectContextItemIDs []string                         `json:"project_context_item_ids"`
	Runner                *string                          `json:"runner"`
	CursorModel           *string                          `json:"cursor_model"`
	PickupNotBefore       *string                          `json:"pickup_not_before,omitempty"`
	Tags                  []string                         `json:"tags,omitempty"`
	Milestone             *string                          `json:"milestone,omitempty"`
	DependsOn             dependsOnWire                    `json:"depends_on,omitempty"`
	ChecklistItems        []store.CreateChecklistItemInput `json:"checklist_items"`
	WorktreeID            *string                          `json:"worktree_id,omitempty"`
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func decodeComposePayload(raw json.RawMessage) (taskComposePayloadJSON, error) {
	var payload taskComposePayloadJSON
	if len(raw) == 0 {
		return payload, fmt.Errorf("%w: payload required", domain.ErrInvalidInput)
	}
	if err := json.Unmarshal(raw, &payload); err != nil {
		return payload, fmt.Errorf("%w: invalid payload: %v", domain.ErrInvalidInput, err)
	}
	return payload, nil
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func composePayloadToRaw(payload taskComposePayloadJSON) (json.RawMessage, error) {
	return json.Marshal(payload)
}
