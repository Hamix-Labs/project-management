package storefake

import (
	"context"
	"time"

	gitcontract "github.com/AlexsanderHamir/Hamix/pkgs/gitinventory/contract"
	gitdomain "github.com/AlexsanderHamir/Hamix/pkgs/gitinventory/domain"
	"github.com/AlexsanderHamir/Hamix/pkgs/gitwork"
)

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (unimplementedHandlerStore) ListAllGitRepositories(context.Context) ([]gitdomain.GitRepository, error) {
	return nil, errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (unimplementedHandlerStore) ListAllGitRepositoriesWithSummary(context.Context) ([]gitcontract.GitRepositoryListSummary, error) {
	return nil, errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (unimplementedHandlerStore) ListGitRepositories(context.Context, string) ([]gitdomain.GitRepository, error) {
	return nil, errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (unimplementedHandlerStore) GetGitRepositoryByID(context.Context, string) (gitdomain.GitRepository, error) {
	return gitdomain.GitRepository{}, errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (unimplementedHandlerStore) GetGitRepository(context.Context, string, string) (gitdomain.GitRepository, error) {
	return gitdomain.GitRepository{}, errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (unimplementedHandlerStore) DeleteGlobalGitRepository(context.Context, string) error {
	return errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (unimplementedHandlerStore) DeleteGitRepository(context.Context, string, string) error {
	return errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (unimplementedHandlerStore) ListGitWorktreesByRepo(context.Context, string) ([]gitdomain.GitWorktree, error) {
	return nil, errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (unimplementedHandlerStore) ListGitWorktrees(context.Context, string, string) ([]gitdomain.GitWorktree, error) {
	return nil, errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (unimplementedHandlerStore) UnregisterGitWorktreeByID(context.Context, string) error {
	return errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (unimplementedHandlerStore) UnregisterGitWorktree(context.Context, string, string) error {
	return errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (unimplementedHandlerStore) ListGitBranchesByRepo(context.Context, string) ([]gitdomain.GitBranch, error) {
	return nil, errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (unimplementedHandlerStore) ListGitBranches(context.Context, string, string) ([]gitdomain.GitBranch, error) {
	return nil, errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (unimplementedHandlerStore) CreateGlobalGitRepository(context.Context, gitcontract.CreateGitRepositoryInput, gitwork.Service) (gitdomain.GitRepository, error) {
	return gitdomain.GitRepository{}, errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (unimplementedHandlerStore) CreateGitRepository(context.Context, string, gitcontract.CreateGitRepositoryInput, gitwork.Service) (gitdomain.GitRepository, error) {
	return gitdomain.GitRepository{}, errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (unimplementedHandlerStore) CreateGitWorktreeForRepo(context.Context, string, gitcontract.CreateGitWorktreeInput, gitwork.Service) (gitdomain.GitWorktree, error) {
	return gitdomain.GitWorktree{}, errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (unimplementedHandlerStore) AllocateTaskWorktree(context.Context, string, string, gitwork.Service) (gitdomain.GitWorktree, error) {
	return gitdomain.GitWorktree{}, errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (unimplementedHandlerStore) SyncGitRepository(context.Context, string, gitwork.Service) (gitcontract.ReconcileGitOutput, error) {
	return gitcontract.ReconcileGitOutput{}, errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (unimplementedHandlerStore) WorktreeStaleMap(context.Context, string, time.Time) (map[string]bool, error) {
	return nil, errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (unimplementedHandlerStore) CreateGitWorktree(context.Context, string, string, gitcontract.CreateGitWorktreeInput, gitwork.Service) (gitdomain.GitWorktree, error) {
	return gitdomain.GitWorktree{}, errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (unimplementedHandlerStore) RemoveGitWorktreeFromDiskByID(context.Context, string, bool, gitwork.Service) error {
	return errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (unimplementedHandlerStore) RemoveGitWorktreeFromDisk(context.Context, string, string, bool, gitwork.Service) error {
	return errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (unimplementedHandlerStore) CreateGitBranch(context.Context, string, string, gitcontract.CreateGitBranchInput, gitwork.Service) (gitdomain.GitBranch, error) {
	return gitdomain.GitBranch{}, errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (unimplementedHandlerStore) DeleteGitBranch(context.Context, string, string, bool, gitwork.Service) error {
	return errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (unimplementedHandlerStore) RepoWorktreeInventory(context.Context, gitdomain.GitRepository, gitwork.Service) ([]gitcontract.WorktreeInventoryRow, error) {
	return nil, errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (unimplementedHandlerStore) RepoWorktreeCheckoutStatus(context.Context, gitdomain.GitRepository, gitwork.Service) ([]gitcontract.WorktreeCheckoutStatusRow, error) {
	return nil, errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (unimplementedHandlerStore) ProbeGitWorktree(context.Context, string, string, gitwork.Service) (gitcontract.GitWorktreeProbeResult, error) {
	return gitcontract.GitWorktreeProbeResult{}, errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (unimplementedHandlerStore) RegisterExistingGitWorktree(context.Context, string, string, string, gitcontract.BindBranchInput, gitwork.Service) (gitdomain.GitWorktree, error) {
	return gitdomain.GitWorktree{}, errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (unimplementedHandlerStore) ReconcileGitRepository(context.Context, string, string, gitcontract.ReconcileGitInput, gitwork.Service) (gitcontract.ReconcileGitOutput, error) {
	return gitcontract.ReconcileGitOutput{}, errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (unimplementedHandlerStore) RelocateGitRepository(context.Context, string, string, string, gitwork.Service) (gitcontract.ReconcileGitOutput, error) {
	return gitcontract.ReconcileGitOutput{}, errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (unimplementedHandlerStore) RelocateGitWorktree(context.Context, string, string, gitwork.Service) (gitdomain.GitWorktree, error) {
	return gitdomain.GitWorktree{}, errNotImplemented
}
