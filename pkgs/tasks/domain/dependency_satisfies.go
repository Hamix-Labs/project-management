package domain

import taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"

type DependencySatisfies = taskcoredomain.DependencySatisfies

const DependencySatisfiesDone = taskcoredomain.DependencySatisfiesDone

var (
	ValidDependencySatisfies     = taskcoredomain.ValidDependencySatisfies
	NormalizeDependencySatisfies = taskcoredomain.NormalizeDependencySatisfies
)
