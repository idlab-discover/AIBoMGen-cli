#!/usr/bin/env bash
set -euo pipefail

# Load .env if present in current directory
ENV_FILE="$(dirname "$0")/.env"
if [[ -f "$ENV_FILE" ]]; then
  # shellcheck disable=SC2046
  export $(grep -v '^#' "$ENV_FILE" | xargs -d '\n' -I{} echo {}) || true
fi

GPU_VAL="${GPU:-true}"
GPU_VAL_LOWER=$(echo "$GPU_VAL" | tr '[:upper:]' '[:lower:]')

if [[ "$GPU_VAL_LOWER" == "true" || "$GPU_VAL_LOWER" == "1" || "$GPU_VAL_LOWER" == "yes" || "$GPU_VAL_LOWER" == "on" ]]; then
  echo "GPU=true: starting stack with GPU worker (same behavior as before)"
  COMPOSE_PROFILES=gpu docker compose up --build "$@"
else
  echo "GPU=false: starting stack in CPU-only mode"
  COMPOSE_PROFILES=cpu docker compose up --build "$@"
fi
