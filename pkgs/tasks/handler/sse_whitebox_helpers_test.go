package handler

import (
	"encoding/json"
	"fmt"
	"sort"
	"testing"
	"time"

	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/realtime"
)

// Minimal whitebox SSE drain helpers (cannot import internal/handlertest — import cycle).

func drainSSE(t *testing.T, ch <-chan string, want int, timeout time.Duration) []realtime.Event {
	t.Helper()
	out := make([]realtime.Event, 0, want)
	deadline := time.After(timeout)
	for len(out) < want {
		select {
		case s, ok := <-ch:
			if !ok {
				return out
			}
			var ev realtime.Event
			if err := json.Unmarshal([]byte(s), &ev); err != nil {
				t.Fatalf("decode sse line %q: %v", s, err)
			}
			out = append(out, ev)
		case <-deadline:
			return out
		}
	}
	select {
	case s := <-ch:
		var ev realtime.Event
		if err := json.Unmarshal([]byte(s), &ev); err == nil {
			out = append(out, ev)
		}
	case <-time.After(50 * time.Millisecond):
	}
	return out
}

func summarize(events []realtime.Event) []string {
	out := make([]string, 0, len(events))
	for _, ev := range events {
		if ev.EventSeq > 0 {
			out = append(out, fmt.Sprintf("%s:%s/%d", ev.Type, ev.ID, ev.EventSeq))
			continue
		}
		if ev.CycleID != "" {
			out = append(out, fmt.Sprintf("%s:%s/%s", ev.Type, ev.ID, ev.CycleID))
			continue
		}
		out = append(out, fmt.Sprintf("%s:%s", ev.Type, ev.ID))
	}
	sort.Strings(out)
	return out
}

func mustEqualEvents(t *testing.T, route string, got, want []string) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("%s: got %d events %v, want %d %v", route, len(got), got, len(want), want)
	}
	for i := range got {
		if got[i] != want[i] {
			t.Fatalf("%s: event[%d]=%q want %q (full got=%v want=%v)", route, i, got[i], want[i], got, want)
		}
	}
}

func mustHaveTaskUpdatedData(t *testing.T, route string, events []realtime.Event, taskID string) {
	t.Helper()
	for _, ev := range events {
		if ev.Type == realtime.TaskUpdated && ev.ID == taskID {
			if ev.Data == nil {
				t.Fatalf("%s: task_updated:%s missing data enrichment", route, taskID)
			}
			return
		}
	}
	t.Fatalf("%s: no task_updated:%s in events %v", route, taskID, summarize(events))
}
