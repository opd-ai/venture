# Multiplayer Networking Guide

This guide covers multiplayer networking in Venture, including standard and high-latency configurations for Tor/onion services.

**For complete Tor/onion service setup instructions, see [TOR_SETUP.md](./TOR_SETUP.md).**

## Quick Start

### Standard Multiplayer (LAN/Internet)

**Start a server:**
```bash
./venture-server --port 8080 --max-players 4
```

**Connect clients:**
```bash
./venture-client -multiplayer -server <host>:8080
```

### High-Latency Multiplayer (Tor/Onion Services)

**Start a server with high-latency support:**
```bash
./venture-server --high-latency --port 8080 --max-players 4
```

**Connect clients through Tor:**
```bash
# Configure client code to use TorClientConfig
# See TOR_SETUP.md for complete setup instructions including:
# - Installing and configuring Tor
# - Setting up hidden services
# - SOCKS5 proxy configuration
# - Security best practices
```

## Network Configurations

Venture provides two network configuration presets optimized for different environments:

### Default Configuration

Optimized for typical LAN and internet connections with <100ms latency:

| Setting | Server | Client |
|---------|--------|--------|
| Read Timeout | 10 seconds | N/A |
| Write Timeout | 5 seconds | N/A |
| Connection Timeout | N/A | 10 seconds |
| Max Latency | N/A | 500ms |
| Ping Interval | N/A | 1 second |
| Buffer Size | 256 messages | 256 messages |
| TCP Keepalive | 30 seconds | 30 seconds |

**Use Case:** Local area networks, cable/fiber internet, most cloud gaming scenarios

### High-Latency Configuration

Optimized for high-latency connections (200-5000ms), including Tor/onion services:

| Setting | Server | Client |
|---------|--------|--------|
| Read Timeout | 60 seconds | N/A |
| Write Timeout | 30 seconds | N/A |
| Connection Timeout | N/A | 60 seconds |
| Max Latency | N/A | 5000ms |
| Ping Interval | N/A | 5 seconds |
| Buffer Size | 512 messages | 512 messages |
| TCP Keepalive | 30 seconds | 30 seconds |
| Lag Compensation Buffer | 300 snapshots (15s) | N/A |
| Prediction History | N/A | 256 states (12.8s) |
| Delta Compression Epsilon | 0.01 (bandwidth-optimized) | 0.01 (bandwidth-optimized) |

**Use Case:** Tor/I2P networks, satellite internet, congested/lossy networks, intercontinental connections

## Configuration Details

### Why High-Latency Configuration?

Standard network timeouts are designed for fast connections and will cause premature disconnections over high-latency networks:

1. **Read Timeout (60s vs 10s):** Prevents timeout when packets are delayed by network conditions. At 5000ms latency, a round-trip takes 10 seconds - the default 10s timeout would disconnect during normal gameplay.

2. **Write Timeout (30s vs 5s):** Allows write operations to complete over slow links. Critical for large state updates.

3. **Connection Timeout (60s vs 10s):** Tor circuit building can take 15-30 seconds. Default timeout would fail before connection establishes.

4. **Buffer Size (512 vs 256):** At 20Hz update rate with 10s round-trip, approximately 200 messages are in-flight. 512-message buffer provides 56% headroom for latency spikes.

5. **TCP Keepalive (30s):** Prevents silent disconnections at NAT/proxy boundaries (typical 5-15 minute timeout). Essential for long-duration Tor connections.

6. **Lag Compensation Buffer (300 vs 200 snapshots):** At 20Hz update rate, 300 snapshots = 15 seconds of history. Required for 5000ms max compensation with 10s round-trip time plus safety margin.

7. **Prediction History (256 vs 128 states):** At 20Hz update rate, 256 states = 12.8 seconds of history. Adequate for high-latency scenarios with 10s round-trip time.

8. **Delta Compression Epsilon (0.01 vs 0.001):** Higher epsilon (10x less sensitive) reduces bandwidth by filtering out small entity movements. Trade-off between bandwidth efficiency and position accuracy.

### Buffer Sizing Formulas

Understanding the rationale behind buffer sizes and timeout values:

**Message Buffer Calculation:**

```
Update Rate: 20 Hz (50ms per update)
Latency: L milliseconds (one-way)
Round-Trip Time: RTT = 2 * L

Messages in-flight = (RTT / 1000) * Update Rate
                   = (2 * L / 1000) * 20
                   = L / 25

For 5000ms latency:
  Messages in-flight = 5000 / 25 = 200 messages

Buffer size = Messages in-flight * Safety Factor
            = 200 * 2.56 = 512 messages (56% headroom)
```

**Timeout Calculations:**

```
Read Timeout = RTT + Processing Time + Safety Margin
             = (2 * L) + 5s + 5s
For L=5000ms: Read Timeout = 10s + 5s + 5s = 20s minimum
Configured:   Read Timeout = 60s (3x safety margin)

Write Timeout = L + Queue Time + Safety Margin
              = L + 10s + 10s
For L=5000ms: Write Timeout = 5s + 10s + 10s = 25s minimum
Configured:   Write Timeout = 30s (20% safety margin)

Connection Timeout = Circuit Building + Handshake + Safety
For Tor:           = 30s (avg circuit) + 10s + 20s = 60s
```

**Snapshot Buffer Calculation:**

```
Max Compensation: C milliseconds
Round-Trip Time: RTT = 2 * C (worst case)
Update Rate: 20 Hz

Required history = (RTT / 1000) * Update Rate + Safety
                 = ((2 * C) / 1000) * 20 + 20%
For C=5000ms:    = (10s) * 20 + 20% = 200 + 40 = 240 snapshots
Configured:      = 300 snapshots (25% safety margin, 15s history)
```

**Prediction History Calculation:**

```
Round-Trip Time: RTT = 2 * L
Update Rate: 20 Hz

Required states = (RTT / 1000) * Update Rate + Safety
                = ((2 * L) / 1000) * 20 + 28%
For L=5000ms:   = (10s) * 20 + 28% = 200 + 56 = 256 states
Configured:     = 256 states (28% safety margin, 12.8s history)
```

**General Formula for Custom Latency:**

```go
// For custom latency L (milliseconds):
bufferSize := int(float64(L) / 25 * 2.56)           // Message buffer
readTimeout := time.Duration(2*L + 10000) * time.Millisecond  // Read timeout
writeTimeout := time.Duration(L + 20000) * time.Millisecond   // Write timeout
snapshotBuffer := int(float64(L) / 50 * 20 * 1.5)  // Snapshot buffer
predictionStates := int(float64(L) / 50 * 20 * 1.28) // Prediction history
```

### Server-Side Setup

**Default server:**
```bash
./venture-server --port 8080 --max-players 4
```

**High-latency server:**
```bash
./venture-server --high-latency --port 8080 --max-players 4
```

**Custom configuration in code:**
```go
import "github.com/opd-ai/venture/pkg/network"

// Standard configuration
config := network.DefaultServerConfig()
config.Address = ":8080"
config.MaxPlayers = 4

// High-latency configuration
config := network.HighLatencyServerConfig()
config.Address = ":8080"
config.MaxPlayers = 4
```

### Client-Side Setup

**Using Default Configuration:**
```go
import "github.com/opd-ai/venture/pkg/network"

config := network.DefaultClientConfig()
config.ServerAddress = "example.com:8080"
client := network.NewClient(config)
```

**Using Tor Configuration:**
```go
import "github.com/opd-ai/venture/pkg/network"

config := network.TorClientConfig()
config.ServerAddress = "exampleonion123456.onion:8080"
client := network.NewClient(config)
```

## Performance Characteristics

### Latency Tolerance

| Configuration | Min Latency | Max Latency | Tolerance |
|--------------|-------------|-------------|-----------|
| Default | 1ms | 500ms | Good |
| High-Latency | 1ms | 5000ms | Excellent |

### Bandwidth Usage

- **Update Rate:** 20 updates/second (50ms intervals)
- **Per-Player Bandwidth:** <100 KB/s (typical), <150 KB/s (peak)
- **Delta Compression:** Only changed entities transmitted
- **Spatial Culling:** Only visible/nearby entities sent

### Connection Stability

| Metric | Default Config | High-Latency Config |
|--------|---------------|---------------------|
| Connection Success Rate | >99% | >95% |
| Disconnection Rate | <0.1/hour | <1/hour |
| Message Drop Rate | <0.1% | <2% |
| Prediction Correction Rate | <1% | <10% |

## Troubleshooting

### Frequent Disconnections

**Symptoms:** Client disconnects every few minutes

**Solutions:**
1. Enable high-latency mode on server: `--high-latency`
2. Use TorClientConfig if connecting through Tor
3. Enable automatic reconnection: `client.ConnectWithRetry(network.TorReconnectConfig())`
4. Check firewall/NAT timeout settings (ensure >60s for TCP)
5. Monitor buffer statistics for congestion: `client.GetBufferStats()`

### High Latency Warnings

**Symptoms:** "High latency detected" messages in logs

**Solutions:**
1. Verify using TorClientConfig (MaxLatency: 5000ms vs 500ms)
2. Check network path (traceroute, ping)
3. Consider this normal for Tor connections (2000-5000ms typical)
4. Ensure lag compensation configured: `HighLatencyLagCompensationConfig()`

### Connection Timeouts

**Symptoms:** "Connection timeout" error during connect

**Solutions:**
1. Increase ConnectionTimeout in client config (60s for Tor)
2. Use ConnectWithRetry with TorReconnectConfig (allows circuit rebuilding)
3. Verify server is running and accessible
4. For Tor: ensure circuit building completes (can take 30s+)

### Message Drops

**Symptoms:** Rubber-banding, entity teleportation, "buffer full" warnings

**Solutions:**
1. Check buffer statistics: `client.GetBufferStats()` or `server.GetBufferStats()`
2. Increase BufferSize (512 for high-latency)
3. Check network packet loss (should be <5%)
4. Verify server update rate matches client expectations
5. Consider increasing DeltaCompressionEpsilon to reduce bandwidth

### Failed Reconnection Attempts

**Symptoms:** Client fails to reconnect after network interruption

**Solutions:**
1. Verify using ConnectWithRetry instead of Connect
2. Increase MaxRetries in ReconnectConfig (10 for Tor)
3. Increase MaxDelay to allow longer circuit rebuilding (120s for Tor)
4. Check logs for specific connection errors
5. Verify server is reachable (firewall, network path)

## Testing High-Latency Connections

### Using Linux Traffic Control (tc)

Simulate high latency on loopback interface:

```bash
# Add 5000ms delay
sudo tc qdisc add dev lo root netem delay 5000ms

# Start high-latency server
./venture-server --high-latency --port 8080

# Connect client
./venture-client -multiplayer -server localhost:8080

# Remove delay when done
sudo tc qdisc del dev lo root
```

### Using Tor for Real-World Testing

1. Set up a Tor hidden service pointing to your server port
2. Start server with `--high-latency`
3. Configure client with TorClientConfig
4. Connect through `.onion` address

Expected latency: 2000-5000ms (sometimes higher)

## Priority-Based Message Handling

Venture's network system uses priority-based message handling to ensure critical game events are delivered before less important updates, especially during high network load or bandwidth-limited scenarios.

### Priority Levels

Messages are assigned priorities from 0 (lowest) to 255 (highest). Four standard priority levels are defined:

| Priority Level | Value | Use Cases | Examples |
|----------------|-------|-----------|----------|
| **Critical** | 255 | Game-critical state changes | Player death, revival, quest completion |
| **High** | 200 | Important gameplay events | Combat hits, damage taken, item pickups |
| **Normal** | 128 | Regular entity updates | Position/velocity updates, basic state |
| **Low** | 64 | Cosmetic changes | Animation state, particle effects, audio cues |

### How It Works

When the server's send buffer has multiple messages queued:

1. Messages are stored in a priority queue (heap-based)
2. Higher priority messages are sent first
3. Lower priority messages wait or are dropped if buffer is full
4. This ensures critical events (deaths, revivals) are never delayed by cosmetic updates

**Example:** During intense combat with many particle effects:
- Player death message (Priority: 255) sent immediately
- Combat damage (Priority: 200) sent next
- Position updates (Priority: 128) sent after
- Particle animations (Priority: 64) sent last or dropped if buffer full

### Using Priority Levels in Code

**Server-side (creating state updates):**

```go
import "github.com/opd-ai/venture/pkg/network"

// Critical: Player death event
deathUpdate := network.NewCriticalUpdate(playerEntityID)
deathUpdate.Components = []network.ComponentData{
    {Type: "death", Data: deathData},
}
server.BroadcastStateUpdate(deathUpdate)

// High: Combat damage
damageUpdate := network.NewHighPriorityUpdate(targetEntityID)
damageUpdate.Components = []network.ComponentData{
    {Type: "health", Data: healthData},
}
server.BroadcastStateUpdate(damageUpdate)

// Normal: Regular position update
posUpdate := network.NewNormalUpdate(entityID)
posUpdate.Components = []network.ComponentData{
    {Type: "position", Data: posData},
}
server.BroadcastStateUpdate(posUpdate)

// Low: Animation state
animUpdate := network.NewLowPriorityUpdate(entityID)
animUpdate.Components = []network.ComponentData{
    {Type: "animation", Data: animData},
}
server.BroadcastStateUpdate(animUpdate)

// Custom priority (0-255)
customUpdate := network.NewStateUpdate(entityID, 180)
```

### Priority Guidelines

**Use Critical (255) for:**
- Player death and revival
- Quest completion/failure
- Game state transitions (level complete, game over)
- Critical UI state (inventory loss, skill reset)

**Use High (200) for:**
- Combat damage and healing
- Item pickups and drops
- Skill activations and cooldowns
- NPC dialogue triggers

**Use Normal (128) for:**
- Entity position and velocity
- Health/mana regeneration
- Generic entity state
- Movement commands

**Use Low (64) for:**
- Animation state changes
- Particle effect updates
- Audio triggers (footsteps, ambient)
- Visual-only effects

### Performance Impact

Priority-based handling has minimal overhead:

- **Push to queue:** ~53ns per message (1 allocation)
- **Pop from queue:** ~25ns per message (0 allocations)
- **Memory:** No additional per-message overhead
- **CPU:** O(log n) insertion/extraction (heap-based)

### Benefits

1. **Reliability:** Critical events never lost during bandwidth limits
2. **User Experience:** Deaths/revivals always communicated instantly
3. **Scalability:** Graceful degradation under network stress
4. **Fairness:** All clients receive critical updates equally fast
5. **Optimization:** Cosmetic updates can be dropped without gameplay impact

## Advanced Topics

### Custom Timeout Configuration

For specific network conditions, you can customize timeouts:

```go
config := network.DefaultServerConfig()
config.ReadTimeout = 30 * time.Second  // Custom read timeout
config.WriteTimeout = 15 * time.Second // Custom write timeout
config.BufferSize = 384                // Custom buffer size
```

### Lag Compensation Tuning

High-latency mode automatically uses extended lag compensation:

```go
// Automatically selected when using --high-latency flag
lagCompConfig := network.HighLatencyLagCompensationConfig()
// MaxCompensation: 5000ms
// SnapshotBufferSize: 300 (15s history at 20Hz)
// DeltaCompressionEpsilon: 0.01 (10x less sensitive for bandwidth efficiency)
```

### Automatic Reconnection

Venture supports automatic reconnection with exponential backoff to recover from transient network failures. This is critical for unstable connections like Tor circuits.

**Configuration Presets:**

```go
// Default reconnection for standard networks
config := network.DefaultReconnectConfig()
// MaxRetries: 5
// InitialDelay: 1 second
// MaxDelay: 30 seconds (caps exponential growth)
// BackoffFactor: 2.0

// Tor reconnection for high-latency networks
config := network.TorReconnectConfig()
// MaxRetries: 10 (more aggressive for unstable Tor circuits)
// InitialDelay: 5 seconds (longer for circuit stability)
// MaxDelay: 120 seconds (2 minutes for Tor circuit rebuilding)
// BackoffFactor: 2.0
```

**Using Reconnection:**

```go
import "github.com/opd-ai/venture/pkg/network"

// Create client with default config
clientConfig := network.DefaultClientConfig()
clientConfig.ServerAddress = "example.com:8080"
client := network.NewClient(clientConfig)

// Connect with automatic retry
reconnectConfig := network.DefaultReconnectConfig()
err := client.ConnectWithRetry(reconnectConfig)
if err != nil {
    log.Fatalf("Failed to connect after %d retries: %v", reconnectConfig.MaxRetries, err)
}

// For Tor connections, use Tor-specific configs
clientConfig = network.TorClientConfig()
clientConfig.ServerAddress = "exampleonion123456.onion:8080"
client = network.NewClient(clientConfig)

reconnectConfig = network.TorReconnectConfig()
err = client.ConnectWithRetry(reconnectConfig)
```

**Exponential Backoff Behavior:**

The reconnection system uses exponential backoff to avoid overwhelming the server or network:

1. First retry after `InitialDelay` (1s default, 5s Tor)
2. Second retry after `InitialDelay * BackoffFactor` (2s default, 10s Tor)
3. Third retry after `InitialDelay * BackoffFactor^2` (4s default, 20s Tor)
4. Continues until `MaxDelay` is reached (30s default, 120s Tor)
5. All subsequent retries use `MaxDelay`
6. Stops after `MaxRetries` attempts

**Benefits:**

- Automatic recovery from network hiccups
- Graceful handling of Tor circuit rebuilding (15-60 seconds)
- Prevents connection storms during server restarts
- Configurable retry strategy per network type

### Buffer Monitoring

Venture includes comprehensive buffer monitoring to detect and diagnose network congestion in real-time.

**Buffer Statistics Tracked:**

Each network channel tracks:
- **Current Size**: Number of messages currently buffered
- **Capacity**: Maximum buffer capacity
- **Sent**: Total messages successfully sent
- **Dropped**: Messages dropped due to full buffer
- **Utilization**: Current size / capacity (0.0-1.0)
- **Drop Rate**: Dropped / (sent + dropped)

**Warning Thresholds:**

Automatic warnings are logged when buffer utilization exceeds 80%:

```
WARN buffer utilization high buffer=state_updates size=410 capacity=512 utilization=0.80
```

**Accessing Buffer Statistics:**

**Client-side:**

```go
client := network.NewClient(config)
// ... after connecting ...

// Get all buffer stats
stats := client.GetBufferStats()

// Check specific buffers
stateUpdateStats := stats["state_updates"]
fmt.Printf("State updates: %d/%d (%.1f%% utilization, %.2f%% drop rate)\n",
    stateUpdateStats.CurrentSize,
    stateUpdateStats.Capacity,
    stateUpdateStats.Utilization * 100,
    stateUpdateStats.DropRate * 100)

inputQueueStats := stats["input_queue"]
errorStats := stats["errors"]
```

**Server-side:**

```go
server := network.NewServer(config)
// ... after starting ...

// Get all buffer stats (includes per-client buffers)
stats := server.GetBufferStats()

// Check global buffers
inputStats := stats["input_commands"]
joinStats := stats["player_joins"]
leaveStats := stats["player_leaves"]
errorStats := stats["errors"]

// Check per-client buffers (named "state_updates_<clientID>")
for name, snapshot := range stats {
    if strings.HasPrefix(name, "state_updates_") {
        fmt.Printf("Client %s: %d/%d messages (%.1f%% utilization)\n",
            name, snapshot.CurrentSize, snapshot.Capacity,
            snapshot.Utilization * 100)
    }
}
```

**Interpreting Results:**

| Utilization | Meaning | Action |
|-------------|---------|--------|
| <50% | Healthy | No action needed |
| 50-79% | Moderate load | Monitor for trends |
| 80-94% | High load | **Warning logged**, investigate |
| 95-99% | Critical | Message drops likely imminent |
| 100% | Full | Messages being dropped |

**When to Act:**

- **Persistent 80%+ utilization**: Increase buffer size or reduce update rate
- **High drop rate (>1%)**: Network cannot keep up, adjust configuration
- **Specific buffer full**: Investigate that subsystem (e.g., state updates vs input queue)

### Delta Compression Optimization

Delta compression controls bandwidth vs accuracy tradeoff by filtering out small entity changes.

**Epsilon Values:**

The `DeltaCompressionEpsilon` parameter controls change sensitivity:

- **0.001 (Default)**: High sensitivity, high accuracy, higher bandwidth
  - Detects position changes >0.001 units (~1/1000 of a tile)
  - Best for low-latency (<100ms) connections
  - Smooth, precise entity movement

- **0.01 (High-Latency)**: Lower sensitivity, bandwidth-optimized, slight accuracy loss
  - Detects position changes >0.01 units (~1/100 of a tile)
  - Best for high-latency (>500ms) connections
  - Reduces bandwidth by ~10-30% (varies by scenario)
  - Small movements filtered out, less jitter transmission

**Configuration:**

```go
// Lag compensation configs automatically set appropriate epsilon
defaultConfig := network.DefaultLagCompensationConfig()
// DeltaCompressionEpsilon: 0.001 (high sensitivity)

highLatencyConfig := network.HighLatencyLagCompensationConfig()
// DeltaCompressionEpsilon: 0.01 (10x less sensitive)

// Custom epsilon for specific needs
customConfig := network.LagCompensationConfig{
    MaxCompensation:         2000 * time.Millisecond,
    MinCompensation:         10 * time.Millisecond,
    SnapshotBufferSize:      200,
    DeltaCompressionEpsilon: 0.005, // Custom value between default and high-latency
}
```

**Performance Impact:**

Example: 100 entities on screen

| Epsilon | Updates Sent/Sec | Bandwidth Saved | Accuracy |
|---------|------------------|-----------------|----------|
| 0.001 | ~400 | 0% (baseline) | Excellent |
| 0.005 | ~320 | ~20% | Very Good |
| 0.01  | ~280 | ~30% | Good |
| 0.05  | ~180 | ~55% | Fair |

**When to Use:**

- **0.001**: LAN, low-latency internet (<100ms), gameplay requiring precision
- **0.01**: Tor, satellite, high-latency (>500ms), bandwidth-constrained
- **Custom**: Special cases (e.g., 0.005 for intercontinental play)

**Trade-offs:**

✅ **Higher epsilon (e.g., 0.01)**:
- Lower bandwidth usage (fewer updates)
- Better performance on slow connections
- Less network congestion

⚠️ **Higher epsilon drawbacks**:
- Small movements not transmitted
- Slightly less smooth remote entity motion
- Potential for "micro-teleports" (<1/100 tile)

### Network Monitoring

Enable verbose logging to monitor network performance:

```bash
LOG_LEVEL=debug ./venture-server --high-latency --port 8080
```

Logs will show:
- Connection establishment (with keepalive status)
- Latency measurements
- Buffer utilization
- Message drop warnings

## Security Considerations

### Tor/Onion Services

**Advantages:**
- Anonymity for players and server operators
- NAT traversal without port forwarding
- Censorship resistance

**Considerations:**
- Higher latency (2000-5000ms typical)
- Connection stability depends on Tor network health
- Circuit building can take 15-60 seconds

### Firewall Configuration

**Required Ports:**
- Default: TCP 8080 (or custom `--port`)
- Must allow bidirectional TCP traffic
- TCP keepalive packets (every 30s)

**Recommended:**
- Rate limiting to prevent DoS attacks
- Connection limit per IP (prevent flood)
- DDoS protection for public servers

## FAQ

**Q: Can I use high-latency config for regular internet connections?**

A: Yes. High-latency config works for all connection types. The trade-off is slightly higher resource usage (larger buffers, longer timeouts). For <100ms connections, default config is more efficient.

**Q: What latency is playable?**

A: With client-side prediction and lag compensation:
- <50ms: Excellent, feels local
- 50-150ms: Good, slight delay noticeable
- 150-500ms: Acceptable, requires prediction
- 500-2000ms: Playable with high-latency config
- 2000-5000ms: Challenging but functional (Tor typical)
- >5000ms: May require custom configuration

**Q: Does high-latency mode affect server performance?**

A: Minimal impact. Slightly larger memory usage for buffers (512 vs 256 messages per client). No CPU or update rate changes.

**Q: Can I mix default and high-latency clients on same server?**

A: Yes. Server configuration applies to all clients, but each client can use appropriate config (TorClientConfig for Tor users, DefaultClientConfig for LAN users).

**Q: How do I know which config to use?**

A: Use high-latency config if:
- Connecting through Tor/I2P/onion services
- Satellite internet or mobile data
- Experiencing frequent disconnections
- Network latency >500ms
- Intercontinental connections

Use default config for:
- Local area network (LAN)
- Cable/fiber internet
- Cloud gaming platforms
- Network latency <100ms

**Q: How does automatic reconnection work?**

A: Use `client.ConnectWithRetry()` instead of `client.Connect()`. The system will:
- Retry connection on failure using exponential backoff
- Wait 1s, 2s, 4s, 8s, etc. (DefaultReconnectConfig) between attempts
- Wait 5s, 10s, 20s, 40s, etc. (TorReconnectConfig) for Tor circuits
- Cap at MaxDelay (30s default, 120s Tor) to avoid excessive waits
- Stop after MaxRetries attempts (5 default, 10 Tor)

**Q: How do I monitor network congestion?**

A: Use buffer statistics APIs:

```go
// Client-side
stats := client.GetBufferStats()
stateUpdates := stats["state_updates"]
if stateUpdates.Utilization > 0.8 {
    fmt.Printf("Warning: %.1f%% buffer utilization\n", stateUpdates.Utilization * 100)
}

// Server-side
stats := server.GetBufferStats()
// Check global and per-client buffers
```

Watch for:
- Utilization >80%: Warning threshold
- DropRate >1%: Network cannot keep up
- Check logs for "buffer utilization high" warnings

**Q: What's the difference between epsilon 0.001 and 0.01?**

A: DeltaCompressionEpsilon controls bandwidth vs accuracy:
- **0.001** (default): Transmits all movements >1/1000 tile, high accuracy, higher bandwidth
- **0.01** (high-latency): Transmits movements >1/100 tile, 10-30% bandwidth savings, slight accuracy loss

Use 0.01 for Tor/high-latency, 0.001 for LAN/low-latency.

**Q: Will the client reconnect if my Tor circuit fails?**

A: Yes, if using `ConnectWithRetry()` with `TorReconnectConfig()`. The system will:
- Detect connection failure
- Wait 5 seconds (longer for circuit stability)
- Retry up to 10 times with increasing delays
- Handle Tor circuit rebuilding (can take 15-60s)
- Max 2-minute wait between attempts

**Q: Can I see real-time network statistics while playing?**

A: Yes, enable debug logging:

```bash
LOG_LEVEL=debug ./venture-client -multiplayer -server example.com:8080
```

Logs show:
- Buffer utilization (every time it hits 80%)
- Message drops (when they occur)
- Latency measurements
- Connection/reconnection events

## Load Testing

### Multi-Client Load Testing Tool

The `loadtest` tool validates server stability under realistic network conditions with multiple concurrent clients at varying latencies.

**Basic Usage:**
```bash
# Build the tool
go build -o build/loadtest ./cmd/loadtest

# Run 20-minute test with 4 clients (50ms-5000ms latency range)
./build/loadtest --server localhost:8080 --clients 4 --duration 20m
```

**Custom Configuration:**
```bash
# Test low-latency clients only
./build/loadtest --server localhost:8080 --clients 4 --duration 10m \
  --min-latency 50ms --max-latency 500ms

# Test high-latency/Tor-like connections
./build/loadtest --server localhost:8080 --clients 4 --duration 20m \
  --min-latency 1000ms --max-latency 5000ms

# Enable verbose logging
./build/loadtest --server localhost:8080 --clients 4 --duration 5m --verbose
```

**Success Criteria:**
- All clients successfully connect
- ≥90% success rate (clients maintain ≥90% uptime)
- <10% disconnect rate
- Continuous message throughput

**Example Output:**
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

✅ LOAD TEST PASSED
```

**Documentation:** See [cmd/loadtest/README.md](../cmd/loadtest/README.md) for detailed usage and troubleshooting.

## See Also

- [Network Package Documentation](../pkg/network/doc.go)
- [PLAN.md Network Audit](PLAN.md) - Technical analysis and rationale
- [Load Testing Tool](../cmd/loadtest/README.md) - Multi-client load testing
- [Server CLI Reference](../cmd/server/main.go)
- [Client CLI Reference](../cmd/client/main.go)
