package contract

import (
	"time"

	"github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
	cyclesdomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcycles/domain"
)

// TaskStats holds the global task counters.
type TaskStats struct {
	Total          int64
	Ready          int64
	Critical       int64
	Scheduled      int64
	ByStatus       map[domain.Status]int64
	ByPriority     map[domain.Priority]int64
	ByScope        map[string]int64
	Cycles         CycleStats
	Phases         PhaseStats
	Runner         RunnerStats
	RecentFailures []RecentFailure
}

// CycleStats aggregates task_cycles for stats consumers.
type CycleStats struct {
	ByStatus      map[cyclesdomain.CycleStatus]int64
	ByTriggeredBy map[domain.Actor]int64
}

// PhaseStats aggregates task_cycle_phases by (phase, status).
type PhaseStats struct {
	ByPhaseStatus map[cyclesdomain.Phase]map[cyclesdomain.PhaseStatus]int64
}

// RunnerStats aggregates terminal cycles by adapter identity.
type RunnerStats struct {
	ByRunner              map[string]RunnerBucket
	ByModel               map[string]RunnerBucket
	ByRunnerModel         map[string]RunnerBucket
	ByRunnerModelResolved map[string]RunnerBucket
}

// RunnerBucket is the per-bucket payload for runner stats.
type RunnerBucket struct {
	ByStatus                    map[cyclesdomain.CycleStatus]int64
	Succeeded                   int64
	DurationP50SucceededSeconds float64
	DurationP95SucceededSeconds float64
}

// RecentFailure is one row in the recent_failures slice on /tasks/stats.
type RecentFailure struct {
	TaskID     string
	EventSeq   int64
	At         time.Time
	CycleID    string
	AttemptSeq int64
	Status     string
	Reason     string
}

// ListCycleFailuresInput is the paginated query for cycle failures.
type ListCycleFailuresInput struct {
	Limit  int
	Offset int
	Sort   string
}

// ListCycleFailuresResult is returned by ListCycleFailures.
type ListCycleFailuresResult struct {
	Total               int64
	Failures            []RecentFailure
	ReasonSortTruncated bool
}
