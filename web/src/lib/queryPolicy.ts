/** Central staleTime / gcTime tiers for TanStack Query. See ADR-0025. */
export const QUERY_POLICY = {
  defaultStaleTimeMs: 15_000,
  gcTimeMs: 5 * 60_000,
  shellStaleTimeMs: 5 * 60_000,
  listStaleTimeMs: 60_000,
  detailStaleTimeMs: 30_000,
  prefetchStaleTimeMs: 30_000,
  persistMaxAgeMs: 30 * 60_000,
} as const;
