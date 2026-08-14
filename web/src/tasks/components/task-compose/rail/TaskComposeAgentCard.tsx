import { AgentBotIcon } from "../../task-create-modal/fields/TaskCreateAgentIcons";
import { TaskCreateModalAgentSection } from "../../task-create-modal/fields/TaskCreateModalAgentSection";
import { TaskCreateModalAutonomyToggle } from "../../task-create-modal/fields/TaskCreateModalAutonomyToggle";
import { useTaskCreateAgentOptions } from "@/tasks/create/hooks/useTaskCreateAgentOptions";
import { ComposeRailSectionTitle } from "../TaskComposeBriefCard";

type Props = {
  disabled: boolean;
  lockRunner?: boolean;
  runner: string;
  cursorModel: string;
  autonomyEnabled: boolean;
  autonomyDisabled: boolean;
  onRunnerChange: (runner: string) => void;
  onCursorModelChange: (v: string) => void;
  onAutonomyChange: (enabled: boolean) => void;
};

export function TaskComposeAgentCard({
  disabled,
  lockRunner = false,
  runner,
  cursorModel,
  autonomyEnabled,
  autonomyDisabled,
  onRunnerChange,
  onCursorModelChange,
  onAutonomyChange,
}: Props) {
  const agentOptions = useTaskCreateAgentOptions(runner);

  return (
    <section
      className="compose-card compose-card--padded compose-agent"
      aria-labelledby="task-compose-agent-heading"
    >
      <ComposeRailSectionTitle icon={<AgentBotIcon />}>
        <span id="task-compose-agent-heading">Agent</span>
      </ComposeRailSectionTitle>
      <div className="compose-agent__fields">
        <TaskCreateModalAgentSection
          disabled={disabled}
          variant="createModal"
          hideHeader
          lockRunner={lockRunner}
          runner={runner}
          cursorModel={cursorModel}
          {...agentOptions}
          onRunnerChange={lockRunner ? () => {} : onRunnerChange}
          onCursorModelChange={onCursorModelChange}
        />
      </div>
      <div className="compose-agent__autonomy">
        <TaskCreateModalAutonomyToggle
          enabled={autonomyEnabled}
          disabled={disabled || autonomyDisabled}
          onChange={onAutonomyChange}
        />
      </div>
    </section>
  );
}
