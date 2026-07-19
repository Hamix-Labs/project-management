import { useState } from "react";
import { Link } from "react-router-dom";
import { useProjectDetailPrefetcher } from "@/app/hooks/usePrefetchOnIntent";
import { useGlobalRepositories } from "@/hooks/useGlobalRepositories";
import { EmptyState } from "@/shared/EmptyState";
import { useDocumentTitle } from "@/shared/useDocumentTitle";
import type { Project } from "@/types";
import { useProjects } from "./hooks";
import { ProjectCreateDialog } from "@/components/projects/ProjectCreateDialog";
import { useCreateProjectMutation } from "./mutations";

/** Last path segment for repo labels (e.g. `C:/proj/hamix` → `hamix`). */
function repositoryBasename(path: string): string {
  const normalized = path.trim().replace(/\\/g, "/").replace(/\/+$/, "");
  if (normalized === "") return path;
  const slash = normalized.lastIndexOf("/");
  return slash >= 0 ? normalized.slice(slash + 1) : normalized;
}

export function ProjectListPage() {
  useDocumentTitle("Projects");
  const { data, isLoading, error } = useProjects({ includeArchived: true });
  const repositoriesQuery = useGlobalRepositories();
  const projects = data?.projects ?? [];
  const activeCount = projects.filter((p) => p.status === "active").length;
  const archivedCount = projects.length - activeCount;
  const [createOpen, setCreateOpen] = useState(false);
  const repoLabelById = new Map(
    (repositoriesQuery.data ?? []).map((repo) => [
      repo.id,
      repositoryBasename(repo.path) || repo.path,
    ]),
  );

  const createMutation = useCreateProjectMutation();

  function openCreateDialog() {
    createMutation.reset();
    setCreateOpen(true);
  }

  function closeCreateDialog() {
    if (createMutation.isPending) return;
    createMutation.reset();
    setCreateOpen(false);
  }

  return (
    <section className="panel task-detail-panel pl">
      <header className="pl__head">
        <div className="pl__head-text">
          <h2 className="task-list-section-title">Projects</h2>
          <p className="pl__subtitle">
            Shared context and memory for tasks in each project.
          </p>
        </div>
        <div className="pl__head-actions">
          <dl className="pl__stats" aria-label="Project summary">
            <div className="pl__stat">
              <dd>{projects.length}</dd>
              <dt>total</dt>
            </div>
            <span className="pl__stat-sep" aria-hidden="true" />
            <div className="pl__stat pl__stat--active">
              <dd>{activeCount}</dd>
              <dt>active</dt>
            </div>
            <span className="pl__stat-sep" aria-hidden="true" />
            <div className="pl__stat">
              <dd>{archivedCount}</dd>
              <dt>archived</dt>
            </div>
          </dl>
          <button
            type="button"
            className="pl__new-btn"
            onClick={openCreateDialog}
          >
            New project
          </button>
        </div>
      </header>

      {createOpen ? (
        <ProjectCreateDialog
          saving={createMutation.isPending}
          error={createMutation.error}
          onCancel={closeCreateDialog}
          onSubmit={(input) =>
            createMutation.mutate(
              {
                name: input.name,
                description: input.description,
                repository_id: input.repository_id,
              },
              { onSuccess: () => setCreateOpen(false) },
            )
          }
        />
      ) : null}

      <div className="pl__list-section">
        {isLoading ? <ProjectListSkeleton /> : null}
        {error ? (
          <div className="pd__inline-error" role="alert">
            {error.message}
          </div>
        ) : null}
        {!isLoading && !error && projects.length === 0 ? (
          <EmptyState
            title="No projects yet"
            description="Create a project to group related tasks."
            density="compact"
            hideIcon
          />
        ) : null}
        {projects.length > 0 ? (
          <div className="pl__list" aria-label="Projects">
            {projects.map((project, i) => (
              <ProjectRow
                key={project.id}
                project={project}
                index={i}
                repositoryLabel={
                  project.repository_id
                    ? repoLabelById.get(project.repository_id)
                    : undefined
                }
              />
            ))}
          </div>
        ) : null}
      </div>
    </section>
  );
}

function ProjectRow({
  project,
  index,
  repositoryLabel,
}: {
  project: Project;
  index: number;
  repositoryLabel?: string;
}) {
  const isArchived = project.status === "archived";
  const displayName =
    project.is_default && repositoryLabel
      ? `${project.name} · ${repositoryLabel}`
      : project.name;
  const openLabel = `Open project ${displayName}`;
  const to = `/projects/${encodeURIComponent(project.id)}`;
  const prefetchProjectDetail = useProjectDetailPrefetcher();
  const onIntent = () => prefetchProjectDetail(project.id);

  return (
    <Link
      to={to}
      className={isArchived ? "pl__row pl__row--archived" : "pl__row pl__row--active"}
      style={{ animationDelay: `${index * 40}ms` }}
      aria-label={openLabel}
      onPointerEnter={onIntent}
      onFocus={onIntent}
    >
      <div className="pl__row-marker" aria-hidden="true" />
      <div className="pl__row-main">
        <span className="pl__row-name">{displayName}</span>
        <span className="pl__row-desc">
          {project.description || project.context_summary || "No description"}
        </span>
      </div>
      <div className="pl__row-meta">
        <span
          className={
            isArchived ? "pd__badge pd__badge--muted" : "pd__badge pd__badge--live"
          }
        >
          <span className="pd__badge-dot" aria-hidden="true" />
          {project.status}
        </span>
        <span className="pl__row-date">{formatDate(project.updated_at)}</span>
      </div>
      <svg className="pl__row-arrow" width="16" height="16" viewBox="0 0 16 16" fill="none" aria-hidden="true">
        <path d="M6 4l4 4-4 4" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round" />
      </svg>
    </Link>
  );
}

function ProjectListSkeleton() {
  return (
    <div className="pl__list" aria-hidden="true">
      {Array.from({ length: 4 }).map((_, i) => (
        <div className="pl__row pl__row--skeleton" key={i}>
          <div className="pl__row-marker" />
          <div className="pl__row-main">
            <span className="pd__shimmer" style={{ width: `${60 - i * 8}%`, height: "0.9rem" }} />
            <span className="pd__shimmer" style={{ width: `${40 + i * 5}%`, height: "0.75rem" }} />
          </div>
          <div className="pl__row-meta">
            <span className="pd__shimmer" style={{ width: "3rem", height: "0.75rem" }} />
          </div>
        </div>
      ))}
    </div>
  );
}

function formatDate(value: string) {
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return "";
  return date.toLocaleDateString(undefined, { month: "short", day: "numeric" });
}
