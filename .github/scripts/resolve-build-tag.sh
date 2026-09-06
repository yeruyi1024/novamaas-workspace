#!/usr/bin/env bash
set -euo pipefail

merged_at=${1:?Pass the PR merge timestamp}
architecture=${2:?Pass the image architecture}

if [[ ! "$merged_at" =~ ^([0-9]{4})-([0-9]{2})-([0-9]{2})T([0-9]{2}):([0-9]{2}):([0-9]{2})Z$ ]]; then
  echo "Invalid UTC merge timestamp: $merged_at" >&2
  exit 1
fi

case "$architecture" in
  amd64 | arm64 | multiarch) ;;
  *)
    echo "Unsupported image architecture: $architecture" >&2
    exit 1
    ;;
esac

timestamp="${BASH_REMATCH[1]}${BASH_REMATCH[2]}${BASH_REMATCH[3]}T${BASH_REMATCH[4]}${BASH_REMATCH[5]}${BASH_REMATCH[6]}Z"
printf 'build_%s_%s\n' "$timestamp" "$architecture"
