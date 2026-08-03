import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { render, screen, waitFor } from "@testing-library/react";
import type { ReactNode } from "react";
import { describe, expect, it } from "vitest";
import type { TaskCycle } from "@/types/cycle";
import { ensureMswListening } from "@/test/mswLifecycle";
import { server } from "@/test/server";
import { taskTokenUsageGet } from "@/test/handlers/cycles";
import { CycleHistoryList } from "./CycleHistoryList";

ensureMswListening();

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

const baseCycle: TaskCycle = {
  id: "cyc-1",
  task_id: "task-1",
  attempt_seq: 1,
  status: "succeeded",
  started_at: "2026-01-01T12:00:00.000Z",
  ended_at: "2026-01-01T12:05:00.000Z",
  triggered_by: "user",
  meta: {},
  cycle_meta: {
    runner: "",
    runner_version: "",
    cursor_model: "",
    cursor_model_effective: "",
    prompt_hash: "",
  },
};

describe("CycleHistoryList token usage", () => {
  it("shows attempt tokens and share of task when known", async () => {
    server.use(
      taskTokenUsageGet("task-1", {
        task_id: "task-1",
        token_usage: {
          consumed_tokens: 10000,
          execute_consumed_tokens: 8000,
          verify_consumed_tokens: 2000,
          input_tokens: 0,
          output_tokens: 0,
          cache_read_tokens: 0,
          cache_write_tokens: 0,
          known: true,
        },
        attempts: [
          {
            cycle_id: "cyc-1",
            attempt_seq: 1,
            token_usage: {
              consumed_tokens: 12400,
              execute_consumed_tokens: 10000,
              verify_consumed_tokens: 2400,
              input_tokens: 0,
              output_tokens: 0,
              cache_read_tokens: 0,
              cache_write_tokens: 0,
              known: true,
            },
            share_of_task_pct: 15.4,
          },
        ],
      }),
    );

    render(
      <CycleHistoryList
        taskId="task-1"
        cycles={[
          {
            ...baseCycle,
            token_usage: {
              consumed_tokens: 12400,
              execute_consumed_tokens: 10000,
              verify_consumed_tokens: 2400,
              input_tokens: 0,
              output_tokens: 0,
              cache_read_tokens: 0,
              cache_write_tokens: 0,
              known: true,
            },
          },
        ]}
      />,
      { wrapper: createWrapper() },
    );

    const summary = await screen.findByTestId("task-cycle-row-tokens");
    await waitFor(() => {
      expect(summary).toHaveTextContent("12.4K · 15.4% of task");
    });
  });

  it("omits the token summary when attempt usage is unknown", async () => {
    server.use(
      taskTokenUsageGet("task-1", {
        task_id: "task-1",
        token_usage: {
          consumed_tokens: 0,
          execute_consumed_tokens: 0,
          verify_consumed_tokens: 0,
          input_tokens: 0,
          output_tokens: 0,
          cache_read_tokens: 0,
          cache_write_tokens: 0,
          known: false,
        },
        attempts: [],
      }),
    );

    render(
      <CycleHistoryList
        taskId="task-1"
        cycles={[baseCycle]}
      />,
      { wrapper: createWrapper() },
    );

    await screen.findByText(/Attempt #1/);
    expect(screen.queryByTestId("task-cycle-row-tokens")).not.toBeInTheDocument();
  });
});
