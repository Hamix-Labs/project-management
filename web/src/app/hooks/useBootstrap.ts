import { useQueryClient } from "@tanstack/react-query";
import { useEffect, useRef } from "react";
import { fetchBootstrap } from "@/api";
import { seedBootstrapCache } from "@/tasks/sync/seedBootstrapCache";

/** Cold-start cache seeding — see docs/web.md §Cold start (sync policy exception). */
export function useBootstrap(): void {
  const queryClient = useQueryClient();
  const hasRunRef = useRef(false);

  useEffect(() => {
    if (hasRunRef.current) return;
    hasRunRef.current = true;
    const controller = new AbortController();

    void (async () => {
      try {
        const payload = await fetchBootstrap({ signal: controller.signal });
        if (controller.signal.aborted) return;
        if (!payload) return;
        seedBootstrapCache(queryClient, payload);
      } catch (err) {
        if (
          import.meta.env.DEV &&
          !(err instanceof DOMException && err.name === "AbortError")
        ) {
          console.warn("[bootstrap] aggregate fetch failed", err);
        }
      }
    })();

    return () => controller.abort();
  }, [queryClient]);
}
