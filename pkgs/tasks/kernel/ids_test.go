package kernel

import "testing"

func TestResolveID(t *testing.T) {
	t.Parallel()
	if got := ResolveID("  abc  "); got != "abc" {
		t.Fatalf("got %q want abc", got)
	}
	if got := ResolveID(""); got == "" {
		t.Fatal("expected generated id for empty input")
	}
	if got := ResolveID("   "); got == "" {
		t.Fatal("expected generated id for whitespace input")
	}
}
