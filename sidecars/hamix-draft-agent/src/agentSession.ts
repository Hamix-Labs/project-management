import { randomUUID } from "node:crypto";
import type {
  AgentPort,
  AgentRun,
  AgentSession,
} from "./agentPort.js";
import { writeBindFile, mcpStdioEntry } from "./mcp.js";
import { DRAFT_AGENT_SYSTEM_PROMPT } from "./promptSystem.js";
import type { FormSnapshot } from "./types.js";

// RunSlot tracks one in-flight run keyed by the wire run_id. The run itself
// is set once agent.send() resolves; cancelledEarly is honoured by the
// event-forwarding loop if a cancel arrived before send() returned.
interface RunSlot {
  run?: AgentRun;
  cancelledEarly: boolean;
}

// SessionRecord binds the SDK session to the bind-file we wrote on its behalf.
// One record per draft-assist session id (from taskapi).
interface SessionRecord {
  sessionID: string;
  agent: AgentSession;
  cwd: string;
  nonce: string;
  bindCleanup: () => void;
  activeRuns: Map<string, RunSlot>;
}

export interface StartRunInput {
  sessionID: string;
  runID: string;
  userMessage: string;
  worktreeCwd: string;
  snapshot?: FormSnapshot;
  apiKey: string;
  model?: string;
  agentIdToResume?: string;
  taskapiBaseUrl?: string;
}

export interface StartRunResult {
  run: AgentRun;
  runID: string;
  agentId: string;
  cleanup: () => Promise<void>;
}

// snapshotSummary is the "short fresh snapshot" the design doc says every
// send() should include so follow-ups see edits the user made by hand.
function snapshotSummary(snap: FormSnapshot | undefined): string | undefined {
  if (!snap) return undefined;
  const parts: string[] = [];
  if (snap.title) parts.push(`title: ${snap.title}`);
  if (snap.priority) parts.push(`priority: ${snap.priority}`);
  if (snap.criteria && snap.criteria.length > 0) {
    parts.push(`criteria:\n- ${snap.criteria.join("\n- ")}`);
  }
  if (snap.tags && snap.tags.length > 0) parts.push(`tags: ${snap.tags.join(", ")}`);
  if (snap.prompt) parts.push(`current prompt:\n${snap.prompt}`);
  if (parts.length === 0) return undefined;
  return `Current form snapshot:\n${parts.join("\n")}`;
}

// AgentSessionRegistry owns the Map<sessionID, AgentSession>. Callers acquire
// a session on the first /runs call and reuse it for follow-ups. Cancel and
// closeAll are wired into the HTTP routes and process shutdown.
export class AgentSessionRegistry {
  private readonly sessions = new Map<string, SessionRecord>();

  constructor(private readonly port: AgentPort) {}

  activeAgents(): number {
    return this.sessions.size;
  }

  // startRun ensures a session exists, then invokes agent.send with the user
  // message and a snapshot digest. The returned AgentRun is the event source
  // the SSE writer forwards. The slot is registered before `agent.send()`
  // resolves so an eager /cancel does not 404-race the run.
  async startRun(input: StartRunInput): Promise<StartRunResult> {
    let rec = this.sessions.get(input.sessionID);
    if (!rec) {
      rec = await this.createSession(input);
      this.sessions.set(input.sessionID, rec);
    }
    const slot: RunSlot = { cancelledEarly: false };
    rec.activeRuns.set(input.runID, slot);
    const snap = snapshotSummary(input.snapshot);
    try {
      const run = await rec.agent.send(input.userMessage, snap);
      slot.run = run;
      if (slot.cancelledEarly) {
        try {
          await run.cancel();
        } catch {
          // best-effort; the mock/SDK may already be in a terminal state.
        }
      }
      const cleanup = async (): Promise<void> => {
        rec?.activeRuns.delete(input.runID);
      };
      return { run, runID: input.runID, agentId: rec.agent.agentId, cleanup };
    } catch (err) {
      rec.activeRuns.delete(input.runID);
      throw err;
    }
  }

  // cancel dispatches to the run of the given id. Returns true when the run
  // was known — 202 accepted contract; the terminal Done frame comes over
  // SSE from the run's event iterator. If the slot exists but agent.send
  // hasn't produced a run yet, cancellation is recorded and applied when
  // the run resolves.
  async cancel(sessionID: string, runID: string): Promise<boolean> {
    const rec = this.sessions.get(sessionID);
    if (!rec) return false;
    const slot = rec.activeRuns.get(runID);
    if (!slot) return false;
    slot.cancelledEarly = true;
    if (slot.run) {
      await slot.run.cancel();
    }
    return true;
  }

  // closeSession disposes one session's SDK executor and removes the bind
  // file. Safe to call on missing ids.
  async closeSession(sessionID: string): Promise<void> {
    const rec = this.sessions.get(sessionID);
    if (!rec) return;
    this.sessions.delete(sessionID);
    for (const slot of rec.activeRuns.values()) {
      slot.cancelledEarly = true;
      if (slot.run) {
        try {
          await slot.run.cancel();
        } catch {
          // ignore; run may already be terminal
        }
      }
    }
    try {
      await rec.agent.close();
    } finally {
      rec.bindCleanup();
    }
  }

  async closeAll(): Promise<void> {
    const ids = Array.from(this.sessions.keys());
    await Promise.all(ids.map((id) => this.closeSession(id)));
  }

  private async createSession(input: StartRunInput): Promise<SessionRecord> {
    const nonce = randomUUID();
    const bind = writeBindFile(input.sessionID, nonce, input.taskapiBaseUrl);
    try {
      const agent = await this.port.create({
        cwd: input.worktreeCwd,
        model: input.model,
        apiKey: input.apiKey,
        systemPrompt: DRAFT_AGENT_SYSTEM_PROMPT,
        mcpServers: {
          "hamix-draft": mcpStdioEntry(bind.bindPath),
        },
        ...(input.agentIdToResume ? { agentIdToResume: input.agentIdToResume } : {}),
      });
      return {
        sessionID: input.sessionID,
        agent,
        cwd: input.worktreeCwd,
        nonce,
        bindCleanup: bind.cleanup,
        activeRuns: new Map(),
      };
    } catch (err) {
      bind.cleanup();
      throw err;
    }
  }
}
