package cursor_test

import (
	"context"
	"strings"
	"testing"

	"github.com/AlexsanderHamir/Hamix/pkgs/agents/runner/cursor"
)

func TestRun_redactionAuthorizationHeader(t *testing.T) {
	t.Parallel()

	stderr := []byte("Authorization: Bearer sk-live-supersecrettoken\n")
	a := newAdapter(fakeExec(&captured{}, []byte(""), stderr, 1, nil, false))

	res, _ := a.Run(context.Background(), defaultRequest())
	if strings.Contains(res.RawOutput, "sk-live-supersecrettoken") {
		t.Errorf("RawOutput leaks bearer token: %q", res.RawOutput)
	}
	if !strings.Contains(res.RawOutput, "[REDACTED]") {
		t.Errorf("RawOutput missing redaction marker: %q", res.RawOutput)
	}
}

// TestRun_redactionCookieHeader proves Cookie and Set-Cookie header
// values are scrubbed from RawOutput. The Authorization header is
// already redacted (TestRun_redactionAuthorizationHeader), but Cookie
// and Set-Cookie headers are equally credential-bearing â€” a session
// cookie is functionally equivalent to a bearer token. Cursor's CLI
// can emit HTTP-style traces in verbose / error paths (or any embedded
// HTTP client logging) where these headers leak verbatim. Treating
// only Authorization as secret-shaped while leaving Cookie /
// Set-Cookie in the clear is a defense-in-depth gap. The fix matches
// both `Cookie:` and `Set-Cookie:` case-insensitively (the latter
// covers the response-side header variant) and consumes the rest of
// the line, mirroring the Authorization redaction shape exactly.
func TestRun_redactionCookieHeader(t *testing.T) {
	t.Parallel()

	stderr := []byte("Cookie: session=abc.def.ghi; csrf=xyz123\n" +
		"Set-Cookie: auth=tok-1234567890; Path=/; HttpOnly\n")
	a := newAdapter(fakeExec(&captured{}, []byte(""), stderr, 1, nil, false))

	res, _ := a.Run(context.Background(), defaultRequest())
	if strings.Contains(res.RawOutput, "session=abc.def.ghi") {
		t.Errorf("RawOutput leaks Cookie value: %q", res.RawOutput)
	}
	if strings.Contains(res.RawOutput, "csrf=xyz123") {
		t.Errorf("RawOutput leaks Cookie attribute: %q", res.RawOutput)
	}
	if strings.Contains(res.RawOutput, "auth=tok-1234567890") {
		t.Errorf("RawOutput leaks Set-Cookie value: %q", res.RawOutput)
	}
	if !strings.Contains(res.RawOutput, "Cookie: [REDACTED]") {
		t.Errorf("missing Cookie redaction marker: %q", res.RawOutput)
	}
	if !strings.Contains(res.RawOutput, "Set-Cookie: [REDACTED]") {
		t.Errorf("missing Set-Cookie redaction marker: %q", res.RawOutput)
	}
}

// TestRun_redactionT2AEnv proves HAMIX_* env values are scrubbed from
// RawOutput. Exact mechanism: stderr accidentally echoing an env line
// like "HAMIX_DATABASE_URL=postgres://...".
func TestRun_redactionT2AEnv(t *testing.T) {
	t.Parallel()

	stderr := []byte("env dump: HAMIX_DATABASE_URL=postgres://user:pw@host/db PATH=/usr/bin\n")
	a := newAdapter(fakeExec(&captured{}, []byte(""), stderr, 1, nil, false))

	res, _ := a.Run(context.Background(), defaultRequest())
	if strings.Contains(res.RawOutput, "postgres://user:pw@host/db") {
		t.Errorf("RawOutput leaks DATABASE_URL value: %q", res.RawOutput)
	}
	if !strings.Contains(res.RawOutput, "HAMIX_DATABASE_URL=[REDACTED]") {
		t.Errorf("expected HAMIX_DATABASE_URL=[REDACTED]: %q", res.RawOutput)
	}
}

// TestRun_redactionHomePath proves absolute home paths are rewritten to
// "~" so RawOutput does not depend on the operator's filesystem layout.
func TestRun_redactionHomePath(t *testing.T) {
	t.Parallel()

	stderr := []byte("error in /home/runner/.cache/cursor/config.json\nalso C:\\Users\\runner\\AppData\\Local\\cursor\\log.txt\n")
	a := newAdapter(fakeExec(&captured{}, []byte(""), stderr, 1, nil, false))

	res, _ := a.Run(context.Background(), defaultRequest())
	if strings.Contains(res.RawOutput, "/home/runner/") {
		t.Errorf("Unix home path not redacted: %q", res.RawOutput)
	}
	if strings.Contains(res.RawOutput, `C:\Users\runner\`) {
		t.Errorf("Windows home path not redacted: %q", res.RawOutput)
	}
	if !strings.Contains(res.RawOutput, "~/.cache/cursor/config.json") {
		t.Errorf("expected ~/-prefixed unix path in: %q", res.RawOutput)
	}
}

// TestRedact_publicHelper covers the exported Redact entry point used by
// future callers (worker logs).
func TestRedact_publicHelper(t *testing.T) {
	t.Parallel()

	in := "Authorization: Bearer abc.def.ghi\nHAMIX_FOO=secretvalue\nCookie: sid=cookie-secret-12345\nSet-Cookie: x=y; HttpOnly\n"
	got := cursor.Redact(in)
	if strings.Contains(got, "abc.def.ghi") || strings.Contains(got, "secretvalue") {
		t.Errorf("Redact leaked secret: %q", got)
	}
	if strings.Contains(got, "cookie-secret-12345") {
		t.Errorf("Redact leaked Cookie value: %q", got)
	}
	if strings.Contains(got, "x=y") {
		t.Errorf("Redact leaked Set-Cookie value: %q", got)
	}
}

// TestRun_envAllowlist asserts that DATABASE_URL and HAMIX_* keys are
// stripped from the env passed to the child process even when the caller
// places them in Request.Env. This is the defense-in-depth guarantee.
//
// Cannot run in parallel: t.Setenv mutates process-global state.
