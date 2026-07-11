package domain

import taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"

type (
	RetryMode    = taskcoredomain.RetryMode
	PendingRetry = taskcoredomain.PendingRetry
)

const (
	RetryFresh  = taskcoredomain.RetryFresh
	RetryResume = taskcoredomain.RetryResume
)
