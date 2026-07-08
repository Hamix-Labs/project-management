import { Link } from "react-router-dom";
import { useCallback } from "react";
import { CustomSelect } from "@/components/custom-select";
import {
  WorktreesBranchIcon,
  WorktreesFolderGitIcon,
} from "@/worktrees/components/WorktreesIcons";
import { ProjectSelect } from "@/projects/ProjectSelect";
import { useComposeGitAssignment } from "@/tasks/create/hooks/useComposeGitAssignment";
import type { ComposeGitAssignment } from "@/tasks/create/composeGitAssignment";

type Props = {
  idsPrefix: string;
  repositoryId: string;
  projectId: string;
  worktreeId: string;
  onAssignmentChange: (next: ComposeGitAssignment) => void;
  onProjectContextClear: () => void;
  disabled?: boolean;
};

export function TaskCreateAssignmentFields({
  idsPrefix,
  repositoryId,
  projectId,
  worktreeId,
  onAssignmentChange,
  onProjectContextClear,
  disabled = false,
}: Props) {
  const onAssignmentChangeStable = useCallback(onAssignmentChange, [onAssignmentChange]);
  const onProjectContextClearStable = useCallback(onProjectContextClear, [onProjectContextClear]);

  const git = useComposeGitAssignment({
    repositoryId,
    projectId,
    worktreeId,
    onAssignmentChange: onAssignmentChangeStable,
    onProjectContextClear: onProjectContextClearStable,
  });

  if (git.noRepositories) {
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
      aria-busy={git.loading ? "true" : undefined}
    >
      <CustomSelect
        id={`${idsPrefix}-repo`}
        label="Repository"
        value={repositoryId}
        options={git.repoOptions}
        disabled={disabled || git.loading || git.repoOptions.length === 0}
        requirement="required"
        leadingIcon={
          <WorktreesFolderGitIcon className="worktrees-git-selector__icon" />
        }
        onChange={git.selectRepository}
      />

      <ProjectSelect
        id={`${idsPrefix}-project`}
        value={projectId}
        projects={git.projects}
        loading={git.loading}
        disabled={disabled || repositoryId === ""}
        onChange={git.selectProject}
      />

      <CustomSelect
        id={`${idsPrefix}-worktree`}
        label="Worktree"
        value={worktreeId}
        options={git.worktreeOptions}
        disabled={
          disabled ||
          repositoryId === "" ||
          git.loading ||
          git.worktreeOptions.length === 0
        }
        requirement="required"
        leadingIcon={
          <WorktreesBranchIcon className="worktrees-git-selector__icon" />
        }
        onChange={git.selectWorktree}
      />

      <p className="worktrees-git-selector__manage">
        <Link to="/worktrees" target="_blank" rel="noopener noreferrer">
          Manage worktrees
        </Link>
      </p>
    </div>
  );
}
