import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { render, screen, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import type { ReactNode } from "react";
import { describe, expect, it, vi } from "vitest";
import type { ProjectContextItem } from "@/types";
import { ModalStackProvider } from "@/shared/ModalStackContext";
import { ProjectContextPicker } from "./ProjectContextPicker";
import { projectQueryKeys } from "@/lib/projectQueryKeys";

const projectId = "project-1";

const contextItems: ProjectContextItem[] = [
  {
    id: "ctx-risk",
    project_id: projectId,
    kind: "risk",
    title: "New One",
    description: "",
    body: "Risk details",
    created_by: "user",
    pinned: false,
    created_at: "2026-04-29T00:00:00Z",
    updated_at: "2026-04-29T00:00:00Z",
  },
  {
    id: "ctx-decision",
    project_id: projectId,
    kind: "decision",
    title: "Decision node",
    description: "",
    body: "Decision details",
    created_by: "user",
    pinned: false,
    created_at: "2026-04-29T00:00:00Z",
    updated_at: "2026-04-29T00:00:00Z",
  },
];

function renderPicker(selectedIds: string[] = []) {
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: { retry: false, staleTime: Infinity },
      mutations: { retry: false },
    },
  });
  queryClient.setQueryData(projectQueryKeys.context(projectId), {
    items: contextItems,
    edges: [],
    limit: 100,
  });
  const onChange = vi.fn();

  function Wrapper({ children }: { children: ReactNode }) {
    return (
      <QueryClientProvider client={queryClient}>
        <ModalStackProvider>{children}</ModalStackProvider>
      </QueryClientProvider>
    );
  }

  render(
    <ProjectContextPicker
      projectId={projectId}
      selectedIds={selectedIds}
      onChange={onChange}
    />,
    { wrapper: Wrapper },
  );

  return { onChange };
}

describe("ProjectContextPicker", () => {
  it("adds a node immediately when selected from the chooser list", async () => {
    const user = userEvent.setup();
    const { onChange } = renderPicker();

    await user.click(screen.getByRole("button", { name: /choose context/i }));
    const dialog = screen.getByRole("dialog", { name: /choose task context/i });

    await user.type(
      within(dialog).getByPlaceholderText(/search by title, body, or kind/i),
      "risk",
    );

    expect(within(dialog).getByText("New One")).toBeInTheDocument();
    expect(within(dialog).queryByText("Decision node")).not.toBeInTheDocument();

    await user.click(within(dialog).getByRole("checkbox", { name: /select new one/i }));

    expect(onChange).toHaveBeenCalledWith(["ctx-risk"]);
    expect(
      screen.queryByRole("dialog", { name: /reference project context/i }),
    ).not.toBeInTheDocument();
  });

  it("removes a selected node from the summary chip list without prompting", async () => {
    const user = userEvent.setup();
    const { onChange } = renderPicker(["ctx-risk", "ctx-decision"]);

    expect(screen.getByText(/2 nodes selected/i)).toBeInTheDocument();
    await user.click(
      screen.getByRole("button", {
        name: /remove reference to new one/i,
      }),
    );
    expect(onChange).toHaveBeenCalledWith(["ctx-decision"]);
  });

  it("renders the selected items with title plus short id in the summary", () => {
    renderPicker(["ctx-risk"]);
    const chip = screen.getByText("New One").closest(".project-context-picker__chip");
    expect(chip).not.toBeNull();
    // The short-id helper strips dashes and lowercases the first six alnums.
    expect(chip?.textContent).toContain("ctxris");
  });

  it("renders compact summary copy for the create-task modal", () => {
    const queryClient = new QueryClient({
      defaultOptions: {
        queries: { retry: false, staleTime: Infinity },
        mutations: { retry: false },
      },
    });
    queryClient.setQueryData(projectQueryKeys.context(projectId), {
      items: contextItems,
      edges: [],
      limit: 100,
    });

    render(
      <QueryClientProvider client={queryClient}>
        <ModalStackProvider>
          <ProjectContextPicker
            projectId={projectId}
            selectedIds={[]}
            compact
            onChange={vi.fn()}
          />
        </ModalStackProvider>
      </QueryClientProvider>,
    );

    expect(screen.getByText("0 selected")).toBeInTheDocument();
    expect(screen.getByText("No context attached yet")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /^choose$/i })).toBeInTheDocument();
  });

  it("lists context nodes without a tree view toggle", async () => {
    const user = userEvent.setup();
    const { onChange } = renderPicker();

    await user.click(screen.getByRole("button", { name: /choose context/i }));
    const dialog = screen.getByRole("dialog", { name: /choose task context/i });

    expect(
      within(dialog).queryByRole("tablist", { name: /context chooser view/i }),
    ).not.toBeInTheDocument();
    expect(within(dialog).getByText("Decision node")).toBeInTheDocument();

    await user.click(
      within(dialog).getByRole("checkbox", { name: /select decision node/i }),
    );

    expect(onChange).toHaveBeenCalledWith(["ctx-decision"]);
    expect(
      screen.queryByRole("dialog", { name: /reference project context/i }),
    ).not.toBeInTheDocument();
  });
});
