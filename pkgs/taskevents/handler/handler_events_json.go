package handler

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"time"

	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
	taskeventsdomain "github.com/AlexsanderHamir/Hamix/pkgs/taskevents/domain"
	taskeventsstore "github.com/AlexsanderHamir/Hamix/pkgs/taskevents/store"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/calltrace"
)

var jsonObjectMessageEmpty = json.RawMessage(`{}`)

type taskEventLine struct {
	Seq            int64                                  `json:"seq"`
	At             time.Time                              `json:"at"`
	Type           taskeventsdomain.EventType             `json:"type"`
	By             taskcoredomain.Actor                   `json:"by"`
	Data           json.RawMessage                        `json:"data"`
	UserResponse   *string                                `json:"user_response,omitempty"`
	UserResponseAt *time.Time                             `json:"user_response_at,omitempty"`
	ResponseThread []taskeventsdomain.ResponseThreadEntry `json:"response_thread,omitempty"`
}

type taskEventsResponse struct {
	TaskID          string          `json:"task_id"`
	Events          []taskEventLine `json:"events"`
	Limit           *int            `json:"limit,omitempty"`
	Total           *int64          `json:"total,omitempty"`
	RangeStart      *int64          `json:"range_start,omitempty"`
	RangeEnd        *int64          `json:"range_end,omitempty"`
	HasMoreNewer    bool            `json:"has_more_newer"`
	HasMoreOlder    bool            `json:"has_more_older"`
	ApprovalPending bool            `json:"approval_pending"`
}

type taskEventDetailResponse struct {
	TaskID         string                                 `json:"task_id"`
	Seq            int64                                  `json:"seq"`
	At             time.Time                              `json:"at"`
	Type           taskeventsdomain.EventType             `json:"type"`
	By             taskcoredomain.Actor                   `json:"by"`
	Data           json.RawMessage                        `json:"data"`
	UserResponse   *string                                `json:"user_response,omitempty"`
	UserResponseAt *time.Time                             `json:"user_response_at,omitempty"`
	ResponseThread []taskeventsdomain.ResponseThreadEntry `json:"response_thread,omitempty"`
}

type taskEventUserResponseJSON struct {
	UserResponse string `json:"user_response"`
}

func taskEventDetailFromDomain(ev *taskeventsdomain.TaskEvent, taskID string) taskEventDetailResponse {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "taskevents.handler.taskEventDetailFromDomain")
	data := normalizeJSONObjectForResponse(ev.Data)
	resp := taskEventDetailResponse{
		TaskID:         taskID,
		Seq:            ev.Seq,
		At:             ev.At,
		Type:           ev.Type,
		By:             ev.By,
		Data:           data,
		UserResponse:   ev.UserResponse,
		UserResponseAt: ev.UserResponseAt,
	}
	if th := taskeventsstore.ThreadEntriesForDisplay(ev); len(th) > 0 {
		resp.ResponseThread = th
	}
	return resp
}

func taskEventLines(evs []taskeventsdomain.TaskEvent) []taskEventLine {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "taskevents.handler.taskEventLines")
	out := make([]taskEventLine, 0, len(evs))
	for _, e := range evs {
		data := normalizeJSONObjectForResponse(e.Data)
		line := taskEventLine{
			Seq:            e.Seq,
			At:             e.At,
			Type:           e.Type,
			By:             e.By,
			Data:           data,
			UserResponse:   e.UserResponse,
			UserResponseAt: e.UserResponseAt,
		}
		if th := taskeventsstore.ThreadEntriesForDisplay(&e); len(th) > 0 {
			line.ResponseThread = th
		}
		out = append(out, line)
	}
	return out
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func normalizeJSONObjectForResponse(raw []byte) json.RawMessage {
	trimmed := bytes.TrimSpace(raw)
	if len(trimmed) == 0 || bytes.Equal(trimmed, []byte("null")) {
		return jsonObjectMessageEmpty
	}
	var probe map[string]json.RawMessage
	if err := json.Unmarshal(trimmed, &probe); err != nil {
		return jsonObjectMessageEmpty
	}
	return json.RawMessage(trimmed)
}
