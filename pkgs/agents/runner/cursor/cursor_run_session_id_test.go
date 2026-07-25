package cursor_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/AlexsanderHamir/Hamix/pkgs/agents/runner"
)

// TestRun_timeoutIncludesSessionIDFromInit pins Layer A behavior for the
// timeout/cancel path: when stdout captured before kill contains only the
// system.init frame, failureDetails must still surface session_id from that
// frame so downstream resume decisions can use it.
func TestRun_timeoutIncludesSessionIDFromInit(t *testing.T) {
	t.Parallel()

	initOnly := []byte(`{"type":"system","subtype":"init","session_id":"sess-init-timeout","model":"Claude 4 Sonnet"}` + "\n")

	exec := func(ctx context.Context, dir string, env []string, stdin []byte, name string, args ...string) ([]byte, []byte, int, error) {
		<-ctx.Done()
		return initOnly, nil, 0, ctx.Err()
	}
	a := newAdapter(exec)
	req := defaultRequest()
	req.Timeout = 25 * time.Millisecond

	res, err := a.Run(context.Background(), req)
	if !errors.Is(err, runner.ErrTimeout) {
		t.Fatalf("err: got %v want errors.Is(_, ErrTimeout)", err)
	}
	var details struct {
		FailureStage string `json:"failure_stage"`
		SessionID    string `json:"session_id"`
	}
	if err := json.Unmarshal(res.Details, &details); err != nil {
		t.Fatalf("Details unmarshal: %v raw=%s", err, res.Details)
	}
	if details.FailureStage != "timeout" {
		t.Fatalf("failure_stage: got %q want timeout", details.FailureStage)
	}
	if details.SessionID != "sess-init-timeout" {
		t.Fatalf("session_id: got %q want sess-init-timeout (details=%s)", details.SessionID, res.Details)
	}
}

// TestRun_nonZeroExitIncludesSessionIDFromInit pins Layer A behavior for the
// non-zero-exit path: init frame in stdout must contribute session_id to
// details even though the CLI never emitted a terminal result event.
func TestRun_nonZeroExitIncludesSessionIDFromInit(t *testing.T) {
	t.Parallel()

	stdout := []byte(`{"type":"system","subtype":"init","session_id":"sess-nonzero","model":"Claude 4 Sonnet"}` + "\n")
	stderr := []byte("boom\n")
	a := newAdapter(fakeExec(&captured{}, stdout, stderr, 3, nil, false))

	res, err := a.Run(context.Background(), defaultRequest())
	if !errors.Is(err, runner.ErrNonZeroExit) {
		t.Fatalf("err: got %v want errors.Is(_, ErrNonZeroExit)", err)
	}
	var details struct {
		SessionID string `json:"session_id"`
	}
	if err := json.Unmarshal(res.Details, &details); err != nil {
		t.Fatalf("Details unmarshal: %v raw=%s", err, res.Details)
	}
	if details.SessionID != "sess-nonzero" {
		t.Fatalf("session_id: got %q want sess-nonzero (details=%s)", details.SessionID, res.Details)
	}
}

// TestRun_successBackfillsSessionIDFromInit pins Layer A behavior for the
// success path: when the terminal result event omits session_id, the id
// captured earlier from init must be preserved in Details.
func TestRun_successBackfillsSessionIDFromInit(t *testing.T) {
	t.Parallel()

	stdout := []byte(
		`{"type":"system","subtype":"init","session_id":"sess-backfill","model":"Claude 4 Sonnet"}` + "\n" +
			`{"type":"assistant","message":{"role":"assistant","content":[{"type":"text","text":"working"}]}}` + "\n" +
			`{"type":"result","subtype":"success","is_error":false,"result":"all good"}` + "\n",
	)
	a := newAdapter(fakeExec(&captured{}, stdout, nil, 0, nil, false))

	res, err := a.Run(context.Background(), defaultRequest())
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	var details struct {
		SessionID string `json:"session_id"`
	}
	if err := json.Unmarshal(res.Details, &details); err != nil {
		t.Fatalf("Details unmarshal: %v raw=%s", err, res.Details)
	}
	if details.SessionID != "sess-backfill" {
		t.Fatalf("session_id: got %q want sess-backfill (details=%s)", details.SessionID, res.Details)
	}
}
