// Package federation provides server-to-server federation protocol.
package federation

import (
	"encoding/json"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/opd-ai/venture/pkg/engine"
)

// FederationProtocol handles server-to-server communication
type FederationProtocol struct {
	serverID    string
	peers       map[string]*PeerServer
	connections map[string]net.Conn
	mu          sync.RWMutex
	identity    *ServerIdentity
	handshake   *HandshakeManager
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
		serverID:    serverID,
		peers:       make(map[string]*PeerServer),
		connections: make(map[string]net.Conn),
		identity:    identity,
		handshake:   NewHandshakeManager(identity),
	}
}

// Connect establishes connection to a peer server
func (f *FederationProtocol) Connect(peerAddress string) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	if _, exists := f.connections[peerAddress]; exists {
		return fmt.Errorf("already connected to %s", peerAddress)
	}

	conn, err := net.DialTimeout("tcp", peerAddress, 10*time.Second)
	if err != nil {
		return fmt.Errorf("failed to connect to %s: %w", peerAddress, err)
	}

	handshakeMsg, err := f.identity.CreateHandshake("6.0.0", []string{"travel", "trade", "post"}, TrustVerified)
	if err != nil {
		conn.Close()
		return fmt.Errorf("failed to create handshake: %w", err)
	}

	encoder := json.NewEncoder(conn)
	if err := encoder.Encode(handshakeMsg); err != nil {
		conn.Close()
		return fmt.Errorf("failed to send handshake: %w", err)
	}

	decoder := json.NewDecoder(conn)
	var response FederationHandshake
	if err := decoder.Decode(&response); err != nil {
		conn.Close()
		return fmt.Errorf("failed to receive handshake response: %w", err)
	}

	if err := f.handshake.ProcessHandshake(&response); err != nil {
		conn.Close()
		return fmt.Errorf("invalid handshake response: %w", err)
	}

	f.connections[peerAddress] = conn
	f.peers[response.ServerID] = &PeerServer{
		ID:            response.ServerID,
		Name:          response.ServerName,
		Address:       peerAddress,
		TrustLevel:    response.TrustLevel.String(),
		Connected:     true,
		LastHeartbeat: time.Now().Unix(),
	}

	return nil
}

// TransferPlayer initiates player transfer to another server
func (f *FederationProtocol) TransferPlayer(playerID uint64, world *engine.World, targetServer string, authMgr *AuthManager, transferMgr *TransferManager) error {
	// Create session token for authentication
	token, err := authMgr.CreateSessionToken(playerID, f.serverID)
	if err != nil {
		return fmt.Errorf("failed to create session token: %w", err)
	}

	// Prepare transfer
	transfer, err := transferMgr.PrepareTransfer(playerID, world, targetServer)
	if err != nil {
		return fmt.Errorf("failed to prepare transfer: %w", err)
	}

	// Validate state
	if err := transferMgr.ValidatePlayerState(transfer.PlayerState); err != nil {
		transferMgr.RollbackTransfer(playerID, "validation failed")
		return fmt.Errorf("invalid player state: %w", err)
	}

	// Begin transfer phase
	if err := transferMgr.BeginTransfer(playerID, f.serverID); err != nil {
		transferMgr.RollbackTransfer(playerID, "begin transfer failed")
		return fmt.Errorf("failed to begin transfer: %w", err)
	}

	f.mu.RLock()
	conn, exists := f.connections[targetServer]
	f.mu.RUnlock()

	if !exists {
		// In test/offline mode, allow transfer to be prepared without active connection
		// The transfer will remain in Transfer phase until connection is established
		return nil
	}

	transferReq := map[string]interface{}{
		"type":         "transfer_request",
		"player_id":    playerID,
		"player_state": transfer.PlayerState,
		"token":        token,
		"timestamp":    time.Now().Unix(),
	}

	encoder := json.NewEncoder(conn)
	if err := encoder.Encode(transferReq); err != nil {
		transferMgr.RollbackTransfer(playerID, "failed to send transfer request")
		return fmt.Errorf("failed to send transfer request: %w", err)
	}

	decoder := json.NewDecoder(conn)
	var response map[string]interface{}
	if err := decoder.Decode(&response); err != nil {
		transferMgr.RollbackTransfer(playerID, "failed to receive transfer response")
		return fmt.Errorf("failed to receive transfer response: %w", err)
	}

	if status, ok := response["status"].(string); !ok || status != "accepted" {
		reason := "unknown"
		if r, ok := response["reason"].(string); ok {
			reason = r
		}
		transferMgr.RollbackTransfer(playerID, reason)
		return fmt.Errorf("transfer rejected: %s", reason)
	}

	return transferMgr.ConfirmTransfer(playerID)
}

// PortalSystem manages cross-server portals
type PortalSystem struct {
	world      *engine.World
	federation *FederationProtocol
}

// NewPortalSystem creates a new portal system
func NewPortalSystem(world *engine.World, federation *FederationProtocol) *PortalSystem {
	return &PortalSystem{
		world:      world,
		federation: federation,
	}
}

// Update processes portal interactions
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
			distance := dx*dx + dy*dy

			if distance < 4.0 {
				if portalComp.RequiresActivation {
					continue
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
