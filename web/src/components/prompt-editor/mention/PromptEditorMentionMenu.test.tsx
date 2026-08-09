import type { SuggestionMenuProps } from "@blocknote/react";
import { render, screen } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import {
  PromptEditorMentionMenu,
  type PromptFileMentionItem,
} from "./PromptEditorMentionMenu";

const menuHeight = 297;
const menuWidth = 340;

/**
 * jsdom performs no layout and ships no ResizeObserver, so the virtualizer
 * would measure the scroller as 0×0 and decide no row fits on screen. Give it
 * the size the stylesheet gives it.
 */
beforeEach(() => {
  vi.stubGlobal(
    "ResizeObserver",
    class {
      observe() {}
      unobserve() {}
      disconnect() {}
    },
  );
  Object.defineProperty(HTMLElement.prototype, "offsetHeight", {
    configurable: true,
    get: () => menuHeight,
  });
  Object.defineProperty(HTMLElement.prototype, "offsetWidth", {
    configurable: true,
    get: () => menuWidth,
  });
});

afterEach(() => {
  vi.unstubAllGlobals();
  Reflect.deleteProperty(HTMLElement.prototype, "offsetHeight");
  Reflect.deleteProperty(HTMLElement.prototype, "offsetWidth");
});

function makeItems(count: number, prefix = "src/widget"): PromptFileMentionItem[] {
  return Array.from({ length: count }, (_, i) => ({
    title: `${prefix}-${i}.ts`,
    query: "widget",
    onItemClick: vi.fn(),
  }));
}

function renderMenu(overrides: Partial<SuggestionMenuProps<PromptFileMentionItem>>) {
  const props = {
    items: [],
    selectedIndex: 0,
    onItemClick: vi.fn(),
    loadingState: "loaded",
    ...overrides,
  } as unknown as SuggestionMenuProps<PromptFileMentionItem>;
  return { ...render(<PromptEditorMentionMenu {...props} />), props };
}

describe("PromptEditorMentionMenu", () => {
  it("reports the full match count rather than a capped page", () => {
    renderMenu({ items: makeItems(1234) });

    expect(screen.getByText("1,234")).toBeInTheDocument();
  });

  it("gives every match scroll room while rendering only a window of rows", () => {
    renderMenu({ items: makeItems(500) });

    const list = screen.getByRole("listbox", { name: "Repository files" });
    const viewport = list.firstElementChild as HTMLElement;

    // Scrollable height covers all 500 rows, so nothing is unreachable.
    expect(parseInt(viewport.style.height, 10)).toBeGreaterThan(500 * 30);
    // But the DOM holds only the visible window plus overscan.
    const rendered = screen.getAllByRole("option");
    expect(rendered.length).toBeGreaterThan(0);
    expect(rendered.length).toBeLessThan(60);
  });

  it("tells assistive tech the true size of the list", () => {
    renderMenu({ items: makeItems(400) });

    const [first] = screen.getAllByRole("option");
    expect(first).toHaveAttribute("aria-setsize", "400");
    expect(first).toHaveAttribute("aria-posinset", "1");
  });

  it("marks the selected row for keyboard navigation", () => {
    renderMenu({ items: makeItems(20), selectedIndex: 3 });

    const options = screen.getAllByRole("option");
    const selected = options.filter(
      (option) => option.getAttribute("aria-selected") === "true",
    );
    expect(selected).toHaveLength(1);
    expect(selected[0]).toHaveTextContent("widget-3.ts");
  });

  it("says nothing matched rather than blaming the worktree binding", () => {
    renderMenu({ items: [], loadingState: "loaded" });

    expect(screen.getByText("No files match that search.")).toBeInTheDocument();
    expect(screen.queryByText(/bind a worktree/i)).not.toBeInTheDocument();
  });

  it("shows the loading label while items are being fetched", () => {
    renderMenu({ items: [], loadingState: "loading-initial" });

    expect(screen.getByText("Searching files…")).toBeInTheDocument();
    expect(screen.queryByText(/No files match/)).not.toBeInTheDocument();
  });
});
