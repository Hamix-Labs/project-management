import { taskQueryKeys } from "@/lib/taskQueryKeys";
import type { QueryInvalidationKey, TaskInvalidationScope } from "./types";

export function decideTaskInvalidationKeys(
  input: TaskInvalidationScope,
): readonly QueryInvalidationKey[] {
  switch (input.scope) {
    case "listStats":
      return [taskQueryKeys.listRoot(), taskQueryKeys.stats()];
    case "detail":
      return [taskQueryKeys.detail(input.taskId)];
    case "checklist":
      return [
        taskQueryKeys.checklist(input.taskId),
        taskQueryKeys.detail(input.taskId),
      ];
    case "events":
      return [taskQueryKeys.eventsRoot(input.taskId)];
    case "cycles":
      return [taskQueryKeys.cycles(input.taskId)];
    case "drafts":
      return [taskQueryKeys.drafts()];
    case "templates":
      return [taskQueryKeys.templates()];
  }
}
