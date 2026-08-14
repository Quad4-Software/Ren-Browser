#!/usr/bin/env python3
# SPDX-License-Identifier: 0BSD

"""Fast git-index tree manifest for renbrowser.rsm signing.

Hashes index blobs (same bytes as git show :path) via git cat-file --batch.
Output format matches build/scripts/tree-manifest.sh generate.
"""

from __future__ import annotations

import hashlib
import os
import subprocess
import sys
import threading
from pathlib import Path

MANIFEST_HEADER = "# renbrowser tree manifest v1"
EXCLUDE_RSM = "renbrowser.rsm"
ALLOWED_MODES = frozenset({"100644", "100755"})


def _repo_root() -> Path:
    return Path(__file__).resolve().parents[2]


def is_excluded_path(path: str) -> bool:
    if path == EXCLUDE_RSM:
        return True
    if path == "vendor" or path.startswith("vendor/"):
        return True
    if "/vendor/" in path or path.endswith("/vendor"):
        return True
    return False


def _parse_ls_files_z(raw: bytes) -> list[tuple[str, str, str]]:
    entries: list[tuple[str, str, str]] = []
    for chunk in raw.split(b"\0"):
        if not chunk:
            continue
        tab = chunk.find(b"\t")
        if tab < 0:
            continue
        meta = chunk[:tab].decode("ascii", errors="replace")
        path = chunk[tab + 1 :].decode("utf-8", errors="surrogateescape")
        parts = meta.split()
        if len(parts) < 3:
            continue
        mode, oid, _stage = parts[0], parts[1], parts[2]
        entries.append((path, mode, oid))
    return entries


def _read_batch_blob(stdout, header: bytes) -> bytes | None:
    parts = header.strip().split()
    if len(parts) == 2 and parts[1] == b"missing":
        return None
    if len(parts) != 3:
        raise RuntimeError(f"unexpected cat-file header: {header!r}")
    _oid, typ, size_s = parts
    if typ == b"missing":
        return None
    size = int(size_s)
    data = stdout.read(size)
    if len(data) != size:
        raise RuntimeError("short read from git cat-file --batch")
    if typ == b"blob":
        delim = stdout.read(1)
        if delim != b"\n":
            raise RuntimeError("expected newline delimiter after cat-file blob")
    return data


def _feed_cat_file_stdin(stdin, requests: list[str]) -> None:
    try:
        for req in requests:
            stdin.write(f"{req}\n".encode("ascii"))
        stdin.close()
    except BrokenPipeError:
        pass


def _run_cat_file_batch(
    root: Path,
    env: dict[str, str],
    requests: list[str],
) -> list[bytes | None]:
    if not requests:
        return []

    proc = subprocess.Popen(
        ["git", "cat-file", "--batch"],
        cwd=root,
        stdin=subprocess.PIPE,
        stdout=subprocess.PIPE,
        env=env,
    )
    assert proc.stdin is not None
    assert proc.stdout is not None

    writer = threading.Thread(
        target=_feed_cat_file_stdin,
        args=(proc.stdin, requests),
        daemon=True,
    )
    writer.start()
    blobs: list[bytes | None] = []
    try:
        for _req in requests:
            header = proc.stdout.readline()
            if not header:
                raise RuntimeError("git cat-file --batch closed stdout early")
            blobs.append(_read_batch_blob(proc.stdout, header))
        writer.join(timeout=30)
        proc.wait()
        if proc.returncode not in (0, None):
            raise RuntimeError(f"git cat-file --batch exited {proc.returncode}")
    except Exception:
        proc.kill()
        writer.join(timeout=5)
        proc.wait()
        raise
    return blobs


def _git_env() -> dict[str, str]:
    env = os.environ.copy()
    env["LC_ALL"] = "C"
    return env


def _indexed_file_rows(root: Path) -> list[tuple[str, str, str]]:
    env = _git_env()
    raw = subprocess.check_output(
        ["git", "ls-files", "-s", "-z"],
        cwd=root,
        env=env,
    )
    rows: list[tuple[str, str, str]] = []
    for path, mode, oid in _parse_ls_files_z(raw):
        if not path or is_excluded_path(path):
            continue
        if mode not in ALLOWED_MODES:
            continue
        rows.append((path, mode, oid))
    rows.sort(key=lambda item: item[0])
    return rows


def generate_manifest(root: Path) -> str:
    rows = _indexed_file_rows(root)
    lines = [MANIFEST_HEADER]
    if not rows:
        return "\n".join(lines) + "\n"

    env = _git_env()
    blobs = _run_cat_file_batch(root, env, [oid for _path, _mode, oid in rows])
    for (path, _mode, _oid), blob in zip(rows, blobs, strict=True):
        if blob is None:
            continue
        digest = hashlib.sha256(blob).hexdigest()
        lines.append(f"{digest}  {path}")

    return "\n".join(lines) + "\n"


def _parse_inventory(text: str) -> dict[str, str]:
    lines = text.splitlines()
    if not lines or lines[0] != MANIFEST_HEADER:
        raise ValueError(f"bad header: {lines[0] if lines else '<empty>'}")

    expected: dict[str, str] = {}
    for line in lines[1:]:
        if not line or line.startswith("#"):
            continue
        if "  " not in line:
            continue
        digest, path = line.split("  ", 1)
        expected[path] = digest
    return expected


def _file_sha256(path: Path) -> str:
    digest = hashlib.sha256()
    with path.open("rb") as fh:
        for chunk in iter(lambda: fh.read(1024 * 1024), b""):
            digest.update(chunk)
    return digest.hexdigest()


def _tracked_inventory_paths(root: Path) -> list[str]:
    env = _git_env()
    raw = subprocess.check_output(["git", "ls-files"], cwd=root, env=env, text=True)
    paths: list[str] = []
    for path in raw.splitlines():
        if not path or is_excluded_path(path):
            continue
        full = root / path
        if not full.is_file() or full.is_symlink():
            continue
        paths.append(path)
    paths.sort()
    return paths


def verify_manifest(root: Path, inv_text: str, check_tracked: bool) -> tuple[bool, int]:
    expected = _parse_inventory(inv_text)
    env = _git_env()
    paths = sorted(expected.keys())
    blobs = _run_cat_file_batch(root, env, [f":{path}" for path in paths])

    fail = False
    for path, blob in zip(paths, blobs, strict=True):
        want = expected[path]
        full = root / path
        if not full.is_file():
            print(f"tree-manifest.sh: missing: {path}", file=sys.stderr)
            fail = True
            continue
        if blob is not None:
            got = hashlib.sha256(blob).hexdigest()
        else:
            got = _file_sha256(full)
        if got != want:
            print(f"tree-manifest.sh: modified: {path}", file=sys.stderr)
            print(f"  expected {want}", file=sys.stderr)
            print(f"  got      {got}", file=sys.stderr)
            fail = True

    if check_tracked:
        tracked = _tracked_inventory_paths(root)
        inv_paths = sorted(expected.keys())
        tracked_set = set(tracked)
        inv_set = set(inv_paths)
        extra = sorted(inv_set - tracked_set)
        missing = sorted(tracked_set - inv_set)
        if extra:
            print(
                "tree-manifest.sh: inventory has paths not tracked (or excluded):",
                file=sys.stderr,
            )
            for path in extra:
                print(path, file=sys.stderr)
            fail = True
        if missing:
            print(
                "tree-manifest.sh: tracked files missing from inventory (added?):",
                file=sys.stderr,
            )
            for path in missing:
                print(path, file=sys.stderr)
            fail = True

    return not fail, len(expected)


def main() -> int:
    root = _repo_root()
    cmd = sys.argv[1] if len(sys.argv) > 1 else "generate"

    try:
        if cmd == "generate":
            sys.stdout.write(generate_manifest(root))
            return 0
        if cmd in ("verify", "verify-tracked"):
            inv_src = sys.argv[2] if len(sys.argv) > 2 else "-"
            if inv_src == "-" or not inv_src:
                inv_text = sys.stdin.read()
            else:
                inv_text = Path(inv_src).read_text()
            ok, count = verify_manifest(root, inv_text, check_tracked=cmd == "verify-tracked")
            if not ok:
                print("tree-manifest.sh: verification failed", file=sys.stderr)
                return 1
            print(f"tree-manifest.sh: OK ({count} files)")
            return 0
        print(
            f"tree_manifest_generate.py: unknown command: {cmd}",
            file=sys.stderr,
        )
        return 2
    except Exception as exc:
        print(f"tree_manifest_generate.py: {exc}", file=sys.stderr)
        return 1


if __name__ == "__main__":
    raise SystemExit(main())
