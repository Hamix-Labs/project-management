import type { ChecklistItemDraft, ChecklistVerifyCommandInput } from "@/types";

export const CREATE_CHECKLIST_REQUIRED_MSG = "Add at least one done criterion.";

export const MAX_VERIFY_COMMANDS_PER_ITEM = 5;

export function nonEmptyChecklistCount(
  items: ReadonlyArray<ChecklistItemDraft | string>,
): number {
  let n = 0;
  for (const raw of items) {
    const text = typeof raw === "string" ? raw : raw.text;
    if (text.trim() !== "") {
      n++;
    }
  }
  return n;
}

export function normalizeVerifyCommands(
  cmds: ReadonlyArray<ChecklistVerifyCommandInput>,
): ChecklistVerifyCommandInput[] {
  const out: ChecklistVerifyCommandInput[] = [];
  for (const raw of cmds) {
    const command = raw.command.trim();
    if (!command) continue;
    const expected = raw.expected_outcome?.trim();
    const timeout =
      typeof raw.timeout_seconds === "number" &&
      Number.isFinite(raw.timeout_seconds) &&
      raw.timeout_seconds > 0
        ? Math.floor(raw.timeout_seconds)
        : undefined;
    out.push({
      command,
      ...(expected ? { expected_outcome: expected } : {}),
      ...(timeout !== undefined ? { timeout_seconds: timeout } : {}),
    });
    if (out.length >= MAX_VERIFY_COMMANDS_PER_ITEM) break;
  }
  return out;
}

export function normalizeChecklistItems(
  items: ReadonlyArray<ChecklistItemDraft>,
): Array<{ text: string; verify_commands?: ChecklistVerifyCommandInput[] }> {
  return items
    .map((item) => {
      const text = item.text.trim();
      if (!text) return null;
      const verify_commands = normalizeVerifyCommands(item.verify_commands ?? []);
      return {
        text,
        ...(verify_commands.length > 0 ? { verify_commands } : {}),
      };
    })
    .filter((item): item is { text: string; verify_commands?: ChecklistVerifyCommandInput[] } =>
      item !== null,
    );
}

export function emptyVerifyCommandRow(): ChecklistVerifyCommandInput {
  return { command: "", expected_outcome: "" };
}
