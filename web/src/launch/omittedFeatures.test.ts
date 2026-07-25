import { describe, expect, it } from "vitest";
import { isUiFeatureOmitted, OMITTED_UI_FEATURES } from "./omittedFeatures";

describe("omittedFeatures", () => {
  it("documents projects as restored in Cycle 6", () => {
    expect(OMITTED_UI_FEATURES.projects).toBe(false);
    expect(isUiFeatureOmitted("projects")).toBe(false);
  });

  it("documents task tags as restored", () => {
    expect(OMITTED_UI_FEATURES.taskTags).toBe(false);
    expect(isUiFeatureOmitted("taskTags")).toBe(false);
  });

  it("documents task dependencies and milestone as omitted for the current launch", () => {
    expect(OMITTED_UI_FEATURES.taskDependencies).toBe(true);
    expect(isUiFeatureOmitted("taskDependencies")).toBe(true);
  });

  it("documents schedule as omitted for the current launch", () => {
    expect(OMITTED_UI_FEATURES.schedule).toBe(true);
    expect(isUiFeatureOmitted("schedule")).toBe(true);
  });

  it("documents release gates as omitted for the current launch", () => {
    expect(OMITTED_UI_FEATURES.releaseGates).toBe(true);
    expect(isUiFeatureOmitted("releaseGates")).toBe(true);
  });
});
