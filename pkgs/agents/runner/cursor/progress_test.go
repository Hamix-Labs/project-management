package cursor

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestToolInputSummary_progressVerbsAndGlobFields(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		tool  string
		input map[string]any
		want  string
	}{
		{
			name: "read file uses Reading",
			tool: "ReadFile",
			input: map[string]any{
				"path": "README.md",
			},
			want: "Reading README.md",
		},
		{
			name: "glob camelCase pattern",
			tool: "Glob",
			input: map[string]any{
				"globPattern":     "Makefile",
				"targetDirectory": "/repo/pkgs/agents/worker",
			},
			want: "Searching for Makefile in worker",
		},
		{
			name: "glob missing pattern does not duplicate files",
			tool: "Glob",
			input: map[string]any{
				"targetDirectory": "/repo",
			},
			want: "Searching files in repo",
		},
		{
			name:  "glob missing pattern and scope",
			tool:  "Glob",
			input: map[string]any{},
			want:  "Searching files",
		},
		{
			name: "delete uses Deleting",
			tool: "Delete",
			input: map[string]any{
				"path": "old.txt",
			},
			want: "Deleting old.txt",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			raw, err := json.Marshal(tt.input)
			if err != nil {
				t.Fatalf("marshal input: %v", err)
			}
			if got := toolInputSummary(tt.tool, raw); got != tt.want {
				t.Fatalf("toolInputSummary() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestToolProgressMessage_fallbackUsesIng(t *testing.T) {
	t.Parallel()

	if got := toolProgressMessage("ReadFile", "started", nil); got != "Starting ReadFile" {
		t.Fatalf("started fallback = %q, want Starting ReadFile", got)
	}
	if got := toolProgressMessage("ReadFile", "completed", nil); got != "Finishing ReadFile" {
		t.Fatalf("completed fallback = %q, want Finishing ReadFile", got)
	}
}

func TestProgressFromLine_assistantPersistsFullMessage(t *testing.T) {
	t.Parallel()

	full := strings.Repeat("Refactor step complete. ", 20) + "Done."
	raw, err := json.Marshal(map[string]any{
		"type": "assistant",
		"message": map[string]any{
			"role": "assistant",
			"content": []map[string]string{
				{"type": "text", "text": full},
			},
		},
	})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	ev, ok := progressFromLine(raw, nil)
	if !ok {
		t.Fatal("expected assistant progress event")
	}
	if ev.Message != full {
		t.Fatalf("Message clipped: got %d runes want %d", len([]rune(ev.Message)), len([]rune(full)))
	}
	if len(ev.Payload) == 0 {
		t.Fatal("expected Cursor-shaped payload")
	}
	var line progressEventLine
	if err := json.Unmarshal(ev.Payload, &line); err != nil {
		t.Fatalf("payload: %v", err)
	}
	if got := textContent(line.Message.Content); got != full {
		t.Fatalf("payload text = %q, want full message", got)
	}
}

func TestAssistantMessagePayload_cursorShape(t *testing.T) {
	t.Parallel()

	msg := "line one\nline two"
	payload := assistantMessagePayload(msg)
	if payload == nil {
		t.Fatal("expected payload")
	}
	var line progressEventLine
	if err := json.Unmarshal(payload, &line); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if line.Type != cursorEventAssistant {
		t.Fatalf("type = %q", line.Type)
	}
	if got := textContent(line.Message.Content); got != msg {
		t.Fatalf("text = %q, want %q", got, msg)
	}
}
