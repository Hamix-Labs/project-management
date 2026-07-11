package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	taskeventsstore "github.com/AlexsanderHamir/Hamix/pkgs/taskevents/store"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/apijson"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/calltrace"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/domain"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/logctx"
)

const (
	maxPathIDBytes             = 128
	maxHTTPLogJSONPreviewBytes = 16384
	maxHTTPLogTextRunes        = 240
	maxTaskEventSeqParamBytes  = 32
)

var jsonObjectMessageEmpty = json.RawMessage(`{}`)

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func decodeJSON(ctx context.Context, r io.Reader, dst any) error {
	dec := json.NewDecoder(r)
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		return fmt.Errorf("json decode: %w", err)
	}
	if err := dec.Decode(&struct{}{}); err != nil {
		if err == io.EOF {
			return nil
		}
		return fmt.Errorf("json trailing data: %w", err)
	}
	return fmt.Errorf("%w: json trailing data", domain.ErrInvalidInput)
}

//funclogmeasure:skip category=delegate-already-logs reason="JSON response helper; HTTP handler chokepoint emits trace."
func writeJSON(w http.ResponseWriter, r *http.Request, op string, code int, v any) {
	apijson.ApplySecurityHeaders(w)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(v); err != nil {
		writeJSONError(w, r, op, http.StatusInternalServerError, "internal server error")
		return
	}
	payload := bytes.TrimSuffix(buf.Bytes(), []byte("\n"))
	w.WriteHeader(code)
	_, _ = w.Write(payload)
	_, _ = w.Write([]byte("\n"))
}

//funclogmeasure:skip category=delegate-already-logs reason="Error response helper; HTTP handler chokepoint emits trace."
func writeJSONError(w http.ResponseWriter, r *http.Request, op string, code int, msg string) {
	apijson.WriteJSONError(w, r, op, code, msg, calltrace.Path)
}

//funclogmeasure:skip category=delegate-already-logs reason="Error response helper; HTTP handler chokepoint emits trace."
func writeError(w http.ResponseWriter, r *http.Request, op string, err error, code int) {
	msg := http.StatusText(code)
	if code == http.StatusBadRequest {
		msg = userFacingJSONError(err)
		if msg == "" {
			msg = "bad request"
		}
	}
	writeJSONError(w, r, op, code, msg)
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func userFacingJSONError(err error) string {
	s := err.Error()
	if strings.HasPrefix(s, "json decode: ") {
		return strings.TrimPrefix(s, "json decode: ")
	}
	if errors.Is(err, domain.ErrInvalidInput) {
		return "request body must contain a single JSON value"
	}
	if strings.HasPrefix(s, "json trailing data:") {
		return "request body must contain a single JSON value"
	}
	return s
}

//funclogmeasure:skip category=delegate-already-logs reason="Error response helper; HTTP handler chokepoint emits trace."
func writeStoreError(w http.ResponseWriter, r *http.Request, op string, err error) {
	code, msg := storeErrHTTP(err)
	if code >= 500 {
		slog.Log(r.Context(), slog.LevelError, "request failed",
			"cmd", calltrace.LogCmd, "operation", op,
			"request_id", logctx.RequestIDFromContext(r.Context()),
			"http_status", code, "err", err,
		)
	} else {
		slog.Log(r.Context(), slog.LevelWarn, "request failed",
			"cmd", calltrace.LogCmd, "operation", op,
			"request_id", logctx.RequestIDFromContext(r.Context()),
			"http_status", code, "err", err,
		)
	}
	writeJSONError(w, r, op, code, msg)
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func storeErrHTTP(err error) (code int, msg string) {
	code = http.StatusInternalServerError
	msg = "internal server error"
	switch {
	case errors.Is(err, context.DeadlineExceeded):
		return http.StatusGatewayTimeout, "request timed out"
	case errors.Is(err, context.Canceled):
		return http.StatusRequestTimeout, "request canceled"
	case errors.Is(err, domain.ErrNotFound):
		return http.StatusNotFound, "not found"
	case errors.Is(err, domain.ErrInvalidInput):
		return http.StatusBadRequest, invalidInputDetail(err)
	case errors.Is(err, domain.ErrConflict):
		return http.StatusConflict, conflictDetail(err)
	default:
		return code, msg
	}
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func invalidInputDetail(err error) string {
	s := err.Error()
	const mark = "tasks: invalid input: "
	if i := strings.Index(s, mark); i >= 0 {
		return strings.TrimSpace(s[i+len(mark):])
	}
	return s
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func conflictDetail(err error) string {
	s := err.Error()
	const mark = "tasks: conflict: "
	if i := strings.Index(s, mark); i >= 0 {
		return strings.TrimSpace(s[i+len(mark):])
	}
	return s
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func parseTaskPathID(id string) (string, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return "", fmt.Errorf("%w: id", domain.ErrInvalidInput)
	}
	if len(id) > maxPathIDBytes {
		return "", fmt.Errorf("%w: id too long", domain.ErrInvalidInput)
	}
	return id, nil
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func actorFromRequest(r *http.Request) domain.Actor {
	if r == nil {
		return domain.ActorUser
	}
	switch strings.ToLower(strings.TrimSpace(r.Header.Get("X-Actor"))) {
	case "agent":
		return domain.ActorAgent
	default:
		return domain.ActorUser
	}
}

func debugHTTPRequest(r *http.Request, op string, extra ...any) {
	if r == nil || !slog.Default().Enabled(r.Context(), slog.LevelDebug) {
		return
	}
	q := r.URL.RawQuery
	args := []any{
		"cmd", calltrace.LogCmd,
		"obs_category", "http_io",
		"operation", op,
		"call_path", calltrace.Path(r.Context()),
		"phase", "in",
		"method", r.Method,
		"path", r.URL.Path,
		"route_pattern", r.Pattern,
		"query", q,
		"content_length", r.ContentLength,
		"x_actor", strings.TrimSpace(r.Header.Get("X-Actor")),
	}
	args = append(args, extra...)
	slog.Log(r.Context(), slog.LevelDebug, "http.io", args...)
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func truncateRunes(s string, maxRunes int) string {
	if maxRunes <= 0 {
		return ""
	}
	var b strings.Builder
	n := 0
	for _, r := range s {
		if n >= maxRunes {
			b.WriteString("…")
			break
		}
		b.WriteRune(r)
		n++
	}
	return b.String()
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

type taskEventLine struct {
	Seq            int64                        `json:"seq"`
	At             time.Time                    `json:"at"`
	Type           domain.EventType             `json:"type"`
	By             domain.Actor                 `json:"by"`
	Data           json.RawMessage              `json:"data"`
	UserResponse   *string                      `json:"user_response,omitempty"`
	UserResponseAt *time.Time                   `json:"user_response_at,omitempty"`
	ResponseThread []domain.ResponseThreadEntry `json:"response_thread,omitempty"`
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
	TaskID         string                       `json:"task_id"`
	Seq            int64                        `json:"seq"`
	At             time.Time                    `json:"at"`
	Type           domain.EventType             `json:"type"`
	By             domain.Actor                 `json:"by"`
	Data           json.RawMessage              `json:"data"`
	UserResponse   *string                      `json:"user_response,omitempty"`
	UserResponseAt *time.Time                   `json:"user_response_at,omitempty"`
	ResponseThread []domain.ResponseThreadEntry `json:"response_thread,omitempty"`
}

type taskEventUserResponseJSON struct {
	UserResponse string `json:"user_response"`
}

func taskEventDetailFromDomain(ev *domain.TaskEvent, taskID string) taskEventDetailResponse {
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

func taskEventLines(evs []domain.TaskEvent) []taskEventLine {
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

func (h *Handler) taskEvent(w http.ResponseWriter, r *http.Request) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "taskevents.handler.taskEvent")
	const op = "tasks.event"
	r = calltrace.WithRequestRoot(r, op)
	id, err := parseTaskPathID(r.PathValue("id"))
	if err != nil {
		writeStoreError(w, r, op, err)
		return
	}
	seqStr := strings.TrimSpace(r.PathValue("seq"))
	if len(seqStr) > maxTaskEventSeqParamBytes {
		writeError(w, r, op, errors.New("seq too long"), http.StatusBadRequest)
		return
	}
	seq, err := strconv.ParseInt(seqStr, 10, 64)
	if err != nil || seq < 1 {
		debugHTTPRequest(r, op, "task_id", id, "seq_param", seqStr, "seq_parse_failed", true)
		writeError(w, r, op, errors.New("seq must be a positive integer"), http.StatusBadRequest)
		return
	}
	debugHTTPRequest(r, op, "task_id", id, "seq", seq)
	ev, err := h.events.GetTaskEvent(r.Context(), id, seq)
	if err != nil {
		writeStoreError(w, r, op, err)
		return
	}
	writeJSON(w, r, op, http.StatusOK, taskEventDetailFromDomain(ev, id))
}

func (h *Handler) taskEvents(w http.ResponseWriter, r *http.Request) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "taskevents.handler.taskEvents")
	const op = "tasks.events"
	r = calltrace.WithRequestRoot(r, op)
	id, err := parseTaskPathID(r.PathValue("id"))
	if err != nil {
		writeStoreError(w, r, op, err)
		return
	}
	debugHTTPRequest(r, op, "task_id", id)
	if _, err := h.tasks.Get(r.Context(), id); err != nil {
		writeStoreError(w, r, op, err)
		return
	}
	pending, err := h.events.ApprovalPending(r.Context(), id)
	if err != nil {
		writeStoreError(w, r, op, err)
		return
	}
	q := r.URL.Query()
	if q.Get("offset") != "" {
		writeStoreError(w, r, op, fmt.Errorf("%w: offset is not supported for task events; use before_seq or after_seq", domain.ErrInvalidInput))
		return
	}
	if q.Get("limit") == "" && q.Get("before_seq") == "" && q.Get("after_seq") == "" {
		h.writeTaskEventsFullList(w, r, op, id, pending)
		return
	}
	h.writeTaskEventsCursorPage(w, r, op, id, pending, q)
}

func (h *Handler) writeTaskEventsFullList(w http.ResponseWriter, r *http.Request, op, id string, pending bool) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "taskevents.handler.writeTaskEventsFullList")
	evs, err := h.events.ListTaskEvents(r.Context(), id)
	if err != nil {
		writeStoreError(w, r, op, err)
		return
	}
	writeJSON(w, r, op, http.StatusOK, taskEventsResponse{
		TaskID:          id,
		Events:          taskEventLines(evs),
		ApprovalPending: pending,
	})
}

func (h *Handler) writeTaskEventsCursorPage(w http.ResponseWriter, r *http.Request, op, id string, pending bool, q url.Values) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "taskevents.handler.writeTaskEventsCursorPage")
	beforeStr := strings.TrimSpace(q.Get("before_seq"))
	afterStr := strings.TrimSpace(q.Get("after_seq"))
	if beforeStr != "" && afterStr != "" {
		writeError(w, r, op, errors.New("before_seq and after_seq cannot both be set"), http.StatusBadRequest)
		return
	}
	if (beforeStr != "" && len(beforeStr) > maxTaskEventSeqParamBytes) || (afterStr != "" && len(afterStr) > maxTaskEventSeqParamBytes) {
		writeError(w, r, op, errors.New("before_seq or after_seq too long"), http.StatusBadRequest)
		return
	}
	limit, err := parseTaskEventsLimit(r.Context(), q)
	if err != nil {
		writeStoreError(w, r, op, err)
		return
	}
	var beforeSeq, afterSeq *int64
	if beforeStr != "" {
		n, err := strconv.ParseInt(beforeStr, 10, 64)
		if err != nil || n < 1 {
			writeError(w, r, op, errors.New("before_seq must be a positive integer"), http.StatusBadRequest)
			return
		}
		beforeSeq = &n
	}
	if afterStr != "" {
		n, err := strconv.ParseInt(afterStr, 10, 64)
		if err != nil || n < 1 {
			writeError(w, r, op, errors.New("after_seq must be a positive integer"), http.StatusBadRequest)
			return
		}
		afterSeq = &n
	}
	page, err := h.events.ListTaskEventsPageCursor(r.Context(), id, limit, beforeSeq, afterSeq)
	if err != nil {
		writeStoreError(w, r, op, err)
		return
	}
	lim := limit
	tot := page.Total
	resp := taskEventsResponse{
		TaskID:          id,
		Events:          taskEventLines(page.Events),
		Limit:           &lim,
		Total:           &tot,
		HasMoreNewer:    page.HasMoreNewer,
		HasMoreOlder:    page.HasMoreOlder,
		ApprovalPending: pending,
	}
	if len(page.Events) > 0 {
		rs := page.RangeStart
		re := page.RangeEnd
		resp.RangeStart = &rs
		resp.RangeEnd = &re
	}
	writeJSON(w, r, op, http.StatusOK, resp)
}

func (h *Handler) patchTaskEventUserResponse(w http.ResponseWriter, r *http.Request) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "taskevents.handler.patchTaskEventUserResponse")
	const op = "tasks.event.user_response"
	r = calltrace.WithRequestRoot(r, op)
	id, err := parseTaskPathID(r.PathValue("id"))
	if err != nil {
		writeStoreError(w, r, op, err)
		return
	}
	seqStr := strings.TrimSpace(r.PathValue("seq"))
	if len(seqStr) > maxTaskEventSeqParamBytes {
		writeError(w, r, op, errors.New("seq too long"), http.StatusBadRequest)
		return
	}
	seq, err := strconv.ParseInt(seqStr, 10, 64)
	if err != nil || seq < 1 {
		debugHTTPRequest(r, op, "task_id", id, "seq_param", seqStr, "seq_parse_failed", true)
		writeError(w, r, op, errors.New("seq must be a positive integer"), http.StatusBadRequest)
		return
	}
	var body taskEventUserResponseJSON
	if err := decodeJSON(r.Context(), r.Body, &body); err != nil {
		debugHTTPRequest(r, op, "task_id", id, "seq", seq, "json_decode_failed", true)
		writeError(w, r, op, err, http.StatusBadRequest)
		return
	}
	debugHTTPRequest(r, op, "task_id", id, "seq", seq,
		"user_response_len", len(body.UserResponse),
		"user_response_preview", truncateRunes(body.UserResponse, maxHTTPLogTextRunes),
	)
	by := actorFromRequest(r)
	if err := h.events.AppendTaskEventResponseMessage(r.Context(), id, seq, body.UserResponse, by); err != nil {
		writeStoreError(w, r, op, err)
		return
	}
	ev, err := h.events.GetTaskEvent(r.Context(), id, seq)
	if err != nil {
		writeStoreError(w, r, op, err)
		return
	}
	if h.notifyTaskEventChanged != nil {
		h.notifyTaskEventChanged(id, seq)
	}
	writeJSON(w, r, op, http.StatusOK, taskEventDetailFromDomain(ev, id))
}

func parseTaskEventsLimit(ctx context.Context, q url.Values) (limit int, err error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "taskevents.handler.parseTaskEventsLimit")
	ctx = calltrace.Push(ctx, "parseTaskEventsLimit")
	calltrace.HelperIOIn(ctx, "parseTaskEventsLimit", "limit_q", q.Get("limit"), "before_seq_q", q.Get("before_seq"), "after_seq_q", q.Get("after_seq"))
	defer func() { calltrace.HelperIOOut(ctx, "parseTaskEventsLimit", "limit", limit, "err", err) }()
	limit = 50
	if v := q.Get("limit"); v != "" {
		if len(v) > maxTaskEventSeqParamBytes {
			return 0, fmt.Errorf("%w: limit too long", domain.ErrInvalidInput)
		}
		n, e := strconv.Atoi(v)
		if e != nil || n < 0 || n > 200 {
			return 0, fmt.Errorf("%w: limit must be integer 0..200", domain.ErrInvalidInput)
		}
		limit = n
	}
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}
	return limit, nil
}
