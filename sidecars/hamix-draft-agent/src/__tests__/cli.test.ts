import { describe, it, expect } from "vitest";
import { parseCli } from "../server.js";

describe("parseCli", () => {
  it("parses --port <n>", () => {
    expect(parseCli(["--port", "12345"])).toEqual({ port: 12345 });
  });

  it("accepts --port 0 for ephemeral", () => {
    expect(parseCli(["--port", "0"])).toEqual({ port: 0 });
  });

  it("defaults to 0 when no --port and no PORT env", () => {
    const prev = process.env.PORT;
    delete process.env.PORT;
    try {
      expect(parseCli([])).toEqual({ port: 0 });
    } finally {
      if (prev !== undefined) process.env.PORT = prev;
    }
  });

  it("reads PORT env when no --port", () => {
    const prev = process.env.PORT;
    process.env.PORT = "9099";
    try {
      expect(parseCli([])).toEqual({ port: 9099 });
    } finally {
      if (prev === undefined) delete process.env.PORT;
      else process.env.PORT = prev;
    }
  });
});
