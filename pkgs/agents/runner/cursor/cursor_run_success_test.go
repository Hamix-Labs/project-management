package cursor_test

import (
	"context"
	"encoding/json"
	"testing"

	cyclesdomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcycles/domain"
)

func TestRun_successPath(t *testing.T) {
	t.Parallel()

	stdout := []byte(`{"type":"result","subtype":"success","is_error":false,"duration_ms":1200,"duration_api_ms":1100,"result":"all good","session_id":"sess-abc","request_id":"req-xyz","usage":{"inputTokens":10,"outputTokens":3}}`)
	var c captured
	a := newAdapter(fakeExec(&c, stdout, nil, 0, nil, false))

	res, err := a.Run(context.Background(), defaultRequest())
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if res.Status != cyclesdomain.PhaseStatusSucceeded {
		t.Errorf("Status: got %q want %q", res.Status, cyclesdomain.PhaseStatusSucceeded)
	}
	if res.Summary != "all good" {
		t.Errorf("Summary: got %q", res.Summary)
	}
	var details struct {
		Type       string          `json:"type"`
		Subtype    string          `json:"subtype"`
		IsError    bool            `json:"is_error"`
		DurationMs int64           `json:"duration_ms"`
		SessionID  string          `json:"session_id"`
		RequestID  string          `json:"request_id"`
		Usage      json.RawMessage `json:"usage"`
	}
	if err := json.Unmarshal(res.Details, &details); err != nil {
		t.Fatalf("Details unmarshal: %v (raw=%s)", err, res.Details)
	}
	if details.Type != "result" || details.Subtype != "success" {
		t.Errorf("Details type/subtype: got %q/%q want result/success", details.Type, details.Subtype)
	}
	if details.IsError {
		t.Errorf("Details.is_error must be false on happy path")
	}
	if details.SessionID != "sess-abc" || details.RequestID != "req-xyz" {
		t.Errorf("Details ids: got session=%q request=%q", details.SessionID, details.RequestID)
	}
	if details.DurationMs != 1200 {
		t.Errorf("Details.duration_ms: got %d", details.DurationMs)
	}
	if len(details.Usage) == 0 {
		t.Errorf("Details.usage missing; got %s", res.Details)
	}
	if c.name != "fake-cursor-agent" {
		t.Errorf("invoked name: got %q", c.name)
	}
	wantArgs := []string{"--print", "--output-format", "stream-json", "--force", "--workspace", "/repo/work"}
	if !equalStrSlice(c.args, wantArgs) {
		t.Errorf("args: got %v want %v", c.args, wantArgs)
	}
	if res.ResolvedModel != "" {
		t.Errorf("ResolvedModel on legacy single-object stdout should be empty; got %q", res.ResolvedModel)
	}
	if string(c.stdin) != "do the thing" {
		t.Errorf("stdin: got %q", c.stdin)
	}
	if c.dir != "/repo/work" {
		t.Errorf("dir: got %q", c.dir)
	}
}

// TestRun_streamJSONCapturesResolvedModel pins the new plumbing: when
// cursor-agent emits its --output-format stream-json NDJSON, the
// adapter must
//   - extract the resolved model from the first `system.init` event
//     (the ONLY surface where cursor-agent reveals what model `auto`
//     routed to â€” the terminal `result` event has no model field; see
//     https://cursor.com/docs/cli/reference/output-format),
//   - still recover the Summary / session_id / timing from the
//     terminal `result` event (wire-identical to the old json format),
//   - surface the captured model both on Result.ResolvedModel (so the
//     worker can record it in cycle MetaJSON as
//     cursor_model_resolved) AND inside Result.Details.resolved_model
//     (so the per-phase details_json audit row carries it too).
