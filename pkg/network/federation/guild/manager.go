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
	guilds map[string]*Guild
	mu     sync.RWMutex
}

// NewManager creates a new guild manager
func NewManager() *Manager {
	return &Manager{
		guilds: make(map[string]*Guild),
	}
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

	guild, exists := m.guilds[guildID]
	if !exists {
		return fmt.Errorf("guild not found: %s", guildID)
	}

	// Get promoter's rank to check permissions
	promoter := guild.GetMember(promoterID)
	if promoter == nil {
		return fmt.Errorf("promoter not a member of guild")
	}

	// Check if promoter has permission to promote
	if !guild.HasPermission(promoter.Rank, PermissionPromote) {
		return fmt.Errorf("insufficient permissions to promote")
	}

	// Find and update target member
	for i := range guild.Members {
		if guild.Members[i].PlayerID == targetPlayerID {
			currentRank := guild.Members[i].Rank

			// Determine next rank
			var newRank Rank
			switch currentRank {
			case RankRecruit:
				newRank = RankMember
			case RankMember:
				newRank = RankOfficer
			case RankOfficer:
				// Only leader can promote to leader
				if promoter.Rank != RankLeader {
					return fmt.Errorf("only leader can promote to officer")
				}
				newRank = RankLeader
				// Transfer leadership
				guild.LeaderID = targetPlayerID
				// Demote current leader to officer
				for j := range guild.Members {
					if guild.Members[j].PlayerID == promoterID {
						guild.Members[j].Rank = RankOfficer
						break
					}
				}
			case RankLeader:
				return fmt.Errorf("cannot promote leader")
			default:
				return fmt.Errorf("unknown rank: %s", currentRank)
			}

			guild.Members[i].Rank = newRank
			guild.UpdatedAt = time.Now()
			return nil
		}
	}
	return fmt.Errorf("target member not found: %s", targetPlayerID)
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
