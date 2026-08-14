import { randomUUID } from "node:crypto";
import type {
  AgentCreateOptions,
  AgentPort,
  AgentRun,
  AgentSession,
} from "./agentPort.js";
import type { EmitEvent, PatchOp } from "./types.js";

// SdkAgentPort is the production adapter that talks to `@cursor/sdk`. The SDK
// is a `optionalDependencies` install and imported dynamically so this file
// compiles even without the package on disk. Tests use MockAgentPort in
// mockAgentPort.ts and never load this module.

// Minimal structural typings of the SDK surface we actually use. We keep this
// small so that a future SDK version bump is easy to spot.
interface SdkModule {
  Agent: {
    create(opts: SdkCreateOpts): Promise<SdkAgent>;
    resume(id: string, opts: SdkResumeOpts): Promise<SdkAgent>;
  };
  CursorAgentError: new (...args: unknown[]) => Error;
}

interface SdkCreateOpts {
  apiKey: string;
  model?: { id: string };
  local: { cwd: string; settingSources?: string[] };
  mcpServers?: Record<string, { command: string; args: string[] }>;
  tools?: string[];
  disallowedTools?: string[];
  systemPrompt?: string;
}

interface SdkResumeOpts {
  apiKey: string;
  mcpServers?: Record<string, { command: string; args: string[] }>;
}

interface SdkAgent {
  agentId: string;
  send(input: string): Promise<SdkRun>;
  close?(): Promise<void>;
  [Symbol.asyncDispose]?: () => Promise<void>;
}

interface SdkRun {
  id: string;
  stream(): AsyncIterable<SdkMessage>;
  wait(): Promise<{ status: string; result?: string }>;
  cancel?(): Promise<void>;
  supports?(op: string): boolean;
}

type SdkMessage =
  | {
      type: "assistant";
      message: { content: Array<{ type: string; text?: string }> };
    }
  | {
      type: "tool_use" | "tool_call";
      name: string;
      input?: unknown;
    }
  | {
      type: "tool_result";
      name: string;
      ok?: boolean;
      error?: string;
    }
  | { type: string; [k: string]: unknown };

let sdkPromise: Promise<SdkModule> | null = null;

async function loadSdk(): Promise<SdkModule> {
  if (!sdkPromise) {
    // Dynamic import via a variable so TypeScript does not resolve the
    // module at compile time. This lets the sidecar build when
    // @cursor/sdk is unavailable offline; the import fails at runtime
    // instead, and readyz reports the reason.
    const spec = "@cursor/sdk";
    sdkPromise = import(/* @vite-ignore */ spec) as unknown as Promise<SdkModule>;
  }
  return sdkPromise;
}

export async function isSdkAvailable(): Promise<boolean> {
  try {
    await loadSdk();
    return true;
  } catch {
    return false;
  }
}

async function readSdkVersion(): Promise<string | undefined> {
  try {
    // Best effort — the SDK does not always export a version constant.
    const mod = (await loadSdk()) as unknown as { version?: string };
    return mod.version;
  } catch {
    return undefined;
  }
}

// tryParsePatch inspects a tool call and returns a normalized PatchEvent when
// the model invoked one of the hamix.draft_* prompt-write tools.
function tryParsePatch(
  name: string,
  input: unknown,
): { op: PatchOp; find?: string; value?: string; summary?: string } | null {
  if (typeof name !== "string") return null;
  const parsed = (input ?? {}) as Record<string, unknown>;
  if (name === "hamix.draft_set_prompt") {
    return {
      op: "set",
      value: typeof parsed.prompt === "string" ? parsed.prompt : "",
      summary: "prompt replaced",
    };
  }
  if (name === "hamix.draft_patch_prompt") {
    const op = typeof parsed.op === "string" ? parsed.op : "find_replace";
    return {
      op: (op === "append" ? "append" : "find_replace") as PatchOp,
      find: typeof parsed.find === "string" ? parsed.find : undefined,
      value: typeof parsed.value === "string" ? parsed.value : undefined,
      summary: typeof parsed.summary === "string" ? parsed.summary : undefined,
    };
  }
  return null;
}

class SdkRunAdapter implements AgentRun {
  constructor(
    public readonly runId: string,
    public readonly agentId: string,
    private readonly run: SdkRun,
  ) {}

  events(): AsyncIterable<EmitEvent> {
    const run = this.run;
    return {
      async *[Symbol.asyncIterator](): AsyncIterator<EmitEvent> {
        yield { kind: "status", data: { status: "thinking" } } as EmitEvent;
        let sawToken = false;
        try {
          for await (const msg of run.stream()) {
            if (msg.type === "assistant") {
              const assistant = msg as {
                message: { content: Array<{ type: string; text?: string }> };
              };
              for (const block of assistant.message.content ?? []) {
                if (block.type === "text" && typeof block.text === "string" && block.text.length > 0) {
                  if (!sawToken) {
                    sawToken = true;
                    yield { kind: "status", data: { status: "streaming" } } as EmitEvent;
                  }
                  yield {
                    kind: "token",
                    data: { delta: block.text },
                  } as EmitEvent;
                }
              }
            } else if (msg.type === "tool_use" || msg.type === "tool_call") {
              const tool = msg as { name: string; input?: unknown };
              yield {
                kind: "tool",
                data: { name: tool.name, phase: "start" },
              } as EmitEvent;
              const patch = tryParsePatch(tool.name, tool.input);
              if (patch) {
                yield {
                  kind: "patch",
                  data: patch,
                } as EmitEvent;
              }
            } else if (msg.type === "tool_result") {
              const tr = msg as { name: string; ok?: boolean; error?: string };
              yield {
                kind: "tool",
                data: {
                  name: tr.name,
                  phase: "end",
                  ok: tr.ok !== false,
                  ...(tr.error ? { error: tr.error } : {}),
                },
              } as EmitEvent;
            }
          }
          const result = await run.wait();
          if (result.status === "cancelled") {
            yield {
              kind: "status",
              data: { status: "cancelling" },
            } as EmitEvent;
            yield { kind: "done", data: { status: "cancelled" } } as EmitEvent;
            return;
          }
          if (result.status === "error") {
            yield {
              kind: "error",
              data: { code: "run_error", message: result.result ?? "run failed" },
            } as EmitEvent;
            yield { kind: "done", data: { status: "failed" } } as EmitEvent;
            return;
          }
          yield { kind: "done", data: { status: "done" } } as EmitEvent;
        } catch (err) {
          yield {
            kind: "error",
            data: {
              code: "sdk_error",
              message: err instanceof Error ? err.message : String(err),
            },
          } as EmitEvent;
          yield { kind: "done", data: { status: "failed" } } as EmitEvent;
        }
      },
    };
  }

  async cancel(): Promise<void> {
    if (this.run.supports && !this.run.supports("cancel")) return;
    if (typeof this.run.cancel === "function") {
      await this.run.cancel();
    }
  }
}

class SdkSessionAdapter implements AgentSession {
  constructor(
    public readonly agentId: string,
    private readonly agent: SdkAgent,
  ) {}

  async send(userMessage: string, snapshotText?: string): Promise<AgentRun> {
    const input = snapshotText
      ? `${snapshotText}\n\n---\n\n${userMessage}`
      : userMessage;
    const run = await this.agent.send(input);
    const runId = run.id ?? randomUUID();
    return new SdkRunAdapter(runId, this.agentId, run);
  }

  async close(): Promise<void> {
    if (typeof this.agent.close === "function") {
      await this.agent.close();
      return;
    }
    const dispose = this.agent[Symbol.asyncDispose];
    if (typeof dispose === "function") {
      await dispose.call(this.agent);
    }
  }
}

export class SdkAgentPort implements AgentPort {
  private cachedVersion: string | undefined;

  name(): string {
    return "sdk";
  }

  sdkVersion(): string | undefined {
    return this.cachedVersion;
  }

  async create(opts: AgentCreateOptions): Promise<AgentSession> {
    const sdk = await loadSdk();
    if (!this.cachedVersion) {
      this.cachedVersion = await readSdkVersion();
    }
    const commonMcp = { mcpServers: opts.mcpServers };
    if (opts.agentIdToResume) {
      const agent = await sdk.Agent.resume(opts.agentIdToResume, {
        apiKey: opts.apiKey,
        ...commonMcp,
      });
      return new SdkSessionAdapter(agent.agentId, agent);
    }
    const agent = await sdk.Agent.create({
      apiKey: opts.apiKey,
      ...(opts.model ? { model: { id: opts.model } } : {}),
      local: { cwd: opts.cwd, settingSources: [] },
      systemPrompt: opts.systemPrompt,
      tools: ["read", "grep", "glob", "ls", "mcp"],
      disallowedTools: ["shell", "edit", "task"],
      ...commonMcp,
    });
    return new SdkSessionAdapter(agent.agentId, agent);
  }
}
