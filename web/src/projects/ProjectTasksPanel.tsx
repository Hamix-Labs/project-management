import { useMemo } from "react";
import { useQuery } from "@tanstack/react-query";
import { Link } from "react-router-dom";
import { listTasks } from "@/api";
import { StatusBadge } from "@/components/task-status";
import { STATUS_META } from "@/lib/taskStatusDisplay";
import { taskQueryKeys } from "@/lib/taskQueryKeys";
import type { Status } from "@/types";

type Props = {
  projectId: string;
};

const DISPLAY_LIMIT = 12;

export function ProjectTasksPanel({ projectId }: Props) {
  const projectTasks = useQuery({
    queryKey: taskQueryKeys.list({ limit: 200, offset: 0 }),
    queryFn: ({ signal }) => listTasks(200, 0, { signal }),
    enabled: Boolean(projectId),
  });

  const memberTasks = useMemo(
    () =>
      (projectTasks.data?.tasks ?? []).filter(
        (task) => task.project_id === projectId,
      ),
    [projectTasks.data?.tasks, projectId],
  );

  const visibleTasks = memberTasks.slice(0, DISPLAY_LIMIT);

  const statusCounts = useMemo(() => {
    const counts = new Map<Status, number>();
    for (const task of memberTasks) {
      counts.set(task.status, (counts.get(task.status) ?? 0) + 1);
    }
    return [...counts.entries()].sort(
      ([a], [b]) => STATUS_META[a].order - STATUS_META[b].order,
    );
  }, [memberTasks]);

  const totalLabel =
    memberTasks.length === 1
      ? "1 task connected to this project"
      : `${memberTasks.length} tasks connected to this project`;

  return (
    <section className="pd__card" aria-labelledby="pd-tasks-title">
      <div className="pd__card-head pd__card-head--tasks">
        <div className="pd__card-head-text">
          <h2 id="pd-tasks-title" className="pd__card-title">
            Linked tasks
          </h2>
          <p className="pd__card-desc">
            {projectTasks.isLoading ? "Loading linked tasks…" : totalLabel}
          </p>
        </div>
        {statusCounts.length > 0 ? (
          <ul className="pd__status-summary" aria-label="Task status summary">
            {statusCounts.map(([status, count]) => (
              <li key={status}>
                <span
                  className={`pd__status-summary-chip pd__status-summary-chip--${STATUS_META[status].tone}`}
                  title={`${STATUS_META[status].label}: ${count}`}
                >
                  <span className="pd__status-summary-dot" aria-hidden="true" />
                  <span className="pd__status-summary-count">{count}</span>
                  <span className="visually-hidden">
                    {STATUS_META[status].label}
                  </span>
                </span>
              </li>
            ))}
          </ul>
        ) : null}
      </div>

      {projectTasks.isLoading ? <TaskListSkeleton /> : null}

      {!projectTasks.isLoading && memberTasks.length === 0 ? (
        <p className="pd__empty">No tasks linked to this project yet</p>
      ) : null}

      {visibleTasks.length > 0 ? (
        <ul className="pd__task-list">
          {visibleTasks.map((task) => (
            <li key={task.id}>
              <Link
                to={`/tasks/${encodeURIComponent(task.id)}`}
                className="pd__task-row"
              >
                <span className="pd__task-title-group">
                  <span className="pd__task-title">{task.title}</span>
                  <svg
                    className="pd__task-external"
                    width="14"
                    height="14"
                    viewBox="0 0 16 16"
                    fill="none"
                    aria-hidden="true"
                  >
                    <path
                      d="M4.5 11.5l7-7M7 4.5h4.5V9"
                      stroke="currentColor"
                      strokeWidth="1.5"
                      strokeLinecap="round"
                      strokeLinejoin="round"
                    />
                  </svg>
                </span>
                <StatusBadge status={task.status} />
              </Link>
            </li>
          ))}
        </ul>
      ) : null}
    </section>
  );
}

function TaskListSkeleton() {
  return (
    <div className="pd__task-skeleton" aria-hidden="true">
      {Array.from({ length: 3 }).map((_, i) => (
        <div key={i} className="pd__task-skeleton-row">
          <span className="pd__shimmer" style={{ width: `${58 - i * 12}%`, height: "0.875rem" }} />
          <span className="pd__shimmer" style={{ width: "3.5rem", height: "0.75rem" }} />
        </div>
      ))}
    </div>
  );
}
