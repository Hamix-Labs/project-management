import type { SettingsSelectOption, SettingsSelectRow } from "./settingsSelectTypes";

function extractModelFamily(label: string): string {
  const codex = label.match(/^(Codex \d+(?:\.\d+)?(?: Max)?)/i);
  if (codex) return codex[1];
  const gpt = label.match(/^(GPT-[\d.]+)/i);
  if (gpt) return gpt[1];
  const composer = label.match(/^(Composer [\d.]+)/i);
  if (composer) return composer[1];
  return "";
}

export function groupModelSelectRows(
  options: SettingsSelectOption[],
): SettingsSelectRow[] {
  const rows: SettingsSelectRow[] = [];
  let lastGroup = "";

  for (const opt of options) {
    if (opt.value === "") {
      rows.push({ type: "option", value: opt.value, label: opt.label });
      continue;
    }
    const group = extractModelFamily(opt.label);
    if (group && group !== lastGroup) {
      rows.push({ type: "header", label: group });
      lastGroup = group;
    }
    rows.push({ type: "option", value: opt.value, label: opt.label });
  }
  return rows;
}
