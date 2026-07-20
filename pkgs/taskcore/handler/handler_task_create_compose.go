package handler

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	settingsdomain "github.com/AlexsanderHamir/Hamix/pkgs/settings/domain"
	"github.com/AlexsanderHamir/Hamix/pkgs/storekernel"
	taskcorecontract "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/contract"
	"github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/calltrace"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/realtime"
)

type CreateTaskComposeOpts struct {
	ID                      string
	DraftID                 string
	Gate                    *domain.TaskGate
	StripDependsOn          bool
	OmitPastPickupNotBefore bool
	InstantiateFromTemplate bool
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func taskCreateJSONToCompose(body taskCreateJSON) taskComposePayloadJSON {
	return taskComposePayloadJSON{
		Title:                 body.Title,
		InitialPrompt:         body.InitialPrompt,
		Status:                body.Status,
		Priority:              body.Priority,
		ProjectID:             body.ProjectID,
		RepositoryID:          body.RepositoryID,
		ProjectContextItemIDs: body.ProjectContextItemIDs,
		Runner:                body.Runner,
		CursorModel:           body.CursorModel,
		PickupNotBefore:       body.PickupNotBefore,
		Tags:                  body.Tags,
		Milestone:             body.Milestone,
		DependsOn:             body.DependsOn,
		ChecklistItems:        body.ChecklistItems,
		WorktreeID:            body.WorktreeID,
	}
}

func (h *Handler) CreateTaskFromComposeJSON(
	ctx context.Context,
	r *http.Request,
	op string,
	payload TaskComposePayloadJSON,
	opts CreateTaskComposeOpts,
	by domain.Actor,
) (*domain.Task, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "handler.Handler.createTaskFromComposeJSON")
	settings, err := h.settings.GetSettings(ctx)
	if err != nil {
		return nil, err
	}
	if err := h.ValidateCompose(r.Context(), payload, settings, ValidateComposeOpts{
		GitMode: ComposeTaskRepoBinding,
	}); err != nil {
		return nil, err
	}
	runner, cursorModel, err := resolveRunnerModelFields(payload.Runner, payload.CursorModel, settings)
	if err != nil {
		return nil, err
	}
	pickupNotBefore, err := resolvePickupNotBeforeForCreate(payload.PickupNotBefore, payload.Status, settings)
	if err != nil {
		return nil, err
	}
	if opts.OmitPastPickupNotBefore {
		pickupNotBefore = omitPastPickupNotBefore(pickupNotBefore)
	}
	var dependsOn []domain.DependencyEdge
	if !opts.StripDependsOn {
		dependsOn, err = parseDependsOnWire(payload.DependsOn)
		if err != nil {
			return nil, err
		}
	}
	checklistItems, err := parseCreateChecklistItems(payload.ChecklistItems)
	if err != nil {
		return nil, err
	}
	draftID := opts.DraftID
	if opts.InstantiateFromTemplate {
		draftID = ""
	}
	taskID := storekernel.ResolveID(opts.ID)
	if strings.TrimSpace(opts.ID) != "" {
		existing, getErr := h.tasks.Get(ctx, taskID)
		if getErr == nil && existing != nil {
			return nil, fmt.Errorf("%w: task id already exists", domain.ErrConflict)
		}
		if getErr != nil && !errors.Is(getErr, domain.ErrNotFound) {
			return nil, getErr
		}
	}
	repoID := strings.TrimSpace(*payload.RepositoryID)
	wtID, err := h.gitCompose.AllocateTaskWorktree(r.Context(), repoID, taskID)
	if err != nil {
		return nil, err
	}
	payload.WorktreeID = &wtID
	if err := h.gitCompose.ValidateTaskGitBindingV2(r.Context(), payload.ProjectID, payload.WorktreeID); err != nil {
		return nil, err
	}
	if err := h.gitCompose.ValidatePromptMentionsForWorktree(r.Context(), payload.WorktreeID, payload.InitialPrompt); err != nil {
		return nil, err
	}
	t, err := h.tasks.Create(ctx, taskcorecontract.CreateTaskInput{
		ID:                    taskID,
		DraftID:               draftID,
		Title:                 payload.Title,
		InitialPrompt:         payload.InitialPrompt,
		Status:                payload.Status,
		Priority:              payload.Priority,
		ProjectID:             payload.ProjectID,
		ProjectContextItemIDs: payload.ProjectContextItemIDs,
		Runner:                runner,
		CursorModel:           cursorModel,
		PickupNotBefore:       pickupNotBefore,
		Tags:                  payload.Tags,
		Milestone:             payload.Milestone,
		Gate:                  opts.Gate,
		DependsOn:             dependsOn,
		ChecklistItems:        checklistItems,
		WorktreeID:            payload.WorktreeID,
	}, by)
	if err != nil {
		return nil, err
	}
	return h.finalizeCreatedTask(ctx, t)
}

func (h *Handler) finalizeCreatedTask(ctx context.Context, t *domain.Task) (*domain.Task, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "handler.Handler.finalizeCreatedTask")
	h.notifyTaskChangedSafe(realtime.TaskCreated, t.ID, t)
	if t.Gate != nil {
		h.notifyChangeSafe(realtime.TaskGateChanged, t.ID)
	}
	if len(t.DependsOn) > 0 {
		h.notifyChangeSafe(realtime.TaskDependencyChanged, t.ID)
	}
	taskapiDomainTasksCreatedTotal.Inc()
	return t, nil
}

// ValidateComposeOpts selects optional git-binding and mention checks around the
// shared compose field validation.
type ValidateComposeOpts struct {
	// GitMode selects which repository/worktree binding check to run (if any).
	GitMode ComposeGitMode
	// Mentions, when true, validates @-mentions against the payload worktree.
	Mentions bool
}

// ComposeGitMode selects which git binding validation ValidateCompose runs.
type ComposeGitMode int

const (
	// ComposeGitNone skips git binding checks (field validation only).
	ComposeGitNone ComposeGitMode = iota
	// ComposeGitBinding runs ValidateComposeGitBinding (drafts/templates).
	ComposeGitBinding
	// ComposeTaskRepoBinding runs ValidateTaskRepositoryBinding (task create).
	ComposeTaskRepoBinding
)

// ValidateCompose validates a compose payload: optional git binding, optional
// prompt mentions, then shared field checks (runner/model, pickup, title, etc.).
//
//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func (h *Handler) ValidateCompose(ctx context.Context, payload TaskComposePayloadJSON, settings settingsdomain.AppSettings, opts ValidateComposeOpts) error {
	switch opts.GitMode {
	case ComposeGitBinding:
		if err := h.gitCompose.ValidateComposeGitBinding(ctx, payload.RepositoryID, payload.ProjectID, payload.WorktreeID); err != nil {
			return err
		}
	case ComposeTaskRepoBinding:
		if err := h.gitCompose.ValidateTaskRepositoryBinding(ctx, payload.ProjectID, payload.RepositoryID); err != nil {
			return err
		}
	}
	if opts.Mentions {
		if err := h.gitCompose.ValidatePromptMentionsForWorktree(ctx, payload.WorktreeID, payload.InitialPrompt); err != nil {
			return err
		}
	}
	if _, _, err := resolveRunnerModelFields(payload.Runner, payload.CursorModel, settings); err != nil {
		return err
	}
	if _, err := resolvePickupNotBeforeForCreate(payload.PickupNotBefore, payload.Status, settings); err != nil {
		return err
	}
	if strings.TrimSpace(payload.Title) == "" {
		return fmt.Errorf("%w: title required", domain.ErrInvalidInput)
	}
	if payload.Priority == "" {
		return fmt.Errorf("%w: priority required", domain.ErrInvalidInput)
	}
	if _, err := parseCreateChecklistItems(payload.ChecklistItems); err != nil {
		return err
	}
	return nil
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func omitPastPickupNotBefore(t *time.Time) *time.Time {
	if t == nil {
		return nil
	}
	if t.Before(time.Now().UTC()) {
		return nil
	}
	return t
}
