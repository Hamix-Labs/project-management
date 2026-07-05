import { useEffect, useMemo, useState } from "react";
import { Link } from "react-router-dom";
import { CustomSelect } from "@/components/custom-select";
import {
  WorktreesBranchIcon,
  WorktreesFolderGitIcon,
} from "@/worktrees/components/WorktreesIcons";
import { useGlobalRepositories } from "@/worktrees/hooks/useGlobalRepositories";
import { useGlobalWorktrees } from "@/worktrees/hooks/useGlobalWorktrees";
import { useGlobalBranches } from "@/worktrees/hooks/useGlobalBranches";
import { isFullyRegisteredWorktree } from "@/worktrees/worktreeRegistration";
import { ProjectSelect } from "@/projects/ProjectSelect";
import { useProjectsByRepository } from "@/projects/useProjectsByRepository";

type Props = {
  idsPrefix: string;
  repositoryId: string;
  projectId: string;
  worktreeId: string;
  onRepositoryChange: (repositoryId: string) => void;
  onProjectChange: (projectId: string) => void;
  onWorktreeChange: (worktreeId: string) => void;
  disabled?: boolean;
};

function defaultProjectId(projects: { id: string; is_default: boolean; status: string }[]) {
  const active = projects.filter((p) => p.status === "active");
  const row = active.find((p) => p.is_default) ?? active[0];
  return row?.id ?? "";
}

export function TaskCreateAssignmentFields({
  idsPrefix,
  repositoryId,
  projectId,
  worktreeId,
  onRepositoryChange,
  onProjectChange,
  onWorktreeChange,
  disabled = false,
}: Props) {
  const repositoriesQuery = useGlobalRepositories();
  const repositories = repositoriesQuery.data ?? [];

  const [selectedRepoId, setSelectedRepoId] = useState(repositoryId ?? "");

  useEffect(() => {
    if (repositoryId !== selectedRepoId) {
      setSelectedRepoId(repositoryId);
    }
  }, [repositoryId, selectedRepoId]);

  useEffect(() => {
    if (repositories.length === 1 && selectedRepoId === "") {
      const id = repositories[0]!.id;
      setSelectedRepoId(id);
      onRepositoryChange(id);
    }
  }, [repositories, selectedRepoId, onRepositoryChange]);

  const projectsQuery = useProjectsByRepository(selectedRepoId, {
    enabled: selectedRepoId !== "",
  });
  const projects = projectsQuery.data?.projects ?? [];

  useEffect(() => {
    if (selectedRepoId === "" || projectsQuery.isLoading) return;
    if (projectId !== "" && projects.some((p) => p.id === projectId)) return;
    const next = defaultProjectId(projects);
    if (next !== "" && next !== projectId) {
      onProjectChange(next);
    }
  }, [selectedRepoId, projects, projectsQuery.isLoading, projectId, onProjectChange]);

  const worktreesQuery = useGlobalWorktrees(selectedRepoId, {
    enabled: selectedRepoId !== "",
  });
  const worktrees = (worktreesQuery.data ?? []).filter(isFullyRegisteredWorktree);

  const branchesQuery = useGlobalBranches(selectedRepoId, {
    enabled: selectedRepoId !== "",
  });
  const branches = branchesQuery.data ?? [];
  const branchById = useMemo(
    () => new Map(branches.map((b) => [b.id, b])),
    [branches],
  );

  useEffect(() => {
    if (worktrees.length === 1 && worktreeId === "") {
      onWorktreeChange(worktrees[0]!.id);
    }
  }, [worktrees, worktreeId, onWorktreeChange]);

  const loading =
    repositoriesQuery.isLoading ||
    projectsQuery.isLoading ||
    worktreesQuery.isLoading ||
    branchesQuery.isLoading;

  const noRepositories = !repositoriesQuery.isLoading && repositories.length === 0;

  const repoOptions = repositories.map((r) => ({
    value: r.id,
    label: r.path,
  }));

  const worktreeOptions = worktrees.map((wt) => {
    const branch = wt.branch_id ? branchById.get(wt.branch_id) : undefined;
    const name = wt.name.trim() || wt.path;
    const label = branch ? `${name} (${branch.name})` : name;
    return { value: wt.id, label };
  });

  if (noRepositories) {
    return (
      <p className="worktrees-git-selector__manage">
        No repositories registered.{" "}
        <Link to="/worktrees?register=1" target="_blank" rel="noopener noreferrer">
          Register repository
        </Link>
      </p>
    );
  }

  return (
    <div
      className="worktrees-git-selector worktrees-git-selector--modal"
      aria-busy={loading ? "true" : undefined}
    >
      <CustomSelect
        id={`${idsPrefix}-repo`}
        label="Repository"
        value={selectedRepoId}
        options={repoOptions}
        disabled={disabled || repositoriesQuery.isLoading || repoOptions.length === 0}
        requirement="required"
        leadingIcon={
          <WorktreesFolderGitIcon className="worktrees-git-selector__icon" />
        }
        onChange={(id) => {
          setSelectedRepoId(id);
          onRepositoryChange(id);
          onProjectChange("");
          onWorktreeChange("");
        }}
      />

      <ProjectSelect
        id={`${idsPrefix}-project`}
        value={projectId}
        projects={projects}
        loading={projectsQuery.isLoading}
        disabled={disabled || selectedRepoId === ""}
        onChange={onProjectChange}
      />

      <CustomSelect
        id={`${idsPrefix}-worktree`}
        label="Worktree"
        value={worktreeId}
        options={worktreeOptions}
        disabled={
          disabled ||
          selectedRepoId === "" ||
          worktreesQuery.isLoading ||
          worktreeOptions.length === 0
        }
        requirement="required"
        leadingIcon={
          <WorktreesBranchIcon className="worktrees-git-selector__icon" />
        }
        onChange={onWorktreeChange}
      />

      <p className="worktrees-git-selector__manage">
        <Link to="/worktrees" target="_blank" rel="noopener noreferrer">
          Manage worktrees
        </Link>
      </p>
    </div>
  );
}
