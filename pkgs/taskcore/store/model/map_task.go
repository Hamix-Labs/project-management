package model

import (
	"github.com/AlexsanderHamir/Hamix/pkgs/storekernel/jsonmap"
	"github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
)

// FromDomainTask copies persisted columns from domain.Task to model.Task.
//
//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func FromDomainTask(d domain.Task) Task {
	return Task{
		ID:                    d.ID,
		Title:                 d.Title,
		InitialPrompt:         d.InitialPrompt,
		Status:                d.Status,
		Priority:              d.Priority,
		ProjectID:             d.ProjectID,
		Number:                d.Number,
		ProjectContextItemIDs: jsonmap.JSONStringSlice(d.ProjectContextItemIDs),
		Tags:                  jsonmap.JSONStringSlice(d.Tags),
		Milestone:             d.Milestone,
		Gate:                  d.Gate,
		Runner:                d.Runner,
		CursorModel:           d.CursorModel,
		RunnerConfig:          jsonmap.DatatypesFromRaw(d.RunnerConfig),
		PickupNotBefore:       d.PickupNotBefore,
		CriteriaSatisfiedAt:   d.CriteriaSatisfiedAt,
		PendingRetry:          d.PendingRetry,
		WorktreeID:            d.WorktreeID,
	}
}

// ToDomainTask copies persisted columns to domain.Task. DependsOn and CreatedAt
// remain zero until hydrate helpers run.
//
//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func ToDomainTask(m Task) domain.Task {
	runnerConfig := jsonmap.RawFromDatatypes(m.RunnerConfig)
	if len(runnerConfig) == 0 {
		runnerConfig = jsonmap.JSONRawObject()
	}
	return domain.Task{
		ID:                    m.ID,
		Title:                 m.Title,
		InitialPrompt:         m.InitialPrompt,
		Status:                m.Status,
		Priority:              m.Priority,
		ProjectID:             m.ProjectID,
		Number:                m.Number,
		ProjectContextItemIDs: jsonmap.StringSliceFromJSON(m.ProjectContextItemIDs),
		Tags:                  jsonmap.StringSliceFromJSON(m.Tags),
		Milestone:             m.Milestone,
		Gate:                  m.Gate,
		Runner:                m.Runner,
		CursorModel:           m.CursorModel,
		RunnerConfig:          runnerConfig,
		PickupNotBefore:       m.PickupNotBefore,
		CriteriaSatisfiedAt:   m.CriteriaSatisfiedAt,
		PendingRetry:          m.PendingRetry,
		WorktreeID:            m.WorktreeID,
	}
}

// ToDomainTasks maps a slice of persistence tasks to domain.Task.
//
//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func ToDomainTasks(rows []Task) []domain.Task {
	if len(rows) == 0 {
		return nil
	}
	out := make([]domain.Task, len(rows))
	for i := range rows {
		out[i] = ToDomainTask(rows[i])
	}
	return out
}

// FromDomainTaskPtr returns nil when d is nil.
//
//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func FromDomainTaskPtr(d *domain.Task) *Task {
	if d == nil {
		return nil
	}
	m := FromDomainTask(*d)
	return &m
}

// ToDomainTaskPtr returns nil when m is nil.
//
//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func ToDomainTaskPtr(m *Task) *domain.Task {
	if m == nil {
		return nil
	}
	d := ToDomainTask(*m)
	return &d
}
