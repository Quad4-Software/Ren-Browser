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

  server_opts="${app_dir}/application_options.go"
  if [ -f "${server_opts}" ] && ! grep -q 'OuterMiddleware' "${server_opts}"; then
    sed -i '/TLS \*TLSOptions$/a\\n\t// OuterMiddleware wraps the entire server mux in headless mode, including\n\t// /wails/events and /health. Used by Ren Browser for HTTP auth.\n\tOuterMiddleware func(http.Handler) http.Handler' "${server_opts}"
  fi

  server_app="${app_dir}/application_server.go"
  if [ -f "${server_app}" ] && ! grep -q 'OuterMiddleware' "${server_app}"; then
    sed -i 's/return mux$/handler := http.Handler(mux)\n\tif wrap := h.app.options.Server.OuterMiddleware; wrap != nil {\n\t\thandler = wrap(handler)\n\t}\n\treturn handler/' "${server_app}"
  fi

  android_app="${app_dir}/application_android.go"
  if [ -f "${android_app}" ] && grep -q 'func (a \*androidApp) destroy() {$' "${android_app}"; then
  if ! grep -q 'quitApp' "${android_app}"; then
    sed -i '/func (a \*androidApp) destroy() {/,/^}$/c\
func (a *androidApp) destroy() {\
\tif globalApplication != nil \&\& globalApplication.shouldQuit() {\
\t\tglobalApplication.cleanup()\
\t}\
\tandroidBridgeVoid("quitApp")\
}' "${android_app}"
  fi
  fi

  ios_app="${app_dir}/application_ios.go"
  if [ -f "${ios_app}" ] && grep -q 'Cleanup iOS resources' "${ios_app}"; then
    sed -i '/func (a \*iosApp) destroy() {/,/^}$/c\
func (a *iosApp) destroy() {\
\tif globalApplication != nil \&\& globalApplication.shouldQuit() {\
\t\tglobalApplication.cleanup()\
\t}\
\ta.parent.platformQuit()\
}' "${ios_app}"
  fi

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

  ws_server="${app_dir}/websocket_server.go"
  if [ -f "${ws_server}" ] && ! grep -q 'func (b \*WebSocketBroadcaster) closeAll()' "${ws_server}"; then
    sed -i 's/b\.app\.info("WebSocket client connected"/b.app.debug("WebSocket client connected"/g' "${ws_server}"
    sed -i 's/b\.app\.info("WebSocket client disconnected"/b.app.debug("WebSocket client disconnected"/g' "${ws_server}"
    sed -i '/^\/\/ DispatchWailsEvent implements WailsEventListener interface\./i\
// closeAll closes every active WebSocket so HTTP server shutdown is not blocked.\
func (b *WebSocketBroadcaster) closeAll() {\
\tb.mu.Lock()\
\tconns := make([]*websocket.Conn, 0, len(b.clients))\
\tfor conn := range b.clients {\
\t\tconns = append(conns, conn)\
\t}\
\tb.clients = make(map[*websocket.Conn]*clientInfo)\
\tb.windows = make(map[string]*BrowserWindow)\
\tb.mu.Unlock()\
\
\tfor _, conn := range conns {\
\t\tconn.Close(websocket.StatusGoingAway, "server shutting down")\
\t}\
}\
' "${ws_server}"
  fi

  linux_cgo_c="${app_dir}/linux_cgo.c"
  linux_cgo_h="${app_dir}/linux_cgo.h"
  linux_cgo_go="${app_dir}/linux_cgo.go"
  linux_cgo_gtk3="${app_dir}/linux_cgo_gtk3.go"
  if [ -f "${linux_cgo_c}" ] && ! grep -q 'window_apply_frameless' "${linux_cgo_c}"; then
    sed -i '/^void attach_action_group_to_widget/i\
void window_apply_frameless(GtkWindow *window, gboolean frameless, const char *title) {\
    if (!frameless) {\
        gtk_window_set_titlebar(window, NULL);\
        gtk_window_set_decorated(window, TRUE);\
        if (title != NULL && title[0] != '\''\\0'\'') {\
            gtk_window_set_title(window, title);\
        }\
        gtk_window_present(window);\
        return;\
    }\
    gtk_window_set_decorated(window, FALSE);\
}\
' "${linux_cgo_c}"
  fi
  if [ -f "${linux_cgo_h}" ] && ! grep -q 'window_apply_frameless' "${linux_cgo_h}"; then
    sed -i '/void window_apply_pending_always_on_top/a void window_apply_frameless(GtkWindow *window, gboolean frameless, const char *title);' "${linux_cgo_h}"
  fi
  if [ -f "${linux_cgo_go}" ] && ! grep -q 'window_apply_frameless' "${linux_cgo_go}"; then
    sed -i '/func (w \*linuxWebviewWindow) setFrameless(frameless bool) {/,/^}$/c\
func (w *linuxWebviewWindow) setFrameless(frameless bool) {\
\ttitle := w.parent.options.Title\
\tif title == "" {\
\t\ttitle = w.parent.options.Name\
\t}\
\tcTitle := C.CString(title)\
\tdefer C.free(unsafe.Pointer(cTitle))\
\tC.window_apply_frameless(w.gtkWindow(), gtkBool(frameless), cTitle)\
\tw.execJS(fmt.Sprintf("if(window._wails&&window._wails.flags)window._wails.flags.frameless=%v;", frameless))\
}' "${linux_cgo_go}"
  fi
  if [ -f "${linux_cgo_gtk3}" ] && ! grep -q 'gtk_window_set_titlebar(w.gtkWindow(), nil)' "${linux_cgo_gtk3}"; then
    sed -i '/func (w \*linuxWebviewWindow) setFrameless(frameless bool) {/,/^}$/c\
func (w *linuxWebviewWindow) setFrameless(frameless bool) {\
\tif !frameless {\
\t\tC.gtk_window_set_titlebar(w.gtkWindow(), nil)\
\t}\
\tC.gtk_window_set_decorated(w.gtkWindow(), gtkBool(!frameless))\
\tif !frameless {\
\t\ttitle := w.parent.options.Title\
\t\tif title == "" {\
\t\t\ttitle = w.parent.options.Name\
\t\t}\
\t\tcTitle := C.CString(title)\
\t\tC.gtk_window_set_title(w.gtkWindow(), cTitle)\
\t\tC.free(unsafe.Pointer(cTitle))\
\t\tC.gtk_window_present(w.gtkWindow())\
\t}\
\tw.execJS(fmt.Sprintf("if(window._wails&&window._wails.flags)window._wails.flags.frameless=%v;", frameless))\
}' "${linux_cgo_gtk3}"
  fi

  server_app="${app_dir}/application_server.go"
  if [ -f "${server_app}" ] && ! grep -q 'Force shutdown' "${server_app}"; then
    sed -i '/\/\/ Wait for shutdown signal or error/,/h\.app\.info("Server stopped")/c\
	// Wait for shutdown signal or error\
	quit := make(chan os.Signal, 2)\
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)\
	defer signal.Stop(quit)\
\
	select {\
	case err := <-errCh:\
		if err != nil && !errors.Is(err, http.ErrServerClosed) {\
			return err\
		}\
	case <-quit:\
		h.app.info("Shutdown signal received")\
	case <-h.app.ctx.Done():\
		h.app.info("Application context cancelled")\
	}\
\
	if h.broadcaster != nil {\
		h.broadcaster.closeAll()\
	}\
	if h.listener != nil {\
		_ = h.listener.Close()\
	}\
\
	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)\
	defer cancel()\
\
	shutdownDone := make(chan error, 1)\
	go func() {\
		shutdownDone <- h.server.Shutdown(ctx)\
	}()\
\
	select {\
	case err := <-shutdownDone:\
		if err != nil && !errors.Is(err, http.ErrServerClosed) {\
			return fmt.Errorf("server shutdown error: %w", err)\
		}\
	case <-quit:\
		h.app.info("Force shutdown")\
		_ = h.server.Close()\
		<-shutdownDone\
	}\
\
	h.app.info("Server stopped")' "${server_app}"
    sed -i '/func (h \*serverApp) destroy() {/,/^}$/c\
func (h *serverApp) destroy() {\
\tif h.broadcaster != nil {\
\t\th.broadcaster.closeAll()\
\t}\
\tif h.server != nil {\
\t\tctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)\
\t\tdefer cancel()\
\t\t_ = h.server.Shutdown(ctx)\
\t}\
\th.app.cleanup()\
}' "${server_app}"
  fi
fi
