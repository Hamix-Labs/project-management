package store

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/AlexsanderHamir/Hamix/pkgs/gitinventory/domain"
	"github.com/AlexsanderHamir/Hamix/pkgs/gitinventory/store/model"
	"github.com/AlexsanderHamir/Hamix/pkgs/gitwork"
	projectsdomain "github.com/AlexsanderHamir/Hamix/pkgs/projects/domain"
	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
	taskcoremodel "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/store/model"
	"github.com/google/uuid"
)

func TestReconcileGitRepository_needsBootstrapWhenPathMissing(t *testing.T) {
	s, ctx, gitSvc := gitTestStore(t)
	main := initGitRepo(t)
	repo, err := s.CreateGitRepository(ctx, projectsdomain.LegacyGlobalDefaultProjectID, CreateGitRepositoryInput{Path: main}, gitSvc)
	if err != nil {
		t.Fatal(err)
	}
	renamed := filepath.Join(filepath.Dir(main), "renamed-main")
	if err := os.Rename(main, renamed); err != nil {
		t.Fatalf("rename main: %v", err)
	}
	t.Cleanup(func() { _ = os.Rename(renamed, main) })

	out, err := s.ReconcileGitRepository(ctx, projectsdomain.LegacyGlobalDefaultProjectID, repo.ID, ReconcileGitInput{
		AllowRemove: true,
	}, gitSvc)
	if err != nil {
		t.Fatalf("ReconcileGitRepository: %v", err)
	}
	if out.Status != reconcileStatusNeedsBootstrapPath {
		t.Fatalf("status=%q want %q", out.Status, reconcileStatusNeedsBootstrapPath)
	}
}

func TestReconcileGitRepository_mainRenamed_autoDiscover(t *testing.T) {
	s, ctx, gitSvc := gitTestStore(t)
	main := initGitRepo(t)
	repo, err := s.CreateGitRepository(ctx, projectsdomain.LegacyGlobalDefaultProjectID, CreateGitRepositoryInput{Path: main}, gitSvc)
	if err != nil {
		t.Fatal(err)
	}
	renamed := filepath.Join(filepath.Dir(main), "renamed-auto")
	if err := os.Rename(main, renamed); err != nil {
		t.Fatalf("rename main: %v", err)
	}
	t.Cleanup(func() { _ = os.Rename(renamed, main) })

	out, err := s.ReconcileGitRepository(ctx, projectsdomain.LegacyGlobalDefaultProjectID, repo.ID, ReconcileGitInput{
		AllowCheckoutDiscover: true,
		RepairGit:             true,
		AllowRemove:           true,
	}, gitSvc)
	if err != nil {
		t.Fatalf("ReconcileGitRepository: %v", err)
	}
	if out.Status != reconcileStatusOK {
		t.Fatalf("status=%q want ok", out.Status)
	}
	if out.Report.ResolutionSource != gitwork.ResolveSourceDiscovered {
		t.Fatalf("resolution_source=%q want %q", out.Report.ResolutionSource, gitwork.ResolveSourceDiscovered)
	}
	if !out.Report.RepoPathUpdated {
		t.Fatal("expected repo path update")
	}
	gotRepo, err := s.GetGitRepository(ctx, projectsdomain.LegacyGlobalDefaultProjectID, repo.ID)
	if err != nil {
		t.Fatal(err)
	}
	if worktreePathKey(gotRepo.Path) != worktreePathKey(renamed) {
		t.Fatalf("repo path=%q want %q", gotRepo.Path, renamed)
	}
}

func TestReconcileGitRepository_mainRenamed_withLinkedWorktreeSibling(t *testing.T) {
	s, ctx, gitSvc := gitTestStore(t)
	main := initGitRepo(t)
	repo, err := s.CreateGitRepository(ctx, projectsdomain.LegacyGlobalDefaultProjectID, CreateGitRepositoryInput{Path: main}, gitSvc)
	if err != nil {
		t.Fatal(err)
	}
	parent := filepath.Dir(main)
	wtPath := filepath.Join(parent, "linked-sibling")
	repoGit, err := gitSvc.OpenRepository(ctx, main)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := gitSvc.AddWorktree(ctx, repoGit, wtPath, gitwork.AddWorktreeOptions{
		Branch:       "feature",
		CreateBranch: true,
	}); err != nil {
		t.Fatalf("AddWorktree: %v", err)
	}
	renamed := filepath.Join(parent, "renamed-with-linked")
	if err := os.Rename(main, renamed); err != nil {
		t.Fatalf("rename main: %v", err)
	}
	t.Cleanup(func() { _ = os.Rename(renamed, main) })

	out, err := s.ReconcileGitRepository(ctx, projectsdomain.LegacyGlobalDefaultProjectID, repo.ID, ReconcileGitInput{
		AllowCheckoutDiscover: true,
		RepairGit:             true,
		AllowRemove:           true,
	}, gitSvc)
	if err != nil {
		t.Fatalf("ReconcileGitRepository: %v", err)
	}
	if out.Status != reconcileStatusOK {
		t.Fatalf("status=%q want ok", out.Status)
	}
	gotRepo, err := s.GetGitRepository(ctx, projectsdomain.LegacyGlobalDefaultProjectID, repo.ID)
	if err != nil {
		t.Fatal(err)
	}
	if worktreePathKey(gotRepo.Path) != worktreePathKey(renamed) {
		t.Fatalf("repo path=%q want %q", gotRepo.Path, renamed)
	}
}

func TestReconcileGitRepository_mainRenamed_withBootstrap(t *testing.T) {
	s, ctx, gitSvc := gitTestStore(t)
	main := initGitRepo(t)
	repo, err := s.CreateGitRepository(ctx, projectsdomain.LegacyGlobalDefaultProjectID, CreateGitRepositoryInput{Path: main}, gitSvc)
	if err != nil {
		t.Fatal(err)
	}
	wtsBefore, err := s.ListGitWorktrees(ctx, projectsdomain.LegacyGlobalDefaultProjectID, repo.ID)
	if err != nil || len(wtsBefore) == 0 {
		t.Fatalf("worktrees before: %v len=%d", err, len(wtsBefore))
	}
	mainID := wtsBefore[0].ID

	renamed := filepath.Join(filepath.Dir(main), "renamed-main-bootstrap")
	if err := os.Rename(main, renamed); err != nil {
		t.Fatalf("rename main: %v", err)
	}
	t.Cleanup(func() { _ = os.Rename(renamed, main) })

	out, err := s.ReconcileGitRepository(ctx, projectsdomain.LegacyGlobalDefaultProjectID, repo.ID, ReconcileGitInput{
		BootstrapPath: renamed,
		RepairGit:     true,
		AllowRemove:   true,
	}, gitSvc)
	if err != nil {
		t.Fatalf("ReconcileGitRepository: %v", err)
	}
	if out.Status != reconcileStatusOK {
		t.Fatalf("status=%q want ok", out.Status)
	}
	if !out.Report.RepoPathUpdated {
		t.Fatal("expected repo path update")
	}

	gotRepo, err := s.GetGitRepository(ctx, projectsdomain.LegacyGlobalDefaultProjectID, repo.ID)
	if err != nil {
		t.Fatal(err)
	}
	if worktreePathKey(gotRepo.Path) != worktreePathKey(renamed) {
		t.Fatalf("repo path=%q want %q", gotRepo.Path, renamed)
	}
	gotWT, err := s.GetGitWorktree(ctx, projectsdomain.LegacyGlobalDefaultProjectID, mainID)
	if err != nil {
		t.Fatal(err)
	}
	if gotWT.ID != mainID {
		t.Fatalf("main worktree id changed: %q", gotWT.ID)
	}
	if worktreePathKey(gotWT.Path) != worktreePathKey(renamed) {
		t.Fatalf("main wt path=%q want %q", gotWT.Path, renamed)
	}
}

func TestReconcileGitRepository_linkedWorktreeMoved_preservesID(t *testing.T) {
	s, ctx, gitSvc := gitTestStore(t)
	main := initGitRepo(t)
	repo, err := s.CreateGitRepository(ctx, projectsdomain.LegacyGlobalDefaultProjectID, CreateGitRepositoryInput{Path: main}, gitSvc)
	if err != nil {
		t.Fatal(err)
	}
	wtPath := filepath.Join(filepath.Dir(main), "wt-move-src")
	wt, err := s.CreateGitWorktree(ctx, projectsdomain.LegacyGlobalDefaultProjectID, repo.ID, CreateGitWorktreeInput{
		Path:         wtPath,
		Branch:       "feature-move",
		CreateBranch: true,
	}, gitSvc)
	if err != nil {
		t.Fatalf("CreateGitWorktree: %v", err)
	}
	wtPath2 := filepath.Join(filepath.Dir(main), "wt-move-dst")
	runGitStore(t, main, "worktree", "move", wtPath, wtPath2)
	t.Cleanup(func() {
		_ = os.RemoveAll(wtPath2)
	})

	out, err := s.ReconcileGitRepository(ctx, projectsdomain.LegacyGlobalDefaultProjectID, repo.ID, ReconcileGitInput{
		RepairGit:   true,
		AllowRemove: true,
	}, gitSvc)
	if err != nil {
		t.Fatalf("ReconcileGitRepository: %v", err)
	}
	if out.Report.WorktreesPathUpdated < 1 {
		t.Fatalf("expected path update report=%+v", out.Report)
	}
	got, err := s.GetGitWorktree(ctx, projectsdomain.LegacyGlobalDefaultProjectID, wt.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.ID != wt.ID {
		t.Fatalf("worktree id changed")
	}
	if worktreePathKey(got.Path) != worktreePathKey(wtPath2) {
		t.Fatalf("path=%q want %q", got.Path, wtPath2)
	}
}

func TestReconcileGitRepository_skipsUnregisteredLiveWorktrees(t *testing.T) {
	s, ctx, gitSvc := gitTestStore(t)
	main := initGitRepo(t)
	repo, err := s.CreateGitRepository(ctx, projectsdomain.LegacyGlobalDefaultProjectID, CreateGitRepositoryInput{Path: main}, gitSvc)
	if err != nil {
		t.Fatal(err)
	}
	extraPath := filepath.Join(filepath.Dir(main), "wt-unregistered")
	runGitStore(t, main, "worktree", "add", extraPath, "-b", "orphan-branch")
	t.Cleanup(func() { _ = os.RemoveAll(extraPath) })

	out, err := s.ReconcileGitRepository(ctx, projectsdomain.LegacyGlobalDefaultProjectID, repo.ID, ReconcileGitInput{
		RepairGit:   true,
		AllowRemove: true,
	}, gitSvc)
	if err != nil {
		t.Fatalf("ReconcileGitRepository: %v", err)
	}
	if out.Report.WorktreesAdded != 0 {
		t.Fatalf("worktrees_added=%d want 0", out.Report.WorktreesAdded)
	}
	wts, err := s.ListGitWorktrees(ctx, projectsdomain.LegacyGlobalDefaultProjectID, repo.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(wts) != 1 {
		t.Fatalf("registered worktrees=%d want 1 (main only)", len(wts))
	}
}

func TestReconcileGitRepository_bootstrapWrongRepo(t *testing.T) {
	s, ctx, gitSvc := gitTestStore(t)
	mainA := initGitRepo(t)
	runGitStore(t, mainA, "commit", "--allow-empty", "-m", "marker-a")
	repoA, err := s.CreateGitRepository(ctx, projectsdomain.LegacyGlobalDefaultProjectID, CreateGitRepositoryInput{Path: mainA}, gitSvc)
	if err != nil {
		t.Fatal(err)
	}
	branches, err := s.ListGitBranches(ctx, projectsdomain.LegacyGlobalDefaultProjectID, repoA.ID)
	if err != nil || len(branches) == 0 || strings.TrimSpace(branches[0].HeadSHA) == "" {
		t.Fatalf("branches for verify: %v len=%d", err, len(branches))
	}
	mainB := initGitRepo(t)
	renamed := filepath.Join(filepath.Dir(mainA), "gone-a")
	if err := os.Rename(mainA, renamed); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Rename(renamed, mainA) })

	_, err = s.ReconcileGitRepository(ctx, projectsdomain.LegacyGlobalDefaultProjectID, repoA.ID, ReconcileGitInput{
		BootstrapPath: mainB,
		AllowRemove:   true,
	}, gitSvc)
	if domain.GitErrCode(err) != domain.GitCodeBootstrapMismatch {
		t.Fatalf("got %v want bootstrap_mismatch", err)
	}
}

func TestStore_CreateGitRepository_setsGitCommonDirAndSingleBranch(t *testing.T) {
	s, ctx, gitSvc := gitTestStore(t)
	main := initGitRepo(t)
	runGitStore(t, main, "branch", "extra")
	repo, err := s.CreateGitRepository(ctx, projectsdomain.LegacyGlobalDefaultProjectID, CreateGitRepositoryInput{Path: main}, gitSvc)
	if err != nil {
		t.Fatalf("CreateGitRepository: %v", err)
	}
	if repo.GitCommonDir == "" {
		t.Fatal("GitCommonDir empty")
	}
	branches, err := s.ListGitBranches(ctx, projectsdomain.LegacyGlobalDefaultProjectID, repo.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(branches) != 1 {
		t.Fatalf("len(branches)=%d want 1 bound branch only", len(branches))
	}
	if branches[0].Name != "main" {
		t.Fatalf("branch name=%q want main", branches[0].Name)
	}
}

func TestReconcileGitRepository_pathMatch_reportsCheckoutMismatch(t *testing.T) {
	s, ctx, gitSvc := gitTestStore(t)
	main := initGitRepo(t)
	repo, err := s.CreateGitRepository(ctx, projectsdomain.LegacyGlobalDefaultProjectID, CreateGitRepositoryInput{Path: main}, gitSvc)
	if err != nil {
		t.Fatal(err)
	}
	wtPath := filepath.Join(filepath.Dir(main), "wt-checkout")
	wt, err := s.CreateGitWorktree(ctx, projectsdomain.LegacyGlobalDefaultProjectID, repo.ID, CreateGitWorktreeInput{
		Path:         wtPath,
		Branch:       "feature-bound",
		CreateBranch: true,
	}, gitSvc)
	if err != nil {
		t.Fatalf("CreateGitWorktree: %v", err)
	}
	runGitStore(t, wtPath, "checkout", "-b", "other-branch")

	out, err := s.ReconcileGitRepository(ctx, projectsdomain.LegacyGlobalDefaultProjectID, repo.ID, ReconcileGitInput{}, gitSvc)
	if err != nil {
		t.Fatalf("ReconcileGitRepository: %v", err)
	}
	if out.Status != reconcileStatusPartial {
		t.Fatalf("status=%q want partial", out.Status)
	}
	found := false
	for _, skip := range out.Report.WorktreesSkipped {
		if skip.WorktreeID == wt.ID && skip.Reason == "branch_checkout_mismatch" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected branch_checkout_mismatch skip report=%+v", out.Report)
	}
}

func TestReconcileGitRepository_dryRun_noWrites(t *testing.T) {
	s, ctx, gitSvc := gitTestStore(t)
	main := initGitRepo(t)
	repo, err := s.CreateGitRepository(ctx, projectsdomain.LegacyGlobalDefaultProjectID, CreateGitRepositoryInput{Path: main}, gitSvc)
	if err != nil {
		t.Fatal(err)
	}
	before, err := s.GetGitRepository(ctx, projectsdomain.LegacyGlobalDefaultProjectID, repo.ID)
	if err != nil {
		t.Fatal(err)
	}
	out, err := s.ReconcileGitRepository(ctx, projectsdomain.LegacyGlobalDefaultProjectID, repo.ID, ReconcileGitInput{
		DryRun:      true,
		AllowRemove: true,
	}, gitSvc)
	if err != nil {
		t.Fatalf("ReconcileGitRepository: %v", err)
	}
	if out.Status != reconcileStatusOK {
		t.Fatalf("status=%q", out.Status)
	}
	after, err := s.GetGitRepository(ctx, projectsdomain.LegacyGlobalDefaultProjectID, repo.ID)
	if err != nil {
		t.Fatal(err)
	}
	if before.Path != after.Path || before.UpdatedAt != after.UpdatedAt {
		t.Fatal("dry run modified repository row")
	}
}

func TestReconcileGitRepository_removesIncompleteDiscoverStubAtMainPath(t *testing.T) {
	s, ctx, gitSvc := gitTestStore(t)
	main := initGitRepo(t)
	repo, err := s.CreateGlobalGitRepository(ctx, CreateGitRepositoryInput{Path: main}, gitSvc)
	if err != nil {
		t.Fatal(err)
	}
	rows, err := s.ListGitWorktreesByRepo(ctx, repo.ID)
	if err != nil {
		t.Fatal(err)
	}
	var mainRow domain.GitWorktree
	for _, row := range rows {
		if row.IsMain {
			mainRow = row
			break
		}
	}
	if mainRow.ID == "" {
		t.Fatal("expected seeded main worktree row")
	}
	if err := s.db.WithContext(ctx).Model(&model.GitWorktree{}).Where("id = ?", mainRow.ID).Updates(map[string]any{
		"branch_id": "",
		"name":      "discovered-" + filepath.Base(main),
		"is_main":   true,
	}).Error; err != nil {
		t.Fatalf("simulate discover stub: %v", err)
	}

	out, err := s.ReconcileGitRepository(ctx, "", repo.ID, ReconcileGitInput{
		AllowRemove: true,
		RepairGit:   true,
	}, gitSvc)
	if err != nil {
		t.Fatalf("ReconcileGitRepository: %v", err)
	}
	if out.Report.WorktreesRemoved != 1 {
		t.Fatalf("worktrees_removed=%d want 1", out.Report.WorktreesRemoved)
	}
	after, err := s.ListGitWorktreesByRepo(ctx, repo.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(after) != 0 {
		t.Fatalf("worktrees after reconcile=%+v want empty", after)
	}
}

func TestReconcileGitRepository_preservesRegisteredMainWorktree(t *testing.T) {
	s, ctx, gitSvc := gitTestStore(t)
	main := initGitRepo(t)
	repo, err := s.CreateGlobalGitRepository(ctx, CreateGitRepositoryInput{Path: main}, gitSvc)
	if err != nil {
		t.Fatal(err)
	}
	out, err := s.ReconcileGitRepository(ctx, "", repo.ID, ReconcileGitInput{
		AllowRemove: true,
		RepairGit:   true,
	}, gitSvc)
	if err != nil {
		t.Fatalf("ReconcileGitRepository: %v", err)
	}
	if out.Report.WorktreesRemoved != 0 {
		t.Fatalf("worktrees_removed=%d want 0", out.Report.WorktreesRemoved)
	}
	rows, err := s.ListGitWorktreesByRepo(ctx, repo.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 || !rows[0].IsMain || rows[0].BranchID == "" {
		t.Fatalf("registered main worktree=%+v", rows)
	}
}

func TestReconcileGitRepository_doesNotDiscoverAtMainCheckoutPath(t *testing.T) {
	s, ctx, gitSvc := gitTestStore(t)
	main := initGitRepo(t)
	repo, err := s.CreateGlobalGitRepository(ctx, CreateGitRepositoryInput{Path: main}, gitSvc)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.ReconcileGitRepository(ctx, "", repo.ID, ReconcileGitInput{
		AllowDiscover: true,
		AllowRemove:   true,
		RepairGit:     true,
	}, gitSvc); err != nil {
		t.Fatalf("ReconcileGitRepository: %v", err)
	}
	rows, err := s.ListGitWorktreesByRepo(ctx, repo.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 {
		t.Fatalf("worktrees=%d want 1 (main only, no discover stub)", len(rows))
	}
	if !rows[0].IsMain || rows[0].BranchID == "" {
		t.Fatalf("main row=%+v", rows[0])
	}
	for _, row := range rows {
		if strings.HasPrefix(row.Name, "discovered-") {
			t.Fatalf("unexpected discover stub at main path: %+v", row)
		}
	}
}

func TestRelocateGitRepository_globalUpdatesRepoPath(t *testing.T) {
	s, ctx, gitSvc := gitTestStore(t)
	main := initGitRepo(t)
	repo, err := s.CreateGlobalGitRepository(ctx, CreateGitRepositoryInput{Path: main}, gitSvc)
	if err != nil {
		t.Fatal(err)
	}
	renamed := filepath.Join(filepath.Dir(main), "relocated-main")
	if err := os.Rename(main, renamed); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Rename(renamed, main) })

	out, err := s.RelocateGitRepository(ctx, "", repo.ID, renamed, gitSvc)
	if err != nil {
		t.Fatalf("RelocateGitRepository: %v", err)
	}
	if out.Status != reconcileStatusOK {
		t.Fatalf("status=%q want ok", out.Status)
	}
	if !out.Report.RepoPathUpdated {
		t.Fatal("expected repo path update")
	}
	got, err := s.GetGitRepositoryByID(ctx, repo.ID)
	if err != nil {
		t.Fatal(err)
	}
	if worktreePathKey(got.Path) != worktreePathKey(renamed) {
		t.Fatalf("repo path=%q want %q", got.Path, renamed)
	}
}

func assertNoObsoleteReconcileSkipReasons(t *testing.T, report ReconcileReport) {
	t.Helper()
	for _, skip := range report.WorktreesSkipped {
		switch skip.Reason {
		case "path_and_branch_not_found", "has_task_ref":
			t.Fatalf("obsolete skip reason %q on worktree %q; report=%+v", skip.Reason, skip.WorktreeID, report)
		}
	}
}

func insertTaskOnWorktree(t *testing.T, s *Store, worktreeID string, status taskcoredomain.Status) {
	t.Helper()
	wtID := worktreeID
	task := taskcoremodel.Task{
		ID:            uuid.NewString(),
		Title:         "reconcile-vanish-task",
		InitialPrompt: "test",
		Status:        status,
		Priority:      taskcoredomain.PriorityMedium,
		Runner:        "cursor",
		WorktreeID:    &wtID,
	}
	if err := s.db.Create(&task).Error; err != nil {
		t.Fatalf("insert task: %v", err)
	}
}

func TestReconcileGitRepository_vanishedLinkedWorktree_removesRow(t *testing.T) {
	s, ctx, gitSvc := gitTestStore(t)
	main := initGitRepo(t)
	repo, err := s.CreateGitRepository(ctx, projectsdomain.LegacyGlobalDefaultProjectID, CreateGitRepositoryInput{Path: main}, gitSvc)
	if err != nil {
		t.Fatal(err)
	}
	wtPath := filepath.Join(filepath.Dir(main), "wt-vanished")
	wt, err := s.CreateGitWorktree(ctx, projectsdomain.LegacyGlobalDefaultProjectID, repo.ID, CreateGitWorktreeInput{
		Path:         wtPath,
		Branch:       "feature-vanished",
		CreateBranch: true,
	}, gitSvc)
	if err != nil {
		t.Fatalf("CreateGitWorktree: %v", err)
	}
	if err := os.RemoveAll(wtPath); err != nil {
		t.Fatalf("RemoveAll vanished path: %v", err)
	}

	out, err := s.ReconcileGitRepository(ctx, projectsdomain.LegacyGlobalDefaultProjectID, repo.ID, ReconcileGitInput{
		RepairGit:   true,
		AllowRemove: true,
	}, gitSvc)
	if err != nil {
		t.Fatalf("ReconcileGitRepository: %v", err)
	}
	assertNoObsoleteReconcileSkipReasons(t, out.Report)
	if out.Report.WorktreesRemoved != 1 {
		t.Fatalf("worktrees_removed=%d want 1 report=%+v", out.Report.WorktreesRemoved, out.Report)
	}
	if _, err := s.GetGitWorktree(ctx, projectsdomain.LegacyGlobalDefaultProjectID, wt.ID); err == nil {
		t.Fatal("expected vanished worktree row to be gone")
	}
	rows, err := s.ListGitWorktrees(ctx, projectsdomain.LegacyGlobalDefaultProjectID, repo.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 || !rows[0].IsMain {
		t.Fatalf("worktrees after reconcile=%+v want main only", rows)
	}
}

func TestReconcileGitRepository_ghostPathInLiveList_removesRow(t *testing.T) {
	s, ctx, gitSvc := gitTestStore(t)
	main := initGitRepo(t)
	repo, err := s.CreateGitRepository(ctx, projectsdomain.LegacyGlobalDefaultProjectID, CreateGitRepositoryInput{Path: main}, gitSvc)
	if err != nil {
		t.Fatal(err)
	}
	wtPath := filepath.Join(filepath.Dir(main), "wt-ghost")
	wt, err := s.CreateGitWorktree(ctx, projectsdomain.LegacyGlobalDefaultProjectID, repo.ID, CreateGitWorktreeInput{
		Path:         wtPath,
		Branch:       "feature-ghost",
		CreateBranch: true,
	}, gitSvc)
	if err != nil {
		t.Fatalf("CreateGitWorktree: %v", err)
	}
	if err := os.RemoveAll(wtPath); err != nil {
		t.Fatalf("RemoveAll ghost path: %v", err)
	}

	// No RepairGit/prune: git still lists the missing path (often prunable); FilterLiveWorktrees must drop it.
	out, err := s.ReconcileGitRepository(ctx, projectsdomain.LegacyGlobalDefaultProjectID, repo.ID, ReconcileGitInput{
		AllowRemove: true,
	}, gitSvc)
	if err != nil {
		t.Fatalf("ReconcileGitRepository: %v", err)
	}
	assertNoObsoleteReconcileSkipReasons(t, out.Report)
	if out.Report.WorktreesRemoved != 1 {
		t.Fatalf("worktrees_removed=%d want 1 report=%+v", out.Report.WorktreesRemoved, out.Report)
	}
	if _, err := s.GetGitWorktree(ctx, projectsdomain.LegacyGlobalDefaultProjectID, wt.ID); err == nil {
		t.Fatal("expected ghost worktree row to be gone")
	}
}

func TestReconcileGitRepository_vanishedWithNonRunningTask_stillRemoves(t *testing.T) {
	s, ctx, gitSvc := gitTestStore(t)
	main := initGitRepo(t)
	repo, err := s.CreateGitRepository(ctx, projectsdomain.LegacyGlobalDefaultProjectID, CreateGitRepositoryInput{Path: main}, gitSvc)
	if err != nil {
		t.Fatal(err)
	}
	wtPath := filepath.Join(filepath.Dir(main), "wt-nonrunning-task")
	wt, err := s.CreateGitWorktree(ctx, projectsdomain.LegacyGlobalDefaultProjectID, repo.ID, CreateGitWorktreeInput{
		Path:         wtPath,
		Branch:       "feature-nonrunning",
		CreateBranch: true,
	}, gitSvc)
	if err != nil {
		t.Fatalf("CreateGitWorktree: %v", err)
	}
	insertTaskOnWorktree(t, s, wt.ID, taskcoredomain.StatusReady)
	if err := os.RemoveAll(wtPath); err != nil {
		t.Fatalf("RemoveAll: %v", err)
	}

	out, err := s.ReconcileGitRepository(ctx, projectsdomain.LegacyGlobalDefaultProjectID, repo.ID, ReconcileGitInput{
		RepairGit:   true,
		AllowRemove: true,
	}, gitSvc)
	if err != nil {
		t.Fatalf("ReconcileGitRepository: %v", err)
	}
	assertNoObsoleteReconcileSkipReasons(t, out.Report)
	if out.Report.WorktreesRemoved != 1 {
		t.Fatalf("worktrees_removed=%d want 1 report=%+v", out.Report.WorktreesRemoved, out.Report)
	}
	for _, skip := range out.Report.WorktreesSkipped {
		if skip.WorktreeID == wt.ID {
			t.Fatalf("unexpected skip for non-running task worktree: %+v", skip)
		}
	}
	if _, err := s.GetGitWorktree(ctx, projectsdomain.LegacyGlobalDefaultProjectID, wt.ID); err == nil {
		t.Fatal("expected worktree with non-running task to be removed")
	}
}

func TestReconcileGitRepository_vanishedWithRunningTask_skipsHasRunningTask(t *testing.T) {
	s, ctx, gitSvc := gitTestStore(t)
	main := initGitRepo(t)
	repo, err := s.CreateGitRepository(ctx, projectsdomain.LegacyGlobalDefaultProjectID, CreateGitRepositoryInput{Path: main}, gitSvc)
	if err != nil {
		t.Fatal(err)
	}
	wtPath := filepath.Join(filepath.Dir(main), "wt-running-task")
	wt, err := s.CreateGitWorktree(ctx, projectsdomain.LegacyGlobalDefaultProjectID, repo.ID, CreateGitWorktreeInput{
		Path:         wtPath,
		Branch:       "feature-running",
		CreateBranch: true,
	}, gitSvc)
	if err != nil {
		t.Fatalf("CreateGitWorktree: %v", err)
	}
	insertTaskOnWorktree(t, s, wt.ID, taskcoredomain.StatusRunning)
	if err := os.RemoveAll(wtPath); err != nil {
		t.Fatalf("RemoveAll: %v", err)
	}

	out, err := s.ReconcileGitRepository(ctx, projectsdomain.LegacyGlobalDefaultProjectID, repo.ID, ReconcileGitInput{
		RepairGit:   true,
		AllowRemove: true,
	}, gitSvc)
	if err != nil {
		t.Fatalf("ReconcileGitRepository: %v", err)
	}
	assertNoObsoleteReconcileSkipReasons(t, out.Report)
	if out.Report.WorktreesRemoved != 0 {
		t.Fatalf("worktrees_removed=%d want 0 report=%+v", out.Report.WorktreesRemoved, out.Report)
	}
	if out.Status != reconcileStatusPartial {
		t.Fatalf("status=%q want %q", out.Status, reconcileStatusPartial)
	}
	found := false
	for _, skip := range out.Report.WorktreesSkipped {
		if skip.WorktreeID == wt.ID && skip.Reason == "has_running_task" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected has_running_task skip report=%+v", out.Report)
	}
	got, err := s.GetGitWorktree(ctx, projectsdomain.LegacyGlobalDefaultProjectID, wt.ID)
	if err != nil {
		t.Fatalf("expected worktree row kept: %v", err)
	}
	if got.ID != wt.ID {
		t.Fatalf("worktree id=%q want %q", got.ID, wt.ID)
	}
}
