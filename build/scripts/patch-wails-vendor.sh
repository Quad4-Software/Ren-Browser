#!/usr/bin/env bash
# Re-apply Ren Browser patches to vendored Wails after `go mod vendor`.
set -euo pipefail

root="$(cd "$(dirname "$0")/../.." && pwd)"
patch_dir="${root}/build/patches/wails"
vendor="${root}/vendor/github.com/wailsapp/wails/v3"
extract="${vendor}/pkg/updater/extract.go"

if [ ! -d "${vendor}" ]; then
  exit 0
fi

# install's -D flag (create dst's parent directory) is GNU-only; BSD/macOS
# install has no such flag, so create the parent directory ourselves first.
install_file() {
  mkdir -p "$(dirname "$2")"
  install -m 0644 "$1" "$2"
}

install_file "${patch_dir}/internal/operatingsystem/os_bsd.go" "${vendor}/internal/operatingsystem/os_bsd.go"
install_file "${patch_dir}/internal/assetserver/assetserver_bsd.go" "${vendor}/internal/assetserver/assetserver_bsd.go"
install_file "${patch_dir}/internal/fileexplorer/fileexplorer_bsd.go" "${vendor}/internal/fileexplorer/fileexplorer_bsd.go"

if [ -f "${extract}" ]; then
  sed -i \
    -e 's/exceeds %d bytes", maxArchiveTotalSize)/exceeds %d bytes", int64(maxArchiveTotalSize))/g' \
    "${extract}"
fi

# Several windows-specific files in pkg/application are missing "!server" in
# their build constraint (unlike their linux/darwin counterparts), so they
# redeclare application_server.go's headless stubs (or reference the native
# windowsApp/windowsWebviewWindow types application_server.go excludes)
# whenever GOOS=windows is combined with `-tags server`. Add the same "&&
# !server" exclusion those files already carry on linux/darwin so `task
# build:server GOOS=windows` compiles. The sed patterns below only match the
# exact unpatched line, so re-running this script is a no-op once applied.
app_dir="${vendor}/pkg/application"
if [ -d "${app_dir}" ]; then
  for f in clipboard_windows.go dialogs_windows.go events_common_windows.go \
    mainthread_windows.go single_instance_windows.go systemtray_windows.go \
    webview_window_windows.go screen_windows.go; do
    target="${app_dir}/${f}"
    if [ -f "${target}" ]; then
      sed -i '1s/^\/\/go:build windows$/\/\/go:build windows \&\& !server/' "${target}"
    fi
  done

  prod="${app_dir}/webview_window_windows_production.go"
  if [ -f "${prod}" ]; then
    sed -i '1s/^\/\/go:build windows && production && !devtools$/\/\/go:build windows \&\& production \&\& !devtools \&\& !server/' "${prod}"
  fi

  devtools="${app_dir}/webview_window_windows_devtools.go"
  if [ -f "${devtools}" ]; then
    sed -i '1s/^\/\/go:build windows && (!production || devtools)$/\/\/go:build windows \&\& (!production || devtools) \&\& !server/' "${devtools}"
  fi

  popupmenu="${app_dir}/popupmenu_windows.go"
  if [ -f "${popupmenu}" ] && ! grep -q '^//go:build' "${popupmenu}"; then
    sed -i '1i //go:build windows \&\& !server\n' "${popupmenu}"
  fi
fi
