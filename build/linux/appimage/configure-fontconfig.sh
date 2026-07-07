#!/usr/bin/env bash
# Point bundled fontconfig at a minimal AppDir config so it does not parse the
# host's /etc/fonts XML (often newer than the linuxdeploy-bundled libfontconfig).
set -euo pipefail

APPDIR="${1:?usage: configure-fontconfig.sh <AppDir>}"

if [ ! -d "${APPDIR}/usr" ]; then
  printf 'No usr/ in AppDir: %s\n' "${APPDIR}" >&2
  exit 1
fi

if ! find "${APPDIR}/usr/lib" -maxdepth 1 -name 'libfontconfig.so*' -type f 2>/dev/null | grep -q .; then
  printf 'No bundled libfontconfig in AppDir; skipping fontconfig setup\n' >&2
  exit 0
fi

CONF_DIR="${APPDIR}/usr/etc/fonts"
mkdir -p "${CONF_DIR}"

DTD_SRC="/usr/share/xml/fontconfig/fonts.dtd"
if [ -f "${DTD_SRC}" ]; then
  mkdir -p "${APPDIR}/usr/share/xml/fontconfig"
  cp "${DTD_SRC}" "${APPDIR}/usr/share/xml/fontconfig/fonts.dtd"
fi

cat >"${CONF_DIR}/fonts.conf" <<'EOF'
<?xml version="1.0"?>
<!DOCTYPE fontconfig SYSTEM "fonts.dtd">
<fontconfig>
  <dir>/usr/share/fonts</dir>
  <dir>/usr/local/share/fonts</dir>
  <dir prefix="xdg">fonts</dir>
  <cachedir prefix="xdg">fontconfig</cachedir>
</fontconfig>
EOF

HOOK_FILE="${APPDIR}/apprun-hooks/linuxdeploy-plugin-gtk.sh"
if [ -f "${HOOK_FILE}" ] && ! grep -q 'renbrowser-appimage-fontconfig' "${HOOK_FILE}"; then
  cat >>"${HOOK_FILE}" <<'EOF'

# renbrowser-appimage-fontconfig: avoid parsing host /etc/fonts with bundled libfontconfig.
export FONTCONFIG_FILE="${APPDIR}/usr/etc/fonts/fonts.conf"
export FONTCONFIG_PATH="${APPDIR}/usr/etc/fonts"
if [ -z "${JSC_SIGNAL_FOR_GC:-}" ]; then
  export JSC_SIGNAL_FOR_GC=36
fi
EOF
  printf 'Injected AppImage fontconfig workaround into %s\n' "${HOOK_FILE#"${APPDIR}/"}"
fi

printf 'Installed AppImage fontconfig at %s\n' "${CONF_DIR#"${APPDIR}/"}"
