import { afterAll, afterEach, beforeAll } from "vitest";
import { server } from "./server";

/**
 * Opt-in MSW for component suites that need network interception.
 * Call once at module top (side effect registers lifecycle hooks).
 */
let listening = false;

export function ensureMswListening(
  onUnhandledRequest: "bypass" | "error" = "bypass",
): void {
  if (listening) return;
  listening = true;

  beforeAll(() => {
    server.listen({ onUnhandledRequest });
  });

  afterEach(() => {
    server.resetHandlers();
  });

  afterAll(() => {
    server.close();
    listening = false;
  });
}
