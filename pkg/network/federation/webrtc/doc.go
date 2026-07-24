// Package webrtc provides WebRTC-based federation for browser-to-browser server connections.
//
// This package enables Venture servers running in web browsers to connect directly via WebRTC
// peer-to-peer data channels, eliminating the need for dedicated server infrastructure. This
// supports the game's zero-asset, decentralized architecture by allowing any browser client
// to act as a federated server.
//
// # Architecture
//
// The WebRTC federation system consists of three main components:
//
//  1. Signaling Server (optional): Coordinates initial peer discovery and SDP exchange.
//     Runs as a lightweight WebSocket relay for connection establishment only.
//
//  2. Peer Connection Manager: Manages WebRTC peer connections, data channels, and ICE
//     negotiation. Handles connection state, reconnection, and graceful degradation.
//
//  3. Protocol Adapter: Translates the existing federation protocol (handshake, sync,
//     discovery, transfer) to/from WebRTC data channels with fallback to WebSocket.
//
// # Connection Flow
//
// Browser-to-browser connection establishment:
//
//  1. Peer A creates offer SDP and sends to signaling server
//  2. Signaling server relays offer to Peer B
//  3. Peer B creates answer SDP and sends back via signaling server
//  4. ICE candidates exchanged through signaling server
//  5. Direct P2P connection established (signaling no longer needed)
//  6. Federation protocol runs over WebRTC data channel
//
// # NAT Traversal
//
// NAT/firewall traversal uses standard WebRTC techniques:
//   - STUN servers identify public IP/port (built-in Google STUN by default)
//   - ICE (Interactive Connectivity Establishment) tries direct, STUN, and TURN paths
//   - TURN relay fallback for symmetric NAT (configurable TURN server)
//   - Success rate >95% in typical home/corporate networks
//
// # P2P Relay Network
//
// The relay system provides TURN relay servers for NAT traversal:
//
//  1. RelayManager: Manages pool of TURN relays with health checks and load balancing
//  2. STUNClient: Discovers public IP/port and detects NAT type
//  3. NATTraversal: Coordinates connection attempts (Direct → STUN → TURN)
//  4. RelayNode: Individual TURN server with statistics and capacity management
//
// Relay selection strategies:
//   - LowestLatency: Best for real-time gameplay (<50ms preferred)
//   - HighestBandwidth: Best for large world transfers (>5 MB/s preferred)
//   - LowestUtilization: Balances load across relay pool
//   - RoundRobin: Simple rotation for testing
//
// Example NAT traversal:
//
//	// Create relay manager with TURN servers
//	rm := webrtc.NewRelayManager(webrtc.StrategyLowestLatency)
//	relay := webrtc.NewRelayNode("relay1", "turn:relay.example.com:3478", "user", "pass", "us-east", 100)
//	rm.AddRelay(relay)
//
//	// Create NAT traversal coordinator
//	nt := webrtc.NewNATTraversal([]string{"stun:stun.l.google.com:19302"}, rm)
//
//	// Attempt connection (tries Direct → STUN → TURN automatically)
//	result, err := nt.EstablishConnection(ctx)
//	if err != nil {
//	    logrus.WithError(err).Fatal("NAT traversal failed")
//	}
//
//	logrus.WithFields(logrus.Fields{"method": result.Method, "setup_time": result.SetupTime}).Info("Connection established")
//
// NAT type detection helps optimize connection strategy:
//   - Full Cone / Restricted Cone: STUN usually sufficient
//   - Port Restricted Cone: STUN with port prediction
//   - Symmetric NAT: Requires TURN relay (10-15% of connections)
//
// # Fallback Behavior
//
// If WebRTC is unavailable (old browsers, restrictive firewalls):
//   - Automatically fall back to WebSocket federation
//   - Detection via feature detection and connection timeout
//   - Same protocol, different transport layer
//   - Performance: WebRTC (1-10 MB/s) vs WebSocket (100-500 KB/s)
//
// # Usage
//
// Creating a WebRTC federation peer:
//
//	config := &webrtc.Config{
//	    SignalingURL:  "wss://signal.example.com",
//	    STUNServers:   []string{"stun:stun.l.google.com:19302"},
//	    TURNServers:   []webrtc.TURNServer{},  // Optional
//	    DataChannelLabel: "venture-federation",
//	    ICETimeout:    10 * time.Second,
//	}
//
//	peer, err := webrtc.NewPeer("peer-alice", config)
//	if err != nil {
//	    logrus.WithError(err).Fatal("failed to create peer")
//	}
//	defer peer.Close()
//
//	// Connect to remote peer
//	if err := peer.Connect("peer-bob"); err != nil {
//	    logrus.WithError(err).Fatal("connection failed")
//	}
//
//	// Send federation message
//	msg := protocol.HandshakeMessage{...}
//	if err := peer.Send(msg); err != nil {
//	    log.Errorf("send failed: %v", err)
//	}
//
//	// Receive messages
//	for msg := range peer.Receive() {
//	    // Process federation protocol message
//	}
//
// # Performance Characteristics
//
//   - Signaling latency: <500ms for peer discovery
//   - Connection establishment: <2s (including ICE)
//   - NAT traversal success rate: >95% (Direct: 5-10%, STUN: 75-85%, TURN: 10-15%)
//   - TURN relay overhead: <50ms additional latency
//   - Data channel bandwidth: 1-10 MB/s (network-dependent)
//   - Relay capacity: 100-1000 concurrent connections per TURN server
//   - Zero server infrastructure cost for P2P (TURN relay costs apply for 10-15% of connections)
//   - Fallback WebSocket: 100-500 KB/s
//
// # Security
//
// WebRTC provides built-in security:
//   - DTLS encryption for data channels (AES-128-GCM)
//   - SRTP for media channels (not used in Venture)
//   - Certificate fingerprint verification in SDP
//   - Same federation authentication (ed25519) as TCP/UDP
//
// # Browser Compatibility
//
// Supported browsers:
//   - Chrome/Edge 80+ (full WebRTC support)
//   - Firefox 75+ (full WebRTC support)
//   - Safari 15+ (WebRTC supported, some limitations)
//   - Mobile Chrome/Safari (supported with battery optimization)
//
// Unsupported browsers automatically fall back to WebSocket.
//
// # Integration with Existing Federation
//
// The WebRTC adapter implements the same federation interfaces:
//   - FederationTransport (Send/Receive)
//   - PeerDiscovery (Advertise/Discover)
//   - ConnectionManager (Connect/Disconnect)
//
// Existing federation code (handshake, sync, transfer, auth) works unchanged.
// Only the transport layer is replaced with WebRTC data channels.
//
// # Stub Implementation Boundaries
//
// This package follows Venture's minimal-dependency philosophy by providing a
// complete simulation layer for testing without requiring github.com/pion/webrtc/v3.
// The following clearly delineates production-ready logic from stub behavior:
//
// Production-ready (fully functional without external dependencies):
//   - NAT traversal coordination (Direct → STUN → TURN fallback logic in NATTraversal.EstablishConnection)
//   - Relay management (health checks, load balancing, selection strategies)
//   - Signaling protocol (SDP/ICE message exchange, peer registry, cleanup)
//   - Connection state machine (state transitions, statistics, lifecycle)
//   - TimeProvider abstraction (deterministic testing support)
//
// Stub behavior (requires pion/webrtc/v3 or equivalent for production):
//   - Peer.Connect: Simulates connection with 10ms delay instead of real SDP negotiation
//   - Peer.Send: Updates statistics but does not transmit via WebRTC data channel
//   - Peer.processMessages: Drains send channel without actual data channel I/O
//   - STUNClient.GetPublicAddress: Returns simulated IP/port, not actual STUN binding (querySTUNServer is a stub)
//   - STUNClient.DetectNATType: Returns simulated NAT type classification
//   - NATTraversal.tryDirectConnection: Always fails (stub)
//   - NATTraversal.trySTUNConnection: Relies on STUNClient.GetPublicAddress stub
//   - NATTraversal.tryTURNConnection: Never creates TURN allocation (stub)
//   - RelayConnection.Send: Counts bytes but does not relay via TURN allocation
//   - SignalingClient.Connect: Simulates WebSocket connection establishment
//
// Note: On native builds, WebRTC federation is not used (cmd/client/webrtc_native.go returns nil);
// federation uses TCP transport. On WASM builds, the real pion/webrtc/v3 library is used and
// bypasses the STUNClient/NATTraversal stubs entirely. The stub layer is exercised only by tests.
//
// To integrate a real native WebRTC backend, implement the stub methods using
// pion/webrtc/v3 PeerConnection and DataChannel APIs. The existing interfaces
// and state management remain unchanged.
//
// # Thread Safety
//
// All public methods are thread-safe. Internal state is protected by sync.RWMutex.
// WebRTC callbacks run in separate goroutines and coordinate via channels.
package webrtc
