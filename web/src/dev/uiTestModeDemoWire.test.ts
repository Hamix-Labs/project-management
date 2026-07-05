import { describe, expect, it } from "vitest";
import { DEMO_PRIMARY_PROJECT_ID } from "./uiTestModeDemoWire";
import {
  parseProjectContextListResponse,
  parseProjectListResponse,
} from "@/api/projects";
import { parseTaskListResponse, parseTaskStatsResponse } from "@/api/parseTaskApi";
import {
  demoContextWire,
  demoProjectsListWire,
  demoTaskStatsWire,
  demoTasksListWire,
} from "./uiTestModeDemoWire";

describe("uiTestModeDemoWire", () => {
  it("parses as valid API payloads", () => {
    expect(() => parseProjectListResponse(demoProjectsListWire())).not.toThrow();
    expect(() => parseProjectContextListResponse(demoContextWire(DEMO_PRIMARY_PROJECT_ID))).not.toThrow();
    expect(() => parseTaskListResponse(demoTasksListWire(200, 0, null))).not.toThrow();
    expect(() => parseTaskStatsResponse(demoTaskStatsWire())).not.toThrow();
  });
});
