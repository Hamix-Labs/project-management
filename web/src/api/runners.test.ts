import { describe, expect, it } from "vitest";
import { fetchRunnerConfigSchema, fetchRunners } from "./runners";

describe("runners parsers via fetchRunners", () => {
  it("throws on invalid config field type", async () => {
    const originalFetch = globalThis.fetch;
    globalThis.fetch = (async () =>
      new Response(
        JSON.stringify([
          {
            id: "cursor",
            label: "Cursor",
            default_binary_hint: "agent",
            config_schema: {
              version: 1,
              fields: [{ key: "x", label: "X", type: "float" }],
            },
          },
        ]),
        { status: 200, headers: { "Content-Type": "application/json" } },
      )) as typeof fetch;
    try {
      await expect(fetchRunners()).rejects.toThrow(/must be one of/);
    } finally {
      globalThis.fetch = originalFetch;
    }
  });

  it("throws when runner id is empty", async () => {
    const originalFetch = globalThis.fetch;
    globalThis.fetch = (async () =>
      new Response(JSON.stringify([{ id: "", label: "X" }]), {
        status: 200,
        headers: { "Content-Type": "application/json" },
      })) as typeof fetch;
    try {
      await expect(fetchRunners()).rejects.toThrow(/non-empty string/);
    } finally {
      globalThis.fetch = originalFetch;
    }
  });

  it("parses a valid config schema", async () => {
    const originalFetch = globalThis.fetch;
    globalThis.fetch = (async () =>
      new Response(
        JSON.stringify({
          version: 1,
          fields: [
            {
              key: "model",
              label: "Model",
              type: "enum",
              enum_values: [{ value: "a", label: "A" }],
            },
          ],
        }),
        { status: 200, headers: { "Content-Type": "application/json" } },
      )) as typeof fetch;
    try {
      await expect(fetchRunnerConfigSchema("cursor")).resolves.toEqual({
        version: 1,
        fields: [
          {
            key: "model",
            label: "Model",
            type: "enum",
            enum_values: [{ value: "a", label: "A" }],
          },
        ],
      });
    } finally {
      globalThis.fetch = originalFetch;
    }
  });
});
