package storefake

import (
	"context"

	checklistcontract "github.com/AlexsanderHamir/Hamix/pkgs/taskchecklist/contract"
	checklistdomain "github.com/AlexsanderHamir/Hamix/pkgs/taskchecklist/domain"
	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
)

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (unimplementedHandlerStore) ListChecklistForSubject(context.Context, string) ([]checklistcontract.ChecklistItemView, error) {
	return nil, errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (unimplementedHandlerStore) IsTaskCycleRunning(context.Context, string) (bool, error) {
	return false, errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (unimplementedHandlerStore) AddChecklistItem(context.Context, string, string, []checklistcontract.VerifyCommandInput, taskcoredomain.Actor) (*checklistdomain.TaskChecklistItem, error) {
	return nil, errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (unimplementedHandlerStore) UpdateChecklistItemText(context.Context, string, string, string, taskcoredomain.Actor) error {
	return errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (unimplementedHandlerStore) ReplaceChecklistVerifyCommands(context.Context, string, string, []checklistcontract.VerifyCommandInput, taskcoredomain.Actor) error {
	return errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (unimplementedHandlerStore) SetChecklistItemDoneWithEvidence(context.Context, string, string, string, checklistdomain.VerifierKind, string, string, taskcoredomain.Actor) error {
	return errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (unimplementedHandlerStore) SetChecklistItemDone(context.Context, string, string, bool, taskcoredomain.Actor) error {
	return errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (unimplementedHandlerStore) DeleteChecklistItem(context.Context, string, string, taskcoredomain.Actor) error {
	return errNotImplemented
}
