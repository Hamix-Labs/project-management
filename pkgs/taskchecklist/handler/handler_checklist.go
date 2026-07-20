package handler

import (
	"context"
	"fmt"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/handlerhttp"
	"log/slog"
	"net/http"
	"strings"

	"github.com/AlexsanderHamir/Hamix/pkgs/obs/calltrace"
	"github.com/AlexsanderHamir/Hamix/pkgs/taskchecklist/contract"
	checklistdomain "github.com/AlexsanderHamir/Hamix/pkgs/taskchecklist/domain"
	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
)

type checklistItemCreateJSON struct {
	Text           string                        `json:"text"`
	VerifyCommands []contract.VerifyCommandInput `json:"verify_commands,omitempty"`
}

type patchChecklistItemBody struct {
	Text           *string                        `json:"text,omitempty"`
	VerifyCommands *[]contract.VerifyCommandInput `json:"verify_commands,omitempty"`
	Done           *bool                          `json:"done,omitempty"`
	Evidence       *string                        `json:"evidence,omitempty"`
	VerifiedBy     *string                        `json:"verified_by,omitempty"`
}

type checklistListResponse struct {
	Items []contract.ChecklistItemView `json:"items"`
}

func (h *Handler) getChecklist(w http.ResponseWriter, r *http.Request) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "taskchecklist.handler.getChecklist")
	const op = "tasks.checklist.list"
	r = calltrace.WithRequestRoot(r, op)
	id, err := parseTaskPathID(r.PathValue("id"))
	if err != nil {
		handlerhttp.WriteStoreError(w, r, op, err)
		return
	}
	handlerhttp.DebugHTTPRequest(r, op, "task_id", id)
	items, err := h.checklist.ListChecklistForSubject(r.Context(), id)
	if err != nil {
		handlerhttp.WriteStoreError(w, r, op, err)
		return
	}
	handlerhttp.WriteJSONWithETag(w, r, op, http.StatusOK, checklistListResponse{Items: items})
}

func (h *Handler) postChecklistItem(w http.ResponseWriter, r *http.Request) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "taskchecklist.handler.postChecklistItem")
	const op = "tasks.checklist.create"
	r = calltrace.WithRequestRoot(r, op)
	id, err := parseTaskPathID(r.PathValue("id"))
	if err != nil {
		handlerhttp.WriteStoreError(w, r, op, err)
		return
	}
	var body checklistItemCreateJSON
	if err := handlerhttp.DecodeJSON(r.Context(), r.Body, &body); err != nil {
		handlerhttp.DebugHTTPRequest(r, op, "task_id", id, "json_decode_failed", true)
		handlerhttp.WriteError(w, r, op, err, http.StatusBadRequest)
		return
	}
	by := handlerhttp.ActorFromRequest(r)
	handlerhttp.DebugHTTPRequest(r, op, "task_id", id, "actor", string(by),
		"text_len", len(body.Text), "text_preview", handlerhttp.TruncateRunes(body.Text, handlerhttp.MaxHTTPLogTextRunes))
	if running, err := h.cycleRunning.IsTaskCycleRunning(r.Context(), id); err != nil {
		handlerhttp.WriteStoreError(w, r, op, err)
		return
	} else if running {
		handlerhttp.WriteStoreError(w, r, op, fmt.Errorf("%w: cycle running; cannot add criteria", taskcoredomain.ErrConflict))
		return
	}
	it, err := h.checklist.AddChecklistItem(r.Context(), id, body.Text, body.VerifyCommands, by)
	if err != nil {
		handlerhttp.WriteStoreError(w, r, op, err)
		return
	}
	if err := h.notifyChecklistChange(r.Context(), id); err != nil {
		handlerhttp.WriteStoreError(w, r, op, err)
		return
	}
	handlerhttp.WriteJSON(w, r, op, http.StatusCreated, it)
}

func (h *Handler) patchChecklistItem(w http.ResponseWriter, r *http.Request) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "taskchecklist.handler.patchChecklistItem")
	const op = "tasks.checklist.patch"
	r = calltrace.WithRequestRoot(r, op)
	taskID, err := parseTaskPathID(r.PathValue("id"))
	if err != nil {
		handlerhttp.WriteStoreError(w, r, op, err)
		return
	}
	itemID, err := parseTaskPathItemID(r.PathValue("itemId"))
	if err != nil {
		handlerhttp.WriteStoreError(w, r, op, err)
		return
	}
	var body patchChecklistItemBody
	if err := handlerhttp.DecodeJSON(r.Context(), r.Body, &body); err != nil {
		handlerhttp.DebugHTTPRequest(r, op, "task_id", taskID, "item_id", itemID, "json_decode_failed", true)
		handlerhttp.WriteError(w, r, op, err, http.StatusBadRequest)
		return
	}
	setCount := 0
	if body.Text != nil {
		setCount++
	}
	if body.VerifyCommands != nil {
		setCount++
	}
	if body.Done != nil {
		setCount++
	}
	if setCount != 1 {
		handlerhttp.WriteStoreError(w, r, op, fmt.Errorf("%w: send exactly one of text, verify_commands, or done", taskcoredomain.ErrInvalidInput))
		return
	}
	by := handlerhttp.ActorFromRequest(r)
	if body.Text != nil {
		if running, err := h.cycleRunning.IsTaskCycleRunning(r.Context(), taskID); err != nil {
			handlerhttp.WriteStoreError(w, r, op, err)
			return
		} else if running {
			handlerhttp.WriteStoreError(w, r, op, fmt.Errorf("%w: cycle running; cannot edit criteria", taskcoredomain.ErrConflict))
			return
		}
	} else if body.VerifyCommands != nil {
		if running, err := h.cycleRunning.IsTaskCycleRunning(r.Context(), taskID); err != nil {
			handlerhttp.WriteStoreError(w, r, op, err)
			return
		} else if running {
			handlerhttp.WriteStoreError(w, r, op, fmt.Errorf("%w: cycle running; cannot edit criteria", taskcoredomain.ErrConflict))
			return
		}
	}
	if body.Text != nil {
		t := strings.TrimSpace(*body.Text)
		handlerhttp.DebugHTTPRequest(r, op, "task_id", taskID, "item_id", itemID, "text_len", len(t), "text_preview", handlerhttp.TruncateRunes(t, handlerhttp.MaxHTTPLogTextRunes), "actor", string(by))
		if t == "" {
			handlerhttp.WriteStoreError(w, r, op, fmt.Errorf("%w: text required", taskcoredomain.ErrInvalidInput))
			return
		}
		if err := h.checklist.UpdateChecklistItemText(r.Context(), taskID, itemID, t, by); err != nil {
			handlerhttp.WriteStoreError(w, r, op, err)
			return
		}
	} else if body.VerifyCommands != nil {
		if err := h.checklist.ReplaceChecklistVerifyCommands(r.Context(), taskID, itemID, *body.VerifyCommands, by); err != nil {
			handlerhttp.WriteStoreError(w, r, op, err)
			return
		}
	} else {
		handlerhttp.DebugHTTPRequest(r, op, "task_id", taskID, "item_id", itemID, "done", *body.Done, "actor", string(by))
		if *body.Done {
			if by != taskcoredomain.ActorAgent {
				handlerhttp.WriteStoreError(w, r, op, fmt.Errorf("%w: only the agent may mark checklist items done", taskcoredomain.ErrInvalidInput))
				return
			}
			evidence := ""
			if body.Evidence != nil {
				evidence = strings.TrimSpace(*body.Evidence)
			}
			if evidence == "" {
				handlerhttp.WriteStoreError(w, r, op, fmt.Errorf("%w: evidence required when marking done", taskcoredomain.ErrInvalidInput))
				return
			}
			verifier := checklistdomain.VerifierAgentSelf
			if body.VerifiedBy != nil {
				verifier = checklistdomain.VerifierKind(strings.TrimSpace(*body.VerifiedBy))
			}
			if !checklistdomain.ValidVerifierKind(verifier) || verifier == checklistdomain.VerifierLegacy {
				handlerhttp.WriteStoreError(w, r, op, fmt.Errorf("%w: invalid verified_by", taskcoredomain.ErrInvalidInput))
				return
			}
			if err := h.checklist.SetChecklistItemDoneWithEvidence(r.Context(), contract.SetDoneWithEvidenceInput{
				TaskID: taskID, ItemID: itemID, Evidence: evidence, Verifier: verifier, By: by,
			}); err != nil {
				handlerhttp.WriteStoreError(w, r, op, err)
				return
			}
		} else if err := h.checklist.SetChecklistItemDone(r.Context(), taskID, itemID, false, by); err != nil {
			handlerhttp.WriteStoreError(w, r, op, err)
			return
		}
	}
	if err := h.notifyChecklistChange(r.Context(), taskID); err != nil {
		handlerhttp.WriteStoreError(w, r, op, err)
		return
	}
	items, err := h.checklist.ListChecklistForSubject(r.Context(), taskID)
	if err != nil {
		handlerhttp.WriteStoreError(w, r, op, err)
		return
	}
	handlerhttp.WriteJSON(w, r, op, http.StatusOK, checklistListResponse{Items: items})
}

func (h *Handler) deleteChecklistItem(w http.ResponseWriter, r *http.Request) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "taskchecklist.handler.deleteChecklistItem")
	const op = "tasks.checklist.delete"
	r = calltrace.WithRequestRoot(r, op)
	id, err := parseTaskPathID(r.PathValue("id"))
	if err != nil {
		handlerhttp.WriteStoreError(w, r, op, err)
		return
	}
	itemID, err := parseTaskPathItemID(r.PathValue("itemId"))
	if err != nil {
		handlerhttp.WriteStoreError(w, r, op, err)
		return
	}
	handlerhttp.DebugHTTPRequest(r, op, "task_id", id, "item_id", itemID)
	if running, err := h.cycleRunning.IsTaskCycleRunning(r.Context(), id); err != nil {
		handlerhttp.WriteStoreError(w, r, op, err)
		return
	} else if running {
		handlerhttp.WriteStoreError(w, r, op, fmt.Errorf("%w: cycle running; cannot delete criteria", taskcoredomain.ErrConflict))
		return
	}
	by := handlerhttp.ActorFromRequest(r)
	if err := h.checklist.DeleteChecklistItem(r.Context(), id, itemID, by); err != nil {
		handlerhttp.WriteStoreError(w, r, op, err)
		return
	}
	if err := h.notifyChecklistChange(r.Context(), id); err != nil {
		handlerhttp.WriteStoreError(w, r, op, err)
		return
	}
	handlerhttp.DebugHTTPOut(r.Context(), op, http.StatusNoContent, "task_id", id, "item_id", itemID, "response_empty", true)
	w.WriteHeader(http.StatusNoContent)
}

//funclogmeasure:skip category=delegate-already-logs reason="SSE notify callback; HTTP handler chokepoint emits trace."
func (h *Handler) notifyChecklistChange(ctx context.Context, taskID string) error {
	if h.notifyTaskUpdated == nil {
		return nil
	}
	return h.notifyTaskUpdated(ctx, taskID)
}
