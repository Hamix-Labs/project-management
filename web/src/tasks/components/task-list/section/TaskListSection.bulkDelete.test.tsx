import "./taskListSection.testMocks";
import { deleteTask } from "@/api";
import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";
import { TaskListSection } from "./TaskListSection";
import {
  listPagerDefaults,
  makeRow,
  renderWithRouter,
} from "./taskListSection.testSetup";

const mockDeleteTask = vi.mocked(deleteTask);

describe("TaskListSection bulk delete", () => {
  it("opens confirm and DELETEs each selected task", async () => {
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
        onRequestDelete={vi.fn()}
      />,
    );
    await user.click(screen.getByTestId("task-list-select-row-a"));
    await user.click(screen.getByTestId("task-list-select-row-b"));
    await user.click(screen.getByTestId("task-list-bulk-bar-delete"));
    expect(
      screen.getByRole("heading", { name: /delete 2 tasks/i }),
    ).toBeInTheDocument();
    await user.click(screen.getByTestId("task-bulk-delete-confirm"));
    await waitFor(() => {
      expect(mockDeleteTask).toHaveBeenCalledTimes(2);
    });
    const ids = mockDeleteTask.mock.calls.map((c) => c[0]).sort();
    expect(ids).toEqual(["a", "b"]);
    await waitFor(() => {
      expect(
        screen.queryByTestId("task-list-bulk-bar"),
      ).not.toBeInTheDocument();
    });
  });
});
