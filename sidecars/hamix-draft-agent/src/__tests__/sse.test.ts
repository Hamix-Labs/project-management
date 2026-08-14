import { describe, expect, it, afterEach } from "vitest";
import { MockAgentPort } from "../mockAgentPort.js";
import { bootTestServer, collectStream } from "./helpers.js";

const teardowns: Array<() => Promise<void>> = [];

afterEach(async () => {
  const list = teardowns.splice(0);
  for (const close of list.reverse()) {
    await close();
  }
});

async function boot(port: MockAgentPort, apiKey = "test-key") {
  const s = await bootTestServer({ agentPort: port, apiKey });
  teardowns.push(s.close);
  return s;
}

describe("SSE contract", () => {
  it("emits session, status, token, done in order with schema_version=1", async () => {
    const port = new MockAgentPort();
    const s = await boot(port);

    const res = await fetch(`${s.baseUrl}/runs`, {
      method: "POST",
      headers: { "content-type": "application/json" },
      body: JSON.stringify({
        session_id: "sess-1",
        user_message: "help me tighten the prompt",
        worktree_cwd: process.cwd(),
      }),
    });
    expect(res.status).toBe(200);
    expect(res.headers.get("content-type")).toContain("text/event-stream");

    const events = await collectStream(res.body!);
    const kinds = events.map((e) => e.kind);
    expect(kinds[0]).toBe("session");
    expect(kinds).toContain("status");
    expect(kinds).toContain("token");
    expect(kinds[kinds.length - 1]).toBe("done");

    const session = events[0]!.data as { schema_version: number; session_id: string };
    expect(session.schema_version).toBe(1);
    expect(session.session_id).toBe("sess-1");
  });

  it("emits patch after hamix.draft_set_prompt tool call", async () => {
    const port = new MockAgentPort({
      script: [
        { type: "emit", event: { kind: "status", data: { status: "thinking" } } },
        {
          type: "emit",
          event: {
            kind: "tool",
            data: { name: "hamix.draft_set_prompt", phase: "start" },
          },
        },
        {
          type: "emit",
          event: {
            kind: "patch",
            data: { op: "set", value: "<p>rewritten</p>", summary: "prompt replaced" },
          },
        },
        {
          type: "emit",
          event: {
            kind: "tool",
            data: { name: "hamix.draft_set_prompt", phase: "end", ok: true },
          },
        },
        { type: "emit", event: { kind: "done", data: { status: "done" } } },
      ],
    });
    const s = await boot(port);
    const res = await fetch(`${s.baseUrl}/runs`, {
      method: "POST",
      headers: { "content-type": "application/json" },
      body: JSON.stringify({
        session_id: "sess-p",
        user_message: "rewrite",
        worktree_cwd: process.cwd(),
      }),
    });
    const events = await collectStream(res.body!);
    const patches = events.filter((e) => e.kind === "patch");
    expect(patches).toHaveLength(1);
    expect((patches[0]!.data as { op: string }).op).toBe("set");
  });

  it("assigns monotonic ids to frames", async () => {
    const port = new MockAgentPort();
    const s = await boot(port);
    const res = await fetch(`${s.baseUrl}/runs`, {
      method: "POST",
      headers: { "content-type": "application/json" },
      body: JSON.stringify({
        session_id: "sess-ids",
        user_message: "hi",
        worktree_cwd: process.cwd(),
      }),
    });
    const reader = res.body!.getReader();
    const decoder = new TextDecoder();
    let text = "";
    while (true) {
      const { value, done } = await reader.read();
      if (done) break;
      text += decoder.decode(value, { stream: true });
      if (text.includes("event: done")) break;
    }
    const ids = Array.from(text.matchAll(/^id: (\d+)$/gm)).map((m) => Number(m[1]));
    expect(ids.length).toBeGreaterThan(1);
    for (let i = 1; i < ids.length; i++) {
      expect(ids[i]!).toBeGreaterThan(ids[i - 1]!);
    }
  });

  it("rejects requests when session_id is missing", async () => {
    const port = new MockAgentPort();
    const s = await boot(port);
    const res = await fetch(`${s.baseUrl}/runs`, {
      method: "POST",
      headers: { "content-type": "application/json" },
      body: JSON.stringify({ user_message: "x", worktree_cwd: process.cwd() }),
    });
    expect(res.status).toBe(400);
  });
});
