package handler

import (
	"time"

	"github.com/AlexsanderHamir/Hamix/pkgs/gitinventory/domain"
)

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func toGitBranchJSON(b domain.GitBranch) gitBranchJSON {
	return gitBranchJSON{
		ID:           b.ID,
		RepositoryID: b.RepositoryID,
		Name:         b.Name,
		HeadSHA:      b.HeadSHA,
		CreatedAt:    b.CreatedAt.UTC().Format(time.RFC3339),
	}
}
