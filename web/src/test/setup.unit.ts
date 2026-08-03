import { cleanup } from "@testing-library/react";
import "@testing-library/jest-dom/vitest";
import { afterEach } from "vitest";

/**
 * Slim unit setup: no global MSW. Most unit files run in node; the few that
 * use renderHook/document set // @vitest-environment jsdom at file top.
 */
afterEach(() => {
  if (typeof document !== "undefined") {
    cleanup();
  }
});
