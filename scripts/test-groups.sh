#!/usr/bin/env bash
# Go test groups for CI matrix and scoped local runs.
# Source of truth for which packages belong to core/tasks/agents/harness.

repo_packages() {
  go list ./cmd/... ./internal/... ./pkgs/...
}

group_packages() {
  local group="$1"
  case "$group" in
    core)
      go list ./cmd/... ./internal/... ./pkgs/repo/... ./pkgs/gitcore/... ./pkgs/gitexec/... ./pkgs/gitwork/... \
        ./pkgs/obs/... \
        | grep -Ev '/(handlertest|agentreconcile)(/|$)'
      ;;
    tasks)
      go list ./pkgs/tasks/... ./pkgs/runners/handler ./pkgs/storekernel/...
      ;;
    handlertest)
      printf '%s\n' \
        github.com/AlexsanderHamir/Hamix/internal/handlertest/events \
        github.com/AlexsanderHamir/Hamix/internal/handlertest/cycles \
        github.com/AlexsanderHamir/Hamix/internal/handlertest/sse
      ;;
    handlertest-ui)
      printf '%s\n' \
        github.com/AlexsanderHamir/Hamix/internal/handlertest \
        github.com/AlexsanderHamir/Hamix/internal/handlertest/checklist \
        github.com/AlexsanderHamir/Hamix/internal/handlertest/compose \
        github.com/AlexsanderHamir/Hamix/internal/handlertest/shell
      ;;
    task-bcs)
      go list ./pkgs/projects/... ./pkgs/gitinventory/... ./pkgs/settings/... ./pkgs/taskcompose/... \
        ./pkgs/taskcore/... ./pkgs/taskchecklist/... ./pkgs/taskcycles/... ./pkgs/taskevents/... ./pkgs/draftassist/...
      ;;
    agents)
      go list ./pkgs/agents/... ./internal/taskapi/agentreconcile/... | grep -Ev '/harness'
      ;;
    harness)
      go list ./pkgs/agents/harness/...
      ;;
    *)
      echo "unknown test group: $group (valid: $(group_names))" >&2
      return 2
      ;;
  esac
}

group_names() {
  echo "core tasks handlertest handlertest-ui task-bcs agents harness"
}

assert_groups_cover_all() {
  local all grouped missing extra pkg count

  if ! command -v go >/dev/null 2>&1; then
    echo "go not on PATH" >&2
    return 1
  fi

  all="$(repo_packages | sort -u)"
  # Prefer here-string over `printf | grep` so `set -o pipefail` cannot treat
  # SIGPIPE from early grep -q exit as a failed membership check.
  count="$(grep -c . <<<"$all" || true)"
  if [[ "$count" -eq 0 ]]; then
    echo "go list returned no packages; is the module tree intact?" >&2
    return 1
  fi
  grouped=""
  for g in $(group_names); do
    grouped+="$(group_packages "$g" | sort -u)"$'\n'
  done
  grouped="$(grep -v '^$' <<<"$grouped" | sort -u || true)"

  missing=""
  while IFS= read -r pkg; do
    [[ -z "$pkg" ]] && continue
    if ! grep -Fxq -- "$pkg" <<<"$grouped"; then
      missing+="${pkg}"$'\n'
    fi
  done <<< "$all"

  extra=""
  while IFS= read -r pkg; do
    [[ -z "$pkg" ]] && continue
    if ! grep -Fxq -- "$pkg" <<<"$all"; then
      extra+="${pkg}"$'\n'
    fi
  done <<< "$grouped"

  if [[ -n "$missing" || -n "$extra" ]]; then
    echo "test group coverage mismatch:" >&2
    if [[ -n "$missing" ]]; then
      echo "  not assigned to any group:" >&2
      printf '%s' "$missing" | sed '/^$/d' | sed 's/^/    /' >&2
    fi
    if [[ -n "$extra" ]]; then
      echo "  assigned but not in repo:" >&2
      printf '%s' "$extra" | sed '/^$/d' | sed 's/^/    /' >&2
    fi
    echo "  fix: scripts/test-groups.sh" >&2
    return 1
  fi
}
