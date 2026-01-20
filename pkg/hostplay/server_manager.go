package hostplay

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/opd-ai/venture/pkg/engine"
	"github.com/opd-ai/venture/pkg/network"
	"github.com/opd-ai/venture/pkg/procgen"
	"github.com/opd-ai/venture/pkg/procgen/terrain"
	"github.com/sirupsen/logrus"
)

// ServerConfig contains configuration for the embedded server.
type ServerConfig struct {
	// Port is the starting port to attempt binding to (default 8080).
	// If this port is in use, ports 8081-8089 will be tried as fallbacks.
	Port int

	// MaxPlayers is the maximum number of concurrent players (default 4).
	MaxPlayers int

	// BindLAN controls whether to bind to all interfaces (0.0.0.0) or just localhost (127.0.0.1).
	// Default is false (localhost only) for security.
	BindLAN bool

	// WorldSeed is the seed for deterministic world generation.
	WorldSeed int64

	// GenreID is the genre for procedural generation (e.g., "fantasy", "scifi").
	GenreID string

	// Difficulty is the difficulty level (0.0 to 1.0).
	Difficulty float64

	// TickRate is the server update rate in Hz (default 20).
	TickRate int
}

// ServerManager manages the lifecycle of an in-process game server.
//
// # Lifecycle and Cleanup
//
// ServerManager uses a context-based shutdown pattern where the cancel function
// is stored in cancelFunc and must be explicitly called via Stop(). This design
// provides graceful shutdown control and allows the server to clean up resources
// in a coordinated manner.
//
// Callers MUST ensure Stop() is called when the server is no longer needed,
// ideally using defer immediately after Start():
//
//	sm, err := NewServerManager(config, logger)
//	if err != nil {
//	    return err
//	}
//	if err := sm.Start(); err != nil {
//	    return err
//	}
//	defer sm.Stop() // Ensures cleanup even if panic occurs
//
// Failure to call Stop() may result in goroutine and context resource leaks.
type ServerManager struct {
	config           *ServerConfig
	logger           *logrus.Logger
	server           *network.TCPServer
	world            *engine.World
	snapshotManager  *network.SnapshotManager
	lagCompensator   *network.LagCompensator
	inputHandler     *InputHandler
	stateBroadcaster *StateBroadcaster
	address          string
	port             int
	cancelFunc       context.CancelFunc // Must be called via Stop() for proper cleanup
	wg               sync.WaitGroup
	mu               sync.RWMutex
	running          bool
	generatedTerrain *terrain.Terrain
	readyChan        chan struct{} // Signals when server goroutine is fully initialized
}

// NewServerManager creates a new ServerManager with the given configuration.
func NewServerManager(config *ServerConfig, logger *logrus.Logger) (*ServerManager, error) {
	if config == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}
	if logger == nil {
		return nil, fmt.Errorf("logger cannot be nil")
	}

	// Set defaults
	if config.Port == 0 {
		config.Port = 8080
	}
	if config.MaxPlayers == 0 {
		config.MaxPlayers = 4
	}
	if config.GenreID == "" {
		config.GenreID = "fantasy"
	}
	if config.TickRate == 0 {
		config.TickRate = 20
	}

	return &ServerManager{
		config: config,
		logger: logger,
	}, nil
}

// Start starts the server in a background goroutine and waits until it's listening.
// Returns an error if the server fails to start or bind to any port in the fallback range.
func (sm *ServerManager) Start() error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if sm.running {
		return fmt.Errorf("server is already running")
	}

	// Determine bind address
	bindAddr := "127.0.0.1"
	if sm.config.BindLAN {
		bindAddr = "0.0.0.0"
		sm.logger.Warn("Server will bind to all interfaces (0.0.0.0) - accessible from LAN")
	}

	// Create ECS world with logger
	sm.world = engine.NewWorldWithLogger(sm.logger)

	// Add gameplay systems
	movementSystem := engine.NewMovementSystem(200.0)
	collisionSystem := engine.NewCollisionSystem(64.0)
	combatSystem := engine.NewCombatSystemWithLogger(sm.config.WorldSeed, sm.logger)
	aiSystem := engine.NewAISystem(sm.world)
	progressionSystem := engine.NewProgressionSystem(sm.world)
	inventorySystem := engine.NewInventorySystem(sm.world)

	sm.world.AddSystem(movementSystem)
	sm.world.AddSystem(collisionSystem)
	sm.world.AddSystem(combatSystem)
	sm.world.AddSystem(aiSystem)
	sm.world.AddSystem(progressionSystem)
	sm.world.AddSystem(inventorySystem)

	// Generate world terrain
	terrainGen := terrain.NewBSPGeneratorWithLogger(sm.logger)
	params := procgen.GenerationParams{
		Difficulty: sm.config.Difficulty,
		Depth:      1,
		GenreID:    sm.config.GenreID,
		Custom: map[string]interface{}{
			"width":  100,
			"height": 100,
		},
	}

	terrainResult, err := terrainGen.Generate(sm.config.WorldSeed, params)
	if err != nil {
		return fmt.Errorf("failed to generate terrain: %w", err)
	}
	// Safe type assertion with error handling
	terrain, ok := terrainResult.(*terrain.Terrain)
	if !ok {
		return fmt.Errorf("terrain generator returned unexpected type: %T", terrainResult)
	}
	sm.generatedTerrain = terrain

	sm.logger.WithFields(logrus.Fields{
		"width":     sm.generatedTerrain.Width,
		"height":    sm.generatedTerrain.Height,
		"roomCount": len(sm.generatedTerrain.Rooms),
	}).Info("world terrain generated")

	// Try to bind to a port (with fallback)
	var port int
	var serverConfig network.ServerConfig
	var lastErr error

	maxPort := sm.config.Port + 9 // Try up to 10 ports
	for port = sm.config.Port; port <= maxPort; port++ {
		addr := fmt.Sprintf("%s:%d", bindAddr, port)
		serverConfig = network.DefaultServerConfig()
		serverConfig.Address = addr
		serverConfig.MaxPlayers = sm.config.MaxPlayers
		serverConfig.UpdateRate = sm.config.TickRate

		// Try to create and start server
		sm.server = network.NewServerWithLogger(serverConfig, sm.logger)
		if err := sm.server.Start(); err == nil {
			sm.logger.Info("Server bound to port", "address", addr, "port", port)
			break
		}
		lastErr = err
		sm.logger.Debug("Port in use, trying next", "port", port, "error", err)
		sm.server = nil
	}

	if sm.server == nil {
		return fmt.Errorf("failed to bind to any port in range %d-%d: %w",
			sm.config.Port, maxPort, lastErr)
	}

	sm.port = port
	sm.address = fmt.Sprintf("localhost:%d", port)

	// Create snapshot manager and lag compensator
	sm.snapshotManager = network.NewSnapshotManager(100)
	lagCompConfig := network.DefaultLagCompensationConfig()
	sm.lagCompensator = network.NewLagCompensator(lagCompConfig)

	// Create input handler and state broadcaster
	loggerEntry := sm.logger.WithFields(logrus.Fields{"component": "server"})
	sm.inputHandler = NewInputHandler(sm.world, loggerEntry)
	sm.stateBroadcaster = NewStateBroadcaster(sm.world, sm.config.TickRate, loggerEntry)

	// Create context for shutdown coordination.
	// The cancel function is stored and MUST be called via Stop() for proper cleanup.
	// This pattern enables graceful shutdown with resource cleanup guarantees.
	ctx, cancel := context.WithCancel(context.Background())
	sm.cancelFunc = cancel

	// Create ready channel for server initialization synchronization.
	// Buffered channel of size 1 prevents the serverLoop from blocking on send.
	sm.readyChan = make(chan struct{}, 1)

	// Start server goroutine
	sm.wg.Add(1)
	go sm.serverLoop(ctx)

	// Wait for server goroutine to signal it's fully initialized.
	// This provides proper synchronization instead of arbitrary sleep delays.
	<-sm.readyChan

	// Set running flag after server is fully initialized to avoid race condition
	// where IsRunning() could return true before the server is actually ready.
	sm.running = true

	sm.logger.WithFields(logrus.Fields{
		"address":     sm.address,
		"max_players": sm.config.MaxPlayers,
		"world_seed":  sm.config.WorldSeed,
		"genre":       sm.config.GenreID,
	}).Info("host-and-play server started")

	return nil
}

// serverLoop runs the server in a goroutine until context is cancelled.
func (sm *ServerManager) serverLoop(ctx context.Context) {
	defer sm.wg.Done()

	sm.logger.Debug("Server loop started")

	ticker := time.NewTicker(time.Duration(1000/sm.config.TickRate) * time.Millisecond)
	defer ticker.Stop()

	inputCommands := sm.server.ReceiveInputCommand()
	playerJoins := sm.server.ReceivePlayerJoin()
	playerLeaves := sm.server.ReceivePlayerLeave()
	errors := sm.server.ReceiveError()

	// Signal that server loop is fully initialized and ready to process events.
	// This unblocks the Start() method which waits for this signal.
	sm.readyChan <- struct{}{}

	for {
		select {
		case <-ctx.Done():
			sm.logger.Info("Server shutting down")
			sm.server.Stop()
			return

		case playerID := <-playerJoins:
			sm.handlePlayerJoin(playerID)

		case playerID := <-playerLeaves:
			sm.handlePlayerLeave(playerID)

		case inputCmd := <-inputCommands:
			sm.handleInputCommand(inputCmd)

		case err := <-errors:
			sm.logger.Error("Network error", "error", err)

		case <-ticker.C:
			sm.handleWorldUpdate()
		}
	}
}

// handlePlayerJoin spawns a player entity when a player joins.
func (sm *ServerManager) handlePlayerJoin(playerID uint64) {
	sm.logger.Info("Player joined", "player_id", playerID)
	sm.spawnPlayer(playerID)
}

// handlePlayerLeave removes a player entity when a player leaves.
func (sm *ServerManager) handlePlayerLeave(playerID uint64) {
	sm.logger.Info("Player left", "player_id", playerID)
	sm.removePlayer(playerID)
}

// handleInputCommand processes player input through the InputHandler.
func (sm *ServerManager) handleInputCommand(inputCmd *network.InputCommand) {
	sm.logger.Debug("Received input command", "player_id", inputCmd.PlayerID, "type", inputCmd.InputType)
	sm.inputHandler.ProcessInputRaw(inputCmd.PlayerID, inputCmd.InputType, inputCmd.Data)
}

// handleWorldUpdate updates the game world and broadcasts state to clients.
func (sm *ServerManager) handleWorldUpdate() {
	dt := float64(1.0 / float64(sm.config.TickRate))
	sm.world.Update(dt)

	snapshot, err := sm.stateBroadcaster.CreateSnapshot()
	if err != nil {
		sm.logger.Error("Failed to create snapshot", "error", err)
		return
	}

	if sm.stateBroadcaster.ShouldBroadcast() {
		sm.broadcastEntityStates(snapshot)
	}
}

// spawnPlayer spawns a player entity at a spawn location.
func (sm *ServerManager) spawnPlayer(playerID uint64) {
	// Find a spawn point in one of the rooms
	if len(sm.generatedTerrain.Rooms) == 0 {
		sm.logger.Error("No rooms available for spawn")
		return
	}

	// Use first room's center as spawn point
	room := sm.generatedTerrain.Rooms[0]
	spawnX := float64(room.X + room.Width/2)
	spawnY := float64(room.Y + room.Height/2)

	// Create player entity
	playerEntity := sm.world.CreateEntity()
	playerEntity.AddComponent(&engine.PositionComponent{X: spawnX, Y: spawnY})
	playerEntity.AddComponent(&engine.VelocityComponent{VX: 0, VY: 0})

	// Add combat stats
	playerEntity.AddComponent(&engine.HealthComponent{Current: 100, Max: 100})

	// Add network component to track player ID
	networkComp := &engine.NetworkComponent{PlayerID: playerID}
	playerEntity.AddComponent(networkComp)

	// Register player with input handler
	sm.inputHandler.RegisterPlayer(playerID, playerEntity)

	sm.logger.WithFields(logrus.Fields{
		"player_id": playerID,
		"entity_id": playerEntity.ID,
		"x":         spawnX,
		"y":         spawnY,
	}).Info("Player spawned")
}

// removePlayer removes a player entity from the world.
func (sm *ServerManager) removePlayer(playerID uint64) {
	// Unregister from input handler
	sm.inputHandler.UnregisterPlayer(playerID)

	// Find entity with matching player ID
	for _, entity := range sm.world.GetEntities() {
		netComp, exists := entity.GetComponent("network")
		if exists && netComp != nil {
			// Safe type assertion with comma-ok idiom
			nc, ok := netComp.(*engine.NetworkComponent)
			if !ok {
				continue // Skip if component is wrong type
			}
			if nc.PlayerID == playerID {
				sm.world.RemoveEntity(entity.ID)
				sm.logger.WithFields(logrus.Fields{
					"player_id": playerID,
					"entity_id": entity.ID,
				}).Info("Player entity removed")
				return
			}
		}
	}

	sm.logger.Warn("Could not find entity for player", "player_id", playerID)
}

// broadcastEntityStates sends entity state updates to all connected clients.
func (sm *ServerManager) broadcastEntityStates(snapshot *WorldState) {
	if snapshot == nil || len(snapshot.Entities) == 0 {
		return
	}

	// Convert each entity state to network StateUpdate and broadcast
	for _, entityState := range snapshot.Entities {
		update := sm.convertToStateUpdate(entityState)
		if update != nil {
			sm.server.BroadcastStateUpdate(update)
		}
	}
}

// convertToStateUpdate converts a WorldState EntityState to a network StateUpdate.
func (sm *ServerManager) convertToStateUpdate(entityState EntityState) *network.StateUpdate {
	timestamp, err := network.NowTimestamp()
	if err != nil {
		// Log warning and use zero timestamp (will be corrected by receiver)
		timestamp = 0
	}

	update := &network.StateUpdate{
		Timestamp: timestamp,
		EntityID:  entityState.ID,
		Priority:  network.PriorityNormal,
	}

	// Add position component
	if entityState.Position != nil {
		posComp := &engine.PositionComponent{
			X: entityState.Position.X,
			Y: entityState.Position.Y,
		}
		posData, err := json.Marshal(posComp)
		if err == nil {
			update.Components = append(update.Components, network.ComponentData{
				Type: "position",
				Data: posData,
			})
		}
	}

	// Add velocity component
	if entityState.Velocity != nil {
		velComp := &engine.VelocityComponent{
			VX: entityState.Velocity.VX,
			VY: entityState.Velocity.VY,
		}
		velData, err := json.Marshal(velComp)
		if err == nil {
			update.Components = append(update.Components, network.ComponentData{
				Type: "velocity",
				Data: velData,
			})
		}
	}

	// Add health component
	if entityState.Health != nil {
		healthComp := &engine.HealthComponent{
			Current: entityState.Health.Current,
			Max:     entityState.Health.Max,
		}
		healthData, err := json.Marshal(healthComp)
		if err == nil {
			update.Components = append(update.Components, network.ComponentData{
				Type: "health",
				Data: healthData,
			})
		}
	}

	// Add rotation component
	if entityState.Rotation != nil {
		rotComp := &engine.RotationComponent{
			Angle: entityState.Rotation.Angle,
		}
		rotData, err := json.Marshal(rotComp)
		if err == nil {
			update.Components = append(update.Components, network.ComponentData{
				Type: "rotation",
				Data: rotData,
			})
		}
	}

	// Only send if we have components
	if len(update.Components) == 0 {
		return nil
	}

	return update
}

// Stop gracefully stops the server and waits for the goroutine to exit.
//
// This method calls the stored cancel function to signal shutdown to the server
// goroutine, then waits (with a 5-second timeout) for all background work to complete.
// Stop is idempotent and safe to call multiple times.
//
// Callers should defer Stop() after successful Start() to ensure cleanup:
//
//	sm, err := NewServerManager(config, logger)
//	if err != nil {
//	    return err
//	}
//	if err := sm.Start(); err != nil {
//	    return err
//	}
//	defer sm.Stop() // Ensures cleanup even on panic
//
// Returns an error if shutdown does not complete within 5 seconds.
func (sm *ServerManager) Stop() error {
	sm.mu.Lock()
	if !sm.running {
		sm.mu.Unlock()
		return nil
	}

	sm.logger.Info("Stopping host-and-play server")

	// Signal shutdown
	if sm.cancelFunc != nil {
		sm.cancelFunc()
	}
	sm.running = false
	sm.mu.Unlock()

	// Wait for server goroutine to exit (with timeout)
	done := make(chan struct{})
	go func() {
		sm.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		sm.logger.Info("Host-and-play server stopped cleanly")
		return nil
	case <-time.After(5 * time.Second):
		return fmt.Errorf("server shutdown timeout after 5 seconds")
	}
}

// Address returns the address the server is listening on (e.g., "localhost:8080").
func (sm *ServerManager) Address() string {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.address
}

// Port returns the port the server is listening on.
func (sm *ServerManager) Port() int {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.port
}

// IsRunning returns whether the server is currently running.
func (sm *ServerManager) IsRunning() bool {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.running
}

// GetLANAddress attempts to determine the LAN IP address for clients on other machines.
// Returns empty string if not bound to LAN or if IP cannot be determined.
func (sm *ServerManager) GetLANAddress() string {
	sm.mu.RLock()
	bindLAN := sm.config.BindLAN
	port := sm.port
	sm.mu.RUnlock()

	if !bindLAN {
		return ""
	}

	// Get local IP addresses
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		sm.logger.Warn("Failed to get interface addresses", "error", err)
		return ""
	}

	// Find first non-loopback IPv4 address
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return fmt.Sprintf("%s:%d", ipnet.IP.String(), port)
			}
		}
	}

	return ""
}
