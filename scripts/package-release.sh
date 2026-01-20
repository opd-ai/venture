#!/bin/bash
set -e

# Package release artifacts for Venture
# Usage: ./scripts/package-release.sh [version]
# Example: ./scripts/package-release.sh 1.0.0

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
BUILD_DIR="$PROJECT_ROOT/build"
DIST_DIR="$PROJECT_ROOT/dist"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo_info() { echo -e "${GREEN}[INFO]${NC} $1"; }
echo_warn() { echo -e "${YELLOW}[WARN]${NC} $1"; }
echo_error() { echo -e "${RED}[ERROR]${NC} $1"; }

VERSION="${1:-1.0.0}"

echo_info "Packaging Venture $VERSION release..."
echo ""

# Create dist directories
mkdir -p "$DIST_DIR"/{linux,darwin,windows,wasm}

cd "$BUILD_DIR"

# Package Linux
for arch in amd64 arm64; do
    if [ -f "venture-server-linux-$arch" ] && [ -f "venture-client-linux-$arch" ]; then
        echo_info "Packaging linux-$arch..."
        tar czf "$DIST_DIR/linux/venture-linux-$arch.tar.gz" \
            "venture-server-linux-$arch" \
            "venture-client-linux-$arch"
    fi
done

# Package macOS
for arch in amd64 arm64; do
    if [ -f "venture-server-darwin-$arch" ] && [ -f "venture-client-darwin-$arch" ]; then
        echo_info "Packaging darwin-$arch..."
        tar czf "$DIST_DIR/darwin/venture-darwin-$arch.tar.gz" \
            "venture-server-darwin-$arch" \
            "venture-client-darwin-$arch"
    fi
done

# Package Windows
if [ -f "venture-server-windows-amd64.exe" ] && [ -f "venture-client-windows-amd64.exe" ]; then
    echo_info "Packaging windows-amd64..."
    zip -q "$DIST_DIR/windows/venture-windows-amd64.zip" \
        "venture-server-windows-amd64.exe" \
        "venture-client-windows-amd64.exe"
fi

# Package WASM
if [ -f "wasm/venture.wasm" ]; then
    echo_info "Packaging wasm..."
    cd wasm
    tar czf "$DIST_DIR/wasm/venture-wasm.tar.gz" venture.wasm wasm_exec.js
    cd ..
fi

echo ""
echo_info "Packaging complete!"
echo_info "Artifacts in: $DIST_DIR"

# List packages
echo ""
echo_info "Created packages:"
find "$DIST_DIR" -type f \( -name "*.tar.gz" -o -name "*.zip" \) -exec ls -lh {} \;
