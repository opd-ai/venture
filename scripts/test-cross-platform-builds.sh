#!/bin/bash
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
    
    if output=$(CGO_ENABLED=0 GOOS=$goos GOARCH=$goarch go build $extra_flags -o /tmp/test-$name $target 2>&1); then
        echo -e "${GREEN}✓ PASS${NC}"
        ((PASS_COUNT++))
        return 0
    else
        if echo "$output" | grep -q "requires external (cgo) linking"; then
            echo -e "${YELLOW}⚠ SKIP (requires CGO)${NC}"
            ((SKIP_COUNT++))
            return 0  # Don't fail the script
        else
            echo -e "${RED}✗ FAIL${NC}"
            echo "$output" | head -5
            ((FAIL_COUNT++))
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
for pkg in procgen procgen/terrain procgen/entity procgen/item procgen/magic \
           procgen/skills procgen/quest procgen/station procgen/environment \
           procgen/genre combat world saveload logging audio audio/music \
           audio/sfx audio/synthesis; do
    echo -n "Testing pkg/$pkg: "
    if GOOS=android GOARCH=arm64 go build ./pkg/$pkg >/dev/null 2>&1; then
        echo -e "${GREEN}✓ PASS${NC}"
        ((PASS_COUNT++))
    else
        echo -e "${RED}✗ FAIL${NC}"
        ((FAIL_COUNT++))
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
