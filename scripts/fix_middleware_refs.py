#!/usr/bin/env python3
"""Fix handler tests after middleware_shim deletion."""
from __future__ import annotations

import pathlib
import re

ROOT = pathlib.Path(__file__).resolve().parents[1]
HANDLER = ROOT / "pkgs" / "tasks" / "handler"

REPLACEMENTS = [
    ("WithAccessLog(", "middleware.WithAccessLog("),
    ("WithRecovery(", "middleware.WithRecovery("),
    ("WithHTTPMetrics(", "middleware.WithHTTPMetrics("),
    ("WithRateLimit(", "middleware.WithRateLimit("),
    ("WithRequestTimeout(", "middleware.WithRequestTimeout("),
    ("WithMaxRequestBody(", "middleware.WithMaxRequestBody("),
    ("WithIdempotency(", "middleware.WithIdempotency("),
    ("WithAPIAuth(", "middleware.WithAPIAuth("),
    ("APIAuthEnabled()", "middleware.APIAuthEnabled()"),
]

ACCESS_LOG_FIX = re.compile(
    r"middleware\.WithAccessLog\(([^)]+)\)(?!\s*,\s*calltrace\.Path)"
)


def ensure_imports(text: str) -> str:
    if "middleware." in text and '"github.com/AlexsanderHamir/Hamix/pkgs/tasks/middleware"' not in text:
        text = text.replace(
            "import (",
            'import (\n\t"github.com/AlexsanderHamir/Hamix/pkgs/tasks/middleware"',
            1,
        )
    if "calltrace.Path" in text and "pkgs/obs/calltrace" not in text:
        text = text.replace(
            "import (",
            'import (\n\t"github.com/AlexsanderHamir/Hamix/pkgs/obs/calltrace"',
            1,
        )
    return text


def fix_access_log(text: str) -> str:
    def repl(m: re.Match[str]) -> str:
        inner = m.group(1)
        if "calltrace.Path" in inner:
            return m.group(0)
        return f"middleware.WithAccessLog({inner}, calltrace.Path)"
    return ACCESS_LOG_FIX.sub(repl, text)


def main() -> None:
    shim = HANDLER / "middleware_shim.go"
    if shim.exists():
        shim.unlink()
        print("deleted", shim)

    for path in HANDLER.rglob("*.go"):
        if path.name == "middleware_shim.go":
            continue
        text = path.read_text(encoding="utf-8")
        orig = text
        text = text.replace("func Testmiddleware.HasValidBearerToken(", "func TestHasValidBearerToken(")
        text = text.replace(
            "func Testmiddleware.HasValidBearerToken_caseInsensitiveScheme(",
            "func TestHasValidBearerToken_caseInsensitiveScheme(",
        )
        for old, new in REPLACEMENTS:
            text = text.replace(old, new)
        text = fix_access_log(text)
        text = ensure_imports(text)
        if text != orig:
            path.write_text(text, encoding="utf-8")
            print("updated", path.relative_to(ROOT))


if __name__ == "__main__":
    main()
