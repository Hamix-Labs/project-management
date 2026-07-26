import type { PriorityChoice } from "@/types";
import { TaskComposeTitlePriorityRow } from "../../task-compose/fields/TaskComposeTitlePriorityRow";
import { TaskCreateAssignmentFields } from "./TaskCreateAssignmentFields";

type Props = {
  idsPrefix: string;
  title: string;
  priority: PriorityChoice;
  repositoryId: string;
  projectId: string;
  worktreeId: string;
  disabled: boolean;
  showWorktree: boolean;
  onTitleChange: (v: string) => void;
  onPriorityChange: (p: PriorityChoice) => void;
  onRepositoryChange: (repositoryId: string) => void;
  onProjectChange: (projectId: string) => void;
  onWorktreeChange: (worktreeId: string) => void;
};

export function TaskCreateModalEssentialsFields({
  idsPrefix,
  title,
  priority,
  repositoryId,
  projectId,
  worktreeId,
  disabled,
  showWorktree,
  onTitleChange,
  onPriorityChange,
  onRepositoryChange,
  onProjectChange,
  onWorktreeChange,
}: Props) {
  return (
    <>
      <TaskComposeTitlePriorityRow
        idsPrefix={idsPrefix}
        title={title}
        priority={priority}
        disabled={disabled}
        onTitleChange={onTitleChange}
        onPriorityChange={onPriorityChange}
        layout="modalEssentials"
      />
      {showWorktree ? (
        <TaskCreateAssignmentFields
          idsPrefix={idsPrefix}
          repositoryId={repositoryId}
          projectId={projectId}
          worktreeId={worktreeId}
          onAssignmentChange={(next) => {
            if (next.repositoryId !== repositoryId) {
              onRepositoryChange(next.repositoryId);
            }
            if (next.projectId !== projectId) {
              onProjectChange(next.projectId);
            }
            if (next.worktreeId !== worktreeId) {
              onWorktreeChange(next.worktreeId);
            }
          }}
          disabled={disabled}
        />
      ) : null}
    </>
  );
}
