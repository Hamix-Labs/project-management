import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { act, renderHook, waitFor } from "@testing-library/react";
import type { ReactNode } from "react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { taskQueryKeys } from "../../../task-query";
import { makeTask } from "@/test/taskDefaults";
import {
  type BulkCloseResult,
  useBulkCloseMutation,
} from "./useBulkCloseMutation";
import {
  type BulkScheduleResult,
  useBulkScheduleMutation,
} from "./useBulkScheduleMutation";

const { mockCloseTask, mockPatchTask } = vi.hoisted(() => ({
  mockCloseTask: vi.fn(),
  mockPatchTask: vi.fn(),
}));

vi.mock("@/api", async (importOriginal) => {
  const actual = await importOriginal<typeof import("@/api")>();
  return {
    ...actual,
    closeTask: mockCloseTask,
    patchTask: mockPatchTask,
  };
});

import { closeTask, patchTask } from "@/api";

const mockedClose = vi.mocked(closeTask);
const mockedPatch = vi.mocked(patchTask);

function deferred<T>() {
  let resolve!: (value: T) => void;
  const promise = new Promise<T>((res) => {
    resolve = res;
  });
  return { promise, resolve };
}

function makeWrapper() {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false }, mutations: { retry: false } },
  });
  function Wrapper({ children }: { children: ReactNode }) {
    return (
      <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
    );
  }
  return { Wrapper, queryClient };
}

describe("useBulkCloseMutation", () => {
  beforeEach(() => {
    mockedClose.mockReset();
    mockedClose.mockImplementation(async (id: string) =>
      makeTask({ id, status: "closed" }),
    );
  });

  it("does not call the API or invalidate when selection is empty", async () => {
    const { Wrapper, queryClient } = makeWrapper();
    const inv = vi.spyOn(queryClient, "invalidateQueries");

    const { result } = renderHook(() => useBulkCloseMutation(), {
      wrapper: Wrapper,
    });

    await act(async () => {
      await result.current.run([]);
    });

    expect(mockedClose).not.toHaveBeenCalled();
    expect(inv).not.toHaveBeenCalled();
  });

  it("still invalidates task queries when some closes fail (partial success)", async () => {
    const { Wrapper, queryClient } = makeWrapper();
    const inv = vi.spyOn(queryClient, "invalidateQueries");

    mockedClose
      .mockResolvedValueOnce(makeTask({ id: "ok-id", status: "closed" }))
      .mockRejectedValueOnce(new Error("server no"));

    const { result } = renderHook(() => useBulkCloseMutation(), {
      wrapper: Wrapper,
    });

    await act(async () => {
      await result.current.run(["ok-id", "bad-id"]);
    });

    expect(inv.mock.calls.map((c) => c[0]?.queryKey)).toEqual([
      taskQueryKeys.listRoot(),
      taskQueryKeys.stats(),
    ]);
    expect(result.current.lastResult).toMatchObject({
      attempted: 2,
      succeeded: 1,
      failed: [{ taskId: "bad-id" }],
    });
  });
});

describe("useBulkScheduleMutation", () => {
  beforeEach(() => {
    mockedPatch.mockReset();
  });

  it("still invalidates when some patches fail", async () => {
    const { Wrapper, queryClient } = makeWrapper();
    const inv = vi.spyOn(queryClient, "invalidateQueries");

    mockedPatch
      .mockResolvedValueOnce({} as never)
      .mockRejectedValueOnce(new Error("conflict"));

    const { result } = renderHook(() => useBulkScheduleMutation(), {
      wrapper: Wrapper,
    });

    await act(async () => {
      await result.current.run(["a", "b"], "2026-01-01T00:00:00Z");
    });

    expect(inv.mock.calls.map((c) => c[0]?.queryKey)).toEqual([
      taskQueryKeys.listRoot(),
      taskQueryKeys.stats(),
    ]);
  });
});

describe("useBulkScheduleMutation overlapping runs", () => {
  beforeEach(() => {
    mockedPatch.mockReset();
  });

  /**
   * Same contract as bulk close: shared `isPending` + in-flight ref.
   * Split `act` boundaries for React 18 flush semantics (see close overlap test).
   */
  it("keeps isPending true until every overlapping bulk run has finished", async () => {
    const dSlow = deferred<void>();
    const { Wrapper } = makeWrapper();
    const when = "2026-01-01T00:00:00Z";

    mockedPatch.mockImplementation(async (id: string) => {
      if (id === "slow") {
        await dSlow.promise;
        return {} as never;
      }
      return {} as never;
    });

    const { result } = renderHook(() => useBulkScheduleMutation(), {
      wrapper: Wrapper,
    });

    let slowDone = false;
    let pSlow!: Promise<BulkScheduleResult>;

    await act(() => {
      pSlow = result.current.run(["slow"], when).finally(() => {
        slowDone = true;
      });
    });

    await waitFor(() =>
      expect(mockedPatch).toHaveBeenCalledWith("slow", {
        pickup_not_before: when,
      }),
    );
    expect(result.current.isPending).toBe(true);

    await act(async () => {
      await result.current.run(["fast"], when);
    });
    expect(slowDone).toBe(false);
    expect(result.current.isPending).toBe(true);

    await act(async () => {
      dSlow.resolve(undefined);
      await pSlow;
    });
    expect(slowDone).toBe(true);
    expect(result.current.isPending).toBe(false);
  });
});

describe("useBulkCloseMutation overlapping runs", () => {
  beforeEach(() => {
    mockedClose.mockReset();
  });

  /**
   * Overlapping `run()` calls must not clear `isPending` until all complete.
   * Avoid one long `await act(async () => …)` around the overlap: React 18 defers
   * state flushes until that act callback returns, so `isPending` would still
   * read `false` mid-flight.
   */
  it("keeps isPending true until every overlapping bulk run has finished", async () => {
    const dSlow = deferred<void>();
    const { Wrapper } = makeWrapper();

    mockedClose.mockImplementation(async (id: string) => {
      if (id === "slow") {
        await dSlow.promise;
      }
      return makeTask({ id, status: "closed" });
    });

    const { result } = renderHook(() => useBulkCloseMutation(), {
      wrapper: Wrapper,
    });

    let slowDone = false;
    let pSlow!: Promise<BulkCloseResult>;

    await act(() => {
      pSlow = result.current.run(["slow"]).finally(() => {
        slowDone = true;
      });
    });

    await waitFor(() => expect(mockedClose).toHaveBeenCalledWith("slow"));
    expect(result.current.isPending).toBe(true);

    await act(async () => {
      await result.current.run(["fast"]);
    });
    expect(slowDone).toBe(false);
    expect(result.current.isPending).toBe(true);

    await act(async () => {
      dSlow.resolve(undefined);
      await pSlow;
    });
    expect(slowDone).toBe(true);
    expect(result.current.isPending).toBe(false);
  });
});
