package domain

import taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"

var (
	ValidateTaskTag       = taskcoredomain.ValidateTaskTag
	ValidateTaskTags      = taskcoredomain.ValidateTaskTags
	ValidateTaskMilestone = taskcoredomain.ValidateTaskMilestone
	NormalizeTaskTags     = taskcoredomain.NormalizeTaskTags
)
