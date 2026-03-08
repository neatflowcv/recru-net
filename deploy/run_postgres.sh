#!/usr/bin/env bash

set -euo pipefail

PG_NAME="recur-db"
PG_DB="recur"
PG_USER="recur"
PG_PASSWORD="recur"
PG_VOLUME="recur"
PG_PORT="5432"

podman run -d \
--name="${PG_NAME:?}" \
-e "POSTGRES_DB=${PG_DB:?}" \
-e "POSTGRES_USER=${PG_USER:?}" \
-e "POSTGRES_PASSWORD=${PG_PASSWORD:?}" \
-p "${PG_PORT:?}:5432" \
-v "${PG_VOLUME:?}:/var/lib/postgresql" \
docker.io/postgres
