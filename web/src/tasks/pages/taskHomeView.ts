export type TaskHomeView = "list" | "board";

/** Parse `view` search param; unknown / missing → list. */
export function parseTaskHomeView(
  viewParam: string | null | undefined,
): TaskHomeView {
  if (viewParam === "board") return "board";
  return "list";
}

/**
 * Returns next search params with `view` set. Omits `view` for list
 * (default). Preserves other params.
 */
export function applyTaskHomeView(
  prev: URLSearchParams,
  view: TaskHomeView,
): URLSearchParams {
  const next = new URLSearchParams(prev);
  if (view === "list") {
    next.delete("view");
  } else {
    next.set("view", view);
  }
  return next;
}
