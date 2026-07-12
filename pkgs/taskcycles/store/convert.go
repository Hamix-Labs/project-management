package store

import (
	"github.com/AlexsanderHamir/Hamix/pkgs/taskcycles/contract"
	"github.com/AlexsanderHamir/Hamix/pkgs/taskcycles/store/internal/commits"
	"github.com/AlexsanderHamir/Hamix/pkgs/taskcycles/store/internal/cycles"
	"github.com/AlexsanderHamir/Hamix/pkgs/taskcycles/store/internal/reports"
)

//funclogmeasure:skip category=hot-path reason="Pure struct mapper; store method emits trace at the chokepoint."
func appendStreamIn(in contract.AppendCycleStreamEventInput) cycles.AppendStreamEventInput {
	return cycles.AppendStreamEventInput{
		TaskID:   in.TaskID,
		CycleID:  in.CycleID,
		PhaseSeq: in.PhaseSeq,
		At:       in.At,
		Source:   in.Source,
		Kind:     in.Kind,
		Subtype:  in.Subtype,
		Message:  in.Message,
		Tool:     in.Tool,
		Payload:  in.Payload,
	}
}

//funclogmeasure:skip category=hot-path reason="Pure struct mapper; store method emits trace at the chokepoint."
func cycleCommitEntries(in []contract.CycleCommitEntry) []commits.Entry {
	if len(in) == 0 {
		return nil
	}
	out := make([]commits.Entry, len(in))
	for i, e := range in {
		out[i] = commits.Entry{
			PhaseSeq:    e.PhaseSeq,
			Seq:         e.Seq,
			Repo:        e.Repo,
			Worktree:    e.Worktree,
			Branch:      e.Branch,
			SHA:         e.SHA,
			CommittedAt: e.CommittedAt,
			Message:     e.Message,
		}
	}
	return out
}

//funclogmeasure:skip category=hot-path reason="Pure struct mapper; store method emits trace at the chokepoint."
func criteriaReportEntries(in []contract.CriteriaReportEntry) []reports.CriteriaEntry {
	if len(in) == 0 {
		return nil
	}
	out := make([]reports.CriteriaEntry, len(in))
	for i, e := range in {
		out[i] = reports.CriteriaEntry{
			CriterionID: e.CriterionID,
			ClaimedDone: e.ClaimedDone,
			Evidence:    e.Evidence,
		}
	}
	return out
}

//funclogmeasure:skip category=hot-path reason="Pure struct mapper; store method emits trace at the chokepoint."
func verifyReportEntries(in []contract.VerifyReportEntry) []reports.VerifyEntry {
	if len(in) == 0 {
		return nil
	}
	out := make([]reports.VerifyEntry, len(in))
	for i, e := range in {
		out[i] = reports.VerifyEntry{
			CriterionID:  e.CriterionID,
			Verified:     e.Verified,
			VerifierKind: e.VerifierKind,
			Reasoning:    e.Reasoning,
		}
	}
	return out
}

//funclogmeasure:skip category=hot-path reason="Pure struct mapper; store method emits trace at the chokepoint."
func commandRunEntries(in []contract.CommandRunEntry) []reports.CommandRunEntry {
	if len(in) == 0 {
		return nil
	}
	out := make([]reports.CommandRunEntry, len(in))
	for i, e := range in {
		out[i] = reports.CommandRunEntry{
			CriterionID: e.CriterionID,
			CommandSeq:  e.CommandSeq,
			ExitCode:    e.ExitCode,
			MetaPath:    e.MetaPath,
		}
	}
	return out
}
