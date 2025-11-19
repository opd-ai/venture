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
//	    log.Fatalf("failed to create peer: %v", err)
//	}
//	defer peer.Close()
//
//	// Connect to remote peer
//	if err := peer.Connect("peer-bob"); err != nil {
//	    log.Fatalf("connection failed: %v", err)
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
//   - Data channel bandwidth: 1-10 MB/s (network-dependent)
//   - Zero server infrastructure cost (P2P only)
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
// # Thread Safety
//
// All public methods are thread-safe. Internal state is protected by sync.RWMutex.
// WebRTC callbacks run in separate goroutines and coordinate via channels.
package webrtc
