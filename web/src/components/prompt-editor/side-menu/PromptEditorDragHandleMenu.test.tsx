import { render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

vi.mock("./promptDragHandleMenuOpenContext", () => ({
  usePromptDragHandleMenuOpenChange: () => vi.fn(),
}));

vi.mock("./PromptEditorTurnIntoItem", () => ({
  PromptEditorTurnIntoItem: () => <div data-testid="turn-into">Turn into</div>,
}));

vi.mock("@blocknote/react", () => ({
  DragHandleMenu: ({ children }: { children?: React.ReactNode }) => (
    <div data-testid="drag-handle-menu">{children}</div>
  ),
  RemoveBlockItem: ({ children }: { children?: React.ReactNode }) => (
    <div data-testid="delete-item">{children}</div>
  ),
  BlockColorsItem: ({ children }: { children?: React.ReactNode }) => (
    <div data-testid="colors-item">{children}</div>
  ),
  useDictionary: () => ({
    drag_handle: {
      delete_menuitem: "Delete",
      colors_menuitem: "Colors",
    },
  }),
}));

import { PromptEditorDragHandleMenu } from "./PromptEditorDragHandleMenu";

describe("PromptEditorDragHandleMenu", () => {
  it("places Turn into between Delete and Colors without dropping the open beacon", () => {
    render(<PromptEditorDragHandleMenu />);

    const menu = screen.getByTestId("drag-handle-menu");
    const labels = Array.from(menu.children).map(
      (node) => node.getAttribute("data-testid") ?? node.textContent,
    );

    expect(labels).toEqual([
      "delete-item",
      "turn-into",
      "colors-item",
    ]);
  });
});
