#!/usr/bin/env bash
# Build a self-contained AppImage from bin/renbrowser using linuxdeploy + appimagetool.
# appimagetool continuous embeds a fuse3-capable runtime for distros without libfuse2.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../.." && pwd)"
VERSION="${1:?usage: package-appimage.sh <version>}"
ARCH="${ARCH:-$(uname -m)}"
case "${ARCH}" in
  x86_64|amd64) ARCH=x86_64; TOOL_SUFFIX=x86_64 ;;
  aarch64|arm64) ARCH=aarch64; TOOL_SUFFIX=aarch64 ;;
  *)
    printf 'Unsupported architecture: %s\n' "${ARCH}" >&2
    exit 1
    ;;
esac

OUTDIR="${ROOT}/bin"
APPDIR="${ROOT}/build/linux/appimage/AppDir"
TOOLSDIR="${ROOT}/build/tools"
LINUXDEPLOY="${TOOLSDIR}/linuxdeploy-${TOOL_SUFFIX}.AppImage"
GTK_PLUGIN="${TOOLSDIR}/linuxdeploy-plugin-gtk.sh"
APPIMAGETOOL="${TOOLSDIR}/appimagetool-${TOOL_SUFFIX}.AppImage"
BINARY="${ROOT}/bin/renbrowser"
ICON="${ROOT}/build/packaging/renbrowser-256.png"
DESKTOP="${ROOT}/build/linux/appimage/renbrowser.desktop"
OUTPUT="${OUTDIR}/renbrowser-${VERSION}-${ARCH}.AppImage"

for tool in "${LINUXDEPLOY}" "${GTK_PLUGIN}" "${APPIMAGETOOL}"; do
  if [[ ! -f "${tool}" ]]; then
    printf 'Missing %s. Run: task linux:fetch-linuxdeploy\n' "${tool}" >&2
    exit 1
  fi
done
if [[ ! -f "${BINARY}" ]]; then
  printf 'Missing %s. Run: task build first\n' "${BINARY}" >&2
  exit 1
fi
if [[ ! -f "${ICON}" ]]; then
  printf 'Missing %s. Run: task linux:packaging-icon\n' "${ICON}" >&2
  exit 1
fi

mkdir -p "${OUTDIR}" "$(dirname "${APPDIR}")"
rm -rf "${APPDIR}"
export ARCH
export LINUXDEPLOY
export DEPLOY_GTK_VERSION=4
export NO_STRIP=1
export APPIMAGE_EXTRACT_AND_RUN=1
export VERSION

cd "${TOOLSDIR}"
chmod +x "${GTK_PLUGIN}"
"${LINUXDEPLOY}" \
  --appdir "${APPDIR}" \
  -e "${BINARY}" \
  -i "${ICON}" \
  -d "${DESKTOP}" \
  --plugin gtk

"${ROOT}/build/linux/appimage/bundle-webkitgtk.sh" "${APPDIR}"
chmod +x "${ROOT}/build/linux/appimage/strip-gpu-libs.sh"
"${ROOT}/build/linux/appimage/strip-gpu-libs.sh" "${APPDIR}"
chmod +x "${ROOT}/build/linux/appimage/configure-fontconfig.sh"
"${ROOT}/build/linux/appimage/configure-fontconfig.sh" "${APPDIR}"

rm -f "${OUTPUT}"
"${APPIMAGETOOL}" --appimage-extract-and-run -n "${APPDIR}" "${OUTPUT}"
chmod a+x "${OUTPUT}"
printf '%s\n' "${OUTPUT}"
