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

# WebKit's sandbox (BubblewrapLauncher) shells out to bwrap and xdg-dbus-proxy
# using the same hard-coded /usr paths patched above, so they must exist at
# $APPDIR/usr/bin for the sandbox to work on any target machine.
#
# These run as separate exec'd processes, not libraries loaded into our own
# process, so they don't need to be "relocatable" copies at all - they just
# need to be found at that path. Bundling real copies pulled a huge, fragile,
# build-host-specific dependency graph (glibc, libsystemd, libmount, ...) that
# WebKit spawns with a sanitized environment where nothing outside an
# explicit RPATH resolves, not even libc; matching that correctly for every
# target distro/kernel is impractical. Instead, install thin wrapper scripts
# that exec the target system's own bwrap/xdg-dbus-proxy via a normal PATH
# lookup, so each helper always runs with its native distro's own toolchain.
install_sandbox_helper_wrapper() {
  local name="$1"
  cat > "${APPDIR}/usr/bin/${name}" <<EOF
#!/bin/sh
# Thin wrapper: WebKit's BubblewrapLauncher expects to find this helper
# relative to the (relocated) AppImage install dir; exec the host's real
# ${name} so it runs with its native distro's own libraries. Deliberately
# does not use "command -v"/\$PATH: AppRun.wrapped puts this very directory
# on PATH, and a self-referential lookup here would recurse into itself.
for dir in /usr/bin /bin /usr/local/bin; do
  if [ -x "\${dir}/${name}" ]; then
    exec "\${dir}/${name}" "\$@"
  fi
done
printf '${name} not found on this system; install the "bubblewrap" package (and "xdg-dbus-proxy") for WebKit sandboxing\n' >&2
exit 127
EOF
  chmod 755 "${APPDIR}/usr/bin/${name}"
  printf 'Installed %s wrapper at %s\n' "$name" "${APPDIR}/usr/bin/${name}"
}

install_sandbox_helper_wrapper bwrap
install_sandbox_helper_wrapper xdg-dbus-proxy

# linuxdeploy's bundled patchelf (0.15.0) corrupts libleancrypto.so.1's ELF
# layout while rewriting RUNPATH: it relocates .dynstr away from the segment
# holding .dynsym/.gnu.hash, which then makes glibc's ld.so segfault inside
# its GNU_HASH symbol lookup the moment the library is loaded (reproduced
# reliably on Arch/CachyOS, where libgnutls depends on libleancrypto for
# post-quantum KEM support). Confirmed via bisection that no other bundled
# library exhibits this failure. RUNPATH patching is redundant for us anyway:
# AppRun.wrapped already exports LD_LIBRARY_PATH=$APPDIR/usr/lib for every
# child process, so repair by restoring the unpatched build-host copy.
mapfile -t leancrypto_libs < <(find "$APPDIR"/usr/lib* -name 'libleancrypto.so*' -type f 2>/dev/null || true)
if [ "${#leancrypto_libs[@]}" -gt 0 ]; then
  for lib in "${leancrypto_libs[@]}"; do
    base="$(basename "$lib")"
    src=""
    for dir in /usr/lib/x86_64-linux-gnu /usr/lib64 /usr/lib /lib/x86_64-linux-gnu /lib64 /lib; do
      if [ -f "${dir}/${base}" ]; then
        src="${dir}/${base}"
        break
      fi
    done
    if [ -z "$src" ]; then
      printf 'Warning: %s is patchelf-corrupted and no build-host copy was found to repair it\n' "$base" >&2
      continue
    fi
    printf 'Repairing patchelf-corrupted library: %s\n' "$base"
    install -Dm755 "$src" "$lib"
  done
fi
