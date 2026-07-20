package policy_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/handler/policy"
)

func TestBootstrapLimits_matchSPAContract(t *testing.T) {
	tests := []struct {
		name  string
		got   int
		want  int
		label string
	}{
		{"list", policy.BootstrapListLimit, 20, "home task list"},
		{"projects", policy.BootstrapProjectsLimit, 100, "AppShell projects"},
		{"drafts", policy.BootstrapDraftsLimit, 50, "task create drafts"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Fatalf("%s: got %d want %d", tt.label, tt.got, tt.want)
			}
		})
	}
}

func TestShellChecklistIncluded(t *testing.T) {
	if !policy.ShellChecklistIncluded {
		t.Fatal("shell route must embed checklist items when shipped")
	}
}

func TestListAndPagingLimits(t *testing.T) {
	tests := []struct {
		name string
		got  int
		want int
	}{
		{"task_list_default", policy.TaskListDefaultLimit, 50},
		{"task_list_max", policy.TaskListMaxLimit, 200},
		{"task_events_default", policy.TaskEventsDefaultLimit, 50},
		{"task_events_max", policy.TaskEventsMaxLimit, 200},
		{"cycle_list_default", policy.CycleListDefaultLimit, 50},
		{"cycle_list_max", policy.CycleListMaxLimit, 200},
		{"cycle_stream_default", policy.CycleStreamDefaultLimit, 100},
		{"cycle_stream_max", policy.CycleStreamMaxLimit, 500},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Fatalf("%s: got %d want %d", tt.name, tt.got, tt.want)
			}
		})
	}
}

// TestReadLimitsContractJSON locks Go constants to testdata/readlimits.json
// (shared with web/src/lib/readLimits.contract.test.ts).
func TestReadLimitsContractJSON(t *testing.T) {
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	path := filepath.Join(filepath.Dir(thisFile), "..", "..", "..", "..", "testdata", "readlimits.json")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read contract: %v", err)
	}
	var contract map[string]int
	if err := json.Unmarshal(raw, &contract); err != nil {
		t.Fatalf("parse contract: %v", err)
	}
	want := map[string]int{
		"bootstrapListLimit":      policy.BootstrapListLimit,
		"bootstrapProjectsLimit":  policy.BootstrapProjectsLimit,
		"bootstrapDraftsLimit":    policy.BootstrapDraftsLimit,
		"taskListDefaultLimit":    policy.TaskListDefaultLimit,
		"taskListMaxLimit":        policy.TaskListMaxLimit,
		"taskEventsDefaultLimit":  policy.TaskEventsDefaultLimit,
		"taskEventsMaxLimit":      policy.TaskEventsMaxLimit,
		"cycleListDefaultLimit":   policy.CycleListDefaultLimit,
		"cycleListMaxLimit":       policy.CycleListMaxLimit,
		"cycleStreamDefaultLimit": policy.CycleStreamDefaultLimit,
		"cycleStreamMaxLimit":     policy.CycleStreamMaxLimit,
	}
	for key, got := range want {
		if contract[key] != got {
			t.Fatalf("%s: contract %d != policy %d", key, contract[key], got)
		}
	}
}
