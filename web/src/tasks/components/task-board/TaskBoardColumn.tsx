import type { BoardColumnDef } from "./boardColumns";
import type { TaskWithDepth } from "../../task-tree";
import { TaskBoardCard } from "./TaskBoardCard";

type Props = {
  column: BoardColumnDef;
  tasks: TaskWithDepth[];
  projectNameById: Record<string, string>;
  showProject: boolean;
  showTags: boolean;
};

export function TaskBoardColumn({
  column,
  tasks,
  projectNameById,
  showProject,
  showTags,
}: Props) {
  const headingId = `task-board-col-${column.id}`;
  return (
    <section
      className={`task-board-column task-board-column--tone-${column.tone}`}
      aria-labelledby={headingId}
    >
      <header className="task-board-column__head">
        <div className="task-board-column__head-main">
          <span className="task-board-column__dot" aria-hidden="true" />
          <h3 id={headingId} className="task-board-column__title">
            {column.label}
          </h3>
          <span
            className="task-board-column__count"
            aria-label={`${tasks.length} tasks`}
          >
            {tasks.length}
          </span>
        </div>
      </header>
      <div className="task-board-column__body">
        {tasks.length === 0 ? (
          <div className="task-board-column__empty">No tasks</div>
        ) : (
          tasks.map((task) => (
            <TaskBoardCard
              key={task.id}
              task={task}
              showProject={showProject}
              showTags={showTags}
              projectName={
                task.project_id
                  ? projectNameById[task.project_id]
                  : undefined
              }
            />
          ))
        )}
      </div>
    </section>
  );
}
