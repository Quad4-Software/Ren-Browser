module renbrowser

go 1.27.4

require (
	git.quad4.io/libraries/msgpack/v5 v5.9.3
	git.quad4.io/reticulum/Reticulum-Go v1.6.1-0.20261126222014-01045d6d96ba
	github.com/adrg/xdg v0.5.3
	github.com/tetratelabs/wazero v1.12.0
	github.com/wailsapp/wails/v3 v3.1.0
	go.uber.org/goleak v1.3.0
	golang.org/x/crypto v0.59.0
	golang.org/x/net v0.63.0
	golang.org/x/sys v0.50.0
	golang.org/x/term v0.48.0
	micron-parser-go v1.3.0
	modernc.org/sqlite v1.63.0
)

require (
	git.quad4.io/libraries/bzip2 v1.0.2 // indirect
	git.quad4.io/libraries/tagparser/v2 v2.2.2 // indirect
	github.com/coder/websocket v1.8.15 // indirect
	github.com/dunglas/httpsfv v1.1.2 // indirect
	github.com/dustin/go-humanize v1.1.0 // indirect
	github.com/ebitengine/purego v0.11.3 // indirect
	github.com/go-ole/go-ole v1.3.0 // indirect
	github.com/godbus/dbus/v5 v5.2.4 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/landlock-lsm/go-landlock v0.11.2 // indirect
	github.com/mattn/go-isatty v0.0.24 // indirect
	github.com/mdlayher/socket v0.8.0 // indirect
	github.com/mdlayher/vsock v1.4.0 // indirect
	github.com/ncruces/go-strftime v1.1.0 // indirect
	github.com/quic-go/qpack v0.6.0 // indirect
	github.com/quic-go/quic-go v0.66.0 // indirect
	github.com/quic-go/webtransport-go v0.13.0 // indirect
	github.com/remyoudompheng/bigfft v0.0.0-20230129092748-24d4a6f8daec // indirect
	go.bug.st/serial v1.8.0 // indirect
	golang.org/x/sync v0.25.0 // indirect
	golang.org/x/text v0.44.0 // indirect
	kernel.org/pub/linux/libs/security/libcap/psx v1.2.78 // indirect
	modernc.org/libc v1.82.14 // indirect
	modernc.org/mathutil v1.7.1 // indirect
	modernc.org/memory v1.13.1 // indirect
)

replace micron-parser-go => github.com/Quad4-Software/Micron-Parser-Go v1.3.0

replace git.quad4.io/reticulum/Reticulum-Go => ./third_party/reticulum-go
