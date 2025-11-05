#!/bin/bash
# WASM Touch Control Verification Script
# This script helps verify that touch controls work in the WASM build

set -e

echo "==================================="
echo "WASM Touch Control Verification"
echo "==================================="
echo ""

# Step 1: Build WASM
echo "[1/5] Building WASM..."
make build-wasm
echo "✓ Build successful"
echo ""

# Step 2: Check files exist
echo "[2/5] Verifying build artifacts..."
if [ ! -f "build/wasm/venture.wasm" ]; then
    echo "✗ venture.wasm not found"
    exit 1
fi

if [ ! -f "build/wasm/wasm_exec.js" ]; then
    echo "✗ wasm_exec.js not found"
    exit 1
fi

if [ ! -f "build/wasm/index.html" ]; then
    echo "✗ index.html not found"
    exit 1
fi

WASM_SIZE=$(stat -f%z "build/wasm/venture.wasm" 2>/dev/null || stat -c%s "build/wasm/venture.wasm" 2>/dev/null)
echo "✓ venture.wasm: $(numfmt --to=iec-i --suffix=B $WASM_SIZE 2>/dev/null || echo "$WASM_SIZE bytes")"
echo "✓ wasm_exec.js: present"
echo "✓ index.html: present"
echo ""

# Step 3: Check HTML has proper meta tags
echo "[3/5] Checking HTML configuration..."
if grep -q 'touch-action: none' build/wasm/game.html; then
    echo "✓ touch-action: none found in CSS"
else
    echo "✗ touch-action: none NOT found - touch may not work"
fi

if grep -q 'viewport' build/wasm/game.html; then
    echo "✓ viewport meta tag found"
else
    echo "⚠ viewport meta tag missing - may affect mobile display"
fi
echo ""

# Step 4: Start local server
echo "[4/5] Starting local server..."
echo ""
echo "Server will start on http://localhost:8080"
echo ""
echo "==================================="
echo "Manual Testing Instructions:"
echo "==================================="
echo ""
echo "Desktop Testing (Chrome DevTools):"
echo "  1. Open http://localhost:8080 in Chrome"
echo "  2. Press F12 to open DevTools"
echo "  3. Click the 'Toggle device toolbar' icon (phone/tablet icon)"
echo "  4. Select 'iPad' or 'iPhone' from the device dropdown"
echo "  5. Reload the page"
echo "  6. Look for:"
echo "     • D-pad (circle) in bottom-left corner"
echo "     • Action buttons (A, B) in bottom-right corner"
echo "     • Menu button (☰) in top-right corner"
echo "  7. Click and drag the D-pad - player should move"
echo "  8. Click action buttons - should trigger attacks"
echo ""
echo "Mobile Device Testing:"
echo "  1. Get your local IP: ip addr show | grep 'inet '"
echo "  2. On mobile browser, open http://YOUR_IP:8080"
echo "  3. Touch the screen - controls should appear"
echo "  4. Touch D-pad to move, buttons to attack"
echo ""
echo "Console Checks (F12 → Console):"
echo "  • Look for: 'virtual controls initialized for touch-capable platform'"
echo "  • NO errors about 'Failed to load WASM'"
echo "  • NO errors about 'touch is not defined'"
echo ""
echo "==================================="
echo ""
echo "[5/5] Press Ctrl+C to stop the server when done testing"
echo ""

# Start server
cd build/wasm
python3 -m http.server 8080 || python -m SimpleHTTPServer 8080
