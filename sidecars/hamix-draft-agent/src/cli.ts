#!/usr/bin/env node
// CLI entry for the built hamix-draft-agent binary. See server.ts for the
// library surface. Kept minimal so esbuild produces a small bundle.
import { runCli } from "./server.js";

runCli(process.argv.slice(2)).catch((err) => {
  // eslint-disable-next-line no-console
  console.error("hamix-draft-agent:", err);
  process.exit(1);
});
