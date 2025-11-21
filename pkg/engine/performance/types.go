package performance

import (
	"sync"
	"time"
)

// LODLevel represents level of detail quality
type LODLevel int

const (
	LODVeryHigh LODLevel = iota
	LODHigh
	LODMedium
	LODLow
	LODVeryLow
)

func (l LODLevel) String() string {
	switch l {
	case LODVeryHigh:
		return "VeryHigh"
	case LODHigh:
		return "High"
	case LODMedium:
		return "Medium"
	case LODLow:
		return "Low"
	case LODVeryLow:
		return "VeryLow"
	default:
		return "Unknown"
	}
}

// MemoryStats holds memory usage statistics
type MemoryStats struct {
	TotalBytes     uint64
	TotalMB        uint64
	Allocations    map[string]uint64
	LargestAlloc   string
	LargestAllocMB uint64
}

// NetworkStats holds network performance metrics
type NetworkStats struct {
	MessagesSent   uint64
	MessagesPerSec float64
	BytesSent      uint64
	BytesPerSec    float64
	BatchCount     uint64
	AvgBatchSize   float64
	LastBatchTime  time.Time
}

// CacheStats holds cache performance metrics
type CacheStats struct {
	CurrentSizeMB uint64
	MaxSizeMB     uint64
	ItemCount     int
	HitRate       float64
	EvictionCount uint64
	LastCleanup   time.Time
}

// LoadRequest represents a background loading request
type LoadRequest struct {
	ID       string
	Type     string // "raid", "guild_hall", "terrain", etc.
	Priority int    // 0=low, 5=high
	Callback func(interface{})
	Data     interface{}
}

// Message represents a network message
type Message struct {
	Type      string
	Data      []byte
	Timestamp time.Time
	PlayerID  string
}

// BatchedMessage groups multiple messages
type BatchedMessage struct {
	Messages  []*Message
	Timestamp time.Time
	Size      int // total bytes
}

// CacheEntry represents a cached resource
type CacheEntry struct {
	Key          string
	Data         interface{}
	SizeBytes    uint64
	LastAccessed time.Time
	AccessCount  uint64
}

// IndexConfig configures database indexing
type IndexConfig struct {
	TableName   string
	IndexFields []string
	Unique      bool
	Type        string // "btree", "hash", "gin", "gist"
}

// PerformanceConfig holds all performance tuning parameters
type PerformanceConfig struct {
	// Memory
	MaxMemoryMB       uint64
	CacheSizeMB       uint64
	SpriteCacheSizeMB uint64
	RaidCacheSizeMB   uint64

	// Network
	BatchWindowMs      int
	MaxBatchSize       int
	TargetBandwidthKBs int

	// Loading
	BackgroundLoaders int
	PreloadDistance   float64
	LoadTimeoutSec    int

	// LOD
	LODDistanceHigh   float64
	LODDistanceMedium float64
	LODDistanceLow    float64

	// Database
	IndexBatchSize int
	QueryCacheSize int
}

// DefaultPerformanceConfig returns the default V9 performance configuration
func DefaultPerformanceConfig() *PerformanceConfig {
	return &PerformanceConfig{
		MaxMemoryMB:        550, // V9 target
		CacheSizeMB:        400, // Expanded from 300MB
		SpriteCacheSizeMB:  300,
		RaidCacheSizeMB:    100,
		BatchWindowMs:      100,   // Batch every 100ms
		MaxBatchSize:       64,    // Max 64 messages per batch
		TargetBandwidthKBs: 85,    // V9 target
		BackgroundLoaders:  4,     // 4 concurrent loaders
		PreloadDistance:    500.0, // Preload within 500 units
		LoadTimeoutSec:     5,
		LODDistanceHigh:    100.0,
		LODDistanceMedium:  300.0,
		LODDistanceLow:     600.0,
		IndexBatchSize:     1000,
		QueryCacheSize:     500,
	}
}

// PerformanceMonitor tracks overall system performance
type PerformanceMonitor struct {
	mu               sync.RWMutex
	config           *PerformanceConfig
	memoryStats      *MemoryStats
	networkStats     *NetworkStats
	cacheStats       *CacheStats
	frameTimeMs      float64
	fps              float64
	lastUpdate       time.Time
	warningThreshold float64 // memory usage % to trigger warnings
}

// NewPerformanceMonitor creates a new performance monitor
func NewPerformanceMonitor() *PerformanceMonitor {
	return &PerformanceMonitor{
		config: DefaultPerformanceConfig(),
		memoryStats: &MemoryStats{
			Allocations: make(map[string]uint64),
		},
		networkStats:     &NetworkStats{},
		cacheStats:       &CacheStats{},
		lastUpdate:       time.Now(),
		warningThreshold: 0.9, // Warn at 90% memory usage
	}
}

// UpdateFrameTime records frame timing
func (pm *PerformanceMonitor) UpdateFrameTime(deltaMs float64) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	pm.frameTimeMs = deltaMs
	pm.fps = 1000.0 / deltaMs
	pm.lastUpdate = time.Now()
}

// GetFrameTime returns current frame time in milliseconds
func (pm *PerformanceMonitor) GetFrameTime() float64 {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	return pm.frameTimeMs
}

// GetFPS returns current frames per second
func (pm *PerformanceMonitor) GetFPS() float64 {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	return pm.fps
}

// UpdateMemoryStats updates memory statistics
func (pm *PerformanceMonitor) UpdateMemoryStats(stats *MemoryStats) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	pm.memoryStats = stats
}

// UpdateNetworkStats updates network statistics
func (pm *PerformanceMonitor) UpdateNetworkStats(stats *NetworkStats) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	pm.networkStats = stats
}

// UpdateCacheStats updates cache statistics
func (pm *PerformanceMonitor) UpdateCacheStats(stats *CacheStats) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	pm.cacheStats = stats
}

// GetMemoryStats returns current memory statistics
func (pm *PerformanceMonitor) GetMemoryStats() *MemoryStats {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	// Deep copy
	stats := &MemoryStats{
		TotalBytes:     pm.memoryStats.TotalBytes,
		TotalMB:        pm.memoryStats.TotalMB,
		Allocations:    make(map[string]uint64),
		LargestAlloc:   pm.memoryStats.LargestAlloc,
		LargestAllocMB: pm.memoryStats.LargestAllocMB,
	}
	for k, v := range pm.memoryStats.Allocations {
		stats.Allocations[k] = v
	}
	return stats
}

// GetNetworkStats returns current network statistics
func (pm *PerformanceMonitor) GetNetworkStats() *NetworkStats {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	stats := *pm.networkStats
	return &stats
}

// GetCacheStats returns current cache statistics
func (pm *PerformanceMonitor) GetCacheStats() *CacheStats {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	stats := *pm.cacheStats
	return &stats
}

// CheckMemoryWarning returns true if memory usage exceeds threshold
func (pm *PerformanceMonitor) CheckMemoryWarning() bool {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	usage := float64(pm.memoryStats.TotalMB) / float64(pm.config.MaxMemoryMB)
	return usage >= pm.warningThreshold
}

// CheckPerformanceTarget returns true if all performance targets are met
func (pm *PerformanceMonitor) CheckPerformanceTarget() bool {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	// Check frame time (60 FPS = 16.67ms)
	if pm.frameTimeMs > 16.67 {
		return false
	}

	// Check memory
	if pm.memoryStats.TotalMB > pm.config.MaxMemoryMB {
		return false
	}

	// Check network bandwidth
	if pm.networkStats.BytesPerSec > float64(pm.config.TargetBandwidthKBs*1024) {
		return false
	}

	return true
}

// GetConfig returns the performance configuration
func (pm *PerformanceMonitor) GetConfig() *PerformanceConfig {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	config := *pm.config
	return &config
}

// SetConfig updates the performance configuration
func (pm *PerformanceMonitor) SetConfig(config *PerformanceConfig) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	pm.config = config
}
