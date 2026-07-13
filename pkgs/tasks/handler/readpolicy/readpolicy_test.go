package readpolicy_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/handler/readpolicy"
)

func TestBootstrapLimits_matchSPAContract(t *testing.T) {
	tests := []struct {
		name  string
		got   int
		want  int
		label string
	}{
		{"list", readpolicy.BootstrapListLimit, 20, "home task list"},
		{"projects", readpolicy.BootstrapProjectsLimit, 100, "AppShell projects"},
		{"drafts", readpolicy.BootstrapDraftsLimit, 50, "task create drafts"},
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
	if !readpolicy.ShellChecklistIncluded {
		t.Fatal("shell route must embed checklist items when shipped")
	}
}

func TestListAndPagingLimits(t *testing.T) {
	tests := []struct {
		name string
		got  int
		want int
	}{
		{"task_list_default", readpolicy.TaskListDefaultLimit, 50},
		{"task_list_max", readpolicy.TaskListMaxLimit, 200},
		{"task_events_default", readpolicy.TaskEventsDefaultLimit, 50},
		{"task_events_max", readpolicy.TaskEventsMaxLimit, 200},
		{"cycle_list_default", readpolicy.CycleListDefaultLimit, 50},
		{"cycle_list_max", readpolicy.CycleListMaxLimit, 200},
		{"cycle_stream_default", readpolicy.CycleStreamDefaultLimit, 100},
		{"cycle_stream_max", readpolicy.CycleStreamMaxLimit, 500},
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
		"bootstrapListLimit":      readpolicy.BootstrapListLimit,
		"bootstrapProjectsLimit":  readpolicy.BootstrapProjectsLimit,
		"bootstrapDraftsLimit":    readpolicy.BootstrapDraftsLimit,
		"taskListDefaultLimit":    readpolicy.TaskListDefaultLimit,
		"taskListMaxLimit":        readpolicy.TaskListMaxLimit,
		"taskEventsDefaultLimit":  readpolicy.TaskEventsDefaultLimit,
		"taskEventsMaxLimit":      readpolicy.TaskEventsMaxLimit,
		"cycleListDefaultLimit":   readpolicy.CycleListDefaultLimit,
		"cycleListMaxLimit":       readpolicy.CycleListMaxLimit,
		"cycleStreamDefaultLimit": readpolicy.CycleStreamDefaultLimit,
		"cycleStreamMaxLimit":     readpolicy.CycleStreamMaxLimit,
	}
	for key, got := range want {
		if contract[key] != got {
			t.Fatalf("%s: contract %d != readpolicy %d", key, contract[key], got)
		}
	}
}
