#!/usr/bin/env bash
# Wrapper around `podman compose up` for hosts without a systemd user session
# (WSL2). podman normally drives container healthchecks with systemd user
# timers; without them the health status never leaves "starting", so
# podman-compose blocks forever on `depends_on: condition: service_healthy`
# (it waits via `podman wait --condition=healthy`).
# This script polls `podman healthcheck run` for every running container while
# compose is bringing the stack up, then stops.
set -euo pipefail

cleanup() {
  [[ -n "${HC_LOOP_PID:-}" ]] && kill "$HC_LOOP_PID" 2>/dev/null || true
}
trap cleanup EXIT INT TERM

hc_loop() {
  while true; do
    for c in $(podman ps --format '{{.Names}}' 2>/dev/null); do
      # Fails for containers without a healthcheck — ignore.
      podman healthcheck run "$c" >/dev/null 2>&1 || true
    done
    sleep 2
  done
}

hc_loop &
HC_LOOP_PID=$!

podman compose up -d "$@"
