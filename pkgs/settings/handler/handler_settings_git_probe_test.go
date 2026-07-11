package handler

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"os/exec"
	"testing"

	"github.com/AlexsanderHamir/Hamix/internal/tasktestdb"
	"github.com/AlexsanderHamir/Hamix/pkgs/gitwork"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/store"
)

func initGitRepo(t *testing.T, dir string) {
	t.Helper()
	if out, err := exec.Command("git", "init", "-b", "main", dir).CombinedOutput(); err != nil {
		t.Fatalf("git init: %v %s", err, out)
	}
	runGit(t, dir, "config", "user.email", "test@example.com")
	runGit(t, dir, "config", "user.name", "Test User")
	if out, err := exec.Command("git", "-C", dir, "commit", "--allow-empty", "-m", "init").CombinedOutput(); err != nil {
		t.Fatalf("git commit: %v %s", err, out)
	}
}

func runGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v %s", args, err, out)
	}
}

func TestHTTP_gitRepositoryProbe_notARepository(t *testing.T) {
	dir := t.TempDir()

	db := tasktestdb.OpenSQLite(t)
	st := store.NewStore(db)
	srv := newSettingsHTTPServer(t, st, Deps{Settings: st, GitRead: st, Git: stubGitService{}})

	res, err := http.Get(srv.URL + "/settings/git-probe?path=" + url.QueryEscape(dir))
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(res.Body)
		t.Fatalf("status %d body=%s", res.StatusCode, b)
	}
	var body gitRepositoryProbeResponse
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if body.IsGitRepository {
		t.Fatalf("expected not a repository, got %+v", body)
	}
	if len(body.Branches) != 0 {
		t.Fatalf("branches=%+v", body.Branches)
	}
}

func TestHTTP_gitRepositoryProbe_listsBranches(t *testing.T) {
	dir := t.TempDir()
	initGitRepo(t, dir)
	runGit(t, dir, "branch", "feature")

	db := tasktestdb.OpenSQLite(t)
	st := store.NewStore(db)
	srv := newSettingsHTTPServer(t, st, Deps{Settings: st, GitRead: st, Git: gitwork.New()})

	res, err := http.Get(srv.URL + "/settings/git-probe?path=" + url.QueryEscape(dir))
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(res.Body)
		t.Fatalf("status %d body=%s", res.StatusCode, b)
	}
	var body gitRepositoryProbeResponse
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if !body.IsGitRepository {
		t.Fatalf("expected git repository, got %+v", body)
	}
	if body.CurrentBranch != "main" {
		t.Fatalf("current_branch=%q want main", body.CurrentBranch)
	}
	if len(body.Branches) < 2 {
		t.Fatalf("branches=%+v", body.Branches)
	}
}

type stubGitService struct{}

func (stubGitService) OpenRepository(_ context.Context, _ string) (*gitwork.Repository, error) {
	return nil, gitwork.ErrNotARepository
}

func (stubGitService) ResolveRegistration(context.Context, string) (string, string, error) {
	panic("unexpected call")
}
func (stubGitService) BelongsToRepository(context.Context, string, string) (bool, error) {
	panic("unexpected call")
}
func (stubGitService) OpenRegisteredCheckout(context.Context, gitwork.ResolveInput) (gitwork.ResolveResult, error) {
	panic("unexpected call")
}
func (stubGitService) VerifySameRepository(context.Context, gitwork.RegisteredCheckout, *gitwork.Repository) error {
	panic("unexpected call")
}
func (stubGitService) DiscoverCheckoutNearby(context.Context, gitwork.RegisteredCheckout) (*gitwork.Repository, error) {
	panic("unexpected call")
}
func (stubGitService) ListWorktrees(context.Context, *gitwork.Repository) ([]gitwork.Worktree, error) {
	panic("unexpected call")
}
func (stubGitService) AddWorktree(context.Context, *gitwork.Repository, string, gitwork.AddWorktreeOptions) (*gitwork.Worktree, error) {
	panic("unexpected call")
}
func (stubGitService) RemoveWorktree(context.Context, *gitwork.Repository, string, bool) error {
	panic("unexpected call")
}
func (stubGitService) RepairWorktrees(context.Context, *gitwork.Repository) error {
	panic("unexpected call")
}
func (stubGitService) PruneWorktrees(context.Context, *gitwork.Repository) error {
	panic("unexpected call")
}
func (stubGitService) BranchHead(context.Context, *gitwork.Repository, string) (string, error) {
	panic("unexpected call")
}
func (stubGitService) ListBranches(context.Context, *gitwork.Repository) ([]gitwork.Branch, error) {
	panic("unexpected call")
}
func (stubGitService) CreateBranch(context.Context, *gitwork.Repository, string, string) (*gitwork.Branch, error) {
	panic("unexpected call")
}
func (stubGitService) DeleteBranch(context.Context, *gitwork.Repository, string, bool) error {
	panic("unexpected call")
}
func (stubGitService) WorktreeCurrentBranch(context.Context, string) (string, error) {
	panic("unexpected call")
}
func (stubGitService) Checkout(context.Context, string, string) error {
	panic("unexpected call")
}
func (stubGitService) CheckoutStatus(context.Context, string) (gitwork.CheckoutStatus, error) {
	panic("unexpected call")
}
