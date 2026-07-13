import type { RefObject } from "react";

export type BulkSelectionProps = {
  isSelected: (id: string) => boolean;
  onRowToggle: (id: string) => void;
  allVisibleSelected: boolean;
  someVisibleSelected: boolean;
  onToggleAllVisible: () => void;
};

export function syncHeaderCheckboxIndeterminate(
  selection: BulkSelectionProps | undefined,
  headerCheckboxRef: RefObject<HTMLInputElement | null>,
): void {
  if (!selection || !headerCheckboxRef.current) return;
  headerCheckboxRef.current.indeterminate = selection.someVisibleSelected;
}
