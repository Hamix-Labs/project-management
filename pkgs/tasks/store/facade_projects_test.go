package store

import (
	"context"
	"encoding/json"
	"errors"
	"os/exec"
	"testing"

	"github.com/AlexsanderHamir/Hamix/internal/tasktestdb"
	"github.com/AlexsanderHamir/Hamix/pkgs/gitwork"
	projectsdomain "github.com/AlexsanderHamir/Hamix/pkgs/projects/domain"
	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
)

func newProjectsFacadeStore(t *testing.T) (*Store, context.Context) {
	t.Helper()
	return NewStore(tasktestdb.OpenSQLite(t)), context.Background()
}

func initGitRepoForProjects(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	runGitForProjects(t, dir, "init")
	runGitForProjects(t, dir, "config", "user.email", "test@example.com")
	runGitForProjects(t, dir, "config", "user.name", "Test")
	runGitForProjects(t, dir, "commit", "--allow-empty", "-m", "init")
	return dir
}

func runGitForProjects(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v in %s: %v\n%s", args, dir, err, out)
	}
}

func mustRepoDefaultProjectViaFacade(t *testing.T, s *Store, ctx context.Context) (repoID string, defaultProj projectsdomain.Project) {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not on PATH")
	}
	main := initGitRepoForProjects(t)
	repo, err := s.CreateGlobalGitRepository(ctx, CreateGitRepositoryInput{Path: main}, gitwork.New())
	if err != nil {
		t.Fatalf("CreateGlobalGitRepository: %v", err)
	}
	defaultProj, err = s.GetDefaultProjectForRepository(ctx, repo.ID)
	if err != nil {
		t.Fatalf("GetDefaultProjectForRepository: %v", err)
	}
	return repo.ID, defaultProj
}

func TestStore_CreateGlobalGitRepository_seedsDefaultProject(t *testing.T) {
	s, ctx := newProjectsFacadeStore(t)
	repoID, defaultProj := mustRepoDefaultProjectViaFacade(t, s, ctx)
	if !defaultProj.IsDefault || defaultProj.RepositoryID == nil || *defaultProj.RepositoryID != repoID {
		t.Fatalf("default project = %#v", defaultProj)
	}
	byRepo, err := s.ListProjectsByRepository(ctx, repoID)
	if err != nil {
		t.Fatalf("ListProjectsByRepository: %v", err)
	}
	if len(byRepo) != 1 || !byRepo[0].IsDefault {
		t.Fatalf("projects by repo = %#v", byRepo)
	}
}

func TestStore_CreateProject_requiresRepositoryID(t *testing.T) {
	s, ctx := newProjectsFacadeStore(t)
	if _, err := s.CreateProject(ctx, CreateProjectInput{Name: "No repo"}); !errors.Is(err, projectsdomain.ErrInvalidInput) {
		t.Fatalf("create without repo err = %v, want ErrInvalidInput", err)
	}
}

func TestStore_ProjectCRUD_roundtrip(t *testing.T) {
	s, ctx := newProjectsFacadeStore(t)
	repoID, defaultProj := mustRepoDefaultProjectViaFacade(t, s, ctx)

	project, err := s.CreateProject(ctx, CreateProjectInput{
		Name:           "Project moat",
		Description:    "Long-running project context",
		ContextSummary: "Shared memory for related tasks",
		RepositoryID:   &repoID,
	})
	if err != nil {
		t.Fatalf("create project: %v", err)
	}
	if project.ID == "" {
		t.Fatal("expected generated project id")
	}
	if project.Status != projectsdomain.ProjectStatusActive {
		t.Fatalf("status = %q, want active", project.Status)
	}

	got, err := s.GetProject(ctx, project.ID)
	if err != nil {
		t.Fatalf("get project: %v", err)
	}
	if got.Name != "Project moat" || got.ContextSummary == "" {
		t.Fatalf("project = %#v", got)
	}

	archived := projectsdomain.ProjectStatusArchived
	renamed := "Project context moat"
	updated, err := s.UpdateProject(ctx, project.ID, UpdateProjectInput{
		Name:   &renamed,
		Status: &archived,
	})
	if err != nil {
		t.Fatalf("update project: %v", err)
	}
	if updated.Name != renamed || updated.Status != archived {
		t.Fatalf("updated = %#v", updated)
	}

	active, err := s.ListProjects(ctx, false, 10)
	if err != nil {
		t.Fatalf("list active projects: %v", err)
	}
	if len(active) != 1 || active[0].ID != defaultProj.ID {
		t.Fatalf("active projects = %#v, want default only after archive", active)
	}

	all, err := s.ListProjects(ctx, true, 10)
	if err != nil {
		t.Fatalf("list all projects: %v", err)
	}
	if len(all) != 2 {
		t.Fatalf("all projects = %#v", all)
	}

	if err := s.DeleteProject(ctx, project.ID); err != nil {
		t.Fatalf("delete project: %v", err)
	}
	_, err = s.GetProject(ctx, project.ID)
	if !errors.Is(err, taskcoredomain.ErrNotFound) {
		t.Fatalf("get deleted err = %v, want ErrNotFound", err)
	}
}

func TestStore_DefaultProject_seededAndProtected(t *testing.T) {
	s, ctx := newProjectsFacadeStore(t)
	_, defaultProj := mustRepoDefaultProjectViaFacade(t, s, ctx)

	project, err := s.GetProject(ctx, defaultProj.ID)
	if err != nil {
		t.Fatalf("get default project: %v", err)
	}
	if project.Name != projectsdomain.DefaultProjectName || project.Status != projectsdomain.ProjectStatusActive {
		t.Fatalf("default project = %#v", project)
	}
	renamed := "Renamed default"
	if _, err := s.UpdateProject(ctx, defaultProj.ID, UpdateProjectInput{
		Name: &renamed,
	}); !errors.Is(err, projectsdomain.ErrConflict) {
		t.Fatalf("rename default err = %v, want ErrConflict", err)
	}
	archived := projectsdomain.ProjectStatusArchived
	if _, err := s.UpdateProject(ctx, defaultProj.ID, UpdateProjectInput{
		Status: &archived,
	}); !errors.Is(err, projectsdomain.ErrConflict) {
		t.Fatalf("archive default err = %v, want ErrConflict", err)
	}
	if err := s.DeleteProject(ctx, defaultProj.ID); !errors.Is(err, projectsdomain.ErrConflict) {
		t.Fatalf("delete default err = %v, want ErrConflict", err)
	}
}

func TestStore_ProjectContextCRUD_roundtrip(t *testing.T) {
	s, ctx := newProjectsFacadeStore(t)
	repoID, _ := mustRepoDefaultProjectViaFacade(t, s, ctx)
	project, err := s.CreateProject(ctx, CreateProjectInput{Name: "Context project", RepositoryID: &repoID})
	if err != nil {
		t.Fatalf("create project: %v", err)
	}

	first, err := s.CreateProjectContext(ctx, project.ID, CreateProjectContextInput{
		Kind:      projectsdomain.ProjectContextKindDecision,
		Title:     "Use relational memory first",
		Body:      "Defer embeddings until explicit context works.",
		CreatedBy: projectsdomain.ActorUser,
		Pinned:    true,
	})
	if err != nil {
		t.Fatalf("create first context: %v", err)
	}
	second, err := s.CreateProjectContext(ctx, project.ID, CreateProjectContextInput{
		Kind:      projectsdomain.ProjectContextKindNote,
		Title:     "Loose note",
		Body:      "Visible but not pinned.",
		CreatedBy: projectsdomain.ActorAgent,
	})
	if err != nil {
		t.Fatalf("create second context: %v", err)
	}

	pinned, err := s.ListProjectContext(ctx, project.ID, false, 10)
	if err != nil {
		t.Fatalf("list pinned context: %v", err)
	}
	if len(pinned) != 1 || pinned[0].ID != first.ID {
		t.Fatalf("pinned context = %#v", pinned)
	}

	all, err := s.ListProjectContext(ctx, project.ID, true, 10)
	if err != nil {
		t.Fatalf("list all context: %v", err)
	}
	if len(all) != 2 || all[0].ID != first.ID {
		t.Fatalf("all context = %#v", all)
	}

	pinSecond := true
	kind := projectsdomain.ProjectContextKindHandoff
	updated, err := s.UpdateProjectContext(ctx, project.ID, second.ID, UpdateProjectContextInput{
		Kind:   &kind,
		Pinned: &pinSecond,
	})
	if err != nil {
		t.Fatalf("update context: %v", err)
	}
	if updated.Kind != kind || !updated.Pinned {
		t.Fatalf("updated context = %#v", updated)
	}

	if err := s.DeleteProjectContext(ctx, project.ID, first.ID); err != nil {
		t.Fatalf("delete context: %v", err)
	}
	remaining, err := s.ListProjectContext(ctx, project.ID, true, 10)
	if err != nil {
		t.Fatalf("list remaining context: %v", err)
	}
	if len(remaining) != 1 || remaining[0].ID != second.ID {
		t.Fatalf("remaining context = %#v", remaining)
	}
}

func TestStore_ProjectContextEdges_roundtripAndValidation(t *testing.T) {
	s, ctx := newProjectsFacadeStore(t)
	repoID, _ := mustRepoDefaultProjectViaFacade(t, s, ctx)
	project, err := s.CreateProject(ctx, CreateProjectInput{Name: "Graph project", RepositoryID: &repoID})
	if err != nil {
		t.Fatalf("create project: %v", err)
	}
	first, err := s.CreateProjectContext(ctx, project.ID, CreateProjectContextInput{
		Kind:      projectsdomain.ProjectContextKindDecision,
		Title:     "Use explicit graph memory",
		Body:      "Nodes are project owned.",
		CreatedBy: projectsdomain.ActorUser,
	})
	if err != nil {
		t.Fatalf("create first context: %v", err)
	}
	second, err := s.CreateProjectContext(ctx, project.ID, CreateProjectContextInput{
		Kind:      projectsdomain.ProjectContextKindConstraint,
		Title:     "No hidden retrieval",
		Body:      "Tasks opt into selected nodes.",
		CreatedBy: projectsdomain.ActorUser,
	})
	if err != nil {
		t.Fatalf("create second context: %v", err)
	}

	edge, err := s.CreateProjectContextEdge(ctx, project.ID, CreateProjectContextEdgeInput{
		SourceContextID: first.ID,
		TargetContextID: second.ID,
		Relation:        projectsdomain.ProjectContextRelationSupports,
		Strength:        4,
		Note:            "Decision reinforces constraint",
	})
	if err != nil {
		t.Fatalf("create edge: %v", err)
	}
	if edge.ProjectID != project.ID || edge.SourceContextID != first.ID || edge.TargetContextID != second.ID {
		t.Fatalf("edge = %#v", edge)
	}

	edges, err := s.ListProjectContextEdges(ctx, project.ID, []string{first.ID, second.ID})
	if err != nil {
		t.Fatalf("list edges: %v", err)
	}
	if len(edges) != 1 || edges[0].ID != edge.ID {
		t.Fatalf("edges = %#v", edges)
	}

	strength := 5
	relation := projectsdomain.ProjectContextRelationRefines
	updated, err := s.UpdateProjectContextEdge(ctx, project.ID, edge.ID, UpdateProjectContextEdgeInput{
		Relation: &relation,
		Strength: &strength,
	})
	if err != nil {
		t.Fatalf("update edge: %v", err)
	}
	if updated.Relation != relation || updated.Strength != strength {
		t.Fatalf("updated edge = %#v", updated)
	}

	if _, err := s.CreateProjectContextEdge(ctx, project.ID, CreateProjectContextEdgeInput{
		SourceContextID: first.ID,
		TargetContextID: first.ID,
		Relation:        projectsdomain.ProjectContextRelationRelated,
		Strength:        3,
	}); !errors.Is(err, projectsdomain.ErrInvalidInput) {
		t.Fatalf("self edge err = %v, want ErrInvalidInput", err)
	}
	if _, err := s.CreateProjectContextEdge(ctx, project.ID, CreateProjectContextEdgeInput{
		SourceContextID: first.ID,
		TargetContextID: second.ID,
		Relation:        projectsdomain.ProjectContextRelationRelated,
		Strength:        6,
	}); !errors.Is(err, projectsdomain.ErrInvalidInput) {
		t.Fatalf("bad strength err = %v, want ErrInvalidInput", err)
	}

	if err := s.DeleteProjectContext(ctx, project.ID, first.ID); err != nil {
		t.Fatalf("delete context: %v", err)
	}
	edges, err = s.ListProjectContextEdges(ctx, project.ID, []string{first.ID, second.ID})
	if err != nil {
		t.Fatalf("list edges after context delete: %v", err)
	}
	if len(edges) != 0 {
		t.Fatalf("edges after context delete = %#v, want none", edges)
	}
}

func TestStore_TaskContextSnapshot_roundtrip(t *testing.T) {
	s, ctx := newProjectsFacadeStore(t)
	repoID, _ := mustRepoDefaultProjectViaFacade(t, s, ctx)
	project, err := s.CreateProject(ctx, CreateProjectInput{Name: "Snapshot project", RepositoryID: &repoID})
	if err != nil {
		t.Fatalf("create project: %v", err)
	}
	task := mustCreateTask(t, s, ctx)
	cycle, err := s.StartCycle(ctx, StartCycleInput{TaskID: task.ID, TriggeredBy: taskcoredomain.ActorAgent})
	if err != nil {
		t.Fatalf("start cycle: %v", err)
	}

	raw := json.RawMessage(`{"project_id":"` + project.ID + `","items":[{"id":"ctx-1"}]}`)
	snapshot, err := s.CreateTaskContextSnapshot(ctx, CreateTaskContextSnapshotInput{
		TaskID:          task.ID,
		CycleID:         cycle.ID,
		ProjectID:       project.ID,
		ContextJSON:     raw,
		RenderedContext: "## Project context\n- Use relational memory first.",
		TokenEstimate:   42,
	})
	if err != nil {
		t.Fatalf("create snapshot: %v", err)
	}
	if snapshot.ID == "" {
		t.Fatal("expected generated snapshot id")
	}

	got, err := s.GetTaskContextSnapshotForCycle(ctx, cycle.ID)
	if err != nil {
		t.Fatalf("get snapshot: %v", err)
	}
	if got.TaskID != task.ID || got.ProjectID != project.ID || got.TokenEstimate != 42 {
		t.Fatalf("snapshot = %#v", got)
	}
	if string(got.ContextJSON) != string(raw) {
		t.Fatalf("context_json = %s, want %s", string(got.ContextJSON), string(raw))
	}
}

func TestStore_Project_validation_errors(t *testing.T) {
	s, ctx := newProjectsFacadeStore(t)

	if _, err := s.CreateProject(ctx, CreateProjectInput{Name: " "}); !errors.Is(err, projectsdomain.ErrInvalidInput) {
		t.Fatalf("create empty name err = %v, want ErrInvalidInput", err)
	}

	repoID, _ := mustRepoDefaultProjectViaFacade(t, s, ctx)
	project, err := s.CreateProject(ctx, CreateProjectInput{Name: "Validation project", RepositoryID: &repoID})
	if err != nil {
		t.Fatalf("create project: %v", err)
	}
	custom, err := s.CreateProjectContext(ctx, project.ID, CreateProjectContextInput{
		Kind:      projectsdomain.ProjectContextKind("memory"),
		Title:     "Custom",
		Body:      "Custom kind",
		CreatedBy: projectsdomain.ActorUser,
	})
	if err != nil {
		t.Fatalf("create custom kind: %v", err)
	}
	blankKind := projectsdomain.ProjectContextKind(" ")
	if _, err := s.UpdateProjectContext(ctx, project.ID, custom.ID, UpdateProjectContextInput{
		Kind: &blankKind,
	}); !errors.Is(err, projectsdomain.ErrInvalidInput) {
		t.Fatalf("patch empty kind err = %v, want ErrInvalidInput", err)
	}

	_, err = s.GetProject(ctx, "missing")
	if !errors.Is(err, taskcoredomain.ErrNotFound) {
		t.Fatalf("get missing project err = %v, want ErrNotFound", err)
	}
}
