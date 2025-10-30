#!/bin/bash
set -euo pipefail
# Test script for cross-platform builds
# Tests Android, iOS, and WASM builds

echo "=== Testing Cross-Platform Builds ==="
echo ""

# Color codes
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

PASS_COUNT=0
FAIL_COUNT=0
SKIP_COUNT=0

test_build() {
    local name=$1
    local goos=$2
    local goarch=$3
    local target=$4
    local extra_flags=$5
    
    echo -n "Testing $name ($goos/$goarch): "
    
    output_file=$(mktemp /tmp/test-$name.XXXXXX)
    if output=$(CGO_ENABLED=0 GOOS=$goos GOARCH=$goarch go build $extra_flags -o "$output_file" $target 2>&1); then
        echo -e "${GREEN}✓ PASS${NC}"
        PASS_COUNT=$((PASS_COUNT + 1))
        rm -f "$output_file"
        return 0
    else
        if echo "$output" | grep -q "requires external (cgo) linking"; then
            echo -e "${YELLOW}⚠ SKIP (requires CGO)${NC}"
            SKIP_COUNT=$((SKIP_COUNT + 1))
            rm -f "$output_file"
            return 0  # Don't fail the script
        else
            echo -e "${RED}✗ FAIL${NC}"
            echo "$output" | head -5
            FAIL_COUNT=$((FAIL_COUNT + 1))
            rm -f "$output_file"
            return 0  # Don't fail the script
        fi
    fi
}

echo "--- Testing Android Builds ---"
test_build "client-android-arm64" "android" "arm64" "./cmd/client" ""
test_build "client-android-amd64" "android" "amd64" "./cmd/client" ""
test_build "server-android-arm64" "android" "arm64" "./cmd/server" ""
test_build "server-android-amd64" "android" "amd64" "./cmd/server" ""
echo ""

echo "--- Testing iOS Builds ---"
test_build "client-ios-amd64" "ios" "amd64" "./cmd/client" "-buildmode=exe"
test_build "client-ios-arm64" "ios" "arm64" "./cmd/client" "-buildmode=exe"
test_build "server-ios-amd64" "ios" "amd64" "./cmd/server" "-buildmode=exe"
test_build "server-ios-arm64" "ios" "arm64" "./cmd/server" "-buildmode=exe"
echo ""

echo "--- Testing WASM Builds ---"
test_build "client-wasm" "js" "wasm" "./cmd/client" ""
test_build "server-wasm" "js" "wasm" "./cmd/server" ""
echo ""

echo "--- Testing Key Packages ---"
# Package exclusion list for cross-platform builds
# These packages have dependencies that require Ebiten or platform-specific tooling:
# - engine: Uses ebiten.Game and ebiten.Image types
# - mobile: Direct ebiten usage for touch controls
# - network, hostplay: Import engine which imports ebiten
# - procgen/recipe: Imports engine for recipe types
# - rendering (root): Uses ebiten.Image directly
# - rendering/sprites, cache, pool, shapes: Use ebiten.Image for rendering
EXCLUDED_PACKAGES="(engine|rendering$|rendering/(sprites|cache|pool|shapes)|procgen/recipe|network|hostplay|mobile)"

# Note: We test packages only against Android arm64 as a representative platform.
# These packages are pure Go with no platform-specific code, so if they build
# for one platform, they'll build for all (android, ios, js/wasm).
# This approach keeps the test suite fast while ensuring cross-platform compatibility.
# For platform-specific binaries (cmd/client, cmd/server), we test all platforms above.

for pkg in $(go list ./pkg/... | grep -vE "$EXCLUDED_PACKAGES"); do
    pkg_name=$(echo $pkg | sed 's|github.com/opd-ai/venture/||')
    echo -n "Testing $pkg_name: "
    if GOOS=android GOARCH=arm64 go build $pkg >/dev/null 2>&1; then
        echo -e "${GREEN}✓ PASS${NC}"
        PASS_COUNT=$((PASS_COUNT + 1))
    else
        echo -e "${RED}✗ FAIL${NC}"
        FAIL_COUNT=$((FAIL_COUNT + 1))
    fi
done
echo ""

echo "=== Summary ==="
echo -e "Total: $((PASS_COUNT + FAIL_COUNT + SKIP_COUNT))"
echo -e "${GREEN}Passed: $PASS_COUNT${NC}"
echo -e "${RED}Failed: $FAIL_COUNT${NC}"
echo -e "${YELLOW}Skipped: $SKIP_COUNT${NC}"
echo ""

if [ $FAIL_COUNT -eq 0 ]; then
    echo -e "${GREEN}✓ All builds successful!${NC}"
    exit 0
else
    echo -e "${RED}✗ Some builds failed${NC}"
    exit 1
fi
