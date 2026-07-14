import "./taskListSection.testMocks";
import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";
import { TASK_TEST_DEFAULTS } from "@/test/taskDefaults";
import { TaskListSection } from "./TaskListSection";
import {
  listPagerDefaults,
  makeRow,
  renderWithRouter,
} from "./taskListSection.testSetup";

describe("TaskListSection", () => {
  it("shows loading status", () => {
    renderWithRouter(
      <TaskListSection
        tasks={[]}
        loading
        refreshing={false}
        saving={false}
        smoothTransitions={false}
        {...listPagerDefaults}
        onEdit={vi.fn()}
        onRequestDelete={vi.fn()}
      />,
    );
    expect(
      screen.getByRole("status", { name: "Loading tasks" }),
    ).toBeInTheDocument();
  });

  it("shows syncing status when refreshing", () => {
    renderWithRouter(
      <TaskListSection
        tasks={[]}
        loading={false}
        refreshing
        saving={false}
        smoothTransitions={false}
        {...listPagerDefaults}
        onEdit={vi.fn()}
        onRequestDelete={vi.fn()}
      />,
    );
    expect(screen.getByText("Syncing with server…")).toBeInTheDocument();
  });

  it("shows the welcome empty state when not loading and no tasks", () => {
    renderWithRouter(
      <TaskListSection
        tasks={[]}
        loading={false}
        refreshing={false}
        saving={false}
        {...listPagerDefaults}
        onEdit={vi.fn()}
        onRequestDelete={vi.fn()}
      />,
    );
    // Title stays "No tasks yet" (precise, used as a page-ready sentinel
    // by App.test.tsx integration tests). Description copy is one short line
    // per frontend/components.mdc — assert on the new phrase so a future copy regression
    // is caught here.
    expect(screen.getByText(/no tasks yet/i)).toBeInTheDocument();
    expect(
      screen.getByRole("tabpanel", { name: /^all tasks$/i }),
    ).toBeInTheDocument();
    expect(
      screen.getByRole("table", {
        name: /all tasks: title with context line, status, priority, created time, project, and row actions/i,
      }),
    ).toBeInTheDocument();
  });

  it("calls emptyListAction when the empty-state CTA is used", async () => {
    const user = userEvent.setup();
    const onCreate = vi.fn();
    renderWithRouter(
      <TaskListSection
        tasks={[]}
        loading={false}
        refreshing={false}
        saving={false}
        {...listPagerDefaults}
        onEdit={vi.fn()}
        onRequestDelete={vi.fn()}
        emptyListAction={{
          label: "Create one",
          onClick: onCreate,
        }}
      />,
    );
    await user.click(screen.getByRole("button", { name: /^create one$/i }));
    expect(onCreate).toHaveBeenCalledTimes(1);
  });

  it("renders rows and calls onEdit", async () => {
    const user = userEvent.setup();
    const onEdit = vi.fn();
    const onRequestDelete = vi.fn();
    const task = {
      id: "1",
      title: "Alpha",
      initial_prompt: "",
      status: "ready" as const,
      priority: "medium" as const,
      ...TASK_TEST_DEFAULTS,
      depth: 0,
    };
    renderWithRouter(
      <TaskListSection
        tasks={[task]}
        loading={false}
        refreshing={false}
        saving={false}
        {...listPagerDefaults}
        rootTasksOnPage={1}
        onEdit={onEdit}
        onRequestDelete={onRequestDelete}
      />,
    );
    await user.click(
      screen.getByRole("button", { name: /^edit task "alpha"$/i }),
    );
    expect(onEdit).toHaveBeenCalledWith(task);
    await user.click(
      screen.getByRole("button", { name: /^delete task "alpha"$/i }),
    );
    expect(onRequestDelete).toHaveBeenCalledWith({
      ...task,
    });
  });

  it("renders heading summary when stats and rows are present", () => {
    renderWithRouter(
      <TaskListSection
        tasks={[makeRow("1", "Alpha")]}
        loading={false}
        refreshing={false}
        saving={false}
        {...listPagerDefaults}
        rootTasksOnPage={1}
        onEdit={vi.fn()}
        onRequestDelete={vi.fn()}
        taskStats={{
          total: 15,
          ready: 7,
          scheduled: 0,
          critical: 0,
          by_status: { review: 2, blocked: 2 },
          by_priority: {},
          cycles: { by_status: {}, by_triggered_by: {} },
          phases: {
            by_phase_status: { execute: {}, verify: {} },
          },
          runner: {
            by_runner: {},
            by_model: {},
            by_runner_model: {},
            by_runner_model_resolved: {},
          },
          recent_failures: [],
        }}
      />,
    );
    expect(
      screen.getByText("1 shown · 7 ready · 2 in review · 2 blocked"),
    ).toBeInTheDocument();
  });

  it("disables edit but keeps delete enabled for running tasks", async () => {
    const user = userEvent.setup();
    const onEdit = vi.fn();
    const onRequestDelete = vi.fn();
    const task = {
      id: "1",
      title: "Running task",
      initial_prompt: "",
      status: "running" as const,
      priority: "medium" as const,
      ...TASK_TEST_DEFAULTS,
      depth: 0,
    };
    renderWithRouter(
      <TaskListSection
        tasks={[task]}
        loading={false}
        refreshing={false}
        saving={false}
        {...listPagerDefaults}
        rootTasksOnPage={1}
        onEdit={onEdit}
        onRequestDelete={onRequestDelete}
      />,
    );
    const editButton = screen.getByRole("button", {
      name: /cannot edit task "running task" while in progress/i,
    });
    expect(editButton).toBeDisabled();
    await user.click(editButton);
    expect(onEdit).not.toHaveBeenCalled();

    await user.click(
      screen.getByRole("button", { name: /^delete task "running task"$/i }),
    );
    expect(onRequestDelete).toHaveBeenCalledWith(task);
  });

  it("filters rows by status and priority", async () => {
    const user = userEvent.setup();
    const tasks = [
      {
        id: "1",
        title: "Low ready",
        initial_prompt: "",
        status: "ready" as const,
        priority: "low" as const,
        ...TASK_TEST_DEFAULTS,
        depth: 0,
      },
      {
        id: "2",
        title: "High done",
        initial_prompt: "",
        status: "done" as const,
        priority: "high" as const,
        ...TASK_TEST_DEFAULTS,
        depth: 0,
      },
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
        onRequestDelete={vi.fn()}
      />,
    );
    expect(screen.getByText("Low ready")).toBeInTheDocument();
    expect(screen.getByText("High done")).toBeInTheDocument();

    await user.click(screen.getByRole("tab", { name: /^ready$/i }));
    expect(screen.getByText("Low ready")).toBeInTheDocument();
    await waitFor(
      () => {
        expect(screen.queryByText("High done")).not.toBeInTheDocument();
      },
      { timeout: 500 },
    );

    await user.click(screen.getByRole("tab", { name: /^all$/i }));
    await user.click(screen.getByRole("combobox", { name: /^priority$/i }));
    await user.click(screen.getByRole("option", { name: /^high$/i }));
    await waitFor(
      () => {
        expect(screen.queryByText("Low ready")).not.toBeInTheDocument();
      },
      { timeout: 500 },
    );
    expect(screen.getByText("High done")).toBeInTheDocument();
  });

  it("renders status tabs for filtering", () => {
    renderWithRouter(
      <TaskListSection
        tasks={[]}
        loading={false}
        refreshing={false}
        saving={false}
        smoothTransitions={false}
        {...listPagerDefaults}
        onEdit={vi.fn()}
        onRequestDelete={vi.fn()}
      />,
    );
    expect(screen.getByRole("tablist", { name: /filter tasks by status/i })).toBeInTheDocument();
    expect(screen.getByRole("tab", { name: /^all$/i })).toBeInTheDocument();
    expect(screen.getByRole("tab", { name: /^ready$/i })).toBeInTheDocument();
    expect(screen.getByRole("tab", { name: /^in progress$/i })).toBeInTheDocument();
  });

  it("filters rows by title search", async () => {
    const user = userEvent.setup();
    const tasks = [
      {
        id: "1",
        title: "Alpha task",
        initial_prompt: "",
        status: "ready" as const,
        priority: "medium" as const,
        ...TASK_TEST_DEFAULTS,
        depth: 0,
      },
      {
        id: "2",
        title: "Beta",
        initial_prompt: "",
        status: "ready" as const,
        priority: "medium" as const,
        ...TASK_TEST_DEFAULTS,
        depth: 0,
      },
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
        onRequestDelete={vi.fn()}
      />,
    );
    expect(screen.getByText("Alpha task")).toBeInTheDocument();
    expect(screen.getByText("Beta")).toBeInTheDocument();

    const search = screen.getByLabelText(/^search titles$/i);
    await user.type(search, "alp");
    expect(screen.getByText("Alpha task")).toBeInTheDocument();
    await waitFor(
      () => {
        expect(screen.queryByText("Beta")).not.toBeInTheDocument();
      },
      { timeout: 500 },
    );

    await user.clear(search);
    await waitFor(() => {
      expect(screen.getByText("Beta")).toBeInTheDocument();
    });
  });

  it("filters rows by project membership", async () => {
    const user = userEvent.setup();
    const tasks = [
      makeRow("1", "Moat task", { project_id: "project-1" }),
      makeRow("2", "Unassigned task"),
    ];
    renderWithRouter(
      <TaskListSection
        tasks={tasks}
        loading={false}
        refreshing={false}
        saving={false}
        {...listPagerDefaults}
        rootTasksOnPage={2}
        projectFilterOptions={[{ id: "project-1", name: "Context moat" }]}
        onEdit={vi.fn()}
        onRequestDelete={vi.fn()}
      />,
    );
    expect(screen.getByText("Moat task")).toBeInTheDocument();
    expect(screen.getByText("Unassigned task")).toBeInTheDocument();
    expect(screen.getByText("Context moat")).toBeInTheDocument();

    await user.click(screen.getByRole("combobox", { name: /^project$/i }));
    await user.click(screen.getByRole("option", { name: /^context moat$/i }));
    expect(screen.getByText("Moat task")).toBeInTheDocument();
    await waitFor(
      () => {
        expect(screen.queryByText("Unassigned task")).not.toBeInTheDocument();
      },
      { timeout: 500 },
    );

    await user.click(screen.getByRole("combobox", { name: /^project$/i }));
    expect(screen.queryByRole("option", { name: /^no project$/i })).not.toBeInTheDocument();
    await user.click(screen.getByRole("option", { name: /^all projects$/i }));
    expect(screen.getByText("Moat task")).toBeInTheDocument();
    expect(screen.getByText("Unassigned task")).toBeInTheDocument();
  });

  it("shows copy when no tasks match filters", async () => {
    const user = userEvent.setup();
    renderWithRouter(
      <TaskListSection
        tasks={[
          {
            id: "1",
            title: "Only ready",
            initial_prompt: "",
            status: "ready" as const,
            priority: "medium" as const,
            ...TASK_TEST_DEFAULTS,
            depth: 0,
          },
        ]}
        loading={false}
        refreshing={false}
        saving={false}
        {...listPagerDefaults}
        rootTasksOnPage={1}
        onEdit={vi.fn()}
        onRequestDelete={vi.fn()}
      />,
    );
    await user.click(screen.getByRole("tab", { name: /^failed$/i }));
    await waitFor(
      () => {
        expect(screen.getByText(/no matching tasks/i)).toBeInTheDocument();
      },
      { timeout: 500 },
    );
  });

});
