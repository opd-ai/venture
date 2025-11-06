# Multiplayer Networking Guide

This guide covers multiplayer networking in Venture, including standard and high-latency configurations for Tor/onion services.

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
# See "Using Tor Configuration" section below
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

**Use Case:** Tor/I2P networks, satellite internet, congested/lossy networks, intercontinental connections

## Configuration Details

### Why High-Latency Configuration?

Standard network timeouts are designed for fast connections and will cause premature disconnections over high-latency networks:

1. **Read Timeout (60s vs 10s):** Prevents timeout when packets are delayed by network conditions. At 5000ms latency, a round-trip takes 10 seconds - the default 10s timeout would disconnect during normal gameplay.

2. **Write Timeout (30s vs 5s):** Allows write operations to complete over slow links. Critical for large state updates.

3. **Connection Timeout (60s vs 10s):** Tor circuit building can take 15-30 seconds. Default timeout would fail before connection establishes.

4. **Buffer Size (512 vs 256):** At 20Hz update rate with 10s round-trip, approximately 200 messages are in-flight. 512-message buffer provides 56% headroom for latency spikes.

5. **TCP Keepalive (30s):** Prevents silent disconnections at NAT/proxy boundaries (typical 5-15 minute timeout). Essential for long-duration Tor connections.

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
3. Check firewall/NAT timeout settings (ensure >60s for TCP)

### High Latency Warnings

**Symptoms:** "High latency detected" messages in logs

**Solutions:**
1. Verify using TorClientConfig (MaxLatency: 5000ms vs 500ms)
2. Check network path (traceroute, ping)
3. Consider this normal for Tor connections (2000-5000ms typical)

### Connection Timeouts

**Symptoms:** "Connection timeout" error during connect

**Solutions:**
1. Increase ConnectionTimeout in client config (60s for Tor)
2. Verify server is running and accessible
3. For Tor: ensure circuit building completes (can take 30s+)

### Message Drops

**Symptoms:** Rubber-banding, entity teleportation

**Solutions:**
1. Increase BufferSize (512 for high-latency)
2. Check network packet loss (should be <5%)
3. Verify server update rate matches client expectations

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
// SnapshotBufferSize: 200 (10s history at 20Hz)
```

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

## See Also

- [Network Package Documentation](../pkg/network/doc.go)
- [PLAN.md Network Audit](PLAN.md) - Technical analysis and rationale
- [Server CLI Reference](../cmd/server/main.go)
- [Client CLI Reference](../cmd/client/main.go)
