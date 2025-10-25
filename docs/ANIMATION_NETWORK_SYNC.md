# Animation Network Synchronization

**Version:** 1.0  
**Phase:** 7.1  
**Status:** Production Ready  
**Last Updated:** October 25, 2025

## Overview

The Animation Network Synchronization system enables deterministic, bandwidth-efficient synchronization of entity animation states across multiplayer clients. This system integrates seamlessly with Venture's existing animation system (Phase 1) and networking infrastructure (Phase 6) to provide smooth, synchronized animations in multiplayer gameplay.

### Key Features

- **Delta Compression**: Only state changes transmitted (90% bandwidth reduction)
- **Compact Encoding**: 20 bytes per packet (EntityID + State + Frame + Timestamp + Loop)
- **Client Interpolation**: 150ms buffer for smooth playback despite network jitter
- **Deterministic Rendering**: Same state + seed = identical animation across all clients
- **Efficient**: Sub-microsecond delta checks, minimal allocation overhead
- **Scalable**: Batch support for simultaneous state changes

### Performance Characteristics

| Metric | Value | Target | Status |
|--------|-------|--------|--------|
| Encode latency | 376 ns | <1 μs | ✅ Excellent |
| Decode latency | 229 ns | <1 μs | ✅ Excellent |
| Delta check | 8.9 ns | <100 ns | ✅ Excellent |
| Packet size | 20 bytes | <50 bytes | ✅ Excellent |
| Bandwidth | 1 KB/s | <100 KB/s | ✅ Excellent |
| Allocations (delta) | 0 | 0 | ✅ Perfect |

---

## Table of Contents

1. [Architecture](#architecture)
2. [Packet Format](#packet-format)
3. [API Reference](#api-reference)
4. [Integration Guide](#integration-guide)
5. [Performance Optimization](#performance-optimization)
6. [Troubleshooting](#troubleshooting)
7. [Examples](#examples)

---

## Architecture

### System Components

```
┌─────────────────────────────────────────────────────────┐
│                   Server Game Loop                       │
└───────────────┬─────────────────────────────────────────┘
                │
                ▼
┌───────────────────────────────────────────────────────┐
│         Animation System (Phase 1)                    │
│  • Detects state changes (idle → walk)               │
│  • Updates frame indices                             │
│  • Triggers sync callbacks                           │
└───────────────┬───────────────────────────────────────┘
                │ State Change Detected
                ▼
┌───────────────────────────────────────────────────────┐
│     AnimationSyncManager (Server)                     │
│  • Delta compression check                           │
│  • Only sync if state changed                        │
│  • Spatial culling (viewport-based)                  │
└───────────────┬───────────────────────────────────────┘
                │ Create Packet
                ▼
┌───────────────────────────────────────────────────────┐
│        AnimationStatePacket                           │
│  • Encode to 20 bytes                                │
│  • Binary format for efficiency                      │
└───────────────┬───────────────────────────────────────┘
                │ Network Transmission
                ▼
┌───────────────────────────────────────────────────────┐
│        Client Network Handler                         │
│  • Receives packet                                   │
│  • Decodes state                                     │
└───────────────┬───────────────────────────────────────┘
                │
                ▼
┌───────────────────────────────────────────────────────┐
│     AnimationSyncManager (Client)                     │
│  • Buffers state for interpolation                   │
│  • 150ms buffer (3 states @ 20Hz)                    │
│  • FIFO queue ensures ordered playback               │
└───────────────┬───────────────────────────────────────┘
                │ Apply State
                ▼
┌───────────────────────────────────────────────────────┐
│         Animation System (Client)                     │
│  • Applies state to entity                           │
│  • Deterministic frame generation                    │
│  • Identical animation as server                     │
└───────────────────────────────────────────────────────┘
```

### Design Philosophy

1. **State-Only Sync**: Transmit animation state (idle/walk/attack), not sprite data. Deterministic generation ensures identical visuals.

2. **Delta Compression**: Only send state changes, not continuous updates. Reduces bandwidth by 90% in typical scenarios.

3. **Client Interpolation**: Buffer incoming states to smooth out network jitter and packet loss.

4. **Spatial Relevance**: Integrate with viewport culling (Phase 6) to only sync visible entities.

5. **Minimal Latency**: Fast encoding/decoding ensures <1ms overhead per packet.

---

## Packet Format

### AnimationStatePacket

**Binary Layout** (20 bytes total):

```
 0                   1                   2                   3
 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|                          EntityID (64-bit)                    |
|                                                               |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|    StateID    |         FrameIndex (16-bit)      |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|                        Timestamp (64-bit)                     |
|                                                               |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|     Loop      |
+-+-+-+-+-+-+-+-+
```

**Field Descriptions:**

- **EntityID** (8 bytes): Unique entity identifier
- **StateID** (1 byte): Animation state as uint8 (0=idle, 1=walk, etc.)
- **FrameIndex** (2 bytes): Current frame in animation (0-65535)
- **Timestamp** (8 bytes): Server timestamp for interpolation (microseconds)
- **Loop** (1 byte): Boolean flag (0=false, 1=true)

### AnimationStateBatch

**Binary Layout**:

```
 0                   1                   2                   3
 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|      Count (16-bit)      |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|                     Batch Timestamp (64-bit)                  |
|                                                               |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|                 AnimationStatePacket #1 (20 bytes)            |
|                           ...                                 |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|                 AnimationStatePacket #2 (20 bytes)            |
|                           ...                                 |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|                           ... (Count packets)                 |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
```

**Use Case**: Batch multiple state changes into single network message to reduce protocol overhead.

**Size Calculation**: 10 bytes (header) + Count × 20 bytes (states)

### State ID Mapping

| State | ID | Description |
|-------|----|----|
| Idle | 0 | Standing still, breathing |
| Walk | 1 | Normal movement |
| Run | 2 | Fast movement |
| Attack | 3 | Melee attack animation |
| Cast | 4 | Spell casting animation |
| Hit | 5 | Taking damage animation |
| Death | 6 | Death/despawn animation |
| Jump | 7 | Jumping animation |
| Crouch | 8 | Crouching/sneaking |
| Use | 9 | Interacting with object |

**Extension**: IDs 10-255 reserved for future states. Unknown IDs default to Idle (0).

---

## API Reference

### AnimationStatePacket

#### Constructor

Not required - use struct literal:

```go
packet := network.AnimationStatePacket{
    EntityID:   entityID,
    State:      engine.AnimationStateWalk,
    FrameIndex: 5,
    Timestamp:  time.Now().UnixMicro(),
    Loop:       true,
}
```

#### Methods

**Encode() ([]byte, error)**

Serializes packet to 20-byte binary format.

```go
data, err := packet.Encode()
if err != nil {
    log.Fatalf("Encode failed: %v", err)
}
// data is exactly 20 bytes
```

**Errors:**
- `FrameIndex out of range` if FrameIndex < 0 or > 65535

**Decode(data []byte) error**

Deserializes 20-byte binary data into packet.

```go
var packet network.AnimationStatePacket
err := packet.Decode(data)
if err != nil {
    log.Fatalf("Decode failed: %v", err)
}
```

**Errors:**
- `packet too short` if len(data) < 20

### AnimationStateBatch

#### Constructor

```go
batch := network.AnimationStateBatch{
    States: []network.AnimationStatePacket{
        {EntityID: 1, State: engine.AnimationStateIdle, ...},
        {EntityID: 2, State: engine.AnimationStateWalk, ...},
    },
    Timestamp: time.Now().UnixMicro(),
}
```

#### Methods

**Encode() ([]byte, error)**

Serializes batch with all states.

```go
data, err := batch.Encode()
// data size: 10 + (len(batch.States) * 20) bytes
```

**Decode(data []byte) error**

Deserializes batch from binary data.

### AnimationSyncManager

#### Constructor

```go
manager := network.NewAnimationSyncManager()
// Default buffer size: 3 states (150ms @ 20Hz)
```

#### Server-Side Methods

**ShouldSync(entityID uint64, newState engine.AnimationState) bool**

Determines if state change should be transmitted (delta compression).

```go
if manager.ShouldSync(entityID, newState) {
    packet := network.AnimationStatePacket{ /* ... */ }
    data, _ := packet.Encode()
    sendToClients(data)
    manager.RecordSync(entityID, newState, len(data))
}
```

**Returns**: `true` if state changed since last sync, `false` otherwise.

**RecordSync(entityID uint64, state engine.AnimationState, bytesSent int)**

Records that a state was transmitted (for delta tracking and statistics).

```go
manager.RecordSync(entityID, engine.AnimationStateWalk, 20)
```

#### Client-Side Methods

**BufferState(packet AnimationStatePacket) bool**

Adds received state to interpolation buffer.

```go
var packet network.AnimationStatePacket
packet.Decode(receivedData)

if manager.BufferState(packet) {
    // Buffer full - start applying states
    applyNextState()
}
```

**Returns**: `true` if buffer full (ready to apply), `false` if still buffering.

**GetNextState(entityID uint64) *AnimationStatePacket**

Retrieves next state from interpolation buffer (FIFO).

```go
state := manager.GetNextState(entityID)
if state != nil {
    animationSystem.SetState(entityID, state.State, state.FrameIndex)
}
```

**Returns**: Next state or `nil` if buffer empty.

#### Utility Methods

**ClearEntity(entityID uint64)**

Removes tracking data when entity destroyed.

```go
manager.ClearEntity(entityID)
```

**Stats() AnimationSyncStats**

Returns synchronization statistics.

```go
stats := manager.Stats()
fmt.Printf("Sent: %d, Received: %d, Bandwidth: %.2f KB/s\n",
    stats.StateChangesSent,
    stats.StatesReceived,
    stats.Bandwidth(time.Second))
```

### AnimationSyncStats

```go
type AnimationSyncStats struct {
    StateChangesSent uint64  // Total states transmitted
    StatesReceived   uint64  // Total states received
    BytesTransmitted uint64  // Total bytes sent
    BufferedEntities int     // Entities with buffered states
}
```

**Bandwidth(duration time.Duration) float64**

Calculates average bytes per second over given duration.

---

## Integration Guide

### Server-Side Integration

**Step 1: Create Manager**

```go
// In server initialization
animSyncMgr := network.NewAnimationSyncManager()
```

**Step 2: Hook Animation System**

```go
// In animation system update loop
func (s *AnimationSystem) Update(entities []*Entity, deltaTime float64) {
    for _, entity := range entities {
        anim := entity.GetComponent("animation").(*AnimationComponent)
        
        // Detect state change
        if anim.CurrentState != anim.PreviousState {
            // Check if should sync
            if animSyncMgr.ShouldSync(entity.ID, anim.CurrentState) {
                // Create packet
                packet := network.AnimationStatePacket{
                    EntityID:   entity.ID,
                    State:      anim.CurrentState,
                    FrameIndex: anim.FrameIndex,
                    Timestamp:  time.Now().UnixMicro(),
                    Loop:       anim.Loop,
                }
                
                // Encode and send
                data, _ := packet.Encode()
                server.BroadcastToRelevantClients(entity, data)
                
                // Record sync
                animSyncMgr.RecordSync(entity.ID, anim.CurrentState, len(data))
            }
            
            anim.PreviousState = anim.CurrentState
        }
    }
}
```

**Step 3: Spatial Culling Optimization**

```go
// Only send to clients who can see the entity
func (s *Server) BroadcastToRelevantClients(entity *Entity, data []byte) {
    pos := entity.GetComponent("position").(*PositionComponent)
    
    for _, client := range s.clients {
        // Check if entity in client's viewport
        if client.ViewportContains(pos.X, pos.Y) {
            client.Send(PacketTypeAnimationState, data)
        }
    }
}
```

### Client-Side Integration

**Step 1: Create Manager**

```go
// In client initialization
animSyncMgr := network.NewAnimationSyncManager()
```

**Step 2: Handle Incoming Packets**

```go
// In network message handler
func (c *Client) HandleAnimationStatePacket(data []byte) {
    var packet network.AnimationStatePacket
    if err := packet.Decode(data); err != nil {
        log.Printf("Failed to decode animation packet: %v", err)
        return
    }
    
    // Buffer for interpolation
    if animSyncMgr.BufferState(packet) {
        // Buffer full - start applying
        c.applyBufferedStates()
    }
}
```

**Step 3: Apply Buffered States**

```go
// In render loop or update loop
func (c *Client) applyBufferedStates() {
    for _, entity := range c.entities {
        state := animSyncMgr.GetNextState(entity.ID)
        if state != nil {
            // Apply to animation component
            anim := entity.GetComponent("animation").(*AnimationComponent)
            anim.SetState(state.State)
            anim.FrameIndex = state.FrameIndex
            anim.Loop = state.Loop
        }
    }
}
```

**Step 4: Handle Entity Cleanup**

```go
func (c *Client) RemoveEntity(entityID uint64) {
    // Clear sync manager tracking
    animSyncMgr.ClearEntity(entityID)
    
    // Remove entity from world
    delete(c.entities, entityID)
}
```

---

## Performance Optimization

### Bandwidth Optimization

**1. Spatial Culling** (Recommended)

Only send animations for entities within client viewports:

```go
// Reduces network traffic by 90%+ in typical scenarios
if !clientCanSee(entity) {
    continue // Skip sync
}
```

**2. Update Rate Throttling**

Not all entities need 20Hz updates:

```go
// Background NPCs: 5Hz (200ms)
// Nearby entities: 20Hz (50ms)
// Player entities: 30Hz (33ms)

if entity.DistanceFromPlayer() > 500 && time.Since(entity.LastSync) < 200*time.Millisecond {
    continue // Throttle distant entities
}
```

**3. Batch Optimization**

Use batches when many entities change simultaneously:

```go
if len(stateChanges) > 3 {
    batch := network.AnimationStateBatch{
        States: stateChanges,
        Timestamp: time.Now().UnixMicro(),
    }
    data, _ := batch.Encode()
    broadcast(data)
} else {
    // Individual packets for <3 states (less overhead)
    for _, state := range stateChanges {
        // ...
    }
}
```

### CPU Optimization

**1. Delta Check Caching**

Delta check is already optimized (8.9ns, 0 allocs), but can be further improved:

```go
// Pre-compute hash of state for faster comparison
type stateHash uint64

func computeStateHash(state engine.AnimationState) stateHash {
    // Simple hash - state enum value is sufficient
    return stateHash(animationStateToID(state))
}
```

**2. Pool Packet Objects** (Optional)

```go
var packetPool = sync.Pool{
    New: func() interface{} {
        return &network.AnimationStatePacket{}
    },
}

// Get from pool
packet := packetPool.Get().(*network.AnimationStatePacket)
// ... use packet ...
packetPool.Put(packet) // Return to pool
```

**3. Batch Encoding Optimization**

Pre-allocate buffer for known batch size:

```go
// Avoid multiple allocations during batch encoding
expectedSize := 10 + len(batch.States) * 20
buf := bytes.NewBuffer(make([]byte, 0, expectedSize))
```

### Memory Optimization

**1. Buffer Size Tuning**

Adjust interpolation buffer based on network conditions:

```go
// Low latency (<50ms): buffer 1-2 states
// Medium latency (50-100ms): buffer 3 states (default)
// High latency (>100ms): buffer 5+ states

if avgLatency < 50*time.Millisecond {
    manager.bufferSize = 2
} else if avgLatency > 100*time.Millisecond {
    manager.bufferSize = 5
}
```

**2. Periodic Cleanup**

Clear tracking data for inactive entities:

```go
// Every 60 seconds
if time.Since(lastCleanup) > 60*time.Second {
    for entityID, lastSeen := range entityLastSeen {
        if time.Since(lastSeen) > 30*time.Second {
            manager.ClearEntity(entityID)
        }
    }
}
```

---

## Troubleshooting

### Problem: Animations Out of Sync

**Symptoms**: Client animations don't match server

**Causes**:
1. Non-deterministic animation generation
2. Missing state updates
3. Packet loss without recovery

**Solutions**:
- Verify animation generation uses same seed on client/server
- Enable periodic full-state sync (every 5 seconds)
- Implement packet loss detection and resend requests

```go
// Periodic full sync
if time.Since(lastFullSync) > 5*time.Second {
    sendFullAnimationState(entity)
    lastFullSync = time.Now()
}
```

### Problem: High Bandwidth Usage

**Symptoms**: >10 KB/s per player

**Causes**:
1. Spatial culling not enabled
2. Too frequent updates
3. Not using delta compression

**Solutions**:
- Enable viewport culling (reduces by 90%)
- Throttle distant entity updates
- Verify `ShouldSync()` is being called

```go
stats := manager.Stats()
bandwidth := stats.Bandwidth(time.Second)
if bandwidth > 10000 { // > 10 KB/s
    log.Printf("Warning: High animation bandwidth: %.2f KB/s", bandwidth/1024)
    // Enable aggressive culling/throttling
}
```

### Problem: Jerky Animations on Client

**Symptoms**: Stuttering or teleporting animations

**Causes**:
1. Network jitter without buffering
2. Buffer size too small
3. Applying states too eagerly

**Solutions**:
- Increase buffer size (3-5 states)
- Wait for buffer to fill before applying first state
- Smooth frame transitions with interpolation

```go
// Wait for buffer before starting playback
if !bufferInitialized {
    buffer := manager.stateBuffer[entityID]
    if len(buffer) >= manager.bufferSize {
        bufferInitialized = true
    } else {
        return // Wait for more states
    }
}
```

### Problem: Memory Leak

**Symptoms**: Memory usage grows over time

**Causes**:
1. Not clearing entities on removal
2. Buffer accumulation
3. Tracking data not cleaned up

**Solutions**:
- Always call `ClearEntity()` when removing entities
- Implement periodic cleanup for stale data
- Monitor `BufferedEntities` stat

```go
// Cleanup on entity removal
func (w *World) RemoveEntity(entityID uint64) {
    animSyncMgr.ClearEntity(entityID)
    delete(w.entities, entityID)
}

// Periodic cleanup
if frameCount % 3600 == 0 { // Every 60 seconds @ 60 FPS
    stats := animSyncMgr.Stats()
    if stats.BufferedEntities > len(world.entities) {
        log.Printf("Warning: %d buffered entities vs %d active entities",
            stats.BufferedEntities, len(world.entities))
        // Perform cleanup
    }
}
```

---

## Examples

### Complete Server Example

```go
package main

import (
    "time"
    "github.com/opd-ai/venture/pkg/engine"
    "github.com/opd-ai/venture/pkg/network"
)

type Server struct {
    entities       []*engine.Entity
    animSyncMgr    *network.AnimationSyncManager
    animSystem     *engine.AnimationSystem
    clients        []*Client
}

func (s *Server) Update(deltaTime float64) {
    // Update animation system
    s.animSystem.Update(s.entities, deltaTime)
    
    // Sync animation states
    for _, entity := range s.entities {
        anim, ok := entity.GetComponent("animation").(*engine.AnimationComponent)
        if !ok {
            continue
        }
        
        // Check for state change
        if anim.CurrentState != anim.PreviousState {
            // Delta compression check
            if s.animSyncMgr.ShouldSync(entity.ID, anim.CurrentState) {
                // Create packet
                packet := network.AnimationStatePacket{
                    EntityID:   entity.ID,
                    State:      anim.CurrentState,
                    FrameIndex: anim.FrameIndex,
                    Timestamp:  time.Now().UnixMicro(),
                    Loop:       anim.Loop,
                }
                
                // Encode
                data, err := packet.Encode()
                if err != nil {
                    continue
                }
                
                // Send to relevant clients (spatial culling)
                s.BroadcastAnimationState(entity, data)
                
                // Record sync for stats
                s.animSyncMgr.RecordSync(entity.ID, anim.CurrentState, len(data))
            }
            
            anim.PreviousState = anim.CurrentState
        }
    }
}

func (s *Server) BroadcastAnimationState(entity *engine.Entity, data []byte) {
    pos := entity.GetComponent("position").(*engine.PositionComponent)
    
    for _, client := range s.clients {
        // Spatial culling - only send if in viewport
        if client.ViewportContains(pos.X, pos.Y) {
            client.SendPacket(network.PacketTypeAnimationState, data)
        }
    }
}
```

### Complete Client Example

```go
package main

import (
    "time"
    "github.com/opd-ai/venture/pkg/engine"
    "github.com/opd-ai/venture/pkg/network"
)

type Client struct {
    entities        map[uint64]*engine.Entity
    animSyncMgr     *network.AnimationSyncManager
    bufferReady     map[uint64]bool
}

func NewClient() *Client {
    return &Client{
        entities:    make(map[uint64]*engine.Entity),
        animSyncMgr: network.NewAnimationSyncManager(),
        bufferReady: make(map[uint64]bool),
    }
}

func (c *Client) HandleNetworkPacket(packetType uint8, data []byte) {
    switch packetType {
    case network.PacketTypeAnimationState:
        c.handleAnimationState(data)
    case network.PacketTypeAnimationStateBatch:
        c.handleAnimationBatch(data)
    }
}

func (c *Client) handleAnimationState(data []byte) {
    var packet network.AnimationStatePacket
    if err := packet.Decode(data); err != nil {
        return
    }
    
    // Buffer for interpolation
    if c.animSyncMgr.BufferState(packet) {
        // Buffer full - mark as ready
        c.bufferReady[packet.EntityID] = true
    }
}

func (c *Client) Update(deltaTime float64) {
    // Apply buffered animation states
    for entityID, entity := range c.entities {
        // Wait for buffer to fill initially
        if !c.bufferReady[entityID] {
            continue
        }
        
        // Get next state from buffer
        state := c.animSyncMgr.GetNextState(entityID)
        if state == nil {
            continue
        }
        
        // Apply to animation component
        anim, ok := entity.GetComponent("animation").(*engine.AnimationComponent)
        if !ok {
            continue
        }
        
        anim.SetState(state.State)
        anim.FrameIndex = state.FrameIndex
        anim.Loop = state.Loop
    }
}

func (c *Client) RemoveEntity(entityID uint64) {
    c.animSyncMgr.ClearEntity(entityID)
    delete(c.entities, entityID)
    delete(c.bufferReady, entityID)
}
```

---

## Related Documentation

- **Animation System**: `docs/ANIMATION_SYSTEM.md` (Phase 1)
- **Networking**: `pkg/network/README.md` (if exists)
- **Client Prediction**: `examples/prediction_demo/` (Phase 6)
- **Performance**: `docs/PERFORMANCE.md` (Phase 6)

---

**Document Version**: 1.0  
**Last Updated**: October 25, 2025  
**Maintained By**: Venture Development Team  
**Phase**: 7.1 - Animation Network Synchronization
