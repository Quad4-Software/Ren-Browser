#!/usr/bin/env bash
# Sign a RenBrowser extension package with Reticulum rnid (.rsg embedded).
set -euo pipefail

root="$(cd "$(dirname "$0")/../.." && pwd)"
identity=""
target=""
kind="auto"
output=""

usage() {
  cat <<EOF
Usage: $(basename "$0") -i IDENTITY PATH

Sign an extension directory, zip, or wasm bundle using rnid and embed the
resulting .rsg signature.

Options:
  -i PATH   Reticulum identity file (required)
  -o PATH   Output path for wasm signing (optional, in-place by default)
  -h        Show help

Examples:
  $(basename "$0") -i ~/.renbrowser/publisher extensions/micron-translator
  $(basename "$0") -i ~/.renbrowser/publisher dist/release/renbrowser-micron-translator.wasm
EOF
}

while getopts ":i:o:h" opt; do
  case "$opt" in
    i) identity="$OPTARG" ;;
    o) output="$OPTARG" ;;
    h)
      usage
      exit 0
      ;;
    *)
      usage >&2
      exit 2
      ;;
  esac
done
shift $((OPTIND - 1))

target="${1:-}"
if [[ -z "$identity" || -z "$target" ]]; then
  usage >&2
  exit 2
fi

if ! command -v rnid >/dev/null 2>&1; then
  echo "$(basename "$0"): rnid not found; install with: pip install rns" >&2
  exit 1
fi

args=(sign -identity "$identity")
if [[ -d "$target" ]]; then
  args+=(-dir "$target")
elif [[ -f "$target" && "$target" == *.wasm ]]; then
  args+=(-wasm "$target")
  if [[ -n "$output" ]]; then
    args+=(-output "$output")
  fi
elif [[ -f "$target" && "$target" == *.zip ]]; then
  args+=(-zip "$target")
else
  echo "$(basename "$0"): unsupported target: $target" >&2
  exit 2
fi

cd "$root"
go run ./cmd/pluginsign "${args[@]}"

verify_args=(verify)
if [[ -d "$target" ]]; then
  verify_args+=(-dir "$target")
elif [[ -f "$target" && "$target" == *.wasm ]]; then
  verify_args+=(-wasm "$target")
else
  verify_args+=(-zip "$target")
fi
go run ./cmd/pluginsign "${verify_args[@]}"
