import { useAppSettingsQuery } from "@/hooks/useAppSettingsQuery";
import type { Task } from "@/types";
import {
  effectiveVerifyChatMode,
  verifyChatModeLabel,
  verifyChatModeSource,
} from "../../task-display/verifyChatModeDisplay";

type Props = {
  task: Pick<Task, "verify_chat_mode">;
};

/**
 * Read-only toolbar chip for the effective verify chat mode
 * (task override, else workspace default).
 */
export function VerifyChatModeChip({ task }: Props) {
  const settingsQuery = useAppSettingsQuery();
  const mode = effectiveVerifyChatMode(
    task.verify_chat_mode,
    settingsQuery.data?.verify_chat_mode,
  );
  const label = verifyChatModeLabel(mode);
  const source = verifyChatModeSource(task.verify_chat_mode);
  const sourceLabel =
    source === "task" ? "Task override" : "Workspace default";

  return (
    <span
      className="task-verify-chat-mode-chip"
      data-testid="task-verify-chat-mode-chip"
      data-mode={mode}
      data-source={source}
      title={`${sourceLabel}: ${label}`}
      aria-label={`Verify chat: ${label} (${sourceLabel.toLowerCase()})`}
    >
      <span className="task-verify-chat-mode-chip-label">Verify chat</span>
      <span className="task-verify-chat-mode-chip-value">{label}</span>
    </span>
  );
}
