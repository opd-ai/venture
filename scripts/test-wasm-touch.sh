#!/bin/bash
# Manual WASM Touch Input Testing Script
# This script helps set up and guide through manual testing of touch input on WASM build

set -e

echo "========================================"
echo "WASM Touch Input Testing Helper"
echo "========================================"
echo ""

# Check if we're in the right directory
if [ ! -f "go.mod" ]; then
    echo "Error: Must run from repository root"
    exit 1
fi

echo "Step 1: Building WASM version..."
echo "Running: make build-wasm"
if make build-wasm 2>&1 | tail -10; then
    echo "✓ WASM build successful"
else
    echo "✗ WASM build failed"
    exit 1
fi

echo ""
echo "Step 2: Starting local server..."
echo "Running: make serve-wasm in background"
echo ""

# Start server in background
make serve-wasm &
SERVER_PID=$!

# Give server time to start
sleep 2

echo "✓ Server started (PID: $SERVER_PID)"
echo ""
echo "========================================"
echo "Testing Instructions"
echo "========================================"
echo ""
echo "1. On this machine (if touch-capable):"
echo "   Open: http://localhost:8080"
echo ""
echo "2. On mobile device (same network):"
echo "   a. Get your local IP:"
if command -v ip &> /dev/null; then
    LOCAL_IP=$(ip addr show | grep "inet " | grep -v 127.0.0.1 | awk '{print $2}' | cut -d/ -f1 | head -1)
    echo "      Your IP: $LOCAL_IP"
    echo "   b. Open: http://$LOCAL_IP:8080"
elif command -v ifconfig &> /dev/null; then
    LOCAL_IP=$(ifconfig | grep "inet " | grep -v 127.0.0.1 | awk '{print $2}' | head -1)
    echo "      Your IP: $LOCAL_IP"
    echo "   b. Open: http://$LOCAL_IP:8080"
else
    echo "      Run: ip addr show (Linux) or ifconfig (Mac)"
    echo "   b. Open: http://<your-ip>:8080"
fi
echo ""
echo "3. Touch the screen - virtual controls should appear"
echo ""
echo "========================================"
echo "Test Checklist"
echo "========================================"
echo ""
echo "□ Virtual controls appear on first touch"
echo "□ D-pad controls character movement"
echo "□ Action button (A) triggers attack"
echo "□ Secondary button (B) uses item"
echo "□ Menu button (☰) opens pause menu"
echo "□ Tap gesture works (quick tap)"
echo "□ Swipe gesture works (drag)"
echo "□ Long press works (hold 500ms+)"
echo "□ Double tap works (tap twice quickly)"
echo "□ Pinch zoom works (two fingers)"
echo "□ Touch switches off when using keyboard"
echo "□ Touch switches back on when touching again"
echo "□ No errors in browser console (F12)"
echo ""
echo "See docs/TESTING_TOUCH_INPUT.md for detailed testing guide"
echo ""
echo "Press Ctrl+C when done testing to stop the server"
echo ""

# Wait for user to stop
trap "echo ''; echo 'Stopping server...'; kill $SERVER_PID 2>/dev/null; exit 0" INT TERM

# Keep script running
wait $SERVER_PID
