import { READ_LIMITS } from "@/lib/readLimits";

/** Server page size for `GET /tasks` from the home list (bootstrap aggregate too). */
export const TASK_LIST_PAGE_SIZE = READ_LIMITS.bootstrapListLimit;

/** Server page size for `GET /tasks/{id}/events` on the task detail timeline. */
export const TASK_EVENTS_PAGE_SIZE = READ_LIMITS.bootstrapListLimit;
