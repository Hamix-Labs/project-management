package handler

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/AlexsanderHamir/Hamix/pkgs/obs/calltrace"
	"github.com/AlexsanderHamir/Hamix/pkgs/taskcore/contract"
	"github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
	"github.com/google/uuid"
)

const (
	maxListIntQueryParamBytes          = 32
	maxListAfterIDParamBytes           = 128
	maxTemplateInstantiateCountPerItem = 25
	maxTemplateInstantiateTotalCreates = 100
)

// MaxListIntQueryParamBytes is the max byte length for list limit/offset query params.
const MaxListIntQueryParamBytes = maxListIntQueryParamBytes

// MaxListAfterIDParamBytes is the max byte length for list after_id query params.
const MaxListAfterIDParamBytes = maxListAfterIDParamBytes

// ParseListParams parses limit, offset, and after_id from task list query values.
func ParseListParams(ctx context.Context, q url.Values) (limit, offset int, afterID string, err error) {
	return parseListParams(ctx, q)
}

func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "handler.Handler.create")
	const op = "tasks.create"
	r = calltrace.WithRequestRoot(r, op)
	var body taskCreateJSON
	if err := h.httpPort.DecodeJSON(r.Context(), r.Body, &body); err != nil {
		h.debugHTTPRequest(r, op, "json_decode_failed", true)
		h.httpPort.WriteError(w, r, op, err, http.StatusBadRequest)
		return
	}
	by := h.httpPort.ActorFromRequest(r)
	h.debugHTTPRequest(r, op, taskCreateInputFields(&body, string(by))...)
	task, err := h.CreateTaskFromComposeJSON(r.Context(), r, op, taskCreateJSONToCompose(body), CreateTaskComposeOpts{
		ID:      body.ID,
		DraftID: body.DraftID,
		Gate:    body.Gate,
	}, by)
	if err != nil {
		h.httpPort.WriteStoreError(w, r, op, err)
		return
	}
	h.httpPort.WriteJSON(w, r, op, http.StatusCreated, task)
}

func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "handler.Handler.get")
	const op = "tasks.get"
	r = calltrace.WithRequestRoot(r, op)
	id, err := h.parseTaskPathID(r.PathValue("id"))
	if err != nil {
		h.httpPort.WriteStoreError(w, r, op, err)
		return
	}
	h.debugHTTPRequest(r, op, "task_id", id)
	t, err := h.tasks.Get(r.Context(), id)
	if err != nil {
		h.httpPort.WriteStoreError(w, r, op, err)
		return
	}
	h.httpPort.WriteJSONWithETag(w, r, op, http.StatusOK, t)
}

// list serves GET /tasks — the hottest read path in taskapi (SPA initial load
// and SSE-driven refetch). Failure contract for operators:
//   - Invalid query params → 400 {"error":"..."} with failure_stage=parse_list_params
//     and raw limit_q/offset_q/after_id_q on the warn-level "request failed" log.
//   - Store/persistence errors (closed DB, driver faults) → 500 with
//     failure_stage=store_list plus resolved limit/offset/after_id/pagination_mode.
//   - Request context canceled or deadline exceeded → 408/504 via storeErrHTTPResponse.
//   - JSON encode failures → 500 with error log msg=response encode failed and
//     failure_stage=response_encode (includes request_id and route when available).
//   - Response-body write failures after headers → truncated body with
//     msg=response write failed and failure_stage body or newline (never silent).
//
// Successful responses never publish SSE events.
func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "handler.Handler.list")
	const op = "tasks.list"
	r = calltrace.WithRequestRoot(r, op)
	q := r.URL.Query()
	limit, offset, afterID, err := parseListParams(r.Context(), q)
	if err != nil {
		h.debugHTTPRequest(r, op, "list_params_invalid", true)
		h.httpPort.WriteStoreError(w, r, op, err,
			"failure_stage", "parse_list_params",
			"limit_q", q.Get("limit"),
			"offset_q", q.Get("offset"),
			"after_id_q", q.Get("after_id"),
		)
		return
	}
	h.debugHTTPRequest(r, op, "limit", limit, "offset", offset, "after_id", afterID)
	var tasks []domain.Task
	var hasMore bool
	if afterID != "" {
		tasks, hasMore, err = h.tasks.ListFlatAfter(r.Context(), limit, afterID)
		offset = 0
	} else {
		tasks, hasMore, err = h.tasks.ListFlatPage(r.Context(), limit, offset, nil)
	}
	if err != nil {
		mode := "offset"
		if afterID != "" {
			mode = "keyset"
		}
		h.httpPort.WriteStoreError(w, r, op, err,
			"failure_stage", "store_list",
			"limit", limit,
			"offset", offset,
			"after_id", afterID,
			"pagination_mode", mode,
		)
		return
	}
	h.httpPort.WriteJSONWithETag(w, r, op, http.StatusOK, buildListResponse(tasks, limit, offset, hasMore))
}

func (h *Handler) stats(w http.ResponseWriter, r *http.Request) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "handler.Handler.stats")
	const op = "tasks.stats"
	r = calltrace.WithRequestRoot(r, op)
	h.debugHTTPRequest(r, op)
	stats, err := h.tasks.TaskStats(r.Context())
	if err != nil {
		h.httpPort.WriteStoreError(w, r, op, err)
		return
	}
	h.httpPort.WriteJSONWithETag(w, r, op, http.StatusOK, taskStatsResponseFromStore(stats))
}

func (h *Handler) delete(w http.ResponseWriter, r *http.Request) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "handler.Handler.delete")
	const op = "tasks.delete"
	r = calltrace.WithRequestRoot(r, op)
	id, err := h.parseTaskPathID(r.PathValue("id"))
	if err != nil {
		h.httpPort.WriteStoreError(w, r, op, err)
		return
	}
	h.debugHTTPRequest(r, op, "task_id", id)
	by := h.httpPort.ActorFromRequest(r)
	deletedIDs, err := h.tasks.Delete(r.Context(), id, by)
	if err != nil {
		h.httpPort.WriteStoreError(w, r, op, err)
		return
	}
	for _, deletedID := range deletedIDs {
		h.notifyChangeSafe(contract.ChangeTaskDeleted, deletedID)
		taskapiDomainTasksDeletedTotal.Inc()
	}
	h.debugHTTPOut(r.Context(), op, http.StatusNoContent,
		"task_id", id,
		"deleted_count", len(deletedIDs),
		"response_empty", true)
	w.WriteHeader(http.StatusNoContent)
}

func parseListParams(ctx context.Context, q url.Values) (limit, offset int, afterID string, err error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "handler.parseListParams")
	ctx = calltrace.Push(ctx, "parseListParams")
	calltrace.HelperIOIn(ctx, "parseListParams", "limit_q", q.Get("limit"), "offset_q", q.Get("offset"), "after_id_q", q.Get("after_id"))
	defer func() {
		calltrace.HelperIOOut(ctx, "parseListParams", "limit", limit, "offset", offset, "after_id", afterID, "err", err)
	}()
	limit = contract.TaskListDefaultLimit
	offset = 0
	afterID = strings.TrimSpace(q.Get("after_id"))
	if afterID != "" && len(afterID) > maxListAfterIDParamBytes {
		return 0, 0, "", fmt.Errorf("%w: after_id too long", domain.ErrInvalidInput)
	}
	if _, ok := q["offset"]; ok && afterID != "" {
		return 0, 0, "", fmt.Errorf("%w: offset cannot be used with after_id", domain.ErrInvalidInput)
	}
	if afterID != "" {
		if _, perr := uuid.Parse(afterID); perr != nil {
			return 0, 0, "", fmt.Errorf("%w: after_id must be a UUID", domain.ErrInvalidInput)
		}
	}
	if v := q.Get("limit"); v != "" {
		if len(v) > maxListIntQueryParamBytes {
			return 0, 0, "", fmt.Errorf("%w: limit value too long", domain.ErrInvalidInput)
		}
		n, e := strconv.Atoi(v)
		if e != nil || n < 0 || n > contract.TaskListMaxLimit {
			return 0, 0, "", fmt.Errorf("%w: limit must be integer 0..%d", domain.ErrInvalidInput, contract.TaskListMaxLimit)
		}
		limit = n
	}
	if limit <= 0 {
		limit = contract.TaskListDefaultLimit
	}
	if v := q.Get("offset"); v != "" {
		if len(v) > maxListIntQueryParamBytes {
			return 0, 0, "", fmt.Errorf("%w: offset value too long", domain.ErrInvalidInput)
		}
		n, e := strconv.Atoi(v)
		if e != nil || n < 0 {
			return 0, 0, "", fmt.Errorf("%w: offset must be non-negative integer", domain.ErrInvalidInput)
		}
		offset = n
	}
	return limit, offset, afterID, nil
}
