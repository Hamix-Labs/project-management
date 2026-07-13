import { useId } from "react";
import { Link } from "react-router-dom";
import {
  AGENT_HEADING_ID,
  type TaskCreateModalAgentSectionProps,
} from "./taskCreateModalAgentShared";
import { TaskCreateModalModelField } from "./TaskCreateModalModelField";
import { TaskCreateModalRunnerField } from "./TaskCreateModalRunnerField";

/**
 * TaskCreateModalAgentSection - runtime configuration panel for the
 * new task. Picks the runner (where the task executes) and the model
 * (which underlying LLM the runner drives).
 *
 * Uses the shared CustomSelect portal dropdown so option lists match
 * Priority and other polished controls (native <select> menus
 * cannot be styled consistently across browsers).
 */
export function TaskCreateModalAgentSection({
  disabled,
  lockRunner = false,
  variant = "default",
  runner,
  cursorModel,
  modelIds,
  modelsForSelect,
  modelSelectBusy,
  modelFetchError,
  modelServerError,
  onRunnerChange,
  onCursorModelChange,
}: TaskCreateModalAgentSectionProps) {
  const baseId = useId();
  const runnerId = `${baseId}-runner`;
  const modelId = `${baseId}-model`;

  const isModelDialog = variant === "modelDialog";
  const isCreateModal = variant === "createModal";

  return (
    <section
      className={[
        isModelDialog
          ? "task-create-agent task-create-agent--model-dialog"
          : "task-create-agent",
        isCreateModal ? "task-create-agent--create-modal" : "",
      ]
        .filter(Boolean)
        .join(" ")}
      aria-labelledby={AGENT_HEADING_ID}
    >
      <h3 id={AGENT_HEADING_ID} className="task-create-subtasks-heading">
        Agent
      </h3>
      <div className="task-create-agent-panel">
        {lockRunner && !isModelDialog ? (
          <p className="task-create-agent-lock-notice" role="note">
            <strong>Runner</strong> is fixed-it was chosen when this task was
            created and can&apos;t be changed here.{" "}
            <strong>Model</strong> below can override the workspace default for
            this task only. Workspace CLI path and default model:{" "}
            <Link
              to="/settings#cursor-agent"
              className="task-create-agent-lock-notice-link"
            >
              Settings → Cursor agent
            </Link>
            .
          </p>
        ) : null}
        <div className="task-create-agent-grid">
          <TaskCreateModalRunnerField
            runnerId={runnerId}
            disabled={disabled}
            lockRunner={lockRunner}
            variant={variant}
            runner={runner}
            onRunnerChange={onRunnerChange}
          />
          <TaskCreateModalModelField
            modelId={modelId}
            disabled={disabled}
            lockRunner={lockRunner}
            variant={variant}
            runner={runner}
            cursorModel={cursorModel}
            modelIds={modelIds}
            modelsForSelect={modelsForSelect}
            modelSelectBusy={modelSelectBusy}
            modelFetchError={modelFetchError}
            modelServerError={modelServerError}
            onCursorModelChange={onCursorModelChange}
          />
        </div>
      </div>
    </section>
  );
}
