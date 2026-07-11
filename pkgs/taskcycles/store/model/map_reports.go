package model

import cyclesdomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcycles/domain"

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func FromDomainTaskCycleCriteriaReport(d cyclesdomain.TaskCycleCriteriaReport) TaskCycleCriteriaReport {
	return TaskCycleCriteriaReport{
		ID:          d.ID,
		CycleID:     d.CycleID,
		AttemptSeq:  d.AttemptSeq,
		CriterionID: d.CriterionID,
		ClaimedDone: d.ClaimedDone,
		Evidence:    d.Evidence,
		WrittenAt:   d.WrittenAt,
	}
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func FromDomainTaskCycleCriteriaReports(rows []cyclesdomain.TaskCycleCriteriaReport) []TaskCycleCriteriaReport {
	if len(rows) == 0 {
		return nil
	}
	out := make([]TaskCycleCriteriaReport, len(rows))
	for i := range rows {
		out[i] = FromDomainTaskCycleCriteriaReport(rows[i])
	}
	return out
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func ToDomainTaskCycleCriteriaReport(m TaskCycleCriteriaReport) cyclesdomain.TaskCycleCriteriaReport {
	return cyclesdomain.TaskCycleCriteriaReport{
		ID:          m.ID,
		CycleID:     m.CycleID,
		AttemptSeq:  m.AttemptSeq,
		CriterionID: m.CriterionID,
		ClaimedDone: m.ClaimedDone,
		Evidence:    m.Evidence,
		WrittenAt:   m.WrittenAt,
	}
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func ToDomainTaskCycleCriteriaReportPtr(m *TaskCycleCriteriaReport) *cyclesdomain.TaskCycleCriteriaReport {
	if m == nil {
		return nil
	}
	d := ToDomainTaskCycleCriteriaReport(*m)
	return &d
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func ToDomainTaskCycleCriteriaReports(rows []TaskCycleCriteriaReport) []cyclesdomain.TaskCycleCriteriaReport {
	if len(rows) == 0 {
		return nil
	}
	out := make([]cyclesdomain.TaskCycleCriteriaReport, len(rows))
	for i := range rows {
		out[i] = ToDomainTaskCycleCriteriaReport(rows[i])
	}
	return out
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func FromDomainTaskCycleVerifyReport(d cyclesdomain.TaskCycleVerifyReport) TaskCycleVerifyReport {
	return TaskCycleVerifyReport{
		ID:           d.ID,
		CycleID:      d.CycleID,
		AttemptSeq:   d.AttemptSeq,
		CriterionID:  d.CriterionID,
		Verified:     d.Verified,
		VerifierKind: d.VerifierKind,
		Reasoning:    d.Reasoning,
		WrittenAt:    d.WrittenAt,
	}
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func FromDomainTaskCycleVerifyReports(rows []cyclesdomain.TaskCycleVerifyReport) []TaskCycleVerifyReport {
	if len(rows) == 0 {
		return nil
	}
	out := make([]TaskCycleVerifyReport, len(rows))
	for i := range rows {
		out[i] = FromDomainTaskCycleVerifyReport(rows[i])
	}
	return out
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func ToDomainTaskCycleVerifyReport(m TaskCycleVerifyReport) cyclesdomain.TaskCycleVerifyReport {
	return cyclesdomain.TaskCycleVerifyReport{
		ID:           m.ID,
		CycleID:      m.CycleID,
		AttemptSeq:   m.AttemptSeq,
		CriterionID:  m.CriterionID,
		Verified:     m.Verified,
		VerifierKind: m.VerifierKind,
		Reasoning:    m.Reasoning,
		WrittenAt:    m.WrittenAt,
	}
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func ToDomainTaskCycleVerifyReports(rows []TaskCycleVerifyReport) []cyclesdomain.TaskCycleVerifyReport {
	if len(rows) == 0 {
		return nil
	}
	out := make([]cyclesdomain.TaskCycleVerifyReport, len(rows))
	for i := range rows {
		out[i] = ToDomainTaskCycleVerifyReport(rows[i])
	}
	return out
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func FromDomainTaskCycleCommandRun(d cyclesdomain.TaskCycleCommandRun) TaskCycleCommandRun {
	return TaskCycleCommandRun{
		ID:          d.ID,
		CycleID:     d.CycleID,
		AttemptSeq:  d.AttemptSeq,
		CriterionID: d.CriterionID,
		CommandSeq:  d.CommandSeq,
		ExitCode:    d.ExitCode,
		MetaPath:    d.MetaPath,
		WrittenAt:   d.WrittenAt,
	}
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func FromDomainTaskCycleCommandRuns(rows []cyclesdomain.TaskCycleCommandRun) []TaskCycleCommandRun {
	if len(rows) == 0 {
		return nil
	}
	out := make([]TaskCycleCommandRun, len(rows))
	for i := range rows {
		out[i] = FromDomainTaskCycleCommandRun(rows[i])
	}
	return out
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func ToDomainTaskCycleCommandRun(m TaskCycleCommandRun) cyclesdomain.TaskCycleCommandRun {
	return cyclesdomain.TaskCycleCommandRun{
		ID:          m.ID,
		CycleID:     m.CycleID,
		AttemptSeq:  m.AttemptSeq,
		CriterionID: m.CriterionID,
		CommandSeq:  m.CommandSeq,
		ExitCode:    m.ExitCode,
		MetaPath:    m.MetaPath,
		WrittenAt:   m.WrittenAt,
	}
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func ToDomainTaskCycleCommandRuns(rows []TaskCycleCommandRun) []cyclesdomain.TaskCycleCommandRun {
	if len(rows) == 0 {
		return nil
	}
	out := make([]cyclesdomain.TaskCycleCommandRun, len(rows))
	for i := range rows {
		out[i] = ToDomainTaskCycleCommandRun(rows[i])
	}
	return out
}
