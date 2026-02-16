// Package webrtc NAT traversal coordination.
// This file coordinates NAT traversal using multiple methods (Direct, STUN, TURN)
// and manages TURN relay connections for symmetric NAT scenarios.
package webrtc

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

// NATTraversal coordinates STUN/TURN for establishing P2P connections.
type NATTraversal struct {
	mu sync.RWMutex

	stunClient   *STUNClient
	relayManager *RelayManager

	// timeProvider abstracts time access for deterministic testing.
	timeProvider TimeProvider

	// Statistics
	totalAttempts    int
	directSuccess    int
	stunSuccess      int
	turnSuccess      int
	failures         int
	averageSetupTime time.Duration
	natTraversalRate float64
}

// NewNATTraversal creates a new NAT traversal coordinator.
func NewNATTraversal(stunServers []string, relayManager *RelayManager) *NATTraversal {
	return &NATTraversal{
		stunClient:   NewSTUNClient(stunServers),
		relayManager: relayManager,
		timeProvider: DefaultTimeProvider(),
	}
}

// TraversalResult represents the result of a NAT traversal attempt.
type TraversalResult struct {
	// Success indicates if traversal was successful.
	Success bool
	// Method is the method that succeeded (Direct, STUN, TURN).
	Method TraversalMethod
	// PublicIP is the discovered public IP (for STUN).
	PublicIP string
	// PublicPort is the discovered public port (for STUN).
	PublicPort int
	// RelayNode is the TURN relay used (if Method == MethodTURN).
	RelayNode *RelayNode
	// SetupTime is how long traversal took.
	SetupTime time.Duration
	// NATType is the detected NAT type.
	NATType NATType
	// Error if traversal failed.
	Error error
}

// TraversalMethod represents the NAT traversal method used.
type TraversalMethod int

// String returns the string representation of TraversalMethod.
func (m TraversalMethod) String() string {
	switch m {
	case MethodDirect:
		return "Direct"
	case MethodSTUN:
		return "STUN"
	case MethodTURN:
		return "TURN"
	case MethodFailed:
		return "Failed"
	default:
		return "Unknown"
	}
}

// EstablishConnection attempts NAT traversal using multiple methods.
func (n *NATTraversal) EstablishConnection(ctx context.Context) (*TraversalResult, error) {
	start := n.timeProvider.Now()

	n.mu.Lock()
	n.totalAttempts++
	n.mu.Unlock()

	result := &TraversalResult{
		Success: false,
		Method:  MethodFailed,
	}

	// Method 1: Try direct connection (no NAT)
	if err := n.tryDirectConnection(ctx); err == nil {
		result.Success = true
		result.Method = MethodDirect
		result.SetupTime = n.timeProvider.Now().Sub(start)
		result.NATType = NATTypeNone

		n.mu.Lock()
		n.directSuccess++
		n.updateAverageSetupTime(result.SetupTime)
		n.mu.Unlock()

		log.WithFields(log.Fields{
			"traversal_method": "direct",
			"setup_time":       result.SetupTime,
		}).Debug("NAT traversal succeeded")

		return result, nil
	}

	// Method 2: Try STUN-assisted connection
	stunResp, err := n.trySTUNConnection(ctx)
	if err == nil {
		result.Success = true
		result.Method = MethodSTUN
		result.PublicIP = stunResp.PublicIP.String()
		result.PublicPort = stunResp.PublicPort
		result.NATType = stunResp.NATType
		result.SetupTime = n.timeProvider.Now().Sub(start)

		n.mu.Lock()
		n.stunSuccess++
		n.updateAverageSetupTime(result.SetupTime)
		n.mu.Unlock()

		log.WithFields(log.Fields{
			"traversal_method": "stun",
			"public_ip":        result.PublicIP,
			"public_port":      result.PublicPort,
			"setup_time":       result.SetupTime,
		}).Debug("NAT traversal succeeded")

		return result, nil
	}

	// Method 3: Fallback to TURN relay
	relay, err := n.tryTURNConnection(ctx)
	if err == nil {
		result.Success = true
		result.Method = MethodTURN
		result.RelayNode = relay
		result.SetupTime = n.timeProvider.Now().Sub(start)

		n.mu.Lock()
		n.turnSuccess++
		n.updateAverageSetupTime(result.SetupTime)
		n.mu.Unlock()

		log.WithFields(log.Fields{
			"traversal_method": "turn",
			"relay_id":         relay.ID,
			"setup_time":       result.SetupTime,
		}).Debug("NAT traversal succeeded")

		return result, nil
	}

	// All methods failed
	result.Error = ErrAllMethodsFailed
	result.SetupTime = n.timeProvider.Now().Sub(start)

	n.mu.Lock()
	n.failures++
	n.mu.Unlock()

	log.WithFields(log.Fields{
		"setup_time": result.SetupTime,
	}).Warn("NAT traversal failed: all methods exhausted")

	return result, ErrNATTraversalFailed
}

// tryDirectConnection attempts a direct P2P connection.
func (n *NATTraversal) tryDirectConnection(ctx context.Context) error {
	// In a real implementation, this would attempt to establish
	// a direct connection without any NAT traversal assistance.
	// For testing, we simulate a failure (most connections need NAT traversal).
	return errors.New("direct connection not available")
}

// trySTUNConnection attempts STUN-assisted NAT traversal.
func (n *NATTraversal) trySTUNConnection(ctx context.Context) (*STUNResponse, error) {
	if n.stunClient == nil {
		return nil, errors.New("STUN client not available")
	}

	resp, err := n.stunClient.GetPublicAddress(ctx)
	if err != nil {
		return nil, fmt.Errorf("STUN query failed: %w", err)
	}

	// Detect NAT type for better connection strategy
	natType, _ := n.stunClient.DetectNATType(ctx)
	resp.NATType = natType

	// Symmetric NAT requires TURN relay
	if natType == NATTypeSymmetric {
		return nil, errors.New("symmetric NAT requires TURN relay")
	}

	return resp, nil
}

// tryTURNConnection attempts TURN relay connection.
func (n *NATTraversal) tryTURNConnection(ctx context.Context) (*RelayNode, error) {
	if n.relayManager == nil {
		return nil, ErrNoRelayAvailable
	}

	relay, err := n.relayManager.SelectRelay()
	if err != nil {
		return nil, fmt.Errorf("relay selection failed: %w", err)
	}

	// Attempt to acquire relay connection
	if err := relay.Acquire(); err != nil {
		return nil, fmt.Errorf("relay acquisition failed: %w", err)
	}

	// In a real implementation, this would establish the TURN allocation
	// and create the relayed connection.

	return relay, nil
}

// updateAverageSetupTime updates the running average setup time.
func (n *NATTraversal) updateAverageSetupTime(setupTime time.Duration) {
	// Simple exponential moving average
	alpha := 0.2
	if n.averageSetupTime == 0 {
		n.averageSetupTime = setupTime
	} else {
		n.averageSetupTime = time.Duration(
			float64(n.averageSetupTime)*(1-alpha) + float64(setupTime)*alpha,
		)
	}

	// Update NAT traversal success rate
	total := n.directSuccess + n.stunSuccess + n.turnSuccess + n.failures
	if total > 0 {
		successes := n.directSuccess + n.stunSuccess + n.turnSuccess
		n.natTraversalRate = float64(successes) / float64(total)
	}
}

// GetStats returns NAT traversal statistics.
func (n *NATTraversal) GetStats() NATTraversalStats {
	n.mu.RLock()
	defer n.mu.RUnlock()

	var turnFallbackRate float64
	totalSuccess := n.directSuccess + n.stunSuccess + n.turnSuccess
	if totalSuccess > 0 {
		turnFallbackRate = float64(n.turnSuccess) / float64(totalSuccess)
	}

	return NATTraversalStats{
		TotalAttempts:    n.totalAttempts,
		DirectSuccess:    n.directSuccess,
		STUNSuccess:      n.stunSuccess,
		TURNSuccess:      n.turnSuccess,
		Failures:         n.failures,
		SuccessRate:      n.natTraversalRate,
		TURNFallbackRate: turnFallbackRate,
		AverageSetupTime: n.averageSetupTime,
	}
}

// NATTraversalStats holds NAT traversal statistics.
type NATTraversalStats struct {
	TotalAttempts    int
	DirectSuccess    int
	STUNSuccess      int
	TURNSuccess      int
	Failures         int
	SuccessRate      float64
	TURNFallbackRate float64
	AverageSetupTime time.Duration
}

// RelayConnection represents an active TURN relay connection.
type RelayConnection struct {
	mu sync.RWMutex

	relay        *RelayNode
	localAddr    string
	relayAddr    string
	expiry       time.Time
	active       bool
	bytesRelay   uint64
	timeProvider TimeProvider
}

// NewRelayConnection creates a new relay connection.
func NewRelayConnection(relay *RelayNode, localAddr, relayAddr string, lifetime time.Duration) *RelayConnection {
	tp := DefaultTimeProvider()
	return &RelayConnection{
		relay:        relay,
		localAddr:    localAddr,
		relayAddr:    relayAddr,
		expiry:       tp.Now().Add(lifetime),
		active:       true,
		timeProvider: tp,
	}
}

// Send sends data through the TURN relay.
func (r *RelayConnection) Send(data []byte) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if !r.active {
		return errors.New("relay connection inactive")
	}

	if r.timeProvider.Now().After(r.expiry) {
		r.active = false
		return errors.New("relay allocation expired")
	}

	// In production, this would send via TURN relay
	r.bytesRelay += uint64(len(data))
	r.relay.AddBytesRelayed(uint64(len(data)))

	return nil
}

// Refresh extends the relay allocation lifetime.
func (r *RelayConnection) Refresh(lifetime time.Duration) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if !r.active {
		return errors.New("relay connection inactive")
	}

	r.expiry = r.timeProvider.Now().Add(lifetime)
	return nil
}

// Close releases the relay connection.
func (r *RelayConnection) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if !r.active {
		return nil
	}

	r.active = false
	r.relay.Release()
	return nil
}

// GetStats returns relay connection statistics.
func (r *RelayConnection) GetStats() RelayConnectionStats {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return RelayConnectionStats{
		RelayNode:  r.relay.ID,
		LocalAddr:  r.localAddr,
		RelayAddr:  r.relayAddr,
		BytesRelay: r.bytesRelay,
		Active:     r.active,
		Expiry:     r.expiry,
	}
}

// RelayConnectionStats holds relay connection statistics.
type RelayConnectionStats struct {
	RelayNode  string
	LocalAddr  string
	RelayAddr  string
	BytesRelay uint64
	Active     bool
	Expiry     time.Time
}
