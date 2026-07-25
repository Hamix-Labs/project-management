import { describe, expect, it } from "vitest";
import { STATUSES } from "@/types";
import { BOARD_COLUMNS, boardColumnIdForStatus } from "./boardColumns";
import { groupTasksByBoardColumn } from "./groupTasksByBoardColumn";
import { makeTask } from "@/test/taskDefaults";

describe("BOARD_COLUMNS", () => {
  it("assigns every status except done to exactly one column", () => {
    const seen = new Map<string, string>();
    for (const col of BOARD_COLUMNS) {
      for (const status of col.statuses) {
        expect(seen.has(status)).toBe(false);
        seen.set(status, col.id);
      }
    }
    for (const status of STATUSES) {
      if (status === "done") {
        expect(seen.has("done")).toBe(false);
        expect(boardColumnIdForStatus("done")).toBeNull();
        continue;
      }
      expect(seen.get(status)).toBeDefined();
      expect(boardColumnIdForStatus(status)).toBe(seen.get(status));
    }
  });
});

describe("groupTasksByBoardColumn", () => {
  it("buckets by workflow column and drops done", () => {
    const tasks = [
      makeTask({ id: "1", status: "ready" }),
      makeTask({ id: "2", status: "on_hold" }),
      makeTask({ id: "3", status: "running" }),
      makeTask({ id: "4", status: "review" }),
      makeTask({ id: "5", status: "blocked" }),
      makeTask({ id: "6", status: "failed" }),
      makeTask({ id: "7", status: "done" }),
      makeTask({ id: "8", status: "closed" }),
    ];
    const groups = groupTasksByBoardColumn(tasks);
    expect(groups.backlog.map((t) => t.id)).toEqual(["1", "2"]);
    expect(groups.in_progress.map((t) => t.id)).toEqual(["3"]);
    expect(groups.verification.map((t) => t.id)).toEqual(["4"]);
    expect(groups.needs_attention.map((t) => t.id)).toEqual(["5", "6"]);
    expect(groups.closed.map((t) => t.id)).toEqual(["8"]);
  });

  it("returns empty bags for empty input", () => {
    const groups = groupTasksByBoardColumn([]);
    expect(groups.backlog).toEqual([]);
    expect(groups.in_progress).toEqual([]);
    expect(groups.verification).toEqual([]);
    expect(groups.needs_attention).toEqual([]);
    expect(groups.closed).toEqual([]);
  });
});
