package federation

import (
	"testing"
	"time"

	"github.com/opd-ai/venture/pkg/engine"
)

func TestTransferPhaseString(t *testing.T) {
	tests := []struct {
		phase    TransferPhase
		expected string
	}{
		{TransferPhasePrepare, "Prepare"},
		{TransferPhaseTransfer, "Transfer"},
		{TransferPhaseConfirm, "Confirm"},
		{TransferPhaseRollback, "Rollback"},
		{TransferPhase(99), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if got := tt.phase.String(); got != tt.expected {
				t.Errorf("TransferPhase.String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestNewTransferManager(t *testing.T) {
	tm := NewTransferManager()

	if tm == nil {
		t.Fatal("NewTransferManager() returned nil")
	}

	if tm.activeTransfers == nil {
		t.Error("activeTransfers map not initialized")
	}

	if tm.stateBackups == nil {
		t.Error("stateBackups map not initialized")
	}

	if tm.transferTimeout != 60*time.Second {
		t.Errorf("transferTimeout = %v, want 60s", tm.transferTimeout)
	}
}

func TestSetTransferTimeout(t *testing.T) {
	tm := NewTransferManager()
	newTimeout := 30 * time.Second

	tm.SetTransferTimeout(newTimeout)

	if tm.transferTimeout != newTimeout {
		t.Errorf("transferTimeout = %v, want %v", tm.transferTimeout, newTimeout)
	}
}

func TestSetCallbacks(t *testing.T) {
	tm := NewTransferManager()

	_ = false // startCalled
	_ = false // commitCalled
	_ = false // failCalled

	tm.SetCallbacks(
		func(playerID uint64, targetServer string) { /* startCalled = true */ },
		func(playerID uint64) { /* commitCalled = true */ },
		func(playerID uint64, reason string) { /* failCalled = true */ },
	)

	if tm.onTransferStart == nil {
		t.Error("onTransferStart not set")
	}
	if tm.onTransferCommit == nil {
		t.Error("onTransferCommit not set")
	}
	if tm.onTransferFail == nil {
		t.Error("onTransferFail not set")
	}
}

func TestPrepareTransfer(t *testing.T) {
	tm := NewTransferManager()
	world := engine.NewWorld()

	// Create a player entity
	player := world.CreateEntity()
	player.AddComponent(&engine.PositionComponent{X: 100, Y: 200})
	player.AddComponent(&engine.HealthComponent{Current: 80, Max: 100})
	player.AddComponent(&engine.InventoryComponent{Items: nil, MaxItems: 20})

	world.Update(0) // Ensure entity is in world
	transfer, err := tm.PrepareTransfer(player.ID, world, "server-2")
	if err != nil {
		t.Fatalf("PrepareTransfer() error = %v", err)
	}

	if transfer == nil {
		t.Fatal("PrepareTransfer() returned nil transfer")
	}

	if transfer.Phase != TransferPhasePrepare {
		t.Errorf("Phase = %v, want %v", transfer.Phase, TransferPhasePrepare)
	}

	if transfer.PlayerState == nil {
		t.Error("PlayerState is nil")
	}

	if transfer.StateHash == "" {
		t.Error("StateHash is empty")
	}

	if transfer.TargetID != "server-2" {
		t.Errorf("TargetID = %v, want server-2", transfer.TargetID)
	}

	// Verify state backup created
	if _, exists := tm.stateBackups[player.ID]; !exists {
		t.Error("State backup not created")
	}
}

func TestPrepareTransfer_AlreadyInProgress(t *testing.T) {
	tm := NewTransferManager()
	world := engine.NewWorld()

	player := world.CreateEntity()
	player.AddComponent(&engine.PositionComponent{X: 100, Y: 200})
	player.AddComponent(&engine.HealthComponent{Current: 80, Max: 100})
	player.AddComponent(engine.NewInventoryComponent(20, 100.0))
	world.Update(0) // Process entity addition

	// First transfer
	world.Update(0) // Ensure entity is in world
	_, err := tm.PrepareTransfer(player.ID, world, "server-2")
	if err != nil {
		t.Fatalf("First PrepareTransfer() error = %v", err)
	}

	// Second transfer should fail
	world.Update(0) // Ensure entity is in world
	_, err = tm.PrepareTransfer(player.ID, world, "server-3")
	if err == nil {
		t.Error("Expected error for duplicate transfer, got nil")
	}
}

func TestPrepareTransfer_InvalidPlayer(t *testing.T) {
	tm := NewTransferManager()
	world := engine.NewWorld()

	world.Update(0) // Ensure entity is in world
	_, err := tm.PrepareTransfer(99999, world, "server-2")
	if err == nil {
		t.Error("Expected error for invalid player, got nil")
	}
}

func TestBeginTransfer(t *testing.T) {
	tm := NewTransferManager()
	world := engine.NewWorld()

	player := world.CreateEntity()
	player.AddComponent(&engine.PositionComponent{X: 100, Y: 200})
	player.AddComponent(&engine.HealthComponent{Current: 80, Max: 100})
	player.AddComponent(engine.NewInventoryComponent(20, 100.0))
	world.Update(0) // Process entity addition

	world.Update(0) // Ensure entity is in world
	_, err := tm.PrepareTransfer(player.ID, world, "server-2")
	if err != nil {
		t.Fatalf("PrepareTransfer() error = %v", err)
	}

	err = tm.BeginTransfer(player.ID, "server-1")
	if err != nil {
		t.Errorf("BeginTransfer() error = %v", err)
	}

	transfer, _ := tm.GetTransfer(player.ID)
	if transfer.Phase != TransferPhaseTransfer {
		t.Errorf("Phase = %v, want %v", transfer.Phase, TransferPhaseTransfer)
	}

	if transfer.OriginID != "server-1" {
		t.Errorf("OriginID = %v, want server-1", transfer.OriginID)
	}
}

func TestBeginTransfer_NotPrepared(t *testing.T) {
	tm := NewTransferManager()

	err := tm.BeginTransfer(99999, "server-1")
	if err == nil {
		t.Error("Expected error for unprepared transfer, got nil")
	}
}

func TestConfirmTransfer(t *testing.T) {
	tm := NewTransferManager()
	world := engine.NewWorld()

	commitCalled := false
	tm.SetCallbacks(nil, func(uint64) { commitCalled = true }, nil)

	player := world.CreateEntity()

	player.AddComponent(&engine.PositionComponent{X: 100, Y: 200})
	player.AddComponent(&engine.HealthComponent{Current: 80, Max: 100})
	player.AddComponent(engine.NewInventoryComponent(20, 100.0))
	world.Update(0) // Process entity addition

	world.Update(0) // Ensure entity is in world
	tm.PrepareTransfer(player.ID, world, "server-2")
	tm.BeginTransfer(player.ID, "server-1")

	err := tm.ConfirmTransfer(player.ID)
	if err != nil {
		t.Errorf("ConfirmTransfer() error = %v", err)
	}

	if !commitCalled {
		t.Error("Commit callback not called")
	}

	// Verify cleanup
	if _, exists := tm.activeTransfers[player.ID]; exists {
		t.Error("Active transfer not cleaned up")
	}

	if _, exists := tm.stateBackups[player.ID]; exists {
		t.Error("State backup not cleaned up")
	}
}

func TestRollbackTransfer(t *testing.T) {
	tm := NewTransferManager()
	world := engine.NewWorld()

	failCalled := false
	var failReason string
	tm.SetCallbacks(nil, nil, func(_ uint64, reason string) {
		failCalled = true
		failReason = reason
	})

	player := world.CreateEntity()

	player.AddComponent(&engine.PositionComponent{X: 100, Y: 200})
	player.AddComponent(&engine.HealthComponent{Current: 80, Max: 100})
	player.AddComponent(engine.NewInventoryComponent(20, 100.0))
	world.Update(0) // Process entity addition

	world.Update(0) // Ensure entity is in world
	tm.PrepareTransfer(player.ID, world, "server-2")
	tm.BeginTransfer(player.ID, "server-1")

	err := tm.RollbackTransfer(player.ID, "timeout")
	if err != nil {
		t.Errorf("RollbackTransfer() error = %v", err)
	}

	if !failCalled {
		t.Error("Fail callback not called")
	}

	if failReason != "timeout" {
		t.Errorf("failReason = %v, want timeout", failReason)
	}

	// Verify active transfer cleaned up but backup kept
	if _, exists := tm.activeTransfers[player.ID]; exists {
		t.Error("Active transfer not cleaned up")
	}

	if _, exists := tm.stateBackups[player.ID]; !exists {
		t.Error("State backup incorrectly removed")
	}
}

func TestRestorePlayerState(t *testing.T) {
	tm := NewTransferManager()
	world := engine.NewWorld()

	player := world.CreateEntity()

	player.AddComponent(&engine.PositionComponent{X: 100, Y: 200})
	player.AddComponent(&engine.HealthComponent{Current: 80, Max: 100})
	player.AddComponent(engine.NewInventoryComponent(20, 100.0))
	world.Update(0) // Process entity addition

	world.Update(0) // Ensure entity is in world
	tm.PrepareTransfer(player.ID, world, "server-2")

	// Modify player state
	pos, _ := player.GetComponent("position")
	pos.(*engine.PositionComponent).X = 500
	pos.(*engine.PositionComponent).Y = 600

	// Restore from backup
	err := tm.RestorePlayerState(player.ID, world)
	if err != nil {
		t.Errorf("RestorePlayerState() error = %v", err)
	}

	// Verify restoration
	pos, _ = player.GetComponent("position")
	posComp := pos.(*engine.PositionComponent)
	if posComp.X != 100 || posComp.Y != 200 {
		t.Errorf("Position = (%v, %v), want (100, 200)", posComp.X, posComp.Y)
	}

	// Verify backup cleaned up
	if _, exists := tm.stateBackups[player.ID]; exists {
		t.Error("Backup not cleaned up after restore")
	}
}

func TestRestorePlayerState_NoBackup(t *testing.T) {
	tm := NewTransferManager()
	world := engine.NewWorld()

	err := tm.RestorePlayerState(99999, world)
	if err == nil {
		t.Error("Expected error for missing backup, got nil")
	}
}

func TestGetTransfer(t *testing.T) {
	tm := NewTransferManager()
	world := engine.NewWorld()

	player := world.CreateEntity()

	player.AddComponent(&engine.PositionComponent{X: 100, Y: 200})
	player.AddComponent(&engine.HealthComponent{Current: 80, Max: 100})
	player.AddComponent(engine.NewInventoryComponent(20, 100.0))
	world.Update(0) // Process entity addition

	world.Update(0) // Ensure entity is in world
	tm.PrepareTransfer(player.ID, world, "server-2")

	transfer, exists := tm.GetTransfer(player.ID)
	if !exists {
		t.Error("Transfer not found")
	}

	if transfer == nil {
		t.Error("Transfer is nil")
	}
}

func TestGetTransfer_NotFound(t *testing.T) {
	tm := NewTransferManager()

	_, exists := tm.GetTransfer(99999)
	if exists {
		t.Error("Transfer should not exist")
	}
}

func TestCheckTimeouts(t *testing.T) {
	tm := NewTransferManager()
	world := engine.NewWorld()

	// Set short timeout
	tm.SetTransferTimeout(1 * time.Second)

	player := world.CreateEntity()

	player.AddComponent(&engine.PositionComponent{X: 100, Y: 200})
	player.AddComponent(&engine.HealthComponent{Current: 80, Max: 100})
	player.AddComponent(engine.NewInventoryComponent(20, 100.0))
	world.Update(0) // Process entity addition

	world.Update(0) // Ensure entity is in world
	tm.PrepareTransfer(player.ID, world, "server-2")

	// Wait for timeout
	time.Sleep(2 * time.Second)

	expired := tm.CheckTimeouts()
	if len(expired) != 1 {
		t.Errorf("len(expired) = %d, want 1", len(expired))
	}

	if len(expired) > 0 && expired[0] != player.ID {
		t.Errorf("expired[0] = %v, want %v", expired[0], player.ID)
	}
}

func TestValidatePlayerState_Valid(t *testing.T) {
	tm := NewTransferManager()

	state := &PlayerState{
		PlayerID: 123,
		Position: &PositionData{X: 100, Y: 200},
		Health:   &HealthData{Current: 80, Max: 100},
		Inventory: []ItemData{
			{ItemID: "1", TypeName: "Sword", Quantity: 1},
		},
	}

	err := tm.ValidatePlayerState(state)
	if err != nil {
		t.Errorf("ValidatePlayerState() error = %v", err)
	}
}

func TestValidatePlayerState_Nil(t *testing.T) {
	tm := NewTransferManager()

	err := tm.ValidatePlayerState(nil)
	if err == nil {
		t.Error("Expected error for nil state, got nil")
	}
}

func TestValidatePlayerState_InvalidPlayerID(t *testing.T) {
	tm := NewTransferManager()

	state := &PlayerState{
		PlayerID: 0,
	}

	err := tm.ValidatePlayerState(state)
	if err == nil {
		t.Error("Expected error for invalid player ID, got nil")
	}
}

func TestValidatePlayerState_NegativeHealth(t *testing.T) {
	tm := NewTransferManager()

	state := &PlayerState{
		PlayerID: 123,
		Health:   &HealthData{Current: -10, Max: 100},
	}

	err := tm.ValidatePlayerState(state)
	if err == nil {
		t.Error("Expected error for negative health, got nil")
	}
}

func TestValidatePlayerState_InvalidMaxHealth(t *testing.T) {
	tm := NewTransferManager()

	state := &PlayerState{
		PlayerID: 123,
		Health:   &HealthData{Current: 50, Max: 0},
	}

	err := tm.ValidatePlayerState(state)
	if err == nil {
		t.Error("Expected error for invalid max health, got nil")
	}
}

func TestValidatePlayerState_ExcessiveHealth(t *testing.T) {
	tm := NewTransferManager()

	state := &PlayerState{
		PlayerID: 123,
		Health:   &HealthData{Current: 120, Max: 100},
	}

	err := tm.ValidatePlayerState(state)
	if err == nil {
		t.Error("Expected error for excessive health, got nil")
	}
}

func TestValidatePlayerState_LargeInventory(t *testing.T) {
	tm := NewTransferManager()

	items := make([]ItemData, 101)
	for i := range items {
		items[i] = ItemData{ItemID: "item", TypeName: "Item", Quantity: 1}
	}

	state := &PlayerState{
		PlayerID:  123,
		Inventory: items,
	}

	err := tm.ValidatePlayerState(state)
	if err == nil {
		t.Error("Expected error for large inventory, got nil")
	}
}

func TestComputeStateHash(t *testing.T) {
	tm := NewTransferManager()

	state := &PlayerState{
		PlayerID: 123,
		Position: &PositionData{X: 100, Y: 200},
		Health:   &HealthData{Current: 80, Max: 100},
	}

	hash1, err := tm.computeStateHash(state)
	if err != nil {
		t.Fatalf("computeStateHash() error = %v", err)
	}

	if hash1 == "" {
		t.Error("Hash is empty")
	}

	// Same state should produce same hash
	hash2, err := tm.computeStateHash(state)
	if err != nil {
		t.Fatalf("computeStateHash() second call error = %v", err)
	}

	if hash1 != hash2 {
		t.Errorf("Hash mismatch: %v != %v", hash1, hash2)
	}

	// Different state should produce different hash
	state.Position.X = 300
	hash3, err := tm.computeStateHash(state)
	if err != nil {
		t.Fatalf("computeStateHash() third call error = %v", err)
	}

	if hash1 == hash3 {
		t.Error("Hash should differ for different state")
	}
}

func TestVerifyStateHash(t *testing.T) {
	tm := NewTransferManager()

	state := &PlayerState{
		PlayerID: 123,
		Position: &PositionData{X: 100, Y: 200},
		Health:   &HealthData{Current: 80, Max: 100},
	}

	hash, _ := tm.computeStateHash(state)

	err := tm.VerifyStateHash(state, hash)
	if err != nil {
		t.Errorf("VerifyStateHash() error = %v", err)
	}
}

func TestVerifyStateHash_Mismatch(t *testing.T) {
	tm := NewTransferManager()

	state := &PlayerState{
		PlayerID: 123,
		Position: &PositionData{X: 100, Y: 200},
	}

	err := tm.VerifyStateHash(state, "invalid-hash")
	if err == nil {
		t.Error("Expected error for hash mismatch, got nil")
	}
}

// Benchmarks

func BenchmarkPrepareTransfer(b *testing.B) {
	tm := NewTransferManager()
	world := engine.NewWorld()

	player := world.CreateEntity()

	player.AddComponent(&engine.PositionComponent{X: 100, Y: 200})
	player.AddComponent(&engine.HealthComponent{Current: 80, Max: 100})
	player.AddComponent(engine.NewInventoryComponent(20, 100.0))
	world.Update(0) // Process entity addition

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Clean up from previous iteration
		tm.mu.Lock()
		delete(tm.activeTransfers, player.ID)
		delete(tm.stateBackups, player.ID)
		tm.mu.Unlock()

		world.Update(0) // Ensure entity is in world
		tm.PrepareTransfer(player.ID, world, "server-2")
	}
}

func BenchmarkBeginTransfer(b *testing.B) {
	tm := NewTransferManager()
	world := engine.NewWorld()

	player := world.CreateEntity()

	player.AddComponent(&engine.PositionComponent{X: 100, Y: 200})
	player.AddComponent(&engine.HealthComponent{Current: 80, Max: 100})
	player.AddComponent(engine.NewInventoryComponent(20, 100.0))
	world.Update(0) // Process entity addition

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		tm.mu.Lock()
		delete(tm.activeTransfers, player.ID)
		delete(tm.stateBackups, player.ID)
		tm.mu.Unlock()
		world.Update(0) // Ensure entity is in world
		tm.PrepareTransfer(player.ID, world, "server-2")
		b.StartTimer()

		tm.BeginTransfer(player.ID, "server-1")
	}
}

func BenchmarkValidatePlayerState(b *testing.B) {
	tm := NewTransferManager()

	state := &PlayerState{
		PlayerID: 123,
		Position: &PositionData{X: 100, Y: 200},
		Health:   &HealthData{Current: 80, Max: 100},
		Inventory: []ItemData{
			{ItemID: "1", TypeName: "Sword", Quantity: 1},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tm.ValidatePlayerState(state)
	}
}

func BenchmarkComputeStateHash(b *testing.B) {
	tm := NewTransferManager()

	state := &PlayerState{
		PlayerID: 123,
		Position: &PositionData{X: 100, Y: 200},
		Health:   &HealthData{Current: 80, Max: 100},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tm.computeStateHash(state)
	}
}
