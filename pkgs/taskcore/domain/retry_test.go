package domain

import (
	"errors"
	"testing"
)

func TestPendingRetry_Validate(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		in      PendingRetry
		wantErr bool
	}{
		{name: "fresh", in: PendingRetry{Mode: RetryFresh, ParentCycleID: "cycle-1"}},
		{name: "resume", in: PendingRetry{Mode: RetryResume, ParentCycleID: "cycle-1"}},
		{name: "trim parent", in: PendingRetry{Mode: RetryFresh, ParentCycleID: "  abc  "}},
		{name: "polish", in: PendingRetry{Kind: PendingKindPolish, Mode: RetryResume, ParentCycleID: "cycle-1", Instructions: "fix spacing"}},
		{name: "polish with flags", in: PendingRetry{Kind: PendingKindPolish, Mode: RetryResume, ParentCycleID: "cycle-1", Instructions: "fix", FlaggedCriterionIDs: []string{"c1", " c1 ", "c2"}}},
		{name: "polish with new", in: PendingRetry{Kind: PendingKindPolish, Mode: RetryResume, ParentCycleID: "cycle-1", Instructions: "add docs", NewCriterionIDs: []string{"n1", " n1 "}}},
		{name: "polish empty instructions", in: PendingRetry{Kind: PendingKindPolish, Mode: RetryResume, ParentCycleID: "c", Instructions: "  "}, wantErr: true},
		{name: "polish empty instructions with flags", in: PendingRetry{Kind: PendingKindPolish, Mode: RetryResume, ParentCycleID: "c", Instructions: "", FlaggedCriterionIDs: []string{"c1"}}, wantErr: true},
		{name: "polish empty instructions with new", in: PendingRetry{Kind: PendingKindPolish, Mode: RetryResume, ParentCycleID: "c", Instructions: "", NewCriterionIDs: []string{"n1"}}, wantErr: true},
		{name: "polish fresh mode", in: PendingRetry{Kind: PendingKindPolish, Mode: RetryFresh, ParentCycleID: "c", Instructions: "x"}, wantErr: true},
		{name: "bad mode", in: PendingRetry{Mode: "nope", ParentCycleID: "c"}, wantErr: true},
		{name: "empty parent", in: PendingRetry{Mode: RetryFresh, ParentCycleID: "  "}, wantErr: true},
		{name: "bad kind", in: PendingRetry{Kind: "other", Mode: RetryResume, ParentCycleID: "c"}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := tt.in.Validate()
			if tt.wantErr {
				if !errors.Is(err, ErrInvalidInput) {
					t.Fatalf("got %v want ErrInvalidInput", err)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if tt.name == "trim parent" && tt.in.ParentCycleID != "abc" {
				t.Fatalf("parent %q want abc", tt.in.ParentCycleID)
			}
			if tt.name == "polish with flags" {
				if tt.in.SkipVerify {
					t.Fatal("expected SkipVerify false when flags present")
				}
				if len(tt.in.FlaggedCriterionIDs) != 2 || tt.in.FlaggedCriterionIDs[0] != "c1" || tt.in.FlaggedCriterionIDs[1] != "c2" {
					t.Fatalf("flagged = %#v", tt.in.FlaggedCriterionIDs)
				}
			}
			if tt.name == "polish with new" {
				if tt.in.SkipVerify {
					t.Fatal("expected SkipVerify false when new criteria present")
				}
				if len(tt.in.NewCriterionIDs) != 1 || tt.in.NewCriterionIDs[0] != "n1" {
					t.Fatalf("new = %#v", tt.in.NewCriterionIDs)
				}
			}
			if tt.name == "polish" && !tt.in.SkipVerify {
				t.Fatal("expected SkipVerify true for instructions-only polish")
			}
		})
	}
}
