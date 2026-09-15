module renbrowser

go 1.27.1

require (
	github.com/adrg/xdg v0.5.3
	github.com/tetratelabs/wazero v1.12.0
	github.com/wailsapp/wails/v3 v3.0.0-beta.20
	go.uber.org/goleak v1.3.0
	golang.org/x/crypto v0.57.0
	golang.org/x/sys v0.48.0
	golang.org/x/term v0.46.0
	micron-parser-go v0.0.0
	modernc.org/sqlite v1.58.0
	quad4/msgpack/v5 v5.8.2
	quad4/reticulum-go v0.0.0
)

require (
	github.com/coder/websocket v1.8.15 // indirect
	github.com/dunglas/httpsfv v1.1.1 // indirect
	github.com/dustin/go-humanize v1.0.1 // indirect
	github.com/go-ole/go-ole v1.3.0 // indirect
	github.com/godbus/dbus/v5 v5.2.2 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/landlock-lsm/go-landlock v0.10.0 // indirect
	github.com/mattn/go-colorable v0.1.15 // indirect
	github.com/mattn/go-isatty v0.0.24 // indirect
	github.com/mdlayher/socket v0.6.1 // indirect
	github.com/mdlayher/vsock v1.3.0 // indirect
	github.com/ncruces/go-strftime v1.0.0 // indirect
	github.com/quic-go/qpack v0.6.0 // indirect
	github.com/quic-go/quic-go v0.60.0 // indirect
	github.com/quic-go/webtransport-go v0.11.1 // indirect
	github.com/remyoudompheng/bigfft v0.0.0-20230129092748-24d4a6f8daec // indirect
	go.bug.st/serial v1.8.0 // indirect
	golang.org/x/net v0.59.0 // indirect
	golang.org/x/sync v0.23.0 // indirect
	golang.org/x/text v0.42.0 // indirect
	kernel.org/pub/linux/libs/security/libcap/psx v1.2.78 // indirect
	modernc.org/libc v1.75.7 // indirect
	modernc.org/mathutil v1.7.1 // indirect
	modernc.org/memory v1.12.1 // indirect
	quad4/bzip2 v0.0.0 // indirect
	quad4/tagparser v0.0.0 // indirect
)

replace (
	micron-parser-go => github.com/Quad4-Software/Micron-Parser-Go v1.2.0
	quad4/bzip2 => github.com/Quad4-Software/bzip2 v0.0.0-20260904205503-6d5ad4dc26c1
	quad4/msgpack/v5 => github.com/Quad4-Software/msgpack/v5 v5.8.2
	quad4/pbt => github.com/Quad4-Software/pbt v0.0.0-20260904205509-622550407798
	quad4/reticulum-go => ./third_party/reticulum-go
	quad4/tagparser => github.com/Quad4-Software/tagparser v0.1.3-0.20260614183136-daa4d5f437ce
)
