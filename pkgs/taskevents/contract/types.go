package contract

import taskeventsdomain "github.com/AlexsanderHamir/Hamix/pkgs/taskevents/domain"

// TaskEventsPage is one window of audit events plus paging metadata.
type TaskEventsPage struct {
	Events       []taskeventsdomain.TaskEvent
	Total        int64
	RangeStart   int64
	RangeEnd     int64
	HasMoreNewer bool
	HasMoreOlder bool
}
