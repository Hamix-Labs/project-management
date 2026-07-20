package apijson

import "strings"

// Canonical domain invalid-input prefixes stripped for client-facing 400 detail.
const (
	TasksInvalidInputMark    = "tasks: invalid input: "
	SettingsInvalidInputMark = "settings: invalid input: "
	ProjectsInvalidInputMark = "projects: invalid input: "
)

// InvalidInputDetail returns the trimmed suffix after the first matching mark in
// err.Error(). Empty string if err is nil or no mark matches.
//
// Shared here (not only in handlerhttp) so httperr and BC handlers can call it
// without importing pkgs/tasks/handlerhttp.
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

// UserFacingMessage returns InvalidInputDetail when a mark matches, otherwise
// err.Error(). Nil err yields "". Shared by repo/settings handlers (B-33).
//
//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func UserFacingMessage(err error, marks ...string) string {
	if err == nil {
		return ""
	}
	if d := InvalidInputDetail(err, marks...); d != "" {
		return d
	}
	return err.Error()
}
