import { describe, expect, it } from "vitest";
import {
  parseDependenciesEnvelope,
  parseDependsOnList,
} from "./parseTaskApiTasks";

describe("parseDependsOnList", () => {
  it("parses string and object edges with satisfies", () => {
    expect(
      parseDependsOnList([
        "a",
        { task_id: "b", satisfies: "done" },
        { task_id: "c" },
      ]),
    ).toEqual([
      { task_id: "a", satisfies: "done" },
      { task_id: "b", satisfies: "done" },
      { task_id: "c", satisfies: "done" },
    ]);
  });

  it("throws on non-array", () => {
    expect(() => parseDependsOnList(undefined)).toThrow(/must be an array/);
    expect(() => parseDependsOnList({})).toThrow(/must be an array/);
  });
});

describe("parseDependenciesEnvelope", () => {
  it("requires a record with depends_on", () => {
    expect(() => parseDependenciesEnvelope(null)).toThrow(/must be an object/);
    expect(() => parseDependenciesEnvelope({})).toThrow(/depends_on is required/);
  });

  it("maps null depends_on to empty list", () => {
    expect(parseDependenciesEnvelope({ depends_on: null })).toEqual([]);
  });

  it("parses depends_on array", () => {
    expect(
      parseDependenciesEnvelope({ depends_on: [{ task_id: "x" }] }),
    ).toEqual([{ task_id: "x", satisfies: "done" }]);
  });

  it("throws when depends_on is not an array or null", () => {
    expect(() =>
      parseDependenciesEnvelope({ depends_on: "bad" }),
    ).toThrow(/must be an array/);
  });
});
