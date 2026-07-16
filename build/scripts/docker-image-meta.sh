#!/usr/bin/env bash
# Emit Docker image tags and OCI labels for GHCR from local CI context only.
# No GitHub API calls (avoids docker/metadata-action and unicorn/500 outages).
#
# Required env:
#   IMAGE           full image name without tag (e.g. ghcr.io/owner/renbrowser)
#   IMAGE_VERSION   product version from build/config.yml
#   GITHUB_REF      refs/heads/... or refs/tags/...
#   GITHUB_SHA      full commit sha
# Optional env:
#   GITHUB_EVENT_NAME
#   GITHUB_REF_NAME
#   DEFAULT_BRANCH  repository default branch (from event payload)
#   GITHUB_REPOSITORY  owner/repo for source URL
#   GITHUB_OUTPUT   Actions output file (if unset, prints to stdout)
set -euo pipefail

image="${IMAGE:?IMAGE is required}"
version="${IMAGE_VERSION:?IMAGE_VERSION is required}"
ref="${GITHUB_REF:-}"
sha="${GITHUB_SHA:-}"
event="${GITHUB_EVENT_NAME:-}"
ref_name="${GITHUB_REF_NAME:-}"
default_branch="${DEFAULT_BRANCH:-master}"
repo="${GITHUB_REPOSITORY:-}"

if [[ -z "${ref}" || -z "${sha}" ]]; then
  echo "GITHUB_REF and GITHUB_SHA are required" >&2
  exit 1
fi

if [[ -z "${ref_name}" ]]; then
  ref_name="${ref##*/}"
fi

short_sha="$(printf '%s' "${sha}" | cut -c1-7)"
created="$(date -u +'%Y-%m-%dT%H:%M:%SZ')"

tags=()
add_tag() {
  local t="$1"
  [[ -n "${t}" ]] || return 0
  tags+=("${image}:${t}")
}

sanitize_ref_tag() {
  # Docker tag rules: lowercase, allowed [a-z0-9._-]
  printf '%s' "$1" | tr '[:upper:]' '[:lower:]' | sed -E 's/[^a-z0-9._-]+/-/g; s/^-+//; s/-+$//'
}

is_default_branch=0
if [[ "${ref}" == "refs/heads/${default_branch}" ]]; then
  is_default_branch=1
fi

case "${ref}" in
  refs/heads/*)
    add_tag "$(sanitize_ref_tag "${ref_name}")"
    ;;
  refs/tags/*)
    add_tag "$(sanitize_ref_tag "${ref_name}")"
    if [[ "${ref_name}" =~ ^v([0-9]+)\.([0-9]+)\.([0-9]+)$ ]]; then
      major="${BASH_REMATCH[1]}"
      minor="${BASH_REMATCH[2]}"
      patch="${BASH_REMATCH[3]}"
      add_tag "${major}.${minor}.${patch}"
      add_tag "${major}.${minor}"
      add_tag "${major}"
    fi
    ;;
esac

add_tag "sha-${short_sha}"

if [[ "${is_default_branch}" -eq 1 ]]; then
  add_tag "latest"
fi

if [[ "${event}" == "schedule" || "${ref_name}" == nightly-* ]]; then
  add_tag "nightly"
fi

if [[ "${ref}" == "refs/heads/beta" ]]; then
  add_tag "beta"
fi

# Deduplicate while preserving order
tags_out=""
declare -A seen=()
for t in "${tags[@]}"; do
  if [[ -n "${seen[$t]+x}" ]]; then
    continue
  fi
  seen[$t]=1
  tags_out+="${t}"$'\n'
done
tags_out="${tags_out%$'\n'}"

source_url=""
if [[ -n "${repo}" ]]; then
  source_url="https://github.com/${repo}"
fi

title="Ren Browser"
description="Reticulum browser for NomadNet pages"

labels=""
append_label() {
  labels+="$1=$2"$'\n'
}
append_label "org.opencontainers.image.title" "${title}"
append_label "org.opencontainers.image.description" "${description}"
append_label "org.opencontainers.image.version" "${version}"
append_label "org.opencontainers.image.revision" "${sha}"
append_label "org.opencontainers.image.created" "${created}"
if [[ -n "${source_url}" ]]; then
  append_label "org.opencontainers.image.source" "${source_url}"
  append_label "org.opencontainers.image.url" "${source_url}"
fi
labels="${labels%$'\n'}"

if [[ -n "${GITHUB_OUTPUT:-}" ]]; then
  {
    echo "tags<<EOF"
    echo "${tags_out}"
    echo "EOF"
    echo "labels<<EOF"
    echo "${labels}"
    echo "EOF"
  } >> "${GITHUB_OUTPUT}"
else
  echo "tags:"
  echo "${tags_out}"
  echo "labels:"
  echo "${labels}"
fi
