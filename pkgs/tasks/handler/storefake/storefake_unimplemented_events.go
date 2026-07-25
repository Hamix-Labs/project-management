package storefake

import (
	"context"

	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
	taskeventscontract "github.com/AlexsanderHamir/Hamix/pkgs/taskevents/contract"
	taskeventsdomain "github.com/AlexsanderHamir/Hamix/pkgs/taskevents/domain"
)

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (unimplementedHandlerStore) GetTaskEvent(context.Context, string, int64) (*taskeventsdomain.TaskEvent, error) {
	return nil, errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (unimplementedHandlerStore) ListTaskEvents(context.Context, string) ([]taskeventsdomain.TaskEvent, error) {
	return nil, errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (unimplementedHandlerStore) ListTaskEventsPageCursor(context.Context, string, int, *int64, *int64) (*taskeventscontract.TaskEventsPage, error) {
	return nil, errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (unimplementedHandlerStore) ApprovalPending(context.Context, string) (bool, error) {
	return false, errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (unimplementedHandlerStore) AppendTaskEventResponseMessage(context.Context, string, int64, string, taskcoredomain.Actor) error {
	return errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (unimplementedHandlerStore) ListTaskActivity(context.Context, taskeventscontract.ListActivityInput) (taskeventscontract.ListActivityResult, error) {
	return taskeventscontract.ListActivityResult{}, errNotImplemented
}
