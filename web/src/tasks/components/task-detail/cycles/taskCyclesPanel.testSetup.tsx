import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { render } from "@testing-library/react";
import type { ReactNode } from "react";
import { TaskCyclesPanel } from "./TaskCyclesPanel";

/**
 * The panel composes useTaskCycles + useTaskCycle and renders five
 * distinct states:
 *   1. loading
 *   2. error
 *   3. empty (no cycles ever recorded)
 *   4. populated, no running cycle (history only, no live ticker)
 *   5. populated with running cycle (live ticker + history; phase
 *      detail fetched per row)
 *
 * We drive the states through fetch mocks rather than mocking the
 * hooks themselves so the parsing layer (api/cycles.ts +
 * parseTaskApi.ts) is exercised by the test too — protecting against
 * a parser regression silently breaking the panel.
 */

type FetchInput = Parameters<typeof fetch>[0];

function createWrapper() {
  const qc = new QueryClient({
    defaultOptions: {
      queries: { retry: false, gcTime: 0, staleTime: 0 },
    },
  });
  return function Wrapper({ children }: { children: ReactNode }) {
    return <QueryClientProvider client={qc}>{children}</QueryClientProvider>;
  };
}

export const okJSON = (body: unknown) =>
  new Response(JSON.stringify(body), {
    status: 200,
    headers: { "Content-Type": "application/json" },
  });

export const reqUrl = (input: FetchInput): string =>
  typeof input === "string"
    ? input
    : input instanceof URL
      ? input.toString()
      : (input as Request).url;

export function renderPanel(taskId = "task-1") {
  const Wrapper = createWrapper();
  return render(
    <Wrapper>
      <TaskCyclesPanel taskId={taskId} />
    </Wrapper>,
  );
}
