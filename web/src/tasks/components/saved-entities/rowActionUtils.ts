/** Returns true when the event target is an interactive control that should not activate the row. */
export function isRowActionExcluded(
  target: EventTarget | null,
  extraSelectors = "",
): boolean {
  if (!(target instanceof Element)) return false;
  const base = "button, a, input, textarea, select, label";
  const selectors = extraSelectors ? `${base}, ${extraSelectors}` : base;
  return Boolean(target.closest(selectors));
}
