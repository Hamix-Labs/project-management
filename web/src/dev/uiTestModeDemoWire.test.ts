import { describe, expect, it } from "vitest";
import { parseProjectListResponse } from "@/api/projects";
import { parseTaskListResponse, parseTaskStatsResponse } from "@/api/parseTaskApi";
import {
  demoProjectsListWire,
  demoTaskStatsWire,
  demoTasksListWire,
} from "./uiTestModeDemoWire";

describe("uiTestModeDemoWire", () => {
  it("parses as valid API payloads", () => {
    expect(() => parseProjectListResponse(demoProjectsListWire())).not.toThrow();
    expect(() => parseTaskListResponse(demoTasksListWire(200, 0, null))).not.toThrow();
    expect(() => parseTaskStatsResponse(demoTaskStatsWire())).not.toThrow();
  });
});
