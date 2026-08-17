import { render, screen } from "@testing-library/react";
import type { ReactNode } from "react";
import { describe, expect, it, vi } from "vitest";
import {
  COMPOSE_BRIEF_PLACEHOLDER,
  TaskComposeBriefCard,
} from "./TaskComposeBriefCard";

vi.mock("@/components/rich-prompt", () => ({
  RichPromptEditor: ({
    placeholder,
    menuVariant,
    menuRight,
  }: {
    placeholder?: string;
    menuVariant?: string;
    menuRight?: ReactNode;
  }) => (
    <div
      data-testid="compose-brief-editor"
      data-placeholder={placeholder}
      data-menu-variant={menuVariant}
    >
      {menuRight}
    </div>
  ),
}));

function renderCard() {
  return render(
    <TaskComposeBriefCard
      idsPrefix="new"
      editorKey="k1"
      title=""
      prompt="<p></p>"
      disabled={false}
      onTitleChange={vi.fn()}
      onPromptChange={vi.fn()}
    />,
  );
}

describe("TaskComposeBriefCard", () => {
  it("uses slash-command empty-state copy instead of a markdown brief placeholder", () => {
    renderCard();
    const editor = screen.getByTestId("compose-brief-editor");
    expect(editor).toHaveAttribute("data-placeholder", COMPOSE_BRIEF_PLACEHOLDER);
    expect(editor).toHaveAttribute(
      "data-placeholder",
      "Press `/` for commands",
    );
    expect(
      screen.queryByPlaceholderText(/describe the full brief/i),
    ).not.toBeInTheDocument();
  });

  it("does not present the icon markdown toolbar on compose", () => {
    renderCard();
    const editor = screen.getByTestId("compose-brief-editor");
    expect(editor).toHaveAttribute("data-menu-variant", "none");
    expect(
      screen.queryByRole("toolbar", { name: /text formatting/i }),
    ).not.toBeInTheDocument();
    expect(
      screen.queryByRole("button", { name: /heading 2/i }),
    ).not.toBeInTheDocument();
    expect(screen.getByText("0 words")).toBeInTheDocument();
  });
});
