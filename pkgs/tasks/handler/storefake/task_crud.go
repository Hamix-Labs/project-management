package storefake

import (
	"context"
	"sync"

	taskcorecontract "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/contract"
	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
)

// GetCall records one TaskCRUDStore.Get invocation.
type GetCall struct {
	ID string
}

// RetryCall records one RequestTaskRetry invocation.
type RetryCall struct {
	Input taskcorecontract.RequestRetryInput
	By    taskcoredomain.Actor
}

// ApproveCall records one RequestTaskApprove invocation.
type ApproveCall struct {
	TaskID string
	By     taskcoredomain.Actor
}

// PolishCall records one RequestTaskPolish invocation.
type PolishCall struct {
	Input taskcorecontract.RequestPolishInput
	By    taskcoredomain.Actor
}

// GateCall records one ApplyTaskGateAction invocation.
type GateCall struct {
	TaskID string
	Action taskcorecontract.GateAction
	By     taskcoredomain.Actor
}

// TaskCRUDFake implements the focused taskcore contracts (and composed
// TaskCRUDStore) with call recording and injectable per-method outcomes
// for handler error-path tests.
type TaskCRUDFake struct {
	mu sync.Mutex

	getErr  error
	getTask *taskcoredomain.Task

	retryErr  error
	retryTask *taskcoredomain.Task

	approveErr  error
	approveTask *taskcoredomain.Task

	polishErr  error
	polishTask *taskcoredomain.Task

	gateErr  error
	gateTask *taskcoredomain.Task

	getCalls     []GetCall
	retryCalls   []RetryCall
	approveCalls []ApproveCall
	polishCalls  []PolishCall
	gateCalls    []GateCall
}

// NewTaskCRUD returns an empty TaskCRUDFake. Get returns taskcoredomain.ErrNotFound
// until OnGet or FailGet configures a response. Retry/Gate return
// errNotImplemented until OnRetry/FailRetry or OnGate/FailGate are set.
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
func (f *TaskCRUDFake) OnGet(task *taskcoredomain.Task) {
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

// FailRetry configures RequestTaskRetry to return err.
//
//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (f *TaskCRUDFake) FailRetry(err error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.retryErr = err
}

// OnRetry configures RequestTaskRetry to return task.
//
//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (f *TaskCRUDFake) OnRetry(task *taskcoredomain.Task) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.retryTask = task
}

// RetryCalls returns a copy of recorded RequestTaskRetry calls.
//
//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (f *TaskCRUDFake) RetryCalls() []RetryCall {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := make([]RetryCall, len(f.retryCalls))
	copy(out, f.retryCalls)
	return out
}

// FailGate configures ApplyTaskGateAction to return err.
//
//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (f *TaskCRUDFake) FailGate(err error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.gateErr = err
}

// OnGate configures ApplyTaskGateAction to return task.
//
//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (f *TaskCRUDFake) OnGate(task *taskcoredomain.Task) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.gateTask = task
}

// GateCalls returns a copy of recorded ApplyTaskGateAction calls.
//
//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (f *TaskCRUDFake) GateCalls() []GateCall {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := make([]GateCall, len(f.gateCalls))
	copy(out, f.gateCalls)
	return out
}

// Get records the call and returns the configured outcome.
//
//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (f *TaskCRUDFake) Get(ctx context.Context, id string) (*taskcoredomain.Task, error) {
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
	return nil, taskcoredomain.ErrNotFound
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (f *TaskCRUDFake) Create(ctx context.Context, in taskcorecontract.CreateTaskInput, by taskcoredomain.Actor) (*taskcoredomain.Task, error) {
	return nil, errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (f *TaskCRUDFake) Update(ctx context.Context, id string, in taskcorecontract.UpdateTaskInput, by taskcoredomain.Actor) (*taskcoredomain.Task, error) {
	return nil, errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (f *TaskCRUDFake) Delete(ctx context.Context, id string, by taskcoredomain.Actor) ([]string, error) {
	return nil, errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (f *TaskCRUDFake) ListFlatPage(ctx context.Context, limit, offset int, filter *taskcorecontract.ListFilter) ([]taskcoredomain.Task, bool, error) {
	return nil, false, errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (f *TaskCRUDFake) ListFlatAfter(ctx context.Context, limit int, afterID string) ([]taskcoredomain.Task, bool, error) {
	return nil, false, errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (f *TaskCRUDFake) TaskStats(ctx context.Context) (taskcorecontract.TaskStats, error) {
	return taskcorecontract.TaskStats{}, errNotImplemented
}

// RequestTaskRetry records the call and returns the configured outcome.
//
//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (f *TaskCRUDFake) RequestTaskRetry(ctx context.Context, in taskcorecontract.RequestRetryInput, by taskcoredomain.Actor) (*taskcoredomain.Task, error) {
	f.mu.Lock()
	f.retryCalls = append(f.retryCalls, RetryCall{Input: in, By: by})
	err := f.retryErr
	task := f.retryTask
	f.mu.Unlock()
	if err != nil {
		return nil, err
	}
	if task != nil {
		return task, nil
	}
	return nil, errNotImplemented
}

// RequestTaskApprove records the call and returns the configured outcome.
//
//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (f *TaskCRUDFake) RequestTaskApprove(ctx context.Context, taskID string, by taskcoredomain.Actor) (*taskcoredomain.Task, error) {
	f.mu.Lock()
	f.approveCalls = append(f.approveCalls, ApproveCall{TaskID: taskID, By: by})
	err := f.approveErr
	task := f.approveTask
	f.mu.Unlock()
	if err != nil {
		return nil, err
	}
	if task != nil {
		return task, nil
	}
	return nil, errNotImplemented
}

// FailPolish configures RequestTaskPolish to return err.
//
//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (f *TaskCRUDFake) FailPolish(err error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.polishErr = err
}

// OnPolish configures RequestTaskPolish to return task.
//
//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (f *TaskCRUDFake) OnPolish(task *taskcoredomain.Task) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.polishTask = task
}

// PolishCalls returns a copy of recorded RequestTaskPolish calls.
//
//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (f *TaskCRUDFake) PolishCalls() []PolishCall {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := make([]PolishCall, len(f.polishCalls))
	copy(out, f.polishCalls)
	return out
}

// RequestTaskPolish records the call and returns the configured outcome.
//
//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (f *TaskCRUDFake) RequestTaskPolish(ctx context.Context, in taskcorecontract.RequestPolishInput, by taskcoredomain.Actor) (*taskcoredomain.Task, error) {
	f.mu.Lock()
	f.polishCalls = append(f.polishCalls, PolishCall{Input: in, By: by})
	err := f.polishErr
	task := f.polishTask
	f.mu.Unlock()
	if err != nil {
		return nil, err
	}
	if task != nil {
		return task, nil
	}
	return nil, errNotImplemented
}

// ApplyTaskGateAction records the call and returns the configured outcome.
//
//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (f *TaskCRUDFake) ApplyTaskGateAction(ctx context.Context, taskID string, action taskcorecontract.GateAction, by taskcoredomain.Actor) (*taskcoredomain.Task, error) {
	f.mu.Lock()
	f.gateCalls = append(f.gateCalls, GateCall{TaskID: taskID, Action: action, By: by})
	err := f.gateErr
	task := f.gateTask
	f.mu.Unlock()
	if err != nil {
		return nil, err
	}
	if task != nil {
		return task, nil
	}
	return nil, errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (f *TaskCRUDFake) ValidateTaskWorktreeBinding(ctx context.Context, projectID *string, worktreeID string) error {
	return errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (f *TaskCRUDFake) ListTaskDependencies(ctx context.Context, taskID string) ([]taskcoredomain.DependencyEdge, error) {
	return nil, errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (f *TaskCRUDFake) AddTaskDependency(ctx context.Context, taskID, dependsOnTaskID string, satisfies taskcoredomain.DependencySatisfies) error {
	return errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (f *TaskCRUDFake) RemoveTaskDependency(ctx context.Context, taskID, dependsOnTaskID string) error {
	return errNotImplemented
}

var (
	_ taskcorecontract.TaskGetter    = (*TaskCRUDFake)(nil)
	_ taskcorecontract.TaskReader    = (*TaskCRUDFake)(nil)
	_ taskcorecontract.TaskWriter    = (*TaskCRUDFake)(nil)
	_ taskcorecontract.TaskDepsStore = (*TaskCRUDFake)(nil)
	_ taskcorecontract.TaskOpsStore  = (*TaskCRUDFake)(nil)
	_ taskcorecontract.TaskCRUDStore = (*TaskCRUDFake)(nil)
)
