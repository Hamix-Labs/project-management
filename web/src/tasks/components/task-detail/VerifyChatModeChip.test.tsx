import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { settingsQueryKeys } from "@/lib/settingsQueryKeys";
import { testAppSettings } from "@/test/settingsDefaults";
import { VerifyChatModeChip } from "./VerifyChatModeChip";

function renderChip(taskMode: string | undefined, settingsMode: string) {
  const qc = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  });
  qc.setQueryData(settingsQueryKeys.app(), {
    ...testAppSettings,
    verify_chat_mode: settingsMode,
  });
  return render(
    <QueryClientProvider client={qc}>
      <VerifyChatModeChip task={{ verify_chat_mode: taskMode }} />
    </QueryClientProvider>,
  );
}

describe("VerifyChatModeChip", () => {
  it("shows task override label", () => {
    renderChip("different_chat", "same_chat");
    const chip = screen.getByTestId("task-verify-chat-mode-chip");
    expect(chip).toHaveAttribute("data-mode", "different_chat");
    expect(chip).toHaveAttribute("data-source", "task");
    expect(chip).toHaveTextContent("Start new chat");
  });

  it("inherits workspace default when task mode empty", () => {
    renderChip("", "different_chat");
    const chip = screen.getByTestId("task-verify-chat-mode-chip");
    expect(chip).toHaveAttribute("data-mode", "different_chat");
    expect(chip).toHaveAttribute("data-source", "workspace");
    expect(chip).toHaveTextContent("Start new chat");
  });
});
