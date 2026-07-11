package repo

import (
	"context"
	"log/slog"
	"strings"

	"github.com/AlexsanderHamir/Hamix/pkgs/gitinventory/domain"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/calltrace"
)

// GitWorktreeResolver loads git worktree rows for repo path resolution.
type GitWorktreeResolver interface {
	GetGitWorktreeByID(ctx context.Context, id string) (domain.GitWorktree, error)
}

// RepoProvider returns the *Root that /repo/* handlers and prompt mention
// validation should consult for the current request. Production wiring
// resolves paths from git_worktrees via OpenWorktreeRoot.
type RepoProvider interface {
	OpenWorktreeRoot(ctx context.Context, worktreeID string) (root *Root, reason string, err error)
}

const (
	// RepoReasonOpenFailed: worktree path is set but OpenRoot rejected it.
	RepoReasonOpenFailed = "worktree_open_failed"
	// RepoReasonWorktreeIDRequired: /repo/* called without worktree_id.
	RepoReasonWorktreeIDRequired = "worktree_id_required"
	// RepoReasonWorktreeNotFound: unknown worktree_id.
	RepoReasonWorktreeNotFound = "worktree_not_found"
)

type staticRepoProvider struct {
	root *Root
}

// NewStaticRepoProvider wraps r so OpenWorktreeRoot returns it when worktreeID is non-empty.
func NewStaticRepoProvider(r *Root) RepoProvider {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "repo.NewStaticRepoProvider")
	return &staticRepoProvider{root: r}
}

func (p *staticRepoProvider) OpenWorktreeRoot(_ context.Context, worktreeID string) (*Root, string, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "repo.staticRepoProvider.OpenWorktreeRoot")
	if strings.TrimSpace(worktreeID) == "" {
		return nil, RepoReasonWorktreeIDRequired, nil
	}
	if p.root == nil {
		return nil, RepoReasonWorktreeNotFound, nil
	}
	return p.root, "", nil
}

type settingsRepoProvider struct {
	resolver GitWorktreeResolver
}

// NewSettingsRepoProvider returns a provider backed by git_worktrees in r.
func NewSettingsRepoProvider(r GitWorktreeResolver) RepoProvider {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "repo.NewSettingsRepoProvider")
	return &settingsRepoProvider{resolver: r}
}

func (p *settingsRepoProvider) OpenWorktreeRoot(ctx context.Context, worktreeID string) (*Root, string, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "repo.settingsRepoProvider.OpenWorktreeRoot")
	if p == nil || p.resolver == nil {
		return nil, RepoReasonWorktreeNotFound, nil
	}
	worktreeID = strings.TrimSpace(worktreeID)
	if worktreeID == "" {
		return nil, RepoReasonWorktreeIDRequired, nil
	}
	wt, err := p.resolver.GetGitWorktreeByID(ctx, worktreeID)
	if err != nil {
		if domain.GitErrCode(err) == domain.GitCodeWorktreeNotFound {
			return nil, RepoReasonWorktreeNotFound, nil
		}
		return nil, "", err
	}
	root, openErr := OpenRoot(wt.Path)
	if openErr != nil {
		return nil, RepoReasonOpenFailed, openErr
	}
	return root, "", nil
}
