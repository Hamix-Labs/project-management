import type { PriorityChoice } from "@/types";
import { TaskCreateModalEssentialsFields } from "./fields/TaskCreateModalEssentialsFields";
import { TaskCreateModalSection } from "./fields/TaskCreateModalSection";
import type { TaskCreateModalPresentation } from "./taskCreateModalPresentation";

type Props = {
  presentation: TaskCreateModalPresentation;
  title: string;
  priority: PriorityChoice;
  repositoryId: string;
  projectId: string;
  worktreeId: string;
  onTitleChange: (v: string) => void;
  onPriorityChange: (p: PriorityChoice) => void;
  onRepositoryChange: (repositoryId: string) => void;
  onProjectChange: (projectId: string) => void;
  onWorktreeChange: (worktreeId: string) => void;
  onProjectContextClear: () => void;
};

export function TaskCreateModalEssentialsSection({
  presentation,
  title,
  priority,
  repositoryId,
  projectId,
  worktreeId,
  onTitleChange,
  onPriorityChange,
  onRepositoryChange,
  onProjectChange,
  onWorktreeChange,
  onProjectContextClear,
}: Props) {
  return (
    <TaskCreateModalSection
      variant="essentials"
      title="Essentials"
      lede="What to do, how urgent it is, and how success is judged."
    >
      <TaskCreateModalEssentialsFields
        idsPrefix={presentation.idsPrefix}
        title={title}
        priority={priority}
        repositoryId={repositoryId}
        projectId={projectId}
        worktreeId={worktreeId}
        disabled={presentation.disabled}
        showWorktree={!presentation.isTaskEdit}
        onTitleChange={onTitleChange}
        onPriorityChange={onPriorityChange}
        onRepositoryChange={onRepositoryChange}
        onProjectChange={onProjectChange}
        onWorktreeChange={onWorktreeChange}
        onProjectContextClear={onProjectContextClear}
      />
    </TaskCreateModalSection>
  );
}
