import { useEffect, useMemo, useRef } from "react";
import { useGlobalRepositories } from "@/hooks/useGlobalRepositories";
import { useGlobalWorktrees } from "@/hooks/useGlobalWorktrees";
import { useGlobalBranches } from "@/hooks/useGlobalBranches";
import { isFullyRegisteredWorktree } from "@/lib/gitWorktreeRegistration";
import { repositoryDisplayName } from "@/lib/repositoryDisplayName";
import { useProjectsByRepository } from "@/hooks/useProjectsByRepository";
import {
  assignmentEquals,
  decideComposeGitAssignment,
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
};

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

  // User project picks beat late-arriving defaults (F-06-07).
  const projectChosenByUserRef = useRef(false);
  const prevRepositoryIdRef = useRef(assignment.repositoryId);

  if (prevRepositoryIdRef.current !== assignment.repositoryId) {
    prevRepositoryIdRef.current = assignment.repositoryId;
    projectChosenByUserRef.current = false;
  }

  useEffect(() => {
    const next = decideComposeGitAssignment(
      assignment,
      {
        repositories,
        repositoriesLoading: repositoriesQuery.isLoading,
        projects,
        projectsLoading: projectsQuery.isLoading,
        worktreesLoading: worktreesQuery.isLoading,
      },
      { projectChosenByUser: projectChosenByUserRef.current },
    );
    if (!assignmentEquals(next, assignment)) {
      input.onAssignmentChange(next);
    }
  }, [
    assignment,
    repositories,
    repositoriesQuery.isLoading,
    projects,
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
    label: repositoryDisplayName(r.path) || r.path,
    title: r.path,
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
      projectChosenByUserRef.current = false;
      input.onAssignmentChange(selectRepositoryState(assignment, repositoryId));
    },
    selectProject: (projectId: string) => {
      projectChosenByUserRef.current = true;
      input.onAssignmentChange(selectProjectState(assignment, projectId));
    },
    selectWorktree: (worktreeId: string) => {
      input.onAssignmentChange(selectWorktreeState(assignment, worktreeId));
    },
  };
}
