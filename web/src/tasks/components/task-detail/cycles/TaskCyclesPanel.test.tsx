import "./taskCyclesPanel.testSetup";
import { screen, waitFor, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { afterEach, describe, expect, it, vi } from "vitest";
import { okJSON, renderPanel, reqUrl } from "./taskCyclesPanel.testSetup";

describe("TaskCyclesPanel", () => {
  afterEach(() => {
    vi.restoreAllMocks();
    vi.useRealTimers();
  });

  it("renders a loading skeleton while the cycles list query is pending", async () => {
    // fetch never resolves → query stays pending forever (test-scoped).
    vi.spyOn(globalThis, "fetch").mockImplementation(
      () => new Promise(() => {}),
    );
    const { container } = renderPanel();

    // The skeleton list must be busy-announced for assistive tech.
    const busy = container.querySelector('[aria-busy="true"]');
    expect(busy).not.toBeNull();
    expect(busy?.getAttribute("aria-label")).toMatch(/Loading execution cycles/i);
    // No live ticker yet: we don't know if there's a running cycle.
    expect(screen.queryByTestId("task-cycle-ticker")).not.toBeInTheDocument();
  });

  it("surfaces an error with a retry control when the cycles fetch fails", async () => {
    let callCount = 0;
    const fetchSpy = vi
      .spyOn(globalThis, "fetch")
      .mockImplementation(async () => {
        callCount += 1;
        if (callCount === 1) {
          return new Response(
            JSON.stringify({ error: "boom" }),
            { status: 500, headers: { "Content-Type": "application/json" } },
          );
        }
        return okJSON({
          task_id: "task-1",
          cycles: [],
          limit: 50,
          has_more: false,
        });
      });

    renderPanel();

    const alert = await screen.findByRole("alert");
    expect(alert).toHaveTextContent(/boom/);

    // Retry button refetches and the error gives way to the empty state.
    await userEvent.click(screen.getByRole("button", { name: /Try again/i }));
    await screen.findByText(/No execution cycles yet/i);
    expect(fetchSpy).toHaveBeenCalledTimes(2);
  });

  it("renders the empty state when the task has no cycles", async () => {
    vi.spyOn(globalThis, "fetch").mockResolvedValue(
      okJSON({
        task_id: "task-1",
        cycles: [],
        limit: 50,
        has_more: false,
      }),
    );

    renderPanel();

    await screen.findByText(/No execution cycles yet/i);
    // No ticker, no list, no error — just the empty state.
    expect(screen.queryByTestId("task-cycle-ticker")).not.toBeInTheDocument();
    expect(screen.queryByTestId("task-cycles-list")).not.toBeInTheDocument();
  });

  it("lists historical cycles newest-first and lazy-loads phases on row expansion", async () => {
    // Two terminal cycles, no running one. Row expansion triggers
    // a per-cycle detail fetch that we count below to assert the
    // panel doesn't waste bandwidth on collapsed rows.
    const detailCalls: string[] = [];
    vi.spyOn(globalThis, "fetch").mockImplementation(async (input) => {
      const url = reqUrl(input);
      if (url.endsWith("/tasks/task-1/cycles")) {
        return okJSON({
          task_id: "task-1",
          cycles: [
            {
              id: "cyc-2",
              task_id: "task-1",
              attempt_seq: 2,
              status: "succeeded",
              started_at: "2026-04-18T11:00:00.000Z",
              ended_at: "2026-04-18T11:01:00.000Z",
              triggered_by: "agent",
              meta: {},
            },
            {
              id: "cyc-1",
              task_id: "task-1",
              attempt_seq: 1,
              status: "failed",
              started_at: "2026-04-18T10:00:00.000Z",
              ended_at: "2026-04-18T10:00:45.000Z",
              triggered_by: "user",
              meta: {},
            },
          ],
          limit: 50,
          has_more: false,
        });
      }
      if (url.endsWith("/verdicts")) {
        const m = url.match(/\/cycles\/([^/]+)\/verdicts$/);
        return okJSON({
          task_id: "task-1",
          cycle_id: m?.[1] ?? "",
          criteria_reports: [],
          verify_reports: [],
          command_runs: [],
          commits: [],
        });
      }
      if (url.startsWith("/tasks/task-1/cycles/")) {
        detailCalls.push(url);
        const id = url.replace("/tasks/task-1/cycles/", "");
        return okJSON({
          id,
          task_id: "task-1",
          attempt_seq: id === "cyc-2" ? 2 : 1,
          status: id === "cyc-2" ? "succeeded" : "failed",
          started_at: "2026-04-18T10:00:00.000Z",
          ended_at: "2026-04-18T10:00:45.000Z",
          triggered_by: "agent",
          meta: {},
          phases: [
            {
              id: `${id}-ph-1`,
              cycle_id: id,
              phase: "execute",
              phase_seq: 1,
              status: id === "cyc-2" ? "succeeded" : "failed",
              started_at: "2026-04-18T10:00:11.000Z",
              ended_at: "2026-04-18T10:00:40.000Z",
              details: {},
              summary: "looked at the failure",
            },
          ],
        });
      }
      return new Response("not found", { status: 404 });
    });

    renderPanel();

    // List shows both cycles; the running ticker is absent.
    const list = await screen.findByTestId("task-cycles-list");
    const items = within(list).getAllByRole("listitem");
    expect(items).toHaveLength(2);
    expect(within(items[0]).getByText(/Attempt #2/)).toBeInTheDocument();
    expect(within(items[1]).getByText(/Attempt #1/)).toBeInTheDocument();
    expect(screen.queryByTestId("task-cycle-ticker")).not.toBeInTheDocument();

    // Phase fetch is lazy — collapsed rows must not have hit /cycles/{id}.
    expect(detailCalls).toEqual([]);

    // Expanding the first row triggers exactly one detail fetch.
    await userEvent.click(within(items[0]).getByText(/Attempt #2/));
    await waitFor(() => expect(detailCalls).toEqual(["/tasks/task-1/cycles/cyc-2"]));
    await within(items[0]).findByText(/looked at the failure/);

    // Expanding the second row triggers a second, distinct detail fetch.
    await userEvent.click(within(items[1]).getByText(/Attempt #1/));
    await waitFor(() =>
      expect(detailCalls).toEqual([
        "/tasks/task-1/cycles/cyc-2",
        "/tasks/task-1/cycles/cyc-1",
      ]),
    );
  });

});
