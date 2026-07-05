package service

import (
	"context"

	"golang.org/x/sync/errgroup"

	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/domain"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/store"
)

// BootstrapLimits caps each sub-read in the cold-start aggregate.
type BootstrapLimits struct {
	TasksLimit    int
	ProjectsLimit int
	DraftsLimit   int
}

// BootstrapData is the HTTP-agnostic cold-start bundle for GET /v1/bootstrap.
type BootstrapData struct {
	Settings store.AppSettings
	Tasks    []domain.Task
	HasMore  bool
	Stats    store.TaskStats
	Projects []domain.Project
	Drafts   []store.DraftSummary
}

// BootstrapStore is the persistence surface required for cold-start aggregation.
type BootstrapStore interface {
	GetSettings(ctx context.Context) (store.AppSettings, error)
	ListFlatPage(ctx context.Context, limit, offset int, filter *store.ListFilter) ([]domain.Task, bool, error)
	TaskStats(ctx context.Context) (store.TaskStats, error)
	ListProjects(ctx context.Context, includeArchived bool, limit int) ([]domain.Project, error)
	ListDrafts(ctx context.Context, limit int) ([]store.DraftSummary, error)
}

// Bootstrap loads settings, tasks page, stats, projects, and drafts in parallel.
// Any sub-read failure aborts the whole operation so callers can fall back to
// per-endpoint fan-out.
//
//funclogmeasure:skip category=delegate-already-logs reason="Bootstrap orchestration; handler bootstrap and store reads emit operation traces."
func Bootstrap(ctx context.Context, st BootstrapStore, limits BootstrapLimits) (BootstrapData, error) {
	var out BootstrapData
	g, gctx := errgroup.WithContext(ctx)
	g.Go(func() error {
		v, err := st.GetSettings(gctx)
		if err == nil {
			out.Settings = v
		}
		return err
	})
	g.Go(func() error {
		rows, more, err := st.ListFlatPage(gctx, limits.TasksLimit, 0, nil)
		if err == nil {
			out.Tasks = rows
			out.HasMore = more
		}
		return err
	})
	g.Go(func() error {
		v, err := st.TaskStats(gctx)
		if err == nil {
			out.Stats = v
		}
		return err
	})
	g.Go(func() error {
		v, err := st.ListProjects(gctx, false, limits.ProjectsLimit)
		if err == nil {
			out.Projects = v
		}
		return err
	})
	g.Go(func() error {
		v, err := st.ListDrafts(gctx, limits.DraftsLimit)
		if err == nil {
			out.Drafts = v
		}
		return err
	})
	if err := g.Wait(); err != nil {
		return BootstrapData{}, err
	}
	return out, nil
}
