package cursor_test

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/AlexsanderHamir/Hamix/pkgs/agents/runner"
	cyclesdomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcycles/domain"
)

func TestRun_isErrorTrueMapsToFailure(t *testing.T) {
	t.Parallel()

	stdout := []byte(`{"type":"result","subtype":"error","is_error":true,"result":"could not authenticate","session_id":"sess-err","request_id":"req-err"}`)
	a := newAdapter(fakeExec(&captured{}, stdout, nil, 0, nil, false))

	res, err := a.Run(context.Background(), defaultRequest())
	if !errors.Is(err, runner.ErrNonZeroExit) {
		t.Fatalf("err: got %v want errors.Is(_, ErrNonZeroExit)", err)
	}
	if res.Status != cyclesdomain.PhaseStatusFailed {
		t.Errorf("Status: got %q want %q", res.Status, cyclesdomain.PhaseStatusFailed)
	}
	if res.Summary != "could not authenticate" {
		t.Errorf("Summary: got %q want the agent's result text", res.Summary)
	}
	var details struct {
		IsError   bool   `json:"is_error"`
		Subtype   string `json:"subtype"`
		SessionID string `json:"session_id"`
	}
	if err := json.Unmarshal(res.Details, &details); err != nil {
		t.Fatalf("Details unmarshal: %v (raw=%s)", err, res.Details)
	}
	if !details.IsError || details.Subtype != "error" {
		t.Errorf("Details mismatch: got is_error=%v subtype=%q", details.IsError, details.Subtype)
	}
	if details.SessionID != "sess-err" {
		t.Errorf("Details.session_id: got %q", details.SessionID)
	}
}

// TestRun_isErrorTrueWithEmptyResultGetsFallbackSummary covers the
// edge case where cursor-agent sets is_error=true but does not emit a
// "result" string. The Summary must still be non-empty so the audit
// row is honest about the failure.
func TestRun_isErrorTrueWithEmptyResultGetsFallbackSummary(t *testing.T) {
	t.Parallel()

	stdout := []byte(`{"type":"result","is_error":true}`)
	a := newAdapter(fakeExec(&captured{}, stdout, nil, 0, nil, false))

	res, err := a.Run(context.Background(), defaultRequest())
	if !errors.Is(err, runner.ErrNonZeroExit) {
		t.Fatalf("err: got %v want errors.Is(_, ErrNonZeroExit)", err)
	}
	if res.Summary == "" {
		t.Errorf("Summary must not be empty on is_error fallback")
	}
}

// TestRun_nonZeroExit asserts the documented error mapping plus the
// stderr_tail-in-Details contract.
func TestRun_nonZeroExit(t *testing.T) {
	t.Parallel()

	stderr := []byte("compile failed\nerror: missing semicolon\n")
	var c captured
	a := newAdapter(fakeExec(&c, []byte(""), stderr, 7, nil, false))

	res, err := a.Run(context.Background(), defaultRequest())
	if !errors.Is(err, runner.ErrNonZeroExit) {
		t.Fatalf("err: got %v want errors.Is(_, ErrNonZeroExit)", err)
	}
	if res.Status != cyclesdomain.PhaseStatusFailed {
		t.Errorf("Status: got %q want %q", res.Status, cyclesdomain.PhaseStatusFailed)
	}
	if !strings.Contains(res.Summary, "exit 7") {
		t.Errorf("Summary should mention exit code: got %q", res.Summary)
	}
	if !strings.Contains(res.Summary, "compile failed") {
		t.Errorf("Summary should include first stderr line hint: got %q", res.Summary)
	}
	var details struct {
		StderrTail string `json:"stderr_tail"`
	}
	if err := json.Unmarshal(res.Details, &details); err != nil {
		t.Fatalf("Details unmarshal: %v (raw=%s)", err, res.Details)
	}
	if !strings.Contains(details.StderrTail, "missing semicolon") {
		t.Errorf("stderr_tail missing expected content: %q", details.StderrTail)
	}
	if !strings.Contains(res.RawOutput, "missing semicolon") {
		t.Errorf("RawOutput should include redacted stderr: %q", res.RawOutput)
	}
}

func TestRun_execErrorIncludesDiagnostics(t *testing.T) {
	t.Parallel()

	execErr := errors.New("fork/exec /home/runner/bin/cursor-agent: no such file or directory")
	a := newAdapter(fakeExec(&captured{}, []byte("stdout before failure"), []byte("stderr before failure"), 0, execErr, false))

	res, err := a.Run(context.Background(), defaultRequest())
	if !errors.Is(err, runner.ErrInvalidOutput) {
		t.Fatalf("err: got %v want errors.Is(_, ErrInvalidOutput)", err)
	}
	if !strings.Contains(res.Summary, "no such file or directory") {
		t.Fatalf("Summary = %q, want exec error hint", res.Summary)
	}
	var details struct {
		FailureStage string   `json:"failure_stage"`
		Error        string   `json:"error"`
		StdoutTail   string   `json:"stdout_tail"`
		StderrTail   string   `json:"stderr_tail"`
		Binary       string   `json:"binary"`
		Argv         []string `json:"argv"`
		WorkingDir   string   `json:"working_dir"`
	}
	if err := json.Unmarshal(res.Details, &details); err != nil {
		t.Fatalf("Details unmarshal: %v raw=%s", err, res.Details)
	}
	if details.FailureStage != "exec" || !strings.Contains(details.Error, "no such file") {
		t.Fatalf("details missing exec error context: %+v", details)
	}
	if strings.Contains(details.Error, "/home/runner") {
		t.Fatalf("details error leaked home path: %+v", details)
	}
	if details.StdoutTail != "stdout before failure" || details.StderrTail != "stderr before failure" {
		t.Fatalf("details tails = stdout %q stderr %q", details.StdoutTail, details.StderrTail)
	}
	if details.Binary == "" || len(details.Argv) == 0 || details.WorkingDir == "" {
		t.Fatalf("details missing invocation context: %+v", details)
	}
}

// TestRun_invalidJSON exercises the parse-failure branch.
func TestRun_invalidJSON(t *testing.T) {
	t.Parallel()

	a := newAdapter(fakeExec(&captured{}, []byte("not json at all"), nil, 0, nil, false))
	res, err := a.Run(context.Background(), defaultRequest())
	if !errors.Is(err, runner.ErrInvalidOutput) {
		t.Fatalf("err: got %v want errors.Is(_, ErrInvalidOutput)", err)
	}
	if res.Status != cyclesdomain.PhaseStatusFailed {
		t.Errorf("Status: got %q", res.Status)
	}
	if !strings.Contains(res.Summary, "invalid output") {
		t.Fatalf("Summary = %q, want invalid output", res.Summary)
	}
	var details struct {
		FailureStage string `json:"failure_stage"`
		Error        string `json:"error"`
		StdoutTail   string `json:"stdout_tail"`
	}
	if err := json.Unmarshal(res.Details, &details); err != nil {
		t.Fatalf("Details unmarshal: %v raw=%s", err, res.Details)
	}
	if details.FailureStage != "parse_stdout" || details.Error == "" || details.StdoutTail != "not json at all" {
		t.Fatalf("parse failure details = %+v", details)
	}
}

// TestRun_emptyStdoutInvalid catches an edge case: 0 exit but no stdout
// must be ErrInvalidOutput, not silent success.
func TestRun_emptyStdoutInvalid(t *testing.T) {
	t.Parallel()

	a := newAdapter(fakeExec(&captured{}, []byte("   "), nil, 0, nil, false))
	_, err := a.Run(context.Background(), defaultRequest())
	if !errors.Is(err, runner.ErrInvalidOutput) {
		t.Errorf("got %v want errors.Is(_, ErrInvalidOutput)", err)
	}
}

// TestRun_timeout drives the per-call timeout path: the ExecFn blocks
// until ctx is cancelled, the adapter must return ErrTimeout with status
// Failed.
func TestRun_timeout(t *testing.T) {
	t.Parallel()

	a := newAdapter(fakeExec(&captured{}, nil, nil, 0, nil, true))
	req := defaultRequest()
	req.Timeout = 25 * time.Millisecond

	res, err := a.Run(context.Background(), req)
	if !errors.Is(err, runner.ErrTimeout) {
		t.Fatalf("err: got %v want errors.Is(_, ErrTimeout)", err)
	}
	if res.Status != cyclesdomain.PhaseStatusFailed {
		t.Errorf("Status on timeout: got %q want %q", res.Status, cyclesdomain.PhaseStatusFailed)
	}
	var details struct {
		FailureStage      string `json:"failure_stage"`
		TimeoutConfigured bool   `json:"timeout_configured"`
		TimeoutNS         int64  `json:"timeout_ns"`
	}
	if err := json.Unmarshal(res.Details, &details); err != nil {
		t.Fatalf("Details unmarshal: %v raw=%s", err, res.Details)
	}
	if details.FailureStage != "timeout" || !details.TimeoutConfigured || details.TimeoutNS <= 0 {
		t.Fatalf("timeout details = %+v", details)
	}
}

// TestRun_alreadyCancelledContext short-circuits without invoking exec.
func TestRun_alreadyCancelledContext(t *testing.T) {
	t.Parallel()

	called := false
	exec := func(ctx context.Context, dir string, env []string, stdin []byte, name string, args ...string) ([]byte, []byte, int, error) {
		called = true
		return nil, nil, 0, nil
	}
	a := newAdapter(exec)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := a.Run(ctx, defaultRequest())
	if !errors.Is(err, runner.ErrTimeout) {
		t.Errorf("got %v want errors.Is(_, ErrTimeout)", err)
	}
	if called {
		t.Errorf("exec must not be invoked when ctx is already cancelled")
	}
}

// TestRun_redactionAuthorizationHeader proves Authorization values are
// scrubbed from RawOutput.
