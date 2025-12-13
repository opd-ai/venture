# Network Package

Multiplayer networking for Venture with binary serialization, client-server communication, state synchronization, client-side prediction, entity interpolation, and lag compensation. Supports high-latency connections (200-5000ms) with authoritative server architecture.

## Features

- **Binary Protocol**: Efficient serialization (<0.5µs encode/decode, ~80 bytes per update)
- **Client-Side Prediction**: Immediate response with server reconciliation
- **Entity Interpolation**: Smooth remote entity movement
- **Lag Compensation**: Server-side rewind for fair hit detection (10ms-5000ms)
- **Buffer Monitoring**: Real-time channel utilization with automatic warnings
- **Delta Compression**: Bandwidth-efficient state synchronization
- **Thread-Safe**: Concurrent client/server operations

## Quick Start

### Server

```go
config := network.DefaultServerConfig()
config.Address = ":8080"
config.MaxPlayers = 32
server := network.NewServer(config)
server.Start()
defer server.Stop()

// Handle inputs and broadcast state updates
go func() {
    for cmd := range server.ReceiveInputCommand() {
        // Process input...
    }
}()
```

### Client

```go
config := network.DefaultClientConfig()
config.ServerAddress = "localhost:8080"
client := network.NewClient(config)
client.Connect()
defer client.Disconnect()

// Handle state updates
go func() {
    for update := range client.ReceiveStateUpdate() {
        applyUpdate(world, update)
    }
}()

client.SendInput("move", encodeMovement(dx, dy))
```

## Architecture

| Layer | Purpose |
|-------|---------|
| Protocol | Binary encoding/decoding (StateUpdate, InputCommand) |
| Client | Connection management, input queuing, latency measurement |
| Server | Multi-client handling, broadcast/unicast, player limits |
| Prediction | Client-side movement prediction, server reconciliation |
| Synchronization | Snapshot buffer, entity interpolation, delta compression |
| Lag Compensation | Historical state recording, hit validation with rewind |

## Configuration

```go
// Client
config := network.ClientConfig{
    ServerAddress:     "localhost:8080",
    ConnectionTimeout: 10 * time.Second,
    BufferSize:        256,
}

// Server  
config := network.ServerConfig{
    Address:    ":8080",
    MaxPlayers: 32,
    UpdateRate: 20,
    BufferSize: 256,
}
```

## Performance

| Operation | Time | Throughput |
|-----------|------|------------|
| StateUpdate encode | ~0.4µs | 2.2M ops/sec |
| StateUpdate decode | ~0.6µs | 1.7M ops/sec |
| InputCommand encode | ~0.2µs | 4.7M ops/sec |
| InputCommand decode | ~0.3µs | 3.6M ops/sec |

**Bandwidth** (20 updates/sec, 32 players):
- Downstream: ~51 KB/s per client
- Upstream: ~0.7 KB/s per client
- Total: ~52 KB/s (within 100 KB/s target)

## Wire Protocol

All messages use length-prefixed framing: `[4 bytes: length][N bytes: data]`

| Message | Format |
|---------|--------|
| StateUpdate | timestamp (8) + entityID (8) + priority (1) + sequence (4) + components |
| InputCommand | playerID (8) + timestamp (8) + sequence (4) + type + data |

## Component Serialization

Supported components via `ComponentSerializer`:
- Position, Velocity (16 bytes each)
- Health, Stats, Level
- Team, Input, Attack, Item

## Thread Safety

- **Client/Server**: Thread-safe for concurrent operations
- **Protocol**: Stateless, safe for concurrent use

## Testing

```bash
go test -tags test ./pkg/network/           # Run tests
go test -tags test -cover ./pkg/network/    # With coverage
go test -tags test -bench=. ./pkg/network/  # Benchmarks
```

Coverage: 82.6%

## Design Rationale

- **Binary Protocol**: 10x faster than JSON, 50% smaller packets
- **Length-Prefixed**: Simple, efficient, reliable framing
- **Authoritative Server**: Prevents cheating, single source of truth
- **Go Channels**: Natural async pattern, thread-safe, testable

## References

- [Gaffer on Games - Multiplayer](https://gafferongames.com/)
- [Source Engine Networking](https://developer.valvesoftware.com/wiki/Source_Multiplayer_Networking)
- [Gabriel Gambetta - Client-Server Architecture](https://www.gabrielgambetta.com/)
