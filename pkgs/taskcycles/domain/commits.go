package domain

import "time"

// TaskCycleCommit is the durable worker-indexed record of one git commit
// declared by the agent in criteria-report.json and validated at execute ingest.
// See docs/domain/cycle-commits.md.
//
// Unique (cycle_id, sha) provides idempotent re-ingest across verify retries.
// cycle_id cascades on delete; task_id is denormalized for list-by-task queries.
type TaskCycleCommit struct {
	ID          string    `json:"id"`
	TaskID      string    `json:"task_id"`
	CycleID     string    `json:"cycle_id"`
	PhaseSeq    int64     `json:"phase_seq"`
	Seq         int64     `json:"seq"`
	Repo        string    `json:"repo"`
	Worktree    string    `json:"worktree"`
	Branch      string    `json:"branch"`
	SHA         string    `json:"sha"`
	CommittedAt time.Time `json:"committed_at"`
	Message     string    `json:"message"`
	RecordedAt  time.Time `json:"recorded_at"`
}
