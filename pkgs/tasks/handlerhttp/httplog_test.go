package handlerhttp

import "testing"

func TestTruncateRunes_preservesShortString(t *testing.T) {
	if got := TruncateRunes("hello", 10); got != "hello" {
		t.Fatalf("got %q", got)
	}
}

func TestTruncateRunes_truncatesWithEllipsis(t *testing.T) {
	got := TruncateRunes("abcdefghij", 4)
	if got != "abcd…" {
		t.Fatalf("got %q", got)
	}
}

func TestTruncateRunes_zeroMax(t *testing.T) {
	if got := TruncateRunes("abc", 0); got != "" {
		t.Fatalf("got %q", got)
	}
}
