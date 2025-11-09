#!/bin/bash
# Automated Package Code Review Script
# Selects and audits a package based on dependency depth

set -e

REPO_ROOT="/home/runner/work/venture/venture"
PKG_DIR="$REPO_ROOT/pkg"
TEMP_DIR="/tmp/audit-$$"
mkdir -p "$TEMP_DIR"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo "=== Venture Package Audit System ==="
echo "Date: $(date -I)"
echo ""

# Function to calculate dependency depth for a package
calculate_depth() {
    local pkg_path="$1"
    local pkg_name="${pkg_path#$PKG_DIR/}"
    
    # Find all Go files (excluding tests)
    local go_files=$(find "$pkg_path" -maxdepth 1 -name "*.go" -not -name "*_test.go" 2>/dev/null)
    
    if [ -z "$go_files" ]; then
        echo "999" # No Go files, deprioritize
        return
    fi
    
    # Count internal dependencies (imports from github.com/opd-ai/venture/pkg/)
    local internal_deps=$(grep -h "^[[:space:]]*\"github.com/opd-ai/venture/pkg/" $go_files 2>/dev/null | \
                          grep -v "\"$pkg_name\"" | \
                          sort -u | wc -l)
    
    echo "$internal_deps"
}

# Function to check if package has AUDIT.md
has_audit() {
    local pkg_path="$1"
    [ -f "$pkg_path/AUDIT.md" ]
}

# Function to list all packages recursively
find_all_packages() {
    find "$PKG_DIR" -type d | while read dir; do
        # Check if directory contains Go files
        if ls "$dir"/*.go 2>/dev/null | grep -v "_test.go" >/dev/null 2>&1; then
            echo "$dir"
        fi
    done
}

# Scan packages and build candidate list
echo "Scanning pkg/ directory for audit candidates..."
CANDIDATES="$TEMP_DIR/candidates.txt"
> "$CANDIDATES"

find_all_packages | while read pkg_path; do
    pkg_name="${pkg_path#$PKG_DIR/}"
    
    # Skip if already audited
    if has_audit "$pkg_path"; then
        echo "  [SKIP] $pkg_name (already audited)"
        continue
    fi
    
    # Calculate depth
    depth=$(calculate_depth "$pkg_path")
    
    echo "$depth|$pkg_name|$pkg_path" >> "$CANDIDATES"
    echo "  [SCAN] $pkg_name (depth: $depth)"
done

# Sort by depth (lowest first), then by name
SORTED="$TEMP_DIR/sorted.txt"
sort -t'|' -k1,1n -k2,2 "$CANDIDATES" > "$SORTED"

# Select the first package
if [ ! -s "$SORTED" ]; then
    echo ""
    echo -e "${YELLOW}No packages available for audit (all have AUDIT.md files)${NC}"
    exit 0
fi

SELECTED=$(head -1 "$SORTED")
DEPTH=$(echo "$SELECTED" | cut -d'|' -f1)
PKG_NAME=$(echo "$SELECTED" | cut -d'|' -f2)
PKG_PATH=$(echo "$SELECTED" | cut -d'|' -f3)

echo ""
echo -e "${GREEN}=== SELECTED PACKAGE ===${NC}"
echo "Package: pkg/$PKG_NAME"
echo "Dependency Depth: $DEPTH"
echo "Path: $PKG_PATH"
echo ""

# Export for the Go audit program
export AUDIT_PKG_NAME="$PKG_NAME"
export AUDIT_PKG_PATH="$PKG_PATH"
export AUDIT_DEPTH="$DEPTH"

# Run the actual audit
cd "$REPO_ROOT"
go run "$REPO_ROOT/cmd/auditrunner/main.go"

echo ""
echo -e "${GREEN}=== AUDIT COMPLETE ===${NC}"
echo "AUDIT.md created at: $PKG_PATH/AUDIT.md"

# Cleanup
rm -rf "$TEMP_DIR"
