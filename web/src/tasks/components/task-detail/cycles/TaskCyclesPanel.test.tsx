import "./taskCyclesPanel.testSetup";
import { screen, within } from "@testing-library/react";
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
    expect(busy?.getAttribute("aria-label")).toMatch(/Loading attempts/i);
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
    await screen.findByText(/No attempts yet/i);
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

    await screen.findByText(/No attempts yet/i);
    // No ticker, no list, no error — just the empty state.
    expect(screen.queryByTestId("task-cycle-ticker")).not.toBeInTheDocument();
    expect(screen.queryByTestId("task-cycles-list")).not.toBeInTheDocument();
  });

  it("lists historical cycles newest-first with a link to run details", async () => {
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
      return new Response("not found", { status: 404 });
    });

    renderPanel();

    const list = await screen.findByTestId("task-cycles-list");
    const items = within(list).getAllByRole("listitem");
    expect(items).toHaveLength(2);
    expect(within(items[0]).getByText(/Attempt #2/)).toBeInTheDocument();
    expect(within(items[1]).getByText(/Attempt #1/)).toBeInTheDocument();
    expect(screen.queryByTestId("task-cycle-ticker")).not.toBeInTheDocument();
    expect(
      within(items[0]).getByRole("link", { name: /Details/i }),
    ).toHaveAttribute("href", "/tasks/task-1/cycles/cyc-2");
  });

  it("defaults open and collapses the cycles body from the section header", async () => {
    const user = userEvent.setup();
    vi.spyOn(globalThis, "fetch").mockResolvedValue(
      okJSON({
        task_id: "task-1",
        cycles: [
          {
            id: "cyc-1",
            task_id: "task-1",
            attempt_seq: 1,
            status: "succeeded",
            started_at: "2026-04-18T10:00:00.000Z",
            ended_at: "2026-04-18T10:00:45.000Z",
            triggered_by: "user",
            meta: {},
          },
        ],
        limit: 50,
        has_more: false,
      }),
    );

    renderPanel();

    const list = await screen.findByTestId("task-cycles-list");
    expect(list).toBeVisible();
    const details = document.querySelector(
      ".task-cycles-panel .task-detail-collapsible",
    );
    expect(details).toHaveAttribute("open");
    expect(screen.getByText("1")).toBeInTheDocument();

    await user.click(screen.getByRole("heading", { name: /^attempts$/i }));
    expect(details).not.toHaveAttribute("open");
    expect(list).not.toBeVisible();
  });

});
