import { describe, expect, it } from "vitest";
import {
  parseBrowseDirsResponse,
  parseGitRepositoryProbeResponse,
  parseWorkspaceRootsResponse,
} from "./settingsBrowse";

describe("parseWorkspaceRootsResponse", () => {
  it("parses roots and environment", () => {
    expect(
      parseWorkspaceRootsResponse({
        environment: "native",
        roots: [
          {
            id: "home",
            path: "/home/me",
            label: "Home",
            available: true,
          },
        ],
      }),
    ).toEqual({
      environment: "native",
      roots: [
        {
          id: "home",
          path: "/home/me",
          label: "Home",
          available: true,
        },
      ],
    });
  });

  it("throws when id is empty", () => {
    expect(() =>
      parseWorkspaceRootsResponse({
        environment: "native",
        roots: [{ id: "", path: "/x", label: "X", available: true }],
      }),
    ).toThrow(/id must be a non-empty string/);
  });

  it("throws on unknown category", () => {
    expect(() =>
      parseWorkspaceRootsResponse({
        environment: "native",
        roots: [
          {
            id: "home",
            path: "/home/me",
            label: "Home",
            category: "not-a-category",
            available: true,
          },
        ],
      }),
    ).toThrow(/known browse category/);
  });
});

describe("parseBrowseDirsResponse", () => {
  it("parses directory entries", () => {
    expect(
      parseBrowseDirsResponse({
        path: "/home/me",
        is_git_repo: true,
        entries: [
          {
            name: "my-app",
            path: "/home/me/my-app",
            has_children: true,
            is_git_repo: true,
          },
        ],
      }),
    ).toEqual({
      path: "/home/me",
      is_git_repo: true,
      entries: [
        {
          name: "my-app",
          path: "/home/me/my-app",
          has_children: true,
          is_git_repo: true,
        },
      ],
    });
  });

  it("throws when entry name is missing", () => {
    expect(() =>
      parseBrowseDirsResponse({
        entries: [{ path: "/x", has_children: false, is_git_repo: false }],
      }),
    ).toThrow(/name must be a non-empty string/);
  });
});

describe("parseGitRepositoryProbeResponse", () => {
  it("parses main_path and is_main for a checkout", () => {
    expect(
      parseGitRepositoryProbeResponse({
        path: "/repos/wt-linked",
        main_path: "/repos/main",
        is_main: false,
        is_git_repository: true,
        current_branch: "linked",
        branches: [{ name: "linked", head_sha: "abc" }],
      }),
    ).toEqual({
      path: "/repos/wt-linked",
      main_path: "/repos/main",
      is_main: false,
      is_git_repository: true,
      current_branch: "linked",
      branches: [{ name: "linked", head_sha: "abc" }],
    });
  });
});
