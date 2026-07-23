import { Link } from "react-router-dom";
import { cycleRunnerChipClass, runnerLabel } from "@/tasks/cycleDisplay/cyclesViewModel";
import type { Task } from "@/types";
import {
  priorityListLabel,
  priorityPillClass,
} from "../../../task-display";
import { TaskDetailGitBinding } from "./TaskDetailGitBinding";

type TaskDetailHeaderTask = Pick<
  Task,
  | "id"
  | "title"
  | "priority"
  | "runner"
  | "cursor_model"
  | "tags"
  | "milestone"
  | "worktree_id"
  | "project_id"
>;

type Props = {
  task: TaskDetailHeaderTask;
};

// formatTaskRuntime renders the header chip copy. Unlike
// `formatRunnerModel` (which reads `cursor_model_effective` off a
// cycle's meta — the truth the runner resolved), the header is about
// the task's INTENT: `task.cursor_model` is what the operator picked
// for the NEXT run, which may differ from any historical cycle's
// effective model. "default model" is the copy for the empty-intent
// case (the runner will fill in its adapter default at start).
function formatTaskRuntime(task: TaskDetailHeaderTask): string {
  const runner = runnerLabel(task.runner);
  if (runner === "unknown runner") {
    return runner;
  }
  const model = (task.cursor_model ?? "").trim();
  if (!model) {
    return `${runner} · default model`;
  }
  return `${runner} · ${model}`;
}

export function TaskDetailHeader({ task }: Props) {
  const milestone = (task.milestone ?? "").trim();
  const tags = task.tags ?? [];
  const hasSecondaryMeta = milestone !== "" || tags.length > 0;

  return (
    <>
      <nav className="task-detail-nav" aria-label="Task navigation">
        <Link to="/" className="pd__back project-context-back-link">
          <span aria-hidden="true">&#8249;</span>
          All tasks
        </Link>
      </nav>

      <header className="task-detail-header">
        <div className="task-detail-identity">
          <h2 className="task-detail-title term-arrow">
            <span>{task.title}</span>
          </h2>
          <div className="task-detail-meta">
            <span className={priorityPillClass(task.priority)}>
              {priorityListLabel(task.priority)} priority
            </span>
          </div>
        </div>

        <div className="task-detail-meta-secondary">
          <span
            className={`cell-pill ${cycleRunnerChipClass()} task-detail-runtime-chip`}
            data-testid="task-detail-runtime"
            aria-label="Agent for this task"
          >
            {formatTaskRuntime(task)}
          </span>
          {hasSecondaryMeta ? (
            <>
              {milestone !== "" ? (
                <span
                  className="cell-pill task-detail-milestone-chip"
                  data-testid="task-detail-milestone"
                >
                  {milestone}
                </span>
              ) : null}
              {tags.map((tag) => (
                <span
                  key={tag}
                  className="cell-pill task-detail-tag-chip"
                  data-testid="task-detail-tag"
                >
                  {tag}
                </span>
              ))}
            </>
          ) : null}
        </div>

        <TaskDetailGitBinding
          taskId={task.id}
          worktreeId={task.worktree_id}
          projectId={task.project_id}
        />
      </header>
    </>
  );
}
