package engine

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestHotReloadSystemNew(t *testing.T) {
	world := NewWorld()
	system := NewHotReloadSystem(world)

	if system == nil {
		t.Fatal("system should not be nil")
	}
	if system.world != world {
		t.Error("world should be set")
	}
}

func TestHotReloadSystemUpdate(t *testing.T) {
	world := NewWorld()
	system := NewHotReloadSystem(world)

	entity := world.CreateEntity()
	comp := NewHotReloadComponent()
	entity.AddComponent(comp)

	// Should not panic with no file watcher
	system.Update([]*Entity{entity}, 0.016)
}

func TestHotReloadSystemUpdateDisabled(t *testing.T) {
	world := NewWorld()
	system := NewHotReloadSystem(world)

	entity := world.CreateEntity()
	comp := NewHotReloadComponent()
	comp.SetEnabled(false)
	entity.AddComponent(comp)

	// Should skip disabled components
	system.Update([]*Entity{entity}, 0.016)
}

func TestHotReloadSystemSetCallbacks(t *testing.T) {
	world := NewWorld()
	system := NewHotReloadSystem(world)

	reloadCalled := false
	rollbackCalled := false
	hashCalled := false

	system.SetReloadCallback(func(modID string, data []byte) error {
		reloadCalled = true
		return nil
	})

	system.SetRollbackCallback(func(modID string, state *ModState) error {
		rollbackCalled = true
		return nil
	})

	system.SetHashCallback(func(modID string) (string, error) {
		hashCalled = true
		return "hash", nil
	})

	// Callbacks are set but not called until needed
	if reloadCalled || rollbackCalled || hashCalled {
		t.Error("callbacks should not be called on set")
	}
}

func TestHotReloadSystemStartWatchingMod(t *testing.T) {
	world := NewWorld()
	system := NewHotReloadSystem(world)

	watcher := NewInMemoryFileWatcher()
	watcher.AddMod("mod1", []byte("content"), "1.0.0")
	system.SetFileWatcher(watcher)

	comp := NewHotReloadComponent()

	err := system.StartWatchingMod(comp, "mod1")
	if err != nil {
		t.Fatalf("start watching failed: %v", err)
	}

	if !comp.IsWatching("mod1") {
		t.Error("mod1 should be watched")
	}

	// Starting again should not error
	err = system.StartWatchingMod(comp, "mod1")
	if err != nil {
		t.Fatalf("start watching again should not error: %v", err)
	}
}

func TestHotReloadSystemStartWatchingModNilComp(t *testing.T) {
	world := NewWorld()
	system := NewHotReloadSystem(world)

	err := system.StartWatchingMod(nil, "mod1")
	if err == nil {
		t.Error("should error on nil component")
	}
}

func TestHotReloadSystemStopWatchingMod(t *testing.T) {
	world := NewWorld()
	system := NewHotReloadSystem(world)

	comp := NewHotReloadComponent()
	comp.StartWatching("mod1", "1.0.0", "hash1")
	comp.SaveStateForRollback("mod1", "1.0.0", nil, nil)

	system.StopWatchingMod(comp, "mod1")

	if comp.IsWatching("mod1") {
		t.Error("mod1 should not be watched")
	}
	if _, exists := comp.GetRollbackState("mod1"); exists {
		t.Error("rollback state should be cleared")
	}
}

func TestHotReloadSystemStopWatchingModNilComp(t *testing.T) {
	world := NewWorld()
	system := NewHotReloadSystem(world)

	// Should not panic
	system.StopWatchingMod(nil, "mod1")
}

func TestHotReloadSystemReloadMod(t *testing.T) {
	world := NewWorld()
	system := NewHotReloadSystem(world)

	watcher := NewInMemoryFileWatcher()
	watcher.AddMod("mod1", []byte("content v1"), "1.0.0")
	system.SetFileWatcher(watcher)

	reloaded := false
	system.SetReloadCallback(func(modID string, data []byte) error {
		reloaded = true
		if modID != "mod1" {
			t.Errorf("expected mod1, got %s", modID)
		}
		return nil
	})

	comp := NewHotReloadComponent()
	comp.StartWatching("mod1", "1.0.0", "hash1")

	err := system.ReloadMod(comp, "mod1")
	if err != nil {
		t.Fatalf("reload failed: %v", err)
	}

	if !reloaded {
		t.Error("reload callback should have been called")
	}

	// Check history entry
	history := comp.GetReloadHistory()
	if len(history) != 1 {
		t.Fatalf("expected 1 history entry, got %d", len(history))
	}
	if !history[0].Success {
		t.Error("reload should be successful")
	}
	if history[0].ModID != "mod1" {
		t.Errorf("expected mod1 in history, got %s", history[0].ModID)
	}
}

func TestHotReloadSystemReloadModNilComp(t *testing.T) {
	world := NewWorld()
	system := NewHotReloadSystem(world)

	err := system.ReloadMod(nil, "mod1")
	if err == nil {
		t.Error("should error on nil component")
	}
}

func TestHotReloadSystemReloadModNotWatched(t *testing.T) {
	world := NewWorld()
	system := NewHotReloadSystem(world)

	comp := NewHotReloadComponent()

	err := system.ReloadMod(comp, "mod1")
	if err == nil {
		t.Error("should error when mod is not being watched")
	}
}

func TestHotReloadSystemReloadModNoWatcher(t *testing.T) {
	world := NewWorld()
	system := NewHotReloadSystem(world)

	comp := NewHotReloadComponent()
	comp.StartWatching("mod1", "1.0.0", "hash1")

	err := system.ReloadMod(comp, "mod1")
	if err == nil {
		t.Error("should error when no file watcher configured")
	}

	// Check failure recorded
	if comp.GetFailedReloads() != 1 {
		t.Error("failure should be recorded")
	}
}

func TestHotReloadSystemReloadModCallbackFails(t *testing.T) {
	world := NewWorld()
	system := NewHotReloadSystem(world)

	watcher := NewInMemoryFileWatcher()
	watcher.AddMod("mod1", []byte("content"), "1.0.0")
	system.SetFileWatcher(watcher)

	system.SetReloadCallback(func(modID string, data []byte) error {
		return fmt.Errorf("reload failed")
	})

	comp := NewHotReloadComponent()
	comp.StartWatching("mod1", "1.0.0", "hash1")

	err := system.ReloadMod(comp, "mod1")
	if err == nil {
		t.Error("should error when reload callback fails")
	}

	// Check failure recorded
	history := comp.GetReloadHistory()
	if len(history) != 1 || history[0].Success {
		t.Error("failure should be recorded")
	}
}

func TestHotReloadSystemReloadWithStateMigration(t *testing.T) {
	world := NewWorld()
	system := NewHotReloadSystem(world)

	watcher := NewInMemoryFileWatcher()
	watcher.AddMod("mod1", []byte("content"), "1.0.0")
	system.SetFileWatcher(watcher)

	migrationHandler := NewInMemoryStateMigrationHandler()
	migrationHandler.SetState("mod1", map[string]any{"script": "data"}, map[string]any{"var": 42})
	system.SetMigrationHandler(migrationHandler)

	system.SetReloadCallback(func(modID string, data []byte) error {
		return nil
	})

	comp := NewHotReloadComponent()
	comp.StartWatching("mod1", "1.0.0", "hash1")

	err := system.ReloadMod(comp, "mod1")
	if err != nil {
		t.Fatalf("reload failed: %v", err)
	}

	// State should be restored
	scripts, vars, _ := migrationHandler.SaveState("mod1")
	if scripts["script"] != "data" {
		t.Error("script state should be preserved")
	}
	if vars["var"] != 42 {
		t.Error("variable state should be preserved")
	}
}

func TestHotReloadSystemRollbackMod(t *testing.T) {
	world := NewWorld()
	system := NewHotReloadSystem(world)

	rollbackCalled := false
	system.SetRollbackCallback(func(modID string, state *ModState) error {
		rollbackCalled = true
		if modID != "mod1" {
			t.Errorf("expected mod1, got %s", modID)
		}
		if state.Version != "1.0.0" {
			t.Errorf("expected version 1.0.0, got %s", state.Version)
		}
		return nil
	})

	comp := NewHotReloadComponent()
	comp.SaveStateForRollback("mod1", "1.0.0", nil, nil)

	err := system.RollbackMod(comp, "mod1")
	if err != nil {
		t.Fatalf("rollback failed: %v", err)
	}

	if !rollbackCalled {
		t.Error("rollback callback should have been called")
	}

	// Check rollback state cleared
	if _, exists := comp.GetRollbackState("mod1"); exists {
		t.Error("rollback state should be cleared after rollback")
	}

	// Check history
	history := comp.GetReloadHistory()
	found := false
	for _, entry := range history {
		if entry.RolledBack {
			found = true
			break
		}
	}
	if !found {
		t.Error("rollback should be recorded in history")
	}
}

func TestHotReloadSystemRollbackModNilComp(t *testing.T) {
	world := NewWorld()
	system := NewHotReloadSystem(world)

	err := system.RollbackMod(nil, "mod1")
	if err == nil {
		t.Error("should error on nil component")
	}
}

func TestHotReloadSystemRollbackModNoState(t *testing.T) {
	world := NewWorld()
	system := NewHotReloadSystem(world)

	comp := NewHotReloadComponent()

	err := system.RollbackMod(comp, "mod1")
	if err == nil {
		t.Error("should error when no rollback state available")
	}
}

func TestHotReloadSystemRollbackModNoCallback(t *testing.T) {
	world := NewWorld()
	system := NewHotReloadSystem(world)

	comp := NewHotReloadComponent()
	comp.SaveStateForRollback("mod1", "1.0.0", nil, nil)

	err := system.RollbackMod(comp, "mod1")
	if err == nil {
		t.Error("should error when no rollback callback configured")
	}
}

func TestHotReloadSystemForceReloadAll(t *testing.T) {
	world := NewWorld()
	system := NewHotReloadSystem(world)

	watcher := NewInMemoryFileWatcher()
	watcher.AddMod("mod1", []byte("content1"), "1.0.0")
	watcher.AddMod("mod2", []byte("content2"), "1.0.0")
	system.SetFileWatcher(watcher)

	reloadedMods := make(map[string]bool)
	var mu sync.Mutex
	system.SetReloadCallback(func(modID string, data []byte) error {
		mu.Lock()
		reloadedMods[modID] = true
		mu.Unlock()
		return nil
	})

	comp := NewHotReloadComponent()
	comp.StartWatching("mod1", "1.0.0", "hash1")
	comp.StartWatching("mod2", "1.0.0", "hash2")

	errors := system.ForceReloadAll(comp)
	if len(errors) != 0 {
		t.Errorf("expected no errors, got %v", errors)
	}

	if !reloadedMods["mod1"] || !reloadedMods["mod2"] {
		t.Error("both mods should be reloaded")
	}
}

func TestHotReloadSystemForceReloadAllNilComp(t *testing.T) {
	world := NewWorld()
	system := NewHotReloadSystem(world)

	errors := system.ForceReloadAll(nil)
	if len(errors) != 1 {
		t.Error("should return error for nil component")
	}
}

func TestHotReloadSystemForceReloadAllPartialFailure(t *testing.T) {
	world := NewWorld()
	system := NewHotReloadSystem(world)

	watcher := NewInMemoryFileWatcher()
	watcher.AddMod("mod1", []byte("content1"), "1.0.0")
	watcher.AddMod("mod2", []byte("content2"), "1.0.0")
	system.SetFileWatcher(watcher)

	system.SetReloadCallback(func(modID string, data []byte) error {
		if modID == "mod2" {
			return fmt.Errorf("mod2 failed")
		}
		return nil
	})

	comp := NewHotReloadComponent()
	comp.StartWatching("mod1", "1.0.0", "hash1")
	comp.StartWatching("mod2", "1.0.0", "hash2")

	errors := system.ForceReloadAll(comp)
	if len(errors) != 1 {
		t.Errorf("expected 1 error, got %d", len(errors))
	}
}

func TestHotReloadSystemGetReloadStatistics(t *testing.T) {
	world := NewWorld()
	system := NewHotReloadSystem(world)

	comp := NewHotReloadComponent()
	comp.AddReloadEntry(ReloadEntry{Success: true})
	comp.AddReloadEntry(ReloadEntry{Success: true})
	comp.AddReloadEntry(ReloadEntry{Success: false})

	total, successful, failed := system.GetReloadStatistics(comp)

	if total != 3 {
		t.Errorf("expected total 3, got %d", total)
	}
	if successful != 2 {
		t.Errorf("expected successful 2, got %d", successful)
	}
	if failed != 1 {
		t.Errorf("expected failed 1, got %d", failed)
	}
}

func TestHotReloadSystemGetReloadStatisticsNilComp(t *testing.T) {
	world := NewWorld()
	system := NewHotReloadSystem(world)

	total, successful, failed := system.GetReloadStatistics(nil)

	if total != 0 || successful != 0 || failed != 0 {
		t.Error("all stats should be 0 for nil component")
	}
}

func TestHotReloadSystemAutoReload(t *testing.T) {
	world := NewWorld()
	system := NewHotReloadSystem(world)

	watcher := NewInMemoryFileWatcher()
	watcher.AddMod("mod1", []byte("content v1"), "1.0.0")
	system.SetFileWatcher(watcher)

	reloaded := false
	system.SetReloadCallback(func(modID string, data []byte) error {
		reloaded = true
		return nil
	})

	entity := world.CreateEntity()
	comp := NewHotReloadComponent()
	comp.SetAutoReload(true)
	comp.SetWatchInterval(1 * time.Millisecond)
	comp.StartWatching("mod1", "1.0.0", ComputeHash([]byte("content v1")))
	entity.AddComponent(comp)

	// Update with file watcher - simulate change
	watcher.UpdateMod("mod1", []byte("content v2"), "1.1.0")

	// First update should detect change
	system.Update([]*Entity{entity}, 0.016)
	time.Sleep(2 * time.Millisecond)
	system.Update([]*Entity{entity}, 0.016)

	// Should have triggered auto reload
	if !reloaded {
		t.Error("auto reload should have triggered")
	}
}

func TestInMemoryFileWatcher(t *testing.T) {
	watcher := NewInMemoryFileWatcher()

	// Test add mod
	watcher.AddMod("mod1", []byte("content"), "1.0.0")

	hash, err := watcher.GetFileHash("mod1")
	if err != nil {
		t.Fatalf("get hash failed: %v", err)
	}
	if hash == "" {
		t.Error("hash should not be empty")
	}

	data, err := watcher.GetModData("mod1")
	if err != nil {
		t.Fatalf("get data failed: %v", err)
	}
	if string(data) != "content" {
		t.Errorf("expected 'content', got %s", string(data))
	}

	version, err := watcher.GetModVersion("mod1")
	if err != nil {
		t.Fatalf("get version failed: %v", err)
	}
	if version != "1.0.0" {
		t.Errorf("expected 1.0.0, got %s", version)
	}

	// Test update mod
	watcher.UpdateMod("mod1", []byte("new content"), "1.1.0")

	newHash, _ := watcher.GetFileHash("mod1")
	if newHash == hash {
		t.Error("hash should change after update")
	}

	newVersion, _ := watcher.GetModVersion("mod1")
	if newVersion != "1.1.0" {
		t.Errorf("expected 1.1.0, got %s", newVersion)
	}

	// Test non-existent mod
	_, err = watcher.GetFileHash("nonexistent")
	if err == nil {
		t.Error("should error for non-existent mod")
	}

	_, err = watcher.GetModData("nonexistent")
	if err == nil {
		t.Error("should error for non-existent mod")
	}

	_, err = watcher.GetModVersion("nonexistent")
	if err == nil {
		t.Error("should error for non-existent mod")
	}
}

func TestComputeHash(t *testing.T) {
	data1 := []byte("hello")
	data2 := []byte("world")

	hash1 := ComputeHash(data1)
	hash2 := ComputeHash(data2)

	if hash1 == "" {
		t.Error("hash should not be empty")
	}
	if hash1 == hash2 {
		t.Error("different data should produce different hashes")
	}

	// Same data should produce same hash (deterministic)
	hash1Again := ComputeHash(data1)
	if hash1 != hash1Again {
		t.Error("same data should produce same hash")
	}
}

func TestInMemoryStateMigrationHandler(t *testing.T) {
	handler := NewInMemoryStateMigrationHandler()

	scripts := map[string]any{"script1": "code"}
	variables := map[string]any{"var1": 42}

	handler.SetState("mod1", scripts, variables)

	savedScripts, savedVars, err := handler.SaveState("mod1")
	if err != nil {
		t.Fatalf("save state failed: %v", err)
	}
	if savedScripts["script1"] != "code" {
		t.Error("scripts not saved correctly")
	}
	if savedVars["var1"] != 42 {
		t.Error("variables not saved correctly")
	}

	// Test restore state
	newScripts := map[string]any{"script2": "new code"}
	newVars := map[string]any{"var2": 100}
	err = handler.RestoreState("mod1", newScripts, newVars)
	if err != nil {
		t.Fatalf("restore state failed: %v", err)
	}

	restoredScripts, restoredVars, _ := handler.SaveState("mod1")
	if restoredScripts["script2"] != "new code" {
		t.Error("scripts not restored correctly")
	}
	if restoredVars["var2"] != 100 {
		t.Error("variables not restored correctly")
	}

	// Test non-existent mod
	scripts, vars, err := handler.SaveState("nonexistent")
	if err != nil {
		t.Fatalf("should not error: %v", err)
	}
	if scripts == nil || vars == nil {
		t.Error("should return empty maps")
	}
}

func TestHotReloadSystemCheckForChanges(t *testing.T) {
	world := NewWorld()
	system := NewHotReloadSystem(world)

	hashCalled := false
	system.SetHashCallback(func(modID string) (string, error) {
		hashCalled = true
		return "newhash", nil
	})

	entity := world.CreateEntity()
	comp := NewHotReloadComponent()
	comp.SetWatchInterval(1 * time.Millisecond)
	comp.StartWatching("mod1", "1.0.0", "oldhash")
	entity.AddComponent(comp)

	// Wait for interval
	time.Sleep(2 * time.Millisecond)

	// Update should check for changes
	system.Update([]*Entity{entity}, 0.016)

	if !hashCalled {
		t.Error("hash callback should be called")
	}

	if !comp.HasPendingUpdates() {
		t.Error("should have pending updates after hash change")
	}
}

func TestHotReloadSystemCheckForChangesThrottled(t *testing.T) {
	world := NewWorld()
	system := NewHotReloadSystem(world)

	callCount := 0
	system.SetHashCallback(func(modID string) (string, error) {
		callCount++
		return "hash", nil
	})

	entity := world.CreateEntity()
	comp := NewHotReloadComponent()
	comp.SetWatchInterval(1 * time.Second)
	comp.StartWatching("mod1", "1.0.0", "hash")
	entity.AddComponent(comp)

	// Multiple rapid updates should be throttled
	system.Update([]*Entity{entity}, 0.016)
	system.Update([]*Entity{entity}, 0.016)
	system.Update([]*Entity{entity}, 0.016)

	if callCount > 1 {
		t.Errorf("hash callback should be throttled, got %d calls", callCount)
	}
}

func BenchmarkReloadMod(b *testing.B) {
	world := NewWorld()
	system := NewHotReloadSystem(world)

	watcher := NewInMemoryFileWatcher()
	watcher.AddMod("mod1", []byte("content"), "1.0.0")
	system.SetFileWatcher(watcher)

	system.SetReloadCallback(func(modID string, data []byte) error {
		return nil
	})

	comp := NewHotReloadComponent()
	comp.StartWatching("mod1", "1.0.0", "hash1")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		system.ReloadMod(comp, "mod1")
	}
}

func BenchmarkCheckForChanges(b *testing.B) {
	world := NewWorld()
	system := NewHotReloadSystem(world)

	system.SetHashCallback(func(modID string) (string, error) {
		return "hash", nil
	})

	entity := world.CreateEntity()
	comp := NewHotReloadComponent()
	comp.SetWatchInterval(0) // No throttling for benchmark
	for i := 0; i < 10; i++ {
		comp.StartWatching(fmt.Sprintf("mod%d", i), "1.0.0", "hash")
	}
	entity.AddComponent(comp)

	entities := []*Entity{entity}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		system.Update(entities, 0.016)
	}
}
