import { useEffect, useId, useRef, useState } from "react";
import {
  TimelineCalendarGlyph,
  TimelineCheckSmallGlyph,
  TimelineChevronGlyph,
} from "./TimelineGlyphs";
import {
  DEFAULT_TIMELINE_RANGE,
  TIMELINE_RANGE_OPTIONS,
  timelineRangeLabel,
} from "./timelineRange";
import type { TimelineRangeId } from "./timelineTypes";

type Props = {
  value: TimelineRangeId;
  onChange: (next: TimelineRangeId) => void;
};

export function TimelineRangeDropdown({
  value = DEFAULT_TIMELINE_RANGE,
  onChange,
}: Props) {
  const [open, setOpen] = useState(false);
  const rootRef = useRef<HTMLDivElement>(null);
  const listId = useId();

  useEffect(() => {
    if (!open) return;
    function onPointerDown(e: MouseEvent) {
      if (rootRef.current && !rootRef.current.contains(e.target as Node)) {
        setOpen(false);
      }
    }
    function onKey(e: KeyboardEvent) {
      if (e.key === "Escape") setOpen(false);
    }
    document.addEventListener("mousedown", onPointerDown);
    document.addEventListener("keydown", onKey);
    return () => {
      document.removeEventListener("mousedown", onPointerDown);
      document.removeEventListener("keydown", onKey);
    };
  }, [open]);

  return (
    <div ref={rootRef} className="task-home-timeline-range">
      <button
        type="button"
        className="task-home-timeline-range__trigger"
        aria-haspopup="listbox"
        aria-expanded={open}
        aria-controls={listId}
        onClick={() => setOpen((o) => !o)}
      >
        <TimelineCalendarGlyph className="task-home-timeline-range__icon" />
        <span>{timelineRangeLabel(value)}</span>
        <TimelineChevronGlyph
          className={
            open
              ? "task-home-timeline-range__chevron task-home-timeline-range__chevron--open"
              : "task-home-timeline-range__chevron"
          }
        />
      </button>
      {open ? (
        <ul
          id={listId}
          role="listbox"
          className="task-home-timeline-range__menu"
          aria-label="Time range"
        >
          {TIMELINE_RANGE_OPTIONS.map((opt) => {
            const selected = opt.id === value;
            return (
              <li key={opt.id} role="presentation">
                <button
                  type="button"
                  role="option"
                  aria-selected={selected}
                  className={
                    selected
                      ? "task-home-timeline-range__option task-home-timeline-range__option--selected"
                      : "task-home-timeline-range__option"
                  }
                  onClick={() => {
                    onChange(opt.id);
                    setOpen(false);
                  }}
                >
                  <span>{opt.label}</span>
                  {selected ? (
                    <TimelineCheckSmallGlyph className="task-home-timeline-range__check" />
                  ) : null}
                </button>
              </li>
            );
          })}
        </ul>
      ) : null}
    </div>
  );
}
