package network

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestDesyncDetector_ComputeChecksum(t *testing.T) {
	detector := NewDesyncDetector()

	tests := []struct {
		name       string
		entityID   uint64
		timestamp  uint64
		components []ComponentData
		wantSame   bool
	}{
		{
			name:      "identical state produces same hash",
			entityID:  1,
			timestamp: 1000,
			components: []ComponentData{
				{Type: "position", Data: []byte{10, 20}},
				{Type: "health", Data: []byte{100}},
			},
			wantSame: true,
		},
		{
			name:      "different timestamp produces different hash",
			entityID:  1,
			timestamp: 2000,
			components: []ComponentData{
				{Type: "position", Data: []byte{10, 20}},
				{Type: "health", Data: []byte{100}},
			},
			wantSame: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checksum1 := detector.ComputeChecksum(tt.entityID, tt.timestamp, tt.components)
			checksum2 := detector.ComputeChecksum(tt.entityID, 1000, []ComponentData{
				{Type: "position", Data: []byte{10, 20}},
				{Type: "health", Data: []byte{100}},
			})

			if tt.wantSame {
				if checksum1.Hash != checksum2.Hash {
					t.Errorf("expected same hash, got different")
				}
			} else {
				if checksum1.Hash == checksum2.Hash {
					t.Errorf("expected different hash, got same")
				}
			}
		})
	}
}

func TestDesyncDetector_ValidateChecksum(t *testing.T) {
	detector := NewDesyncDetector()

	components := []ComponentData{
		{Type: "position", Data: []byte{10, 20}},
	}

	checksum1 := detector.ComputeChecksum(1, 1000, components)
	checksum2 := detector.ComputeChecksum(1, 1000, components)
	checksum3 := detector.ComputeChecksum(1, 2000, components)

	if !detector.ValidateChecksum(checksum1, checksum2) {
		t.Error("identical checksums should validate")
	}

	if detector.ValidateChecksum(checksum1, checksum3) {
		t.Error("different checksums should not validate")
	}
}

func TestDesyncDetector_DetectDesync(t *testing.T) {
	detector := NewDesyncDetector()

	serverHash := [32]byte{1, 2, 3}
	clientHash := [32]byte{4, 5, 6}

	detected := detector.DetectDesync(DesyncCombat, 123, clientHash, serverHash, "kill attribution mismatch")
	if !detected {
		t.Error("expected desync to be detected")
	}

	events := detector.GetEvents()
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}

	event := events[0]
	if event.Type != DesyncCombat {
		t.Errorf("expected type %v, got %v", DesyncCombat, event.Type)
	}
	if event.EntityID != 123 {
		t.Errorf("expected entity ID 123, got %d", event.EntityID)
	}
	if event.Recovered {
		t.Error("event should not be marked as recovered")
	}
	if event.Details != "kill attribution mismatch" {
		t.Errorf("expected details 'kill attribution mismatch', got %q", event.Details)
	}
}

func TestDesyncDetector_DetectDesync_NoDesyncWhenHashesMatch(t *testing.T) {
	detector := NewDesyncDetector()

	sameHash := [32]byte{1, 2, 3}

	detected := detector.DetectDesync(DesyncInventory, 456, sameHash, sameHash, "")
	if detected {
		t.Error("should not detect desync when hashes match")
	}

	events := detector.GetEvents()
	if len(events) != 0 {
		t.Errorf("expected 0 events, got %d", len(events))
	}
}

func TestDesyncDetector_RecordRecovery(t *testing.T) {
	detector := NewDesyncDetector()

	serverHash := [32]byte{1}
	clientHash := [32]byte{2}

	detector.DetectDesync(DesyncPosition, 789, clientHash, serverHash, "")

	recoveryTime := 5 * time.Second
	detector.RecordRecovery(DesyncPosition, 789, recoveryTime)

	events := detector.GetEvents()
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}

	if !events[0].Recovered {
		t.Error("event should be marked as recovered")
	}
	if events[0].RecoveryTime != recoveryTime {
		t.Errorf("expected recovery time %v, got %v", recoveryTime, events[0].RecoveryTime)
	}
}

func TestDesyncDetector_NeedsFullSync(t *testing.T) {
	detector := NewDesyncDetector()
	detector.SetFullSyncInterval(100 * time.Millisecond)

	if detector.NeedsFullSync() {
		t.Error("should not need full sync immediately after creation")
	}

	time.Sleep(150 * time.Millisecond)

	if !detector.NeedsFullSync() {
		t.Error("should need full sync after interval elapsed")
	}

	detector.MarkFullSync()

	if detector.NeedsFullSync() {
		t.Error("should not need full sync immediately after marking sync")
	}
}

func TestDesyncDetector_GetEventsByType(t *testing.T) {
	detector := NewDesyncDetector()

	hash1 := [32]byte{1}
	hash2 := [32]byte{2}

	detector.DetectDesync(DesyncCombat, 1, hash1, hash2, "combat1")
	detector.DetectDesync(DesyncInventory, 2, hash1, hash2, "inventory1")
	detector.DetectDesync(DesyncCombat, 3, hash1, hash2, "combat2")

	combatEvents := detector.GetEventsByType(DesyncCombat)
	if len(combatEvents) != 2 {
		t.Errorf("expected 2 combat events, got %d", len(combatEvents))
	}

	inventoryEvents := detector.GetEventsByType(DesyncInventory)
	if len(inventoryEvents) != 1 {
		t.Errorf("expected 1 inventory event, got %d", len(inventoryEvents))
	}

	questEvents := detector.GetEventsByType(DesyncQuest)
	if len(questEvents) != 0 {
		t.Errorf("expected 0 quest events, got %d", len(questEvents))
	}
}

func TestDesyncDetector_GetStats(t *testing.T) {
	detector := NewDesyncDetector()

	hash1 := [32]byte{1}
	hash2 := [32]byte{2}

	// Add 3 desyncs, recover 2
	detector.DetectDesync(DesyncCombat, 1, hash1, hash2, "")
	detector.DetectDesync(DesyncInventory, 2, hash1, hash2, "")
	detector.DetectDesync(DesyncPosition, 3, hash1, hash2, "")

	detector.RecordRecovery(DesyncCombat, 1, 3*time.Second)
	detector.RecordRecovery(DesyncInventory, 2, 7*time.Second)

	stats := detector.GetStats()

	if stats.TotalEvents != 3 {
		t.Errorf("expected 3 total events, got %d", stats.TotalEvents)
	}
	if stats.RecoveredEvents != 2 {
		t.Errorf("expected 2 recovered events, got %d", stats.RecoveredEvents)
	}

	expectedAvg := (3*time.Second + 7*time.Second) / 2
	if stats.AverageRecovery != expectedAvg {
		t.Errorf("expected average recovery %v, got %v", expectedAvg, stats.AverageRecovery)
	}
}

func TestDesyncDetector_ClearEvents(t *testing.T) {
	detector := NewDesyncDetector()

	hash1 := [32]byte{1}
	hash2 := [32]byte{2}

	detector.DetectDesync(DesyncCombat, 1, hash1, hash2, "")
	detector.DetectDesync(DesyncInventory, 2, hash1, hash2, "")

	if len(detector.GetEvents()) != 2 {
		t.Fatalf("expected 2 events before clear")
	}

	detector.ClearEvents()

	if len(detector.GetEvents()) != 0 {
		t.Errorf("expected 0 events after clear, got %d", len(detector.GetEvents()))
	}
}

func TestDesyncDetector_Configuration(t *testing.T) {
	detector := NewDesyncDetector()

	// Test default values
	if detector.fullSyncInterval != 5*time.Minute {
		t.Errorf("expected default full sync interval 5m, got %v", detector.fullSyncInterval)
	}
	if detector.detectionDeadline != 30*time.Second {
		t.Errorf("expected default detection deadline 30s, got %v", detector.detectionDeadline)
	}
	if detector.recoveryDeadline != 10*time.Second {
		t.Errorf("expected default recovery deadline 10s, got %v", detector.recoveryDeadline)
	}

	// Test setters
	detector.SetFullSyncInterval(1 * time.Minute)
	detector.SetDetectionDeadline(15 * time.Second)
	detector.SetRecoveryDeadline(5 * time.Second)

	if detector.fullSyncInterval != 1*time.Minute {
		t.Errorf("expected full sync interval 1m, got %v", detector.fullSyncInterval)
	}
	if detector.detectionDeadline != 15*time.Second {
		t.Errorf("expected detection deadline 15s, got %v", detector.detectionDeadline)
	}
	if detector.recoveryDeadline != 5*time.Second {
		t.Errorf("expected recovery deadline 5s, got %v", detector.recoveryDeadline)
	}
}

func TestRollbackRecovery_Recover(t *testing.T) {
	var appliedEntityID uint64
	var appliedComponents []ComponentData

	serverState := func(entityID uint64) ([]ComponentData, error) {
		return []ComponentData{
			{Type: "position", Data: []byte{10, 20}},
			{Type: "health", Data: []byte{100}},
		}, nil
	}

	applyState := func(entityID uint64, components []ComponentData) error {
		appliedEntityID = entityID
		appliedComponents = components
		return nil
	}

	recovery := NewRollbackRecovery(serverState, applyState)

	event := DesyncEvent{
		Type:     DesyncPosition,
		EntityID: 123,
	}

	err := recovery.Recover(event)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if appliedEntityID != 123 {
		t.Errorf("expected entity ID 123, got %d", appliedEntityID)
	}
	if len(appliedComponents) != 2 {
		t.Errorf("expected 2 components, got %d", len(appliedComponents))
	}
}

func TestPeriodicSyncManager_StartStop(t *testing.T) {
	detector := NewDesyncDetector()
	detector.SetFullSyncInterval(50 * time.Millisecond)

	var mu sync.Mutex
	syncCount := 0
	syncFunc := func() error {
		mu.Lock()
		syncCount++
		mu.Unlock()
		return nil
	}

	manager := NewPeriodicSyncManager(detector, syncFunc, nil)

	manager.Start()
	time.Sleep(150 * time.Millisecond)
	manager.Stop()

	mu.Lock()
	count := syncCount
	mu.Unlock()

	if count < 2 {
		t.Errorf("expected at least 2 syncs, got %d", count)
	}

	// Should not sync after stop - wait for any in-flight sync to complete
	time.Sleep(20 * time.Millisecond)
	mu.Lock()
	oldCount := syncCount
	mu.Unlock()

	time.Sleep(100 * time.Millisecond)
	mu.Lock()
	finalCount := syncCount
	mu.Unlock()

	if finalCount != oldCount {
		t.Errorf("sync occurred after stop: before=%d, after=%d", oldCount, finalCount)
	}
}

func TestPeriodicSyncManager_ContinuesOnError(t *testing.T) {
	detector := NewDesyncDetector()
	detector.SetFullSyncInterval(50 * time.Millisecond)

	var mu sync.Mutex
	syncCallCount := 0
	syncFunc := func() error {
		mu.Lock()
		syncCallCount++
		mu.Unlock()
		// Return an error to verify sync continues despite errors
		return fmt.Errorf("test sync error")
	}

	// Create manager with nil logger (uses default)
	manager := NewPeriodicSyncManager(detector, syncFunc, nil)

	manager.Start()
	time.Sleep(150 * time.Millisecond)
	manager.Stop()

	mu.Lock()
	count := syncCallCount
	mu.Unlock()

	// Verify that sync was called despite errors
	if count < 2 {
		t.Errorf("expected at least 2 sync attempts, got %d", count)
	}
}

func TestDesyncDetector_Concurrent(t *testing.T) {
	detector := NewDesyncDetector()

	// Concurrent checksum computation
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func(id int) {
			components := []ComponentData{
				{Type: "test", Data: []byte{byte(id)}},
			}
			_ = detector.ComputeChecksum(uint64(id), uint64(id*1000), components)
			done <- true
		}(i)
	}

	for i := 0; i < 10; i++ {
		<-done
	}

	// Concurrent event recording
	for i := 0; i < 10; i++ {
		go func(id int) {
			hash1 := [32]byte{byte(id)}
			hash2 := [32]byte{byte(id + 1)}
			detector.DetectDesync(DesyncCombat, uint64(id), hash1, hash2, "concurrent test")
			done <- true
		}(i)
	}

	for i := 0; i < 10; i++ {
		<-done
	}

	events := detector.GetEvents()
	if len(events) != 10 {
		t.Errorf("expected 10 events, got %d", len(events))
	}
}

// Benchmark checksum computation
func BenchmarkDesyncDetector_ComputeChecksum(b *testing.B) {
	detector := NewDesyncDetector()
	components := []ComponentData{
		{Type: "position", Data: []byte{10, 20, 30, 40}},
		{Type: "velocity", Data: []byte{1, 2, 3, 4}},
		{Type: "health", Data: []byte{100}},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = detector.ComputeChecksum(1, 1000, components)
	}
}

// Benchmark desync detection
func BenchmarkDesyncDetector_DetectDesync(b *testing.B) {
	detector := NewDesyncDetector()
	hash1 := [32]byte{1, 2, 3}
	hash2 := [32]byte{4, 5, 6}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		detector.DetectDesync(DesyncCombat, uint64(i), hash1, hash2, "benchmark")
	}
}
