package apijson

import (
	"errors"
	"fmt"
	"testing"
)

func TestInvalidInputDetail(t *testing.T) {
	t.Parallel()

	tasksErr := fmt.Errorf("%swrong status", TasksInvalidInputMark)
	settingsErr := fmt.Errorf("%sempty model", SettingsInvalidInputMark)
	projectsErr := fmt.Errorf("%sproject name required", ProjectsInvalidInputMark)
	wrapped := fmt.Errorf("wrap: %w", fmt.Errorf("%sbad id", TasksInvalidInputMark))

	tests := []struct {
		name  string
		err   error
		marks []string
		want  string
	}{
		{name: "nil", err: nil, marks: []string{TasksInvalidInputMark}, want: ""},
		{name: "no marks", err: tasksErr, marks: nil, want: ""},
		{name: "tasks mark", err: tasksErr, marks: []string{TasksInvalidInputMark}, want: "wrong status"},
		{name: "settings preferred", err: settingsErr, marks: []string{SettingsInvalidInputMark, TasksInvalidInputMark}, want: "empty model"},
		{name: "projects preferred", err: projectsErr, marks: []string{ProjectsInvalidInputMark, SettingsInvalidInputMark, TasksInvalidInputMark}, want: "project name required"},
		{name: "tasks after settings miss", err: tasksErr, marks: []string{SettingsInvalidInputMark, TasksInvalidInputMark}, want: "wrong status"},
		{name: "no match", err: errors.New("other"), marks: []string{TasksInvalidInputMark}, want: ""},
		{name: "wrapped message still scanned", err: wrapped, marks: []string{TasksInvalidInputMark}, want: "bad id"},
		{name: "trims detail", err: errors.New(TasksInvalidInputMark + "  spaced  "), marks: []string{TasksInvalidInputMark}, want: "spaced"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := InvalidInputDetail(tt.err, tt.marks...)
			if got != tt.want {
				t.Fatalf("InvalidInputDetail() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestUserFacingMessage(t *testing.T) {
	t.Parallel()
	if got := UserFacingMessage(nil, TasksInvalidInputMark); got != "" {
		t.Fatalf("nil = %q", got)
	}
	marked := fmt.Errorf("%swrong status", TasksInvalidInputMark)
	if got := UserFacingMessage(marked, TasksInvalidInputMark); got != "wrong status" {
		t.Fatalf("marked = %q", got)
	}
	other := errors.New("plain")
	if got := UserFacingMessage(other, TasksInvalidInputMark); got != "plain" {
		t.Fatalf("fallback = %q", got)
	}
}
