import type { Editor } from "@tiptap/core";
import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { RichPromptMenuBar } from "./RichPromptMenuBar";

const stubEditor = {} as Editor;

describe("RichPromptMenuBar", () => {
  it("renders nothing when editor is null", () => {
    const { container } = render(<RichPromptMenuBar editor={null} />);
    expect(container.firstChild).toBeNull();
  });

  it("hides formatting buttons when variant is none", () => {
    render(
      <RichPromptMenuBar
        editor={stubEditor}
        variant="none"
        right={<span>3 words</span>}
      />,
    );
    expect(
      screen.queryByRole("toolbar", { name: /text formatting/i }),
    ).not.toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "Bold" })).not.toBeInTheDocument();
    expect(
      screen.queryByRole("button", { name: "Heading 2" }),
    ).not.toBeInTheDocument();
    expect(screen.getByText("3 words")).toBeInTheDocument();
    expect(document.querySelector('[data-variant="none"]')).not.toBeNull();
  });

  it("renders nothing when variant is none and there is no trailing slot", () => {
    const { container } = render(
      <RichPromptMenuBar editor={stubEditor} variant="none" />,
    );
    expect(container.firstChild).toBeNull();
  });
});

