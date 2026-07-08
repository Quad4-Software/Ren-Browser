module renbrowser

go 1.26.5

require (
	github.com/adrg/xdg v0.5.3
	github.com/tetratelabs/wazero v1.9.0
	github.com/wailsapp/wails/v3 v3.0.0-alpha2.111
	go.uber.org/goleak v1.3.0
	golang.org/x/crypto v0.52.0
	golang.org/x/sys v0.46.0
	golang.org/x/term v0.44.0
	micron-parser-go v0.0.0
	modernc.org/sqlite v1.53.0
	quad4/msgpack/v5 v5.8.0
	quad4/reticulum-go v0.0.0
)

require (
	github.com/coder/websocket v1.8.14 // indirect
	github.com/dustin/go-humanize v1.0.1 // indirect
	github.com/go-ole/go-ole v1.3.0 // indirect
	github.com/godbus/dbus/v5 v5.2.2 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/jchv/go-winloader v0.0.0-20250406163304-c1995be93bd1 // indirect
	github.com/mattn/go-colorable v0.1.14 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/ncruces/go-strftime v1.0.0 // indirect
	github.com/remyoudompheng/bigfft v0.0.0-20230129092748-24d4a6f8daec // indirect
	github.com/wailsapp/wails/webview2 v1.0.27 // indirect
	modernc.org/libc v1.73.4 // indirect
	modernc.org/mathutil v1.7.1 // indirect
	modernc.org/memory v1.11.0 // indirect
	quad4/bzip2 v0.0.0 // indirect
	quad4/tagparser v0.0.0 // indirect
)

replace (
	micron-parser-go => github.com/Quad4-Software/Micron-Parser-Go v1.0.6
	quad4/bzip2 => github.com/Quad4-Software/bzip2 v0.0.0-20260704225916-ca8b2bb66059
	quad4/msgpack/v5 => github.com/Quad4-Software/msgpack/v5 v5.8.0
	quad4/pbt => github.com/Quad4-Software/pbt v0.0.0-20260614183135-abe0cfc4e604
	quad4/reticulum-go => ./third_party/reticulum-go
	quad4/tagparser => github.com/Quad4-Software/tagparser v0.1.3-0.20260614183136-daa4d5f437ce
)
