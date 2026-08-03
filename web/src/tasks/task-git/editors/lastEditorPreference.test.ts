// @vitest-environment jsdom
import { beforeEach, describe, expect, it } from "vitest";
import {
  editorsForMenu,
  getLastEditorId,
  LAST_EDITOR_STORAGE_KEY,
  setLastEditorId,
} from "./lastEditorPreference";

describe("lastEditorPreference", () => {
  beforeEach(() => {
    localStorage.clear();
  });

  it("defaults to cursor when unset", () => {
    expect(getLastEditorId()).toBe("cursor");
  });

  it("persists and reads a valid editor id", () => {
    setLastEditorId("vscode");
    expect(localStorage.getItem(LAST_EDITOR_STORAGE_KEY)).toBe("vscode");
    expect(getLastEditorId()).toBe("vscode");
  });

  it("ignores unknown stored ids", () => {
    localStorage.setItem(LAST_EDITOR_STORAGE_KEY, "notepad");
    expect(getLastEditorId()).toBe("cursor");
  });

  it("orders editors with preferred first", () => {
    const ordered = editorsForMenu("vscode");
    expect(ordered.map((e) => e.id)).toEqual(["vscode", "cursor"]);
  });
});
