package model

import (
	"encoding/json"
	"reflect"
	"testing"
	"time"

	"github.com/AlexsanderHamir/Hamix/pkgs/settings/domain"
)

func TestAppSettings_roundTrip(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 3, 1, 12, 0, 0, 0, time.UTC)
	cfg := json.RawMessage(`{"cursor":{"binary_path":"/bin/cursor"}}`)
	orig := domain.AppSettings{
		ID:                         domain.AppSettingsRowID,
		AgentPaused:                true,
		Runner:                     "cursor",
		CursorBin:                  "/bin/cursor",
		CursorModel:                "opus",
		VerifyModel:                "composer-2.5-fast",
		MaxRunDurationSeconds:      120,
		AgentTaskParallelism:       4,
		AgentPickupDelaySeconds:    3,
		DisplayTimezone:            "America/Los_Angeles",
		OptimisticMutationsEnabled: true,
		SSEReplayEnabled:           true,
		RunnerConfigs:              cfg,
		CursorSessionResumeEnabled: false,
		AgentMCPEnabled:            true,
		UpdatedAt:                  now,
	}
	m := FromDomainAppSettings(orig)
	back := ToDomainAppSettings(m)
	if !reflect.DeepEqual(orig, back) {
		t.Fatalf("round-trip mismatch:\norig=%+v\nback=%+v", orig, back)
	}
	m2 := FromDomainAppSettings(back)
	if !appSettingsModelEqual(m, m2) {
		t.Fatalf("model round-trip mismatch")
	}
}

func appSettingsModelEqual(a, b AppSettings) bool {
	return reflect.DeepEqual(a, b)
}

func TestAppSettings_emptyRunnerConfigs(t *testing.T) {
	t.Parallel()
	orig := domain.DefaultAppSettings()
	m := FromDomainAppSettings(orig)
	m2 := FromDomainAppSettings(ToDomainAppSettings(m))
	if !reflect.DeepEqual(m, m2) {
		t.Fatal("empty runner configs round-trip failed")
	}
}
