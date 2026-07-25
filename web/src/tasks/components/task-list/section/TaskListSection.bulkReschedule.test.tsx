import "./taskListSection.testMocks";
import { closeTask, patchTask } from "@/api";
import { screen, waitFor, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";
import { TASK_TEST_DEFAULTS } from "@/test/taskDefaults";
import { TaskListSection } from "./TaskListSection";
import {
  listPagerDefaults,
  makeRow,
  renderWithRouter,
} from "./taskListSection.testSetup";

const mockPatchTask = vi.mocked(patchTask);
const mockCloseTask = vi.mocked(closeTask);

describe("TaskListSection bulk reschedule", () => {
  describe("bulk reschedule", () => {
    it("disables Reschedule when any selected task is done", async () => {
      const user = userEvent.setup();
      const tasks = [
        makeRow("a", "Alpha", { status: "done" }),
        makeRow("b", "Beta", { status: "ready" }),
      ];
      renderWithRouter(
        <TaskListSection
          tasks={tasks}
          loading={false}
          refreshing={false}
          saving={false}
          {...listPagerDefaults}
          rootTasksOnPage={2}
          onEdit={vi.fn()}
          onRequestClose={vi.fn()}
        />,
      );
      await user.click(screen.getByTestId("task-list-select-row-a"));
      const reschedule = screen.getByTestId("task-list-bulk-bar-reschedule");
      expect(reschedule).toBeDisabled();
      expect(reschedule).toHaveAttribute(
        "title",
        "Completed tasks cannot be rescheduled from the list.",
      );
    });

    it("shows the bulk action bar only after at least one row is selected", async () => {
      const user = userEvent.setup();
      const tasks = [
        makeRow("a", "Alpha"),
        makeRow("b", "Beta"),
      ];
      renderWithRouter(
        <TaskListSection
          tasks={tasks}
          loading={false}
          refreshing={false}
          saving={false}
          {...listPagerDefaults}
          rootTasksOnPage={2}
          onEdit={vi.fn()}
          onRequestClose={vi.fn()}
        />,
      );
      expect(
        screen.queryByTestId("task-list-bulk-bar"),
      ).not.toBeInTheDocument();
      await user.click(screen.getByTestId("task-list-select-row-a"));
      const bar = screen.getByTestId("task-list-bulk-bar");
      expect(bar).toBeInTheDocument();
      expect(
        within(bar).getByTestId("task-list-bulk-bar-summary"),
      ).toHaveTextContent("1 task selected");
    });

    it("select-all-visible toggles every row in the filtered set", async () => {
      const user = userEvent.setup();
      const tasks = [
        makeRow("a", "Alpha"),
        makeRow("b", "Beta"),
        makeRow("c", "Gamma"),
      ];
      renderWithRouter(
        <TaskListSection
          tasks={tasks}
          loading={false}
          refreshing={false}
          saving={false}
          {...listPagerDefaults}
          rootTasksOnPage={3}
          onEdit={vi.fn()}
          onRequestClose={vi.fn()}
        />,
      );
      await user.click(screen.getByTestId("task-list-select-all"));
      const bar = screen.getByTestId("task-list-bulk-bar");
      expect(
        within(bar).getByTestId("task-list-bulk-bar-summary"),
      ).toHaveTextContent("3 tasks selected");
      await user.click(screen.getByTestId("task-list-select-all"));
      expect(
        screen.queryByTestId("task-list-bulk-bar"),
      ).not.toBeInTheDocument();
    });

    it("Reschedule bulk action PATCHes every selected task with the same ISO", async () => {
      const user = userEvent.setup();
      mockPatchTask.mockImplementation(async (id: string, body: { pickup_not_before?: string | null }) => ({
        id,
        title: id,
        initial_prompt: "",
        status: "ready",
        priority: "medium",
        pickup_not_before: body.pickup_not_before ?? undefined,
        ...TASK_TEST_DEFAULTS,
      }));
      const tasks = [
        makeRow("a", "Alpha"),
        makeRow("b", "Beta"),
        makeRow("c", "Gamma"),
      ];
      renderWithRouter(
        <TaskListSection
          tasks={tasks}
          loading={false}
          refreshing={false}
          saving={false}
          {...listPagerDefaults}
          rootTasksOnPage={3}
          onEdit={vi.fn()}
          onRequestClose={vi.fn()}
        />,
      );
      await user.click(screen.getByTestId("task-list-select-row-a"));
      await user.click(screen.getByTestId("task-list-select-row-b"));
      await user.click(screen.getByTestId("task-list-select-row-c"));
      await user.click(screen.getByTestId("task-list-bulk-bar-reschedule"));
      const fixedNow = Date.UTC(2026, 3, 19, 12, 0, 0);
      const dateNow = vi.spyOn(Date, "now").mockReturnValue(fixedNow);
      await user.click(screen.getByTestId("schedule-picker-quick-trigger"));
      await user.click(screen.getByTestId("schedule-picker-quick-hour-1"));
      dateNow.mockRestore();
      await user.click(screen.getByTestId("task-bulk-reschedule-submit"));
      await waitFor(() => {
        expect(mockPatchTask).toHaveBeenCalledTimes(3);
      });
      const calls = mockPatchTask.mock.calls;
      const ids = calls.map((c) => c[0]).sort();
      expect(ids).toEqual(["a", "b", "c"]);
      const bodies = calls.map((c) => c[1]);
      const isoSet = new Set(
        bodies.map((b) => (b as { pickup_not_before: string }).pickup_not_before),
      );
      expect(isoSet.size).toBe(1);
      const iso = [...isoSet][0];
      expect(typeof iso).toBe("string");
      expect(new Date(iso).toISOString()).toBe(iso);
      await waitFor(() => {
        expect(
          screen.queryByTestId("task-list-bulk-bar"),
        ).not.toBeInTheDocument();
      });
    });

    it("Clear schedule bulk action sends pickup_not_before=null to scheduled rows only", async () => {
      const user = userEvent.setup();
      mockPatchTask.mockResolvedValue({
        id: "a",
        title: "Alpha",
        initial_prompt: "",
        status: "ready",
        priority: "medium",
        ...TASK_TEST_DEFAULTS,
      });
      const futureIso = new Date(Date.now() + 60 * 60 * 1000).toISOString();
      const tasks = [
        makeRow("a", "Alpha", { pickup_not_before: futureIso }),
        makeRow("b", "Beta", { pickup_not_before: futureIso }),
        makeRow("c", "Gamma"),
      ];
      renderWithRouter(
        <TaskListSection
          tasks={tasks}
          loading={false}
          refreshing={false}
          saving={false}
          {...listPagerDefaults}
          rootTasksOnPage={3}
          onEdit={vi.fn()}
          onRequestClose={vi.fn()}
        />,
      );
      await user.click(screen.getByTestId("task-list-select-all"));
      await user.click(screen.getByTestId("task-list-bulk-bar-clear"));
      await waitFor(() => {
        expect(mockPatchTask).toHaveBeenCalledTimes(2);
      });
      const ids = mockPatchTask.mock.calls.map((c) => c[0]).sort();
      expect(ids).toEqual(["a", "b"]);
      for (const c of mockPatchTask.mock.calls) {
        expect((c[1] as { pickup_not_before: unknown }).pickup_not_before).toBe(null);
      }
    });

    it("Clear schedule button is disabled when no selected row has a schedule", async () => {
      const user = userEvent.setup();
      const tasks = [makeRow("a", "Alpha"), makeRow("b", "Beta")];
      renderWithRouter(
        <TaskListSection
          tasks={tasks}
          loading={false}
          refreshing={false}
          saving={false}
          {...listPagerDefaults}
          rootTasksOnPage={2}
          onEdit={vi.fn()}
          onRequestClose={vi.fn()}
        />,
      );
      await user.click(screen.getByTestId("task-list-select-all"));
      const clearBtn = screen.getByTestId("task-list-bulk-bar-clear");
      expect(clearBtn).toBeDisabled();
    });

    it("partial failure surfaces an aggregate error and keeps the bar visible", async () => {
      const user = userEvent.setup();
      mockPatchTask
        .mockResolvedValueOnce({
          id: "a",
          title: "Alpha",
          initial_prompt: "",
          status: "ready",
          priority: "medium",
          ...TASK_TEST_DEFAULTS,
        })
        .mockRejectedValueOnce(new Error("boom"))
        .mockResolvedValueOnce({
          id: "c",
          title: "Gamma",
          initial_prompt: "",
          status: "ready",
          priority: "medium",
          ...TASK_TEST_DEFAULTS,
        });
      const tasks = [
        makeRow("a", "Alpha"),
        makeRow("b", "Beta"),
        makeRow("c", "Gamma"),
      ];
      renderWithRouter(
        <TaskListSection
          tasks={tasks}
          loading={false}
          refreshing={false}
          saving={false}
          {...listPagerDefaults}
          rootTasksOnPage={3}
          onEdit={vi.fn()}
          onRequestClose={vi.fn()}
        />,
      );
      await user.click(screen.getByTestId("task-list-select-all"));
      await user.click(screen.getByTestId("task-list-bulk-bar-reschedule"));
      const fixedNow = Date.UTC(2026, 3, 19, 12, 0, 0);
      const dateNow = vi.spyOn(Date, "now").mockReturnValue(fixedNow);
      await user.click(screen.getByTestId("schedule-picker-quick-trigger"));
      await user.click(screen.getByTestId("schedule-picker-quick-hour-1"));
      dateNow.mockRestore();
      await user.click(screen.getByTestId("task-bulk-reschedule-submit"));
      await waitFor(() => {
        expect(mockPatchTask).toHaveBeenCalledTimes(3);
      });
      // Modal closes only on full success — partial failure leaves
      // it open so the operator can retry.
      await waitFor(() => {
        const banner = screen.getByTestId("task-list-bulk-error");
        expect(banner.textContent).toMatch(/1 of 3/);
      });
    });

    it("Cancel clears the running selection without firing PATCHes", async () => {
      const user = userEvent.setup();
      const tasks = [makeRow("a", "Alpha"), makeRow("b", "Beta")];
      renderWithRouter(
        <TaskListSection
          tasks={tasks}
          loading={false}
          refreshing={false}
          saving={false}
          {...listPagerDefaults}
          rootTasksOnPage={2}
          onEdit={vi.fn()}
          onRequestClose={vi.fn()}
        />,
      );
      await user.click(screen.getByTestId("task-list-select-all"));
      await user.click(screen.getByTestId("task-list-bulk-bar-cancel"));
      expect(mockPatchTask).not.toHaveBeenCalled();
      expect(mockCloseTask).not.toHaveBeenCalled();
      expect(
        screen.queryByTestId("task-list-bulk-bar"),
      ).not.toBeInTheDocument();
    });

    it("changing a filter clears the running selection", async () => {
      const user = userEvent.setup();
      const tasks = [
        makeRow("a", "Alpha", { status: "ready" }),
        makeRow("b", "Beta", { status: "done" }),
      ];
      renderWithRouter(
        <TaskListSection
          tasks={tasks}
          loading={false}
          refreshing={false}
          saving={false}
          {...listPagerDefaults}
          rootTasksOnPage={2}
          onEdit={vi.fn()}
          onRequestClose={vi.fn()}
        />,
      );
      await user.click(screen.getByTestId("task-list-select-row-a"));
      expect(screen.getByTestId("task-list-bulk-bar")).toBeInTheDocument();
      await user.click(screen.getByRole("combobox", { name: /^status$/i }));
      await user.click(screen.getByRole("option", { name: /^done$/i }));
      expect(
        screen.queryByTestId("task-list-bulk-bar"),
      ).not.toBeInTheDocument();
    });

    it("Scheduled (deferred) filter limits rows to ready+future", async () => {
      const user = userEvent.setup();
      const future = new Date(Date.now() + 60 * 60 * 1000).toISOString();
      const past = new Date(Date.now() - 60 * 60 * 1000).toISOString();
      const tasks = [
        makeRow("a", "Future ready", { pickup_not_before: future }),
        makeRow("b", "Past ready", { pickup_not_before: past }),
        makeRow("c", "Plain ready"),
      ];
      renderWithRouter(
        <TaskListSection
          tasks={tasks}
          loading={false}
          refreshing={false}
          saving={false}
          {...listPagerDefaults}
          rootTasksOnPage={3}
          onEdit={vi.fn()}
          onRequestClose={vi.fn()}
        />,
      );
      await user.click(screen.getByRole("combobox", { name: /^status$/i }));
      await user.click(
        screen.getByRole("option", { name: /^scheduled \(deferred\)$/i }),
      );
      expect(screen.getByText("Future ready")).toBeInTheDocument();
      await waitFor(
        () => {
          expect(screen.queryByText("Past ready")).not.toBeInTheDocument();
          expect(screen.queryByText("Plain ready")).not.toBeInTheDocument();
        },
        { timeout: 500 },
      );
    });
  });

});
