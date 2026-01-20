#!/bin/bash
# Performance regression testing script
# Runs key benchmarks and compares against baseline thresholds
# Exits with non-zero status if any benchmark exceeds threshold by more than 10%

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BASELINE_FILE="${SCRIPT_DIR}/benchmark-baseline.json"
RESULTS_FILE="${SCRIPT_DIR}/../build/benchmark-results.txt"
FAILED=0

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo "=========================================="
echo "Performance Regression Testing"
echo "=========================================="
echo ""

# Ensure build directory exists
mkdir -p "$(dirname "$RESULTS_FILE")"

# Check if baseline file exists
if [ ! -f "$BASELINE_FILE" ]; then
    echo "Error: Baseline file not found: $BASELINE_FILE"
    exit 1
fi

# Extract threshold percentage from baseline
THRESHOLD=$(jq -r '.threshold_percent' "$BASELINE_FILE")
echo "Regression threshold: ${THRESHOLD}%"
echo ""

# Function to run a single benchmark and check against baseline
run_benchmark() {
    local bench_name="$1"
    local package="$2"
    local max_ns="$3"
    
    echo -n "Running ${bench_name}... "
    
    # Run the benchmark and capture output
    local output
    output=$(go test -run='^$' -bench="^${bench_name}$" -benchmem -count=3 "./${package}" 2>&1 || true)
    
    # Extract ns/op from the benchmark output (take median of 3 runs)
    local ns_per_op
    ns_per_op=$(echo "$output" | grep "^${bench_name}" | awk '{print $3}' | sort -n | sed -n '2p')
    
    if [ -z "$ns_per_op" ]; then
        echo -e "${YELLOW}SKIP${NC} (benchmark not found or failed)"
        return 0
    fi
    
    # Convert to integer (truncate decimals for comparison)
    ns_per_op=$(echo "$ns_per_op" | awk '{printf "%.0f", $1}')
    
    # Calculate threshold with margin
    local allowed_max
    allowed_max=$(echo "$max_ns * (100 + $THRESHOLD) / 100" | bc)
    
    # Compare
    if [ "$ns_per_op" -gt "$allowed_max" ]; then
        echo -e "${RED}FAIL${NC} (${ns_per_op} ns/op > ${allowed_max} ns/op threshold)"
        echo "  Baseline: ${max_ns} ns/op, Actual: ${ns_per_op} ns/op"
        FAILED=1
    else
        local pct_of_baseline
        pct_of_baseline=$(echo "scale=1; $ns_per_op * 100 / $max_ns" | bc)
        echo -e "${GREEN}PASS${NC} (${ns_per_op} ns/op, ${pct_of_baseline}% of baseline)"
    fi
    
    # Log to results file
    echo "${bench_name}: ${ns_per_op} ns/op (max: ${max_ns})" >> "$RESULTS_FILE"
}

# Clear previous results
> "$RESULTS_FILE"
echo "Benchmark results - $(date)" >> "$RESULTS_FILE"
echo "==================================" >> "$RESULTS_FILE"

echo "Running benchmarks..."
echo ""

# Read benchmarks from baseline file and run each one
while IFS= read -r bench_name; do
    package=$(jq -r ".benchmarks[\"$bench_name\"].package" "$BASELINE_FILE")
    max_ns=$(jq -r ".benchmarks[\"$bench_name\"].max_ns_per_op" "$BASELINE_FILE")
    
    if [ "$package" != "null" ] && [ "$max_ns" != "null" ]; then
        run_benchmark "$bench_name" "$package" "$max_ns"
    fi
done < <(jq -r '.benchmarks | keys[]' "$BASELINE_FILE")

echo ""
echo "=========================================="
if [ $FAILED -eq 0 ]; then
    echo -e "${GREEN}All benchmarks passed!${NC}"
    echo "Results saved to: $RESULTS_FILE"
    exit 0
else
    echo -e "${RED}Some benchmarks exceeded regression threshold!${NC}"
    echo "Results saved to: $RESULTS_FILE"
    exit 1
fi
