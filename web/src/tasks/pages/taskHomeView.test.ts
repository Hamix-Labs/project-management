import { describe, expect, it } from "vitest";
import { applyTaskHomeView, parseTaskHomeView } from "./taskHomeView";

describe("parseTaskHomeView", () => {
  it("defaults unknown and missing to list", () => {
    expect(parseTaskHomeView(null)).toBe("list");
    expect(parseTaskHomeView(undefined)).toBe("list");
    expect(parseTaskHomeView("")).toBe("list");
    expect(parseTaskHomeView("table")).toBe("list");
    expect(parseTaskHomeView("list")).toBe("list");
  });

  it("accepts board", () => {
    expect(parseTaskHomeView("board")).toBe("board");
  });
});

describe("applyTaskHomeView", () => {
  it("omits view for list and sets board, preserving other params", () => {
    const base = new URLSearchParams("project=p1&view=board");
    expect(applyTaskHomeView(base, "list").toString()).toBe("project=p1");
    expect(applyTaskHomeView(base, "board").toString()).toBe(
      "project=p1&view=board",
    );
  });
});
