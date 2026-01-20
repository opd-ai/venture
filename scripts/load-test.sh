#!/bin/bash
# Load Testing Script for Venture Multiplayer Server
# Tests capacity with multiple concurrent clients and monitors resource usage.
#
# Usage:
#   ./scripts/load-test.sh                    # Default: 10 clients, 5 minutes
#   ./scripts/load-test.sh --clients 20       # Test with 20 clients
#   ./scripts/load-test.sh --duration 10m     # Test for 10 minutes
#   ./scripts/load-test.sh --server host:port # Test against specific server
#
# This script:
# 1. Builds and runs the load test tool from examples/loadtest
# 2. Monitors CPU, memory, and network usage during the test
# 3. Generates a capacity report in docs/capacity/

set -e

# Default configuration
CLIENTS=10
DURATION="5m"
SERVER="localhost:8080"
OUTPUT_DIR="docs/capacity"
VERBOSE=false
MIN_LATENCY="50ms"
MAX_LATENCY="500ms"

# Parse command-line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --clients|-c)
            CLIENTS="$2"
            shift 2
            ;;
        --duration|-d)
            DURATION="$2"
            shift 2
            ;;
        --server|-s)
            SERVER="$2"
            shift 2
            ;;
        --output|-o)
            OUTPUT_DIR="$2"
            shift 2
            ;;
        --min-latency)
            MIN_LATENCY="$2"
            shift 2
            ;;
        --max-latency)
            MAX_LATENCY="$2"
            shift 2
            ;;
        --verbose|-v)
            VERBOSE=true
            shift
            ;;
        --help|-h)
            echo "Usage: $0 [OPTIONS]"
            echo ""
            echo "Options:"
            echo "  -c, --clients NUM      Number of concurrent clients (default: 10)"
            echo "  -d, --duration TIME    Test duration (default: 5m)"
            echo "  -s, --server ADDR      Server address (default: localhost:8080)"
            echo "  -o, --output DIR       Output directory (default: docs/capacity)"
            echo "      --min-latency MS   Minimum simulated latency (default: 50ms)"
            echo "      --max-latency MS   Maximum simulated latency (default: 500ms)"
            echo "  -v, --verbose          Enable verbose logging"
            echo "  -h, --help             Show this help message"
            exit 0
            ;;
        *)
            echo "Unknown option: $1"
            exit 1
            ;;
    esac
done

# Create output directory
mkdir -p "$OUTPUT_DIR"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
REPORT_FILE="$OUTPUT_DIR/load_test_${TIMESTAMP}.md"
METRICS_FILE="$OUTPUT_DIR/metrics_${TIMESTAMP}.txt"

echo "=== Venture Load Testing ==="
echo "Server:    $SERVER"
echo "Clients:   $CLIENTS"
echo "Duration:  $DURATION"
echo "Latency:   $MIN_LATENCY - $MAX_LATENCY"
echo "Output:    $OUTPUT_DIR"
echo ""

# Check if loadtest binary exists or needs to be built
LOADTEST_BIN="./bin/loadtest"
if [[ ! -f "$LOADTEST_BIN" ]]; then
    echo "[1/4] Building load test tool..."
    go build -o "$LOADTEST_BIN" ./examples/loadtest
    echo "      Built: $LOADTEST_BIN"
else
    echo "[1/4] Using existing load test binary: $LOADTEST_BIN"
fi

# Start resource monitoring in background
echo "[2/4] Starting resource monitoring..."
(
    echo "# Resource Metrics - Load Test $TIMESTAMP" > "$METRICS_FILE"
    echo "# Timestamp, CPU%, MEM_MB, NET_RX_KB, NET_TX_KB" >> "$METRICS_FILE"
    
    # Get initial network counters
    if [[ -f /proc/net/dev ]]; then
        INITIAL_RX=$(grep -E '^\s*eth0:|^\s*ens|^\s*wlan|^\s*lo:' /proc/net/dev | head -1 | awk '{print $2}')
        INITIAL_TX=$(grep -E '^\s*eth0:|^\s*ens|^\s*wlan|^\s*lo:' /proc/net/dev | head -1 | awk '{print $10}')
        INITIAL_RX=${INITIAL_RX:-0}
        INITIAL_TX=${INITIAL_TX:-0}
    else
        INITIAL_RX=0
        INITIAL_TX=0
    fi
    
    while true; do
        # Get CPU usage (Linux)
        if command -v mpstat &> /dev/null; then
            CPU=$(mpstat 1 1 2>/dev/null | tail -1 | awk '{print 100 - $NF}' || echo "N/A")
        else
            CPU=$(top -bn1 2>/dev/null | grep "Cpu(s)" | awk '{print $2}' || echo "N/A")
        fi
        
        # Get memory usage (Linux)
        if [[ -f /proc/meminfo ]]; then
            MEM_TOTAL=$(grep MemTotal /proc/meminfo | awk '{print $2}')
            MEM_FREE=$(grep MemAvailable /proc/meminfo | awk '{print $2}')
            MEM_USED=$(( (MEM_TOTAL - MEM_FREE) / 1024 ))
        else
            MEM_USED="N/A"
        fi
        
        # Get network usage
        if [[ -f /proc/net/dev ]]; then
            CURRENT_RX=$(grep -E '^\s*eth0:|^\s*ens|^\s*wlan|^\s*lo:' /proc/net/dev | head -1 | awk '{print $2}')
            CURRENT_TX=$(grep -E '^\s*eth0:|^\s*ens|^\s*wlan|^\s*lo:' /proc/net/dev | head -1 | awk '{print $10}')
            CURRENT_RX=${CURRENT_RX:-0}
            CURRENT_TX=${CURRENT_TX:-0}
            NET_RX=$(( (CURRENT_RX - INITIAL_RX) / 1024 ))
            NET_TX=$(( (CURRENT_TX - INITIAL_TX) / 1024 ))
        else
            NET_RX="N/A"
            NET_TX="N/A"
        fi
        
        echo "$(date +%H:%M:%S), ${CPU}%, ${MEM_USED}MB, ${NET_RX}KB, ${NET_TX}KB" >> "$METRICS_FILE"
        sleep 5
    done
) &
MONITOR_PID=$!

# Function to cleanup background processes
cleanup() {
    echo ""
    echo "Stopping resource monitoring..."
    kill $MONITOR_PID 2>/dev/null || true
}
trap cleanup EXIT

# Run load test
echo "[3/4] Running load test ($DURATION)..."
echo ""

LOADTEST_ARGS="--server $SERVER --clients $CLIENTS --duration $DURATION --min-latency $MIN_LATENCY --max-latency $MAX_LATENCY"
if $VERBOSE; then
    LOADTEST_ARGS="$LOADTEST_ARGS --verbose"
fi

# Capture load test output
LOADTEST_OUTPUT=$(mktemp)
$LOADTEST_BIN $LOADTEST_ARGS 2>&1 | tee "$LOADTEST_OUTPUT"
LOADTEST_EXIT_CODE=${PIPESTATUS[0]}

# Stop monitoring
kill $MONITOR_PID 2>/dev/null || true
trap - EXIT

# Generate report
echo ""
echo "[4/4] Generating capacity report..."

# Calculate resource statistics from metrics file
if [[ -f "$METRICS_FILE" ]]; then
    AVG_CPU=$(grep -v '^#' "$METRICS_FILE" | awk -F', ' '{gsub(/%/,"",$2); sum+=$2; count++} END {if(count>0) printf "%.1f", sum/count; else print "N/A"}')
    MAX_CPU=$(grep -v '^#' "$METRICS_FILE" | awk -F', ' '{gsub(/%/,"",$2); if($2>max) max=$2} END {printf "%.1f", max}')
    AVG_MEM=$(grep -v '^#' "$METRICS_FILE" | awk -F', ' '{gsub(/MB/,"",$3); sum+=$3; count++} END {if(count>0) printf "%.0f", sum/count; else print "N/A"}')
    MAX_MEM=$(grep -v '^#' "$METRICS_FILE" | awk -F', ' '{gsub(/MB/,"",$3); if($3>max) max=$3} END {printf "%.0f", max}')
else
    AVG_CPU="N/A"
    MAX_CPU="N/A"
    AVG_MEM="N/A"
    MAX_MEM="N/A"
fi

# Parse load test results
SUCCESSFUL_CLIENTS=$(grep "Successful Clients:" "$LOADTEST_OUTPUT" | awk '{print $3}' | head -1)
TOTAL_RECONNECTS=$(grep "Total Reconnects:" "$LOADTEST_OUTPUT" | awk '{print $3}' | head -1)
TOTAL_ERRORS=$(grep "Total Errors:" "$LOADTEST_OUTPUT" | awk '{print $3}' | head -1)
MESSAGES_SENT=$(grep "Messages Sent:" "$LOADTEST_OUTPUT" | awk '{print $3}' | head -1)
TEST_RESULT=$(grep -E "LOAD TEST (PASSED|FAILED)" "$LOADTEST_OUTPUT" | head -1)

cat > "$REPORT_FILE" << EOF
# Load Test Report

**Date:** $(date '+%Y-%m-%d %H:%M:%S')  
**Duration:** $DURATION  
**Clients:** $CLIENTS  
**Server:** $SERVER  
**Latency Range:** $MIN_LATENCY - $MAX_LATENCY  

## Test Results

| Metric | Value |
|--------|-------|
| Successful Clients | ${SUCCESSFUL_CLIENTS:-N/A} / $CLIENTS |
| Total Reconnects | ${TOTAL_RECONNECTS:-N/A} |
| Total Errors | ${TOTAL_ERRORS:-N/A} |
| Messages Sent | ${MESSAGES_SENT:-N/A} |
| Test Result | ${TEST_RESULT:-Unknown} |

## Resource Usage

| Resource | Average | Peak |
|----------|---------|------|
| CPU | ${AVG_CPU}% | ${MAX_CPU}% |
| Memory | ${AVG_MEM} MB | ${MAX_MEM} MB |

## Capacity Analysis

Based on this test with $CLIENTS concurrent clients over $DURATION:

- **CPU Headroom:** $(echo "$MAX_CPU" | awk '{if($1<50) print "Good (>50% available)"; else if($1<80) print "Moderate (20-50% available)"; else print "Limited (<20% available)"}')
- **Memory Headroom:** $(echo "$MAX_MEM" | awk '{if($1<250) print "Good (<50% of 500MB budget)"; else if($1<400) print "Moderate (50-80% of budget)"; else print "Limited (>80% of budget)"}')
- **Stability:** $(if [[ "${SUCCESSFUL_CLIENTS:-0}" == "$CLIENTS" ]]; then echo "All clients maintained connection"; else echo "Some clients experienced disconnections"; fi)

## Recommendations

$(if [[ "$MAX_CPU" != "N/A" ]] && (( $(echo "$MAX_CPU < 50" | bc -l 2>/dev/null || echo 0) )); then
    echo "- **Scale Up:** System can likely handle $(( CLIENTS * 2 )) concurrent clients"
elif [[ "$MAX_CPU" != "N/A" ]] && (( $(echo "$MAX_CPU > 80" | bc -l 2>/dev/null || echo 0) )); then
    echo "- **Scale Down:** Consider reducing max clients or optimizing server"
else
    echo "- **Current Load:** System is operating within acceptable parameters"
fi)

## Raw Metrics

See \`metrics_${TIMESTAMP}.txt\` for detailed time-series data.

---

*Generated by scripts/load-test.sh*
EOF

echo ""
echo "=== Load Test Complete ==="
echo "Report saved to: $REPORT_FILE"
echo "Metrics saved to: $METRICS_FILE"
echo ""

# Cleanup temp file
rm -f "$LOADTEST_OUTPUT"

# Exit with load test result
exit $LOADTEST_EXIT_CODE
