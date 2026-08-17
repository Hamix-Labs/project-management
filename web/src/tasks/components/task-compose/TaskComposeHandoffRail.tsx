import { TaskCreateModalAdvancedOptions } from "../task-create-modal/TaskCreateModalAdvancedOptions";
import type { TaskCreateModalProps } from "../task-create-modal/taskCreateModalProps";
import type { TaskCreateModalPresentation } from "../task-create-modal/taskCreateModalPresentation";
import { TaskComposeAgentCard } from "./rail/TaskComposeAgentCard";
import { TaskComposeDestinationCard } from "./rail/TaskComposeDestinationCard";
import { TaskComposePriorityCard } from "./rail/TaskComposePriorityCard";
import { TaskComposeTagsCard } from "./rail/TaskComposeTagsCard";

type Props = {
  presentation: TaskCreateModalPresentation;
  session: TaskCreateModalProps["session"];
  essentials: TaskCreateModalProps["essentials"];
  criteria: TaskCreateModalProps["criteria"];
  git: TaskCreateModalProps["git"];
  execution: TaskCreateModalProps["execution"];
  appTimezone: TaskCreateModalProps["appTimezone"];
  editingTaskRunner: string;
};

/** Right-hand handoff rail for compose v2. */
export function TaskComposeHandoffRail({
  presentation,
  session,
  essentials,
  criteria,
  git,
  execution,
  appTimezone,
  editingTaskRunner,
}: Props) {
  const autonomyDisabled = execution.autonomyDisabled ?? false;
  const agentRunner = presentation.isTaskEdit
    ? editingTaskRunner
    : execution.taskRunner;
  const showTagsCard = presentation.tagsUiEnabled;
  const showMoreOptions =
    presentation.scheduleUiEnabled ||
    presentation.dependenciesUiEnabled ||
    Boolean(presentation.isTaskEdit && session.onComposeStatusChange);

  return (
    <>
      <div className="compose-handoff">
        <TaskComposeDestinationCard
          idsPrefix={presentation.idsPrefix}
          repositoryId={git.repositoryId}
          projectId={git.projectId}
          worktreeId={git.worktreeId}
          assignmentLocked={git.assignmentLocked === true}
          disabled={presentation.disabled}
          showWorktree={!presentation.isTaskEdit}
          onRepositoryChange={git.onRepositoryChange}
          onProjectChange={git.onProjectChange}
          onWorktreeChange={git.onWorktreeChange}
        />
        <TaskComposePriorityCard
          value={essentials.priority}
          disabled={presentation.disabled}
          onChange={essentials.onPriorityChange}
        />
        <TaskComposeAgentCard
          disabled={presentation.disabled}
          lockRunner={presentation.isTaskEdit}
          runner={agentRunner}
          cursorModel={execution.taskCursorModel}
          autonomyEnabled={execution.autonomyEnabled}
          autonomyDisabled={autonomyDisabled}
          onRunnerChange={execution.onTaskRunnerChange}
          onCursorModelChange={execution.onTaskCursorModelChange}
          onAutonomyChange={execution.onAutonomyChange}
        />
        {showTagsCard ? (
          <TaskComposeTagsCard
            tagsCsv={criteria.tagsCsv}
            disabled={presentation.disabled}
            onTagsCsvChange={criteria.onTagsCsvChange}
          />
        ) : null}
      </div>
      {showMoreOptions ? (
        <div className="compose-more-options">
          <TaskCreateModalAdvancedOptions
            presentation={presentation}
            editingTaskRunner={editingTaskRunner}
            taskRunner={execution.taskRunner}
            taskCursorModel={execution.taskCursorModel}
            onTaskRunnerChange={execution.onTaskRunnerChange}
            onTaskCursorModelChange={execution.onTaskCursorModelChange}
            onComposeStatusChange={session.onComposeStatusChange}
            schedule={execution.schedule}
            onScheduleChange={execution.onScheduleChange}
            appTimezone={appTimezone}
            tagsCsv={criteria.tagsCsv}
            milestone={execution.milestone}
            projectId={git.projectId}
            dependsOn={execution.dependsOn}
            onTagsCsvChange={criteria.onTagsCsvChange}
            onMilestoneChange={execution.onMilestoneChange}
            onDependsOnChange={execution.onDependsOnChange}
            omitAgent
            omitTags
          />
        </div>
      ) : null}
    </>
  );
}
