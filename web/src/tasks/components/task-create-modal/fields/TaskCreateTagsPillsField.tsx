import {
  useId,
  useRef,
  useState,
  type KeyboardEvent as ReactKeyboardEvent,
  type MouseEvent as ReactMouseEvent,
} from "react";
import { FieldLabel } from "@/shared/FieldLabel";
import {
  formatTagsCsv,
  isValidTaskTag,
  parseTagsFromCsv,
} from "@/tasks/create/composePayload";

type Props = {
  id?: string;
  disabled?: boolean;
  tagsCsv: string;
  onTagsCsvChange: (value: string) => void;
};

const TAG_HINT =
  "Comma-separated. 1–32 characters each: letters, numbers, and . _ - only (no spaces). Capitals are lowercased on save.";

/**
 * Pill-style tags editor that keeps the form's `tagsCsv` string in sync.
 * Validates each tag against the server wire rule before committing.
 */
export function TaskCreateTagsPillsField({
  id,
  disabled = false,
  tagsCsv,
  onTagsCsvChange,
}: Props) {
  const generatedId = useId();
  const inputId = id ?? generatedId;
  const hintId = `${inputId}-hint`;
  const errorId = `${inputId}-error`;
  const inputRef = useRef<HTMLInputElement>(null);
  const [draft, setDraft] = useState("");
  const [error, setError] = useState<string | null>(null);

  const tags = parseTagsFromCsv(tagsCsv);

  function commitTags(raw: string) {
    const candidates = raw
      .split(/[,;\n]+/)
      .map((t) => t.trim().toLowerCase())
      .filter(Boolean);

    if (candidates.length === 0) return;

    const next = [...tags];
    for (const tag of candidates) {
      if (!isValidTaskTag(tag)) {
        setError(
          `"${tag}" is invalid. Use 1–32 letters, numbers, . _ - only; must start with a letter or number.`,
        );
        return;
      }
      if (!next.includes(tag)) next.push(tag);
    }
    onTagsCsvChange(formatTagsCsv(next));
    setDraft("");
    setError(null);
  }

  function removeTag(tag: string) {
    onTagsCsvChange(formatTagsCsv(tags.filter((t) => t !== tag)));
    setError(null);
  }

  function handleKeyDown(e: ReactKeyboardEvent<HTMLInputElement>) {
    if (e.nativeEvent.isComposing || e.keyCode === 229) return;
    if (e.key === "Enter" || e.key === ",") {
      e.preventDefault();
      commitTags(draft);
    } else if (e.key === "Backspace" && draft === "" && tags.length > 0) {
      removeTag(tags[tags.length - 1]!);
    }
  }

  function focusInput(e: ReactMouseEvent<HTMLDivElement>) {
    if (disabled) return;
    if ((e.target as HTMLElement).closest("button")) return;
    inputRef.current?.focus();
  }

  return (
    <div className="task-create-tags-field">
      <FieldLabel htmlFor={inputId}>Tags</FieldLabel>
      <div
        className={[
          "task-create-tags-pills",
          disabled ? "task-create-tags-pills--disabled" : "",
          error ? "task-create-tags-pills--error" : "",
        ]
          .filter(Boolean)
          .join(" ")}
        onClick={focusInput}
      >
        {tags.map((tag) => (
          <span key={tag} className="task-create-tags-pill">
            {tag}
            <button
              type="button"
              className="task-create-tags-pill__remove"
              aria-label={`Remove ${tag}`}
              disabled={disabled}
              onClick={() => removeTag(tag)}
            >
              <span aria-hidden="true">×</span>
            </button>
          </span>
        ))}
        <input
          ref={inputRef}
          id={inputId}
          className="task-create-tags-pills__input"
          value={draft}
          disabled={disabled}
          placeholder={tags.length === 0 ? "e.g. backend, api" : "Add a tag"}
          aria-invalid={error ? true : undefined}
          aria-describedby={error ? errorId : hintId}
          onChange={(e) => {
            setDraft(e.target.value);
            setError(null);
          }}
          onKeyDown={handleKeyDown}
          onBlur={() => commitTags(draft)}
        />
      </div>
      {error ? (
        <p id={errorId} className="task-create-tags-field__error" role="alert">
          {error}
        </p>
      ) : (
        <p id={hintId} className="task-create-tags-field__hint">
          {TAG_HINT}
        </p>
      )}
    </div>
  );
}
