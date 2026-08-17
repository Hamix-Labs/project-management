import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";
import {
  COMPOSE_BRIEF_PLACEHOLDER,
  TaskComposeBriefCard,
} from "./TaskComposeBriefCard";
import { COMPOSE_BRIEF_EDITOR_MIN_PX } from "./useComposeBriefVerticalResize";

vi.mock("@/components/rich-prompt", async () => {
  const actual = await vi.importActual<typeof import("@/components/rich-prompt")>(
    "@/components/rich-prompt",
  );
  return {
    ...actual,
    RichPromptEditor: ({ placeholder }: { placeholder?: string }) => (
      <div className="rich-prompt-wrap">
        <div
          className="tiptap ProseMirror rich-prompt-editor notion-like-editor"
          data-testid="compose-brief-editor"
          data-placeholder={placeholder}
        />
      </div>
    ),
  };
});

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
    expect(screen.getByTestId("compose-brief-editor")).toHaveAttribute(
      "data-placeholder",
      COMPOSE_BRIEF_PLACEHOLDER,
    );
    expect(
      screen.queryByPlaceholderText(/describe the full brief/i),
    ).not.toBeInTheDocument();
  });

  it("keeps the live word count in the title row", () => {
    renderCard();
    expect(screen.getByText("0 words")).toBeInTheDocument();
  });
});

describe("TaskComposeBriefCard resize", () => {
  it("renders a pointer-only corner grip on the brief card", () => {
    const { container } = renderCard();
    const handle = grip(container);
    expect(handle).toHaveAttribute("aria-hidden", "true");
    expect(handle.querySelectorAll("path")).toHaveLength(2);
  });

  it("grows the editor height downward when the grip is dragged", () => {
    const { container } = renderCard();
    dispatchPointer(grip(container), "pointerdown", 400);
    dispatchPointer(grip(container), "pointermove", 520);
    dispatchPointer(grip(container), "pointerup", 520);
    expect(briefCard(container).style.getPropertyValue("--compose-brief-editor-h")).toBe(
      `${COMPOSE_BRIEF_EDITOR_MIN_PX + 120}px`,
    );
  });

  it("does not shrink the editor below the 320px minimum", () => {
    const { container } = renderCard();
    dispatchPointer(grip(container), "pointerdown", 400);
    dispatchPointer(grip(container), "pointermove", 100);
    dispatchPointer(grip(container), "pointerup", 100);
    expect(briefCard(container).style.getPropertyValue("--compose-brief-editor-h")).toBe(
      `${COMPOSE_BRIEF_EDITOR_MIN_PX}px`,
    );
  });
});

describe("TaskComposeBriefCard focused writing", () => {
  it("expands, closes on Done, and restores focus to Expand", async () => {
    const user = userEvent.setup();
    renderCard();
    const expand = screen.getByRole("button", { name: "Expand" });
    await user.click(expand);
    expect(screen.getByRole("dialog", { name: "Editing brief" })).toBeInTheDocument();
    await user.click(screen.getByRole("button", { name: /done/i }));
    expect(screen.queryByRole("dialog")).not.toBeInTheDocument();
    expect(expand).toHaveFocus();
  });

  it("closes on Escape", async () => {
    const user = userEvent.setup();
    renderCard();
    await user.click(screen.getByRole("button", { name: "Expand" }));
    await user.keyboard("{Escape}");
    expect(screen.queryByRole("dialog")).not.toBeInTheDocument();
  });
});
