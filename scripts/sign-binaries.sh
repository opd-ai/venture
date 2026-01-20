#!/bin/bash
set -e

# Sign release binaries with GPG
# Usage: ./scripts/sign-binaries.sh [directory] [gpg-key-id]
# Example: ./scripts/sign-binaries.sh dist ABCD1234

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
NC='\033[0m'

echo_info() { echo -e "${GREEN}[INFO]${NC} $1"; }
echo_error() { echo -e "${RED}[ERROR]${NC} $1"; }

TARGET_DIR="${1:-$PROJECT_ROOT/dist}"
GPG_KEY="${2:-}"

if [ -z "$GPG_KEY" ]; then
    echo_error "GPG key ID required"
    echo "Usage: $0 <directory> <gpg-key-id>"
    exit 1
fi

if [ ! -d "$TARGET_DIR" ]; then
    echo_error "Directory $TARGET_DIR does not exist"
    exit 1
fi

if ! command -v gpg &> /dev/null; then
    echo_error "gpg is not installed"
    exit 1
fi

echo_info "Signing artifacts in $TARGET_DIR with key $GPG_KEY..."

cd "$TARGET_DIR"

# Sign the checksums file
if [ -f "SHA256SUMS.txt" ]; then
    echo_info "Signing SHA256SUMS.txt..."
    gpg --default-key "$GPG_KEY" --armor --detach-sign SHA256SUMS.txt
    echo_info "Created SHA256SUMS.txt.asc"
fi

# Optionally sign individual files
# for f in *.tar.gz *.zip *.deb *.rpm; do
#     if [ -f "$f" ]; then
#         echo_info "Signing $f..."
#         gpg --default-key "$GPG_KEY" --armor --detach-sign "$f"
#     fi
# done

echo_info "Signing complete!"
echo ""
echo "Verify signatures with:"
echo "  gpg --verify SHA256SUMS.txt.asc SHA256SUMS.txt"
