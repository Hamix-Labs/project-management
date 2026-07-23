import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";
import {
  RepoFileSuggestionList,
  splitPath,
} from "./repoFileSuggestionList";

describe("splitPath", () => {
  it("splits directory and basename", () => {
    expect(splitPath(".cursor/rules/CODE_STANDARDS.mdc")).toEqual([
      ".cursor/rules/",
      "CODE_STANDARDS.mdc",
    ]);
  });

  it("returns empty dir for a bare filename", () => {
    expect(splitPath("README.md")).toEqual(["", "README.md"]);
  });
});

describe("RepoFileSuggestionList", () => {
  it("shows the search placeholder when query is empty", () => {
    render(
      <RepoFileSuggestionList items={[]} command={vi.fn()} query="" />,
    );
    expect(screen.getByText(/search repository files/i)).toBeInTheDocument();
    expect(screen.getByText(/no matching files/i)).toBeInTheDocument();
    expect(screen.getByText(/to navigate/i)).toBeInTheDocument();
    expect(screen.getByText(/to select/i)).toBeInTheDocument();
  });

  it("shows the live query in the search header", () => {
    render(
      <RepoFileSuggestionList
        items={[]}
        command={vi.fn()}
        query="concurrency"
      />,
    );
    expect(screen.getByText("concurrency")).toBeInTheDocument();
    expect(
      screen.getByText(/no files match .concurrency./i),
    ).toBeInTheDocument();
  });

  it("renders split path rows and marks the selected index", () => {
    render(
      <RepoFileSuggestionList
        items={[
          { path: ".codegraph/.gitignore" },
          { path: ".cursor/rules/CODE_STANDARDS.mdc" },
        ]}
        command={vi.fn()}
        selectedIndex={1}
      />,
    );
    expect(screen.getByText(".codegraph/")).toBeInTheDocument();
    expect(screen.getByText(".gitignore")).toBeInTheDocument();
    expect(screen.getByText(".cursor/rules/")).toBeInTheDocument();
    expect(screen.getByText("CODE_STANDARDS.mdc")).toBeInTheDocument();

    const options = screen.getAllByRole("option");
    expect(options[0]).toHaveAttribute("aria-selected", "false");
    expect(options[1]).toHaveAttribute("aria-selected", "true");
    expect(options[1]).toHaveClass("mention-option--active");
  });

  it("invokes command when a row is clicked", async () => {
    const user = userEvent.setup();
    const command = vi.fn();
    render(
      <RepoFileSuggestionList
        items={[{ path: "cmd/server/main.go" }]}
        command={command}
      />,
    );
    await user.click(screen.getByRole("button", { name: /main\.go/i }));
    expect(command).toHaveBeenCalledWith({ path: "cmd/server/main.go" });
  });
});
