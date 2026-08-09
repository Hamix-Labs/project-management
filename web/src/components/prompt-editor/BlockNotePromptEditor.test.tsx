import type { ReactNode } from "react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
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
  ApiError: class ApiError extends Error {},
  listRepoFiles: vi.fn(async () => ({
    paths: [],
    truncated: false,
    source: "git" as const,
  })),
  repoQueryKeys: {
    files: (worktreeId: string) => ["repo", "files", worktreeId],
  },
}));

vi.mock("./code/useEnhanceCodeBlockToolbars", () => ({
  useEnhanceCodeBlockToolbars: () => undefined,
}));

import { BlockNotePromptEditor } from "./BlockNotePromptEditor";

function renderEditor(ui: ReactNode) {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  });
  return render(ui, {
    wrapper: ({ children }: { children: ReactNode }) => (
      <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
    ),
  });
}

describe("BlockNotePromptEditor emit", () => {
  it("calls onChange with HTML on the first content change after mount", () => {
    const onChange = vi.fn();
    viewOnChange = undefined;

    renderEditor(
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

    renderEditor(
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
