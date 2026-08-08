import { StrictMode, useState } from "react";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";
import {
  useEnhanceCodeBlockToolbars,
  type CodeBlockLanguageEditor,
} from "./useEnhanceCodeBlockToolbars";

function CodeBlockHost({
  disabled = false,
  editor,
}: {
  disabled?: boolean;
  editor?: CodeBlockLanguageEditor;
}) {
  const [host, setHost] = useState<HTMLDivElement | null>(null);
  useEnhanceCodeBlockToolbars(host, disabled, editor);
  return (
    <div ref={setHost} className="blocknote-prompt-editor">
      <div className="bn-block" data-id="block-1">
        <div className="bn-block-content" data-content-type="codeBlock">
          <div contentEditable={false}>
            <select defaultValue="go">
              <option value="text">Plain Text</option>
              <option value="go">Go</option>
              <option value="typescript">TypeScript</option>
            </select>
          </div>
          <pre>
            <code>func main() {"{}"}</code>
          </pre>
        </div>
      </div>
    </div>
  );
}

/** Stands in for BlockNote: language lives in block props, not in the <select>. */
function fakeEditor(initialLanguage: string) {
  const languages = new Map([["block-1", initialLanguage]]);
  const listeners = new Set<() => void>();
  const updateBlock = vi.fn(
    (id: string, update: { props: { language: string } }) => {
      languages.set(id, update.props.language);
      listeners.forEach((listener) => listener());
    },
  );
  const editor: CodeBlockLanguageEditor = {
    getBlock: (id) => ({ props: { language: languages.get(id) } }),
    updateBlock,
    onChange: (callback) => {
      listeners.add(callback);
      return () => listeners.delete(callback);
    },
  };
  return { editor, updateBlock, languages };
}

describe("useEnhanceCodeBlockToolbars", () => {
  it("mounts the language toolbar under StrictMode remounts", async () => {
    render(
      <StrictMode>
        <CodeBlockHost />
      </StrictMode>,
    );

    await waitFor(() => {
      expect(screen.getByRole("button", { name: /go/i })).toBeInTheDocument();
    });
    expect(
      screen.getByRole("button", { name: /copy code/i }),
    ).toBeInTheDocument();
    expect(
      document.querySelector(
        ".blocknote-prompt-editor > .prompt-code-toolbar-root",
      ),
    ).toBeTruthy();
  });

  it("changes the hidden select when a language is chosen", async () => {
    const user = userEvent.setup();
    const { container } = render(<CodeBlockHost />);

    await waitFor(() => {
      expect(screen.getByRole("button", { name: /go/i })).toBeInTheDocument();
    });

    await user.click(screen.getByRole("button", { name: /go/i }));
    await user.click(screen.getByRole("button", { name: /^typescript$/i }));

    const select = container.querySelector("select");
    expect(select).toBeInstanceOf(HTMLSelectElement);
    expect(select?.value).toBe("typescript");
  });

  it("writes the language to the block props and follows the model", async () => {
    const user = userEvent.setup();
    const { editor, updateBlock, languages } = fakeEditor("go");
    render(<CodeBlockHost editor={editor} />);

    await waitFor(() => {
      expect(screen.getByRole("button", { name: /go/i })).toBeInTheDocument();
    });

    await user.click(screen.getByRole("button", { name: /go/i }));
    await user.click(screen.getByRole("button", { name: /^typescript$/i }));

    expect(updateBlock).toHaveBeenCalledWith("block-1", {
      props: { language: "typescript" },
    });
    expect(languages.get("block-1")).toBe("typescript");
    await waitFor(() => {
      expect(
        screen.getByRole("button", { name: /typescript/i }),
      ).toBeInTheDocument();
    });
  });

  it("keeps the label in sync when the block props change externally", async () => {
    const { editor } = fakeEditor("go");
    render(<CodeBlockHost editor={editor} />);

    await waitFor(() => {
      expect(screen.getByRole("button", { name: /go/i })).toBeInTheDocument();
    });

    editor.updateBlock("block-1", { props: { language: "text" } });

    await waitFor(() => {
      expect(
        screen.getByRole("button", { name: /plain text/i }),
      ).toBeInTheDocument();
    });
  });

  it("copies code text from the block", async () => {
    const user = userEvent.setup();
    const writeText = vi.fn().mockResolvedValue(undefined);
    vi.stubGlobal("navigator", {
      ...navigator,
      clipboard: { writeText },
    });

    render(<CodeBlockHost />);
    await waitFor(() => {
      expect(
        screen.getByRole("button", { name: /copy code/i }),
      ).toBeInTheDocument();
    });

    await user.click(screen.getByRole("button", { name: /copy code/i }));
    expect(writeText).toHaveBeenCalledWith("func main() {}");

    vi.unstubAllGlobals();
  });
});
