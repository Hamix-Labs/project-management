type Props = {
  disabled: boolean;
  saveDisabled?: boolean;
  form?: string;
  onClose: () => void;
};

export function TaskCreateModalEditFooterActions({
  disabled,
  saveDisabled = false,
  form,
  onClose,
}: Props) {
  return (
    <div className="task-create-modal-actions">
      <div className="task-create-modal-actions__start">
        <button
          type="button"
          className="secondary task-create-cancel-btn"
          disabled={disabled}
          onClick={onClose}
        >
          Cancel
        </button>
      </div>
      <div className="task-create-modal-actions__end">
        <button
          type="submit"
          className="task-create-submit"
          form={form}
          disabled={disabled || saveDisabled}
        >
          Save
        </button>
      </div>
    </div>
  );
}
