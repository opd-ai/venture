package mobile

import (
	"context"
	"sync"
	"time"
)

// Platform represents the mobile platform type
type Platform string

const (
	PlatformIOS     Platform = "ios"
	PlatformAndroid Platform = "android"
	PlatformUnknown Platform = "unknown"
)

// String returns the platform name
func (p Platform) String() string {
	switch p {
	case PlatformIOS:
		return "iOS"
	case PlatformAndroid:
		return "Android"
	default:
		return "Unknown"
	}
}

// BatteryMode represents the battery optimization mode
type BatteryMode string

const (
	BatteryModeNormal   BatteryMode = "normal"   // >50% battery
	BatteryModeLow      BatteryMode = "low"      // 20-50% battery
	BatteryModeCritical BatteryMode = "critical" // <20% battery
)

// String returns the battery mode name
func (m BatteryMode) String() string {
	switch m {
	case BatteryModeNormal:
		return "Normal"
	case BatteryModeLow:
		return "Low"
	case BatteryModeCritical:
		return "Critical"
	default:
		return "Unknown"
	}
}

// SyncStatus represents the current sync status
type SyncStatus string

const (
	SyncStatusIdle   SyncStatus = "idle"
	SyncStatusActive SyncStatus = "active"
	SyncStatusPaused SyncStatus = "paused"
	SyncStatusError  SyncStatus = "error"
)

// String returns the sync status name
func (s SyncStatus) String() string {
	switch s {
	case SyncStatusIdle:
		return "Idle"
	case SyncStatusActive:
		return "Active"
	case SyncStatusPaused:
		return "Paused"
	case SyncStatusError:
		return "Error"
	default:
		return "Unknown"
	}
}

// Config holds mobile federation configuration
type Config struct {
	DeviceID          string        // Unique device identifier
	Platform          Platform      // Mobile platform
	EnableBackground  bool          // Enable background sync
	BatteryThreshold  float64       // Battery level to switch to low-power mode (0.0-1.0)
	SyncInterval      time.Duration // Base sync interval
	MaxBandwidth      int64         // Maximum bandwidth in bytes/sec (0 = unlimited)
	TimeoutMultiplier float64       // Multiplier for mobile network timeouts (default: 2.0)
}

// DefaultConfig returns default mobile configuration
func DefaultConfig() *Config {
	return &Config{
		DeviceID:          "mobile-device",
		Platform:          PlatformUnknown,
		EnableBackground:  true,
		BatteryThreshold:  0.5,
		SyncInterval:      60 * time.Second,
		MaxBandwidth:      0,
		TimeoutMultiplier: 2.0,
	}
}

// State represents the current adapter state
type State struct {
	mu              sync.RWMutex
	BatteryLevel    float64     // Current battery level (0.0-1.0)
	BatteryMode     BatteryMode // Current battery optimization mode
	SyncStatus      SyncStatus  // Current sync status
	LastSyncTime    time.Time   // Last successful sync time
	SyncErrors      int         // Consecutive sync errors
	BytesSent       int64       // Total bytes sent
	BytesReceived   int64       // Total bytes received
	SyncCount       int64       // Total sync operations
	BackgroundCount int64       // Total background syncs
}

// NewState creates a new adapter state
func NewState() *State {
	return &State{
		BatteryLevel: 1.0,
		BatteryMode:  BatteryModeNormal,
		SyncStatus:   SyncStatusIdle,
		LastSyncTime: time.Time{},
	}
}

// GetBatteryLevel returns the current battery level (thread-safe)
func (s *State) GetBatteryLevel() float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.BatteryLevel
}

// SetBatteryLevel updates the battery level (thread-safe)
func (s *State) SetBatteryLevel(level float64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.BatteryLevel = level
}

// GetBatteryMode returns the current battery mode (thread-safe)
func (s *State) GetBatteryMode() BatteryMode {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.BatteryMode
}

// SetBatteryMode updates the battery mode (thread-safe)
func (s *State) SetBatteryMode(mode BatteryMode) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.BatteryMode = mode
}

// GetSyncStatus returns the current sync status (thread-safe)
func (s *State) GetSyncStatus() SyncStatus {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.SyncStatus
}

// SetSyncStatus updates the sync status (thread-safe)
func (s *State) SetSyncStatus(status SyncStatus) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.SyncStatus = status
}

// RecordSyncSuccess records a successful sync (thread-safe)
func (s *State) RecordSyncSuccess(bytesSent, bytesReceived int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.LastSyncTime = time.Now()
	s.SyncErrors = 0
	s.BytesSent += bytesSent
	s.BytesReceived += bytesReceived
	s.SyncCount++
}

// RecordSyncError records a sync error (thread-safe)
func (s *State) RecordSyncError() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.SyncErrors++
}

// RecordBackgroundSync records a background sync (thread-safe)
func (s *State) RecordBackgroundSync() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.BackgroundCount++
}

// GetStats returns current statistics (thread-safe)
func (s *State) GetStats() (bytesSent, bytesReceived, syncCount, backgroundCount int64) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.BytesSent, s.BytesReceived, s.SyncCount, s.BackgroundCount
}

// SyncHandler is called to perform federation sync
type SyncHandler func(ctx context.Context) error

// BackgroundTask represents a background sync task
type BackgroundTask struct {
	ID          string
	ScheduledAt time.Time
	ExecutedAt  time.Time
	Status      string
	Error       error
}
