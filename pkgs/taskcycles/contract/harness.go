package contract

import (
	"context"
	"time"

	checklistdomain "github.com/AlexsanderHamir/Hamix/pkgs/taskchecklist/domain"
	cyclesdomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcycles/domain"
)

// CriteriaReportEntry is the per-criterion criteria-report upsert payload.
type CriteriaReportEntry struct {
	CriterionID string
	ClaimedDone bool
	Evidence    string
}

// VerifyReportEntry is the per-criterion verify-report upsert payload.
type VerifyReportEntry struct {
	CriterionID  string
	Verified     bool
	VerifierKind checklistdomain.VerifierKind
	Reasoning    string
}

// CommandRunEntry is one verify-phase shell command execution row.
type CommandRunEntry struct {
	CriterionID string
	CommandSeq  int64
	ExitCode    int
	MetaPath    string
}

// CycleCommitEntry is a commit upsert payload for cycle commit indexing.
type CycleCommitEntry struct {
	PhaseSeq    int64
	Seq         int64
	Repo        string
	Worktree    string
	Branch      string
	SHA         string
	CommittedAt time.Time
	Message     string
}

// AppendCycleStreamEventInput captures one durable per-attempt stream event.
type AppendCycleStreamEventInput struct {
	TaskID   string
	CycleID  string
	PhaseSeq int64
	At       time.Time
	Source   string
	Kind     string
	Subtype  string
	Message  string
	Tool     string
	Payload  []byte
}

// CycleHarnessStore extends CycleStore with harness/worker write and resume paths.
type CycleHarnessStore interface {
	CycleStore
	ListCyclesForTask(ctx context.Context, taskID string, limit int) ([]cyclesdomain.TaskCycle, error)
	LastSessionID(ctx context.Context, cycleID string, phase cyclesdomain.Phase) (string, error)
	UpsertCriteriaReports(ctx context.Context, cycleID string, attemptSeq int64, entries []CriteriaReportEntry) error
	UpsertVerifyReports(ctx context.Context, cycleID string, attemptSeq int64, entries []VerifyReportEntry) error
	UpsertCommandRuns(ctx context.Context, cycleID string, attemptSeq int64, entries []CommandRunEntry) error
	UpsertCycleCommits(ctx context.Context, taskID, cycleID string, entries []CycleCommitEntry) error
	AppendCycleStreamEvent(ctx context.Context, in AppendCycleStreamEventInput) (*cyclesdomain.TaskCycleStreamEvent, error)
	ListRunningCycles(ctx context.Context) ([]cyclesdomain.TaskCycle, error)
	ListRunningCyclePhases(ctx context.Context) ([]cyclesdomain.TaskCyclePhase, error)
	// PatchPhaseDetails shallow-merges patch into details_json for a
	// non-terminal phase row. See taskcycles/store/internal/cycles.PatchPhaseDetails
	// for merge / first-wins semantics (ADR-0031).
	PatchPhaseDetails(ctx context.Context, cycleID string, phaseSeq int64, patch []byte) error
}

// CycleWorkerStore lists in-flight cycles for reconcile and worker startup sweeps.
type CycleWorkerStore interface {
	ListRunningCycles(ctx context.Context) ([]cyclesdomain.TaskCycle, error)
	ListRunningCyclePhases(ctx context.Context) ([]cyclesdomain.TaskCyclePhase, error)
}
