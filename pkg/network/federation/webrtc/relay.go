// Package webrtc TURN relay management.
// This file manages a pool of TURN relay servers with health checking,
// load balancing, and multiple selection strategies for optimal relay selection.
package webrtc

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/opd-ai/venture/pkg/recovery"
)

// RelayNode represents a single TURN relay server.
type RelayNode struct {
	// ID is the unique identifier for this relay.
	ID string
	// URL is the TURN server URL (turn:host:port or turns:host:port).
	URL string
	// Username for TURN authentication.
	Username string
	// Credential for TURN authentication.
	Credential string
	// Region is the geographic region (e.g., "us-east", "eu-west").
	Region string
	// MaxConnections is the maximum concurrent connections supported.
	MaxConnections int

	mu sync.RWMutex

	// activeConnections is the current number of active connections.
	activeConnections int
	// totalConnections is the total connections handled.
	totalConnections int
	// bytesRelayed is the total bytes relayed through this node.
	bytesRelayed uint64
	// latency is the measured latency to this relay.
	latency time.Duration
	// bandwidth is the measured available bandwidth (bytes/sec).
	bandwidth uint64
	// lastHealthCheck is when the relay was last checked.
	lastHealthCheck time.Time
	// healthy indicates if the relay is currently healthy.
	healthy bool
	// timeProvider abstracts time access for deterministic testing.
	timeProvider TimeProvider
}

// NewRelayNode creates a new relay node with the given configuration.
func NewRelayNode(id, url, username, credential, region string, maxConn int) *RelayNode {
	return &RelayNode{
		ID:             id,
		URL:            url,
		Username:       username,
		Credential:     credential,
		Region:         region,
		MaxConnections: maxConn,
		healthy:        true,
		timeProvider:   DefaultTimeProvider(),
	}
}

// Acquire attempts to allocate a connection slot on this relay.
func (r *RelayNode) Acquire() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if !r.healthy {
		return fmt.Errorf("relay %s is unhealthy", r.ID)
	}

	if r.activeConnections >= r.MaxConnections {
		return ErrRelayFull
	}

	r.activeConnections++
	r.totalConnections++
	return nil
}

// Release releases a connection slot on this relay.
func (r *RelayNode) Release() {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.activeConnections > 0 {
		r.activeConnections--
	}
}

// AddBytesRelayed increments the bytes relayed counter.
func (r *RelayNode) AddBytesRelayed(bytes uint64) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.bytesRelayed += bytes
}

// UpdateLatency updates the measured latency to this relay.
func (r *RelayNode) UpdateLatency(latency time.Duration) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.latency = latency
}

// UpdateBandwidth updates the measured bandwidth for this relay.
func (r *RelayNode) UpdateBandwidth(bandwidth uint64) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.bandwidth = bandwidth
}

// SetHealthy marks this relay as healthy or unhealthy.
func (r *RelayNode) SetHealthy(healthy bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.healthy = healthy
	r.lastHealthCheck = r.timeProvider.Now()
}

// GetStats returns current statistics for this relay.
func (r *RelayNode) GetStats() RelayStats {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return RelayStats{
		ID:                r.ID,
		URL:               r.URL,
		Region:            r.Region,
		ActiveConnections: r.activeConnections,
		TotalConnections:  r.totalConnections,
		BytesRelayed:      r.bytesRelayed,
		Latency:           r.latency,
		Bandwidth:         r.bandwidth,
		Healthy:           r.healthy,
		Utilization:       float64(r.activeConnections) / float64(r.MaxConnections),
	}
}

// IsAvailable returns true if the relay can accept new connections.
func (r *RelayNode) IsAvailable() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.healthy && r.activeConnections < r.MaxConnections
}

// RelayStats holds statistics about a relay node.
type RelayStats struct {
	ID                string
	URL               string
	Region            string
	ActiveConnections int
	TotalConnections  int
	BytesRelayed      uint64
	Latency           time.Duration
	Bandwidth         uint64
	Healthy           bool
	Utilization       float64
}

// RelayManager manages a pool of TURN relay nodes.
type RelayManager struct {
	mu    sync.RWMutex
	nodes map[string]*RelayNode

	// Selection strategy
	strategy SelectionStrategy

	// Round-robin counter
	rrCounter int

	// Health check interval
	healthCheckInterval time.Duration

	// timeProvider abstracts time access for deterministic testing.
	timeProvider TimeProvider

	// Context for health checks
	ctx    context.Context
	cancel context.CancelFunc
}

// SelectionStrategy defines how relays are selected.
type SelectionStrategy int

// String returns the string representation of SelectionStrategy.
func (s SelectionStrategy) String() string {
	switch s {
	case StrategyLowestLatency:
		return "LowestLatency"
	case StrategyHighestBandwidth:
		return "HighestBandwidth"
	case StrategyLowestUtilization:
		return "LowestUtilization"
	case StrategyRoundRobin:
		return "RoundRobin"
	default:
		return "Unknown"
	}
}

// NewRelayManager creates a new relay manager with the given strategy.
func NewRelayManager(strategy SelectionStrategy) *RelayManager {
	ctx, cancel := context.WithCancel(context.Background())
	rm := &RelayManager{
		nodes:               make(map[string]*RelayNode),
		strategy:            strategy,
		healthCheckInterval: 30 * time.Second,
		timeProvider:        DefaultTimeProvider(),
		ctx:                 ctx,
		cancel:              cancel,
	}

	// Start health check loop
	go func() {
		defer recovery.RecoverPanicWithLogger("webrtc_relay", "health check loop", nil)()
		rm.healthCheckLoop()
	}()

	return rm
}

// AddRelay adds a relay node to the manager.
func (rm *RelayManager) AddRelay(node *RelayNode) error {
	if node == nil {
		return fmt.Errorf("relay node cannot be nil")
	}

	rm.mu.Lock()
	defer rm.mu.Unlock()

	if _, exists := rm.nodes[node.ID]; exists {
		return fmt.Errorf("relay %s already exists", node.ID)
	}

	rm.nodes[node.ID] = node
	return nil
}

// RemoveRelay removes a relay node from the manager.
func (rm *RelayManager) RemoveRelay(id string) error {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	if _, exists := rm.nodes[id]; !exists {
		return fmt.Errorf("relay %s not found", id)
	}

	delete(rm.nodes, id)
	return nil
}

// SelectRelay selects the best relay based on the configured strategy.
func (rm *RelayManager) SelectRelay() (*RelayNode, error) {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	if len(rm.nodes) == 0 {
		return nil, ErrNoRelayAvailable
	}

	var candidates []*RelayNode
	for _, node := range rm.nodes {
		if node.IsAvailable() {
			candidates = append(candidates, node)
		}
	}

	if len(candidates) == 0 {
		return nil, ErrRelayFull
	}

	switch rm.strategy {
	case StrategyLowestLatency:
		return rm.selectLowestLatency(candidates), nil
	case StrategyHighestBandwidth:
		return rm.selectHighestBandwidth(candidates), nil
	case StrategyLowestUtilization:
		return rm.selectLowestUtilization(candidates), nil
	case StrategyRoundRobin:
		return rm.selectRoundRobin(candidates), nil
	default:
		return candidates[0], nil
	}
}

// selectLowestLatency finds the relay with lowest latency.
// Snapshots each node's latency individually to avoid holding multiple locks.
func (rm *RelayManager) selectLowestLatency(candidates []*RelayNode) *RelayNode {
	if len(candidates) == 0 {
		return nil
	}

	best := candidates[0]
	best.mu.RLock()
	bestLatency := best.latency
	best.mu.RUnlock()

	for _, node := range candidates[1:] {
		node.mu.RLock()
		nodeLatency := node.latency
		node.mu.RUnlock()

		if nodeLatency < bestLatency && nodeLatency > 0 {
			best = node
			bestLatency = nodeLatency
		}
	}
	return best
}

// selectHighestBandwidth finds the relay with highest bandwidth.
// Snapshots each node's bandwidth individually to avoid holding multiple locks.
func (rm *RelayManager) selectHighestBandwidth(candidates []*RelayNode) *RelayNode {
	if len(candidates) == 0 {
		return nil
	}

	best := candidates[0]
	best.mu.RLock()
	bestBandwidth := best.bandwidth
	best.mu.RUnlock()

	for _, node := range candidates[1:] {
		node.mu.RLock()
		nodeBandwidth := node.bandwidth
		node.mu.RUnlock()

		if nodeBandwidth > bestBandwidth {
			best = node
			bestBandwidth = nodeBandwidth
		}
	}
	return best
}

// selectLowestUtilization finds the relay with most available capacity.
// Snapshots each node's utilization individually to avoid holding multiple locks.
func (rm *RelayManager) selectLowestUtilization(candidates []*RelayNode) *RelayNode {
	if len(candidates) == 0 {
		return nil
	}

	best := candidates[0]
	best.mu.RLock()
	bestUtil := float64(best.activeConnections) / float64(best.MaxConnections)
	best.mu.RUnlock()

	for _, node := range candidates[1:] {
		node.mu.RLock()
		nodeUtil := float64(node.activeConnections) / float64(node.MaxConnections)
		node.mu.RUnlock()

		if nodeUtil < bestUtil {
			best = node
			bestUtil = nodeUtil
		}
	}
	return best
}

// selectRoundRobin rotates through available relays.
// Resets counter on overflow to prevent negative modulo results.
func (rm *RelayManager) selectRoundRobin(candidates []*RelayNode) *RelayNode {
	if len(candidates) == 0 {
		return nil
	}

	idx := rm.rrCounter % len(candidates)
	rm.rrCounter++
	if rm.rrCounter < 0 {
		rm.rrCounter = 0
	}
	return candidates[idx]
}

// GetRelayStats returns statistics for all relays.
func (rm *RelayManager) GetRelayStats() []RelayStats {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	stats := make([]RelayStats, 0, len(rm.nodes))
	for _, node := range rm.nodes {
		stats = append(stats, node.GetStats())
	}
	return stats
}

// healthCheckLoop periodically checks relay health.
func (rm *RelayManager) healthCheckLoop() {
	ticker := time.NewTicker(rm.healthCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-rm.ctx.Done():
			return
		case <-ticker.C:
			rm.performHealthChecks()
		}
	}
}

// performHealthChecks checks the health of all relays.
func (rm *RelayManager) performHealthChecks() {
	rm.mu.RLock()
	nodes := make([]*RelayNode, 0, len(rm.nodes))
	for _, node := range rm.nodes {
		nodes = append(nodes, node)
	}
	rm.mu.RUnlock()

	for _, node := range nodes {
		go func(n *RelayNode) {
			defer recovery.RecoverPanicWithLogger("webrtc_relay", "check relay health", nil)()
			rm.checkRelayHealth(n)
		}(node)
	}
}

// checkRelayHealth checks if a relay is reachable and measures latency.
func (rm *RelayManager) checkRelayHealth(node *RelayNode) {
	ctx, cancel := context.WithTimeout(rm.ctx, 5*time.Second)
	defer cancel()

	start := rm.timeProvider.Now()
	healthy := rm.pingRelay(ctx, node.URL)
	latency := rm.timeProvider.Now().Sub(start)

	node.SetHealthy(healthy)
	if healthy {
		node.UpdateLatency(latency)
	}

	log.WithFields(log.Fields{
		"relay_id": node.ID,
		"healthy":  healthy,
		"latency":  latency,
	}).Debug("relay health check completed")
}

// pingRelay attempts to connect to a relay to verify it's reachable.
func (rm *RelayManager) pingRelay(ctx context.Context, url string) bool {
	// Parse TURN URL to extract host:port
	// Format: turn:host:port or turns:host:port
	// Validate URL has proper prefix before slicing
	if len(url) < 5 || (url[:5] != "turn:" && url[:6] != "turns:") {
		log.WithFields(log.Fields{
			"url": url,
		}).Warn("invalid TURN URL format")
		return false
	}

	// Skip "turn:" (5 chars) or "turns:" (6 chars)
	urlSuffix := url[5:]
	if url[:6] == "turns:" {
		urlSuffix = url[6:]
	}

	host, _, err := net.SplitHostPort(urlSuffix)
	if err != nil {
		log.WithFields(log.Fields{
			"url":   url,
			"error": err,
		}).Debug("failed to parse TURN URL")
		return false
	}

	// Attempt connection with timeout
	dialer := net.Dialer{Timeout: 3 * time.Second}
	conn, err := dialer.DialContext(ctx, "udp", host)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

// Close shuts down the relay manager.
func (rm *RelayManager) Close() {
	rm.cancel()
}
