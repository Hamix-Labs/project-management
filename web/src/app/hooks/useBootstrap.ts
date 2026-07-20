import { useQueryClient } from "@tanstack/react-query";
import { useEffect, useState } from "react";
import { fetchBootstrap } from "@/api";
import { seedBootstrapCache } from "@/tasks/sync/seedBootstrapCache";

/**
 * Cold-start cache seeding — see docs/web.md §Cold start (sync policy exception).
 *
 * @returns `true` once bootstrap has settled (seeded, unavailable, or failed)
 * so list/stats/settings queries can start without racing the aggregate GET.
 */
export function useBootstrap(): boolean {
  const queryClient = useQueryClient();
  const [settled, setSettled] = useState(false);

  useEffect(() => {
    const controller = new AbortController();

    void (async () => {
      try {
        const payload = await fetchBootstrap({ signal: controller.signal });
        if (controller.signal.aborted) return;
        if (payload) {
          seedBootstrapCache(queryClient, payload);
        }
      } catch (err) {
        if (controller.signal.aborted) return;
        if (
          import.meta.env.DEV &&
          !(err instanceof DOMException && err.name === "AbortError")
        ) {
          console.warn("[bootstrap] aggregate fetch failed", err);
        }
      } finally {
        // Abort (Strict Mode remount / unmount) must not leave queries gated.
        if (!controller.signal.aborted) {
          setSettled(true);
        }
      }
    })();

    return () => controller.abort();
  }, [queryClient]);

  return settled;
}
