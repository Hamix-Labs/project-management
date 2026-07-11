package domain

import taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"

type (
	Status     = taskcoredomain.Status
	Priority   = taskcoredomain.Priority
	GateStatus = taskcoredomain.GateStatus
)

const (
	StatusReady   = taskcoredomain.StatusReady
	StatusRunning = taskcoredomain.StatusRunning
	StatusBlocked = taskcoredomain.StatusBlocked
	StatusReview  = taskcoredomain.StatusReview
	StatusDone    = taskcoredomain.StatusDone
	StatusFailed  = taskcoredomain.StatusFailed
	StatusOnHold  = taskcoredomain.StatusOnHold

	PriorityLow      = taskcoredomain.PriorityLow
	PriorityMedium   = taskcoredomain.PriorityMedium
	PriorityHigh     = taskcoredomain.PriorityHigh
	PriorityCritical = taskcoredomain.PriorityCritical

	GateStatusLocked         = taskcoredomain.GateStatusLocked
	GateStatusActive         = taskcoredomain.GateStatusActive
	GateStatusPendingRelease = taskcoredomain.GateStatusPendingRelease
	GateStatusReleased       = taskcoredomain.GateStatusReleased
)
