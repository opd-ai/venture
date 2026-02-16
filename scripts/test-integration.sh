#!/bin/bash
# test-integration.sh — Run integration tests that require a display server.
#
# Several integration packages transitively import Ebiten (via pkg/world/housing),
# which requires GLFW initialization. On headless CI/Linux environments, GLFW panics
# if the DISPLAY environment variable is not set. This script handles two strategies:
#
#   1. If xvfb-run is available, use it for a real virtual framebuffer.
#   2. Otherwise, set DISPLAY=:99 as a lightweight fallback that satisfies
#      GLFW's environment check without a running X server.
#
# Usage:
#   ./scripts/test-integration.sh              # Run all integration tests
#   ./scripts/test-integration.sh -v -race     # Pass extra flags to go test
#
# Affected packages:
#   - pkg/integration/guild_housing     (imports pkg/world/housing → ebiten)
#   - pkg/integration/narrative_world   (imports pkg/engine → ebiten)
#   - pkg/integration/political_warfare (imports pkg/engine → ebiten)
#   - pkg/integration/trade_routes      (imports pkg/world/housing → ebiten)

set -euo pipefail

EXTRA_FLAGS="${*}"

# If DISPLAY is already set, run tests directly.
if [ -n "${DISPLAY:-}" ]; then
    echo "DISPLAY=$DISPLAY — running integration tests directly"
    go test ${EXTRA_FLAGS} ./pkg/integration/...
    exit $?
fi

# Strategy 1: xvfb-run (provides a full virtual framebuffer)
if command -v xvfb-run &>/dev/null; then
    echo "Using xvfb-run for headless display"
    xvfb-run -a go test ${EXTRA_FLAGS} ./pkg/integration/...
    exit $?
fi

# Strategy 2: Set DISPLAY to a placeholder value.
# GLFW checks for DISPLAY presence but these tests don't actually render,
# so a non-existent display server is sufficient.
echo "No display server found — setting DISPLAY=:99 as fallback"
export DISPLAY=:99
go test ${EXTRA_FLAGS} ./pkg/integration/...
