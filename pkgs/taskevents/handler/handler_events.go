package handler

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
	"github.com/AlexsanderHamir/Hamix/pkgs/obs/calltrace"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/handler/readpolicy"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/handlerhttp"
)

func (h *Handler) taskEvent(w http.ResponseWriter, r *http.Request) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "taskevents.handler.taskEvent")
	const op = "tasks.event"
	r = calltrace.WithRequestRoot(r, op)
	id, err := parseTaskPathID(r.PathValue("id"))
	if err != nil {
		handlerhttp.WriteStoreError(w, r, op, err)
		return
	}
	seqStr := strings.TrimSpace(r.PathValue("seq"))
	if len(seqStr) > maxTaskEventSeqParamBytes {
		handlerhttp.WriteError(w, r, op, errors.New("seq too long"), http.StatusBadRequest)
		return
	}
	seq, err := strconv.ParseInt(seqStr, 10, 64)
	if err != nil || seq < 1 {
		debugHTTPRequest(r, op, "task_id", id, "seq_param", seqStr, "seq_parse_failed", true)
		handlerhttp.WriteError(w, r, op, errors.New("seq must be a positive integer"), http.StatusBadRequest)
		return
	}
	debugHTTPRequest(r, op, "task_id", id, "seq", seq)
	ev, err := h.events.GetTaskEvent(r.Context(), id, seq)
	if err != nil {
		handlerhttp.WriteStoreError(w, r, op, err)
		return
	}
	handlerhttp.WriteJSON(w, r, op, http.StatusOK, taskEventDetailFromDomain(ev, id))
}

func (h *Handler) taskEvents(w http.ResponseWriter, r *http.Request) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "taskevents.handler.taskEvents")
	const op = "tasks.events"
	r = calltrace.WithRequestRoot(r, op)
	id, err := parseTaskPathID(r.PathValue("id"))
	if err != nil {
		handlerhttp.WriteStoreError(w, r, op, err)
		return
	}
	debugHTTPRequest(r, op, "task_id", id)
	if _, err := h.tasks.Get(r.Context(), id); err != nil {
		handlerhttp.WriteStoreError(w, r, op, err)
		return
	}
	pending, err := h.events.ApprovalPending(r.Context(), id)
	if err != nil {
		handlerhttp.WriteStoreError(w, r, op, err)
		return
	}
	q := r.URL.Query()
	if q.Get("offset") != "" {
		handlerhttp.WriteStoreError(w, r, op, fmt.Errorf("%w: offset is not supported for task events; use before_seq or after_seq", taskcoredomain.ErrInvalidInput))
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
		handlerhttp.WriteStoreError(w, r, op, err)
		return
	}
	handlerhttp.WriteJSON(w, r, op, http.StatusOK, taskEventsResponse{
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
		handlerhttp.WriteError(w, r, op, errors.New("before_seq and after_seq cannot both be set"), http.StatusBadRequest)
		return
	}
	if (beforeStr != "" && len(beforeStr) > maxTaskEventSeqParamBytes) || (afterStr != "" && len(afterStr) > maxTaskEventSeqParamBytes) {
		handlerhttp.WriteError(w, r, op, errors.New("before_seq or after_seq too long"), http.StatusBadRequest)
		return
	}
	limit, err := parseTaskEventsLimit(r.Context(), q)
	if err != nil {
		handlerhttp.WriteStoreError(w, r, op, err)
		return
	}
	var beforeSeq, afterSeq *int64
	if beforeStr != "" {
		n, err := strconv.ParseInt(beforeStr, 10, 64)
		if err != nil || n < 1 {
			handlerhttp.WriteError(w, r, op, errors.New("before_seq must be a positive integer"), http.StatusBadRequest)
			return
		}
		beforeSeq = &n
	}
	if afterStr != "" {
		n, err := strconv.ParseInt(afterStr, 10, 64)
		if err != nil || n < 1 {
			handlerhttp.WriteError(w, r, op, errors.New("after_seq must be a positive integer"), http.StatusBadRequest)
			return
		}
		afterSeq = &n
	}
	page, err := h.events.ListTaskEventsPageCursor(r.Context(), id, limit, beforeSeq, afterSeq)
	if err != nil {
		handlerhttp.WriteStoreError(w, r, op, err)
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
	handlerhttp.WriteJSON(w, r, op, http.StatusOK, resp)
}

func (h *Handler) patchTaskEventUserResponse(w http.ResponseWriter, r *http.Request) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "taskevents.handler.patchTaskEventUserResponse")
	const op = "tasks.event.user_response"
	r = calltrace.WithRequestRoot(r, op)
	id, err := parseTaskPathID(r.PathValue("id"))
	if err != nil {
		handlerhttp.WriteStoreError(w, r, op, err)
		return
	}
	seqStr := strings.TrimSpace(r.PathValue("seq"))
	if len(seqStr) > maxTaskEventSeqParamBytes {
		handlerhttp.WriteError(w, r, op, errors.New("seq too long"), http.StatusBadRequest)
		return
	}
	seq, err := strconv.ParseInt(seqStr, 10, 64)
	if err != nil || seq < 1 {
		debugHTTPRequest(r, op, "task_id", id, "seq_param", seqStr, "seq_parse_failed", true)
		handlerhttp.WriteError(w, r, op, errors.New("seq must be a positive integer"), http.StatusBadRequest)
		return
	}
	var body taskEventUserResponseJSON
	if err := handlerhttp.DecodeJSON(r.Context(), r.Body, &body); err != nil {
		debugHTTPRequest(r, op, "task_id", id, "seq", seq, "json_decode_failed", true)
		handlerhttp.WriteError(w, r, op, err, http.StatusBadRequest)
		return
	}
	debugHTTPRequest(r, op, "task_id", id, "seq", seq,
		"user_response_len", len(body.UserResponse),
		"user_response_preview", truncateRunes(body.UserResponse, maxHTTPLogTextRunes),
	)
	by := handlerhttp.ActorFromRequest(r)
	if err := h.events.AppendTaskEventResponseMessage(r.Context(), id, seq, body.UserResponse, by); err != nil {
		handlerhttp.WriteStoreError(w, r, op, err)
		return
	}
	ev, err := h.events.GetTaskEvent(r.Context(), id, seq)
	if err != nil {
		handlerhttp.WriteStoreError(w, r, op, err)
		return
	}
	if h.notifyTaskEventChanged != nil {
		h.notifyTaskEventChanged(id, seq)
	}
	handlerhttp.WriteJSON(w, r, op, http.StatusOK, taskEventDetailFromDomain(ev, id))
}

func parseTaskEventsLimit(ctx context.Context, q url.Values) (limit int, err error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "taskevents.handler.parseTaskEventsLimit")
	ctx = calltrace.Push(ctx, "parseTaskEventsLimit")
	calltrace.HelperIOIn(ctx, "parseTaskEventsLimit", "limit_q", q.Get("limit"), "before_seq_q", q.Get("before_seq"), "after_seq_q", q.Get("after_seq"))
	defer func() { calltrace.HelperIOOut(ctx, "parseTaskEventsLimit", "limit", limit, "err", err) }()
	limit = readpolicy.TaskEventsDefaultLimit
	if v := q.Get("limit"); v != "" {
		if len(v) > maxTaskEventSeqParamBytes {
			return 0, fmt.Errorf("%w: limit too long", taskcoredomain.ErrInvalidInput)
		}
		n, e := strconv.Atoi(v)
		if e != nil || n < 0 || n > readpolicy.TaskEventsMaxLimit {
			return 0, fmt.Errorf("%w: limit must be integer 0..%d", taskcoredomain.ErrInvalidInput, readpolicy.TaskEventsMaxLimit)
		}
		limit = n
	}
	if limit <= 0 {
		limit = readpolicy.TaskEventsDefaultLimit
	}
	if limit > readpolicy.TaskEventsMaxLimit {
		limit = readpolicy.TaskEventsMaxLimit
	}
	return limit, nil
}
