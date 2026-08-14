import { randomUUID } from "node:crypto";
import type {
  AgentCreateOptions,
  AgentPort,
  AgentRun,
  AgentSession,
} from "./agentPort.js";
import type { EmitEvent } from "./types.js";

// MockAgentPort is the deterministic fake used by tests. It mirrors the shape
// of the SDK adapter so contract tests can assert on SSE frame order without
// installing @cursor/sdk. Tests can override behaviour by passing a script.

export type MockScriptStep =
  | { type: "wait"; ms: number }
  | { type: "emit"; event: EmitEvent };

export type MockScript = MockScriptStep[];

export interface MockAgentOptions {
  script?: MockScript;
  scriptFactory?: (userMessage: string) => MockScript;
  createDelayMs?: number;
}

const defaultScript: MockScript = [
  { type: "emit", event: { kind: "status", data: { status: "thinking" } } },
  { type: "wait", ms: 5 },
  { type: "emit", event: { kind: "status", data: { status: "streaming" } } },
  {
    type: "emit",
    event: { kind: "token", data: { delta: "Here is a tightened prompt draft." } },
  },
  { type: "emit", event: { kind: "done", data: { status: "done" } } },
];

class MockRun implements AgentRun {
  private cancelled = false;
  private cancelResolvers: Array<() => void> = [];

  constructor(
    public readonly runId: string,
    public readonly agentId: string,
    private readonly script: MockScript,
  ) {}

  private wait(ms: number): Promise<void> {
    return new Promise((resolve) => {
      const t = setTimeout(() => {
        this.cancelResolvers = this.cancelResolvers.filter((r) => r !== resolve);
        resolve();
      }, ms);
      const cancelResolve = (): void => {
        clearTimeout(t);
        resolve();
      };
      this.cancelResolvers.push(cancelResolve);
    });
  }

  events(): AsyncIterable<EmitEvent> {
    const self = this;
    return {
      async *[Symbol.asyncIterator](): AsyncIterator<EmitEvent> {
        for (const step of self.script) {
          if (self.cancelled) {
            yield {
              kind: "status",
              data: { status: "cancelling" },
            } as EmitEvent;
            yield { kind: "done", data: { status: "cancelled" } } as EmitEvent;
            return;
          }
          if (step.type === "wait") {
            await self.wait(step.ms);
            if (self.cancelled) {
              yield {
                kind: "status",
                data: { status: "cancelling" },
              } as EmitEvent;
              yield {
                kind: "done",
                data: { status: "cancelled" },
              } as EmitEvent;
              return;
            }
            continue;
          }
          yield step.event;
        }
      },
    };
  }

  async cancel(): Promise<void> {
    this.cancelled = true;
    for (const r of this.cancelResolvers.splice(0)) r();
  }
}

class MockSession implements AgentSession {
  constructor(
    public readonly agentId: string,
    private readonly opts: MockAgentOptions,
  ) {}

  async send(userMessage: string): Promise<AgentRun> {
    const script = this.opts.scriptFactory
      ? this.opts.scriptFactory(userMessage)
      : this.opts.script ?? defaultScript;
    return new MockRun(randomUUID(), this.agentId, script);
  }

  async close(): Promise<void> {
    // no-op; nothing to release for the mock.
  }
}

export class MockAgentPort implements AgentPort {
  constructor(private readonly opts: MockAgentOptions = {}) {}

  name(): string {
    return "mock";
  }

  sdkVersion(): string | undefined {
    return "mock-1.0.0";
  }

  async create(opts: AgentCreateOptions): Promise<AgentSession> {
    if (this.opts.createDelayMs && this.opts.createDelayMs > 0) {
      await new Promise((r) => setTimeout(r, this.opts.createDelayMs));
    }
    const agentId = opts.agentIdToResume ?? `mock-${randomUUID()}`;
    return new MockSession(agentId, this.opts);
  }
}
