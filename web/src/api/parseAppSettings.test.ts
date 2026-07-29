import { describe, expect, it } from "vitest";
import { parseAppSettings } from "./settings";

const requiredSettings = {
  runner: "cursor",
  cursor_bin: "cursor-agent",
  cursor_model: "",
  max_run_duration_seconds: 600,
  agent_task_parallelism: 150,
  agent_pickup_delay_seconds: 5,
};

describe("parseAppSettings", () => {
  it("parses required settings fields", () => {
    const settings = parseAppSettings(requiredSettings);
    expect(settings.runner).toBe("cursor");
    expect(settings.verify_model).toBe("");
    expect(settings.max_run_duration_seconds).toBe(600);
  });
});
