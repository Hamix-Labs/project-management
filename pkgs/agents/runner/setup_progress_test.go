package runner

import "testing"

func TestSetupProgressEvent(t *testing.T) {
	ev := SetupProgressEvent(ProgressRunStateSetupStarted, "Preparing execute…")
	if ev.Kind != ProgressRunStateKind {
		t.Fatalf("kind: got %q want %q", ev.Kind, ProgressRunStateKind)
	}
	if ev.Subtype != ProgressRunStateSetupStarted {
		t.Fatalf("subtype: got %q want %q", ev.Subtype, ProgressRunStateSetupStarted)
	}
	if ev.Message != "Preparing execute…" {
		t.Fatalf("message: got %q", ev.Message)
	}
	if ev.Tool != ProgressToolHarnessSetup {
		t.Fatalf("tool: got %q want %q", ev.Tool, ProgressToolHarnessSetup)
	}
}
