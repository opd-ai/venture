package federation

import (
	"testing"

	"github.com/opd-ai/venture/pkg/engine"
)

// TestNewFederationProtocol tests creation of federation protocol handler
func TestNewFederationProtocol(t *testing.T) {
	tests := []struct {
		name     string
		serverID string
	}{
		{"valid server ID", "server-123"},
		{"empty server ID", ""},
		{"long server ID", "server-with-a-very-long-identifier-0123456789"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fp := NewFederationProtocol(tt.serverID, testIdentity())
			if fp == nil {
				t.Fatal("NewFederationProtocol returned nil")
			}
			if fp.serverID != tt.serverID {
				t.Errorf("serverID = %v, want %v", fp.serverID, tt.serverID)
			}
			if fp.peers == nil {
				t.Error("peers map is nil")
			}
			if len(fp.peers) != 0 {
				t.Errorf("peers map not empty, got %d entries", len(fp.peers))
			}
		})
	}
}

func TestFederationProtocolCloseClosesHandshakeManager(t *testing.T) {
	fp := NewFederationProtocol("test-server", testIdentity())

	if err := fp.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}

	select {
	case <-fp.handshake.stopCh:
	default:
		t.Fatal("expected handshake manager stop channel to be closed")
	}
}

// TestConnect tests peer connection functionality
func TestConnect(t *testing.T) {
	tests := []struct {
		name        string
		peerAddress string
		wantErr     bool
	}{
		{"valid address", "192.168.1.100:8080", true}, // true because not implemented
		{"localhost", "localhost:8080", true},
		{"empty address", "", true},
		{"invalid format", "not-a-valid-address", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fp := NewFederationProtocol("test-server", testIdentity())
			err := fp.Connect(tt.peerAddress)
			if (err != nil) != tt.wantErr {
				t.Errorf("Connect() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestTransferPlayer tests player transfer functionality
func TestTransferPlayer(t *testing.T) {
	tests := []struct {
		name         string
		playerID     uint64
		targetServer string
		wantErr      bool
	}{
		{"valid transfer", 1, "target-server", true}, // true because stub not fully implemented
		{"zero player ID", 0, "target-server", true},
		{"empty target", 1, "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fp := NewFederationProtocol("test-server", testIdentity())
			world := engine.NewWorld()
			authMgr := NewAuthManager()
			transferMgr := NewTransferManager()

			err := fp.TransferPlayer(tt.playerID, world, tt.targetServer, authMgr, transferMgr)
			if (err != nil) != tt.wantErr {
				t.Errorf("TransferPlayer() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestNewPortalSystem_TableDriven tests creation of portal system with various parameter combinations
func TestNewPortalSystem_TableDriven(t *testing.T) {
	world := engine.NewWorld()
	fp := NewFederationProtocol("test-server", testIdentity())

	tests := []struct {
		name       string
		world      *engine.World
		federation *FederationProtocol
	}{
		{"valid parameters", world, fp},
		{"nil world", nil, fp},
		{"nil federation", world, nil},
		{"both nil", nil, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ps := NewPortalSystem(tt.world, tt.federation)
			if ps == nil {
				t.Fatal("NewPortalSystem returned nil")
			}
			if ps.world != tt.world {
				t.Error("world not set correctly")
			}
			if ps.federation != tt.federation {
				t.Error("federation not set correctly")
			}
		})
	}
}

// TestPortalSystemUpdate tests portal system update loop
func TestPortalSystemUpdate(t *testing.T) {
	world := engine.NewWorld()
	fp := NewFederationProtocol("test-server", testIdentity())
	ps := NewPortalSystem(world, fp)

	tests := []struct {
		name      string
		deltaTime float64
		setup     func(*engine.World)
	}{
		{
			name:      "no portals",
			deltaTime: 0.016,
			setup:     func(w *engine.World) {},
		},
		{
			name:      "portal without position",
			deltaTime: 0.016,
			setup: func(w *engine.World) {
				entity := w.CreateEntity()
				entity.AddComponent(&engine.PortalComponent{
					DestinationServer: "local",
					DestinationX:      100,
					DestinationY:      200,
				})
			},
		},
		{
			name:      "portal with position",
			deltaTime: 0.016,
			setup: func(w *engine.World) {
				entity := w.CreateEntity()
				entity.AddComponent(&engine.PositionComponent{X: 50, Y: 50})
				entity.AddComponent(&engine.PortalComponent{
					DestinationServer: "local",
					DestinationX:      100,
					DestinationY:      200,
				})
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset world
			world = engine.NewWorld()
			ps.world = world
			tt.setup(world)

			// Should not panic
			ps.Update(tt.deltaTime)
		})
	}
}

// TestActivatePortal tests portal activation
func TestActivatePortal(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(*engine.World) (playerID, portalID uint64)
		wantErr bool
		errMsg  string
	}{
		{
			name: "portal not found",
			setup: func(w *engine.World) (uint64, uint64) {
				player := w.CreateEntity()
				return player.ID, 9999 // non-existent portal
			},
			wantErr: true,
			errMsg:  "portal not found",
		},
		{
			name: "entity is not a portal",
			setup: func(w *engine.World) (uint64, uint64) {
				player := w.CreateEntity()
				notPortal := w.CreateEntity()
				notPortal.AddComponent(&engine.PositionComponent{X: 10, Y: 10})
				return player.ID, notPortal.ID
			},
			wantErr: true,
			errMsg:  "entity is not a portal",
		},
		{
			name: "local teleport - player not found",
			setup: func(w *engine.World) (uint64, uint64) {
				portal := w.CreateEntity()
				portal.AddComponent(&engine.PortalComponent{
					DestinationServer: "local",
					DestinationX:      100,
					DestinationY:      200,
				})
				return 9999, portal.ID // non-existent player
			},
			wantErr: true,
			errMsg:  "player not found",
		},
		{
			name: "local teleport - player has no position",
			setup: func(w *engine.World) (uint64, uint64) {
				player := w.CreateEntity()
				portal := w.CreateEntity()
				portal.AddComponent(&engine.PortalComponent{
					DestinationServer: "local",
					DestinationX:      100,
					DestinationY:      200,
				})
				return player.ID, portal.ID
			},
			wantErr: true,
			errMsg:  "player has no position",
		},
		{
			name: "local teleport - success",
			setup: func(w *engine.World) (uint64, uint64) {
				player := w.CreateEntity()
				player.AddComponent(&engine.PositionComponent{X: 10, Y: 10})
				portal := w.CreateEntity()
				portal.AddComponent(&engine.PortalComponent{
					DestinationServer: "local",
					DestinationX:      100,
					DestinationY:      200,
				})
				return player.ID, portal.ID
			},
			wantErr: false,
		},
		{
			name: "cross-server transfer - preparation succeeds",
			setup: func(w *engine.World) (uint64, uint64) {
				player := w.CreateEntity()
				player.AddComponent(&engine.PositionComponent{X: 10, Y: 10})
				portal := w.CreateEntity()
				portal.AddComponent(&engine.PortalComponent{
					DestinationServer: "remote-server",
					DestinationX:      100,
					DestinationY:      200,
				})
				return player.ID, portal.ID
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			world := engine.NewWorld()
			fp := NewFederationProtocol("test-server", testIdentity())
			ps := NewPortalSystem(world, fp)
			authMgr := NewAuthManager()
			transferMgr := NewTransferManager()

			playerID, portalID := tt.setup(world)
			// Process pending entity additions
			world.Update(0)
			err := ps.ActivatePortal(playerID, portalID, authMgr, transferMgr)

			if (err != nil) != tt.wantErr {
				t.Errorf("ActivatePortal() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && tt.errMsg != "" {
				if err == nil {
					t.Errorf("expected error containing %q, got nil", tt.errMsg)
				} else if err.Error() != tt.errMsg {
					t.Errorf("error message = %q, want %q", err.Error(), tt.errMsg)
				}
			}

			// If successful local teleport, verify position changed
			if !tt.wantErr && tt.name == "local teleport - success" {
				player, ok := world.GetEntity(playerID)
				if !ok {
					t.Fatal("player not found after teleport")
				}
				posComp, ok := player.GetComponent("position")
				if !ok {
					t.Fatal("player lost position component after teleport")
				}
				pos := posComp.(*engine.PositionComponent)
				portal, _ := world.GetEntity(portalID)
				portalComp, _ := portal.GetComponent("portal")
				expectedPos := portalComp.(*engine.PortalComponent)
				if pos.X != expectedPos.DestinationX || pos.Y != expectedPos.DestinationY {
					t.Errorf("player position = (%v, %v), want (%v, %v)",
						pos.X, pos.Y, expectedPos.DestinationX, expectedPos.DestinationY)
				}
			}
		})
	}
}

// TestLocalTeleport tests the local teleport helper function
func TestLocalTeleport(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(*engine.World) (playerID uint64, destX, destY float64)
		wantErr  bool
		checkPos bool
	}{
		{
			name: "valid teleport",
			setup: func(w *engine.World) (uint64, float64, float64) {
				player := w.CreateEntity()
				player.AddComponent(&engine.PositionComponent{X: 10, Y: 20})
				return player.ID, 100.5, 200.5
			},
			wantErr:  false,
			checkPos: true,
		},
		{
			name: "player not found",
			setup: func(w *engine.World) (uint64, float64, float64) {
				return 9999, 100, 200
			},
			wantErr:  true,
			checkPos: false,
		},
		{
			name: "player without position component",
			setup: func(w *engine.World) (uint64, float64, float64) {
				player := w.CreateEntity()
				return player.ID, 100, 200
			},
			wantErr:  true,
			checkPos: false,
		},
		{
			name: "teleport to zero coordinates",
			setup: func(w *engine.World) (uint64, float64, float64) {
				player := w.CreateEntity()
				player.AddComponent(&engine.PositionComponent{X: 50, Y: 50})
				return player.ID, 0, 0
			},
			wantErr:  false,
			checkPos: true,
		},
		{
			name: "teleport to negative coordinates",
			setup: func(w *engine.World) (uint64, float64, float64) {
				player := w.CreateEntity()
				player.AddComponent(&engine.PositionComponent{X: 50, Y: 50})
				return player.ID, -10, -20
			},
			wantErr:  false,
			checkPos: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			world := engine.NewWorld()
			fp := NewFederationProtocol("test-server", testIdentity())
			ps := NewPortalSystem(world, fp)

			playerID, destX, destY := tt.setup(world)
			// Process pending entity additions
			world.Update(0)
			err := ps.localTeleport(playerID, destX, destY)

			if (err != nil) != tt.wantErr {
				t.Errorf("localTeleport() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.checkPos {
				player, ok := world.GetEntity(playerID)
				if !ok {
					t.Fatal("player not found after teleport")
				}
				posComp, ok := player.GetComponent("position")
				if !ok {
					t.Fatal("position component not found")
				}
				pos := posComp.(*engine.PositionComponent)
				if pos.X != destX || pos.Y != destY {
					t.Errorf("position = (%v, %v), want (%v, %v)", pos.X, pos.Y, destX, destY)
				}
			}
		})
	}
}

// TestPeerServerFields tests PeerServer struct fields
func TestPeerServerFields(t *testing.T) {
	peer := &PeerServer{
		ID:            "peer-123",
		Name:          "Test Peer",
		Address:       "192.168.1.100:8080",
		TrustLevel:    "trusted",
		Connected:     true,
		LastHeartbeat: 1234567890,
	}

	if peer.ID != "peer-123" {
		t.Errorf("ID = %v, want peer-123", peer.ID)
	}
	if peer.Name != "Test Peer" {
		t.Errorf("Name = %v, want Test Peer", peer.Name)
	}
	if peer.Address != "192.168.1.100:8080" {
		t.Errorf("Address = %v, want 192.168.1.100:8080", peer.Address)
	}
	if peer.TrustLevel != "trusted" {
		t.Errorf("TrustLevel = %v, want trusted", peer.TrustLevel)
	}
	if !peer.Connected {
		t.Error("Connected should be true")
	}
	if peer.LastHeartbeat != 1234567890 {
		t.Errorf("LastHeartbeat = %v, want 1234567890", peer.LastHeartbeat)
	}
}

// BenchmarkNewFederationProtocol benchmarks federation protocol creation
func BenchmarkNewFederationProtocol(b *testing.B) {
	for i := 0; i < b.N; i++ {
		NewFederationProtocol("test-server", testIdentity())
	}
}

// BenchmarkNewPortalSystem benchmarks portal system creation
func BenchmarkNewPortalSystem(b *testing.B) {
	world := engine.NewWorld()
	fp := NewFederationProtocol("test-server", testIdentity())

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		NewPortalSystem(world, fp)
	}
}

// BenchmarkLocalTeleport benchmarks local teleportation
func BenchmarkLocalTeleport(b *testing.B) {
	world := engine.NewWorld()
	fp := NewFederationProtocol("test-server", testIdentity())
	ps := NewPortalSystem(world, fp)

	player := world.CreateEntity()
	player.AddComponent(&engine.PositionComponent{X: 10, Y: 20})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ps.localTeleport(player.ID, 100, 200)
	}
}

// BenchmarkPortalSystemUpdate benchmarks portal system update
func BenchmarkPortalSystemUpdate(b *testing.B) {
	world := engine.NewWorld()
	fp := NewFederationProtocol("test-server", testIdentity())
	ps := NewPortalSystem(world, fp)

	// Create some portals
	for i := 0; i < 10; i++ {
		entity := world.CreateEntity()
		entity.AddComponent(&engine.PositionComponent{X: float64(i * 10), Y: float64(i * 10)})
		entity.AddComponent(&engine.PortalComponent{
			DestinationServer: "local",
			DestinationX:      float64(i * 100),
			DestinationY:      float64(i * 100),
		})
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ps.Update(0.016)
	}
}

// TestGuildUpdateMessage tests guild update message serialization
func TestGuildUpdateMessage(t *testing.T) {
	tests := []struct {
		name      string
		guildID   string
		guildData []byte
		wantErr   bool
	}{
		{"valid message", "guild-123", []byte("compressed-data"), false},
		{"empty guild ID", "", []byte("data"), false},
		{"empty data", "guild-123", []byte{}, false},
		{"large data", "guild-456", make([]byte, 10000), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := &GuildUpdateMessage{
				Type:      "guild_update",
				GuildID:   tt.guildID,
				GuildData: tt.guildData,
				Timestamp: 1234567890,
			}

			if msg.Type != "guild_update" {
				t.Errorf("Type = %v, want guild_update", msg.Type)
			}
			if msg.GuildID != tt.guildID {
				t.Errorf("GuildID = %v, want %v", msg.GuildID, tt.guildID)
			}
			if len(msg.GuildData) != len(tt.guildData) {
				t.Errorf("GuildData length = %v, want %v", len(msg.GuildData), len(tt.guildData))
			}
		})
	}
}

// TestBroadcastGuildUpdate tests broadcasting guild updates to peers
func TestBroadcastGuildUpdate(t *testing.T) {
	fp := NewFederationProtocol("test-server", testIdentity())

	tests := []struct {
		name      string
		guildID   string
		guildData []byte
		wantErr   bool
	}{
		{"valid broadcast with no peers", "guild-123", []byte("data"), false},
		{"empty guild ID", "", []byte("data"), false},
		{"empty data", "guild-123", []byte{}, false},
		{"nil data", "guild-123", nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := fp.BroadcastGuildUpdate(tt.guildID, tt.guildData)
			if (err != nil) != tt.wantErr {
				t.Errorf("BroadcastGuildUpdate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestReceiveGuildUpdate tests receiving and validating guild updates
func TestReceiveGuildUpdate(t *testing.T) {
	fp := NewFederationProtocol("test-server", testIdentity())

	tests := []struct {
		name    string
		msg     *GuildUpdateMessage
		wantErr bool
	}{
		{
			name: "valid message",
			msg: &GuildUpdateMessage{
				Type:      "guild_update",
				GuildID:   "guild-123",
				GuildData: []byte("compressed-data"),
				Timestamp: 1234567890,
			},
			wantErr: false,
		},
		{
			name: "invalid type",
			msg: &GuildUpdateMessage{
				Type:      "wrong_type",
				GuildID:   "guild-123",
				GuildData: []byte("data"),
			},
			wantErr: true,
		},
		{
			name: "missing guild ID",
			msg: &GuildUpdateMessage{
				Type:      "guild_update",
				GuildID:   "",
				GuildData: []byte("data"),
			},
			wantErr: true,
		},
		{
			name: "missing guild data",
			msg: &GuildUpdateMessage{
				Type:    "guild_update",
				GuildID: "guild-123",
			},
			wantErr: true,
		},
		{
			name: "empty guild data",
			msg: &GuildUpdateMessage{
				Type:      "guild_update",
				GuildID:   "guild-123",
				GuildData: []byte{},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := fp.ReceiveGuildUpdate(tt.msg)
			if (err != nil) != tt.wantErr {
				t.Errorf("ReceiveGuildUpdate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// BenchmarkBroadcastGuildUpdate benchmarks guild update broadcasting
func BenchmarkBroadcastGuildUpdate(b *testing.B) {
	fp := NewFederationProtocol("test-server", testIdentity())
	guildData := make([]byte, 1000) // 1KB guild data

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = fp.BroadcastGuildUpdate("guild-123", guildData)
	}
}

// BenchmarkReceiveGuildUpdate benchmarks guild update reception
func BenchmarkReceiveGuildUpdate(b *testing.B) {
	fp := NewFederationProtocol("test-server", testIdentity())
	msg := &GuildUpdateMessage{
		Type:      "guild_update",
		GuildID:   "guild-123",
		GuildData: make([]byte, 1000),
		Timestamp: 1234567890,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = fp.ReceiveGuildUpdate(msg)
	}
}
