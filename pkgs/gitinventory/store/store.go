package store

import (
	"log/slog"

	"github.com/AlexsanderHamir/Hamix/pkgs/gitwork"
	"github.com/AlexsanderHamir/Hamix/pkgs/obs/calltrace"
	"gorm.io/gorm"
)

// Store is the GORM-backed persistence facade for git inventory
// (repositories, worktrees, branches, reconcile).
type Store struct {
	db  *gorm.DB
	git gitwork.Service
}

// NewStore returns a Store backed by db. When gitSvc is nil, a default
// gitwork.New() service is used (tests may pass an explicit stub).
func NewStore(db *gorm.DB, gitSvc gitwork.Service) *Store {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "gitinventory.store.NewStore")
	if gitSvc == nil {
		gitSvc = gitwork.New()
	}
	return &Store{db: db, git: gitSvc}
}

//funclogmeasure:skip category=hot-path reason="Pure accessor; git I/O traces emit at call sites."
func (s *Store) gitSvc() gitwork.Service {
	if s != nil && s.git != nil {
		return s.git
	}
	return gitwork.New()
}
