#!/bin/bash
# Automated Code Review Script for Venture Packages
# Performs comprehensive code review based on CODE_REVIEW_PLAN.md

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Parse arguments
PACKAGE_PATH=""
VERBOSE=0

while [[ $# -gt 0 ]]; do
    case $1 in
        -p|--package)
            PACKAGE_PATH="$2"
            shift 2
            ;;
        -v|--verbose)
            VERBOSE=1
            shift
            ;;
        *)
            echo "Unknown option: $1"
            echo "Usage: $0 -p <package_path> [-v]"
            exit 1
            ;;
    esac
done

if [ -z "$PACKAGE_PATH" ]; then
    echo "Error: Package path is required"
    echo "Usage: $0 -p <package_path> [-v]"
    exit 1
fi

# Ensure package path is absolute
if [[ ! "$PACKAGE_PATH" = /* ]]; then
    PACKAGE_PATH="$REPO_ROOT/$PACKAGE_PATH"
fi

if [ ! -d "$PACKAGE_PATH" ]; then
    echo "Error: Package directory does not exist: $PACKAGE_PATH"
    exit 1
fi

PKG_NAME=$(basename "$PACKAGE_PATH")
PKG_REL_PATH=$(realpath --relative-to="$REPO_ROOT" "$PACKAGE_PATH")
GO_PKG_PATH="github.com/opd-ai/venture/$PKG_REL_PATH"

echo "==================================================================="
echo "Automated Code Review for: $PKG_REL_PATH"
echo "==================================================================="
echo

# Create temporary directory for review outputs
REVIEW_DIR=$(mktemp -d)
trap "rm -rf $REVIEW_DIR" EXIT

# Results tracking
declare -A QUALITY_GATES
CRITICAL_ISSUES=0
MAJOR_ISSUES=0
MINOR_ISSUES=0

log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

run_check() {
    local gate_name=$1
    local command=$2
    local output_file=$3
    
    echo -n "  Checking $gate_name... "
    
    if eval "$command" > "$output_file" 2>&1; then
        echo -e "${GREEN}PASS${NC}"
        QUALITY_GATES[$gate_name]="PASS"
        return 0
    else
        echo -e "${RED}FAIL${NC}"
        QUALITY_GATES[$gate_name]="FAIL"
        [ $VERBOSE -eq 1 ] && cat "$output_file"
        return 1
    fi
}

# Phase 1: Static Analysis
log_info "Phase 1: Static Analysis & Structure Review"

# 1.1 Go vet
run_check "go_vet" \
    "cd $REPO_ROOT && go vet ./$PKG_REL_PATH" \
    "$REVIEW_DIR/vet.txt"

# 1.2 Go fmt
run_check "go_fmt" \
    "cd $PACKAGE_PATH && test -z \"\$(gofmt -l .)\"" \
    "$REVIEW_DIR/fmt.txt"
    
# 1.3 Build test
run_check "build" \
    "cd $REPO_ROOT && go build ./$PKG_REL_PATH" \
    "$REVIEW_DIR/build.txt"

# Phase 2: Testing
log_info "Phase 2: Testing & Coverage"

# 2.1 Test execution
run_check "tests" \
    "cd $REPO_ROOT && go test ./$PKG_REL_PATH" \
    "$REVIEW_DIR/test.txt"

# 2.2 Race detection
run_check "race" \
    "cd $REPO_ROOT && go test -race ./$PKG_REL_PATH" \
    "$REVIEW_DIR/race.txt"

# 2.3 Coverage analysis
echo -n "  Checking coverage... "
cd $REPO_ROOT
go test -cover -coverprofile="$REVIEW_DIR/coverage.out" ./$PKG_REL_PATH > "$REVIEW_DIR/coverage.txt" 2>&1
COVERAGE=$(go tool cover -func="$REVIEW_DIR/coverage.out" 2>/dev/null | tail -1 | awk '{print $3}' | sed 's/%//')
if [ -n "$COVERAGE" ]; then
    if (( $(echo "$COVERAGE >= 65.0" | bc -l) )); then
        echo -e "${GREEN}PASS${NC} (${COVERAGE}%)"
        QUALITY_GATES["coverage"]="PASS"
    else
        echo -e "${YELLOW}WARN${NC} (${COVERAGE}% < 65%)"
        QUALITY_GATES["coverage"]="WARN"
    fi
else
    echo -e "${YELLOW}N/A${NC}"
    QUALITY_GATES["coverage"]="N/A"
fi

# Phase 3: Documentation & API
log_info "Phase 3: Documentation & API Review"

# 3.1 Check for doc.go
if [ -f "$PACKAGE_PATH/doc.go" ]; then
    echo -e "  Package documentation: ${GREEN}PASS${NC}"
    QUALITY_GATES["doc_go"]="PASS"
else
    echo -e "  Package documentation: ${RED}FAIL${NC} (missing doc.go)"
    QUALITY_GATES["doc_go"]="FAIL"
    ((MAJOR_ISSUES++))
fi

# 3.2 Check godoc coverage
echo -n "  Checking godoc coverage... "
cd $PACKAGE_PATH
UNDOCUMENTED=$(go doc -all 2>/dev/null | grep -c "^func\|^type" || echo "0")
if [ "$UNDOCUMENTED" -eq 0 ]; then
    echo -e "${GREEN}PASS${NC}"
    QUALITY_GATES["godoc"]="PASS"
else
    echo -e "${YELLOW}WARN${NC} (some exports may lack documentation)"
    QUALITY_GATES["godoc"]="WARN"
fi

# Phase 4: Pattern Compliance
log_info "Phase 4: Pattern Compliance Checks"

# 4.1 Check for deterministic RNG patterns
echo -n "  Checking deterministic generation... "
if grep -r "time.Now()" "$PACKAGE_PATH"/*.go 2>/dev/null | grep -v "_test.go" > "$REVIEW_DIR/time_now.txt"; then
    echo -e "${RED}FAIL${NC} (uses time.Now())"
    QUALITY_GATES["determinism"]="FAIL"
    ((CRITICAL_ISSUES++))
else
    echo -e "${GREEN}PASS${NC}"
    QUALITY_GATES["determinism"]="PASS"
fi

# 4.2 Check for global rand usage
echo -n "  Checking RNG isolation... "
if grep -r "rand\\.Intn\|rand\\.Float64" "$PACKAGE_PATH"/*.go 2>/dev/null | grep -v "_test.go" | grep -v "rand.New" > "$REVIEW_DIR/global_rand.txt"; then
    echo -e "${YELLOW}WARN${NC} (may use global rand)"
    QUALITY_GATES["rng_isolation"]="WARN"
else
    echo -e "${GREEN}PASS${NC}"
    QUALITY_GATES["rng_isolation"]="PASS"
fi

# 4.3 Error handling check
echo -n "  Checking error handling... "
if grep -r "_, *err *:=\|err *:=.*\$" "$PACKAGE_PATH"/*.go 2>/dev/null | grep -v "_test.go" > "$REVIEW_DIR/errors.txt"; then
    # Check if errors are actually handled
    UNCHECKED=$(grep -r "_ *= *.*().*err" "$PACKAGE_PATH"/*.go 2>/dev/null | wc -l || echo "0")
    if [ "$UNCHECKED" -gt 0 ]; then
        echo -e "${YELLOW}WARN${NC} (potential unchecked errors)"
        QUALITY_GATES["error_handling"]="WARN"
    else
        echo -e "${GREEN}PASS${NC}"
        QUALITY_GATES["error_handling"]="PASS"
    fi
else
    echo -e "${GREEN}PASS${NC}"
    QUALITY_GATES["error_handling"]="PASS"
fi

# Generate dependency count
log_info "Analyzing dependencies..."
cd $REPO_ROOT
DEPS=$(go list -f '{{join .Imports "\n"}}' ./$PKG_REL_PATH | grep "github.com/opd-ai/venture/pkg/" | wc -l)
echo "  Internal dependencies: $DEPS"

# Summary
echo
echo "==================================================================="
echo "Review Summary"
echo "==================================================================="
echo "Package: $PKG_REL_PATH"
echo "Dependency Depth: $DEPS"
echo "Coverage: ${COVERAGE:-N/A}%"
echo
echo "Quality Gates:"
for gate in "${!QUALITY_GATES[@]}"; do
    status="${QUALITY_GATES[$gate]}"
    case $status in
        PASS)
            echo -e "  [${GREEN}✓${NC}] $gate"
            ;;
        WARN)
            echo -e "  [${YELLOW}!${NC}] $gate"
            ;;
        FAIL)
            echo -e "  [${RED}✗${NC}] $gate"
            ;;
        N/A)
            echo -e "  [-] $gate"
            ;;
    esac
done
echo
echo "Issues Found:"
echo "  Critical: $CRITICAL_ISSUES"
echo "  Major: $MAJOR_ISSUES"
echo "  Minor: $MINOR_ISSUES"
echo

# Export results for AUDIT.md generation
cat > "$REVIEW_DIR/results.env" <<EOF
PKG_REL_PATH=$PKG_REL_PATH
PKG_NAME=$PKG_NAME
DEPS=$DEPS
COVERAGE=${COVERAGE:-N/A}
CRITICAL_ISSUES=$CRITICAL_ISSUES
MAJOR_ISSUES=$MAJOR_ISSUES
MINOR_ISSUES=$MINOR_ISSUES
EOF

# Export quality gates
for gate in "${!QUALITY_GATES[@]}"; do
    echo "GATE_${gate}=${QUALITY_GATES[$gate]}" >> "$REVIEW_DIR/results.env"
done

echo "Review complete. Results saved to: $REVIEW_DIR"
echo "Results environment file: $REVIEW_DIR/results.env"

# Copy results to a known location
RESULTS_FILE="$REPO_ROOT/.code_review_results"
cp "$REVIEW_DIR/results.env" "$RESULTS_FILE"
echo "Results copied to: $RESULTS_FILE"
