import { Modal } from "@/shared/Modal";

/** Shown while the create-modal chunk loads after open. */
export function CreateModalChunkFallback({ onClose }: { onClose: () => void }) {
  return (
    <Modal
      labelledBy="create-modal-chunk-title"
      onClose={onClose}
      size="wide"
      busy
      busyLabel="Loading form…"
      dismissibleWhileBusy
    >
      <h2 id="create-modal-chunk-title" className="visually-hidden">
        Loading form
      </h2>
    </Modal>
  );
}
