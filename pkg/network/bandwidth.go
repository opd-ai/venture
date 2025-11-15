package network

import (
	"sync"
	"time"
)

// BandwidthMonitor tracks network bandwidth usage.
type BandwidthMonitor struct {
	mu sync.RWMutex

	// Per-player bandwidth tracking
	playerStats map[uint64]*PlayerBandwidthStats

	// Global bandwidth stats
	totalBytesSent     uint64
	totalBytesReceived uint64

	// Sampling window
	windowStart time.Time
	windowSize  time.Duration
}

// PlayerBandwidthStats tracks bandwidth for a single player.
type PlayerBandwidthStats struct {
	PlayerID      uint64
	BytesSent     uint64
	BytesReceived uint64
	MessagesSent  uint64
	MessagesRecv  uint64
	LastActivity  time.Time
	WindowStart   time.Time
	CurrentRate   float64 // Bytes per second
	PeakRate      float64 // Peak bytes per second observed
}

// NewBandwidthMonitor creates a new bandwidth monitor.
func NewBandwidthMonitor(windowSize time.Duration) *BandwidthMonitor {
	return &BandwidthMonitor{
		playerStats: make(map[uint64]*PlayerBandwidthStats),
		windowStart: time.Now(),
		windowSize:  windowSize,
	}
}

// RecordSent records bytes sent for a player.
func (bm *BandwidthMonitor) RecordSent(playerID uint64, bytes uint64) {
	bm.mu.Lock()
	defer bm.mu.Unlock()

	stats := bm.getOrCreateStats(playerID)
	stats.BytesSent += bytes
	stats.MessagesSent++
	stats.LastActivity = time.Now()

	bm.totalBytesSent += bytes
}

// RecordReceived records bytes received for a player.
func (bm *BandwidthMonitor) RecordReceived(playerID uint64, bytes uint64) {
	bm.mu.Lock()
	defer bm.mu.Unlock()

	stats := bm.getOrCreateStats(playerID)
	stats.BytesReceived += bytes
	stats.MessagesRecv++
	stats.LastActivity = time.Now()

	bm.totalBytesReceived += bytes
}

// GetPlayerStats returns bandwidth stats for a player.
func (bm *BandwidthMonitor) GetPlayerStats(playerID uint64) *PlayerBandwidthStats {
	bm.mu.RLock()
	defer bm.mu.RUnlock()

	stats, exists := bm.playerStats[playerID]
	if !exists {
		return &PlayerBandwidthStats{PlayerID: playerID}
	}

	// Return a copy to avoid race conditions
	statsCopy := *stats
	return &statsCopy
}

// GetGlobalStats returns global bandwidth statistics.
func (bm *BandwidthMonitor) GetGlobalStats() (sent, received uint64, duration time.Duration) {
	bm.mu.RLock()
	defer bm.mu.RUnlock()

	duration = time.Since(bm.windowStart)
	return bm.totalBytesSent, bm.totalBytesReceived, duration
}

// UpdateRates updates current bandwidth rates for all players.
func (bm *BandwidthMonitor) UpdateRates() {
	bm.mu.Lock()
	defer bm.mu.Unlock()

	now := time.Now()
	windowElapsed := now.Sub(bm.windowStart).Seconds()

	for _, stats := range bm.playerStats {
		playerWindowElapsed := now.Sub(stats.WindowStart).Seconds()
		if playerWindowElapsed > 0 {
			totalBytes := stats.BytesSent + stats.BytesReceived
			stats.CurrentRate = float64(totalBytes) / playerWindowElapsed

			if stats.CurrentRate > stats.PeakRate {
				stats.PeakRate = stats.CurrentRate
			}
		}
	}

	// Reset window if elapsed
	if windowElapsed >= bm.windowSize.Seconds() {
		bm.resetWindow()
	}
}

// ResetPlayerStats resets stats for a player.
func (bm *BandwidthMonitor) ResetPlayerStats(playerID uint64) {
	bm.mu.Lock()
	defer bm.mu.Unlock()

	if stats, exists := bm.playerStats[playerID]; exists {
		stats.BytesSent = 0
		stats.BytesReceived = 0
		stats.MessagesSent = 0
		stats.MessagesRecv = 0
		stats.WindowStart = time.Now()
		stats.CurrentRate = 0
	}
}

// GetAllPlayerStats returns stats for all players.
func (bm *BandwidthMonitor) GetAllPlayerStats() map[uint64]*PlayerBandwidthStats {
	bm.mu.RLock()
	defer bm.mu.RUnlock()

	result := make(map[uint64]*PlayerBandwidthStats)
	for id, stats := range bm.playerStats {
		statsCopy := *stats
		result[id] = &statsCopy
	}

	return result
}

// Private methods

func (bm *BandwidthMonitor) getOrCreateStats(playerID uint64) *PlayerBandwidthStats {
	stats, exists := bm.playerStats[playerID]
	if !exists {
		stats = &PlayerBandwidthStats{
			PlayerID:    playerID,
			WindowStart: time.Now(),
		}
		bm.playerStats[playerID] = stats
	}
	return stats
}

func (bm *BandwidthMonitor) resetWindow() {
	// Reset global counters
	bm.totalBytesSent = 0
	bm.totalBytesReceived = 0
	bm.windowStart = time.Now()

	// Reset player windows
	for _, stats := range bm.playerStats {
		stats.BytesSent = 0
		stats.BytesReceived = 0
		stats.MessagesSent = 0
		stats.MessagesRecv = 0
		stats.WindowStart = time.Now()
		stats.CurrentRate = 0
	}
}

// CalculateBandwidthKBps calculates bandwidth in KB/s.
func CalculateBandwidthKBps(bytes uint64, duration time.Duration) float64 {
	seconds := duration.Seconds()
	if seconds == 0 {
		return 0
	}
	bytesPerSecond := float64(bytes) / seconds
	return bytesPerSecond / 1024.0
}

// CalculateBandwidthMBps calculates bandwidth in MB/s.
func CalculateBandwidthMBps(bytes uint64, duration time.Duration) float64 {
	return CalculateBandwidthKBps(bytes, duration) / 1024.0
}
