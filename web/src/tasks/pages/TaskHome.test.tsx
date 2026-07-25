import { render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { ROUTER_FUTURE_FLAGS } from "@/lib/routerFutureFlags";
import { projectsListEmpty } from "@/test/handlers/projects";
import { tasksListEmpty } from "@/test/handlers/tasks";
import { server } from "@/test/server";
import type { useTasksApp } from "../hooks/useTasksApp";
import { TasksAppProvider } from "../app/TasksAppProvider";
import { TaskHome } from "./TaskHome";

vi.mock("../components/task-list", () => ({
  TaskListSection: () => <div data-testid="task-list-section" />,
}));

vi.mock("../components/task-board/TaskBoardSection", () => ({
  TaskBoardSection: () => <div data-testid="task-board-section" />,
}));

vi.mock("../components/task-board/TaskHomeViewToggle", () => ({
  TaskHomeViewToggle: ({
    value,
    onChange,
  }: {
    value: string;
    onChange: (v: "list" | "board") => void;
  }) => (
    <div data-testid="view-toggle">
      <button type="button" onClick={() => onChange("list")}>
        List
      </button>
      <button type="button" onClick={() => onChange("board")}>
        Board
      </button>
      <span data-testid="view-value">{value}</span>
    </div>
  ),
}));

type App = ReturnType<typeof useTasksApp>;

function makeApp(overrides: Partial<App> = {}): App {
  return {
    tasks: [],
    rootTasksOnPage: [],
    loading: false,
    listRefreshing: false,
    saving: false,
    sseLive: false,
    taskListPage: 0,
    taskListPageSize: 50,
    hasNextTaskPage: false,
    hasPrevTaskPage: false,
    setTaskListPage: () => {},
    resetTaskListPage: () => {},
    openEdit: () => {},
    requestDelete: () => {},
    patchPending: false,
    deletePending: false,
    openCreateModal: () => {},
    closeCreateModal: () => {},
    createModalOpen: false,
    createEntryDraftErrorHint: false,
    retryCreateEntryDraftLoad: () => {},
    draftPickerOpen: false,
    setDraftPickerOpen: () => {},
    taskDrafts: [],
    draftListLoading: false,
    draftListError: null,
    retryDraftList: () => {},
    startFreshDraft: async () => {},
    resumeDraftByID: async () => {},
    resumeDraftPending: false,
    resumeDraftError: null,
    taskStats: undefined,
    taskStatsLoading: true,
    homeDataReady: true,
    ...overrides,
  } as unknown as App;
}

function renderHome(app: App, entries: string[] = ["/"]) {
  const client = new QueryClient({
    defaultOptions: { queries: { retry: false, staleTime: Infinity } },
  });
  return render(
    <QueryClientProvider client={client}>
      <TasksAppProvider value={app}>
        <MemoryRouter initialEntries={entries} future={ROUTER_FUTURE_FLAGS}>
          <TaskHome />
        </MemoryRouter>
      </TasksAppProvider>
    </QueryClientProvider>,
  );
}

describe("TaskHome", () => {
  beforeEach(() => {
    server.use(projectsListEmpty(), tasksListEmpty());
  });

  it("renders the task list without KPI stats cards", () => {
    renderHome(makeApp());

    expect(screen.getByTestId("task-list-section")).toBeInTheDocument();
    expect(screen.queryByLabelText("Task overview")).not.toBeInTheDocument();
    expect(screen.queryByText(/total tasks/i)).not.toBeInTheDocument();
  });

  it("renders the board section when view=board", () => {
    renderHome(makeApp(), ["/?view=board"]);
    expect(screen.getByTestId("task-board-section")).toBeInTheDocument();
    expect(screen.queryByTestId("task-list-section")).not.toBeInTheDocument();
  });
});
