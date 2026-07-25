import { listTasks as listTasksDefault } from "@/api";
import type { Task, TaskListResponse } from "@/types";
import {
  BOARD_ACTIVE_CAP,
  BOARD_MAX_PAGES,
  BOARD_PAGE_SIZE,
} from "./boardConstants";

export type ListTasksFn = typeof listTasksDefault;

export type FetchActiveTasksForBoardOptions = {
  signal?: AbortSignal;
  listTasks?: ListTasksFn;
  activeCap?: number;
  maxPages?: number;
  pageSize?: number;
};

function assertNotAborted(signal?: AbortSignal): void {
  if (signal?.aborted) {
    const err = new Error("Board task walk aborted");
    err.name = "AbortError";
    throw err;
  }
}

/**
 * Walks `GET /tasks` newest-first pages and keeps non-done tasks until
 * caps. Returns a `TaskListResponse` so the result can live under
 * `taskQueryKeys.board()` / `listRoot()` optimistic helpers.
 *
 * `has_more` is true when the walk stopped early (active cap or page
 * cap with more server rows) so the UI can show a truncation hint.
 */
export async function fetchActiveTasksForBoard(
  options: FetchActiveTasksForBoardOptions = {},
): Promise<TaskListResponse> {
  const listTasks = options.listTasks ?? listTasksDefault;
  const activeCap = options.activeCap ?? BOARD_ACTIVE_CAP;
  const maxPages = options.maxPages ?? BOARD_MAX_PAGES;
  const pageSize = options.pageSize ?? BOARD_PAGE_SIZE;
  const { signal } = options;

  const active: Task[] = [];
  let afterId: string | undefined;
  let pages = 0;
  let hitActiveCap = false;
  let serverMayHaveMore = false;

  while (pages < maxPages && active.length < activeCap) {
    assertNotAborted(signal);
    const res = await listTasks(pageSize, 0, {
      signal,
      afterId,
    });
    pages += 1;

    for (const task of res.tasks) {
      if (task.status === "done") continue;
      active.push(task);
      if (active.length >= activeCap) {
        hitActiveCap = true;
        break;
      }
    }

    const pageMayContinue = res.has_more && res.tasks.length > 0;
    if (hitActiveCap) {
      serverMayHaveMore = pageMayContinue || active.length >= activeCap;
      break;
    }
    if (!pageMayContinue) {
      serverMayHaveMore = false;
      break;
    }
    serverMayHaveMore = true;
    afterId = res.tasks[res.tasks.length - 1]!.id;
  }

  const truncated =
    hitActiveCap || (serverMayHaveMore && pages >= maxPages);

  return {
    tasks: active,
    limit: active.length,
    offset: 0,
    has_more: truncated,
  };
}
