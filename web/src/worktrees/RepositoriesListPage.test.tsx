import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter, Route, Routes } from "react-router-dom";
import { describe, expect, it } from "vitest";
import { ROUTER_FUTURE_FLAGS } from "@/lib/routerFutureFlags";
import { ModalStackProvider } from "@/shared/ModalStackContext";
import { gitRepositoryFactory } from "@/test/factories/git";
import {
  FACTORY_GIT_REPO_ID,
  gitRepositoriesListEmpty,
  gitRepositoriesListError,
  gitRepositoriesListOk,
} from "@/test/handlers/worktrees";
import { server } from "@/test/server";
import { RepositoriesListPage } from "./RepositoriesListPage";
import { RegisterRepositoryModal } from "./modals/RegisterRepositoryModal";
import { worktreeGitCopy } from "./worktreeGitCopy";

const repoId = FACTORY_GIT_REPO_ID;
const repoId2 = "00000000-0000-4000-8000-000000000011";

function renderListPage(initialEntries: string[] = ["/worktrees"]) {
  const client = new QueryClient({
    defaultOptions: { queries: { retry: false, gcTime: 0 } },
  });
  return render(
    <QueryClientProvider client={client}>
      <MemoryRouter future={ROUTER_FUTURE_FLAGS} initialEntries={initialEntries}>
        <ModalStackProvider>
          <Routes>
            <Route path="/worktrees" element={<RepositoriesListPage />} />
            <Route path="/worktrees/:repositoryId" element={<div>Detail</div>} />
          </Routes>
        </ModalStackProvider>
      </MemoryRouter>
    </QueryClientProvider>,
  );
}

describe("RepositoriesListPage", () => {
  it("shows repository setup copy when no repositories are registered", async () => {
    server.use(gitRepositoriesListEmpty());

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
    server.use(gitRepositoriesListError(404, "Not Found"));

    renderListPage();
    const alert = await screen.findByRole("alert");
    expect(alert).toHaveTextContent(/could not load repositories/i);
    expect(alert).toHaveTextContent(/git API may be unavailable/i);
    expect(screen.getByRole("button", { name: /try again/i })).toBeInTheDocument();
    expect(screen.queryByText(/register a repository to get started/i)).not.toBeInTheDocument();
  });

  it("opens register modal from ?register=1 and strips the query param", async () => {
    server.use(gitRepositoriesListEmpty());

    renderListPage(["/worktrees?register=1"]);
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

  it("lists one repository with branch badge and worktree count from API", async () => {
    server.use(
      gitRepositoriesListOk([
        gitRepositoryFactory({
          linked_worktree_count: 1,
          main_branch_name: "main",
        }),
      ]),
    );

    renderListPage();
    expect(
      await screen.findByRole("heading", { level: 1, name: /^repositories$/i }),
    ).toBeInTheDocument();
    expect(await screen.findByText("main", { selector: ".repositories-list-row__name" })).toBeInTheDocument();
    expect(await screen.findByText("main", { selector: ".repositories-list-row__branch" })).toBeInTheDocument();
    expect(await screen.findByRole("gridcell", { name: /1 worktree/i })).toHaveTextContent("1");
    expect(await screen.findByText("1 of 1 repository")).toBeInTheDocument();
  });

  it("navigates to repository detail when a row is clicked", async () => {
    server.use(gitRepositoriesListOk([gitRepositoryFactory()]));

    renderListPage();
    const row = await screen.findByRole("row", { name: /main, 0 worktrees/i });
    await userEvent.click(row);
    expect(await screen.findByText("Detail")).toBeInTheDocument();
  });

  it("filters repositories with the search field and shows empty search state", async () => {
    server.use(
      gitRepositoriesListOk([
        gitRepositoryFactory({
          id: repoId,
          path: "/repo/hamix",
          host_path: "C:/Users/dev/Documents/hamix",
        }),
        gitRepositoryFactory({
          id: repoId2,
          path: "/repo/other",
          host_path: "C:/Users/dev/Documents/other",
        }),
      ]),
    );

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
