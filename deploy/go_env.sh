#!/usr/bin/env bash

set -euo pipefail

GO_CACHE_ROOT="${GO_CACHE_ROOT:-/tmp/recru-net-go}"

export GOCACHE="${GOCACHE:-${GO_CACHE_ROOT}/build}"
export GOMODCACHE="${GOMODCACHE:-${GO_CACHE_ROOT}/mod}"
export GOTMPDIR="${GOTMPDIR:-${GO_CACHE_ROOT}/tmp}"

mkdir -p "${GOCACHE}" "${GOMODCACHE}" "${GOTMPDIR}"

cat <<EOF
Configured Go environment for this shell:
  GOCACHE=${GOCACHE}
  GOMODCACHE=${GOMODCACHE}
  GOTMPDIR=${GOTMPDIR}

Usage:
  source deploy/go_env.sh
EOF
