package git

import "github.com/AlexsanderHamir/Hamix/pkgs/obs/calltrace"
import (
	"context"
	"encoding/json"
	"log/slog"
	"strings"
	"time"
)

const gitSnapshotProbeTimeout = 30 * time.Second

// PhaseSnapshot captures repo/worktree/HEAD anchors at execute start.
type PhaseSnapshot struct {
	Skipped      bool
	Repo         string
	Worktree     string
	BaseSHA      string
	CycleBaseSHA string
	BaseBranch   string
	CapturedAt   time.Time
}

// CaptureExecuteGitSnapshot records git anchors for this execute phase.
// When workdir is not a git repo, Skipped is true and ingest/gates are bypassed.
func CaptureExecuteGitSnapshot(ctx context.Context, repo GitRepo, repoRoot, workdir, priorCycleBase string) (PhaseSnapshot, error) {
	if repo == nil {
		repo = DefaultRepo()
	}
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "agent.harness.git.CaptureExecuteGitSnapshot",
		"repo_root", repoRoot, "workdir", workdir)
	workdir = strings.TrimSpace(workdir)
	if workdir == "" {
		return PhaseSnapshot{Skipped: true}, nil
	}
	probeCtx, cancel := context.WithTimeout(ctx, gitSnapshotProbeTimeout)
	defer cancel()

	head, err := repo.Run(probeCtx, workdir, "rev-parse", "HEAD")
	if err != nil {
		if IsNotAGitRepoErr(err) {
			return PhaseSnapshot{Skipped: true}, nil
		}
		return PhaseSnapshot{}, err
	}
	worktree, err := repo.Run(probeCtx, workdir, "rev-parse", "--show-toplevel")
	if err != nil {
		if IsNotAGitRepoErr(err) {
			return PhaseSnapshot{Skipped: true}, nil
		}
		return PhaseSnapshot{}, err
	}
	branch, err := repo.Run(probeCtx, workdir, "branch", "--show-current")
	if err != nil && !IsNotAGitRepoErr(err) {
		return PhaseSnapshot{}, err
	}
	cycleBase := strings.TrimSpace(priorCycleBase)
	if cycleBase == "" {
		cycleBase = head
	}
	return PhaseSnapshot{
		Repo:         strings.TrimSpace(repoRoot),
		Worktree:     worktree,
		BaseSHA:      head,
		CycleBaseSHA: cycleBase,
		BaseBranch:   strings.TrimSpace(branch),
		CapturedAt:   time.Now().UTC(),
	}, nil
}

// CycleBaseFromPhaseDetails reads cycle_base_sha or base_sha from execute phase details.
//
//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func CycleBaseFromPhaseDetails(details []byte) string {
	if len(details) == 0 {
		return ""
	}
	var root map[string]json.RawMessage
	if err := json.Unmarshal(details, &root); err != nil {
		return ""
	}
	raw, ok := root["git"]
	if !ok {
		return ""
	}
	var gitMeta struct {
		CycleBaseSHA string `json:"cycle_base_sha"`
		BaseSHA      string `json:"base_sha"`
	}
	if err := json.Unmarshal(raw, &gitMeta); err != nil {
		return ""
	}
	if v := strings.TrimSpace(gitMeta.CycleBaseSHA); v != "" {
		return v
	}
	return strings.TrimSpace(gitMeta.BaseSHA)
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func snapshotToMap(s PhaseSnapshot, commitCount int) map[string]any {
	m := map[string]any{
		"repo":           s.Repo,
		"worktree":       s.Worktree,
		"base_sha":       s.BaseSHA,
		"cycle_base_sha": s.CycleBaseSHA,
		"base_branch":    s.BaseBranch,
		"captured_at":    s.CapturedAt.Format(time.RFC3339),
	}
	if commitCount > 0 {
		m["commit_count"] = commitCount
	}
	if s.Skipped {
		m["skipped"] = true
	}
	return m
}

// CriteriaReportProbeErrFromPhaseDetails reads criteria_report_probe_err from execute phase details.
//
//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func CriteriaReportProbeErrFromPhaseDetails(details []byte) string {
	if len(details) == 0 {
		return ""
	}
	var root map[string]json.RawMessage
	if err := json.Unmarshal(details, &root); err != nil {
		return ""
	}
	raw, ok := root["criteria_report_probe_err"]
	if !ok {
		return ""
	}
	var s string
	if err := json.Unmarshal(raw, &s); err != nil {
		return ""
	}
	return strings.TrimSpace(s)
}

// MergeCriteriaReportProbeErr attaches a criteria-report parse error for cross-cycle resume.
//
//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func MergeCriteriaReportProbeErr(baseDetails []byte, probeErr string) []byte {
	probeErr = strings.TrimSpace(probeErr)
	if probeErr == "" {
		return baseDetails
	}
	root := map[string]any{}
	if len(baseDetails) > 0 {
		if err := json.Unmarshal(baseDetails, &root); err != nil {
			slog.Warn("MergeCriteriaReportProbeErr: keeping base details; unmarshal failed",
				"operation", "agent.harness.git.MergeCriteriaReportProbeErr", "err", err)
			return baseDetails
		}
	}
	root["criteria_report_probe_err"] = probeErr
	out, err := json.Marshal(root)
	if err != nil {
		return baseDetails
	}
	return out
}

// MergeRunnerDetailsWithGit attaches git snapshot metadata to execute phase details.
//
//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func MergeRunnerDetailsWithGit(baseDetails []byte, snap PhaseSnapshot, commitCount int) []byte {
	if snap.Skipped && commitCount == 0 {
		if len(baseDetails) == 0 {
			return baseDetails
		}
		return baseDetails
	}
	root := map[string]any{}
	if len(baseDetails) > 0 {
		if err := json.Unmarshal(baseDetails, &root); err != nil {
			slog.Warn("MergeRunnerDetailsWithGit: keeping base details; unmarshal failed",
				"operation", "agent.harness.git.MergeRunnerDetailsWithGit", "err", err)
			return baseDetails
		}
	}
	root["git"] = snapshotToMap(snap, commitCount)
	out, err := json.Marshal(root)
	if err != nil {
		return baseDetails
	}
	return out
}
