#!/bin/bash
set -e

# Windows build script for Venture
# Usage: ./scripts/build-windows.sh [amd64]

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
BUILD_DIR="$PROJECT_ROOT/build"
OUTPUT_DIR="$PROJECT_ROOT/dist/windows"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

echo_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

echo_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Parse architecture
ARCH="${1:-amd64}"
if [[ ! "$ARCH" =~ ^(amd64)$ ]]; then
    echo_error "Invalid architecture: $ARCH. Must be amd64"
    exit 1
fi

echo_info "Building for Windows $ARCH..."

# Check prerequisites
if ! command -v go &> /dev/null; then
    echo_error "Go is not installed"
    exit 1
fi

# Create build directory
mkdir -p "$BUILD_DIR"
mkdir -p "$OUTPUT_DIR"

cd "$PROJECT_ROOT"

# Build server
echo_info "Building server..."
GOOS=windows GOARCH="$ARCH" go build -ldflags="-s -w" \
    -o "$BUILD_DIR/venture-server-windows-$ARCH.exe" \
    ./cmd/server

# Build client
echo_info "Building client..."
GOOS=windows GOARCH="$ARCH" go build -ldflags="-s -w" \
    -o "$BUILD_DIR/venture-client-windows-$ARCH.exe" \
    ./cmd/client

# Create archives
echo_info "Creating archives..."
cd "$BUILD_DIR"

if command -v zip &> /dev/null; then
    zip "$OUTPUT_DIR/venture-server-windows-$ARCH.zip" "venture-server-windows-$ARCH.exe"
    zip "$OUTPUT_DIR/venture-client-windows-$ARCH.zip" "venture-client-windows-$ARCH.exe"
else
    echo_warn "zip not found, using tar instead"
    tar czf "$OUTPUT_DIR/venture-server-windows-$ARCH.tar.gz" "venture-server-windows-$ARCH.exe"
    tar czf "$OUTPUT_DIR/venture-client-windows-$ARCH.tar.gz" "venture-client-windows-$ARCH.exe"
fi

echo_info "Build complete!"
echo_info "Server: $OUTPUT_DIR/venture-server-windows-$ARCH.*"
echo_info "Client: $OUTPUT_DIR/venture-client-windows-$ARCH.*"
