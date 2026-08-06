import { act, renderHook, waitFor } from "@testing-library/react";
import { describe, expect, it, vi, beforeEach } from "vitest";
import { usePromptEditorDocumentLoad } from "./usePromptEditorDocumentLoad";
import type { PromptDocumentAdapter } from "./types";
import {
  FACTORY_GIT_REPO_ID,
  FACTORY_GIT_WORKTREE_ID,
  gitRepositoryFactory,
  gitWorktreeDetailFactory,
} from "@/test/factories/git";

vi.mock("@/api/gitGlobal", () => ({
  getGlobalGitWorktree: vi.fn(),
  getGlobalGitRepository: vi.fn(),
}));

import {
  getGlobalGitRepository,
  getGlobalGitWorktree,
} from "@/api/gitGlobal";

const mockGetWorktree = vi.mocked(getGlobalGitWorktree);
const mockGetRepository = vi.mocked(getGlobalGitRepository);

describe("usePromptEditorDocumentLoad", () => {
  beforeEach(() => {
    mockGetWorktree.mockReset();
    mockGetRepository.mockReset();
  });

  it("commits snapshot and reaches ready", async () => {
    const adapter: PromptDocumentAdapter = {
      load: vi.fn(async () => ({ html: "<p>hello there friend</p>" })),
      save: vi.fn(),
    };
    const onCommit = vi.fn();
    const onLoadError = vi.fn();
    const onStatus = vi.fn();
    const dirtyRef = { current: false };

    const { result } = renderHook(() =>
      usePromptEditorDocumentLoad({
        adapter,
        launch: null,
        loadNonce: 0,
        dirtyRef,
        onCommit,
        onLoadError,
        onStatus,
      }),
    );

    await waitFor(() => expect(onStatus).toHaveBeenCalledWith("ready"));
    expect(onCommit).toHaveBeenCalledWith({
      html: "<p>hello there friend</p>",
      worktreeId: undefined,
      repositoryId: undefined,
    });
    expect(onLoadError).not.toHaveBeenCalled();
    expect(result.current.repoLabel).toBe("No repo");
  });

  it("prefers non-empty seedHtml over adapter html", async () => {
    const adapter: PromptDocumentAdapter = {
      load: vi.fn(async () => ({ html: "<p>server</p>" })),
      save: vi.fn(),
    };
    const onCommit = vi.fn();

    renderHook(() =>
      usePromptEditorDocumentLoad({
        adapter,
        launch: { seedHtml: "<p>seeded content here</p>" },
        loadNonce: 0,
        dirtyRef: { current: false },
        onCommit,
        onLoadError: vi.fn(),
        onStatus: vi.fn(),
      }),
    );

    await waitFor(() =>
      expect(onCommit).toHaveBeenCalledWith({
        html: "<p>seeded content here</p>",
        worktreeId: undefined,
        repositoryId: undefined,
      }),
    );
  });

  it("settles on error when load fails", async () => {
    const adapter: PromptDocumentAdapter = {
      load: vi.fn(async () => {
        throw new Error("draft unavailable");
      }),
      save: vi.fn(),
    };
    const onLoadError = vi.fn();
    const onStatus = vi.fn();

    renderHook(() =>
      usePromptEditorDocumentLoad({
        adapter,
        launch: null,
        loadNonce: 0,
        dirtyRef: { current: false },
        onCommit: vi.fn(),
        onLoadError,
        onStatus,
      }),
    );

    await waitFor(() => expect(onStatus).toHaveBeenCalledWith("error"));
    expect(onLoadError).toHaveBeenCalled();
    expect(onLoadError.mock.calls[0][0].phase).toBe("load");
    expect(onLoadError.mock.calls[0][0].detail).toContain("draft unavailable");
  });

  it("ignores stale load after remount and settles on the latest", async () => {
    let resolveFirst!: (v: { html: string }) => void;
    const first = new Promise<{ html: string }>((r) => {
      resolveFirst = r;
    });
    const load = vi
      .fn()
      .mockImplementationOnce(() => first)
      .mockImplementationOnce(async () => ({ html: "<p>second</p>" }));

    const adapter: PromptDocumentAdapter = { load, save: vi.fn() };
    const onCommit = vi.fn();
    const onStatus = vi.fn();

    const { rerender, unmount } = renderHook(
      ({ nonce }: { nonce: number }) =>
        usePromptEditorDocumentLoad({
          adapter,
          launch: null,
          loadNonce: nonce,
          dirtyRef: { current: false },
          onCommit,
          onLoadError: vi.fn(),
          onStatus,
        }),
      { initialProps: { nonce: 0 } },
    );

    await act(async () => {
      rerender({ nonce: 1 });
    });

    await waitFor(() =>
      expect(onCommit).toHaveBeenCalledWith({
        html: "<p>second</p>",
        worktreeId: undefined,
        repositoryId: undefined,
      }),
    );

    await act(async () => {
      resolveFirst({ html: "<p>first-stale</p>" });
    });

    expect(onCommit).toHaveBeenCalledTimes(1);
    expect(onStatus).toHaveBeenCalledWith("ready");
    unmount();
  });

  it("resolves repo label from repositoryId when worktree is unset", async () => {
    mockGetRepository.mockResolvedValue(
      gitRepositoryFactory({ path: "/repos/my-app" }),
    );
    const adapter: PromptDocumentAdapter = {
      load: vi.fn(async () => ({ html: "<p>brief</p>" })),
      save: vi.fn(),
    };

    const { result } = renderHook(() =>
      usePromptEditorDocumentLoad({
        adapter,
        launch: null,
        loadNonce: 0,
        dirtyRef: { current: false },
        onCommit: vi.fn(),
        onLoadError: vi.fn(),
        onStatus: vi.fn(),
        repositoryId: FACTORY_GIT_REPO_ID,
      }),
    );

    await waitFor(() => expect(result.current.repoLabel).toBe("my-app repo"));
    expect(mockGetRepository).toHaveBeenCalledWith(FACTORY_GIT_REPO_ID);
    expect(mockGetWorktree).not.toHaveBeenCalled();
  });

  it("prefers worktree label when both worktree and repository are set", async () => {
    mockGetWorktree.mockResolvedValue(
      gitWorktreeDetailFactory({
        repository_path: "/repos/from-worktree",
      }),
    );
    mockGetRepository.mockResolvedValue(
      gitRepositoryFactory({ path: "/repos/from-repo" }),
    );
    const adapter: PromptDocumentAdapter = {
      load: vi.fn(async () => ({ html: "<p>brief</p>" })),
      save: vi.fn(),
    };

    const { result } = renderHook(() =>
      usePromptEditorDocumentLoad({
        adapter,
        launch: null,
        loadNonce: 0,
        dirtyRef: { current: false },
        onCommit: vi.fn(),
        onLoadError: vi.fn(),
        onStatus: vi.fn(),
        worktreeId: FACTORY_GIT_WORKTREE_ID,
        repositoryId: FACTORY_GIT_REPO_ID,
      }),
    );

    await waitFor(() =>
      expect(result.current.repoLabel).toBe("from-worktree repo"),
    );
    expect(mockGetWorktree).toHaveBeenCalledWith(FACTORY_GIT_WORKTREE_ID);
    expect(mockGetRepository).not.toHaveBeenCalled();
  });
});
