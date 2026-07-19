import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { render, screen, waitFor } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { requestUrl } from "@/test/requestUrl";
import type { ProjectContextItem } from "@/types";
import { ProjectContextPanel } from "./ProjectContextPanel";
import { projectQueryKeys } from "./queryKeys";

type FetchInput = RequestInfo | URL;

function jsonResponse(body: unknown, init: ResponseInit = { status: 200 }): Response {
  return new Response(JSON.stringify(body), {
    ...init,
    headers: { "content-type": "application/json", ...(init.headers ?? {}) },
  });
}

const contextItem: ProjectContextItem = {
  id: "ctx-1",
  project_id: "project-1",
  kind: "note",
  title: "API plan",
  body: "Use REST for v1.",
  created_by: "user",
  pinned: false,
  created_at: "2026-04-27T00:00:00Z",
  updated_at: "2026-04-27T00:00:00Z",
};

function renderPanel(projectId = "project-1") {
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: { retry: false, staleTime: Infinity },
      mutations: { retry: false },
    },
  });
  queryClient.setQueryData(projectQueryKeys.context(projectId), {
    items: [contextItem],
    edges: [],
    limit: 100,
  });

  return render(
    <QueryClientProvider client={queryClient}>
      <ProjectContextPanel projectId={projectId} />
    </QueryClientProvider>,
  );
}

describe("ProjectContextPanel", () => {
  beforeEach(() => {
    vi.spyOn(globalThis, "fetch").mockImplementation(async (input: FetchInput) => {
      const u = requestUrl(input);
      if (/\/projects\/[^/]+\/context/.test(u)) {
        return jsonResponse({ items: [contextItem], edges: [], limit: 100 });
      }
      return new Response(`unexpected fetch ${u}`, { status: 500 });
    });
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  it("renders loaded context items with list actions", async () => {
    renderPanel();

    await waitFor(() => {
      expect(screen.getByRole("button", { name: "Import memory file" })).toBeInTheDocument();
    });
    expect(screen.getByText("API plan")).toBeInTheDocument();
    expect(screen.queryByRole("tablist", { name: "Context view" })).not.toBeInTheDocument();
    expect(screen.queryByText(/at least two memory nodes/i)).not.toBeInTheDocument();
  });
});
