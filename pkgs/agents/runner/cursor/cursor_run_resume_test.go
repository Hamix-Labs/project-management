package cursor_test

import (
	"context"
	"errors"
	"testing"

	"github.com/AlexsanderHamir/Hamix/pkgs/agents/runner"
)

func TestRun_resumeArgv(t *testing.T) {
	t.Parallel()
	stdout := []byte(`{"type":"result","subtype":"success","is_error":false,"result":"ok","session_id":"sess-abc"}`)
	var c captured
	a := newAdapter(fakeExec(&c, stdout, nil, 0, nil, false))
	req := defaultRequest()
	req.ResumeSessionID = "sess-prior"
	_, err := a.Run(context.Background(), req)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	want := []string{"--print", "--output-format", "stream-json", "--force", "--resume", "sess-prior", "--workspace", "/repo/work"}
	if !equalStrSlice(c.args, want) {
		t.Fatalf("args: got %v want %v", c.args, want)
	}
}

func TestRun_resumeSessionFailure(t *testing.T) {
	t.Parallel()
	stderr := []byte("Error: resume session not found\n")
	var c captured
	a := newAdapter(fakeExec(&c, nil, stderr, 1, nil, false))
	req := defaultRequest()
	req.ResumeSessionID = "gone"
	_, err := a.Run(context.Background(), req)
	if !errors.Is(err, runner.ErrResumeSession) {
		t.Fatalf("err: got %v want ErrResumeSession", err)
	}
}

// TestRun_successPath covers the happy path: 0 exit + valid JSON stdout
// produces a Result with PhaseStatusSucceeded and the parsed Summary /
// Details intact.
