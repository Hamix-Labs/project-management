import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter, Route, Routes } from "react-router-dom";
import { describe, expect, it, vi } from "vitest";
import { ApiError } from "@/api";
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

vi.mock("@/components/workspace-picker", () => ({
  WorkspaceDirPickerModal: ({
    open,
    onSelect,
    onClose,
  }: {
    open: boolean;
    onSelect: (path: string) => void;
    onClose: () => void;
  }) =>
    open ? (
      <div role="dialog" aria-label="Choose a repository">
        <button type="button" onClick={() => onSelect("/repos/hamix")}>
          Use this repository
        </button>
        <button type="button" onClick={onClose}>
          Cancel picker
        </button>
      </div>
    ) : null,
}));

const repoId = FACTORY_GIT_REPO_ID;
const repoId2 = "00000000-0000-4000-8000-000000000011";

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
  it("shows repository setup copy when no repositories are registered", async () => {
    server.use(gitRepositoriesListEmpty());

    renderListPage();
    expect(await screen.findByRole("heading", { name: /^repositories$/i })).toBeInTheDocument();
    expect(
      await screen.findByText(
        /point hamix at a git repository on disk/i,
      ),
    ).toBeInTheDocument();
    expect(
      await screen.findByText(/register a repository to get started/i),
    ).toBeInTheDocument();
    const registerButtons = screen.getAllByRole("button", { name: /Register repository/i });
    expect(registerButtons).toHaveLength(1);
    await userEvent.click(registerButtons[0]!);
    expect(
      await screen.findByRole("button", { name: /Choose repository/i }),
    ).toBeInTheDocument();
    expect(screen.getByText(/No repository selected yet/i)).toBeInTheDocument();
    expect(screen.getByText(/^Browse$/)).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /^Register$/i })).toBeDisabled();
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

    renderListPage(["/repositories?register=1"]);
    expect(
      await screen.findByRole("button", { name: /Choose repository/i }),
    ).toBeInTheDocument();
  });

  it("renders register repository modal empty chrome when open", () => {
    render(
      <ModalStackProvider>
        <RegisterRepositoryModal
          open
          pending={false}
          error={null}
          registeredRepositories={[]}
          onClose={() => {}}
          onSubmit={() => {}}
        />
      </ModalStackProvider>,
    );
    expect(screen.getByRole("heading", { name: /Register repository/i })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /Choose repository/i })).toBeInTheDocument();
    expect(screen.getByText(/Browse your folders to select a Git repository/i)).toBeInTheDocument();
    expect(screen.getByText(/^Browse$/)).toBeInTheDocument();
    expect(screen.getByText(/No repository selected yet/i)).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /^Register$/i })).toBeDisabled();
    expect(screen.getByRole("button", { name: /^Close$/i })).toBeInTheDocument();
  });

  it("shows selected chrome after choosing a repository path", async () => {
    const onSubmit = vi.fn();
    render(
      <ModalStackProvider>
        <RegisterRepositoryModal
          open
          pending={false}
          error={null}
          registeredRepositories={[]}
          onClose={() => {}}
          onSubmit={onSubmit}
        />
      </ModalStackProvider>,
    );

    await userEvent.click(screen.getByRole("button", { name: /Choose repository/i }));
    await userEvent.click(screen.getByRole("button", { name: /Use this repository/i }));

    expect(await screen.findByText("hamix")).toBeInTheDocument();
    expect(screen.getByText("/repos/hamix")).toBeInTheDocument();
    expect(screen.getByText(/^Change$/)).toBeInTheDocument();
    expect(screen.getByText(/Ready to register this repository/i)).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /Change repository/i })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /^Register$/i })).toBeEnabled();

    await userEvent.click(screen.getByRole("button", { name: /^Register$/i }));
    expect(onSubmit).toHaveBeenCalledWith({ path: "/repos/hamix" });
  });

  it("blocks register when the selected path is already registered", async () => {
    const onSubmit = vi.fn();
    render(
      <ModalStackProvider>
        <RegisterRepositoryModal
          open
          pending={false}
          error={null}
          registeredRepositories={[
            { path: "/repos/hamix", host_path: "C:/Users/dev/Documents/hamix" },
          ]}
          onClose={() => {}}
          onSubmit={onSubmit}
        />
      </ModalStackProvider>,
    );

    await userEvent.click(screen.getByRole("button", { name: /Choose repository/i }));
    await userEvent.click(screen.getByRole("button", { name: /Use this repository/i }));

    expect(await screen.findByText("/repos/hamix")).toBeInTheDocument();
    expect(screen.getByText(/This repository is already registered/i)).toBeInTheDocument();
    expect(screen.queryByText(/Ready to register this repository/i)).not.toBeInTheDocument();
    expect(screen.getByRole("button", { name: /^Register$/i })).toBeDisabled();
    expect(screen.queryByRole("alert")).not.toBeInTheDocument();

    await userEvent.click(screen.getByRole("button", { name: /^Register$/i }));
    expect(onSubmit).not.toHaveBeenCalled();
  });

  it("shows already-registered status for a duplicate mutation error without Ready", async () => {
    const { rerender } = render(
      <ModalStackProvider>
        <RegisterRepositoryModal
          open
          pending={false}
          error={null}
          registeredRepositories={[]}
          onClose={() => {}}
          onSubmit={() => {}}
        />
      </ModalStackProvider>,
    );

    await userEvent.click(screen.getByRole("button", { name: /Choose repository/i }));
    await userEvent.click(screen.getByRole("button", { name: /Use this repository/i }));
    expect(await screen.findByText(/Ready to register this repository/i)).toBeInTheDocument();

    rerender(
      <ModalStackProvider>
        <RegisterRepositoryModal
          open
          pending={false}
          error={
            new ApiError("repository already registered (request abc)", {
              status: 409,
              code: "duplicate",
              requestId: "abc",
            })
          }
          registeredRepositories={[]}
          onClose={() => {}}
          onSubmit={() => {}}
        />
      </ModalStackProvider>,
    );

    expect(screen.getByText(/This repository is already registered/i)).toBeInTheDocument();
    expect(screen.queryByText(/Ready to register this repository/i)).not.toBeInTheDocument();
    expect(screen.queryByText(/request abc/i)).not.toBeInTheDocument();
    expect(screen.getByRole("button", { name: /^Register$/i })).toBeDisabled();
    expect(screen.queryByRole("alert")).not.toBeInTheDocument();
  });

  it("clears the selected path when the modal is closed and reopened", async () => {
    const { rerender } = render(
      <ModalStackProvider>
        <RegisterRepositoryModal
          open
          pending={false}
          error={null}
          registeredRepositories={[]}
          onClose={() => {}}
          onSubmit={() => {}}
        />
      </ModalStackProvider>,
    );

    await userEvent.click(screen.getByRole("button", { name: /Choose repository/i }));
    await userEvent.click(screen.getByRole("button", { name: /Use this repository/i }));
    expect(await screen.findByText("/repos/hamix")).toBeInTheDocument();

    rerender(
      <ModalStackProvider>
        <RegisterRepositoryModal
          open={false}
          pending={false}
          error={null}
          registeredRepositories={[]}
          onClose={() => {}}
          onSubmit={() => {}}
        />
      </ModalStackProvider>,
    );
    expect(screen.queryByRole("heading", { name: /Register repository/i })).not.toBeInTheDocument();

    rerender(
      <ModalStackProvider>
        <RegisterRepositoryModal
          open
          pending={false}
          error={null}
          registeredRepositories={[]}
          onClose={() => {}}
          onSubmit={() => {}}
        />
      </ModalStackProvider>,
    );

    expect(await screen.findByRole("button", { name: /Choose repository/i })).toBeInTheDocument();
    expect(screen.getByText(/No repository selected yet/i)).toBeInTheDocument();
    expect(screen.queryByText("/repos/hamix")).not.toBeInTheDocument();
    expect(screen.getByRole("button", { name: /^Register$/i })).toBeDisabled();
  });

  it("lists one repository with delete action", async () => {
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
    expect(screen.queryByText("main", { selector: ".repositories-list-row__branch" })).not.toBeInTheDocument();
    expect(screen.queryByRole("button", { name: /sync main/i })).not.toBeInTheDocument();
    expect(await screen.findByRole("button", { name: /delete main/i })).toBeInTheDocument();
    expect(await screen.findByText("1 of 1 repository")).toBeInTheDocument();
  });

  it("opens delete confirmation when delete is clicked", async () => {
    server.use(gitRepositoriesListOk([gitRepositoryFactory()]));

    renderListPage();
    await userEvent.click(await screen.findByRole("button", { name: /delete main/i }));
    expect(await screen.findByRole("heading", { name: /delete repository/i })).toBeInTheDocument();
    expect(screen.getByRole("dialog")).toHaveTextContent("/repo/main");
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
