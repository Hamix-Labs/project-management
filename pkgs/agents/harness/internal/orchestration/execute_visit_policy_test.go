package orchestration

import (
	"testing"

	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
)

func TestResolveExecuteVisitPolicy(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name       string
		runKind    taskcoredomain.PendingRunKind
		skipClaim  bool
		wantIngest CommitIngestMode
		wantPath   PostExecutePath
	}{
		{
			name:       "default_retry",
			runKind:    taskcoredomain.PendingKindRetry,
			wantIngest: CommitIngestRequireRegistered,
			wantPath:   PostExecuteClaimAcceptance,
		},
		{
			name:       "empty_kind",
			wantIngest: CommitIngestRequireRegistered,
			wantPath:   PostExecuteClaimAcceptance,
		},
		{
			name:       "open_pr",
			runKind:    taskcoredomain.PendingKindOpenPR,
			wantIngest: CommitIngestAllowEmptyWhenNoHeadDelta,
			wantPath:   PostExecuteOpenPR,
		},
		{
			name:       "open_pr_ignores_skip_claim_flag",
			runKind:    taskcoredomain.PendingKindOpenPR,
			skipClaim:  false,
			wantIngest: CommitIngestAllowEmptyWhenNoHeadDelta,
			wantPath:   PostExecuteOpenPR,
		},
		{
			name:       "polish_skip_claim",
			runKind:    taskcoredomain.PendingKindPolish,
			skipClaim:  true,
			wantIngest: CommitIngestAllowEmptyWhenNoHeadDelta,
			wantPath:   PostExecuteReviewSkipClaims,
		},
		{
			name:       "polish_with_flags_requires_commits",
			runKind:    taskcoredomain.PendingKindPolish,
			skipClaim:  false,
			wantIngest: CommitIngestRequireRegistered,
			wantPath:   PostExecuteClaimAcceptance,
		},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := ResolveExecuteVisitPolicy(tc.runKind, tc.skipClaim)
			if got.CommitIngest != tc.wantIngest || got.PostExecute != tc.wantPath {
				t.Fatalf("got %+v want ingest=%v path=%v", got, tc.wantIngest, tc.wantPath)
			}
		})
	}
}
