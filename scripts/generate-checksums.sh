#!/bin/bash
set -e

# Generate SHA256 checksums for release artifacts
# Usage: ./scripts/generate-checksums.sh [directory]
# Example: ./scripts/generate-checksums.sh dist

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

# Colors for output
GREEN='\033[0;32m'
NC='\033[0m'

echo_info() { echo -e "${GREEN}[INFO]${NC} $1"; }

TARGET_DIR="${1:-$PROJECT_ROOT/dist}"

if [ ! -d "$TARGET_DIR" ]; then
    echo "Error: Directory $TARGET_DIR does not exist"
    exit 1
fi

echo_info "Generating SHA256 checksums for $TARGET_DIR..."

cd "$TARGET_DIR"

# Find all release artifacts and generate checksums
find . -type f \( -name "*.tar.gz" -o -name "*.zip" -o -name "*.deb" -o -name "*.rpm" -o -name "*.exe" \) \
    -exec sha256sum {} \; | sed 's|  \./|  |' > SHA256SUMS.txt

echo_info "Checksums written to $TARGET_DIR/SHA256SUMS.txt"
echo ""
cat SHA256SUMS.txt
