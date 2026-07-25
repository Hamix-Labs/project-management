import { describe, expect, it, vi } from "vitest";
import type { Task, TaskListResponse } from "@/types";
import { makeTask } from "@/test/taskDefaults";
import { fetchActiveTasksForBoard } from "./fetchActiveTasksForBoard";

function page(
  tasks: Task[],
  has_more: boolean,
): TaskListResponse {
  return { tasks, limit: tasks.length, offset: 0, has_more };
}

describe("fetchActiveTasksForBoard", () => {
  it("returns empty response when the first page is empty", async () => {
    const listTasks = vi.fn().mockResolvedValue(page([], false));
    const out = await fetchActiveTasksForBoard({ listTasks });
    expect(out).toEqual({
      tasks: [],
      limit: 0,
      offset: 0,
      has_more: false,
    });
    expect(listTasks).toHaveBeenCalledTimes(1);
  });

  it("walks pages and keeps only non-done, non-closed tasks", async () => {
    const listTasks = vi
      .fn()
      .mockResolvedValueOnce(
        page(
          [
            makeTask({ id: "d1", status: "done" }),
            makeTask({ id: "c1", status: "closed" }),
            makeTask({ id: "a1", status: "ready" }),
          ],
          true,
        ),
      )
      .mockResolvedValueOnce(
        page(
          [
            makeTask({ id: "a2", status: "running" }),
            makeTask({ id: "d2", status: "done" }),
          ],
          false,
        ),
      );

    const out = await fetchActiveTasksForBoard({
      listTasks,
      pageSize: 3,
    });
    expect(out.tasks.map((t) => t.id)).toEqual(["a1", "a2"]);
    expect(out.has_more).toBe(false);
    expect(listTasks).toHaveBeenCalledTimes(2);
    expect(listTasks.mock.calls[1]![2]).toMatchObject({ afterId: "a1" });
  });

  it("uses last task id on page as after_id even when it is done", async () => {
    const listTasks = vi
      .fn()
      .mockResolvedValueOnce(
        page(
          [
            makeTask({ id: "a1", status: "ready" }),
            makeTask({ id: "d1", status: "done" }),
          ],
          true,
        ),
      )
      .mockResolvedValueOnce(page([], false));

    await fetchActiveTasksForBoard({ listTasks, pageSize: 2 });
    expect(listTasks.mock.calls[1]![2].afterId).toBe("d1");
  });

  it("stops at active cap and sets has_more", async () => {
    const listTasks = vi.fn().mockResolvedValue(
      page(
        [
          makeTask({ id: "1", status: "ready" }),
          makeTask({ id: "2", status: "ready" }),
          makeTask({ id: "3", status: "ready" }),
        ],
        true,
      ),
    );
    const out = await fetchActiveTasksForBoard({
      listTasks,
      activeCap: 2,
      pageSize: 3,
    });
    expect(out.tasks.map((t) => t.id)).toEqual(["1", "2"]);
    expect(out.has_more).toBe(true);
    expect(listTasks).toHaveBeenCalledTimes(1);
  });

  it("stops at max pages with has_more when server still has rows", async () => {
    const listTasks = vi
      .fn()
      .mockResolvedValueOnce(
        page([makeTask({ id: "1", status: "ready" })], true),
      )
      .mockResolvedValueOnce(
        page([makeTask({ id: "2", status: "ready" })], true),
      );

    const out = await fetchActiveTasksForBoard({
      listTasks,
      maxPages: 2,
      pageSize: 1,
    });
    expect(out.tasks.map((t) => t.id)).toEqual(["1", "2"]);
    expect(out.has_more).toBe(true);
    expect(listTasks).toHaveBeenCalledTimes(2);
  });

  it("does not call further pages after abort", async () => {
    const ac = new AbortController();
    const listTasks = vi.fn().mockImplementation(async () => {
      ac.abort();
      return page([makeTask({ id: "1", status: "ready" })], true);
    });
    await expect(
      fetchActiveTasksForBoard({
        listTasks,
        signal: ac.signal,
        pageSize: 1,
        maxPages: 5,
      }),
    ).rejects.toMatchObject({ name: "AbortError" });
    expect(listTasks).toHaveBeenCalledTimes(1);
  });
});
