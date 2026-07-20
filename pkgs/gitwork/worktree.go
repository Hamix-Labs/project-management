package gitwork

import "github.com/AlexsanderHamir/Hamix/pkgs/obs/calltrace"
import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
)

func (s *DefaultService) ListWorktrees(ctx context.Context, repo *Repository) ([]Worktree, error) {
	slog.DebugContext(ctx, "trace", "cmd", calltrace.LogCmd, "operation", "gitwork.ListWorktrees")
	if repo == nil {
		return nil, ErrNotARepository
	}
	out, err := s.runGit(ctx, repo.Root, "worktree", "list", "--porcelain")
	if err != nil {
		return nil, err
	}
	return parseWorktreePorcelain(out, repo.Root)
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func parseWorktreePorcelain(out, mainRoot string) ([]Worktree, error) {
	mainRoot, err := absPath(mainRoot)
	if err != nil {
		return nil, err
	}
	var worktrees []Worktree
	var cur *Worktree
	flush := func() error {
		if cur == nil {
			return nil
		}
		wtPath, err := absPath(cur.Path)
		if err != nil {
			return err
		}
		cur.Path = wtPath
		cur.IsMain = wtPath == mainRoot
		worktrees = append(worktrees, *cur)
		cur = nil
		return nil
	}
	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			if err := flush(); err != nil {
				return nil, err
			}
			continue
		}
		switch {
		case strings.HasPrefix(line, "worktree "):
			if err := flush(); err != nil {
				return nil, err
			}
			cur = &Worktree{Path: strings.TrimPrefix(line, "worktree ")}
		case strings.HasPrefix(line, "branch refs/heads/"):
			if cur != nil {
				cur.Branch = strings.TrimPrefix(line, "branch refs/heads/")
			}
		case line == "bare", line == "detached":
			if cur != nil {
				cur.Branch = ""
			}
		case strings.HasPrefix(line, "locked"):
			if cur != nil {
				cur.Locked = true
			}
		case strings.HasPrefix(line, "prunable"):
			if cur != nil {
				cur.Prunable = true
			}
		}
	}
	if err := flush(); err != nil {
		return nil, err
	}
	return worktrees, nil
}

func (s *DefaultService) AddWorktree(ctx context.Context, repo *Repository, path string, opts AddWorktreeOptions) (*Worktree, error) {
	slog.DebugContext(ctx, "trace", "cmd", calltrace.LogCmd, "operation", "gitwork.AddWorktree")
	if repo == nil {
		return nil, ErrNotARepository
	}
	absWT, err := absPath(path)
	if err != nil {
		return nil, err
	}
	if _, err := os.Stat(absWT); err == nil {
		return nil, ErrWorktreeExists
	} else if !os.IsNotExist(err) {
		return nil, err
	}
	args := []string{"worktree", "add"}
	if opts.CreateBranch {
		args = append(args, "-b", opts.Branch)
		args = append(args, absWT)
		if opts.StartPoint != "" {
			args = append(args, opts.StartPoint)
		}
	} else {
		args = append(args, absWT, opts.Branch)
	}
	if _, err := s.runGit(ctx, repo.Root, args...); err != nil {
		return nil, mapWorktreeAddErr(err)
	}
	list, err := s.ListWorktrees(ctx, repo)
	if err != nil {
		return nil, err
	}
	for _, wt := range list {
		if filepath.Clean(wt.Path) == filepath.Clean(absWT) {
			return &wt, nil
		}
	}
	return &Worktree{Path: absWT, Branch: opts.Branch, IsMain: absWT == repo.Root}, nil
}

func (s *DefaultService) RemoveWorktree(ctx context.Context, repo *Repository, path string, force bool) error {
	slog.DebugContext(ctx, "trace", "cmd", calltrace.LogCmd, "operation", "gitwork.RemoveWorktree")
	if repo == nil {
		return ErrNotARepository
	}
	absWT, err := absPath(path)
	if err != nil {
		return err
	}
	args := []string{"worktree", "remove"}
	if force {
		args = append(args, "-f")
	}
	args = append(args, absWT)
	if err := s.runGitRemove(ctx, repo.Root, args...); err != nil {
		return mapWorktreeRemoveErr(err)
	}
	return nil
}

//funclogmeasure:skip category=delegate-already-logs reason="Thin wrapper around runGit which emits git command trace."
func (s *DefaultService) runGitRemove(ctx context.Context, dir string, args ...string) error {
	_, err := s.runGit(ctx, dir, args...)
	return err
}
