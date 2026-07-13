package runner_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/AlexsanderHamir/Hamix/pkgs/agents/runner"
	cyclesdomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcycles/domain"
)

// TestRequest_jsonShape pins the on-the-wire keys for Request. Adapters that
// serialise a Request must produce exactly these keys in any order.
func TestRequest_jsonShape(t *testing.T) {
	t.Parallel()

	req := runner.Request{
		TaskID:     "11111111-1111-4111-8111-111111111111",
		AttemptSeq: 3,
		Phase:      cyclesdomain.PhaseExecute,
		Prompt:     "do the thing",
		WorkingDir: "/repo",
		Timeout:    5 * time.Second,
		Env:        map[string]string{"PATH": "/usr/bin"},
	}

	raw, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var generic map[string]any
	if err := json.Unmarshal(raw, &generic); err != nil {
		t.Fatalf("unmarshal generic: %v", err)
	}

	wantKeys := []string{"task_id", "attempt_seq", "phase", "prompt", "working_dir", "timeout_ns", "env"}
	for _, k := range wantKeys {
		if _, ok := generic[k]; !ok {
			t.Errorf("missing JSON key %q in %s", k, raw)
		}
	}
	for k := range generic {
		if !contains(wantKeys, k) && k != "cursor_model" && k != "run_correlation_id" {
			t.Errorf("unexpected JSON key %q (full payload: %s)", k, raw)
		}
	}

	if got := generic["phase"].(string); got != "execute" {
		t.Errorf("phase wire value: got %q want %q", got, "execute")
	}
	if got := generic["timeout_ns"].(float64); got != float64(5*time.Second) {
		t.Errorf("timeout_ns: got %v want %v", got, float64(5*time.Second))
	}
}

// TestRequest_jsonRoundtrip checks Request survives a full round-trip with
// no field loss.
func TestRequest_jsonRoundtrip(t *testing.T) {
	t.Parallel()

	want := runner.Request{
		TaskID:      "22222222-2222-4222-8222-222222222222",
		AttemptSeq:  7,
		Phase:       cyclesdomain.PhaseExecute,
		Prompt:      "execute the change",
		WorkingDir:  "/work",
		Timeout:     250 * time.Millisecond,
		Env:         map[string]string{"PATH": "/bin", "HOME": "/home/runner"},
		CursorModel: "opus-4.1",
	}

	raw, err := json.Marshal(want)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var got runner.Request
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got.TaskID != want.TaskID || got.AttemptSeq != want.AttemptSeq ||
		got.Phase != want.Phase || got.Prompt != want.Prompt ||
		got.WorkingDir != want.WorkingDir || got.Timeout != want.Timeout ||
		got.CursorModel != want.CursorModel {
		t.Errorf("scalar mismatch: got %+v want %+v", got, want)
	}
	if len(got.Env) != len(want.Env) {
		t.Fatalf("env length: got %d want %d", len(got.Env), len(want.Env))
	}
	for k, v := range want.Env {
		if got.Env[k] != v {
			t.Errorf("env[%q]: got %q want %q", k, got.Env[k], v)
		}
	}
}

// TestResult_jsonShape_omitempty ensures optional fields drop out of the
// payload when zero. Audit-log consumers rely on this to keep cycle/phase
// rows compact.
func TestResult_jsonShape_omitempty(t *testing.T) {
	t.Parallel()

	res := runner.Result{Status: cyclesdomain.PhaseStatusSucceeded}
	raw, err := json.Marshal(res)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var generic map[string]any
	if err := json.Unmarshal(raw, &generic); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if _, ok := generic["status"]; !ok {
		t.Errorf("status must always serialise: %s", raw)
	}
	for _, k := range []string{"summary", "details", "raw_output", "truncated"} {
		if _, ok := generic[k]; ok {
			t.Errorf("zero-value field %q must omit from payload: %s", k, raw)
		}
	}
}

// TestResult_jsonShape_full pins the keys when every field is populated.
func TestResult_jsonShape_full(t *testing.T) {
	t.Parallel()

	res := runner.Result{
		Status:    cyclesdomain.PhaseStatusFailed,
		Summary:   "exit 1",
		Details:   json.RawMessage(`{"exit_code":1}`),
		RawOutput: "boom",
		Truncated: true,
	}
	raw, err := json.Marshal(res)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var generic map[string]any
	if err := json.Unmarshal(raw, &generic); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	wantKeys := []string{"status", "summary", "details", "raw_output", "truncated"}
	for _, k := range wantKeys {
		if _, ok := generic[k]; !ok {
			t.Errorf("missing JSON key %q in %s", k, raw)
		}
	}
	for k := range generic {
		if !contains(wantKeys, k) {
			t.Errorf("unexpected JSON key %q (full payload: %s)", k, raw)
		}
	}
}

// TestNewResult_passesSmallValuesThrough is the happy path: nothing is
// clipped and Truncated stays false.
