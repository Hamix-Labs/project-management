import { useMutation, useQueryClient } from "@tanstack/react-query";
import { Link } from "react-router-dom";
import { createProject } from "@/api";
import { ProjectCreateDialog } from "@/components/projects/ProjectCreateDialog";
import { useProjectsByRepository } from "@/hooks/useProjectsByRepository";
import { gitQueryKeys } from "@/lib/gitQueryKeys";
import { projectQueryKeys } from "@/lib/projectQueryKeys";
import { useState } from "react";

type Props = {
  repositoryId: string;
};

export function RepositoryProjectsSection({ repositoryId }: Props) {
  const queryClient = useQueryClient();
  const [createOpen, setCreateOpen] = useState(false);
  const projectsQuery = useProjectsByRepository(repositoryId);
  const projects = projectsQuery.data?.projects ?? [];

  const createMutation = useMutation({
    mutationFn: (input: { name: string; description: string }) =>
      createProject({ ...input, repository_id: repositoryId }),
    onSuccess: async () => {
      setCreateOpen(false);
      await queryClient.invalidateQueries({ queryKey: projectQueryKeys.all });
      await queryClient.invalidateQueries({
        queryKey: gitQueryKeys.projectsByRepo(repositoryId),
      });
    },
  });

  return (
    <section className="repository-projects" aria-labelledby="repository-projects-heading">
      <div className="repository-projects__head">
        <h2 id="repository-projects-heading" className="repository-projects__title">
          Projects
        </h2>
        <button
          type="button"
          className="secondary"
          onClick={() => {
            createMutation.reset();
            setCreateOpen(true);
          }}
        >
          New project
        </button>
      </div>

      {projectsQuery.isLoading ? (
        <p className="repository-projects__loading">Loading projects…</p>
      ) : null}
      {projectsQuery.isError ? (
        <p className="err" role="alert">
          Could not load projects for this repository.
        </p>
      ) : null}

      {!projectsQuery.isLoading && projects.length === 0 ? (
        <p className="repository-projects__empty">
          No projects yet. The system default appears after registration.
        </p>
      ) : null}

      {projects.length > 0 ? (
        <ul className="repository-projects__list">
          {projects.map((project) => (
            <li key={project.id}>
              <Link to={`/projects/${encodeURIComponent(project.id)}`}>
                {project.is_default ? "Default" : project.name}
              </Link>
            </li>
          ))}
        </ul>
      ) : null}

      {createOpen ? (
        <ProjectCreateDialog
          repositoryId={repositoryId}
          saving={createMutation.isPending}
          error={createMutation.error}
          onCancel={() => {
            if (createMutation.isPending) return;
            createMutation.reset();
            setCreateOpen(false);
          }}
          onSubmit={(input) => createMutation.mutate(input)}
        />
      ) : null}
    </section>
  );
}
