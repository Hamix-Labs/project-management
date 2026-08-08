import { render, waitFor } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import {
  PromptEditorSideMenuController,
  promptSideMenuPositionKey,
} from "./PromptEditorSideMenuController";

let sideMenuMounts = 0;

vi.mock("@blocknote/react", async () => {
  const React = await vi.importActual<typeof import("react")>("react");

  return {
    SideMenuController: () => {
      React.useState(() => {
        sideMenuMounts += 1;
        return 0;
      });
      return <div data-testid="side-menu-controller" />;
    },
    useExtensionState: () => "block-1:0:0:100:24",
  };
});

function rect(init: DOMRectInit) {
  return DOMRect.fromRect(init);
}

describe("promptSideMenuPositionKey", () => {
  it("changes when the same block moves", () => {
    const first = promptSideMenuPositionKey({
      show: true,
      block: { id: "block-1" },
      referencePos: rect({ x: 10, y: 20, width: 120, height: 28 }),
    });
    const moved = promptSideMenuPositionKey({
      show: true,
      block: { id: "block-1" },
      referencePos: rect({ x: 10, y: 84, width: 120, height: 28 }),
    });

    expect(moved).not.toBe(first);
  });
});

describe("PromptEditorSideMenuController", () => {
  it("reanchors the side menu after a prompt-editor drag finishes", async () => {
    const editorHost = document.createElement("div");
    document.body.appendChild(editorHost);
    sideMenuMounts = 0;

    try {
      render(<PromptEditorSideMenuController editorHost={editorHost} />);

      await waitFor(() => expect(sideMenuMounts).toBe(1));

      editorHost.dispatchEvent(new Event("dragstart", { bubbles: true }));
      editorHost.dispatchEvent(new Event("drop", { bubbles: true }));

      await waitFor(() => expect(sideMenuMounts).toBeGreaterThan(1));
    } finally {
      editorHost.remove();
    }
  });
});
