package contract

import taskcorecontract "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/contract"

type (
	TaskStats               = taskcorecontract.TaskStats
	CycleStats              = taskcorecontract.CycleStats
	PhaseStats              = taskcorecontract.PhaseStats
	RunnerStats             = taskcorecontract.RunnerStats
	RunnerBucket            = taskcorecontract.RunnerBucket
	RecentFailure           = taskcorecontract.RecentFailure
	ListCycleFailuresInput  = taskcorecontract.ListCycleFailuresInput
	ListCycleFailuresResult = taskcorecontract.ListCycleFailuresResult
)
