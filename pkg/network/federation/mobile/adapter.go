// Package mobile adapter implements battery-aware mobile federation management.
// This file contains the Adapter type and all its methods for coordinating
// mobile device federation with dynamic sync intervals based on battery level.
package mobile

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/opd-ai/venture/pkg/network/federation/internal/timerutil"
)

// Adapter manages mobile federation with battery optimization
type Adapter struct {
	config       *Config
	state        *State
	syncHandler  SyncHandler
	syncTicker   *time.Ticker
	stopChan     chan struct{}
	wg           sync.WaitGroup
	mu           sync.RWMutex
	running      bool
	ctx          context.Context
	cancel       context.CancelFunc
	timeProvider TimeProvider
}

// NewAdapter creates a new mobile federation adapter with real system time.
// For deterministic behavior, use NewAdapterWithTimeProvider instead.
func NewAdapter(config *Config) *Adapter {
	return NewAdapterWithTimeProvider(config, DefaultTimeProvider())
}

// NewAdapterWithTimeProvider creates a new mobile federation adapter with a custom time source.
func NewAdapterWithTimeProvider(config *Config, tp TimeProvider) *Adapter {
	if config == nil {
		config = DefaultConfig()
	}
	if tp == nil {
		tp = DefaultTimeProvider()
	}
	if config.MaxBandwidth < 0 {
		config.MaxBandwidth = 0
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &Adapter{
		config:       config,
		state:        NewStateWithTimeProvider(tp),
		stopChan:     make(chan struct{}),
		ctx:          ctx,
		cancel:       cancel,
		timeProvider: tp,
	}
}

// Start begins mobile federation
func (a *Adapter) Start() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.running {
		return fmt.Errorf("adapter already running")
	}

	// Initialize sync ticker with base interval
	a.syncTicker = time.NewTicker(a.config.SyncInterval)
	a.running = true

	// Start sync loop
	a.wg.Add(1)
	go a.syncLoop()

	return nil
}

// Stop halts mobile federation
func (a *Adapter) Stop() error {
	a.mu.Lock()

	if !a.running {
		a.mu.Unlock()
		return fmt.Errorf("adapter not running")
	}

	a.running = false
	a.cancel()
	close(a.stopChan)

	if a.syncTicker != nil {
		a.syncTicker.Stop()
	}
	a.mu.Unlock()

	a.wg.Wait()
	return nil
}

// RegisterSyncHandler sets the sync handler function
func (a *Adapter) RegisterSyncHandler(handler SyncHandler) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.syncHandler = handler
}

// UpdateBatteryLevel updates battery level and adjusts sync mode
func (a *Adapter) UpdateBatteryLevel(level float64) {
	if level < 0.0 {
		level = 0.0
	}
	if level > 1.0 {
		level = 1.0
	}
	a.state.SetBatteryLevel(level)

	// Determine battery mode
	var mode BatteryMode
	if level >= a.config.BatteryThreshold {
		mode = BatteryModeNormal
	} else if level >= 0.2 {
		mode = BatteryModeLow
	} else {
		mode = BatteryModeCritical
	}

	oldMode := a.state.GetBatteryMode()
	if mode != oldMode {
		a.state.SetBatteryMode(mode)
		a.adjustSyncInterval(mode)
	}
}

// adjustSyncInterval changes sync interval based on battery mode
func (a *Adapter) adjustSyncInterval(mode BatteryMode) {
	a.mu.Lock()
	defer a.mu.Unlock()

	if !a.running || a.syncTicker == nil {
		return
	}

	var interval time.Duration
	switch mode {
	case BatteryModeNormal:
		interval = a.config.SyncInterval // 1 minute default
	case BatteryModeLow:
		interval = 5 * time.Minute
	case BatteryModeCritical:
		interval = 15 * time.Minute
	default:
		interval = a.config.SyncInterval
	}

	// Reset ticker with new interval
	a.syncTicker.Stop()
	a.syncTicker = time.NewTicker(interval)
}

// syncLoop runs periodic sync operations
func (a *Adapter) syncLoop() {
	defer a.wg.Done()

	for {
		// Get current ticker atomically
		a.mu.RLock()
		ticker := a.syncTicker
		a.mu.RUnlock()

		if ticker == nil {
			return
		}

		select {
		case <-a.stopChan:
			return
		case <-a.ctx.Done():
			return
		case <-ticker.C:
			a.performSync()
		}
	}
}

// performSync executes a sync operation
func (a *Adapter) performSync() {
	a.mu.RLock()
	handler := a.syncHandler
	a.mu.RUnlock()

	if handler == nil {
		return
	}

	a.state.SetSyncStatus(SyncStatusActive)

	// Create timeout based on battery mode and config
	timeout := a.getSyncTimeout()
	ctx, cancel := context.WithTimeout(a.ctx, timeout)
	defer cancel()

	// Execute sync with bandwidth limiting
	err := a.executeSyncWithBandwidthLimit(ctx, handler)

	if err != nil {
		a.state.SetSyncStatus(SyncStatusError)
		a.state.RecordSyncError()
	} else {
		a.state.SetSyncStatus(SyncStatusIdle)
		// Record success with simulated bytes (real implementation would track actual bytes)
		a.state.RecordSyncSuccess(1024, 2048)
	}
}

// executeSyncWithBandwidthLimit executes sync with bandwidth limits using
// an iterative token bucket algorithm to avoid recursive stack growth.
func (a *Adapter) executeSyncWithBandwidthLimit(ctx context.Context, handler SyncHandler) error {
	// If no bandwidth limit, execute directly
	if a.config.MaxBandwidth == 0 {
		return handler(ctx)
	}

	// Implement token bucket algorithm for bandwidth limiting
	// Bucket capacity = MaxBandwidth (bytes per second)
	// Tokens refill at rate of MaxBandwidth per second
	// Each sync operation consumes tokens equal to estimated data transfer

	// Estimate data transfer size (conservative estimate: 10KB per sync)
	estimatedBytes := int64(10 * 1024)

	for {
		// Calculate token refill based on time since last sync
		a.mu.Lock()
		now := a.timeProvider.Now()
		lastSync := a.state.GetLastSyncTime()
		timeDelta := now.Sub(lastSync)

		// Refill tokens based on elapsed time
		tokensToAdd := int64(timeDelta.Seconds() * float64(a.config.MaxBandwidth))
		currentTokens := a.state.GetBytesAvailable() + tokensToAdd

		// Cap tokens at bucket capacity (MaxBandwidth)
		if currentTokens > int64(a.config.MaxBandwidth) {
			currentTokens = int64(a.config.MaxBandwidth)
		}

		// Check if we have enough tokens
		if currentTokens >= estimatedBytes {
			// Consume tokens and execute
			a.state.SetBytesAvailable(currentTokens - estimatedBytes)
			a.mu.Unlock()
			return handler(ctx)
		}

		a.mu.Unlock()

		// Not enough bandwidth available - calculate wait time
		tokensNeeded := estimatedBytes - currentTokens
		waitTime := time.Duration(float64(tokensNeeded)/float64(a.config.MaxBandwidth)) * time.Second

		// Wait for tokens to refill or context cancellation
		if err := waitForBandwidthRefill(ctx, waitTime); err != nil {
			return err
		}
	}
}

// waitForBandwidthRefill waits for the computed bandwidth-delay window while
// honoring context cancellation and ensuring timer resources are cleaned up.
func waitForBandwidthRefill(ctx context.Context, waitTime time.Duration) error {
	timer := time.NewTimer(waitTime)
	defer timerutil.StopAndDrain(timer)

	select {
	case <-timer.C:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// getSyncTimeout returns appropriate timeout based on battery mode and config
func (a *Adapter) getSyncTimeout() time.Duration {
	baseTimeout := 30 * time.Second
	mode := a.state.GetBatteryMode()

	// Reduce timeout in critical battery mode
	if mode == BatteryModeCritical {
		baseTimeout = 15 * time.Second
	}

	return time.Duration(float64(baseTimeout) * a.config.TimeoutMultiplier)
}

// ScheduleBackgroundTask schedules a background sync task
func (a *Adapter) ScheduleBackgroundTask() (*BackgroundTask, error) {
	if !a.config.EnableBackground {
		return nil, fmt.Errorf("background sync not enabled")
	}

	mode := a.state.GetBatteryMode()
	var delay time.Duration

	// Determine delay based on battery mode
	switch mode {
	case BatteryModeNormal:
		delay = time.Minute
	case BatteryModeLow:
		delay = 5 * time.Minute
	case BatteryModeCritical:
		delay = 15 * time.Minute
	default:
		delay = time.Minute
	}

	task := &BackgroundTask{
		ID:          fmt.Sprintf("bg-%d", a.timeProvider.Now().Unix()),
		ScheduledAt: a.timeProvider.Now().Add(delay),
		Status:      "scheduled",
	}

	return task, nil
}

// ExecuteBackgroundTask executes a background sync task
func (a *Adapter) ExecuteBackgroundTask(task *BackgroundTask) error {
	if task == nil {
		return fmt.Errorf("task is nil")
	}

	a.mu.RLock()
	handler := a.syncHandler
	a.mu.RUnlock()

	if handler == nil {
		return fmt.Errorf("no sync handler registered")
	}

	task.ExecutedAt = a.timeProvider.Now()
	task.Status = "executing"

	// Use shorter timeout for background tasks
	ctx, cancel := context.WithTimeout(a.ctx, 15*time.Second)
	defer cancel()

	err := handler(ctx)
	if err != nil {
		task.Status = "failed"
		task.Error = err
		return err
	}

	task.Status = "completed"
	a.state.RecordBackgroundSync()
	a.state.RecordSyncSuccess(512, 1024) // Simulated bytes for background sync

	return nil
}

// GetState returns current adapter state (thread-safe copy)
func (a *Adapter) GetState() State {
	a.mu.RLock()
	defer a.mu.RUnlock()

	// Get all fields atomically from state
	batteryLevel, batteryMode, syncStatus, lastSyncTime, syncErrors, bytesSent, bytesReceived, syncCount, backgroundCount := a.state.GetAllFields()

	// Return a copy to avoid race conditions
	return State{
		BatteryLevel:    batteryLevel,
		BatteryMode:     batteryMode,
		SyncStatus:      syncStatus,
		LastSyncTime:    lastSyncTime,
		SyncErrors:      syncErrors,
		BytesSent:       bytesSent,
		BytesReceived:   bytesReceived,
		SyncCount:       syncCount,
		BackgroundCount: backgroundCount,
	}
}

// GetConfig returns adapter configuration (copy)
func (a *Adapter) GetConfig() Config {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return *a.config
}

// IsRunning returns whether adapter is currently running
func (a *Adapter) IsRunning() bool {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.running
}

// PauseSync pauses sync operations
func (a *Adapter) PauseSync() {
	a.state.SetSyncStatus(SyncStatusPaused)
}

// ResumeSync resumes sync operations
func (a *Adapter) ResumeSync() {
	if a.state.GetSyncStatus() == SyncStatusPaused {
		a.state.SetSyncStatus(SyncStatusIdle)
	}
}
