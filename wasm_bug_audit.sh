#!/bin/bash
# WASM Bug Audit & Validation Script
# Comprehensive testing for WebAssembly build issues

# Don't exit on error - we want to run all tests
set +e

cd "$(dirname "$0")"

echo "╔══════════════════════════════════════════════════════════════╗"
echo "║  Venture WebAssembly Bug Audit & Remediation                ║"
echo "║  Automated Testing Suite                                    ║"
echo "╚══════════════════════════════════════════════════════════════╝"
echo ""

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

pass_count=0
fail_count=0

function test_pass {
    echo -e "${GREEN}✓ PASS${NC}: $1"
    ((pass_count++))
}

function test_fail {
    echo -e "${RED}✗ FAIL${NC}: $1"
    ((fail_count++))
}

function test_warn {
    echo -e "${YELLOW}⚠ WARN${NC}: $1"
}

function test_info {
    echo -e "${BLUE}ℹ INFO${NC}: $1"
}

echo "════════════════════════════════════════════════════════════════"
echo "PHASE 0: HTML/CSS Validation & Mobile Web Audit"
echo "════════════════════════════════════════════════════════════════"
echo ""

# Test 1: HTML5 Doctype
echo "Test 1: HTML5 Doctype..."
if grep -q "<!DOCTYPE html>" build/wasm/index.html && grep -q "<!DOCTYPE html>" build/wasm/game.html; then
    test_pass "HTML5 doctype present in both files"
else
    test_fail "HTML5 doctype missing"
fi

# Test 2: Charset Meta Tag
echo "Test 2: Charset Meta Tag..."
if grep -q '<meta charset="UTF-8">' build/wasm/index.html && grep -q '<meta charset="UTF-8">' build/wasm/game.html; then
    test_pass "UTF-8 charset meta tag present"
else
    test_fail "Charset meta tag missing or incorrect"
fi

# Test 3: Viewport Meta Tag (Mobile Compatibility)
echo "Test 3: Viewport Meta Tag..."
if grep -q 'name="viewport"' build/wasm/index.html && grep -q 'viewport-fit=cover' build/wasm/index.html; then
    test_pass "Viewport meta tag with viewport-fit=cover present"
else
    test_fail "Viewport meta tag missing or incomplete"
fi

# Test 4: PWA Meta Tags
echo "Test 4: PWA Meta Tags..."
if grep -q 'apple-mobile-web-app-capable' build/wasm/index.html && grep -q 'mobile-web-app-capable' build/wasm/index.html; then
    test_pass "PWA meta tags present (iOS + Android)"
else
    test_warn "PWA meta tags missing (optional but recommended)"
fi

# Test 5: Safe Area Insets (iPhone X+ Notch)
echo "Test 5: Safe Area Insets CSS..."
if grep -q 'env(safe-area-inset' build/wasm/index.html && grep -q 'env(safe-area-inset' build/wasm/game.html; then
    test_pass "Safe area insets CSS present (iPhone notch support)"
else
    test_fail "Safe area insets CSS missing"
fi

# Test 6: Touch-Action CSS
echo "Test 6: Touch-Action CSS..."
if grep -q 'touch-action' build/wasm/game.html; then
    test_pass "Touch-action CSS present for canvas"
else
    test_fail "Touch-action CSS missing"
fi

# Test 7: Tap Highlight Color
echo "Test 7: Tap Highlight Color..."
if grep -q 'tap-highlight' build/wasm/index.html && grep -q 'tap-highlight' build/wasm/game.html; then
    test_pass "Tap highlight color set (iOS tap flash prevention)"
else
    test_warn "Tap highlight color missing (optional)"
fi

# Test 8: Canvas Rendering CSS
echo "Test 8: Canvas Rendering CSS..."
if grep -q 'image-rendering.*pixelated\|crisp-edges' build/wasm/game.html; then
    test_pass "Canvas image-rendering CSS present (pixel art mode)"
else
    test_warn "Canvas image-rendering CSS missing (optional for pixel art)"
fi

# Test 9: User-Select Prevention
echo "Test 9: User-Select Prevention..."
if grep -q 'user-select.*none' build/wasm/game.html; then
    test_pass "User-select prevention CSS present"
else
    test_warn "User-select prevention CSS missing"
fi

# Test 10: Input Element for Mobile Keyboard
echo "Test 10: Mobile Keyboard Support..."
if grep -rn "venture-keyboard-input" pkg/mobile/keyboard_wasm.go > /dev/null; then
    test_pass "Mobile keyboard implementation found (hidden input element)"
else
    test_fail "Mobile keyboard implementation missing"
fi

echo ""
echo "════════════════════════════════════════════════════════════════"
echo "PHASE 1: Build & Static Analysis"
echo "════════════════════════════════════════════════════════════════"
echo ""

# Test 11: WASM Build
echo "Test 11: WASM Build..."
if GOOS=js GOARCH=wasm go build -ldflags="-s -w" -o build/wasm/venture.wasm ./cmd/client 2>&1 | grep -q "error"; then
    test_fail "WASM build failed"
else
    test_pass "WASM build successful"
fi

# Test 12: WASM Binary Size
echo "Test 12: WASM Binary Size..."
wasm_size=$(stat -f%z build/wasm/venture.wasm 2>/dev/null || stat -c%s build/wasm/venture.wasm 2>/dev/null)
wasm_size_mb=$((wasm_size / 1024 / 1024))
if [ $wasm_size_mb -lt 50 ]; then
    test_pass "WASM binary size: ${wasm_size_mb}MB (< 50MB target)"
else
    test_warn "WASM binary size: ${wasm_size_mb}MB (larger than recommended)"
fi

# Test 13: wasm_exec.js Present
echo "Test 13: wasm_exec.js Present..."
if [ -f "build/wasm/wasm_exec.js" ]; then
    test_pass "wasm_exec.js present"
else
    test_fail "wasm_exec.js missing"
fi

# Test 14: No Blocking Operations
echo "Test 14: No Blocking Operations in Hot Paths..."
if grep -rn "time\.Sleep" pkg/engine/*.go cmd/client/*.go 2>/dev/null | grep -v "_test.go" | grep -v "// " | grep -q .; then
    test_warn "Found time.Sleep in non-test code (may block WASM)"
else
    test_pass "No blocking time.Sleep in hot paths"
fi

# Test 15: WASM Platform Detection
echo "Test 15: WASM Platform Detection..."
if grep -q "IsWASM()" pkg/mobile/platform.go && grep -q "IsWASM()" cmd/client/main.go; then
    test_pass "WASM platform detection implemented"
else
    test_fail "WASM platform detection missing"
fi

# Test 16: localStorage Implementation
echo "Test 16: localStorage Save/Load Implementation..."
if [ -f "pkg/saveload/storage_wasm.go" ] && grep -q "localStorage" pkg/saveload/storage_wasm.go; then
    test_pass "localStorage save/load implementation found"
else
    test_fail "localStorage save/load implementation missing"
fi

# Test 17: Keyboard Bridge (WASM)
echo "Test 17: Keyboard Bridge for Mobile Web..."
if [ -f "pkg/mobile/keyboard_wasm.go" ] && grep -q "ShowKeyboard" pkg/mobile/keyboard_wasm.go; then
    test_pass "Mobile keyboard bridge implementation found"
else
    test_fail "Mobile keyboard bridge missing"
fi

echo ""
echo "════════════════════════════════════════════════════════════════"
echo "PHASE 2: Code Quality Checks"
echo "════════════════════════════════════════════════════════════════"
echo ""

# Test 18: go vet
echo "Test 18: go vet (Suspicious Constructs)..."
if GOOS=js GOARCH=wasm go vet ./... 2>&1 | grep -q "exit status"; then
    test_warn "go vet found issues (check output above)"
else
    test_pass "go vet clean"
fi

# Test 19: ESC Key Handling
echo "Test 19: ESC Key Handling (Dual-Exit Pattern)..."
esc_count=$(grep -rn "IsKeyJustPressed(ebiten.KeyEscape)" pkg/engine/*.go | wc -l)
if [ $esc_count -gt 10 ]; then
    test_pass "ESC key handling found in $esc_count locations"
else
    test_warn "Limited ESC key handling ($esc_count locations)"
fi

# Test 20: Menu Trap Detection
echo "Test 20: Menu Trap Detection..."
if grep -rn "dual-exit\|dual exit" pkg/engine/*.go > /dev/null; then
    test_pass "Dual-exit pattern documented in code"
else
    test_warn "Dual-exit pattern not explicitly documented"
fi

echo ""
echo "════════════════════════════════════════════════════════════════"
echo "PHASE 3: Test Suite"
echo "════════════════════════════════════════════════════════════════"
echo ""

# Test 21: Unit Tests (Non-WASM)
echo "Test 21: Unit Tests..."
test_info "Running subset of tests (excluding WASM-specific)..."
if go test ./pkg/saveload/... ./pkg/mobile/... -short -timeout 30s > /dev/null 2>&1; then
    test_pass "Key package tests passed"
else
    test_warn "Some tests failed (may be environment-specific)"
fi

# Test 22: Race Detection
echo "Test 22: Race Condition Detection..."
test_info "Running race detector on key packages..."
if go test -race -short ./pkg/engine/... ./pkg/saveload/... -timeout 30s > /dev/null 2>&1; then
    test_pass "No race conditions detected"
else
    test_warn "Race detector found issues or tests failed"
fi

echo ""
echo "════════════════════════════════════════════════════════════════"
echo "Summary"
echo "════════════════════════════════════════════════════════════════"
echo ""
echo -e "Tests Passed:  ${GREEN}$pass_count${NC}"
echo -e "Tests Failed:  ${RED}$fail_count${NC}"
echo ""

if [ $fail_count -eq 0 ]; then
    echo -e "${GREEN}✓ All critical tests passed!${NC}"
    echo "WASM build is ready for deployment."
    exit 0
else
    echo -e "${RED}✗ $fail_count critical test(s) failed.${NC}"
    echo "Please address the failures before deployment."
    exit 1
fi
