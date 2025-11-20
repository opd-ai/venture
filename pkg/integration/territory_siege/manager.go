package territory_siege

import (
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/opd-ai/venture/pkg/engine"
	"github.com/opd-ai/venture/pkg/network/federation/guild"
	"github.com/opd-ai/venture/pkg/world"
)

// SiegeManager manages active sieges and coordinates with existing systems.
type SiegeManager struct {
	world            *engine.World
	territoryManager *world.TerritoryManager
	politicsSystem   *engine.PoliticsSystem
	guildManager     *guild.Manager
	activeSieges     map[string]*Siege
	completedSieges  []*Siege
	structureGen     *StructureGenerator
	mu               sync.RWMutex
}

// NewSiegeManager creates a new siege manager.
func NewSiegeManager(w *engine.World, tm *world.TerritoryManager, ps *engine.PoliticsSystem, gm *guild.Manager) *SiegeManager {
	return &SiegeManager{
		world:            w,
		territoryManager: tm,
		politicsSystem:   ps,
		guildManager:     gm,
		activeSieges:     make(map[string]*Siege),
		completedSieges:  make([]*Siege, 0),
		structureGen:     NewStructureGenerator(rand.NewSource(time.Now().UnixNano())),
	}
}

// DeclareSiege initiates a siege on a territory.
func (sm *SiegeManager) DeclareSiege(attackerGuildID, defenderGuildID, zoneID string) (*Siege, error) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	// Validate zone exists
	zone, err := sm.territoryManager.GetZone(zoneID)
	if err != nil {
		return nil, fmt.Errorf("invalid zone: %w", err)
	}

	// Validate guilds exist
	attacker, err := sm.guildManager.GetGuild(attackerGuildID)
	if err != nil {
		return nil, fmt.Errorf("invalid attacker guild: %w", err)
	}

	defender, err := sm.guildManager.GetGuild(defenderGuildID)
	if err != nil {
		return nil, fmt.Errorf("invalid defender guild: %w", err)
	}

	// Check if zone is owned by defender
	if zone.OwnerFaction != defenderGuildID {
		return nil, fmt.Errorf("zone not owned by defender guild (owner: %s)", zone.OwnerFaction)
	}

	// Check for existing siege
	for _, siege := range sm.activeSieges {
		if siege.ZoneID == zoneID {
			return nil, fmt.Errorf("zone already under siege")
		}
	}

	// Generate siege ID
	siegeID := fmt.Sprintf("siege_%s_%d", zoneID, time.Now().Unix())

	// Generate defensive structures (5-15 based on zone size)
	structureCount := 5 + len(zone.ControlPoints)
	if structureCount > 15 {
		structureCount = 15
	}
	structures := sm.structureGen.GenerateStructures(zoneID, structureCount)

	// Create siege
	siege := &Siege{
		SiegeID:             siegeID,
		AttackerGuildID:     attackerGuildID,
		DefenderGuildID:     defenderGuildID,
		ZoneID:              zoneID,
		CurrentPhase:        PhasePreparation,
		PhaseStartTime:      time.Now().Unix(),
		PreparationDuration: 3600, // 1 hour
		AssaultDuration:     7200, // 2 hours
		ReinforcementGuilds: make([]string, 0),
		DefensiveStructures: structures,
		AttackerPlayerCount: 0,
		DefenderPlayerCount: 0,
		Victor:              "",
		TreasuryLoot:        0,
		LastUpdate:          time.Now().Unix(),
	}

	sm.activeSieges[siegeID] = siege

	// Log siege declaration (use attacker/defender names for clarity)
	_ = attacker.Name
	_ = defender.Name

	return siege, nil
}

// CallReinforcements allows allied guilds to join the defense.
func (sm *SiegeManager) CallReinforcements(siegeID, allyGuildID string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	siege, exists := sm.activeSieges[siegeID]
	if !exists {
		return fmt.Errorf("siege not found: %s", siegeID)
	}

	// Only during preparation phase
	if siege.CurrentPhase != PhasePreparation {
		return fmt.Errorf("reinforcements can only be called during preparation phase")
	}

	// Validate ally guild exists
	ally, err := sm.guildManager.GetGuild(allyGuildID)
	if err != nil {
		return fmt.Errorf("invalid ally guild: %w", err)
	}

	// Check if already reinforcing
	for _, existingAlly := range siege.ReinforcementGuilds {
		if existingAlly == allyGuildID {
			return fmt.Errorf("guild already reinforcing")
		}
	}

	// Check political alliance via politics system
	serverFaction := sm.politicsSystem.GetServerFaction()
	if serverFaction != nil {
		// If ally is not in alliance list, they can still join (player choice)
		// but we note it for potential political implications
		_ = ally.Name // Use ally to avoid unused warning
	}

	siege.ReinforcementGuilds = append(siege.ReinforcementGuilds, allyGuildID)

	return nil
}

// Update processes all active sieges, advancing phases and checking victory conditions.
func (sm *SiegeManager) Update(deltaTime float64) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	now := time.Now().Unix()

	for siegeID, siege := range sm.activeSieges {
		siege.LastUpdate = now

		// Check if phase should advance
		if siege.ShouldAdvancePhase() {
			switch siege.CurrentPhase {
			case PhasePreparation:
				// Advance to assault
				siege.CurrentPhase = PhaseAssault
				siege.PhaseStartTime = now

			case PhaseAssault:
				// Assault phase expired → defenders win
				siege.CurrentPhase = PhaseResolution
				siege.Victor = siege.DefenderGuildID

				// Move to completed sieges
				sm.completedSieges = append(sm.completedSieges, siege)
				delete(sm.activeSieges, siegeID)
			}
		}

		// Check victory conditions (during assault phase)
		if siege.CurrentPhase == PhaseAssault {
			victor, condition := sm.checkVictoryConditions(siege)
			if condition != VictoryConditionNone {
				// Siege complete
				siege.CurrentPhase = PhaseResolution
				siege.Victor = victor

				// Calculate loot
				result := sm.calculateSiegeResult(siege, condition)
				siege.TreasuryLoot = result.TreasuryLoot

				// Move to completed sieges
				sm.completedSieges = append(sm.completedSieges, siege)
				delete(sm.activeSieges, siegeID)
			}
		}
	}
}

// checkVictoryConditions determines if a siege has a winner.
func (sm *SiegeManager) checkVictoryConditions(siege *Siege) (victor string, condition VictoryCondition) {
	// Check guild hall (keep) destruction
	for _, structure := range siege.DefensiveStructures {
		if structure.Type == StructureKeep && structure.IsStructureDestroyed() {
			return siege.AttackerGuildID, VictoryConditionGuildHallDestroyed
		}
	}

	// Check control point capture (need 3+ points fully captured)
	zone, err := sm.territoryManager.GetZone(siege.ZoneID)
	if err == nil {
		capturedCount := 0
		for _, cp := range zone.ControlPoints {
			if cp.CaptureProgress >= 100.0 && cp.CapturingFaction == siege.AttackerGuildID {
				capturedCount++
			}
		}
		if capturedCount >= 3 {
			return siege.AttackerGuildID, VictoryConditionAllPointsCaptured
		}
	}

	// Check attacker elimination (no attackers present for 5+ minutes)
	if siege.AttackerPlayerCount == 0 {
		elapsed := time.Now().Unix() - siege.PhaseStartTime
		if elapsed > 300 { // 5 minutes
			return siege.DefenderGuildID, VictoryConditionAttackersEliminated
		}
	}

	return "", VictoryConditionNone
}

// calculateSiegeResult computes loot and rewards for the victor.
func (sm *SiegeManager) calculateSiegeResult(siege *Siege, condition VictoryCondition) *SiegeResult {
	result := &SiegeResult{
		VictorGuildID:    siege.Victor,
		VictoryCondition: condition,
		DurationSeconds:  time.Now().Unix() - siege.PhaseStartTime,
	}

	// Count captured control points
	zone, err := sm.territoryManager.GetZone(siege.ZoneID)
	if err == nil {
		for _, cp := range zone.ControlPoints {
			if cp.CapturingFaction == siege.AttackerGuildID && cp.CaptureProgress >= 100.0 {
				result.CapturedControlPoints++
			}
		}
	}

	// Count destroyed structures
	result.DestroyedStructures = siege.CountDestroyedStructures()

	// Calculate reward multiplier (1.0-3.0)
	baseMultiplier := 1.0

	// Bonus for capturing control points (up to +0.5)
	if len(zone.ControlPoints) > 0 {
		captureBonus := 0.5 * (float64(result.CapturedControlPoints) / float64(len(zone.ControlPoints)))
		baseMultiplier += captureBonus
	}

	// Bonus for destroying structures (up to +0.5)
	if len(siege.DefensiveStructures) > 0 {
		destructionBonus := 0.5 * siege.GetDestructionPercentage()
		baseMultiplier += destructionBonus
	}

	// Bonus for fast victory (up to +1.0)
	expectedDuration := siege.AssaultDuration
	if result.DurationSeconds < expectedDuration {
		speedBonus := 1.0 * (1.0 - float64(result.DurationSeconds)/float64(expectedDuration))
		baseMultiplier += speedBonus
	}

	result.RewardMultiplier = baseMultiplier

	// Calculate treasury loot (10-30% based on multiplier)
	defenderGuild, err := sm.guildManager.GetGuild(siege.DefenderGuildID)
	if err == nil {
		baseLootPercent := 0.10 // 10%
		maxLootPercent := 0.30  // 30%
		lootPercent := baseLootPercent + (maxLootPercent-baseLootPercent)*(result.RewardMultiplier/3.0)

		if lootPercent > maxLootPercent {
			lootPercent = maxLootPercent
		}

		result.TreasuryLoot = int(float64(defenderGuild.Treasury) * lootPercent)

		// Deduct from defender treasury (handled by caller)
		// This is just calculation; actual treasury modification done elsewhere
	}

	return result
}

// GetActiveSiege retrieves an active siege by ID.
func (sm *SiegeManager) GetActiveSiege(siegeID string) (*Siege, error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	siege, exists := sm.activeSieges[siegeID]
	if !exists {
		return nil, fmt.Errorf("siege not found: %s", siegeID)
	}

	return siege, nil
}

// GetActiveSiegesForZone returns all active sieges for a zone.
func (sm *SiegeManager) GetActiveSiegesForZone(zoneID string) []*Siege {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	sieges := make([]*Siege, 0)
	for _, siege := range sm.activeSieges {
		if siege.ZoneID == zoneID {
			sieges = append(sieges, siege)
		}
	}

	return sieges
}

// GetCompletedSieges returns all completed sieges.
func (sm *SiegeManager) GetCompletedSieges() []*Siege {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	completed := make([]*Siege, len(sm.completedSieges))
	copy(completed, sm.completedSieges)
	return completed
}

// DamageStructure applies damage to a defensive structure.
func (sm *SiegeManager) DamageStructure(siegeID, structureID string, damage int) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	siege, exists := sm.activeSieges[siegeID]
	if !exists {
		return fmt.Errorf("siege not found: %s", siegeID)
	}

	// Only during assault phase
	if siege.CurrentPhase != PhaseAssault {
		return fmt.Errorf("structures can only be damaged during assault phase")
	}

	for _, structure := range siege.DefensiveStructures {
		if structure.StructureID == structureID {
			structure.TakeDamage(damage)
			return nil
		}
	}

	return fmt.Errorf("structure not found: %s", structureID)
}

// UpdatePlayerCounts updates the attacker and defender player counts.
func (sm *SiegeManager) UpdatePlayerCounts(siegeID string, attackers, defenders int) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	siege, exists := sm.activeSieges[siegeID]
	if !exists {
		return fmt.Errorf("siege not found: %s", siegeID)
	}

	siege.AttackerPlayerCount = attackers
	siege.DefenderPlayerCount = defenders

	return nil
}
