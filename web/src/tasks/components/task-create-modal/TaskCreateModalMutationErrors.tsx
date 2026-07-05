import { MutationErrorBanner } from "../../../shared/MutationErrorBanner";

type Props = {
  isTaskEdit: boolean;
  createFormError?: string | null;
  createError?: Error | null;
  formError?: string | null;
  patchError?: string | null;
};

export function TaskCreateModalMutationErrors({
  isTaskEdit,
  createFormError,
  createError,
  formError,
  patchError,
}: Props) {
  if (!isTaskEdit) {
    return (
      <>
        <MutationErrorBanner
          error={createFormError}
          className="task-create-modal-err task-create-modal-err--create"
        />

        <MutationErrorBanner
          error={createError}
          fallback="Could not create task."
          className="task-create-modal-err task-create-modal-err--create"
        />
      </>
    );
  }

  return (
    <>
      <MutationErrorBanner
        error={formError}
        className="task-create-modal-err task-create-modal-err--edit"
      />
      <MutationErrorBanner
        error={patchError}
        className="task-edit-form-err task-create-modal-err task-create-modal-err--edit"
      />
    </>
  );
}
