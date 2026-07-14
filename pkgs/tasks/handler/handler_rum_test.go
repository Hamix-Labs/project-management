package handler

import (
	"testing"
)

func TestFoldRUMEvent_validatesDurationAndStatus(t *testing.T) {
	cases := []struct {
		name string
		ev   rumEvent
		ok   bool
	}{
		{"mutation_started accepted", rumEvent{Type: "mutation_started", MutationKind: "task_patch"}, true},
		{"optimistic with neg duration dropped", rumEvent{Type: "mutation_optimistic_applied", MutationKind: "task_patch", DurationSeconds: -0.1}, false},
		{"settled with huge duration dropped", rumEvent{Type: "mutation_settled", MutationKind: "task_patch", DurationSeconds: 99999, StatusCode: 200}, false},
		{"settled with 5xx accepted (errors are observable)", rumEvent{Type: "mutation_settled", MutationKind: "task_patch", DurationSeconds: 0.1, StatusCode: 503}, true},
		{"reconnect with zero duration accepted", rumEvent{Type: "sse_reconnected", DurationSeconds: 0}, true},
		{"reconnect with negative duration dropped", rumEvent{Type: "sse_reconnected", DurationSeconds: -1}, false},
		{"web_vitals unknown name dropped", rumEvent{Type: "web_vitals", Name: "BOGUS", Value: 1}, false},
		{"web_vitals known name accepted", rumEvent{Type: "web_vitals", Name: "INP", Value: 50}, true},
		{"unknown type dropped", rumEvent{Type: "garbage"}, false},
	}
	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			if got := foldRUMEvent(c.ev); got != c.ok {
				t.Fatalf("foldRUMEvent=%v want %v", got, c.ok)
			}
		})
	}
}
