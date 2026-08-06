import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";
import { CodeLanguageToolbar } from "./CodeLanguageToolbar";
import type { PromptCodeLanguage } from "./promptCodeBlockOptions";

const LANGUAGES: PromptCodeLanguage[] = [
  { id: "go", name: "Go" },
  { id: "text", name: "Plain Text" },
  { id: "typescript", name: "TypeScript" },
];

describe("CodeLanguageToolbar", () => {
  it("shows the human language name for the current value", () => {
    render(
      <CodeLanguageToolbar
        languages={LANGUAGES}
        value="go"
        onChange={() => {}}
        onCopy={() => {}}
      />,
    );

    expect(
      screen.getByRole("button", { name: /go/i }),
    ).toBeInTheDocument();
  });

  it("falls back to Plain Text when value is empty", () => {
    render(
      <CodeLanguageToolbar
        languages={LANGUAGES}
        value=""
        onChange={() => {}}
        onCopy={() => {}}
      />,
    );

    expect(
      screen.getByRole("button", { name: /plain text/i }),
    ).toBeInTheDocument();
  });

  it("calls onChange with the selected language id", async () => {
    const user = userEvent.setup();
    const onChange = vi.fn();

    render(
      <CodeLanguageToolbar
        languages={LANGUAGES}
        value="text"
        onChange={onChange}
        onCopy={() => {}}
      />,
    );

    await user.click(screen.getByRole("button", { name: /plain text/i }));
    expect(
      screen.getByRole("listbox", { name: /languages/i }),
    ).toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: /^go$/i }));
    expect(onChange).toHaveBeenCalledWith("go");
  });

  it("calls onCopy when the copy button is clicked", async () => {
    const user = userEvent.setup();
    const onCopy = vi.fn();

    render(
      <CodeLanguageToolbar
        languages={LANGUAGES}
        value="go"
        onChange={() => {}}
        onCopy={onCopy}
      />,
    );

    await user.click(screen.getByRole("button", { name: /copy code/i }));
    expect(onCopy).toHaveBeenCalledTimes(1);
  });
});
