import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter, Route, Routes } from "react-router-dom";
import { describe, expect, it, vi, afterEach } from "vitest";
import { ROUTER_FUTURE_FLAGS } from "@/lib/routerFutureFlags";
import { ModalStackProvider } from "@/shared/ModalStackContext";
import { requestUrl } from "@/test/requestUrl";
import { respondGlobalGitApi } from "@/test/handlers/gitGlobal";
import { RepositoriesListPage } from "./RepositoriesListPage";
import { RegisterRepositoryModal } from "./modals/RegisterRepositoryModal";
import { worktreeGitCopy } from "./worktreeGitCopy";

const repoId = "00000000-0000-4000-8000-000000000010";
const repoId2 = "00000000-0000-4000-8000-000000000011";

function jsonResponse(body: unknown, init: ResponseInit = { status: 200 }): Response {
  return new Response(JSON.stringify(body), {
    ...init,
    headers: { "content-type": "application/json", ...(init.headers ?? {}) },
  });
}

function repositoryJson(
  overrides: Record<string, unknown> = {},
): Record<string, unknown> {
  return {
    id: repoId,
    path: "/repo/main",
    git_common_dir: "",
    host_path: "",
    default_branch: "main",
    main_branch_name: "main",
    linked_worktree_count: 0,
    created_at: "2026-06-22T12:00:00Z",
    updated_at: "2026-06-22T12:00:00Z",
    ...overrides,
  };
}

function renderListPage(initialEntries: string[] = ["/repositories"]) {
  const client = new QueryClient({
    defaultOptions: { queries: { retry: false, gcTime: 0 } },
  });
  return render(
    <QueryClientProvider client={client}>
      <MemoryRouter future={ROUTER_FUTURE_FLAGS} initialEntries={initialEntries}>
        <ModalStackProvider>
          <Routes>
            <Route path="/repositories" element={<RepositoriesListPage />} />
          </Routes>
        </ModalStackProvider>
      </MemoryRouter>
    </QueryClientProvider>,
  );
}

describe("RepositoriesListPage", () => {
  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it("shows repository setup copy when no repositories are registered", async () => {
    vi.spyOn(globalThis, "fetch").mockImplementation(async (input: RequestInfo | URL) => {
      const url = requestUrl(input);
      if (url.endsWith("/git/repositories")) {
        return jsonResponse({ repositories: [] });
      }
      const res = respondGlobalGitApi(url, "GET");
      if (res) return res;
      return jsonResponse({ error: "not found" }, { status: 404 });
    });

    renderListPage();
    expect(await screen.findByRole("heading", { name: /^repositories$/i })).toBeInTheDocument();
    expect(
      await screen.findByText(/register repositories; hamix allocates worktrees/i),
    ).toBeInTheDocument();
    expect(
      await screen.findByText(/register a repository to get started/i),
    ).toBeInTheDocument();
    const registerButtons = screen.getAllByRole("button", { name: /Register repository/i });
    expect(registerButtons).toHaveLength(1);
    await userEvent.click(registerButtons[0]!);
    expect(
      await screen.findByRole("button", { name: /Choose folder/i }),
    ).toBeInTheDocument();
  });

  it("shows only an error callout when repository fetch fails with Not Found", async () => {
    vi.spyOn(globalThis, "fetch").mockImplementation(async (input: RequestInfo | URL) => {
      const url = requestUrl(input);
      if (url.endsWith("/git/repositories")) {
        return jsonResponse({ error: "Not Found" }, { status: 404 });
      }
      return jsonResponse({ error: "not found" }, { status: 404 });
    });

    renderListPage();
    const alert = await screen.findByRole("alert");
    expect(alert).toHaveTextContent(/could not load repositories/i);
    expect(alert).toHaveTextContent(/git API may be unavailable/i);
    expect(screen.getByRole("button", { name: /try again/i })).toBeInTheDocument();
    expect(screen.queryByText(/register a repository to get started/i)).not.toBeInTheDocument();
  });

  it("opens register modal from ?register=1 and strips the query param", async () => {
    vi.spyOn(globalThis, "fetch").mockImplementation(async (input: RequestInfo | URL) => {
      const url = requestUrl(input);
      if (url.endsWith("/git/repositories")) {
        return jsonResponse({ repositories: [] });
      }
      return jsonResponse({ error: "not found" }, { status: 404 });
    });

    renderListPage(["/repositories?register=1"]);
    expect(
      await screen.findByRole("button", { name: /Choose folder/i }),
    ).toBeInTheDocument();
  });

  it("renders register repository modal when open", () => {
    render(
      <ModalStackProvider>
        <RegisterRepositoryModal
          open
          pending={false}
          error={null}
          onClose={() => {}}
          onSubmit={() => {}}
        />
      </ModalStackProvider>,
    );
    expect(screen.getByRole("button", { name: /Choose folder/i })).toBeInTheDocument();
  });

  it("lists one repository with branch badge and sync/delete actions", async () => {
    vi.spyOn(globalThis, "fetch").mockImplementation(async (input: RequestInfo | URL) => {
      const url = requestUrl(input);
      if (url.endsWith("/git/repositories")) {
        return jsonResponse({
          repositories: [
            repositoryJson({
              linked_worktree_count: 1,
              main_branch_name: "main",
            }),
          ],
        });
      }
      return jsonResponse({ error: "not found" }, { status: 404 });
    });

    renderListPage();
    expect(
      await screen.findByRole("heading", { level: 1, name: /^repositories$/i }),
    ).toBeInTheDocument();
    expect(await screen.findByText("main", { selector: ".repositories-list-row__name" })).toBeInTheDocument();
    expect(await screen.findByText("main", { selector: ".repositories-list-row__branch" })).toBeInTheDocument();
    expect(await screen.findByRole("button", { name: /sync main/i })).toBeInTheDocument();
    expect(await screen.findByRole("button", { name: /delete main/i })).toBeInTheDocument();
    expect(await screen.findByText("1 of 1 repository")).toBeInTheDocument();
  });

  it("opens delete confirmation when delete is clicked", async () => {
    vi.spyOn(globalThis, "fetch").mockImplementation(async (input: RequestInfo | URL) => {
      const url = requestUrl(input);
      if (url.endsWith("/git/repositories")) {
        return jsonResponse({ repositories: [repositoryJson()] });
      }
      return jsonResponse({ error: "not found" }, { status: 404 });
    });

    renderListPage();
    await userEvent.click(await screen.findByRole("button", { name: /delete main/i }));
    expect(await screen.findByRole("heading", { name: /delete repository/i })).toBeInTheDocument();
    expect(screen.getByRole("dialog")).toHaveTextContent("/repo/main");
  });

  it("filters repositories with the search field and shows empty search state", async () => {
    vi.spyOn(globalThis, "fetch").mockImplementation(async (input: RequestInfo | URL) => {
      const url = requestUrl(input);
      if (url.endsWith("/git/repositories")) {
        return jsonResponse({
          repositories: [
            repositoryJson({
              id: repoId,
              path: "/repo/hamix",
              host_path: "C:/Users/dev/Documents/hamix",
            }),
            repositoryJson({
              id: repoId2,
              path: "/repo/other",
              host_path: "C:/Users/dev/Documents/other",
            }),
          ],
        });
      }
      return jsonResponse({ error: "not found" }, { status: 404 });
    });

    renderListPage();
    const search = await screen.findByRole("searchbox", { name: /search repositories/i });
    expect(await screen.findByText("hamix", { selector: ".repositories-list-row__name" })).toBeInTheDocument();
    expect(screen.getByText("other", { selector: ".repositories-list-row__name" })).toBeInTheDocument();
    expect(await screen.findByText("2 of 2 repositories")).toBeInTheDocument();

    await userEvent.clear(search);
    await userEvent.type(search, "hamix");
    await waitFor(() => {
      expect(screen.queryByText("other", { selector: ".repositories-list-row__name" })).not.toBeInTheDocument();
    });
    expect(screen.getByText("hamix", { selector: ".repositories-list-row__name" })).toBeInTheDocument();
    expect(await screen.findByText("1 of 2 repositories")).toBeInTheDocument();

    await userEvent.clear(search);
    await userEvent.type(search, "nomatch");
    await waitFor(() => {
      expect(screen.queryByText("hamix", { selector: ".repositories-list-row__name" })).not.toBeInTheDocument();
    });
    expect(await screen.findByText(/no repositories found/i)).toBeInTheDocument();
    await userEvent.click(
      screen.getByRole("button", { name: worktreeGitCopy.clearSearch }),
    );
    await waitFor(() => {
      expect(screen.getByText("hamix", { selector: ".repositories-list-row__name" })).toBeInTheDocument();
    });
  });
});
