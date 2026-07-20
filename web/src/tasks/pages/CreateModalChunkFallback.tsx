import { Modal } from "@/shared/Modal";

/** Shown while the create-modal / TipTap chunk loads after open. */
export function CreateModalChunkFallback({ onClose }: { onClose: () => void }) {
  return (
    <Modal
      labelledBy="create-modal-chunk-title"
      onClose={onClose}
      size="wide"
      busy
      busyLabel="Loading editor…"
      dismissibleWhileBusy
    >
      <h2 id="create-modal-chunk-title" className="visually-hidden">
        Loading editor
      </h2>
    </Modal>
  );
}
