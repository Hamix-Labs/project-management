package cursorresume

import (
	"testing"

	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
	cyclesdomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcycles/domain"
)

func TestDecide_table(t *testing.T) {
	t.Parallel()
	base := Facts{
		SessionResumeEnabled: true,
		RetryMode:            taskcoredomain.RetryResume,
		Phase:                cyclesdomain.PhaseExecute,
		HeadMatchesAnchor:    true,
		SessionID:            "sess-1",
		WorkingDir:           "/tmp/ws",
	}
	tests := []struct {
		name      string
		mutate    func(*Facts)
		wantMode  Mode
		wantDeny  string
		wantAllow bool
	}{
		{
			name:      "continue when all gates pass",
			wantMode:  ModeContinue,
			wantAllow: true,
		},
		{
			name: "force fresh",
			mutate: func(f *Facts) {
				f.ForceFresh = true
			},
			wantMode: ModeFresh,
			wantDeny: "resume_failed",
		},
		{
			name: "settings disabled",
			mutate: func(f *Facts) {
				f.SessionResumeEnabled = false
			},
			wantMode: ModeFresh,
			wantDeny: "settings_disabled",
		},
		{
			name: "retry fresh",
			mutate: func(f *Facts) {
				f.RetryMode = taskcoredomain.RetryFresh
			},
			wantMode: ModeFresh,
			wantDeny: "retry_fresh",
		},
		{
			name: "verify continues when session id present",
			mutate: func(f *Facts) {
				f.Phase = cyclesdomain.PhaseVerify
			},
			wantMode:  ModeContinue,
			wantAllow: true,
		},
		{
			name: "verify fresh after execute",
			mutate: func(f *Facts) {
				f.Phase = cyclesdomain.PhaseVerify
				f.FirstVerifyAfterExecute = true
			},
			wantMode: ModeFresh,
			wantDeny: "verify_fresh_after_execute",
		},
		{
			name: "head drift",
			mutate: func(f *Facts) {
				f.HasPostExecuteHead = true
				f.HeadMatchesAnchor = false
			},
			wantMode: ModeFresh,
			wantDeny: "head_drift",
		},
		{
			name: "tamper",
			mutate: func(f *Facts) {
				f.ReportTampered = true
			},
			wantMode: ModeFresh,
			wantDeny: "tamper",
		},
		{
			name: "no session id",
			mutate: func(f *Facts) {
				f.SessionID = ""
			},
			wantMode: ModeFresh,
			wantDeny: "no_session_id",
		},
		{
			name: "workspace mismatch",
			mutate: func(f *Facts) {
				f.WorkingDir = "  "
			},
			wantMode: ModeFresh,
			wantDeny: "workspace_mismatch",
		},
		{
			name: "skip head check when git skipped",
			mutate: func(f *Facts) {
				f.GitSkipped = true
				f.HasPostExecuteHead = true
				f.HeadMatchesAnchor = false
			},
			wantMode:  ModeContinue,
			wantAllow: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			in := base
			if tt.mutate != nil {
				tt.mutate(&in)
			}
			got := Decide(in)
			if got.Mode != tt.wantMode || got.DenyReason != tt.wantDeny || got.AllowResume != tt.wantAllow {
				t.Fatalf("got %+v want mode=%s deny=%q allow=%v", got, tt.wantMode, tt.wantDeny, tt.wantAllow)
			}
		})
	}
}
