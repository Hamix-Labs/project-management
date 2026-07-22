import "./taskCyclesPanel.testSetup";
import { act, screen, waitFor, within } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";
import { pushAgentRunProgress } from "../../../hooks/useAgentRunProgress";
import { okJSON, renderPanel, reqUrl } from "./taskCyclesPanel.testSetup";

describe("TaskCyclesPanel live ticker", () => {
  afterEach(() => {
    vi.restoreAllMocks();
    vi.useRealTimers();
  });

  it("renders the live ticker for the running cycle with the currently running phase", async () => {
    vi.useFakeTimers({ shouldAdvanceTime: true });
    vi.setSystemTime(new Date("2026-04-18T11:00:30.000Z"));

    vi.spyOn(globalThis, "fetch").mockImplementation(async (input) => {
      const url = reqUrl(input);
      if (url.endsWith("/tasks/task-1/cycles")) {
        return okJSON({
          task_id: "task-1",
          cycles: [
            {
              id: "cyc-live",
              task_id: "task-1",
              attempt_seq: 3,
              status: "running",
              started_at: "2026-04-18T11:00:00.000Z",
              triggered_by: "agent",
              meta: {},
            },
          ],
          limit: 50,
          has_more: false,
        });
      }
      if (url === "/tasks/task-1/cycles/cyc-live") {
        return okJSON({
          id: "cyc-live",
          task_id: "task-1",
          attempt_seq: 3,
          status: "running",
          started_at: "2026-04-18T11:00:00.000Z",
          triggered_by: "agent",
          meta: {},
          phases: [
            {
              id: "p-e",
              cycle_id: "cyc-live",
              phase: "execute",
              phase_seq: 1,
              status: "running",
              started_at: "2026-04-18T11:00:11.000Z",
              details: {},
            },
          ],
        });
      }
      return new Response("not found", { status: 404 });
    });

    renderPanel();
    await act(async () => {});

    const ticker = await screen.findByTestId("task-cycle-ticker");
    expect(ticker).not.toHaveAttribute("aria-live");
    expect(within(ticker).getByTestId("task-cycle-ticker-status")).toHaveTextContent(
      /Running/,
    );
    expect(within(ticker).getByText(/Attempt #3/)).toBeInTheDocument();
    // Started 30 s ago at fake-now (11:00:30 - 11:00:00).
    expect(
      within(ticker).getByTestId("task-cycle-ticker-elapsed"),
    ).toHaveTextContent(/Started 30\.0 s ago/);

    // The phase line resolves to the running execute phase.
    const phaseLine = await within(ticker).findByTestId(
      "task-cycle-ticker-phase",
    );
    expect(phaseLine).toHaveTextContent(/Execute/);
    // Phase started 19 s ago (11:00:30 - 11:00:11 = 19 s).
    expect(phaseLine).toHaveTextContent(/19\.0 s/);
    expect(within(phaseLine).getByText(/19\.0 s/)).toHaveAttribute(
      "aria-hidden",
      "true",
    );

    act(() => {
      vi.advanceTimersByTime(5000);
    });
    expect(
      within(ticker).getByTestId("task-cycle-ticker-elapsed"),
    ).toHaveTextContent(/Started 35\.0 s ago/);
    expect(phaseLine).toHaveTextContent(/24\.0 s/);

    // The running cycle ALSO appears in the history list, with a
    // small "↑ live" hint pointing the user up to the ticker.
    const list = screen.getByTestId("task-cycles-list");
    expect(within(list).getByLabelText(/shown in the live ticker above/i)).toBeInTheDocument();
  });

  it("renders a runner/model chip on the live ticker only", async () => {
    vi.spyOn(globalThis, "fetch").mockImplementation(async (input) => {
      const url = reqUrl(input);
      if (url.endsWith("/tasks/task-1/cycles")) {
        return okJSON({
          task_id: "task-1",
          cycles: [
            {
              id: "cyc-live",
              task_id: "task-1",
              attempt_seq: 2,
              status: "running",
              started_at: "2026-04-18T11:00:00.000Z",
              triggered_by: "agent",
              meta: {},
              cycle_meta: {
                runner: "cursor",
                runner_version: "v1.2.3",
                cursor_model: "opus-4",
                cursor_model_effective: "opus-4",
                prompt_hash: "abc",
              },
            },
            {
              id: "cyc-hist",
              task_id: "task-1",
              attempt_seq: 1,
              status: "succeeded",
              started_at: "2026-04-18T10:00:00.000Z",
              ended_at: "2026-04-18T10:01:00.000Z",
              triggered_by: "user",
              meta: {},
              cycle_meta: {
                runner: "cursor",
                runner_version: "v1.2.3",
                cursor_model: "",
                cursor_model_effective: "sonnet-4.5",
                prompt_hash: "def",
              },
            },
          ],
          limit: 50,
          has_more: false,
        });
      }
      if (url === "/tasks/task-1/cycles/cyc-live") {
        return okJSON({
          id: "cyc-live",
          task_id: "task-1",
          attempt_seq: 2,
          status: "running",
          started_at: "2026-04-18T11:00:00.000Z",
          triggered_by: "agent",
          meta: {},
          cycle_meta: {
            runner: "cursor",
            runner_version: "v1.2.3",
            cursor_model: "opus-4",
            cursor_model_effective: "opus-4",
            prompt_hash: "abc",
          },
          phases: [],
        });
      }
      return new Response("not found", { status: 404 });
    });

    renderPanel();

    const ticker = await screen.findByTestId("task-cycle-ticker");
    expect(within(ticker).getByTestId("task-cycle-ticker-runner")).toHaveTextContent(
      "Cursor CLI · opus-4",
    );

    const list = screen.getByTestId("task-cycles-list");
    expect(within(list).queryByTestId("task-cycle-row-runner")).not.toBeInTheDocument();
  });

  it("falls back to a 'between phases' line when the running cycle has no in-flight phase", async () => {
    // Cycle is running but every phase has already terminated —
    // the worker is between StartCycle/StartPhase frames. The
    // ticker must still resolve gracefully and show the most
    // recent (highest phase_seq) phase rather than going blank.
    const fakeNow = Date.parse("2026-04-18T11:01:00.000Z");
    vi.spyOn(Date, "now").mockReturnValue(fakeNow);

    vi.spyOn(globalThis, "fetch").mockImplementation(async (input) => {
      const url = reqUrl(input);
      if (url.endsWith("/tasks/task-1/cycles")) {
        return okJSON({
          task_id: "task-1",
          cycles: [
            {
              id: "cyc-tween",
              task_id: "task-1",
              attempt_seq: 4,
              status: "running",
              started_at: "2026-04-18T11:00:00.000Z",
              triggered_by: "agent",
              meta: {},
            },
          ],
          limit: 50,
          has_more: false,
        });
      }
      if (url === "/tasks/task-1/cycles/cyc-tween") {
        return okJSON({
          id: "cyc-tween",
          task_id: "task-1",
          attempt_seq: 4,
          status: "running",
          started_at: "2026-04-18T11:00:00.000Z",
          triggered_by: "agent",
          meta: {},
          phases: [
            {
              id: "p-e",
              cycle_id: "cyc-tween",
              phase: "execute",
              phase_seq: 1,
              status: "succeeded",
              started_at: "2026-04-18T11:00:11.000Z",
              ended_at: "2026-04-18T11:00:40.000Z",
              details: {},
            },
          ],
        });
      }
      return new Response("not found", { status: 404 });
    });

    renderPanel();

    const phaseLine = await screen.findByTestId("task-cycle-ticker-phase");
    await waitFor(() => {
      expect(phaseLine).toHaveTextContent(/Between phases/);
    });
    expect(phaseLine).toHaveTextContent(/Execute/);
    expect(phaseLine).toHaveTextContent(/succeeded/);
  });

  it("renders bounded live progress under the running phase", async () => {
    vi.spyOn(globalThis, "fetch").mockImplementation(async (input) => {
      const url = reqUrl(input);
      if (url.endsWith("/tasks/task-1/cycles")) {
        return okJSON({
          task_id: "task-1",
          cycles: [
            {
              id: "cyc-live-progress",
              task_id: "task-1",
              attempt_seq: 1,
              status: "running",
              started_at: "2026-04-18T11:00:00.000Z",
              triggered_by: "agent",
              meta: {},
            },
          ],
          limit: 50,
          has_more: false,
        });
      }
      if (url === "/tasks/task-1/cycles/cyc-live-progress") {
        return okJSON({
          id: "cyc-live-progress",
          task_id: "task-1",
          attempt_seq: 1,
          status: "running",
          started_at: "2026-04-18T11:00:00.000Z",
          triggered_by: "agent",
          meta: {},
          phases: [
            {
              id: "p-e",
              cycle_id: "cyc-live-progress",
              phase: "execute",
              phase_seq: 2,
              status: "running",
              started_at: "2026-04-18T11:00:05.000Z",
              details: {},
            },
          ],
        });
      }
      return new Response("not found", { status: 404 });
    });

    renderPanel();

    const ticker = await screen.findByTestId("task-cycle-ticker");
    expect(await within(ticker).findByTestId("task-cycle-progress-empty")).toHaveTextContent(
      /Waiting for the next agent update/,
    );

    act(() => {
      pushAgentRunProgress({
        taskId: "task-1",
        cycleId: "cyc-live-progress",
        phaseSeq: 2,
        progress: {
          kind: "tool_call",
          subtype: "started",
          tool: "ReadFile",
          message: "Reading README.md",
        },
      });
    });

    const progressList = await within(ticker).findByTestId("task-cycle-progress-list");
    expect(progressList).toHaveAttribute("aria-label", "Recent agent progress");
    expect(progressList).toHaveTextContent(/Tool call/);
    expect(progressList).toHaveTextContent(/Reading README\.md/);
    expect(
      progressList.querySelector(".task-cycle-progress-item--latest"),
    ).not.toBeNull();
  });

});
