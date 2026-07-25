import "./useTasksApp.testSetup";
import { act, renderHook, waitFor } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { useTasksApp } from "./useTasksApp";
import { stubEventSource } from "../../test/browserMocks";
import { makeTask } from "@/test/taskDefaults";
import {
  makeWrapper,
  mockedCloseTask,
  mockedEnsureRepos,
  mockedGetStats,
  mockedListDrafts,
  mockedListTasks,
} from "./useTasksApp.testSetup";

describe("useTasksApp close lifecycle", () => {
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
    mockedCloseTask.mockReset();
  });

  it("keeps successful close variables available after the confirm dialog closes", async () => {
    mockedCloseTask.mockResolvedValueOnce(
      makeTask({ id: "root-task", status: "closed" }),
    );
    const { Wrapper } = makeWrapper();
    const { result } = renderHook(() => useTasksApp({ sseLive: false }), {
      wrapper: Wrapper,
    });

    act(() => {
      result.current.requestClose({
        id: "root-task",
        title: "Root task",
      });
    });
    act(() => {
      result.current.confirmClose();
    });

    await waitFor(() => {
      expect(result.current.closeTarget).toBeNull();
    });
    await waitFor(() => {
      expect(result.current.closeSuccess).toBe(true);
      expect(result.current.closeVariables).toEqual({ id: "root-task" });
    });
  });
});
