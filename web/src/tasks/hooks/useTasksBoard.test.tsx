import { describe, expect, it } from "vitest";
import { renderHook, waitFor } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import type { ReactNode } from "react";
import { http, HttpResponse } from "msw";
import { useTasksBoard } from "./useTasksBoard";
import { server } from "@/test/server";
import { makeTask } from "@/test/taskDefaults";

function wrapper(client: QueryClient) {
  return function W({ children }: { children: ReactNode }) {
    return (
      <QueryClientProvider client={client}>{children}</QueryClientProvider>
    );
  };
}

describe("useTasksBoard", () => {
  it("does not fetch when view is list", () => {
    let hits = 0;
    server.use(
      http.get("*/tasks", () => {
        hits += 1;
        return HttpResponse.json({
          tasks: [],
          limit: 0,
          offset: 0,
          has_more: false,
        });
      }),
    );
    const client = new QueryClient({
      defaultOptions: { queries: { retry: false } },
    });
    const { result } = renderHook(
      () => useTasksBoard({ view: "list", dataEnabled: true }),
      { wrapper: wrapper(client) },
    );
    expect(result.current.loading).toBe(false);
    expect(result.current.tasks).toEqual([]);
    expect(hits).toBe(0);
  });

  it("fetches when view is board", async () => {
    server.use(
      http.get("*/tasks", () =>
        HttpResponse.json({
          tasks: [makeTask({ id: "a1", status: "ready" })],
          limit: 1,
          offset: 0,
          has_more: false,
        }),
      ),
    );
    const client = new QueryClient({
      defaultOptions: { queries: { retry: false } },
    });
    const { result } = renderHook(
      () => useTasksBoard({ view: "board", dataEnabled: true }),
      { wrapper: wrapper(client) },
    );
    await waitFor(() => expect(result.current.loading).toBe(false));
    expect(result.current.tasks).toHaveLength(1);
    expect(result.current.tasks[0]?.id).toBe("a1");
  });
});
