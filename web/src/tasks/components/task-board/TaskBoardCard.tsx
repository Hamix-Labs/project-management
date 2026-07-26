import { Link } from "react-router-dom";
import { PriorityBadge } from "@/tasks/task-display";
import { ProjectsStackIcon } from "@/components/project/ProjectsStackIcon";
import { taskDisplayRef } from "@/lib/taskShortId";
import { previewTextFromPrompt } from "@/lib/promptFormat";
import { taskListRowSubtitle } from "../task-list/table/taskListRowSubtitle";
import { formatRelativeTime } from "@/shared/time/relativeTime";
import { TASK_LIST_TAG_CHIP_LIMIT } from "../task-list/filters/taskListClientFilter";
import { taskBranchName } from "@/tasks/task-git/taskWorktreeNames";
import type { TaskWithDepth } from "../../task-tree";

type Props = {
  task: TaskWithDepth;
  projectName?: string;
  showProject: boolean;
  showTags: boolean;
};

function BranchGlyph() {
  return (
    <svg
      className="task-board-card__chip-icon"
      width="12"
      height="12"
      viewBox="0 0 16 16"
      fill="none"
      aria-hidden="true"
    >
      <path
        d="M4 3a1.5 1.5 0 1 1 0 3 1.5 1.5 0 0 1 0-3ZM12 10a1.5 1.5 0 1 1 0 3 1.5 1.5 0 0 1 0-3ZM4 10a1.5 1.5 0 1 1 0 3 1.5 1.5 0 0 1 0-3Z"
        stroke="currentColor"
        strokeWidth="1.2"
      />
      <path
        d="M4 6v4M4 7.5c0 2 2 3 4.5 3H12"
        stroke="currentColor"
        strokeWidth="1.2"
        strokeLinecap="round"
      />
    </svg>
  );
}

export function TaskBoardCard({
  task,
  projectName,
  showProject,
  showTags,
}: Props) {
  const summary = taskListRowSubtitle({
    promptPreview: previewTextFromPrompt(task.initial_prompt),
  });
  const tags = showTags ? (task.tags ?? []) : [];
  const visibleTags = tags.slice(0, TASK_LIST_TAG_CHIP_LIMIT);
  const overflow = tags.length - visibleTags.length;
  const created = formatRelativeTime(task.created_at);
  const branch = taskBranchName(task.id);
  const showProjectChip = Boolean(showProject && projectName);
  const showChips =
    Boolean(branch) || showProjectChip || visibleTags.length > 0;

  return (
    <Link
      to={`/tasks/${encodeURIComponent(task.id)}`}
      className="task-board-card"
    >
      <div className="task-board-card__top">
        <span className="task-board-card__id">{taskDisplayRef(task)}</span>
        <PriorityBadge priority={task.priority} />
      </div>

      <span className="task-board-card__title">{task.title}</span>

      {summary ? (
        <span className="task-board-card__summary">{summary}</span>
      ) : null}

      {showChips ? (
        <div className="task-board-card__chips">
          <span className="task-board-card__branch-chip">
            <BranchGlyph />
            {branch}
          </span>
          {showProjectChip ? (
            <span className="task-board-card__project-chip">
              <ProjectsStackIcon className="task-board-card__chip-icon" />
              {projectName}
            </span>
          ) : null}
          {visibleTags.map((tag) => (
            <span key={tag} className="task-board-card__tag-chip">
              {tag}
            </span>
          ))}
          {overflow > 0 ? (
            <span className="task-board-card__tag-overflow">+{overflow}</span>
          ) : null}
        </div>
      ) : null}

      {created ? (
        <div className="task-board-card__footer">
          <span className="task-board-card__time">{created}</span>
        </div>
      ) : null}
    </Link>
  );
}
