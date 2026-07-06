#!/usr/bin/env bash
# Create a new GitHub release with assets. Never updates or replaces assets on an
# existing release (required when repository release immutability is enabled).
set -euo pipefail

channel="${1:?channel required (nightly|beta|release)}"
tag="${2:?tag required}"
title="${3:?title required}"
prerelease="${4:?prerelease required (true|false)}"
make_latest="${5:?make_latest required (true|false)}"
notes_file="${6:?notes_file required}"
release_dir="${7:?release_dir required}"

if ! command -v gh >/dev/null 2>&1; then
  echo "gh CLI is required to publish releases." >&2
  exit 1
fi

if [ ! -d "$release_dir" ] || [ -z "$(find "$release_dir" -maxdepth 1 -type f | head -n 1)" ]; then
  echo "No release files found in ${release_dir}." >&2
  exit 1
fi

if gh release view "$tag" >/dev/null 2>&1; then
  case "$channel" in
    release)
      echo "Release ${tag} already exists; refusing to modify an immutable release." >&2
      exit 1
      ;;
    *)
      tag="${tag}-r${GITHUB_RUN_ID}"
      if gh release view "$tag" >/dev/null 2>&1; then
        echo "Release ${tag} already exists." >&2
        exit 1
      fi
      ;;
  esac
fi

args=(
  release create "$tag"
  --target "${GITHUB_SHA}"
  --title "$title"
  --notes-file "$notes_file"
)

if [ "$prerelease" = "true" ]; then
  args+=(--prerelease)
fi
if [ "$make_latest" = "true" ]; then
  args+=(--latest)
fi

while IFS= read -r -d '' file; do
  args+=("$file")
done < <(find "$release_dir" -maxdepth 1 -type f -print0)

gh "${args[@]}"
printf 'Published immutable release %s\n' "$tag"
