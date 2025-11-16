package federation

import (
	"testing"
	"time"

	"github.com/opd-ai/venture/pkg/engine"
)

// TestAcceptance_TransferTime verifies transfer completes within time limits
func TestAcceptance_TransferTime(t *testing.T) {
	tm := NewTransferManager()
	world := engine.NewWorld()

	// Create player with all components
	player := world.CreateEntity()
	player.AddComponent(&engine.PositionComponent{X: 100, Y: 200})
	player.AddComponent(&engine.HealthComponent{Current: 80, Max: 100})
	player.AddComponent(engine.NewInventoryComponent(20, 100.0))
	world.Update(0) // Process entity addition

	start := time.Now()

	// Prepare transfer
	_, err := tm.PrepareTransfer(player.ID, world, "server-2")
	if err != nil {
		t.Fatalf("PrepareTransfer() error = %v", err)
	}

	// Begin transfer
	err = tm.BeginTransfer(player.ID, "server-1")
	if err != nil {
		t.Fatalf("BeginTransfer() error = %v", err)
	}

	// Confirm transfer
	err = tm.ConfirmTransfer(player.ID)
	if err != nil {
		t.Fatalf("ConfirmTransfer() error = %v", err)
	}

	elapsed := time.Since(start)

	// At 200ms latency, target is <5s, but local operation should be <100ms
	if elapsed > 100*time.Millisecond {
		t.Errorf("Transfer took %v, want <100ms for local operation", elapsed)
	}
}

// TestAcceptance_RollbackRate verifies rollback success rate
func TestAcceptance_RollbackRate(t *testing.T) {
	tm := NewTransferManager()
	world := engine.NewWorld()

	totalTransfers := 100
	failures := 0

	for i := 0; i < totalTransfers; i++ {
		player := world.CreateEntity()
		player.AddComponent(&engine.PositionComponent{X: 100, Y: 200})
		player.AddComponent(&engine.HealthComponent{Current: 80, Max: 100})
		player.AddComponent(engine.NewInventoryComponent(20, 100.0))
		world.Update(0) // Process entity addition

		// Prepare transfer
		_, err := tm.PrepareTransfer(player.ID, world, "server-2")
		if err != nil {
			t.Fatalf("PrepareTransfer() error = %v", err)
		}

		// Simulate 5% failure rate
		if i%20 == 0 {
			tm.RollbackTransfer(player.ID, "simulated failure")

			// Verify player can be restored
			err = tm.RestorePlayerState(player.ID, world)
			if err != nil {
				failures++
			}
		} else {
			tm.BeginTransfer(player.ID, "server-1")
			tm.ConfirmTransfer(player.ID)
		}
	}

	// Target: <1% rollback failure rate
	failureRate := float64(failures) / float64(totalTransfers/20)
	if failureRate > 0.01 {
		t.Errorf("Rollback failure rate = %.2f%%, want <1%%", failureRate*100)
	}
}

// TestAcceptance_StateValidation verifies all validation rules
func TestAcceptance_StateValidation(t *testing.T) {
	tm := NewTransferManager()

	tests := []struct {
		name    string
		state   *PlayerState
		wantErr bool
	}{
		{
			name: "valid state",
			state: &PlayerState{
				PlayerID: 123,
				Position: &PositionData{X: 100, Y: 200},
				Health:   &HealthData{Current: 80, Max: 100},
				Inventory: []ItemData{
					{ItemID: "sword-1", TypeName: "Sword", Quantity: 1},
				},
			},
			wantErr: false,
		},
		{
			name: "health within 10% variance",
			state: &PlayerState{
				PlayerID: 123,
				Health:   &HealthData{Current: 109, Max: 100},
			},
			wantErr: false,
		},
		{
			name: "health exceeds 10% variance",
			state: &PlayerState{
				PlayerID: 123,
				Health:   &HealthData{Current: 120, Max: 100},
			},
			wantErr: true,
		},
		{
			name: "100 items allowed",
			state: &PlayerState{
				PlayerID:  123,
				Inventory: make([]ItemData, 100),
			},
			wantErr: false,
		},
		{
			name: "101 items rejected",
			state: &PlayerState{
				PlayerID:  123,
				Inventory: make([]ItemData, 101),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tm.ValidatePlayerState(tt.state)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidatePlayerState() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestAcceptance_SessionTokenExpiry verifies token lifecycle
func TestAcceptance_SessionTokenExpiry(t *testing.T) {
	am := NewAuthManager()
	am.SetTTL(1 * time.Second)

	// Create token
	token, err := am.CreateSessionToken(123, "server-1")
	if err != nil {
		t.Fatalf("CreateSessionToken() error = %v", err)
	}

	// Should be valid immediately
	_, err = am.ValidateToken(token.Token)
	if err != nil {
		t.Errorf("ValidateToken() error = %v immediately after creation", err)
	}

	// Wait for expiry
	time.Sleep(2 * time.Second)

	// Should be invalid after expiry
	_, err = am.ValidateToken(token.Token)
	if err == nil {
		t.Error("Expected error for expired token, got nil")
	}
}

// TestAcceptance_NonceReplayPrevention verifies replay attack prevention
func TestAcceptance_NonceReplayPrevention(t *testing.T) {
	am := NewAuthManager()

	token, err := am.CreateSessionToken(123, "server-1")
	if err != nil {
		t.Fatalf("CreateSessionToken() error = %v", err)
	}

	// First validation should succeed
	err = am.ValidateNonce(token.Nonce)
	if err != nil {
		t.Errorf("First ValidateNonce() error = %v", err)
	}

	// Mark as used
	err = am.MarkNonceUsed(token.Nonce)
	if err != nil {
		t.Errorf("MarkNonceUsed() error = %v", err)
	}

	// Second validation should fail
	err = am.ValidateNonce(token.Nonce)
	if err == nil {
		t.Error("Expected error for used nonce, got nil")
	}
}

// TestAcceptance_PortalActivation verifies portal activation workflow
func TestAcceptance_PortalActivation(t *testing.T) {
	world := engine.NewWorld()
	federation := NewFederationProtocol("server-1")
	ps := NewPortalSystem(world, federation)
	authMgr := NewAuthManager()
	transferMgr := NewTransferManager()

	// Create player
	player := world.CreateEntity()
	player.AddComponent(&engine.PositionComponent{X: 100, Y: 100})
	player.AddComponent(&engine.HealthComponent{Current: 80, Max: 100})
	player.AddComponent(engine.NewInventoryComponent(20, 100.0))
	world.Update(0) // Process entity addition

	// Test local portal
	t.Run("local teleport", func(t *testing.T) {
		portal := world.CreateEntity()
		portal.AddComponent(&engine.PortalComponent{
			DestinationServer: "local",
			DestinationX:      500,
			DestinationY:      600,
		})
		portal.AddComponent(&engine.PositionComponent{X: 100, Y: 100})
		world.Update(0) // Process portal addition

		err := ps.ActivatePortal(player.ID, portal.ID, authMgr, transferMgr)
		if err != nil {
			t.Errorf("ActivatePortal() error = %v", err)
		}

		posCompRaw, _ := player.GetComponent("position")
		posComp := posCompRaw.(*engine.PositionComponent)
		if posComp.X != 500 || posComp.Y != 600 {
			t.Errorf("Position = (%v, %v), want (500, 600)", posComp.X, posComp.Y)
		}
	})

	// Test cross-server portal
	t.Run("cross-server transfer", func(t *testing.T) {
		portal := world.CreateEntity()
		portal.AddComponent(&engine.PortalComponent{
			DestinationServer: "server-2",
			DestinationX:      700,
			DestinationY:      800,
		})
		portal.AddComponent(&engine.PositionComponent{X: 100, Y: 100})
		world.Update(0) // Process portal addition

		err := ps.ActivatePortal(player.ID, portal.ID, authMgr, transferMgr)
		if err != nil {
			t.Errorf("ActivatePortal() error = %v", err)
		}

		transfer, exists := transferMgr.GetTransfer(player.ID)
		if !exists {
			t.Error("Transfer not found")
			return
		}
		if transfer == nil {
			t.Error("Transfer is nil")
			return
		}

		if transfer.Phase != TransferPhaseTransfer {
			t.Errorf("Phase = %v, want %v", transfer.Phase, TransferPhaseTransfer)
		}
	})
}

// TestAcceptance_TransferTimeout verifies timeout detection
func TestAcceptance_TransferTimeout(t *testing.T) {
	tm := NewTransferManager()
	tm.SetTransferTimeout(1 * time.Second)
	world := engine.NewWorld()

	// Create multiple players
	playerIDs := make([]uint64, 3)
	for i := range playerIDs {
		player := world.CreateEntity()
		player.AddComponent(&engine.PositionComponent{X: 100, Y: 200})
		player.AddComponent(&engine.HealthComponent{Current: 80, Max: 100})
		player.AddComponent(engine.NewInventoryComponent(20, 100.0))
		playerIDs[i] = player.ID
	}
	world.Update(0) // Process all entity additions

	for _, playerID := range playerIDs {
		tm.PrepareTransfer(playerID, world, "server-2")
	}

	// Wait for timeout
	time.Sleep(2 * time.Second)

	// Check for timeouts
	expired := tm.CheckTimeouts()
	if len(expired) != 3 {
		t.Errorf("CheckTimeouts() returned %d expired, want 3", len(expired))
	}
}

// TestAcceptance_ConcurrentTransfers verifies concurrent transfer handling
func TestAcceptance_ConcurrentTransfers(t *testing.T) {
	tm := NewTransferManager()
	world := engine.NewWorld()

	// Create 10 players
	players := make([]uint64, 10)
	for i := range players {
		player := world.CreateEntity()
		player.AddComponent(&engine.PositionComponent{X: 100, Y: 200})
		player.AddComponent(&engine.HealthComponent{Current: 80, Max: 100})
		player.AddComponent(engine.NewInventoryComponent(20, 100.0))
		players[i] = player.ID
	}
	world.Update(0) // Process all entity additions

	// Prepare all transfers concurrently
	done := make(chan bool)
	for _, playerID := range players {
		go func(pid uint64) {
			_, err := tm.PrepareTransfer(pid, world, "server-2")
			if err != nil {
				t.Errorf("PrepareTransfer() error = %v", err)
			}
			done <- true
		}(playerID)
	}

	// Wait for all to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify all transfers were created
	for _, playerID := range players {
		if _, exists := tm.GetTransfer(playerID); !exists {
			t.Errorf("Transfer for player %d not found", playerID)
		}
	}
}

// Benchmarks

func BenchmarkAcceptance_FullTransferCycle(b *testing.B) {
	tm := NewTransferManager()
	world := engine.NewWorld()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		player := world.CreateEntity()
		player.AddComponent(&engine.PositionComponent{X: 100, Y: 200})
		player.AddComponent(&engine.HealthComponent{Current: 80, Max: 100})
		player.AddComponent(engine.NewInventoryComponent(20, 100.0))
		b.StartTimer()

		tm.PrepareTransfer(player.ID, world, "server-2")
		tm.BeginTransfer(player.ID, "server-1")
		tm.ConfirmTransfer(player.ID)
	}
}

func BenchmarkAcceptance_PortalActivation(b *testing.B) {
	world := engine.NewWorld()
	federation := NewFederationProtocol("server-1")
	ps := NewPortalSystem(world, federation)
	authMgr := NewAuthManager()
	transferMgr := NewTransferManager()

	player := world.CreateEntity()
	player.AddComponent(&engine.PositionComponent{X: 100, Y: 100})

	portal := world.CreateEntity()
	portal.AddComponent(&engine.PortalComponent{
		DestinationServer: "local",
		DestinationX:      500,
		DestinationY:      600,
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ps.ActivatePortal(player.ID, portal.ID, authMgr, transferMgr)
	}
}
