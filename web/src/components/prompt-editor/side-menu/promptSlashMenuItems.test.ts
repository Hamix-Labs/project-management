import { describe, expect, it } from "vitest";
import { slashKeyToBlockTypeTarget } from "./promptSlashMenuItems";
import { PROMPT_SLASH_INSERT_ONLY_KEYS } from "./promptBlockTypeTargets";

describe("slashKeyToBlockTypeTarget", () => {
  it("maps catalog slash keys", () => {
    expect(slashKeyToBlockTypeTarget("paragraph")?.type).toBe("paragraph");
    expect(slashKeyToBlockTypeTarget("heading_2")?.props).toEqual({
      level: 2,
      isToggleable: false,
    });
    expect(slashKeyToBlockTypeTarget("code_block")?.type).toBe("codeBlock");
  });

  it("maps toggle headings and H4–H6 for in-place slash conversion", () => {
    expect(slashKeyToBlockTypeTarget("toggle_heading")?.props).toEqual({
      level: 1,
      isToggleable: true,
    });
    expect(slashKeyToBlockTypeTarget("heading_5")?.props).toEqual({
      level: 5,
      isToggleable: false,
    });
  });

  it("leaves insert-only keys unmapped so defaults keep inserting", () => {
    for (const key of PROMPT_SLASH_INSERT_ONLY_KEYS) {
      expect(slashKeyToBlockTypeTarget(key)).toBeUndefined();
    }
  });
});
