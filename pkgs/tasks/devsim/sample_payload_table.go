package devsim

import (
	"encoding/json"
	taskeventsdomain "github.com/AlexsanderHamir/Hamix/pkgs/taskevents/domain"
)

// samplePayloadByType returns deterministic JSON payloads for the dev ticker.
// default branch in samplePayload handles unknown types.
var samplePayloadByType = map[taskeventsdomain.EventType]func() ([]byte, error){
	taskeventsdomain.EventStatusChanged: func() ([]byte, error) {
		return json.Marshal(map[string]string{"from": "ready", "to": "running"})
	},
	taskeventsdomain.EventPriorityChanged: func() ([]byte, error) {
		return json.Marshal(map[string]string{"from": "medium", "to": "high"})
	},
	taskeventsdomain.EventPromptAppended: func() ([]byte, error) {
		return json.Marshal(map[string]string{"from": "<p>a</p>", "to": "<p>a</p><p>b</p>"})
	},
	taskeventsdomain.EventMessageAdded: func() ([]byte, error) {
		return json.Marshal(map[string]string{"from": "Title A", "to": "Title B"})
	},
	taskeventsdomain.EventContextAdded: func() ([]byte, error) {
		return json.Marshal(map[string]string{"summary": "Repo layout", "detail": "Tasks live under pkgs/tasks."})
	},
	taskeventsdomain.EventConstraintAdded: func() ([]byte, error) {
		return json.Marshal(map[string]string{"text": "Must keep default go test ./... green."})
	},
	taskeventsdomain.EventSuccessCriterionAdded: func() ([]byte, error) {
		return json.Marshal(map[string]string{"text": "UI timeline renders without console errors."})
	},
	taskeventsdomain.EventNonGoalAdded: func() ([]byte, error) {
		return json.Marshal(map[string]string{"text": "No production deploy in this iteration."})
	},
	taskeventsdomain.EventPlanAdded: func() ([]byte, error) {
		return json.Marshal(map[string]any{
			"title": "Dev sim plan",
			"steps": []string{"Sketch", "Implement", "Verify"},
		})
	},
	taskeventsdomain.EventChecklistItemAdded: func() ([]byte, error) {
		return json.Marshal(map[string]string{"item_id": "cli-dev-1", "text": "Run go test ./..."})
	},
	taskeventsdomain.EventChecklistItemToggled: func() ([]byte, error) {
		return json.Marshal(map[string]string{"item_id": "cli-dev-1", "done": "true"})
	},
	taskeventsdomain.EventChecklistItemUpdated: func() ([]byte, error) {
		return json.Marshal(map[string]string{"item_id": "cli-dev-1", "text": "Run go test ./... (updated)"})
	},
	taskeventsdomain.EventChecklistItemRemoved: func() ([]byte, error) {
		return json.Marshal(map[string]string{"item_id": "cli-dev-1", "text": "Removed criterion (synthetic)"})
	},
	taskeventsdomain.EventArtifactAdded: func() ([]byte, error) {
		return json.Marshal(map[string]string{"name": "notes.md", "uri": "file:///tmp/hamix-devsim"})
	},
	taskeventsdomain.EventApprovalRequested: func() ([]byte, error) {
		return json.Marshal(map[string]string{"reason": "Checkpoint ready", "checkpoint": "plan_review"})
	},
	taskeventsdomain.EventApprovalGranted: func() ([]byte, error) {
		return json.Marshal(map[string]string{"grantor": "lead", "note": "LGTM (synthetic)"})
	},
	taskeventsdomain.EventOnTaskDone: func() ([]byte, error) {
		return json.Marshal(map[string]any{
			"worktree_id": "wt-devsim",
			"branch_id":   "br-devsim",
			"commits":     []map[string]string{{"sha": "abc1234", "message": "Synthetic commit."}},
		})
	},
	taskeventsdomain.EventTaskFailed: func() ([]byte, error) {
		return json.Marshal(map[string]string{"error": "Simulated failure", "retryable": "true"})
	},
	taskeventsdomain.EventSyncPing: func() ([]byte, error) {
		return json.Marshal(map[string]string{"source": "devsim"})
	},
}
