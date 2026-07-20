import { cleanup } from "@testing-library/react";
import "@testing-library/jest-dom/vitest";
import { afterAll, afterEach, beforeAll } from "vitest";
import { server } from "./server";

/**
 * Full-app Vitest projects set HAMIX_MSW_UNHANDLED=error so a forgotten
 * handler fails loudly. unit/components keep bypass so legacy
 * vi.spyOn(fetch) suites still work until migrated.
 */
const unhandledMode =
  process.env.HAMIX_MSW_UNHANDLED === "error" ? "error" : "bypass";

beforeAll(() => {
  server.listen({ onUnhandledRequest: unhandledMode });
});

afterEach(() => {
  server.resetHandlers();
  cleanup();
});

afterAll(() => {
  server.close();
});
