import type { AppSettings } from "@/api/settings";
import { DEFAULT_NEW_TASK_STATUS, type Status } from "@/types";

export function generateTaskDraftID(): string {
  return typeof crypto !== "undefined" && typeof crypto.randomUUID === "function"
    ? crypto.randomUUID()
    : `draft-${Date.now().toString(36)}-${Math.random().toString(36).slice(2, 8)}`;
}

export function defaultRunnerFromSettings(settings: AppSettings | undefined): string {
  return (settings?.runner ?? "cursor").trim() || "cursor";
}

export function defaultCursorModelFromSettings(settings: AppSettings | undefined): string {
  return settings?.cursor_model ?? "";
}

export function createSubmitStatusForAutonomy(autonomyEnabled: boolean): Status {
  return autonomyEnabled ? DEFAULT_NEW_TASK_STATUS : "on_hold";
}