package policy_test

import (
	"testing"

	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/handler/policy"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/realtime"
)

func TestEnrichedTaskChangeEvent(t *testing.T) {
	tests := []struct {
		typ  realtime.ChangeType
		want bool
	}{
		{realtime.TaskCreated, true},
		{realtime.TaskUpdated, true},
		{realtime.TaskDeleted, false},
		{realtime.TaskGateChanged, false},
		{realtime.TaskCycleChanged, false},
		{realtime.AgentRunProgress, false},
		{realtime.Resync, false},
	}
	for _, tt := range tests {
		t.Run(string(tt.typ), func(t *testing.T) {
			if got := policy.EnrichedTaskChangeEvent(tt.typ); got != tt.want {
				t.Fatalf("EnrichedTaskChangeEvent(%q) = %v, want %v", tt.typ, got, tt.want)
			}
		})
	}
}

func TestIsHintOnly(t *testing.T) {
	hintSet := map[realtime.ChangeType]bool{
		realtime.TaskDeleted:           true,
		realtime.TaskGateChanged:       true,
		realtime.TaskDependencyChanged: true,
		realtime.TaskEventChanged:      true,
		realtime.ProjectCreated:        true,
		realtime.ProjectUpdated:        true,
		realtime.ProjectDeleted:        true,
		realtime.ProjectContextChanged: true,
		realtime.SettingsChanged:       true,
	}

	all := []realtime.ChangeType{
		realtime.TaskCreated,
		realtime.TaskUpdated,
		realtime.TaskDeleted,
		realtime.TaskGateChanged,
		realtime.TaskDependencyChanged,
		realtime.TaskEventChanged,
		realtime.TaskCycleChanged,
		realtime.AgentRunProgress,
		realtime.ProjectCreated,
		realtime.ProjectUpdated,
		realtime.ProjectDeleted,
		realtime.ProjectContextChanged,
		realtime.SettingsChanged,
		realtime.AgentRunCancelled,
		realtime.Resync,
	}

	for _, typ := range all {
		t.Run(string(typ), func(t *testing.T) {
			want := hintSet[typ]
			if got := policy.IsHintOnly(typ); got != want {
				t.Fatalf("IsHintOnly(%q) = %v, want %v", typ, got, want)
			}
		})
	}
}

func TestHintOnlyChangeTypes_matchesIsHintOnly(t *testing.T) {
	for _, typ := range policy.HintOnlyChangeTypes {
		if !policy.IsHintOnly(typ) {
			t.Fatalf("HintOnlyChangeTypes includes %q but IsHintOnly is false", typ)
		}
	}
}
