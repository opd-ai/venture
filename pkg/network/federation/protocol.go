// Package federation provides server-to-server federation protocol.
package federation

import (
	"encoding/json"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/opd-ai/venture/pkg/engine"
	"github.com/opd-ai/venture/pkg/version"
	log "github.com/sirupsen/logrus"
)

const (
	// DefaultConnectionTimeout is the timeout for establishing TCP connections
	DefaultConnectionTimeout = 10 * time.Second
)

// DefaultProtocolVersion returns the current federation protocol version from the version package.
// This ensures protocol version is centrally managed and consistent across the codebase.
var DefaultProtocolVersion = version.ProtocolVersion

// DefaultProtocolFeatures lists the features supported by this server
var DefaultProtocolFeatures = []string{"travel", "trade", "post"}

// FederationProtocol handles server-to-server communication
type FederationProtocol struct {
	serverID       string
	peers          map[string]*PeerServer
	connections    map[string]net.Conn
	mu             sync.RWMutex
	identity       *ServerIdentity
	handshake      *HandshakeManager
	connectionPool *ConnectionPool
	health         *FederationHealth
	retryStrategy  *RetryStrategy
}

// PeerServer represents a connected peer server
type PeerServer struct {
	ID            string
	Name          string
	Address       string
	TrustLevel    string
	Connected     bool
	LastHeartbeat int64
}

// NewFederationProtocol creates a new federation protocol handler
func NewFederationProtocol(serverID string, identity *ServerIdentity) *FederationProtocol {
	return &FederationProtocol{
		serverID:       serverID,
		peers:          make(map[string]*PeerServer),
		connections:    make(map[string]net.Conn),
		identity:       identity,
		handshake:      NewHandshakeManager(identity),
		connectionPool: NewConnectionPool(DefaultConnectionConfig()),
		health:         NewFederationHealth(),
		retryStrategy:  NewRetryStrategy(DefaultRetryConfig()),
	}
}

// Connect establishes connection to a peer server with retry logic and circuit breaker
func (f *FederationProtocol) Connect(peerAddress string) error {
	if err := f.validateConnectionEligibility(peerAddress); err != nil {
		return err
	}

	conn, handshakeResponse, err := f.establishConnectionWithRetry(peerAddress)
	if err != nil {
		return fmt.Errorf("failed to connect to %s after retries: %w", peerAddress, err)
	}

	f.registerPeerConnection(peerAddress, conn, handshakeResponse)
	return nil
}

// validateConnectionEligibility checks if the connection can be established.
func (f *FederationProtocol) validateConnectionEligibility(peerAddress string) error {
	if f.health.IsLocalOnly() {
		return fmt.Errorf("federation is in local-only mode, cannot connect to remote servers")
	}

	f.mu.Lock()
	_, exists := f.connections[peerAddress]
	f.mu.Unlock()

	if exists {
		return fmt.Errorf("already connected to %s", peerAddress)
	}

	return nil
}

// establishConnectionWithRetry attempts to connect and handshake with retry logic.
func (f *FederationProtocol) establishConnectionWithRetry(peerAddress string) (net.Conn, FederationHandshake, error) {
	var conn net.Conn
	var handshakeResponse FederationHandshake

	err := f.retryStrategy.Execute(func() error {
		var attemptErr error
		conn, handshakeResponse, attemptErr = f.attemptConnection(peerAddress)
		return attemptErr
	}, IsNetworkError)

	return conn, handshakeResponse, err
}

// attemptConnection performs a single connection and handshake attempt.
func (f *FederationProtocol) attemptConnection(peerAddress string) (net.Conn, FederationHandshake, error) {
	conn, err := f.dialPeer(peerAddress)
	if err != nil {
		return nil, FederationHandshake{}, err
	}

	handshakeResponse, err := f.performHandshake(conn)
	if err != nil {
		conn.Close()
		return nil, FederationHandshake{}, err
	}

	return conn, handshakeResponse, nil
}

// dialPeer establishes a TCP connection to the peer server.
func (f *FederationProtocol) dialPeer(peerAddress string) (net.Conn, error) {
	conn, err := net.DialTimeout("tcp", peerAddress, DefaultConnectionTimeout)
	if err != nil {
		f.health.RecordFailure()
		return nil, err
	}
	return conn, nil
}

// performHandshake exchanges handshake messages with the peer.
func (f *FederationProtocol) performHandshake(conn net.Conn) (FederationHandshake, error) {
	handshakeMsg, err := f.identity.CreateHandshake(DefaultProtocolVersion, DefaultProtocolFeatures, TrustVerified)
	if err != nil {
		f.health.RecordFailure()
		return FederationHandshake{}, fmt.Errorf("failed to create handshake: %w", err)
	}

	if err := f.sendHandshake(conn, *handshakeMsg); err != nil {
		return FederationHandshake{}, err
	}

	handshakeResponse, err := f.receiveHandshake(conn)
	if err != nil {
		return FederationHandshake{}, err
	}

	if err := f.handshake.ProcessHandshake(&handshakeResponse); err != nil {
		f.health.RecordFailure()
		return FederationHandshake{}, fmt.Errorf("invalid handshake response: %w", err)
	}

	return handshakeResponse, nil
}

// sendHandshake sends the handshake message to the peer.
func (f *FederationProtocol) sendHandshake(conn net.Conn, handshakeMsg FederationHandshake) error {
	encoder := json.NewEncoder(conn)
	if err := encoder.Encode(handshakeMsg); err != nil {
		f.health.RecordFailure()
		return fmt.Errorf("failed to send handshake: %w", err)
	}
	return nil
}

// receiveHandshake receives and decodes the handshake response from the peer.
func (f *FederationProtocol) receiveHandshake(conn net.Conn) (FederationHandshake, error) {
	var handshakeResponse FederationHandshake
	decoder := json.NewDecoder(conn)
	if err := decoder.Decode(&handshakeResponse); err != nil {
		f.health.RecordFailure()
		return FederationHandshake{}, fmt.Errorf("failed to receive handshake response: %w", err)
	}
	return handshakeResponse, nil
}

// registerPeerConnection records the successful connection in federation state.
func (f *FederationProtocol) registerPeerConnection(peerAddress string, conn net.Conn, handshakeResponse FederationHandshake) {
	f.health.RecordSuccess()

	f.mu.Lock()
	defer f.mu.Unlock()

	f.connections[peerAddress] = conn
	f.connectionPool.Add(handshakeResponse.ServerID, peerAddress, conn)

	f.peers[handshakeResponse.ServerID] = &PeerServer{
		ID:            handshakeResponse.ServerID,
		Name:          handshakeResponse.ServerName,
		Address:       peerAddress,
		TrustLevel:    handshakeResponse.TrustLevel.String(),
		Connected:     true,
		LastHeartbeat: time.Now().Unix(),
	}
}

// TransferPlayer initiates player transfer to another server
func (f *FederationProtocol) TransferPlayer(playerID uint64, world *engine.World, targetServer string, authMgr *AuthManager, transferMgr *TransferManager) error {
	token, transfer, err := f.preparePlayerTransfer(playerID, world, targetServer, authMgr, transferMgr)
	if err != nil {
		return err
	}

	conn, exists := f.getTargetConnection(targetServer)
	if !exists {
		return nil
	}

	if err := f.executeTransferRequest(playerID, token, transfer, conn, transferMgr); err != nil {
		return err
	}

	return transferMgr.ConfirmTransfer(playerID)
}

// preparePlayerTransfer creates auth token and prepares transfer state.
func (f *FederationProtocol) preparePlayerTransfer(playerID uint64, world *engine.World, targetServer string, authMgr *AuthManager, transferMgr *TransferManager) (string, *PlayerTransfer, error) {
	token, err := authMgr.CreateSessionToken(playerID, f.serverID)
	if err != nil {
		return "", nil, fmt.Errorf("failed to create session token: %w", err)
	}

	transfer, err := transferMgr.PrepareTransfer(playerID, world, targetServer)
	if err != nil {
		return "", nil, fmt.Errorf("failed to prepare transfer: %w", err)
	}

	if err := f.validateAndBeginTransfer(playerID, transfer, transferMgr); err != nil {
		return "", nil, err
	}

	return token.Token, transfer, nil
}

// validateAndBeginTransfer validates player state and begins the transfer.
func (f *FederationProtocol) validateAndBeginTransfer(playerID uint64, transfer *PlayerTransfer, transferMgr *TransferManager) error {
	if err := transferMgr.ValidatePlayerState(transfer.PlayerState); err != nil {
		transferMgr.RollbackTransfer(playerID, "validation failed")
		return fmt.Errorf("invalid player state: %w", err)
	}

	if err := transferMgr.BeginTransfer(playerID, f.serverID); err != nil {
		transferMgr.RollbackTransfer(playerID, "begin transfer failed")
		return fmt.Errorf("failed to begin transfer: %w", err)
	}

	return nil
}

// getTargetConnection retrieves the connection to the target server.
func (f *FederationProtocol) getTargetConnection(targetServer string) (net.Conn, bool) {
	f.mu.RLock()
	conn, exists := f.connections[targetServer]
	f.mu.RUnlock()
	return conn, exists
}

// executeTransferRequest sends the transfer request and handles response.
func (f *FederationProtocol) executeTransferRequest(playerID uint64, token string, transfer *PlayerTransfer, conn net.Conn, transferMgr *TransferManager) error {
	transferReq := f.buildTransferRequest(playerID, token, transfer)

	if err := f.sendTransferRequest(transferReq, conn, playerID, transferMgr); err != nil {
		return err
	}

	return f.receiveTransferResponse(conn, playerID, transferMgr)
}

// buildTransferRequest creates the transfer request payload.
func (f *FederationProtocol) buildTransferRequest(playerID uint64, token string, transfer *PlayerTransfer) map[string]interface{} {
	return map[string]interface{}{
		"type":         "transfer_request",
		"player_id":    playerID,
		"player_state": transfer.PlayerState,
		"token":        token,
		"timestamp":    time.Now().Unix(),
	}
}

// sendTransferRequest encodes and sends the transfer request.
func (f *FederationProtocol) sendTransferRequest(transferReq map[string]interface{}, conn net.Conn, playerID uint64, transferMgr *TransferManager) error {
	encoder := json.NewEncoder(conn)
	if err := encoder.Encode(transferReq); err != nil {
		transferMgr.RollbackTransfer(playerID, "failed to send transfer request")
		return fmt.Errorf("failed to send transfer request: %w", err)
	}
	return nil
}

// receiveTransferResponse decodes and validates the transfer response.
func (f *FederationProtocol) receiveTransferResponse(conn net.Conn, playerID uint64, transferMgr *TransferManager) error {
	decoder := json.NewDecoder(conn)
	var response map[string]interface{}
	if err := decoder.Decode(&response); err != nil {
		transferMgr.RollbackTransfer(playerID, "failed to receive transfer response")
		return fmt.Errorf("failed to receive transfer response: %w", err)
	}

	if status, ok := response["status"].(string); !ok || status != "accepted" {
		reason := f.extractRejectionReason(response)
		transferMgr.RollbackTransfer(playerID, reason)
		return fmt.Errorf("transfer rejected: %s", reason)
	}

	return nil
}

// extractRejectionReason extracts the rejection reason from response.
func (f *FederationProtocol) extractRejectionReason(response map[string]interface{}) string {
	if r, ok := response["reason"].(string); ok {
		return r
	}
	return "unknown"
}

// PortalSystem manages cross-server portals
type PortalSystem struct {
	world       *engine.World
	federation  *FederationProtocol
	authMgr     *AuthManager
	transferMgr *TransferManager
}

// NewPortalSystem creates a new portal system
func NewPortalSystem(world *engine.World, federation *FederationProtocol) *PortalSystem {
	return &PortalSystem{
		world:      world,
		federation: federation,
	}
}

// SetManagers configures the auth and transfer managers needed for portal activation.
// Must be called before Update() can trigger automatic portal transfers.
func (s *PortalSystem) SetManagers(authMgr *AuthManager, transferMgr *TransferManager) {
	s.authMgr = authMgr
	s.transferMgr = transferMgr
}

// Update processes portal interactions. When a player is within activation
// range (distance < 2.0) and the portal does not require manual activation,
// the portal is activated automatically.
func (s *PortalSystem) Update(deltaTime float64) {
	portals := s.world.GetEntitiesWith("portal", "position")
	players := s.world.GetEntitiesWith("player", "position")

	for _, portal := range portals {
		portalCompRaw, ok := portal.GetComponent("portal")
		if !ok {
			continue
		}
		portalComp := portalCompRaw.(*engine.PortalComponent)

		portalPosRaw, ok := portal.GetComponent("position")
		if !ok {
			continue
		}
		portalPos := portalPosRaw.(*engine.PositionComponent)

		for _, player := range players {
			playerPosRaw, ok := player.GetComponent("position")
			if !ok {
				continue
			}
			playerPos := playerPosRaw.(*engine.PositionComponent)

			dx := playerPos.X - portalPos.X
			dy := playerPos.Y - portalPos.Y
			distSq := dx*dx + dy*dy

			if distSq < 4.0 {
				if portalComp.RequiresActivation {
					// Portal requires manual activation via ActivatePortal
					continue
				}
				// Auto-activate portal for players in range
				if s.authMgr == nil || s.transferMgr == nil {
					continue
				}
				if err := s.ActivatePortal(player.ID, portal.ID, s.authMgr, s.transferMgr); err != nil {
					log.WithFields(log.Fields{
						"player_id": player.ID,
						"portal_id": portal.ID,
						"error":     err.Error(),
					}).Warn("Auto portal activation failed")
				}
			}
		}
	}
}

// ActivatePortal transfers a player through a portal
func (s *PortalSystem) ActivatePortal(playerID, portalID uint64, authMgr *AuthManager, transferMgr *TransferManager) error {
	portal, ok := s.world.GetEntity(portalID)
	if !ok || portal == nil {
		return fmt.Errorf("portal not found")
	}

	portalCompRaw, ok := portal.GetComponent("portal")
	if !ok {
		return fmt.Errorf("entity is not a portal")
	}

	portalComp := portalCompRaw.(*engine.PortalComponent)

	// Check if player meets requirements
	if portalComp.RequiredItem != "" {
		if err := s.checkRequiredItem(playerID, portalComp.RequiredItem); err != nil {
			return fmt.Errorf("portal requirement not met: %w", err)
		}
	}

	if portalComp.DestinationServer == "local" {
		// Local teleport
		return s.localTeleport(playerID, portalComp.DestinationX, portalComp.DestinationY)
	}

	// Cross-server transfer
	return s.federation.TransferPlayer(playerID, s.world, portalComp.DestinationServer, authMgr, transferMgr)
}

func (s *PortalSystem) localTeleport(playerID uint64, destX, destY float64) error {
	player, ok := s.world.GetEntity(playerID)
	if !ok || player == nil {
		return fmt.Errorf("player not found")
	}

	posCompRaw, ok := player.GetComponent("position")
	if !ok {
		return fmt.Errorf("player has no position")
	}

	posComp := posCompRaw.(*engine.PositionComponent)
	posComp.X = destX
	posComp.Y = destY

	return nil
}

func (s *PortalSystem) checkRequiredItem(playerID uint64, itemName string) error {
	player, ok := s.world.GetEntity(playerID)
	if !ok || player == nil {
		return fmt.Errorf("player not found")
	}

	invCompRaw, ok := player.GetComponent("inventory")
	if !ok {
		return fmt.Errorf("player has no inventory")
	}

	inv := invCompRaw.(*engine.InventoryComponent)
	for _, item := range inv.Items {
		if item.Name == itemName {
			return nil
		}
	}

	return fmt.Errorf("required item not found: %s", itemName)
}

// GuildUpdateMessage represents a guild state update for cross-server sync
type GuildUpdateMessage struct {
	Type      string `json:"type"`       // "guild_update"
	GuildID   string `json:"guild_id"`   // Guild identifier
	GuildData []byte `json:"guild_data"` // Serialized guild state (gzip-compressed JSON)
	Timestamp int64  `json:"timestamp"`  // Update timestamp for conflict resolution
}

// BroadcastGuildUpdate broadcasts a guild update to all connected peer servers
func (f *FederationProtocol) BroadcastGuildUpdate(guildID string, guildData []byte) error {
	f.mu.RLock()
	defer f.mu.RUnlock()

	if len(f.connections) == 0 {
		return nil
	}

	msg := GuildUpdateMessage{
		Type:      "guild_update",
		GuildID:   guildID,
		GuildData: guildData,
		Timestamp: time.Now().Unix(),
	}

	for peerAddr, conn := range f.connections {
		encoder := json.NewEncoder(conn)
		if err := encoder.Encode(msg); err != nil {
			f.mu.RUnlock()
			f.mu.Lock()
			conn.Close()
			delete(f.connections, peerAddr)
			f.mu.Unlock()
			f.mu.RLock()
			continue
		}
	}

	return nil
}

// ReceiveGuildUpdate processes an incoming guild update from a peer server
func (f *FederationProtocol) ReceiveGuildUpdate(msg *GuildUpdateMessage) error {
	if msg.Type != "guild_update" {
		return fmt.Errorf("invalid message type: %s", msg.Type)
	}

	if msg.GuildID == "" {
		return fmt.Errorf("missing guild ID")
	}

	if len(msg.GuildData) == 0 {
		return fmt.Errorf("missing guild data")
	}

	return nil
}

// GetHealth returns the current federation health tracker
func (f *FederationProtocol) GetHealth() *FederationHealth {
	return f.health
}

// SendGossip sends a gossip message to a specific peer server via its TCP/TLS connection.
// This implements the GossipTransport interface for integration with DiscoverySystem.
func (f *FederationProtocol) SendGossip(peerID string, msg *GossipMessage) error {
	f.mu.RLock()
	defer f.mu.RUnlock()

	peer, exists := f.peers[peerID]
	if !exists {
		return fmt.Errorf("unknown peer: %s", peerID)
	}

	conn, hasConn := f.connections[peer.Address]
	if !hasConn {
		return fmt.Errorf("no connection to peer %s at %s", peerID, peer.Address)
	}

	if err := json.NewEncoder(conn).Encode(msg); err != nil {
		return fmt.Errorf("failed to encode gossip for peer %s: %w", peerID, err)
	}

	return nil
}

// GetConnectionPool returns the connection pool
func (f *FederationProtocol) GetConnectionPool() *ConnectionPool {
	return f.connectionPool
}

// Close closes all connections and cleans up resources
func (f *FederationProtocol) Close() error {
	f.mu.Lock()
	defer f.mu.Unlock()

	// Close all direct connections
	for addr, conn := range f.connections {
		conn.Close()
		delete(f.connections, addr)
	}

	// Close connection pool
	if f.connectionPool != nil {
		f.connectionPool.Close()
	}

	if f.handshake != nil {
		f.handshake.Close()
	}

	// Clear peers
	f.peers = make(map[string]*PeerServer)

	return nil
}
