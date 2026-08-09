import { useEffect, useState, type ReactNode } from "react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { act, render, screen, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import type { FileWorktreeResolution } from "../usePromptEditorFileWorktree";
import { usePromptFileMentionSearch } from "./usePromptFileMentionSearch";

const { listRepoFiles, FakeApiError } = vi.hoisted(() => ({
  listRepoFiles: vi.fn(),
  FakeApiError: class FakeApiError extends Error {
    readonly status: number;
    constructor(status: number) {
      super(`http ${status}`);
      this.status = status;
    }
  },
}));

vi.mock("@/api", () => ({
  ApiError: FakeApiError,
  listRepoFiles,
  repoQueryKeys: {
    files: (worktreeId: string) => ["repo", "files", worktreeId],
  },
}));

/**
 * Mirrors `useLoadSuggestionMenuItems` in @blocknote/react: `getItems` is a
 * dependency of the loading effect, so an unstable identity reloads forever.
 */
function FakeSuggestionMenuController({
  query,
  getItems,
}: {
  query: string;
  getItems: (query: string) => Promise<{ title: string }[]>;
}) {
  const [items, setItems] = useState<{ title: string }[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    setLoading(true);
    void getItems(query).then((next) => {
      setItems(next);
      setLoading(false);
    });
  }, [query, getItems]);

  return (
    <div>
      <span data-testid="loading">{loading ? "loading" : "settled"}</span>
      <span data-testid="count">{items.length}</span>
      <span data-testid="items">{items.map((i) => i.title).join(",")}</span>
    </div>
  );
}

function worktreeStub(
  overrides: Partial<FileWorktreeResolution> = {},
): FileWorktreeResolution {
  return {
    worktreeId: "wt-1",
    resolving: false,
    whenResolved: async () => overrides.worktreeId ?? "wt-1",
    ...overrides,
  };
}

function Harness({
  worktree,
  query = "",
  onSelectPath = () => {},
}: {
  worktree: FileWorktreeResolution;
  query?: string;
  onSelectPath?: (path: string) => void;
}) {
  const { getItems, status } = usePromptFileMentionSearch({
    worktree,
    onSelectPath,
  });

  return (
    <>
      <span data-testid="status">{status.kind}</span>
      <FakeSuggestionMenuController query={query} getItems={getItems} />
    </>
  );
}

function renderWithClient(ui: ReactNode) {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  });
  const wrapper = ({ children }: { children: ReactNode }) => (
    <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
  );
  return { ...render(ui, { wrapper }), queryClient };
}

function fileList(paths: string[], truncated = false) {
  return { paths, truncated, source: "git" as const };
}

describe("usePromptFileMentionSearch", () => {
  beforeEach(() => {
    listRepoFiles.mockReset();
    listRepoFiles.mockResolvedValue(fileList(["web/src/main.tsx"]));
  });

  it("issues exactly one request per worktree even though status changes re-render", async () => {
    const { rerender } = renderWithClient(<Harness worktree={worktreeStub()} />);

    await waitFor(() =>
      expect(screen.getByTestId("status")).toHaveTextContent("ready"),
    );

    // Writing status re-renders the host. An unstable getItems would restart
    // the load on every one of these and never stop.
    for (let i = 0; i < 5; i += 1) {
      rerender(<Harness worktree={worktreeStub()} />);
      await act(async () => {});
    }

    expect(listRepoFiles).toHaveBeenCalledTimes(1);
  });

  it("keeps the same getItems identity across renders", async () => {
    const seen = new Set<unknown>();

    function IdentityProbe() {
      const { getItems, status } = usePromptFileMentionSearch({
        worktree: worktreeStub(),
        onSelectPath: () => {},
      });
      seen.add(getItems);
      return (
        <>
          <span data-testid="status">{status.kind}</span>
          <FakeSuggestionMenuController query="" getItems={getItems} />
        </>
      );
    }

    renderWithClient(<IdentityProbe />);
    await waitFor(() =>
      expect(screen.getByTestId("status")).toHaveTextContent("ready"),
    );

    expect(seen.size).toBe(1);
  });

  it("serves later queries from the cache without another request", async () => {
    listRepoFiles.mockResolvedValue(
      fileList(["web/src/main.tsx", "docs/api.md"]),
    );
    const { rerender } = renderWithClient(<Harness worktree={worktreeStub()} />);

    await waitFor(() =>
      expect(screen.getByTestId("count")).toHaveTextContent("2"),
    );

    rerender(<Harness worktree={worktreeStub()} query="api" />);

    await waitFor(() =>
      expect(screen.getByTestId("items")).toHaveTextContent("docs/api.md"),
    );
    expect(listRepoFiles).toHaveBeenCalledTimes(1);
  });

  it("waits for an in-flight worktree resolution instead of reporting a failure", async () => {
    let release: (worktreeId: string | undefined) => void = () => {};
    const pending = new Promise<string | undefined>((resolve) => {
      release = resolve;
    });

    renderWithClient(
      <Harness
        worktree={{
          worktreeId: undefined,
          resolving: true,
          whenResolved: () => pending,
        }}
      />,
    );

    await waitFor(() =>
      expect(screen.getByTestId("status")).toHaveTextContent("resolving"),
    );
    expect(listRepoFiles).not.toHaveBeenCalled();

    release("wt-late");
    await waitFor(() =>
      expect(screen.getByTestId("items")).toHaveTextContent("web/src/main.tsx"),
    );
    expect(listRepoFiles).toHaveBeenCalledWith(
      "wt-late",
      expect.objectContaining({ signal: expect.any(AbortSignal) }),
    );
  });

  it("reports the binding gap rather than a search failure", async () => {
    renderWithClient(
      <Harness
        worktree={{
          worktreeId: undefined,
          resolving: false,
          gap: "no-main-worktree",
          whenResolved: async () => undefined,
        }}
      />,
    );

    await waitFor(() =>
      expect(screen.getByTestId("status")).toHaveTextContent("no-main-worktree"),
    );
    expect(listRepoFiles).not.toHaveBeenCalled();
  });

  it("maps a 404 to a missing worktree and a 409/503 to an unconfigured repo", async () => {
    listRepoFiles.mockRejectedValueOnce(new FakeApiError(404));
    const missing = renderWithClient(<Harness worktree={worktreeStub()} />);
    await waitFor(() =>
      expect(screen.getByTestId("status")).toHaveTextContent("worktree-missing"),
    );
    missing.unmount();

    listRepoFiles.mockResolvedValueOnce(null);
    renderWithClient(<Harness worktree={worktreeStub({ worktreeId: "wt-2" })} />);
    await waitFor(() =>
      expect(screen.getByTestId("status")).toHaveTextContent("no-repo"),
    );
  });

  it("distinguishes an empty repository from a query with no matches", async () => {
    listRepoFiles.mockResolvedValue(fileList([]));
    const emptyRepo = renderWithClient(<Harness worktree={worktreeStub()} />);
    await waitFor(() =>
      expect(screen.getByTestId("status")).toHaveTextContent("empty-repo"),
    );
    emptyRepo.unmount();

    listRepoFiles.mockResolvedValue(fileList(["a.ts"]));
    renderWithClient(
      <Harness worktree={worktreeStub({ worktreeId: "wt-3" })} query="zzznope" />,
    );
    await waitFor(() =>
      expect(screen.getByTestId("status")).toHaveTextContent("ready"),
    );
    expect(screen.getByTestId("count")).toHaveTextContent("0");
  });

  it("settles the menu when the request throws an unexpected error", async () => {
    listRepoFiles.mockRejectedValue(new Error("boom"));
    renderWithClient(<Harness worktree={worktreeStub()} />);

    await waitFor(() =>
      expect(screen.getByTestId("loading")).toHaveTextContent("settled"),
    );
    expect(screen.getByTestId("status")).toHaveTextContent("failed");
  });

  it("drops a status that belonged to a previous worktree", async () => {
    listRepoFiles.mockResolvedValue(null);
    const { rerender } = renderWithClient(<Harness worktree={worktreeStub()} />);

    await waitFor(() =>
      expect(screen.getByTestId("status")).toHaveTextContent("no-repo"),
    );

    rerender(<Harness worktree={worktreeStub({ worktreeId: "wt-other" })} />);
    expect(screen.getByTestId("status")).toHaveTextContent("idle");
  });

  it("offers every match without a cap", async () => {
    listRepoFiles.mockResolvedValue(
      fileList(Array.from({ length: 250 }, (_, i) => `src/widget-${i}.ts`)),
    );
    renderWithClient(<Harness worktree={worktreeStub()} query="widget" />);

    await waitFor(() =>
      expect(screen.getByTestId("count")).toHaveTextContent("250"),
    );
  });
});
