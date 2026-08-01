import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { render, screen, waitFor, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter } from "react-router-dom";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { ROUTER_FUTURE_FLAGS } from "@/lib/routerFutureFlags";
import {
  FACTORY_GIT_REPO_ID,
  gitRepositoriesList,
  projectCreate,
  projectsListEmpty,
} from "@/test/handlers/projects";
import { gitRepositoryFactory } from "@/test/factories/git";
import { server } from "@/test/server";
import { type Project } from "@/types";
import { ProjectListPage } from "./ProjectListPage";
import { projectQueryKeys } from "./queryKeys";

function project(index: number, overrides: Partial<Project> = {}): Project {
  return {
    id: `project-${index}`,
    name: `Project ${index}`,
    description: `Context space ${index}`,
    status: "active",
    is_default: false,
    created_at: "2026-04-27T00:00:00Z",
    updated_at: "2026-04-27T00:00:00Z",
    ...overrides,
  };
}

function renderPage(projects: Project[]) {
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: { retry: false, staleTime: Infinity },
      mutations: { retry: false },
    },
  });
  queryClient.setQueryData(projectQueryKeys.list(true, 50), {
    projects,
    limit: 50,
  });

  return render(
    <QueryClientProvider client={queryClient}>
      <MemoryRouter future={ROUTER_FUTURE_FLAGS}>
        <ProjectListPage />
      </MemoryRouter>
    </QueryClientProvider>,
  );
}

describe("ProjectListPage", () => {
  beforeEach(() => {
    // Strict MSW (HAMIX_MSW_UNHANDLED=error): page may fetch repos for default labels.
    server.use(gitRepositoriesList(), projectsListEmpty());
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  it("renders a dense library for larger project collections", async () => {
    const projects = Array.from({ length: 10 }, (_, index) =>
      project(index + 1, index > 7 ? { status: "archived" } : {}),
    );

    renderPage(projects);

    const summary = screen.getByLabelText("Project summary");
    expect(within(summary).getByText("10")).toBeInTheDocument();
    expect(within(summary).getByText("8")).toBeInTheDocument();
    expect(within(summary).getByText("2")).toBeInTheDocument();

    const library = await screen.findByLabelText("Projects");
    expect(within(library).getAllByRole("link")).toHaveLength(10);
    expect(
      within(library).queryAllByRole("button", { name: /^Delete project / }),
    ).toHaveLength(0);
    expect(
      within(library).getByRole("link", { name: /^Open project Project 10$/ }),
    ).toHaveAttribute("href", "/projects/project-10");
  });

  it("does not surface row delete controls on the list", () => {
    const projects: Project[] = [
      project(0, {
        name: "Default",
        is_default: true,
        repository_id: "00000000-0000-4000-8000-000000000010",
      }),
      project(1, { id: "custom-a", name: "Alpha" }),
      project(2, { id: "custom-b", name: "Beta" }),
    ];
    renderPage(projects);
    const library = screen.getByLabelText("Projects");
    expect(
      within(library).queryAllByRole("button", { name: /^Delete project / }),
    ).toHaveLength(0);
  });

  it("labels the global default project as Default", async () => {
    renderPage([
      project(0, {
        name: "Default",
        is_default: true,
        description: "Built-in project for tasks not assigned to a custom project.",
      }),
    ]);

    expect(
      await screen.findByRole("link", { name: /^Open project Default$/ }),
    ).toBeInTheDocument();
  });

  it("creates a project with name, description, and repository via the dialog", async () => {
    const repoId = FACTORY_GIT_REPO_ID;
    const created = {
      id: "new-1",
      name: "Payments",
      description: "Card flow",
      status: "active",
      repository_id: repoId,
      is_default: false,
      created_at: "2026-05-31T00:00:00Z",
      updated_at: "2026-05-31T00:00:00Z",
    };

    let captured: { name?: unknown; description?: unknown; repository_id?: unknown } | null =
      null;
    server.use(
      gitRepositoriesList([
        gitRepositoryFactory({
          id: repoId,
          path: "/repo/main",
          git_common_dir: "/repo/main/.git",
          main_branch_name: "main",
          created_at: "2026-05-31T00:00:00Z",
          updated_at: "2026-05-31T00:00:00Z",
        }),
      ]),
      projectCreate(created, (body) => {
        captured = body as {
          name?: unknown;
          description?: unknown;
          repository_id?: unknown;
        };
      }),
    );

    renderPage([]);

    const user = userEvent.setup();
    await user.click(screen.getByRole("button", { name: /new project/i }));

    const dialog = await screen.findByRole("dialog");
    await waitFor(() => {
      expect(within(dialog).getByLabelText(/^repository$/i)).toBeInTheDocument();
    });
    await user.type(
      within(dialog).getByLabelText(/^name$/i),
      created.name,
    );
    await user.type(
      within(dialog).getByLabelText(/^description/i),
      created.description,
    );
    await user.click(
      within(dialog).getByRole("button", { name: /create project/i }),
    );

    await waitFor(() => {
      expect(captured).not.toBeNull();
    });
    expect(captured).toMatchObject({
      name: created.name,
      description: created.description,
      repository_id: repoId,
    });
  });
});
