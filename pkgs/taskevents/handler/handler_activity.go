package handler

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/AlexsanderHamir/Hamix/pkgs/obs/calltrace"
	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
	"github.com/AlexsanderHamir/Hamix/pkgs/taskevents/contract"
	taskeventsdomain "github.com/AlexsanderHamir/Hamix/pkgs/taskevents/domain"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/handlerhttp"
)

type activityEntry struct {
	TaskID        string                     `json:"task_id"`
	Seq           int64                      `json:"seq"`
	At            time.Time                  `json:"at"`
	Type          taskeventsdomain.EventType `json:"type"`
	By            taskcoredomain.Actor       `json:"by"`
	Data          json.RawMessage            `json:"data"`
	TaskTitle     *string                    `json:"task_title,omitempty"`
	TaskNumber    *int                       `json:"task_number,omitempty"`
	TaskPriority  *taskcoredomain.Priority   `json:"task_priority,omitempty"`
	TaskProjectID *string                    `json:"task_project_id,omitempty"`
	TaskTags      []string                   `json:"task_tags,omitempty"`
}

type activityResponse struct {
	Total  int64           `json:"total"`
	Limit  int             `json:"limit"`
	Offset int             `json:"offset"`
	Events []activityEntry `json:"events"`
}

func (h *Handler) taskActivity(w http.ResponseWriter, r *http.Request) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "taskevents.handler.taskActivity")
	const op = "tasks.activity"
	r = calltrace.WithRequestRoot(r, op)
	in, err := parseActivityQuery(r.URL.Query())
	if err != nil {
		handlerhttp.WriteStoreError(w, r, op, err)
		return
	}
	handlerhttp.DebugHTTPRequest(r, op, "limit", in.Limit, "offset", in.Offset, "since", in.Since)
	out, err := h.activity.ListTaskActivity(r.Context(), in)
	if err != nil {
		handlerhttp.WriteStoreError(w, r, op, err)
		return
	}
	handlerhttp.WriteJSON(w, r, op, http.StatusOK, activityResponse{
		Total:  out.Total,
		Limit:  out.Limit,
		Offset: out.Offset,
		Events: activityEntriesToJSON(out.Events),
	})
}

//funclogmeasure:skip category=hot-path reason="Pure DTO mapper without I/O; operation trace is emitted by the calling chokepoint."
func activityEntriesToJSON(events []contract.ActivityEvent) []activityEntry {
	out := make([]activityEntry, 0, len(events))
	for _, e := range events {
		tags := e.TaskTags
		if tags == nil {
			tags = []string{}
		}
		entry := activityEntry{
			TaskID:        e.TaskID,
			Seq:           e.Seq,
			At:            e.At,
			Type:          e.Type,
			By:            e.By,
			Data:          normalizeJSONObjectForResponse(e.Data),
			TaskTitle:     e.TaskTitle,
			TaskNumber:    e.TaskNumber,
			TaskPriority:  e.TaskPriority,
			TaskProjectID: e.TaskProjectID,
		}
		if len(tags) > 0 {
			entry.TaskTags = tags
		}
		out = append(out, entry)
	}
	return out
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func parseActivityQuery(q url.Values) (contract.ListActivityInput, error) {
	in := contract.ListActivityInput{
		Limit:  50,
		Offset: 0,
	}
	if v := q.Get("limit"); v != "" {
		if len(v) > maxTaskEventSeqParamBytes {
			return contract.ListActivityInput{}, fmt.Errorf("%w: limit value too long", taskcoredomain.ErrInvalidInput)
		}
		n, e := strconv.Atoi(v)
		if e != nil || n < 1 || n > 200 {
			return contract.ListActivityInput{}, fmt.Errorf("%w: limit must be integer 1..200", taskcoredomain.ErrInvalidInput)
		}
		in.Limit = n
	}
	if v := q.Get("offset"); v != "" {
		if len(v) > maxTaskEventSeqParamBytes {
			return contract.ListActivityInput{}, fmt.Errorf("%w: offset value too long", taskcoredomain.ErrInvalidInput)
		}
		n, e := strconv.Atoi(v)
		if e != nil || n < 0 {
			return contract.ListActivityInput{}, fmt.Errorf("%w: offset must be non-negative integer", taskcoredomain.ErrInvalidInput)
		}
		in.Offset = n
	}
	if v := q.Get("since"); v != "" {
		if len(v) > 64 {
			return contract.ListActivityInput{}, fmt.Errorf("%w: since value too long", taskcoredomain.ErrInvalidInput)
		}
		t, e := time.Parse(time.RFC3339, v)
		if e != nil {
			return contract.ListActivityInput{}, fmt.Errorf("%w: since must be RFC3339 timestamp", taskcoredomain.ErrInvalidInput)
		}
		in.Since = &t
	}
	return in, nil
}
