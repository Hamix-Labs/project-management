import { useCallback, useMemo, useState } from "react";
import type { TaskWithDepth } from "../../../task-tree";
import {
  useBulkCloseMutation,
  useBulkScheduleMutation,
  useTaskListSelection,
} from "../bulk";
import {
  formatBulkCloseFailure,
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
  const bulkClose = useBulkCloseMutation();
  const [rescheduleModalOpen, setRescheduleModalOpen] = useState(false);
  const [bulkCloseModalOpen, setBulkCloseModalOpen] = useState(false);
  const [bulkErrorBanner, setBulkErrorBanner] = useState<string | null>(null);
  const [bulkCloseError, setBulkCloseError] = useState<string | null>(null);

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

  const selectedRowsForBulkClose = useMemo(() => {
    const visibleSelected = new Set(selection.selectedVisibleIds);
    return filteredTasks
      .filter((t) => visibleSelected.has(t.id))
      .map((t) => ({
        id: t.id,
        title: t.title,
        number: t.number,
      }));
  }, [filteredTasks, selection.selectedVisibleIds]);

  const closeReschedule = useCallback(() => {
    setRescheduleModalOpen(false);
    bulkSchedule.reset();
  }, [bulkSchedule]);

  const closeBulkClose = useCallback(() => {
    setBulkCloseModalOpen(false);
    bulkClose.reset();
    setBulkCloseError(null);
  }, [bulkClose]);

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

  const handleBulkCloseConfirm = useCallback(async () => {
    const ids = selection.selectedVisibleIds;
    if (ids.length === 0) {
      closeBulkClose();
      return;
    }
    const result = await bulkClose.run(ids);
    if (result.failed.length === 0) {
      setBulkCloseModalOpen(false);
      bulkClose.reset();
      selection.clearSelection();
      setBulkCloseError(null);
    } else {
      setBulkCloseError(
        formatBulkCloseFailure(result.failed.length, result.attempted),
      );
    }
  }, [bulkClose, closeBulkClose, selection]);

  const handleCancelSelection = useCallback(() => {
    selection.clearSelection();
    setBulkErrorBanner(null);
    setBulkCloseModalOpen(false);
    bulkClose.reset();
    setBulkCloseError(null);
    setRescheduleModalOpen(false);
    bulkSchedule.reset();
  }, [bulkClose, bulkSchedule, selection]);

  const openRescheduleModal = useCallback(() => {
    setBulkCloseModalOpen(false);
    bulkClose.reset();
    if (selectedIncludesDone) return;
    setBulkErrorBanner(null);
    setRescheduleModalOpen(true);
  }, [bulkClose, selectedIncludesDone]);

  const openBulkCloseModal = useCallback(() => {
    setRescheduleModalOpen(false);
    bulkSchedule.reset();
    setBulkErrorBanner(null);
    setBulkCloseError(null);
    setBulkCloseModalOpen(true);
  }, [bulkSchedule]);

  return {
    selection,
    bulkSchedule,
    bulkClose,
    rescheduleModalOpen: scheduleUiEnabled && rescheduleModalOpen,
    bulkCloseModalOpen,
    bulkErrorBanner,
    bulkCloseError,
    selectedScheduledIds,
    selectedIncludesDone,
    selectedRowsForBulkClose,
    closeReschedule,
    closeBulkClose,
    handleRescheduleSubmit,
    handleClearSchedule,
    handleBulkCloseConfirm,
    handleCancelSelection,
    openRescheduleModal,
    openBulkCloseModal,
  };
}
