import "./taskCyclesPanel.testSetup";
import { screen, within } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";
import { okJSON, renderPanel, reqUrl } from "./taskCyclesPanel.testSetup";

describe("TaskCyclesPanel lineage", () => {
  afterEach(() => {
    vi.restoreAllMocks();
    vi.useRealTimers();
  });

  it("shows retry lineage on the live ticker, not on history rows", async () => {
    vi.spyOn(globalThis, "fetch").mockImplementation(async (input) => {
      const url = reqUrl(input);
      if (url.endsWith("/tasks/task-1/cycles")) {
        return okJSON({
          task_id: "task-1",
          cycles: [
            {
              id: "cyc-3",
              task_id: "task-1",
              attempt_seq: 3,
              parent_cycle_id: "cyc-2",
              status: "running",
              started_at: "2026-04-18T12:00:00.000Z",
              triggered_by: "agent",
              meta: { retry_mode: "resume" },
            },
            {
              id: "cyc-2",
              task_id: "task-1",
              attempt_seq: 2,
              status: "failed",
              started_at: "2026-04-18T11:00:00.000Z",
              ended_at: "2026-04-18T11:01:00.000Z",
              triggered_by: "agent",
              meta: {},
            },
          ],
          limit: 50,
          has_more: false,
        });
      }
      if (url.endsWith("/tasks/task-1/cycles/cyc-3")) {
        return okJSON({
          id: "cyc-3",
          task_id: "task-1",
          attempt_seq: 3,
          parent_cycle_id: "cyc-2",
          status: "running",
          started_at: "2026-04-18T12:00:00.000Z",
          triggered_by: "agent",
          meta: { retry_mode: "resume" },
          phases: [
            {
              id: "cyc-3-ph-1",
              cycle_id: "cyc-3",
              phase: "execute",
              phase_seq: 1,
              status: "running",
              started_at: "2026-04-18T12:00:01.000Z",
              details: {},
              summary: "",
            },
          ],
        });
      }
      return new Response("not found", { status: 404 });
    });

    renderPanel();

    const ticker = await screen.findByTestId("task-cycle-ticker");
    expect(within(ticker).getByText(/resumed from attempt 2/i)).toBeInTheDocument();

    const list = await screen.findByTestId("task-cycles-list");
    expect(within(list).queryByText(/resumed from attempt/i)).not.toBeInTheDocument();
  });
});
