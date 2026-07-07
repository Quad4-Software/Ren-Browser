#!/usr/bin/env bash
# Remove GPU-driver-dependent libraries from an AppDir before packaging.
#
# linuxdeploy bundles Ubuntu-compiled EGL/Mesa/Wayland copies that ABI-mismatch
# with rolling distros (Arch, CachyOS, Fedora). WebKitGPUProcess then fails with
# "Could not create default EGL display: EGL_BAD_PARAMETER".
#
# Deleting these libs forces the host stack to be used via normal linker search.
set -euo pipefail

APPDIR="${1:?usage: strip-gpu-libs.sh <AppDir>}"

if [ ! -d "${APPDIR}/usr" ]; then
  printf 'No usr/ in AppDir: %s\n' "${APPDIR}" >&2
  exit 1
fi

mapfile -t removed < <(
  find "${APPDIR}/usr" -type f \( \
    -name 'libEGL.so*' -o \
    -name 'libEGL_mesa.so*' -o \
    -name 'libGLESv2.so*' -o \
    -name 'libgbm.so*' -o \
    -name 'libGLX.so*' -o \
    -name 'libGLdispatch.so*' -o \
    -name 'libdrm.so*' -o \
    -name 'libwayland-client.so*' -o \
    -name 'libwayland-server.so*' -o \
    -name 'libwayland-egl.so*' -o \
    -name 'libwayland-cursor.so*' \
  \) -print
)

if [ "${#removed[@]}" -eq 0 ]; then
  printf 'No GPU/Wayland libs to strip in %s\n' "${APPDIR}" >&2
  exit 0
fi

for lib in "${removed[@]}"; do
  rm -f "$lib"
  printf 'Stripped %s\n' "${lib#"${APPDIR}/"}"
done

HOOK_FILE="${APPDIR}/apprun-hooks/linuxdeploy-plugin-gtk.sh"
if [ -f "${HOOK_FILE}" ] && ! grep -q 'renbrowser-appimage-gpu-workaround' "${HOOK_FILE}"; then
  cat >>"${HOOK_FILE}" <<'EOF'

# renbrowser-appimage-gpu-workaround: host EGL/Mesa/Wayland (see strip-gpu-libs.sh).
if [ -z "${WEBKIT_DISABLE_DMABUF_RENDERER+x}" ]; then
  export WEBKIT_DISABLE_DMABUF_RENDERER=1
fi
case "$(uname -m)" in
  x86_64|amd64) _rb_wl_dirs="/usr/lib/x86_64-linux-gnu /usr/lib64 /usr/lib" ;;
  aarch64|arm64) _rb_wl_dirs="/usr/lib/aarch64-linux-gnu /usr/lib64 /usr/lib" ;;
  *) _rb_wl_dirs="/usr/lib" ;;
esac
if [ -z "${LD_PRELOAD:-}" ]; then
  for _rb_dir in ${_rb_wl_dirs}; do
    if [ -e "${_rb_dir}/libwayland-client.so.0" ]; then
      export LD_PRELOAD="${_rb_dir}/libwayland-client.so.0"
      break
    fi
  done
fi
unset _rb_dir _rb_wl_dirs
EOF
  printf 'Injected AppImage GPU workaround into %s\n' "${HOOK_FILE#"${APPDIR}/"}"
fi
