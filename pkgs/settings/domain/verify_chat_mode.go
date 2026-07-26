package domain

import "strings"

// VerifyChatMode selects how PhaseVerify relates to the execute Cursor chat.
type VerifyChatMode string

const (
	// VerifyChatModeSameChat resumes the execute session (ADR-0085 default).
	VerifyChatModeSameChat VerifyChatMode = "same_chat"
	// VerifyChatModeDifferentChat uses a separate verify session chain (ADR-0031).
	VerifyChatModeDifferentChat VerifyChatMode = "different_chat"
)

// DefaultVerifyChatMode is the seed and fallback when unset or invalid.
const DefaultVerifyChatMode = VerifyChatModeSameChat

// ValidVerifyChatMode reports whether v is a known mode wire value.
//
//funclogmeasure:skip category=hot-path reason="Pure enum check without I/O."
func ValidVerifyChatMode(v string) bool {
	switch VerifyChatMode(strings.TrimSpace(v)) {
	case VerifyChatModeSameChat, VerifyChatModeDifferentChat:
		return true
	default:
		return false
	}
}

// NormalizeVerifyChatMode trims v. Empty stays empty (task inherit sentinel).
// Invalid non-empty values return ("", false).
//
//funclogmeasure:skip category=hot-path reason="Pure normalize without I/O."
func NormalizeVerifyChatMode(v string) (string, bool) {
	trimmed := strings.TrimSpace(v)
	if trimmed == "" {
		return "", true
	}
	if !ValidVerifyChatMode(trimmed) {
		return "", false
	}
	return trimmed, true
}

// EffectiveVerifyChatMode resolves task override over settings default.
// Empty task value inherits settings; empty/invalid settings fall back to same_chat.
//
//funclogmeasure:skip category=hot-path reason="Pure resolve without I/O."
func EffectiveVerifyChatMode(taskMode, settingsMode string) VerifyChatMode {
	if n, ok := NormalizeVerifyChatMode(taskMode); ok && n != "" {
		return VerifyChatMode(n)
	}
	if n, ok := NormalizeVerifyChatMode(settingsMode); ok && n != "" {
		return VerifyChatMode(n)
	}
	return DefaultVerifyChatMode
}
