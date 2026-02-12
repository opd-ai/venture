// Package engine provides the core game engine functionality.
package engine

import (
	"sort"
	"sync"
	"time"
)

// FrameTimeTracker tracks frame times to detect performance issues and stuttering.
// It maintains a rolling window of frame durations and calculates statistics
// including percentiles to identify frame time variance (jank).
// Incrementally tracks min/max/sum to avoid expensive full-window recalculation.
// Thread-safe: protected by a mutex for concurrent RecordFrame/GetStats access.
type FrameTimeTracker struct {
	mu         sync.Mutex
	frameTimes []time.Duration
	maxSamples int
	index      int

	// Incrementally maintained running statistics to avoid per-call allocations.
	runningSum time.Duration // Sum of all samples in the window
	runningMin time.Duration // Min frame time in window (approximate, refreshed periodically)
	runningMax time.Duration // Max frame time in window

	// Cached sorted copy reused across GetStats calls to reduce allocations.
	// Only re-sorted when dirty (new samples recorded since last GetStats).
	sortedCache []time.Duration
	sortDirty   bool
}

// NewFrameTimeTracker creates a new frame time tracker with the specified sample window size.
// maxSamples determines how many frames to track (e.g., 1000 frames = ~16 seconds at 60 FPS).
func NewFrameTimeTracker(maxSamples int) *FrameTimeTracker {
	return &FrameTimeTracker{
		frameTimes:  make([]time.Duration, 0, maxSamples),
		maxSamples:  maxSamples,
		index:       0,
		runningMin:  time.Hour,
		sortedCache: make([]time.Duration, 0, maxSamples),
		sortDirty:   true,
	}
}

// RecordFrame records the duration of a single frame.
// This should be called at the end of each frame's Update() method.
// Thread-safe: can be called concurrently with GetStats.
func (f *FrameTimeTracker) RecordFrame(duration time.Duration) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if len(f.frameTimes) < f.maxSamples {
		f.frameTimes = append(f.frameTimes, duration)
		f.runningSum += duration
	} else {
		// Subtract the old sample being overwritten, add the new one
		f.runningSum -= f.frameTimes[f.index]
		f.runningSum += duration
		f.frameTimes[f.index] = duration
		f.index = (f.index + 1) % f.maxSamples
	}

	// Update running max (always accurate)
	if duration > f.runningMax {
		f.runningMax = duration
	}
	// Update running min (always accurate for new minimums)
	if duration < f.runningMin {
		f.runningMin = duration
	}

	f.sortDirty = true
}

// GetStats calculates comprehensive frame time statistics including percentiles.
// Returns empty stats if no frames have been recorded.
// Reuses a cached sorted buffer to minimize allocations and only re-sorts when dirty.
// Thread-safe: can be called concurrently with RecordFrame.
func (f *FrameTimeTracker) GetStats() FrameTimeStats {
	f.mu.Lock()
	defer f.mu.Unlock()

	if len(f.frameTimes) == 0 {
		return FrameTimeStats{}
	}

	count := len(f.frameTimes)

	// Reuse sorted cache buffer — only copy and sort if new frames have been recorded
	if f.sortDirty {
		if cap(f.sortedCache) < count {
			f.sortedCache = make([]time.Duration, count)
		}
		f.sortedCache = f.sortedCache[:count]
		copy(f.sortedCache, f.frameTimes)
		sort.Slice(f.sortedCache, func(i, j int) bool { return f.sortedCache[i] < f.sortedCache[j] })
		// Refresh exact min/max from sorted data
		f.runningMin = f.sortedCache[0]
		f.runningMax = f.sortedCache[count-1]
		f.sortDirty = false
	}

	sorted := f.sortedCache

	// Use incrementally maintained sum for O(1) average
	avg := f.runningSum / time.Duration(count)

	// Calculate standard deviation
	var variance float64
	for _, ft := range sorted {
		diff := float64(ft - avg)
		variance += diff * diff
	}
	stdDev := time.Duration(variance / float64(count))

	// Calculate percentiles
	idx1Pct := int(float64(count) * 0.99)
	if idx1Pct >= count {
		idx1Pct = count - 1
	}
	idx99Pct := idx1Pct
	idx999Pct := int(float64(count) * 0.999)
	if idx999Pct >= count {
		idx999Pct = count - 1
	}

	return FrameTimeStats{
		Average:       avg,
		Min:           f.runningMin,
		Max:           f.runningMax,
		Percentile1:   sorted[idx1Pct],
		Percentile01:  sorted[count-1],
		Percentile99:  sorted[idx99Pct],
		Percentile999: sorted[idx999Pct],
		StdDev:        stdDev,
		SampleCount:   count,
	}
}

// FrameTimeStats contains comprehensive frame time statistics.
// The Percentile1 represents the 99th percentile (1% worst/slowest frames).
// Higher percentile values indicate worse performance (slower frames).
type FrameTimeStats struct {
	Average       time.Duration // Average frame time
	Min           time.Duration // Fastest frame
	Max           time.Duration // Slowest frame
	Percentile1   time.Duration // 99th percentile (1% worst frames - should be <20ms for smooth gameplay)
	Percentile01  time.Duration // Worst frame (0.1% low)
	Percentile99  time.Duration // 99th percentile
	Percentile999 time.Duration // 99.9th percentile
	StdDev        time.Duration // Standard deviation (measure of consistency)
	SampleCount   int           // Number of samples used
}

// IsStuttering returns true if frame time variance indicates perceptible stuttering.
// Target: 60 FPS = 16.67ms per frame. Stuttering occurs if 1% worst frames exceed target significantly.
func (s FrameTimeStats) IsStuttering() bool {
	// Target: 60 FPS = 16.67ms per frame
	// Stuttering if 1% worst frames are significantly above target (>20ms)
	targetFrameTime := 20 * time.Millisecond
	return s.Percentile1 > targetFrameTime
}

// GetFPS returns the average FPS based on the average frame time.
func (s FrameTimeStats) GetFPS() float64 {
	if s.Average == 0 {
		return 0
	}
	return float64(time.Second) / float64(s.Average)
}

// GetWorstFPS returns the FPS of the worst 1% of frames (99th percentile).
// This is more indicative of perceived performance than average FPS.
// Lower values indicate worse stuttering.
func (s FrameTimeStats) GetWorstFPS() float64 {
	if s.Percentile1 == 0 {
		return 0
	}
	return float64(time.Second) / float64(s.Percentile1)
}
