package verify

import (
	"context"
	"errors"
	"log/slog"

	"github.com/AlexsanderHamir/Hamix/pkgs/obs/calltrace"
	checklistcontract "github.com/AlexsanderHamir/Hamix/pkgs/taskchecklist/contract"
	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
)

// LoadSnapshot reads checklist criteria for claim-acceptance gating.
func (s *Service) LoadSnapshot(ctx context.Context, taskID string) (Snapshot, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "agent.harness.verify.LoadSnapshot",
		"task_id", taskID)
	items, err := s.store.ListChecklistForVerify(ctx, taskID)
	if err != nil {
		return Snapshot{}, err
	}
	return Snapshot{
		Enabled:  len(items) > 0,
		Criteria: items,
	}, nil
}

// CompleteChecklistLegacy marks every checklist item done when verify is disabled.
func (s *Service) CompleteChecklistLegacy(ctx context.Context, taskID string) error {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "agent.harness.verify.CompleteChecklistLegacy", "task_id", taskID)
	items, err := s.store.ListChecklistForSubject(ctx, taskID)
	if err != nil {
		return err
	}
	for _, it := range items {
		if it.Done {
			continue
		}
		if err := s.store.SetChecklistItemDone(ctx, taskID, it.ID, true, taskcoredomain.ActorAgent); err != nil {
			return err
		}
	}
	return nil
}

// ApplyVerifiedCompletions writes checklist completions for passed verdicts.
func (s *Service) ApplyVerifiedCompletions(ctx context.Context, taskID, cycleID string, verdicts []Verdict) error {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "agent.harness.verify.ApplyVerifiedCompletions",
		"task_id", taskID, "cycle_id", cycleID, "verdict_count", len(verdicts))
	for _, v := range verdicts {
		if !v.Passed {
			continue
		}
		err := s.store.SetChecklistItemDoneWithEvidence(ctx, checklistcontract.SetDoneWithEvidenceInput{
			TaskID:    taskID,
			ItemID:    v.ID,
			Evidence:  v.Evidence,
			Verifier:  v.Verifier,
			Reasoning: v.Reasoning,
			CycleID:   cycleID,
			By:        taskcoredomain.ActorAgent,
		})
		if err != nil && !errors.Is(err, taskcoredomain.ErrNotFound) {
			return err
		}
	}
	return nil
}
