import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { beforeEach, describe, expect, it } from "vitest";
import { LAST_EDITOR_STORAGE_KEY } from "@/tasks/task-git/editors/lastEditorPreference";
import { OpenInEditorMenu } from "./OpenInEditorMenu";

describe("OpenInEditorMenu", () => {
  beforeEach(() => {
    localStorage.clear();
  });

  it("opens a menu with Cursor and VS Code deep links", async () => {
    const user = userEvent.setup();
    render(<OpenInEditorMenu openPath="/repo/main" />);

    expect(
      screen.queryByTestId("task-detail-open-in-menu"),
    ).not.toBeInTheDocument();

    await user.click(screen.getByTestId("task-detail-open-in-trigger"));

    const menu = screen.getByTestId("task-detail-open-in-menu");
    expect(menu).toBeInTheDocument();

    const cursor = screen.getByRole("menuitem", {
      name: /open worktree in cursor/i,
    });
    const vscode = screen.getByRole("menuitem", {
      name: /open worktree in vs code/i,
    });
    expect(cursor).toHaveAttribute(
      "href",
      "cursor://file/repo/main/?windowId=_blank",
    );
    expect(vscode).toHaveAttribute(
      "href",
      "vscode://file/repo/main/?windowId=_blank",
    );
  });

  it("persists the chosen editor and reorders on remount", async () => {
    const user = userEvent.setup();
    const { unmount } = render(<OpenInEditorMenu openPath="/repo/main" />);

    await user.click(screen.getByTestId("task-detail-open-in-trigger"));
    const vscodeItem = screen.getByRole("menuitem", {
      name: /open worktree in vs code/i,
    });
    // jsdom cannot follow custom protocol hrefs; keep click handlers under test.
    vscodeItem.addEventListener("click", (e) => e.preventDefault());
    await user.click(vscodeItem);

    expect(localStorage.getItem(LAST_EDITOR_STORAGE_KEY)).toBe("vscode");

    unmount();
    render(<OpenInEditorMenu openPath="/repo/main" />);
    await user.click(screen.getByTestId("task-detail-open-in-trigger"));

    const items = screen.getAllByRole("menuitem");
    expect(items[0]).toHaveAttribute("data-editor-id", "vscode");
    expect(items[0]).toHaveAttribute("data-preferred", "true");
    expect(items[1]).toHaveAttribute("data-editor-id", "cursor");
  });

  it("closes on Escape", async () => {
    const user = userEvent.setup();
    render(<OpenInEditorMenu openPath="/repo/main" />);

    await user.click(screen.getByTestId("task-detail-open-in-trigger"));
    expect(screen.getByTestId("task-detail-open-in-menu")).toBeInTheDocument();

    await user.keyboard("{Escape}");
    expect(
      screen.queryByTestId("task-detail-open-in-menu"),
    ).not.toBeInTheDocument();
  });
});
