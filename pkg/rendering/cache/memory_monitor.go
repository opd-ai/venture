package cache

import (
	"runtime"
	"sync"
	"time"
)

// MemoryMonitor tracks cache memory usage and triggers cleanup when needed.
// Phase 44: Monitors sprite cache to ensure <300MB limit for 64x64 sprites.
type MemoryMonitor struct {
	mu sync.RWMutex

	// Cache reference
	cache *SpriteCache

	// Memory limits (bytes)
	softLimit int64 // Trigger cleanup
	hardLimit int64 // Force eviction

	// Monitoring interval
	interval time.Duration

	// Statistics
	stats MemoryStats

	// Stop channel
	stopCh   chan struct{}
	doneCh   chan struct{}
	stopOnce sync.Once
}

// MemoryStats tracks memory monitoring metrics.
type MemoryStats struct {
	CurrentUsage   int64
	PeakUsage      int64
	CleanupCount   uint64
	EvictionCount  uint64
	LastCleanupAt  time.Time
	SystemMemoryMB uint64
}

// NewMemoryMonitor creates memory monitor with default limits.
// Default: 250MB soft limit, 300MB hard limit (Phase 44 target).
func NewMemoryMonitor(cache *SpriteCache) *MemoryMonitor {
	return &MemoryMonitor{
		cache:     cache,
		softLimit: 250 * 1024 * 1024, // 250MB
		hardLimit: 300 * 1024 * 1024, // 300MB
		interval:  5 * time.Second,
		stopCh:    make(chan struct{}),
		doneCh:    make(chan struct{}),
	}
}

// SetLimits configures memory limits in bytes.
// If softLimit > hardLimit, softLimit is clamped to hardLimit.
func (m *MemoryMonitor) SetLimits(softLimit, hardLimit int64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if softLimit > hardLimit {
		softLimit = hardLimit
	}
	m.softLimit = softLimit
	m.hardLimit = hardLimit
}

// SetInterval configures monitoring interval.
func (m *MemoryMonitor) SetInterval(interval time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.interval = interval
}

// Start begins background monitoring.
func (m *MemoryMonitor) Start() {
	go m.monitor()
}

// Stop terminates background monitoring. Safe to call multiple times.
func (m *MemoryMonitor) Stop() {
	m.stopOnce.Do(func() {
		close(m.stopCh)
		<-m.doneCh
	})
}

// monitor runs periodic memory checks.
func (m *MemoryMonitor) monitor() {
	defer close(m.doneCh)

	ticker := time.NewTicker(m.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			m.checkMemory()
		case <-m.stopCh:
			return
		}
	}
}

// checkMemory verifies memory usage and triggers cleanup if needed.
func (m *MemoryMonitor) checkMemory() {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Get cache size
	cacheSize := m.cache.Size()

	// Update stats
	m.stats.CurrentUsage = cacheSize
	if cacheSize > m.stats.PeakUsage {
		m.stats.PeakUsage = cacheSize
	}

	// Get system memory
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	m.stats.SystemMemoryMB = memStats.Alloc / 1024 / 1024

	// Check limits
	if cacheSize > m.hardLimit {
		// Hard limit exceeded - force eviction to soft limit
		m.forceEviction()
	} else if cacheSize > m.softLimit {
		// Soft limit exceeded - gentle cleanup
		m.softCleanup()
	}
}

// forceEviction aggressively evicts entries to meet hard limit.
func (m *MemoryMonitor) forceEviction() {
	targetSize := m.softLimit
	m.cache.SetMaxSize(targetSize)

	m.stats.EvictionCount++
	m.stats.LastCleanupAt = time.Now() // time.Now() is acceptable for non-deterministic performance monitoring.
}

// softCleanup gradually reduces cache to soft limit.
func (m *MemoryMonitor) softCleanup() {
	// Reduce max size slightly to trigger LRU eviction
	targetSize := m.softLimit * 95 / 100 // 95% of soft limit

	currentMax := m.cache.MaxSize()
	if targetSize < currentMax {
		m.cache.SetMaxSize(targetSize)
	}

	m.stats.CleanupCount++
	m.stats.LastCleanupAt = time.Now() // time.Now() is acceptable for non-deterministic performance monitoring.
}

// Stats returns current monitoring statistics.
func (m *MemoryMonitor) Stats() MemoryStats {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.stats
}

// IsHealthy returns true if cache is within limits.
func (m *MemoryMonitor) IsHealthy() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.stats.CurrentUsage <= m.hardLimit
}

// UsagePercentage returns cache usage as percentage of hard limit.
func (m *MemoryMonitor) UsagePercentage() float64 {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.hardLimit == 0 {
		return 0.0
	}

	return float64(m.stats.CurrentUsage) / float64(m.hardLimit) * 100.0
}

// EstimatedSpriteCapacity estimates max 64x64 sprites at current usage.
// 64x64 RGBA = 16KB per sprite.
func (m *MemoryMonitor) EstimatedSpriteCapacity() int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	const sprite64Size = 64 * 64 * 4 // 16KB
	return int(m.hardLimit / sprite64Size)
}
