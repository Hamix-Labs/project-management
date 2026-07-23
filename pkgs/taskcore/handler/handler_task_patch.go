package handler

import (
	"log/slog"
	"net/http"
	"strings"

	"github.com/AlexsanderHamir/Hamix/pkgs/obs/calltrace"
	"github.com/AlexsanderHamir/Hamix/pkgs/taskcore/contract"
	"github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
	"github.com/AlexsanderHamir/Hamix/pkgs/taskcore/store"
)

func (h *Handler) patch(w http.ResponseWriter, r *http.Request) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "handler.Handler.patch")
	const op = "tasks.patch"
	r = calltrace.WithRequestRoot(r, op)
	id, err := h.parseTaskPathID(r.PathValue("id"))
	if err != nil {
		h.httpPort.WriteStoreError(w, r, op, err)
		return
	}
	var body taskPatchJSON
	if err := h.httpPort.DecodeJSON(r.Context(), r.Body, &body); err != nil {
		h.debugHTTPRequest(r, op, "task_id", id, "json_decode_failed", true)
		h.httpPort.WriteError(w, r, op, err, http.StatusBadRequest)
		return
	}
	h.debugHTTPRequest(r, op, append(append([]any{}, "task_id", id), taskPatchInputFields(&body)...)...)
	var dependsOnPatch *[]domain.DependencyEdge
	if body.DependsOn != nil && body.DependsOn.set {
		dependsOnPatch = &body.DependsOn.value
	}
	in := store.UpdateTaskInput{
		Title:                 body.Title,
		InitialPrompt:         body.InitialPrompt,
		Status:                body.Status,
		Priority:              body.Priority,
		Project:               projectFieldPatchToStore(body.ProjectID),
		ProjectContextItemIDs: body.ProjectContextItemIDs,
		PickupNotBefore:       pickupNotBeforePatchToStore(body.PickupNotBefore),
		CursorModel:           body.CursorModel,
		Tags:                  body.Tags,
		Milestone:             body.Milestone,
		Gate:                  gateFieldPatchToStore(body.Gate),
		DependsOn:             dependsOnPatch,
		WorktreeID:            body.WorktreeID,
	}
	if body.InitialPrompt != nil {
		cur, getErr := h.tasks.Get(r.Context(), id)
		if getErr != nil {
			h.httpPort.WriteStoreError(w, r, op, getErr)
			return
		}
		wt := body.WorktreeID
		if wt == nil {
			wt = cur.WorktreeID
		}
		projectID := cur.ProjectID
		if body.ProjectID.Defined && !body.ProjectID.Clear {
			projectID = &body.ProjectID.SetID
		}
		var err error
		if wt != nil && strings.TrimSpace(*wt) != "" {
			err = h.gitCompose.ValidatePromptMentionsForWorktree(r.Context(), wt, *body.InitialPrompt)
		} else {
			err = h.gitCompose.ValidatePromptMentionsForProject(r.Context(), projectID, *body.InitialPrompt)
		}
		if err != nil {
			h.httpPort.WriteStoreError(w, r, op, err)
			return
		}
	}
	if body.WorktreeID != nil {
		cur, getErr := h.tasks.Get(r.Context(), id)
		if getErr != nil {
			h.httpPort.WriteStoreError(w, r, op, getErr)
			return
		}
		projectID := cur.ProjectID
		if body.ProjectID.Defined && !body.ProjectID.Clear {
			projectID = &body.ProjectID.SetID
		}
		wt := cur.WorktreeID
		if body.WorktreeID != nil {
			wt = body.WorktreeID
		}
		if err := h.gitCompose.ValidateTaskGitBindingV2(r.Context(), projectID, wt); err != nil {
			h.httpPort.WriteStoreError(w, r, op, err)
			return
		}
	}
	by := h.httpPort.ActorFromRequest(r)
	_, err = h.tasks.Update(r.Context(), id, in, by)
	if err != nil {
		h.httpPort.WriteStoreError(w, r, op, err)
		return
	}
	task, err := h.tasks.Get(r.Context(), id)
	if err != nil {
		h.httpPort.WriteStoreError(w, r, op, err)
		return
	}
	h.notifyTaskChangedSafe(contract.ChangeTaskUpdated, id, task)
	if body.Gate.Defined {
		h.notifyChangeSafe(contract.ChangeTaskGateChanged, id)
	}
	if body.DependsOn != nil && body.DependsOn.set {
		h.notifyChangeSafe(contract.ChangeTaskDependencyChanged, id)
	}
	taskapiDomainTasksUpdatedTotal.Inc()
	h.httpPort.WriteJSON(w, r, op, http.StatusOK, task)
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func projectFieldPatchToStore(p patchProjectField) *store.ProjectFieldPatch {
	if !p.Defined {
		return nil
	}
	if p.Clear {
		return &store.ProjectFieldPatch{Clear: true}
	}
	return &store.ProjectFieldPatch{ID: p.SetID}
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func gateFieldPatchToStore(p patchGateField) **domain.TaskGate {
	if !p.Defined {
		return nil
	}
	if p.Clear {
		var cleared *domain.TaskGate
		return &cleared
	}
	return &p.Set
}
