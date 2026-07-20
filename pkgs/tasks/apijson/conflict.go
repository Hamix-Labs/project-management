package apijson

import "strings"

// Canonical domain conflict prefixes stripped for client-facing 409 detail.
const (
	TasksConflictMark    = "tasks: conflict: "
	ProjectsConflictMark = "projects: conflict: "
)

// ConflictDetail returns the trimmed suffix after the first matching mark in
// err.Error(). Empty string if err is nil or no mark matches.
//
//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func ConflictDetail(err error, marks ...string) string {
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
