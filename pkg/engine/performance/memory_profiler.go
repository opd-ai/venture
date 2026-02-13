package performance

import (
	"runtime"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

// MemoryProfiler tracks memory allocations across V9 systems
type MemoryProfiler struct {
	mu           sync.RWMutex
	allocations  map[string]uint64
	totalBytes   uint64
	startTime    time.Time
	snapshots    []*MemorySnapshot
	maxSnapshots int
}

// MemorySnapshot represents memory state at a point in time
type MemorySnapshot struct {
	Timestamp   time.Time
	TotalMB     uint64
	Allocations map[string]uint64
	HeapAlloc   uint64
	HeapInuse   uint64
	StackInuse  uint64
}

// NewMemoryProfiler creates a new memory profiler
func NewMemoryProfiler() *MemoryProfiler {
	return &MemoryProfiler{
		allocations:  make(map[string]uint64),
		startTime:    time.Now(),
		snapshots:    make([]*MemorySnapshot, 0),
		maxSnapshots: 100, // Keep last 100 snapshots
	}
}

// TrackAllocation records a memory allocation
func (mp *MemoryProfiler) TrackAllocation(name string, bytes uint64) {
	mp.mu.Lock()
	defer mp.mu.Unlock()

	mp.allocations[name] += bytes
	mp.totalBytes += bytes
}

// ReleaseAllocation records memory release
func (mp *MemoryProfiler) ReleaseAllocation(name string, bytes uint64) {
	mp.mu.Lock()
	defer mp.mu.Unlock()

	if current, exists := mp.allocations[name]; exists {
		if bytes > current {
			// Log underflow for debugging - requested release exceeds tracked allocation
			log.WithFields(log.Fields{
				"allocation_name": name,
				"requested_bytes": bytes,
				"current_bytes":   current,
			}).Warn("memory release underflow, clamping to current allocation")
			bytes = current
		}
		mp.allocations[name] -= bytes
		mp.totalBytes -= bytes
	}
}

// GetStats returns current memory statistics
func (mp *MemoryProfiler) GetStats() *MemoryStats {
	mp.mu.RLock()
	defer mp.mu.RUnlock()

	stats := &MemoryStats{
		TotalBytes:  mp.totalBytes,
		TotalMB:     mp.totalBytes / (1024 * 1024),
		Allocations: make(map[string]uint64),
	}

	// Copy allocations and find largest
	var largest uint64
	for name, bytes := range mp.allocations {
		stats.Allocations[name] = bytes
		if bytes > largest {
			largest = bytes
			stats.LargestAlloc = name
			stats.LargestAllocMB = bytes / (1024 * 1024)
		}
	}

	return stats
}

// TakeSnapshot captures current memory state
func (mp *MemoryProfiler) TakeSnapshot() *MemorySnapshot {
	mp.mu.Lock()
	defer mp.mu.Unlock()

	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	snapshot := &MemorySnapshot{
		Timestamp:   time.Now(),
		TotalMB:     mp.totalBytes / (1024 * 1024),
		Allocations: make(map[string]uint64),
		HeapAlloc:   memStats.HeapAlloc / (1024 * 1024),
		HeapInuse:   memStats.HeapInuse / (1024 * 1024),
		StackInuse:  memStats.StackInuse / (1024 * 1024),
	}

	for name, bytes := range mp.allocations {
		snapshot.Allocations[name] = bytes
	}

	// Add snapshot and maintain limit
	mp.snapshots = append(mp.snapshots, snapshot)
	if len(mp.snapshots) > mp.maxSnapshots {
		mp.snapshots = mp.snapshots[1:]
	}

	return snapshot
}

// GetSnapshots returns all recorded snapshots
func (mp *MemoryProfiler) GetSnapshots() []*MemorySnapshot {
	mp.mu.RLock()
	defer mp.mu.RUnlock()

	result := make([]*MemorySnapshot, len(mp.snapshots))
	copy(result, mp.snapshots)
	return result
}

// GetAllocationTrend returns memory growth for a specific allocation
func (mp *MemoryProfiler) GetAllocationTrend(name string, samples int) []uint64 {
	mp.mu.RLock()
	defer mp.mu.RUnlock()

	if samples > len(mp.snapshots) {
		samples = len(mp.snapshots)
	}

	start := len(mp.snapshots) - samples
	result := make([]uint64, samples)

	for i := 0; i < samples; i++ {
		snapshot := mp.snapshots[start+i]
		if bytes, exists := snapshot.Allocations[name]; exists {
			result[i] = bytes
		}
	}

	return result
}

// IdentifyLeaks detects memory allocations that never decrease
func (mp *MemoryProfiler) IdentifyLeaks(minGrowthMB uint64) []string {
	mp.mu.RLock()
	defer mp.mu.RUnlock()

	if len(mp.snapshots) < 10 {
		return nil // Need more data
	}

	leaks := make([]string, 0)

	// Compare first and last snapshots
	first := mp.snapshots[0]
	last := mp.snapshots[len(mp.snapshots)-1]

	for name := range last.Allocations {
		firstBytes, _ := first.Allocations[name]
		lastBytes := last.Allocations[name]

		growthBytes := lastBytes - firstBytes
		growthMB := growthBytes / (1024 * 1024)

		if growthMB >= minGrowthMB {
			// Check if allocation ever decreased
			neverDecreased := true
			prevBytes := firstBytes

			for _, snapshot := range mp.snapshots[1:] {
				currentBytes, exists := snapshot.Allocations[name]
				if !exists {
					continue
				}
				if currentBytes < prevBytes {
					neverDecreased = false
					break
				}
				prevBytes = currentBytes
			}

			if neverDecreased {
				leaks = append(leaks, name)
			}
		}
	}

	return leaks
}

// Reset clears all tracked allocations
func (mp *MemoryProfiler) Reset() {
	mp.mu.Lock()
	defer mp.mu.Unlock()

	mp.allocations = make(map[string]uint64)
	mp.totalBytes = 0
	mp.snapshots = make([]*MemorySnapshot, 0)
	mp.startTime = time.Now()
}

// GetUptime returns profiler uptime duration
func (mp *MemoryProfiler) GetUptime() time.Duration {
	mp.mu.RLock()
	defer mp.mu.RUnlock()
	return time.Since(mp.startTime)
}
