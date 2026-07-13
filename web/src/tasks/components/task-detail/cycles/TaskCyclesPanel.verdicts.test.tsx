import "./taskCyclesPanel.testSetup";
import { screen, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { afterEach, describe, expect, it, vi } from "vitest";
import { okJSON, renderPanel, reqUrl } from "./taskCyclesPanel.testSetup";

describe("TaskCyclesPanel verdicts", () => {
  afterEach(() => {
    vi.restoreAllMocks();
    vi.useRealTimers();
  });

  it("renders per-criterion verdict reasoning when a cycle row is expanded", async () => {
    // The verdict block lives below the phase strip and only mounts
    // when the row is expanded — so we mock both the cycles list,
    // the per-cycle detail (phases), and the new verdicts envelope.
    vi.spyOn(globalThis, "fetch").mockImplementation(async (input) => {
      const url = reqUrl(input);
      if (url.endsWith("/tasks/task-1/cycles")) {
        return okJSON({
          task_id: "task-1",
          cycles: [
            {
              id: "cyc-verdicts",
              task_id: "task-1",
              attempt_seq: 1,
              status: "succeeded",
              started_at: "2026-04-18T10:00:00.000Z",
              ended_at: "2026-04-18T10:01:00.000Z",
              triggered_by: "agent",
              meta: {},
            },
          ],
          limit: 50,
          has_more: false,
        });
      }
      if (url === "/tasks/task-1/cycles/cyc-verdicts") {
        return okJSON({
          id: "cyc-verdicts",
          task_id: "task-1",
          attempt_seq: 1,
          status: "succeeded",
          started_at: "2026-04-18T10:00:00.000Z",
          ended_at: "2026-04-18T10:01:00.000Z",
          triggered_by: "agent",
          meta: {},
          phases: [],
        });
      }
      if (url === "/tasks/task-1/cycles/cyc-verdicts/verdicts") {
        return okJSON({
          task_id: "task-1",
          cycle_id: "cyc-verdicts",
          criteria_reports: [
            {
              id: "cr-1",
              cycle_id: "cyc-verdicts",
              attempt_seq: 1,
              criterion_id: "crit-a",
              claimed_done: true,
              evidence: "ran 7 tests, all passed",
              written_at: "2026-04-18T10:00:55.000Z",
            },
          ],
          verify_reports: [
            {
              id: "vr-1",
              cycle_id: "cyc-verdicts",
              attempt_seq: 1,
              criterion_id: "crit-a",
              verified: true,
              verifier_kind: "verify_agent",
              reasoning: "verifier ran the test suite independently",
              written_at: "2026-04-18T10:00:58.000Z",
            },
          ],
          command_runs: [],
          commits: [],
        });
      }
      return new Response("not found", { status: 404 });
    });

    renderPanel();

    const list = await screen.findByTestId("task-cycles-list");
    const row = within(list).getAllByRole("listitem")[0];
    await userEvent.click(within(row).getByText(/Attempt #1/));

    const verdicts = await within(row).findByTestId("task-cycle-verdicts");
    expect(verdicts).toHaveTextContent(/Verdicts/);
    expect(verdicts).toHaveTextContent(/Verified/);
    expect(verdicts).toHaveTextContent(/Verify agent/);
    expect(verdicts).toHaveTextContent(/verifier ran the test suite independently/);
    expect(verdicts).toHaveTextContent(/crit-a/);
  });

  it("renders an empty-state note when a cycle has no verdicts captured", async () => {
    vi.spyOn(globalThis, "fetch").mockImplementation(async (input) => {
      const url = reqUrl(input);
      if (url.endsWith("/tasks/task-1/cycles")) {
        return okJSON({
          task_id: "task-1",
          cycles: [
            {
              id: "cyc-empty",
              task_id: "task-1",
              attempt_seq: 1,
              status: "succeeded",
              started_at: "2026-04-18T10:00:00.000Z",
              ended_at: "2026-04-18T10:01:00.000Z",
              triggered_by: "agent",
              meta: {},
            },
          ],
          limit: 50,
          has_more: false,
        });
      }
      if (url === "/tasks/task-1/cycles/cyc-empty") {
        return okJSON({
          id: "cyc-empty",
          task_id: "task-1",
          attempt_seq: 1,
          status: "succeeded",
          started_at: "2026-04-18T10:00:00.000Z",
          ended_at: "2026-04-18T10:01:00.000Z",
          triggered_by: "agent",
          meta: {},
          phases: [],
        });
      }
      if (url === "/tasks/task-1/cycles/cyc-empty/verdicts") {
        return okJSON({
          task_id: "task-1",
          cycle_id: "cyc-empty",
          criteria_reports: [],
          verify_reports: [],
          command_runs: [],
          commits: [],
        });
      }
      return new Response("not found", { status: 404 });
    });

    renderPanel();
    const list = await screen.findByTestId("task-cycles-list");
    const row = within(list).getAllByRole("listitem")[0];
    await userEvent.click(within(row).getByText(/Attempt #1/));
    expect(
      await within(row).findByTestId("task-cycle-verdicts-empty"),
    ).toHaveTextContent(/No verdicts captured/);
  });

  it("shows retry lineage on cycle rows when parent_cycle_id and retry_mode are set", async () => {
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
    const matches = await screen.findAllByText(/resumed from attempt 2/i);
    expect(matches.length).toBeGreaterThanOrEqual(1);
  });
});
