import { afterEach, describe, expect, it, vi } from "vitest";
import {
  getProject,
  listProjects,
  patchProject,
  parseProject,
} from "./projects";

const projectWire = {
  id: "11111111-1111-4111-8111-111111111111",
  name: "Context moat",
  description: "Long-running work",
  status: "active",
  context_summary: "Shared memory",
  created_at: "2026-04-26T00:00:00Z",
  updated_at: "2026-04-26T00:00:00Z",
};

describe("project API parsers", () => {
  afterEach(() => {
    vi.restoreAllMocks();
  });

  it("parses project rows", () => {
    expect(parseProject(projectWire).name).toBe("Context moat");
  });

  it("rejects unknown project statuses", () => {
    expect(() => parseProject({ ...projectWire, status: "paused" })).toThrow(
      /project status/,
    );
  });

  it("lists projects with query params", async () => {
    const spy = vi.spyOn(globalThis, "fetch").mockResolvedValue(
      new Response(JSON.stringify({ projects: [projectWire], limit: 10 }), {
        status: 200,
        headers: { "Content-Type": "application/json" },
      }),
    );

    const out = await listProjects({ limit: 10, includeArchived: true });

    expect(out.projects).toHaveLength(1);
    expect(String(spy.mock.calls[0][0])).toContain("include_archived=true");
  });

  it("gets one project", async () => {
    vi.spyOn(globalThis, "fetch").mockResolvedValue(
      new Response(JSON.stringify(projectWire), {
        status: 200,
        headers: { "Content-Type": "application/json" },
      }),
    );

    await expect(getProject(projectWire.id)).resolves.toMatchObject({
      id: projectWire.id,
    });
  });

  it("patches projects", async () => {
    const spy = vi.spyOn(globalThis, "fetch").mockResolvedValue(
      new Response(JSON.stringify({ ...projectWire, name: "Renamed" }), {
        status: 200,
        headers: { "Content-Type": "application/json" },
      }),
    );

    const out = await patchProject(projectWire.id, { name: "Renamed" });

    expect(out.name).toBe("Renamed");
    expect(spy.mock.calls[0][1]).toMatchObject({ method: "PATCH" });
  });
});
