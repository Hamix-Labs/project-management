import { describe, expect, it } from "vitest";
import { parseAppSettings } from "./settings";
import { DEFAULT_VERIFY_MAX_RETRIES } from "@/types/task";

const requiredSettings = {
  runner: "cursor",
  cursor_bin: "cursor-agent",
  cursor_model: "",
  max_run_duration_seconds: 600,
  stream_idle_stuck_seconds: 120,
  agent_pickup_delay_seconds: 5,
};

describe("parseAppSettings", () => {
  it("defaults verify_max_retries to DEFAULT_VERIFY_MAX_RETRIES when omitted", () => {
    const settings = parseAppSettings(requiredSettings);
    expect(settings.verify_max_retries).toBe(DEFAULT_VERIFY_MAX_RETRIES);
  });

  it("preserves an explicit verify_max_retries value", () => {
    const settings = parseAppSettings({
      ...requiredSettings,
      verify_max_retries: 4,
    });
    expect(settings.verify_max_retries).toBe(4);
  });
});
