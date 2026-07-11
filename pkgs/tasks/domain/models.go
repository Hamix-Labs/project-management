package domain

import taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"

type (
	Task                = taskcoredomain.Task
	TaskContextSnapshot = taskcoredomain.TaskContextSnapshot
	TaskDependency      = taskcoredomain.TaskDependency
	DependencyEdge      = taskcoredomain.DependencyEdge
)
