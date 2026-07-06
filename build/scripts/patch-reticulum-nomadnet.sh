#!/usr/bin/env bash
# Patches the vendored reticulum-go link package so that:
#   - Link.Request() accepts arbitrary msgpack-able request data (e.g. a
#     map[string]string) instead of forcing callers to pre-encode request
#     parameters as a []byte blob. Nomad Network nodes only read var_*/field_*
#     request parameters when the wire payload unpacks as a dict; shipping
#     bytes there silently drops every request parameter (form fields, link
#     vars) sent from Ren Browser.
#   - Resource-transfer responses that carry metadata (as Nomad Network does
#     for /file/ downloads, attaching {"name": <filename>}) are split into
#     their metadata and payload instead of being treated as a corrupt
#     [request_id, response] envelope, which silently prepended the metadata
#     bytes onto every downloaded file.
#
# See build/patches/reticulum-go/pkg/link for the patched sources and
# quad4/reticulum-go's pkg/link/nomadnet_response_test.go for regression tests.
set -euo pipefail

# install_file copies src to dst, creating dst's parent directory first.
# GNU install's -D flag does this in one step, but BSD/macOS install has no
# -D flag (it uses -D for something else entirely), so do it in two portable
# steps instead.
install_file() {
  mkdir -p "$(dirname "$2")"
  install -m 0644 "$1" "$2"
}

root="$(cd "$(dirname "$0")/../.." && pwd)"
vendor_dir="${root}/third_party/reticulum-go"
patch_dir="${root}/build/patches/reticulum-go"
link_dir="${vendor_dir}/pkg/link"
gomod_vendor_link_dir="${root}/vendor/quad4/reticulum-go/pkg/link"

bash "${root}/build/scripts/fetch-reticulum-go.sh"

if [[ ! -d "${link_dir}" ]]; then
  echo "patch-reticulum-nomadnet: ${link_dir} not found" >&2
  exit 1
fi

if ! grep -q 'func splitResourceMetadata' "${link_dir}/incoming_resource.go" 2>/dev/null; then
  chmod -R u+w "${link_dir}"
  install_file "${patch_dir}/pkg/link/link.go" "${link_dir}/link.go"
  install_file "${patch_dir}/pkg/link/incoming_resource.go" "${link_dir}/incoming_resource.go"
  install_file "${patch_dir}/pkg/link/nomadnet_response_test.go" "${link_dir}/nomadnet_response_test.go"

  if ! grep -q 'func splitResourceMetadata' "${link_dir}/incoming_resource.go"; then
    echo "patch-reticulum-nomadnet: failed to patch ${link_dir}/incoming_resource.go" >&2
    exit 1
  fi
fi

# go.mod's replace directive points quad4/reticulum-go at third_party/reticulum-go,
# so `go mod vendor` is what normally copies these fixes into vendor/. That step is
# only run by the vendor:go task, not by go:mod:tidy/ci-prep-go, so also patch the
# committed vendor/ copy directly here to keep -mod=vendor builds (the default,
# since vendor/modules.txt is present) from silently reverting to the buggy behavior.
if [[ -d "${gomod_vendor_link_dir}" ]] && ! grep -q 'func splitResourceMetadata' "${gomod_vendor_link_dir}/incoming_resource.go" 2>/dev/null; then
  chmod -R u+w "${gomod_vendor_link_dir}"
  install_file "${patch_dir}/pkg/link/link.go" "${gomod_vendor_link_dir}/link.go"
  install_file "${patch_dir}/pkg/link/incoming_resource.go" "${gomod_vendor_link_dir}/incoming_resource.go"

  if ! grep -q 'func splitResourceMetadata' "${gomod_vendor_link_dir}/incoming_resource.go"; then
    echo "patch-reticulum-nomadnet: failed to patch ${gomod_vendor_link_dir}/incoming_resource.go" >&2
    exit 1
  fi
fi
