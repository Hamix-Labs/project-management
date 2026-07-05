package storefake

import (
	"context"
	"encoding/json"

	"github.com/AlexsanderHamir/Hamix/pkgs/gitwork"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/contract"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/domain"
)

// HandlerStoreFake composes TaskCRUDFake with stub implementations for the
// remaining HandlerStore slices. Use NewHandlerStore for handler tests that
// only exercise one store slice (e.g. GET /tasks/{id} → Get).
type HandlerStoreFake struct {
	*TaskCRUDFake
	unimplementedHandlerStore
}

// NewHandlerStore returns a HandlerStoreFake with an embedded TaskCRUDFake.
//
//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func NewHandlerStore() *HandlerStoreFake {
	return &HandlerStoreFake{TaskCRUDFake: NewTaskCRUD()}
}

// NewHandlerStoreFromTaskCRUD wraps an existing TaskCRUDFake.
//
//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func NewHandlerStoreFromTaskCRUD(crud *TaskCRUDFake) *HandlerStoreFake {
	if crud == nil {
		crud = NewTaskCRUD()
	}
	return &HandlerStoreFake{TaskCRUDFake: crud}
}

var _ contract.HandlerStore = (*HandlerStoreFake)(nil)

type unimplementedHandlerStore struct{}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (unimplementedHandlerStore) Ready(context.Context) error {
	return errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (unimplementedHandlerStore) CountGitRepositories(context.Context) (int64, error) {
	return 0, errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (unimplementedHandlerStore) GetSettings(context.Context) (domain.AppSettings, error) {
	return domain.AppSettings{}, errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (unimplementedHandlerStore) UpdateSettings(context.Context, contract.SettingsPatch) (domain.AppSettings, error) {
	return domain.AppSettings{}, errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (unimplementedHandlerStore) GetTaskEvent(context.Context, string, int64) (*domain.TaskEvent, error) {
	return nil, errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (unimplementedHandlerStore) ListTaskEvents(context.Context, string) ([]domain.TaskEvent, error) {
	return nil, errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (unimplementedHandlerStore) ListTaskEventsPageCursor(context.Context, string, int, *int64, *int64) (*contract.TaskEventsPage, error) {
	return nil, errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (unimplementedHandlerStore) ApprovalPending(context.Context, string) (bool, error) {
	return false, errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (unimplementedHandlerStore) AppendTaskEventResponseMessage(context.Context, string, int64, string, domain.Actor) error {
	return errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (unimplementedHandlerStore) ListChecklistForSubject(context.Context, string) ([]contract.ChecklistItemView, error) {
	return nil, errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (unimplementedHandlerStore) IsTaskCycleRunning(context.Context, string) (bool, error) {
	return false, errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (unimplementedHandlerStore) AddChecklistItem(context.Context, string, string, []contract.VerifyCommandInput, domain.Actor) (*domain.TaskChecklistItem, error) {
	return nil, errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (unimplementedHandlerStore) UpdateChecklistItemText(context.Context, string, string, string, domain.Actor) error {
	return errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (unimplementedHandlerStore) ReplaceChecklistVerifyCommands(context.Context, string, string, []contract.VerifyCommandInput, domain.Actor) error {
	return errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (unimplementedHandlerStore) SetChecklistItemDoneWithEvidence(context.Context, string, string, string, domain.VerifierKind, string, string, domain.Actor) error {
	return errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (unimplementedHandlerStore) SetChecklistItemDone(context.Context, string, string, bool, domain.Actor) error {
	return errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (unimplementedHandlerStore) DeleteChecklistItem(context.Context, string, string, domain.Actor) error {
	return errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (unimplementedHandlerStore) StartCycle(context.Context, contract.StartCycleInput) (*domain.TaskCycle, error) {
	return nil, errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (unimplementedHandlerStore) GetCycle(context.Context, string) (*domain.TaskCycle, error) {
	return nil, errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (unimplementedHandlerStore) ListCyclesForTaskBefore(context.Context, string, int64, int) ([]domain.TaskCycle, error) {
	return nil, errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (unimplementedHandlerStore) TerminateCycle(context.Context, string, domain.CycleStatus, string, domain.Actor) (*domain.TaskCycle, error) {
	return nil, errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (unimplementedHandlerStore) StartPhase(context.Context, string, domain.Phase, domain.Actor) (*domain.TaskCyclePhase, error) {
	return nil, errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (unimplementedHandlerStore) CompletePhase(context.Context, contract.CompletePhaseInput) (*domain.TaskCyclePhase, error) {
	return nil, errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (unimplementedHandlerStore) ListPhasesForCycle(context.Context, string) ([]domain.TaskCyclePhase, error) {
	return nil, errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (unimplementedHandlerStore) ListCycleStreamEvents(context.Context, string, int64, int) ([]domain.TaskCycleStreamEvent, error) {
	return nil, errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (unimplementedHandlerStore) ListCriteriaReportsForCycle(context.Context, string) ([]domain.TaskCycleCriteriaReport, error) {
	return nil, errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (unimplementedHandlerStore) ListVerifyReportsForCycle(context.Context, string) ([]domain.TaskCycleVerifyReport, error) {
	return nil, errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (unimplementedHandlerStore) ListCommandRunsForCycle(context.Context, string) ([]domain.TaskCycleCommandRun, error) {
	return nil, errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (unimplementedHandlerStore) ListCommitsForCycle(context.Context, string) ([]domain.TaskCycleCommit, error) {
	return nil, errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (unimplementedHandlerStore) ListCommitsForTask(context.Context, string) ([]domain.TaskCycleCommit, error) {
	return nil, errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (unimplementedHandlerStore) CreateProject(context.Context, contract.CreateProjectInput) (domain.Project, error) {
	return domain.Project{}, errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (unimplementedHandlerStore) ListProjects(context.Context, bool, int) ([]domain.Project, error) {
	return nil, errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (unimplementedHandlerStore) GetProject(context.Context, string) (domain.Project, error) {
	return domain.Project{}, errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (unimplementedHandlerStore) UpdateProject(context.Context, string, contract.UpdateProjectInput) (domain.Project, error) {
	return domain.Project{}, errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (unimplementedHandlerStore) DeleteProject(context.Context, string) error {
	return errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (unimplementedHandlerStore) CreateProjectContext(context.Context, string, contract.CreateProjectContextInput) (domain.ProjectContextItem, error) {
	return domain.ProjectContextItem{}, errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (unimplementedHandlerStore) ListProjectContext(context.Context, string, bool, int) ([]domain.ProjectContextItem, error) {
	return nil, errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (unimplementedHandlerStore) ListProjectContextEdges(context.Context, string, []string) ([]domain.ProjectContextEdge, error) {
	return nil, errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (unimplementedHandlerStore) CreateProjectContextEdge(context.Context, string, contract.CreateProjectContextEdgeInput) (domain.ProjectContextEdge, error) {
	return domain.ProjectContextEdge{}, errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (unimplementedHandlerStore) UpdateProjectContextEdge(context.Context, string, string, contract.UpdateProjectContextEdgeInput) (domain.ProjectContextEdge, error) {
	return domain.ProjectContextEdge{}, errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (unimplementedHandlerStore) DeleteProjectContextEdge(context.Context, string, string) error {
	return errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (unimplementedHandlerStore) UpdateProjectContext(context.Context, string, string, contract.UpdateProjectContextInput) (domain.ProjectContextItem, error) {
	return domain.ProjectContextItem{}, errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (unimplementedHandlerStore) DeleteProjectContext(context.Context, string, string) error {
	return errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (unimplementedHandlerStore) ListProjectsByRepository(context.Context, string) ([]domain.Project, error) {
	return nil, errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (unimplementedHandlerStore) ListDrafts(context.Context, int) ([]contract.DraftSummary, error) {
	return nil, errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (unimplementedHandlerStore) SaveDraft(context.Context, string, string, json.RawMessage) (*contract.DraftSummary, error) {
	return nil, errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (unimplementedHandlerStore) GetDraft(context.Context, string) (*contract.DraftDetail, error) {
	return nil, errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (unimplementedHandlerStore) DeleteDraft(context.Context, string) error {
	return errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (unimplementedHandlerStore) ListTemplates(context.Context, int, string, string, string, string) ([]contract.TemplateSummary, error) {
	return nil, errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (unimplementedHandlerStore) SaveTemplate(context.Context, string, string, json.RawMessage) (*contract.TemplateSummary, error) {
	return nil, errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (unimplementedHandlerStore) GetTemplate(context.Context, string) (*contract.TemplateDetail, error) {
	return nil, errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (unimplementedHandlerStore) PatchTemplate(context.Context, string, *string, json.RawMessage) (*contract.TemplateDetail, error) {
	return nil, errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (unimplementedHandlerStore) DeleteTemplate(context.Context, string) error {
	return errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (unimplementedHandlerStore) IncrementTemplateInstantiateCounts(context.Context, map[string]int) error {
	return errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (unimplementedHandlerStore) ListAllGitRepositories(context.Context) ([]domain.GitRepository, error) {
	return nil, errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (unimplementedHandlerStore) ListAllGitRepositoriesWithSummary(context.Context) ([]contract.GitRepositoryListSummary, error) {
	return nil, errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (unimplementedHandlerStore) ListGitRepositories(context.Context, string) ([]domain.GitRepository, error) {
	return nil, errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (unimplementedHandlerStore) GetGitRepositoryByID(context.Context, string) (domain.GitRepository, error) {
	return domain.GitRepository{}, errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (unimplementedHandlerStore) GetGitRepository(context.Context, string, string) (domain.GitRepository, error) {
	return domain.GitRepository{}, errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (unimplementedHandlerStore) DeleteGlobalGitRepository(context.Context, string) error {
	return errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (unimplementedHandlerStore) DeleteGitRepository(context.Context, string, string) error {
	return errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (unimplementedHandlerStore) ListGitWorktreesByRepo(context.Context, string) ([]domain.GitWorktree, error) {
	return nil, errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (unimplementedHandlerStore) ListGitWorktrees(context.Context, string, string) ([]domain.GitWorktree, error) {
	return nil, errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (unimplementedHandlerStore) UnregisterGitWorktreeByID(context.Context, string) error {
	return errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (unimplementedHandlerStore) UnregisterGitWorktree(context.Context, string, string) error {
	return errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (unimplementedHandlerStore) ListGitBranchesByRepo(context.Context, string) ([]domain.GitBranch, error) {
	return nil, errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (unimplementedHandlerStore) ListGitBranches(context.Context, string, string) ([]domain.GitBranch, error) {
	return nil, errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (unimplementedHandlerStore) CreateGlobalGitRepository(context.Context, contract.CreateGitRepositoryInput, gitwork.Service) (domain.GitRepository, error) {
	return domain.GitRepository{}, errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (unimplementedHandlerStore) CreateGitRepository(context.Context, string, contract.CreateGitRepositoryInput, gitwork.Service) (domain.GitRepository, error) {
	return domain.GitRepository{}, errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (unimplementedHandlerStore) CreateGitWorktreeForRepo(context.Context, string, contract.CreateGitWorktreeInput, gitwork.Service) (domain.GitWorktree, error) {
	return domain.GitWorktree{}, errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (unimplementedHandlerStore) CreateGitWorktree(context.Context, string, string, contract.CreateGitWorktreeInput, gitwork.Service) (domain.GitWorktree, error) {
	return domain.GitWorktree{}, errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (unimplementedHandlerStore) RemoveGitWorktreeFromDiskByID(context.Context, string, bool, gitwork.Service) error {
	return errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (unimplementedHandlerStore) RemoveGitWorktreeFromDisk(context.Context, string, string, bool, gitwork.Service) error {
	return errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (unimplementedHandlerStore) CreateGitBranch(context.Context, string, string, contract.CreateGitBranchInput, gitwork.Service) (domain.GitBranch, error) {
	return domain.GitBranch{}, errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (unimplementedHandlerStore) DeleteGitBranch(context.Context, string, string, bool, gitwork.Service) error {
	return errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (unimplementedHandlerStore) RepoWorktreeInventory(context.Context, domain.GitRepository, gitwork.Service) ([]contract.WorktreeInventoryRow, error) {
	return nil, errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (unimplementedHandlerStore) RepoWorktreeCheckoutStatus(context.Context, domain.GitRepository, gitwork.Service) ([]contract.WorktreeCheckoutStatusRow, error) {
	return nil, errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (unimplementedHandlerStore) ProbeGitWorktree(context.Context, string, string, gitwork.Service) (contract.GitWorktreeProbeResult, error) {
	return contract.GitWorktreeProbeResult{}, errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (unimplementedHandlerStore) RegisterExistingGitWorktree(context.Context, string, string, string, contract.BindBranchInput, gitwork.Service) (domain.GitWorktree, error) {
	return domain.GitWorktree{}, errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (unimplementedHandlerStore) ReconcileGitRepository(context.Context, string, string, contract.ReconcileGitInput, gitwork.Service) (contract.ReconcileGitOutput, error) {
	return contract.ReconcileGitOutput{}, errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (unimplementedHandlerStore) RelocateGitRepository(context.Context, string, string, string, gitwork.Service) (contract.ReconcileGitOutput, error) {
	return contract.ReconcileGitOutput{}, errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (unimplementedHandlerStore) RelocateGitWorktree(context.Context, string, string, gitwork.Service) (domain.GitWorktree, error) {
	return domain.GitWorktree{}, errNotImplemented
}
