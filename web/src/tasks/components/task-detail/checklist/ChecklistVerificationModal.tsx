import { Modal } from "@/shared/Modal";
import { CHECKLIST_EVIDENCE_DISPLAY_CAP } from "@/types/task";

type Props = {
  /** The criterion text — shown as the modal heading. */
  criterionText: string;
  /** Raw evidence payload captured at verification time. */
  evidence?: string;
  onClose: () => void;
};

/**
 * Popup for evidence on a satisfied checklist criterion. Inline
 * `<details>` disclosures made each criterion row visually heavy;
 * promoting that detail into an on-demand sheet keeps the row
 * scannable while still making the audit trail one click away.
 */
export function ChecklistVerificationModal({
  criterionText,
  evidence,
  onClose,
}: Props) {
  const hasEvidence = typeof evidence === "string" && evidence.length > 0;

  return (
    <Modal
      onClose={onClose}
      labelledBy="checklist-verification-title"
      size="wide"
    >
      <section className="panel modal-sheet checklist-verification-modal">
        <header className="checklist-verification-modal-head">
          <p className="checklist-verification-modal-eyebrow">
            Verification details
          </p>
          <h2
            id="checklist-verification-title"
            className="checklist-verification-modal-title"
          >
            {criterionText}
          </h2>
        </header>

        <div className="checklist-verification-modal-body">
          {hasEvidence ? (
            <section
              className="checklist-verification-modal-section"
              aria-labelledby="checklist-verification-evidence-heading"
            >
              <h3
                id="checklist-verification-evidence-heading"
                className="checklist-verification-modal-section-heading"
              >
                Evidence
              </h3>
              <pre className="checklist-verification-modal-pre">
                {evidence!.slice(0, CHECKLIST_EVIDENCE_DISPLAY_CAP)}
              </pre>
              {evidence!.length > CHECKLIST_EVIDENCE_DISPLAY_CAP ? (
                <p className="checklist-verification-modal-truncated">
                  Truncated to {CHECKLIST_EVIDENCE_DISPLAY_CAP.toLocaleString()}{" "}
                  characters for display. The full payload is preserved on the
                  server.
                </p>
              ) : null}
            </section>
          ) : (
            <p className="checklist-verification-modal-empty muted">
              No additional verification detail was captured for this
              criterion.
            </p>
          )}
        </div>

        <div className="row stack-row-actions checklist-verification-modal-footer">
          <button
            type="button"
            className="secondary"
            onClick={onClose}
            autoFocus
          >
            Close
          </button>
        </div>
      </section>
    </Modal>
  );
}
