// Nested module boundary so toolchains under build/tools (go-mte, AppImage
// helpers, etc.) are never part of the renbrowser module's go test ./... or
// go mod tidy walk.
module renbrowser.tools

go 1.26
