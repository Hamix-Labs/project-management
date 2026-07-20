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
	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
	"github.com/AlexsanderHamir/Hamix/pkgs/taskcycles/contract"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/handler/policy"
)

// maxCycleListLimitParamBytes mirrors maxTaskEventSeqParamBytes — keep
// list-paging limit query strings short.
const maxCycleListLimitParamBytes = 32

func parseCyclePathPair(r *http.Request) (string, string, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "taskcycles.handler.parseCyclePathPair")
	taskID, err := parseTaskPathID(r.PathValue("id"))
	if err != nil {
		return "", "", err
	}
	cycleID, err := parseTaskPathCycleID(r.PathValue("cycleId"))
	if err != nil {
		return "", "", err
	}
	return taskID, cycleID, nil
}

// assertCycleBelongsToTask preflights write routes so a cycleId from a
// different task surfaces as 404 instead of mutating the wrong row. The
// store does not enforce this implicitly because cycleId is unique on its
// own, so the handler must check.
func assertCycleBelongsToTask(ctx context.Context, s contract.CycleStore, taskID, cycleID string) error {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "taskcycles.handler.assertCycleBelongsToTask")
	c, err := s.GetCycle(ctx, cycleID)
	if err != nil {
		return err
	}
	if c.TaskID != taskID {
		return taskcoredomain.ErrNotFound
	}
	return nil
}

// parseCycleListBeforeAttemptSeq parses the optional ?before_attempt_seq=
// keyset cursor for GET /tasks/{id}/cycles. Mirrors the validation used
// by ?before_seq= on /tasks/{id}/events: 32-byte abuse guard, must be a
// strictly positive int64. Returns 0 (no cursor / first page) when the
// param is absent or empty after trim.
func parseCycleListBeforeAttemptSeq(ctx context.Context, q url.Values) (before int64, err error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "taskcycles.handler.parseCycleListBeforeAttemptSeq")
	ctx = calltrace.Push(ctx, "parseCycleListBeforeAttemptSeq")
	calltrace.HelperIOIn(ctx, "parseCycleListBeforeAttemptSeq", "before_q", q.Get("before_attempt_seq"))
	defer func() {
		calltrace.HelperIOOut(ctx, "parseCycleListBeforeAttemptSeq", "before_attempt_seq", before, "err", err)
	}()
	v := strings.TrimSpace(q.Get("before_attempt_seq"))
	if v == "" {
		return 0, nil
	}
	if len(v) > maxCycleListLimitParamBytes {
		return 0, fmt.Errorf("%w: before_attempt_seq too long", taskcoredomain.ErrInvalidInput)
	}
	n, e := strconv.ParseInt(v, 10, 64)
	if e != nil || n < 1 {
		return 0, fmt.Errorf("%w: before_attempt_seq must be a positive integer", taskcoredomain.ErrInvalidInput)
	}
	return n, nil
}

// parseCycleListLimit is the GET /tasks/{id}/cycles equivalent of
// parseTaskEventsLimit. Same 0..200 cap and 32-byte abuse guard.
func parseCycleListLimit(ctx context.Context, q url.Values) (int, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "taskcycles.handler.parseCycleListLimit")
	ctx = calltrace.Push(ctx, "parseCycleListLimit")
	calltrace.HelperIOIn(ctx, "parseCycleListLimit", "limit_q", q.Get("limit"))
	var (
		limit = policy.CycleListDefaultLimit
		err   error
	)
	defer func() { calltrace.HelperIOOut(ctx, "parseCycleListLimit", "limit", limit, "err", err) }()
	v := strings.TrimSpace(q.Get("limit"))
	if v == "" {
		return limit, nil
	}
	if len(v) > maxCycleListLimitParamBytes {
		err = fmt.Errorf("%w: limit too long", taskcoredomain.ErrInvalidInput)
		return 0, err
	}
	n, e := strconv.Atoi(v)
	if e != nil || n < 0 || n > policy.CycleListMaxLimit {
		err = fmt.Errorf("%w: limit must be integer 0..%d", taskcoredomain.ErrInvalidInput, policy.CycleListMaxLimit)
		return 0, err
	}
	if n == 0 {
		return policy.CycleListDefaultLimit, nil
	}
	limit = n
	return limit, nil
}

func parseCycleStreamAfterSeq(ctx context.Context, q url.Values) (after int64, err error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "taskcycles.handler.parseCycleStreamAfterSeq")
	ctx = calltrace.Push(ctx, "parseCycleStreamAfterSeq")
	calltrace.HelperIOIn(ctx, "parseCycleStreamAfterSeq", "after_q", q.Get("after_seq"))
	defer func() { calltrace.HelperIOOut(ctx, "parseCycleStreamAfterSeq", "after_seq", after, "err", err) }()
	v := strings.TrimSpace(q.Get("after_seq"))
	if v == "" {
		return 0, nil
	}
	if len(v) > maxCycleListLimitParamBytes {
		return 0, fmt.Errorf("%w: after_seq too long", taskcoredomain.ErrInvalidInput)
	}
	n, e := strconv.ParseInt(v, 10, 64)
	if e != nil || n < 1 {
		return 0, fmt.Errorf("%w: after_seq must be a positive integer", taskcoredomain.ErrInvalidInput)
	}
	return n, nil
}

func parseCycleStreamLimit(ctx context.Context, q url.Values) (int, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "taskcycles.handler.parseCycleStreamLimit")
	ctx = calltrace.Push(ctx, "parseCycleStreamLimit")
	calltrace.HelperIOIn(ctx, "parseCycleStreamLimit", "limit_q", q.Get("limit"))
	var (
		limit = policy.CycleStreamDefaultLimit
		err   error
	)
	defer func() { calltrace.HelperIOOut(ctx, "parseCycleStreamLimit", "limit", limit, "err", err) }()
	v := strings.TrimSpace(q.Get("limit"))
	if v == "" {
		return limit, nil
	}
	if len(v) > maxCycleListLimitParamBytes {
		err = fmt.Errorf("%w: limit too long", taskcoredomain.ErrInvalidInput)
		return 0, err
	}
	n, e := strconv.Atoi(v)
	if e != nil || n < 0 || n > policy.CycleStreamMaxLimit {
		err = fmt.Errorf("%w: limit must be integer 0..%d", taskcoredomain.ErrInvalidInput, policy.CycleStreamMaxLimit)
		return 0, err
	}
	if n == 0 {
		return policy.CycleStreamDefaultLimit, nil
	}
	limit = n
	return limit, nil
}

// paginateMappedRows applies limit+1 paging: trims overflow, maps domain rows to
// wire responses, and returns an optional cursor from the last mapped item.
//
//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling cycle list handler."
func paginateMappedRows[T any, R any](
	rows []T,
	limit int,
	mapFn func(*T) R,
	cursorFn func(R) int64,
) (out []R, hasMore bool, next *int64) {
	if len(rows) > limit {
		hasMore = true
		rows = rows[:limit]
	}
	out = make([]R, 0, len(rows))
	for i := range rows {
		out = append(out, mapFn(&rows[i]))
	}
	if hasMore && len(out) > 0 {
		n := cursorFn(out[len(out)-1])
		next = &n
	}
	return out, hasMore, next
}
