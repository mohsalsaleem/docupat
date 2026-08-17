#!/usr/bin/env bash
set -euo pipefail

threshold="${COVERAGE_THRESHOLD:-80}"
profile="$(mktemp -t docpatch-coverage.XXXXXX)"
cache="${GOCACHE:-$(mktemp -d -t docpatch-go-cache.XXXXXX)}"

cleanup() {
  rm -f "$profile"
  if [[ -z "${GOCACHE:-}" ]]; then
    rm -rf "$cache"
  fi
}
trap cleanup EXIT

GOCACHE="$cache" go test -coverpkg=./internal/... -coverprofile="$profile" ./internal/...
coverage="$(GOCACHE="$cache" go tool cover -func="$profile" | awk '/^total:/ { gsub("%", "", $3); print $3 }')"

echo "Internal backend statement coverage: ${coverage}% (minimum ${threshold}%)"
awk -v actual="$coverage" -v minimum="$threshold" 'BEGIN { exit !(actual + 0 >= minimum + 0) }' || {
  echo "Coverage gate failed: ${coverage}% is below ${threshold}%" >&2
  exit 1
}
