package network

import (
	"testing"
	"time"
)

// TestNewProjectileNetworkSync verifies initialization.
func TestNewProjectileNetworkSync(t *testing.T) {
	sync := NewProjectileNetworkSync()

	if sync == nil {
		t.Fatal("NewProjectileNetworkSync returned nil")
	}

	if sync.sequenceNumber != 0 {
		t.Errorf("Expected initial sequence number 0, got %d", sync.sequenceNumber)
	}

	if sync.serverTime != 0.0 {
		t.Errorf("Expected initial server time 0.0, got %.2f", sync.serverTime)
	}

	if sync.predictedProjectiles == nil {
		t.Error("predictedProjectiles map not initialized")
	}

	if sync.confirmedProjectiles == nil {
		t.Error("confirmedProjectiles map not initialized")
	}

	if sync.projectileHistory == nil {
		t.Error("projectileHistory map not initialized")
	}
}

// TestUpdateServerTime verifies server time tracking.
func TestUpdateServerTime(t *testing.T) {
	sync := NewProjectileNetworkSync()

	// Initial time
	if sync.GetServerTime() != 0.0 {
		t.Errorf("Expected initial server time 0.0, got %.2f", sync.GetServerTime())
	}

	// Update time
	sync.UpdateServerTime(0.016) // 16ms frame
	if sync.GetServerTime() != 0.016 {
		t.Errorf("Expected server time 0.016, got %.2f", sync.GetServerTime())
	}

	// Multiple updates
	sync.UpdateServerTime(0.016)
	sync.UpdateServerTime(0.016)
	expected := 0.048
	if abs(sync.GetServerTime()-expected) > 0.001 {
		t.Errorf("Expected server time %.3f, got %.3f", expected, sync.GetServerTime())
	}
}

// TestCreateSpawnMessage verifies spawn message creation.
func TestCreateSpawnMessage(t *testing.T) {
	sync := NewProjectileNetworkSync()
	sync.UpdateServerTime(10.0)

	msg := sync.CreateSpawnMessage(
		1000,       // projectileID
		100,        // ownerID
		50.0, 50.0, // position
		300.0, 0.0, // velocity
		25.0,  // damage
		300.0, // speed
		2.0,   // lifetime
		0,     // pierce
		0,     // bounce
		false, // explosive
		0.0,   // explosionRadius
		"arrow",
	)

	if msg.ProjectileID != 1000 {
		t.Errorf("Expected ProjectileID 1000, got %d", msg.ProjectileID)
	}

	if msg.OwnerID != 100 {
		t.Errorf("Expected OwnerID 100, got %d", msg.OwnerID)
	}

	if msg.PositionX != 50.0 || msg.PositionY != 50.0 {
		t.Errorf("Expected position (50, 50), got (%.2f, %.2f)", msg.PositionX, msg.PositionY)
	}

	if msg.VelocityX != 300.0 || msg.VelocityY != 0.0 {
		t.Errorf("Expected velocity (300, 0), got (%.2f, %.2f)", msg.VelocityX, msg.VelocityY)
	}

	if msg.Damage != 25.0 {
		t.Errorf("Expected damage 25.0, got %.2f", msg.Damage)
	}

	if msg.SpawnTime != 10.0 {
		t.Errorf("Expected spawn time 10.0, got %.2f", msg.SpawnTime)
	}

	if msg.SequenceNumber != 1 {
		t.Errorf("Expected sequence number 1, got %d", msg.SequenceNumber)
	}

	if msg.ProjectileType != "arrow" {
		t.Errorf("Expected projectile type 'arrow', got '%s'", msg.ProjectileType)
	}

	// Verify projectile is tracked
	_, exists := sync.GetConfirmedProjectile(1000)
	if !exists {
		t.Error("Projectile not added to confirmed projectiles")
	}

	// Verify history initialized
	snapshot := sync.GetHistoricalState(1000, 10.0)
	if snapshot == nil {
		t.Error("Projectile history not initialized")
	} else {
		if snapshot.X != 50.0 || snapshot.Y != 50.0 {
			t.Errorf("Initial snapshot position incorrect: (%.2f, %.2f)", snapshot.X, snapshot.Y)
		}
	}
}

// TestCreateHitMessage verifies hit message creation.
func TestCreateHitMessage(t *testing.T) {
	sync := NewProjectileNetworkSync()
	sync.UpdateServerTime(10.0)

	// Create projectile first
	sync.CreateSpawnMessage(1000, 100, 50.0, 50.0, 300.0, 0.0, 25.0, 300.0, 2.0, 0, 0, false, 0.0, "arrow")

	sync.UpdateServerTime(0.5) // Advance time

	msg := sync.CreateHitMessage(
		1000, // projectileID
		500,  // hitEntityID
		"entity",
		25.0,  // damageDealt
		200.0, // x
		50.0,  // y
		true,  // projectileDestroyed
		false, // explosionTriggered
		nil,   // explosionEntities
		nil,   // explosionDamages
	)

	if msg.ProjectileID != 1000 {
		t.Errorf("Expected ProjectileID 1000, got %d", msg.ProjectileID)
	}

	if msg.HitEntityID != 500 {
		t.Errorf("Expected HitEntityID 500, got %d", msg.HitEntityID)
	}

	if msg.HitType != "entity" {
		t.Errorf("Expected HitType 'entity', got '%s'", msg.HitType)
	}

	if msg.DamageDealt != 25.0 {
		t.Errorf("Expected DamageDealt 25.0, got %.2f", msg.DamageDealt)
	}

	if msg.HitTime != 10.5 {
		t.Errorf("Expected HitTime 10.5, got %.2f", msg.HitTime)
	}

	if msg.SequenceNumber != 2 {
		t.Errorf("Expected sequence number 2, got %d", msg.SequenceNumber)
	}

	if !msg.ProjectileDestroyed {
		t.Error("Expected ProjectileDestroyed to be true")
	}

	// Verify projectile removed from confirmed
	_, exists := sync.GetConfirmedProjectile(1000)
	if exists {
		t.Error("Destroyed projectile should be removed from confirmed projectiles")
	}
}

// TestCreateHitMessage_Explosive verifies explosive hit message creation.
func TestCreateHitMessage_Explosive(t *testing.T) {
	sync := NewProjectileNetworkSync()
	sync.UpdateServerTime(10.0)

	// Create explosive projectile
	sync.CreateSpawnMessage(1003, 100, 50.0, 50.0, 250.0, 0.0, 50.0, 250.0, 2.0, 0, 0, true, 80.0, "fireball")

	sync.UpdateServerTime(0.8)

	explosionEntities := []uint64{500, 501, 502}
	explosionDamages := []float64{50.0, 35.0, 20.0}

	msg := sync.CreateHitMessage(
		1003,
		500,
		"entity",
		50.0,
		250.0,
		50.0,
		true,
		true,
		explosionEntities,
		explosionDamages,
	)

	if !msg.ExplosionTriggered {
		t.Error("Expected ExplosionTriggered to be true")
	}

	if len(msg.ExplosionEntities) != 3 {
		t.Errorf("Expected 3 explosion entities, got %d", len(msg.ExplosionEntities))
	}

	if len(msg.ExplosionDamages) != 3 {
		t.Errorf("Expected 3 explosion damages, got %d", len(msg.ExplosionDamages))
	}

	if len(msg.ExplosionEntities) != len(msg.ExplosionDamages) {
		t.Error("ExplosionEntities and ExplosionDamages length mismatch")
	}
}

// TestCreateDespawnMessage verifies despawn message creation.
func TestCreateDespawnMessage(t *testing.T) {
	sync := NewProjectileNetworkSync()
	sync.UpdateServerTime(12.0)

	// Create projectile first
	sync.CreateSpawnMessage(1004, 100, 50.0, 50.0, 300.0, 0.0, 25.0, 300.0, 2.0, 0, 0, false, 0.0, "arrow")

	msg := sync.CreateDespawnMessage(1004, "expired")

	if msg.ProjectileID != 1004 {
		t.Errorf("Expected ProjectileID 1004, got %d", msg.ProjectileID)
	}

	if msg.Reason != "expired" {
		t.Errorf("Expected Reason 'expired', got '%s'", msg.Reason)
	}

	if msg.DespawnTime != 12.0 {
		t.Errorf("Expected DespawnTime 12.0, got %.2f", msg.DespawnTime)
	}

	if msg.SequenceNumber != 2 {
		t.Errorf("Expected sequence number 2, got %d", msg.SequenceNumber)
	}

	// Verify projectile removed
	_, exists := sync.GetConfirmedProjectile(1004)
	if exists {
		t.Error("Despawned projectile should be removed from confirmed projectiles")
	}
}

// TestProjectileRecordSnapshot verifies snapshot recording.
func TestProjectileRecordSnapshot(t *testing.T) {
	sync := NewProjectileNetworkSync()
	sync.UpdateServerTime(10.0)

	// Create projectile
	sync.CreateSpawnMessage(1000, 100, 50.0, 50.0, 300.0, 0.0, 25.0, 300.0, 2.0, 0, 0, false, 0.0, "arrow")

	// Record snapshots at different times
	sync.UpdateServerTime(0.1)
	sync.RecordSnapshot(1000, 80.0, 50.0, 300.0, 0.0, 0.1)

	sync.UpdateServerTime(0.1)
	sync.RecordSnapshot(1000, 110.0, 50.0, 300.0, 0.0, 0.2)

	// Retrieve snapshot at t=10.1
	snapshot := sync.GetHistoricalState(1000, 10.1)
	if snapshot == nil {
		t.Fatal("Failed to retrieve snapshot at t=10.1")
	}

	if snapshot.X != 80.0 || snapshot.Y != 50.0 {
		t.Errorf("Expected position (80, 50), got (%.2f, %.2f)", snapshot.X, snapshot.Y)
	}

	// Retrieve snapshot at t=10.2
	snapshot = sync.GetHistoricalState(1000, 10.2)
	if snapshot == nil {
		t.Fatal("Failed to retrieve snapshot at t=10.2")
	}

	if snapshot.X != 110.0 || snapshot.Y != 50.0 {
		t.Errorf("Expected position (110, 50), got (%.2f, %.2f)", snapshot.X, snapshot.Y)
	}
}

// TestGetHistoricalState_Interpolation verifies snapshot interpolation.
func TestGetHistoricalState_Interpolation(t *testing.T) {
	sync := NewProjectileNetworkSync()
	sync.UpdateServerTime(10.0)

	// Create projectile
	sync.CreateSpawnMessage(1000, 100, 0.0, 0.0, 100.0, 0.0, 25.0, 100.0, 2.0, 0, 0, false, 0.0, "arrow")

	// Record snapshots at t=10.0 (x=0) and t=11.0 (x=100)
	sync.UpdateServerTime(1.0)
	sync.RecordSnapshot(1000, 100.0, 0.0, 100.0, 0.0, 1.0)

	// Retrieve snapshot at t=10.5 (should interpolate to x=50)
	snapshot := sync.GetHistoricalState(1000, 10.5)
	if snapshot == nil {
		t.Fatal("Failed to retrieve interpolated snapshot")
	}

	expected := 50.0
	tolerance := 1.0
	if abs(snapshot.X-expected) > tolerance {
		t.Errorf("Expected interpolated X ~%.2f, got %.2f", expected, snapshot.X)
	}
}

// TestCleanupOldHistory verifies history cleanup.
func TestCleanupOldHistory(t *testing.T) {
	sync := NewProjectileNetworkSync()
	sync.UpdateServerTime(0.0)

	// Create projectile and record snapshots
	sync.CreateSpawnMessage(1000, 100, 0.0, 0.0, 100.0, 0.0, 25.0, 100.0, 2.0, 0, 0, false, 0.0, "arrow")

	for i := 0; i < 10; i++ {
		sync.UpdateServerTime(0.2)
		sync.RecordSnapshot(1000, float64(i)*20.0, 0.0, 100.0, 0.0, float64(i)*0.2)
	}

	// Advance time beyond cleanup threshold
	sync.UpdateServerTime(2.0)
	
	// Record a snapshot after the time advance so we have recent data
	sync.RecordSnapshot(1000, 400.0, 0.0, 100.0, 0.0, 4.0)
	
	sync.CleanupOldHistory()

	// Old snapshots should be removed
	snapshot := sync.GetHistoricalState(1000, 0.0)
	if snapshot != nil {
		t.Error("Old snapshot at t=0.0 should be cleaned up")
	}

	// Recent snapshots should remain
	// Check for the snapshot we just recorded at serverTime=4.0
	snapshot = sync.GetHistoricalState(1000, sync.GetServerTime())
	if snapshot == nil {
		t.Error("Recent snapshot should not be cleaned up")
	}
}

// TestPredictProjectile verifies client-side prediction.
func TestPredictProjectile(t *testing.T) {
	sync := NewProjectileNetworkSync()

	predictionID := uint64(9999)
	msg := ProjectileSpawnMessage{
		ProjectileID:   predictionID,
		OwnerID:        100,
		PositionX:      50.0,
		PositionY:      50.0,
		ProjectileType: "arrow",
	}

	sync.PredictProjectile(predictionID, msg)

	// Verify prediction stored
	stats := sync.GetStats()
	if stats.PredictedProjectiles != 1 {
		t.Errorf("Expected 1 predicted projectile, got %d", stats.PredictedProjectiles)
	}
}

// TestConfirmPrediction verifies prediction confirmation.
func TestConfirmPrediction(t *testing.T) {
	sync := NewProjectileNetworkSync()

	// Client predicts projectile
	predictionID := uint64(9999)
	predictedMsg := ProjectileSpawnMessage{
		ProjectileID:   predictionID,
		OwnerID:        100,
		PositionX:      50.0,
		PositionY:      50.0,
		VelocityX:      300.0,
		VelocityY:      0.0,
		ProjectileType: "arrow",
	}
	sync.PredictProjectile(predictionID, predictedMsg)

	// Server confirms with authoritative ID
	serverProjectileID := uint64(1000)
	serverMsg := ProjectileSpawnMessage{
		ProjectileID:   serverProjectileID,
		OwnerID:        100,
		PositionX:      51.0, // Slight difference
		PositionY:      50.0,
		VelocityX:      300.0,
		VelocityY:      0.0,
		ProjectileType: "arrow",
	}

	confirmed := sync.ConfirmPrediction(predictionID, serverProjectileID, serverMsg)
	if !confirmed {
		t.Error("Prediction should be confirmed")
	}

	// Verify prediction removed and server projectile added
	stats := sync.GetStats()
	if stats.PredictedProjectiles != 0 {
		t.Errorf("Expected 0 predicted projectiles after confirmation, got %d", stats.PredictedProjectiles)
	}

	if stats.ConfirmedProjectiles != 1 {
		t.Errorf("Expected 1 confirmed projectile, got %d", stats.ConfirmedProjectiles)
	}

	// Verify server projectile retrievable
	msg, exists := sync.GetConfirmedProjectile(serverProjectileID)
	if !exists {
		t.Error("Server projectile should exist after confirmation")
	}

	if msg.ProjectileID != serverProjectileID {
		t.Errorf("Expected ProjectileID %d, got %d", serverProjectileID, msg.ProjectileID)
	}
}

// TestConfirmPrediction_NotFound verifies behavior when prediction not found.
func TestConfirmPrediction_NotFound(t *testing.T) {
	sync := NewProjectileNetworkSync()

	// Try to confirm non-existent prediction
	confirmed := sync.ConfirmPrediction(9999, 1000, ProjectileSpawnMessage{})
	if confirmed {
		t.Error("Confirmation should fail for non-existent prediction")
	}
}

// TestRemoveProjectile verifies projectile removal.
func TestRemoveProjectile(t *testing.T) {
	sync := NewProjectileNetworkSync()

	// Create projectile
	sync.CreateSpawnMessage(1000, 100, 50.0, 50.0, 300.0, 0.0, 25.0, 300.0, 2.0, 0, 0, false, 0.0, "arrow")

	// Verify exists
	_, exists := sync.GetConfirmedProjectile(1000)
	if !exists {
		t.Error("Projectile should exist after creation")
	}

	// Remove projectile
	sync.RemoveProjectile(1000)

	// Verify removed
	_, exists = sync.GetConfirmedProjectile(1000)
	if exists {
		t.Error("Projectile should not exist after removal")
	}
}

// TestProjectileGetStats verifies statistics retrieval.
func TestProjectileGetStats(t *testing.T) {
	sync := NewProjectileNetworkSync()
	sync.UpdateServerTime(10.0)

	// Initial stats
	stats := sync.GetStats()
	if stats.ConfirmedProjectiles != 0 {
		t.Errorf("Expected 0 confirmed projectiles, got %d", stats.ConfirmedProjectiles)
	}

	// Create projectiles
	sync.CreateSpawnMessage(1000, 100, 50.0, 50.0, 300.0, 0.0, 25.0, 300.0, 2.0, 0, 0, false, 0.0, "arrow")
	sync.CreateSpawnMessage(1001, 100, 60.0, 60.0, 400.0, 0.0, 30.0, 400.0, 3.0, 2, 0, false, 0.0, "bolt")

	// Predict one
	sync.PredictProjectile(9999, ProjectileSpawnMessage{})

	// Check stats
	stats = sync.GetStats()
	if stats.ConfirmedProjectiles != 2 {
		t.Errorf("Expected 2 confirmed projectiles, got %d", stats.ConfirmedProjectiles)
	}

	if stats.PredictedProjectiles != 1 {
		t.Errorf("Expected 1 predicted projectile, got %d", stats.PredictedProjectiles)
	}

	if stats.HistoryEntries != 2 {
		t.Errorf("Expected 2 history entries, got %d", stats.HistoryEntries)
	}

	if stats.ServerTime != 10.0 {
		t.Errorf("Expected server time 10.0, got %.2f", stats.ServerTime)
	}
}

// TestCleanupTask verifies background cleanup task.
func TestCleanupTask(t *testing.T) {
	sync := NewProjectileNetworkSync()

	// Start cleanup task
	stopChan := sync.CleanupTask()
	defer func() { stopChan <- struct{}{} }()

	// Create old projectile and record initial snapshot
	sync.CreateSpawnMessage(1000, 100, 0.0, 0.0, 100.0, 0.0, 25.0, 100.0, 2.0, 0, 0, false, 0.0, "arrow")
	sync.RecordSnapshot(1000, 0.0, 0.0, 100.0, 0.0, 0.0)

	// Advance time significantly and record snapshots
	for i := 0; i < 30; i++ {
		sync.UpdateServerTime(0.1)
		sync.RecordSnapshot(1000, float64(i+1)*10.0, 0.0, 100.0, 0.0, float64(i+1)*0.1)
	}

	// Wait for cleanup task to run (runs every 1 second)
	time.Sleep(1100 * time.Millisecond)

	// Old snapshots should be cleaned
	snapshot := sync.GetHistoricalState(1000, 0.0)
	if snapshot != nil {
		t.Error("Old snapshot should be cleaned up by background task")
	}
	
	// Recent snapshots should still exist
	snapshot = sync.GetHistoricalState(1000, sync.GetServerTime())
	if snapshot == nil {
		t.Error("Recent snapshot should not be cleaned up")
	}
}

// TestSequenceNumberIncrement verifies sequence numbers increment correctly.
func TestSequenceNumberIncrement(t *testing.T) {
	sync := NewProjectileNetworkSync()

	msg1 := sync.CreateSpawnMessage(1000, 100, 50.0, 50.0, 300.0, 0.0, 25.0, 300.0, 2.0, 0, 0, false, 0.0, "arrow")
	msg2 := sync.CreateSpawnMessage(1001, 100, 60.0, 60.0, 400.0, 0.0, 30.0, 400.0, 3.0, 0, 0, false, 0.0, "bolt")
	msg3 := sync.CreateHitMessage(1000, 500, "entity", 25.0, 200.0, 50.0, true, false, nil, nil)
	msg4 := sync.CreateDespawnMessage(1001, "expired")

	if msg1.SequenceNumber != 1 {
		t.Errorf("Expected sequence 1, got %d", msg1.SequenceNumber)
	}

	if msg2.SequenceNumber != 2 {
		t.Errorf("Expected sequence 2, got %d", msg2.SequenceNumber)
	}

	if msg3.SequenceNumber != 3 {
		t.Errorf("Expected sequence 3, got %d", msg3.SequenceNumber)
	}

	if msg4.SequenceNumber != 4 {
		t.Errorf("Expected sequence 4, got %d", msg4.SequenceNumber)
	}
}
