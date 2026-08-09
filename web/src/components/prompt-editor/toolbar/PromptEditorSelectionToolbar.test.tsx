import { render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

// File and table buttons render null unless such a block is selected.
vi.mock("@blocknote/react", () => ({
  FormattingToolbar: ({ children }: { children: React.ReactNode }) => (
    <div data-testid="toolbar">{children}</div>
  ),
  ColorStyleButton: () => <button type="button">color</button>,
  CreateLinkButton: () => <button type="button">link</button>,
  BasicTextStyleButton: ({ basicTextStyle }: { basicTextStyle: string }) => (
    <button type="button">{basicTextStyle}</button>
  ),
  TableCellMergeButton: () => null,
  FileCaptionButton: () => null,
  FileReplaceButton: () => null,
  FileRenameButton: () => null,
  FileDeleteButton: () => null,
  FileDownloadButton: () => null,
  FilePreviewButton: () => null,
}));

vi.mock("./PromptEditorBlockTypeSelect", () => ({
  PromptEditorBlockTypeSelect: () => (
    <button type="button">Paragraph</button>
  ),
}));

import { PromptEditorSelectionToolbar } from "./PromptEditorSelectionToolbar";

function renderToolbar() {
  render(<PromptEditorSelectionToolbar />);
  return screen.getByTestId("toolbar");
}

describe("PromptEditorSelectionToolbar", () => {
  it("stacks the block type above the formatting controls", () => {
    const toolbar = renderToolbar();

    const rows = toolbar.querySelectorAll(".prompt-selection-toolbar__row");
    expect(rows).toHaveLength(3);
    expect(rows[0]).toHaveClass("prompt-selection-toolbar__row--block-type");
    expect(rows[0]).toContainElement(
      screen.getByRole("button", { name: "Paragraph" }),
    );
  });

  it("orders formatting controls to match the Notion selection panel", () => {
    const toolbar = renderToolbar();

    const formatRow = toolbar.querySelector(
      ".prompt-selection-toolbar__row--format",
    );
    const labels = Array.from(formatRow?.querySelectorAll("button") ?? []).map(
      (button) => button.textContent,
    );

    expect(labels).toEqual([
      "color",
      "bold",
      "italic",
      "underline",
      "link",
      "strike",
      "code",
    ]);
  });

  it("leaves the block action row empty for a text selection", () => {
    const toolbar = renderToolbar();

    const blockActions = toolbar.querySelector(
      ".prompt-selection-toolbar__row--block-actions",
    );
    expect(blockActions?.children).toHaveLength(0);
  });

  it("omits alignment, indent, and comment controls", () => {
    renderToolbar();

    for (const name of [/align/i, /indent/i, /comment/i]) {
      expect(screen.queryByRole("button", { name })).toBeNull();
    }
  });
});
