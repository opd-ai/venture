# Buffer Monitoring System

## Overview

The network package implements comprehensive buffer monitoring to detect and warn about channel congestion before it causes message drops or performance degradation.

## Architecture

### BufferStats

Thread-safe statistics tracker for buffered channels:

```go
type BufferStats struct {
    Name     string  // Channel name for logging
    Capacity int     // Maximum channel capacity
    
    sent    uint64  // Total messages sent (atomic)
    dropped uint64  // Messages dropped (atomic)
    
    currentSize int  // Current estimated size (mutex-protected)
    // ... internal fields
}
```

### Monitoring Points

**Client Buffers:**
- `state_updates`: State updates from server (capacity: BufferSize)
- `input_queue`: Input commands to server (capacity: BufferSize)
- `errors`: Error events (capacity: 16)

**Server Buffers:**
- `input_commands`: Input commands from all clients (capacity: BufferSize × MaxPlayers)
- `player_joins`: Player connection events (capacity: MaxPlayers)
- `player_leaves`: Player disconnection events (capacity: MaxPlayers)
- `errors`: Error events (capacity: 64)
- Per-client `state_updates`: State updates per client (capacity: BufferSize)

## Usage

### Getting Buffer Statistics

**Client:**
```go
client := NewClient(DefaultClientConfig())
stats := client.GetBufferStats()

for name, snapshot := range stats {
    fmt.Printf("%s: %d/%d (%.1f%% full, %.2f%% dropped)\n",
        name,
        snapshot.CurrentSize,
        snapshot.Capacity,
        snapshot.Utilization * 100,
        snapshot.DropRate * 100,
    )
}
```

**Server:**
```go
server := NewServer(DefaultServerConfig())
stats := server.GetBufferStats()

// Server stats include per-client buffers
for name, snapshot := range stats {
    if snapshot.Utilization > 0.8 {
        fmt.Printf("WARNING: %s is %.1f%% full!\n", name, snapshot.Utilization * 100)
    }
}
```

### BufferSnapshot Fields

```go
type BufferSnapshot struct {
    Name        string  // Buffer identifier
    Capacity    int     // Maximum capacity
    CurrentSize int     // Current approximate size
    Sent        uint64  // Total messages sent
    Dropped     uint64  // Total messages dropped
    Utilization float64 // Current utilization (0.0-1.0)
    DropRate    float64 // Percentage dropped (0.0-1.0)
}
```

## Automatic Warning Logs

Warnings are logged automatically at 80% utilization:

```
level=warning msg="buffer utilization high" 
  buffer=client_state_updates 
  size=205 
  capacity=256 
  utilization=0.80 
  total_sent=1520 
  total_drops=32
```

### Spam Prevention

Warnings use spam prevention:
- Only log when utilization crosses threshold (79% → 80%)
- Only log when size increases above previous warning
- Reset warning when utilization drops below threshold

## Performance Characteristics

### Overhead

- **RecordSend/RecordDrop**: ~1-2ns (atomic increment + check)
- **RecordReceive**: ~50ns (mutex lock/unlock + decrement)
- **Snapshot**: ~100ns (read all fields with locks)
- **Warning Check**: ~10ns (comparison + optional log)

### Memory

- **Per BufferStats**: ~120 bytes
- **Client**: ~360 bytes (3 buffers)
- **Server**: ~600 bytes + ~120 bytes per client

## Thread Safety

All operations are thread-safe:
- Atomic counters for `sent` and `dropped`
- RWMutex for `currentSize` tracking
- Lock-free reads via `atomic.LoadUint64()`

## Best Practices

### Monitoring in Production

```go
// Periodic stats collection
ticker := time.NewTicker(30 * time.Second)
for range ticker.C {
    stats := server.GetBufferStats()
    for name, snapshot := range stats {
        metrics.RecordGauge("network.buffer.utilization", snapshot.Utilization, map[string]string{
            "buffer": name,
        })
        metrics.RecordCounter("network.buffer.drops", snapshot.Dropped, map[string]string{
            "buffer": name,
        })
    }
}
```

### Alert Thresholds

Recommended alert levels:
- **Info**: 60% utilization (capacity planning)
- **Warning**: 80% utilization (potential congestion)
- **Error**: 95% utilization (critical congestion)
- **Critical**: >1% drop rate (message loss occurring)

### Interpreting Metrics

**High Utilization, Low Drop Rate:**
- Normal under load
- Consider increasing buffer size if sustained

**High Drop Rate:**
- Consumer too slow
- Network congestion
- Consider backpressure or rate limiting

**Oscillating Utilization:**
- Bursty traffic pattern
- Normal for game events (combat, lots of entities)
- Consider smoothing with larger buffers

## Integration Example

```go
// Custom monitoring with alerting
func monitorBuffers(client *TCPClient) {
    ticker := time.NewTicker(5 * time.Second)
    defer ticker.Stop()
    
    for range ticker.C {
        stats := client.GetBufferStats()
        
        for name, snapshot := range stats {
            // Alert on high drop rate
            if snapshot.DropRate > 0.01 { // 1%
                log.Errorf("Buffer %s dropping messages: %.2f%% drop rate",
                    name, snapshot.DropRate * 100)
            }
            
            // Alert on sustained high utilization
            if snapshot.Utilization > 0.9 && snapshot.Sent > 100 {
                log.Warnf("Buffer %s near capacity: %.1f%% full",
                    name, snapshot.Utilization * 100)
            }
        }
    }
}
```

## Testing

### Unit Tests

See `buffer_stats_test.go`:
- Record operations (send, receive, drop)
- Utilization calculations
- Drop rate calculations
- Concurrent access
- Warning threshold behavior

### Integration Tests

See `buffer_monitoring_test.go`:
- Client buffer monitoring
- Server buffer monitoring
- Per-client state update tracking
- High utilization scenarios

## Troubleshooting

### Buffer Full Warnings

**Symptom:** Frequent "buffer utilization high" warnings

**Causes:**
1. Consumer too slow (game loop lag)
2. Producer too fast (network spike)
3. Buffer too small for workload

**Solutions:**
1. Profile consumer (game loop, rendering)
2. Implement backpressure or rate limiting
3. Increase buffer size in config:
   ```go
   config := DefaultClientConfig()
   config.BufferSize = 512  // 2x default
   ```

### Message Drops

**Symptom:** Non-zero drop rate in stats

**Causes:**
1. Sustained buffer overflow
2. Latency spikes
3. Insufficient buffer capacity

**Solutions:**
1. Use high-latency configs for slow networks:
   ```go
   config := TorClientConfig()  // 512 buffer size
   ```
2. Implement selective message dropping (deprioritize old updates)
3. Add client-side prediction to mask drops

## Implementation Notes

### Why Not Always Use RecordReceive?

RecordReceive requires a mutex lock, which adds overhead to high-frequency operations. We use it selectively:

- **Use RecordReceive**: When you control the receive (e.g., `case msg := <-ch`)
- **Skip RecordReceive**: For sends where receives are in separate goroutines (tracking approximate size is sufficient)

Current implementation tracks receives in:
- Client: `sendLoop` (inputQueue receive)
- Server: `handleClientSend` (per-client stateUpdates receive)

This provides accurate size tracking for the most critical buffers while minimizing overhead.

### Future Enhancements

Possible improvements:
1. **Adaptive buffer sizing**: Automatically grow buffers under sustained load
2. **Buffer pressure callbacks**: Custom handlers for high utilization
3. **Histogram metrics**: Track utilization distribution over time
4. **Per-message priority**: Track drops by message priority
5. **Rate limiting**: Automatic producer throttling at threshold

## References

- Network Stack Audit: `docs/PLAN.md` (Week 3)
- Buffer Pool: `pkg/network/buffer_pool.go`
- Client Implementation: `pkg/network/client.go`
- Server Implementation: `pkg/network/server.go`
