package handler

import (
	repopkg "github.com/AlexsanderHamir/Hamix/pkgs/repo"
	repohandler "github.com/AlexsanderHamir/Hamix/pkgs/repo/handler"
)

// RepoProvider resolves workspace roots for /repo/* and @-mention validation.
type RepoProvider = repopkg.RepoProvider

// GitWorktreeResolver loads git worktree rows for repo path resolution.
type GitWorktreeResolver = repopkg.GitWorktreeResolver

var (
	NewStaticRepoProvider   = repopkg.NewStaticRepoProvider
	NewSettingsRepoProvider = repopkg.NewSettingsRepoProvider
)

const (
	RepoReasonOpenFailed         = repopkg.RepoReasonOpenFailed
	RepoReasonWorktreeIDRequired = repopkg.RepoReasonWorktreeIDRequired
	RepoReasonWorktreeNotFound   = repopkg.RepoReasonWorktreeNotFound
	maxRepoSearchQueryBytes      = repohandler.MaxSearchQueryBytes
	maxRepoRelPathQueryBytes     = repohandler.MaxRelPathQueryBytes
	maxRepoLineQueryParamBytes   = repohandler.MaxLineQueryParamBytes
)
