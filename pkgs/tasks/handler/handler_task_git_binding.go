package handler

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/AlexsanderHamir/Hamix/pkgs/obs/calltrace"
	"github.com/AlexsanderHamir/Hamix/pkgs/repo"
	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
)

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func trimmedOptionalID(id *string) string {
	if id == nil {
		return ""
	}
	return strings.TrimSpace(*id)
}

func (h *Handler) validateTaskGitBindingV2(
	ctx context.Context,
	projectID *string,
	worktreeID *string,
) error {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "handler.Handler.validateTaskGitBindingV2")
	pid := trimmedOptionalID(projectID)
	if pid == "" {
		return fmt.Errorf("%w: project_id required", taskcoredomain.ErrInvalidInput)
	}
	wtID := trimmedOptionalID(worktreeID)
	if wtID == "" {
		return fmt.Errorf("%w: worktree_id required", taskcoredomain.ErrInvalidInput)
	}
	proj, err := h.store.GetProject(ctx, pid)
	if err != nil {
		return err
	}
	if !proj.IsDefault {
		if proj.RepositoryID == nil || strings.TrimSpace(*proj.RepositoryID) == "" {
			return fmt.Errorf("%w: project not bound to repository", taskcoredomain.ErrInvalidInput)
		}
	}
	return h.store.ValidateTaskWorktreeBinding(ctx, projectID, wtID)
}

func (h *Handler) validateTaskRepositoryBinding(
	ctx context.Context,
	projectID *string,
	repositoryID *string,
) error {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "handler.Handler.validateTaskRepositoryBinding")
	pid := trimmedOptionalID(projectID)
	if pid == "" {
		return fmt.Errorf("%w: project_id required", taskcoredomain.ErrInvalidInput)
	}
	repoID := trimmedOptionalID(repositoryID)
	if repoID == "" {
		return fmt.Errorf("%w: repository_id required", taskcoredomain.ErrInvalidInput)
	}
	proj, err := h.store.GetProject(ctx, pid)
	if err != nil {
		return err
	}
	if !proj.IsDefault {
		if proj.RepositoryID == nil || strings.TrimSpace(*proj.RepositoryID) == "" {
			return fmt.Errorf("%w: project not bound to repository", taskcoredomain.ErrInvalidInput)
		}
		if strings.TrimSpace(*proj.RepositoryID) != repoID {
			return fmt.Errorf("%w: repository_id does not match project", taskcoredomain.ErrInvalidInput)
		}
	}
	if _, err := h.store.GetGitRepositoryByID(ctx, repoID); err != nil {
		return err
	}
	return nil
}

func (h *Handler) allocateTaskWorktree(ctx context.Context, repositoryID, taskID string) (string, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "handler.Handler.allocateTaskWorktree")
	wt, err := h.store.AllocateTaskWorktree(ctx, repositoryID, taskID)
	if err != nil {
		return "", err
	}
	return wt.ID, nil
}

func (h *Handler) validateComposeGitBinding(
	ctx context.Context,
	repositoryID *string,
	projectID *string,
	worktreeID *string,
) error {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "handler.Handler.validateComposeGitBinding")
	if err := h.validateTaskRepositoryBinding(ctx, projectID, repositoryID); err != nil {
		return err
	}
	wtID := trimmedOptionalID(worktreeID)
	if wtID == "" {
		return nil
	}
	return h.store.ValidateTaskWorktreeBinding(ctx, projectID, wtID)
}

func (h *Handler) validatePromptMentionsForWorktree(ctx context.Context, worktreeID *string, prompt string) error {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "handler.Handler.validatePromptMentionsForWorktree")
	wtID := trimmedOptionalID(worktreeID)
	if wtID == "" {
		if len(repo.ParseFileMentions(prompt)) > 0 {
			return fmt.Errorf("%w: worktree_id required for @-mentions", taskcoredomain.ErrInvalidInput)
		}
		return nil
	}
	return h.validatePromptMentionsForWorktreeID(ctx, wtID, prompt)
}

func (h *Handler) validatePromptMentionsForRepository(ctx context.Context, repositoryID, prompt string) error {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "handler.Handler.validatePromptMentionsForRepository")
	if len(repo.ParseFileMentions(prompt)) == 0 {
		return nil
	}
	repositoryID = strings.TrimSpace(repositoryID)
	if repositoryID == "" {
		return fmt.Errorf("%w: repository_id required for @-mentions", taskcoredomain.ErrInvalidInput)
	}
	wts, err := h.store.ListGitWorktreesByRepo(ctx, repositoryID)
	if err != nil {
		return err
	}
	var mainID string
	for _, wt := range wts {
		if wt.IsMain {
			mainID = wt.ID
			break
		}
	}
	if mainID == "" {
		return fmt.Errorf("%w: repository has no main worktree for @-mentions", taskcoredomain.ErrInvalidInput)
	}
	return h.validatePromptMentionsForWorktreeID(ctx, mainID, prompt)
}

func (h *Handler) validatePromptMentionsForProject(ctx context.Context, projectID *string, prompt string) error {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "handler.Handler.validatePromptMentionsForProject")
	if len(repo.ParseFileMentions(prompt)) == 0 {
		return nil
	}
	pid := trimmedOptionalID(projectID)
	if pid == "" {
		return fmt.Errorf("%w: project_id required for @-mentions", taskcoredomain.ErrInvalidInput)
	}
	proj, err := h.store.GetProject(ctx, pid)
	if err != nil {
		return err
	}
	if proj.RepositoryID == nil || strings.TrimSpace(*proj.RepositoryID) == "" {
		return fmt.Errorf("%w: project not bound to repository", taskcoredomain.ErrInvalidInput)
	}
	return h.validatePromptMentionsForRepository(ctx, *proj.RepositoryID, prompt)
}

func (h *Handler) validatePromptMentionsForWorktreeID(ctx context.Context, worktreeID, prompt string) error {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "handler.Handler.validatePromptMentionsForWorktreeID")
	if h.repoProv == nil {
		return nil
	}
	worktreeID = strings.TrimSpace(worktreeID)
	if worktreeID == "" {
		if len(repo.ParseFileMentions(prompt)) > 0 {
			return fmt.Errorf("%w: worktree_id required for @-mentions", taskcoredomain.ErrInvalidInput)
		}
		return nil
	}
	root, reason, err := h.repoProv.OpenWorktreeRoot(ctx, worktreeID)
	if err != nil {
		return err
	}
	if root == nil {
		if reason == RepoReasonWorktreeNotFound {
			return fmt.Errorf("%w: worktree not found", taskcoredomain.ErrNotFound)
		}
		return fmt.Errorf("%w: %s", taskcoredomain.ErrInvalidInput, reason)
	}
	return root.ValidatePromptMentions(prompt)
}
