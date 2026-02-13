// Core guild management system.
//
// This file implements the Manager struct which coordinates all guild operations:
// - Guild creation with procedural identities
// - Member management (add, remove, promote)
// - Permission-based access control
// - Message of the day (MOTD) updates
//
// The Manager is thread-safe using RWMutex for concurrent access. Methods that
// modify state use exclusive locks, while read-only operations use shared locks.
//
// Specialized operations are delegated to other files:
// - treasury.go: Treasury deposit/withdrawal
// - federation.go: Cross-server sync
// - persistence.go: Save/load operations
// - identity.go: Procedural guild generation
package guild

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// ErrGuildDataSizeExceeded indicates the decompressed guild data exceeds MaxGuildDataSize.
// Originally defined in: manager.go
var ErrGuildDataSizeExceeded = errors.New("guild data exceeds maximum allowed size")

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
	logger           *logrus.Entry
	guildCounter     int64 // Counter for deterministic guild ID generation
}

// NewManager creates a new guild manager
func NewManager() *Manager {
	m := &Manager{
		guilds:           make(map[string]*Guild),
		federatedServers: make([]string, 0),
		messageHandlers:  make(map[MessageType]func(msg GuildMessage) error),
		serverID:         fmt.Sprintf("server-%s", uuid.New().String()),
		logger:           logrus.WithField("component", "guild_manager"),
		guildCounter:     0,
	}

	// Register message handlers
	m.messageHandlers[MsgTypeGuildSync] = m.handleGuildSync
	m.messageHandlers[MsgTypeMemberJoin] = m.handleMemberJoin
	m.messageHandlers[MsgTypeMemberLeave] = m.handleMemberLeave
	m.messageHandlers[MsgTypeTerritoryChange] = m.handleTerritoryChange

	return m
}

// CreateGuild creates a new guild with procedural identity.
// The seed parameter ensures deterministic guild ID and identity generation.
// Same seed + genre + leaderID produces the same guild identity.
func (m *Manager) CreateGuild(genre, leaderID string, seed int64) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Generate deterministic guild ID using seed and counter
	m.guildCounter++
	guildID := fmt.Sprintf("guild-%d-%s-%d", seed, leaderID, m.guildCounter)

	// Generate procedural guild identity using the provided seed
	identity := GenerateIdentity(genre, seed)

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

	m.logger.WithFields(logrus.Fields{
		"guild_id":  guildID,
		"leader_id": leaderID,
		"genre":     genre,
		"seed":      seed,
	}).Info("guild created")

	return guildID, nil
}

// GetGuild retrieves a guild by ID
func (m *Manager) GetGuild(guildID string) (*Guild, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	guild, exists := m.guilds[guildID]
	if !exists {
		m.logger.WithField("guild_id", guildID).Debug("guild not found")
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
		m.logger.WithFields(logrus.Fields{
			"guild_id":  guildID,
			"player_id": playerID,
		}).Warn("add member failed: guild not found")
		return fmt.Errorf("guild not found: %s", guildID)
	}

	// Check if already a member
	for _, member := range guild.Members {
		if member.PlayerID == playerID {
			m.logger.WithFields(logrus.Fields{
				"guild_id":  guildID,
				"player_id": playerID,
			}).Debug("add member failed: already a member")
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
		m.logger.WithFields(logrus.Fields{
			"guild_id":  guildID,
			"player_id": playerID,
		}).Warn("remove member failed: guild not found")
		return fmt.Errorf("guild not found: %s", guildID)
	}

	// Cannot remove guild leader
	if playerID == guild.LeaderID {
		m.logger.WithFields(logrus.Fields{
			"guild_id":  guildID,
			"player_id": playerID,
		}).Warn("remove member failed: cannot remove leader")
		return fmt.Errorf("cannot remove guild leader")
	}

	for i, member := range guild.Members {
		if member.PlayerID == playerID {
			guild.Members = append(guild.Members[:i], guild.Members[i+1:]...)
			guild.UpdatedAt = time.Now()
			return nil
		}
	}
	m.logger.WithFields(logrus.Fields{
		"guild_id":  guildID,
		"player_id": playerID,
	}).Debug("remove member failed: member not found")
	return fmt.Errorf("member not found: %s", playerID)
}

// PromoteMember promotes a guild member to the next rank
func (m *Manager) PromoteMember(guildID, targetPlayerID, promoterID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	guild, err := m.getGuildUnsafe(guildID)
	if err != nil {
		m.logger.WithFields(logrus.Fields{
			"guild_id":    guildID,
			"target_id":   targetPlayerID,
			"promoter_id": promoterID,
		}).Warn("promote member failed: guild not found")
		return err
	}

	promoter, err := m.validatePromoterPermissions(guild, promoterID)
	if err != nil {
		m.logger.WithFields(logrus.Fields{
			"guild_id":    guildID,
			"target_id":   targetPlayerID,
			"promoter_id": promoterID,
			"error":       err.Error(),
		}).Warn("promote member failed: permission validation error")
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

// SetMOTD updates the message of the day
func (m *Manager) SetMOTD(guildID, motd string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	guild, exists := m.guilds[guildID]
	if !exists {
		m.logger.WithField("guild_id", guildID).Warn("set MOTD failed: guild not found")
		return fmt.Errorf("guild not found: %s", guildID)
	}

	guild.MOTD = motd
	guild.UpdatedAt = time.Now()
	return nil
}
