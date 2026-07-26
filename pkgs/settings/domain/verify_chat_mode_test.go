package domain

import "testing"

func TestEffectiveVerifyChatMode(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name, task, settings string
		want                 VerifyChatMode
	}{
		{"defaults", "", "", VerifyChatModeSameChat},
		{"settings different", "", "different_chat", VerifyChatModeDifferentChat},
		{"task wins", "same_chat", "different_chat", VerifyChatModeSameChat},
		{"task different", "different_chat", "same_chat", VerifyChatModeDifferentChat},
		{"invalid settings fallback", "", "nope", VerifyChatModeSameChat},
		{"invalid task falls to settings", "nope", "different_chat", VerifyChatModeDifferentChat},
		{"whitespace inherit", "  ", "different_chat", VerifyChatModeDifferentChat},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := EffectiveVerifyChatMode(tc.task, tc.settings)
			if got != tc.want {
				t.Fatalf("EffectiveVerifyChatMode(%q,%q)=%q want %q", tc.task, tc.settings, got, tc.want)
			}
		})
	}
}

func TestNormalizeVerifyChatMode(t *testing.T) {
	t.Parallel()
	if got, ok := NormalizeVerifyChatMode(""); !ok || got != "" {
		t.Fatalf("empty: got %q ok=%v", got, ok)
	}
	if got, ok := NormalizeVerifyChatMode(" same_chat "); !ok || got != "same_chat" {
		t.Fatalf("same_chat: got %q ok=%v", got, ok)
	}
	if _, ok := NormalizeVerifyChatMode("fresh"); ok {
		t.Fatal("expected invalid")
	}
}
