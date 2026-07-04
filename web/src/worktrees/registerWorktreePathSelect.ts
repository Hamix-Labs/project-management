const COPY = {
  loading: "Loading linked worktrees…",
  unavailable: "Registered checkout unavailable",
  empty: "No unregistered worktrees for this repository.",
  prompt: "Select a linked worktree",
} as const;

export type RegisterWorktreePathSelectState = {
  loading: boolean;
  optionCount: number;
  inventoryError: boolean;
};

export function registerWorktreePathPlaceholder(
  state: RegisterWorktreePathSelectState,
): string {
  if (state.loading) return COPY.loading;
  if (state.inventoryError) return COPY.unavailable;
  if (state.optionCount === 0) return COPY.empty;
  return COPY.prompt;
}

export function registerWorktreePathDisabled(
  state: RegisterWorktreePathSelectState & {
    pending: boolean;
    inventoryRefreshPending?: boolean;
  },
): boolean {
  return (
    state.pending ||
    state.loading ||
    state.inventoryError ||
    state.optionCount === 0 ||
    state.inventoryRefreshPending === true
  );
}
