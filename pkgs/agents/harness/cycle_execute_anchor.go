package harness

import "github.com/AlexsanderHamir/Hamix/pkgs/obs/calltrace"
import (
	"context"
	"log/slog"
	"strings"

	"github.com/AlexsanderHamir/Hamix/pkgs/agents/harness/internal/git"
)

func (h *Harness) anchorPostExecuteState(
	ctx context.Context,
	state *processState,
	execPhaseSeq int64,
	snap git.PhaseSnapshot,
	ingestAttempted bool,
	ingestOutcome executeCommitIngestOutcome,
	ingestErr error,
) {
	state.phase.executeReachedClaimAcceptance = true
	state.git.lastCommitIngestOK = commitIngestOK(snap, ingestAttempted, ingestOutcome, ingestErr)
	head, ok, err := h.resolveCurrentHeadSHA(ctx, snap)
	if err != nil {
		slog.Warn("agent harness post-execute head anchor failed", "cmd", calltrace.LogCmd,
			"operation", "agent.harness.Harness.anchorPostExecuteState.head",
			"cycle_id", state.cycle.cycleID, "err", err)
		return
	}
	if ok {
		state.git.postExecuteHeadSHA = head
	}
	state.phase.lastCompletedExecutePhaseSeq = execPhaseSeq
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func commitIngestOK(
	snap git.PhaseSnapshot,
	ingestAttempted bool,
	ingestOutcome executeCommitIngestOutcome,
	ingestErr error,
) bool {
	if snap.Skipped {
		return true
	}
	if !ingestAttempted {
		return true
	}
	if ingestErr != nil {
		return false
	}
	return ingestOutcome.FailReason == ""
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func (h *Harness) resolveCurrentHeadSHA(ctx context.Context, snap git.PhaseSnapshot) (head string, ok bool, err error) {
	if snap.Skipped {
		return "", false, nil
	}
	workdir := strings.TrimSpace(h.opts.WorkingDir)
	if workdir == "" {
		return "", false, nil
	}
	repo := h.gitSvc().Repo()
	if repo == nil {
		return "", false, nil
	}
	out, err := repo.Run(ctx, workdir, "rev-parse", "HEAD")
	if err != nil {
		if git.IsNotAGitRepoErr(err) {
			return "", false, nil
		}
		return "", false, err
	}
	return strings.TrimSpace(out), true, nil
}
