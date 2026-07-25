import type { Status } from "@/types";
import {
  BOARD_COLUMNS,
  boardColumnIdForStatus,
  type BoardColumnDef,
  type BoardColumnId,
} from "./boardColumns";

export type BoardColumnGroups<T extends { status: Status }> = Record<
  BoardColumnId,
  T[]
>;

/** Empty column bags keyed by every configured column id. */
export function emptyBoardColumnGroups<
  T extends { status: Status },
>(): BoardColumnGroups<T> {
  const out = {} as BoardColumnGroups<T>;
  for (const col of BOARD_COLUMNS) {
    out[col.id] = [];
  }
  return out;
}

/**
 * Groups tasks into board columns. Drops `done` and any status that is
 * not listed on a column (defense in depth if cache briefly holds done).
 */
export function groupTasksByBoardColumn<T extends { status: Status }>(
  tasks: readonly T[],
  columns: readonly BoardColumnDef[] = BOARD_COLUMNS,
): BoardColumnGroups<T> {
  const groups = emptyBoardColumnGroups<T>();
  const allowed = new Set(columns.map((c) => c.id));
  for (const task of tasks) {
    const id = boardColumnIdForStatus(task.status);
    if (!id || !allowed.has(id)) continue;
    groups[id].push(task);
  }
  return groups;
}
