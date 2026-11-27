#!/usr/bin/env bash
# Fetch the pinned AppImage packaging tools used by package-appimage.sh.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../.." && pwd)"
TOOLSDIR="${ROOT}/build/tools"
LINUXDEPLOY_TAG="${LINUXDEPLOY_TAG:?set LINUXDEPLOY_TAG}"
APPIMAGETOOL_URL="${APPIMAGETOOL_URL:?set APPIMAGETOOL_URL}"
GTK_PLUGIN_COMMIT="${GTK_PLUGIN_COMMIT:-7a3fbc31a9e5075073ff8790f26effbac5f84453}"
GTK_PLUGIN_SHA256="${GTK_PLUGIN_SHA256:-b0f4cbc684a0103a9651f0955b635eaea0096b3a66c0f5a2c2aa337960375171}"
GTK_PLUGIN_PATCHED_SHA256="${GTK_PLUGIN_PATCHED_SHA256:-3ece20d8c6a696dab848fdddca775411a5b0185799696d133dd152d064604f71}"

arch="$(uname -m)"
case "${arch}" in
  x86_64|amd64) suffix=x86_64 ;;
  aarch64|arm64) suffix=aarch64 ;;
  *)
    printf 'Unsupported architecture: %s\n' "${arch}" >&2
    exit 1
    ;;
esac

mkdir -p "${TOOLSDIR}"

fetch() {
  local url="$1"
  local dest="$2"
  if [ ! -f "${dest}" ]; then
    local tmp="${dest}.tmp"
    rm -f "${tmp}"
    curl -fsSL -o "${tmp}" "${url}"
    mv "${tmp}" "${dest}"
  fi
  chmod +x "${dest}"
}

fetch "https://github.com/linuxdeploy/linuxdeploy/releases/download/${LINUXDEPLOY_TAG}/linuxdeploy-${suffix}.AppImage" "${TOOLSDIR}/linuxdeploy-${suffix}.AppImage"
fetch "${APPIMAGETOOL_URL}-${suffix}.AppImage" "${TOOLSDIR}/appimagetool-${suffix}.AppImage"

plugin="${TOOLSDIR}/linuxdeploy-plugin-gtk.sh"
plugin_url="https://raw.githubusercontent.com/linuxdeploy/linuxdeploy-plugin-gtk/${GTK_PLUGIN_COMMIT}/linuxdeploy-plugin-gtk.sh"
if [ ! -f "${plugin}" ] || ! printf '%s  %s\n' "${GTK_PLUGIN_PATCHED_SHA256}" "${plugin}" | sha256sum -c --status - 2>/dev/null; then
  tmp="${plugin}.tmp"
  rm -f "${tmp}"
  curl -fsSL -o "${tmp}" "${plugin_url}"
  printf '%s  %s\n' "${GTK_PLUGIN_SHA256}" "${tmp}" | sha256sum -c --status -
  mv "${tmp}" "${plugin}"

  patched="${plugin}.patched"
  awk '
$0 == "        copy_lib_tree \"$gtk4_libdir\" \"$APPDIR/\"" {
  print "        if [ -d \"$gtk4_libdir\" ]; then"
  print "            copy_lib_tree \"$gtk4_libdir\" \"$APPDIR/\""
  print "        else"
  print "            echo \"GTK 4.0 module directory not present; skipping\""
  print "        fi"
  gtk4++
  next
}
$0 == "copy_lib_tree \"$gdk_pixbuf_binarydir\" \"$APPDIR/\"" {
  print "if [ -d \"$gdk_pixbuf_binarydir\" ]; then"
  print
  pixbuf_open++
  next
}
$0 == "sed -i \"s|$gdk_pixbuf_moduledir/||g\" \"$APPDIR/${gdk_pixbuf_cache_file/$LD_GTK_LIBRARY_PATH//usr/lib}\"" {
  print
  print "else"
  print "    echo \"GDK Pixbuf loader directory not present; skipping\""
  print "fi"
  pixbuf_close++
  next
}
$0 == "for directory in \"${PATCH_ARRAY[@]}\"; do" {
  print
  print "    [ -d \"$directory\" ] || continue"
  patch_array++
  next
}
{ print }
END {
  if (gtk4 != 1 || pixbuf_open != 1 || pixbuf_close != 1 || patch_array != 1) {
    printf "unexpected GTK plugin patch counts: gtk4=%d pixbuf_open=%d pixbuf_close=%d patch_array=%d\n", gtk4, pixbuf_open, pixbuf_close, patch_array > "/dev/stderr"
    exit 1
  }
}
  ' "${plugin}" > "${patched}"
  mv "${patched}" "${plugin}"
  printf '%s  %s\n' "${GTK_PLUGIN_PATCHED_SHA256}" "${plugin}" | sha256sum -c --status -
fi
chmod +x "${plugin}"
bash -n "${plugin}"
printf 'linuxdeploy tools ready in %s\n' "${TOOLSDIR}"
