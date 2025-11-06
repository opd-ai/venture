// Package network provides buffer monitoring for network channels.
// This file implements BufferStats for tracking channel utilization and detecting congestion.
package network

import (
	"sync"
	"sync/atomic"

	"github.com/sirupsen/logrus"
)

// BufferStats tracks statistics for a buffered channel.
// It monitors current utilization, total operations, and drops to help detect network congestion.
type BufferStats struct {
	Name     string // Channel name for logging
	Capacity int    // Maximum channel capacity

	// Atomic counters for thread-safe updates
	sent    uint64 // Total messages sent to channel
	dropped uint64 // Messages dropped due to full channel
	
	// Current size tracking (requires lock)
	currentSize int
	mu          sync.RWMutex
	
	// Logger for warnings
	logger *logrus.Entry
	
	// Warning threshold tracking
	warnThreshold float64 // Percentage (0.0-1.0) to trigger warnings
	lastWarnSize  int     // Last size that triggered a warning (prevents spam)
}

// NewBufferStats creates a new buffer statistics tracker.
// name: descriptive name for logging (e.g., "client_state_updates")
// capacity: maximum channel capacity
// logger: optional logger for warnings (can be nil)
func NewBufferStats(name string, capacity int, logger *logrus.Entry) *BufferStats {
	return &BufferStats{
		Name:          name,
		Capacity:      capacity,
		warnThreshold: 0.8, // 80% utilization triggers warnings
		logger:        logger,
	}
}

// RecordSend records a successful send to the channel.
// Call this after successfully sending to a channel.
func (bs *BufferStats) RecordSend() {
	atomic.AddUint64(&bs.sent, 1)
	
	bs.mu.Lock()
	bs.currentSize++
	size := bs.currentSize
	bs.mu.Unlock()
	
	// Check if we should warn about high utilization
	bs.checkWarnThreshold(size)
}

// RecordReceive records a successful receive from the channel.
// Call this after successfully receiving from a channel.
func (bs *BufferStats) RecordReceive() {
	bs.mu.Lock()
	if bs.currentSize > 0 {
		bs.currentSize--
	}
	bs.mu.Unlock()
}

// RecordDrop records a message drop due to full channel.
// Call this when a non-blocking send fails due to full channel.
func (bs *BufferStats) RecordDrop() {
	atomic.AddUint64(&bs.dropped, 1)
	
	// Log drop event
	if bs.logger != nil {
		bs.logger.WithFields(logrus.Fields{
			"buffer":       bs.Name,
			"utilization":  bs.Utilization(),
			"total_drops":  bs.GetDropped(),
			"total_sent":   bs.GetSent(),
		}).Warn("buffer full, message dropped")
	}
}

// checkWarnThreshold checks if utilization exceeds warning threshold and logs if needed.
func (bs *BufferStats) checkWarnThreshold(currentSize int) {
	if bs.logger == nil {
		return
	}
	
	utilization := float64(currentSize) / float64(bs.Capacity)
	
	// Only warn if:
	// 1. Utilization exceeds threshold
	// 2. Size has increased since last warning (prevents spam)
	if utilization >= bs.warnThreshold && currentSize > bs.lastWarnSize {
		bs.lastWarnSize = currentSize
		bs.logger.WithFields(logrus.Fields{
			"buffer":       bs.Name,
			"size":         currentSize,
			"capacity":     bs.Capacity,
			"utilization":  utilization,
			"total_sent":   bs.GetSent(),
			"total_drops":  bs.GetDropped(),
		}).Warn("buffer utilization high")
	}
	
	// Reset lastWarnSize if utilization drops below threshold
	if utilization < bs.warnThreshold {
		bs.lastWarnSize = 0
	}
}

// GetSent returns total messages sent to the channel.
func (bs *BufferStats) GetSent() uint64 {
	return atomic.LoadUint64(&bs.sent)
}

// GetDropped returns total messages dropped due to full channel.
func (bs *BufferStats) GetDropped() uint64 {
	return atomic.LoadUint64(&bs.dropped)
}

// GetCurrentSize returns current estimated channel size.
// This is an approximation due to concurrent access.
func (bs *BufferStats) GetCurrentSize() int {
	bs.mu.RLock()
	defer bs.mu.RUnlock()
	return bs.currentSize
}

// Utilization returns current buffer utilization as a percentage (0.0-1.0).
func (bs *BufferStats) Utilization() float64 {
	if bs.Capacity == 0 {
		return 0.0
	}
	size := bs.GetCurrentSize()
	return float64(size) / float64(bs.Capacity)
}

// DropRate returns the percentage of messages dropped (0.0-1.0).
// Returns 0.0 if no messages have been sent or dropped.
func (bs *BufferStats) DropRate() float64 {
	sent := bs.GetSent()
	dropped := bs.GetDropped()
	total := sent + dropped
	if total == 0 {
		return 0.0
	}
	return float64(dropped) / float64(total)
}

// Reset resets all statistics to zero.
// Useful for testing or periodic metric collection.
func (bs *BufferStats) Reset() {
	atomic.StoreUint64(&bs.sent, 0)
	atomic.StoreUint64(&bs.dropped, 0)
	bs.mu.Lock()
	bs.currentSize = 0
	bs.lastWarnSize = 0
	bs.mu.Unlock()
}

// Snapshot returns a point-in-time copy of the statistics.
// Useful for metrics collection without holding locks.
type BufferSnapshot struct {
	Name        string
	Capacity    int
	CurrentSize int
	Sent        uint64
	Dropped     uint64
	Utilization float64
	DropRate    float64
}

// Snapshot returns a snapshot of current buffer statistics.
func (bs *BufferStats) Snapshot() BufferSnapshot {
	return BufferSnapshot{
		Name:        bs.Name,
		Capacity:    bs.Capacity,
		CurrentSize: bs.GetCurrentSize(),
		Sent:        bs.GetSent(),
		Dropped:     bs.GetDropped(),
		Utilization: bs.Utilization(),
		DropRate:    bs.DropRate(),
	}
}
