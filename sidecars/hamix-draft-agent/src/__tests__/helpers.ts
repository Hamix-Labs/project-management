// Test helpers for the hamix-draft-agent HTTP + SSE surface. These are not
// vitest fixtures — they are just plain functions the specs import.
import { startServer } from "../server.js";
import type { AgentPort } from "../agentPort.js";

export async function bootTestServer(opts: {
  agentPort: AgentPort;
  apiKey?: string;
}) {
  const running = await startServer({
    port: 0,
    agentPort: opts.agentPort,
    getApiKey: () => opts.apiKey ?? "test-key",
  });
  return {
    port: running.port,
    baseUrl: `http://127.0.0.1:${running.port}`,
    async close() {
      await running.close();
    },
  };
}

// parseSseStream consumes a fetch Response body and yields each named event
// frame as { kind, data }. Heartbeat comment lines are ignored. Stops when
// the connection closes.
export async function* parseSseStream(
  body: ReadableStream<Uint8Array>,
): AsyncGenerator<{ id?: string; kind: string; data: unknown }> {
  const reader = body.getReader();
  const decoder = new TextDecoder();
  let buf = "";
  let currentId: string | undefined;
  let currentEvent: string | undefined;
  let dataLines: string[] = [];
  while (true) {
    const { value, done } = await reader.read();
    if (done) break;
    buf += decoder.decode(value, { stream: true });
    let idx = buf.indexOf("\n");
    while (idx >= 0) {
      const line = buf.slice(0, idx).replace(/\r$/, "");
      buf = buf.slice(idx + 1);
      if (line === "") {
        if (currentEvent && dataLines.length > 0) {
          const raw = dataLines.join("\n");
          let parsed: unknown = raw;
          try {
            parsed = JSON.parse(raw);
          } catch {
            // leave raw when it is not JSON
          }
          yield {
            ...(currentId ? { id: currentId } : {}),
            kind: currentEvent,
            data: parsed,
          };
        }
        currentEvent = undefined;
        currentId = undefined;
        dataLines = [];
      } else if (line.startsWith(":")) {
        // comment / heartbeat
      } else if (line.startsWith("id:")) {
        currentId = line.slice(3).trim();
      } else if (line.startsWith("event:")) {
        currentEvent = line.slice(6).trim();
      } else if (line.startsWith("data:")) {
        dataLines.push(line.slice(5).trim());
      }
      idx = buf.indexOf("\n");
    }
  }
}

export async function collectStream(
  body: ReadableStream<Uint8Array>,
): Promise<Array<{ kind: string; data: unknown }>> {
  const out: Array<{ kind: string; data: unknown }> = [];
  for await (const ev of parseSseStream(body)) {
    out.push({ kind: ev.kind, data: ev.data });
    if (ev.kind === "done") break;
  }
  return out;
}
