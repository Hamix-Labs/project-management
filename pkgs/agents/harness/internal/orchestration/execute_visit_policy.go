package orchestration

import (
	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
)

// CommitIngestMode selects post-execute MCP commit-register validation.
type CommitIngestMode int

const (
	// CommitIngestRequireRegistered requires a non-empty register (ADR-0093 default).
	CommitIngestRequireRegistered CommitIngestMode = iota
	// CommitIngestAllowEmptyWhenNoHeadDelta allows an empty register when HEAD
	// has no commits beyond cycle_base (open-pr / instructions-only polish).
	CommitIngestAllowEmptyWhenNoHeadDelta
)

// PostExecutePath selects what the cycle loop does after a successful execute.
type PostExecutePath int

const (
	// PostExecuteClaimAcceptance runs claim acceptance then finalize to review.
	PostExecuteClaimAcceptance PostExecutePath = iota
	// PostExecuteReviewSkipClaims skips claim acceptance and finalizes to review.
	PostExecuteReviewSkipClaims
	// PostExecuteOpenPR skips claim acceptance and finalizes to pr_ready.
	PostExecuteOpenPR
)

// ExecuteVisitPolicy is the run-kind policy for one execute visit.
type ExecuteVisitPolicy struct {
	CommitIngest CommitIngestMode
	PostExecute  PostExecutePath
}

// ResolveExecuteVisitPolicy maps pending-run kind + skip-claim flag to visit policy.
//
//funclogmeasure:skip category=hot-path reason="Pure policy table; callers emit operation traces."
func ResolveExecuteVisitPolicy(runKind taskcoredomain.PendingRunKind, skipClaimAcceptance bool) ExecuteVisitPolicy {
	switch runKind {
	case taskcoredomain.PendingKindOpenPR:
		return ExecuteVisitPolicy{
			CommitIngest: CommitIngestAllowEmptyWhenNoHeadDelta,
			PostExecute:  PostExecuteOpenPR,
		}
	case taskcoredomain.PendingKindPolish:
		if skipClaimAcceptance {
			return ExecuteVisitPolicy{
				CommitIngest: CommitIngestAllowEmptyWhenNoHeadDelta,
				PostExecute:  PostExecuteReviewSkipClaims,
			}
		}
	}
	return ExecuteVisitPolicy{
		CommitIngest: CommitIngestRequireRegistered,
		PostExecute:  PostExecuteClaimAcceptance,
	}
}
