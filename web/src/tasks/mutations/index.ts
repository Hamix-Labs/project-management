export { mergePatchIntoTask, patchTaskInList } from "./optimisticTaskDetail";
export type { TaskDetailPatchFields } from "./optimisticTaskDetail";
export {
  applyCreatedTaskToCache,
  applyCreatedTasksToCache,
  insertTaskInList,
  patchTaskPickupInList,
  patchTaskPickupInListCaches,
  removeTaskFromList,
  removeTaskFromListCaches,
} from "./optimisticTaskList";
export { invalidateTaskListAndStats } from "./invalidateTaskListCoherence";
export {
  beginGuardedTaskWrite,
  cancelQueriesForKeys,
  endGuardedTaskWrite,
  recordOptimisticApplied,
} from "./guardedTaskWrite";
export type { GuardedWriteContext } from "./guardedTaskWrite";
export { runGuardedTaskMutation } from "./runGuardedMutation";
export type {
  RunGuardedTaskMutationOptions,
  RunGuardedTaskMutationResult,
} from "./runGuardedMutation";
