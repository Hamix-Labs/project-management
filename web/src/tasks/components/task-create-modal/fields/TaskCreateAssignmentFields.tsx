import { Link } from "react-router-dom";
import { useCallback } from "react";
import { CustomSelect } from "@/components/custom-select";
import { FolderGitIcon } from "@/components/icons/FolderGitIcon";
import { ProjectSelect } from "@/components/project/ProjectSelect";
import { useComposeGitAssignment } from "@/hooks/useComposeGitAssignment";
import type { ComposeGitAssignment } from "@/lib/composeGitAssignment";

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
        <Link to="/repositories?register=1" target="_blank" rel="noopener noreferrer">
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
          <FolderGitIcon className="worktrees-git-selector__icon" />
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

      <p className="worktrees-git-selector__manage">
        Hamix allocates a worktree and branch when the task is created.{" "}
        <Link to="/repositories" target="_blank" rel="noopener noreferrer">
          Inspect repositories
        </Link>
      </p>
    </div>
  );
}
