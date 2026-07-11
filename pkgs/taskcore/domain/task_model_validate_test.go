package domain

import (
	"errors"
	"strings"
	"testing"
)

func TestValidateTaskTag(t *testing.T) {
	t.Parallel()
	tests := []struct {
		tag string
		ok  bool
	}{
		{"design", true},
		{"launch-v1", true},
		{"a.b_c", true},
		{"", false},
		{"Design", false},
		{strings.Repeat("a", 33), false},
	}
	for _, tt := range tests {
		t.Run(tt.tag, func(t *testing.T) {
			t.Parallel()
			err := ValidateTaskTag(tt.tag)
			if tt.ok && err != nil {
				t.Fatalf("ValidateTaskTag(%q) = %v, want nil", tt.tag, err)
			}
			if !tt.ok && !errors.Is(err, ErrInvalidInput) {
				t.Fatalf("ValidateTaskTag(%q) = %v, want ErrInvalidInput", tt.tag, err)
			}
		})
	}
}

func TestValidateTaskMilestone(t *testing.T) {
	t.Parallel()
	if err := ValidateTaskMilestone(""); err != nil {
		t.Fatalf("empty: %v", err)
	}
	if err := ValidateTaskMilestone("Launch v1"); err != nil {
		t.Fatalf("valid: %v", err)
	}
	if err := ValidateTaskMilestone("-bad"); !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("invalid prefix: %v", err)
	}
}

func TestTaskGate_GateBlocksWorker(t *testing.T) {
	t.Parallel()
	if (&TaskGate{Status: GateStatusReleased}).GateBlocksWorker() {
		t.Fatal("released should not block")
	}
	if !(&TaskGate{Status: GateStatusActive}).GateBlocksWorker() {
		t.Fatal("active should block")
	}
	var nilGate *TaskGate
	if nilGate.GateBlocksWorker() {
		t.Fatal("nil gate should not block")
	}
}

func TestNormalizeTaskTags_dedupes(t *testing.T) {
	t.Parallel()
	got := NormalizeTaskTags([]string{"a", "a", " b ", ""})
	if len(got) != 2 || got[0] != "a" || got[1] != "b" {
		t.Fatalf("got %v", got)
	}
}
