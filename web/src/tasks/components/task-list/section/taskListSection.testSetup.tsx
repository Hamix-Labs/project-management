import { render } from "@testing-library/react";
import type { ReactElement } from "react";
import { MemoryRouter } from "react-router-dom";
import { vi } from "vitest";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { ROUTER_FUTURE_FLAGS } from "../../../../lib/routerFutureFlags";
import { TASK_TEST_DEFAULTS } from "@/test/taskDefaults";

export function renderWithRouter(ui: ReactElement) {
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: { retry: false, staleTime: Infinity },
      mutations: { retry: false },
    },
  });
  return render(
    <QueryClientProvider client={queryClient}>
      <MemoryRouter future={ROUTER_FUTURE_FLAGS}>{ui}</MemoryRouter>
    </QueryClientProvider>,
  );
}

export function makeRow(
  id: string,
  title: string,
  extras: Partial<{
    status: import("@/types").Status;
    priority: import("@/types").Priority;
    pickup_not_before?: string;
    project_id?: string;
  }> = {},
) {
  return {
    id,
    title,
    initial_prompt: "",
    status: extras.status ?? ("ready" as const),
    priority: extras.priority ?? ("medium" as const),
    pickup_not_before: extras.pickup_not_before,
    project_id: extras.project_id,
    ...TASK_TEST_DEFAULTS,
    depth: 0,
  };
}

export const listPagerDefaults = {
  listPage: 0,
  listPageSize: 20,
  onListPageChange: vi.fn(),
  onListFiltersChange: vi.fn(),
  hasNextPage: false,
  hasPrevPage: false,
  rootTasksOnPage: 0,
};
