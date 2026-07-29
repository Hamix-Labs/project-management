import { useQuery } from "@tanstack/react-query";
import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter } from "react-router-dom";
import { describe, expect, it, vi, beforeEach } from "vitest";
import { ROUTER_FUTURE_FLAGS } from "@/lib/routerFutureFlags";
import type { useTasksApp } from "../hooks/useTasksApp";
import { TasksAppProvider } from "../app/TasksAppProvider";
import { TaskTemplatesPage } from "./TaskTemplatesPage";

vi.mock("@tanstack/react-query", async (importOriginal) => {
  const actual = await importOriginal<typeof import("@tanstack/react-query")>();
  return {
    ...actual,
    useQuery: vi.fn(),
  };
});

vi.mock("@/hooks/useProjects", () => ({
  useProjects: () => ({ data: { projects: [] }, isPending: false, isError: false }),
}));

vi.mock("@/hooks/useGlobalRepositories", () => ({
  useGlobalRepositories: () => ({ data: [], isPending: false, isError: false }),
}));

const mockedUseQuery = vi.mocked(useQuery);

type App = ReturnType<typeof useTasksApp>;

function makeApp(overrides: Partial<App> = {}): App {
  return {
    openTemplateCreateModal: vi.fn(),
    editTemplateByID: vi.fn(),
    deleteTemplateByID: vi.fn().mockResolvedValue(undefined),
    instantiateTemplates: vi.fn(),
    instantiateTemplatesPending: false,
    loadTemplatePending: false,
    deleteTemplatePending: false,
    ...overrides,
  } as unknown as App;
}

function renderPage(app: App) {
  return render(
    <TasksAppProvider value={app}>
      <MemoryRouter future={ROUTER_FUTURE_FLAGS}>
        <TaskTemplatesPage />
      </MemoryRouter>
    </TasksAppProvider>,
  );
}

function createTasksButton(_total: number) {
  return screen.getByRole("button", { name: new RegExp(`create tasks`, "i") });
}

function createTasksTotalBadge(total: number) {
  return screen.getByText(String(total), { selector: ".template-batch-bar__create-badge" });
}

const templates = [
  {
    id: "tmpl-1",
    name: "Alpha template",
    created_at: "2026-01-01T00:00:00Z",
    updated_at: "2026-01-02T00:00:00Z",
    instantiate_count: 0,
  },
  {
    id: "tmpl-2",
    name: "Beta template",
    created_at: "2026-01-01T00:00:00Z",
    updated_at: "2026-01-02T00:00:00Z",
    instantiate_count: 0,
  },
];

describe("TaskTemplatesPage", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockedUseQuery.mockReturnValue({
      data: templates,
      isPending: false,
      isError: false,
      error: null,
      refetch: vi.fn(),
    } as unknown as ReturnType<typeof useQuery>);
  });

  it("lists templates from the query result", () => {
    renderPage(makeApp());
    expect(screen.getByText("Alpha template")).toBeInTheDocument();
    expect(screen.getByText("Beta template")).toBeInTheDocument();
  });

  it("shows batch action when rows are selected", async () => {
    const user = userEvent.setup();
    renderPage(makeApp());

    await user.click(screen.getByLabelText(/select alpha template/i));
    expect(createTasksButton(1)).toBeInTheDocument();
    expect(createTasksTotalBadge(1)).toBeInTheDocument();
  });

  it("calls instantiate with selected template items in order", async () => {
    const user = userEvent.setup();
    const instantiateTemplates = vi.fn().mockResolvedValue({ tasks: [{}], errors: [] });
    renderPage(makeApp({ instantiateTemplates }));

    await user.click(screen.getByLabelText(/select alpha template/i));
    await user.click(screen.getByLabelText(/select beta template/i));
    await user.click(createTasksButton(2));

    await waitFor(() => {
      expect(instantiateTemplates).toHaveBeenCalledWith([
        { template_id: "tmpl-1", count: 1 },
        { template_id: "tmpl-2", count: 1 },
      ]);
    });
  });

  it("applies batch default count to selected templates immediately", async () => {
    const user = userEvent.setup();
    const instantiateTemplates = vi.fn().mockResolvedValue({ tasks: [{}], errors: [] });
    renderPage(makeApp({ instantiateTemplates }));

    await user.click(screen.getByLabelText(/select alpha template/i));
    await user.click(screen.getByLabelText(/select beta template/i));

    const batchDefault = screen.getByLabelText(/instances per selected template/i);
    fireEvent.change(batchDefault, { target: { value: "5" } });
    await waitFor(() => {
      expect(batchDefault).toHaveValue("5");
    });

    await waitFor(() => {
      expect(createTasksTotalBadge(10)).toBeInTheDocument();
    });

    await user.click(createTasksButton(10));

    await waitFor(() => {
      expect(instantiateTemplates).toHaveBeenCalledWith([
        { template_id: "tmpl-1", count: 5 },
        { template_id: "tmpl-2", count: 5 },
      ]);
    });
  });

  it("allows per-row instance override", async () => {
    const user = userEvent.setup();
    const instantiateTemplates = vi.fn().mockResolvedValue({ tasks: [{}], errors: [] });
    renderPage(makeApp({ instantiateTemplates }));

    await user.click(screen.getByLabelText(/select alpha template/i));
    await user.click(screen.getByLabelText(/select beta template/i));

    const alphaQty = screen.getByLabelText(/instances for alpha template/i);
    fireEvent.change(alphaQty, { target: { value: "3" } });
    await waitFor(() => {
      expect(alphaQty).toHaveValue("3");
    });

    await waitFor(() => {
      expect(createTasksTotalBadge(4)).toBeInTheDocument();
    });

    await user.click(createTasksButton(4));

    await waitFor(() => {
      expect(instantiateTemplates).toHaveBeenCalledWith([
        { template_id: "tmpl-1", count: 3 },
        { template_id: "tmpl-2", count: 1 },
      ]);
    });
  });
});
