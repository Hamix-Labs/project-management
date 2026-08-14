import type { EmitEvent } from "./types.js";

// AgentPort is the thin seam over @cursor/sdk. It lets us:
//   - swap in a mock in tests without installing the real SDK, and
//   - keep TypeScript compilation working when @cursor/sdk is not on disk
//     (optionalDependencies). The real adapter dynamic-imports the package.
//
// One AgentSession corresponds to one draft-assist session; the sidecar keeps
// a Map<sessionID, AgentSession> in agentSession.ts.

export interface AgentCreateOptions {
  cwd: string;
  model?: string;
  apiKey: string;
  systemPrompt: string;
  mcpServers: Record<string, { command: string; args: string[] }>;
  agentIdToResume?: string;
}

// AgentRun is one turn of the conversation. The port surfaces the normalized
// SSE events the sidecar forwards to the SPA. cancel() should trigger a
// terminal Done event via events().
export interface AgentRun {
  runId: string;
  agentId: string;
  events(): AsyncIterable<EmitEvent>;
  cancel(): Promise<void>;
}

// AgentSession is the durable Agent handle. send() starts one run; close()
// disposes the SDK executor. Reflect await using / Symbol.asyncDispose is not
// exposed here because the sidecar owns the lifetime explicitly.
export interface AgentSession {
  agentId: string;
  send(userMessage: string, snapshotText?: string): Promise<AgentRun>;
  close(): Promise<void>;
}

// AgentPort constructs AgentSessions. This is the only surface the HTTP
// routes depend on.
export interface AgentPort {
  name(): string;
  sdkVersion(): string | undefined;
  create(opts: AgentCreateOptions): Promise<AgentSession>;
}
