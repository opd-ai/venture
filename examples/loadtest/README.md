# Multi-Client Load Testing Tool

A comprehensive load testing tool for Venture's multiplayer server that simulates multiple clients with varying network latencies.

## Purpose

This tool validates connection stability and multiplayer server performance under realistic network conditions, including:
- Mixed latency clients (50ms to 5000ms)
- Prolonged connection stability (up to hours)
- Automatic reconnection behavior
- Message throughput tracking
- Error rate monitoring

## Usage

### Basic Test
```bash
# Build the tool
go build -o build/loadtest ./cmd/loadtest

# Run a 20-minute test with 4 clients
./build/loadtest --server localhost:8080 --clients 4 --duration 20m
```

### Custom Latency Range
```bash
# Test low-latency connections only (50-500ms)
./build/loadtest --server localhost:8080 --clients 4 --duration 10m \
  --min-latency 50ms --max-latency 500ms

# Test high-latency/Tor-like connections (1000-5000ms)
./build/loadtest --server localhost:8080 --clients 4 --duration 20m \
  --min-latency 1000ms --max-latency 5000ms
```

### Verbose Logging
```bash
# Enable detailed logging for debugging
./build/loadtest --server localhost:8080 --clients 4 --duration 5m --verbose
```

## Command-Line Options

| Flag | Default | Description |
|------|---------|-------------|
| `--server` | `localhost:8080` | Server address (host:port) |
| `--clients` | `4` | Number of concurrent clients |
| `--duration` | `20m` | Test duration (e.g., 5s, 10m, 1h) |
| `--min-latency` | `50ms` | Minimum simulated latency |
| `--max-latency` | `5000ms` | Maximum simulated latency |
| `--verbose` | `false` | Enable verbose logging |

## Output

The tool provides real-time progress updates every 30 seconds and a comprehensive final report:

### Progress Updates
```
[30s] Connected: 4/4 | Reconnects: 0 | Errors: 0
[1m0s] Connected: 4/4 | Reconnects: 1 | Errors: 2
```

### Final Report
```
=== Load Test Results ===
Test Duration: 20m0s
Total Clients: 4
Successful Clients: 4 (100.0%)
Total Reconnects: 3
Premature Disconnects: 0
Messages Sent: 48000
Messages Received: 48000
Total Errors: 3

=== Success Criteria ===
✓ All clients connected: true
✓ Success rate ≥90%: true (100.0%)
✓ Disconnect rate <10%: true (0.0%)
✓ Messages sent: true (48000 messages)

✅ LOAD TEST PASSED
```

## Success Criteria

The test passes if:
1. **All clients connected**: Every client successfully established initial connection
2. **Success rate ≥90%**: At least 90% of clients maintained ≥90% uptime
3. **Disconnect rate <10%**: Less than 10% of clients disconnected prematurely
4. **Messages sent**: Clients successfully sent messages throughout the test

## Implementation Details

### Client Simulation
- Each client simulates network behavior at its target latency
- Higher latency clients have proportionally higher failure rates
- Automatic reconnection with exponential backoff on connection loss
- 20 updates/second per client (matching server update rate)

### Latency Distribution
Clients are distributed evenly across the latency range:
- With 4 clients and 50ms-500ms range: 50ms, 200ms, 350ms, 500ms
- With 4 clients and 50ms-5000ms range: 50ms, 1700ms, 3350ms, 5000ms

### Failure Simulation
- Connection failures scale with latency (0.0001% at 50ms to 0.01% at 5000ms)
- 5% chance of initial connection failure for high-latency clients (>1000ms)
- Random connection drops during gameplay based on latency

## Testing the Tool

Run the test suite:
```bash
go test -v ./cmd/loadtest/...
```

Check test coverage:
```bash
go test -cover ./cmd/loadtest/...
```

## Integration with PLAN.md

This tool implements **Week 3, Task 4** from `docs/PLAN.md`:
- Setup: 4 clients, mixed latencies (50ms-5000ms)
- Duration: 20 minutes
- Metrics: All clients remain connected

To validate Week 3 completion, run:
```bash
./build/loadtest --server localhost:8080 --clients 4 --duration 20m
```

Expected result: ✅ LOAD TEST PASSED

## Troubleshooting

### Server Not Running
```
ERROR Failed to connect: connection timeout after 60s
```
**Solution**: Start the Venture server before running the test:
```bash
./venture-server --port 8080
```

### All Clients Failing
```
❌ LOAD TEST FAILED
Successful Clients: 0 (0.0%)
```
**Solution**: Check server logs for errors. Ensure server is configured for expected load.

### High Reconnect Rate
```
Total Reconnects: 50
```
**Solution**: This may indicate server instability or network issues. Check server performance metrics.

## Performance Expectations

Based on testing:
- **Low latency (50-500ms)**: <1% error rate, 0-2 reconnects per client
- **High latency (1000-5000ms)**: 1-5% error rate, 2-10 reconnects per client
- **Message throughput**: ~1200 messages/minute per client (20 updates/sec)

## Related Documentation

- [MULTIPLAYER.md](../../docs/MULTIPLAYER.md) - Network architecture and configuration
- [TOR_SETUP.md](../../docs/TOR_SETUP.md) - High-latency network setup
- [PLAN.md](../../docs/PLAN.md) - Implementation plan and status
