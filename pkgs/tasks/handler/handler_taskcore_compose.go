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

func (h *Handler) ValidateTaskRepositoryBinding(ctx context.Context, projectID, repositoryID *string) error {
	return h.validateTaskRepositoryBinding(ctx, projectID, repositoryID)
}

func (h *Handler) AllocateTaskWorktree(ctx context.Context, repositoryID, taskID string) (string, error) {
	return h.allocateTaskWorktree(ctx, repositoryID, taskID)
}

func (h *Handler) ValidatePromptMentionsForWorktree(ctx context.Context, worktreeID *string, prompt string) error {
	return h.validatePromptMentionsForWorktree(ctx, worktreeID, prompt)
}

func (h *Handler) ValidatePromptMentionsForRepository(ctx context.Context, repositoryID, prompt string) error {
	return h.validatePromptMentionsForRepository(ctx, repositoryID, prompt)
}

func (h *Handler) ValidatePromptMentionsForProject(ctx context.Context, projectID *string, prompt string) error {
	return h.validatePromptMentionsForProject(ctx, projectID, prompt)
}

//funclogmeasure:skip category=delegate-already-logs reason="Validation delegate; taskcore handler emits trace at the HTTP chokepoint."
func (h *Handler) validateComposePayload(ctx context.Context, payload taskcorehandler.TaskComposePayloadJSON, settings settingsdomain.AppSettings) error {
	return h.taskcoreHandler().ValidateCompose(ctx, taskcorehandler.TaskComposePayloadJSON(payload), settings, taskcorehandler.ValidateComposeOpts{
		GitMode: taskcorehandler.ComposeGitBinding,
	})
}
