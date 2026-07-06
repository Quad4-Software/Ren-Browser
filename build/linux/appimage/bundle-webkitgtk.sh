#!/usr/bin/env bash
# Bundle WebKitGTK 6 helper processes and patch libwebkit*.so for relocatable AppImages.
set -euo pipefail

APPDIR="${1:?usage: bundle-webkitgtk.sh <AppDir>}"

mapfile -t webkit_libs < <(find "$APPDIR"/usr/lib* -name 'libwebkit*.so*' -type f 2>/dev/null || true)
if [ "${#webkit_libs[@]}" -eq 0 ]; then
  printf 'No libwebkit libraries in AppDir; skipping WebKit helper bundling\n' >&2
  exit 0
fi

arch="$(uname -m)"
WEBKIT_DIRS=(
  "/usr/lib/${arch}-linux-gnu/webkitgtk-6.0"
  "/usr/lib/webkitgtk-6.0"
  "/usr/lib64/webkitgtk-6.0"
  "/usr/libexec/webkitgtk-6.0"
)

webkit_dir=""
for dir in "${WEBKIT_DIRS[@]}"; do
  if [ -x "${dir}/WebKitNetworkProcess" ]; then
    webkit_dir="$dir"
    break
  fi
done

if [ -z "$webkit_dir" ]; then
  printf 'webkitgtk-6.0 helpers not found on build host\n' >&2
  exit 1
fi

install_helpers() {
  local dest="$1"
  mkdir -p "$dest"
  cp -a "${webkit_dir}/." "$dest/"
}

printf 'Bundling WebKitGTK helpers from %s\n' "$webkit_dir"

# Debian/Ubuntu WebKit builds hard-code the multiarch helper directory.
install_helpers "${APPDIR}/usr/lib/${arch}-linux-gnu/webkitgtk-6.0"

# Arch and some other distros install helpers next to /usr/lib/webkitgtk-6.0.
if [ "${webkit_dir}" != "${APPDIR}/usr/lib/webkitgtk-6.0" ]; then
  install_helpers "${APPDIR}/usr/lib/webkitgtk-6.0"
fi

printf 'Patching libwebkit paths for AppImage relocation\n'
find "$APPDIR"/usr/lib* -name 'libwebkit*.so*' -type f -exec sed -i -e 's|/usr|././|g' '{}' +
