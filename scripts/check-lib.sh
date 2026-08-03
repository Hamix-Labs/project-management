# Shared step runner for check-go.sh and check-web.sh.
# Sourced from repo root after cd. Not executed directly.

: "${CHECK_BANNER:=Hamix check}"
: "${CHECK_SECTION:=go}"
: "${VERBOSE:=0}"
: "${STEP:=0}"
: "${PASSED:=0}"
: "${TOTAL:=0}"
: "${SECTION_TIME:=0}"
: "${STEP_STATS:=}"
: "${CHECK_STEP_BUDGET_SECS:=120}"

if [[ -t 1 ]]; then
  C_RESET=$'\033[0m'
  C_CYAN=$'\033[36m'
  C_GREEN=$'\033[32m'
  C_RED=$'\033[31m'
else
  C_RESET= C_CYAN= C_GREEN= C_RED=
fi

format_secs() {
  local s="$1"
  if [[ "$s" -lt 60 ]]; then
    printf '%ss' "$s"
  else
    printf '%sm%02ss' $((s / 60)) $((s % 60))
  fi
}

step_prefix() {
  STEP=$((STEP + 1))
  printf '[%d/%d] ' "$STEP" "$TOTAL"
}

fail_step() {
  local name="$1"
  local code="${2:-1}"
  local fix="${3:-}"
  echo ""
  echo "${C_RED}check FAILED: ${name} (${STEP}/${TOTAL})${C_RESET}" >&2
  if [[ -n "$fix" ]]; then
    echo "  fix: $fix" >&2
  fi
  exit "$code"
}

# Fail when a check-script step exceeds CHECK_STEP_BUDGET_SECS (install excluded).
enforce_step_budget() {
  local label="$1"
  local elapsed="$2"

  case "$label" in
    "npm ci") return 0 ;;
  esac

  if [[ "$elapsed" -le "$CHECK_STEP_BUDGET_SECS" ]]; then
    return 0
  fi

  if [[ -n "${GITHUB_ACTIONS:-}" ]]; then
    printf '::error title=check step budget::%s exceeded step budget (%s > %ss). Split the suite, slim the harness, or speed up the tests.\n' \
      "$label" "$(format_secs "$elapsed")" "$CHECK_STEP_BUDGET_SECS"
  fi

  echo ""
  echo "${C_RED}check FAILED: ${label} exceeded step budget ($(format_secs "$elapsed") > ${CHECK_STEP_BUDGET_SECS}s)${C_RESET}" >&2
  echo "  This step must finish within ${CHECK_STEP_BUDGET_SECS}s (keep check steps under two minutes)." >&2
  echo "  Split the suite, slim the harness, or speed up the tests — do not raise the budget casually." >&2
  exit 1
}

complete_ok() {
  local detail="${1:-}"
  local elapsed=$((SECONDS - CHECK_START))
  echo ""
  printf '%bcheck OK  %d/%d passed  %s%b\n' "$C_GREEN" "$PASSED" "$TOTAL" "$(format_secs "$elapsed")" "$C_RESET"
  if [[ -n "$detail" ]]; then
    echo "  ($detail)"
  fi
  exit 0
}

print_ok_line() {
  local label="$1"
  local elapsed="$2"
  local stats="${3:-}"
  local pad=$((22 - ${#label}))
  if [[ $pad -lt 1 ]]; then pad=1; fi
  printf '%*s%b ok %5s%b' "$pad" '' "$C_GREEN" "$(format_secs "$elapsed")" "$C_RESET"
  if [[ -n "$stats" ]]; then
    printf '  (%s)' "$stats"
  fi
  echo
}

add_section_time() {
  local elapsed="$1"
  SECTION_TIME=$((SECTION_TIME + elapsed))
}

run_capture() {
  local label="$1"
  shift
  local start=$SECONDS
  local log
  log="$(mktemp "${TMPDIR:-/tmp}/hamix-check.XXXXXX")"

  step_prefix
  printf '%s ' "$label"

  set +e
  "$@" >"$log" 2>&1
  local code=$?
  set -e

  local elapsed=$((SECONDS - start))
  add_section_time "$elapsed"

  if [[ $code -eq 0 ]]; then
    PASSED=$((PASSED + 1))
    print_ok_line "$label" "$elapsed" "${STEP_STATS:-}"
    STEP_STATS=""
    rm -f "$log"
    enforce_step_budget "$label" "$elapsed"
    return 0
  fi

  echo "${C_RED}FAILED${C_RESET}"
  cat "$log"
  rm -f "$log"
  fail_step "$label" "$code"
}

run_stream() {
  local label="$1"
  shift
  local start=$SECONDS

  step_prefix
  echo "${C_CYAN}${label} ...${C_RESET}"

  set +e
  "$@"
  local code=$?
  set -e

  local elapsed=$((SECONDS - start))
  add_section_time "$elapsed"

  if [[ $code -eq 0 ]]; then
    PASSED=$((PASSED + 1))
    enforce_step_budget "$label" "$elapsed"
    return 0
  fi
  fail_step "$label" "$code"
}

run_cmd() {
  if [[ "$VERBOSE" == "1" ]]; then
    run_stream "$@"
  else
    run_capture "$@"
  fi
}

print_banner() {
  echo "$CHECK_BANNER"
  echo ""
}
