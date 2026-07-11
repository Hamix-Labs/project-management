package model

import cyclesdomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcycles/domain"

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func FromDomainTaskCycle(d cyclesdomain.TaskCycle) TaskCycle {
	return TaskCycle{
		ID:            d.ID,
		TaskID:        d.TaskID,
		AttemptSeq:    d.AttemptSeq,
		Status:        d.Status,
		StartedAt:     d.StartedAt,
		EndedAt:       d.EndedAt,
		TriggeredBy:   d.TriggeredBy,
		ParentCycleID: d.ParentCycleID,
		MetaJSON:      datatypesFromRaw(d.MetaJSON),
	}
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func FromDomainTaskCyclePtr(d *cyclesdomain.TaskCycle) *TaskCycle {
	if d == nil {
		return nil
	}
	m := FromDomainTaskCycle(*d)
	return &m
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func ToDomainTaskCycle(m TaskCycle) cyclesdomain.TaskCycle {
	return cyclesdomain.TaskCycle{
		ID:            m.ID,
		TaskID:        m.TaskID,
		AttemptSeq:    m.AttemptSeq,
		Status:        m.Status,
		StartedAt:     m.StartedAt,
		EndedAt:       m.EndedAt,
		TriggeredBy:   m.TriggeredBy,
		ParentCycleID: m.ParentCycleID,
		MetaJSON:      rawFromDatatypes(m.MetaJSON),
	}
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func ToDomainTaskCyclePtr(m *TaskCycle) *cyclesdomain.TaskCycle {
	if m == nil {
		return nil
	}
	d := ToDomainTaskCycle(*m)
	return &d
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func ToDomainTaskCycles(rows []TaskCycle) []cyclesdomain.TaskCycle {
	if len(rows) == 0 {
		return nil
	}
	out := make([]cyclesdomain.TaskCycle, len(rows))
	for i := range rows {
		out[i] = ToDomainTaskCycle(rows[i])
	}
	return out
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func FromDomainTaskCyclePhase(d cyclesdomain.TaskCyclePhase) TaskCyclePhase {
	return TaskCyclePhase{
		ID:          d.ID,
		CycleID:     d.CycleID,
		Phase:       d.Phase,
		PhaseSeq:    d.PhaseSeq,
		Status:      d.Status,
		StartedAt:   d.StartedAt,
		EndedAt:     d.EndedAt,
		Summary:     d.Summary,
		DetailsJSON: datatypesFromRaw(d.DetailsJSON),
		EventSeq:    d.EventSeq,
	}
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func FromDomainTaskCyclePhasePtr(d *cyclesdomain.TaskCyclePhase) *TaskCyclePhase {
	if d == nil {
		return nil
	}
	m := FromDomainTaskCyclePhase(*d)
	return &m
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func ToDomainTaskCyclePhase(m TaskCyclePhase) cyclesdomain.TaskCyclePhase {
	return cyclesdomain.TaskCyclePhase{
		ID:          m.ID,
		CycleID:     m.CycleID,
		Phase:       m.Phase,
		PhaseSeq:    m.PhaseSeq,
		Status:      m.Status,
		StartedAt:   m.StartedAt,
		EndedAt:     m.EndedAt,
		Summary:     m.Summary,
		DetailsJSON: rawFromDatatypes(m.DetailsJSON),
		EventSeq:    m.EventSeq,
	}
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func ToDomainTaskCyclePhasePtr(m *TaskCyclePhase) *cyclesdomain.TaskCyclePhase {
	if m == nil {
		return nil
	}
	d := ToDomainTaskCyclePhase(*m)
	return &d
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func ToDomainTaskCyclePhases(rows []TaskCyclePhase) []cyclesdomain.TaskCyclePhase {
	if len(rows) == 0 {
		return nil
	}
	out := make([]cyclesdomain.TaskCyclePhase, len(rows))
	for i := range rows {
		out[i] = ToDomainTaskCyclePhase(rows[i])
	}
	return out
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func FromDomainTaskCycleStreamEvent(d cyclesdomain.TaskCycleStreamEvent) TaskCycleStreamEvent {
	return TaskCycleStreamEvent{
		ID:          d.ID,
		TaskID:      d.TaskID,
		CycleID:     d.CycleID,
		PhaseSeq:    d.PhaseSeq,
		StreamSeq:   d.StreamSeq,
		At:          d.At,
		Source:      d.Source,
		Kind:        d.Kind,
		Subtype:     d.Subtype,
		Message:     d.Message,
		Tool:        d.Tool,
		PayloadJSON: datatypesFromRaw(d.PayloadJSON),
	}
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func FromDomainTaskCycleStreamEventPtr(d *cyclesdomain.TaskCycleStreamEvent) *TaskCycleStreamEvent {
	if d == nil {
		return nil
	}
	m := FromDomainTaskCycleStreamEvent(*d)
	return &m
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func ToDomainTaskCycleStreamEvent(m TaskCycleStreamEvent) cyclesdomain.TaskCycleStreamEvent {
	return cyclesdomain.TaskCycleStreamEvent{
		ID:          m.ID,
		TaskID:      m.TaskID,
		CycleID:     m.CycleID,
		PhaseSeq:    m.PhaseSeq,
		StreamSeq:   m.StreamSeq,
		At:          m.At,
		Source:      m.Source,
		Kind:        m.Kind,
		Subtype:     m.Subtype,
		Message:     m.Message,
		Tool:        m.Tool,
		PayloadJSON: rawFromDatatypes(m.PayloadJSON),
	}
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func ToDomainTaskCycleStreamEvents(rows []TaskCycleStreamEvent) []cyclesdomain.TaskCycleStreamEvent {
	if len(rows) == 0 {
		return nil
	}
	out := make([]cyclesdomain.TaskCycleStreamEvent, len(rows))
	for i := range rows {
		out[i] = ToDomainTaskCycleStreamEvent(rows[i])
	}
	return out
}
