import { describe, expect, it } from "vitest";
import { buildCursorOpenFolderUri } from "./cursorOpenUri";

describe("buildCursorOpenFolderUri", () => {
  it("builds a Windows drive path URI", () => {
    expect(
      buildCursorOpenFolderUri(
        String.raw`C:\Users\a\.hamix\repo\worktrees\hamix-task-x`,
      ),
    ).toBe("cursor://file/C:/Users/a/.hamix/repo/worktrees/hamix-task-x/");
  });

  it("builds a macOS absolute path URI", () => {
    expect(
      buildCursorOpenFolderUri("/Users/a/.hamix/repo/worktrees/hamix-task-x"),
    ).toBe("cursor://file/Users/a/.hamix/repo/worktrees/hamix-task-x/");
  });

  it("normalizes mixed separators and trailing slashes", () => {
    expect(
      buildCursorOpenFolderUri(String.raw`D:\wt\task\\`),
    ).toBe("cursor://file/D:/wt/task/");
    expect(buildCursorOpenFolderUri("/tmp/wt///")).toBe("cursor://file/tmp/wt/");
  });

  it("rejects empty path", () => {
    expect(() => buildCursorOpenFolderUri("   ")).toThrow(/path is required/);
  });
});
