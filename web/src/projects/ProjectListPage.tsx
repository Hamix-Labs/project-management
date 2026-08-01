import { useState } from "react";
import { Button } from "@/components/ui";
import { EmptyState } from "@/shared/EmptyState";
import { useDocumentTitle } from "@/shared/useDocumentTitle";
import { useProjects } from "./hooks";
import { ProjectCreateDialog } from "./ProjectCreateDialog";
import { useCreateProjectMutation } from "./mutations";
import { ProjectListRow, ProjectListSkeleton } from "./ProjectListRow";

export function ProjectListPage() {
  useDocumentTitle("Projects");
  const { data, isLoading, error } = useProjects({ includeArchived: true });
  const projects = data?.projects ?? [];
  const activeCount = projects.filter((p) => p.status === "active").length;
  const archivedCount = projects.length - activeCount;
  const [createOpen, setCreateOpen] = useState(false);

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
          <Button
            type="button"
            variant="primary"
            className="pl__new-btn"
            onClick={openCreateDialog}
          >
            New project
          </Button>
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
              <ProjectListRow
                key={project.id}
                project={project}
                index={i}
              />
            ))}
          </div>
        ) : null}
      </div>
    </section>
  );
}
