import "./taskListSection.testMocks";
import { render, screen, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { MemoryRouter } from "react-router-dom";
import { ROUTER_FUTURE_FLAGS } from "../../../../lib/routerFutureFlags";
import { TASK_TEST_DEFAULTS } from "@/test/taskDefaults";
import { TaskListSection } from "./TaskListSection";
import {
  listPagerDefaults,
  makeRow,
  renderWithRouter,
} from "./taskListSection.testSetup";

describe("TaskListSection refetch and pager", () => {
  it("renders newly refetched tasks in created_at order after bulk template create", () => {
    const oldTask = makeRow("old-task", "Older task", { status: "done" });
    Object.assign(oldTask, { created_at: "2026-01-01T00:00:00Z" });

    const queryClient = new QueryClient({
      defaultOptions: {
        queries: { retry: false, staleTime: Infinity },
        mutations: { retry: false },
      },
    });

    const { rerender } = render(
      <QueryClientProvider client={queryClient}>
        <MemoryRouter future={ROUTER_FUTURE_FLAGS}>
          <TaskListSection
            tasks={[oldTask]}
            loading={false}
            refreshing={false}
            saving={false}
            smoothTransitions={false}
            {...listPagerDefaults}
            rootTasksOnPage={1}
            onEdit={vi.fn()}
            onRequestDelete={vi.fn()}
          />
        </MemoryRouter>
      </QueryClientProvider>,
    );

    const newTaskA = makeRow("new-a", "Refactor module", { status: "running" });
    Object.assign(newTaskA, { created_at: "2026-06-20T12:00:00Z" });
    const newTaskB = makeRow("new-b", "Split function", { status: "ready" });
    Object.assign(newTaskB, { created_at: "2026-06-20T11:59:00Z" });

    rerender(
      <QueryClientProvider client={queryClient}>
        <MemoryRouter future={ROUTER_FUTURE_FLAGS}>
          <TaskListSection
            tasks={[newTaskA, newTaskB, oldTask]}
            loading={false}
            refreshing={false}
            saving={false}
            smoothTransitions={false}
            {...listPagerDefaults}
            rootTasksOnPage={3}
            onEdit={vi.fn()}
            onRequestDelete={vi.fn()}
          />
        </MemoryRouter>
      </QueryClientProvider>,
    );

    const table = screen.getByRole("table", {
      name: /all tasks: title with context line/i,
    });
    const titles = within(table)
      .getAllByRole("row")
      .slice(1)
      .map(
        (row) =>
          row.querySelector(".cell-title-text--primary")?.textContent?.trim() ?? "",
      );

    expect(titles[0]).toMatch(/refactor module/i);
    expect(titles[1]).toMatch(/split function/i);
    expect(titles[2]).toMatch(/older task/i);
  });

  it("shows list pager when another server page may exist", async () => {
    const user = userEvent.setup();
    const onListPageChange = vi.fn();
    const task = {
      id: "1",
      title: "One",
      initial_prompt: "",
      status: "ready" as const,
      priority: "medium" as const,
      ...TASK_TEST_DEFAULTS,
      depth: 0,
    };
    const filler = Array.from({ length: 19 }, (_, i) => ({
      id: `x${i}`,
      title: `T${i}`,
      initial_prompt: "",
      status: "ready" as const,
      priority: "medium" as const,
      ...TASK_TEST_DEFAULTS,
      depth: 0,
    }));
    renderWithRouter(
      <TaskListSection
        tasks={[task, ...filler]}
        loading={false}
        refreshing={false}
        saving={false}
        listPage={0}
        listPageSize={20}
        onListPageChange={onListPageChange}
        onListFiltersChange={vi.fn()}
        hasNextPage
        hasPrevPage={false}
        rootTasksOnPage={20}
        onEdit={vi.fn()}
        onRequestDelete={vi.fn()}
      />,
    );
    await user.click(screen.getByRole("button", { name: /^next$/i }));
    expect(onListPageChange).toHaveBeenCalledWith(1);
  });
});
