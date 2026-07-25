import { Link } from "react-router-dom";
import { PriorityBadge } from "@/tasks/task-display";
import type { TaskWithDepth } from "../../task-tree";

type Props = {
  task: TaskWithDepth;
  projectName?: string;
  showProject: boolean;
};

export function TaskBoardCard({ task, projectName, showProject }: Props) {
  return (
    <Link
      to={`/tasks/${encodeURIComponent(task.id)}`}
      className="task-board-card"
    >
      <span className="task-board-card__title">{task.title}</span>
      <span className="task-board-card__meta">
        <PriorityBadge priority={task.priority} />
        {showProject && projectName ? (
          <span className="task-board-card__project">{projectName}</span>
        ) : null}
      </span>
    </Link>
  );
}
