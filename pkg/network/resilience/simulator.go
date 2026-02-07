// Package resilience simulator implements network impairment simulation
// for testing including latency injection, packet loss, jitter simulation,
// and bandwidth limiting with delayed packet delivery queues.
package resilience

import (
	"errors"
	"math/rand"
	"sync"
	"time"
)

var (
	// ErrPacketDropped indicates a packet was dropped due to packet loss simulation
	ErrPacketDropped = errors.New("packet dropped")

	// ErrBandwidthExceeded indicates bandwidth limit was exceeded
	ErrBandwidthExceeded = errors.New("bandwidth limit exceeded")
)

// NetworkSimulator simulates network impairments for testing.
type NetworkSimulator struct {
	mu sync.RWMutex

	config NetworkConfig
	rng    *rand.Rand

	// Delayed packets queue (for latency simulation)
	delayQueue []*delayedPacket
	queueMu    sync.Mutex

	// Bandwidth tracking
	bytesSentLastSecond int
	bandwidthResetTime  time.Time
	bandwidthMu         sync.Mutex

	// Statistics
	packetsSent    uint64
	packetsDropped uint64
	bytesProcessed uint64
}

// delayedPacket represents a packet waiting to be delivered.
type delayedPacket struct {
	packet    *Packet
	deliverAt time.Time
}

// NewNetworkSimulator creates a new network simulator with default config.
// Uses time-based seed for backward compatibility. For deterministic testing,
// use NewNetworkSimulatorWithSeed instead.
func NewNetworkSimulator() *NetworkSimulator {
	return NewNetworkSimulatorWithSeed(time.Now().UnixNano())
}

// NewNetworkSimulatorWithSeed creates a network simulator with deterministic seed.
// Same seed produces same random sequence for reproducible test scenarios.
func NewNetworkSimulatorWithSeed(seed int64) *NetworkSimulator {
	return &NetworkSimulator{
		config: NetworkConfig{
			Latency:        0,
			PacketLossRate: 0,
			Jitter:         0,
			BandwidthLimit: 0,
		},
		rng:                rand.New(rand.NewSource(seed)),
		delayQueue:         make([]*delayedPacket, 0, 1000),
		bandwidthResetTime: time.Now(),
	}
}

// NewNetworkSimulatorWithConfig creates a simulator with the given config.
// Uses time-based seed. For deterministic testing, use NewNetworkSimulatorWithConfigAndSeed.
func NewNetworkSimulatorWithConfig(config NetworkConfig) (*NetworkSimulator, error) {
	return NewNetworkSimulatorWithConfigAndSeed(config, time.Now().UnixNano())
}

// NewNetworkSimulatorWithConfigAndSeed creates a simulator with config and deterministic seed.
func NewNetworkSimulatorWithConfigAndSeed(config NetworkConfig, seed int64) (*NetworkSimulator, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}

	sim := NewNetworkSimulatorWithSeed(seed)
	sim.config = config
	return sim, nil
}

// SetConfig updates the simulator configuration.
func (ns *NetworkSimulator) SetConfig(config NetworkConfig) error {
	if err := config.Validate(); err != nil {
		return err
	}

	ns.mu.Lock()
	ns.config = config
	ns.mu.Unlock()

	return nil
}

// SetLatency sets the simulated one-way latency.
func (ns *NetworkSimulator) SetLatency(latency time.Duration) {
	ns.mu.Lock()
	ns.config.Latency = latency
	ns.mu.Unlock()
}

// SetPacketLoss sets the packet loss rate (0.0-1.0).
func (ns *NetworkSimulator) SetPacketLoss(rate float64) {
	ns.mu.Lock()
	if rate < 0 {
		rate = 0
	}
	if rate > 1 {
		rate = 1
	}
	ns.config.PacketLossRate = rate
	ns.mu.Unlock()
}

// SetJitter sets the latency variance.
func (ns *NetworkSimulator) SetJitter(jitter time.Duration) {
	ns.mu.Lock()
	ns.config.Jitter = jitter
	ns.mu.Unlock()
}

// SetBandwidthLimit sets the maximum bytes per second (0 = unlimited).
func (ns *NetworkSimulator) SetBandwidthLimit(bytesPerSecond int) {
	ns.mu.Lock()
	ns.config.BandwidthLimit = bytesPerSecond
	ns.mu.Unlock()
}

// GetConfig returns the current configuration.
func (ns *NetworkSimulator) GetConfig() NetworkConfig {
	ns.mu.RLock()
	defer ns.mu.RUnlock()
	return ns.config
}

// Send simulates sending a packet through the network.
// The packet may be delayed, dropped, or blocked by bandwidth limits.
func (ns *NetworkSimulator) Send(data []byte) error {
	if data == nil || len(data) == 0 {
		return errors.New("cannot send empty packet")
	}

	ns.mu.Lock()
	packetSize := len(data)
	packetLossRate := ns.config.PacketLossRate
	latency := ns.config.Latency
	jitter := ns.config.Jitter
	bandwidthLimit := ns.config.BandwidthLimit
	ns.mu.Unlock()

	// Simulate packet loss
	if ns.shouldDropPacket(packetLossRate) {
		ns.mu.Lock()
		ns.packetsDropped++
		ns.mu.Unlock()
		return ErrPacketDropped
	}

	// Check bandwidth limit
	if bandwidthLimit > 0 {
		if err := ns.checkBandwidth(packetSize, bandwidthLimit); err != nil {
			return err
		}
	}

	// Calculate delivery time with jitter
	deliveryDelay := ns.calculateDelay(latency, jitter)

	// Create packet
	packet := &Packet{
		Data:      data,
		Timestamp: time.Now(),
		Size:      packetSize,
		Dropped:   false,
	}

	// Add to delay queue if latency > 0
	if deliveryDelay > 0 {
		delayed := &delayedPacket{
			packet:    packet,
			deliverAt: time.Now().Add(deliveryDelay),
		}

		ns.queueMu.Lock()
		ns.delayQueue = append(ns.delayQueue, delayed)
		ns.queueMu.Unlock()
	}

	// Update statistics
	ns.mu.Lock()
	ns.packetsSent++
	ns.bytesProcessed += uint64(packetSize)
	ns.mu.Unlock()

	return nil
}

// shouldDropPacket determines if a packet should be dropped.
func (ns *NetworkSimulator) shouldDropPacket(lossRate float64) bool {
	if lossRate <= 0 {
		return false
	}

	ns.mu.RLock()
	drop := ns.rng.Float64() < lossRate
	ns.mu.RUnlock()

	return drop
}

// calculateDelay calculates the delivery delay with jitter.
func (ns *NetworkSimulator) calculateDelay(latency, jitter time.Duration) time.Duration {
	if latency == 0 && jitter == 0 {
		return 0
	}

	baseDelay := latency

	if jitter > 0 {
		ns.mu.RLock()
		// Add random jitter: ±jitter
		jitterNs := ns.rng.Int63n(int64(jitter)*2) - int64(jitter)
		ns.mu.RUnlock()

		baseDelay += time.Duration(jitterNs)
		if baseDelay < 0 {
			baseDelay = 0
		}
	}

	return baseDelay
}

// checkBandwidth checks if sending would exceed bandwidth limit.
func (ns *NetworkSimulator) checkBandwidth(packetSize, limit int) error {
	ns.bandwidthMu.Lock()
	defer ns.bandwidthMu.Unlock()

	now := time.Now()

	// Reset counter every second
	if now.Sub(ns.bandwidthResetTime) >= time.Second {
		ns.bytesSentLastSecond = 0
		ns.bandwidthResetTime = now
	}

	// Check if packet would exceed limit
	if ns.bytesSentLastSecond+packetSize > limit {
		return ErrBandwidthExceeded
	}

	ns.bytesSentLastSecond += packetSize
	return nil
}

// ProcessDelayedPackets returns packets ready for delivery and removes them from queue.
func (ns *NetworkSimulator) ProcessDelayedPackets() []*Packet {
	ns.queueMu.Lock()
	defer ns.queueMu.Unlock()

	now := time.Now()
	ready := make([]*Packet, 0)
	remaining := make([]*delayedPacket, 0, len(ns.delayQueue))

	for _, delayed := range ns.delayQueue {
		if now.After(delayed.deliverAt) || now.Equal(delayed.deliverAt) {
			ready = append(ready, delayed.packet)
		} else {
			remaining = append(remaining, delayed)
		}
	}

	ns.delayQueue = remaining
	return ready
}

// GetStats returns current simulator statistics.
func (ns *NetworkSimulator) GetStats() (packetsSent, packetsDropped, bytesProcessed uint64) {
	ns.mu.RLock()
	defer ns.mu.RUnlock()
	return ns.packetsSent, ns.packetsDropped, ns.bytesProcessed
}

// Reset clears all statistics and queues.
func (ns *NetworkSimulator) Reset() {
	ns.mu.Lock()
	ns.packetsSent = 0
	ns.packetsDropped = 0
	ns.bytesProcessed = 0
	ns.mu.Unlock()

	ns.queueMu.Lock()
	ns.delayQueue = make([]*delayedPacket, 0, 1000)
	ns.queueMu.Unlock()

	ns.bandwidthMu.Lock()
	ns.bytesSentLastSecond = 0
	ns.bandwidthResetTime = time.Now()
	ns.bandwidthMu.Unlock()
}

// QueueSize returns the current number of delayed packets.
func (ns *NetworkSimulator) QueueSize() int {
	ns.queueMu.Lock()
	defer ns.queueMu.Unlock()
	return len(ns.delayQueue)
}
