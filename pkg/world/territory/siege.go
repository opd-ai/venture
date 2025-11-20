package territory

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

// SiegePhase represents the current phase of a siege.
type SiegePhase int

const (
	// PhasePreparation is the 1-hour preparation period before assault
	PhasePreparation SiegePhase = iota
	// PhaseAssault is the active combat phase (2 hours)
	PhaseAssault
	// PhaseResolution is the final calculation and loot distribution
	PhaseResolution
	// PhaseEnded indicates the siege has concluded
	PhaseEnded
)

// String returns the human-readable name for the siege phase.
func (sp SiegePhase) String() string {
	switch sp {
	case PhasePreparation:
		return "Preparation"
	case PhaseAssault:
		return "Assault"
	case PhaseResolution:
		return "Resolution"
	case PhaseEnded:
		return "Ended"
	default:
		return "Unknown"
	}
}

// VictoryCondition represents how a siege can be won.
type VictoryCondition int

const (
	// VictoryCapturePoints means capturing all control points
	VictoryCapturePoints VictoryCondition = iota
	// VictoryDestroyHall means destroying the enemy guild hall
	VictoryDestroyHall
	// VictoryDefenseTimeout means defenders held for entire assault phase
	VictoryDefenseTimeout
	// VictorySurrender means one side surrendered
	VictorySurrender
)

// String returns the human-readable name for the victory condition.
func (vc VictoryCondition) String() string {
	switch vc {
	case VictoryCapturePoints:
		return "Captured All Points"
	case VictoryDestroyHall:
		return "Guild Hall Destroyed"
	case VictoryDefenseTimeout:
		return "Defense Held"
	case VictorySurrender:
		return "Surrender"
	default:
		return "Unknown"
	}
}

// Siege represents an active siege on a territory.
type Siege struct {
	ID               string
	TerritoryID      string
	AttackerGuildID  string
	DefenderGuildID  string
	Phase            SiegePhase
	StartTime        time.Time
	PhaseStartTime   time.Time
	EndTime          time.Time
	VictoryCondition VictoryCondition
	WinnerGuildID    string

	// Participants
	Attackers      map[string]bool     // Player IDs
	Defenders      map[string]bool     // Player IDs
	Reinforcements map[string][]string // Guild ID -> Player IDs

	// Progress tracking
	ControlPointsCaptured int
	TotalControlPoints    int
	GuildHallHP           float64
	GuildHallMaxHP        float64

	// Loot distribution
	DefenderTreasury int
	LootPercentage   float64
	LootDistributed  bool
}

// NewSiege creates a new siege instance.
func NewSiege(territoryID, attackerGuild, defenderGuild string, defenderTreasury int) *Siege {
	now := time.Now()
	return &Siege{
		ID:                    fmt.Sprintf("siege_%s_%d", territoryID, now.Unix()),
		TerritoryID:           territoryID,
		AttackerGuildID:       attackerGuild,
		DefenderGuildID:       defenderGuild,
		Phase:                 PhasePreparation,
		StartTime:             now,
		PhaseStartTime:        now,
		EndTime:               time.Time{},
		Attackers:             make(map[string]bool),
		Defenders:             make(map[string]bool),
		Reinforcements:        make(map[string][]string),
		ControlPointsCaptured: 0,
		TotalControlPoints:    5, // Default 5 control points
		GuildHallHP:           10000.0,
		GuildHallMaxHP:        10000.0,
		DefenderTreasury:      defenderTreasury,
		LootPercentage:        0.15, // Default 15% loot
		LootDistributed:       false,
	}
}

// CanJoin checks if a player can join the siege.
func (s *Siege) CanJoin(playerID string, isAttacker bool) error {
	if s.Phase == PhaseEnded {
		return fmt.Errorf("siege has ended")
	}

	// Check if already in the siege
	if s.Attackers[playerID] || s.Defenders[playerID] {
		return fmt.Errorf("player already in siege")
	}

	// Check participant cap (50-100 players per siege)
	totalParticipants := len(s.Attackers) + len(s.Defenders)
	if totalParticipants >= 100 {
		return fmt.Errorf("siege is at maximum capacity")
	}

	return nil
}

// JoinSiege adds a player to the siege.
func (s *Siege) JoinSiege(playerID string, isAttacker bool) error {
	if err := s.CanJoin(playerID, isAttacker); err != nil {
		return err
	}

	if isAttacker {
		s.Attackers[playerID] = true
	} else {
		s.Defenders[playerID] = true
	}

	return nil
}

// AddReinforcements adds allied guild members to the siege.
func (s *Siege) AddReinforcements(guildID string, playerIDs []string) error {
	if s.Phase == PhaseEnded {
		return fmt.Errorf("siege has ended")
	}

	// Check if guild already has reinforcements
	if _, exists := s.Reinforcements[guildID]; exists {
		return fmt.Errorf("guild %s already has reinforcements", guildID)
	}

	// Cap reinforcements at 5 guilds per side
	if len(s.Reinforcements) >= 5 {
		return fmt.Errorf("maximum reinforcement guilds reached (5)")
	}

	// Add reinforcements
	s.Reinforcements[guildID] = playerIDs

	// Add players to defenders (reinforcements always defend)
	for _, playerID := range playerIDs {
		s.Defenders[playerID] = true
	}

	return nil
}

// AdvancePhase moves the siege to the next phase.
func (s *Siege) AdvancePhase() error {
	now := time.Now()

	switch s.Phase {
	case PhasePreparation:
		// Check if 1 hour has passed
		if now.Sub(s.PhaseStartTime) < time.Hour {
			return fmt.Errorf("preparation phase not complete yet")
		}
		s.Phase = PhaseAssault
		s.PhaseStartTime = now

	case PhaseAssault:
		// Check if 2 hours have passed or victory condition met
		if now.Sub(s.PhaseStartTime) < 2*time.Hour && s.WinnerGuildID == "" {
			return fmt.Errorf("assault phase not complete yet")
		}
		s.Phase = PhaseResolution
		s.PhaseStartTime = now

	case PhaseResolution:
		s.Phase = PhaseEnded
		s.EndTime = now

	default:
		return fmt.Errorf("cannot advance from phase %s", s.Phase)
	}

	return nil
}

// CaptureControlPoint marks a control point as captured.
func (s *Siege) CaptureControlPoint() error {
	if s.Phase != PhaseAssault {
		return fmt.Errorf("control points can only be captured during assault phase")
	}

	s.ControlPointsCaptured++

	// Check for victory condition
	if s.ControlPointsCaptured >= s.TotalControlPoints {
		s.VictoryCondition = VictoryCapturePoints
		s.WinnerGuildID = s.AttackerGuildID
	}

	return nil
}

// DamageGuildHall applies damage to the guild hall.
func (s *Siege) DamageGuildHall(damage float64) error {
	if s.Phase != PhaseAssault {
		return fmt.Errorf("guild hall can only be damaged during assault phase")
	}

	s.GuildHallHP -= damage
	if s.GuildHallHP < 0 {
		s.GuildHallHP = 0
	}

	// Check for victory condition
	if s.GuildHallHP <= 0 {
		s.VictoryCondition = VictoryDestroyHall
		s.WinnerGuildID = s.AttackerGuildID
	}

	return nil
}

// DistributeLoot calculates and distributes loot to winners.
func (s *Siege) DistributeLoot() (int, error) {
	if s.Phase != PhaseResolution {
		return 0, fmt.Errorf("loot can only be distributed during resolution phase")
	}

	if s.LootDistributed {
		return 0, fmt.Errorf("loot already distributed")
	}

	if s.WinnerGuildID == "" {
		return 0, fmt.Errorf("no winner determined")
	}

	// Calculate loot amount (10-30% of defender treasury)
	loot := int(float64(s.DefenderTreasury) * s.LootPercentage)
	s.LootDistributed = true

	return loot, nil
}

// SiegeManager manages multiple sieges across territories.
type SiegeManager struct {
	sieges map[string]*Siege // Siege ID -> Siege
	mu     sync.RWMutex
}

// NewSiegeManager creates a new siege manager.
func NewSiegeManager() *SiegeManager {
	return &SiegeManager{
		sieges: make(map[string]*Siege),
	}
}

// CreateSiege initiates a new siege.
func (sm *SiegeManager) CreateSiege(territoryID, attackerGuild, defenderGuild string, defenderTreasury int) (*Siege, error) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	// Check if territory already has an active siege
	for _, siege := range sm.sieges {
		if siege.TerritoryID == territoryID && siege.Phase != PhaseEnded {
			return nil, fmt.Errorf("territory %s already has an active siege", territoryID)
		}
	}

	siege := NewSiege(territoryID, attackerGuild, defenderGuild, defenderTreasury)
	sm.sieges[siege.ID] = siege

	return siege, nil
}

// GetSiege retrieves a siege by ID.
func (sm *SiegeManager) GetSiege(siegeID string) (*Siege, error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	siege, exists := sm.sieges[siegeID]
	if !exists {
		return nil, fmt.Errorf("siege not found: %s", siegeID)
	}

	return siege, nil
}

// GetActiveSieges returns all active sieges.
func (sm *SiegeManager) GetActiveSieges() []*Siege {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	active := make([]*Siege, 0)
	for _, siege := range sm.sieges {
		if siege.Phase != PhaseEnded {
			active = append(active, siege)
		}
	}

	return active
}

// GetSiegeForTerritory returns the active siege for a territory, if any.
func (sm *SiegeManager) GetSiegeForTerritory(territoryID string) (*Siege, bool) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	for _, siege := range sm.sieges {
		if siege.TerritoryID == territoryID && siege.Phase != PhaseEnded {
			return siege, true
		}
	}

	return nil, false
}

// Update processes all active sieges and advances phases as needed.
func (sm *SiegeManager) Update(deltaTime float64) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	now := time.Now()

	for _, siege := range sm.sieges {
		if siege.Phase == PhaseEnded {
			continue
		}

		// Check for phase advancement
		switch siege.Phase {
		case PhasePreparation:
			if now.Sub(siege.PhaseStartTime) >= time.Hour {
				siege.AdvancePhase()
			}

		case PhaseAssault:
			// Check for timeout victory (defenders held for 2 hours)
			if now.Sub(siege.PhaseStartTime) >= 2*time.Hour {
				if siege.WinnerGuildID == "" {
					siege.VictoryCondition = VictoryDefenseTimeout
					siege.WinnerGuildID = siege.DefenderGuildID
				}
				siege.AdvancePhase()
			}

		case PhaseResolution:
			// Auto-advance after resolution calculations
			if now.Sub(siege.PhaseStartTime) >= 5*time.Minute {
				siege.AdvancePhase()
			}
		}
	}
}

// GenerateDefensiveStructures procedurally generates defensive structures for a territory.
func GenerateDefensiveStructures(territoryID string, seed int64, count int) []*DefensiveStructure {
	rng := rand.New(rand.NewSource(seed))

	structures := make([]*DefensiveStructure, 0, count)

	// Generate 5-15 defensive structures
	structureCount := count
	if structureCount < 5 {
		structureCount = 5
	}
	if structureCount > 15 {
		structureCount = 15
	}

	for i := 0; i < structureCount; i++ {
		structType := StructureType(rng.Intn(3)) // Wall, Tower, or Guard

		hp := 1000.0 + float64(rng.Intn(4000))  // 1000-5000 HP
		damage := 50.0 + float64(rng.Intn(150)) // 50-200 damage
		level := 1 + rng.Intn(5)                // Level 1-5

		structure := &DefensiveStructure{
			ID:            fmt.Sprintf("%s_struct_%d", territoryID, i),
			Type:          structType,
			X:             float64(rng.Intn(500)),
			Y:             float64(rng.Intn(500)),
			HP:            hp,
			MaxHP:         hp,
			Damage:        damage,
			Level:         level,
			ConstructedAt: time.Now(),
		}

		structures = append(structures, structure)
	}

	return structures
}
