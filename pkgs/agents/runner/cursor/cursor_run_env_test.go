package cursor_test

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/AlexsanderHamir/Hamix/pkgs/agents/runner/cursor"
)

func TestRun_envAllowlist(t *testing.T) {
	t.Setenv("PATH", "/test/path")
	t.Setenv("HOME", "/home/runner")
	t.Setenv("DATABASE_URL", "postgres://should-not-leak")
	t.Setenv("HAMIX_SECRET_TOKEN", "should-not-leak")
	t.Setenv("ALLOWED_EXTRA", "yes-please")

	var c captured
	a := newAdapter(
		fakeExec(&c, []byte(`{"type":"result","subtype":"success","result":"ok"}`), nil, 0, nil, false),
		func(o *cursor.Options) {
			o.ExtraAllowedEnvKeys = []string{"ALLOWED_EXTRA"}
		},
	)
	req := defaultRequest()
	req.Env = map[string]string{
		"DATABASE_URL":     "from-request-must-also-be-stripped",
		"HAMIX_BACKDOOR":   "must-be-stripped",
		"REQUEST_PROVIDED": "request-wins-over-parent",
	}

	if _, err := a.Run(context.Background(), req); err != nil {
		t.Fatalf("Run: %v", err)
	}

	envMap := envSliceToMap(c.env)
	if _, present := envMap["DATABASE_URL"]; present {
		t.Errorf("DATABASE_URL must never be passed to child: %v", envMap)
	}
	for k := range envMap {
		if strings.HasPrefix(k, "HAMIX_") {
			t.Errorf("HAMIX_* keys must never be passed to child: %s", k)
		}
	}
	if envMap["PATH"] != "/test/path" {
		t.Errorf("PATH not passed through: got %q", envMap["PATH"])
	}
	if envMap["HOME"] != "/home/runner" {
		t.Errorf("HOME not passed through: got %q", envMap["HOME"])
	}
	if envMap["ALLOWED_EXTRA"] != "yes-please" {
		t.Errorf("ExtraAllowedEnvKeys not honoured: got %q", envMap["ALLOWED_EXTRA"])
	}
	if envMap["REQUEST_PROVIDED"] != "request-wins-over-parent" {
		t.Errorf("Request.Env not merged: got %q", envMap["REQUEST_PROVIDED"])
	}
}

// TestRun_envRequestShadowsParent: Request.Env overrides the inherited
// parent value for the same key.
//
// Cannot run in parallel: t.Setenv mutates process-global state.
func TestRun_envRequestShadowsParent(t *testing.T) {
	t.Setenv("PATH", "/parent/path")
	var c captured
	a := newAdapter(fakeExec(&c, []byte(`{"type":"result","subtype":"success","result":"ok"}`), nil, 0, nil, false))
	req := defaultRequest()
	req.Env = map[string]string{"PATH": "/request/path"}

	if _, err := a.Run(context.Background(), req); err != nil {
		t.Fatalf("Run: %v", err)
	}
	if envSliceToMap(c.env)["PATH"] != "/request/path" {
		t.Errorf("Request.Env did not shadow parent PATH: %v", c.env)
	}
}

// TestRun_envSystemVarsForwarded asserts the curated default-passthrough
// list (defaultPassthroughEnvKeys) actually reaches the child for the
// Windows system vars that were missing from the original {PATH, HOME,
// USERPROFILE} seed.
//
// 2026-04-19 incident: with SYSTEMDRIVE / SYSTEMROOT / TEMP / etc.
// stripped from the child env block, components in the cursor-agent
// process tree (Software Licensing Service, ETW, .NET CLR config
// loaders, Defender hooks) called ExpandEnvironmentStrings on hardcoded
// paths like "%SystemDrive%\\ProgramData\\Microsoft\\Windows\\Caches\\..."
// against the empty env, got the literal "%SystemDrive%\\..." string
// back, and CreateFile resolved it as a relative path under the child's
// cwd â€” which is AppSettings.RepoRoot. The child silently wrote a
// literal "%SystemDrive%" directory tree into the operator's worktree,
// surfacing as junk in `git status` and forcing a manual cleanup. This
// test pins the wider passthrough so a future "minimise the env" refactor
// cannot reintroduce the regression without flagging it loudly in CI.
//
// Cannot run in parallel: t.Setenv mutates process-global state.
func TestRun_envSystemVarsForwarded(t *testing.T) {
	cases := []struct {
		key, value string
	}{
		// Windows process model + shell (the canonical "system" set).
		{"SYSTEMDRIVE", "C:"},
		{"SYSTEMROOT", `C:\Windows`},
		{"WINDIR", `C:\Windows`},
		{"COMSPEC", `C:\Windows\System32\cmd.exe`},
		{"PATHEXT", ".COM;.EXE;.BAT;.CMD"},
		// Known folders.
		{"LOCALAPPDATA", `C:\Users\runner\AppData\Local`},
		{"APPDATA", `C:\Users\runner\AppData\Roaming`},
		{"PROGRAMDATA", `C:\ProgramData`},
		{"ALLUSERSPROFILE", `C:\ProgramData`},
		{"PUBLIC", `C:\Users\Public`},
		{"TEMP", `C:\Users\runner\AppData\Local\Temp`},
		{"TMP", `C:\Users\runner\AppData\Local\Temp`},
		// Program / DLL lookup.
		{"PROGRAMFILES", `C:\Program Files`},
		{"PROGRAMFILES(X86)", `C:\Program Files (x86)`},
		{"PROGRAMW6432", `C:\Program Files`},
		{"COMMONPROGRAMFILES", `C:\Program Files\Common Files`},
		{"COMMONPROGRAMFILES(X86)", `C:\Program Files (x86)\Common Files`},
		// Identity.
		{"USERNAME", "runner"},
		{"USERDOMAIN", "BUILDBOX"},
		{"COMPUTERNAME", "BUILDBOX"},
		{"LOGONSERVER", `\\BUILDBOX`},
		{"SESSIONNAME", "Console"},
		// Architecture / CPU.
		{"OS", "Windows_NT"},
		{"PROCESSOR_ARCHITECTURE", "AMD64"},
		{"PROCESSOR_IDENTIFIER", "Intel64 Family 6 Model 142 Stepping 12, GenuineIntel"},
		{"PROCESSOR_LEVEL", "6"},
		{"PROCESSOR_REVISION", "8e0c"},
		{"NUMBER_OF_PROCESSORS", "8"},
	}
	for _, tc := range cases {
		t.Setenv(tc.key, tc.value)
	}

	var c captured
	a := newAdapter(fakeExec(&c, []byte(`{"type":"result","subtype":"success","result":"ok"}`), nil, 0, nil, false))

	if _, err := a.Run(context.Background(), defaultRequest()); err != nil {
		t.Fatalf("Run: %v", err)
	}

	envMap := envSliceToMap(c.env)
	for _, tc := range cases {
		got, ok := envMap[tc.key]
		if !ok {
			t.Errorf("env key %q must be forwarded by default; child saw env=%v", tc.key, envMap)
			continue
		}
		// The buildEnv contract is "forward whatever os.Getenv returns at
		// lookup time", not "preserve every t.Setenv override exactly":
		// a handful of Windows-reserved vars (NUMBER_OF_PROCESSORS is
		// the canonical example, computed by the Session Manager from
		// physical CPU topology) ignore SetEnvironmentVariable for the
		// session, so the value buildEnv sees can differ from what
		// t.Setenv just wrote. Asserting "child env value matches
		// parent env value at the call site" still proves the
		// passthrough policy without coupling the test to those quirks.
		// The t.Setenv calls above are still doing real work â€” they
		// guarantee the key is non-empty so buildEnv does not skip it.
		want := os.Getenv(tc.key)
		if got != want {
			t.Errorf("env key %q: child got %q, parent has %q (buildEnv must passthrough verbatim)", tc.key, got, want)
		}
	}
}

// TestRun_workingDirPropagated already covered indirectly by the success
// path; this test pins it explicitly with a non-default value.
