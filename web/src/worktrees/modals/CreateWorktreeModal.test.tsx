import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";
import { worktreeGitCopy } from "../worktreeGitCopy";
import { CreateWorktreeModal } from "./CreateWorktreeModal";

vi.mock("@/components/workspace-picker", () => ({
  WorkspaceDirPickerModal: ({
    open,
    onSelect,
  }: {
    open: boolean;
    onSelect: (path: string) => void;
  }) =>
    open ? (
      <button type="button" onClick={() => onSelect("/repo")}>
        Mock pick parent folder
      </button>
    ) : null,
}));

function jsonResponse(body: unknown, init: ResponseInit = { status: 200 }): Response {
  return new Response(JSON.stringify(body), {
    ...init,
    headers: { "content-type": "application/json", ...(init.headers ?? {}) },
  });
}

function renderModal(overrides: Partial<Parameters<typeof CreateWorktreeModal>[0]> = {}) {
  const onSubmit = vi.fn();
  const client = new QueryClient({
    defaultOptions: { queries: { retry: false, gcTime: 0 } },
  });
  render(
    <QueryClientProvider client={client}>
      <CreateWorktreeModal
        open
        pending={false}
        error={null}
        repositoryId="00000000-0000-4000-8000-000000000010"
        storedPath="/repo/main"
        onReconcile={() => {}}
        onClose={() => {}}
        onSubmit={onSubmit}
        {...overrides}
      />
    </QueryClientProvider>,
  );
  return { onSubmit };
}

describe("CreateWorktreeModal", () => {
  it("shows new checkout location copy instead of worktree path wording", async () => {
    const fetchMock = vi.spyOn(globalThis, "fetch").mockImplementation(async (input) => {
      const url = String(input);
      if (url.includes("/worktrees/live")) {
        return jsonResponse({ worktrees: [] });
      }
      if (url.includes("/branches")) {
        return jsonResponse({ branches: [] });
      }
      return new Response("not found", { status: 404 });
    });

    renderModal();

    expect(await screen.findByText(worktreeGitCopy.createModalLocationLabel)).toBeInTheDocument();
    expect(screen.getByText(worktreeGitCopy.createModalLocationHint)).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /change parent folder/i })).toBeInTheDocument();
    expect(screen.getByText(worktreeGitCopy.createModalFolderNameHint)).toBeInTheDocument();
    expect(screen.getByText(worktreeGitCopy.createModalDisplayNameHint)).toBeInTheDocument();
    expect(screen.getByLabelText(/checkout folder name/i)).toBeRequired();
    expect(screen.queryByText(/choose worktree path/i)).not.toBeInTheDocument();

    fetchMock.mockRestore();
  });

  it("preselects the main repository parent folder when the modal opens", async () => {
    const fetchMock = vi.spyOn(globalThis, "fetch").mockImplementation(async (input) => {
      const url = String(input);
      if (url.includes("/worktrees/live")) {
        return jsonResponse({ worktrees: [] });
      }
      if (url.includes("/branches")) {
        return jsonResponse({ branches: [] });
      }
      return new Response("not found", { status: 404 });
    });

    renderModal({ storedPath: "C:/Users/gomes/OneDrive/Documents/Hamix" });

    expect(await screen.findByText(worktreeGitCopy.createModalLocationLabel)).toBeInTheDocument();
    expect(screen.getByText("Default", { selector: ".worktrees-form-modal__parent-default-badge" })).toBeInTheDocument();
    expect(screen.getByText("C:/Users/gomes/OneDrive/Documents", { selector: "code" })).toBeInTheDocument();

    fetchMock.mockRestore();
  });

  it("submits combined parent folder and new directory name as path", async () => {
    const fetchMock = vi.spyOn(globalThis, "fetch").mockImplementation(async (input) => {
      const url = String(input);
      if (url.includes("/worktrees/live")) {
        return jsonResponse({ worktrees: [] });
      }
      if (url.includes("/branches/live")) {
        return jsonResponse({ branches: [{ name: "main", head: "abc" }] });
      }
      if (url.includes("/branches")) {
        return jsonResponse({ branches: [] });
      }
      return new Response("not found", { status: 404 });
    });

    const user = userEvent.setup();
    const { onSubmit } = renderModal();

    await screen.findByText("/repo", { selector: "code" });
    await user.type(screen.getByLabelText(/checkout folder name/i), "feature-a");
    await user.click(screen.getByRole("checkbox", { name: /create a new branch/i }));
    await user.type(screen.getByLabelText(/new branch name/i), "feature-a");
    await user.click(screen.getByRole("button", { name: /^create worktree$/i }));

    await waitFor(() => {
      expect(onSubmit).toHaveBeenCalledWith({
        path: "/repo/feature-a",
        branch: "feature-a",
        create_branch: true,
      });
    });

    fetchMock.mockRestore();
  });
});
