# Quick Start: Task 1.1 - Network Server Implementation

## Overview
**Goal:** Implement actual TCP network server for multiplayer functionality  
**Priority:** CRITICAL (Gap #1, Score 162.5)  
**Effort:** 2-3 days  
**Status:** ⭕ Not Started

## Why This First?
- Blocks all multiplayer functionality
- Highest priority gap (162.5 score)
- Documentation claims "Full multiplayer support" but server is a stub
- No dependencies - can start immediately
- Will raise network package coverage from 66.8% toward 80% target

## Current Problem
```go
// cmd/server/main.go:100
log.Printf("Server running on port %s (not accepting connections yet - network layer stub)", *port)
```

Server runs authoritative game loop but has no network listener. Cannot accept client connections.

## What to Implement

### 1. Network Protocol (pkg/network/protocol.go)
```go
package network

// Message types for client-server communication
type MessageType uint8

const (
    MsgTypeHandshake MessageType = iota
    MsgTypeAuthenticate
    MsgTypeWorldState
    MsgTypePlayerInput
    MsgTypeDisconnect
    MsgTypeError
)

// Message wraps all network messages
type Message struct {
    Type      MessageType
    Timestamp time.Time
    Payload   []byte
}

// HandshakeRequest sent by client on connect
type HandshakeRequest struct {
    ProtocolVersion uint32
    PlayerName      string
}

// HandshakeResponse sent by server after validation
type HandshakeResponse struct {
    Accepted      bool
    PlayerID      uint64
    WorldSeed     int64
    GenreID       string
    ErrorMessage  string // if not accepted
}

// WorldStateMessage broadcasts game state
type WorldStateMessage struct {
    Snapshot WorldSnapshot
}

// PlayerInputMessage from client to server
type PlayerInputMessage struct {
    PlayerID  uint64
    Sequence  uint32
    MoveX     float64
    MoveY     float64
    Actions   uint32 // bitfield for action buttons
}

// Serialize/Deserialize methods for each message type
```

### 2. Client Connection Handler (pkg/network/client_connection.go)
```go
package network

import (
    "net"
    "sync"
)

// ClientConnection represents a connected client
type ClientConnection struct {
    conn      net.Conn
    playerID  uint64
    playerName string
    
    // Input buffering
    inputQueue []PlayerInputMessage
    inputLock  sync.Mutex
    
    // Send channel for outgoing messages
    sendChan chan Message
    
    // Graceful shutdown
    closeChan chan struct{}
    closed    bool
}

// NewClientConnection creates a new client handler
func NewClientConnection(conn net.Conn) *ClientConnection {
    return &ClientConnection{
        conn:      conn,
        sendChan:  make(chan Message, 100),
        closeChan: make(chan struct{}),
    }
}

// ReadLoop continuously reads messages from client
func (c *ClientConnection) ReadLoop() {
    defer c.Close()
    
    for {
        msg, err := c.readMessage()
        if err != nil {
            log.Printf("Client read error: %v", err)
            return
        }
        
        c.handleMessage(msg)
    }
}

// WriteLoop continuously sends messages to client
func (c *ClientConnection) WriteLoop() {
    defer c.Close()
    
    for {
        select {
        case msg := <-c.sendChan:
            if err := c.writeMessage(msg); err != nil {
                log.Printf("Client write error: %v", err)
                return
            }
        case <-c.closeChan:
            return
        }
    }
}

// SendMessage queues a message for sending
func (c *ClientConnection) SendMessage(msg Message) {
    if !c.closed {
        c.sendChan <- msg
    }
}

// Close gracefully shuts down the connection
func (c *ClientConnection) Close() {
    if !c.closed {
        c.closed = true
        close(c.closeChan)
        c.conn.Close()
    }
}

// Helper methods for reading/writing protocol messages
func (c *ClientConnection) readMessage() (Message, error) {
    // Read message size (4 bytes)
    // Read message type (1 byte)
    // Read timestamp (8 bytes)
    // Read payload (size - 9 bytes)
    // Deserialize and return
}

func (c *ClientConnection) writeMessage(msg Message) error {
    // Serialize message
    // Write size (4 bytes)
    // Write type (1 byte)
    // Write timestamp (8 bytes)
    // Write payload
}

func (c *ClientConnection) handleMessage(msg Message) {
    switch msg.Type {
    case MsgTypePlayerInput:
        // Parse player input and queue for game loop
    case MsgTypeDisconnect:
        // Graceful disconnect
        c.Close()
    default:
        log.Printf("Unknown message type: %v", msg.Type)
    }
}
```

### 3. Network Server (pkg/network/server.go)
```go
package network

import (
    "fmt"
    "net"
    "sync"
)

// Server manages client connections and state broadcasting
type Server struct {
    listener  net.Listener
    config    ServerConfig
    
    // Connected clients
    clients   map[uint64]*ClientConnection
    clientsMu sync.RWMutex
    
    // Player ID assignment
    nextPlayerID uint64
    
    // Shutdown coordination
    shutdownChan chan struct{}
    wg           sync.WaitGroup
}

// NewServer creates a new game server
func NewServer(config ServerConfig) (*Server, error) {
    listener, err := net.Listen("tcp", config.Address)
    if err != nil {
        return nil, fmt.Errorf("failed to bind port: %w", err)
    }
    
    return &Server{
        listener:     listener,
        config:       config,
        clients:      make(map[uint64]*ClientConnection),
        nextPlayerID: 1,
        shutdownChan: make(chan struct{}),
    }, nil
}

// AcceptClients listens for new client connections
func (s *Server) AcceptClients() {
    s.wg.Add(1)
    defer s.wg.Done()
    
    log.Printf("Server listening on %s", s.listener.Addr())
    
    for {
        select {
        case <-s.shutdownChan:
            return
        default:
        }
        
        conn, err := s.listener.Accept()
        if err != nil {
            log.Printf("Accept error: %v", err)
            continue
        }
        
        log.Printf("New connection from %s", conn.RemoteAddr())
        
        // Handle connection in goroutine
        go s.handleNewConnection(conn)
    }
}

// handleNewConnection processes authentication and adds client
func (s *Server) handleNewConnection(conn net.Conn) {
    client := NewClientConnection(conn)
    
    // Perform handshake
    if err := s.authenticateClient(client); err != nil {
        log.Printf("Authentication failed: %v", err)
        conn.Close()
        return
    }
    
    // Add to clients map
    s.clientsMu.Lock()
    client.playerID = s.nextPlayerID
    s.nextPlayerID++
    s.clients[client.playerID] = client
    s.clientsMu.Unlock()
    
    log.Printf("Client authenticated: player %d (%s)", client.playerID, client.playerName)
    
    // Start client read/write loops
    s.wg.Add(2)
    go func() {
        defer s.wg.Done()
        client.ReadLoop()
        s.removeClient(client.playerID)
    }()
    go func() {
        defer s.wg.Done()
        client.WriteLoop()
    }()
}

// authenticateClient performs handshake protocol
func (s *Server) authenticateClient(client *ClientConnection) error {
    // Read handshake request
    msg, err := client.readMessage()
    if err != nil {
        return fmt.Errorf("read handshake: %w", err)
    }
    
    if msg.Type != MsgTypeHandshake {
        return fmt.Errorf("expected handshake, got %v", msg.Type)
    }
    
    var req HandshakeRequest
    // Deserialize msg.Payload into req
    
    // Validate protocol version
    if req.ProtocolVersion != CurrentProtocolVersion {
        response := HandshakeResponse{
            Accepted:     false,
            ErrorMessage: "Protocol version mismatch",
        }
        client.writeMessage(Message{Type: MsgTypeAuthenticate, Payload: serialize(response)})
        return fmt.Errorf("protocol mismatch")
    }
    
    // Check player count
    s.clientsMu.RLock()
    count := len(s.clients)
    s.clientsMu.RUnlock()
    
    if count >= s.config.MaxPlayers {
        response := HandshakeResponse{
            Accepted:     false,
            ErrorMessage: "Server full",
        }
        client.writeMessage(Message{Type: MsgTypeAuthenticate, Payload: serialize(response)})
        return fmt.Errorf("server full")
    }
    
    // Accept client
    client.playerName = req.PlayerName
    
    response := HandshakeResponse{
        Accepted:  true,
        PlayerID:  client.playerID,
        WorldSeed: s.config.WorldSeed,
        GenreID:   s.config.GenreID,
    }
    
    return client.writeMessage(Message{Type: MsgTypeAuthenticate, Payload: serialize(response)})
}

// BroadcastState sends world state to all clients
func (s *Server) BroadcastState(snapshot WorldSnapshot) {
    s.clientsMu.RLock()
    defer s.clientsMu.RUnlock()
    
    msg := Message{
        Type:      MsgTypeWorldState,
        Timestamp: snapshot.Timestamp,
        Payload:   serializeSnapshot(snapshot),
    }
    
    for _, client := range s.clients {
        client.SendMessage(msg)
    }
}

// removeClient removes a disconnected client
func (s *Server) removeClient(playerID uint64) {
    s.clientsMu.Lock()
    defer s.clientsMu.Unlock()
    
    if client, ok := s.clients[playerID]; ok {
        log.Printf("Client disconnected: player %d", playerID)
        client.Close()
        delete(s.clients, playerID)
    }
}

// Shutdown gracefully stops the server
func (s *Server) Shutdown() {
    close(s.shutdownChan)
    s.listener.Close()
    
    // Close all client connections
    s.clientsMu.Lock()
    for _, client := range s.clients {
        client.Close()
    }
    s.clientsMu.Unlock()
    
    // Wait for all goroutines
    s.wg.Wait()
}
```

### 4. Update Server Main (cmd/server/main.go)
```go
// After line 95, replace stub message with actual server start

// Create and start network server
server, err := network.NewServer(serverConfig)
if err != nil {
    log.Fatalf("Failed to create server: %v", err)
}

log.Println("Server initialized successfully")
log.Printf("Server listening on port %s", *port)
log.Printf("Max players: %d, Update rate: %d Hz", *maxPlayers, *tickRate)
log.Printf("Game world ready with %d entities", len(world.GetEntities()))

// Start accepting clients in background
go server.AcceptClients()

// Handle graceful shutdown
sigChan := make(chan os.Signal, 1)
signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

// Run authoritative game loop
tickDuration := time.Duration(1000000000 / *tickRate)
ticker := time.NewTicker(tickDuration)
defer ticker.Stop()

lastUpdate := time.Now()

log.Printf("Starting authoritative game loop at %d Hz...", *tickRate)

for {
    select {
    case <-ticker.C:
        // Calculate delta time
        now := time.Now()
        deltaTime := now.Sub(lastUpdate).Seconds()
        lastUpdate = now

        // Update game world
        world.Update(deltaTime)

        // Record snapshot for lag compensation and state sync
        snapshot := buildWorldSnapshot(world, now)
        snapshotManager.AddSnapshot(snapshot)
        lagCompensator.RecordSnapshot(snapshot)

        // Broadcast state to connected clients
        server.BroadcastState(snapshot)

        if *verbose && int(now.Unix())%10 == 0 {
            log.Printf("Server tick: %d entities, %d players",
                len(world.GetEntities()), server.ClientCount())
        }

    case <-sigChan:
        log.Println("Shutting down server...")
        server.Shutdown()
        return
    }
}
```

## Implementation Checklist

### Day 1: Protocol and Client Connection
- [ ] Create `pkg/network/protocol.go`
  - [ ] Define message types (HandshakeRequest, WorldStateMessage, etc.)
  - [ ] Implement serialization/deserialization
  - [ ] Write unit tests for protocol messages
- [ ] Create `pkg/network/client_connection.go`
  - [ ] Implement ClientConnection struct
  - [ ] Implement ReadLoop/WriteLoop
  - [ ] Write unit tests for connection handling

### Day 2: Server Implementation
- [ ] Create `pkg/network/server.go`
  - [ ] Implement Server struct with net.Listener
  - [ ] Implement AcceptClients() goroutine
  - [ ] Implement authentication handshake
  - [ ] Implement BroadcastState()
  - [ ] Implement graceful shutdown
  - [ ] Write unit tests for server logic
- [ ] Update `cmd/server/main.go`
  - [ ] Replace stub message with actual server start
  - [ ] Add signal handling for graceful shutdown
  - [ ] Update logging

### Day 3: Integration and Testing
- [ ] Write integration tests
  - [ ] Test server start/stop
  - [ ] Test client connection/disconnection
  - [ ] Test authentication success/failure
  - [ ] Test state broadcasting
  - [ ] Test max players limit
  - [ ] Test protocol version mismatch
- [ ] Load testing
  - [ ] Connect 4 clients simultaneously
  - [ ] Verify state broadcast rate (20 Hz)
  - [ ] Measure bandwidth usage (<100KB/s per player)
- [ ] Update documentation
  - [ ] Remove "stub" warnings
  - [ ] Document network protocol
  - [ ] Update README.md

## Testing Commands

```bash
# Build server
go build ./cmd/server

# Run server with verbose logging
./venture-server -port 8080 -max-players 4 -verbose

# In another terminal, verify port is listening
netstat -an | grep 8080
# Should show: tcp 0 0 0.0.0.0:8080 0.0.0.0:* LISTEN

# Test with telnet (basic connectivity)
telnet localhost 8080

# Run unit tests
go test -tags test ./pkg/network/...

# Run with race detector
go test -tags test -race ./pkg/network/...

# Measure coverage
go test -tags test -cover ./pkg/network/...
```

## Success Criteria

✅ Server binds to specified port  
✅ Server accepts TCP connections  
✅ Handshake protocol works  
✅ Client authentication succeeds/fails correctly  
✅ State broadcasts at 20 Hz  
✅ Graceful shutdown works  
✅ Unit tests pass with >80% coverage  
✅ Integration tests pass  
✅ `netstat` shows port in LISTEN state  
✅ No "stub" messages in logs  
✅ Network package coverage increases from 66.8% toward target

## Common Pitfalls

❌ **Blocking Accept()** - Run AcceptClients() in goroutine  
❌ **No timeout on reads** - Add read deadlines to detect disconnects  
❌ **Unbuffered sends** - Use buffered channels for sendChan  
❌ **Missing locks** - Protect clients map with RWMutex  
❌ **No graceful shutdown** - Handle SIGINT/SIGTERM signals  
❌ **Infinite message size** - Limit max message size to prevent DoS  
❌ **No protocol versioning** - Check version in handshake  
❌ **Panics in goroutines** - Recover from panics in client loops

## References

- Existing network code: `pkg/network/`
- Current server stub: `cmd/server/main.go:95-142`
- Client-side prediction: `pkg/network/prediction.go`
- Lag compensation: `pkg/network/lag_compensation.go`
- Snapshot management: `pkg/network/snapshot.go`

## Next Steps After Completion

After Task 1.1 is complete:
1. Test multiplayer with actual client connections
2. Move to Task 1.2 (Inventory UI) or Task 2.2 (Audio Integration)
3. Update TASK-TRACKER.md with completion status
4. Write learnings/notes for future reference

---

**Ready to start? Begin with protocol.go and work through the checklist!** 🚀
