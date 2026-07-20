package tasks

import (
	"context"
	"log/slog"
	"time"

	"github.com/AlexsanderHamir/Hamix/pkgs/obs/calltrace"
	"github.com/AlexsanderHamir/Hamix/pkgs/taskcore/contract"
	"github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
	"gorm.io/gorm"
)

// ReadyForAgentPickup applies the same predicates as ListQueueCandidates for one task row.
func ReadyForAgentPickup(ctx context.Context, db *gorm.DB, t *domain.Task, now time.Time) (bool, contract.FailedPredicate, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.tasks.ReadyForAgentPickup")
	partial := contract.EvaluateWorkerReadiness(t, now, false)
	if !partial.Ready && partial.FailedPredicate != contract.FailedPredicateDependencies {
		return false, partial.FailedPredicate, nil
	}
	depsMet, err := DependenciesSatisfied(ctx, db, t.ID)
	if err != nil {
		return false, contract.FailedPredicateDependencies, err
	}
	result := contract.EvaluateWorkerReadiness(t, now, depsMet)
	return result.Ready, result.FailedPredicate, nil
}
