import { ConfirmDialog } from "@/components/feedback/ConfirmDialog";
import { deleteEntityConfirmProps } from "@/components/feedback/confirmDialogPresets";

type Props = {
  taskTitle: string;
  saving: boolean;
  deletePending: boolean;
  error?: string | null;
  onCancel: () => void;
  onConfirm: () => void;
};

export function DeleteConfirmDialog({
  taskTitle,
  saving,
  deletePending,
  error = null,
  onCancel,
  onConfirm,
}: Props) {
  const preset = deleteEntityConfirmProps({
    noun: "task",
    name: taskTitle,
    titleId: "delete-dialog-title",
    descriptionId: "delete-dialog-description",
  });
  return (
    <ConfirmDialog
      {...preset}
      busy={deletePending}
      cancelDisabled={saving}
      confirmDisabled={saving}
      error={error}
      onCancel={onCancel}
      onConfirm={onConfirm}
    />
  );
}
