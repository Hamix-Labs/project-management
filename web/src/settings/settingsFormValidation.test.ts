import { describe, expect, it } from "vitest";
import type { SettingsFormState } from "./settingsForm";
import { parseSettingsNumericValidation } from "./settingsFormValidation";

function form(overrides: Partial<SettingsFormState> = {}): SettingsFormState {
  return {
    runner: "cursor",
    cursorBin: "",
    cursorModel: "",
    verifyModel: "",
    maxRunDurationSeconds: "3600",
    streamIdleStuckSeconds: "900",
    agentTaskParallelism: "150",
    agentPickupDelaySeconds: "0",
    displayTimezone: "",
    ...overrides,
  };
}

describe("parseSettingsNumericValidation", () => {
  it("returns no invalid flags when form is null", () => {
    expect(parseSettingsNumericValidation(null)).toEqual({
      maxInvalid: false,
      streamIdleInvalid: false,
      parallelismInvalid: false,
      pickupInvalid: false,
    });
  });

  it("flags negative max run duration", () => {
    expect(parseSettingsNumericValidation(form({ maxRunDurationSeconds: "-1" }))).toMatchObject({
      maxInvalid: true,
    });
  });

  it("flags negative stream idle timeout", () => {
    expect(
      parseSettingsNumericValidation(form({ streamIdleStuckSeconds: "-1" })),
    ).toMatchObject({
      streamIdleInvalid: true,
    });
  });

  it("flags parallelism below one", () => {
    expect(
      parseSettingsNumericValidation(form({ agentTaskParallelism: "0" })),
    ).toMatchObject({
      parallelismInvalid: true,
    });
  });

  it("flags pickup delay above one week", () => {
    expect(
      parseSettingsNumericValidation(form({ agentPickupDelaySeconds: "604801" })),
    ).toMatchObject({
      pickupInvalid: true,
    });
  });
});
