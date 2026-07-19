import { FieldLabel } from "@/shared/FieldLabel";
import { Modal } from "@/shared/Modal";
import {
  useCallback,
  useEffect,
  useId,
  useRef,
  useState,
  type DragEvent,
  type FormEvent,
  type KeyboardEvent,
} from "react";
import {
  MAX_PROJECT_CONTEXT_BODY_BYTES,
  MEMORY_IMPORT_ACCEPT,
} from "./projectContextLimits";
import {
  formatMemoryImportBytes,
  MemoryImportError,
  previewMemoryImportText,
  readMemoryImportFile,
  sanitizeMemoryAlias,
  validateMemoryAlias,
  type MemoryImportFileResult,
} from "./readMemoryImportFile";

type Props = {
  open: boolean;
  onClose: () => void;
  isPending: boolean;
  onImport: (input: {
    title: string;
    body: string;
  }) => void;
};

export function ProjectContextImportMemoryModal({
  open,
  onClose,
  isPending,
  onImport,
}: Props) {
  const fileInputRef = useRef<HTMLInputElement>(null);
  const dropzoneId = useId();
  const aliasId = useId();
  const errorId = useId();
  const helperId = useId();

  const [imported, setImported] = useState<MemoryImportFileResult | null>(null);
  const [alias, setAlias] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [reading, setReading] = useState(false);
  const [dragActive, setDragActive] = useState(false);

  useEffect(() => {
    if (!open) {
      setImported(null);
      setAlias("");
      setError(null);
      setReading(false);
      setDragActive(false);
      if (fileInputRef.current) fileInputRef.current.value = "";
    }
  }, [open]);

  const applyFile = useCallback(async (file: File | null | undefined) => {
    if (!file) return;
    setReading(true);
    setError(null);
    try {
      const result = await readMemoryImportFile(file);
      setImported(result);
      setAlias(result.defaultAlias);
    } catch (err) {
      setImported(null);
      setAlias("");
      setError(
        err instanceof MemoryImportError
          ? err.message
          : "Could not read that file. Try a smaller .txt or .md file.",
      );
    } finally {
      setReading(false);
    }
  }, []);

  if (!open) return null;

  const busy = isPending || reading;
  const aliasError = alias ? validateMemoryAlias(alias) : null;
  const canSubmit =
    Boolean(imported) &&
    !aliasError &&
    sanitizeMemoryAlias(alias).length > 0 &&
    !busy;

  function openFilePicker() {
    if (busy) return;
    fileInputRef.current?.click();
  }

  function onDropzoneKeyDown(event: KeyboardEvent<HTMLDivElement>) {
    if (event.key === "Enter" || event.key === " ") {
      event.preventDefault();
      openFilePicker();
    }
  }

  function onDragOver(event: DragEvent<HTMLDivElement>) {
    event.preventDefault();
    event.stopPropagation();
    if (!busy) setDragActive(true);
  }

  function onDragLeave(event: DragEvent<HTMLDivElement>) {
    event.preventDefault();
    event.stopPropagation();
    setDragActive(false);
  }

  function onDrop(event: DragEvent<HTMLDivElement>) {
    event.preventDefault();
    event.stopPropagation();
    setDragActive(false);
    if (busy) return;
    const file = event.dataTransfer.files?.[0];
    void applyFile(file);
  }

  function clearFile() {
    if (busy) return;
    setImported(null);
    setAlias("");
    setError(null);
    if (fileInputRef.current) fileInputRef.current.value = "";
  }

  function onSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!imported || !canSubmit) return;
    const title = sanitizeMemoryAlias(alias);
    const titleErr = validateMemoryAlias(title);
    if (titleErr) {
      setError(titleErr);
      return;
    }
    onImport({ title, body: imported.text });
  }

  return (
    <Modal
      onClose={onClose}
      labelledBy="project-context-import-title"
      describedBy="project-context-import-desc"
      size="wide"
      busy={busy}
      busyLabel="Importing memory file..."
    >
      <form
        className="panel modal-sheet modal-sheet--edit project-context-form project-context-import-modal"
        onSubmit={onSubmit}
      >
        <div className="project-context-form__heading">
          <div>
            <h2 id="project-context-import-title">Import memory file</h2>
            <p id="project-context-import-desc" className="muted">
              Import a local .txt or .md file as a project memory node. The
              alias is what you and the agent see when selecting it.
            </p>
          </div>
        </div>

        <div className="field grow">
          <FieldLabel htmlFor={`${dropzoneId}-input`} requirement="required">
            File
          </FieldLabel>
          <input
            ref={fileInputRef}
            id={`${dropzoneId}-input`}
            type="file"
            accept={MEMORY_IMPORT_ACCEPT}
            className="pc__file-input"
            disabled={busy}
            onChange={(event) => {
              void applyFile(event.target.files?.[0]);
            }}
          />
          <div
            className={[
              "pc__dropzone",
              dragActive ? "pc__dropzone--active" : "",
              imported ? "pc__dropzone--selected" : "",
            ]
              .filter(Boolean)
              .join(" ")}
            role="button"
            tabIndex={busy ? -1 : 0}
            aria-label="Choose a .txt or .md memory file"
            aria-describedby={`${helperId}${error ? ` ${errorId}` : ""}`}
            aria-disabled={busy}
            onClick={openFilePicker}
            onKeyDown={onDropzoneKeyDown}
            onDragOver={onDragOver}
            onDragLeave={onDragLeave}
            onDrop={onDrop}
          >
            {imported ? (
              <div className="pc__dropzone-selected">
                <span className="pc__dropzone-filename">{imported.fileName}</span>
                <span className="muted">
                  {formatMemoryImportBytes(imported.byteSize)}
                </span>
                <button
                  type="button"
                  className="pc__btn-ghost"
                  disabled={busy}
                  onClick={(event) => {
                    event.stopPropagation();
                    clearFile();
                  }}
                >
                  Clear
                </button>
              </div>
            ) : (
              <div className="pc__dropzone-prompt">
                <span>Drop a .txt or .md file here, or browse</span>
                <span className="pc__dropzone-browse">Browse</span>
              </div>
            )}
          </div>
          <p id={helperId} className="pc__field-hint">
            Up to {formatMemoryImportBytes(MAX_PROJECT_CONTEXT_BODY_BYTES)} ·
            plain text or Markdown
          </p>
        </div>

        <div className="field grow">
          <FieldLabel htmlFor={aliasId} requirement="required">
            Alias
          </FieldLabel>
          <input
            id={aliasId}
            name="alias"
            value={alias}
            required
            aria-required="true"
            aria-invalid={Boolean(aliasError)}
            disabled={busy || !imported}
            maxLength={200}
            onChange={(event) => setAlias(event.target.value)}
          />
          {aliasError ? (
            <p className="pd__inline-error" role="alert">
              {aliasError}
            </p>
          ) : null}
        </div>

        {imported ? (
          <div className="field grow">
            <FieldLabel htmlFor={`${dropzoneId}-preview`}>Preview</FieldLabel>
            <pre
              id={`${dropzoneId}-preview`}
              className="pc__import-preview"
              tabIndex={0}
            >
              {previewMemoryImportText(imported.text)}
            </pre>
          </div>
        ) : null}

        {error ? (
          <div id={errorId} className="pd__inline-error" role="alert">
            {error}
          </div>
        ) : null}

        <div className="row stack-row-actions">
          <button type="submit" disabled={!canSubmit}>
            {isPending ? "Importing..." : reading ? "Reading..." : "Import"}
          </button>
          <button
            type="button"
            className="secondary"
            disabled={isPending}
            onClick={onClose}
          >
            Cancel
          </button>
        </div>
      </form>
    </Modal>
  );
}
