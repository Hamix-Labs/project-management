import "./useTasksApp.testSetup";
import { act, renderHook, waitFor } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { useTasksApp } from "./useTasksApp";
import { stubEventSource } from "../../test/browserMocks";
import {
  makeWrapper,
  mockedDeleteTask,
  mockedEnsureRepos,
  mockedGetStats,
  mockedListDrafts,
  mockedListTasks,
} from "./useTasksApp.testSetup";

describe("useTasksApp delete lifecycle", () => {
  beforeEach(() => {
    stubEventSource();
    mockedListTasks.mockResolvedValue({
      tasks: [],
      limit: 200,
      offset: 0,
      has_more: false,
    });
    mockedGetStats.mockResolvedValue(null as unknown as never);
    mockedListDrafts.mockResolvedValue([]);
    mockedEnsureRepos.mockResolvedValue(true);
  });

  afterEach(() => {
    vi.restoreAllMocks();
    vi.unstubAllGlobals();
    mockedDeleteTask.mockReset();
  });

  it("keeps successful delete variables available after the confirm dialog closes", async () => {
    mockedDeleteTask.mockResolvedValueOnce(undefined as unknown as void);
    const { Wrapper } = makeWrapper();
    const { result } = renderHook(() => useTasksApp({ sseLive: false }), {
      wrapper: Wrapper,
    });

    act(() => {
      result.current.requestDelete({
        id: "root-task",
        title: "Root task",
      });
    });
    act(() => {
      result.current.confirmDelete();
    });

    await waitFor(() => {
      expect(result.current.deleteTarget).toBeNull();
    });
    await waitFor(() => {
      expect(result.current.deleteSuccess).toBe(true);
      expect(result.current.deleteVariables).toEqual({ id: "root-task" });
    });
  });
});
