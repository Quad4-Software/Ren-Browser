#!/usr/bin/env bash
# Install GTK4/WebKitGTK 6.0 build deps for Linux desktop CI jobs.
# apt-get is wrapped in a wall-clock timeout so a stalled Ubuntu mirror
# cannot hold a job until the 6h runner cap.
set -euo pipefail

export DEBIAN_FRONTEND=noninteractive

apt_opts=(
  -o Acquire::Retries=5
  -o Acquire::http::Timeout=30
  -o Acquire::https::Timeout=30
  -o Dpkg::Use-Pty=0
)

packages=(
  pkg-config gcc libc6-dev
  libglib2.0-dev
  libgtk-4-dev libwebkitgtk-6.0-dev
  bubblewrap xdg-dbus-proxy
  curl librsvg2-bin patchelf
)

attempt=1
max_attempts=3
while [ "${attempt}" -le "${max_attempts}" ]; do
  echo "ci-install-linux-deps: attempt ${attempt}/${max_attempts}"
  if sudo timeout 8m apt-get "${apt_opts[@]}" update \
    && sudo timeout 10m apt-get "${apt_opts[@]}" install -y --no-install-recommends "${packages[@]}"; then
    pkg-config --modversion gtk4
    pkg-config --modversion webkitgtk-6.0
    exit 0
  fi
  echo "ci-install-linux-deps: attempt ${attempt} failed" >&2
  attempt=$((attempt + 1))
  if [ "${attempt}" -le "${max_attempts}" ]; then
    sleep 15
  fi
done

echo "ci-install-linux-deps: failed after ${max_attempts} attempts" >&2
exit 1
