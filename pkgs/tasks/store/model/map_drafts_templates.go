package model

import "github.com/AlexsanderHamir/Hamix/pkgs/tasks/domain"

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func FromDomainTaskDraft(d domain.TaskDraft) TaskDraft {
	return TaskDraft{
		ID:          d.ID,
		Name:        d.Name,
		PayloadJSON: d.PayloadJSON,
		CreatedAt:   d.CreatedAt,
		UpdatedAt:   d.UpdatedAt,
	}
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func FromDomainTaskDraftPtr(d *domain.TaskDraft) *TaskDraft {
	return MapPtr(d, FromDomainTaskDraft)
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func ToDomainTaskDraft(m TaskDraft) domain.TaskDraft {
	return domain.TaskDraft{
		ID:          m.ID,
		Name:        m.Name,
		PayloadJSON: m.PayloadJSON,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func ToDomainTaskDrafts(rows []TaskDraft) []domain.TaskDraft {
	return MapSlice(rows, ToDomainTaskDraft)
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func FromDomainTaskTemplate(d domain.TaskTemplate) TaskTemplate {
	return TaskTemplate{
		ID:               d.ID,
		Name:             d.Name,
		PayloadJSON:      d.PayloadJSON,
		InstantiateCount: d.InstantiateCount,
		CreatedAt:        d.CreatedAt,
		UpdatedAt:        d.UpdatedAt,
	}
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func FromDomainTaskTemplatePtr(d *domain.TaskTemplate) *TaskTemplate {
	return MapPtr(d, FromDomainTaskTemplate)
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func ToDomainTaskTemplate(m TaskTemplate) domain.TaskTemplate {
	return domain.TaskTemplate{
		ID:               m.ID,
		Name:             m.Name,
		PayloadJSON:      m.PayloadJSON,
		InstantiateCount: m.InstantiateCount,
		CreatedAt:        m.CreatedAt,
		UpdatedAt:        m.UpdatedAt,
	}
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func ToDomainTaskTemplates(rows []TaskTemplate) []domain.TaskTemplate {
	return MapSlice(rows, ToDomainTaskTemplate)
}
