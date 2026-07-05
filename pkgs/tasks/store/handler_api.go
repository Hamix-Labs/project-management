package store

import (
	"context"
	"encoding/json"

	"github.com/AlexsanderHamir/Hamix/pkgs/gitwork"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/domain"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/store/internal/checklist"
)

// HandlerAPI is the persistence contract required by pkgs/tasks/handler.
// *Store implements it; tests may pass a narrower fake.
type HandlerAPI interface {
	// Health
	Ready(ctx context.Context) error
	CountGitRepositories(ctx context.Context) (int64, error)

	// Settings
	GetSettings(ctx context.Context) (AppSettings, error)
	UpdateSettings(ctx context.Context, patch SettingsPatch) (AppSettings, error)

	// Tasks
	Get(ctx context.Context, id string) (*domain.Task, error)
	Create(ctx context.Context, in CreateTaskInput, by domain.Actor) (*domain.Task, error)
	Update(ctx context.Context, id string, in UpdateTaskInput, by domain.Actor) (*domain.Task, error)
	Delete(ctx context.Context, id string, by domain.Actor) ([]string, error)
	ListFlatPage(ctx context.Context, limit, offset int, filter *ListFilter) ([]domain.Task, bool, error)
	ListFlatAfter(ctx context.Context, limit int, afterID string) ([]domain.Task, bool, error)
	TaskStats(ctx context.Context) (TaskStats, error)
	RequestTaskRetry(ctx context.Context, in RequestRetryInput, by domain.Actor) (*domain.Task, error)
	ApplyTaskGateAction(ctx context.Context, taskID, action string, by domain.Actor) (*domain.Task, error)
	ValidateTaskWorktreeBinding(ctx context.Context, projectID *string, worktreeID string) error
	ListTaskDependencies(ctx context.Context, taskID string) ([]domain.DependencyEdge, error)
	AddTaskDependency(ctx context.Context, taskID, dependsOnTaskID string, satisfies domain.DependencySatisfies) error
	RemoveTaskDependency(ctx context.Context, taskID, dependsOnTaskID string) error
	ListCycleFailures(ctx context.Context, in ListCycleFailuresInput) (ListCycleFailuresResult, error)

	// Task events
	GetTaskEvent(ctx context.Context, taskID string, seq int64) (*domain.TaskEvent, error)
	ListTaskEvents(ctx context.Context, taskID string) ([]domain.TaskEvent, error)
	ListTaskEventsPageCursor(ctx context.Context, taskID string, limit int, beforeSeq, afterSeq *int64) (*TaskEventsPage, error)
	ApprovalPending(ctx context.Context, taskID string) (bool, error)
	AppendTaskEventResponseMessage(ctx context.Context, taskID string, seq int64, text string, by domain.Actor) error

	// Checklist
	ListChecklistForSubject(ctx context.Context, taskID string) ([]ChecklistItemView, error)
	IsTaskCycleRunning(ctx context.Context, taskID string) (bool, error)
	AddChecklistItem(ctx context.Context, taskID, text string, verifyCommands []checklist.VerifyCommandInput, by domain.Actor) (*domain.TaskChecklistItem, error)
	UpdateChecklistItemText(ctx context.Context, taskID, itemID, text string, by domain.Actor) error
	ReplaceChecklistVerifyCommands(ctx context.Context, taskID, itemID string, cmds []checklist.VerifyCommandInput, by domain.Actor) error
	SetChecklistItemDoneWithEvidence(ctx context.Context, subjectTaskID, itemID string, evidence string, verifier domain.VerifierKind, reasoning, cycleID string, by domain.Actor) error
	SetChecklistItemDone(ctx context.Context, subjectTaskID, itemID string, done bool, by domain.Actor) error
	DeleteChecklistItem(ctx context.Context, taskID, itemID string, by domain.Actor) error

	// Cycles
	StartCycle(ctx context.Context, in StartCycleInput) (*domain.TaskCycle, error)
	GetCycle(ctx context.Context, cycleID string) (*domain.TaskCycle, error)
	ListCyclesForTaskBefore(ctx context.Context, taskID string, beforeAttemptSeq int64, limit int) ([]domain.TaskCycle, error)
	TerminateCycle(ctx context.Context, cycleID string, status domain.CycleStatus, reason string, by domain.Actor) (*domain.TaskCycle, error)
	StartPhase(ctx context.Context, cycleID string, phase domain.Phase, by domain.Actor) (*domain.TaskCyclePhase, error)
	CompletePhase(ctx context.Context, in CompletePhaseInput) (*domain.TaskCyclePhase, error)
	ListPhasesForCycle(ctx context.Context, cycleID string) ([]domain.TaskCyclePhase, error)
	ListCycleStreamEvents(ctx context.Context, cycleID string, afterSeq int64, limit int) ([]domain.TaskCycleStreamEvent, error)
	ListCriteriaReportsForCycle(ctx context.Context, cycleID string) ([]domain.TaskCycleCriteriaReport, error)
	ListVerifyReportsForCycle(ctx context.Context, cycleID string) ([]domain.TaskCycleVerifyReport, error)
	ListCommandRunsForCycle(ctx context.Context, cycleID string) ([]domain.TaskCycleCommandRun, error)
	ListCommitsForCycle(ctx context.Context, cycleID string) ([]domain.TaskCycleCommit, error)
	ListCommitsForTask(ctx context.Context, taskID string) ([]domain.TaskCycleCommit, error)

	// Projects
	CreateProject(ctx context.Context, input CreateProjectInput) (domain.Project, error)
	ListProjects(ctx context.Context, includeArchived bool, limit int) ([]domain.Project, error)
	GetProject(ctx context.Context, id string) (domain.Project, error)
	UpdateProject(ctx context.Context, id string, input UpdateProjectInput) (domain.Project, error)
	DeleteProject(ctx context.Context, id string) error
	CreateProjectContext(ctx context.Context, projectID string, input CreateProjectContextInput) (domain.ProjectContextItem, error)
	ListProjectContext(ctx context.Context, projectID string, includeUnpinned bool, limit int) ([]domain.ProjectContextItem, error)
	ListProjectContextEdges(ctx context.Context, projectID string, nodeIDs []string) ([]domain.ProjectContextEdge, error)
	CreateProjectContextEdge(ctx context.Context, projectID string, input CreateProjectContextEdgeInput) (domain.ProjectContextEdge, error)
	UpdateProjectContextEdge(ctx context.Context, projectID, edgeID string, input UpdateProjectContextEdgeInput) (domain.ProjectContextEdge, error)
	DeleteProjectContextEdge(ctx context.Context, projectID, edgeID string) error
	UpdateProjectContext(ctx context.Context, projectID, itemID string, input UpdateProjectContextInput) (domain.ProjectContextItem, error)
	DeleteProjectContext(ctx context.Context, projectID, itemID string) error
	ListProjectsByRepository(ctx context.Context, repoID string) ([]domain.Project, error)

	// Drafts & templates
	ListDrafts(ctx context.Context, limit int) ([]DraftSummary, error)
	SaveDraft(ctx context.Context, id, name string, payload json.RawMessage) (*DraftSummary, error)
	GetDraft(ctx context.Context, id string) (*DraftDetail, error)
	DeleteDraft(ctx context.Context, id string) error
	ListTemplates(ctx context.Context, limit int, q, sort, order, tag string) ([]TemplateSummary, error)
	SaveTemplate(ctx context.Context, id, name string, payload json.RawMessage) (*TemplateSummary, error)
	GetTemplate(ctx context.Context, id string) (*TemplateDetail, error)
	PatchTemplate(ctx context.Context, id string, name *string, payload json.RawMessage) (*TemplateDetail, error)
	DeleteTemplate(ctx context.Context, id string) error
	IncrementTemplateInstantiateCounts(ctx context.Context, counts map[string]int) error

	// Git
	ListAllGitRepositories(ctx context.Context) ([]domain.GitRepository, error)
	ListAllGitRepositoriesWithSummary(ctx context.Context) ([]GitRepositoryListSummary, error)
	ListGitRepositories(ctx context.Context, projectID string) ([]domain.GitRepository, error)
	CreateGlobalGitRepository(ctx context.Context, input CreateGitRepositoryInput, gitSvc gitwork.Service) (domain.GitRepository, error)
	CreateGitRepository(ctx context.Context, projectID string, input CreateGitRepositoryInput, gitSvc gitwork.Service) (domain.GitRepository, error)
	GetGitRepositoryByID(ctx context.Context, repoID string) (domain.GitRepository, error)
	GetGitRepository(ctx context.Context, projectID, repoID string) (domain.GitRepository, error)
	DeleteGlobalGitRepository(ctx context.Context, repoID string) error
	DeleteGitRepository(ctx context.Context, projectID, repoID string) error
	ListGitWorktreesByRepo(ctx context.Context, repoID string) ([]domain.GitWorktree, error)
	ListGitWorktrees(ctx context.Context, projectID, repoID string) ([]domain.GitWorktree, error)
	CreateGitWorktreeForRepo(ctx context.Context, repoID string, input CreateGitWorktreeInput, gitSvc gitwork.Service) (domain.GitWorktree, error)
	CreateGitWorktree(ctx context.Context, projectID, repoID string, input CreateGitWorktreeInput, gitSvc gitwork.Service) (domain.GitWorktree, error)
	RemoveGitWorktreeFromDiskByID(ctx context.Context, worktreeID string, force bool, gitSvc gitwork.Service) error
	RemoveGitWorktreeFromDisk(ctx context.Context, projectID, worktreeID string, force bool, gitSvc gitwork.Service) error
	UnregisterGitWorktreeByID(ctx context.Context, worktreeID string) error
	UnregisterGitWorktree(ctx context.Context, projectID, worktreeID string) error
	ListGitBranchesByRepo(ctx context.Context, repoID string) ([]domain.GitBranch, error)
	ListGitBranches(ctx context.Context, projectID, repoID string) ([]domain.GitBranch, error)
	CreateGitBranch(ctx context.Context, projectID, repoID string, input CreateGitBranchInput, gitSvc gitwork.Service) (domain.GitBranch, error)
	DeleteGitBranch(ctx context.Context, projectID, branchID string, force bool, gitSvc gitwork.Service) error
	RepoWorktreeInventory(ctx context.Context, repo domain.GitRepository, gitSvc gitwork.Service) ([]WorktreeInventoryRow, error)
	RepoWorktreeCheckoutStatus(ctx context.Context, repo domain.GitRepository, gitSvc gitwork.Service) ([]WorktreeCheckoutStatusRow, error)
	ProbeGitWorktree(ctx context.Context, repoID, path string, gitSvc gitwork.Service) (GitWorktreeProbeResult, error)
	RegisterExistingGitWorktree(ctx context.Context, repoID, path, name string, bind BindBranchInput, gitSvc gitwork.Service) (domain.GitWorktree, error)
	ReconcileGitRepository(ctx context.Context, projectID, repoID string, input ReconcileGitInput, gitSvc gitwork.Service) (ReconcileGitOutput, error)
	RelocateGitRepository(ctx context.Context, projectID, repoID, path string, gitSvc gitwork.Service) (ReconcileGitOutput, error)
	RelocateGitWorktree(ctx context.Context, worktreeID, path string, gitSvc gitwork.Service) (domain.GitWorktree, error)
}

// HandlerAPI compliance is checked at compile time in handler/handler_store_assert_test.go.
