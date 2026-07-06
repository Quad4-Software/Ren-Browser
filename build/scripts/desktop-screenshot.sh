#!/usr/bin/env bash
set -euo pipefail

if [[ $# -lt 1 ]]; then
  printf 'usage: desktop-screenshot.sh OUTPUT_DIR [THEME]\n' >&2
  exit 2
fi

out_dir=$1
theme=${2:-dark}
mkdir -p "${out_dir}/${theme}"

title=${REN_BROWSER_WINDOW_TITLE:-Ren Browser}
file="${out_dir}/${theme}/home.png"

find_window() {
  if command -v xdotool >/dev/null 2>&1; then
    xdotool search --name "${title}" 2>/dev/null | head -n1
    return
  fi
  if command -v wmctrl >/dev/null 2>&1; then
    wmctrl -l | grep -F "${title}" | awk '{print $1}' | head -n1
    return
  fi
  printf ''
}

wid=$(find_window)
if [[ -z "${wid}" ]]; then
  printf 'desktop screenshot: window %q not found (install xdotool or wmctrl)\n' "${title}" >&2
  exit 1
fi

if command -v grim >/dev/null 2>&1 && command -v slurp >/dev/null 2>&1; then
  geom=$(xdotool getwindowgeometry --shell "${wid}" 2>/dev/null || true)
  # shellcheck disable=SC1090
  eval "${geom}"
  grim -g "${WIDTH},${HEIGHT}+${X},${Y}" "${file}"
elif command -v import >/dev/null 2>&1; then
  import -window "${wid}" "${file}"
elif command -v scrot >/dev/null 2>&1; then
  scrot -u "${file}"
else
  printf 'desktop screenshot: install grim, ImageMagick import, or scrot\n' >&2
  exit 1
fi

printf '%s\n' "${file}"
