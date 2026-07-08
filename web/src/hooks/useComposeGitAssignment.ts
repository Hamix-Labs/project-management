import { useEffect, useMemo, useRef } from "react";
import { useGlobalRepositories } from "@/hooks/useGlobalRepositories";
import { useGlobalWorktrees } from "@/hooks/useGlobalWorktrees";
import { useGlobalBranches } from "@/hooks/useGlobalBranches";
import { isFullyRegisteredWorktree } from "@/lib/gitWorktreeRegistration";
import { useProjectsByRepository } from "@/hooks/useProjectsByRepository";
import {
  applyRepoScopedDefaults,
  assignmentEquals,
  initFreshAssignment,
  isFreshAssignment,
  selectProject as selectProjectState,
  selectRepository as selectRepositoryState,
  selectWorktree as selectWorktreeState,
  type ComposeGitAssignment,
} from "@/lib/composeGitAssignment";

type Input = {
  repositoryId: string;
  projectId: string;
  worktreeId: string;
  onAssignmentChange: (next: ComposeGitAssignment) => void;
  onProjectContextClear: () => void;
};

function needsFreshRepoDefaults(assignment: ComposeGitAssignment): boolean {
  return (
    assignment.repositoryId !== "" &&
    assignment.projectId === "" &&
    assignment.worktreeId === ""
  );
}

export function useComposeGitAssignment(input: Input) {
  const assignment: ComposeGitAssignment = useMemo(
    () => ({
      repositoryId: input.repositoryId,
      projectId: input.projectId,
      worktreeId: input.worktreeId,
    }),
    [input.repositoryId, input.projectId, input.worktreeId],
  );

  const repositoriesQuery = useGlobalRepositories();
  const repositories = useMemo(
    () => repositoriesQuery.data ?? [],
    [repositoriesQuery.data],
  );

  const projectsQuery = useProjectsByRepository(assignment.repositoryId, {
    enabled: assignment.repositoryId !== "",
  });
  const projects = useMemo(
    () => projectsQuery.data?.projects ?? [],
    [projectsQuery.data?.projects],
  );

  const worktreesQuery = useGlobalWorktrees(assignment.repositoryId, {
    enabled: assignment.repositoryId !== "",
  });
  const worktrees = (worktreesQuery.data ?? []).filter(isFullyRegisteredWorktree);

  const branchesQuery = useGlobalBranches(assignment.repositoryId, {
    enabled: assignment.repositoryId !== "",
  });
  const branches = useMemo(
    () => branchesQuery.data ?? [],
    [branchesQuery.data],
  );
  const branchById = useMemo(
    () => new Map(branches.map((b) => [b.id, b])),
    [branches],
  );

  const freshDefaultsDoneRef = useRef(false);

  useEffect(() => {
    if (repositoriesQuery.isLoading || !isFreshAssignment(assignment)) {
      return;
    }
    const next = initFreshAssignment(assignment, repositories);
    if (!assignmentEquals(next, assignment)) {
      input.onAssignmentChange(next);
    }
  }, [assignment, repositories, repositoriesQuery.isLoading, input]);

  useEffect(() => {
    if (!needsFreshRepoDefaults(assignment)) {
      freshDefaultsDoneRef.current = false;
      return;
    }
    if (projectsQuery.isLoading || worktreesQuery.isLoading) {
      return;
    }
    if (freshDefaultsDoneRef.current) {
      return;
    }
    const next = applyRepoScopedDefaults(assignment, projects, worktrees);
    freshDefaultsDoneRef.current = true;
    if (!assignmentEquals(next, assignment)) {
      input.onAssignmentChange(next);
    }
  }, [
    assignment,
    projects,
    worktrees,
    projectsQuery.isLoading,
    worktreesQuery.isLoading,
    input,
  ]);

  useEffect(() => {
    if (assignment.repositoryId === "" || needsFreshRepoDefaults(assignment)) {
      return;
    }
    if (projectsQuery.isLoading || worktreesQuery.isLoading) {
      return;
    }
    const next = applyRepoScopedDefaults(assignment, projects, worktrees);
    if (!assignmentEquals(next, assignment)) {
      input.onAssignmentChange(next);
    }
  }, [
    assignment,
    projects,
    worktrees,
    projectsQuery.isLoading,
    worktreesQuery.isLoading,
    input,
  ]);

  const loading =
    repositoriesQuery.isLoading ||
    projectsQuery.isLoading ||
    worktreesQuery.isLoading ||
    branchesQuery.isLoading;

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

  return {
    loading,
    noRepositories: !repositoriesQuery.isLoading && repositories.length === 0,
    repoOptions,
    projects,
    worktreeOptions,
    selectRepository: (repositoryId: string) => {
      input.onProjectContextClear();
      input.onAssignmentChange(selectRepositoryState(assignment, repositoryId));
    },
    selectProject: (projectId: string) => {
      input.onProjectContextClear();
      input.onAssignmentChange(selectProjectState(assignment, projectId));
    },
    selectWorktree: (worktreeId: string) => {
      input.onAssignmentChange(selectWorktreeState(assignment, worktreeId));
    },
  };
}
