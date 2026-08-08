import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";
import { PromptEditorDocHeader } from "./PromptEditorDocHeader";

describe("PromptEditorDocHeader", () => {
  it("commits on Enter and reverts on Escape", async () => {
    const user = userEvent.setup();
    const onTitleCommit = vi.fn();
    render(
      <PromptEditorDocHeader
        title="Original"
        editedLabel="Just now"
        wordCountLabel="3 words"
        repoLabel="repo"
        onTitleCommit={onTitleCommit}
      />,
    );

    const input = screen.getByRole("textbox", { name: /document title/i });
    await user.clear(input);
    await user.type(input, "Renamed");
    await user.keyboard("{Enter}");
    expect(onTitleCommit).toHaveBeenCalledWith("Renamed");

    onTitleCommit.mockClear();
    await user.clear(input);
    await user.type(input, "Scratch");
    await user.keyboard("{Escape}");
    expect(onTitleCommit).not.toHaveBeenCalled();
    expect(input).toHaveValue("Original");
  });

  it("rejects empty blur by restoring the committed title", async () => {
    const user = userEvent.setup();
    const onTitleCommit = vi.fn();
    render(
      <PromptEditorDocHeader
        title="Keep"
        editedLabel="Just now"
        wordCountLabel="1 word"
        repoLabel="repo"
        onTitleCommit={onTitleCommit}
      />,
    );
    const input = screen.getByRole("textbox", { name: /document title/i });
    await user.clear(input);
    await user.tab();
    expect(onTitleCommit).not.toHaveBeenCalled();
    expect(input).toHaveValue("Keep");
  });
});
