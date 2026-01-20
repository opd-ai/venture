#!/bin/bash
set -e

# Build all platforms for Venture release
# Usage: ./scripts/build-all-platforms.sh [version]
# Example: ./scripts/build-all-platforms.sh 1.0.0

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
BUILD_DIR="$PROJECT_ROOT/build"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo_info() { echo -e "${GREEN}[INFO]${NC} $1"; }
echo_warn() { echo -e "${YELLOW}[WARN]${NC} $1"; }
echo_error() { echo -e "${RED}[ERROR]${NC} $1"; }

VERSION="${1:-1.0.0}"

echo_info "Building Venture $VERSION for all platforms..."
echo ""

# Create build directory
mkdir -p "$BUILD_DIR"
cd "$PROJECT_ROOT"

# Build matrix
PLATFORMS=(
    "linux:amd64:"
    "linux:arm64:"
    "darwin:amd64:"
    "darwin:arm64:"
    "windows:amd64:.exe"
)

FAILED=0

for platform in "${PLATFORMS[@]}"; do
    IFS=':' read -r os arch ext <<< "$platform"
    
    echo_info "Building for $os/$arch..."
    
    # Build server
    if GOOS="$os" GOARCH="$arch" CGO_ENABLED=0 go build -ldflags="-s -w" \
        -o "$BUILD_DIR/venture-server-$os-$arch$ext" ./cmd/server 2>/dev/null; then
        echo_info "  ✓ Server built"
    else
        echo_error "  ✗ Server build failed"
        FAILED=1
    fi
    
    # Build client
    if GOOS="$os" GOARCH="$arch" CGO_ENABLED=0 go build -ldflags="-s -w" \
        -o "$BUILD_DIR/venture-client-$os-$arch$ext" ./cmd/client 2>/dev/null; then
        echo_info "  ✓ Client built"
    else
        echo_error "  ✗ Client build failed"
        FAILED=1
    fi
done

# Build WebAssembly
echo_info "Building WebAssembly..."
mkdir -p "$BUILD_DIR/wasm"
if GOOS=js GOARCH=wasm go build -ldflags="-s -w" \
    -o "$BUILD_DIR/wasm/venture.wasm" ./cmd/client 2>/dev/null; then
    cp "$(go env GOROOT)/lib/wasm/wasm_exec.js" "$BUILD_DIR/wasm/"
    echo_info "  ✓ WASM built"
else
    echo_error "  ✗ WASM build failed"
    FAILED=1
fi

echo ""
if [ $FAILED -eq 0 ]; then
    echo_info "All platforms built successfully!"
    echo_info "Artifacts in: $BUILD_DIR"
else
    echo_error "Some builds failed!"
    exit 1
fi

# List built files
echo ""
echo_info "Built artifacts:"
ls -lh "$BUILD_DIR"/venture-* 2>/dev/null || true
ls -lh "$BUILD_DIR/wasm"/*.wasm 2>/dev/null || true
