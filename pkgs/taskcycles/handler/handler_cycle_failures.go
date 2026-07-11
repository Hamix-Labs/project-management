package handler

import (
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/calltrace"
	taskscontract "github.com/AlexsanderHamir/Hamix/pkgs/tasks/contract"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/domain"
)

const cycleFailureSortAtDesc = "at_desc"

type cycleFailureEntry struct {
	TaskID     string    `json:"task_id"`
	EventSeq   int64     `json:"event_seq"`
	At         time.Time `json:"at"`
	CycleID    string    `json:"cycle_id"`
	AttemptSeq int64     `json:"attempt_seq"`
	Status     string    `json:"status"`
	Reason     string    `json:"reason"`
}

type cycleFailuresResponse struct {
	Total               int64               `json:"total"`
	Limit               int                 `json:"limit"`
	Offset              int                 `json:"offset"`
	Sort                string              `json:"sort"`
	ReasonSortTruncated bool                `json:"reason_sort_truncated"`
	Failures            []cycleFailureEntry `json:"failures"`
}

func recentFailuresToJSON(failures []taskscontract.RecentFailure) []cycleFailureEntry {
	out := make([]cycleFailureEntry, 0, len(failures))
	for _, f := range failures {
		out = append(out, cycleFailureEntry{
			TaskID:     f.TaskID,
			EventSeq:   f.EventSeq,
			At:         f.At,
			CycleID:    f.CycleID,
			AttemptSeq: f.AttemptSeq,
			Status:     f.Status,
			Reason:     f.Reason,
		})
	}
	return out
}

func (h *Handler) cycleFailures(w http.ResponseWriter, r *http.Request) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "taskcycles.handler.cycleFailures")
	const op = "tasks.cycle_failures"
	r = calltrace.WithRequestRoot(r, op)
	limit, offset, sort, err := parseCycleFailuresQuery(r.URL.Query())
	if err != nil {
		writeStoreError(w, r, op, err)
		return
	}
	debugHTTPRequest(r, op, "limit", limit, "offset", offset, "sort", sort)
	out, err := h.failures.ListCycleFailures(r.Context(), taskscontract.ListCycleFailuresInput{
		Limit:  limit,
		Offset: offset,
		Sort:   sort,
	})
	if err != nil {
		writeStoreError(w, r, op, err)
		return
	}
	writeJSON(w, r, op, http.StatusOK, cycleFailuresResponse{
		Total:               out.Total,
		Limit:               limit,
		Offset:              offset,
		Sort:                sort,
		ReasonSortTruncated: out.ReasonSortTruncated,
		Failures:            recentFailuresToJSON(out.Failures),
	})
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func parseCycleFailuresQuery(q url.Values) (limit, offset int, sort string, err error) {
	limit = 50
	offset = 0
	sort = cycleFailureSortAtDesc
	if v := q.Get("limit"); v != "" {
		if len(v) > maxListIntQueryParamBytes {
			return 0, 0, "", fmt.Errorf("%w: limit value too long", domain.ErrInvalidInput)
		}
		n, e := strconv.Atoi(v)
		if e != nil || n < 1 || n > 200 {
			return 0, 0, "", fmt.Errorf("%w: limit must be integer 1..200", domain.ErrInvalidInput)
		}
		limit = n
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
	if v := strings.TrimSpace(q.Get("sort")); v != "" {
		sort = v
	}
	return limit, offset, sort, nil
}
