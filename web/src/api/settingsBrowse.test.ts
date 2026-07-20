import { describe, expect, it } from "vitest";
import {
  parseBrowseDirsResponse,
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
