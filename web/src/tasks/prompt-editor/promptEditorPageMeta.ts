import type { PromptSourceKind } from "./types";
import { previewTextFromPrompt } from "@/lib/promptFormat";

export function crumbKindLabel(kind: PromptSourceKind | string): string {
  switch (kind) {
    case "draft":
      return "Draft";
    case "task":
      return "Task";
    case "template":
      return "Template";
    case "ephemeral":
      return "New task";
    default:
      return "Prompt";
  }
}

export function formatRelativeEdited(at: number | null): string {
  if (at == null) return "Not edited yet";
  const sec = Math.max(0, Math.floor((Date.now() - at) / 1000));
  if (sec < 45) return "Edited just now";
  if (sec < 90) return "Edited 1 minute ago";
  if (sec < 3600) return `Edited ${Math.floor(sec / 60)} minutes ago`;
  if (sec < 7200) return "Edited 1 hour ago";
  return `Edited ${Math.floor(sec / 3600)} hours ago`;
}

export function wordCountFromHtml(html: string): number {
  const text = previewTextFromPrompt(html);
  if (!text) return 0;
  return text.split(/\s+/).filter(Boolean).length;
}

export function repoBasename(path: string): string {
  const norm = path.replace(/\\/g, "/").replace(/\/+$/, "");
  const parts = norm.split("/");
  const base = parts[parts.length - 1] || path;
  return base.endsWith(" repo") ? base : `${base} repo`;
}
