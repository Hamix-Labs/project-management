package cursor_test

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/AlexsanderHamir/Hamix/pkgs/agents/runner"
	"github.com/AlexsanderHamir/Hamix/pkgs/agents/runner/cursor"
	cyclesdomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcycles/domain"
)

func TestRun_streamJSONCapturesResolvedModel(t *testing.T) {
	t.Parallel()

	stdout := []byte(
		`{"type":"system","subtype":"init","apiKeySource":"login","cwd":"/tmp","session_id":"sess-abc","model":"Claude 4 Sonnet","permissionMode":"default"}` + "\n" +
			`{"type":"user","message":{"role":"user","content":[{"type":"text","text":"do the thing"}]},"session_id":"sess-abc"}` + "\n" +
			`{"type":"assistant","message":{"role":"assistant","content":[{"type":"text","text":"working"}]},"session_id":"sess-abc"}` + "\n" +
			`{"type":"result","subtype":"success","is_error":false,"duration_ms":1200,"duration_api_ms":1100,"result":"all good","session_id":"sess-abc","request_id":"req-xyz"}` + "\n",
	)
	a := newAdapter(fakeExec(&captured{}, stdout, nil, 0, nil, false))

	res, err := a.Run(context.Background(), defaultRequest())
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if res.Status != cyclesdomain.PhaseStatusSucceeded {
		t.Errorf("Status: got %q want %q", res.Status, cyclesdomain.PhaseStatusSucceeded)
	}
	if res.Summary != "all good" {
		t.Errorf("Summary: got %q want %q", res.Summary, "all good")
	}
	if res.ResolvedModel != "Claude 4 Sonnet" {
		t.Errorf("ResolvedModel: got %q want %q", res.ResolvedModel, "Claude 4 Sonnet")
	}
	var details struct {
		ResolvedModel string `json:"resolved_model"`
		SessionID     string `json:"session_id"`
	}
	if err := json.Unmarshal(res.Details, &details); err != nil {
		t.Fatalf("Details unmarshal: %v (raw=%s)", err, res.Details)
	}
	if details.ResolvedModel != "Claude 4 Sonnet" {
		t.Errorf("Details.resolved_model: got %q want %q", details.ResolvedModel, "Claude 4 Sonnet")
	}
	if details.SessionID != "sess-abc" {
		t.Errorf("Details.session_id should come from the terminal result event: got %q", details.SessionID)
	}
}

func TestRun_streamJSONEmitsLiveProgress(t *testing.T) {
	t.Parallel()

	stdout := []byte(
		`{"type":"system","subtype":"init","model":"Claude 4 Sonnet"}` + "\n" +
			`{"type":"assistant","message":{"role":"assistant","content":[{"type":"text","text":"Reading the task files."}]}}` + "\n" +
			`{"type":"tool_call","subtype":"started","name":"ReadFile"}` + "\n" +
			`{"type":"result","subtype":"success","is_error":false,"result":"done","session_id":"sess-live"}` + "\n",
	)
	var c captured
	a := newAdapter(nil, func(opts *cursor.Options) {
		opts.StreamExecFn = fakeStreamExec(&c, stdout, nil, 0, nil)
	})
	var progress []runner.ProgressEvent
	req := defaultRequest()
	req.OnProgress = func(ev runner.ProgressEvent) {
		progress = append(progress, ev)
	}

	res, err := a.Run(context.Background(), req)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if res.Summary != "done" {
		t.Fatalf("Summary: got %q want done", res.Summary)
	}
	if len(progress) != 4 {
		t.Fatalf("progress count: got %d want 4 (%+v)", len(progress), progress)
	}
	if progress[0].Kind != runner.ProgressRunStateKind || progress[0].Subtype != runner.ProgressRunStateSetupSpawn {
		t.Errorf("setup spawn progress: %+v", progress[0])
	}
	if progress[1].Kind != "system" || progress[1].Message != "Using Claude 4 Sonnet" {
		t.Errorf("system progress: %+v", progress[1])
	}
	if progress[2].Kind != "assistant" || progress[2].Message != "Reading the task files." {
		t.Errorf("assistant progress: %+v", progress[2])
	}
	if progress[3].Kind != "tool_call" || progress[3].Subtype != "started" || progress[3].Tool != "ReadFile" {
		t.Errorf("tool progress: %+v", progress[3])
	}
	var payload map[string]any
	if err := json.Unmarshal(progress[3].Payload, &payload); err != nil {
		t.Fatalf("tool progress payload: %v raw=%s", err, progress[3].Payload)
	}
	if payload["type"] != "tool_call" || payload["name"] != "ReadFile" {
		t.Fatalf("tool progress payload=%v", payload)
	}
}

func TestRun_streamJSONSummarizesToolInputsForLiveProgress(t *testing.T) {
	t.Parallel()

	stdout := []byte(
		`{"type":"tool_call","subtype":"started","input":{"glob_pattern":"*.go","target_directory":"/repo/pkgs/agents/worker"}}` + "\n" +
			`{"type":"tool_call","subtype":"started","input":{"glob_pattern":"*.go","target_directory":"/repo/pkgs/agents/runner"}}` + "\n" +
			`{"type":"tool_call","subtype":"started","input":{"path":"/repo/pkgs/tasks/store/facade_projects.go","offset":1,"limit":91}}` + "\n" +
			`{"type":"tool_call","subtype":"started","name":"Shell","input":{"command":"go test ./pkgs/agents/runner/cursor","description":"Run cursor parser tests"}}` + "\n" +
			`{"type":"result","subtype":"success","is_error":false,"result":"done","session_id":"sess-live"}` + "\n",
	)
	var c captured
	a := newAdapter(nil, func(opts *cursor.Options) {
		opts.StreamExecFn = fakeStreamExec(&c, stdout, nil, 0, nil)
	})
	var progress []runner.ProgressEvent
	req := defaultRequest()
	req.OnProgress = func(ev runner.ProgressEvent) {
		progress = append(progress, ev)
	}

	if _, err := a.Run(context.Background(), req); err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(progress) != 5 {
		t.Fatalf("progress count: got %d want 5 (%+v)", len(progress), progress)
	}
	got := []string{
		progress[0].Message,
		progress[1].Message,
		progress[2].Message,
		progress[3].Message,
		progress[4].Message,
	}
	want := []string{
		"Launching cursor-agent…",
		"Searching for *.go in worker",
		"Searching for *.go in runner",
		"Reading facade_projects.go L1-91",
		"Run cursor parser tests",
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("progress[%d].Message = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestRun_streamJSONSummarizesNestedToolCallsForLiveProgress(t *testing.T) {
	t.Parallel()

	stdout := []byte(
		`{"type":"tool_call","subtype":"started","tool_call":{"readToolCall":{"args":{"path":"README.md"}}}}` + "\n" +
			`{"type":"tool_call","subtype":"completed","tool_call":{"writeToolCall":{"args":{"path":"summary.txt","fileText":"# Summary"},"result":{"success":{"path":"/repo/summary.txt","linesCreated":1}}}}}` + "\n" +
			`{"type":"tool_call","subtype":"started","tool_call":{"editToolCall":{"args":{"path":"/repo/improvements_01.md","streamContent":"Enforce a non-zero default cap"}}}}` + "\n" +
			`{"type":"tool_call","subtype":"started","tool_call":{"function":{"name":"rg","arguments":"{\"pattern\":\"TODO\",\"path\":\"/repo/pkgs/agents/worker\"}"}}}` + "\n" +
			`{"type":"result","subtype":"success","is_error":false,"result":"done","session_id":"sess-live"}` + "\n",
	)
	var c captured
	a := newAdapter(nil, func(opts *cursor.Options) {
		opts.StreamExecFn = fakeStreamExec(&c, stdout, nil, 0, nil)
	})
	var progress []runner.ProgressEvent
	req := defaultRequest()
	req.OnProgress = func(ev runner.ProgressEvent) {
		progress = append(progress, ev)
	}

	if _, err := a.Run(context.Background(), req); err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(progress) != 5 {
		t.Fatalf("progress count: got %d want 5 (%+v)", len(progress), progress)
	}
	got := []string{
		progress[0].Message,
		progress[1].Message,
		progress[2].Message,
		progress[3].Message,
		progress[4].Message,
	}
	want := []string{
		"Launching cursor-agent…",
		"Reading README.md",
		"Writing summary.txt",
		"Editing improvements_01.md",
		"Searching TODO in worker",
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("progress[%d].Message = %q, want %q", i, got[i], want[i])
		}
	}
	gotTools := []string{
		progress[1].Tool,
		progress[2].Tool,
		progress[3].Tool,
		progress[4].Tool,
	}
	wantTools := []string{"ReadFile", "WriteFile", "EditFile", "rg"}
	for i := range wantTools {
		if gotTools[i] != wantTools[i] {
			t.Fatalf("progress[%d].Tool = %q, want %q", i+1, gotTools[i], wantTools[i])
		}
	}
}

func TestRun_streamJSONIgnoresClosedPipeAfterResult(t *testing.T) {
	t.Parallel()

	stdout := []byte(
		`{"type":"assistant","message":{"role":"assistant","content":[{"type":"text","text":"done"}]}}` + "\n" +
			`{"type":"result","subtype":"success","is_error":false,"result":"all done","session_id":"sess-closed-pipe"}` + "\n",
	)
	a := newAdapter(nil, func(opts *cursor.Options) {
		opts.StreamExecFn = fakeStreamExec(&captured{}, stdout, nil, 0, os.ErrClosed)
	})

	res, err := a.Run(context.Background(), defaultRequest())
	if err != nil {
		t.Fatalf("Run: got err %v, want success", err)
	}
	if res.Status != cyclesdomain.PhaseStatusSucceeded || res.Summary != "all done" {
		t.Fatalf("Result = %+v, want succeeded/all done", res)
	}
}

func TestRun_streamJSONClosedPipeWithoutResultUsesAssistantFallback(t *testing.T) {
	t.Parallel()

	stdout := []byte(`{"type":"assistant","message":{"role":"assistant","content":[{"type":"text","text":"finished but no result"}]}}` + "\n")
	a := newAdapter(nil, func(opts *cursor.Options) {
		opts.StreamExecFn = fakeStreamExec(&captured{}, stdout, nil, 0, os.ErrClosed)
	})

	res, err := a.Run(context.Background(), defaultRequest())
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if res.Status != cyclesdomain.PhaseStatusSucceeded || res.Summary != "finished but no result" {
		t.Fatalf("Result = %+v, want degraded success from assistant text", res)
	}
	var details struct {
		MissingTerminalResult bool `json:"missing_terminal_result"`
	}
	if err := json.Unmarshal(res.Details, &details); err != nil {
		t.Fatalf("Details unmarshal: %v raw=%s", err, res.Details)
	}
	if !details.MissingTerminalResult {
		t.Fatalf("Details = %+v", details)
	}
}

func TestRun_streamJSONMissingResultWithOpenToolCallStaysInvalid(t *testing.T) {
	t.Parallel()

	stdout := []byte(
		`{"type":"assistant","message":{"role":"assistant","content":[{"type":"text","text":"starting"}]}}` + "\n" +
			`{"type":"tool_call","subtype":"started","call_id":"tool-1"}` + "\n",
	)
	a := newAdapter(nil, func(opts *cursor.Options) {
		opts.StreamExecFn = fakeStreamExec(&captured{}, stdout, nil, 0, os.ErrClosed)
	})

	res, err := a.Run(context.Background(), defaultRequest())
	if !errors.Is(err, runner.ErrInvalidOutput) {
		t.Fatalf("err: got %v want errors.Is(_, ErrInvalidOutput)", err)
	}
	if !strings.Contains(res.Summary, "open tool call") {
		t.Fatalf("Summary = %q, want open tool call hint", res.Summary)
	}
}

func TestRun_streamJSONMissingResultAfterCompletedToolUsesLastAssistant(t *testing.T) {
	t.Parallel()

	stdout := []byte(
		`{"type":"assistant","message":{"role":"assistant","content":[{"type":"text","text":"starting"}]}}` + "\n" +
			`{"type":"tool_call","subtype":"started","call_id":"tool-1"}` + "\n" +
			`{"type":"tool_call","subtype":"completed","call_id":"tool-1"}` + "\n" +
			`{"type":"assistant","message":{"role":"assistant","content":[{"type":"text","text":"done after tool"}]}}` + "\n",
	)
	a := newAdapter(nil, func(opts *cursor.Options) {
		opts.StreamExecFn = fakeStreamExec(&captured{}, stdout, nil, 0, os.ErrClosed)
	})

	res, err := a.Run(context.Background(), defaultRequest())
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if res.Summary != "done after tool" {
		t.Fatalf("Summary = %q, want last assistant message", res.Summary)
	}
}

// TestRun_streamJSONMissingSystemEventLeavesResolvedModelEmpty covers
// the forward-compat gap: an older cursor-agent build might emit a
// stream without a `system.init` event (e.g. if a future version
// reshapes the init shape or drops it entirely). The adapter must
// still succeed by finding the terminal result event, and record
// ResolvedModel="" â€” truthful "unknown" rather than a fabricated
// fallback.
func TestRun_streamJSONMissingSystemEventLeavesResolvedModelEmpty(t *testing.T) {
	t.Parallel()

	stdout := []byte(
		`{"type":"assistant","message":{"role":"assistant","content":[{"type":"text","text":"working"}]},"session_id":"sess-no-init"}` + "\n" +
			`{"type":"result","subtype":"success","is_error":false,"result":"done","session_id":"sess-no-init"}` + "\n",
	)
	a := newAdapter(fakeExec(&captured{}, stdout, nil, 0, nil, false))

	res, err := a.Run(context.Background(), defaultRequest())
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if res.Status != cyclesdomain.PhaseStatusSucceeded {
		t.Errorf("Status: got %q", res.Status)
	}
	if res.ResolvedModel != "" {
		t.Errorf("ResolvedModel must stay empty when the stream lacks a system.init event; got %q", res.ResolvedModel)
	}
	if res.Summary != "done" {
		t.Errorf("Summary: got %q", res.Summary)
	}
}

// TestRun_isErrorTrueMapsToFailure pins the contract that an exit-0
// run with cursor-agent reporting "is_error": true is treated as a
// recoverable runner failure (ErrNonZeroExit + PhaseStatusFailed) so
// the worker writes a failed cycle instead of silently treating the
// run as success. The agent's own "result" text becomes the Summary.
