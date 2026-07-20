package failures

import "github.com/AlexsanderHamir/Hamix/pkgs/obs/calltrace"
import (
	"context"
	"fmt"
	"log/slog"
	"sort"
	"strings"

	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
	"github.com/AlexsanderHamir/Hamix/pkgs/taskcycles/contract"
	taskeventsdomain "github.com/AlexsanderHamir/Hamix/pkgs/taskevents/domain"
	eventsmodel "github.com/AlexsanderHamir/Hamix/pkgs/taskevents/store/model"
	"gorm.io/gorm"
)

// Cycle failure list sort modes for GET /tasks/cycle-failures.
const (
	SortAtDesc     = "at_desc"
	SortAtAsc      = "at_asc"
	SortReasonAsc  = "reason_asc"
	SortReasonDesc = "reason_desc"
	defaultLimit   = 50
	maxLimit       = 200
	// reasonSortFetchCap bounds how many newest cycle_failed rows we load
	// when sorting by reason (enrichment + sort happen in memory).
	reasonSortFetchCap = 2000
)

// ListInput is the paginated / sorted query for the dedicated cycle failures view.
type ListInput = contract.ListCycleFailuresInput

// ListResult is returned by List.
type ListResult = contract.ListCycleFailuresResult

// List returns cycle_failed mirror rows with the same enrichment as
// /tasks/stats recent_failures. Time-based sorts use SQL pagination;
// reason sorts load up to reasonSortFetchCap newest rows, enrich, sort
// in memory, then slice for offset/limit.
func List(ctx context.Context, db *gorm.DB, in ListInput) (ListResult, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "taskcycles.store.failures.List",
		"limit", in.Limit, "offset", in.Offset, "sort", in.Sort)
	limit := in.Limit
	if limit <= 0 {
		limit = defaultLimit
	}
	if limit > maxLimit {
		limit = maxLimit
	}
	offset := in.Offset
	if offset < 0 {
		offset = 0
	}
	sortKey := strings.TrimSpace(in.Sort)
	if sortKey == "" {
		sortKey = SortAtDesc
	}
	switch sortKey {
	case SortAtDesc, SortAtAsc, SortReasonAsc, SortReasonDesc:
	default:
		return ListResult{}, fmt.Errorf("%w: invalid sort", taskcoredomain.ErrInvalidInput)
	}

	var total int64
	if err := db.WithContext(ctx).Model(&eventsmodel.TaskEvent{}).
		Where("type = ?", string(taskeventsdomain.EventCycleFailed)).
		Count(&total).Error; err != nil {
		return ListResult{}, fmt.Errorf("count cycle failures: %w", err)
	}

	switch sortKey {
	case SortAtDesc, SortAtAsc:
		var rows []cycleFailedRow
		q := db.WithContext(ctx).Model(&eventsmodel.TaskEvent{}).
			Select("task_id, seq, at, data_json").
			Where("type = ?", string(taskeventsdomain.EventCycleFailed))
		if sortKey == SortAtDesc {
			q = q.Order("at DESC, seq DESC")
		} else {
			q = q.Order("at ASC, seq ASC")
		}
		if err := q.Limit(limit).Offset(offset).Scan(&rows).Error; err != nil {
			return ListResult{}, fmt.Errorf("list cycle failures: %w", err)
		}
		listed := decodeCycleFailedRows(rows)
		enrichFromPhaseEvents(ctx, db, listed)
		return ListResult{Total: total, Failures: listed}, nil

	case SortReasonAsc, SortReasonDesc:
		var rows []cycleFailedRow
		if err := db.WithContext(ctx).Model(&eventsmodel.TaskEvent{}).
			Select("task_id, seq, at, data_json").
			Where("type = ?", string(taskeventsdomain.EventCycleFailed)).
			Order("at DESC, seq DESC").
			Limit(reasonSortFetchCap).
			Scan(&rows).Error; err != nil {
			return ListResult{}, fmt.Errorf("list cycle failures for reason sort: %w", err)
		}
		listed := decodeCycleFailedRows(rows)
		enrichFromPhaseEvents(ctx, db, listed)
		if sortKey == SortReasonAsc {
			sort.SliceStable(listed, func(i, j int) bool {
				return strings.ToLower(listed[i].Reason) < strings.ToLower(listed[j].Reason)
			})
		} else {
			sort.SliceStable(listed, func(i, j int) bool {
				return strings.ToLower(listed[i].Reason) > strings.ToLower(listed[j].Reason)
			})
		}
		truncated := total > int64(reasonSortFetchCap)
		if offset >= len(listed) {
			return ListResult{
				Total:               total,
				Failures:            nil,
				ReasonSortTruncated: truncated,
			}, nil
		}
		end := offset + limit
		if end > len(listed) {
			end = len(listed)
		}
		page := listed[offset:end]
		return ListResult{
			Total:               total,
			Failures:            page,
			ReasonSortTruncated: truncated,
		}, nil

	default:
		return ListResult{}, fmt.Errorf("%w: invalid sort", taskcoredomain.ErrInvalidInput)
	}
}
