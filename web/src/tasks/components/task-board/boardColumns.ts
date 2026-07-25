import type { Status } from "@/types";
import type { StatusTone } from "@/lib/taskStatusDisplay";

export type BoardColumnId =
  | "backlog"
  | "in_progress"
  | "verification"
  | "needs_attention"
  | "closed";

export type BoardColumnDef = {
  id: BoardColumnId;
  label: string;
  /**
   * Disjoint status sets; never includes `done` (filtered out of the
   * board walk). `closed` has its own column so operator exits stay
   * findable without leaving the board.
   */
  statuses: readonly Status[];
  tone: StatusTone;
};

/**
 * Workflow columns for the read-only board. Adding a column later is a
 * config change here plus any CSS that keys off `id`.
 */
export const BOARD_COLUMNS: readonly BoardColumnDef[] = [
  {
    id: "backlog",
    label: "Backlog",
    statuses: ["ready", "on_hold"],
    tone: "neutral",
  },
  {
    id: "in_progress",
    label: "In Progress",
    statuses: ["running"],
    tone: "info",
  },
  {
    id: "verification",
    label: "Verification",
    statuses: ["review"],
    tone: "warning",
  },
  {
    id: "needs_attention",
    label: "Needs Attention",
    statuses: ["blocked", "failed"],
    tone: "danger",
  },
  {
    id: "closed",
    label: "Closed",
    statuses: ["closed"],
    tone: "neutral",
  },
] as const;

/** Map each board-eligible status to its column id. */
export function boardColumnIdForStatus(
  status: Status,
): BoardColumnId | null {
  for (const col of BOARD_COLUMNS) {
    if (col.statuses.includes(status)) return col.id;
  }
  return null;
}
