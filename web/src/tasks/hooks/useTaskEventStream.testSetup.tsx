import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import type { ReactNode } from "react";
import { afterEach, beforeEach, vi } from "vitest";

export type MockES = {
  onopen: (() => void) | null;
  onmessage: ((ev: { data?: string }) => void) | null;
  onerror: (() => void) | null;
  close: ReturnType<typeof vi.fn>;
  readyState: number;
};

export let getCurrentMockES: () => MockES | null;

export function createWrapper(qc: QueryClient) {
  return function Wrapper({ children }: { children: ReactNode }) {
    return <QueryClientProvider client={qc}>{children}</QueryClientProvider>;
  };
}

beforeEach(() => {
  vi.useFakeTimers();
  class MockEventSource implements MockES {
    static latest: MockEventSource | null = null;
    onopen: (() => void) | null = null;
    onmessage: ((ev: { data?: string }) => void) | null = null;
    onerror: (() => void) | null = null;
    close = vi.fn();
    readyState = 0;
    constructor() {
      MockEventSource.latest = this;
    }
  }
  getCurrentMockES = () => MockEventSource.latest;
  vi.stubGlobal("EventSource", MockEventSource);
});

afterEach(() => {
  vi.unstubAllGlobals();
  vi.useRealTimers();
});
