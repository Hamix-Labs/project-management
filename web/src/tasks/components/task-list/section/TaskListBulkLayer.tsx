import {
  TaskBulkCloseConfirmModal,
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
        busy={bulk.bulkSchedule.isPending || bulk.bulkClose.isPending}
        onReschedule={bulk.openRescheduleModal}
        onClearSchedule={bulk.handleClearSchedule}
        onClose={bulk.openBulkCloseModal}
        onCancel={bulk.handleCancelSelection}
      />
      {bulk.bulkCloseModalOpen && bulk.selectedRowsForBulkClose.length > 0 ? (
        <TaskBulkCloseConfirmModal
          tasks={bulk.selectedRowsForBulkClose}
          busy={bulk.bulkClose.isPending}
          error={bulk.bulkCloseError}
          onCancel={bulk.closeBulkClose}
          onConfirm={bulk.handleBulkCloseConfirm}
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
