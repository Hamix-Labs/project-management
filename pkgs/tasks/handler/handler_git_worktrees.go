package handler

import (
	"time"

	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/domain"
)

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func (h *Handler) gitWorktreeJSON(w domain.GitWorktree) gitWorktreeJSON {
	return gitWorktreeJSON{
		ID:           w.ID,
		RepositoryID: w.RepositoryID,
		Path:         w.Path,
		HostPath:     h.pathMap.DisplayHostPath(w.Path),
		Name:         w.Name,
		IsMain:       w.IsMain,
		BranchID:     w.BranchID,
		CreatedAt:    w.CreatedAt.UTC().Format(time.RFC3339),
	}
}
