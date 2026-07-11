// Package handler registers /git/* REST routes for taskapi.
package handler

import (
	"net/http"

	"github.com/AlexsanderHamir/Hamix/pkgs/gitwork"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/contract"
)

// HostPathMapper translates container paths for JSON host_path fields.
type HostPathMapper interface {
	DisplayHostPath(containerPath string) string
}

type noopHostPathMapper struct{}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func (noopHostPathMapper) DisplayHostPath(p string) string { return p }

// Deps wires git inventory HTTP handlers into the taskapi mux.
type Deps struct {
	Read       contract.GitReadStore
	Write      contract.GitWriteStore
	Projects   contract.ProjectStore
	GitService gitwork.Service
	HostPaths  HostPathMapper
}

// Handler serves global git inventory REST routes.
type Handler struct {
	read     contract.GitReadStore
	write    contract.GitWriteStore
	projects contract.ProjectStore
	git      gitwork.Service
	paths    HostPathMapper
}

// Register mounts /git/* routes on m.
//
//funclogmeasure:skip category=hot-path reason="Route table wiring only; operation trace is emitted by registered handlers."
func Register(m *http.ServeMux, deps Deps) {
	paths := deps.HostPaths
	if paths == nil {
		paths = noopHostPathMapper{}
	}
	gitSvc := deps.GitService
	if gitSvc == nil {
		gitSvc = gitwork.New()
	}
	h := &Handler{
		read:     deps.Read,
		write:    deps.Write,
		projects: deps.Projects,
		git:      gitSvc,
		paths:    paths,
	}
	m.Handle("GET /git/repositories", http.HandlerFunc(h.listGlobalGitRepositories))
	m.Handle("POST /git/repositories", http.HandlerFunc(h.createGlobalGitRepository))
	m.Handle("GET /git/repositories/{repoId}", http.HandlerFunc(h.getGlobalGitRepository))
	m.Handle("DELETE /git/repositories/{repoId}", http.HandlerFunc(h.deleteGlobalGitRepository))
	m.Handle("GET /git/repositories/{repoId}/worktrees", http.HandlerFunc(h.listGlobalGitWorktrees))
	m.Handle("GET /git/repositories/{repoId}/worktrees/live", http.HandlerFunc(h.listGlobalGitWorktreesLive))
	m.Handle("GET /git/repositories/{repoId}/worktrees/checkout-status", http.HandlerFunc(h.listGlobalGitWorktreesCheckoutStatus))
	m.Handle("GET /git/repositories/{repoId}/worktrees/probe", http.HandlerFunc(h.probeGlobalGitWorktree))
	m.Handle("POST /git/repositories/{repoId}/worktrees", http.HandlerFunc(h.createGlobalGitWorktree))
	m.Handle("POST /git/repositories/{repoId}/worktrees/register", http.HandlerFunc(h.registerGlobalGitWorktree))
	m.Handle("POST /git/repositories/{repoId}/reconcile", http.HandlerFunc(h.reconcileGlobalGitRepository))
	m.Handle("POST /git/repositories/{repoId}/relocate", http.HandlerFunc(h.relocateGlobalGitRepository))
	m.Handle("POST /git/worktrees/{worktreeId}/relocate", http.HandlerFunc(h.relocateGlobalGitWorktree))
	m.Handle("DELETE /git/worktrees/{worktreeId}", http.HandlerFunc(h.deleteGlobalGitWorktree))
	m.Handle("GET /git/repositories/{repoId}/branches", http.HandlerFunc(h.listGlobalGitBranches))
	m.Handle("GET /git/repositories/{repoId}/branches/live", http.HandlerFunc(h.listGlobalGitBranchesLive))
	m.Handle("GET /git/repositories/{repoId}/projects", http.HandlerFunc(h.listRepoProjects))
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func (h *Handler) gitService() gitwork.Service {
	if h.git != nil {
		return h.git
	}
	return gitwork.New()
}
