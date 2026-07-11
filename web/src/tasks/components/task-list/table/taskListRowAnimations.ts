import { useEffect, useLayoutEffect, useMemo, useRef, useState } from "react";
import type { MutableRefObject } from "react";
import type { TaskWithDepth } from "../../../task-tree";
import { computeTaskListDisplayOrder } from "./taskListDisplayOrder";

/**
 * Matches the `task-list-row-fade-out` keyframe duration in
 * app-task-list-and-mentions.css (--duration-normal ≈ 200ms).
 */
export const ROW_EXIT_MS = 220;

export type ExitingRow = { task: TaskWithDepth; timeoutId: number };

export type TaskListRowRenderState = {
  task: TaskWithDepth;
  isEntering: boolean;
  isExiting: boolean;
  isFilterExit: boolean;
};

type TaskListRowAnimationRefs = {
  seenIdsRef: MutableRefObject<Set<string>>;
  exitingRef: MutableRefObject<Map<string, ExitingRow>>;
  filterExitingRef: MutableRefObject<Map<string, TaskWithDepth>>;
  displayOrderRef: MutableRefObject<TaskWithDepth[]>;
  prevFilteredRef: MutableRefObject<TaskWithDepth[]>;
};

function scheduleRemovedTaskExits(
  prevOrder: TaskWithDepth[],
  tasksIds: Set<string>,
  exitingRef: MutableRefObject<Map<string, ExitingRow>>,
  seenIdsRef: MutableRefObject<Set<string>>,
  setExitingTick: (updater: (value: number) => number) => void,
): void {
  for (const pr of prevOrder) {
    if (tasksIds.has(pr.id)) continue;
    if (exitingRef.current.has(pr.id)) continue;
    const timeoutId = window.setTimeout(() => {
      exitingRef.current.delete(pr.id);
      seenIdsRef.current.delete(pr.id);
      setExitingTick((x) => x + 1);
    }, ROW_EXIT_MS);
    exitingRef.current.set(pr.id, { task: pr, timeoutId });
  }
}

function scheduleFilterRemovedRowExits(
  prevOrder: TaskWithDepth[],
  nextIds: Set<string>,
  tasksIds: Set<string>,
  filterExitingRef: MutableRefObject<Map<string, TaskWithDepth>>,
  displayOrderRef: MutableRefObject<TaskWithDepth[]>,
  seenIdsRef: MutableRefObject<Set<string>>,
  setExitingTick: (updater: (value: number) => number) => void,
): boolean {
  let scheduledFilterExit = false;
  for (const t of prevOrder) {
    if (nextIds.has(t.id)) continue;
    if (!tasksIds.has(t.id)) continue;
    if (filterExitingRef.current.has(t.id)) continue;
    filterExitingRef.current.set(t.id, t);
    window.setTimeout(() => {
      filterExitingRef.current.delete(t.id);
      displayOrderRef.current = displayOrderRef.current.filter(
        (row) => row.id !== t.id,
      );
      seenIdsRef.current.delete(t.id);
      setExitingTick((x) => x + 1);
    }, ROW_EXIT_MS);
    scheduledFilterExit = true;
  }
  return scheduledFilterExit;
}

function syncTaskListEnteringIds(
  filteredTasks: TaskWithDepth[],
  filteredIds: Set<string>,
  refs: TaskListRowAnimationRefs,
  setEnteringIds: (updater: (prev: Set<string>) => Set<string>) => void,
): void {
  const newlyEntering = new Set<string>();
  for (const t of filteredTasks) {
    if (!refs.seenIdsRef.current.has(t.id)) {
      newlyEntering.add(t.id);
      refs.seenIdsRef.current.add(t.id);
    }
    const pendingExit = refs.exitingRef.current.get(t.id);
    if (pendingExit) {
      clearTimeout(pendingExit.timeoutId);
      refs.exitingRef.current.delete(t.id);
    }
    refs.filterExitingRef.current.delete(t.id);
  }

  const clientExitIds = new Set(refs.filterExitingRef.current.keys());
  for (const id of Array.from(refs.seenIdsRef.current)) {
    if (filteredIds.has(id)) continue;
    if (refs.exitingRef.current.has(id)) continue;
    if (clientExitIds.has(id)) continue;
    refs.seenIdsRef.current.delete(id);
  }

  setEnteringIds((prevEntering) => {
    if (prevEntering.size === 0 && newlyEntering.size === 0) return prevEntering;
    return newlyEntering;
  });
}

export function buildTaskListRowsToRender(
  renderOrder: TaskWithDepth[],
  filteredTasks: TaskWithDepth[],
  filteredMap: Map<string, TaskWithDepth>,
  filterExitIds: Set<string>,
  filterExitingRef: MutableRefObject<Map<string, TaskWithDepth>>,
  exitingRef: MutableRefObject<Map<string, ExitingRow>>,
  enteringIds: Set<string>,
): TaskListRowRenderState[] {
  const rowsToRender: TaskListRowRenderState[] = [];
  const processed = new Set<string>();
  for (const t of renderOrder) {
    const visible = filteredMap.get(t.id);
    if (visible) {
      rowsToRender.push({
        task: visible,
        isEntering: enteringIds.has(t.id),
        isExiting: false,
        isFilterExit: false,
      });
      processed.add(t.id);
      continue;
    }
    if (filterExitIds.has(t.id)) {
      const exitingTask = filterExitingRef.current.get(t.id) ?? t;
      rowsToRender.push({
        task: exitingTask,
        isEntering: false,
        isExiting: true,
        isFilterExit: true,
      });
      processed.add(t.id);
    }
  }
  for (const t of filteredTasks) {
    if (processed.has(t.id)) continue;
    rowsToRender.push({
      task: t,
      isEntering: enteringIds.has(t.id),
      isExiting: false,
      isFilterExit: false,
    });
    processed.add(t.id);
  }
  for (const { task } of exitingRef.current.values()) {
    if (processed.has(task.id)) continue;
    rowsToRender.push({
      task,
      isEntering: false,
      isExiting: true,
      isFilterExit: false,
    });
  }
  return rowsToRender;
}

export function useTaskListRowAnimations(
  filteredTasks: TaskWithDepth[],
  tasks: TaskWithDepth[],
): TaskListRowRenderState[] {
  const seenIdsRef = useRef<Set<string>>(new Set());
  const [enteringIds, setEnteringIds] = useState<Set<string>>(new Set());
  const exitingRef = useRef<Map<string, ExitingRow>>(new Map());
  const filterExitingRef = useRef<Map<string, TaskWithDepth>>(new Map());
  const displayOrderRef = useRef<TaskWithDepth[]>([]);
  const prevFilteredRef = useRef<TaskWithDepth[]>([]);
  const [exitingTick, setExitingTick] = useState(0);

  const filteredIds = useMemo(
    () => new Set(filteredTasks.map((t) => t.id)),
    [filteredTasks],
  );
  const tasksIds = useMemo(() => new Set(tasks.map((t) => t.id)), [tasks]);
  const animationRefs: TaskListRowAnimationRefs = {
    seenIdsRef,
    exitingRef,
    filterExitingRef,
    displayOrderRef,
    prevFilteredRef,
  };

  useLayoutEffect(() => {
    const prevOrder =
      displayOrderRef.current.length > 0
        ? displayOrderRef.current
        : prevFilteredRef.current;
    const nextIds = new Set(filteredTasks.map((t) => t.id));
    const scheduledFilterExit = scheduleFilterRemovedRowExits(
      prevOrder,
      nextIds,
      tasksIds,
      filterExitingRef,
      displayOrderRef,
      seenIdsRef,
      setExitingTick,
    );

    for (const t of filteredTasks) {
      filterExitingRef.current.delete(t.id);
    }

    scheduleRemovedTaskExits(
      prevOrder,
      tasksIds,
      exitingRef,
      seenIdsRef,
      setExitingTick,
    );

    displayOrderRef.current = computeTaskListDisplayOrder(
      prevOrder,
      filteredTasks,
      new Set(filterExitingRef.current.keys()),
      filterExitingRef.current,
    );
    prevFilteredRef.current = filteredTasks;
    if (scheduledFilterExit) {
      setExitingTick((x) => x + 1);
    }
  }, [filteredTasks, tasksIds, filteredIds]);

  useEffect(() => {
    syncTaskListEnteringIds(filteredTasks, filteredIds, animationRefs, setEnteringIds);
  }, [filteredTasks, filteredIds, tasksIds]);

  useEffect(() => {
    const exiting = exitingRef.current;
    return () => {
      for (const { timeoutId } of exiting.values()) {
        clearTimeout(timeoutId);
      }
      exiting.clear();
    };
  }, []);

  const filteredMap = useMemo(
    () => new Map(filteredTasks.map((t) => [t.id, t])),
    [filteredTasks],
  );
  const filterExitIds = useMemo(
    () => new Set(filterExitingRef.current.keys()),
    // eslint-disable-next-line react-hooks/exhaustive-deps -- ref map is the source of truth
    [exitingTick, filteredTasks],
  );
  const renderOrder =
    displayOrderRef.current.length > 0
      ? displayOrderRef.current
      : filteredTasks;
  const rowsToRender = buildTaskListRowsToRender(
    renderOrder,
    filteredTasks,
    filteredMap,
    filterExitIds,
    filterExitingRef,
    exitingRef,
    enteringIds,
  );
  void exitingTick;

  return rowsToRender;
}
