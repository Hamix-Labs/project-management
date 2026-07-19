import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter, Route, Routes } from "react-router-dom";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { ROUTER_FUTURE_FLAGS } from "@/lib/routerFutureFlags";
import { requestUrl } from "@/test/requestUrl";
import { makeTask } from "@/test/taskDefaults";
import { type Project } from "@/types";
import { ProjectDetailPage } from "./ProjectDetailPage";
import { projectQueryKeys } from "./queryKeys";

type FetchInput = RequestInfo | URL;

function jsonResponse(body: unknown, init: ResponseInit = { status: 200 }): Response {
  return new Response(JSON.stringify(body), {
    ...init,
    headers: { "content-type": "application/json", ...(init.headers ?? {}) },
  });
}

const testProject: Project = {
  id: "project-1",
  name: "Default project",
  description: "Shared context",
  status: "active",
  context_summary: "Shared context",
  is_default: false,
  created_at: "2026-04-27T00:00:00Z",
  updated_at: "2026-04-27T00:00:00Z",
};

let taskListBody: { tasks: ReturnType<typeof makeTask>[]; limit: number; offset: number; has_more: boolean };

function renderPage(
  project: Project = testProject,
  initialPath = `/projects/${project.id}`,
) {
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: { retry: false, staleTime: Infinity },
      mutations: { retry: false },
    },
  });
  queryClient.setQueryData(projectQueryKeys.detail(project.id), project);

  return render(
    <QueryClientProvider client={queryClient}>
      <MemoryRouter future={ROUTER_FUTURE_FLAGS} initialEntries={[initialPath]}>
        <Routes>
          <Route path="/projects/:projectId" element={<ProjectDetailPage />} />
        </Routes>
      </MemoryRouter>
    </QueryClientProvider>,
  );
}

describe("ProjectDetailPage", () => {
  beforeEach(() => {
    taskListBody = { tasks: [], limit: 200, offset: 0, has_more: false };
    vi.spyOn(globalThis, "fetch").mockImplementation(async (input: FetchInput) => {
      const u = requestUrl(input);
      if (u.startsWith("/tasks?")) {
        return jsonResponse(taskListBody);
      }
      if (/\/projects\/[^/]+\/context\?/.test(u) || /\/projects\/[^/]+\/context$/.test(u)) {
        return jsonResponse({ items: [], edges: [], limit: 100 });
      }
      return new Response(`unexpected fetch ${u}`, { status: 500 });
    });
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  it("presents settings, context, and linked work as distinct sections", async () => {
    renderPage();

    expect(screen.getByRole("heading", { name: "Default project" })).toBeInTheDocument();
    expect(screen.getByText("active")).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "Project settings" })).toBeInTheDocument();
    expect(screen.getByText("Manage the core details for this project")).toBeInTheDocument();
    expect(screen.queryByLabelText(/^Status$/)).not.toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Cancel" })).toBeInTheDocument();
    expect(screen.getByText(/Memory nodes/)).toBeInTheDocument();
    await waitFor(() => {
      expect(screen.getByText("0 nodes")).toBeInTheDocument();
    });
    expect(screen.getByRole("link", { name: /Project context/ })).toHaveAttribute(
      "href",
      "/projects/project-1/context",
    );
    expect(screen.getByRole("heading", { name: /Linked tasks/ })).toBeInTheDocument();
    await waitFor(() => {
      expect(screen.getByText("0 tasks connected to this project")).toBeInTheDocument();
    });
    expect(screen.queryByRole("link", { name: "Goals" })).not.toBeInTheDocument();
    expect(screen.queryByRole("link", { name: "Steps" })).not.toBeInTheDocument();
  });

  it("shows colored status badges for linked tasks and a status summary", async () => {
    taskListBody = {
      tasks: [
        makeTask({
          id: "t-run",
          title: "Auth refactor rollout",
          status: "running",
          project_id: "project-1",
        }),
        makeTask({
          id: "t-ready",
          title: "Session invalidation sweep",
          status: "ready",
          project_id: "project-1",
        }),
        makeTask({
          id: "t-other",
          title: "Other project task",
          status: "failed",
          project_id: "project-other",
        }),
      ],
      limit: 200,
      offset: 0,
      has_more: false,
    };

    renderPage();

    await waitFor(() => {
      expect(screen.getByText("2 tasks connected to this project")).toBeInTheDocument();
    });
    expect(screen.getByText("Auth refactor rollout")).toBeInTheDocument();
    expect(screen.getByText("Session invalidation sweep")).toBeInTheDocument();
    expect(screen.queryByText("Other project task")).not.toBeInTheDocument();
    expect(document.querySelector(".task-status-badge--tone-info")).toBeTruthy();
    expect(document.querySelector(".task-status-badge--tone-success")).toBeTruthy();
    expect(screen.getByLabelText("Task status summary")).toBeInTheDocument();
    expect(screen.getByTitle("Running: 1")).toBeInTheDocument();
    expect(screen.getByTitle("Ready: 1")).toBeInTheDocument();
  });

  it("resets settings fields when Cancel is pressed", async () => {
    const user = userEvent.setup();
    renderPage();

    const nameInput = screen.getByLabelText(/^Name$/);
    await user.clear(nameInput);
    await user.type(nameInput, "Renamed project");
    expect(nameInput).toHaveValue("Renamed project");

    await user.click(screen.getByRole("button", { name: "Cancel" }));
    expect(nameInput).toHaveValue("Default project");
  });

  it("shows delete project action for non-default projects", () => {
    renderPage();
    expect(screen.getByRole("button", { name: /^Delete project$/ })).toBeInTheDocument();
  });

  it("does not show delete project action for the built-in default project", () => {
    const builtIn: Project = {
      ...testProject,
      id: "00000000-0000-4000-8000-000000000099",
      name: "Default",
      is_default: true,
    };
    renderPage(builtIn, `/projects/${encodeURIComponent(builtIn.id)}`);
    expect(screen.queryByRole("button", { name: /^Delete project$/ })).not.toBeInTheDocument();
  });

  it("shows muted badge for archived projects", () => {
    renderPage({ ...testProject, status: "archived" });
    const badge = screen.getByText("archived");
    expect(badge.closest(".pd__badge")).toHaveClass("pd__badge--muted");
  });
});
