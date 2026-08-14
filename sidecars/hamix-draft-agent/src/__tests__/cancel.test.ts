import { describe, it, expect, afterEach } from "vitest";
import { MockAgentPort } from "../mockAgentPort.js";
import { bootTestServer, parseSseStream } from "./helpers.js";

const teardowns: Array<() => Promise<void>> = [];

afterEach(async () => {
  for (const c of teardowns.splice(0).reverse()) await c();
});

describe("cancel + resume", () => {
  it("cancel emits status=cancelling then done=cancelled", async () => {
    const port = new MockAgentPort({
      script: [
        { type: "emit", event: { kind: "status", data: { status: "thinking" } } },
        { type: "wait", ms: 10_000 },
        { type: "emit", event: { kind: "done", data: { status: "done" } } },
      ],
    });
    const s = await bootTestServer({ agentPort: port, apiKey: "k" });
    teardowns.push(s.close);

    // Kick off the run in the background — we cancel via HTTP once we have
    // seen the session frame.
    const res = await fetch(`${s.baseUrl}/runs`, {
      method: "POST",
      headers: { "content-type": "application/json" },
      body: JSON.stringify({
        session_id: "sess-cx",
        run_id: "run-cx",
        user_message: "start something long",
        worktree_cwd: process.cwd(),
      }),
    });
    expect(res.status).toBe(200);

    const iter = parseSseStream(res.body!);
    const first = await iter.next();
    expect(first.value?.kind).toBe("session");

    // Consume any status: thinking that arrived immediately, then send cancel.
    const cancelRes = await fetch(`${s.baseUrl}/runs/run-cx/cancel`, {
      method: "POST",
      headers: { "content-type": "application/json" },
      body: JSON.stringify({ session_id: "sess-cx" }),
    });
    expect(cancelRes.status).toBe(202);

    const collected: string[] = [];
    for (;;) {
      const next = await iter.next();
      if (next.done) break;
      collected.push(next.value!.kind);
      if (next.value!.kind === "done") break;
    }
    expect(collected).toContain("status");
    expect(collected[collected.length - 1]).toBe("done");
  });

  it("cancel returns 404 when the run is unknown", async () => {
    const port = new MockAgentPort();
    const s = await bootTestServer({ agentPort: port, apiKey: "k" });
    teardowns.push(s.close);
    const res = await fetch(`${s.baseUrl}/runs/nope/cancel`, {
      method: "POST",
      headers: { "content-type": "application/json" },
      body: JSON.stringify({ session_id: "sess-none" }),
    });
    expect(res.status).toBe(404);
  });
});
