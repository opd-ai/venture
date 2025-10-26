package hostplay

import (
	"context"
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
type ServerManager struct {
	config           *ServerConfig
	logger           *logrus.Logger
	server           *network.TCPServer
	world            *engine.World
	snapshotManager  *network.SnapshotManager
	lagCompensator   *network.LagCompensator
	address          string
	port             int
	cancelFunc       context.CancelFunc
	wg               sync.WaitGroup
	mu               sync.RWMutex
	running          bool
	generatedTerrain *terrain.Terrain
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
	sm.generatedTerrain = terrainResult.(*terrain.Terrain)

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

	// Create context for shutdown
	ctx, cancel := context.WithCancel(context.Background())
	sm.cancelFunc = cancel

	// Start server goroutine
	sm.wg.Add(1)
	go sm.serverLoop(ctx)

	sm.running = true

	// Wait a moment to ensure server is fully initialized
	time.Sleep(100 * time.Millisecond)

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

	// Channels from network server
	inputCommands := sm.server.ReceiveInputCommand()
	playerJoins := sm.server.ReceivePlayerJoin()
	playerLeaves := sm.server.ReceivePlayerLeave()
	errors := sm.server.ReceiveError()

	for {
		select {
		case <-ctx.Done():
			sm.logger.Info("Server shutting down")
			sm.server.Stop()
			return

		case playerID := <-playerJoins:
			sm.logger.Info("Player joined", "player_id", playerID)
			// Spawn player entity at a random spawn point
			sm.spawnPlayer(playerID)

		case playerID := <-playerLeaves:
			sm.logger.Info("Player left", "player_id", playerID)
			// Remove player entity from world
			sm.removePlayer(playerID)

		case inputCmd := <-inputCommands:
			sm.logger.Debug("Received input command", "player_id", inputCmd.PlayerID, "type", inputCmd.InputType)
			// Process player input (TODO: implement input handling)

		case err := <-errors:
			sm.logger.Error("Network error", "error", err)

		case <-ticker.C:
			// Update game world
			dt := float64(1.0 / float64(sm.config.TickRate))
			sm.world.Update(dt)

			// TODO: Broadcast state updates to clients
			// For now, we just run the simulation
		}
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

	sm.logger.WithFields(logrus.Fields{
		"player_id": playerID,
		"entity_id": playerEntity.ID,
		"x":         spawnX,
		"y":         spawnY,
	}).Info("Player spawned")
}

// removePlayer removes a player entity from the world.
func (sm *ServerManager) removePlayer(playerID uint64) {
	// Find entity with matching player ID
	for _, entity := range sm.world.GetEntities() {
		netComp, exists := entity.GetComponent("network")
		if exists && netComp != nil {
			nc := netComp.(*engine.NetworkComponent)
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

// Stop gracefully stops the server and waits for the goroutine to exit.
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
