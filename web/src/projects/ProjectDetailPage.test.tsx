import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter, Route, Routes } from "react-router-dom";
import { describe, expect, it } from "vitest";
import { ROUTER_FUTURE_FLAGS } from "@/lib/routerFutureFlags";
import { type Project } from "@/types";
import { ProjectDetailPage } from "./ProjectDetailPage";
import { projectQueryKeys } from "./queryKeys";

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
  it("presents project settings without a memory context section", () => {
    renderPage();

    expect(screen.getByRole("heading", { name: "Default project" })).toBeInTheDocument();
    expect(screen.getByText("active")).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "Project settings" })).toBeInTheDocument();
    expect(screen.getByText("Manage the core details for this project.")).toBeInTheDocument();
    expect(screen.queryByLabelText(/^Status$/)).not.toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Cancel" })).toBeInTheDocument();
    expect(screen.queryByRole("link", { name: /Project context/ })).not.toBeInTheDocument();
    expect(screen.queryByText(/Memory nodes/)).not.toBeInTheDocument();
    expect(screen.queryByRole("heading", { name: /Linked tasks/ })).not.toBeInTheDocument();
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
