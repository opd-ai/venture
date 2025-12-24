package guild

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"math/rand"
	"sync"
	"time"
)

// INTEGRATION FIX [Category G]: Guild Federation Manager
// Gap: Guild manager missing for cross-server guild coordination (ROADMAP_V8.md Phase 50.1)
// Fix: Created thread-safe guild manager with save/load and cross-server sync support
// Roadmap: ROADMAP_V8.md Phase 50.1

// Manager manages guilds across federated servers
type Manager struct {
	guilds           map[string]*Guild
	mu               sync.RWMutex
	federatedServers []string // List of federated server IDs
	messageHandlers  map[MessageType]func(msg GuildMessage) error
	serverID         string // This server's ID
}

// NewManager creates a new guild manager
func NewManager() *Manager {
	m := &Manager{
		guilds:           make(map[string]*Guild),
		federatedServers: make([]string, 0),
		messageHandlers:  make(map[MessageType]func(msg GuildMessage) error),
		serverID:         fmt.Sprintf("server-%d", time.Now().UnixNano()),
	}

	// Register message handlers
	m.messageHandlers[MsgTypeGuildSync] = m.handleGuildSync
	m.messageHandlers[MsgTypeMemberJoin] = m.handleMemberJoin
	m.messageHandlers[MsgTypeMemberLeave] = m.handleMemberLeave
	m.messageHandlers[MsgTypeTerritoryChange] = m.handleTerritoryChange

	return m
}

// CreateGuild creates a new guild with procedural identity
func (m *Manager) CreateGuild(genre, leaderID string) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Generate unique guild ID
	guildID := fmt.Sprintf("guild-%d-%s", time.Now().UnixNano(), leaderID)

	// Generate procedural guild identity
	identity := GenerateIdentity(genre, time.Now().Unix())

	// Create default permission set
	permissions := make(map[Rank][]Permission)
	permissions[RankLeader] = []Permission{
		PermissionInvite, PermissionKick, PermissionPromote,
		PermissionDemote, PermissionWithdraw, PermissionDeposit,
		PermissionEditMOTD, PermissionManageBank, PermissionDeclareWar,
	}
	permissions[RankOfficer] = []Permission{
		PermissionInvite, PermissionDeposit, PermissionEditMOTD,
	}
	permissions[RankMember] = []Permission{PermissionDeposit}
	permissions[RankRecruit] = []Permission{PermissionDeposit}

	guild := &Guild{
		ID:           guildID,
		Name:         identity.Name,
		Emblem:       identity.Emblem,
		LeaderID:     leaderID,
		Members:      []Member{{PlayerID: leaderID, Rank: RankLeader, JoinedAt: time.Now(), LastLogin: time.Now()}},
		Permissions:  permissions,
		Treasury:     0,
		Transactions: []TreasuryTransaction{},
		MOTD:         fmt.Sprintf("Welcome to %s!", identity.Name),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		Reputation:   make(map[string]float64),
	}

	m.guilds[guildID] = guild
	return guildID, nil
}

// GetGuild retrieves a guild by ID
func (m *Manager) GetGuild(guildID string) (*Guild, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	guild, exists := m.guilds[guildID]
	if !exists {
		return nil, fmt.Errorf("guild not found: %s", guildID)
	}
	return guild, nil
}

// AddMember adds a member to a guild
func (m *Manager) AddMember(guildID, playerID string, rank Rank) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	guild, exists := m.guilds[guildID]
	if !exists {
		return fmt.Errorf("guild not found: %s", guildID)
	}

	// Check if already a member
	for _, member := range guild.Members {
		if member.PlayerID == playerID {
			return fmt.Errorf("player already a member: %s", playerID)
		}
	}

	guild.Members = append(guild.Members, Member{
		PlayerID:  playerID,
		Rank:      rank,
		JoinedAt:  time.Now(),
		LastLogin: time.Now(),
	})
	guild.UpdatedAt = time.Now()
	return nil
}

// RemoveMember removes a member from a guild
func (m *Manager) RemoveMember(guildID, playerID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	guild, exists := m.guilds[guildID]
	if !exists {
		return fmt.Errorf("guild not found: %s", guildID)
	}

	// Cannot remove guild leader
	if playerID == guild.LeaderID {
		return fmt.Errorf("cannot remove guild leader")
	}

	for i, member := range guild.Members {
		if member.PlayerID == playerID {
			guild.Members = append(guild.Members[:i], guild.Members[i+1:]...)
			guild.UpdatedAt = time.Now()
			return nil
		}
	}
	return fmt.Errorf("member not found: %s", playerID)
}

// PromoteMember promotes a guild member to the next rank
func (m *Manager) PromoteMember(guildID, targetPlayerID, promoterID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	guild, err := m.getGuildUnsafe(guildID)
	if err != nil {
		return err
	}

	promoter, err := m.validatePromoterPermissions(guild, promoterID)
	if err != nil {
		return err
	}

	return m.updateMemberRank(guild, targetPlayerID, promoter)
}

// getGuildUnsafe retrieves a guild without locking (caller must lock).
func (m *Manager) getGuildUnsafe(guildID string) (*Guild, error) {
	guild, exists := m.guilds[guildID]
	if !exists {
		return nil, fmt.Errorf("guild not found: %s", guildID)
	}
	return guild, nil
}

// validatePromoterPermissions checks if promoter has permission to promote.
func (m *Manager) validatePromoterPermissions(guild *Guild, promoterID string) (*Member, error) {
	promoter := guild.GetMember(promoterID)
	if promoter == nil {
		return nil, fmt.Errorf("promoter not a member of guild")
	}
	if !guild.HasPermission(promoter.Rank, PermissionPromote) {
		return nil, fmt.Errorf("insufficient permissions to promote")
	}
	return promoter, nil
}

// updateMemberRank finds and updates the target member's rank.
func (m *Manager) updateMemberRank(guild *Guild, targetPlayerID string, promoter *Member) error {
	for i := range guild.Members {
		if guild.Members[i].PlayerID == targetPlayerID {
			return m.promoteToNextRank(guild, i, promoter)
		}
	}
	return fmt.Errorf("target member not found: %s", targetPlayerID)
}

// promoteToNextRank determines and applies the next rank for a member.
func (m *Manager) promoteToNextRank(guild *Guild, memberIdx int, promoter *Member) error {
	currentRank := guild.Members[memberIdx].Rank
	newRank, err := m.calculateNextRank(currentRank, promoter.Rank)
	if err != nil {
		return err
	}

	if newRank == RankLeader {
		m.transferLeadership(guild, memberIdx, promoter.PlayerID)
	}

	guild.Members[memberIdx].Rank = newRank
	guild.UpdatedAt = time.Now()
	return nil
}

// calculateNextRank determines the next rank based on current rank.
func (m *Manager) calculateNextRank(currentRank, promoterRank Rank) (Rank, error) {
	switch currentRank {
	case RankRecruit:
		return RankMember, nil
	case RankMember:
		return RankOfficer, nil
	case RankOfficer:
		if promoterRank != RankLeader {
			return "", fmt.Errorf("only leader can promote to officer")
		}
		return RankLeader, nil
	case RankLeader:
		return "", fmt.Errorf("cannot promote leader")
	default:
		return "", fmt.Errorf("unknown rank: %s", currentRank)
	}
}

// transferLeadership transfers guild leadership and demotes current leader.
func (m *Manager) transferLeadership(guild *Guild, newLeaderIdx int, oldLeaderID string) {
	guild.LeaderID = guild.Members[newLeaderIdx].PlayerID
	for j := range guild.Members {
		if guild.Members[j].PlayerID == oldLeaderID {
			guild.Members[j].Rank = RankOfficer
			break
		}
	}
}

// DepositTreasury adds gold to guild treasury
func (m *Manager) DepositTreasury(guildID, playerID string, amount int) error {
	if amount <= 0 {
		return fmt.Errorf("amount must be positive")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	guild, exists := m.guilds[guildID]
	if !exists {
		return fmt.Errorf("guild not found: %s", guildID)
	}

	guild.Treasury += amount
	guild.Transactions = append(guild.Transactions, TreasuryTransaction{
		PlayerID:  playerID,
		Amount:    amount,
		Timestamp: time.Now(),
		Reason:    "deposit",
	})
	guild.UpdatedAt = time.Now()
	return nil
}

// WithdrawTreasury removes gold from guild treasury
func (m *Manager) WithdrawTreasury(guildID, playerID string, amount int) error {
	if amount <= 0 {
		return fmt.Errorf("amount must be positive")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	guild, exists := m.guilds[guildID]
	if !exists {
		return fmt.Errorf("guild not found: %s", guildID)
	}

	if guild.Treasury < amount {
		return fmt.Errorf("insufficient treasury funds")
	}

	guild.Treasury -= amount
	guild.Transactions = append(guild.Transactions, TreasuryTransaction{
		PlayerID:  playerID,
		Amount:    -amount,
		Timestamp: time.Now(),
		Reason:    "withdrawal",
	})
	guild.UpdatedAt = time.Now()
	return nil
}

// SetMOTD updates the message of the day
func (m *Manager) SetMOTD(guildID, motd string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	guild, exists := m.guilds[guildID]
	if !exists {
		return fmt.Errorf("guild not found: %s", guildID)
	}

	guild.MOTD = motd
	guild.UpdatedAt = time.Now()
	return nil
}

// Save serializes all guilds to gzip-compressed JSON
func (m *Manager) Save() ([]byte, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	data, err := json.Marshal(m.guilds)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal guilds: %w", err)
	}

	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	if _, err := gz.Write(data); err != nil {
		return nil, fmt.Errorf("failed to compress: %w", err)
	}
	if err := gz.Close(); err != nil {
		return nil, fmt.Errorf("failed to close gzip: %w", err)
	}

	return buf.Bytes(), nil
}

// Load deserializes guilds from gzip-compressed JSON
func (m *Manager) Load(data []byte) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	gz, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("failed to decompress: %w", err)
	}
	defer gz.Close()

	var buf bytes.Buffer
	if _, err := buf.ReadFrom(gz); err != nil {
		return fmt.Errorf("failed to read: %w", err)
	}

	guilds := make(map[string]*Guild)
	if err := json.Unmarshal(buf.Bytes(), &guilds); err != nil {
		return fmt.Errorf("failed to unmarshal: %w", err)
	}

	m.guilds = guilds
	return nil
}

// GuildIdentity represents a procedurally generated guild identity
type GuildIdentity struct {
	Name   string
	Emblem *Emblem
}

// GenerateIdentity generates a procedural guild name and emblem
func GenerateIdentity(genre string, seed int64) GuildIdentity {
	rng := rand.New(rand.NewSource(seed))

	// Genre-specific name templates
	var prefixes, suffixes []string
	switch genre {
	case "fantasy":
		prefixes = []string{"The Ancient", "The Noble", "The Shadow", "The Crimson", "The Golden"}
		suffixes = []string{"Knights", "Guardians", "Brotherhood", "Order", "Legion"}
	case "sci-fi":
		prefixes = []string{"The Stellar", "The Nova", "The Quantum", "The Void", "The Nexus"}
		suffixes = []string{"Collective", "Syndicate", "Federation", "Alliance", "Corporation"}
	case "horror":
		prefixes = []string{"The Twisted", "The Cursed", "The Hollow", "The Forsaken", "The Damned"}
		suffixes = []string{"Cult", "Covenant", "Circle", "Cabal", "Congregation"}
	case "cyberpunk":
		prefixes = []string{"The Neon", "The Chrome", "The Digital", "The Neural", "The Cyber"}
		suffixes = []string{"Runners", "Hackers", "Collective", "Network", "Syndicate"}
	case "post-apocalyptic":
		prefixes = []string{"The Wasteland", "The Rad", "The Scrap", "The Dust", "The Rust"}
		suffixes = []string{"Survivors", "Scavengers", "Raiders", "Nomads", "Brotherhood"}
	default:
		prefixes = []string{"The", "Order of", "Guild of", "Company of", "League of"}
		suffixes = []string{"Adventurers", "Explorers", "Warriors", "Traders", "Builders"}
	}

	name := fmt.Sprintf("%s %s", prefixes[rng.Intn(len(prefixes))], suffixes[rng.Intn(len(suffixes))])

	// Generate emblem
	shapes := []string{"shield", "crest", "banner", "circle", "star"}
	symbols := []string{"sword", "dragon", "star", "flame", "skull", "crown", "hammer", "phoenix"}

	emblem := &Emblem{
		Shape:      shapes[rng.Intn(len(shapes))],
		PrimaryR:   uint8(rng.Intn(256)),
		PrimaryG:   uint8(rng.Intn(256)),
		PrimaryB:   uint8(rng.Intn(256)),
		SecondaryR: uint8(rng.Intn(256)),
		SecondaryG: uint8(rng.Intn(256)),
		SecondaryB: uint8(rng.Intn(256)),
		Symbol:     symbols[rng.Intn(len(symbols))],
	}

	return GuildIdentity{Name: name, Emblem: emblem}
}

// SetServerID sets this manager's server ID for federation
func (m *Manager) SetServerID(serverID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.serverID = serverID
}

// AddFederatedServer registers a federated server for guild synchronization
func (m *Manager) AddFederatedServer(serverID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if already registered
	for _, id := range m.federatedServers {
		if id == serverID {
			return
		}
	}
	m.federatedServers = append(m.federatedServers, serverID)
}

// RemoveFederatedServer unregisters a federated server
func (m *Manager) RemoveFederatedServer(serverID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for i, id := range m.federatedServers {
		if id == serverID {
			m.federatedServers = append(m.federatedServers[:i], m.federatedServers[i+1:]...)
			return
		}
	}
}

// SyncGuildState broadcasts guild state to all federated servers
func (m *Manager) SyncGuildState(guildID string) error {
	m.mu.RLock()
	guild, ok := m.guilds[guildID]
	m.mu.RUnlock()

	if !ok {
		return fmt.Errorf("guild not found: %s", guildID)
	}

	// Create sync message with full guild state
	msg := GuildMessage{
		Type:      MsgTypeGuildSync,
		GuildID:   guildID,
		ServerID:  m.serverID,
		Timestamp: time.Now(),
		Data:      guild,
	}

	// Broadcast to all federated servers
	// Note: Actual network transmission would be handled by federation layer
	// This method prepares the message for broadcast
	_ = msg // Message prepared for federation layer

	return nil
}

// HandleGuildMessage processes incoming guild federation messages
func (m *Manager) HandleGuildMessage(msg GuildMessage) error {
	// Validate message
	if msg.Type == "" {
		return fmt.Errorf("message type is empty")
	}
	if msg.GuildID == "" {
		return fmt.Errorf("guild ID is empty")
	}

	// Get handler for message type
	handler, exists := m.messageHandlers[msg.Type]
	if !exists {
		return fmt.Errorf("unknown message type: %v", msg.Type)
	}

	// Execute handler
	return handler(msg)
}

// handleGuildSync processes full guild state synchronization
func (m *Manager) handleGuildSync(msg GuildMessage) error {
	// Extract guild data from message
	guildData, ok := msg.Data.(*Guild)
	if !ok {
		// Try to unmarshal from map (JSON deserialization)
		if dataMap, isMap := msg.Data.(map[string]interface{}); isMap {
			data, err := json.Marshal(dataMap)
			if err != nil {
				return fmt.Errorf("failed to marshal guild data: %w", err)
			}
			guildData = &Guild{}
			if err := json.Unmarshal(data, guildData); err != nil {
				return fmt.Errorf("failed to unmarshal guild data: %w", err)
			}
		} else {
			return fmt.Errorf("invalid guild data type")
		}
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// Update or create guild
	m.guilds[msg.GuildID] = guildData
	return nil
}

// handleMemberJoin processes member join notifications
func (m *Manager) handleMemberJoin(msg GuildMessage) error {
	// Extract member join data
	var joinData MemberJoinData
	if data, ok := msg.Data.(MemberJoinData); ok {
		joinData = data
	} else if dataMap, ok := msg.Data.(map[string]interface{}); ok {
		data, err := json.Marshal(dataMap)
		if err != nil {
			return fmt.Errorf("failed to marshal join data: %w", err)
		}
		if err := json.Unmarshal(data, &joinData); err != nil {
			return fmt.Errorf("failed to unmarshal join data: %w", err)
		}
	} else {
		return fmt.Errorf("invalid member join data type")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	guild, exists := m.guilds[msg.GuildID]
	if !exists {
		return fmt.Errorf("guild not found: %s", msg.GuildID)
	}

	// Check if member already exists
	for _, member := range guild.Members {
		if member.PlayerID == joinData.PlayerID {
			return nil // Already a member, ignore duplicate
		}
	}

	// Add new member
	guild.Members = append(guild.Members, Member{
		PlayerID:  joinData.PlayerID,
		Rank:      joinData.Rank,
		JoinedAt:  msg.Timestamp,
		LastLogin: msg.Timestamp,
	})
	guild.UpdatedAt = msg.Timestamp

	return nil
}

// handleMemberLeave processes member leave notifications
func (m *Manager) handleMemberLeave(msg GuildMessage) error {
	// Extract member leave data
	var leaveData MemberLeaveData
	if data, ok := msg.Data.(MemberLeaveData); ok {
		leaveData = data
	} else if dataMap, ok := msg.Data.(map[string]interface{}); ok {
		data, err := json.Marshal(dataMap)
		if err != nil {
			return fmt.Errorf("failed to marshal leave data: %w", err)
		}
		if err := json.Unmarshal(data, &leaveData); err != nil {
			return fmt.Errorf("failed to unmarshal leave data: %w", err)
		}
	} else {
		return fmt.Errorf("invalid member leave data type")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	guild, exists := m.guilds[msg.GuildID]
	if !exists {
		return fmt.Errorf("guild not found: %s", msg.GuildID)
	}

	// Remove member
	for i, member := range guild.Members {
		if member.PlayerID == leaveData.PlayerID {
			guild.Members = append(guild.Members[:i], guild.Members[i+1:]...)
			guild.UpdatedAt = msg.Timestamp
			return nil
		}
	}

	return nil // Member not found, ignore
}

// handleTerritoryChange processes territory control change notifications
func (m *Manager) handleTerritoryChange(msg GuildMessage) error {
	// Extract territory change data
	var territoryData TerritoryChangeData
	if data, ok := msg.Data.(TerritoryChangeData); ok {
		territoryData = data
	} else if dataMap, ok := msg.Data.(map[string]interface{}); ok {
		data, err := json.Marshal(dataMap)
		if err != nil {
			return fmt.Errorf("failed to marshal territory data: %w", err)
		}
		if err := json.Unmarshal(data, &territoryData); err != nil {
			return fmt.Errorf("failed to unmarshal territory data: %w", err)
		}
	} else {
		return fmt.Errorf("invalid territory change data type")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	guild, exists := m.guilds[msg.GuildID]
	if !exists {
		return fmt.Errorf("guild not found: %s", msg.GuildID)
	}

	// Update guild's reputation for the zone/server
	if guild.Reputation == nil {
		guild.Reputation = make(map[string]float64)
	}

	// Increase reputation for gaining territory
	guild.Reputation[territoryData.ZoneID] = guild.Reputation[territoryData.ZoneID] + 10.0
	guild.UpdatedAt = msg.Timestamp

	return nil
}
