import "./taskListSection.testMocks";
import { closeTask } from "@/api";
import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";
import { TaskListSection } from "./TaskListSection";
import {
  listPagerDefaults,
  makeRow,
  renderWithRouter,
} from "./taskListSection.testSetup";

const mockCloseTask = vi.mocked(closeTask);

describe("TaskListSection bulk close", () => {
  it("opens confirm and POSTs close for each selected task", async () => {
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
    await user.click(screen.getByTestId("task-list-select-row-a"));
    await user.click(screen.getByTestId("task-list-select-row-b"));
    await user.click(screen.getByTestId("task-list-bulk-bar-close"));
    expect(
      screen.getByRole("heading", { name: /close 2 tasks/i }),
    ).toBeInTheDocument();
    await user.click(screen.getByTestId("task-bulk-close-confirm"));
    await waitFor(() => {
      expect(mockCloseTask).toHaveBeenCalledTimes(2);
    });
    const ids = mockCloseTask.mock.calls.map((c) => c[0]).sort();
    expect(ids).toEqual(["a", "b"]);
    await waitFor(() => {
      expect(
        screen.queryByTestId("task-list-bulk-bar"),
      ).not.toBeInTheDocument();
    });
  });
});
