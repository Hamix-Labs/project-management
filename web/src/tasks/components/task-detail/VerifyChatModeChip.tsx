import { useAppSettingsQuery } from "@/hooks/useAppSettingsQuery";
import type { Task } from "@/types";
import {
  effectiveVerifyChatMode,
  verifyChatModeLabel,
  verifyChatModeSource,
} from "../../task-display/verifyChatModeDisplay";
import {
  GitBranchGlyph,
  MessagesSquareGlyph,
} from "./ExecutionBarGlyphs";

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
  const ModeGlyph =
    mode === "different_chat" ? GitBranchGlyph : MessagesSquareGlyph;

  return (
    <span
      className="task-verify-chat-mode-chip"
      data-testid="task-verify-chat-mode-chip"
      data-mode={mode}
      data-source={source}
      title={`${sourceLabel}: ${label}`}
      aria-label={`Verification mode: ${label} (${sourceLabel.toLowerCase()})`}
    >
      <ModeGlyph className="task-verify-chat-mode-chip-icon" />
      <span className="task-verify-chat-mode-chip-label">
        Verification mode
      </span>
      <span className="task-verify-chat-mode-chip-sep" aria-hidden="true">
        ·
      </span>
      <span className="task-verify-chat-mode-chip-value">{label}</span>
    </span>
  );
}
