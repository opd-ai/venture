package guild

import (
	"encoding/json"
	"fmt"
	"time"
)

// Cross-server guild federation and synchronization.
//
// This file implements the federation protocol for syncing guild state across
// multiple game servers. Includes server registration, guild state broadcasting,
// and message handling for member changes and territory updates.
//
// Note: Actual network transport integration is pending. Messages are prepared
// but not yet broadcast (see ARCHITECTURE NOTE in SyncGuildState).
//
// Code relocated from: manager.go

// SetServerID sets this manager's server ID for federation
// Originally defined in: manager.go
func (m *Manager) SetServerID(serverID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.serverID = serverID
}

// AddFederatedServer registers a federated server for guild synchronization
// Originally defined in: manager.go
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
// Originally defined in: manager.go
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
// Originally defined in: manager.go
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

	// ARCHITECTURE NOTE: Guild message broadcast over federation network
	// This method currently creates the guild sync message but does not broadcast it.
	// Actual broadcast requires integration with the federation transport layer which
	// handles authenticated connections and message routing between federated servers.
	//
	// Implementation roadmap:
	// 1. Federation transport layer must provide a Broadcast(message) method
	// 2. Manager should accept a FederationTransport dependency in constructor
	// 3. Guild messages should be serialized and broadcast via the transport layer
	// 4. Transport layer handles encryption, authentication, and reliable delivery
	//
	// Current behavior: The method prepares guild sync messages for broadcast but
	// returns success without transmission. This allows guild state to be managed
	// locally while the federation transport layer is being developed. When the
	// transport layer is ready, this will be a one-line change to call
	// m.transport.Broadcast(msg) instead of discarding it.
	_ = msg // Guild message prepared but not yet broadcast

	return nil
}

// HandleGuildMessage processes incoming guild federation messages
// Originally defined in: manager.go
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
// Originally defined in: manager.go
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
// Originally defined in: manager.go
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
// Originally defined in: manager.go
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
// Originally defined in: manager.go
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
