import type { QueryClient, UseMutationOptions } from "@tanstack/react-query";
import { patchTaskEventUserResponse } from "@/api";
import type { TaskEventDetail } from "@/types";
import { taskQueryKeys } from "../task-query";
import { beginGuardedTaskWrite, endGuardedTaskWrite } from "./guardedTaskWrite";
import { invalidateTaskCache } from "./invalidateTaskCache";

export function buildPatchTaskEventUserResponseMutationOptions(deps: {
  taskId: string;
  eventSeq: number;
  queryClient: QueryClient;
  optimisticMutationsEnabled: boolean;
  onDraftCleared?: () => void;
}): UseMutationOptions<TaskEventDetail, unknown, string, { guarded: boolean }> {
  const { taskId, eventSeq, queryClient, optimisticMutationsEnabled, onDraftCleared } = deps;
  const eventDetailKey = taskQueryKeys.eventDetail(taskId, eventSeq);

  return {
    mutationFn: (text: string) => patchTaskEventUserResponse(taskId, eventSeq, text),
    onMutate: async () => {
      const guard = beginGuardedTaskWrite({
        taskId,
        optimisticEnabled: optimisticMutationsEnabled,
        rumKind: "task_patch",
      });
      return { guarded: guard.guarded };
    },
    onSuccess: (updated) => {
      queryClient.setQueryData(eventDetailKey, updated);
      onDraftCleared?.();
      invalidateTaskCache(queryClient, { scope: "events", taskId });
    },
    onSettled: (_data, _err, _vars, context) => {
      if (context?.guarded) endGuardedTaskWrite(taskId);
    },
  };
}
