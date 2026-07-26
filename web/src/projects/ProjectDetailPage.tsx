import { useState } from "react";
import { Link, useNavigate, useParams } from "react-router-dom";
import { EmptyState } from "@/shared/EmptyState";
import { useDocumentTitle } from "@/shared/useDocumentTitle";
import { useProject } from "./hooks";
import { ProjectDeleteConfirmDialog } from "./ProjectDeleteConfirmDialog";
import { ProjectSettingsPanel } from "./ProjectSettingsPanel";
import { useDeleteProjectMutation } from "./mutations";

export function ProjectDetailPage() {
  const { projectId = "" } = useParams();
  const navigate = useNavigate();
  const project = useProject(projectId);
  const [deleteOpen, setDeleteOpen] = useState(false);
  const title = project.data?.name ? `${project.data.name} project` : "Project";
  useDocumentTitle(title);

  const isDefaultProject = project.data?.is_default === true;

  const deleteProjectMutation = useDeleteProjectMutation(projectId);

  if (!projectId) {
    return (
      <section className="panel task-detail-panel">
        <EmptyState
          title="Missing project id"
          description="Choose a project from the project list."
          density="compact"
          hideIcon
        />
      </section>
    );
  }

  return (
    <section className="panel task-detail-panel pd">
      <header className="pd__header">
        <Link to="/projects" className="pd__back">
          <span aria-hidden="true">&#8249;</span>
          All projects
        </Link>
        {!isDefaultProject && project.data ? (
          <button
            type="button"
            className="pd__header-delete"
            aria-label="Delete project"
            title="Delete project"
            disabled={deleteProjectMutation.isPending}
            onClick={() => setDeleteOpen(true)}
          >
            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" aria-hidden="true">
              <path
                d="M3 6h18M8 6V4a1 1 0 011-1h6a1 1 0 011 1v2m3 0v14a2 2 0 01-2 2H7a2 2 0 01-2-2V6M10 11v6M14 11v6"
                stroke="currentColor"
                strokeWidth="1.75"
                strokeLinecap="round"
                strokeLinejoin="round"
              />
            </svg>
          </button>
        ) : null}
      </header>

      {project.data ? (
        <div className="pd__title-block">
          <div className="pd__title-row">
            <h1 className="pd__title">{project.data.name}</h1>
            <span
              className={
                project.data.status === "archived"
                  ? "pd__badge pd__badge--muted"
                  : "pd__badge pd__badge--live"
              }
            >
              <span className="pd__badge-dot" aria-hidden="true" />
              {project.data.status}
            </span>
          </div>
          {project.data.description ? (
            <p className="pd__subtitle">{project.data.description}</p>
          ) : null}
        </div>
      ) : null}

      {deleteOpen && project.data && !isDefaultProject ? (
        <ProjectDeleteConfirmDialog
          projectName={project.data.name}
          deletePending={deleteProjectMutation.isPending}
          error={deleteProjectMutation.error?.message ?? null}
          onCancel={() => {
            if (!deleteProjectMutation.isPending) {
              deleteProjectMutation.reset();
              setDeleteOpen(false);
            }
          }}
          onConfirm={() =>
            void deleteProjectMutation.mutate(undefined, {
              onSuccess: () => {
                setDeleteOpen(false);
                navigate("/projects");
              },
            })
          }
        />
      ) : null}

      {project.isLoading ? <ProjectDetailSkeleton /> : null}

      {project.error ? (
        <div className="pd__error" role="alert">
          <div className="pd__error-dot" aria-hidden="true" />
          <div>
            <p className="pd__error-title">Unable to load this project</p>
            <p className="pd__error-message">{project.error.message}</p>
          </div>
        </div>
      ) : null}

      {project.data ? (
        <div className="pd__grid">
          <ProjectSettingsPanel project={project.data} />
        </div>
      ) : null}
    </section>
  );
}

function ProjectDetailSkeleton() {
  return (
    <div className="pd__skeleton" aria-hidden="true">
      <div className="pd__shimmer pd__shimmer--card" />
      <div className="pd__shimmer pd__shimmer--card pd__shimmer--card-sm" />
    </div>
  );
}
