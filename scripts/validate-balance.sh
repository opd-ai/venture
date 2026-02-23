#!/bin/bash
# validate-balance.sh - Run balance validation for CI/CD pipelines
#
# This script runs the balance validator with appropriate settings for CI.
# It uses xvfb-run to provide a virtual display for Ebiten dependencies.
#
# Exit codes:
#   0 - All validations passed
#   1 - One or more validations failed
#   2 - Script error (missing dependencies, build failure)
#
# Usage:
#   ./scripts/validate-balance.sh                    # Run all domains with defaults
#   ./scripts/validate-balance.sh combat             # Run specific domain
#   ./scripts/validate-balance.sh all --quick        # Run with reduced simulations
#   ./scripts/validate-balance.sh all --json         # Output JSON for parsing

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

# Default settings
DOMAIN="${1:-all}"
SEED="${BALANCE_SEED:-12345}"
SIMULATIONS=""
JSON_FLAG=""
VERBOSE_FLAG=""

# Parse additional arguments
shift || true
while [[ $# -gt 0 ]]; do
    case "$1" in
        --quick)
            SIMULATIONS="--simulations 100"
            ;;
        --json)
            JSON_FLAG="--json"
            ;;
        --verbose)
            VERBOSE_FLAG="--verbose"
            ;;
        --seed)
            shift
            SEED="$1"
            ;;
        *)
            echo "Unknown option: $1" >&2
            exit 2
            ;;
    esac
    shift
done

# Build the validator if needed
VALIDATOR="$PROJECT_ROOT/balance-validator"
if [[ ! -f "$VALIDATOR" ]] || [[ "$PROJECT_ROOT/cmd/balance-validator/main.go" -nt "$VALIDATOR" ]]; then
    echo "Building balance-validator..."
    cd "$PROJECT_ROOT"
    go build -tags headless -o balance-validator ./cmd/balance-validator/
fi

# Check for xvfb-run (required for Ebiten)
if ! command -v xvfb-run &> /dev/null; then
    echo "Error: xvfb-run is required but not installed" >&2
    echo "Install with: apt-get install xvfb" >&2
    exit 2
fi

# Run validation
echo "Running balance validation..."
echo "  Domain: $DOMAIN"
echo "  Seed: $SEED"
if [[ -n "$SIMULATIONS" ]]; then
    echo "  Mode: quick (reduced simulations)"
fi
echo ""

cd "$PROJECT_ROOT"
exec xvfb-run -a "$VALIDATOR" \
    --domain "$DOMAIN" \
    --seed "$SEED" \
    $SIMULATIONS \
    $JSON_FLAG \
    $VERBOSE_FLAG
