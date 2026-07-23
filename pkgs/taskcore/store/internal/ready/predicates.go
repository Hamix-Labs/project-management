package ready

import (
	"github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
	"gorm.io/gorm"
)

// SQL dequeuable predicates MUST stay aligned with taskcore/contract readiness helpers
// (and the pkgs/tasks/scheduling re-exports used by composition).
// Contract tests: store/scheduling_parity_test.go
//
//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func applyDequeuableTaskPredicates(q *gorm.DB, db *gorm.DB) *gorm.DB {
	q = q.Where("tasks.worktree_id IS NOT NULL AND tasks.worktree_id <> ''")
	q = q.Where(`NOT EXISTS (
		SELECT 1 FROM task_dependencies td
		INNER JOIN tasks dep ON dep.id = td.depends_on_task_id
		WHERE td.task_id = tasks.id AND dep.status <> ?
	)`, domain.StatusDone)
	if UseSQLiteEventRowID(db) {
		return q.Where("(tasks.gate IS NULL OR json_extract(tasks.gate, '$.status') = ?)", string(domain.GateStatusReleased))
	}
	return q.Where("(tasks.gate IS NULL OR tasks.gate->>'status' = ?)", string(domain.GateStatusReleased))
}
