import { act, render } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

const exportedHtml = "<p>first edit words here</p>";
let viewOnChange: (() => void) | undefined;

vi.mock("@blocknote/core/style.css", () => ({}));
vi.mock("@blocknote/ariakit/style.css", () => ({}));

vi.mock("./promptEditorHtml", () => ({
  htmlToInitialBlocks: () => ({
    blocks: [{ type: "paragraph", content: [] }],
    usedFallback: false,
  }),
}));

vi.mock("./blockNoteSchema", () => ({
  promptEditorSchema: {},
}));

vi.mock("@blocknote/react", () => ({
  useCreateBlockNote: () => ({
    document: [{ id: "b1", type: "paragraph" }],
    blocksToHTMLLossy: () => exportedHtml,
    insertInlineContent: vi.fn(),
    insertBlocks: vi.fn(),
    getTextCursorPosition: () => ({ block: { id: "b1" } }),
  }),
  SuggestionMenuController: () => null,
  FormattingToolbarController: () => null,
}));

vi.mock("./toolbar/PromptEditorSelectionToolbar", () => ({
  PromptEditorSelectionToolbar: () => null,
}));

vi.mock("@blocknote/ariakit", () => ({
  BlockNoteView: ({
    onChange,
  }: {
    onChange?: () => void;
  }) => {
    viewOnChange = onChange;
    return <div data-testid="mock-blocknote-view" />;
  },
}));

vi.mock("@/api", () => ({
  searchRepoFiles: vi.fn(async () => []),
}));

vi.mock("@/components/rich-prompt/useRepoWorkspaceProbe", () => ({
  useRepoWorkspaceProbe: () => "unavailable" as const,
}));

vi.mock("./code/useEnhanceCodeBlockToolbars", () => ({
  useEnhanceCodeBlockToolbars: () => undefined,
}));

import { BlockNotePromptEditor } from "./BlockNotePromptEditor";

describe("BlockNotePromptEditor emit", () => {
  it("calls onChange with HTML on the first content change after mount", () => {
    const onChange = vi.fn();
    viewOnChange = undefined;

    render(
      <BlockNotePromptEditor
        id="emit-test"
        initialHtml="<p>seed</p>"
        onChange={onChange}
      />,
    );

    expect(viewOnChange).toBeTypeOf("function");
    act(() => {
      viewOnChange?.();
    });

    expect(onChange).toHaveBeenCalledTimes(1);
    expect(onChange).toHaveBeenCalledWith(exportedHtml);
  });

  it("emits again on a second content change", () => {
    const onChange = vi.fn();
    viewOnChange = undefined;

    render(
      <BlockNotePromptEditor
        id="emit-test-2"
        initialHtml="<p>seed</p>"
        onChange={onChange}
      />,
    );

    act(() => {
      viewOnChange?.();
      viewOnChange?.();
    });

    expect(onChange).toHaveBeenCalledTimes(2);
  });
});
