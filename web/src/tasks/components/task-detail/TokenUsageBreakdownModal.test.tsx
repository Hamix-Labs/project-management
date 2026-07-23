import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { TokenUsageBreakdownModal } from "./TokenUsageBreakdownModal";

describe("TokenUsageBreakdownModal", () => {
  it("renders execute, verify, and total rows with compact counts", () => {
    render(
      <TokenUsageBreakdownModal
        tokenUsage={{
          consumed_tokens: 8200,
          execute_consumed_tokens: 7000,
          verify_consumed_tokens: 1200,
          input_tokens: 5000,
          output_tokens: 3200,
          cache_read_tokens: 0,
          cache_write_tokens: 0,
          known: true,
        }}
        onClose={() => {}}
      />,
    );

    expect(screen.getByRole("heading", { name: "Token usage" })).toBeInTheDocument();
    expect(screen.getByText("Execute agent")).toBeInTheDocument();
    expect(screen.getByText("7K")).toBeInTheDocument();
    expect(screen.getByText("1.2K")).toBeInTheDocument();
    expect(screen.getByText("8.2K")).toBeInTheDocument();
  });
});
