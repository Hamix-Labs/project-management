package storefake

import (
	"context"
	"encoding/json"

	gitcontract "github.com/AlexsanderHamir/Hamix/pkgs/gitinventory/contract"
	gitdomain "github.com/AlexsanderHamir/Hamix/pkgs/gitinventory/domain"
	"github.com/AlexsanderHamir/Hamix/pkgs/gitwork"
	projectscontract "github.com/AlexsanderHamir/Hamix/pkgs/projects/contract"
	projectsdomain "github.com/AlexsanderHamir/Hamix/pkgs/projects/domain"
	settingscontract "github.com/AlexsanderHamir/Hamix/pkgs/settings/contract"
	settingsdomain "github.com/AlexsanderHamir/Hamix/pkgs/settings/domain"
	checklistcontract "github.com/AlexsanderHamir/Hamix/pkgs/taskchecklist/contract"
	checklistdomain "github.com/AlexsanderHamir/Hamix/pkgs/taskchecklist/domain"
	composecontract "github.com/AlexsanderHamir/Hamix/pkgs/taskcompose/contract"
	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
	cyclescontract "github.com/AlexsanderHamir/Hamix/pkgs/taskcycles/contract"
	cyclesdomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcycles/domain"
	taskeventscontract "github.com/AlexsanderHamir/Hamix/pkgs/taskevents/contract"
	taskeventsdomain "github.com/AlexsanderHamir/Hamix/pkgs/taskevents/domain"
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
func (unimplementedHandlerStore) GetSettings(context.Context) (settingsdomain.AppSettings, error) {
	return settingsdomain.AppSettings{}, errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (unimplementedHandlerStore) UpdateSettings(context.Context, settingscontract.SettingsPatch) (settingsdomain.AppSettings, error) {
	return settingsdomain.AppSettings{}, errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (unimplementedHandlerStore) GetTaskEvent(context.Context, string, int64) (*taskeventsdomain.TaskEvent, error) {
	return nil, errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (unimplementedHandlerStore) ListTaskEvents(context.Context, string) ([]taskeventsdomain.TaskEvent, error) {
	return nil, errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (unimplementedHandlerStore) ListTaskEventsPageCursor(context.Context, string, int, *int64, *int64) (*taskeventscontract.TaskEventsPage, error) {
	return nil, errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (unimplementedHandlerStore) ApprovalPending(context.Context, string) (bool, error) {
	return false, errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (unimplementedHandlerStore) AppendTaskEventResponseMessage(context.Context, string, int64, string, taskcoredomain.Actor) error {
	return errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (unimplementedHandlerStore) ListChecklistForSubject(context.Context, string) ([]checklistcontract.ChecklistItemView, error) {
	return nil, errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (unimplementedHandlerStore) IsTaskCycleRunning(context.Context, string) (bool, error) {
	return false, errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (unimplementedHandlerStore) AddChecklistItem(context.Context, string, string, []checklistcontract.VerifyCommandInput, taskcoredomain.Actor) (*checklistdomain.TaskChecklistItem, error) {
	return nil, errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (unimplementedHandlerStore) UpdateChecklistItemText(context.Context, string, string, string, taskcoredomain.Actor) error {
	return errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (unimplementedHandlerStore) ReplaceChecklistVerifyCommands(context.Context, string, string, []checklistcontract.VerifyCommandInput, taskcoredomain.Actor) error {
	return errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (unimplementedHandlerStore) SetChecklistItemDoneWithEvidence(context.Context, string, string, string, checklistdomain.VerifierKind, string, string, taskcoredomain.Actor) error {
	return errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (unimplementedHandlerStore) SetChecklistItemDone(context.Context, string, string, bool, taskcoredomain.Actor) error {
	return errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (unimplementedHandlerStore) DeleteChecklistItem(context.Context, string, string, taskcoredomain.Actor) error {
	return errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (unimplementedHandlerStore) StartCycle(context.Context, cyclescontract.StartCycleInput) (*cyclesdomain.TaskCycle, error) {
	return nil, errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (unimplementedHandlerStore) GetCycle(context.Context, string) (*cyclesdomain.TaskCycle, error) {
	return nil, errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (unimplementedHandlerStore) ListCyclesForTaskBefore(context.Context, string, int64, int) ([]cyclesdomain.TaskCycle, error) {
	return nil, errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (unimplementedHandlerStore) TerminateCycle(context.Context, string, cyclesdomain.CycleStatus, string, taskcoredomain.Actor) (*cyclesdomain.TaskCycle, error) {
	return nil, errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (unimplementedHandlerStore) StartPhase(context.Context, string, cyclesdomain.Phase, taskcoredomain.Actor) (*cyclesdomain.TaskCyclePhase, error) {
	return nil, errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (unimplementedHandlerStore) CompletePhase(context.Context, cyclescontract.CompletePhaseInput) (*cyclesdomain.TaskCyclePhase, error) {
	return nil, errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (unimplementedHandlerStore) ListPhasesForCycle(context.Context, string) ([]cyclesdomain.TaskCyclePhase, error) {
	return nil, errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (unimplementedHandlerStore) ListCycleStreamEvents(context.Context, string, int64, int) ([]cyclesdomain.TaskCycleStreamEvent, error) {
	return nil, errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (unimplementedHandlerStore) ListCriteriaReportsForCycle(context.Context, string) ([]cyclesdomain.TaskCycleCriteriaReport, error) {
	return nil, errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (unimplementedHandlerStore) ListVerifyReportsForCycle(context.Context, string) ([]cyclesdomain.TaskCycleVerifyReport, error) {
	return nil, errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (unimplementedHandlerStore) ListCommandRunsForCycle(context.Context, string) ([]cyclesdomain.TaskCycleCommandRun, error) {
	return nil, errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (unimplementedHandlerStore) ListCommitsForCycle(context.Context, string) ([]cyclesdomain.TaskCycleCommit, error) {
	return nil, errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (unimplementedHandlerStore) ListCommitsForTask(context.Context, string) ([]cyclesdomain.TaskCycleCommit, error) {
	return nil, errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (unimplementedHandlerStore) CreateProject(context.Context, projectscontract.CreateProjectInput) (projectsdomain.Project, error) {
	return projectsdomain.Project{}, errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (unimplementedHandlerStore) ListProjects(context.Context, bool, int) ([]projectsdomain.Project, error) {
	return nil, errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (unimplementedHandlerStore) GetProject(context.Context, string) (projectsdomain.Project, error) {
	return projectsdomain.Project{}, errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (unimplementedHandlerStore) UpdateProject(context.Context, string, projectscontract.UpdateProjectInput) (projectsdomain.Project, error) {
	return projectsdomain.Project{}, errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (unimplementedHandlerStore) DeleteProject(context.Context, string) error {
	return errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (unimplementedHandlerStore) CreateProjectContext(context.Context, string, projectscontract.CreateProjectContextInput) (projectsdomain.ProjectContextItem, error) {
	return projectsdomain.ProjectContextItem{}, errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (unimplementedHandlerStore) ListProjectContext(context.Context, string, bool, int) ([]projectsdomain.ProjectContextItem, error) {
	return nil, errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (unimplementedHandlerStore) ListProjectContextEdges(context.Context, string, []string) ([]projectsdomain.ProjectContextEdge, error) {
	return nil, errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (unimplementedHandlerStore) CreateProjectContextEdge(context.Context, string, projectscontract.CreateProjectContextEdgeInput) (projectsdomain.ProjectContextEdge, error) {
	return projectsdomain.ProjectContextEdge{}, errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (unimplementedHandlerStore) UpdateProjectContextEdge(context.Context, string, string, projectscontract.UpdateProjectContextEdgeInput) (projectsdomain.ProjectContextEdge, error) {
	return projectsdomain.ProjectContextEdge{}, errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (unimplementedHandlerStore) DeleteProjectContextEdge(context.Context, string, string) error {
	return errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (unimplementedHandlerStore) UpdateProjectContext(context.Context, string, string, projectscontract.UpdateProjectContextInput) (projectsdomain.ProjectContextItem, error) {
	return projectsdomain.ProjectContextItem{}, errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (unimplementedHandlerStore) DeleteProjectContext(context.Context, string, string) error {
	return errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (unimplementedHandlerStore) ListProjectsByRepository(context.Context, string) ([]projectsdomain.Project, error) {
	return nil, errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (unimplementedHandlerStore) ListDrafts(context.Context, int) ([]composecontract.DraftSummary, error) {
	return nil, errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (unimplementedHandlerStore) SaveDraft(context.Context, string, string, json.RawMessage) (*composecontract.DraftSummary, error) {
	return nil, errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (unimplementedHandlerStore) GetDraft(context.Context, string) (*composecontract.DraftDetail, error) {
	return nil, errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (unimplementedHandlerStore) DeleteDraft(context.Context, string) error {
	return errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (unimplementedHandlerStore) ListTemplates(context.Context, int, string, string, string, string) ([]composecontract.TemplateSummary, error) {
	return nil, errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (unimplementedHandlerStore) SaveTemplate(context.Context, string, string, json.RawMessage) (*composecontract.TemplateSummary, error) {
	return nil, errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (unimplementedHandlerStore) GetTemplate(context.Context, string) (*composecontract.TemplateDetail, error) {
	return nil, errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (unimplementedHandlerStore) PatchTemplate(context.Context, string, *string, json.RawMessage) (*composecontract.TemplateDetail, error) {
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
func (unimplementedHandlerStore) ListAllGitRepositories(context.Context) ([]gitdomain.GitRepository, error) {
	return nil, errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (unimplementedHandlerStore) ListAllGitRepositoriesWithSummary(context.Context) ([]gitcontract.GitRepositoryListSummary, error) {
	return nil, errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (unimplementedHandlerStore) ListGitRepositories(context.Context, string) ([]gitdomain.GitRepository, error) {
	return nil, errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (unimplementedHandlerStore) GetGitRepositoryByID(context.Context, string) (gitdomain.GitRepository, error) {
	return gitdomain.GitRepository{}, errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (unimplementedHandlerStore) GetGitRepository(context.Context, string, string) (gitdomain.GitRepository, error) {
	return gitdomain.GitRepository{}, errNotImplemented
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
func (unimplementedHandlerStore) ListGitWorktreesByRepo(context.Context, string) ([]gitdomain.GitWorktree, error) {
	return nil, errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (unimplementedHandlerStore) ListGitWorktrees(context.Context, string, string) ([]gitdomain.GitWorktree, error) {
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
func (unimplementedHandlerStore) ListGitBranchesByRepo(context.Context, string) ([]gitdomain.GitBranch, error) {
	return nil, errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (unimplementedHandlerStore) ListGitBranches(context.Context, string, string) ([]gitdomain.GitBranch, error) {
	return nil, errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (unimplementedHandlerStore) CreateGlobalGitRepository(context.Context, gitcontract.CreateGitRepositoryInput, gitwork.Service) (gitdomain.GitRepository, error) {
	return gitdomain.GitRepository{}, errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (unimplementedHandlerStore) CreateGitRepository(context.Context, string, gitcontract.CreateGitRepositoryInput, gitwork.Service) (gitdomain.GitRepository, error) {
	return gitdomain.GitRepository{}, errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (unimplementedHandlerStore) CreateGitWorktreeForRepo(context.Context, string, gitcontract.CreateGitWorktreeInput, gitwork.Service) (gitdomain.GitWorktree, error) {
	return gitdomain.GitWorktree{}, errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (unimplementedHandlerStore) CreateGitWorktree(context.Context, string, string, gitcontract.CreateGitWorktreeInput, gitwork.Service) (gitdomain.GitWorktree, error) {
	return gitdomain.GitWorktree{}, errNotImplemented
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
func (unimplementedHandlerStore) CreateGitBranch(context.Context, string, string, gitcontract.CreateGitBranchInput, gitwork.Service) (gitdomain.GitBranch, error) {
	return gitdomain.GitBranch{}, errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (unimplementedHandlerStore) DeleteGitBranch(context.Context, string, string, bool, gitwork.Service) error {
	return errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (unimplementedHandlerStore) RepoWorktreeInventory(context.Context, gitdomain.GitRepository, gitwork.Service) ([]gitcontract.WorktreeInventoryRow, error) {
	return nil, errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (unimplementedHandlerStore) RepoWorktreeCheckoutStatus(context.Context, gitdomain.GitRepository, gitwork.Service) ([]gitcontract.WorktreeCheckoutStatusRow, error) {
	return nil, errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (unimplementedHandlerStore) ProbeGitWorktree(context.Context, string, string, gitwork.Service) (gitcontract.GitWorktreeProbeResult, error) {
	return gitcontract.GitWorktreeProbeResult{}, errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (unimplementedHandlerStore) RegisterExistingGitWorktree(context.Context, string, string, string, gitcontract.BindBranchInput, gitwork.Service) (gitdomain.GitWorktree, error) {
	return gitdomain.GitWorktree{}, errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (unimplementedHandlerStore) ReconcileGitRepository(context.Context, string, string, gitcontract.ReconcileGitInput, gitwork.Service) (gitcontract.ReconcileGitOutput, error) {
	return gitcontract.ReconcileGitOutput{}, errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (unimplementedHandlerStore) RelocateGitRepository(context.Context, string, string, string, gitwork.Service) (gitcontract.ReconcileGitOutput, error) {
	return gitcontract.ReconcileGitOutput{}, errNotImplemented
}

//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func (unimplementedHandlerStore) RelocateGitWorktree(context.Context, string, string, gitwork.Service) (gitdomain.GitWorktree, error) {
	return gitdomain.GitWorktree{}, errNotImplemented
}
