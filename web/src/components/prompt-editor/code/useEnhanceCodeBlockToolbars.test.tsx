import { StrictMode, useState } from "react";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";
import { useEnhanceCodeBlockToolbars } from "./useEnhanceCodeBlockToolbars";

function CodeBlockHost({ disabled = false }: { disabled?: boolean }) {
  const [host, setHost] = useState<HTMLDivElement | null>(null);
  useEnhanceCodeBlockToolbars(host, disabled);
  return (
    <div ref={setHost} className="blocknote-prompt-editor">
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
  );
}

describe("useEnhanceCodeBlockToolbars", () => {
  it("mounts the language toolbar under StrictMode remounts", async () => {
    render(
      <StrictMode>
        <CodeBlockHost />
      </StrictMode>,
    );

    await waitFor(() => {
      expect(
        screen.getByRole("button", { name: /go/i }),
      ).toBeInTheDocument();
    });
    expect(
      screen.getByRole("button", { name: /copy code/i }),
    ).toBeInTheDocument();
  });

  it("changes the hidden select when a language is chosen", async () => {
    const user = userEvent.setup();
    const { container } = render(<CodeBlockHost />);

    await waitFor(() => {
      expect(
        screen.getByRole("button", { name: /go/i }),
      ).toBeInTheDocument();
    });

    await user.click(screen.getByRole("button", { name: /go/i }));
    await user.click(screen.getByRole("button", { name: /^typescript$/i }));

    const select = container.querySelector("select");
    expect(select).toBeInstanceOf(HTMLSelectElement);
    expect(select?.value).toBe("typescript");
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
