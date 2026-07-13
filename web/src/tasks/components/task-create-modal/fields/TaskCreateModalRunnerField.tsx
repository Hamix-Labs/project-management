import { CustomSelect } from "@/components/custom-select";
import {
  RUNNER_OPTIONS,
  type TaskCreateModalAgentSectionVariant,
} from "./taskCreateModalAgentShared";

type Props = {
  runnerId: string;
  disabled: boolean;
  lockRunner: boolean;
  variant: TaskCreateModalAgentSectionVariant;
  runner: string;
  onRunnerChange: (runner: string) => void;
};

export function TaskCreateModalRunnerField({
  runnerId,
  disabled,
  lockRunner,
  variant,
  runner,
  onRunnerChange,
}: Props) {
  const isCreateModal = variant === "createModal";
  const showRunnerHelp = !isCreateModal || lockRunner;

  return (
    <div className="task-create-agent-field">
      <CustomSelect
        id={runnerId}
        label="Runner"
        value={runner}
        options={RUNNER_OPTIONS}
        disabled={disabled || lockRunner}
        onChange={onRunnerChange}
        className="task-create-agent-custom-select"
      />
      {showRunnerHelp ? (
        <p className="task-create-agent-help">
          {lockRunner
            ? "Set when this task was created; the runner can't be changed for an existing task."
            : isCreateModal
              ? "Saved with the task and fixed after creation."
              : "Pick the runtime for this task. It's saved when the task is created and can't be changed later."}
        </p>
      ) : null}
    </div>
  );
}
