import {
  useMutation,
  useQueryClient,
  type QueryClient,
} from "@tanstack/react-query";
import { useCallback } from "react";
import {
  rumMutationRolledBack,
  rumMutationSettled,
  type RUMMutationKind,
} from "@/observability";
import { useOptionalToast } from "@/shared/toast";
import { useRolloutFlags } from "@/settings";
import type { TaskInvalidationScope } from "@/lib/queryInvalidation";
import {
  beginGuardedTaskWrite,
  endGuardedTaskWrite,
  invalidateTaskCacheAsync,
  type GuardedWriteContext,
} from "@/tasks/mutations";

export type GuardedMutationContextBase = {
  startedAtMs: number;
  guarded: boolean;
};

export type UseGuardedTaskMutationOptions<
  TVariables extends { id: string },
  TContext extends GuardedMutationContextBase,
  TData = unknown,
> = {
  rumKind: RUMMutationKind;
  mutationFn: (variables: TVariables) => Promise<TData>;
  /**
   * Apply optimistic cache writes when `guard.guarded`. Always return a
   * context (including the empty/unguarded shape) so onError/onSettled
   * share one lifecycle.
   */
  applyOptimistic: (args: {
    queryClient: QueryClient;
    variables: TVariables;
    guard: GuardedWriteContext;
  }) => Promise<TContext> | TContext;
  restoreOptimistic: (args: {
    queryClient: QueryClient;
    variables: TVariables;
    context: TContext;
  }) => void;
  /** True when restore actually undid client state (RUM rollback numerator). */
  didRollBack: (context: TContext) => boolean;
  errorToast: string;
  invalidateScopes?: TaskInvalidationScope[];
  onSuccessSideEffect?: (args: {
    variables: TVariables;
    context: TContext | undefined;
  }) => void | Promise<void>;
};

/**
 * Shared useMutation Apply glue for guarded optimistic task writes:
 * begin guard → optimistic → rollback/toast → invalidate → end guard.
 */
export function useGuardedTaskMutation<
  TVariables extends { id: string },
  TContext extends GuardedMutationContextBase,
  TData = unknown,
>(options: UseGuardedTaskMutationOptions<TVariables, TContext, TData>) {
  const queryClient = useQueryClient();
  const toast = useOptionalToast();
  const { optimisticMutationsEnabled } = useRolloutFlags();
  const {
    rumKind,
    mutationFn,
    applyOptimistic,
    restoreOptimistic,
    didRollBack,
    errorToast,
    invalidateScopes = [{ scope: "listStats" }],
    onSuccessSideEffect,
  } = options;

  const mutation = useMutation<TData, unknown, TVariables, TContext>({
    mutationFn,
    onMutate: async (variables) => {
      const guard = beginGuardedTaskWrite({
        taskId: variables.id,
        optimisticEnabled: optimisticMutationsEnabled,
        rumKind,
      });
      return applyOptimistic({ queryClient, variables, guard });
    },
    onError: (_err, variables, context) => {
      if (context) {
        restoreOptimistic({ queryClient, variables, context });
        if (didRollBack(context)) {
          rumMutationRolledBack(
            rumKind,
            performance.now() - context.startedAtMs,
          );
        }
      }
      toast.error(errorToast);
      rumMutationSettled(
        rumKind,
        context ? performance.now() - context.startedAtMs : 0,
        0,
      );
    },
    onSuccess: async (_data, variables, context) => {
      await invalidateTaskCacheAsync(queryClient, ...invalidateScopes);
      await onSuccessSideEffect?.({ variables, context });
      if (context) {
        rumMutationSettled(
          rumKind,
          performance.now() - context.startedAtMs,
          200,
        );
      }
    },
    onSettled: (_data, _err, variables, context) => {
      if (context?.guarded) {
        endGuardedTaskWrite(variables.id);
      }
    },
  });

  const mutate = useCallback(
    (variables: TVariables) => {
      mutation.mutate(variables);
    },
    [mutation],
  );

  const reset = useCallback(() => {
    mutation.reset();
  }, [mutation]);

  return {
    mutate,
    mutation,
    reset,
    isPending: mutation.isPending,
    isError: mutation.isError,
    isSuccess: mutation.isSuccess,
    error: mutation.error,
    variables: mutation.variables,
    isIdle: mutation.isIdle,
  };
}
