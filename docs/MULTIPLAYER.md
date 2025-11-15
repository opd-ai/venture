# Multiplayer Networking Guide

Multiplayer networking for standard and high-latency (Tor/onion) connections.

**For Tor setup, see [TOR_SETUP.md](./TOR_SETUP.md).**

## Quick Start

**Standard (LAN/Internet):**
```bash
# Server
./venture-server -port 8080 -max-players 4

# Client
./venture-client -multiplayer -server <host>:8080
```

**High-Latency (Tor/Satellite):**
```bash
# Server
./venture-server -high-latency -port 8080 -max-players 4

# Client (requires Tor configuration - see TOR_SETUP.md)
./venture-client -multiplayer -server <onion_address>:8080
```

## Network Configurations

### Default (LAN/Internet, <100ms latency)

| Setting | Server | Client |
|---------|--------|--------|
| Read/Write Timeout | 10s / 5s | - |
| Connection Timeout | - | 10s |
| Max Latency | - | 500ms |
| Ping Interval | - | 1s |
| Buffer Size | 256 msgs | 256 msgs |
| TCP Keepalive | 30s | 30s |

### High-Latency (Tor/Satellite, 200-5000ms)

| Setting | Server | Client |
|---------|--------|--------|
| Read/Write Timeout | 60s / 30s | - |
| Connection Timeout | - | 60s |
| Max Latency | - | 5000ms |
| Ping Interval | - | 5s |
| Buffer Size | 512 msgs | 512 msgs |
| TCP Keepalive | 30s | 30s |
| Lag Compensation | 300 snapshots (15s) | - |
| Prediction History | - | 256 states (12.8s) |
| Delta Compression ε | 0.01 | 0.01 |

## Configuration Rationale

**High-latency adjustments:**
- **60s read timeout:** 5s RTT + processing + margin (vs 10s default)
- **512 msg buffer:** 200 in-flight at 5s latency + 56% headroom (vs 256)
- **60s connection:** Tor circuit building takes 15-30s (vs 10s)
- **300 snapshot buffer:** 15s history for 5s max compensation (vs 10s/200)
- **256 prediction states:** 12.8s history for 5s RTT scenarios (vs 6.4s/128)
- **0.01 epsilon:** 10x less sensitive = reduced bandwidth (vs 0.001)

**Buffer sizing formula:**
```
Messages in-flight = (RTT_ms / 1000) * UpdateRate_Hz
Buffer = Messages * SafetyFactor

Example (5000ms latency, 20 Hz):
  In-flight = (10s RTT) * 20 = 200
  Buffer = 200 * 2.56 = 512 messages
```

**Timeout formula:**
```
Read = (2 * Latency) + Processing + Margin
For 5000ms: 10s + 5s + 5s = 20s min → 60s (3x margin)

Connection = Circuit + Handshake + Safety
For Tor: 30s + 10s + 20s = 60s
```

## Code Configuration

**Server:**
```go
import "github.com/opd-ai/venture/pkg/network"

// Standard
config := network.DefaultServerConfig()
config.Address = ":8080"
config.MaxPlayers = 4

// High-latency
config := network.HighLatencyServerConfig()
config.Address = ":8080"
config.MaxPlayers = 4

server := network.NewTCPServer(config)
server.Start()
```

**Client:**
```go
// Standard
config := network.DefaultClientConfig()
config.ServerAddress = "host:8080"

// High-latency (Tor)
config := network.TorClientConfig()
config.ServerAddress = "abcdef.onion:8080"

client := network.NewTCPClient(config)
client.Connect()
```

## Architecture

**Client-Server Model:**
- Server: Authoritative game state, validates inputs, sends snapshots
- Client: Predicts local player, interpolates remote entities, reconciles on misprediction

**Networking Systems:**
- **Client-side prediction:** Immediate local response despite latency
- **Server reconciliation:** Corrects mispredictions, replays inputs
- **Entity interpolation:** Smooth remote player movement (100-200ms buffer)
- **Lag compensation:** Rewinds to client's view for hit detection
- **Delta compression:** Send only changed values (0.01 position threshold)

**Message Types:**
- Input: Client → Server (player actions)
- Snapshot: Server → Client (game state, 20 Hz)
- Ping/Pong: Latency measurement
- Connect/Disconnect: Session management

## Performance

**Bandwidth (per player):**
- Standard: ~50 KB/s at 20 Hz
- High-latency: ~100 KB/s (less compression due to 0.01 epsilon)
- 4 players: ~200-400 KB/s total

**Latency Handling:**
- <50ms: Imperceptible
- 50-200ms: Client prediction masks latency
- 200-500ms: Noticeable but playable with interpolation
- 500-5000ms: Requires high-latency config, still playable

**Update Rate:**
- Default: 20 Hz (50ms/update)
- Low bandwidth: 10 Hz (100ms/update) - reduce `UpdateRate` in config
- High quality: 30 Hz (33ms/update) - increase CPU/bandwidth

## Monitoring

**Server logs:**
```bash
tail -f venture-server.log | grep -i "buffer\|timeout\|disconnect"
```

**Client logs:**
```bash
tail -f venture-client.log | grep -i "latency\|prediction\|reconcile"
```

**Buffer stats:**
```go
stats := server.GetBufferStats()
// Check stats["state_updates"].Dropped, Utilization
```

## Troubleshooting

**Frequent disconnects:**
- Enable `-high-latency` flag
- Check buffer utilization (should be <80%)
- Reduce update rate: `config.UpdateRate = 10`

**High latency warnings:**
- Increase `MaxLatency`: `config.MaxLatency = 10000 * time.Millisecond`
- Check network: `ping <server>` for non-Tor

**Prediction errors:**
- Check client logs for reconciliation frequency
- High frequency = server-client state divergence
- May need to adjust prediction parameters

**Bandwidth issues:**
- Reduce update rate to 10 Hz
- Increase delta epsilon: `config.DeltaEpsilon = 0.1`
- Spatial culling: Send only nearby entities (already implemented)

## Security

**Server:**
- Validate all client inputs (already implemented)
- Rate limit connections (DDoS protection)
- Use Tor hidden service for IP privacy

**Client:**
- Verify server authenticity (onion address)
- Don't trust client predictions from other players
- Use encrypted connections (future: TLS over Tor)

## Advanced

**Custom latency config:**
```go
L := 3000  // 3000ms latency
config.ReadTimeout = time.Duration(2*L + 10000) * time.Millisecond  // 16s
config.WriteTimeout = time.Duration(L + 20000) * time.Millisecond   // 23s
config.BufferSize = int(float64(L) / 25 * 2.56)  // ~307 messages
```

**Monitor network:**
```go
latency := client.GetLatency()
packetLoss := client.GetPacketLoss()
bandwidth := client.GetBandwidth()
```

**State synchronization frequency:**
- Full snapshot: Every 5s (low priority)
- Delta updates: 20 Hz (positions, health)
- Critical events: Immediate (death, damage)

---

**Last Updated:** November 14, 2025
