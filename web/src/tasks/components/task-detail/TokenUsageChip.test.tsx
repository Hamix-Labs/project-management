import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import type { ReactNode } from "react";
import { describe, expect, it } from "vitest";
import { server } from "@/test/server";
import { taskTokenUsageGet } from "@/test/handlers/cycles";
import { TokenUsageChip } from "./TokenUsageChip";

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

describe("TokenUsageChip", () => {
  it("shows a compact consumed count and opens the breakdown modal", async () => {
    server.use(
      taskTokenUsageGet("task-1", {
        task_id: "task-1",
        token_usage: {
          consumed_tokens: 8200,
          execute_consumed_tokens: 7000,
          verify_consumed_tokens: 1200,
          input_tokens: 5000,
          output_tokens: 3200,
          cache_read_tokens: 0,
          cache_write_tokens: 0,
          known: true,
        },
        attempts: [],
      }),
    );

    render(<TokenUsageChip taskId="task-1" />, { wrapper: createWrapper() });

    const chip = await screen.findByTestId("task-token-usage-chip");
    await waitFor(() => {
      expect(chip).toHaveTextContent("8.2K");
    });

    await userEvent.click(chip);
    expect(screen.getByRole("dialog")).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "Token usage" })).toBeInTheDocument();
    expect(screen.getByText("Execute agent")).toBeInTheDocument();
    expect(screen.getByText("Verify agent")).toBeInTheDocument();
    expect(screen.getByText("Total")).toBeInTheDocument();
  });

  it("shows a muted placeholder when usage is unknown", async () => {
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

    render(<TokenUsageChip taskId="task-1" />, { wrapper: createWrapper() });

    const chip = await screen.findByTestId("task-token-usage-chip");
    await waitFor(() => {
      expect(chip).toHaveTextContent("Tokens —");
    });
    await userEvent.click(chip);
    await waitFor(() => {
      expect(screen.getByRole("dialog")).toBeInTheDocument();
    });
  });
});
