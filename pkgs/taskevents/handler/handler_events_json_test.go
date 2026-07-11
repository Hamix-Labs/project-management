package handler

import (
	"context"
	"encoding/json"
	"errors"
	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
	taskeventsdomain "github.com/AlexsanderHamir/Hamix/pkgs/taskevents/domain"
	"net/url"
	"strings"
	"testing"
	"time"
)

func nonObjectJSONFixtures() map[string][]byte {
	return map[string][]byte{
		"nil":        nil,
		"empty":      {},
		"whitespace": []byte("  \t\n  "),
		"null":       []byte("null"),
		"string":     []byte(`"hello"`),
		"number":     []byte(`42`),
		"array":      []byte(`[1,2]`),
		"malformed":  []byte(`{not json`),
	}
}

func assertObjectMessage(t *testing.T, label string, raw json.RawMessage) {
	t.Helper()
	if !json.Valid(raw) {
		t.Fatalf("%s: invalid JSON %q", label, raw)
	}
	trimmed := json.RawMessage(raw)
	if string(trimmed) != "{}" && string(trimmed) != `{"":null}` {
		var probe map[string]json.RawMessage
		if err := json.Unmarshal(trimmed, &probe); err != nil {
			t.Fatalf("%s: not a JSON object: %v raw=%q", label, err, raw)
		}
	}
}

func TestTaskEventDetailFromDomain_normalizes_non_object_data(t *testing.T) {
	for name, raw := range nonObjectJSONFixtures() {
		t.Run(name, func(t *testing.T) {
			ev := &taskeventsdomain.TaskEvent{
				TaskID: "tsk_3",
				Seq:    1,
				At:     time.Now().UTC(),
				Type:   taskeventsdomain.EventStatusChanged,
				By:     taskcoredomain.ActorUser,
				Data:   raw,
			}
			resp := taskEventDetailFromDomain(ev, "tsk_3")
			assertObjectMessage(t, "taskEventDetailResponse.Data", resp.Data)
		})
	}
}

func TestTaskEventLines_normalizes_non_object_data(t *testing.T) {
	for name, raw := range nonObjectJSONFixtures() {
		t.Run(name, func(t *testing.T) {
			evs := []taskeventsdomain.TaskEvent{{
				TaskID: "tsk_4",
				Seq:    1,
				At:     time.Now().UTC(),
				Type:   taskeventsdomain.EventStatusChanged,
				By:     taskcoredomain.ActorUser,
				Data:   raw,
			}}
			lines := taskEventLines(evs)
			if len(lines) != 1 {
				t.Fatalf("expected 1 line, got %d", len(lines))
			}
			assertObjectMessage(t, "taskEventLine.Data", lines[0].Data)
		})
	}
}

func TestParseTaskEventsLimit_reject_overlong_limit(t *testing.T) {
	long := strings.Repeat("1", maxTaskEventSeqParamBytes+1)
	_, err := parseTaskEventsLimit(context.Background(), url.Values{"limit": {long}})
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, taskcoredomain.ErrInvalidInput) {
		t.Fatalf("got %v", err)
	}
}
