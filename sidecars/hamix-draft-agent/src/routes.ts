import type { IncomingMessage, ServerResponse } from "node:http";
import { randomUUID } from "node:crypto";
import { AgentSessionRegistry } from "./agentSession.js";
import { SseWriter } from "./sse.js";
import type { RunRequestBody, SessionEventData } from "./types.js";
import { SESSION_SCHEMA_VERSION } from "./types.js";
import type { AgentPort } from "./agentPort.js";

export interface RouterDeps {
  port: AgentPort;
  registry: AgentSessionRegistry;
  getApiKey(): string | undefined;
  taskapiBaseUrl?: string;
}

// readJsonBody buffers the request body and JSON-parses it. Rejects on
// oversized payloads (1 MiB) to keep the loopback socket well-behaved even if
// the caller misbehaves.
async function readJsonBody(req: IncomingMessage): Promise<unknown> {
  const chunks: Buffer[] = [];
  let size = 0;
  for await (const chunk of req) {
    const buf = Buffer.isBuffer(chunk) ? chunk : Buffer.from(chunk as string);
    size += buf.length;
    if (size > 1024 * 1024) {
      throw new Error("request body too large");
    }
    chunks.push(buf);
  }
  if (chunks.length === 0) return {};
  const raw = Buffer.concat(chunks).toString("utf8");
  return JSON.parse(raw);
}

function sendJson(res: ServerResponse, status: number, body: unknown): void {
  const payload = JSON.stringify(body);
  res.writeHead(status, {
    "content-type": "application/json; charset=utf-8",
    "content-length": Buffer.byteLength(payload),
  });
  res.end(payload);
}

function methodNotAllowed(res: ServerResponse): void {
  sendJson(res, 405, { error: "method not allowed" });
}

function notFound(res: ServerResponse): void {
  sendJson(res, 404, { error: "not found" });
}

// buildRouter wires the HTTP method + path dispatch. Kept dependency-light so
// tests can drive it with a stub AgentPort.
export function buildRouter(deps: RouterDeps) {
  return async function handle(req: IncomingMessage, res: ServerResponse): Promise<void> {
    const url = new URL(req.url ?? "/", "http://localhost");
    const path = url.pathname;
    const method = (req.method ?? "GET").toUpperCase();

    try {
      if (path === "/healthz" && method === "GET") return healthz(res, deps);
      if (path === "/readyz" && method === "GET") return readyz(res, deps);
      if (path === "/runs" && method === "POST") return postRun(req, res, deps);
      if (path.startsWith("/runs/") && path.endsWith("/cancel") && method === "POST") {
        const runId = path.slice("/runs/".length, -"/cancel".length);
        return postCancel(req, res, deps, runId);
      }
      if (path === "/runs" || path.startsWith("/runs/")) return methodNotAllowed(res);
      return notFound(res);
    } catch (err) {
      const msg = err instanceof Error ? err.message : String(err);
      sendJson(res, 500, { error: msg });
    }
  };
}

function healthz(res: ServerResponse, deps: RouterDeps): void {
  sendJson(res, 200, {
    ok: true,
    sdk_version: deps.port.sdkVersion(),
    agents_active: deps.registry.activeAgents(),
  });
}

function readyz(res: ServerResponse, deps: RouterDeps): void {
  const apiKey = deps.getApiKey();
  if (!apiKey) {
    sendJson(res, 200, { ready: false, reason: "missing_key" });
    return;
  }
  sendJson(res, 200, { ready: true });
}

async function postRun(
  req: IncomingMessage,
  res: ServerResponse,
  deps: RouterDeps,
): Promise<void> {
  const body = (await readJsonBody(req)) as Partial<RunRequestBody>;
  if (!body || typeof body.session_id !== "string" || !body.session_id) {
    return sendJson(res, 400, { error: "session_id is required" });
  }
  if (typeof body.user_message !== "string" || !body.user_message) {
    return sendJson(res, 400, { error: "user_message is required" });
  }
  if (typeof body.worktree_cwd !== "string" || !body.worktree_cwd) {
    return sendJson(res, 400, { error: "worktree_cwd is required" });
  }
  const apiKey = deps.getApiKey();
  if (!apiKey) {
    return sendJson(res, 503, { error: "CURSOR_API_KEY missing" });
  }

  const sse = new SseWriter(res);
  const runId = body.run_id ?? randomUUID();

  // Open the stream immediately so the client sees the session frame before
  // Agent.create finishes on cold sessions.
  sse.start();

  const sessionEvent: SessionEventData = {
    session_id: body.session_id,
    snapshot: body.snapshot ?? {},
    schema_version: SESSION_SCHEMA_VERSION,
  };
  sse.emit({ kind: "session", data: sessionEvent });

  const abortHandler = (): void => {
    void deps.registry.cancel(body.session_id!, runId).catch(() => undefined);
  };
  req.on("close", abortHandler);

  try {
    const started = await deps.registry.startRun({
      sessionID: body.session_id,
      runID: runId,
      userMessage: body.user_message,
      worktreeCwd: body.worktree_cwd,
      snapshot: body.snapshot,
      apiKey,
      ...(body.model ? { model: body.model } : {}),
      ...(body.agent_id ? { agentIdToResume: body.agent_id } : {}),
      ...(deps.taskapiBaseUrl ? { taskapiBaseUrl: deps.taskapiBaseUrl } : {}),
    });
    try {
      for await (const ev of started.run.events()) {
        if (sse.isClosed) break;
        sse.emit(ev);
      }
    } finally {
      await started.cleanup();
    }
  } catch (err) {
    const message = err instanceof Error ? err.message : String(err);
    sse.emit({ kind: "error", data: { code: "start_failed", message } });
    sse.emit({ kind: "done", data: { status: "failed" } });
  } finally {
    req.off("close", abortHandler);
    sse.close();
  }
}

async function postCancel(
  req: IncomingMessage,
  res: ServerResponse,
  deps: RouterDeps,
  runId: string,
): Promise<void> {
  // Body may carry { session_id }; without it we cannot look up the run.
  const body = (await readJsonBody(req).catch(() => ({}))) as {
    session_id?: string;
  };
  if (!body.session_id) {
    return sendJson(res, 400, { error: "session_id is required" });
  }
  const known = await deps.registry.cancel(body.session_id, runId);
  if (!known) {
    return sendJson(res, 404, { error: "run not found" });
  }
  res.writeHead(202, { "content-type": "application/json; charset=utf-8" });
  res.end(JSON.stringify({ accepted: true, run_id: runId }));
}
