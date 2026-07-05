#!/usr/bin/env python3
"""Add SPDX-License-Identifier: MIT to project source files."""

from __future__ import annotations

import os
from pathlib import Path

ROOT = Path(__file__).resolve().parents[2]
SPDX_MARKER = "SPDX-License-Identifier"

SKIP_DIR_NAMES = {
    "node_modules",
    "bindings",
    "third_party",
    ".git",
    "bin",
    "dist",
    "frontend/dist",
}

SKIP_FILES = {
    "gradlew",
    "gradlew.bat",
}


def should_skip(path: Path) -> bool:
    rel = path.relative_to(ROOT)
    parts = rel.parts
    if any(part in SKIP_DIR_NAMES for part in parts):
        return True
    if path.name in SKIP_FILES:
        return True
    if "frontend" in parts and "bindings" in parts:
        return True
    return False


def add_spdx_go(text: str) -> str:
    lines = text.splitlines(keepends=True)
    insert_at = 0
    while insert_at < len(lines):
        line = lines[insert_at]
        stripped = line.strip()
        if stripped.startswith("//go:build") or stripped.startswith("// +build"):
            insert_at += 1
            continue
        break
    header = "// SPDX-License-Identifier: MIT\n"
    if insert_at > 0 and lines[insert_at - 1].endswith("\n") and insert_at < len(lines) and lines[insert_at].strip() == "":
        lines.insert(insert_at, header)
    elif insert_at > 0:
        lines.insert(insert_at, "\n" + header)
    else:
        lines.insert(0, header)
    return "".join(lines)


def add_spdx_js(text: str) -> str:
    return "// SPDX-License-Identifier: MIT\n" + text


def add_spdx_svelte(text: str) -> str:
    return "<!-- SPDX-License-Identifier: MIT -->\n" + text


def process_file(path: Path) -> bool:
    text = path.read_text(encoding="utf-8")
    if SPDX_MARKER in text:
        return False
    suffix = path.suffix.lower()
    if suffix == ".go":
        updated = add_spdx_go(text)
    elif suffix in {".ts", ".js", ".mjs", ".cjs"}:
        updated = add_spdx_js(text)
    elif suffix == ".svelte":
        updated = add_spdx_svelte(text)
    else:
        return False
    path.write_text(updated, encoding="utf-8")
    return True


def iter_sources() -> list[Path]:
    patterns = ["**/*.go", "**/*.ts", "**/*.js", "**/*.mjs", "**/*.svelte"]
    found: list[Path] = []
    for pattern in patterns:
        for path in ROOT.glob(pattern):
            if path.is_file() and not should_skip(path):
                found.append(path)
    return sorted(set(found))


def main() -> None:
    changed = 0
    for path in iter_sources():
        if process_file(path):
            changed += 1
            print(path.relative_to(ROOT))
    print(f"updated {changed} files")


if __name__ == "__main__":
    main()
