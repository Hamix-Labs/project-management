package policy_test

import (
	"testing"

	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/handler/policy"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/realtime"
)

func TestIsScopelessHint(t *testing.T) {
	scopelessSet := map[realtime.ChangeType]bool{
		realtime.SettingsChanged:   true,
		realtime.AgentRunCancelled: true,
	}

	all := []realtime.ChangeType{
		realtime.TaskCreated,
		realtime.TaskUpdated,
		realtime.TaskDeleted,
		realtime.TaskGateChanged,
		realtime.TaskDependencyChanged,
		realtime.TaskCycleChanged,
		realtime.AgentRunProgress,
		realtime.ProjectCreated,
		realtime.ProjectUpdated,
		realtime.ProjectDeleted,
		realtime.SettingsChanged,
		realtime.AgentRunCancelled,
		realtime.Resync,
	}

	for _, typ := range all {
		t.Run(string(typ), func(t *testing.T) {
			want := scopelessSet[typ]
			if got := policy.IsScopelessHint(typ); got != want {
				t.Fatalf("IsScopelessHint(%q) = %v, want %v", typ, got, want)
			}
		})
	}
}

func TestScopelessHintChangeTypes_matchesIsScopelessHint(t *testing.T) {
	for _, typ := range policy.ScopelessHintChangeTypes {
		if !policy.IsScopelessHint(typ) {
			t.Fatalf("ScopelessHintChangeTypes includes %q but IsScopelessHint is false", typ)
		}
	}
}
