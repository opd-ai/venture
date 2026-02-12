#!/usr/bin/env bash
# validate-network-types.sh
# 
# Validates that code follows the networking best practices defined in README.md
# by detecting concrete network types that should be replaced with interfaces.
#
# Usage:
#   ./scripts/validate-network-types.sh          # Check all Go files in pkg/
#   ./scripts/validate-network-types.sh <path>   # Check specific path
#
# Exit codes:
#   0 - No violations found
#   1 - Violations found

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
NC='\033[0m' # No Color

# Target path (default to pkg/)
TARGET_PATH="${1:-pkg/}"

echo "🔍 Validating network types in ${TARGET_PATH}..."
echo ""

# Pattern to find concrete network types
PATTERN='\*net\.(UDPAddr|TCPAddr|UDPConn|TCPConn|TCPListener|UnixAddr|UnixConn|UnixListener|IPAddr)'

# Find violations using grep
# Exclude: vendor/, generated files, comments, and imports
VIOLATIONS=$(find "$TARGET_PATH" -name "*.go" -type f \
    ! -path "*/vendor/*" \
    ! -name "*.pb.go" \
    -exec grep -Hn -E "$PATTERN" {} \; 2>/dev/null | \
    grep -v '//.*\*net\.' | \
    grep -v 'import ' | \
    grep -v 'mockAddr' || true)

if [ -z "$VIOLATIONS" ]; then
    FILECOUNT=$(find "$TARGET_PATH" -name "*.go" -type f ! -path "*/vendor/*" ! -name "*.pb.go" 2>/dev/null | wc -l)
    echo -e "${GREEN}✓${NC} All network types follow interface-based design"
    echo -e "${GREEN}✓${NC} Checked ${FILECOUNT} files"
    exit 0
else
    echo -e "${RED}✗${NC} Found violations of network type interface policy:"
    echo ""
    echo "$VIOLATIONS"
    echo ""
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo "See README.md 'Networking Best Practices' for guidelines:"
    echo "  - Use net.Addr (not *net.UDPAddr, *net.TCPAddr)"
    echo "  - Use net.Conn (not *net.TCPConn)"
    echo "  - Use net.PacketConn (not *net.UDPConn)"
    echo "  - Use net.Listener (not *net.TCPListener)"
    exit 1
fi
