import { render, screen } from "@testing-library/react";
import type { ReactNode } from "react";
import { describe, expect, it, vi } from "vitest";
import {
  COMPOSE_BRIEF_PLACEHOLDER,
  TaskComposeBriefCard,
} from "./TaskComposeBriefCard";
import { COMPOSE_BRIEF_EDITOR_MIN_PX } from "./useComposeBriefVerticalResize";

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
    <div className="rich-prompt-wrap">
      <div
        className="tiptap ProseMirror rich-prompt-editor"
        data-testid="compose-brief-editor"
        data-placeholder={placeholder}
        data-menu-variant={menuVariant}
      >
        {menuRight}
      </div>
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

function grip(container: HTMLElement): HTMLElement {
  const el = container.querySelector(".compose-brief__resize-grip");
  if (!(el instanceof HTMLElement)) {
    throw new Error("expected compose brief resize grip");
  }
  return el;
}

function briefCard(container: HTMLElement): HTMLElement {
  const el = container.querySelector(".compose-brief");
  if (!(el instanceof HTMLElement)) {
    throw new Error("expected compose brief card");
  }
  return el;
}

function dispatchPointer(el: HTMLElement, type: string, clientY: number) {
  // jsdom has no PointerEvent; MouseEvent still carries clientY for React.
  el.dispatchEvent(
    new MouseEvent(type, {
      bubbles: true,
      cancelable: true,
      clientY,
      button: 0,
      buttons: type === "pointerup" ? 0 : 1,
    }),
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

  it("presents the icon markdown toolbar and keeps the word count", () => {
    renderCard();
    const editor = screen.getByTestId("compose-brief-editor");
    expect(editor).toHaveAttribute("data-menu-variant", "icons");
    expect(screen.getByText("0 words")).toBeInTheDocument();
  });
});

describe("TaskComposeBriefCard resize", () => {
  it("renders a pointer-only corner grip on the brief card", () => {
    const { container } = renderCard();
    const handle = grip(container);
    expect(handle).toHaveAttribute("aria-hidden", "true");
    expect(handle.querySelector("path")).not.toBeNull();
  });

  it("grows the editor height downward when the grip is dragged", () => {
    const { container } = renderCard();
    const handle = grip(container);
    const card = briefCard(container);

    dispatchPointer(handle, "pointerdown", 400);
    dispatchPointer(handle, "pointermove", 520);
    dispatchPointer(handle, "pointerup", 520);

    expect(card.style.getPropertyValue("--compose-brief-editor-h")).toBe(
      `${COMPOSE_BRIEF_EDITOR_MIN_PX + 120}px`,
    );
    expect(card.style.width).toBe("");
  });

  it("does not shrink the editor below the 320px minimum", () => {
    const { container } = renderCard();
    const handle = grip(container);
    const card = briefCard(container);

    dispatchPointer(handle, "pointerdown", 400);
    dispatchPointer(handle, "pointermove", 100);
    dispatchPointer(handle, "pointerup", 100);

    expect(card.style.getPropertyValue("--compose-brief-editor-h")).toBe(
      `${COMPOSE_BRIEF_EDITOR_MIN_PX}px`,
    );
  });
});
