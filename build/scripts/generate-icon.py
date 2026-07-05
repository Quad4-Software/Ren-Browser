#!/usr/bin/env python3
"""Generate Ren Browser pixel icon SVG and platform raster assets."""

from __future__ import annotations

import shutil
import subprocess
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parents[2]
BUILD = ROOT / "build"
ANDROID_RES = BUILD / "android" / "app" / "src" / "main" / "res"

CANVAS = 128
CORNER_RADIUS = 28
BG = "#09090b"
FG = "#60a5fa"
MONOCHROME = "#FFFFFFFF"

PIXEL_R = [
    "................",
    "................",
    "................",
    "....#######.....",
    "....##....##....",
    "....##....##....",
    "....##....##....",
    "....#######.....",
    "....##.##.......",
    "....##..##......",
    "....##...##.....",
    "....##....##....",
    "....##.....##...",
    "................",
    "................",
    "................",
]

ADAPTIVE_ICON_XML = """<?xml version="1.0" encoding="utf-8"?>
<adaptive-icon xmlns:android="http://schemas.android.com/apk/res/android">
    <background android:drawable="@color/ic_launcher_background" />
    <foreground android:drawable="@drawable/ic_launcher_foreground" />
    <monochrome android:drawable="@drawable/ic_launcher_monochrome" />
</adaptive-icon>
"""


def hex_to_rgb(value: str) -> str:
    value = value.lstrip("#")
    r = int(value[0:2], 16)
    g = int(value[2:4], 16)
    b = int(value[4:6], 16)
    return f"rgb({r} {g} {b})"


def render_svg(
    grid: list[str],
    *,
    rgb: bool = False,
    corner_radius: int | None = None,
    background: bool = True,
) -> str:
    size = len(grid)
    if any(len(row) != size for row in grid):
        raise ValueError("pixel grid must be square")

    cell = CANVAS / size
    bg = hex_to_rgb(BG) if rgb else BG
    fg = hex_to_rgb(FG) if rgb else FG
    radius = CORNER_RADIUS if corner_radius is None else corner_radius

    pixels: list[str] = []
    for y, row in enumerate(grid):
        for x, ch in enumerate(row):
            if ch != "#":
                continue
            px = x * cell
            py = y * cell
            pixels.append(
                f'    <rect x="{px:g}" y="{py:g}" width="{cell:g}" height="{cell:g}"/>'
            )

    pixel_markup = "\n".join(pixels)
    bg_rect = ""
    if background:
        rx_attr = f' rx="{radius}"' if radius > 0 else ""
        bg_rect = f'  <rect width="{CANVAS}" height="{CANVAS}"{rx_attr} fill="{bg}"/>\n'
    return f"""<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 {CANVAS} {CANVAS}" role="img" aria-label="Ren Browser">
{bg_rect}  <g fill="{fg}">
{pixel_markup}
  </g>
</svg>
"""


def render_android_vector(grid: list[str], fill: str, *, dp: int = 108) -> str:
    size = len(grid)
    cell = CANVAS / size
    paths: list[str] = []
    for y, row in enumerate(grid):
        for x, ch in enumerate(row):
            if ch != "#":
                continue
            px = x * cell
            py = y * cell
            paths.append(
                f'    <path android:fillColor="{fill}" '
                f'android:pathData="M{px:g},{py:g}h{cell:g}v{cell:g}h{-cell:g}z"/>'
            )
    body = "\n".join(paths)
    return f"""<?xml version="1.0" encoding="utf-8"?>
<vector xmlns:android="http://schemas.android.com/apk/res/android"
    android:width="{dp}dp"
    android:height="{dp}dp"
    android:viewportWidth="{CANVAS}"
    android:viewportHeight="{CANVAS}">
{body}
</vector>
"""


def write_text(path: Path, content: str) -> None:
    path.parent.mkdir(parents=True, exist_ok=True)
    path.write_text(content, encoding="utf-8")


def run(cmd: list[str], *, cwd: Path | None = None) -> None:
    subprocess.run(cmd, check=True, cwd=cwd)


def rsvg(svg: Path, width: int, height: int, output: Path) -> None:
    output.parent.mkdir(parents=True, exist_ok=True)
    run(["rsvg-convert", "-w", str(width), "-h", str(height), "-o", str(output), str(svg)])


def generate_android_icons(master: Path) -> None:
    legacy = render_svg(PIXEL_R, corner_radius=0)
    legacy_path = BUILD / ".tmp_android_launcher.svg"
    write_text(legacy_path, legacy)

    for px, dpi in (
        (48, "mdpi"),
        (72, "hdpi"),
        (96, "xhdpi"),
        (144, "xxhdpi"),
        (192, "xxxhdpi"),
    ):
        base = ANDROID_RES / f"mipmap-{dpi}"
        rsvg(legacy_path, px, px, base / "ic_launcher.png")
        rsvg(legacy_path, px, px, base / "ic_launcher_round.png")

    write_text(ANDROID_RES / "drawable" / "ic_launcher_foreground.xml", render_android_vector(PIXEL_R, FG))
    write_text(ANDROID_RES / "drawable" / "ic_launcher_monochrome.xml", render_android_vector(PIXEL_R, MONOCHROME))
    write_text(ANDROID_RES / "drawable" / "ic_notification.xml", render_android_vector(PIXEL_R, MONOCHROME, dp=24))

    adaptive_dir = ANDROID_RES / "mipmap-anydpi-v26"
    write_text(adaptive_dir / "ic_launcher.xml", ADAPTIVE_ICON_XML)
    write_text(adaptive_dir / "ic_launcher_round.xml", ADAPTIVE_ICON_XML)

    legacy_path.unlink(missing_ok=True)


def main() -> int:
    master = BUILD / "appicon.svg"
    svg = render_svg(PIXEL_R)
    flatpak_svg = render_svg(PIXEL_R, rgb=True)

    targets = [
        master,
        ROOT / "frontend" / "public" / "favicon.svg",
        BUILD / "appicon.icon" / "Assets" / "ren_icon.svg",
        ROOT / "flatpak" / "io.quad4.renbrowser.svg",
    ]

    for path in targets:
        write_text(path, flatpak_svg if path.name == "io.quad4.renbrowser.svg" else svg)

    if shutil.which("rsvg-convert") is None:
        print("rsvg-convert not found; wrote SVG only", file=sys.stderr)
        return 0

    rsvg(master, 512, 512, BUILD / "appicon.png")
    rsvg(master, 256, 256, BUILD / "packaging" / "renbrowser-256.png")
    rsvg(master, 256, 256, BUILD / "ios" / "icon.png")
    rsvg(
        master,
        256,
        256,
        BUILD / "linux" / "appimage" / "AppDir" / "renbrowser-256.png",
    )
    rsvg(
        master,
        256,
        256,
        BUILD
        / "linux"
        / "appimage"
        / "AppDir"
        / "usr"
        / "share"
        / "icons"
        / "hicolor"
        / "256x256"
        / "apps"
        / "renbrowser-256.png",
    )

    generate_android_icons(master)

    staging = BUILD / "flatpak" / "staging" / "io.quad4.renbrowser.svg"
    write_text(staging, flatpak_svg)
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
