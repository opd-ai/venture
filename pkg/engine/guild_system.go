package engine

import (
	"fmt"
	"strconv"
	"time"

	"github.com/opd-ai/venture/pkg/network/federation/guild"
	"github.com/sirupsen/logrus"
)

// GuildSystem manages guild operations and cross-server synchronization
type GuildSystem struct {
	world      *World
	manager    *guild.Manager
	logger     *logrus.Entry
	federation FederationBroadcaster

	// Sync tracking
	lastSync       time.Time
	syncInterval   time.Duration
	pendingUpdates []string // Guild IDs needing sync
}

// NewGuildSystem creates a new guild management system
func NewGuildSystem(world *World, manager *guild.Manager) *GuildSystem {
	return &GuildSystem{
		world:          world,
		manager:        manager,
		logger:         logrus.WithField("system", "guild"),
		syncInterval:   time.Second * 5, // Sync every 5 seconds
		pendingUpdates: make([]string, 0),
	}
}

// SetFederation configures the federation protocol for cross-server sync
func (s *GuildSystem) SetFederation(federation FederationBroadcaster) {
	s.federation = federation
	s.logger.Info("federation protocol configured for guild sync")
}

// Update processes guild operations and synchronization
func (s *GuildSystem) Update(entities []*Entity, deltaTime float64) {
	// Check if it's time to sync
	now := time.Now()
	if now.Sub(s.lastSync) >= s.syncInterval {
		s.syncGuilds()
		s.lastSync = now
	}
}

// CreateGuild creates a new guild with the given leader
func (s *GuildSystem) CreateGuild(leaderEntity *Entity, genre string) error {
	// Verify entity has player component
	playerComp, ok := leaderEntity.GetComponent("player")
	if !ok || playerComp == nil {
		return fmt.Errorf("entity is not a player")
	}

	// Check if already in a guild
	guildComp, ok := leaderEntity.GetComponent("guild")
	if ok && guildComp != nil {
		gc := guildComp.(*GuildComponent)
		if gc.GuildID != "" {
			return fmt.Errorf("player already in a guild")
		}
	}

	// Create guild through federation manager
	guildID, err := s.manager.CreateGuild(genre, strconv.FormatUint(leaderEntity.ID, 10))
	if err != nil {
		return fmt.Errorf("failed to create guild: %w", err)
	}

	// Add guild component to leader
	leaderEntity.AddComponent(&GuildComponent{
		GuildID:  guildID,
		Rank:     string(guild.RankLeader),
		JoinedAt: time.Now().Unix(),
	})

	// Mark for sync
	s.markForSync(guildID)

	s.logger.WithFields(logrus.Fields{
		"guild_id": guildID,
		"leader":   leaderEntity.ID,
	}).Info("guild created")

	return nil
}

// JoinGuild adds a player to an existing guild
func (s *GuildSystem) JoinGuild(playerEntity *Entity, guildID string) error {
	// Verify entity has player component
	playerComp, ok := playerEntity.GetComponent("player")
	if !ok || playerComp == nil {
		return fmt.Errorf("entity is not a player")
	}

	// Check if already in a guild
	guildComp, ok := playerEntity.GetComponent("guild")
	if ok && guildComp != nil {
		gc := guildComp.(*GuildComponent)
		if gc.GuildID != "" {
			return fmt.Errorf("player already in guild: %s", gc.GuildID)
		}
	}

	// Verify guild exists
	targetGuild, err := s.manager.GetGuild(guildID)
	if err != nil {
		return fmt.Errorf("guild not found: %w", err)
	}

	// Add member to guild (as Recruit rank)
	err = s.manager.AddMember(guildID, strconv.FormatUint(playerEntity.ID, 10), guild.RankRecruit)
	if err != nil {
		return fmt.Errorf("failed to add member: %w", err)
	}

	// Add or update guild component
	if !ok || guildComp == nil {
		playerEntity.AddComponent(&GuildComponent{
			GuildID:  guildID,
			Rank:     string(guild.RankRecruit),
			JoinedAt: time.Now().Unix(),
		})
	} else {
		gc := guildComp.(*GuildComponent)
		gc.GuildID = guildID
		gc.Rank = string(guild.RankRecruit)
		gc.JoinedAt = time.Now().Unix()
	}

	// Mark for sync
	s.markForSync(guildID)

	s.logger.WithFields(logrus.Fields{
		"guild_id":   guildID,
		"guild_name": targetGuild.Name,
		"player":     playerEntity.ID,
	}).Info("player joined guild")

	return nil
}

// LeaveGuild removes a player from their guild
func (s *GuildSystem) LeaveGuild(playerEntity *Entity) error {
	guildComp, ok := playerEntity.GetComponent("guild")
	if !ok || guildComp == nil {
		return fmt.Errorf("player not in a guild")
	}

	gc := guildComp.(*GuildComponent)
	if gc.GuildID == "" {
		return fmt.Errorf("player not in a guild")
	}

	guildID := gc.GuildID

	// Remove from guild manager
	err := s.manager.RemoveMember(guildID, strconv.FormatUint(playerEntity.ID, 10))
	if err != nil {
		return fmt.Errorf("failed to remove member: %w", err)
	}

	// Clear guild component
	gc.GuildID = ""
	gc.Rank = ""
	gc.JoinedAt = 0

	// Mark for sync
	s.markForSync(guildID)

	s.logger.WithFields(logrus.Fields{
		"guild_id": guildID,
		"player":   playerEntity.ID,
	}).Info("player left guild")

	return nil
}

// PromoteMember promotes a guild member to a higher rank
func (s *GuildSystem) PromoteMember(promoterEntity *Entity, targetPlayerID string) error {
	// Get promoter's guild component
	guildComp, ok := promoterEntity.GetComponent("guild")
	if !ok || guildComp == nil {
		return fmt.Errorf("promoter not in a guild")
	}

	gc := guildComp.(*GuildComponent)
	guildID := gc.GuildID

	// Promote through manager (checks permissions internally)
	err := s.manager.PromoteMember(guildID, targetPlayerID, strconv.FormatUint(promoterEntity.ID, 10))
	if err != nil {
		return fmt.Errorf("failed to promote member: %w", err)
	}

	// Update target player's component if they're in this world
	entities := s.world.GetEntitiesWith("guild", "player")
	for _, entity := range entities {
		if strconv.FormatUint(entity.ID, 10) == targetPlayerID {
			targetGuildComp, ok := entity.GetComponent("guild")
			if !ok || targetGuildComp == nil {
				continue
			}
			gc := targetGuildComp.(*GuildComponent)

			// Get updated rank from manager
			targetGuild, _ := s.manager.GetGuild(guildID)
			member := targetGuild.GetMember(targetPlayerID)
			if member != nil {
				gc.Rank = string(member.Rank)
			}
			break
		}
	}

	// Mark for sync
	s.markForSync(guildID)

	return nil
}

// GetGuildInfo retrieves guild information
func (s *GuildSystem) GetGuildInfo(guildID string) (*guild.Guild, error) {
	return s.manager.GetGuild(guildID)
}

// DepositTreasury deposits gold into guild treasury
func (s *GuildSystem) DepositTreasury(playerEntity *Entity, amount int) error {
	guildComp, ok := playerEntity.GetComponent("guild")
	if !ok || guildComp == nil {
		return fmt.Errorf("player not in a guild")
	}

	gc := guildComp.(*GuildComponent)
	if gc.GuildID == "" {
		return fmt.Errorf("player not in a guild")
	}

	err := s.manager.DepositTreasury(gc.GuildID, strconv.FormatUint(playerEntity.ID, 10), amount)
	if err != nil {
		return fmt.Errorf("failed to deposit: %w", err)
	}

	// Mark for sync
	s.markForSync(gc.GuildID)

	return nil
}

// WithdrawTreasury withdraws gold from guild treasury
func (s *GuildSystem) WithdrawTreasury(playerEntity *Entity, amount int) error {
	guildComp, ok := playerEntity.GetComponent("guild")
	if !ok || guildComp == nil {
		return fmt.Errorf("player not in a guild")
	}

	gc := guildComp.(*GuildComponent)
	if gc.GuildID == "" {
		return fmt.Errorf("player not in a guild")
	}

	err := s.manager.WithdrawTreasury(gc.GuildID, strconv.FormatUint(playerEntity.ID, 10), amount)
	if err != nil {
		return fmt.Errorf("failed to withdraw: %w", err)
	}

	// Mark for sync
	s.markForSync(gc.GuildID)

	return nil
}

// markForSync adds a guild ID to pending sync updates
func (s *GuildSystem) markForSync(guildID string) {
	// Avoid duplicates
	for _, id := range s.pendingUpdates {
		if id == guildID {
			return
		}
	}
	s.pendingUpdates = append(s.pendingUpdates, guildID)
}

// syncGuilds synchronizes pending guild updates across federated servers
func (s *GuildSystem) syncGuilds() {
	if len(s.pendingUpdates) == 0 {
		return
	}

	s.logger.WithField("count", len(s.pendingUpdates)).Debug("syncing guilds")

	// Broadcast guild updates to federation
	for _, guildID := range s.pendingUpdates {
		guildObj, err := s.manager.GetGuild(guildID)
		if err != nil {
			s.logger.WithError(err).WithField("guild_id", guildID).Error("failed to get guild for sync")
			continue
		}

		// Serialize guild data for transmission
		guildData, err := s.manager.Save()
		if err != nil {
			s.logger.WithError(err).WithField("guild_id", guildID).Error("failed to serialize guild data")
			continue
		}

		// Broadcast to federated servers if available
		if s.federation != nil {
			if err := s.federation.BroadcastGuildUpdate(guildID, guildData); err != nil {
				s.logger.WithError(err).WithField("guild_id", guildID).Warn("failed to broadcast guild update")
				// Don't fail - guild state is still valid locally
			} else {
				s.logger.WithFields(logrus.Fields{
					"guild_id":   guildID,
					"guild_name": guildObj.Name,
					"data_size":  len(guildData),
				}).Debug("guild update broadcast to federation")
			}
		}
	}

	// Clear pending updates
	s.pendingUpdates = s.pendingUpdates[:0]
}

// LoadGuildData loads guild data from serialized format
// Used for cross-server sync and persistence
func (s *GuildSystem) LoadGuildData(data []byte) error {
	return s.manager.Load(data)
}

// SaveGuildData serializes all guild data
// Used for cross-server sync and persistence
func (s *GuildSystem) SaveGuildData() ([]byte, error) {
	return s.manager.Save()
}

// ApplyGuildUpdate receives and applies a guild update from a peer server
func (s *GuildSystem) ApplyGuildUpdate(guildID string, guildData []byte) error {
	s.logger.WithFields(logrus.Fields{
		"guild_id":  guildID,
		"data_size": len(guildData),
	}).Debug("applying guild update from federation")

	// Load guild data (merges with existing guilds)
	if err := s.manager.Load(guildData); err != nil {
		return fmt.Errorf("failed to load guild data: %w", err)
	}

	// Update local entities with new guild state
	guildObj, err := s.manager.GetGuild(guildID)
	if err != nil {
		return fmt.Errorf("failed to retrieve updated guild: %w", err)
	}

	// Update all local entities that are members of this guild
	entities := s.world.GetEntitiesWith("guild")
	for _, entity := range entities {
		guildComp, ok := entity.GetComponent("guild")
		if !ok || guildComp == nil {
			continue
		}

		gc := guildComp.(*GuildComponent)
		if gc.GuildID == guildID {
			// Find member in updated guild data
			playerID := strconv.FormatUint(entity.ID, 10)
			member := guildObj.GetMember(playerID)
			if member != nil {
				// Update rank if changed
				gc.Rank = string(member.Rank)
			}
		}
	}

	s.logger.WithFields(logrus.Fields{
		"guild_id":   guildID,
		"guild_name": guildObj.Name,
		"members":    len(guildObj.Members),
	}).Info("guild update applied from federation")

	return nil
}
