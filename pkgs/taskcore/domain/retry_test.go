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
		{name: "polish empty instructions", in: PendingRetry{Kind: PendingKindPolish, Mode: RetryResume, ParentCycleID: "c", Instructions: "  "}, wantErr: true},
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
		})
	}
}
