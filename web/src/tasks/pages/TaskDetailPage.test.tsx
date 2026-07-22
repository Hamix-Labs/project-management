import { http, HttpResponse } from "msw";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { render, screen, waitFor, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter, Route, Routes } from "react-router-dom";
import { beforeEach, describe, expect, it, vi } from "vitest";
import type { useTasksApp } from "../hooks/useTasksApp";
import { TasksAppProvider } from "../app/TasksAppProvider";
import { stubEventSource } from "../../test/browserMocks";
import { ROUTER_FUTURE_FLAGS } from "../../lib/routerFutureFlags";
import { DEFAULT_DOCUMENT_TITLE } from "../../shared/useDocumentTitle";
import { setupAppTest } from "@/test/integration/appHarness";
import {
  taskChecklist,
  taskChecklistItemPatch,
  taskGet,
  taskGetFlaky,
  taskGetPending,
  taskEventsListEmpty,
} from "@/test/handlers/tasks";
import { server } from "@/test/server";
import type { Task } from "@/types/task";
import { TaskDetailPage } from "./TaskDetailPage";

const { mockNavigate } = vi.hoisted(() => ({ mockNavigate: vi.fn() }));

const isUiFeatureOmitted = vi.hoisted(() => vi.fn((_feature: string) => false));

vi.mock("@/launch/omittedFeatures", () => ({
  isUiFeatureOmitted: (feature: string) => isUiFeatureOmitted(feature),
}));

vi.mock("react-router-dom", async (importOriginal) => {
  const actual = await importOriginal<typeof import("react-router-dom")>();
  return {
    ...actual,
    useNavigate: () => mockNavigate,
  };
});

function mockApp(): ReturnType<typeof useTasksApp> {
  return {
    deleteSuccess: false,
    deleteVariables: undefined,
    openEdit: vi.fn(),
    requestDelete: vi.fn(),
    saving: false,
  } as unknown as ReturnType<typeof useTasksApp>;
}

function appWithDeleteSuccess(
  variables: { id: string },
): ReturnType<typeof useTasksApp> {
  return {
    ...mockApp(),
    deleteSuccess: true,
    deleteVariables: variables,
  } as unknown as ReturnType<typeof useTasksApp>;
}

function renderDetail(
  initialPath: string,
  app: ReturnType<typeof useTasksApp>,
) {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false }, mutations: { retry: false } },
  });
  return render(
    <QueryClientProvider client={queryClient}>
      <TasksAppProvider value={app}>
        <MemoryRouter
          future={ROUTER_FUTURE_FLAGS}
          initialEntries={[initialPath]}
        >
          <Routes>
            <Route path="/tasks/:taskId" element={<TaskDetailPage />} />
          </Routes>
        </MemoryRouter>
      </TasksAppProvider>
    </QueryClientProvider>,
  );
}

type MockTaskDetailData = Partial<Task> & Pick<Task, "id" | "title">;

function taskDetail(
  id: string,
  title: string,
  overrides: Partial<Omit<Task, "id" | "title">> = {},
): MockTaskDetailData {
  return {
    id,
    title,
    initial_prompt: "",
    status: "ready",
    priority: "medium",
    runner: "cursor",
    cursor_model: "",
    ...overrides,
  };
}

function useTaskDetailHandlers(
  task: MockTaskDetailData,
  checklistItems: unknown[] = [],
) {
  server.use(
    taskGet(task.id, task),
    taskChecklist(task.id, checklistItems),
    taskEventsListEmpty(task.id),
  );
}

describe("TaskDetailPage", () => {
  beforeEach(() => {
    setupAppTest();
    stubEventSource();
    mockNavigate.mockClear();
    isUiFeatureOmitted.mockImplementation(() => false);
  });

  it("shows a loading skeleton while the task query is pending", () => {
    const [pendingHandler] = taskGetPending("t1");
    server.use(pendingHandler);

    renderDetail("/tasks/t1", mockApp());
    expect(
      screen.getByRole("status", { name: /loading task/i }),
    ).toBeInTheDocument();
  });

  it("shows task load error with retry and refetches successfully", async () => {
    const user = userEvent.setup();
    const task = taskDetail("t1", "Recovered title");
    server.use(
      taskGetFlaky("t1", task),
      taskChecklist("t1", []),
      taskEventsListEmpty("t1"),
    );

    renderDetail("/tasks/t1", mockApp());

    expect(await screen.findByRole("alert")).toBeInTheDocument();
    await user.click(screen.getByRole("button", { name: /try again/i }));

    expect(
      await screen.findByRole("heading", { name: /^recovered title$/i }),
    ).toBeInTheDocument();
  });

  it("collapses initial prompt by default and expands on demand", async () => {
    const user = userEvent.setup();
    useTaskDetailHandlers(
      taskDetail("t1", "Testing", {
        initial_prompt: "<p>Secret long body text</p>",
        priority: "critical",
      }),
    );

    renderDetail("/tasks/t1", mockApp());

    expect(await screen.findByRole("heading", { name: /^testing$/i })).toBeInTheDocument();
    expect(document.title).toBe(`Testing · ${DEFAULT_DOCUMENT_TITLE}`);
    expect(
      screen.getByRole("button", { name: /edit task/i }),
    ).toBeInTheDocument();

    const details = document.querySelector(
      ".task-detail-prompt .task-detail-collapsible",
    );
    expect(details).not.toBeNull();
    expect(details).not.toHaveAttribute("open");

    expect(
      screen.getByRole("heading", { name: /^initial prompt$/i }),
    ).toBeInTheDocument();
    expect(screen.queryByText("Secret long body text")).not.toBeVisible();

    await user.click(screen.getByRole("heading", { name: /^initial prompt$/i }));
    expect(details).toHaveAttribute("open");
    expect(screen.getByText("Secret long body text")).toBeVisible();
  });

  it("sanitizes unsafe HTML from initial prompt before rendering", async () => {
    useTaskDetailHandlers(
      taskDetail("txss", "Unsafe prompt", {
        initial_prompt:
          '<p>Safe text</p><img src=x onerror="window.__xss = 1" /><script>window.__xss_script = 1</script><a href="javascript:alert(1)">bad</a>',
      }),
    );

    renderDetail("/tasks/txss", mockApp());
    expect(
      await screen.findByRole("heading", { name: /^unsafe prompt$/i }),
    ).toBeInTheDocument();

    const promptBody = document.querySelector(
      ".task-detail-prompt-body",
    ) as HTMLElement | null;
    expect(promptBody).not.toBeNull();
    expect(promptBody!.innerHTML).not.toContain("<script");
    expect(promptBody!.innerHTML).not.toContain("onerror=");
    expect(promptBody!.innerHTML).not.toContain("javascript:");
    expect(promptBody).toHaveTextContent(/safe text/i);
  });

  it("shows an em dash when there is no visible initial prompt", async () => {
    useTaskDetailHandlers(taskDetail("t2", "Empty prompt"));

    renderDetail("/tasks/t2", mockApp());

    expect(
      await screen.findByRole("heading", { name: /^empty prompt$/i }),
    ).toBeInTheDocument();
    expect(
      screen.getByRole("heading", { name: /^initial prompt$/i }),
    ).toBeInTheDocument();
    expect(
      document.querySelector(".task-detail-prompt .task-detail-collapsible"),
    ).toBeNull();
    const empty = screen.getByText("—");
    expect(empty).toBeInTheDocument();
    expect(empty).toHaveClass("task-detail-prompt-empty");
  });

  it("surfaces the needs-user signal via the status pill", async () => {
    useTaskDetailHandlers(
      taskDetail("tb", "Blocked task", {
        status: "blocked",
      }),
    );

    renderDetail("/tasks/tb", mockApp());

    expect(
      await screen.findByRole("heading", { name: /^blocked task$/i }),
    ).toBeInTheDocument();
    expect(
      screen.getByText("Blocked", { selector: ".task-status-badge" }),
    ).toHaveAttribute(
      "data-needs-user",
      "true",
    );
    expect(screen.queryByText("Agent needs input")).not.toBeInTheDocument();
    expect(
      screen.queryByText(/the agent is blocked/i),
    ).not.toBeInTheDocument();
  });

  it("renders the Dependencies section with an empty state when there are none", async () => {
    useTaskDetailHandlers(taskDetail("tnd", "No deps task"));

    renderDetail("/tasks/tnd", mockApp());

    expect(
      await screen.findByRole("heading", { name: /^no deps task$/i }),
    ).toBeInTheDocument();
    expect(screen.getByTestId("task-deps-empty")).toBeInTheDocument();
    expect(
      screen.getByText(/no upstream dependencies/i),
    ).toBeInTheDocument();
  });

  it("hides Dependencies and Release gate when launch omits them", async () => {
    isUiFeatureOmitted.mockImplementation(
      (feature) => feature === "tagsAndDependencies" || feature === "releaseGates",
    );
    useTaskDetailHandlers(taskDetail("tnd", "No deps task"));

    renderDetail("/tasks/tnd", mockApp());

    expect(
      await screen.findByRole("heading", { name: /^no deps task$/i }),
    ).toBeInTheDocument();
    expect(screen.queryByTestId("task-deps-empty")).not.toBeInTheDocument();
    expect(screen.queryByTestId("task-gate-empty")).not.toBeInTheDocument();
    expect(screen.queryByText(/release gate/i)).not.toBeInTheDocument();
  });

  it("shows done criteria as read-only with progress counts", async () => {
    useTaskDetailHandlers(taskDetail("tc", "Checklist task"), [
      {
        id: "i1",
        sort_order: 0,
        text: "First",
        done: true,
      },
      {
        id: "i2",
        sort_order: 1,
        text: "Second",
        done: false,
      },
    ]);

    renderDetail("/tasks/tc", mockApp());

    expect(
      await screen.findByRole("heading", { name: /^checklist task$/i }),
    ).toBeInTheDocument();
    expect(
      await screen.findByRole("status", {
        name: /checklist progress: 1 of 2 requirements satisfied/i,
      }),
    ).toBeInTheDocument();
    expect(screen.queryByRole("checkbox")).not.toBeInTheDocument();
    expect(screen.getByText("First")).toBeInTheDocument();
    expect(screen.getByText("Second")).toBeInTheDocument();
  });

  it("shows checklist fetch error with try again and refetches", async () => {
    const user = userEvent.setup();
    const task = taskDetail("cf", "Checklist fetch");
    // Keep failing across remount/refetch until the alert is visible, then allow success.
    let failChecklist = true;
    server.use(
      taskGet(task.id, task),
      http.get(`/tasks/${task.id}/checklist`, () => {
        if (failChecklist) {
          return new HttpResponse(null, { status: 500 });
        }
        return HttpResponse.json({ items: [] });
      }),
      taskEventsListEmpty(task.id),
    );

    renderDetail(`/tasks/${task.id}`, mockApp());

    expect(
      await screen.findByRole("heading", { name: /^checklist fetch$/i }),
    ).toBeInTheDocument();

    const checklistSection = document.querySelector("#task-detail-checklist");
    expect(checklistSection).not.toBeNull();
    expect(
      await within(checklistSection as HTMLElement).findByRole("alert"),
    ).toBeInTheDocument();
    failChecklist = false;
    await user.click(
      within(checklistSection as HTMLElement).getByRole("button", {
        name: /try again/i,
      }),
    );
    expect(
      await within(checklistSection as HTMLElement).findByText(/no criteria yet/i),
    ).toBeInTheDocument();
  });

  it("navigates home after successful delete", () => {
    useTaskDetailHandlers(taskDetail("root1", "Root"));

    const app = appWithDeleteSuccess({ id: "root1" });

    renderDetail("/tasks/root1", app);

    expect(mockNavigate).toHaveBeenCalledWith("/", { replace: true });
  });

  it("disables Add criterion when the task is running or done", async () => {
    useTaskDetailHandlers(
      taskDetail("tr", "Running task", { status: "running" }),
    );

    renderDetail("/tasks/tr", mockApp());

    expect(
      await screen.findByRole("heading", { name: /^running task$/i }),
    ).toBeInTheDocument();

    const addBtn = screen.getByRole("button", { name: /^add criterion$/i });
    expect(addBtn).toBeDisabled();
  });

  it("edits a checklist criterion via PATCH text", async () => {
    const user = userEvent.setup();
    let patchBody: string | null = null;
    let checklistText = "Before";
    server.use(
      taskGet("te", taskDetail("te", "Edit checklist")),
      http.get("/tasks/te/checklist", () =>
        HttpResponse.json({
          items: [
            {
              id: "item-1",
              sort_order: 0,
              text: checklistText,
              done: false,
            },
          ],
        }),
      ),
      taskChecklistItemPatch("te", "item-1", (body) => {
        patchBody = body;
        checklistText = "After";
      }, "After"),
      taskEventsListEmpty("te"),
    );

    renderDetail("/tasks/te", mockApp());

    expect(
      await screen.findByRole("heading", { name: /^edit checklist$/i }),
    ).toBeInTheDocument();

    expect(await screen.findByText("Before")).toBeInTheDocument();

    const checklistSection = document.querySelector("#task-detail-checklist");
    expect(checklistSection).not.toBeNull();
    await user.click(
      await within(checklistSection as HTMLElement).findByRole("button", {
        name: /^edit$/i,
      }),
    );

    const dialog = await screen.findByRole("dialog");
    const input = within(dialog).getByLabelText(/^criterion$/i);
    await user.clear(input);
    await user.type(input, "After");

    await user.click(
      within(dialog).getByRole("button", { name: /^save changes$/i }),
    );

    await waitFor(() => {
      expect(patchBody).toBe(JSON.stringify({ text: "After" }));
    });
    expect(await screen.findByText("After")).toBeInTheDocument();
  });
});
