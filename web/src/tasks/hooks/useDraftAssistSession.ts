import { useCallback, useEffect, useRef, useState } from "react";
import {
  createSession as apiCreateSession,
  deleteSession as apiDeleteSession,
} from "@/api/draftAssist";
import type {
  DraftAssistSession,
  DraftAssistSnapshot,
} from "@/types/draftAssist";

/** Lifecycle phase surfaced to callers. */
export type DraftAssistSessionStatus =
  | "idle"
  | "creating"
  | "ready"
  | "error"
  | "closed";

export type UseDraftAssistSessionReturn = {
  session: DraftAssistSession | null;
  status: DraftAssistSessionStatus;
  error: Error | null;
  /** Idempotent within a compose session: first call POSTs `/sessions`; later calls are no-ops. */
  open: (snapshot: DraftAssistSnapshot, worktreeId?: string) => Promise<DraftAssistSession>;
  /** Aborts an in-flight create and DELETEs the session (best effort). */
  close: () => void;
};

/**
 * Owns one draft-assist session per compose modal. First `open(snapshot)`
 * POSTs `/draft-assist/sessions` and caches `{id, nonce}`. Subsequent
 * `open` calls return the cached session so callers can trigger the
 * stream lazily without racing a second POST.
 *
 * On unmount (or explicit `close()`): aborts an in-flight create, then
 * fires `DELETE /draft-assist/sessions/{id}` best-effort. Failures are
 * swallowed — the server GC-s idle sessions anyway.
 */
export function useDraftAssistSession(): UseDraftAssistSessionReturn {
  const [session, setSession] = useState<DraftAssistSession | null>(null);
  const [status, setStatus] = useState<DraftAssistSessionStatus>("idle");
  const [error, setError] = useState<Error | null>(null);

  // Refs mirror the reactive state so cleanup and idempotent-open both
  // see the latest values without re-running the effect closure.
  const sessionRef = useRef<DraftAssistSession | null>(null);
  const pendingRef = useRef<Promise<DraftAssistSession> | null>(null);
  const createAbortRef = useRef<AbortController | null>(null);
  const closedRef = useRef(false);

  const open = useCallback(
    async (
      snapshot: DraftAssistSnapshot,
      worktreeId?: string,
    ): Promise<DraftAssistSession> => {
      if (closedRef.current) {
        throw new Error("draft-assist session is closed");
      }
      if (sessionRef.current) {
        return sessionRef.current;
      }
      if (pendingRef.current) {
        return pendingRef.current;
      }
      const controller = new AbortController();
      createAbortRef.current = controller;
      setStatus("creating");
      setError(null);
      const p = apiCreateSession(
        { snapshot, ...(worktreeId ? { worktree_id: worktreeId } : {}) },
        { signal: controller.signal },
      )
        .then((created) => {
          if (closedRef.current) {
            // Race: caller closed while POST was in-flight. Fire-and-forget
            // the DELETE so we do not leak a session on the server.
            void apiDeleteSession(created.id).catch(() => {});
            throw new Error("draft-assist session was closed before create resolved");
          }
          sessionRef.current = created;
          setSession(created);
          setStatus("ready");
          return created;
        })
        .catch((e: unknown) => {
          const err = e instanceof Error ? e : new Error(String(e));
          if (!closedRef.current) {
            setError(err);
            setStatus("error");
          }
          throw err;
        })
        .finally(() => {
          if (pendingRef.current === p) {
            pendingRef.current = null;
          }
          if (createAbortRef.current === controller) {
            createAbortRef.current = null;
          }
        });
      pendingRef.current = p;
      return p;
    },
    [],
  );

  const close = useCallback(() => {
    if (closedRef.current) return;
    closedRef.current = true;
    createAbortRef.current?.abort();
    const cached = sessionRef.current;
    sessionRef.current = null;
    setStatus("closed");
    if (cached) {
      void apiDeleteSession(cached.id).catch(() => {});
    }
  }, []);

  useEffect(() => {
    return () => {
      if (closedRef.current) return;
      closedRef.current = true;
      createAbortRef.current?.abort();
      const cached = sessionRef.current;
      sessionRef.current = null;
      if (cached) {
        void apiDeleteSession(cached.id).catch(() => {});
      }
    };
  }, []);

  return { session, status, error, open, close };
}
