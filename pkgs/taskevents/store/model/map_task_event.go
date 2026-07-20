package model

import (
	"github.com/AlexsanderHamir/Hamix/pkgs/storekernel/jsonmap"
	taskeventsdomain "github.com/AlexsanderHamir/Hamix/pkgs/taskevents/domain"
)

// FromDomainTaskEvent copies a domain row to its persistence model.
//
//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func FromDomainTaskEvent(d taskeventsdomain.TaskEvent) TaskEvent {
	return TaskEvent{
		TaskID:         d.TaskID,
		Seq:            d.Seq,
		At:             d.At,
		Type:           d.Type,
		By:             d.By,
		Data:           jsonmap.DatatypesFromRaw(d.Data),
		UserResponse:   d.UserResponse,
		UserResponseAt: d.UserResponseAt,
		ResponseThread: jsonmap.DatatypesFromRaw(d.ResponseThread),
	}
}

// ToDomainTaskEvent copies a persistence row to domain.TaskEvent.
//
//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func ToDomainTaskEvent(m TaskEvent) taskeventsdomain.TaskEvent {
	return taskeventsdomain.TaskEvent{
		TaskID:         m.TaskID,
		Seq:            m.Seq,
		At:             m.At,
		Type:           m.Type,
		By:             m.By,
		Data:           jsonmap.RawJSONObjectFromDatatypes(m.Data),
		UserResponse:   m.UserResponse,
		UserResponseAt: m.UserResponseAt,
		ResponseThread: jsonmap.RawFromDatatypes(m.ResponseThread),
	}
}

// ToDomainTaskEvents maps a slice of persistence rows to domain.TaskEvent.
//
//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func ToDomainTaskEvents(rows []TaskEvent) []taskeventsdomain.TaskEvent {
	if len(rows) == 0 {
		return nil
	}
	out := make([]taskeventsdomain.TaskEvent, len(rows))
	for i := range rows {
		out[i] = ToDomainTaskEvent(rows[i])
	}
	return out
}
