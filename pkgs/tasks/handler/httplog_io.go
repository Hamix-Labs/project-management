package handler

import (
	"context"
	"net/http"
	"strings"
	"time"

	taskcorehandler "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/handler"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/handlerhttp"
)

const (
	maxHTTPLogTitleRunes  = 160
	maxHTTPLogPromptRunes = 400
)

// taskCreateInputFields builds the body_* slog attribute slice for the
// debugHTTPRequest http.io trace. Pure transformation; the trace itself
// logs and is gated by Enabled() upstream so this helper never runs when
// debug is off. Skip-listed in cmd/funclogmeasure/analyze.go.
//
//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func taskCreateInputFields(body *taskcorehandler.TaskCreateJSON, actor string) []any {
	if body == nil {
		return nil
	}
	out := []any{
		"body_task_id", strings.TrimSpace(body.ID),
		"body_draft_id", strings.TrimSpace(body.DraftID),
		"body_status", string(body.Status),
		"body_priority", string(body.Priority),
		"body_title_len", len(body.Title),
		"body_title_preview", handlerhttp.TruncateRunes(body.Title, maxHTTPLogTitleRunes),
		"body_initial_prompt_len", len(body.InitialPrompt),
		"body_initial_prompt_preview", handlerhttp.TruncateRunes(body.InitialPrompt, maxHTTPLogPromptRunes),
		"body_project_id_set", body.ProjectID != nil,
		"actor", actor,
	}
	if body.ProjectID != nil {
		out = append(out, "body_project_id", strings.TrimSpace(*body.ProjectID))
	}
	if body.PickupNotBefore != nil {
		out = append(out, "body_pickup_not_before", strings.TrimSpace(*body.PickupNotBefore), "body_pickup_not_before_set", true)
	} else {
		out = append(out, "body_pickup_not_before_set", false)
	}
	out = append(out, "body_checklist_items_count", len(body.ChecklistItems))
	return out
}

// taskPatchInputFields is the PATCH /tasks/{id} mirror of
// taskCreateInputFields above; same pure-helper rationale and skip-list
// entry.
//
//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func taskPatchInputFields(body *taskcorehandler.TaskPatchJSON) []any {
	if body == nil {
		return nil
	}
	out := []any{}
	if body.Title != nil {
		out = append(out, "patch_title", true, "patch_title_len", len(*body.Title), "patch_title_preview", handlerhttp.TruncateRunes(*body.Title, maxHTTPLogTitleRunes))
	}
	if body.InitialPrompt != nil {
		out = append(out, "patch_initial_prompt", true, "patch_initial_prompt_len", len(*body.InitialPrompt),
			"patch_initial_prompt_preview", handlerhttp.TruncateRunes(*body.InitialPrompt, maxHTTPLogPromptRunes))
	}
	if body.Status != nil {
		out = append(out, "patch_status", string(*body.Status))
	}
	if body.Priority != nil {
		out = append(out, "patch_priority", string(*body.Priority))
	}
	if body.ProjectID.Defined {
		if body.ProjectID.Clear {
			out = append(out, "patch_project_id", "clear")
		} else {
			out = append(out, "patch_project_id", body.ProjectID.SetID)
		}
	}
	if body.PickupNotBefore.Defined {
		if body.PickupNotBefore.Clear {
			out = append(out, "patch_pickup_not_before", "clear")
		} else {
			out = append(out, "patch_pickup_not_before", body.PickupNotBefore.Set.UTC().Format(time.RFC3339))
		}
	}
	if body.CursorModel != nil {
		out = append(out, "patch_cursor_model", true, "patch_cursor_model_len", len(strings.TrimSpace(*body.CursorModel)))
	}
	return out
}

//funclogmeasure:skip category=hot-path reason="Thin re-export of handlerhttp.DebugHTTPRequest; shared package emits the http.io trace."
func debugHTTPRequest(r *http.Request, op string, extra ...any) {
	handlerhttp.DebugHTTPRequest(r, op, extra...)
}

//funclogmeasure:skip category=hot-path reason="Thin re-export of handlerhttp.DebugHTTPOut; shared package emits the http.io trace."
func debugHTTPOut(ctx context.Context, op string, httpStatus int, extra ...any) {
	handlerhttp.DebugHTTPOut(ctx, op, httpStatus, extra...)
}
