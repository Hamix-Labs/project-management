import {
  TaskBulkDeleteConfirmModal,
  TaskBulkRescheduleModal,
  TaskListBulkActionBar,
} from "../bulk";
import type { useTaskListSectionBulkActions } from "./useTaskListSectionBulkActions";

type Bulk = ReturnType<typeof useTaskListSectionBulkActions>;

type Props = {
  scheduleUiEnabled: boolean;
  appTimezone: string;
  bulk: Bulk;
};

export function TaskListBulkLayer({
  scheduleUiEnabled,
  appTimezone,
  bulk,
}: Props) {
  return (
    <>
      <TaskListBulkActionBar
        selectedCount={bulk.selection.selectedVisibleIds.length}
        scheduledCount={bulk.selectedScheduledIds.length}
        rescheduleDisabled={bulk.selectedIncludesDone}
        showScheduleActions={scheduleUiEnabled}
        busy={bulk.bulkSchedule.isPending || bulk.bulkDelete.isPending}
        onReschedule={bulk.openRescheduleModal}
        onClearSchedule={bulk.handleClearSchedule}
        onDelete={bulk.openBulkDeleteModal}
        onCancel={bulk.handleCancelSelection}
      />
      {bulk.bulkDeleteModalOpen && bulk.selectedRowsForBulkDelete.length > 0 ? (
        <TaskBulkDeleteConfirmModal
          tasks={bulk.selectedRowsForBulkDelete}
          busy={bulk.bulkDelete.isPending}
          error={bulk.bulkDeleteError}
          onCancel={bulk.closeBulkDelete}
          onConfirm={bulk.handleBulkDeleteConfirm}
        />
      ) : null}
      {bulk.rescheduleModalOpen ? (
        <TaskBulkRescheduleModal
          selectedCount={bulk.selection.selectedVisibleIds.length}
          appTimezone={appTimezone}
          busy={bulk.bulkSchedule.isPending}
          error={bulk.bulkErrorBanner}
          onClose={bulk.closeReschedule}
          onSubmit={bulk.handleRescheduleSubmit}
        />
      ) : null}
    </>
  );
}
