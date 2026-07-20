package storefake

import (
	"context"
	"sync"

	cyclesdomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcycles/domain"
	taskeventsdomain "github.com/AlexsanderHamir/Hamix/pkgs/taskevents/domain"
)

// ListCyclesCall records one ListCyclesForTaskBefore invocation.
type ListCyclesCall struct {
	TaskID           string
	BeforeAttemptSeq int64
	Limit            int
}

// ListEventsCall records one ListTaskEvents invocation.
type ListEventsCall struct {
	TaskID string
}

// HandlerStoreFake composes TaskCRUDFake with scriptable ListCycles /
// ListEvents slices and stubs for the remaining HandlerStore methods.
// Use NewHandlerStore for handler tests that exercise Get/Retry/Gate/
// cycles/events without SQLite.
type HandlerStoreFake struct {
	*TaskCRUDFake
	unimplementedHandlerStore

	mu sync.Mutex

	listCyclesErr  error
	listCycles     []cyclesdomain.TaskCycle
	listCyclesCalls []ListCyclesCall

	listEventsErr   error
	listEvents      []taskeventsdomain.TaskEvent
	listEventsCalls []ListEventsCall
}

// NewHandlerStore returns a HandlerStoreFake with an embedded TaskCRUDFake.
//
//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func NewHandlerStore() *HandlerStoreFake {
	return &HandlerStoreFake{TaskCRUDFake: NewTaskCRUD()}
}

// NewHandlerStoreFromTaskCRUD wraps an existing TaskCRUDFake.
//
//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func NewHandlerStoreFromTaskCRUD(crud *TaskCRUDFake) *HandlerStoreFake {
	if crud == nil {
		crud = NewTaskCRUD()
	}
	return &HandlerStoreFake{TaskCRUDFake: crud}
}

// FailListCycles configures ListCyclesForTaskBefore to return err.
//
//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (f *HandlerStoreFake) FailListCycles(err error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.listCyclesErr = err
}

// OnListCycles configures ListCyclesForTaskBefore to return cycles.
//
//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (f *HandlerStoreFake) OnListCycles(cycles []cyclesdomain.TaskCycle) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.listCycles = append([]cyclesdomain.TaskCycle(nil), cycles...)
}

// ListCyclesCalls returns a copy of recorded ListCyclesForTaskBefore calls.
//
//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (f *HandlerStoreFake) ListCyclesCalls() []ListCyclesCall {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := make([]ListCyclesCall, len(f.listCyclesCalls))
	copy(out, f.listCyclesCalls)
	return out
}

// FailListEvents configures ListTaskEvents to return err.
//
//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (f *HandlerStoreFake) FailListEvents(err error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.listEventsErr = err
}

// OnListEvents configures ListTaskEvents to return events.
//
//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (f *HandlerStoreFake) OnListEvents(events []taskeventsdomain.TaskEvent) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.listEvents = append([]taskeventsdomain.TaskEvent(nil), events...)
}

// ListEventsCalls returns a copy of recorded ListTaskEvents calls.
//
//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (f *HandlerStoreFake) ListEventsCalls() []ListEventsCall {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := make([]ListEventsCall, len(f.listEventsCalls))
	copy(out, f.listEventsCalls)
	return out
}

// ListCyclesForTaskBefore records the call and returns the configured outcome.
//
//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (f *HandlerStoreFake) ListCyclesForTaskBefore(ctx context.Context, taskID string, beforeAttemptSeq int64, limit int) ([]cyclesdomain.TaskCycle, error) {
	f.mu.Lock()
	f.listCyclesCalls = append(f.listCyclesCalls, ListCyclesCall{
		TaskID: taskID, BeforeAttemptSeq: beforeAttemptSeq, Limit: limit,
	})
	err := f.listCyclesErr
	cycles := append([]cyclesdomain.TaskCycle(nil), f.listCycles...)
	f.mu.Unlock()
	if err != nil {
		return nil, err
	}
	if cycles != nil {
		return cycles, nil
	}
	return nil, errNotImplemented
}

// ListTaskEvents records the call and returns the configured outcome.
//
//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (f *HandlerStoreFake) ListTaskEvents(ctx context.Context, taskID string) ([]taskeventsdomain.TaskEvent, error) {
	f.mu.Lock()
	f.listEventsCalls = append(f.listEventsCalls, ListEventsCall{TaskID: taskID})
	err := f.listEventsErr
	events := append([]taskeventsdomain.TaskEvent(nil), f.listEvents...)
	f.mu.Unlock()
	if err != nil {
		return nil, err
	}
	if events != nil {
		return events, nil
	}
	return nil, errNotImplemented
}

type unimplementedHandlerStore struct{}
