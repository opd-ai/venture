#!/bin/bash
# V7.0 Release Validation Script
# Validates all success criteria from ROADMAP_V7.md

set -e

echo "========================================"
echo "V7.0 Release Validation"
echo "========================================"
echo

# Color codes for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

PASS_COUNT=0
FAIL_COUNT=0
WARN_COUNT=0

pass() {
    echo -e "${GREEN}✓${NC} $1"
    PASS_COUNT=$((PASS_COUNT + 1))
}

fail() {
    echo -e "${RED}✗${NC} $1"
    FAIL_COUNT=$((FAIL_COUNT + 1))
}

warn() {
    echo -e "${YELLOW}!${NC} $1"
    WARN_COUNT=$((WARN_COUNT + 1))
}

echo "1. Version Check"
echo "----------------"
VERSION=$(grep 'Version = ' pkg/version/version.go | grep -o '".*"' | tr -d '"')
if [ "$VERSION" = "7.0.0" ]; then
    pass "Version is 7.0.0"
else
    fail "Version is $VERSION, expected 7.0.0"
fi
echo

echo "2. Test Coverage"
echo "----------------"
echo "Running test suite..."
if go test ./pkg/... -coverprofile=coverage.out -timeout=10m > /dev/null 2>&1; then
    pass "All tests passing"
    COVERAGE=$(go tool cover -func=coverage.out | grep total | awk '{print $3}' | tr -d '%')
    if (( $(echo "$COVERAGE >= 65.0" | bc -l) )); then
        pass "Test coverage: ${COVERAGE}% (≥65% required)"
    else
        fail "Test coverage: ${COVERAGE}% (<65% required)"
    fi
else
    fail "Test suite has failures"
fi
echo

echo "3. Display System Validation"
echo "-----------------------------"
if [ -d "pkg/rendering/display" ]; then
    pass "Display package exists"
    
    # Check for key files
    for file in manager.go scaler.go config.go; do
        if [ -f "pkg/rendering/display/$file" ]; then
            pass "  - $file present"
        else
            fail "  - $file missing"
        fi
    done
else
    fail "Display package missing"
fi
echo

echo "4. Sprite Enhancement Validation"
echo "---------------------------------"
# Check for 64x64 sprite support
if grep -q "Enhanced64HumanoidTemplate" pkg/rendering/sprites/anatomy_template.go; then
    pass "64x64 sprite templates implemented"
else
    fail "64x64 sprite templates missing"
fi

if grep -q "SelectTemplate64" pkg/rendering/sprites/anatomy_template.go; then
    pass "64x64 template selector implemented"
else
    fail "64x64 template selector missing"
fi
echo

echo "5. Animation System Validation"
echo "-------------------------------"
if [ -d "pkg/rendering/animation" ]; then
    pass "Animation package exists"
    
    # Check for key components
    if grep -q "Direction8" pkg/rendering/animation/*.go; then
        pass "  - 8-direction support"
    else
        fail "  - 8-direction support missing"
    fi
    
    if grep -q "BodyPart" pkg/rendering/animation/*.go; then
        pass "  - Body part articulation"
    else
        fail "  - Body part articulation missing"
    fi
    
    if grep -q "AnimationCache" pkg/rendering/animation/*.go; then
        pass "  - Animation caching"
    else
        fail "  - Animation caching missing"
    fi
else
    fail "Animation package missing"
fi
echo

echo "6. Wall Rendering Validation"
echo "-----------------------------"
if grep -q "GenerateEnhancedWall" pkg/rendering/tiles/walls.go; then
    pass "Enhanced wall rendering implemented"
else
    fail "Enhanced wall rendering missing"
fi

if grep -q "CornerType" pkg/rendering/tiles/walls.go; then
    pass "Corner detection implemented"
else
    fail "Corner detection missing"
fi

if grep -q "downsample2x2" pkg/rendering/tiles/walls.go; then
    pass "Anti-aliasing support"
else
    fail "Anti-aliasing support missing"
fi
echo

echo "7. Collision System Validation"
echo "-------------------------------"
if [ -f "pkg/engine/collision_precise.go" ]; then
    pass "Pixel-perfect collision file exists"
    
    if grep -q "CollisionPrecision.*0.1" pkg/engine/collision_precise.go; then
        pass "  - 0.1px precision constant"
    else
        fail "  - 0.1px precision constant missing"
    fi
    
    if grep -q "PreciseColliderComponent" pkg/engine/collision_precise.go; then
        pass "  - Precise collider component"
    else
        fail "  - Precise collider component missing"
    fi
    
    if grep -q "ApplyWallSlide" pkg/engine/collision_precise.go; then
        pass "  - Wall sliding support"
    else
        fail "  - Wall sliding support missing"
    fi
else
    fail "Pixel-perfect collision file missing"
fi
echo

echo "8. Build Validation"
echo "-------------------"
echo "Building client..."
if go build -o build/venture-client ./cmd/client > /dev/null 2>&1; then
    pass "Client builds successfully"
    SIZE=$(du -h build/venture-client | awk '{print $1}')
    echo "  Client size: $SIZE"
else
    fail "Client build failed"
fi

echo "Building server..."
if go build -o build/venture-server ./cmd/server > /dev/null 2>&1; then
    pass "Server builds successfully"
    SIZE=$(du -h build/venture-server | awk '{print $1}')
    echo "  Server size: $SIZE"
else
    fail "Server build failed"
fi
echo

echo "9. Documentation Validation"
echo "----------------------------"
DOCS=(
    "docs/ARCHITECTURE.md"
    "docs/TECHNICAL_SPEC.md"
    "docs/USER_MANUAL.md"
    "docs/API_REFERENCE.md"
    "docs/ROADMAP_V7.md"
)

for doc in "${DOCS[@]}"; do
    if [ -f "$doc" ]; then
        pass "$doc exists"
    else
        fail "$doc missing"
    fi
done
echo

echo "10. CLI Test Tools Validation"
echo "------------------------------"
TOOLS=(
    "cmd/sprite64test"
    "cmd/walltest"
    "cmd/collisiontest"
)

for tool in "${TOOLS[@]}"; do
    if [ -d "$tool" ]; then
        pass "$tool exists"
    else
        warn "$tool missing (optional)"
    fi
done
echo

echo "========================================"
echo "Validation Summary"
echo "========================================"
echo -e "${GREEN}Passed:${NC}  $PASS_COUNT"
echo -e "${RED}Failed:${NC}  $FAIL_COUNT"
echo -e "${YELLOW}Warnings:${NC} $WARN_COUNT"
echo

if [ $FAIL_COUNT -eq 0 ]; then
    echo -e "${GREEN}✓ V7.0 Release Validation PASSED${NC}"
    exit 0
else
    echo -e "${RED}✗ V7.0 Release Validation FAILED${NC}"
    echo "Please address failures before release"
    exit 1
fi
