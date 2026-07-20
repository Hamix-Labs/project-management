package resume

import "github.com/AlexsanderHamir/Hamix/pkgs/obs/calltrace"
import (
	"context"
	"errors"
	"fmt"
	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
	taskcorestore "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/store"
	cyclescontract "github.com/AlexsanderHamir/Hamix/pkgs/taskcycles/contract"
	cyclesdomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcycles/domain"
	cyclesstore "github.com/AlexsanderHamir/Hamix/pkgs/taskcycles/store"
	"log/slog"
)

const crossCycleExecuteCarriedForward = "cross_cycle_execute_carried_forward"

// SeedCrossCycleExecuteFromParent records a succeeded execute phase on a new
// retry cycle so verify can start when the parent attempt already passed execute.
func (s *Service) SeedCrossCycleExecuteFromParent(ctx context.Context, cycle *cyclesdomain.TaskCycle, parentCycleID string) error {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "agent.harness.resume.SeedCrossCycleExecuteFromParent",
		"cycle_id", cycle.ID, "parent_cycle_id", parentCycleID)
	phases, err := s.store.ListPhasesForCycle(ctx, parentCycleID)
	if err != nil {
		return err
	}
	var parentExec *cyclesdomain.TaskCyclePhase
	for i := range phases {
		p := &phases[i]
		if p.Phase != cyclesdomain.PhaseExecute {
			continue
		}
		if parentExec == nil || p.PhaseSeq > parentExec.PhaseSeq {
			parentExec = p
		}
	}
	if parentExec == nil || parentExec.Status != cyclesdomain.PhaseStatusSucceeded {
		return fmt.Errorf("parent %q has no succeeded execute phase", parentCycleID)
	}
	exec, err := s.store.StartPhase(ctx, cycle.ID, cyclesdomain.PhaseExecute, taskcoredomain.ActorAgent)
	if err != nil {
		return err
	}
	summary := crossCycleExecuteCarriedForward
	details := []byte(parentExec.DetailsJSON)
	if len(details) == 0 {
		details = nil
	}
	_, err = s.store.CompletePhase(ctx, cyclescontract.CompletePhaseInput{
		CycleID: cycle.ID, PhaseSeq: exec.PhaseSeq,
		Status: cyclesdomain.PhaseStatusSucceeded, Summary: &summary,
		Details: details, By: taskcoredomain.ActorAgent,
	})
	return err
}

// MirrorParentCriteriaForVerifyOnly copies parent execute criteria reports into a child cycle.
func (s *Service) MirrorParentCriteriaForVerifyOnly(ctx context.Context, childCycleID, parentCycleID string) error {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "agent.harness.resume.MirrorParentCriteriaForVerifyOnly",
		"child_cycle_id", childCycleID, "parent_cycle_id", parentCycleID)
	rows, err := s.store.ListCriteriaReportsForCycle(ctx, parentCycleID)
	if err != nil {
		return err
	}
	byAttempt := map[int64][]cyclesstore.CriteriaReportEntry{}
	for _, row := range rows {
		byAttempt[row.AttemptSeq] = append(byAttempt[row.AttemptSeq], cyclesstore.CriteriaReportEntry{
			CriterionID: row.CriterionID, ClaimedDone: row.ClaimedDone, Evidence: row.Evidence,
		})
	}
	for attempt, entries := range byAttempt {
		if len(entries) == 0 {
			continue
		}
		if err := s.store.UpsertCriteriaReports(ctx, childCycleID, attempt, entries); err != nil {
			return err
		}
	}
	return nil
}

// FailTaskAfterRetryPrep marks the task failed when operator retry preparation fails.
func (s *Service) FailTaskAfterRetryPrep(ctx context.Context, taskID, reason string) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "agent.harness.resume.FailTaskAfterRetryPrep",
		"task_id", taskID, "reason", reason)
	failed := taskcoredomain.StatusFailed
	if _, err := s.store.Update(ctx, taskID, taskcorestore.UpdateTaskInput{Status: &failed}, taskcoredomain.ActorAgent); err != nil {
		level := slog.LevelWarn
		if errors.Is(err, taskcoredomain.ErrNotFound) {
			level = slog.LevelInfo
		}
		slog.Log(ctx, level, "agent harness retry prep task transition failed",
			"cmd", calltrace.LogCmd, "operation", "agent.harness.resume.FailTaskAfterRetryPrep.err",
			"task_id", taskID, "reason", reason, "err", err)
	}
}
