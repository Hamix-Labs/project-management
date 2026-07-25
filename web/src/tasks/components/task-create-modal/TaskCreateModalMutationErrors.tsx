import { MutationErrorBanner } from "../../../shared/MutationErrorBanner";
import { taskMutationErrorMessage } from "../../create/taskTagValidation";

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
          error={
            createError
              ? taskMutationErrorMessage(createError, "Could not create task.")
              : null
          }
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
        error={patchError ? taskMutationErrorMessage(patchError) : null}
        className="task-edit-form-err task-create-modal-err task-create-modal-err--edit"
      />
    </>
  );
}
