import type { Status } from "@/types";
import { SchedulePicker } from "@/shared/time/SchedulePicker";
import { advancedSummaryLine } from "./advancedSummaryLine";
import { TaskCreateModalAgentSection } from "./fields/TaskCreateModalAgentSection";
import { TaskCreateModalPickupScheduleField } from "./fields/TaskCreateModalPickupScheduleField";
import { TaskCreateModalSchedulingFields } from "./fields/TaskCreateModalSchedulingFields";
import { TaskCreateModalStatusField } from "./fields/TaskCreateModalStatusField";
import type { TaskCreateModalPresentation } from "./taskCreateModalPresentation";

const noopOnDependsOnChange = (): void => {};

type Props = {
  presentation: TaskCreateModalPresentation;
  editingTaskRunner: string;
  taskRunner: string;
  taskCursorModel: string;
  onTaskRunnerChange: (runner: string) => void;
  onTaskCursorModelChange: (v: string) => void;
  onComposeStatusChange?: (status: Status) => void;
  schedule: string | null;
  onScheduleChange: (next: string | null) => void;
  appTimezone: string;
  tagsCsv: string;
  milestone: string;
  projectId: string;
  dependsOn: string[];
  onTagsCsvChange: (value: string) => void;
  onMilestoneChange: (value: string) => void;
  onDependsOnChange: (value: string[]) => void;
};

export function TaskCreateModalAdvancedOptions({
  presentation,
  editingTaskRunner,
  taskRunner,
  taskCursorModel,
  onTaskRunnerChange,
  onTaskCursorModelChange,
  onComposeStatusChange,
  schedule,
  onScheduleChange,
  appTimezone,
  tagsCsv,
  milestone,
  projectId,
  dependsOn,
  onTagsCsvChange,
  onMilestoneChange,
  onDependsOnChange,
}: Props) {
  return (
    <details className="task-create-advanced">
      <summary
        className="task-create-advanced__summary"
        data-testid="task-create-more-options-toggle"
      >
        <span className="task-create-advanced__chevron" aria-hidden="true" />
        <span className="task-create-advanced__label">More options</span>
        <span className="task-create-advanced__hint">
          {advancedSummaryLine({
            runner: presentation.isTaskEdit ? editingTaskRunner : taskRunner,
            cursorModel: taskCursorModel,
            schedule,
            tagsCsv,
            milestone,
            dependsOn,
            includeSchedule: presentation.scheduleUiEnabled,
            includeTagsAndDependencies: presentation.tagsAndDependenciesUiEnabled,
          })}
        </span>
      </summary>
      <div className="task-create-advanced__body">
        {presentation.isTaskEdit && onComposeStatusChange ? (
          <TaskCreateModalStatusField
            id={`${presentation.idsPrefix}-status`}
            status={presentation.status}
            disabled={presentation.disabled}
            onChange={onComposeStatusChange}
          />
        ) : null}

        <TaskCreateModalAgentSection
          disabled={presentation.disabled}
          variant="createModal"
          lockRunner={presentation.isTaskEdit}
          runner={presentation.isTaskEdit ? editingTaskRunner : taskRunner}
          cursorModel={taskCursorModel}
          onRunnerChange={presentation.isTaskEdit ? () => {} : onTaskRunnerChange}
          onCursorModelChange={onTaskCursorModelChange}
        />

        {presentation.scheduleUiEnabled ? (
          presentation.isTaskEdit ? (
            <TaskCreateModalPickupScheduleField
              status={presentation.status}
              value={schedule}
              onChange={onScheduleChange}
              appTimezone={appTimezone}
              disabled={presentation.disabled}
              idPrefix={`${presentation.idsPrefix}-modal`}
            />
          ) : (
            <SchedulePicker
              value={schedule}
              onChange={onScheduleChange}
              appTimezone={appTimezone}
              disabled={presentation.disabled}
              idPrefix="task-create-modal"
            />
          )
        ) : null}

        {presentation.tagsAndDependenciesUiEnabled ? (
          <TaskCreateModalSchedulingFields
            disabled={presentation.disabled}
            tagsCsv={tagsCsv}
            milestone={milestone}
            projectId={projectId}
            dependsOn={dependsOn}
            showDependsOn
            dependsOnDisabled={presentation.isTaskEdit}
            onTagsCsvChange={onTagsCsvChange}
            onMilestoneChange={onMilestoneChange}
            onDependsOnChange={
              presentation.isTaskEdit ? noopOnDependsOnChange : onDependsOnChange
            }
          />
        ) : null}
      </div>
    </details>
  );
}
