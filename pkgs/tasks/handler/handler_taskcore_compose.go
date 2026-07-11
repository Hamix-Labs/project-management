package handler

import (
	"context"

	settingsdomain "github.com/AlexsanderHamir/Hamix/pkgs/settings/domain"
	taskcorehandler "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/handler"
)

func (h *Handler) ValidateTaskGitBindingV2(ctx context.Context, projectID *string, worktreeID *string) error {
	return h.validateTaskGitBindingV2(ctx, projectID, worktreeID)
}

func (h *Handler) ValidateComposeGitBinding(ctx context.Context, repositoryID, projectID, worktreeID *string) error {
	return h.validateComposeGitBinding(ctx, repositoryID, projectID, worktreeID)
}

func (h *Handler) ValidatePromptMentionsForWorktree(ctx context.Context, worktreeID *string, prompt string) error {
	return h.validatePromptMentionsForWorktree(ctx, worktreeID, prompt)
}

//funclogmeasure:skip category=delegate-already-logs reason="Validation delegate; taskcore handler emits trace at the HTTP chokepoint."
func (h *Handler) validateComposePayload(ctx context.Context, payload taskComposePayloadJSON, settings settingsdomain.AppSettings) error {
	return h.taskcoreHandler().ValidateComposePayload(ctx, taskcorehandler.TaskComposePayloadJSON(payload), settings)
}
