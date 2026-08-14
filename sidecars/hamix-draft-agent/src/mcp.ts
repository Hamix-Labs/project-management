import { mkdirSync, writeFileSync, rmSync } from "node:fs";
import { tmpdir } from "node:os";
import { join } from "node:path";
import { randomUUID } from "node:crypto";
import type { BindFile } from "./types.js";

// BindWriteResult carries the path and cleanup callback for a per-session bind
// file. taskapi will own the canonical WriteBind in Plan 4; here we stub the
// filesystem write so the sidecar is self-sufficient in tests and dev.
export interface BindWriteResult {
  bindPath: string;
  cleanup: () => void;
}

// writeBindFile drops the bind JSON into a per-session temp directory. The
// caller passes the path as `--bind` to hamix-draft-mcp. Schema mirrors
// pkgs/draftassist/mcp.BindFile so the Go tool accepts what we produce.
export function writeBindFile(
  sessionID: string,
  nonce: string,
  taskapiBaseUrl?: string,
): BindWriteResult {
  const dir = join(tmpdir(), `hamix-draft-agent-${randomUUID()}`);
  mkdirSync(dir, { recursive: true });
  const bindPath = join(dir, "bind.json");
  const bind: BindFile = {
    bind_schema_version: 1,
    session_id: sessionID,
    nonce,
    ...(taskapiBaseUrl ? { taskapi_base_url: taskapiBaseUrl } : {}),
  };
  writeFileSync(bindPath, JSON.stringify(bind, null, 2), { encoding: "utf8" });
  return {
    bindPath,
    cleanup: () => {
      try {
        rmSync(dir, { recursive: true, force: true });
      } catch {
        // best effort; the OS will reap the tmp dir eventually.
      }
    },
  };
}

// mcpStdioEntry builds the value for mcpServers["hamix-draft"] on Agent.create.
// hamix-draft-mcp is expected to be on PATH (built alongside taskapi in
// scripts/dev.*). The binary loads the bind JSON at startup.
export function mcpStdioEntry(bindPath: string): {
  command: string;
  args: string[];
} {
  return {
    command: "hamix-draft-mcp",
    args: ["--bind", bindPath],
  };
}
