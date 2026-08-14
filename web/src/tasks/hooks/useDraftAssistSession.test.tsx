import { act, renderHook, waitFor } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import { useDraftAssistSession } from "./useDraftAssistSession";
import { ensureMswListening } from "@/test/mswLifecycle";
import { server } from "@/test/server";
import {
  draftAssistCreateSession,
  draftAssistCreateSessionCapture,
  draftAssistDeleteSessionCapture,
  makeDraftAssistSessionFixture,
} from "@/test/handlers/draftAssist";
import { beforeEach } from "vitest";
import { HttpResponse, http } from "msw";

ensureMswListening("bypass");

/**
 * DELETE is fire-and-forget from the hook's useEffect cleanup. Its fetch
 * is issued after the test body finishes; per-test `server.use` handlers
 * are cleared by `resetHandlers` before React's unmount lands the DELETE,
 * so we re-register a catch-all at the top of each test to keep MSW's
 * "no handler" logs off stderr without hiding the create path.
 */
beforeEach(() => {
  server.use(
    http.delete(
      /\/draft-assist\/sessions\/[^/]+$/,
      () => new HttpResponse(null, { status: 204 }),
    ),
  );
});

describe("useDraftAssistSession", () => {
  it("starts idle with no session", () => {
    const { result } = renderHook(() => useDraftAssistSession());
    expect(result.current.session).toBeNull();
    expect(result.current.status).toBe("idle");
    expect(result.current.error).toBeNull();
  });

  it("POSTs /sessions on first open and exposes id + nonce", async () => {
    let bodyRaw = "";
    const fixture = makeDraftAssistSessionFixture({ id: "sess-lazy" });
    server.use(
      draftAssistCreateSessionCapture((b) => {
        bodyRaw = b;
      }, fixture),
      http.delete("/draft-assist/sessions/sess-lazy", () =>
        new HttpResponse(null, { status: 204 }),
      ),
    );
    const { result } = renderHook(() => useDraftAssistSession());

    await act(async () => {
      await result.current.open({ prompt: "hello" }, "wt-x");
    });

    expect(result.current.session?.id).toBe("sess-lazy");
    expect(result.current.session?.nonce).toBe(fixture.nonce);
    expect(result.current.status).toBe("ready");
    const parsed = JSON.parse(bodyRaw) as {
      worktree_id: string;
      snapshot: { prompt?: string };
    };
    expect(parsed.worktree_id).toBe("wt-x");
    expect(parsed.snapshot.prompt).toBe("hello");
  });

  it("is idempotent: second open returns the cached session (only one POST)", async () => {
    const fixture = makeDraftAssistSessionFixture({ id: "sess-once" });
    let posts = 0;
    server.use(
      http.post("/draft-assist/sessions", () => {
        posts += 1;
        return HttpResponse.json(fixture, { status: 201 });
      }),
      http.delete(`/draft-assist/sessions/${fixture.id}`, () =>
        new HttpResponse(null, { status: 204 }),
      ),
    );
    const { result } = renderHook(() => useDraftAssistSession());
    let first: unknown;
    let second: unknown;
    await act(async () => {
      first = await result.current.open({ prompt: "a" });
    });
    await act(async () => {
      second = await result.current.open({ prompt: "b" });
    });
    expect(first).toBe(second);
    expect(posts).toBe(1);
  });

  it("dedupes concurrent open() calls to a single POST", async () => {
    const fixture = makeDraftAssistSessionFixture({ id: "sess-dedupe" });
    let posts = 0;
    server.use(
      http.post("/draft-assist/sessions", async () => {
        posts += 1;
        return HttpResponse.json(fixture, { status: 201 });
      }),
      http.delete(`/draft-assist/sessions/${fixture.id}`, () =>
        new HttpResponse(null, { status: 204 }),
      ),
    );
    const { result } = renderHook(() => useDraftAssistSession());
    await act(async () => {
      const [a, b] = await Promise.all([
        result.current.open({ prompt: "a" }),
        result.current.open({ prompt: "b" }),
      ]);
      expect(a).toBe(b);
    });
    expect(posts).toBe(1);
  });

  it("close() DELETEs the session and flips status to closed", async () => {
    const fixture = makeDraftAssistSessionFixture({ id: "sess-close" });
    const onDelete = vi.fn();
    server.use(
      draftAssistCreateSession(fixture),
      draftAssistDeleteSessionCapture(fixture.id, onDelete),
    );
    const { result } = renderHook(() => useDraftAssistSession());
    await act(async () => {
      await result.current.open({});
    });
    act(() => {
      result.current.close();
    });
    expect(result.current.status).toBe("closed");
    await waitFor(() => {
      expect(onDelete).toHaveBeenCalledTimes(1);
    });
  });

  it("unmount DELETEs the session best-effort", async () => {
    const fixture = makeDraftAssistSessionFixture({ id: "sess-unmount" });
    const onDelete = vi.fn();
    server.use(
      draftAssistCreateSession(fixture),
      draftAssistDeleteSessionCapture(fixture.id, onDelete),
    );
    const { result, unmount } = renderHook(() => useDraftAssistSession());
    await act(async () => {
      await result.current.open({});
    });
    unmount();
    await waitFor(() => {
      expect(onDelete).toHaveBeenCalledTimes(1);
    });
  });

  it("captures errors on create failure without leaving pending state", async () => {
    server.use(
      http.post("/draft-assist/sessions", () =>
        HttpResponse.json({ error: "boom" }, { status: 500 }),
      ),
      http.delete(/\/draft-assist\/sessions\/.+/, () =>
        new HttpResponse(null, { status: 204 }),
      ),
    );
    const { result } = renderHook(() => useDraftAssistSession());
    await act(async () => {
      await expect(result.current.open({})).rejects.toBeInstanceOf(Error);
    });
    expect(result.current.status).toBe("error");
    expect(result.current.error).not.toBeNull();
  });
});
