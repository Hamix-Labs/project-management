import { createServer } from "node:http";
import type { AddressInfo } from "node:net";
import { AgentSessionRegistry } from "./agentSession.js";
import { buildRouter } from "./routes.js";
import { SdkAgentPort } from "./sdkAgentPort.js";
import type { AgentPort } from "./agentPort.js";

export interface StartOptions {
  port: number;
  host?: string;
  agentPort?: AgentPort;
  getApiKey?: () => string | undefined;
  taskapiBaseUrl?: string;
}

export interface RunningServer {
  port: number;
  close(): Promise<void>;
}

// startServer boots the HTTP listener and returns handles for tests. In
// production the CLI wrapper below calls this, prints the actual bound port
// (so the Go supervisor can discover a random port), and installs shutdown
// handlers.
export async function startServer(opts: StartOptions): Promise<RunningServer> {
  const agentPort: AgentPort = opts.agentPort ?? new SdkAgentPort();
  const registry = new AgentSessionRegistry(agentPort);
  const getApiKey = opts.getApiKey ?? (() => process.env.CURSOR_API_KEY);
  const handler = buildRouter({
    port: agentPort,
    registry,
    getApiKey,
    ...(opts.taskapiBaseUrl ? { taskapiBaseUrl: opts.taskapiBaseUrl } : {}),
  });
  const server = createServer((req, res) => {
    void handler(req, res).catch((err) => {
      // Last-resort log; individual routes catch and translate their own
      // errors. This only fires if the router itself throws synchronously.
      // eslint-disable-next-line no-console
      console.error("router error:", err);
      if (!res.headersSent) {
        res.writeHead(500, { "content-type": "application/json" });
        res.end(JSON.stringify({ error: "internal error" }));
      }
    });
  });

  await new Promise<void>((resolve, reject) => {
    server.once("error", reject);
    server.listen(opts.port, opts.host ?? "127.0.0.1", () => {
      server.off("error", reject);
      resolve();
    });
  });

  const addr = server.address() as AddressInfo | null;
  const boundPort = addr?.port ?? opts.port;

  return {
    port: boundPort,
    async close(): Promise<void> {
      await registry.closeAll();
      await new Promise<void>((resolve) => server.close(() => resolve()));
    },
  };
}

// parseCli reads --port <n> from argv (0 = ephemeral). Falls back to the PORT
// env var, then 0.
export function parseCli(argv: string[]): { port: number } {
  let port: number | null = null;
  for (let i = 0; i < argv.length; i++) {
    if (argv[i] === "--port" && i + 1 < argv.length) {
      const next = argv[i + 1];
      if (typeof next === "string") {
        const parsed = Number.parseInt(next, 10);
        if (Number.isFinite(parsed) && parsed >= 0) port = parsed;
      }
    }
  }
  if (port === null) {
    const envPort = process.env.PORT;
    if (envPort) {
      const parsed = Number.parseInt(envPort, 10);
      if (Number.isFinite(parsed) && parsed >= 0) port = parsed;
    }
  }
  if (port === null) port = 0;
  return { port };
}

// runCli is the executable entry, invoked by src/cli.ts. It prints
// "listening on <port>" so a Go supervisor can discover an ephemeral port,
// then blocks until a signal disposes the server.
export async function runCli(argv: string[]): Promise<void> {
  const { port } = parseCli(argv);
  const running = await startServer({ port });
  // eslint-disable-next-line no-console
  console.log(`listening on ${running.port}`);
  const shutdown = async (): Promise<void> => {
    try {
      await running.close();
    } finally {
      process.exit(0);
    }
  };
  process.on("SIGINT", () => void shutdown());
  process.on("SIGTERM", () => void shutdown());
  await new Promise(() => undefined);
}
