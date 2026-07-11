import { useCallback, useMemo, useState } from "react";
import type { TaskWithDepth } from "../../../task-tree";
import {
  useBulkDeleteMutation,
  useBulkScheduleMutation,
  useTaskListSelection,
} from "../bulk";
import {
  formatBulkDeleteFailure,
  formatBulkFailure,
} from "./taskListSectionBulkUtils";

type UseTaskListSectionBulkActionsArgs = {
  filteredTasks: TaskWithDepth[];
  visibleIds: string[];
  scheduleUiEnabled: boolean;
};

export function useTaskListSectionBulkActions({
  filteredTasks,
  visibleIds,
  scheduleUiEnabled,
}: UseTaskListSectionBulkActionsArgs) {
  const selection = useTaskListSelection(visibleIds);
  const bulkSchedule = useBulkScheduleMutation();
  const bulkDelete = useBulkDeleteMutation();
  const [rescheduleModalOpen, setRescheduleModalOpen] = useState(false);
  const [bulkDeleteModalOpen, setBulkDeleteModalOpen] = useState(false);
  const [bulkErrorBanner, setBulkErrorBanner] = useState<string | null>(null);
  const [bulkDeleteError, setBulkDeleteError] = useState<string | null>(null);

  const selectedScheduledIds = useMemo(() => {
    const visibleSelected = new Set(selection.selectedVisibleIds);
    return filteredTasks
      .filter(
        (t) =>
          visibleSelected.has(t.id) && Boolean(t.pickup_not_before),
      )
      .map((t) => t.id);
  }, [filteredTasks, selection.selectedVisibleIds]);

  const selectedIncludesDone = useMemo(() => {
    const visibleSelected = new Set(selection.selectedVisibleIds);
    return filteredTasks.some(
      (t) => visibleSelected.has(t.id) && t.status === "done",
    );
  }, [filteredTasks, selection.selectedVisibleIds]);

  const selectedRowsForBulkDelete = useMemo(() => {
    const visibleSelected = new Set(selection.selectedVisibleIds);
    return filteredTasks
      .filter((t) => visibleSelected.has(t.id))
      .map((t) => ({
        id: t.id,
        title: t.title,
      }));
  }, [filteredTasks, selection.selectedVisibleIds]);

  const closeReschedule = useCallback(() => {
    setRescheduleModalOpen(false);
    bulkSchedule.reset();
  }, [bulkSchedule]);

  const closeBulkDelete = useCallback(() => {
    setBulkDeleteModalOpen(false);
    bulkDelete.reset();
    setBulkDeleteError(null);
  }, [bulkDelete]);

  const handleRescheduleSubmit = useCallback(
    async (next: string | null) => {
      const ids = selection.selectedVisibleIds;
      if (ids.length === 0) {
        setRescheduleModalOpen(false);
        return;
      }
      if (
        ids.some(
          (id) => filteredTasks.find((t) => t.id === id)?.status === "done",
        )
      ) {
        setRescheduleModalOpen(false);
        return;
      }
      const result = await bulkSchedule.run(ids, next);
      if (result.failed.length === 0) {
        setRescheduleModalOpen(false);
        selection.clearSelection();
        setBulkErrorBanner(null);
      } else {
        setBulkErrorBanner(formatBulkFailure(result.failed.length, result.attempted));
      }
    },
    [bulkSchedule, filteredTasks, selection],
  );

  const handleClearSchedule = useCallback(async () => {
    const ids = selectedScheduledIds;
    if (ids.length === 0) return;
    if (ids.length > 5) {
      const ok = window.confirm(
        `Clear scheduled pickup on ${ids.length} tasks? They will be eligible for the agent immediately.`,
      );
      if (!ok) return;
    }
    const result = await bulkSchedule.run(ids, null);
    if (result.failed.length === 0) {
      selection.clearSelection();
      setBulkErrorBanner(null);
    } else {
      setBulkErrorBanner(formatBulkFailure(result.failed.length, result.attempted));
    }
  }, [bulkSchedule, selectedScheduledIds, selection]);

  const handleBulkDeleteConfirm = useCallback(async () => {
    const ids = selection.selectedVisibleIds;
    if (ids.length === 0) {
      closeBulkDelete();
      return;
    }
    const result = await bulkDelete.run(ids);
    if (result.failed.length === 0) {
      setBulkDeleteModalOpen(false);
      bulkDelete.reset();
      selection.clearSelection();
      setBulkDeleteError(null);
    } else {
      setBulkDeleteError(
        formatBulkDeleteFailure(result.failed.length, result.attempted),
      );
    }
  }, [bulkDelete, closeBulkDelete, selection]);

  const handleCancelSelection = useCallback(() => {
    selection.clearSelection();
    setBulkErrorBanner(null);
    setBulkDeleteModalOpen(false);
    bulkDelete.reset();
    setBulkDeleteError(null);
    setRescheduleModalOpen(false);
    bulkSchedule.reset();
  }, [bulkDelete, bulkSchedule, selection]);

  const openRescheduleModal = useCallback(() => {
    setBulkDeleteModalOpen(false);
    bulkDelete.reset();
    if (selectedIncludesDone) return;
    setBulkErrorBanner(null);
    setRescheduleModalOpen(true);
  }, [bulkDelete, selectedIncludesDone]);

  const openBulkDeleteModal = useCallback(() => {
    setRescheduleModalOpen(false);
    bulkSchedule.reset();
    setBulkErrorBanner(null);
    setBulkDeleteError(null);
    setBulkDeleteModalOpen(true);
  }, [bulkSchedule]);

  return {
    selection,
    bulkSchedule,
    bulkDelete,
    rescheduleModalOpen: scheduleUiEnabled && rescheduleModalOpen,
    bulkDeleteModalOpen,
    bulkErrorBanner,
    bulkDeleteError,
    selectedScheduledIds,
    selectedIncludesDone,
    selectedRowsForBulkDelete,
    closeReschedule,
    closeBulkDelete,
    handleRescheduleSubmit,
    handleClearSchedule,
    handleBulkDeleteConfirm,
    handleCancelSelection,
    openRescheduleModal,
    openBulkDeleteModal,
  };
}
