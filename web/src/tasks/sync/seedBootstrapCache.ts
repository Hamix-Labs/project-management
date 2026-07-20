import type { QueryClient } from "@tanstack/react-query";
import type { Bootstrap } from "@/api";
import { projectQueryKeys } from "@/lib/projectQueryKeys";
import { settingsQueryKeys } from "@/lib/settingsQueryKeys";
import { TASK_LIST_PAGE_SIZE } from "@/tasks/task-paging";
import { taskQueryKeys } from "@/tasks/task-query";

export function seedBootstrapCache(queryClient: QueryClient, payload: Bootstrap): void {
  queryClient.setQueryData(settingsQueryKeys.app(), payload.settings);
  queryClient.setQueryData(
    taskQueryKeys.list({ limit: TASK_LIST_PAGE_SIZE, offset: 0 }),
    payload.tasks,
  );
  queryClient.setQueryData(taskQueryKeys.stats(), payload.stats);
  queryClient.setQueryData(projectQueryKeys.list(false, 100), payload.projects);
  queryClient.setQueryData(taskQueryKeys.drafts(), payload.drafts);
}
