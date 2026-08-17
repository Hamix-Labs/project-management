import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { useRef, useState } from "react";
import { describe, expect, it } from "vitest";
import { PromptFocusFrame } from "./PromptFocusFrame";

function Harness() {
  const restoreFocusRef = useRef<HTMLButtonElement>(null);
  const [expanded, setExpanded] = useState(false);
  return (
    <>
      <button
        ref={restoreFocusRef}
        type="button"
        onClick={() => setExpanded(true)}
      >
        Expand
      </button>
      <PromptFocusFrame
        expanded={expanded}
        onExpandedChange={setExpanded}
        label="Editing brief"
        wordCount={3}
        restoreFocusRef={restoreFocusRef}
        title={<p>Title field</p>}
      >
        <p>Editor surface</p>
      </PromptFocusFrame>
    </>
  );
}

describe("PromptFocusFrame", () => {
  it("opens as a dialog and closes on Done, restoring focus to Expand", async () => {
    const user = userEvent.setup();
    render(<Harness />);
    await user.click(screen.getByRole("button", { name: "Expand" }));
    expect(screen.getByRole("dialog", { name: "Editing brief" })).toBeInTheDocument();
    expect(screen.getByText("3 words")).toBeInTheDocument();
    await user.click(screen.getByRole("button", { name: /done/i }));
    expect(screen.queryByRole("dialog")).not.toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Expand" })).toHaveFocus();
  });

  it("closes on Escape", async () => {
    const user = userEvent.setup();
    render(<Harness />);
    await user.click(screen.getByRole("button", { name: "Expand" }));
    await user.keyboard("{Escape}");
    expect(screen.queryByRole("dialog")).not.toBeInTheDocument();
  });
});
