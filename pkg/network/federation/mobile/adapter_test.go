package mobile

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestNewAdapter(t *testing.T) {
	tests := []struct {
		name   string
		config *Config
	}{
		{"nil config", nil},
		{"custom config", &Config{
			DeviceID:         "test-device",
			Platform:         PlatformAndroid,
			EnableBackground: true,
			BatteryThreshold: 0.4,
			SyncInterval:     30 * time.Second,
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adapter := NewAdapter(tt.config)
			if adapter == nil {
				t.Fatal("NewAdapter() returned nil")
			}
			if adapter.state == nil {
				t.Error("state is nil")
			}
			if adapter.config == nil {
				t.Error("config is nil")
			}
		})
	}
}

func TestAdapter_StartStop(t *testing.T) {
	adapter := NewAdapter(DefaultConfig())

	// Test start
	if err := adapter.Start(); err != nil {
		t.Fatalf("Start() failed: %v", err)
	}

	if !adapter.IsRunning() {
		t.Error("IsRunning() = false, want true after Start()")
	}

	// Test double start
	if err := adapter.Start(); err == nil {
		t.Error("Start() should fail when already running")
	}

	// Test stop
	if err := adapter.Stop(); err != nil {
		t.Fatalf("Stop() failed: %v", err)
	}

	if adapter.IsRunning() {
		t.Error("IsRunning() = true, want false after Stop()")
	}

	// Test double stop
	if err := adapter.Stop(); err == nil {
		t.Error("Stop() should fail when not running")
	}
}

func TestAdapter_RegisterSyncHandler(t *testing.T) {
	adapter := NewAdapter(DefaultConfig())

	called := false
	handler := func(ctx context.Context) error {
		called = true
		return nil
	}

	adapter.RegisterSyncHandler(handler)

	// Trigger sync manually
	adapter.mu.RLock()
	h := adapter.syncHandler
	adapter.mu.RUnlock()

	if h == nil {
		t.Fatal("handler not registered")
	}

	if err := h(context.Background()); err != nil {
		t.Errorf("handler() failed: %v", err)
	}

	if !called {
		t.Error("handler was not called")
	}
}

func TestAdapter_UpdateBatteryLevel(t *testing.T) {
	tests := []struct {
		name         string
		threshold    float64
		batteryLevel float64
		wantMode     BatteryMode
	}{
		{"high battery", 0.5, 0.8, BatteryModeNormal},
		{"at threshold", 0.5, 0.5, BatteryModeNormal},
		{"low battery", 0.5, 0.3, BatteryModeLow},
		{"critical battery", 0.5, 0.1, BatteryModeCritical},
		{"zero battery", 0.5, 0.0, BatteryModeCritical},
		{"full battery", 0.5, 1.0, BatteryModeNormal},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := DefaultConfig()
			config.BatteryThreshold = tt.threshold
			adapter := NewAdapter(config)

			adapter.UpdateBatteryLevel(tt.batteryLevel)

			if got := adapter.state.GetBatteryLevel(); got != tt.batteryLevel {
				t.Errorf("BatteryLevel = %v, want %v", got, tt.batteryLevel)
			}

			if got := adapter.state.GetBatteryMode(); got != tt.wantMode {
				t.Errorf("BatteryMode = %v, want %v", got, tt.wantMode)
			}
		})
	}
}

func TestAdapter_PerformSync(t *testing.T) {
	adapter := NewAdapter(DefaultConfig())

	syncCount := 0
	var mu sync.Mutex

	adapter.RegisterSyncHandler(func(ctx context.Context) error {
		mu.Lock()
		syncCount++
		mu.Unlock()
		return nil
	})

	// Perform sync manually
	adapter.performSync()

	mu.Lock()
	count := syncCount
	mu.Unlock()

	if count != 1 {
		t.Errorf("syncCount = %v, want 1", count)
	}

	// Verify state was updated
	if adapter.state.GetSyncStatus() != SyncStatusIdle {
		t.Errorf("SyncStatus = %v, want %v", adapter.state.GetSyncStatus(), SyncStatusIdle)
	}
}

func TestAdapter_ScheduleBackgroundTask(t *testing.T) {
	tests := []struct {
		name             string
		enableBackground bool
		batteryMode      BatteryMode
		wantErr          bool
		minDelay         time.Duration
	}{
		{"disabled", false, BatteryModeNormal, true, 0},
		{"normal battery", true, BatteryModeNormal, false, time.Minute},
		{"low battery", true, BatteryModeLow, false, 5 * time.Minute},
		{"critical battery", true, BatteryModeCritical, false, 15 * time.Minute},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := DefaultConfig()
			config.EnableBackground = tt.enableBackground
			adapter := NewAdapter(config)

			adapter.state.SetBatteryMode(tt.batteryMode)

			task, err := adapter.ScheduleBackgroundTask()

			if tt.wantErr {
				if err == nil {
					t.Error("ScheduleBackgroundTask() should return error")
				}
				return
			}

			if err != nil {
				t.Fatalf("ScheduleBackgroundTask() failed: %v", err)
			}

			if task == nil {
				t.Fatal("task is nil")
			}

			if task.ID == "" {
				t.Error("task.ID is empty")
			}

			if task.Status != "scheduled" {
				t.Errorf("task.Status = %v, want scheduled", task.Status)
			}

			delay := task.ScheduledAt.Sub(time.Now())
			if delay < tt.minDelay-time.Second || delay > tt.minDelay+time.Second {
				t.Errorf("delay = %v, want ~%v", delay, tt.minDelay)
			}
		})
	}
}

func TestAdapter_ExecuteBackgroundTask(t *testing.T) {
	adapter := NewAdapter(DefaultConfig())

	called := false
	adapter.RegisterSyncHandler(func(ctx context.Context) error {
		called = true
		return nil
	})

	task := &BackgroundTask{
		ID:          "test-task",
		ScheduledAt: time.Now(),
		Status:      "scheduled",
	}

	if err := adapter.ExecuteBackgroundTask(task); err != nil {
		t.Fatalf("ExecuteBackgroundTask() failed: %v", err)
	}

	if !called {
		t.Error("sync handler was not called")
	}

	if task.Status != "completed" {
		t.Errorf("task.Status = %v, want completed", task.Status)
	}

	if task.ExecutedAt.IsZero() {
		t.Error("task.ExecutedAt is zero")
	}

	// Verify background sync was recorded
	_, _, _, bgCount := adapter.state.GetStats()
	if bgCount != 1 {
		t.Errorf("BackgroundCount = %v, want 1", bgCount)
	}
}

func TestAdapter_ExecuteBackgroundTask_NilTask(t *testing.T) {
	adapter := NewAdapter(DefaultConfig())

	if err := adapter.ExecuteBackgroundTask(nil); err == nil {
		t.Error("ExecuteBackgroundTask(nil) should return error")
	}
}

func TestAdapter_ExecuteBackgroundTask_NoHandler(t *testing.T) {
	adapter := NewAdapter(DefaultConfig())

	task := &BackgroundTask{
		ID:          "test-task",
		ScheduledAt: time.Now(),
		Status:      "scheduled",
	}

	if err := adapter.ExecuteBackgroundTask(task); err == nil {
		t.Error("ExecuteBackgroundTask() should fail with no handler")
	}
}

func TestAdapter_GetState(t *testing.T) {
	adapter := NewAdapter(DefaultConfig())

	adapter.state.SetBatteryLevel(0.75)
	adapter.state.SetBatteryMode(BatteryModeLow)
	adapter.state.RecordSyncSuccess(1024, 2048)

	state := adapter.GetState()

	if state.BatteryLevel != 0.75 {
		t.Errorf("BatteryLevel = %v, want 0.75", state.BatteryLevel)
	}

	if state.BatteryMode != BatteryModeLow {
		t.Errorf("BatteryMode = %v, want %v", state.BatteryMode, BatteryModeLow)
	}

	if state.SyncCount != 1 {
		t.Errorf("SyncCount = %v, want 1", state.SyncCount)
	}
}

func TestAdapter_GetConfig(t *testing.T) {
	config := &Config{
		DeviceID:         "test-123",
		Platform:         PlatformIOS,
		EnableBackground: false,
		BatteryThreshold: 0.3,
		SyncInterval:     45 * time.Second,
	}

	adapter := NewAdapter(config)
	got := adapter.GetConfig()

	if got.DeviceID != config.DeviceID {
		t.Errorf("DeviceID = %v, want %v", got.DeviceID, config.DeviceID)
	}

	if got.Platform != config.Platform {
		t.Errorf("Platform = %v, want %v", got.Platform, config.Platform)
	}

	if got.EnableBackground != config.EnableBackground {
		t.Errorf("EnableBackground = %v, want %v", got.EnableBackground, config.EnableBackground)
	}
}

func TestAdapter_PauseResumeSync(t *testing.T) {
	adapter := NewAdapter(DefaultConfig())

	// Initial status
	if adapter.state.GetSyncStatus() != SyncStatusIdle {
		t.Errorf("initial status = %v, want %v", adapter.state.GetSyncStatus(), SyncStatusIdle)
	}

	// Pause
	adapter.PauseSync()
	if adapter.state.GetSyncStatus() != SyncStatusPaused {
		t.Errorf("status after pause = %v, want %v", adapter.state.GetSyncStatus(), SyncStatusPaused)
	}

	// Resume
	adapter.ResumeSync()
	if adapter.state.GetSyncStatus() != SyncStatusIdle {
		t.Errorf("status after resume = %v, want %v", adapter.state.GetSyncStatus(), SyncStatusIdle)
	}
}

func TestAdapter_ConcurrentAccess(t *testing.T) {
	adapter := NewAdapter(DefaultConfig())

	if err := adapter.Start(); err != nil {
		t.Fatalf("Start() failed: %v", err)
	}
	defer adapter.Stop()

	var wg sync.WaitGroup
	iterations := 100

	// Concurrent battery updates
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < iterations; i++ {
			adapter.UpdateBatteryLevel(0.5)
		}
	}()

	// Concurrent state reads
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < iterations; i++ {
			_ = adapter.GetState()
		}
	}()

	// Concurrent config reads
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < iterations; i++ {
			_ = adapter.GetConfig()
		}
	}()

	wg.Wait()
}

// Benchmarks

func BenchmarkAdapter_UpdateBatteryLevel(b *testing.B) {
	adapter := NewAdapter(DefaultConfig())
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		adapter.UpdateBatteryLevel(0.75)
	}
}

func BenchmarkAdapter_GetState(b *testing.B) {
	adapter := NewAdapter(DefaultConfig())
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = adapter.GetState()
	}
}

func BenchmarkAdapter_ScheduleBackgroundTask(b *testing.B) {
	adapter := NewAdapter(DefaultConfig())
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = adapter.ScheduleBackgroundTask()
	}
}

func BenchmarkAdapter_PerformSync(b *testing.B) {
	adapter := NewAdapter(DefaultConfig())
	adapter.RegisterSyncHandler(func(ctx context.Context) error {
		return nil
	})
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		adapter.performSync()
	}
}
