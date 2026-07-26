package model

import (
	checklistdomain "github.com/AlexsanderHamir/Hamix/pkgs/taskchecklist/domain"
)

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func FromDomainTaskChecklistItem(d checklistdomain.TaskChecklistItem) TaskChecklistItem {
	return TaskChecklistItem{
		ID:        d.ID,
		TaskID:    d.TaskID,
		SortOrder: d.SortOrder,
		Text:      d.Text,
	}
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func FromDomainTaskChecklistItemPtr(d *checklistdomain.TaskChecklistItem) *TaskChecklistItem {
	if d == nil {
		return nil
	}
	m := FromDomainTaskChecklistItem(*d)
	return &m
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func ToDomainTaskChecklistItem(m TaskChecklistItem) checklistdomain.TaskChecklistItem {
	return checklistdomain.TaskChecklistItem{
		ID:        m.ID,
		TaskID:    m.TaskID,
		SortOrder: m.SortOrder,
		Text:      m.Text,
	}
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func ToDomainTaskChecklistItems(rows []TaskChecklistItem) []checklistdomain.TaskChecklistItem {
	if len(rows) == 0 {
		return nil
	}
	out := make([]checklistdomain.TaskChecklistItem, len(rows))
	for i := range rows {
		out[i] = ToDomainTaskChecklistItem(rows[i])
	}
	return out
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func FromDomainTaskChecklistCompletion(d checklistdomain.TaskChecklistCompletion) TaskChecklistCompletion {
	return TaskChecklistCompletion{
		TaskID:            d.TaskID,
		ItemID:            d.ItemID,
		At:                d.At,
		By:                d.By,
		Evidence:          d.Evidence,
		VerifiedBy:        d.VerifiedBy,
		VerifierReasoning: d.VerifierReasoning,
		CycleID:           d.CycleID,
	}
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func ToDomainTaskChecklistCompletion(m TaskChecklistCompletion) checklistdomain.TaskChecklistCompletion {
	return checklistdomain.TaskChecklistCompletion{
		TaskID:            m.TaskID,
		ItemID:            m.ItemID,
		At:                m.At,
		By:                m.By,
		Evidence:          m.Evidence,
		VerifiedBy:        m.VerifiedBy,
		VerifierReasoning: m.VerifierReasoning,
		CycleID:           m.CycleID,
	}
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func ToDomainTaskChecklistCompletions(rows []TaskChecklistCompletion) []checklistdomain.TaskChecklistCompletion {
	if len(rows) == 0 {
		return nil
	}
	out := make([]checklistdomain.TaskChecklistCompletion, len(rows))
	for i := range rows {
		out[i] = ToDomainTaskChecklistCompletion(rows[i])
	}
	return out
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func FromDomainTaskChecklistItemCommand(d checklistdomain.TaskChecklistItemCommand) TaskChecklistItemCommand {
	return TaskChecklistItemCommand{
		ID:              d.ID,
		ItemID:          d.ItemID,
		SortOrder:       d.SortOrder,
		Command:         d.Command,
		ExpectedOutcome: d.ExpectedOutcome,
		TimeoutSeconds:  cloneTimeoutSeconds(d.TimeoutSeconds),
	}
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func ToDomainTaskChecklistItemCommand(m TaskChecklistItemCommand) checklistdomain.TaskChecklistItemCommand {
	return checklistdomain.TaskChecklistItemCommand{
		ID:              m.ID,
		ItemID:          m.ItemID,
		SortOrder:       m.SortOrder,
		Command:         m.Command,
		ExpectedOutcome: m.ExpectedOutcome,
		TimeoutSeconds:  cloneTimeoutSeconds(m.TimeoutSeconds),
	}
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func cloneTimeoutSeconds(in *int) *int {
	if in == nil {
		return nil
	}
	v := *in
	return &v
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func ToDomainTaskChecklistItemCommands(rows []TaskChecklistItemCommand) []checklistdomain.TaskChecklistItemCommand {
	if len(rows) == 0 {
		return nil
	}
	out := make([]checklistdomain.TaskChecklistItemCommand, len(rows))
	for i := range rows {
		out[i] = ToDomainTaskChecklistItemCommand(rows[i])
	}
	return out
}
