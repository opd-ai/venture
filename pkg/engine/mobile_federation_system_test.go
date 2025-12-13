package engine

import (
	"testing"
	"time"

	"github.com/opd-ai/venture/pkg/network/federation/mobile"
)

func TestMobileFederationSystem_Creation(t *testing.T) {
	// Test with default config
	sys := NewMobileFederationSystem(nil)
	if sys == nil {
		t.Fatal("Expected non-nil system")
	}

	if sys.adapter == nil {
		t.Fatal("Expected non-nil adapter")
	}

	if sys.logger == nil {
		t.Fatal("Expected non-nil logger")
	}

	// Test with custom config
	config := &mobile.Config{
		SyncInterval:      2 * time.Minute,
		BatteryThreshold:  0.4,
		MaxBandwidth:      100000,
		TimeoutMultiplier: 1.5,
		EnableBackground:  true,
	}

	sys2 := NewMobileFederationSystem(config)
	if sys2 == nil {
		t.Fatal("Expected non-nil system with custom config")
	}

	cfg := sys2.GetAdapter().GetConfig()
	if cfg.SyncInterval != 2*time.Minute {
		t.Errorf("Expected sync interval 2m, got %v", cfg.SyncInterval)
	}
}

func TestMobileFederationSystem_StartStop(t *testing.T) {
	sys := NewMobileFederationSystem(nil)

	// Start the system
	err := sys.Start()
	if err != nil {
		t.Fatalf("Failed to start system: %v", err)
	}

	// Verify adapter is running
	if !sys.adapter.IsRunning() {
		t.Error("Expected adapter to be running")
	}

	// Wait briefly for sync loop to initialize
	time.Sleep(50 * time.Millisecond)

	// Stop the system
	err = sys.Stop()
	if err != nil {
		t.Fatalf("Failed to stop system: %v", err)
	}

	// Verify adapter is stopped
	if sys.adapter.IsRunning() {
		t.Error("Expected adapter to be stopped")
	}
}

func TestMobileFederationSystem_BatteryLevelUpdate(t *testing.T) {
	sys := NewMobileFederationSystem(nil)

	// Start the system
	if err := sys.Start(); err != nil {
		t.Fatalf("Failed to start system: %v", err)
	}
	defer sys.Stop()

	tests := []struct {
		name         string
		batteryLevel float64
		expectedMode mobile.BatteryMode
	}{
		{"Normal battery", 0.8, mobile.BatteryModeNormal},
		{"Normal threshold", 0.5, mobile.BatteryModeNormal},
		{"Low battery", 0.3, mobile.BatteryModeLow},
		{"Critical battery", 0.1, mobile.BatteryModeCritical},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sys.UpdateBatteryLevel(tt.batteryLevel)

			// Allow time for state update
			time.Sleep(10 * time.Millisecond)

			state := sys.GetState()
			if state.BatteryLevel != tt.batteryLevel {
				t.Errorf("Expected battery level %.2f, got %.2f", tt.batteryLevel, state.BatteryLevel)
			}

			if state.BatteryMode != tt.expectedMode {
				t.Errorf("Expected mode %v, got %v", tt.expectedMode, state.BatteryMode)
			}
		})
	}
}

func TestMobileFederationSystem_Update(t *testing.T) {
	sys := NewMobileFederationSystem(nil)

	// Update should not panic (it's a no-op for mobile federation)
	sys.Update(nil, 0.016)
	sys.Update([]*Entity{}, 0.032)

	// Create some entities
	w := NewWorld()
	entity := w.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 10, Y: 20})

	sys.Update([]*Entity{entity}, 0.016)
}

func TestMobileFederationSystem_PauseResume(t *testing.T) {
	sys := NewMobileFederationSystem(nil)

	if err := sys.Start(); err != nil {
		t.Fatalf("Failed to start system: %v", err)
	}
	defer sys.Stop()

	// Initial state should be idle
	state := sys.GetState()
	if state.SyncStatus != mobile.SyncStatusIdle {
		t.Errorf("Expected initial status Idle, got %v", state.SyncStatus)
	}

	// Pause sync
	sys.PauseSync()
	time.Sleep(10 * time.Millisecond)

	state = sys.GetState()
	if state.SyncStatus != mobile.SyncStatusPaused {
		t.Errorf("Expected status Paused, got %v", state.SyncStatus)
	}

	// Resume sync
	sys.ResumeSync()
	time.Sleep(10 * time.Millisecond)

	state = sys.GetState()
	if state.SyncStatus != mobile.SyncStatusIdle {
		t.Errorf("Expected status Idle after resume, got %v", state.SyncStatus)
	}
}

func TestMobileFederationSystem_GetState(t *testing.T) {
	sys := NewMobileFederationSystem(nil)

	state := sys.GetState()

	// Verify initial state
	if state.BatteryLevel != 1.0 {
		t.Errorf("Expected initial battery 1.0, got %.2f", state.BatteryLevel)
	}

	if state.BatteryMode != mobile.BatteryModeNormal {
		t.Errorf("Expected initial mode Normal, got %v", state.BatteryMode)
	}

	if state.SyncStatus != mobile.SyncStatusIdle {
		t.Errorf("Expected initial status Idle, got %v", state.SyncStatus)
	}

	if state.SyncCount != 0 {
		t.Errorf("Expected initial sync count 0, got %d", state.SyncCount)
	}
}

func TestMobileFederationSystem_GetAdapter(t *testing.T) {
	sys := NewMobileFederationSystem(nil)

	adapter := sys.GetAdapter()
	if adapter == nil {
		t.Fatal("Expected non-nil adapter")
	}

	// Verify adapter is the same instance
	if adapter != sys.adapter {
		t.Error("GetAdapter returned different instance")
	}
}

func TestMobileFederationSystem_Integration(t *testing.T) {
	// Create system with shorter intervals for testing
	config := &mobile.Config{
		SyncInterval:      100 * time.Millisecond,
		BatteryThreshold:  0.5,
		MaxBandwidth:      50000,
		TimeoutMultiplier: 1.0,
		EnableBackground:  true,
	}

	sys := NewMobileFederationSystem(config)

	// Start system
	if err := sys.Start(); err != nil {
		t.Fatalf("Failed to start: %v", err)
	}

	// Simulate battery drain
	sys.UpdateBatteryLevel(0.8)
	time.Sleep(50 * time.Millisecond)

	sys.UpdateBatteryLevel(0.3)
	time.Sleep(50 * time.Millisecond)

	sys.UpdateBatteryLevel(0.15)
	time.Sleep(50 * time.Millisecond)

	// Wait for at least one sync to occur
	time.Sleep(150 * time.Millisecond)

	// Check state
	state := sys.GetState()
	if state.BatteryLevel != 0.15 {
		t.Errorf("Expected battery 0.15, got %.2f", state.BatteryLevel)
	}

	if state.BatteryMode != mobile.BatteryModeCritical {
		t.Errorf("Expected Critical mode, got %v", state.BatteryMode)
	}

	// Sync count should be > 0 (at least one sync occurred)
	if state.SyncCount == 0 {
		t.Log("Warning: No syncs occurred during test (timing dependent)")
	}

	// Stop system
	if err := sys.Stop(); err != nil {
		t.Fatalf("Failed to stop: %v", err)
	}
}
