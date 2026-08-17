import { describe, expect, it } from "vitest";
import { PRIORITIES } from "@/types";
import { TEST_SCENARIOS, findTestScenarioById } from "./testScenarios";

describe("TEST_SCENARIOS catalog", () => {
  it("has the three sample-task ids in display order", () => {
    expect(TEST_SCENARIOS.map((s) => s.id)).toEqual([
      "observability",
      "flaky-test",
      "dep-upgrade",
    ]);
  });

  it("every scenario has a unique id", () => {
    const ids = TEST_SCENARIOS.map((s) => s.id);
    const unique = new Set(ids);
    expect(unique.size).toBe(ids.length);
  });

  it("every scenario picks a known Priority", () => {
    for (const scenario of TEST_SCENARIOS) {
      expect(PRIORITIES).toContain(scenario.priority);
    }
  });

  it("every scenario has non-empty name, title, description, prompt, and at least one criterion", () => {
    for (const scenario of TEST_SCENARIOS) {
      expect(scenario.name.trim()).not.toBe("");
      expect(scenario.title.trim()).not.toBe("");
      expect(scenario.description.trim()).not.toBe("");
      expect(scenario.prompt.trim()).not.toBe("");
      expect(scenario.criteria.length).toBeGreaterThan(0);
      for (const item of scenario.criteria) {
        expect(item.text.trim()).not.toBe("");
        for (const cmd of item.verify_commands ?? []) {
          expect(cmd.command.trim()).not.toBe("");
        }
      }
    }
  });

  it("findTestScenarioById returns the matching scenario or undefined", () => {
    const first = TEST_SCENARIOS[0]!;
    expect(findTestScenarioById(first.id)?.id).toBe(first.id);
    expect(findTestScenarioById("does-not-exist")).toBeUndefined();
  });
});
