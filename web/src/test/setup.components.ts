import { cleanup } from "@testing-library/react";
import "@testing-library/jest-dom/vitest";
import { afterEach } from "vitest";
import { installEditorDomStubs } from "./setup-editor-dom";

installEditorDomStubs();

/**
 * Slim components setup: jsdom + jest-dom, no global MSW.
 * Suites that need interception call ensureMswListening() from mswLifecycle.
 */
afterEach(() => {
  cleanup();
});
