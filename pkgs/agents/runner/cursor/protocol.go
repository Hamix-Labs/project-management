package cursor

import "github.com/AlexsanderHamir/Hamix/pkgs/obs/calltrace"
import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"
)

type cursorOutput struct {
	Type                  string          `json:"type,omitempty"`
	Subtype               string          `json:"subtype,omitempty"`
	IsError               bool            `json:"is_error,omitempty"`
	Result                string          `json:"result,omitempty"`
	DurationMs            int64           `json:"duration_ms,omitempty"`
	DurationAPIMs         int64           `json:"duration_api_ms,omitempty"`
	SessionID             string          `json:"session_id,omitempty"`
	RequestID             string          `json:"request_id,omitempty"`
	Usage                 json.RawMessage `json:"usage,omitempty"`
	ResolvedModel         string          `json:"-"`
	MissingTerminalResult bool            `json:"-"`
}

type streamEventHead struct {
	Type      string          `json:"type,omitempty"`
	Subtype   string          `json:"subtype,omitempty"`
	Model     string          `json:"model,omitempty"`
	CallID    string          `json:"call_id,omitempty"`
	SessionID string          `json:"session_id,omitempty"`
	Message   progressMessage `json:"message,omitempty"`
}

func parseStdout(stdout []byte) (cursorOutput, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "cursor.parseStdout", "bytes", len(stdout))
	stdout = bytes.TrimSpace(stdout)
	if len(stdout) == 0 {
		return cursorOutput{}, errors.New("empty stdout")
	}

	var single cursorOutput
	if err := json.Unmarshal(stdout, &single); err == nil && single.Type == cursorEventResult {
		return single, nil
	}

	var (
		out                cursorOutput
		gotResult          bool
		lastDecErr         error
		lastAssistantText  string
		lastSessionID      string
		openToolCalls      = map[string]struct{}{}
		openAnonymousTools int
	)
	for _, raw := range splitNDJSON(stdout) {
		if len(raw) == 0 {
			continue
		}
		var head streamEventHead
		if err := json.Unmarshal(raw, &head); err != nil {
			lastDecErr = err
			continue
		}
		switch head.Type {
		case cursorEventSystem:
			if head.Subtype == cursorSubtypeInit && out.ResolvedModel == "" {
				out.ResolvedModel = strings.TrimSpace(head.Model)
			}
			if lastSessionID == "" {
				lastSessionID = strings.TrimSpace(head.SessionID)
			}
		case cursorEventAssistant:
			if msg := strings.TrimSpace(textContent(head.Message.Content)); msg != "" {
				lastAssistantText = msg
			}
			if lastSessionID == "" {
				lastSessionID = strings.TrimSpace(head.SessionID)
			}
		case cursorEventToolCall:
			updateOpenToolCalls(openToolCalls, &openAnonymousTools, head)
			if lastSessionID == "" {
				lastSessionID = strings.TrimSpace(head.SessionID)
			}
		case cursorEventResult:
			var evt cursorOutput
			if err := json.Unmarshal(raw, &evt); err != nil {
				lastDecErr = err
				continue
			}
			resolved := out.ResolvedModel
			out = evt
			out.ResolvedModel = resolved
			gotResult = true
		}
	}

	if !gotResult {
		if lastDecErr != nil {
			return cursorOutput{}, fmt.Errorf("decode stdout: %w", lastDecErr)
		}
		if open := openToolCallCount(openToolCalls, openAnonymousTools); open > 0 {
			return cursorOutput{}, fmt.Errorf("stream-json: no terminal result event; %d open tool call(s)", open)
		}
		if lastAssistantText != "" {
			return cursorOutput{
				Type:                  cursorEventResult,
				Subtype:               cursorSubtypeSuccess,
				Result:                lastAssistantText,
				SessionID:             lastSessionID,
				ResolvedModel:         out.ResolvedModel,
				MissingTerminalResult: true,
			}, nil
		}
		return cursorOutput{}, errors.New("stream-json: no terminal result event")
	}
	// ADR-0031: keep the earliest observed session_id when the terminal
	// result event omits it, so success details always carry the id
	// captured from the init frame.
	if strings.TrimSpace(out.SessionID) == "" && lastSessionID != "" {
		out.SessionID = lastSessionID
	}
	return out, nil
}

// sessionIDFromLine returns the trimmed session_id from a single NDJSON
// stream event line, or "" when the line is blank, malformed, or has no
// session_id field. Pure helper; safe on partial buffers.
//
//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by callers."
func sessionIDFromLine(raw []byte) string {
	raw = bytes.TrimSpace(raw)
	if len(raw) == 0 {
		return ""
	}
	var head streamEventHead
	if err := json.Unmarshal(raw, &head); err != nil {
		return ""
	}
	return strings.TrimSpace(head.SessionID)
}

// sessionIDFromStdout scans NDJSON stdout and returns the first non-empty
// session_id observed on any stream-json event, or "" when none exists.
// Pure helper reusing splitNDJSON; safe on partial or malformed buffers.
//
//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by callers."
func sessionIDFromStdout(stdout []byte) string {
	if len(stdout) == 0 {
		return ""
	}
	for _, raw := range splitNDJSON(stdout) {
		if id := sessionIDFromLine(raw); id != "" {
			return id
		}
	}
	return ""
}

func updateOpenToolCalls(open map[string]struct{}, openAnonymous *int, head streamEventHead) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "cursor.updateOpenToolCalls",
		"subtype", head.Subtype, "call_id", head.CallID)
	callID := strings.TrimSpace(head.CallID)
	switch head.Subtype {
	case cursorSubtypeStarted, cursorSubtypeStart:
		if callID == "" {
			*openAnonymous = *openAnonymous + 1
			return
		}
		open[callID] = struct{}{}
	case cursorSubtypeCompleted, cursorSubtypeSuccess, cursorSubtypeDone, cursorSubtypeFailed, cursorSubtypeError:
		if callID == "" {
			if *openAnonymous > 0 {
				*openAnonymous = *openAnonymous - 1
			}
			return
		}
		delete(open, callID)
	}
}

func openToolCallCount(open map[string]struct{}, anonymous int) int {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "cursor.openToolCallCount",
		"open", len(open), "anonymous", anonymous)
	return len(open) + anonymous
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func splitNDJSON(b []byte) [][]byte {
	if len(b) == 0 {
		return nil
	}
	out := make([][]byte, 0, 8)
	start := 0
	for i := 0; i < len(b); i++ {
		if b[i] != '\n' {
			continue
		}
		line := b[start:i]
		if len(line) > 0 && line[len(line)-1] == '\r' {
			line = line[:len(line)-1]
		}
		line = bytes.TrimSpace(line)
		if len(line) > 0 {
			out = append(out, line)
		}
		start = i + 1
	}
	if start < len(b) {
		tail := bytes.TrimSpace(b[start:])
		if len(tail) > 0 {
			out = append(out, tail)
		}
	}
	return out
}
