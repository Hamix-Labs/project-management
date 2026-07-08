import type { GitWorktree } from "@/types/git";
import { pickDefaultWorktreeId } from "@/lib/gitWorktreeRegistration";

export type ComposeGitAssignment = {
  repositoryId: string;
  projectId: string;
  worktreeId: string;
};

export type ComposeGitProjectOption = {
  id: string;
  is_default: boolean;
  status: string;
};

export const emptyComposeGitAssignment = (): ComposeGitAssignment => ({
  repositoryId: "",
  projectId: "",
  worktreeId: "",
});

function defaultProjectId(projects: ComposeGitProjectOption[]): string {
  const active = projects.filter((p) => p.status === "active");
  const row = active.find((p) => p.is_default) ?? active[0];
  return row?.id ?? "";
}

export function initFreshAssignment(
  current: ComposeGitAssignment,
  repositories: { id: string }[],
): ComposeGitAssignment {
  if (repositories.length !== 1 || current.repositoryId !== "") {
    return current;
  }
  return { ...current, repositoryId: repositories[0]!.id };
}

export function hydrateAssignmentFromPayload(payload: {
  repositoryId?: string;
  projectId?: string;
  worktreeId?: string;
}): ComposeGitAssignment {
  return {
    repositoryId: payload.repositoryId?.trim() ?? "",
    projectId: payload.projectId?.trim() ?? "",
    worktreeId: payload.worktreeId?.trim() ?? "",
  };
}

export function selectRepository(_current: ComposeGitAssignment, repositoryId: string): ComposeGitAssignment {
  return {
    repositoryId,
    projectId: "",
    worktreeId: "",
  };
}

export function selectProject(current: ComposeGitAssignment, projectId: string): ComposeGitAssignment {
  return { ...current, projectId };
}

export function selectWorktree(current: ComposeGitAssignment, worktreeId: string): ComposeGitAssignment {
  return { ...current, worktreeId };
}

export function applyRepoScopedDefaults(
  current: ComposeGitAssignment,
  projects: ComposeGitProjectOption[],
  worktrees: GitWorktree[],
): ComposeGitAssignment {
  let next = current;
  const projectValid =
    next.projectId !== "" && projects.some((p) => p.id === next.projectId);
  if (!projectValid) {
    const projectId = defaultProjectId(projects);
    if (projectId !== "" && projectId !== next.projectId) {
      next = { ...next, projectId };
    }
  }
  const worktreeValid =
    next.worktreeId !== "" && worktrees.some((wt) => wt.id === next.worktreeId);
  if (!worktreeValid) {
    const worktreeId = pickDefaultWorktreeId(worktrees);
    if (worktreeId !== "" && worktreeId !== next.worktreeId) {
      next = { ...next, worktreeId };
    }
  }
  return next;
}

export function assignmentEquals(a: ComposeGitAssignment, b: ComposeGitAssignment): boolean {
  return (
    a.repositoryId === b.repositoryId &&
    a.projectId === b.projectId &&
    a.worktreeId === b.worktreeId
  );
}

export function isFreshAssignment(assignment: ComposeGitAssignment): boolean {
  return (
    assignment.repositoryId === "" &&
    assignment.projectId === "" &&
    assignment.worktreeId === ""
  );
}

export function applyAssignmentPatch(
  current: ComposeGitAssignment,
  patch: Partial<ComposeGitAssignment>,
): ComposeGitAssignment {
  return {
    repositoryId: patch.repositoryId ?? current.repositoryId,
    projectId: patch.projectId ?? current.projectId,
    worktreeId: patch.worktreeId ?? current.worktreeId,
  };
}
