import { useId, useMemo, useRef, useState, type KeyboardEvent } from "react";
import { FieldLabel } from "@/shared/FieldLabel";
import {
  MAX_PROJECT_CONTEXT_TAG_CHARS,
  validateProjectContextTag,
} from "./projectContextLimits";

type Props = {
  id?: string;
  value: string;
  onChange: (value: string) => void;
  existingTags: string[];
  disabled?: boolean;
  /** Hidden input name for native form posts when needed. */
  name?: string;
};

type Option =
  | { type: "existing"; value: string }
  | { type: "create"; value: string };

export function ProjectContextTagPicker({
  id: idProp,
  value,
  onChange,
  existingTags,
  disabled = false,
  name = "tag",
}: Props) {
  const generatedId = useId();
  const id = idProp ?? generatedId;
  const listId = `${id}-listbox`;
  const hintId = `${id}-hint`;
  const inputRef = useRef<HTMLInputElement>(null);
  const [open, setOpen] = useState(false);
  const [activeIndex, setActiveIndex] = useState(0);

  const trimmed = value.trim();
  const error = validateProjectContextTag(value);
  const normalizedExisting = useMemo(() => {
    const seen = new Map<string, string>();
    for (const tag of existingTags) {
      const key = tag.trim().toLowerCase();
      if (!key || seen.has(key)) continue;
      const display = tag.trim();
      seen.set(key, display);
    }
    return [...seen.entries()]
      .sort((a, b) => a[1].localeCompare(b[1]))
      .map(([key, display]) => ({ key, display }));
  }, [existingTags]);

  const options = useMemo((): Option[] => {
    const q = trimmed.toLowerCase();
    const matches = normalizedExisting.filter((tag) =>
      q ? tag.key.includes(q) : true,
    );
    const exact = q
      ? normalizedExisting.some((tag) => tag.key === q)
      : false;
    const out: Option[] = matches.map((tag) => ({
      type: "existing",
      value: tag.display,
    }));
    if (trimmed && !exact) {
      out.unshift({ type: "create", value: trimmed });
    }
    return out;
  }, [normalizedExisting, trimmed]);

  function selectOption(option: Option) {
    onChange(option.value);
    setOpen(false);
    inputRef.current?.focus();
  }

  function onKeyDown(event: KeyboardEvent<HTMLInputElement>) {
    if (!open && (event.key === "ArrowDown" || event.key === "ArrowUp")) {
      setOpen(true);
      setActiveIndex(0);
      event.preventDefault();
      return;
    }
    if (!open) return;
    if (event.key === "Escape") {
      setOpen(false);
      event.preventDefault();
      return;
    }
    if (event.key === "ArrowDown") {
      setActiveIndex((i) => Math.min(i + 1, Math.max(options.length - 1, 0)));
      event.preventDefault();
      return;
    }
    if (event.key === "ArrowUp") {
      setActiveIndex((i) => Math.max(i - 1, 0));
      event.preventDefault();
      return;
    }
    if (event.key === "Enter" && options[activeIndex]) {
      selectOption(options[activeIndex]);
      event.preventDefault();
    }
  }

  return (
    <div className="field grow pc__tag-picker">
      <FieldLabel htmlFor={id} requirement="required">
        Tag
      </FieldLabel>
      <div className="pc__tag-picker-control">
        <input
          ref={inputRef}
          id={id}
          name={name}
          value={value}
          disabled={disabled}
          required
          aria-required="true"
          aria-invalid={Boolean(error) && trimmed.length > 0}
          aria-autocomplete="list"
          aria-controls={listId}
          aria-expanded={open}
          aria-describedby={hintId}
          maxLength={MAX_PROJECT_CONTEXT_TAG_CHARS}
          placeholder="e.g. Codebase tour"
          autoComplete="off"
          onChange={(event) => {
            onChange(event.target.value);
            setOpen(true);
            setActiveIndex(0);
          }}
          onFocus={() => setOpen(true)}
          onBlur={() => {
            // Defer so option click can fire first.
            window.setTimeout(() => setOpen(false), 120);
          }}
          onKeyDown={onKeyDown}
        />
        {open && options.length > 0 ? (
          <ul
            id={listId}
            className="pc__tag-picker-list"
            role="listbox"
            aria-label="Tag choices"
          >
            {options.map((option, index) => {
              const isCreate = option.type === "create";
              return (
                <li
                  key={`${option.type}:${option.value}`}
                  role="option"
                  aria-selected={index === activeIndex}
                  className={[
                    "pc__tag-picker-option",
                    isCreate ? "pc__tag-picker-option--create" : "",
                    index === activeIndex ? "pc__tag-picker-option--active" : "",
                  ]
                    .filter(Boolean)
                    .join(" ")}
                >
                  <button
                    type="button"
                    tabIndex={-1}
                    disabled={disabled}
                    onMouseDown={(event) => event.preventDefault()}
                    onClick={() => selectOption(option)}
                  >
                    <span className="pc__tag-picker-option-label">
                      {isCreate ? `Create “${option.value}”` : option.value}
                    </span>
                    <span className="pc__tag-picker-option-meta">
                      {isCreate ? "New tag" : "Existing tag"}
                    </span>
                  </button>
                </li>
              );
            })}
          </ul>
        ) : null}
      </div>
      <p id={hintId} className="pc__field-hint">
        Groups this memory with others that share the same tag.{" "}
        {trimmed.length}/{MAX_PROJECT_CONTEXT_TAG_CHARS}
      </p>
      {error && trimmed.length > 0 ? (
        <p className="pd__inline-error" role="alert">
          {error}
        </p>
      ) : null}
    </div>
  );
}
