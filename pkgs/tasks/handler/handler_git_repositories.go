package handler

import (
	"time"

	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/domain"
)

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func (h *Handler) gitRepositoryJSON(r domain.GitRepository) gitRepositoryJSON {
	return h.gitRepositorySummaryJSON(r, "", 0)
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func (h *Handler) gitRepositorySummaryJSON(r domain.GitRepository, mainBranch string, linkedWorktrees int) gitRepositoryJSON {
	return gitRepositoryJSON{
		ID:                  r.ID,
		Path:                r.Path,
		GitCommonDir:        r.GitCommonDir,
		HostPath:            h.pathMap.DisplayHostPath(r.Path),
		DefaultBranch:       r.DefaultBranch,
		MainBranchName:      mainBranch,
		LinkedWorktreeCount: linkedWorktrees,
		CreatedAt:           r.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:           r.UpdatedAt.UTC().Format(time.RFC3339),
	}
}
