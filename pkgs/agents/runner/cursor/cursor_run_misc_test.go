package cursor_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/AlexsanderHamir/Hamix/pkgs/agents/runner"
	"github.com/AlexsanderHamir/Hamix/pkgs/agents/runner/cursor"
)

// path; this test pins it explicitly with a non-default value.
func TestRun_workingDirPropagated(t *testing.T) {
	t.Parallel()

	var c captured
	a := newAdapter(fakeExec(&c, []byte(`{"type":"result","subtype":"success","result":"ok"}`), nil, 0, nil, false))
	req := defaultRequest()
	req.WorkingDir = "/some/other/repo"

	if _, err := a.Run(context.Background(), req); err != nil {
		t.Fatalf("Run: %v", err)
	}
	if c.dir != "/some/other/repo" {
		t.Errorf("dir: got %q want %q", c.dir, "/some/other/repo")
	}
}

// TestNew_defaults pins the documented default values so wire-up code
// can rely on them.
func TestNew_defaults(t *testing.T) {
	t.Parallel()

	a := cursor.New(cursor.Options{})
	if a.Name() != "cursor-cli" {
		t.Errorf("default Name: got %q", a.Name())
	}
	if a.Version() != "0.0.0-unknown" {
		t.Errorf("default Version: got %q", a.Version())
	}
}

// TestNew_overridesNameVersion confirms the worker can override the
// MetaJSON-recorded identity values at construction.
func TestNew_overridesNameVersion(t *testing.T) {
	t.Parallel()

	a := cursor.New(cursor.Options{Name: "cursor-prod", Version: "1.42.0"})
	if a.Name() != "cursor-prod" {
		t.Errorf("Name override lost: %q", a.Name())
	}
	if a.Version() != "1.42.0" {
		t.Errorf("Version override lost: %q", a.Version())
	}
}

// TestRunner_compileTimeConformance keeps the interface surface stable.
func TestRunner_compileTimeConformance(t *testing.T) {
	t.Parallel()
	var _ runner.Runner = cursor.New(cursor.Options{})
}

// TestRun_stderrTailDoesNotSplitMultibyteUTF8 pins the byte-cap tail
// slicing in stderrTailDetails to a UTF-8 rune boundary. Mirrors the
// fix landed in pkgs/repo/read_preview_io.go for file previews
// (TestReadFilePreview_truncatedDoesNotSplitMultibyteUTF8): both call
// sites take the last/first N bytes of a buffer that may contain
// multibyte runes at the boundary, and both must drop the partial
// leading rune before the bytes become a Go string. Without the fix
// json.Marshal silently rewrites the dangling continuation byte to
// U+FFFD, leaking a corrupted diagnostic into the audit
// task_cycle_phases.details_json payload (and into every API response
// that surfaces it).
func TestRun_stderrTailDoesNotSplitMultibyteUTF8(t *testing.T) {
	t.Parallel()

	// Place a 3-byte UTF-8 character ("中" = E4 B8 AD) at offset 0 so
	// the 8 KiB tail truncation lands exactly at index 1 — the second
	// byte of the rune (B8, a continuation byte). Total stderr length
	// = 3 + 8190 = 8193, so tail = stderr[8193-8192:] = stderr[1:].
	const trailing = 8190
	stderrIn := append([]byte("中"), bytes.Repeat([]byte("y"), trailing)...)
	if len(stderrIn) != 8193 {
		t.Fatalf("test setup: stderr len = %d, want 8193", len(stderrIn))
	}

	var c captured
	a := newAdapter(fakeExec(&c, []byte(""), stderrIn, 7, nil, false))

	res, err := a.Run(context.Background(), defaultRequest())
	if !errors.Is(err, runner.ErrNonZeroExit) {
		t.Fatalf("err: got %v want errors.Is(_, ErrNonZeroExit)", err)
	}
	var details struct {
		StderrTail string `json:"stderr_tail"`
	}
	if err := json.Unmarshal(res.Details, &details); err != nil {
		t.Fatalf("Details unmarshal: %v (raw=%s)", err, res.Details)
	}
	if !utf8.ValidString(details.StderrTail) {
		t.Fatalf("stderr_tail must be valid UTF-8 after byte-cap truncation; got %q", details.StderrTail)
	}
	if strings.HasPrefix(details.StderrTail, "\uFFFD") {
		t.Errorf("stderr_tail must not start with the U+FFFD replacement char (truncation split a UTF-8 rune); got prefix %q",
			details.StderrTail[:min(len(details.StderrTail), 6)])
	}
}

// than a substituted placeholder.
func TestAdapter_EffectiveModel(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name         string
		defaultModel string
		reqModel     string
		want         string
	}{
		{"both empty", "", "", ""},
		{"default only", "opus", "", "opus"},
		{"request overrides default", "opus", "sonnet-4.5", "sonnet-4.5"},
		{"request whitespace falls back", "opus", "   ", "opus"},
		{"request trimmed", "opus", "  sonnet-4.5  ", "sonnet-4.5"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			a := cursor.New(cursor.Options{
				DefaultCursorModel: tc.defaultModel,
				ExecFn:             fakeExec(&captured{}, []byte("{}"), nil, 0, nil, false),
			})
			got := a.EffectiveModel(runner.Request{CursorModel: tc.reqModel})
			if got != tc.want {
				t.Errorf("EffectiveModel(req.CursorModel=%q) with default %q: got %q want %q",
					tc.reqModel, tc.defaultModel, got, tc.want)
			}
		})
	}
}
