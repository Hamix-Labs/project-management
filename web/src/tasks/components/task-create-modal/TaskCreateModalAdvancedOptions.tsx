import type { Status } from "@/types";
import { SchedulePicker } from "@/shared/time/SchedulePicker";
import { useTaskCreateAgentOptions } from "@/tasks/create/hooks/useTaskCreateAgentOptions";
import { advancedSummaryLine } from "./advancedSummaryLine";
import { TaskCreateModalAgentSection } from "./fields/TaskCreateModalAgentSection";
import {
  AgentCalendarIcon,
  AgentListIcon,
} from "./fields/TaskCreateAgentIcons";
import { TaskCreateConfigSectionHeader } from "./fields/TaskCreateConfigSectionHeader";
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
  /**
   * When true, skip Agent section (compose page promotes it to the rail).
   * Default false preserves modal behavior.
   */
  omitAgent?: boolean;
  /**
   * When true, hide Tags inside scheduling fields (compose page promotes Tags).
   * Milestone / depends-on still render when enabled. Default false.
   */
  omitTags?: boolean;
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
  omitAgent = false,
  omitTags = false,
}: Props) {
  const agentRunner = presentation.isTaskEdit ? editingTaskRunner : taskRunner;
  const agentOptions = useTaskCreateAgentOptions(agentRunner);

  const showStatus = Boolean(
    presentation.isTaskEdit && onComposeStatusChange,
  );
  const showSchedule = presentation.scheduleUiEnabled;
  const showTags = presentation.tagsUiEnabled && !omitTags;
  const showDeps =
    presentation.dependenciesUiEnabled;
  const showTagsOrDeps = showTags || showDeps;
  const showAfterTags = showStatus || showSchedule;

  const summaryLine = advancedSummaryLine({
    runner: presentation.isTaskEdit ? editingTaskRunner : taskRunner,
    cursorModel: taskCursorModel,
    schedule,
    tagsCsv: omitTags ? "" : tagsCsv,
    milestone,
    dependsOn,
    includeSchedule: presentation.scheduleUiEnabled,
    includeTags: presentation.tagsUiEnabled && !omitTags,
    includeDependencies: presentation.dependenciesUiEnabled,
  });

  return (
    <details className="task-create-advanced">
      <summary
        className="task-create-advanced__summary"
        data-testid="task-create-more-options-toggle"
      >
        <span className="task-create-advanced__chevron" aria-hidden="true" />
        <span className="task-create-advanced__label">More options</span>
        <span className="task-create-advanced__hint">{summaryLine}</span>
      </summary>
      <div className="task-create-advanced__body">
        {!omitAgent ? (
          <TaskCreateModalAgentSection
            disabled={presentation.disabled}
            variant="createModal"
            lockRunner={presentation.isTaskEdit}
            runner={agentRunner}
            cursorModel={taskCursorModel}
            {...agentOptions}
            onRunnerChange={
              presentation.isTaskEdit ? () => {} : onTaskRunnerChange
            }
            onCursorModelChange={onTaskCursorModelChange}
          />
        ) : null}

        {showTagsOrDeps ? (
          <>
            {!omitAgent ? (
              <div className="task-create-advanced__divider" role="separator" />
            ) : null}
            <TaskCreateModalSchedulingFields
              disabled={presentation.disabled}
              tagsCsv={tagsCsv}
              milestone={milestone}
              projectId={projectId}
              dependsOn={dependsOn}
              showTags={showTags}
              showMilestone={showDeps}
              showDependsOn={showDeps}
              dependsOnDisabled={presentation.isTaskEdit}
              configChrome
              onTagsCsvChange={onTagsCsvChange}
              onMilestoneChange={onMilestoneChange}
              onDependsOnChange={
                presentation.isTaskEdit
                  ? noopOnDependsOnChange
                  : onDependsOnChange
              }
            />
          </>
        ) : null}

        {showAfterTags ? (
          <div className="task-create-advanced__divider" role="separator" />
        ) : null}

        {showSchedule ? (
          <div className="task-create-config-section">
            <TaskCreateConfigSectionHeader
              id="task-create-schedule-heading"
              title="Schedule"
              icon={<AgentCalendarIcon />}
            />
            {presentation.isTaskEdit ? (
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
            )}
          </div>
        ) : null}

        {showStatus && onComposeStatusChange ? (
          <>
            {showSchedule ? (
              <div className="task-create-advanced__divider" role="separator" />
            ) : null}
            <div className="task-create-config-section">
              <TaskCreateConfigSectionHeader
                id="task-create-status-heading"
                title="Status"
                icon={<AgentListIcon />}
              />
              <TaskCreateModalStatusField
                id={`${presentation.idsPrefix}-status`}
                status={presentation.status}
                disabled={presentation.disabled}
                onChange={onComposeStatusChange}
              />
            </div>
          </>
        ) : null}
      </div>
    </details>
  );
}
