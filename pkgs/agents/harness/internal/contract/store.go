// Package contract defines the persistence surface required by harness
// production code and harness test doubles. It lives in internal/contract
// so harness root and internal/{verify,resume,git} can share the type
// without an import cycle.
package contract

import (
	"context"

	projectsdomain "github.com/AlexsanderHamir/Hamix/pkgs/projects/domain"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/domain"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/store"
)

// Store is the persistence contract for harness orchestration, verify,
// resume, and git subpackages.
type Store interface {
	// Tasks
	Create(ctx context.Context, in store.CreateTaskInput, by domain.Actor) (*domain.Task, error)
	Get(ctx context.Context, id string) (*domain.Task, error)
	Update(ctx context.Context, id string, in store.UpdateTaskInput, by domain.Actor) (*domain.Task, error)

	// Cycles
	StartCycle(ctx context.Context, in store.StartCycleInput) (*domain.TaskCycle, error)
	GetCycle(ctx context.Context, cycleID string) (*domain.TaskCycle, error)
	ListCyclesForTask(ctx context.Context, taskID string, limit int) ([]domain.TaskCycle, error)
	TerminateCycle(ctx context.Context, cycleID string, status domain.CycleStatus, reason string, by domain.Actor) (*domain.TaskCycle, error)

	// Phases
	StartPhase(ctx context.Context, cycleID string, phase domain.Phase, by domain.Actor) (*domain.TaskCyclePhase, error)
	CompletePhase(ctx context.Context, in store.CompletePhaseInput) (*domain.TaskCyclePhase, error)
	ListPhasesForCycle(ctx context.Context, cycleID string) ([]domain.TaskCyclePhase, error)
	LastSessionID(ctx context.Context, cycleID string, phase domain.Phase) (string, error)

	// Verify reports
	ListCriteriaReportsForCycle(ctx context.Context, cycleID string) ([]domain.TaskCycleCriteriaReport, error)
	UpsertCriteriaReports(ctx context.Context, cycleID string, attemptSeq int64, entries []store.CriteriaReportEntry) error
	ListVerifyReportsForCycle(ctx context.Context, cycleID string) ([]domain.TaskCycleVerifyReport, error)
	UpsertVerifyReports(ctx context.Context, cycleID string, attemptSeq int64, entries []store.VerifyReportEntry) error
	UpsertCommandRuns(ctx context.Context, cycleID string, attemptSeq int64, entries []store.CommandRunEntry) error

	// Commits
	ListCommitsForTask(ctx context.Context, taskID string) ([]domain.TaskCycleCommit, error)
	ListCommitsForCycle(ctx context.Context, cycleID string) ([]domain.TaskCycleCommit, error)
	UpsertCycleCommits(ctx context.Context, taskID, cycleID string, entries []store.CycleCommitEntry) error

	// Checklist
	AddChecklistItem(ctx context.Context, taskID, text string, verifyCommands []store.VerifyCommandInput, by domain.Actor) (*domain.TaskChecklistItem, error)
	ListChecklistForVerify(ctx context.Context, taskID string) ([]store.ChecklistVerifyItem, error)
	ListChecklistForSubject(ctx context.Context, taskID string) ([]store.ChecklistItemView, error)
	SetChecklistItemDone(ctx context.Context, subjectTaskID, itemID string, done bool, by domain.Actor) error
	SetChecklistItemDoneWithEvidence(ctx context.Context, subjectTaskID, itemID string, evidence string, verifier domain.VerifierKind, reasoning, cycleID string, by domain.Actor) error

	// Snapshots
	GetTaskContextSnapshotForCycle(ctx context.Context, cycleID string) (domain.TaskContextSnapshot, error)
	CreateTaskContextSnapshot(ctx context.Context, input store.CreateTaskContextSnapshotInput) (domain.TaskContextSnapshot, error)

	// Settings
	GetSettings(ctx context.Context) (store.AppSettings, error)
	UpdateSettings(ctx context.Context, patch store.SettingsPatch) (store.AppSettings, error)

	// Events
	AppendCycleStreamEvent(ctx context.Context, in store.AppendCycleStreamEventInput) (*domain.TaskCycleStreamEvent, error)
	ListCycleStreamEvents(ctx context.Context, cycleID string, afterSeq int64, limit int) ([]domain.TaskCycleStreamEvent, error)
	AppendTaskEvent(ctx context.Context, taskID string, typ domain.EventType, by domain.Actor, data []byte) error
	ListTaskEvents(ctx context.Context, taskID string) ([]domain.TaskEvent, error)

	// Projects
	GetProject(ctx context.Context, id string) (projectsdomain.Project, error)
	ListProjectContextByIDs(ctx context.Context, projectID string, ids []string) ([]projectsdomain.ProjectContextItem, error)
	ListProjectContextEdges(ctx context.Context, projectID string, nodeIDs []string) ([]projectsdomain.ProjectContextEdge, error)
}
