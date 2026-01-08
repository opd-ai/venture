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

// Test battery mode adjustment with running adapter
func TestAdapter_AdjustSyncInterval_Running(t *testing.T) {
	adapter := NewAdapter(DefaultConfig())

	// Start the adapter
	if err := adapter.Start(); err != nil {
		t.Fatalf("Start() failed: %v", err)
	}
	defer adapter.Stop()

	// Wait a bit for sync loop to start
	time.Sleep(50 * time.Millisecond)

	tests := []struct {
		name        string
		batteryMode BatteryMode
	}{
		{"normal mode", BatteryModeNormal},
		{"low mode", BatteryModeLow},
		{"critical mode", BatteryModeCritical},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Update battery which triggers interval adjustment
			switch tt.batteryMode {
			case BatteryModeNormal:
				adapter.UpdateBatteryLevel(0.8)
			case BatteryModeLow:
				adapter.UpdateBatteryLevel(0.3)
			case BatteryModeCritical:
				adapter.UpdateBatteryLevel(0.05)
			}

			// Verify state was updated
			if got := adapter.state.GetBatteryMode(); got != tt.batteryMode {
				t.Errorf("BatteryMode = %v, want %v", got, tt.batteryMode)
			}
		})
	}
}

// Test battery mode adjustment when not running
func TestAdapter_AdjustSyncInterval_NotRunning(t *testing.T) {
	adapter := NewAdapter(DefaultConfig())

	// Don't start the adapter
	// Update battery should still work but won't adjust interval
	adapter.UpdateBatteryLevel(0.3)

	if got := adapter.state.GetBatteryMode(); got != BatteryModeLow {
		t.Errorf("BatteryMode = %v, want %v", got, BatteryModeLow)
	}
}

// Test performSync with various battery modes
func TestAdapter_PerformSync_BatteryModes(t *testing.T) {
	tests := []struct {
		name         string
		batteryLevel float64
		batteryMode  BatteryMode
	}{
		{"normal battery", 0.8, BatteryModeNormal},
		{"low battery", 0.3, BatteryModeLow},
		{"critical battery", 0.05, BatteryModeCritical},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adapter := NewAdapter(DefaultConfig())

			syncCount := 0
			adapter.RegisterSyncHandler(func(ctx context.Context) error {
				syncCount++
				return nil
			})

			adapter.state.SetBatteryLevel(tt.batteryLevel)
			adapter.state.SetBatteryMode(tt.batteryMode)

			adapter.performSync()

			if syncCount != 1 {
				t.Errorf("syncCount = %d, want 1", syncCount)
			}
		})
	}
}

// Test performSync with paused state
func TestAdapter_PerformSync_Paused(t *testing.T) {
	adapter := NewAdapter(DefaultConfig())

	syncCount := 0
	adapter.RegisterSyncHandler(func(ctx context.Context) error {
		syncCount++
		return nil
	})

	// Pause sync
	adapter.PauseSync()

	// Attempt sync while paused
	adapter.performSync()

	// Note: performSync doesn't check paused status in the implementation
	// It sets status to Active and executes the handler regardless
	// The paused status is only checked in the sync loop
	if syncCount != 1 {
		t.Errorf("syncCount = %d, want 1 (performSync executes even when paused)", syncCount)
	}

	// Status should be changed from Paused to Idle after sync
	if adapter.state.GetSyncStatus() != SyncStatusIdle {
		t.Errorf("status = %v, want %v after sync", adapter.state.GetSyncStatus(), SyncStatusIdle)
	}
}

// Test getSyncTimeout with different battery modes
func TestAdapter_GetSyncTimeout_BatteryModes(t *testing.T) {
	tests := []struct {
		name        string
		batteryMode BatteryMode
		multiplier  float64
		expected    time.Duration
	}{
		{"normal with default multiplier", BatteryModeNormal, 1.0, 30 * time.Second},
		{"normal with 2x multiplier", BatteryModeNormal, 2.0, 60 * time.Second},
		{"low with default multiplier", BatteryModeLow, 1.0, 30 * time.Second},
		{"critical with default multiplier", BatteryModeCritical, 1.0, 15 * time.Second},
		{"critical with 2x multiplier", BatteryModeCritical, 2.0, 30 * time.Second},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := DefaultConfig()
			config.TimeoutMultiplier = tt.multiplier
			adapter := NewAdapter(config)
			adapter.state.SetBatteryMode(tt.batteryMode)

			timeout := adapter.getSyncTimeout()

			if timeout != tt.expected {
				t.Errorf("timeout = %v, want %v", timeout, tt.expected)
			}
		})
	}
}

// Test performSync with handler error
func TestAdapter_PerformSync_HandlerError(t *testing.T) {
	adapter := NewAdapter(DefaultConfig())

	expectedErr := context.DeadlineExceeded
	adapter.RegisterSyncHandler(func(ctx context.Context) error {
		return expectedErr
	})

	// Perform sync with handler that returns error
	adapter.performSync()

	// Should set error status
	if adapter.state.GetSyncStatus() != SyncStatusError {
		t.Errorf("status = %v, want %v after error", adapter.state.GetSyncStatus(), SyncStatusError)
	}

	// Check that error was recorded via state
	// Note: GetStats returns (bytesSent, bytesReceived, syncCount, backgroundCount)
	// Errors are tracked separately via SyncErrors field
	_, _, syncCount, _ := adapter.state.GetStats()

	// Sync count should not increase on error
	if syncCount != 0 {
		t.Errorf("syncCount = %d, want 0 (no successful sync)", syncCount)
	}
}

// Test performSync records bytes transferred
func TestAdapter_PerformSync_BytesTracking(t *testing.T) {
	adapter := NewAdapter(DefaultConfig())

	adapter.RegisterSyncHandler(func(ctx context.Context) error {
		// Simulate successful sync with data transfer
		return nil
	})

	adapter.performSync()

	// Note: GetStats returns (bytesSent, bytesReceived, syncCount, backgroundCount)
	bytesSent, bytesReceived, syncCount, _ := adapter.state.GetStats()

	if syncCount != 1 {
		t.Errorf("syncCount = %d, want 1", syncCount)
	}

	// RecordSyncSuccess is called with (1024, 2048) in the implementation
	if bytesSent != 1024 {
		t.Errorf("bytesSent = %d, want 1024", bytesSent)
	}
	if bytesReceived != 2048 {
		t.Errorf("bytesReceived = %d, want 2048", bytesReceived)
	}

	// Verify state is idle after successful sync
	if adapter.state.GetSyncStatus() != SyncStatusIdle {
		t.Errorf("status = %v, want %v after sync", adapter.state.GetSyncStatus(), SyncStatusIdle)
	}
}

// Test sync loop executes periodically
func TestAdapter_SyncLoop_Periodic(t *testing.T) {
	config := DefaultConfig()
	config.SyncInterval = 100 * time.Millisecond // Short interval for testing
	adapter := NewAdapter(config)

	syncCount := 0
	var mu sync.Mutex

	adapter.RegisterSyncHandler(func(ctx context.Context) error {
		mu.Lock()
		syncCount++
		mu.Unlock()
		return nil
	})

	// Start adapter (starts sync loop)
	if err := adapter.Start(); err != nil {
		t.Fatalf("Start() failed: %v", err)
	}

	// Wait for at least 2 syncs to occur
	time.Sleep(350 * time.Millisecond)

	// Stop adapter
	if err := adapter.Stop(); err != nil {
		t.Fatalf("Stop() failed: %v", err)
	}

	mu.Lock()
	count := syncCount
	mu.Unlock()

	// Should have at least 2 syncs (possibly 3 depending on timing)
	if count < 2 {
		t.Errorf("syncCount = %d, want at least 2 (periodic sync)", count)
	}
}

// Test sync loop respects context cancellation
func TestAdapter_SyncLoop_ContextCancellation(t *testing.T) {
	adapter := NewAdapter(DefaultConfig())

	syncCount := 0
	var mu sync.Mutex

	adapter.RegisterSyncHandler(func(ctx context.Context) error {
		mu.Lock()
		syncCount++
		mu.Unlock()
		return nil
	})

	// Start and quickly stop
	if err := adapter.Start(); err != nil {
		t.Fatalf("Start() failed: %v", err)
	}

	time.Sleep(50 * time.Millisecond)

	if err := adapter.Stop(); err != nil {
		t.Fatalf("Stop() failed: %v", err)
	}

	// Get final count
	mu.Lock()
	finalCount := syncCount
	mu.Unlock()

	// Wait a bit more
	time.Sleep(200 * time.Millisecond)

	// Count should not increase after stop
	mu.Lock()
	laterCount := syncCount
	mu.Unlock()

	if laterCount != finalCount {
		t.Errorf("sync continued after stop: initial=%d, later=%d", finalCount, laterCount)
	}
}

// Test sync loop with battery mode changes during operation
func TestAdapter_SyncLoop_BatteryModeChanges(t *testing.T) {
	config := DefaultConfig()
	config.SyncInterval = 100 * time.Millisecond
	adapter := NewAdapter(config)

	syncCount := 0
	var mu sync.Mutex

	adapter.RegisterSyncHandler(func(ctx context.Context) error {
		mu.Lock()
		syncCount++
		mu.Unlock()
		return nil
	})

	if err := adapter.Start(); err != nil {
		t.Fatalf("Start() failed: %v", err)
	}
	defer adapter.Stop()

	// Wait for initial sync
	time.Sleep(200 * time.Millisecond)

	// Change battery modes during operation
	adapter.UpdateBatteryLevel(0.3) // Switch to low battery
	time.Sleep(200 * time.Millisecond)

	adapter.UpdateBatteryLevel(0.05) // Switch to critical
	time.Sleep(200 * time.Millisecond)

	mu.Lock()
	count := syncCount
	mu.Unlock()

	// Should have at least 1 sync (possibly more depending on timing)
	if count < 1 {
		t.Errorf("syncCount = %d, want at least 1", count)
	}

	// Verify final battery mode
	if adapter.state.GetBatteryMode() != BatteryModeCritical {
		t.Errorf("BatteryMode = %v, want %v", adapter.state.GetBatteryMode(), BatteryModeCritical)
	}
}
