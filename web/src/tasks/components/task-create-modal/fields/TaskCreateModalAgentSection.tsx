import { useId, useMemo } from "react";
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

function buildVerifyModeOptions(
  workspaceDefault: "same_chat" | "different_chat",
): CustomSelectOption[] {
  return (["same_chat", "different_chat"] as const).map((value) => ({
    value,
    label: verifyChatModeLabel(value),
    ...(value === workspaceDefault ? { rowTag: "Default" } : {}),
  }));
}

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
  workspaceVerifyChatMode = "same_chat",
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

  const verifyOptions = useMemo(
    () => buildVerifyModeOptions(workspaceVerifyChatMode),
    [workspaceVerifyChatMode],
  );

  // Empty task value inherits workspace settings — show that concrete
  // mode in the select rather than a third "Use workspace default" row.
  const verifySelectValue =
    verifyChatMode === "same_chat" || verifyChatMode === "different_chat"
      ? verifyChatMode
      : workspaceVerifyChatMode;

  function handleVerifyModeChange(next: string) {
    if (!onVerifyChatModeChange) return;
    // Selecting the workspace default keeps inherit (empty) so later
    // settings changes still apply; picking the other mode pins override.
    onVerifyChatModeChange(
      next === workspaceVerifyChatMode ? "" : next,
    );
  }

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
              label="Verification mode"
              value={verifySelectValue}
              options={verifyOptions}
              disabled={disabled}
              onChange={handleVerifyModeChange}
              className="task-create-agent-custom-select"
              triggerTestId="task-verify-chat-mode"
              leadingIcon={<AgentShieldCheckIcon />}
            />
            <p className="task-create-agent-help">
              Controls whether PhaseVerify continues in the same chat or starts
              a different one. The workspace default is marked in the list.
            </p>
          </div>
        ) : null}
      </div>
    </section>
  );
}
