import { fireEvent, render } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";

type Block = {
  id: string;
  type: string;
  content?: unknown;
  children?: unknown[];
};

let document_: Block[] = [];
let hovered: Block | undefined;

const openSuggestionMenu = vi.fn();
const setTextCursorPosition = vi.fn();
const insertBlocks = vi.fn((blocks: { type: string }[], reference: Block) => {
  // BlockNote normalizes a bare `{ type: "paragraph" }` into a full block, so
  // the double must too — without `content: []` it reads back as contentless.
  const inserted = {
    id: `inserted-${document_.length}`,
    content: [],
    ...blocks[0],
  };
  document_.splice(
    document_.findIndex((block) => block.id === reference.id) + 1,
    0,
    inserted,
  );
  return [inserted];
});

vi.mock("@blocknote/core/extensions", () => ({
  SideMenuExtension: { key: "sideMenu" },
  SuggestionMenu: { key: "suggestionMenu" },
}));

vi.mock("@blocknote/react", () => ({
  useComponentsContext: () => ({
    SideMenu: {
      Button: ({
        label,
        onClick,
        icon,
      }: {
        label: string;
        onClick?: () => void;
        icon?: React.ReactNode;
      }) => (
        <button type="button" aria-label={label} onClick={onClick}>
          {icon}
        </button>
      ),
    },
  }),
  useDictionary: () => ({ side_menu: { add_block_label: "Add block" } }),
  useBlockNoteEditor: () => ({
    getNextBlock: (block: Block) =>
      document_[document_.findIndex((b) => b.id === block.id) + 1],
    insertBlocks,
    setTextCursorPosition,
  }),
  useExtension: () => ({ openSuggestionMenu }),
  useExtensionState: (
    _extension: unknown,
    ctx: { selector: (state: unknown) => unknown },
  ) => ctx.selector({ block: hovered }),
}));

import { PromptEditorAddBlockButton } from "./PromptEditorAddBlockButton";

function clickAddBlock() {
  const view = render(<PromptEditorAddBlockButton />);
  const button = view.getByLabelText("Add block");
  // The handler must live on the button: upstream binds it to the icon, which
  // breaks keyboard activation and the stylesheet's expanded hit area.
  fireEvent.click(button);
  view.unmount();
}

beforeEach(() => {
  vi.clearAllMocks();
  hovered = { id: "a", type: "paragraph", content: [{ type: "text" }] };
  document_ = [hovered];
});

describe("PromptEditorAddBlockButton", () => {
  it("inserts a paragraph on the first click below a non-empty block", () => {
    clickAddBlock();

    expect(insertBlocks).toHaveBeenCalledTimes(1);
    expect(setTextCursorPosition).toHaveBeenCalledWith(
      expect.objectContaining({ type: "paragraph" }),
    );
    expect(openSuggestionMenu).toHaveBeenCalledWith("/");
  });

  it("reuses that paragraph instead of stacking another on repeat clicks", () => {
    clickAddBlock();
    clickAddBlock();
    clickAddBlock();

    expect(insertBlocks).toHaveBeenCalledTimes(1);
    expect(document_).toHaveLength(2);
    expect(setTextCursorPosition).toHaveBeenLastCalledWith(
      expect.objectContaining({ id: "inserted-1" }),
    );
  });

  it("focuses the hovered block when it is already empty", () => {
    hovered = { id: "a", type: "paragraph", content: [] };
    document_ = [hovered];

    clickAddBlock();

    expect(insertBlocks).not.toHaveBeenCalled();
    expect(setTextCursorPosition).toHaveBeenCalledWith(hovered);
  });

  it("inserts rather than reusing an empty heading below", () => {
    document_ = [hovered!, { id: "b", type: "heading", content: [] }];

    clickAddBlock();

    expect(insertBlocks).toHaveBeenCalledTimes(1);
  });

  it("renders nothing while no block is hovered", () => {
    hovered = undefined;

    const view = render(<PromptEditorAddBlockButton />);

    expect(view.queryByLabelText("Add block")).toBeNull();
  });
});
