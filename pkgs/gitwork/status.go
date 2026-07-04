package gitwork

import "github.com/AlexsanderHamir/Hamix/pkgs/tasks/calltrace"
import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"
)

// CheckoutStatus is live git state for one worktree checkout directory.
type CheckoutStatus struct {
	Dirty        bool
	Detached     bool
	HeadSHA      string
	HeadCommitAt time.Time
	Ahead        int
	Behind       int
	Upstream     string
	HasUpstream  bool
}

// CheckoutStatus reads dirty/clean, HEAD commit time, and upstream ahead/behind for worktreePath.
func (s *DefaultService) CheckoutStatus(ctx context.Context, worktreePath string) (CheckoutStatus, error) {
	slog.DebugContext(ctx, "trace", "cmd", calltrace.LogCmd, "operation", "gitwork.CheckoutStatus", "path", worktreePath)
	var st CheckoutStatus

	porcelain, err := s.runGit(ctx, worktreePath, "status", "--porcelain")
	if err != nil {
		return st, err
	}
	st.Dirty = strings.TrimSpace(porcelain) != ""

	headRef, err := s.runGit(ctx, worktreePath, "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return st, err
	}
	headRef = strings.TrimSpace(headRef)
	st.Detached = headRef == "HEAD"

	logOut, err := s.runGit(ctx, worktreePath, "log", "-1", "--format=%H%x09%cI", "HEAD")
	if err != nil {
		return st, err
	}
	sha, commitAt, err := parseHeadLogLine(logOut)
	if err != nil {
		return st, fmt.Errorf("parse HEAD log: %w", err)
	}
	st.HeadSHA = sha
	st.HeadCommitAt = commitAt

	if st.Detached {
		return st, nil
	}

	upstream, err := s.runGit(ctx, worktreePath, "rev-parse", "--abbrev-ref", "@{u}")
	if err != nil || strings.TrimSpace(upstream) == "" {
		return st, nil
	}
	st.HasUpstream = true
	st.Upstream = strings.TrimSpace(upstream)

	countOut, err := s.runGit(ctx, worktreePath, "rev-list", "--left-right", "--count", "@{u}...HEAD")
	if err != nil {
		return st, nil
	}
	behind, ahead, err := parseUpstreamCounts(countOut)
	if err != nil {
		return st, nil
	}
	st.Behind = behind
	st.Ahead = ahead
	return st, nil
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by CheckoutStatus."
func parseHeadLogLine(out string) (sha string, commitAt time.Time, err error) {
	line := strings.TrimSpace(out)
	if line == "" {
		return "", time.Time{}, fmt.Errorf("empty log output")
	}
	parts := strings.SplitN(line, "\t", 2)
	if len(parts) != 2 {
		return "", time.Time{}, fmt.Errorf("unexpected log format %q", line)
	}
	sha = strings.TrimSpace(parts[0])
	ts, parseErr := time.Parse(time.RFC3339, strings.TrimSpace(parts[1]))
	if parseErr != nil {
		return "", time.Time{}, parseErr
	}
	return sha, ts.UTC(), nil
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by CheckoutStatus."
func parseUpstreamCounts(out string) (behind, ahead int, err error) {
	line := strings.TrimSpace(out)
	if line == "" {
		return 0, 0, fmt.Errorf("empty rev-list count")
	}
	parts := strings.Fields(line)
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("unexpected count format %q", line)
	}
	behind, err = strconv.Atoi(parts[0])
	if err != nil {
		return 0, 0, fmt.Errorf("parse behind: %w", err)
	}
	ahead, err = strconv.Atoi(parts[1])
	if err != nil {
		return 0, 0, fmt.Errorf("parse ahead: %w", err)
	}
	return behind, ahead, nil
}
