package contract

import "github.com/AlexsanderHamir/Hamix/pkgs/tasks/domain"

// TaskEventsPage is one window of audit events plus paging metadata.
type TaskEventsPage struct {
	Events       []domain.TaskEvent
	Total        int64
	RangeStart   int64
	RangeEnd     int64
	HasMoreNewer bool
	HasMoreOlder bool
}
