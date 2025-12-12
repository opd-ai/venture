#!/bin/bash
# Build script for keyboard test WASM application

set -e

echo "Building keyboard test for WebAssembly..."

# Create output directory
mkdir -p build/keyboardtest

# Build WASM binary
GOOS=js GOARCH=wasm go build -o build/keyboardtest/keyboardtest.wasm ./examples/keyboardtest

# Copy wasm_exec.js from Go installation
cp "$(go env GOROOT)/lib/wasm/wasm_exec.js" build/keyboardtest/

# Copy HTML file
cp examples/keyboardtest/keyboardtest.html build/keyboardtest/index.html

echo "Build complete!"
echo "Output directory: build/keyboardtest"
echo ""
echo "To test locally, run:"
echo "  cd build/keyboardtest && python3 -m http.server 8080"
echo "Then open: http://localhost:8080"
