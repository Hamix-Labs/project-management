import { useEffect, useState } from "react";
import { act, render, screen, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import type { FileWorktreeResolution } from "../usePromptEditorFileWorktree";
import { usePromptFileMentionSearch } from "./usePromptFileMentionSearch";

const { searchRepoFiles, FakeApiError } = vi.hoisted(() => ({
  searchRepoFiles: vi.fn(),
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
  maxRepoSearchQueryBytes: 512,
  searchRepoFiles,
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

describe("usePromptFileMentionSearch", () => {
  beforeEach(() => {
    searchRepoFiles.mockReset();
    searchRepoFiles.mockResolvedValue(["web/src/main.tsx"]);
  });

  it("issues exactly one request per query even though status changes re-render", async () => {
    const { rerender } = render(<Harness worktree={worktreeStub()} />);

    await waitFor(() =>
      expect(screen.getByTestId("loading")).toHaveTextContent("settled"),
    );
    await waitFor(() =>
      expect(screen.getByTestId("status")).toHaveTextContent("ready"),
    );

    // Writing status re-renders the host. An unstable getItems would restart
    // the load on every one of these and never stop.
    for (let i = 0; i < 5; i += 1) {
      rerender(<Harness worktree={worktreeStub()} />);
      await act(async () => {});
    }

    expect(searchRepoFiles).toHaveBeenCalledTimes(1);
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

    render(<IdentityProbe />);
    await waitFor(() =>
      expect(screen.getByTestId("status")).toHaveTextContent("ready"),
    );

    expect(seen.size).toBe(1);
  });

  it("waits for an in-flight worktree resolution instead of reporting a failure", async () => {
    let release: (worktreeId: string | undefined) => void = () => {};
    const pending = new Promise<string | undefined>((resolve) => {
      release = resolve;
    });

    render(
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
    expect(searchRepoFiles).not.toHaveBeenCalled();

    release("wt-late");
    await waitFor(() =>
      expect(screen.getByTestId("items")).toHaveTextContent("web/src/main.tsx"),
    );
    expect(searchRepoFiles).toHaveBeenCalledWith(
      "",
      expect.objectContaining({ worktreeId: "wt-late" }),
    );
  });

  it("reports the binding gap rather than a search failure", async () => {
    render(
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
    expect(searchRepoFiles).not.toHaveBeenCalled();
  });

  it("maps a 404 to a missing worktree and a 409/503 to an unconfigured repo", async () => {
    searchRepoFiles.mockRejectedValueOnce(new FakeApiError(404));
    const missing = render(<Harness worktree={worktreeStub()} />);
    await waitFor(() =>
      expect(screen.getByTestId("status")).toHaveTextContent("worktree-missing"),
    );
    missing.unmount();

    searchRepoFiles.mockResolvedValueOnce(null);
    render(<Harness worktree={worktreeStub({ worktreeId: "wt-2" })} />);
    await waitFor(() =>
      expect(screen.getByTestId("status")).toHaveTextContent("no-repo"),
    );
  });

  it("settles the menu instead of rejecting when the query is too long", async () => {
    render(<Harness worktree={worktreeStub()} query={"a".repeat(513)} />);

    await waitFor(() =>
      expect(screen.getByTestId("status")).toHaveTextContent("query-rejected"),
    );
    expect(screen.getByTestId("loading")).toHaveTextContent("settled");
    expect(searchRepoFiles).not.toHaveBeenCalled();
  });

  it("settles the menu when the request throws an unexpected error", async () => {
    searchRepoFiles.mockRejectedValue(new Error("boom"));
    render(<Harness worktree={worktreeStub()} />);

    await waitFor(() =>
      expect(screen.getByTestId("loading")).toHaveTextContent("settled"),
    );
    expect(screen.getByTestId("status")).toHaveTextContent("failed");
  });

  it("drops a status that belonged to a previous worktree", async () => {
    searchRepoFiles.mockResolvedValue(null);
    const { rerender } = render(<Harness worktree={worktreeStub()} />);

    await waitFor(() =>
      expect(screen.getByTestId("status")).toHaveTextContent("no-repo"),
    );

    rerender(<Harness worktree={worktreeStub({ worktreeId: "wt-other" })} />);
    expect(screen.getByTestId("status")).toHaveTextContent("idle");
  });
});
