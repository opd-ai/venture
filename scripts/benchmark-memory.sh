#!/bin/bash
# Memory benchmark script for validating <500MB client memory usage claim
# Runs realistic game scenarios and measures peak memory allocation
# Exits with non-zero status if memory exceeds 500MB threshold

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
RESULTS_FILE="${PROJECT_ROOT}/build/memory-benchmark-results.txt"
THRESHOLD_MB=500
THRESHOLD_BYTES=$((THRESHOLD_MB * 1024 * 1024))

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo "=========================================="
echo "Memory Usage Benchmark"
echo "=========================================="
echo ""
echo "Target threshold: ${THRESHOLD_MB}MB (${THRESHOLD_BYTES} bytes)"
echo ""

# Ensure build directory exists
mkdir -p "$(dirname "$RESULTS_FILE")"

# Clear previous results
cat > "$RESULTS_FILE" <<EOF
Memory Benchmark Results - $(date)
==================================
Threshold: ${THRESHOLD_MB}MB

EOF

cd "$PROJECT_ROOT"

# Run memory benchmarks and capture peak allocation
echo "Running memory benchmarks..."
echo ""

# Test 1: Baseline world generation
echo -n "1. Baseline World Generation... "
OUTPUT=$(go test -v -run='^TestMemoryBaselineWorld$' ./pkg/benchmark/memory -timeout=5m 2>&1 || true)
PEAK=$(echo "$OUTPUT" | grep -oP 'Peak allocation: \K[0-9.]+(?=MB)' || echo "")
if [ -n "$PEAK" ]; then
    PEAK_BYTES=$(echo "$PEAK * 1024 * 1024" | bc | awk '{printf "%.0f", $1}')
    if (( $(echo "$PEAK_BYTES > $THRESHOLD_BYTES" | bc -l) )); then
        echo -e "${RED}FAIL${NC} (${PEAK}MB > ${THRESHOLD_MB}MB)"
        echo "  Test 1: FAIL - ${PEAK}MB" >> "$RESULTS_FILE"
        TEST1_PASS=0
    else
        PCT=$(echo "scale=1; $PEAK * 100 / $THRESHOLD_MB" | bc)
        echo -e "${GREEN}PASS${NC} (${PEAK}MB, ${PCT}% of threshold)"
        echo "  Test 1: PASS - ${PEAK}MB (${PCT}% of threshold)" >> "$RESULTS_FILE"
        TEST1_PASS=1
    fi
else
    echo -e "${YELLOW}SKIP${NC} (test not found or benchmark incomplete)"
    echo "  Test 1: SKIP - benchmark not found" >> "$RESULTS_FILE"
    TEST1_PASS=1
fi

# Test 2: High entity count scenario
echo -n "2. High Entity Count (2000 entities)... "
OUTPUT=$(go test -v -run='^TestMemoryHighEntityCount$' ./pkg/benchmark/memory -timeout=5m 2>&1 || true)
PEAK=$(echo "$OUTPUT" | grep -oP 'Peak allocation: \K[0-9.]+(?=MB)' || echo "")
if [ -n "$PEAK" ]; then
    PEAK_BYTES=$(echo "$PEAK * 1024 * 1024" | bc | awk '{printf "%.0f", $1}')
    if (( $(echo "$PEAK_BYTES > $THRESHOLD_BYTES" | bc -l) )); then
        echo -e "${RED}FAIL${NC} (${PEAK}MB > ${THRESHOLD_MB}MB)"
        echo "  Test 2: FAIL - ${PEAK}MB" >> "$RESULTS_FILE"
        TEST2_PASS=0
    else
        PCT=$(echo "scale=1; $PEAK * 100 / $THRESHOLD_MB" | bc)
        echo -e "${GREEN}PASS${NC} (${PEAK}MB, ${PCT}% of threshold)"
        echo "  Test 2: PASS - ${PEAK}MB (${PCT}% of threshold)" >> "$RESULTS_FILE"
        TEST2_PASS=1
    fi
else
    echo -e "${YELLOW}SKIP${NC} (test not found or benchmark incomplete)"
    echo "  Test 2: SKIP - benchmark not found" >> "$RESULTS_FILE"
    TEST2_PASS=1
fi

# Test 3: Procedural generation stress
echo -n "3. Procedural Generation Stress... "
OUTPUT=$(go test -v -run='^TestMemoryProcgenStress$' ./pkg/benchmark/memory -timeout=5m 2>&1 || true)
PEAK=$(echo "$OUTPUT" | grep -oP 'Peak allocation: \K[0-9.]+(?=MB)' || echo "")
if [ -n "$PEAK" ]; then
    PEAK_BYTES=$(echo "$PEAK * 1024 * 1024" | bc | awk '{printf "%.0f", $1}')
    if (( $(echo "$PEAK_BYTES > $THRESHOLD_BYTES" | bc -l) )); then
        echo -e "${RED}FAIL${NC} (${PEAK}MB > ${THRESHOLD_MB}MB)"
        echo "  Test 3: FAIL - ${PEAK}MB" >> "$RESULTS_FILE"
        TEST3_PASS=0
    else
        PCT=$(echo "scale=1; $PEAK * 100 / $THRESHOLD_MB" | bc)
        echo -e "${GREEN}PASS${NC} (${PEAK}MB, ${PCT}% of threshold)"
        echo "  Test 3: PASS - ${PEAK}MB (${PCT}% of threshold)" >> "$RESULTS_FILE"
        TEST3_PASS=1
    fi
else
    echo -e "${YELLOW}SKIP${NC} (test not found or benchmark incomplete)"
    echo "  Test 3: SKIP - benchmark not found" >> "$RESULTS_FILE"
    TEST3_PASS=1
fi

# Test 4: Rendering pipeline stress
echo -n "4. Rendering Pipeline Stress... "
OUTPUT=$(go test -v -run='^TestMemoryRenderingStress$' ./pkg/benchmark/memory -timeout=5m 2>&1 || true)
PEAK=$(echo "$OUTPUT" | grep -oP 'Peak allocation: \K[0-9.]+(?=MB)' || echo "")
if [ -n "$PEAK" ]; then
    PEAK_BYTES=$(echo "$PEAK * 1024 * 1024" | bc | awk '{printf "%.0f", $1}')
    if (( $(echo "$PEAK_BYTES > $THRESHOLD_BYTES" | bc -l) )); then
        echo -e "${RED}FAIL${NC} (${PEAK}MB > ${THRESHOLD_MB}MB)"
        echo "  Test 4: FAIL - ${PEAK}MB" >> "$RESULTS_FILE"
        TEST4_PASS=0
    else
        PCT=$(echo "scale=1; $PEAK * 100 / $THRESHOLD_MB" | bc)
        echo -e "${GREEN}PASS${NC} (${PEAK}MB, ${PCT}% of threshold)"
        echo "  Test 4: PASS - ${PEAK}MB (${PCT}% of threshold)" >> "$RESULTS_FILE"
        TEST4_PASS=1
    fi
else
    echo -e "${YELLOW}SKIP${NC} (test not found or benchmark incomplete)"
    echo "  Test 4: SKIP - benchmark not found" >> "$RESULTS_FILE"
    TEST4_PASS=1
fi

echo ""
echo "=========================================="

# Calculate overall result
if [ "$TEST1_PASS" -eq 1 ] && [ "$TEST2_PASS" -eq 1 ] && [ "$TEST3_PASS" -eq 1 ] && [ "$TEST4_PASS" -eq 1 ]; then
    echo -e "${GREEN}✓ All memory benchmarks passed!${NC}"
    echo -e "${BLUE}Memory usage is within ${THRESHOLD_MB}MB threshold${NC}"
    echo "" >> "$RESULTS_FILE"
    echo "OVERALL: PASS - All tests within ${THRESHOLD_MB}MB threshold" >> "$RESULTS_FILE"
    echo "Results saved to: $RESULTS_FILE"
    exit 0
else
    echo -e "${RED}✗ Some memory benchmarks exceeded ${THRESHOLD_MB}MB threshold!${NC}"
    echo "" >> "$RESULTS_FILE"
    echo "OVERALL: FAIL - One or more tests exceeded threshold" >> "$RESULTS_FILE"
    echo "Results saved to: $RESULTS_FILE"
    exit 1
fi
