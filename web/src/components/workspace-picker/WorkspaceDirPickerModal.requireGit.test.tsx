import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";
import { WorkspaceDirPickerModal } from "./WorkspaceDirPickerModal";
import { browseRouter, jsonResponse } from "./workspacePickerTestUtils";

describe("WorkspaceDirPickerModal requireGit", () => {
  it("blocks confirmation when requireGitRepository and folder is not a git checkout", async () => {
    const onSelect = vi.fn();
    const fetchMock = vi.spyOn(globalThis, "fetch").mockImplementation(async (input) => {
      const url = String(input);
      if (url.endsWith("/settings/workspace-roots")) {
        return jsonResponse({
          environment: "native",
          roots: [{ id: "home", path: "/roots", label: "Home", category: "home", available: true }],
        });
      }
      if (url.includes("/settings/browse-dirs")) {
        return browseRouter({
          "/roots": {
            path: "/roots",
            parent_path: "",
            is_git_repo: false,
            entries: [
              {
                name: "my-app",
                path: "/roots/my-app",
                has_children: false,
                is_git_repo: true,
              },
            ],
          },
          "/roots/my-app": {
            path: "/roots/my-app",
            parent_path: "/roots",
            is_git_repo: true,
            entries: [],
          },
        })(url);
      }
      if (url.includes("/settings/git-probe")) {
        const u = new URL(url, "http://test.local");
        const path = u.searchParams.get("path") ?? "";
        return jsonResponse({
          path,
          main_path: path,
          is_main: true,
          is_git_repository: true,
          current_branch: "main",
          branches: [{ name: "main", head_sha: "abc" }],
        });
      }
      return new Response("not found", { status: 404 });
    });

    render(
      <WorkspaceDirPickerModal
        open
        requireGitRepository
        currentPath=""
        onClose={() => {}}
        onSelect={onSelect}
      />,
    );

    await userEvent.click(await screen.findByRole("button", { name: /Home/ }));

    await waitFor(() => {
      expect(screen.getByRole("button", { name: /Use this repository/ })).toBeDisabled();
    });
    expect(screen.getByText("Select a repository above to continue.")).toBeInTheDocument();

    await userEvent.click(screen.getByRole("button", { name: /^my-app/ }));

    await waitFor(() => {
      expect(screen.getByRole("button", { name: /Use this repository/ })).toBeEnabled();
    });
    expect(screen.getByText("/roots/my-app")).toBeInTheDocument();
    expect(screen.getByText("Repository to register")).toBeInTheDocument();

    await userEvent.click(screen.getByRole("button", { name: /Use this repository/ }));
    expect(onSelect).toHaveBeenCalledWith("/roots/my-app");
    fetchMock.mockRestore();
  });

  it("resolves a linked git folder to the main path for registration", async () => {
    const onSelect = vi.fn();
    const fetchMock = vi.spyOn(globalThis, "fetch").mockImplementation(async (input) => {
      const url = String(input);
      if (url.endsWith("/settings/workspace-roots")) {
        return jsonResponse({
          environment: "native",
          roots: [{ id: "home", path: "/roots", label: "Home", category: "home", available: true }],
        });
      }
      if (url.includes("/settings/browse-dirs")) {
        return browseRouter({
          "/roots": {
            path: "/roots",
            parent_path: "",
            is_git_repo: false,
            entries: [
              {
                name: "wt-linked",
                path: "/roots/wt-linked",
                has_children: false,
                is_git_repo: true,
              },
            ],
          },
        })(url);
      }
      if (url.includes("/settings/git-probe")) {
        return jsonResponse({
          path: "/roots/wt-linked",
          main_path: "/roots/my-app",
          is_main: false,
          is_git_repository: true,
          current_branch: "linked",
          branches: [{ name: "linked", head_sha: "def" }],
        });
      }
      return new Response("not found", { status: 404 });
    });

    render(
      <WorkspaceDirPickerModal
        open
        requireGitRepository
        currentPath=""
        onClose={() => {}}
        onSelect={onSelect}
      />,
    );

    await userEvent.click(await screen.findByRole("button", { name: /Home/ }));
    await userEvent.click(await screen.findByRole("button", { name: /^wt-linked/ }));

    await waitFor(() => {
      expect(screen.getByText("/roots/my-app")).toBeInTheDocument();
    });
    expect(screen.getByText("Repository to register")).toBeInTheDocument();

    await userEvent.click(screen.getByRole("button", { name: /Use this repository/ }));
    expect(onSelect).toHaveBeenCalledWith("/roots/my-app");
    fetchMock.mockRestore();
  });

  it("opens a git folder via the Open control without selecting it", async () => {
    const fetchMock = vi.spyOn(globalThis, "fetch").mockImplementation(async (input) => {
      const url = String(input);
      if (url.endsWith("/settings/workspace-roots")) {
        return jsonResponse({
          environment: "native",
          roots: [{ id: "home", path: "/roots", label: "Home", category: "home", available: true }],
        });
      }
      if (url.includes("/settings/browse-dirs")) {
        return browseRouter({
          "/roots": {
            path: "/roots",
            parent_path: "",
            is_git_repo: false,
            entries: [
              {
                name: "my-app",
                path: "/roots/my-app",
                has_children: true,
                is_git_repo: true,
              },
            ],
          },
          "/roots/my-app": {
            path: "/roots/my-app",
            parent_path: "/roots",
            is_git_repo: true,
            entries: [
              {
                name: "src",
                path: "/roots/my-app/src",
                has_children: false,
                is_git_repo: false,
              },
            ],
          },
        })(url);
      }
      if (url.includes("/settings/git-probe")) {
        return jsonResponse({
          path: "/roots/my-app",
          main_path: "/roots/my-app",
          is_main: true,
          is_git_repository: true,
          current_branch: "main",
          branches: [{ name: "main", head_sha: "abc" }],
        });
      }
      return new Response("not found", { status: 404 });
    });

    render(
      <WorkspaceDirPickerModal
        open
        requireGitRepository
        currentPath=""
        onClose={() => {}}
        onSelect={() => {}}
      />,
    );

    await userEvent.click(await screen.findByRole("button", { name: /Home/ }));
    await userEvent.click(await screen.findByRole("button", { name: /Open my-app/ }));

    await waitFor(() => {
      expect(screen.getByRole("button", { name: /src/ })).toBeInTheDocument();
    });
    await waitFor(() => {
      expect(screen.getByRole("button", { name: /Use this repository/ })).toBeEnabled();
    });
    expect(screen.getByText("/roots/my-app")).toBeInTheDocument();
    fetchMock.mockRestore();
  });

  it("skips roots and opens at initialBrowsePath when set", async () => {
    const fetchMock = vi.spyOn(globalThis, "fetch").mockImplementation(async (input) => {
      const url = String(input);
      if (url.endsWith("/settings/workspace-roots")) {
        return jsonResponse({
          environment: "native",
          roots: [
            {
              id: "repo-main",
              path: "/stale/repo",
              label: "my-repo",
              category: "install",
              available: false,
              unavailable_reason: "directory is not accessible",
            },
          ],
        });
      }
      if (url.includes("/settings/browse-dirs")) {
        return browseRouter({
          "/parent": {
            path: "/parent",
            parent_path: "",
            entries: [
              {
                name: "Hamix",
                path: "/parent/Hamix",
                has_children: false,
                is_git_repo: true,
              },
            ],
          },
          "/parent/Hamix": {
            path: "/parent/Hamix",
            parent_path: "/parent",
            is_git_repo: true,
            entries: [],
          },
        })(url);
      }
      return new Response("not found", { status: 404 });
    });

    render(
      <WorkspaceDirPickerModal
        open
        currentPath="/stale/repo"
        initialBrowsePath="/parent"
        onClose={() => {}}
        onSelect={() => {}}
      />,
    );

    await waitFor(() => {
      expect(screen.getAllByText("/parent").length).toBeGreaterThan(0);
    });
    expect(screen.queryByRole("button", { name: /my-repo/ })).not.toBeInTheDocument();
    expect(await screen.findByRole("button", { name: /Hamix/ })).toBeInTheDocument();

    fetchMock.mockRestore();
  });

  it("requests expanded workspace roots when rootsScope is expanded", async () => {
    const fetchMock = vi.spyOn(globalThis, "fetch").mockImplementation(async (input) => {
      const url = String(input);
      if (url.includes("/settings/workspace-roots")) {
        expect(url).toContain("scope=expanded");
        return jsonResponse({
          environment: "native",
          roots: [
            { id: "repo", path: "/repo", label: "repo", category: "registered", available: true },
            { id: "home", path: "/roots", label: "Home", category: "home", available: true },
          ],
        });
      }
      return new Response("not found", { status: 404 });
    });

    render(
      <WorkspaceDirPickerModal
        open
        rootsScope="expanded"
        currentPath=""
        onClose={() => {}}
        onSelect={() => {}}
      />,
    );

    expect(await screen.findByRole("button", { name: /repo/ })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /Home/ })).toBeInTheDocument();

    fetchMock.mockRestore();
  });
});
