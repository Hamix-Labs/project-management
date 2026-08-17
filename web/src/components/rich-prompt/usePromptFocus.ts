import { useEffect, useRef, type RefObject } from "react";

const CHROME_INERT = [
  ".task-compose-page__sticky-footer",
  ".task-compose-page__topbar",
  ".task-compose-page__rail",
  ".task-compose-page__assist",
] as const;

function setInert(el: Element, on: boolean) {
  if (on) el.setAttribute("inert", "");
  else el.removeAttribute("inert");
}

function applyChromeInert(frame: HTMLElement, on: boolean) {
  for (const selector of CHROME_INERT) {
    document.querySelectorAll(selector).forEach((el) => {
      if (el.contains(frame) || frame.contains(el)) return;
      setInert(el, on);
    });
  }
  const parent = frame.parentElement;
  if (!parent) return;
  for (const sibling of Array.from(parent.children)) {
    if (sibling === frame || sibling.contains(frame)) continue;
    setInert(sibling, on);
  }
}

type Args = {
  expanded: boolean;
  onClose: () => void;
  frameRef: RefObject<HTMLElement | null>;
  restoreFocusRef: RefObject<HTMLElement | null>;
};

/** Escape, body scroll lock, inert on compose chrome, restore focus to Expand. */
export function usePromptFocus({
  expanded,
  onClose,
  frameRef,
  restoreFocusRef,
}: Args) {
  const wasExpanded = useRef(false);

  useEffect(() => {
    if (!expanded) return;
    const onKey = (e: KeyboardEvent) => {
      if (e.key === "Escape") {
        e.preventDefault();
        onClose();
      }
    };
    window.addEventListener("keydown", onKey);
    return () => window.removeEventListener("keydown", onKey);
  }, [expanded, onClose]);

  useEffect(() => {
    if (!expanded) return;
    const prev = document.body.style.overflow;
    document.body.style.overflow = "hidden";
    return () => {
      document.body.style.overflow = prev;
    };
  }, [expanded]);

  useEffect(() => {
    const frame = frameRef.current;
    if (!expanded || !frame) return;
    applyChromeInert(frame, true);
    return () => applyChromeInert(frame, false);
  }, [expanded, frameRef]);

  useEffect(() => {
    if (expanded) {
      wasExpanded.current = true;
      return;
    }
    if (!wasExpanded.current) return;
    restoreFocusRef.current?.focus();
  }, [expanded, restoreFocusRef]);
}
