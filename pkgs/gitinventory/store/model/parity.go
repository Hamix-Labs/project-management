package model

import (
	gitdomain "github.com/AlexsanderHamir/Hamix/pkgs/gitinventory/domain"
	"github.com/AlexsanderHamir/Hamix/pkgs/storekernel/parity"
)

// ParityPair is the BC-local name for the shared parity registry entry type.
type ParityPair = parity.Pair

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
