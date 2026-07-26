import { useId } from "react";
import { Link } from "react-router-dom";
import { CustomSelect, type CustomSelectOption } from "@/components/custom-select";
import { verifyChatModeLabel } from "../../../task-display/verifyChatModeDisplay";
import { AgentBotIcon, AgentShieldCheckIcon } from "./TaskCreateAgentIcons";
import { TaskCreateConfigSectionHeader } from "./TaskCreateConfigSectionHeader";
import {
  AGENT_HEADING_ID,
  type TaskCreateModalAgentSectionProps,
} from "./taskCreateModalAgentShared";
import { TaskCreateModalModelField } from "./TaskCreateModalModelField";
import { TaskCreateModalRunnerField } from "./TaskCreateModalRunnerField";

const VERIFY_CHAT_OPTIONS: CustomSelectOption[] = [
  { value: "", label: "Use workspace default" },
  { value: "same_chat", label: verifyChatModeLabel("same_chat") },
  { value: "different_chat", label: verifyChatModeLabel("different_chat") },
];

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
  verifyChatMode = "",
  modelIds,
  modelsForSelect,
  modelSelectBusy,
  modelFetchError,
  modelServerError,
  onRunnerChange,
  onCursorModelChange,
  onVerifyChatModeChange,
}: TaskCreateModalAgentSectionProps) {
  const baseId = useId();
  const runnerId = `${baseId}-runner`;
  const modelId = `${baseId}-model`;
  const verifyChatId = `${baseId}-verify-chat`;

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
      {isCreateModal ? (
        <TaskCreateConfigSectionHeader
          id={AGENT_HEADING_ID}
          title="Agent"
          icon={<AgentBotIcon />}
        />
      ) : (
        <h3 id={AGENT_HEADING_ID} className="task-create-subtasks-heading">
          Agent
        </h3>
      )}
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
        {!isModelDialog && onVerifyChatModeChange ? (
          <div className="task-create-agent-verify-chat">
            <CustomSelect
              id={verifyChatId}
              label="Verify chat"
              value={verifyChatMode}
              options={VERIFY_CHAT_OPTIONS}
              disabled={disabled}
              onChange={onVerifyChatModeChange}
              className="task-create-agent-custom-select"
              triggerTestId="task-verify-chat-mode"
              leadingIcon={<AgentShieldCheckIcon />}
            />
            <p className="task-create-agent-help">
              Controls whether PhaseVerify continues the execute chat or starts
              a new one.
            </p>
          </div>
        ) : null}
      </div>
    </section>
  );
}
