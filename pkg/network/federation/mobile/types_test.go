package mobile

import (
	"testing"
)

func TestPlatform_String(t *testing.T) {
	tests := []struct {
		name     string
		platform Platform
		want     string
	}{
		{"iOS", PlatformIOS, "iOS"},
		{"Android", PlatformAndroid, "Android"},
		{"Unknown", PlatformUnknown, "Unknown"},
		{"Invalid", Platform("invalid"), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.platform.String(); got != tt.want {
				t.Errorf("Platform.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBatteryMode_String(t *testing.T) {
	tests := []struct {
		name string
		mode BatteryMode
		want string
	}{
		{"Normal", BatteryModeNormal, "Normal"},
		{"Low", BatteryModeLow, "Low"},
		{"Critical", BatteryModeCritical, "Critical"},
		{"Invalid", BatteryMode("invalid"), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.mode.String(); got != tt.want {
				t.Errorf("BatteryMode.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSyncStatus_String(t *testing.T) {
	tests := []struct {
		name   string
		status SyncStatus
		want   string
	}{
		{"Idle", SyncStatusIdle, "Idle"},
		{"Active", SyncStatusActive, "Active"},
		{"Paused", SyncStatusPaused, "Paused"},
		{"Error", SyncStatusError, "Error"},
		{"Invalid", SyncStatus("invalid"), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.status.String(); got != tt.want {
				t.Errorf("SyncStatus.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.DeviceID != "mobile-device" {
		t.Errorf("DeviceID = %v, want mobile-device", cfg.DeviceID)
	}

	if cfg.Platform != PlatformUnknown {
		t.Errorf("Platform = %v, want %v", cfg.Platform, PlatformUnknown)
	}

	if !cfg.EnableBackground {
		t.Error("EnableBackground should be true")
	}

	if cfg.BatteryThreshold != 0.5 {
		t.Errorf("BatteryThreshold = %v, want 0.5", cfg.BatteryThreshold)
	}

	if cfg.TimeoutMultiplier != 2.0 {
		t.Errorf("TimeoutMultiplier = %v, want 2.0", cfg.TimeoutMultiplier)
	}
}

func TestNewState(t *testing.T) {
	state := NewState()

	if state.BatteryLevel != 1.0 {
		t.Errorf("BatteryLevel = %v, want 1.0", state.BatteryLevel)
	}

	if state.BatteryMode != BatteryModeNormal {
		t.Errorf("BatteryMode = %v, want %v", state.BatteryMode, BatteryModeNormal)
	}

	if state.SyncStatus != SyncStatusIdle {
		t.Errorf("SyncStatus = %v, want %v", state.SyncStatus, SyncStatusIdle)
	}
}

func TestState_BatteryLevel(t *testing.T) {
	state := NewState()

	// Test setting and getting
	state.SetBatteryLevel(0.75)
	if got := state.GetBatteryLevel(); got != 0.75 {
		t.Errorf("GetBatteryLevel() = %v, want 0.75", got)
	}

	state.SetBatteryLevel(0.25)
	if got := state.GetBatteryLevel(); got != 0.25 {
		t.Errorf("GetBatteryLevel() = %v, want 0.25", got)
	}
}

func TestState_BatteryMode(t *testing.T) {
	state := NewState()

	// Test setting and getting
	state.SetBatteryMode(BatteryModeLow)
	if got := state.GetBatteryMode(); got != BatteryModeLow {
		t.Errorf("GetBatteryMode() = %v, want %v", got, BatteryModeLow)
	}

	state.SetBatteryMode(BatteryModeCritical)
	if got := state.GetBatteryMode(); got != BatteryModeCritical {
		t.Errorf("GetBatteryMode() = %v, want %v", got, BatteryModeCritical)
	}
}

func TestState_SyncStatus(t *testing.T) {
	state := NewState()

	// Test setting and getting
	state.SetSyncStatus(SyncStatusActive)
	if got := state.GetSyncStatus(); got != SyncStatusActive {
		t.Errorf("GetSyncStatus() = %v, want %v", got, SyncStatusActive)
	}

	state.SetSyncStatus(SyncStatusPaused)
	if got := state.GetSyncStatus(); got != SyncStatusPaused {
		t.Errorf("GetSyncStatus() = %v, want %v", got, SyncStatusPaused)
	}
}

func TestState_RecordSyncSuccess(t *testing.T) {
	state := NewState()

	// Record first sync
	state.RecordSyncSuccess(1024, 2048)

	if state.SyncErrors != 0 {
		t.Errorf("SyncErrors = %v, want 0", state.SyncErrors)
	}

	bytesSent, bytesReceived, syncCount, _ := state.GetStats()
	if bytesSent != 1024 {
		t.Errorf("BytesSent = %v, want 1024", bytesSent)
	}
	if bytesReceived != 2048 {
		t.Errorf("BytesReceived = %v, want 2048", bytesReceived)
	}
	if syncCount != 1 {
		t.Errorf("SyncCount = %v, want 1", syncCount)
	}

	// Record second sync
	state.RecordSyncSuccess(512, 1024)
	bytesSent, bytesReceived, syncCount, _ = state.GetStats()
	if bytesSent != 1536 {
		t.Errorf("BytesSent = %v, want 1536", bytesSent)
	}
	if bytesReceived != 3072 {
		t.Errorf("BytesReceived = %v, want 3072", bytesReceived)
	}
	if syncCount != 2 {
		t.Errorf("SyncCount = %v, want 2", syncCount)
	}
}

func TestState_RecordSyncError(t *testing.T) {
	state := NewState()

	// Record errors
	state.RecordSyncError()
	if state.SyncErrors != 1 {
		t.Errorf("SyncErrors = %v, want 1", state.SyncErrors)
	}

	state.RecordSyncError()
	if state.SyncErrors != 2 {
		t.Errorf("SyncErrors = %v, want 2", state.SyncErrors)
	}

	// Success resets error count
	state.RecordSyncSuccess(100, 200)
	if state.SyncErrors != 0 {
		t.Errorf("SyncErrors = %v, want 0 after success", state.SyncErrors)
	}
}

func TestState_RecordBackgroundSync(t *testing.T) {
	state := NewState()

	// Record background syncs
	state.RecordBackgroundSync()
	_, _, _, bgCount := state.GetStats()
	if bgCount != 1 {
		t.Errorf("BackgroundCount = %v, want 1", bgCount)
	}

	state.RecordBackgroundSync()
	_, _, _, bgCount = state.GetStats()
	if bgCount != 2 {
		t.Errorf("BackgroundCount = %v, want 2", bgCount)
	}
}

func TestState_GetStats(t *testing.T) {
	state := NewState()

	// Record various operations
	state.RecordSyncSuccess(1000, 2000)
	state.RecordSyncSuccess(500, 1000)
	state.RecordBackgroundSync()
	state.RecordBackgroundSync()
	state.RecordBackgroundSync()

	bytesSent, bytesReceived, syncCount, bgCount := state.GetStats()

	if bytesSent != 1500 {
		t.Errorf("BytesSent = %v, want 1500", bytesSent)
	}
	if bytesReceived != 3000 {
		t.Errorf("BytesReceived = %v, want 3000", bytesReceived)
	}
	if syncCount != 2 {
		t.Errorf("SyncCount = %v, want 2", syncCount)
	}
	if bgCount != 3 {
		t.Errorf("BackgroundCount = %v, want 3", bgCount)
	}
}

// Benchmarks

func BenchmarkState_SetBatteryLevel(b *testing.B) {
	state := NewState()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		state.SetBatteryLevel(0.75)
	}
}

func BenchmarkState_GetBatteryLevel(b *testing.B) {
	state := NewState()
	state.SetBatteryLevel(0.75)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = state.GetBatteryLevel()
	}
}

func BenchmarkState_RecordSyncSuccess(b *testing.B) {
	state := NewState()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		state.RecordSyncSuccess(1024, 2048)
	}
}

func BenchmarkState_GetStats(b *testing.B) {
	state := NewState()
	state.RecordSyncSuccess(1024, 2048)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _, _ = state.GetStats()
	}
}
