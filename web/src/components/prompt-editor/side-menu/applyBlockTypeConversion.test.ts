import { describe, expect, it, vi } from "vitest";
import { applyBlockTypeConversion } from "./applyBlockTypeConversion";
import { promptBlockTypeTargetByKey } from "./promptBlockTypeTargets";

function target(key: string) {
  const found = promptBlockTypeTargetByKey(key);
  if (!found) {
    throw new Error(`missing ${key}`);
  }
  return found;
}

describe("applyBlockTypeConversion", () => {
  it("does not open a transaction when every block already matches", () => {
    const transact = vi.fn();
    const updateBlock = vi.fn();
    const editor = {
      getSelection: () => undefined,
      transact,
      updateBlock,
      insertBlocks: vi.fn(),
    };

    applyBlockTypeConversion(editor, target("paragraph"), {
      id: "a",
      type: "paragraph",
      content: [],
    });

    expect(transact).not.toHaveBeenCalled();
    expect(updateBlock).not.toHaveBeenCalled();
  });

  it("converts every selected block inside one transaction", () => {
    const updateBlock = vi.fn();
    const editor = {
      getSelection: () => ({
        blocks: [
          {
            id: "a",
            type: "paragraph",
            content: [{ type: "text", text: "a", styles: {} }],
          },
          {
            id: "b",
            type: "paragraph",
            content: [{ type: "text", text: "b", styles: {} }],
          },
        ],
      }),
      transact: (fn: () => void) => fn(),
      updateBlock,
      insertBlocks: vi.fn(),
    };

    applyBlockTypeConversion(editor, target("heading_2"), {
      id: "a",
      type: "paragraph",
      content: [{ type: "text", text: "a", styles: {} }],
    });

    expect(updateBlock).toHaveBeenCalledTimes(2);
    expect(updateBlock).toHaveBeenNthCalledWith(
      1,
      expect.objectContaining({ id: "a" }),
      {
        type: "heading",
        props: { level: 2, isToggleable: false },
      },
    );
    expect(updateBlock).toHaveBeenNthCalledWith(
      2,
      expect.objectContaining({ id: "b" }),
      {
        type: "heading",
        props: { level: 2, isToggleable: false },
      },
    );
  });

  it("skips non-convertible blocks inside a multi-block selection", () => {
    const updateBlock = vi.fn();
    const editor = {
      getSelection: () => ({
        blocks: [
          {
            id: "a",
            type: "paragraph",
            content: [{ type: "text", text: "a", styles: {} }],
          },
          { id: "embed", type: "repoFileEmbed" },
        ],
      }),
      transact: (fn: () => void) => fn(),
      updateBlock,
      insertBlocks: vi.fn(),
    };

    applyBlockTypeConversion(editor, target("quote"), {
      id: "a",
      type: "paragraph",
      content: [{ type: "text", text: "a", styles: {} }],
    });

    expect(updateBlock).toHaveBeenCalledTimes(1);
    expect(updateBlock.mock.calls[0][0]).toEqual(
      expect.objectContaining({ id: "a" }),
    );
  });
});
