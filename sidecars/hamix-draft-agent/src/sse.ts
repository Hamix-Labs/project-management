import type { ServerResponse } from "node:http";
import type { EmitEvent } from "./types.js";

// SSE heartbeat interval in ms. The Go domain contract requires ≤3s so a dead
// socket cannot masquerade as "still thinking".
export const SSE_HEARTBEAT_MS = 3000;

// SseWriter serializes named SSE frames to an HTTP response. Each frame gets a
// monotonic id so a reconnecting client can Last-Event-ID replay from an
// upstream buffer (the Go taskapi ring keeps 256 events).
export class SseWriter {
  private nextId = 1;
  private heartbeat: NodeJS.Timeout | null = null;
  private closed = false;

  constructor(private readonly res: ServerResponse) {}

  // start writes the SSE headers, flushes them, and begins the heartbeat.
  start(): void {
    this.res.writeHead(200, {
      "content-type": "text/event-stream; charset=utf-8",
      "cache-control": "no-cache, no-transform",
      connection: "keep-alive",
      "x-accel-buffering": "no",
    });
    if (typeof (this.res as { flushHeaders?: () => void }).flushHeaders === "function") {
      (this.res as { flushHeaders?: () => void }).flushHeaders?.();
    }
    this.heartbeat = setInterval(() => this.writeHeartbeat(), SSE_HEARTBEAT_MS);
  }

  // emit writes one named SSE frame with a monotonic id.
  emit(ev: EmitEvent): void {
    if (this.closed) return;
    const id = this.nextId++;
    const payload = JSON.stringify(ev.data);
    // SSE frame: id, event, data. Ends with a blank line.
    this.res.write(`id: ${id}\nevent: ${ev.kind}\ndata: ${payload}\n\n`);
  }

  // writeHeartbeat writes an SSE comment line. Not a named event so the SPA
  // does not confuse it with a status frame.
  private writeHeartbeat(): void {
    if (this.closed) return;
    this.res.write(`: heartbeat ${Date.now()}\n\n`);
  }

  // close stops the heartbeat and ends the response. Idempotent.
  close(): void {
    if (this.closed) return;
    this.closed = true;
    if (this.heartbeat) {
      clearInterval(this.heartbeat);
      this.heartbeat = null;
    }
    try {
      this.res.end();
    } catch {
      // response may already be torn down; nothing to do.
    }
  }

  get isClosed(): boolean {
    return this.closed;
  }
}
