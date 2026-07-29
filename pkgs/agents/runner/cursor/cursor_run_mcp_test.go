package cursor_test

import (
	"context"
	"testing"
)

func TestRun_approveMCPsAndAddDir(t *testing.T) {
	t.Parallel()
	stdout := []byte(`{"type":"result","subtype":"success","is_error":false,"duration_ms":1,"result":"ok","session_id":"s1"}`)
	var c captured
	a := newAdapter(fakeExec(&c, stdout, nil, 0, nil, false))
	req := defaultRequest()
	req.ApproveMCPs = true
	req.TrustWorkspace = true
	req.AddDirs = []string{"/tmp/hamix-worker/cycle-1"}
	if _, err := a.Run(context.Background(), req); err != nil {
		t.Fatalf("Run: %v", err)
	}
	want := []string{
		"--print", "--output-format", "stream-json", "--force", "--approve-mcps", "--trust",
		"--workspace", "/repo/work",
		"--add-dir", "/tmp/hamix-worker/cycle-1",
	}
	if !equalStrSlice(c.args, want) {
		t.Fatalf("args: got %v want %v", c.args, want)
	}
}
