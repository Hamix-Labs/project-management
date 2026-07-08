import type { Status } from "@/types";

const TERMINAL_TASK_STATUSES: ReadonlySet<Status> = new Set(["done", "failed"]);

export function isTerminalTaskStatus(status: Status): boolean {
  return TERMINAL_TASK_STATUSES.has(status);
}
