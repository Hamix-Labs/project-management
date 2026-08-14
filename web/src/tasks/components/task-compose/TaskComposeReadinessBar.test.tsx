import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import {
  buildComposeReadinessChecks,
  TaskComposeReadinessBar,
} from "./TaskComposeReadinessBar";

describe("buildComposeReadinessChecks", () => {
  it("reports all four checks ready when essentials are filled", () => {
    const checks = buildComposeReadinessChecks({
      title: "Ship observability",
      brief: "Add structured logs",
      repositoryId: "repo-1",
      checklistItems: [{ text: "Docs updated" }],
    });
    expect(checks.every((c) => c.ok)).toBe(true);
    expect(checks.map((c) => c.label)).toEqual([
      "Title",
      "Brief",
      "Repository",
      "Done criteria",
    ]);
  });

  it("mirrors TaskCreateModalFooterActions git + checklist gates", () => {
    const incomplete = buildComposeReadinessChecks({
      title: "Title",
      brief: "Brief",
      repositoryId: "",
      checklistItems: [],
    });
    expect(incomplete.find((c) => c.label === "Repository")?.ok).toBe(false);
    expect(incomplete.find((c) => c.label === "Done criteria")?.ok).toBe(
      false,
    );

    const whitespaceCriteria = buildComposeReadinessChecks({
      title: "Title",
      brief: "Brief",
      repositoryId: "repo-1",
      checklistItems: [{ text: "   " }],
    });
    expect(
      whitespaceCriteria.find((c) => c.label === "Done criteria")?.ok,
    ).toBe(false);
  });
});

describe("TaskComposeReadinessBar", () => {
  it("shows Ready to hand off when all checks pass", () => {
    render(
      <TaskComposeReadinessBar
        title="Ship observability"
        brief="Add structured logs"
        repositoryId="repo-1"
        checklistItems={[{ text: "Docs updated" }]}
      />,
    );
    expect(screen.getByText("Ready to hand off")).toBeInTheDocument();
  });

  it("shows N/4 essentials ready when incomplete", () => {
    render(
      <TaskComposeReadinessBar
        title="Ship observability"
        brief=""
        repositoryId=""
        checklistItems={[]}
      />,
    );
    expect(screen.getByText("1/4")).toBeInTheDocument();
    expect(screen.getByText(/essentials ready/)).toBeInTheDocument();
  });
});
