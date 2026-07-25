package policy_test

import (
	"context"
	settingsdomain "github.com/AlexsanderHamir/Hamix/pkgs/settings/domain"
	"testing"

	"github.com/AlexsanderHamir/Hamix/internal/taskapi/agentworker/policy"
)

func TestDecideIdle(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	noRepos := func(context.Context) (bool, string, error) {
		return true, "no_repository_registered", nil
	}
	allInvalid := func(context.Context) (bool, string, error) {
		return true, "all_worktrees_invalid", nil
	}
	okGit := func(context.Context) (bool, string, error) {
		return false, "", nil
	}

	tests := []struct {
		name   string
		cfg    settingsdomain.AppSettings
		check  policy.GitRegistrationChecker
		idle   bool
		reason string
	}{
		{
			name:   "paused",
			cfg:    settingsdomain.AppSettings{AgentPaused: true},
			check:  okGit,
			idle:   true,
			reason: "paused_by_operator",
		},
		{
			name:   "no repositories",
			cfg:    settingsdomain.AppSettings{},
			check:  noRepos,
			idle:   true,
			reason: "no_repository_registered",
		},
		{
			name:   "all worktrees invalid",
			cfg:    settingsdomain.AppSettings{},
			check:  allInvalid,
			idle:   true,
			reason: "all_worktrees_invalid",
		},
		{
			name:   "configured",
			cfg:    settingsdomain.AppSettings{Runner: "cursor"},
			check:  okGit,
			idle:   false,
			reason: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			idle, reason := policy.DecideIdle(ctx, tt.cfg, tt.check)
			if idle != tt.idle || reason != tt.reason {
				t.Fatalf("DecideIdle() = (%v, %q), want (%v, %q)", idle, reason, tt.idle, tt.reason)
			}
		})
	}
}

func TestDecideSchedulingIdleHint(t *testing.T) {
	t.Parallel()
	if got := policy.DecideSchedulingIdleHint(true, 3); got != policy.SchedulingIdleHintReason {
		t.Fatalf("got %q", got)
	}
	if got := policy.DecideSchedulingIdleHint(false, 3); got != "" {
		t.Fatalf("got %q", got)
	}
	if got := policy.DecideSchedulingIdleHint(true, 0); got != "" {
		t.Fatalf("got %q", got)
	}
}

func TestInstanceMatchesSettings(t *testing.T) {
	t.Parallel()
	base := settingsdomain.AppSettings{
		Runner:                "cursor",
		CursorBin:             "/bin/cursor",
		CursorModel:           "gpt",
		MaxRunDurationSeconds: 600,
	}
	inst := &policy.InstanceSnapshot{
		Settings:      base,
		RunnerVersion: "1.0",
	}
	if !policy.InstanceMatchesSettings(inst, base, "1.0") {
		t.Fatal("expected match for identical settings")
	}
	changed := base
	changed.CursorModel = "other"
	if policy.InstanceMatchesSettings(inst, changed, "1.0") {
		t.Fatal("expected mismatch on cursor model")
	}
	if policy.InstanceMatchesSettings(inst, base, "2.0") {
		t.Fatal("expected mismatch on runner version")
	}
}
