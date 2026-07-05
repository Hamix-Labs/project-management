import { EmptyState } from "@/shared/EmptyState";

type PhaseFilteredEmptyKind = "cursor" | "audit";

const COPY: Record<
  PhaseFilteredEmptyKind,
  { filteredTitle: (label: string) => string; emptyTitle: string; emptyDescription: string }
> = {
  cursor: {
    filteredTitle: (label) => `No Cursor output for ${label}`,
    emptyTitle: "No Cursor output yet",
    emptyDescription: "Stream lines appear here as the agent runs.",
  },
  audit: {
    filteredTitle: (label) => `No audit events for ${label}`,
    emptyTitle: "No audit events yet",
    emptyDescription: "System events for this attempt appear here.",
  },
};

type Props = {
  filterLabel: string | null;
  kind: PhaseFilteredEmptyKind;
  onClearPhaseFilter: () => void;
};

export function PhaseFilteredEmpty({
  filterLabel,
  kind,
  onClearPhaseFilter,
}: Props) {
  const copy = COPY[kind];
  if (filterLabel) {
    return (
      <EmptyState
        title={copy.filteredTitle(filterLabel)}
        description="Try another phase or show all activity."
        density="compact"
        hideIcon
        action={{ label: "Show all phases", onClick: onClearPhaseFilter }}
      />
    );
  }
  return (
    <EmptyState
      title={copy.emptyTitle}
      description={copy.emptyDescription}
      density="compact"
      hideIcon
    />
  );
}
