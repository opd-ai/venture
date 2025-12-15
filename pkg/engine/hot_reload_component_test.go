package engine

import (
	"encoding/json"
	"testing"
	"time"
)

func TestHotReloadComponentType(t *testing.T) {
	comp := NewHotReloadComponent()
	if comp.Type() != "hot_reload" {
		t.Errorf("expected type 'hot_reload', got %s", comp.Type())
	}
}

func TestHotReloadComponentNew(t *testing.T) {
	comp := NewHotReloadComponent()

	if comp.WatchedMods == nil {
		t.Error("WatchedMods should not be nil")
	}
	if comp.PendingUpdates == nil {
		t.Error("PendingUpdates should not be nil")
	}
	if comp.ReloadHistory == nil {
		t.Error("ReloadHistory should not be nil")
	}
	if comp.RollbackData == nil {
		t.Error("RollbackData should not be nil")
	}
	if !comp.Enabled {
		t.Error("Enabled should be true by default")
	}
	if comp.WatchInterval != DefaultWatchInterval {
		t.Errorf("expected default watch interval %v, got %v", DefaultWatchInterval, comp.WatchInterval)
	}
}

func TestHotReloadComponentWatching(t *testing.T) {
	comp := NewHotReloadComponent()

	// Test start watching
	comp.StartWatching("mod1", "1.0.0", "hash123")

	if !comp.IsWatching("mod1") {
		t.Error("mod1 should be watched")
	}

	watched, exists := comp.GetWatchedMod("mod1")
	if !exists {
		t.Error("mod1 should exist in watched mods")
	}
	if watched.ModID != "mod1" {
		t.Errorf("expected mod1, got %s", watched.ModID)
	}
	if watched.Version != "1.0.0" {
		t.Errorf("expected version 1.0.0, got %s", watched.Version)
	}
	if watched.FileHash != "hash123" {
		t.Errorf("expected hash123, got %s", watched.FileHash)
	}

	// Test get watched mod IDs
	ids := comp.GetWatchedModIDs()
	if len(ids) != 1 || ids[0] != "mod1" {
		t.Errorf("expected [mod1], got %v", ids)
	}

	// Test stop watching
	comp.StopWatching("mod1")
	if comp.IsWatching("mod1") {
		t.Error("mod1 should not be watched after stop")
	}
}

func TestHotReloadComponentDetectChange(t *testing.T) {
	comp := NewHotReloadComponent()
	comp.StartWatching("mod1", "1.0.0", "hash123")

	// Same hash - no change
	changed := comp.DetectChange("mod1", "hash123")
	if changed {
		t.Error("should not detect change with same hash")
	}

	// Different hash - change detected
	changed = comp.DetectChange("mod1", "hash456")
	if !changed {
		t.Error("should detect change with different hash")
	}

	watched, _ := comp.GetWatchedMod("mod1")
	if watched.FileHash != "hash456" {
		t.Error("hash should be updated")
	}
	if watched.ChangeCount != 1 {
		t.Errorf("expected change count 1, got %d", watched.ChangeCount)
	}
	if !watched.PendingReload {
		t.Error("pending reload should be true")
	}

	// Should be in pending updates
	if !comp.HasPendingUpdates() {
		t.Error("should have pending updates")
	}

	pending := comp.GetPendingUpdates()
	if len(pending) != 1 || pending[0] != "mod1" {
		t.Errorf("expected [mod1] in pending, got %v", pending)
	}

	// Non-watched mod
	changed = comp.DetectChange("nonexistent", "hash")
	if changed {
		t.Error("should not detect change for non-watched mod")
	}
}

func TestHotReloadComponentPendingUpdates(t *testing.T) {
	comp := NewHotReloadComponent()
	comp.StartWatching("mod1", "1.0.0", "hash1")
	comp.StartWatching("mod2", "1.0.0", "hash2")

	comp.DetectChange("mod1", "newhash1")
	comp.DetectChange("mod2", "newhash2")

	if len(comp.GetPendingUpdates()) != 2 {
		t.Error("should have 2 pending updates")
	}

	// Clear single pending update
	comp.ClearPendingUpdate("mod1")
	pending := comp.GetPendingUpdates()
	if len(pending) != 1 || pending[0] != "mod2" {
		t.Errorf("expected [mod2], got %v", pending)
	}

	watched, _ := comp.GetWatchedMod("mod1")
	if watched.PendingReload {
		t.Error("mod1 should not have pending reload")
	}

	// Clear all pending updates
	comp.DetectChange("mod1", "anotherhash")
	comp.ClearAllPendingUpdates()
	if comp.HasPendingUpdates() {
		t.Error("should not have pending updates after clear all")
	}
}

func TestHotReloadComponentRollbackState(t *testing.T) {
	comp := NewHotReloadComponent()

	scripts := map[string]any{"script1": "code"}
	variables := map[string]any{"var1": 42}

	comp.SaveStateForRollback("mod1", "1.0.0", scripts, variables)

	if !comp.RollbackAvailable {
		t.Error("rollback should be available")
	}

	state, exists := comp.GetRollbackState("mod1")
	if !exists {
		t.Error("rollback state should exist")
	}
	if state.ModID != "mod1" {
		t.Errorf("expected mod1, got %s", state.ModID)
	}
	if state.Version != "1.0.0" {
		t.Errorf("expected version 1.0.0, got %s", state.Version)
	}
	if state.Scripts["script1"] != "code" {
		t.Error("scripts not saved correctly")
	}
	if state.Variables["var1"] != 42 {
		t.Error("variables not saved correctly")
	}

	// Clear single rollback state
	comp.ClearRollbackState("mod1")
	_, exists = comp.GetRollbackState("mod1")
	if exists {
		t.Error("rollback state should be cleared")
	}
	if comp.RollbackAvailable {
		t.Error("rollback should not be available after clearing all")
	}

	// Test clear all
	comp.SaveStateForRollback("mod1", "1.0.0", scripts, variables)
	comp.SaveStateForRollback("mod2", "1.0.0", scripts, variables)
	comp.ClearAllRollbackState()
	if comp.RollbackAvailable {
		t.Error("rollback should not be available")
	}
	if len(comp.RollbackData) != 0 {
		t.Error("rollback data should be empty")
	}
}

func TestHotReloadComponentReloadHistory(t *testing.T) {
	comp := NewHotReloadComponent()

	entry := ReloadEntry{
		ModID:       "mod1",
		Timestamp:   time.Now().Unix(),
		OldVersion:  "1.0.0",
		NewVersion:  "1.1.0",
		Success:     true,
		Duration:    100,
		StateChange: "reloaded",
	}

	comp.AddReloadEntry(entry)

	if comp.GetTotalReloads() != 1 {
		t.Errorf("expected 1 reload, got %d", comp.GetTotalReloads())
	}
	if comp.GetSuccessfulReloads() != 1 {
		t.Errorf("expected 1 successful reload, got %d", comp.GetSuccessfulReloads())
	}
	if comp.GetFailedReloads() != 0 {
		t.Errorf("expected 0 failed reloads, got %d", comp.GetFailedReloads())
	}

	history := comp.GetReloadHistory()
	if len(history) != 1 {
		t.Errorf("expected 1 history entry, got %d", len(history))
	}
	if history[0].ModID != "mod1" {
		t.Errorf("expected mod1, got %s", history[0].ModID)
	}

	// Test history for specific mod
	comp.AddReloadEntry(ReloadEntry{ModID: "mod2", Timestamp: time.Now().Unix(), Success: false})
	mod1History := comp.GetReloadHistoryForMod("mod1")
	if len(mod1History) != 1 {
		t.Errorf("expected 1 entry for mod1, got %d", len(mod1History))
	}

	// Test failed reload counting
	if comp.GetFailedReloads() != 1 {
		t.Errorf("expected 1 failed reload, got %d", comp.GetFailedReloads())
	}

	// Test max history limit
	for i := 0; i < MaxReloadHistory+10; i++ {
		comp.AddReloadEntry(ReloadEntry{ModID: "modX", Timestamp: int64(i)})
	}
	if len(comp.GetReloadHistory()) > MaxReloadHistory {
		t.Errorf("history should not exceed %d entries", MaxReloadHistory)
	}
}

func TestHotReloadComponentSettings(t *testing.T) {
	comp := NewHotReloadComponent()

	// Test auto reload
	comp.SetAutoReload(true)
	if !comp.IsAutoReloadEnabled() {
		t.Error("auto reload should be enabled")
	}
	comp.SetAutoReload(false)
	if comp.IsAutoReloadEnabled() {
		t.Error("auto reload should be disabled")
	}

	// Test enabled
	comp.SetEnabled(false)
	if comp.IsEnabled() {
		t.Error("hot reload should be disabled")
	}
	comp.SetEnabled(true)
	if !comp.IsEnabled() {
		t.Error("hot reload should be enabled")
	}

	// Test watch interval
	newInterval := 2 * time.Second
	comp.SetWatchInterval(newInterval)
	if comp.GetWatchInterval() != newInterval {
		t.Errorf("expected interval %v, got %v", newInterval, comp.GetWatchInterval())
	}
}

func TestHotReloadComponentUpdateModVersion(t *testing.T) {
	comp := NewHotReloadComponent()
	comp.StartWatching("mod1", "1.0.0", "hash1")

	comp.UpdateModVersion("mod1", "2.0.0", "newhash")

	watched, _ := comp.GetWatchedMod("mod1")
	if watched.Version != "2.0.0" {
		t.Errorf("expected version 2.0.0, got %s", watched.Version)
	}
	if watched.FileHash != "newhash" {
		t.Errorf("expected newhash, got %s", watched.FileHash)
	}
}

func TestHotReloadComponentGetWatchedModCount(t *testing.T) {
	comp := NewHotReloadComponent()

	if comp.GetWatchedModCount() != 0 {
		t.Error("should have 0 watched mods initially")
	}

	comp.StartWatching("mod1", "1.0.0", "hash1")
	comp.StartWatching("mod2", "1.0.0", "hash2")

	if comp.GetWatchedModCount() != 2 {
		t.Errorf("expected 2 watched mods, got %d", comp.GetWatchedModCount())
	}
}

func TestHotReloadComponentSerialize(t *testing.T) {
	comp := NewHotReloadComponent()
	comp.StartWatching("mod1", "1.0.0", "hash1")
	comp.SetAutoReload(true)
	comp.SetWatchInterval(1 * time.Second)
	comp.AddReloadEntry(ReloadEntry{
		ModID:      "mod1",
		Timestamp:  12345,
		OldVersion: "1.0.0",
		NewVersion: "1.1.0",
		Success:    true,
	})

	data, err := comp.Serialize()
	if err != nil {
		t.Fatalf("serialize failed: %v", err)
	}

	// Verify JSON structure
	var parsed HotReloadData
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("failed to parse serialized data: %v", err)
	}

	if !parsed.AutoReload {
		t.Error("auto reload not serialized")
	}
	if parsed.WatchInterval != 1000 {
		t.Errorf("expected watch interval 1000ms, got %d", parsed.WatchInterval)
	}
	if len(parsed.WatchedMods) != 1 {
		t.Errorf("expected 1 watched mod, got %d", len(parsed.WatchedMods))
	}
	if len(parsed.ReloadHistory) != 1 {
		t.Errorf("expected 1 history entry, got %d", len(parsed.ReloadHistory))
	}
}

func TestHotReloadComponentDeserialize(t *testing.T) {
	// Create serialized data
	original := NewHotReloadComponent()
	original.StartWatching("mod1", "1.0.0", "hash1")
	original.SetAutoReload(true)
	original.SetWatchInterval(2 * time.Second)
	original.AddReloadEntry(ReloadEntry{
		ModID:      "mod1",
		Timestamp:  12345,
		OldVersion: "1.0.0",
		NewVersion: "1.1.0",
		Success:    true,
	})

	data, _ := original.Serialize()

	// Deserialize into new component
	comp := NewHotReloadComponent()
	err := comp.Deserialize(data)
	if err != nil {
		t.Fatalf("deserialize failed: %v", err)
	}

	if !comp.IsAutoReloadEnabled() {
		t.Error("auto reload not restored")
	}
	if comp.GetWatchInterval() != 2*time.Second {
		t.Errorf("expected interval 2s, got %v", comp.GetWatchInterval())
	}
	if comp.GetWatchedModCount() != 1 {
		t.Errorf("expected 1 watched mod, got %d", comp.GetWatchedModCount())
	}
	if comp.GetTotalReloads() != 1 {
		t.Errorf("expected 1 reload, got %d", comp.GetTotalReloads())
	}

	// Transient state should be reset
	if len(comp.PendingUpdates) != 0 {
		t.Error("pending updates should be empty after deserialize")
	}
	if comp.RollbackAvailable {
		t.Error("rollback should not be available after deserialize")
	}
}

func TestHotReloadComponentDeserializeEmpty(t *testing.T) {
	comp := NewHotReloadComponent()

	data := []byte(`{}`)
	err := comp.Deserialize(data)
	if err != nil {
		t.Fatalf("deserialize empty failed: %v", err)
	}

	if comp.WatchedMods == nil {
		t.Error("WatchedMods should not be nil")
	}
	if comp.ReloadHistory == nil {
		t.Error("ReloadHistory should not be nil")
	}
}

func TestHotReloadComponentDeserializeInvalid(t *testing.T) {
	comp := NewHotReloadComponent()

	err := comp.Deserialize([]byte("invalid json"))
	if err == nil {
		t.Error("should fail on invalid JSON")
	}
}

func TestHotReloadComponentConcurrency(t *testing.T) {
	comp := NewHotReloadComponent()
	done := make(chan bool)

	// Concurrent operations
	go func() {
		for i := 0; i < 100; i++ {
			comp.StartWatching("mod1", "1.0.0", "hash")
			comp.IsWatching("mod1")
			comp.GetWatchedModIDs()
		}
		done <- true
	}()

	go func() {
		for i := 0; i < 100; i++ {
			comp.DetectChange("mod1", "newhash")
			comp.HasPendingUpdates()
			comp.GetPendingUpdates()
		}
		done <- true
	}()

	go func() {
		for i := 0; i < 100; i++ {
			comp.SaveStateForRollback("mod1", "1.0.0", nil, nil)
			comp.GetRollbackState("mod1")
			comp.ClearRollbackState("mod1")
		}
		done <- true
	}()

	go func() {
		for i := 0; i < 100; i++ {
			comp.AddReloadEntry(ReloadEntry{ModID: "mod1"})
			comp.GetReloadHistory()
			comp.GetTotalReloads()
		}
		done <- true
	}()

	// Wait for all goroutines
	for i := 0; i < 4; i++ {
		<-done
	}
}

func TestWatchedModFields(t *testing.T) {
	comp := NewHotReloadComponent()
	comp.StartWatching("mod1", "1.0.0", "hash1")

	// Detect multiple changes
	comp.DetectChange("mod1", "hash2")
	comp.DetectChange("mod1", "hash3")

	watched, _ := comp.GetWatchedMod("mod1")
	if watched.ChangeCount != 2 {
		t.Errorf("expected change count 2, got %d", watched.ChangeCount)
	}
	if watched.LastChange == 0 {
		t.Error("last change should be set")
	}
	if watched.WatchStarted == 0 {
		t.Error("watch started should be set")
	}
}

func TestGetLastReloadTime(t *testing.T) {
	comp := NewHotReloadComponent()

	if comp.GetLastReloadTime() != 0 {
		t.Error("last reload time should be 0 initially")
	}

	now := time.Now().Unix()
	comp.AddReloadEntry(ReloadEntry{
		ModID:     "mod1",
		Timestamp: now,
	})

	if comp.GetLastReloadTime() != now {
		t.Errorf("expected last reload time %d, got %d", now, comp.GetLastReloadTime())
	}
}

func TestStopWatchingRemovesPendingUpdates(t *testing.T) {
	comp := NewHotReloadComponent()
	comp.StartWatching("mod1", "1.0.0", "hash1")
	comp.DetectChange("mod1", "newhash")

	if !comp.HasPendingUpdates() {
		t.Error("should have pending updates")
	}

	comp.StopWatching("mod1")

	if comp.HasPendingUpdates() {
		t.Error("pending updates should be cleared after stop watching")
	}
}

func TestDuplicatePendingUpdates(t *testing.T) {
	comp := NewHotReloadComponent()
	comp.StartWatching("mod1", "1.0.0", "hash1")

	// Multiple changes should not duplicate pending updates
	comp.DetectChange("mod1", "hash2")
	comp.DetectChange("mod1", "hash3")
	comp.DetectChange("mod1", "hash4")

	pending := comp.GetPendingUpdates()
	if len(pending) != 1 {
		t.Errorf("expected 1 pending update (no duplicates), got %d", len(pending))
	}
}

func BenchmarkDetectChange(b *testing.B) {
	comp := NewHotReloadComponent()
	comp.StartWatching("mod1", "1.0.0", "hash0")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		comp.DetectChange("mod1", "hash"+string(rune(i%100)))
	}
}

func BenchmarkGetWatchedMod(b *testing.B) {
	comp := NewHotReloadComponent()
	comp.StartWatching("mod1", "1.0.0", "hash1")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		comp.GetWatchedMod("mod1")
	}
}

func BenchmarkSerialize(b *testing.B) {
	comp := NewHotReloadComponent()
	for i := 0; i < 10; i++ {
		comp.StartWatching("mod"+string(rune('0'+i)), "1.0.0", "hash")
	}
	for i := 0; i < 50; i++ {
		comp.AddReloadEntry(ReloadEntry{ModID: "mod0", Success: true})
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		comp.Serialize()
	}
}
