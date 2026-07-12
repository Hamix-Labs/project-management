#!/usr/bin/env python3
"""One-off: replace pkgs/tasks/store imports with internal/taskapi/composition."""
from __future__ import annotations

import pathlib
import re

ROOT = pathlib.Path(__file__).resolve().parents[1]
SKIP_PREFIXES = ("pkgs/tasks/store", ".git", "node_modules", ".codegraph")

REPLACEMENTS = [
    (
        "github.com/AlexsanderHamir/Hamix/pkgs/tasks/store",
        "github.com/AlexsanderHamir/Hamix/internal/taskapi/composition",
    ),
    ("*store.Store", "*composition.API"),
    ("store.NewStore(", "composition.NewAPI("),
    ("store.DefaultReadyTimeout", "taskcorestore.DefaultReadyTimeout"),
    ("store.ShouldNotifyReadyNow", "taskcorestore.ShouldNotifyReadyNow"),
    ("store.ThreadEntriesForDisplay", "taskeventsstore.ThreadEntriesForDisplay"),
    ("store.FindWorktreeInInventory", "gitinventorystore.FindWorktreeInInventory"),
    ("store.CreateDefaultProjectForRepo", "projectsstore.CreateDefaultProjectForRepo"),
]


def should_skip(rel: str) -> bool:
    return any(rel.startswith(p) or f"/{p}/" in rel for p in SKIP_PREFIXES)


def main() -> None:
    for path in ROOT.rglob("*.go"):
        rel = path.relative_to(ROOT).as_posix()
        if should_skip(rel):
            continue
        text = path.read_text(encoding="utf-8")
        if "pkgs/tasks/store" not in text and "store.NewStore" not in text and "*store.Store" not in text:
            if not any(
                x in text
                for x in (
                    "store.DefaultReadyTimeout",
                    "store.ShouldNotifyReadyNow",
                    "store.ThreadEntriesForDisplay",
                )
            ):
                continue
        orig = text
        for old, new in REPLACEMENTS:
            text = text.replace(old, new)
        if "internal/taskapi/composition" in text:
            text = re.sub(
                r'\sstore "github.com/AlexsanderHamir/Hamix/internal/taskapi/composition"',
                ' composition "github.com/AlexsanderHamir/Hamix/internal/taskapi/composition"',
                text,
            )
        if text != orig:
            path.write_text(text, encoding="utf-8")
            print("updated", rel)


if __name__ == "__main__":
    main()
