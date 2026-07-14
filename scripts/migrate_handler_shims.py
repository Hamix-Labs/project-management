#!/usr/bin/env python3
"""Remove handler compat shims: point at taskcorehandler, realtime, middleware."""
from __future__ import annotations

import pathlib
import re

ROOT = pathlib.Path(__file__).resolve().parents[1]
HANDLER = ROOT / "pkgs" / "tasks" / "handler"

TASKCORE_REPLACEMENTS = [
    ("listResponse", "taskcorehandler.ListResponse"),
    ("taskStatsResponse", "taskcorehandler.TaskStatsResponse"),
    ("dependsOnWire", "taskcorehandler.DependsOnWire"),
    ("taskCreateJSON", "taskcorehandler.TaskCreateJSON"),
    ("taskPatchJSON", "taskcorehandler.TaskPatchJSON"),
    ("buildListResponse", "taskcorehandler.BuildListResponse"),
    ("taskStatsResponseFromStore", "taskcorehandler.TaskStatsResponseFromStore"),
    ("decodeComposePayload", "taskcorehandler.DecodeComposePayload"),
    ("composePayloadToRaw", "taskcorehandler.ComposePayloadToRaw"),
    ("maxListIntQueryParamBytes", "taskcorehandler.MaxListIntQueryParamBytes"),
    ("maxListAfterIDParamBytes", "taskcorehandler.MaxListAfterIDParamBytes"),
    ("parseListParams", "taskcorehandler.ParseListParams"),
    ("pickupNotBeforeMinAllowed", "taskcorehandler.PickupNotBeforeMinAllowed"),
]

SSE_REPLACEMENTS = [
    ("TaskChangeEvent", "realtime.Event"),
    ("TaskChangeType", "realtime.ChangeType"),
    ("TaskCreated", "realtime.TaskCreated"),
    ("TaskUpdated", "realtime.TaskUpdated"),
    ("TaskDeleted", "realtime.TaskDeleted"),
    ("TaskCycleChanged", "realtime.TaskCycleChanged"),
    ("TaskEventChanged", "realtime.TaskEventChanged"),
    ("AgentRunProgress", "realtime.AgentRunProgress"),
    ("SettingsChanged", "realtime.SettingsChanged"),
    ("Resync", "realtime.Resync"),
]

MIDDLEWARE_REPLACEMENTS = [
    ("handler.WithRecovery", "middleware.WithRecovery"),
    ("handler.WithHTTPMetrics", "middleware.WithHTTPMetrics"),
    ("handler.WithAccessLog", "middleware.WithAccessLog"),
    ("handler.RateLimitPerMinuteConfigured", "middleware.RateLimitPerMinuteConfigured"),
    ("handler.APIAuthEnabled", "middleware.APIAuthEnabled"),
    ("handler.MaxRequestBodyBytesConfigured", "middleware.MaxRequestBodyBytesConfigured"),
    ("handler.RequestTimeout", "middleware.RequestTimeout"),
    ("handler.IdempotencyTTL", "middleware.IdempotencyTTL"),
    ("handler.IdempotencyCacheLimits", "middleware.IdempotencyCacheLimits"),
    ("clearIdempotencyStateForTest", "middleware.ClearIdempotencyStateForTest"),
    ("HasValidBearerToken", "middleware.HasValidBearerToken"),
]


def ensure_import(text: str, imp: str) -> str:
    if imp in text:
        return text
    # insert after package handler line block
    m = re.search(r"(package handler\n\nimport \(\n)", text)
    if m:
        return text.replace(m.group(1), m.group(1) + "\t" + imp + "\n")
    m = re.search(r"(package handler\n\nimport )", text)
    if m:
        return text.replace(m.group(1), m.group(1) + imp + "\n")
    return text


def process_handler_go(path: pathlib.Path) -> None:
    if path.name in {
        "handler_taskcore_compat.go",
        "middleware_shim.go",
        "sse_types.go",
        "git_store_adapter.go",
    }:
        return
    text = path.read_text(encoding="utf-8")
    orig = text
    for old, new in TASKCORE_REPLACEMENTS:
        text = re.sub(r"\b" + re.escape(old) + r"\b", new, text)
    for old, new in SSE_REPLACEMENTS:
        text = re.sub(r"\b" + re.escape(old) + r"\b", new, text)
    for old, new in MIDDLEWARE_REPLACEMENTS:
        text = text.replace(old, new)
    if "taskcorehandler." in text and "taskcorehandler" not in text:
        pass
    if "taskcorehandler." in text and 'taskcorehandler "' not in text:
        text = ensure_import(text, 'taskcorehandler "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/handler"')
    if "realtime." in text and 'realtime "' not in text and "pkgs/tasks/realtime" not in text:
        text = ensure_import(text, '"github.com/AlexsanderHamir/Hamix/pkgs/tasks/realtime"')
    if text != orig:
        path.write_text(text, encoding="utf-8")
        print("updated", path.relative_to(ROOT))


def process_other(path: pathlib.Path) -> None:
    text = path.read_text(encoding="utf-8")
    orig = text
    for old, new in SSE_REPLACEMENTS:
        text = re.sub(r"\bhandler\." + re.escape(old) + r"\b", "realtime." + new.split(".")[1], text)
    for old, new in MIDDLEWARE_REPLACEMENTS:
        text = text.replace("handler." + old.split(".")[-1] if "." in old else old, new)
    if text != orig:
        path.write_text(text, encoding="utf-8")
        print("updated", path.relative_to(ROOT))


def main() -> None:
    for path in HANDLER.rglob("*.go"):
        process_handler_go(path)
    for rel in [
        "cmd/taskapi/run_helpers.go",
        "cmd/taskapi/run_http.go",
        "internal/handlertest/health_test.go",
        "internal/taskapi/agentreconcile/agent_real_cursor_e2e_test.go",
        "internal/taskapi/agentworker/sse.go",
    ]:
        p = ROOT / rel
        if p.exists():
            process_other(p)
    for shim in [
        "handler_taskcore_compat.go",
        "middleware_shim.go",
        "sse_types.go",
        "git_store_adapter.go",
    ]:
        p = HANDLER / shim
        if p.exists():
            p.unlink()
            print("deleted", p.relative_to(ROOT))


if __name__ == "__main__":
    main()
