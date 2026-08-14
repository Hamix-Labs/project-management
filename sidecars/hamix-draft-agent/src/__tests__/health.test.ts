import { describe, it, expect, afterEach } from "vitest";
import { MockAgentPort } from "../mockAgentPort.js";
import { bootTestServer } from "./helpers.js";

const teardowns: Array<() => Promise<void>> = [];

afterEach(async () => {
  for (const c of teardowns.splice(0).reverse()) await c();
});

describe("healthz + readyz", () => {
  it("healthz reports ok, sdk_version, and agents_active", async () => {
    const port = new MockAgentPort();
    const s = await bootTestServer({ agentPort: port, apiKey: "k" });
    teardowns.push(s.close);

    const res = await fetch(`${s.baseUrl}/healthz`);
    expect(res.status).toBe(200);
    const body = (await res.json()) as {
      ok: boolean;
      sdk_version?: string;
      agents_active: number;
    };
    expect(body.ok).toBe(true);
    expect(body.sdk_version).toBe("mock-1.0.0");
    expect(body.agents_active).toBe(0);
  });

  it("readyz returns missing_key when apiKey is absent", async () => {
    const port = new MockAgentPort();
    const s = await bootTestServer({ agentPort: port, apiKey: "" });
    teardowns.push(s.close);
    const res = await fetch(`${s.baseUrl}/readyz`);
    expect(res.status).toBe(200);
    const body = (await res.json()) as { ready: boolean; reason?: string };
    expect(body.ready).toBe(false);
    expect(body.reason).toBe("missing_key");
  });

  it("readyz returns ready:true when apiKey is set", async () => {
    const port = new MockAgentPort();
    const s = await bootTestServer({ agentPort: port, apiKey: "cursor_abc" });
    teardowns.push(s.close);
    const res = await fetch(`${s.baseUrl}/readyz`);
    const body = (await res.json()) as { ready: boolean };
    expect(body.ready).toBe(true);
  });
});
