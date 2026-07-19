import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { render, screen, waitFor, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter, Route, Routes } from "react-router-dom";
import { describe, expect, it, vi, afterEach } from "vitest";
import { ROUTER_FUTURE_FLAGS } from "@/lib/routerFutureFlags";
import { ModalStackProvider } from "@/shared/ModalStackContext";
import { requestUrl } from "@/test/requestUrl";
import { RepositoryWorktreesPage } from "./RepositoryWorktreesPage";
import { worktreeGitCopy } from "./worktreeGitCopy";

const repoId = "00000000-0000-4000-8000-000000000010";
const wtMain = "00000000-0000-4000-8000-000000000020";
const wtB = "00000000-0000-4000-8000-000000000030";
const branchId = "00000000-0000-4000-8000-000000000040";
const mainBranchId = "00000000-0000-4000-8000-000000000041";

function detailWorktreesResponse() {
  return {
    worktrees: [
      {
        id: wtMain,
        repository_id: repoId,
        path: "/repo/main",
        name: "main",
        is_main: true,
        branch_id: mainBranchId,
        created_at: "2026-06-22T12:00:00Z",
      },
      {
        id: wtB,
        repository_id: repoId,
        path: "/repo/feature",
        name: "feature",
        is_main: false,
        branch_id: branchId,
        created_at: "2026-06-22T12:00:00Z",
      },
    ],
  };
}

function detailBranchesResponse() {
  return {
    branches: [
      {
        id: mainBranchId,
        repository_id: repoId,
        name: "main",
        head_sha: "def456",
        created_at: "2026-06-22T12:00:00Z",
      },
      {
        id: branchId,
        repository_id: repoId,
        name: "feature",
        head_sha: "abc123",
        created_at: "2026-06-22T12:00:00Z",
      },
    ],
  };
}

function detailCheckoutStatusResponse() {
  return {
    worktrees: [
      {
        worktree_id: wtMain,
        available: true,
        dirty: false,
        detached: false,
        head_commit_at: "2026-06-22T12:00:00Z",
        has_upstream: true,
        ahead: 0,
        behind: 0,
        upstream: "origin/main",
      },
      {
        worktree_id: wtB,
        available: true,
        dirty: false,
        detached: false,
        head_commit_at: "2026-06-22T12:00:00Z",
        has_upstream: true,
        ahead: 0,
        behind: 0,
        upstream: "origin/feature",
      },
    ],
  };
}

function jsonResponse(body: unknown, init: ResponseInit = { status: 200 }): Response {
  return new Response(JSON.stringify(body), {
    ...init,
    headers: { "content-type": "application/json", ...(init.headers ?? {}) },
  });
}

function renderDetailPage() {
  const client = new QueryClient({
    defaultOptions: { queries: { retry: false, gcTime: 0 } },
  });
  return render(
    <QueryClientProvider client={client}>
      <MemoryRouter
        future={ROUTER_FUTURE_FLAGS}
        initialEntries={[`/worktrees/${repoId}`]}
      >
        <ModalStackProvider>
          <Routes>
            <Route path="/worktrees/:repositoryId" element={<RepositoryWorktreesPage />} />
          </Routes>
        </ModalStackProvider>
      </MemoryRouter>
    </QueryClientProvider>,
  );
}

function mockRepositoryDetailFetch(options?: {
  onRequest?: (method: string, url: string) => void;
}) {
  vi.spyOn(globalThis, "fetch").mockImplementation(async (input: RequestInfo | URL, init?: RequestInit) => {
    const url = requestUrl(input);
    const method = init?.method ?? "GET";
    options?.onRequest?.(method, url);
    if (method === "GET" && url.endsWith(`/git/repositories/${repoId}`)) {
      return jsonResponse({
        id: repoId,
        path: "/repo/main",
        git_common_dir: "",
        host_path: "",
        default_branch: "main",
        created_at: "2026-06-22T12:00:00Z",
        updated_at: "2026-06-22T12:00:00Z",
      });
    }
    if (method === "GET" && url.endsWith(`/git/repositories/${repoId}/worktrees/checkout-status`)) {
      return jsonResponse(detailCheckoutStatusResponse());
    }
    if (method === "GET" && url.endsWith(`/git/repositories/${repoId}/worktrees`)) {
      return jsonResponse(detailWorktreesResponse());
    }
    if (method === "GET" && url.includes(`/git/repositories/${repoId}/branches`)) {
      return jsonResponse(detailBranchesResponse());
    }
    if (method === "POST" && url.endsWith(`/git/repositories/${repoId}/sync`)) {
      return jsonResponse({
        status: "ok",
        report: {
          repo_path_updated: false,
          worktrees_path_updated: 0,
          worktrees_added: 0,
          worktrees_removed: 0,
          branches_head_updated: 0,
          worktrees_skipped: [],
          needs_branch_bind: [],
        },
      });
    }
    if (method === "DELETE") {
      return jsonResponse(
        { error: "task still running", code: "has_running_task" },
        { status: 409 },
      );
    }
    return jsonResponse({ error: "not found" }, { status: 404 });
  });
}

describe("RepositoryWorktreesPage", () => {
  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it("renders repository title and worktree rows", async () => {
    mockRepositoryDetailFetch();
    renderDetailPage();
    expect(await screen.findByRole("heading", { level: 1, name: "main" })).toBeInTheDocument();
    expect(await screen.findByText("main", { selector: ".worktree-row__label" })).toBeInTheDocument();
    expect(screen.getByText(worktreeGitCopy.primaryWorktreeBadge)).toBeInTheDocument();
    expect(screen.getByText("feature", { selector: ".worktree-row__label" })).toBeInTheDocument();
    expect(screen.getByText("2 of 2 worktrees")).toBeInTheDocument();
    expect(screen.getByText("2 branches checked out")).toBeInTheDocument();
    expect(screen.getByRole("navigation", { name: /repository navigation/i })).toBeInTheDocument();
    expect(screen.getByRole("search", { name: /search worktrees/i })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: worktreeGitCopy.reconcile })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: worktreeGitCopy.deleteRepository })).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: /add worktree/i })).not.toBeInTheDocument();
  });

  it("filters worktrees by search query", async () => {
    mockRepositoryDetailFetch();
    renderDetailPage();
    await screen.findByText("feature", { selector: ".worktree-row__label" });
    const search = screen.getByRole("searchbox", { name: /search worktrees/i });
    await userEvent.type(search, "nomatch");
    await waitFor(() => {
      expect(screen.queryByText("feature", { selector: ".worktree-row__label" })).not.toBeInTheDocument();
    });
    expect(screen.getByText(worktreeGitCopy.noMatchingWorktreesTitle)).toBeInTheDocument();
  });

  it("maps unregister 409 has_running_task to dialog copy", async () => {
    mockRepositoryDetailFetch();
    renderDetailPage();
    await screen.findByText("feature", { selector: ".worktree-row__label" });
    await userEvent.click(
      screen.getByRole("button", { name: /Worktree actions for feature/i }),
    );
    await userEvent.click(screen.getByRole("menuitem", { name: /Unregister worktree/i }));
    const dialog = screen.getByRole("dialog");
    await userEvent.click(within(dialog).getByRole("button", { name: /^Unregister$/i }));
    await waitFor(() => {
      expect(within(dialog).getByText(/task still running/i)).toBeInTheDocument();
    });
    expect(within(dialog).getByRole("button", { name: /^Unregister$/i })).toBeDisabled();
  });

  it("posts Sync to the sync endpoint", async () => {
    const calls: string[] = [];
    mockRepositoryDetailFetch({
      onRequest(method, url) {
        if (method === "POST" && url.includes("/sync")) {
          calls.push("sync");
        }
      },
    });
    renderDetailPage();
    await screen.findByRole("heading", { level: 1, name: "main" });
    await userEvent.click(screen.getByRole("button", { name: worktreeGitCopy.reconcile }));
    await waitFor(() => expect(calls).toContain("sync"));
  });
});
