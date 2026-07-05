package storefake

import (
	"context"
	"sync"

	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/contract"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/domain"
)

// GetCall records one TaskCRUDStore.Get invocation.
type GetCall struct {
	ID string
}

// TaskCRUDFake implements contract.TaskCRUDStore with call recording and
// injectable per-method outcomes for handler error-path tests.
type TaskCRUDFake struct {
	mu sync.Mutex

	getErr  error
	getTask *domain.Task

	getCalls []GetCall
}

// NewTaskCRUD returns an empty TaskCRUDFake. Get returns domain.ErrNotFound
// until OnGet or FailGet configures a response.
//
//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func NewTaskCRUD() *TaskCRUDFake {
	return &TaskCRUDFake{}
}

// FailGet configures Get to return err.
//
//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (f *TaskCRUDFake) FailGet(err error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.getErr = err
}

// OnGet configures Get to return task (copied by pointer on read).
//
//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (f *TaskCRUDFake) OnGet(task *domain.Task) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.getTask = task
}

// GetCalls returns a copy of recorded Get calls.
//
//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (f *TaskCRUDFake) GetCalls() []GetCall {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := make([]GetCall, len(f.getCalls))
	copy(out, f.getCalls)
	return out
}

// Get records the call and returns the configured outcome.
//
//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (f *TaskCRUDFake) Get(ctx context.Context, id string) (*domain.Task, error) {
	f.mu.Lock()
	f.getCalls = append(f.getCalls, GetCall{ID: id})
	err := f.getErr
	task := f.getTask
	f.mu.Unlock()
	if err != nil {
		return nil, err
	}
	if task != nil {
		return task, nil
	}
	return nil, domain.ErrNotFound
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (f *TaskCRUDFake) Create(ctx context.Context, in contract.CreateTaskInput, by domain.Actor) (*domain.Task, error) {
	return nil, errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (f *TaskCRUDFake) Update(ctx context.Context, id string, in contract.UpdateTaskInput, by domain.Actor) (*domain.Task, error) {
	return nil, errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (f *TaskCRUDFake) Delete(ctx context.Context, id string, by domain.Actor) ([]string, error) {
	return nil, errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (f *TaskCRUDFake) ListFlatPage(ctx context.Context, limit, offset int, filter *contract.ListFilter) ([]domain.Task, bool, error) {
	return nil, false, errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (f *TaskCRUDFake) ListFlatAfter(ctx context.Context, limit int, afterID string) ([]domain.Task, bool, error) {
	return nil, false, errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (f *TaskCRUDFake) TaskStats(ctx context.Context) (contract.TaskStats, error) {
	return contract.TaskStats{}, errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (f *TaskCRUDFake) RequestTaskRetry(ctx context.Context, in contract.RequestRetryInput, by domain.Actor) (*domain.Task, error) {
	return nil, errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (f *TaskCRUDFake) ApplyTaskGateAction(ctx context.Context, taskID, action string, by domain.Actor) (*domain.Task, error) {
	return nil, errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (f *TaskCRUDFake) ValidateTaskWorktreeBinding(ctx context.Context, projectID *string, worktreeID string) error {
	return errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (f *TaskCRUDFake) ListTaskDependencies(ctx context.Context, taskID string) ([]domain.DependencyEdge, error) {
	return nil, errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (f *TaskCRUDFake) AddTaskDependency(ctx context.Context, taskID, dependsOnTaskID string, satisfies domain.DependencySatisfies) error {
	return errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (f *TaskCRUDFake) RemoveTaskDependency(ctx context.Context, taskID, dependsOnTaskID string) error {
	return errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (f *TaskCRUDFake) ListCycleFailures(ctx context.Context, in contract.ListCycleFailuresInput) (contract.ListCycleFailuresResult, error) {
	return contract.ListCycleFailuresResult{}, errNotImplemented
}

var _ contract.TaskCRUDStore = (*TaskCRUDFake)(nil)
