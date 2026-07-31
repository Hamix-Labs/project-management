package git

import "github.com/AlexsanderHamir/Hamix/pkgs/obs/calltrace"
import (
	"context"
	"errors"
	"fmt"
	cyclesstore "github.com/AlexsanderHamir/Hamix/pkgs/taskcycles/store"
	"log/slog"
	"strings"
	"time"

	"github.com/AlexsanderHamir/Hamix/pkgs/agents/sidecar"
	cyclesdomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcycles/domain"
)

const (
	// ExecuteInvalidCommitReason is recorded when register SHAs cannot be resolved
	// or R\H is non-empty (ADR-0093).
	ExecuteInvalidCommitReason = "execute_invalid_commit"
	// ExecuteMissingCommitsReason is recorded when the commit register is empty.
	ExecuteMissingCommitsReason = "execute_missing_commits"
	// ExecuteUnregisteredCommitsReason is recorded when HEAD advanced outside the register.
	ExecuteUnregisteredCommitsReason = "execute_unregistered_commits"

	RetryResetAnchorMissing = "retry_reset_anchor_missing"
	RetryGitResetFailed     = "retry_git_reset_failed"
)

// ErrRetryResetAnchorMissing is returned when a fresh retry cannot resolve a git reset anchor.
var ErrRetryResetAnchorMissing = errors.New(RetryResetAnchorMissing)

// ExecuteCommitIngestOutcome summarizes commit indexing after execute.
type ExecuteCommitIngestOutcome struct {
	FailReason  string
	CommitCount int
}

// FreshRetryResetOutcome reports whether fresh-retry git reset was skipped.
type FreshRetryResetOutcome struct {
	Skipped bool
}

type phaseContext struct {
	Repo         string
	Worktree     string
	BaseSHA      string
	CycleBaseSHA string
	BaseBranch   string
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func (s *Service) commitDetails(ctx context.Context, worktree, sha string) (message string, committedAt time.Time, err error) {
	out, err := s.repo().Run(ctx, worktree, "log", "-1", "--format=%s%n%ci", sha)
	if err != nil {
		return "", time.Time{}, err
	}
	lines := strings.Split(strings.TrimSpace(out), "\n")
	if len(lines) == 0 {
		return "", time.Time{}, nil
	}
	msg := strings.TrimSpace(lines[0])
	if len(lines) < 2 {
		return msg, time.Time{}, nil
	}
	ts, parseErr := time.Parse("2006-01-02 15:04:05 -0700", strings.TrimSpace(lines[1]))
	if parseErr != nil {
		ts, parseErr = time.Parse(time.RFC3339, strings.TrimSpace(lines[1]))
	}
	if parseErr != nil {
		return msg, time.Time{}, nil
	}
	return msg, ts.UTC(), nil
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func (s *Service) branchContaining(ctx context.Context, worktree, sha string) (string, error) {
	out, err := s.repo().Run(ctx, worktree, "branch", "--contains", sha, "--format=%(refname:short)")
	if err != nil {
		return "", err
	}
	lines := strings.Split(strings.TrimSpace(out), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			return line, nil
		}
	}
	return "", nil
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func (s *Service) commitExists(ctx context.Context, worktree, sha string) bool {
	sha = strings.TrimSpace(sha)
	if sha == "" {
		return false
	}
	_, err := s.repo().Run(ctx, worktree, "cat-file", "-e", sha+"^{commit}")
	return err == nil
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func (s *Service) normalizeSHA(ctx context.Context, worktree, sha string) (string, error) {
	sha = strings.TrimSpace(sha)
	if sha == "" {
		return "", fmt.Errorf("empty sha")
	}
	out, err := s.repo().Run(ctx, worktree, "rev-parse", sha+"^{commit}")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(out), nil
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func (s *Service) revListCycleRange(ctx context.Context, worktree, cycleBaseSHA string) ([]string, error) {
	cycleBaseSHA = strings.TrimSpace(cycleBaseSHA)
	if cycleBaseSHA == "" {
		return nil, fmt.Errorf("empty cycle_base_sha")
	}
	out, err := s.repo().Run(ctx, worktree, "rev-list", "--reverse", cycleBaseSHA+"..HEAD")
	if err != nil {
		return nil, err
	}
	out = strings.TrimSpace(out)
	if out == "" {
		return nil, nil
	}
	lines := strings.Split(out, "\n")
	shas := make([]string, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			shas = append(shas, line)
		}
	}
	return shas, nil
}

func (s *Service) warnMissingIndexedCommits(ctx context.Context, taskID string, worktree string) {
	prior, err := s.store.ListCommitsForTask(ctx, taskID)
	if err != nil {
		return
	}
	for _, row := range prior {
		if !s.commitExists(ctx, worktree, row.SHA) {
			slog.Warn("indexed commit no longer in repository",
				"cmd", calltrace.LogCmd, "operation", "agent.harness.git.IngestExecuteCommits.missing_prior_sha",
				"task_id", taskID, "sha", row.SHA)
		}
	}
}

// PhaseContext carries git anchors for commit resolution.
type PhaseContext struct {
	Repo         string
	Worktree     string
	BaseSHA      string
	CycleBaseSHA string
	BaseBranch   string
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func phaseContextFromSnapshot(snap PhaseSnapshot) phaseContext {
	return phaseContext{
		Repo:         snap.Repo,
		Worktree:     snap.Worktree,
		BaseSHA:      snap.BaseSHA,
		CycleBaseSHA: snap.CycleBaseSHA,
		BaseBranch:   snap.BaseBranch,
	}
}

//funclogmeasure:skip category=hot-path reason="Git sub-step; operation trace is emitted by IngestExecuteCommits."
func (s *Service) resolveRegisterEntries(
	ctx context.Context,
	g phaseContext,
	entries []sidecar.CommitRegisterEntry,
	execPhaseSeq int64,
) ([]cyclesstore.CycleCommitEntry, error) {
	out := make([]cyclesstore.CycleCommitEntry, 0, len(entries))
	for i, entry := range entries {
		sha := strings.TrimSpace(entry.SHA)
		full, err := s.normalizeSHA(ctx, g.Worktree, sha)
		if err != nil {
			return nil, fmt.Errorf("commit %s not found in repository: %w", sha, err)
		}
		msg, at, err := s.commitDetails(ctx, g.Worktree, full)
		if err != nil {
			return nil, err
		}
		branch := strings.TrimSpace(entry.Branch)
		if branch == "" {
			resolved, branchErr := s.branchContaining(ctx, g.Worktree, full)
			if branchErr != nil {
				slog.Warn("branchContaining failed; falling back to base branch",
					"cmd", calltrace.LogCmd, "operation", "agent.harness.git.resolveRegisterEntries.branch_err",
					"sha", full, "err", branchErr)
			} else {
				branch = resolved
			}
		}
		if branch == "" {
			branch = g.BaseBranch
		}
		out = append(out, cyclesstore.CycleCommitEntry{
			Seq:         int64(i + 1),
			Repo:        g.Repo,
			Worktree:    g.Worktree,
			Branch:      branch,
			SHA:         full,
			CommittedAt: at,
			Message:     msg,
			PhaseSeq:    execPhaseSeq,
		})
	}
	return out, nil
}

// IngestExecuteCommits validates the MCP commit register against cycle_base..HEAD
// (exact set equality) and upserts register SHAs into task_cycle_commits (ADR-0093).
func (s *Service) IngestExecuteCommits(
	ctx context.Context,
	taskID string,
	cycle *cyclesdomain.TaskCycle,
	execPhaseSeq int64,
	snap PhaseSnapshot,
	publish func(taskID, cycleID string),
) (ExecuteCommitIngestOutcome, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "agent.harness.git.IngestExecuteCommits",
		"task_id", taskID, "cycle_id", cycle.ID, "phase_seq", execPhaseSeq)
	if snap.Skipped {
		return ExecuteCommitIngestOutcome{}, nil
	}
	g := phaseContextFromSnapshot(snap)
	s.warnMissingIndexedCommits(ctx, taskID, g.Worktree)

	if strings.TrimSpace(g.CycleBaseSHA) == "" {
		return ExecuteCommitIngestOutcome{FailReason: ExecuteInvalidCommitReason}, nil
	}

	regEntries, err := sidecar.ParseCommitRegister(s.reportDir, cycle.ID)
	if err != nil {
		if errors.Is(err, sidecar.ErrCommitRegisterInvalid) {
			return ExecuteCommitIngestOutcome{FailReason: ExecuteInvalidCommitReason}, nil
		}
		return ExecuteCommitIngestOutcome{}, err
	}
	if len(regEntries) == 0 {
		return ExecuteCommitIngestOutcome{FailReason: ExecuteMissingCommitsReason}, nil
	}

	headSHAs, err := s.revListCycleRange(ctx, g.Worktree, g.CycleBaseSHA)
	if err != nil {
		return ExecuteCommitIngestOutcome{FailReason: ExecuteInvalidCommitReason}, nil
	}

	resolved := make([]sidecar.CommitRegisterEntry, 0, len(regEntries))
	rSet := make(map[string]struct{}, len(regEntries))
	for _, e := range regEntries {
		full, nerr := s.normalizeSHA(ctx, g.Worktree, e.SHA)
		if nerr != nil {
			return ExecuteCommitIngestOutcome{FailReason: ExecuteInvalidCommitReason}, nil
		}
		if _, dup := rSet[full]; dup {
			return ExecuteCommitIngestOutcome{FailReason: ExecuteInvalidCommitReason}, nil
		}
		rSet[full] = struct{}{}
		e.SHA = full
		resolved = append(resolved, e)
	}

	hSet := make(map[string]struct{}, len(headSHAs))
	for _, sha := range headSHAs {
		hSet[sha] = struct{}{}
	}

	for sha := range hSet {
		if _, ok := rSet[sha]; !ok {
			return ExecuteCommitIngestOutcome{FailReason: ExecuteUnregisteredCommitsReason}, nil
		}
	}
	for sha := range rSet {
		if _, ok := hSet[sha]; !ok {
			return ExecuteCommitIngestOutcome{FailReason: ExecuteInvalidCommitReason}, nil
		}
	}

	entries, err := s.resolveRegisterEntries(ctx, g, resolved, execPhaseSeq)
	if err != nil {
		return ExecuteCommitIngestOutcome{FailReason: ExecuteInvalidCommitReason}, nil
	}
	if err := s.store.UpsertCycleCommits(ctx, taskID, cycle.ID, entries); err != nil {
		return ExecuteCommitIngestOutcome{}, err
	}
	if publish != nil {
		publish(taskID, cycle.ID)
	}
	return ExecuteCommitIngestOutcome{CommitCount: len(entries)}, nil
}

// PriorCycleBaseSHA reads cycle_base_sha from the earliest prior execute phase.
//
//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func (s *Service) PriorCycleBaseSHA(ctx context.Context, cycleID string, currentPhaseSeq int64) (string, error) {
	phases, err := s.store.ListPhasesForCycle(ctx, cycleID)
	if err != nil {
		return "", err
	}
	var first *cyclesdomain.TaskCyclePhase
	for i := range phases {
		p := &phases[i]
		if p.Phase != cyclesdomain.PhaseExecute || p.PhaseSeq >= currentPhaseSeq {
			continue
		}
		if first == nil || p.PhaseSeq < first.PhaseSeq {
			first = p
		}
	}
	if first == nil {
		return "", nil
	}
	return CycleBaseFromPhaseDetails(first.DetailsJSON), nil
}
