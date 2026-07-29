import type { Status } from "@/types";
import { TaskCreateModalAdvancedOptions } from "./TaskCreateModalAdvancedOptions";
import { TaskCreateModalAutonomyToggle } from "./fields/TaskCreateModalAutonomyToggle";
import { TaskCreateModalSection } from "./fields/TaskCreateModalSection";
import type { TaskCreateModalPresentation } from "./taskCreateModalPresentation";

type Props = {
  presentation: TaskCreateModalPresentation;
  editingTaskRunner: string;
  autonomyEnabled: boolean;
  autonomyDisabled: boolean;
  onAutonomyChange: (enabled: boolean) => void;
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

export function TaskCreateModalExecutionSection({
  presentation,
  editingTaskRunner,
  autonomyEnabled,
  autonomyDisabled,
  onAutonomyChange,
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
    <TaskCreateModalSection
      variant="execution"
      title="Execution"
      lede="Choose whether the agent should start on its own and which runner to use."
    >
      <TaskCreateModalAutonomyToggle
        enabled={autonomyEnabled}
        disabled={presentation.disabled || autonomyDisabled}
        onChange={onAutonomyChange}
      />

      <TaskCreateModalAdvancedOptions
        presentation={presentation}
        editingTaskRunner={editingTaskRunner}
        taskRunner={taskRunner}
        taskCursorModel={taskCursorModel}
        onTaskRunnerChange={onTaskRunnerChange}
        onTaskCursorModelChange={onTaskCursorModelChange}
        onComposeStatusChange={onComposeStatusChange}
        schedule={schedule}
        onScheduleChange={onScheduleChange}
        appTimezone={appTimezone}
        tagsCsv={tagsCsv}
        milestone={milestone}
        projectId={projectId}
        dependsOn={dependsOn}
        onTagsCsvChange={onTagsCsvChange}
        onMilestoneChange={onMilestoneChange}
        onDependsOnChange={onDependsOnChange}
      />
    </TaskCreateModalSection>
  );
}
