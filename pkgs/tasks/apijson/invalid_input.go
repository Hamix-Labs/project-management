package apijson

import "strings"

// Canonical domain invalid-input prefixes stripped for client-facing 400 detail.
const (
	TasksInvalidInputMark    = "tasks: invalid input: "
	SettingsInvalidInputMark = "settings: invalid input: "
)

// InvalidInputDetail returns the trimmed suffix after the first matching mark in
// err.Error(). Empty string if err is nil or no mark matches.
//
// Shared here (not handlerhttp) so gitinventory can call it without an import
// cycle: handlerhttp imports gitinventory/handler for WriteGitStoreError.
//
//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func InvalidInputDetail(err error, marks ...string) string {
	if err == nil || len(marks) == 0 {
		return ""
	}
	s := err.Error()
	for _, mark := range marks {
		if mark == "" {
			continue
		}
		if i := strings.Index(s, mark); i >= 0 {
			return strings.TrimSpace(s[i+len(mark):])
		}
	}
	return ""
}
