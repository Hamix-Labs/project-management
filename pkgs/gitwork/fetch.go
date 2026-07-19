package gitwork

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/calltrace"
)

const defaultRemote = "origin"

func remoteOrDefault(remote string) string {
	remote = strings.TrimSpace(remote)
	if remote == "" {
		return defaultRemote
	}
	return remote
}

// Fetch runs `git fetch <remote>` in the repository root.
func (s *DefaultService) Fetch(ctx context.Context, repo *Repository, remote string) error {
	slog.DebugContext(ctx, "trace", "cmd", calltrace.LogCmd, "operation", "gitwork.Fetch")
	if repo == nil {
		return ErrNotARepository
	}
	remote = remoteOrDefault(remote)
	_, err := s.runGit(ctx, repo.Root, "fetch", remote)
	if err != nil {
		return fmt.Errorf("git fetch %s: %w", remote, err)
	}
	return nil
}

// ResolveDefaultBranch detects the remote default branch.
// Order: symbolic-ref refs/remotes/<remote>/HEAD, then <remote>/main, <remote>/master,
// then the current local branch name, finally "main".
func (s *DefaultService) ResolveDefaultBranch(ctx context.Context, repo *Repository, remote string) (string, error) {
	slog.DebugContext(ctx, "trace", "cmd", calltrace.LogCmd, "operation", "gitwork.ResolveDefaultBranch")
	if repo == nil {
		return "", ErrNotARepository
	}
	remote = remoteOrDefault(remote)

	out, err := s.runGit(ctx, repo.Root, "symbolic-ref", "--quiet", "--short", "refs/remotes/"+remote+"/HEAD")
	if err == nil {
		ref := strings.TrimSpace(out)
		prefix := remote + "/"
		if strings.HasPrefix(ref, prefix) {
			name := strings.TrimPrefix(ref, prefix)
			if name != "" {
				return name, nil
			}
		}
	}

	for _, candidate := range []string{"main", "master"} {
		if _, err := s.runGit(ctx, repo.Root, "rev-parse", "--verify", "--quiet", remote+"/"+candidate); err == nil {
			return candidate, nil
		}
	}

	branches, err := s.ListBranches(ctx, repo)
	if err == nil {
		for _, b := range branches {
			if b.IsCurrent && strings.TrimSpace(b.Name) != "" {
				return b.Name, nil
			}
		}
	}

	return "main", nil
}
