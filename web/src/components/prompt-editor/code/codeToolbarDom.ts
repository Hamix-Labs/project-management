import ReactDOM from "react-dom/client";

/** Vite's CJS/ESM interop can nest the client entry under `default`. */
export function resolveCreateRoot(): typeof ReactDOM.createRoot {
  const mod = ReactDOM as typeof ReactDOM & {
    default?: { createRoot?: typeof ReactDOM.createRoot };
  };
  const fn = mod.createRoot ?? mod.default?.createRoot;
  if (typeof fn !== "function") {
    throw new Error("react-dom/client createRoot is unavailable");
  }
  return fn;
}

/**
 * BlockNote positions its <select> absolutely inside a zero-width wrapper. Left
 * visible it overlaps our toolbar, so collapse both once the toolbar is mounted.
 */
export function hideNativeSelect(select: HTMLSelectElement) {
  select.hidden = true;
  select.setAttribute("aria-hidden", "true");
  select.tabIndex = -1;
  select.style.setProperty("display", "none", "important");
  const wrap = select.parentElement;
  if (wrap instanceof HTMLElement) {
    wrap.style.cssText =
      "position:absolute;width:0;height:0;overflow:hidden;opacity:0;pointer-events:none;border:none;margin:0;padding:0;";
  }
}

/** Anchor a portaled toolbar to the top-right of its code block. */
export function positionMount(
  host: HTMLElement,
  block: HTMLElement,
  mountEl: HTMLElement,
) {
  const hostRect = host.getBoundingClientRect();
  const blockRect = block.getBoundingClientRect();
  const top = blockRect.top - hostRect.top + host.scrollTop + 8;
  const right = hostRect.right - blockRect.right + host.scrollLeft + 10;
  mountEl.style.top = `${Math.max(0, top)}px`;
  mountEl.style.right = `${Math.max(0, right)}px`;
  mountEl.style.left = "auto";
}

export function findSelect(block: Element): HTMLSelectElement | null {
  const select = block.querySelector(":scope > div select");
  return select instanceof HTMLSelectElement ? select : null;
}

export function blockIdOf(block: HTMLElement): string | null {
  return block.closest("[data-id]")?.getAttribute("data-id") ?? null;
}

/**
 * Mounts are keyed by block id so a node view swap reuses its toolbar instead of
 * remounting. Detached fixtures without a block id fall back to a per-element key.
 */
export function createBlockKeyFactory(): (block: HTMLElement) => string {
  const fallbackKeys = new WeakMap<Element, string>();
  let seq = 0;
  return (block) => {
    const id = blockIdOf(block);
    if (id) return `id:${id}`;
    let key = fallbackKeys.get(block);
    if (!key) {
      seq += 1;
      key = `el:${seq}`;
      fallbackKeys.set(block, key);
    }
    return key;
  };
}
