package tasks

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/AlexsanderHamir/Hamix/pkgs/obs/calltrace"
	"github.com/AlexsanderHamir/Hamix/pkgs/storekernel"
	"github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
	"github.com/AlexsanderHamir/Hamix/pkgs/taskcore/store/model"
	taskeventsdomain "github.com/AlexsanderHamir/Hamix/pkgs/taskevents/domain"
	eventsaudit "github.com/AlexsanderHamir/Hamix/pkgs/taskevents/store/audit"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Close marks the task closed (idempotent if already closed). Clears
// pending_retry. Does not cancel runners — composition does that first.
func Close(ctx context.Context, db *gorm.DB, id string, by domain.Actor) (*domain.Task, error) {
	defer storekernel.DeferLatency(storekernel.OpUpdateTask)()
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.tasks.Close", "task_id", id)
	if err := domain.ValidateActor(by); err != nil {
		return nil, err
	}
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, fmt.Errorf("%w: id", domain.ErrInvalidInput)
	}
	var updated *domain.Task
	err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var cur model.Task
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", id).First(&cur).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return domain.ErrNotFound
			}
			return fmt.Errorf("load task: %w", err)
		}
		dcur := model.ToDomainTask(cur)
		if dcur.Status == domain.StatusClosed {
			if err := hydrateDependsOn(ctx, tx, &dcur); err != nil {
				return err
			}
			updated = &dcur
			return nil
		}
		nextSeq, err := eventsaudit.NextEventSeq(tx, id)
		if err != nil {
			return err
		}
		from := dcur.Status
		dcur.PendingRetry = nil
		dcur.Status = domain.StatusClosed
		b, err := storekernel.EventPairJSON(string(from), string(domain.StatusClosed))
		if err != nil {
			return err
		}
		if err := eventsaudit.AppendEvent(tx, id, nextSeq, taskeventsdomain.EventStatusChanged, by, b); err != nil {
			return err
		}
		cur = model.FromDomainTask(dcur)
		if err := tx.Save(&cur).Error; err != nil {
			return fmt.Errorf("save task: %w", err)
		}
		if err := hydrateDependsOn(ctx, tx, &dcur); err != nil {
			return err
		}
		updated = &dcur
		return nil
	})
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("close task: %w", err)
	}
	_ = hydrateCreatedAt(ctx, db, updated)
	return updated, nil
}

// Reopen transitions a closed task back to ready. Non-closed → ErrConflict.
func Reopen(ctx context.Context, db *gorm.DB, id string, by domain.Actor) (*domain.Task, error) {
	defer storekernel.DeferLatency(storekernel.OpUpdateTask)()
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "tasks.store.tasks.Reopen", "task_id", id)
	if err := domain.ValidateActor(by); err != nil {
		return nil, err
	}
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, fmt.Errorf("%w: id", domain.ErrInvalidInput)
	}
	var updated *domain.Task
	err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var cur model.Task
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", id).First(&cur).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return domain.ErrNotFound
			}
			return fmt.Errorf("load task: %w", err)
		}
		dcur := model.ToDomainTask(cur)
		if dcur.Status != domain.StatusClosed {
			return fmt.Errorf("%w: only closed tasks can be reopened", domain.ErrConflict)
		}
		nextSeq, err := eventsaudit.NextEventSeq(tx, id)
		if err != nil {
			return err
		}
		b, err := storekernel.EventPairJSON(string(domain.StatusClosed), string(domain.StatusReady))
		if err != nil {
			return err
		}
		if err := eventsaudit.AppendEvent(tx, id, nextSeq, taskeventsdomain.EventStatusChanged, by, b); err != nil {
			return err
		}
		dcur.Status = domain.StatusReady
		cur = model.FromDomainTask(dcur)
		if err := tx.Save(&cur).Error; err != nil {
			return fmt.Errorf("save task: %w", err)
		}
		if err := hydrateDependsOn(ctx, tx, &dcur); err != nil {
			return err
		}
		updated = &dcur
		return nil
	})
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, domain.ErrNotFound
		}
		if errors.Is(err, domain.ErrConflict) {
			return nil, err
		}
		return nil, fmt.Errorf("reopen task: %w", err)
	}
	_ = hydrateCreatedAt(ctx, db, updated)
	return updated, nil
}
