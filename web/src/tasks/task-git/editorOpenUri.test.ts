import { describe, expect, it } from "vitest";
import { buildEditorOpenFolderUri } from "./editorOpenUri";

describe("buildEditorOpenFolderUri", () => {
  it("builds a Windows drive path URI for cursor in a new window", () => {
    expect(
      buildEditorOpenFolderUri(
        String.raw`C:\Users\a\.hamix\repo\worktrees\hamix-task-x`,
        "cursor",
      ),
    ).toBe(
      "cursor://file/C:/Users/a/.hamix/repo/worktrees/hamix-task-x/?windowId=_blank",
    );
  });

  it("builds a Windows drive path URI for vscode in a new window", () => {
    expect(
      buildEditorOpenFolderUri(
        String.raw`C:\Users\a\.hamix\repo\worktrees\hamix-task-x`,
        "vscode",
      ),
    ).toBe(
      "vscode://file/C:/Users/a/.hamix/repo/worktrees/hamix-task-x/?windowId=_blank",
    );
  });

  it("builds a macOS absolute path URI for cursor in a new window", () => {
    expect(
      buildEditorOpenFolderUri(
        "/Users/a/.hamix/repo/worktrees/hamix-task-x",
        "cursor",
      ),
    ).toBe(
      "cursor://file/Users/a/.hamix/repo/worktrees/hamix-task-x/?windowId=_blank",
    );
  });

  it("builds a macOS absolute path URI for vscode in a new window", () => {
    expect(
      buildEditorOpenFolderUri(
        "/Users/a/.hamix/repo/worktrees/hamix-task-x",
        "vscode",
      ),
    ).toBe(
      "vscode://file/Users/a/.hamix/repo/worktrees/hamix-task-x/?windowId=_blank",
    );
  });

  it("normalizes mixed separators and trailing slashes", () => {
    expect(buildEditorOpenFolderUri(String.raw`D:\wt\task\\`, "cursor")).toBe(
      "cursor://file/D:/wt/task/?windowId=_blank",
    );
    expect(buildEditorOpenFolderUri("/tmp/wt///", "vscode")).toBe(
      "vscode://file/tmp/wt/?windowId=_blank",
    );
  });

  it("rejects empty path", () => {
    expect(() => buildEditorOpenFolderUri("   ", "cursor")).toThrow(
      /path is required/,
    );
  });

  it("rejects empty scheme", () => {
    expect(() => buildEditorOpenFolderUri("/tmp/wt", "  ")).toThrow(
      /scheme is required/,
    );
  });
});
