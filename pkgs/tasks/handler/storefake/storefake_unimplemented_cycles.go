package storefake

import (
	"context"

	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
	cyclescontract "github.com/AlexsanderHamir/Hamix/pkgs/taskcycles/contract"
	cyclesdomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcycles/domain"
)

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (unimplementedHandlerStore) StartCycle(context.Context, cyclescontract.StartCycleInput) (*cyclesdomain.TaskCycle, error) {
	return nil, errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (unimplementedHandlerStore) GetCycle(context.Context, string) (*cyclesdomain.TaskCycle, error) {
	return nil, errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (unimplementedHandlerStore) ListCyclesForTaskBefore(context.Context, string, int64, int) ([]cyclesdomain.TaskCycle, error) {
	return nil, errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (unimplementedHandlerStore) TerminateCycle(context.Context, string, cyclesdomain.CycleStatus, string, taskcoredomain.Actor) (*cyclesdomain.TaskCycle, error) {
	return nil, errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (unimplementedHandlerStore) StartPhase(context.Context, string, cyclesdomain.Phase, taskcoredomain.Actor) (*cyclesdomain.TaskCyclePhase, error) {
	return nil, errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (unimplementedHandlerStore) CompletePhase(context.Context, cyclescontract.CompletePhaseInput) (*cyclesdomain.TaskCyclePhase, error) {
	return nil, errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (unimplementedHandlerStore) ListPhasesForCycle(context.Context, string) ([]cyclesdomain.TaskCyclePhase, error) {
	return nil, errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (unimplementedHandlerStore) ListCycleStreamEvents(context.Context, string, int64, int) ([]cyclesdomain.TaskCycleStreamEvent, error) {
	return nil, errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (unimplementedHandlerStore) ListCriteriaReportsForCycle(context.Context, string) ([]cyclesdomain.TaskCycleCriteriaReport, error) {
	return nil, errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (unimplementedHandlerStore) ListVerifyReportsForCycle(context.Context, string) ([]cyclesdomain.TaskCycleVerifyReport, error) {
	return nil, errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (unimplementedHandlerStore) ListCommandRunsForCycle(context.Context, string) ([]cyclesdomain.TaskCycleCommandRun, error) {
	return nil, errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (unimplementedHandlerStore) ListCommitsForCycle(context.Context, string) ([]cyclesdomain.TaskCycleCommit, error) {
	return nil, errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (unimplementedHandlerStore) ListCommitsForTask(context.Context, string) ([]cyclesdomain.TaskCycleCommit, error) {
	return nil, errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (unimplementedHandlerStore) ListCycleFailures(context.Context, cyclescontract.ListCycleFailuresInput) (cyclescontract.ListCycleFailuresResult, error) {
	return cyclescontract.ListCycleFailuresResult{}, errNotImplemented
}
