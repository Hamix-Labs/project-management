import { describe, expect, it } from "vitest";
import {
  effectiveVerifyChatMode,
  verifyChatModeLabel,
  verifyChatModeSource,
} from "./verifyChatModeDisplay";

describe("effectiveVerifyChatMode", () => {
  it("uses task override when set", () => {
    expect(effectiveVerifyChatMode("different_chat", "same_chat")).toBe(
      "different_chat",
    );
    expect(effectiveVerifyChatMode("same_chat", "different_chat")).toBe(
      "same_chat",
    );
  });

  it("inherits settings when task mode is empty", () => {
    expect(effectiveVerifyChatMode("", "different_chat")).toBe(
      "different_chat",
    );
    expect(effectiveVerifyChatMode(undefined, "different_chat")).toBe(
      "different_chat",
    );
    expect(effectiveVerifyChatMode("  ", "same_chat")).toBe("same_chat");
  });

  it("defaults to same_chat when both empty or invalid", () => {
    expect(effectiveVerifyChatMode("", "")).toBe("same_chat");
    expect(effectiveVerifyChatMode("nope", "also-nope")).toBe("same_chat");
  });
});

describe("verifyChatModeLabel", () => {
  it("matches create/settings copy", () => {
    expect(verifyChatModeLabel("same_chat")).toBe("Same chat");
    expect(verifyChatModeLabel("different_chat")).toBe("Different chat");
  });
});

describe("verifyChatModeSource", () => {
  it("reports task vs workspace", () => {
    expect(verifyChatModeSource("same_chat")).toBe("task");
    expect(verifyChatModeSource("")).toBe("workspace");
    expect(verifyChatModeSource(undefined)).toBe("workspace");
  });
});
