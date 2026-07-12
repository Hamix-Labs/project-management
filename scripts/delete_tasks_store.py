#!/usr/bin/env python3
from __future__ import annotations

import pathlib
import shutil

ROOT = pathlib.Path(__file__).resolve().parents[1]
STORE = ROOT / "pkgs" / "tasks" / "store"
DST = ROOT / "internal" / "taskapi" / "composition" / "store_task_git_test.go"

src = STORE / "store_task_git_test.go"
if src.exists():
    text = src.read_text(encoding="utf-8")
    text = text.replace("package store_test", "package composition_test")
    text = text.replace("store.NewStore", "composition.NewAPI")
    text = text.replace(
        '"github.com/AlexsanderHamir/Hamix/pkgs/tasks/store"',
        '"github.com/AlexsanderHamir/Hamix/internal/taskapi/composition"',
    )
    DST.write_text(text, encoding="utf-8")

if STORE.exists():
    shutil.rmtree(STORE)
    print("deleted", STORE)
