package model

import (
	gitdomain "github.com/AlexsanderHamir/Hamix/pkgs/gitinventory/domain"
)

// ParityPair binds a domain struct prototype to its model counterpart for
// schema- and field-parity guards.
type ParityPair struct {
	Name   string
	Domain any
	Model  any
	Table  string
	// ModelMigrateExtra lists additional model structs AutoMigrate must run
	// before the primary model type (e.g. parent tables for association FKs).
	ModelMigrateExtra []any
}

// ParityPairs is the single registry both parity tests iterate.
var ParityPairs = []ParityPair{
	{
		Name:   "GitRepository",
		Domain: &gitdomain.GitRepository{},
		Model:  &GitRepository{},
		Table:  "git_repositories",
	},
	{
		Name:   "GitWorktree",
		Domain: &gitdomain.GitWorktree{},
		Model:  &GitWorktree{},
		Table:  "git_worktrees",
		ModelMigrateExtra: []any{
			&GitRepository{},
		},
	},
	{
		Name:   "GitBranch",
		Domain: &gitdomain.GitBranch{},
		Model:  &GitBranch{},
		Table:  "git_branches",
		ModelMigrateExtra: []any{
			&GitRepository{},
		},
	},
}
