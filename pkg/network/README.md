# Network Package

The network package provides multiplayer networking functionality for Venture, including binary serialization, client-server communication, state synchronization, client-side prediction, and entity interpolation. Designed to support high-latency connections (200-5000ms) with authoritative server architecture.

## Features

- **Binary Protocol**: Efficient binary serialization for state updates and input commands
- **Client Networking**: Connection management, input sending, state receiving
- **Server Networking**: Client management, broadcasting, authoritative state
- **Client-Side Prediction**: Immediate response to player input with server reconciliation
- **Entity Interpolation**: Smooth remote entity movement between server snapshots
- **Snapshot Management**: Efficient state history with delta compression
- **Low Latency**: Optimized for real-time multiplayer (sub-millisecond serialization)
- **High Bandwidth**: Minimal packet sizes (<100 bytes typical)
- **Thread-Safe**: Concurrent client/server operations

## Architecture

### Protocol Layer

The `BinaryProtocol` implements efficient binary encoding/decoding:

- **StateUpdate**: Server → Client (entity state changes)
- **InputCommand**: Client → Server (player inputs)

### Client Layer

The `Client` manages connection to a game server:

- Async send/receive loops
- Input queuing with sequence tracking
- Latency measurement
- Automatic reconnection (future)

### Server Layer

The `Server` handles multiple client connections:

- Accept loop for new connections
- Per-client send/receive handlers
- Broadcast and unicast state updates
- Player limit enforcement

### Prediction Layer

The `ClientPredictor` implements client-side prediction:

- Predicts movement immediately for responsive controls
- Maintains history of predicted states
- Reconciles with authoritative server state
- Replays unacknowledged inputs after correction

### Synchronization Layer

The `SnapshotManager` handles state synchronization:

- Maintains circular buffer of world snapshots
- Interpolates entity positions between snapshots
- Creates delta updates for bandwidth efficiency
- Retrieves historical states for lag compensation

## Usage

### Basic Server

```go
import "github.com/opd-ai/venture/pkg/network"

// Create server
config := network.DefaultServerConfig()
config.Address = ":8080"
config.MaxPlayers = 32
server := network.NewServer(config)

// Start listening
if err := server.Start(); err != nil {
    log.Fatal(err)
}
defer server.Stop()

// Handle input commands
go func() {
    for cmd := range server.ReceiveInputCommand() {
        log.Printf("Player %d: %s", cmd.PlayerID, cmd.InputType)
        // Process input...
    }
}()

// Game loop: broadcast state
ticker := time.NewTicker(50 * time.Millisecond) // 20 updates/sec
for range ticker.C {
    update := &network.StateUpdate{
        Timestamp:  uint64(time.Now().UnixNano()),
        EntityID:   1,
        Priority:   128,
        Components: []network.ComponentData{
            {Type: "position", Data: encodePosition(...)},
        },
    }
    server.BroadcastStateUpdate(update)
}
```

### Basic Client

```go
import "github.com/opd-ai/venture/pkg/network"

// Create client
config := network.DefaultClientConfig()
config.ServerAddress = "localhost:8080"
client := network.NewClient(config)

// Connect
if err := client.Connect(); err != nil {
    log.Fatal(err)
}
defer client.Disconnect()

// Set player ID (from auth response)
client.SetPlayerID(1)

// Handle state updates from server
go func() {
    for update := range client.ReceiveStateUpdate() {
        // Apply update to game world
        applyUpdate(world, update)
    }
}()

// Send player input
client.SendInput("move", encodeMovement(dx, dy))
```

### Client-Side Prediction

```go
import "github.com/opd-ai/venture/pkg/network"

// Create predictor
predictor := network.NewClientPredictor()
predictor.SetInitialState(network.Position{X: 0, Y: 0}, network.Velocity{VX: 0, VY: 0})

// In game loop: predict player movement immediately
func handleInput(dx, dy float64, deltaTime float64) {
    // Predict locally for immediate response
    predicted := predictor.PredictInput(dx*100, dy*100, deltaTime)
    
    // Update local player position
    player.Position = predicted.Position
    
    // Send input to server
    client.SendInput("move", encodeMovement(dx, dy))
}

// When server update arrives: reconcile
func onServerUpdate(update *network.StateUpdate) {
    if update.EntityID == playerID {
        // Extract position and velocity from server
        serverPos := decodePosition(update.Components)
        serverVel := decodeVelocity(update.Components)
        
        // Reconcile with server authority
        corrected := predictor.ReconcileServerState(
            update.Sequence,
            serverPos,
            serverVel,
        )
        
        // Apply corrected position
        player.Position = corrected.Position
    }
}
```

### Entity Interpolation

```go
import "github.com/opd-ai/venture/pkg/network"

// Create snapshot manager
snapshots := network.NewSnapshotManager(100) // Keep 100 snapshots

// When server update arrives: store snapshot
func onServerUpdate(update *network.StateUpdate) {
    // Build snapshot from update
    snapshot := network.WorldSnapshot{
        Entities: map[uint64]network.EntitySnapshot{
            update.EntityID: {
                EntityID: update.EntityID,
                Position: decodePosition(update.Components),
                Velocity: decodeVelocity(update.Components),
            },
        },
    }
    
    snapshots.AddSnapshot(snapshot)
}

// In render loop: interpolate remote entities
func renderEntity(entityID uint64) {
    // Render time is slightly in the past (interpolation delay)
    renderTime := time.Now().Add(-100 * time.Millisecond)
    
    // Interpolate entity position
    interpolated := snapshots.InterpolateEntity(entityID, renderTime)
    
    if interpolated != nil {
        // Render entity at interpolated position
        drawEntity(entityID, interpolated.Position)
    }
}
```

### State Synchronization with Delta Compression

```go
import "github.com/opd-ai/venture/pkg/network"

// Server-side: send delta updates instead of full snapshots
func broadcastDelta(server *network.Server, snapshots *network.SnapshotManager) {
    currentSeq := snapshots.GetCurrentSequence()
    previousSeq := currentSeq - 1
    
    // Create delta between previous and current snapshot
    delta := snapshots.CreateDelta(previousSeq, currentSeq)
    
    if delta != nil {
        // Send only the changes (much smaller than full snapshot)
        // For added/changed entities, send full data
        for entityID := range delta.Changed {
            // Create and broadcast state update
            update := createStateUpdate(entityID, delta.Changed[entityID])
            server.BroadcastStateUpdate(update)
        }
    }
}

// Handle state updates
go func() {
    for update := range client.ReceiveStateUpdate() {
        log.Printf("Entity %d updated", update.EntityID)
        // Apply state...
    }
}()

// Handle errors
go func() {
    for err := range client.ReceiveError() {
        log.Printf("Network error: %v", err)
    }
}()

// Game loop: send input
for {
    // On player input
    if keyPressed("W") {
        client.SendInput("move", encodeMovement(0, -1))
    }
}
```

### Binary Protocol

```go
import "github.com/opd-ai/venture/pkg/network"

// Create protocol
protocol := network.NewBinaryProtocol()

// Encode state update
update := &network.StateUpdate{
    Timestamp:      12345,
    EntityID:       100,
    Priority:       128,
    SequenceNumber: 42,
    Components: []network.ComponentData{
        {Type: "position", Data: []byte{1, 2, 3, 4}},
    },
}

data, err := protocol.EncodeStateUpdate(update)
// Send data over network...

// Decode state update
decoded, err := protocol.DecodeStateUpdate(data)
// Use decoded update...

// Encode input command
cmd := &network.InputCommand{
    PlayerID:       1,
    Timestamp:      5555,
    SequenceNumber: 10,
    InputType:      "move",
    Data:           []byte{1, 0}, // dx, dy
}

data, err = protocol.EncodeInputCommand(cmd)
// Send data over network...

// Decode input command
decoded, err := protocol.DecodeInputCommand(data)
// Process input...
```

## Configuration

### Client Config

```go
config := network.ClientConfig{
    ServerAddress:     "localhost:8080",  // Server to connect to
    ConnectionTimeout: 10 * time.Second,  // Connection timeout
    PingInterval:      1 * time.Second,   // Ping frequency
    MaxLatency:        500 * time.Millisecond, // Latency warning threshold
    BufferSize:        256,               // Channel buffer size
}
```

### Server Config

```go
config := network.ServerConfig{
    Address:      ":8080",         // Listen address
    MaxPlayers:   32,              // Maximum concurrent players
    ReadTimeout:  10 * time.Second, // Read timeout per client
    WriteTimeout: 5 * time.Second, // Write timeout per client
    UpdateRate:   20,              // State updates per second
    BufferSize:   256,             // Channel buffer size per client
}
```

## Performance

### Serialization Benchmarks

```
BenchmarkEncodeStateUpdate-4    	 2697820	       448.0 ns/op	     288 B/op
BenchmarkDecodeStateUpdate-4    	 2038742	       589.4 ns/op	     344 B/op
BenchmarkEncodeInputCommand-4   	 5722370	       210.8 ns/op	     144 B/op
BenchmarkDecodeInputCommand-4   	 4468304	       279.0 ns/op	     160 B/op
```

- **StateUpdate encode**: ~0.4 µs (2.2M ops/sec)
- **StateUpdate decode**: ~0.6 µs (1.7M ops/sec)
- **InputCommand encode**: ~0.2 µs (4.7M ops/sec)
- **InputCommand decode**: ~0.3 µs (3.6M ops/sec)

### Packet Sizes

- **StateUpdate** (3 components): ~80 bytes
- **InputCommand** (5 bytes data): ~35 bytes
- **Overhead**: 4 bytes (length prefix)

### Bandwidth Estimation

At 20 updates/second with 32 players:

- **Downstream** (server → client): ~80 bytes × 32 entities × 20 = ~51 KB/s
- **Upstream** (client → server): ~35 bytes × 20 = ~0.7 KB/s
- **Total per player**: ~52 KB/s (well within 100 KB/s target)

## Wire Protocol

### Message Framing

All messages use length-prefixed framing:

```
[4 bytes: message length][N bytes: message data]
```

Length is uint32 little-endian.

### StateUpdate Format

```
[8 bytes: timestamp]
[8 bytes: entity ID]
[1 byte: priority]
[4 bytes: sequence number]
[2 bytes: component count]
For each component:
  [2 bytes: type length]
  [N bytes: type string]
  [4 bytes: data length]
  [M bytes: data]
```

### InputCommand Format

```
[8 bytes: player ID]
[8 bytes: timestamp]
[4 bytes: sequence number]
[2 bytes: input type length]
[N bytes: input type string]
[4 bytes: data length]
[M bytes: data]
```

## Thread Safety

- **Client**: Thread-safe. Multiple goroutines can call `SendInput()` concurrently.
- **Server**: Thread-safe. Multiple goroutines can call broadcast/send methods.
- **Protocol**: Stateless, safe for concurrent use.

## Error Handling

Errors are reported through dedicated channels:

```go
// Client errors
for err := range client.ReceiveError() {
    log.Printf("Client error: %v", err)
}

// Server errors
for err := range server.ReceiveError() {
    log.Printf("Server error: %v", err)
}
```

Common errors:

- **Connection failures**: Network issues, server down
- **Timeout errors**: Client/server not responding
- **Decode errors**: Malformed packets
- **Buffer full**: Too much data queued

## Integration with ECS

### Component Serialization

The `ComponentSerializer` provides ready-to-use methods for serializing common ECS components:

```go
import (
    "github.com/opd-ai/venture/pkg/engine"
    "github.com/opd-ai/venture/pkg/network"
)

// Create serializer
serializer := network.NewComponentSerializer()

// Serialize position
pos := &engine.PositionComponent{X: 123.45, Y: 678.90}
posData := serializer.SerializePosition(pos.X, pos.Y)

// Deserialize position
x, y, err := serializer.DeserializePosition(posData)

// Serialize health
health := &engine.HealthComponent{Current: 75, Max: 100}
healthData := serializer.SerializeHealth(health.Current, health.Max)

// Deserialize health
current, max, err := serializer.DeserializeHealth(healthData)

// Serialize stats
statsData := serializer.SerializeStats(attack, defense, magicPower)

// Serialize input
inputData := serializer.SerializeInput(dx, dy)

// Serialize attack command
attackData := serializer.SerializeAttack(targetID)
```

**Supported Components:**
- **Position**: X, Y coordinates (16 bytes)
- **Velocity**: VX, VY velocities (16 bytes)
- **Health**: Current, Max health (16 bytes)
- **Stats**: Attack, Defense, MagicPower (24 bytes)
- **Team**: Team ID (8 bytes)
- **Level**: Level, XP (8 bytes)
- **Input**: Movement dx, dy (2 bytes)
- **Attack**: Target entity ID (8 bytes)
- **Item**: Item ID (8 bytes)

### Creating Entity Updates

```go
// Create state update for entity
func CreateEntityUpdate(entity *engine.Entity, serializer *network.ComponentSerializer) *network.StateUpdate {
    components := []network.ComponentData{}
    
    // Serialize position
    if pos, ok := entity.GetComponent("position"); ok {
        position := pos.(*engine.PositionComponent)
        components = append(components, network.ComponentData{
            Type: "position",
            Data: serializer.SerializePosition(position.X, position.Y),
        })
    }
    
    // Serialize health
    if hp, ok := entity.GetComponent("health"); ok {
        health := hp.(*engine.HealthComponent)
        components = append(components, network.ComponentData{
            Type: "health",
            Data: serializer.SerializeHealth(health.Current, health.Max),
        })
    }
    
    return &network.StateUpdate{
        Timestamp:  uint64(time.Now().UnixNano()),
        EntityID:   entity.ID,
        Priority:   128,
        Components: components,
    }
}
```

## Future Enhancements

### Planned Features

- [ ] Client-side prediction
- [ ] Server-side interpolation
- [ ] Lag compensation
- [ ] Delta compression
- [ ] Connection encryption (TLS)
- [ ] Authentication system
- [ ] Matchmaking
- [ ] Reconnection handling
- [ ] Bandwidth throttling
- [ ] Packet prioritization

### Advanced Features

- [ ] UDP transport option
- [ ] WebSocket support
- [ ] NAT traversal
- [ ] Relay servers
- [ ] Anti-cheat validation
- [ ] Replay recording

## Testing

```bash
# Run tests
go test -tags test ./pkg/network/

# Run with coverage
go test -tags test -cover ./pkg/network/

# Run benchmarks
go test -tags test -bench=. -benchmem ./pkg/network/
```

Current coverage: 82.6% (serialization fully tested, network I/O needs integration tests)

## Design Decisions

### Why Binary Protocol?

✅ **Performance**: 10x faster than JSON  
✅ **Bandwidth**: 50% smaller packets  
✅ **Deterministic**: Fixed byte layout  
✅ **Type-Safe**: Schema enforced

### Why Length-Prefixed Framing?

✅ **Simple**: Easy to parse  
✅ **Efficient**: No escaping needed  
✅ **Reliable**: Detects truncation  
✅ **Compatible**: Works with TCP streams

### Why Channels for Communication?

✅ **Go Idiom**: Natural Go pattern  
✅ **Async**: Non-blocking I/O  
✅ **Safe**: No shared memory  
✅ **Testable**: Easy to mock

### Why Authoritative Server?

✅ **Security**: Prevents cheating  
✅ **Consistency**: Single source of truth  
✅ **Scalability**: Server controls state  
✅ **Validation**: Server validates inputs

## References

- [Multiplayer Networking](https://gafferongames.com/post/what_every_programmer_needs_to_know_about_game_networking/)
- [Source Engine Networking](https://developer.valvesoftware.com/wiki/Source_Multiplayer_Networking)
- [Fast-Paced Multiplayer](https://www.gabrielgambetta.com/client-server-game-architecture.html)

## License

See [LICENSE](../../LICENSE) file for details.
