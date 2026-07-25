export function SettingsActions({
  isDirty,
  maxInvalid,
  pickupInvalid,
  patchPending,
  onDiscard,
}: {
  isDirty: boolean;
  maxInvalid: boolean;
  pickupInvalid: boolean;
  patchPending: boolean;
  onDiscard: () => void;
}) {
  return (
    <div className="settings-actions" data-dirty={isDirty ? "true" : "false"}>
      <div className="settings-actions-status" aria-hidden="true">
        {maxInvalid || pickupInvalid ? (
          <span className="settings-actions-hint settings-actions-hint--warn">
            Resolve the errors above to save.
          </span>
        ) : isDirty ? (
          <span className="settings-actions-hint settings-actions-hint--dirty">
            <span className="settings-actions-dot" />
            Unsaved changes
          </span>
        ) : (
          <span className="settings-actions-hint settings-actions-hint--clean">
            All changes saved
          </span>
        )}
      </div>
      <div className="settings-actions-buttons">
        {isDirty ? (
          <button
            type="button"
            className="settings-btn settings-btn--ghost"
            onClick={onDiscard}
            disabled={patchPending}
          >
            Discard
          </button>
        ) : null}
        <button
          type="submit"
          className="settings-btn settings-btn--primary"
          disabled={
            !isDirty ||
            patchPending ||
            maxInvalid ||
            pickupInvalid
          }
        >
          {patchPending ? "Saving…" : "Save changes"}
        </button>
      </div>
    </div>
  );
}
