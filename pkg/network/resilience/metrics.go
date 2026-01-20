// Package resilience metrics implements network performance metrics collection
// including latency tracking, bandwidth monitoring, packet statistics, and
// gameplay metrics (mispredictions, desyncs, reconnections).
package resilience

import (
	"math"
	"sort"
	"sync"
	"time"
)

// MetricsCollector collects and aggregates network performance metrics.
type MetricsCollector struct {
	mu sync.RWMutex

	startTime time.Time

	// Latency samples (for percentile calculation)
	latencySamples []time.Duration
	maxSamples     int

	// Packet counters
	packetsSent     uint64
	packetsReceived uint64
	packetsDropped  uint64

	// Bandwidth tracking
	bytesSent     uint64
	bytesReceived uint64

	// Bandwidth samples (bytes per second, sampled every second)
	bandwidthSamples   []float64
	lastBandwidthCheck time.Time
	bytesThisSecond    uint64

	// Gameplay metrics
	mispredictions uint64
	predictions    uint64
	desyncs        uint64
	reconnects     uint64
	reconnectTimes []time.Duration
}

// NewMetricsCollector creates a new metrics collector.
func NewMetricsCollector() *MetricsCollector {
	return &MetricsCollector{
		startTime:          time.Now(),
		latencySamples:     make([]time.Duration, 0, 10000),
		maxSamples:         10000,
		bandwidthSamples:   make([]float64, 0, 1000),
		reconnectTimes:     make([]time.Duration, 0, 100),
		lastBandwidthCheck: time.Now(),
	}
}

// RecordLatency records a latency measurement.
func (mc *MetricsCollector) RecordLatency(latency time.Duration) {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	// Keep max samples using sliding window
	if len(mc.latencySamples) >= mc.maxSamples {
		mc.latencySamples = mc.latencySamples[1:]
	}
	mc.latencySamples = append(mc.latencySamples, latency)
}

// RecordPacketSent increments the sent packet counter.
func (mc *MetricsCollector) RecordPacketSent(bytes int) {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	mc.packetsSent++
	mc.bytesSent += uint64(bytes)
	mc.bytesThisSecond += uint64(bytes)
	mc.updateBandwidthSample()
}

// RecordPacketReceived increments the received packet counter.
func (mc *MetricsCollector) RecordPacketReceived(bytes int) {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	mc.packetsReceived++
	mc.bytesReceived += uint64(bytes)
}

// RecordPacketLoss increments the dropped packet counter.
func (mc *MetricsCollector) RecordPacketLoss() {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	mc.packetsDropped++
}

// RecordPrediction records a client-side prediction event.
func (mc *MetricsCollector) RecordPrediction(mispredicted bool) {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	mc.predictions++
	if mispredicted {
		mc.mispredictions++
	}
}

// RecordDesync records a client-server desynchronization event.
func (mc *MetricsCollector) RecordDesync() {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	mc.desyncs++
}

// RecordReconnect records a reconnection event.
func (mc *MetricsCollector) RecordReconnect(duration time.Duration) {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	mc.reconnects++
	mc.reconnectTimes = append(mc.reconnectTimes, duration)
}

// updateBandwidthSample updates the bandwidth sample if a second has passed.
func (mc *MetricsCollector) updateBandwidthSample() {
	now := time.Now()
	if now.Sub(mc.lastBandwidthCheck) >= time.Second {
		// Record bytes per second
		bps := float64(mc.bytesThisSecond)
		mc.bandwidthSamples = append(mc.bandwidthSamples, bps)

		// Reset for next second
		mc.bytesThisSecond = 0
		mc.lastBandwidthCheck = now
	}
}

// GetStats calculates and returns current statistics.
func (mc *MetricsCollector) GetStats() NetworkStats {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	stats := NetworkStats{
		StartTime:          mc.startTime,
		EndTime:            time.Now(),
		PacketsSent:        mc.packetsSent,
		PacketsReceived:    mc.packetsReceived,
		PacketsDropped:     mc.packetsDropped,
		BytesSent:          mc.bytesSent,
		BytesReceived:      mc.bytesReceived,
		MispredictionCount: mc.mispredictions,
		DesyncCount:        mc.desyncs,
		ReconnectCount:     mc.reconnects,
	}

	stats.Duration = stats.EndTime.Sub(stats.StartTime)

	// Calculate packet loss rate
	totalPackets := mc.packetsSent + mc.packetsDropped
	if totalPackets > 0 {
		stats.PacketLossRate = float64(mc.packetsDropped) / float64(totalPackets)
	}

	// Calculate misprediction rate
	if mc.predictions > 0 {
		stats.MispredictionRate = float64(mc.mispredictions) / float64(mc.predictions)
	}

	// Calculate latency statistics
	if len(mc.latencySamples) > 0 {
		stats.AvgLatency = mc.calculateAvgLatency()
		stats.MinLatency = mc.calculateMinLatency()
		stats.MaxLatency = mc.calculateMaxLatency()
		stats.P95Latency = mc.calculatePercentile(0.95)
		stats.P99Latency = mc.calculatePercentile(0.99)
	}

	// Calculate bandwidth statistics
	if len(mc.bandwidthSamples) > 0 {
		stats.AvgBandwidth = mc.calculateAvgBandwidth()
		stats.PeakBandwidth = mc.calculatePeakBandwidth()
	}

	// Calculate average reconnect time
	if len(mc.reconnectTimes) > 0 {
		stats.AvgReconnectTime = mc.calculateAvgReconnectTime()
	}

	return stats
}

// calculateAvgLatency calculates average latency.
func (mc *MetricsCollector) calculateAvgLatency() time.Duration {
	if len(mc.latencySamples) == 0 {
		return 0
	}

	var sum int64
	for _, lat := range mc.latencySamples {
		sum += int64(lat)
	}

	return time.Duration(sum / int64(len(mc.latencySamples)))
}

// calculateMinLatency finds minimum latency.
func (mc *MetricsCollector) calculateMinLatency() time.Duration {
	if len(mc.latencySamples) == 0 {
		return 0
	}

	min := mc.latencySamples[0]
	for _, lat := range mc.latencySamples[1:] {
		if lat < min {
			min = lat
		}
	}

	return min
}

// calculateMaxLatency finds maximum latency.
func (mc *MetricsCollector) calculateMaxLatency() time.Duration {
	if len(mc.latencySamples) == 0 {
		return 0
	}

	max := mc.latencySamples[0]
	for _, lat := range mc.latencySamples[1:] {
		if lat > max {
			max = lat
		}
	}

	return max
}

// calculatePercentile calculates the Nth percentile latency.
func (mc *MetricsCollector) calculatePercentile(percentile float64) time.Duration {
	if len(mc.latencySamples) == 0 {
		return 0
	}

	// Copy and sort samples
	sorted := make([]time.Duration, len(mc.latencySamples))
	copy(sorted, mc.latencySamples)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i] < sorted[j]
	})

	// Calculate index
	index := int(math.Ceil(percentile*float64(len(sorted)))) - 1
	if index < 0 {
		index = 0
	}
	if index >= len(sorted) {
		index = len(sorted) - 1
	}

	return sorted[index]
}

// calculateAvgBandwidth calculates average bandwidth.
func (mc *MetricsCollector) calculateAvgBandwidth() float64 {
	if len(mc.bandwidthSamples) == 0 {
		return 0
	}

	var sum float64
	for _, bw := range mc.bandwidthSamples {
		sum += bw
	}

	return sum / float64(len(mc.bandwidthSamples))
}

// calculatePeakBandwidth finds peak bandwidth.
func (mc *MetricsCollector) calculatePeakBandwidth() float64 {
	if len(mc.bandwidthSamples) == 0 {
		return 0
	}

	peak := mc.bandwidthSamples[0]
	for _, bw := range mc.bandwidthSamples[1:] {
		if bw > peak {
			peak = bw
		}
	}

	return peak
}

// calculateAvgReconnectTime calculates average reconnection time.
func (mc *MetricsCollector) calculateAvgReconnectTime() time.Duration {
	if len(mc.reconnectTimes) == 0 {
		return 0
	}

	var sum int64
	for _, t := range mc.reconnectTimes {
		sum += int64(t)
	}

	return time.Duration(sum / int64(len(mc.reconnectTimes)))
}

// Reset clears all collected metrics.
func (mc *MetricsCollector) Reset() {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	mc.startTime = time.Now()
	mc.latencySamples = make([]time.Duration, 0, mc.maxSamples)
	mc.packetsSent = 0
	mc.packetsReceived = 0
	mc.packetsDropped = 0
	mc.bytesSent = 0
	mc.bytesReceived = 0
	mc.bandwidthSamples = make([]float64, 0, 1000)
	mc.lastBandwidthCheck = time.Now()
	mc.bytesThisSecond = 0
	mc.mispredictions = 0
	mc.predictions = 0
	mc.desyncs = 0
	mc.reconnects = 0
	mc.reconnectTimes = make([]time.Duration, 0, 100)
}
