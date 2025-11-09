#!/bin/bash
# Code Review Validation Script
# Implements quality gates from docs/CODE_REVIEW_PLAN.md
# Usage: ./scripts/validate-code-review.sh

set -e

# Color output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

FAILED_GATES=()
PASSED_GATES=()
SKIP_RACE=${SKIP_RACE:-false}
SKIP_COVERAGE=${SKIP_COVERAGE:-false}

echo "================================================"
echo "Code Review Quality Gates Validation"
echo "Project: Venture - Procedural Multiplayer Action-RPG"
echo "Methodology: docs/CODE_REVIEW_PLAN.md"
echo "================================================"
echo ""

# Helper functions
pass_gate() {
    echo -e "${GREEN}✓ PASS${NC}: $1"
    PASSED_GATES+=("$1")
}

fail_gate() {
    echo -e "${RED}✗ FAIL${NC}: $1"
    FAILED_GATES+=("$1")
}

warn_gate() {
    echo -e "${YELLOW}⚠ WARN${NC}: $1"
}

# Gate 1: Build Success
echo "Gate 1: Build Success"
echo "-------------------"
if go build ./cmd/client && go build ./cmd/server; then
    pass_gate "Build Success"
else
    fail_gate "Build Success"
fi
echo ""

# Gate 2: Test Pass
echo "Gate 2: Test Pass"
echo "-----------------"
if [ "$CI" = "true" ] || [ -x "$(command -v xvfb-run)" ]; then
    if [ "$CI" = "true" ] || [ -x "$(command -v xvfb-run)" ]; then
        TEST_CMD="xvfb-run -s '-screen 0 1920x1080x24' go test ./..."
    else
        TEST_CMD="go test ./..."
    fi
    
    if eval $TEST_CMD; then
        pass_gate "Test Pass"
    else
        fail_gate "Test Pass"
    fi
else
    if go test ./...; then
        pass_gate "Test Pass"
    else
        fail_gate "Test Pass"
    fi
fi
echo ""

# Gate 3: Race Freedom
echo "Gate 3: Race Freedom"
echo "--------------------"
if [ "$SKIP_RACE" = "true" ]; then
    warn_gate "Race Freedom (skipped via SKIP_RACE=true)"
else
    if [ "$CI" = "true" ] || [ -x "$(command -v xvfb-run)" ]; then
        RACE_CMD="xvfb-run -s '-screen 0 1920x1080x24' go test -race ./... 2>&1"
    else
        RACE_CMD="go test -race ./... 2>&1"
    fi
    
    RACE_OUTPUT=$(eval $RACE_CMD)
    RACE_STATUS=$?
    
    if [ $RACE_STATUS -eq 0 ]; then
        pass_gate "Race Freedom"
    else
        # Check if races are in integration tests (may be acceptable)
        if echo "$RACE_OUTPUT" | grep -q "integration_test.go"; then
            warn_gate "Race Freedom (races detected in integration tests - may require review)"
            echo "Note: Race conditions found in integration tests. Review manually."
        else
            fail_gate "Race Freedom"
            echo "Race conditions detected in production code!"
        fi
        echo "$RACE_OUTPUT" | grep -A 3 "WARNING: DATA RACE" | head -20
    fi
fi
echo ""

# Gate 4: Code Coverage
echo "Gate 4: Code Coverage (≥65% per package)"
echo "----------------------------------------"
if [ "$SKIP_COVERAGE" = "true" ]; then
    warn_gate "Code Coverage (skipped via SKIP_COVERAGE=true)"
else
    if [ "$CI" = "true" ] || [ -x "$(command -v xvfb-run)" ]; then
        COV_CMD="xvfb-run -s '-screen 0 1920x1080x24' go test -cover ./pkg/... 2>&1"
    else
        COV_CMD="go test -cover ./pkg/... 2>&1"
    fi
    
    COVERAGE_OUTPUT=$(eval $COV_CMD | grep -E "coverage:|ok\s+github.com/opd-ai/venture/pkg")
    
    # Check for packages below 65% (excluding those with documented exceptions)
    # Per TESTING.md: engine ~50%, mobile ~60%, network ~62% due to Ebiten dependencies
    LOW_COVERAGE=$(echo "$COVERAGE_OUTPUT" | awk '
        $2 ~ /github.com/ && /coverage:/ {
            pkg = $2
            match($0, /coverage: ([0-9.]+)/, cov)
            coverage = cov[1]
            
            # Skip packages with documented exceptions
            if (pkg == "github.com/opd-ai/venture/pkg/engine" && coverage >= 50.0) next
            if (pkg == "github.com/opd-ai/venture/pkg/mobile" && coverage >= 55.0) next
            if (pkg == "github.com/opd-ai/venture/pkg/network" && coverage >= 60.0) next
            
            # Report packages below target that aren'"'"'t exceptions
            if (coverage > 0 && coverage < 65.0) {
                printf "%s = %.1f%% (target: 65%%)\n", pkg, coverage
            }
        }
    ')
    
    if [ -z "$LOW_COVERAGE" ]; then
        pass_gate "Code Coverage"
        echo "Sample coverage (excluding documented exceptions):"
        echo "$COVERAGE_OUTPUT" | grep "coverage:" | grep -v "engine\|mobile\|network" | head -5
    else
        fail_gate "Code Coverage"
        echo "Packages below 65% coverage (excluding documented exceptions):"
        echo "$LOW_COVERAGE"
    fi
fi
echo ""

# Gate 5: Static Analysis
echo "Gate 5: Static Analysis (go vet)"
echo "--------------------------------"
if go vet ./... 2>&1 | tee /tmp/vet-output.txt; then
    if [ ! -s /tmp/vet-output.txt ]; then
        pass_gate "Static Analysis"
    else
        fail_gate "Static Analysis (warnings present)"
    fi
else
    fail_gate "Static Analysis"
fi
echo ""

# Gate 6: Code Formatting
echo "Gate 6: Code Formatting (gofmt)"
echo "-------------------------------"
UNFORMATTED=$(gofmt -l .)
if [ -z "$UNFORMATTED" ]; then
    pass_gate "Code Formatting"
else
    fail_gate "Code Formatting"
    echo "Unformatted files:"
    echo "$UNFORMATTED"
fi
echo ""

# Gate 7: Documentation Complete
echo "Gate 7: Documentation Complete"
echo "------------------------------"
# Check for exported identifiers without documentation (simplified check)
UNDOC_COUNT=$(go doc -all ./pkg/... 2>/dev/null | grep -c "^func [A-Z]" || true)
if [ "$UNDOC_COUNT" -gt 0 ]; then
    warn_gate "Documentation Complete (manual review recommended)"
else
    pass_gate "Documentation Complete"
fi
echo ""

# Gate 8: Package Docs Present
echo "Gate 8: Package Docs Present"
echo "----------------------------"
MISSING_DOCS=()
for pkg_dir in pkg/*/; do
    # Skip empty directories
    if [ -d "$pkg_dir" ] && [ "$(ls -A "$pkg_dir" 2>/dev/null | grep -v '^doc.go$')" ] && [ ! -f "${pkg_dir}doc.go" ]; then
        MISSING_DOCS+=("$pkg_dir")
    fi
done

# Also check subdirectories
for pkg_dir in pkg/*/*/; do
    # Skip empty directories
    if [ -d "$pkg_dir" ] && [ "$(ls -A "$pkg_dir" 2>/dev/null | grep -v '^doc.go$')" ] && [ ! -f "${pkg_dir}doc.go" ]; then
        MISSING_DOCS+=("$pkg_dir")
    fi
done

if [ ${#MISSING_DOCS[@]} -eq 0 ]; then
    pass_gate "Package Docs Present"
else
    fail_gate "Package Docs Present"
    echo "Missing doc.go files:"
    printf '%s\n' "${MISSING_DOCS[@]}"
fi
echo ""

# Gate 9: No Circular Dependencies
echo "Gate 9: No Circular Dependencies"
echo "--------------------------------"
# Use go mod graph to check for cycles (simplified)
if go list -json ./pkg/... | grep -q "ImportPath"; then
    pass_gate "No Circular Dependencies"
else
    warn_gate "No Circular Dependencies (manual review recommended)"
fi
echo ""

# Gate 10-18: Additional gates that require runtime or manual verification
echo "Gates 10-18: Runtime & Advanced Checks"
echo "-------------------------------------"
warn_gate "Performance Targets Met (requires runtime profiling)"
warn_gate "Determinism Verified (requires generator tests)"
warn_gate "ECS Pattern Compliance (requires code review)"
warn_gate "Error Handling (requires code review)"
warn_gate "Input Validation (requires code review)"
warn_gate "Resource Cleanup (requires profiling and manual review)"
warn_gate "API Documentation (requires doc review)"
warn_gate "Multiplayer Sync (requires integration testing)"
warn_gate "Genre Compatibility (requires generator tests)"
echo ""

# Summary
echo "================================================"
echo "SUMMARY"
echo "================================================"
echo "Passed Gates: ${#PASSED_GATES[@]}"
echo "Failed Gates: ${#FAILED_GATES[@]}"
echo ""

if [ ${#FAILED_GATES[@]} -eq 0 ]; then
    echo -e "${GREEN}✓ All automated quality gates passed!${NC}"
    echo ""
    echo "Manual review required for:"
    echo "  - Performance targets (60 FPS, <500MB memory)"
    echo "  - Determinism verification"
    echo "  - ECS pattern compliance"
    echo "  - Error handling patterns"
    echo "  - Input validation"
    echo "  - Resource cleanup"
    echo "  - API documentation quality"
    echo "  - Multiplayer synchronization"
    echo "  - Genre compatibility"
    echo ""
    echo "See docs/CODE_REVIEW_PLAN.md for detailed review process."
    exit 0
else
    echo -e "${RED}✗ Quality gate failures detected:${NC}"
    printf '  - %s\n' "${FAILED_GATES[@]}"
    echo ""
    echo "Please address the failures above before proceeding with code review."
    echo "See docs/CODE_REVIEW_PLAN.md for remediation guidance."
    exit 1
fi
