// Package engine provides the core game engine functionality.
package engine

import (
	"sort"
	"time"
)

// FrameTimeTracker tracks frame times to detect performance issues and stuttering.
// It maintains a rolling window of frame durations and calculates statistics
// including percentiles to identify frame time variance (jank).
type FrameTimeTracker struct {
	frameTimes []time.Duration
	maxSamples int
	index      int
}

// NewFrameTimeTracker creates a new frame time tracker with the specified sample window size.
// maxSamples determines how many frames to track (e.g., 1000 frames = ~16 seconds at 60 FPS).
func NewFrameTimeTracker(maxSamples int) *FrameTimeTracker {
	return &FrameTimeTracker{
		frameTimes: make([]time.Duration, 0, maxSamples),
		maxSamples: maxSamples,
		index:      0,
	}
}

// RecordFrame records the duration of a single frame.
// This should be called at the end of each frame's Update() method.
func (f *FrameTimeTracker) RecordFrame(duration time.Duration) {
	if len(f.frameTimes) < f.maxSamples {
		f.frameTimes = append(f.frameTimes, duration)
	} else {
		f.frameTimes[f.index] = duration
		f.index = (f.index + 1) % f.maxSamples
	}
}

// GetStats calculates comprehensive frame time statistics including percentiles.
// Returns empty stats if no frames have been recorded.
func (f *FrameTimeTracker) GetStats() FrameTimeStats {
	if len(f.frameTimes) == 0 {
		return FrameTimeStats{}
	}

	// Copy and sort for percentile calculation
	sorted := make([]time.Duration, len(f.frameTimes))
	copy(sorted, f.frameTimes)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i] < sorted[j] })

	// Calculate average
	var total time.Duration
	for _, ft := range sorted {
		total += ft
	}

	count := len(sorted)
	avg := total / time.Duration(count)

	// Calculate standard deviation
	var variance float64
	for _, ft := range sorted {
		diff := float64(ft - avg)
		variance += diff * diff
	}
	stdDev := time.Duration(variance / float64(count))

	// Calculate percentiles
	// For 1% low, we want the 1st percentile (worst 1% of frames)
	// This should be a HIGH value (slow frames), not low
	idx1Pct := int(float64(count) * 0.99) // 99th percentile index
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
		Min:           sorted[0],
		Max:           sorted[count-1],
		Percentile1:   sorted[idx1Pct],   // 99th percentile (1% worst frames)
		Percentile01:  sorted[count-1],   // 0.1% low (worst frame)
		Percentile99:  sorted[idx99Pct],  // 99th percentile
		Percentile999: sorted[idx999Pct], // 99.9th percentile
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
	return 1000.0 / float64(s.Average.Milliseconds())
}

// GetWorstFPS returns the FPS of the worst 1% of frames (99th percentile).
// This is more indicative of perceived performance than average FPS.
// Lower values indicate worse stuttering.
func (s FrameTimeStats) GetWorstFPS() float64 {
	if s.Percentile1 == 0 {
		return 0
	}
	return 1000.0 / float64(s.Percentile1.Milliseconds())
}
