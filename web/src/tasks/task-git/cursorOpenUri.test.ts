import { describe, expect, it } from "vitest";
import { buildCursorOpenFolderUri } from "./cursorOpenUri";

describe("buildCursorOpenFolderUri", () => {
  it("builds a Windows drive path URI in a new window", () => {
    expect(
      buildCursorOpenFolderUri(
        String.raw`C:\Users\a\.hamix\repo\worktrees\hamix-task-x`,
      ),
    ).toBe(
      "cursor://file/C:/Users/a/.hamix/repo/worktrees/hamix-task-x/?windowId=_blank",
    );
  });

  it("builds a macOS absolute path URI in a new window", () => {
    expect(
      buildCursorOpenFolderUri("/Users/a/.hamix/repo/worktrees/hamix-task-x"),
    ).toBe(
      "cursor://file/Users/a/.hamix/repo/worktrees/hamix-task-x/?windowId=_blank",
    );
  });

  it("normalizes mixed separators and trailing slashes", () => {
    expect(buildCursorOpenFolderUri(String.raw`D:\wt\task\\`)).toBe(
      "cursor://file/D:/wt/task/?windowId=_blank",
    );
    expect(buildCursorOpenFolderUri("/tmp/wt///")).toBe(
      "cursor://file/tmp/wt/?windowId=_blank",
    );
  });

  it("rejects empty path", () => {
    expect(() => buildCursorOpenFolderUri("   ")).toThrow(/path is required/);
  });
});
