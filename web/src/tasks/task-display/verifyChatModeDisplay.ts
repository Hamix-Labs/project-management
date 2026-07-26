export type VerifyChatMode = "same_chat" | "different_chat";

const LABELS: Record<VerifyChatMode, string> = {
  same_chat: "Same chat",
  different_chat: "Different chat",
};

/**
 * Resolves effective verify chat mode: non-empty task override wins,
 * else settings default, else same_chat (matches worker EffectiveVerifyChatMode).
 */
export function effectiveVerifyChatMode(
  taskMode: string | undefined | null,
  settingsMode: string | undefined | null,
): VerifyChatMode {
  const task = (taskMode ?? "").trim();
  if (task === "same_chat" || task === "different_chat") {
    return task;
  }
  const settings = (settingsMode ?? "").trim();
  if (settings === "different_chat") {
    return "different_chat";
  }
  return "same_chat";
}

export function verifyChatModeLabel(mode: VerifyChatMode): string {
  return LABELS[mode];
}

export function verifyChatModeSource(
  taskMode: string | undefined | null,
): "task" | "workspace" {
  const task = (taskMode ?? "").trim();
  if (task === "same_chat" || task === "different_chat") {
    return "task";
  }
  return "workspace";
}
