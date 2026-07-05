import type { PriorityChoice } from "@/types";
import { TaskComposeTitlePriorityRow } from "../../task-compose/fields/TaskComposeTitlePriorityRow";
import { WorktreeSelector } from "./WorktreeSelector";

type Props = {
  idsPrefix: string;
  title: string;
  priority: PriorityChoice;
  projectId: string;
  worktreeId: string;
  disabled: boolean;
  showWorktree: boolean;
  onTitleChange: (v: string) => void;
  onPriorityChange: (p: PriorityChoice) => void;
  onWorktreeChange: (worktreeId: string) => void;
};

export function TaskCreateModalEssentialsFields({
  idsPrefix,
  title,
  priority,
  projectId,
  worktreeId,
  disabled,
  showWorktree,
  onTitleChange,
  onPriorityChange,
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
        <WorktreeSelector
          idsPrefix={idsPrefix}
          projectId={projectId}
          worktreeId={worktreeId}
          onWorktreeChange={onWorktreeChange}
          disabled={disabled}
        />
      ) : null}
    </>
  );
}
