package store

import (
	"log/slog"

	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/calltrace"
	"gorm.io/gorm"
)

// Store is the GORM-backed persistence facade for git inventory
// (repositories, worktrees, branches, reconcile).
type Store struct {
	db *gorm.DB
}

// NewStore returns a Store backed by db.
func NewStore(db *gorm.DB) *Store {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "gitinventory.store.NewStore")
	return &Store{db: db}
}
