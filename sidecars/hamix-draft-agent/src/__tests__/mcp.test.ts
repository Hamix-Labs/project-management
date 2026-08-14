import { describe, it, expect } from "vitest";
import { readFileSync, existsSync } from "node:fs";
import { writeBindFile, mcpStdioEntry } from "../mcp.js";

describe("MCP bind file", () => {
  it("writes bind_schema_version=1 with session_id and nonce", () => {
    const { bindPath, cleanup } = writeBindFile("sess-1", "nonce-1", "http://127.0.0.1:8080");
    try {
      expect(existsSync(bindPath)).toBe(true);
      const parsed = JSON.parse(readFileSync(bindPath, "utf8")) as {
        bind_schema_version: number;
        session_id: string;
        nonce: string;
        taskapi_base_url?: string;
      };
      expect(parsed.bind_schema_version).toBe(1);
      expect(parsed.session_id).toBe("sess-1");
      expect(parsed.nonce).toBe("nonce-1");
      expect(parsed.taskapi_base_url).toBe("http://127.0.0.1:8080");
    } finally {
      cleanup();
    }
    expect(existsSync(bindPath)).toBe(false);
  });

  it("omits taskapi_base_url when not provided", () => {
    const { bindPath, cleanup } = writeBindFile("s", "n");
    try {
      const parsed = JSON.parse(readFileSync(bindPath, "utf8")) as {
        taskapi_base_url?: string;
      };
      expect(parsed.taskapi_base_url).toBeUndefined();
    } finally {
      cleanup();
    }
  });

  it("mcpStdioEntry points at hamix-draft-mcp with --bind", () => {
    const entry = mcpStdioEntry("/tmp/bind.json");
    expect(entry.command).toBe("hamix-draft-mcp");
    expect(entry.args).toEqual(["--bind", "/tmp/bind.json"]);
  });
});
