import { useQueryClient } from "@tanstack/react-query";
import { useCallback, useEffect, useRef, useState } from "react";
import { setSseLiveForQueries } from "@/tasks/sync/connectionPolicy";
import { connectTaskEventSource } from "./sseConnection";
import {
  createTaskSyncCoordinator,
  type TaskSyncCoordinator,
} from "../sync/taskSyncCoordinator";

/**
 * App-wide task SSE transport. Debounce timers and pending maps live in
 * `taskSyncCoordinator` (ADR-0022). This hook connects EventSource and
 * publishes the live flag for refetch-on-focus policy.
 *
 * Unmount disposes the coordinator without flushing (drop pending).
 */
export function useTaskEventStream(): boolean {
  const queryClient = useQueryClient();
  const coordinatorRef = useRef<TaskSyncCoordinator | null>(null);
  const streamEffectActiveRef = useRef(false);
  const [sseLive, setSseLive] = useState(false);

  const onMessage = useCallback((data: string) => {
    coordinatorRef.current?.handleRawFrame(data);
  }, []);

  useEffect(() => {
    streamEffectActiveRef.current = true;
    const coordinator = createTaskSyncCoordinator(queryClient);
    coordinatorRef.current = coordinator;
    const disconnect = connectTaskEventSource({
      isActive: () => streamEffectActiveRef.current,
      onMessage,
      onLiveChange: setSseLive,
    });
    return () => {
      streamEffectActiveRef.current = false;
      disconnect();
      coordinator.dispose();
      if (coordinatorRef.current === coordinator) {
        coordinatorRef.current = null;
      }
    };
  }, [queryClient, onMessage]);

  useEffect(() => {
    setSseLiveForQueries(sseLive);
    return () => {
      setSseLiveForQueries(false);
    };
  }, [sseLive]);

  return sseLive;
}
