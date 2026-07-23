import { Modal } from "@/shared/Modal";
import { MentionRangePanel } from "./MentionRangePanel";
import type { PendingFileInsert } from "./richPromptInsertHelpers";

type Props = {
  id: string;
  pendingInsert: PendingFileInsert;
  disabled?: boolean;
  /** Scopes file preview + range validation to the task worktree. */
  worktreeId?: string;
  rangeWarning: string | null;
  onClose: () => void;
  onInsertWithRange: (startLine: number, endLine: number) => Promise<void>;
  onInsertPathOnly: () => void;
};

export function RichPromptFileReferenceModal({
  id,
  pendingInsert,
  disabled,
  worktreeId,
  rangeWarning,
  onClose,
  onInsertWithRange,
  onInsertPathOnly,
}: Props) {
  const rangeModalTitleId = `${id}-mention-range-modal-title`;
  const rangeModalDescId = `${id}-mention-range-modal-desc`;

  return (
    <Modal
      onClose={onClose}
      labelledBy={rangeModalTitleId}
      describedBy={rangeModalDescId}
      size="wide"
    >
      <section className="panel modal-sheet mention-range-modal">
        <h2 id={rangeModalTitleId}>Insert file reference</h2>
        <p id={rangeModalDescId} className="mention-range-modal-desc muted">
          Review the file, optionally choose a line range, then insert it into
          your prompt.
        </p>
        <MentionRangePanel
          id={id}
          path={pendingInsert.path}
          worktreeId={worktreeId}
          disabled={disabled}
          rangeWarning={rangeWarning}
          onInsertWithRange={onInsertWithRange}
          onInsertPathOnly={onInsertPathOnly}
          onCancel={onClose}
        />
      </section>
    </Modal>
  );
}
