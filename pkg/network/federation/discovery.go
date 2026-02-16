package federation

import (
	"encoding/json"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/opd-ai/venture/pkg/recovery"
	log "github.com/sirupsen/logrus"
)

const (
	// DiscoveryPort is the UDP port for LAN discovery broadcasts
	DiscoveryPort = 8090

	// DiscoveryInterval is how often to broadcast discovery packets
	DiscoveryInterval = 30 * time.Second

	// DiscoveryTimeout is how long to wait before removing unresponsive peers
	DiscoveryTimeout = 90 * time.Second

	// MaxGossipHops is the maximum number of hops for gossip protocol
	MaxGossipHops = 3
)

// DiscoveryPacket represents a LAN discovery broadcast message
type DiscoveryPacket struct {
	ServerID   string   `json:"server_id"`   // Server fingerprint
	ServerName string   `json:"server_name"` // Human-readable name
	Address    string   `json:"address"`     // IP:port
	Version    string   `json:"version"`     // Protocol version
	Features   []string `json:"features"`    // Supported features
	Timestamp  int64    `json:"timestamp"`   // Unix milliseconds
	Hops       int      `json:"hops"`        // Gossip hop count (0 = direct)
}

// GossipMessage represents a multi-hop server discovery message
type GossipMessage struct {
	MessageID string            `json:"message_id"` // UUID to prevent loops
	Servers   []DiscoveryPacket `json:"servers"`    // Known servers
	Timestamp int64             `json:"timestamp"`  // Unix milliseconds
	MaxHops   int               `json:"max_hops"`   // Remaining hops
	OriginID  string            `json:"origin_id"`  // Original sender
}

// GossipTransport defines the interface for sending gossip messages to peer servers.
// Implementations handle the actual network transport (TCP/TLS) for multi-hop discovery.
type GossipTransport interface {
	// SendGossip sends a gossip message to the specified peer server.
	SendGossip(peerID string, msg *GossipMessage) error
}

// DiscoverySystem handles peer discovery and gossip protocol
type DiscoverySystem struct {
	identity         *ServerIdentity
	listenAddr       string         // UDP listen address (e.g., ":8090")
	broadcastAddr    string         // Broadcast address (e.g., "255.255.255.255:8090")
	federationAddr   string         // TCP federation server address (e.g., "localhost:8080")
	conn             net.PacketConn // UDP connection
	knownPeers       map[string]*DiscoveredPeer
	mu               sync.RWMutex
	running          bool
	stopChan         chan struct{}
	seenMessages     map[string]int64 // MessageID -> timestamp (for dedup)
	cleanupTicker    *time.Ticker
	broadcastTicker  *time.Ticker
	onPeerDiscovered func(*DiscoveredPeer) // Callback for new peers
	gossipTransport  GossipTransport       // Transport for sending gossip to peers
}

// DiscoveredPeer represents a peer discovered via LAN or gossip
type DiscoveredPeer struct {
	ServerID   string
	ServerName string
	Address    string
	Version    string
	Features   []string
	LastSeen   time.Time
	Hops       int    // 0 = direct LAN, >0 = via gossip
	Via        string // ServerID of relay (empty if direct)
}

// NewDiscoverySystem creates a new discovery system
// Parameters:
//   - identity: Server identity for authentication and identification
//   - listenAddr: UDP address for discovery broadcasts (empty uses default port)
//   - federationAddr: TCP address where federation server listens (e.g., "localhost:8080")
func NewDiscoverySystem(identity *ServerIdentity, listenAddr, federationAddr string) (*DiscoverySystem, error) {
	if identity == nil {
		return nil, fmt.Errorf("identity cannot be nil")
	}
	if listenAddr == "" {
		listenAddr = fmt.Sprintf(":%d", DiscoveryPort)
	}
	if federationAddr == "" {
		federationAddr = "localhost:8080" // Default federation port
	}

	return &DiscoverySystem{
		identity:       identity,
		listenAddr:     listenAddr,
		broadcastAddr:  fmt.Sprintf("255.255.255.255:%d", DiscoveryPort),
		federationAddr: federationAddr,
		knownPeers:     make(map[string]*DiscoveredPeer),
		seenMessages:   make(map[string]int64),
		stopChan:       make(chan struct{}),
	}, nil
}

// Start begins listening for discovery broadcasts
func (ds *DiscoverySystem) Start() error {
	ds.mu.Lock()
	if ds.running {
		ds.mu.Unlock()
		return fmt.Errorf("discovery system already running")
	}

	// Listen on UDP
	conn, err := net.ListenPacket("udp", ds.listenAddr)
	if err != nil {
		ds.mu.Unlock()
		return fmt.Errorf("failed to listen on %s: %w", ds.listenAddr, err)
	}
	ds.conn = conn
	ds.running = true

	// Start cleanup ticker
	ds.cleanupTicker = time.NewTicker(30 * time.Second)

	// Start broadcast ticker
	ds.broadcastTicker = time.NewTicker(DiscoveryInterval)

	ds.mu.Unlock()

	// Start receive goroutine
	go ds.receiveLoop()

	// Start broadcast goroutine
	go ds.broadcastLoop()

	// Start cleanup goroutine
	go ds.cleanupLoop()

	return nil
}

// Stop halts the discovery system
func (ds *DiscoverySystem) Stop() error {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	if !ds.running {
		return fmt.Errorf("discovery system not running")
	}

	// Signal stop
	close(ds.stopChan)
	ds.running = false

	// Stop tickers
	if ds.cleanupTicker != nil {
		ds.cleanupTicker.Stop()
	}
	if ds.broadcastTicker != nil {
		ds.broadcastTicker.Stop()
	}

	// Close connection
	if ds.conn != nil {
		return ds.conn.Close()
	}

	return nil
}

// receiveLoop listens for incoming discovery packets
func (ds *DiscoverySystem) receiveLoop() {
	defer recovery.RecoverPanicWithLogger("federation_discovery", "receive loop", nil)()

	buf := make([]byte, 4096)

	for {
		select {
		case <-ds.stopChan:
			return
		default:
			// Set read deadline to avoid blocking indefinitely
			ds.conn.SetReadDeadline(time.Now().Add(1 * time.Second))

			n, addr, err := ds.conn.ReadFrom(buf)
			if err != nil {
				// Timeout is expected, continue
				if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
					continue
				}
				// Other errors, log and continue
				continue
			}

			// Process packet
			ds.processPacket(buf[:n], addr)
		}
	}
}

// processPacket handles an incoming discovery packet
func (ds *DiscoverySystem) processPacket(data []byte, addr net.Addr) {
	var packet DiscoveryPacket
	if err := json.Unmarshal(data, &packet); err != nil {
		// Invalid packet, ignore
		return
	}

	// Ignore our own broadcasts
	if packet.ServerID == ds.identity.ServerID {
		return
	}

	// Validate timestamp (within 60 seconds)
	now := time.Now().UnixMilli()
	if abs64(now-packet.Timestamp) > 60000 {
		return // Too old or too far in future
	}

	// Update or add peer
	ds.mu.Lock()
	defer ds.mu.Unlock()

	peer, exists := ds.knownPeers[packet.ServerID]
	if !exists {
		// New peer discovered
		peer = &DiscoveredPeer{
			ServerID:   packet.ServerID,
			ServerName: packet.ServerName,
			Address:    packet.Address,
			Version:    packet.Version,
			Features:   packet.Features,
			Hops:       packet.Hops,
		}
		ds.knownPeers[packet.ServerID] = peer

		// Trigger callback if set
		if ds.onPeerDiscovered != nil {
			go func(p *DiscoveredPeer) {
				defer recovery.RecoverPanicWithLogger("federation_discovery", "peer discovered callback", nil)()
				ds.onPeerDiscovered(p)
			}(peer)
		}
	} else {
		// Update existing peer
		peer.ServerName = packet.ServerName
		peer.Address = packet.Address
		peer.Version = packet.Version
		peer.Features = packet.Features
		peer.Hops = packet.Hops
	}

	peer.LastSeen = time.Now()
}

// broadcastLoop sends periodic discovery broadcasts
func (ds *DiscoverySystem) broadcastLoop() {
	defer recovery.RecoverPanicWithLogger("federation_discovery", "broadcast loop", nil)()

	// Send initial broadcast immediately
	ds.broadcastDiscovery()

	for {
		select {
		case <-ds.stopChan:
			return
		case <-ds.broadcastTicker.C:
			ds.broadcastDiscovery()
		}
	}
}

// broadcastDiscovery sends a discovery packet to the broadcast address
func (ds *DiscoverySystem) broadcastDiscovery() {
	packet := DiscoveryPacket{
		ServerID:   ds.identity.ServerID,
		ServerName: ds.identity.ServerName,
		Address:    ds.getLocalAddress(),
		Version:    "6.0.0",                             // V6.0 protocol version
		Features:   []string{"travel", "trade", "post"}, // Default features
		Timestamp:  time.Now().UnixMilli(),
		Hops:       0, // Direct broadcast
	}

	data, err := json.Marshal(packet)
	if err != nil {
		return
	}

	addr, err := resolveUDPAddr(ds.broadcastAddr)
	if err != nil {
		log.WithFields(log.Fields{
			"broadcast_addr": ds.broadcastAddr,
			"error":          err.Error(),
		}).Warn("Failed to resolve UDP broadcast address")
		return
	}

	ds.conn.WriteTo(data, addr)
}

// resolveUDPAddr resolves a UDP address string and returns it as net.Addr.
// Uses net.ResolveUDPAddr internally (the only stdlib option for UDP resolution)
// but returns the interface type to maintain interface-based networking.
func resolveUDPAddr(address string) (net.Addr, error) {
	return net.ResolveUDPAddr("udp", address)
}

// getLocalAddress returns the server's listening address for federation
func (ds *DiscoverySystem) getLocalAddress() string {
	return ds.federationAddr
}

// cleanupLoop removes stale peers
func (ds *DiscoverySystem) cleanupLoop() {
	defer recovery.RecoverPanicWithLogger("federation_discovery", "cleanup loop", nil)()

	for {
		select {
		case <-ds.stopChan:
			return
		case <-ds.cleanupTicker.C:
			ds.cleanupStalePeers()
		}
	}
}

// cleanupStalePeers removes peers not seen within DiscoveryTimeout
func (ds *DiscoverySystem) cleanupStalePeers() {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	now := time.Now()
	for id, peer := range ds.knownPeers {
		if now.Sub(peer.LastSeen) > DiscoveryTimeout {
			delete(ds.knownPeers, id)
		}
	}

	// Also cleanup old seen messages (older than 5 minutes)
	cutoff := now.Add(-5 * time.Minute).UnixMilli()
	for msgID, timestamp := range ds.seenMessages {
		if timestamp < cutoff {
			delete(ds.seenMessages, msgID)
		}
	}
}

// GetPeers returns all currently known peers
func (ds *DiscoverySystem) GetPeers() []*DiscoveredPeer {
	ds.mu.RLock()
	defer ds.mu.RUnlock()

	peers := make([]*DiscoveredPeer, 0, len(ds.knownPeers))
	for _, peer := range ds.knownPeers {
		// Deep copy to avoid external mutation
		peerCopy := *peer
		peerCopy.Features = append([]string{}, peer.Features...)
		peers = append(peers, &peerCopy)
	}

	return peers
}

// GetPeer returns a specific peer by server ID
func (ds *DiscoverySystem) GetPeer(serverID string) (*DiscoveredPeer, bool) {
	ds.mu.RLock()
	defer ds.mu.RUnlock()

	peer, exists := ds.knownPeers[serverID]
	if !exists {
		return nil, false
	}

	// Deep copy
	peerCopy := *peer
	peerCopy.Features = append([]string{}, peer.Features...)
	return &peerCopy, true
}

// AddManualPeer adds a peer manually (not via discovery)
func (ds *DiscoverySystem) AddManualPeer(serverID, serverName, address, version string, features []string) error {
	if serverID == "" || address == "" {
		return fmt.Errorf("serverID and address are required")
	}

	ds.mu.Lock()
	defer ds.mu.Unlock()

	peer := &DiscoveredPeer{
		ServerID:   serverID,
		ServerName: serverName,
		Address:    address,
		Version:    version,
		Features:   features,
		LastSeen:   time.Now(),
		Hops:       0, // Direct connection
	}

	ds.knownPeers[serverID] = peer

	// Trigger callback if set
	if ds.onPeerDiscovered != nil {
		go func(p *DiscoveredPeer) {
			defer recovery.RecoverPanicWithLogger("federation_discovery", "manual peer discovered callback", nil)()
			ds.onPeerDiscovered(p)
		}(peer)
	}

	return nil
}

// RemovePeer removes a peer from the known peers list
func (ds *DiscoverySystem) RemovePeer(serverID string) {
	ds.mu.Lock()
	defer ds.mu.Unlock()
	delete(ds.knownPeers, serverID)
}

// OnPeerDiscovered sets a callback for when new peers are discovered
func (ds *DiscoverySystem) OnPeerDiscovered(callback func(*DiscoveredPeer)) {
	ds.mu.Lock()
	defer ds.mu.Unlock()
	ds.onPeerDiscovered = callback
}

// SetGossipTransport sets the transport used for sending gossip messages to peers.
// When set, PropagateGossip will transmit messages via the transport layer.
// If nil, PropagateGossip will prepare messages but skip transmission.
func (ds *DiscoverySystem) SetGossipTransport(transport GossipTransport) {
	ds.mu.Lock()
	defer ds.mu.Unlock()
	ds.gossipTransport = transport
}

// PropagateGossip forwards discovery information to connected peers
func (ds *DiscoverySystem) PropagateGossip(targetPeer string) error {
	ds.mu.RLock()
	defer ds.mu.RUnlock()

	// Create gossip message with all known peers
	servers := make([]DiscoveryPacket, 0, len(ds.knownPeers))
	for _, peer := range ds.knownPeers {
		// Don't gossip about the target peer to itself
		if peer.ServerID == targetPeer {
			continue
		}

		// Skip peers that have exceeded max gossip hops
		if peer.Hops >= MaxGossipHops {
			continue
		}

		// Increment hop count
		servers = append(servers, DiscoveryPacket{
			ServerID:   peer.ServerID,
			ServerName: peer.ServerName,
			Address:    peer.Address,
			Version:    peer.Version,
			Features:   peer.Features,
			Timestamp:  time.Now().UnixMilli(),
			Hops:       peer.Hops + 1,
		})
	}

	if len(servers) == 0 {
		return nil
	}

	msg := &GossipMessage{
		MessageID: fmt.Sprintf("gossip-%s-%d", ds.identity.ServerID, time.Now().UnixNano()),
		Servers:   servers,
		Timestamp: time.Now().UnixMilli(),
		MaxHops:   MaxGossipHops,
		OriginID:  ds.identity.ServerID,
	}

	if ds.gossipTransport == nil {
		log.WithFields(log.Fields{
			"target_peer":  targetPeer,
			"server_count": len(servers),
		}).Debug("gossip transport not configured, skipping transmission")
		return nil
	}

	if err := ds.gossipTransport.SendGossip(targetPeer, msg); err != nil {
		log.WithFields(log.Fields{
			"target_peer": targetPeer,
			"error":       err,
		}).Warn("failed to send gossip message")
		return fmt.Errorf("gossip transmission to %s failed: %w", targetPeer, err)
	}

	log.WithFields(log.Fields{
		"target_peer":  targetPeer,
		"server_count": len(servers),
		"message_id":   msg.MessageID,
	}).Debug("gossip message propagated")

	return nil
}

// abs64 returns absolute value of int64
func abs64(x int64) int64 {
	if x < 0 {
		return -x
	}
	return x
}
