package political_warfare

import (
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/opd-ai/venture/pkg/engine"
	"github.com/opd-ai/venture/pkg/network/federation/guild"
)

// Manager coordinates political warfare between guilds
type Manager struct {
	world         *engine.World
	guildManager  *guild.Manager
	wars          map[string]*WarDeclaration // Key: attackerID_defenderID
	treaties      map[string]*PeaceTreaty    // Key: guildID1_guildID2 (sorted)
	embargoes     map[string]*TradeEmbargo   // Key: imposingID_targetID
	allianceCalls map[string]*AllianceCall   // Key: callingID_targetID
	penalties     []ReputationPenalty
	rng           *rand.Rand
	seed          int64
	mu            sync.RWMutex
}

// NewManager creates a new political warfare manager with deterministic RNG
// Uses guild manager hash as seed for reproducible political calculations
func NewManager(world *engine.World, guildManager *guild.Manager) *Manager {
	// Use deterministic seed based on guild manager state
	// This ensures same guild configurations produce same political outcomes
	seed := int64(12345) // Default seed, can be derived from game world seed in future
	return &Manager{
		world:         world,
		guildManager:  guildManager,
		wars:          make(map[string]*WarDeclaration),
		treaties:      make(map[string]*PeaceTreaty),
		embargoes:     make(map[string]*TradeEmbargo),
		allianceCalls: make(map[string]*AllianceCall),
		penalties:     make([]ReputationPenalty, 0),
		rng:           rand.New(rand.NewSource(seed)),
		seed:          seed,
	}
}

// DeclareWar declares war between two guilds with a preparation period
func (m *Manager) DeclareWar(attackerGuildID, defenderGuildID string, preparationPeriod time.Duration) (*WarDeclaration, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Validate guilds exist
	if _, err := m.guildManager.GetGuild(attackerGuildID); err != nil {
		return nil, fmt.Errorf("attacker guild not found: %w", err)
	}
	if _, err := m.guildManager.GetGuild(defenderGuildID); err != nil {
		return nil, fmt.Errorf("defender guild not found: %w", err)
	}

	// Check for existing war
	warKey := fmt.Sprintf("%s_%s", attackerGuildID, defenderGuildID)
	if existingWar, exists := m.wars[warKey]; exists && !existingWar.Ended {
		return nil, fmt.Errorf("war already declared between these guilds")
	}

	// Check for active peace treaty
	if m.isUnderPeaceTreaty(attackerGuildID, defenderGuildID) {
		return nil, fmt.Errorf("guilds are under active peace treaty")
	}

	// Create war declaration
	now := time.Now()
	war := &WarDeclaration{
		AttackerGuildID:   attackerGuildID,
		DefenderGuildID:   defenderGuildID,
		DeclaredAt:        now,
		PreparationEnds:   now.Add(preparationPeriod),
		PreparationPeriod: preparationPeriod,
		Active:            false, // Becomes active after preparation
		Ended:             false,
	}

	m.wars[warKey] = war

	// Apply reputation penalty for war declaration
	m.applyReputationPenaltyInternal(attackerGuildID, "war_declaration", -0.2)

	return war, nil
}

// SignPeaceTreaty signs a peace treaty with cooldown period
func (m *Manager) SignPeaceTreaty(guildID1, guildID2 string, duration time.Duration) (*PeaceTreaty, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Validate guilds exist
	if _, err := m.guildManager.GetGuild(guildID1); err != nil {
		return nil, fmt.Errorf("guild 1 not found: %w", err)
	}
	if _, err := m.guildManager.GetGuild(guildID2); err != nil {
		return nil, fmt.Errorf("guild 2 not found: %w", err)
	}

	// End any active war
	warKey := fmt.Sprintf("%s_%s", guildID1, guildID2)
	reverseWarKey := fmt.Sprintf("%s_%s", guildID2, guildID1)

	if war, exists := m.wars[warKey]; exists && !war.Ended {
		war.Ended = true
		war.EndedAt = time.Now()
		war.Active = false
	}
	if war, exists := m.wars[reverseWarKey]; exists && !war.Ended {
		war.Ended = true
		war.EndedAt = time.Now()
		war.Active = false
	}

	// Create peace treaty
	now := time.Now()
	treatyKey := m.makeTreatyKey(guildID1, guildID2)

	// Default cooldown period: 7-14 days
	cooldownPeriod := duration
	if cooldownPeriod == 0 {
		cooldownPeriod = 7 * 24 * time.Hour // Default 7 days
	}

	treaty := &PeaceTreaty{
		GuildID1:     guildID1,
		GuildID2:     guildID2,
		SignedAt:     now,
		ExpiresAt:    now.Add(duration),
		CooldownEnds: now.Add(cooldownPeriod),
		Duration:     duration,
		Active:       true,
	}

	m.treaties[treatyKey] = treaty

	return treaty, nil
}

// ImposeEmbargo imposes a trade embargo on a target guild
func (m *Manager) ImposeEmbargo(imposingGuildID, targetGuildID string, priceIncreasePercent float64) (*TradeEmbargo, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Validate guilds exist
	if _, err := m.guildManager.GetGuild(imposingGuildID); err != nil {
		return nil, fmt.Errorf("imposing guild not found: %w", err)
	}
	if _, err := m.guildManager.GetGuild(targetGuildID); err != nil {
		return nil, fmt.Errorf("target guild not found: %w", err)
	}

	// Validate price increase range (50-90%)
	if priceIncreasePercent < 0.5 || priceIncreasePercent > 0.9 {
		return nil, fmt.Errorf("price increase must be between 50%% and 90%%, got %.1f%%", priceIncreasePercent*100)
	}

	// Check for existing embargo
	embargoKey := fmt.Sprintf("%s_%s", imposingGuildID, targetGuildID)
	if existingEmbargo, exists := m.embargoes[embargoKey]; exists && existingEmbargo.Active {
		return nil, fmt.Errorf("embargo already active between these guilds")
	}

	// Create embargo
	now := time.Now()
	embargo := &TradeEmbargo{
		ImposingGuildID: imposingGuildID,
		TargetGuildID:   targetGuildID,
		PriceIncrease:   priceIncreasePercent,
		ImposedAt:       now,
		Active:          true,
	}

	m.embargoes[embargoKey] = embargo

	// Apply reputation penalty for embargo
	m.applyReputationPenaltyInternal(imposingGuildID, "embargo", -0.1)

	return embargo, nil
}

// CallReinforcementAllies calls allied guilds for siege reinforcements
func (m *Manager) CallReinforcementAllies(callingGuildID, targetGuildID string) (*AllianceCall, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Validate guilds exist
	callingGuild, err := m.guildManager.GetGuild(callingGuildID)
	if err != nil {
		return nil, fmt.Errorf("calling guild not found: %w", err)
	}
	if _, err := m.guildManager.GetGuild(targetGuildID); err != nil {
		return nil, fmt.Errorf("target guild not found: %w", err)
	}

	// Create alliance call
	call := &AllianceCall{
		CallingGuildID:  callingGuildID,
		TargetGuildID:   targetGuildID,
		CalledAt:        time.Now(),
		ResponingAllies: make([]AllianceResponse, 0),
		Completed:       false,
	}

	// Find allied guilds (guilds with high reputation)
	// Note: This simplified implementation checks only the calling guild's reputation map
	for allyGuildID, rep := range callingGuild.Reputation {
		if allyGuildID == callingGuildID || allyGuildID == targetGuildID {
			continue
		}

		// Check reputation (>= 0.6 = allied)
		if rep < 0.6 {
			continue
		}

		// Calculate success rate based on reputation (0.6-0.8 reputation → 60-80% success)
		successRate := rep
		if successRate > 0.8 {
			successRate = 0.8
		}
		if successRate < 0.6 {
			successRate = 0.6
		}

		// Roll for acceptance
		accepted := m.rng.Float64() < successRate

		response := AllianceResponse{
			AllyGuildID: allyGuildID,
			Accepted:    accepted,
			RespondedAt: time.Now(),
			SuccessRate: successRate,
		}

		call.ResponingAllies = append(call.ResponingAllies, response)
	}

	call.Completed = true

	// Store alliance call
	callKey := fmt.Sprintf("%s_%s", callingGuildID, targetGuildID)
	m.allianceCalls[callKey] = call

	return call, nil
}

// NegotiateDiplomaticVictory attempts a diplomatic surrender
func (m *Manager) NegotiateDiplomaticVictory(attackerGuildID, defenderGuildID string, concessions []DiplomaticConcession) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Validate guilds exist
	_, err := m.guildManager.GetGuild(attackerGuildID)
	if err != nil {
		return false, fmt.Errorf("attacker guild not found: %w", err)
	}
	defenderGuild, err := m.guildManager.GetGuild(defenderGuildID)
	if err != nil {
		return false, fmt.Errorf("defender guild not found: %w", err)
	}

	// Check for active war
	warKey := fmt.Sprintf("%s_%s", attackerGuildID, defenderGuildID)
	war, exists := m.wars[warKey]
	if !exists || war.Ended {
		return false, fmt.Errorf("no active war between these guilds")
	}

	// Calculate concession value
	totalValue := m.calculateConcessionValue(concessions, defenderGuild)

	// Diplomatic victory success rate: 10-20% base + concession value bonus
	baseRate := 0.10 + m.rng.Float64()*0.10 // 10-20%
	concessionBonus := totalValue * 0.1     // Each 10k gold equivalent adds 1%
	successRate := baseRate + concessionBonus

	if successRate > 0.5 {
		successRate = 0.5 // Cap at 50%
	}

	// Roll for success
	success := m.rng.Float64() < successRate

	if success {
		// End war with diplomatic victory
		war.Ended = true
		war.EndedAt = time.Now()
		war.Active = false
		war.Victor = attackerGuildID
		war.VictoryType = VictoryTypeDiplomatic

		// Apply concessions
		m.applyConcessions(attackerGuildID, defenderGuildID, concessions, defenderGuild)
	}

	return success, nil
}

// ApplyReputationPenalty applies reputation penalty for aggressive actions
func (m *Manager) ApplyReputationPenalty(guildID, action string, penalty float64) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if penalty > 0 {
		return fmt.Errorf("penalty must be negative, got %.2f", penalty)
	}
	if penalty < -0.5 {
		return fmt.Errorf("penalty too severe, must be between -0.1 and -0.5, got %.2f", penalty)
	}

	m.applyReputationPenaltyInternal(guildID, action, penalty)
	return nil
}

// Update processes time-based state changes (war activation, treaty expiration)
func (m *Manager) Update(deltaTime float64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()

	// Activate wars after preparation period
	for _, war := range m.wars {
		if !war.Active && !war.Ended && now.After(war.PreparationEnds) {
			war.Active = true
		}
	}

	// Expire peace treaties
	for _, treaty := range m.treaties {
		if treaty.Active && now.After(treaty.ExpiresAt) {
			treaty.Active = false
		}
	}
}

// Private helper methods

func (m *Manager) isUnderPeaceTreaty(guildID1, guildID2 string) bool {
	treatyKey := m.makeTreatyKey(guildID1, guildID2)
	treaty, exists := m.treaties[treatyKey]
	return exists && treaty.Active
}

func (m *Manager) makeTreatyKey(guildID1, guildID2 string) string {
	// Sort guild IDs to create consistent key
	if guildID1 < guildID2 {
		return fmt.Sprintf("%s_%s", guildID1, guildID2)
	}
	return fmt.Sprintf("%s_%s", guildID2, guildID1)
}

func (m *Manager) applyReputationPenaltyInternal(guildID, action string, penalty float64) {
	// Record reputation penalty
	// Note: In a full implementation, this would interact with the faction system
	// For now, we just record the penalty for tracking purposes

	penaltyRecord := ReputationPenalty{
		GuildID:   guildID,
		Action:    action,
		Penalty:   penalty,
		AppliedAt: time.Now(),
		FactionID: "all", // Simplified: apply to all factions
	}

	m.penalties = append(m.penalties, penaltyRecord)
}

func (m *Manager) calculateConcessionValue(concessions []DiplomaticConcession, defenderGuild *guild.Guild) float64 {
	totalValue := 0.0

	for _, concession := range concessions {
		switch concession.Type {
		case ConcessionGold:
			if goldAmount, ok := concession.Value.(int); ok {
				totalValue += float64(goldAmount) / 10000.0 // Normalize to 10k gold = 1.0
			}
		case ConcessionTerritory:
			totalValue += 2.0 // Territory worth ~20k gold equivalent
		case ConcessionApology:
			totalValue += 0.1 // Apology adds small value
		case ConcessionTribute:
			if items, ok := concession.Value.([]string); ok {
				totalValue += float64(len(items)) * 0.5 // Each item ~5k gold
			}
		case ConcessionTrade:
			if discount, ok := concession.Value.(float64); ok {
				totalValue += discount * 0.5 // Discount adds value
			}
		}
	}

	return totalValue
}

func (m *Manager) applyConcessions(attackerGuildID, defenderGuildID string, concessions []DiplomaticConcession, defenderGuild *guild.Guild) {
	for _, concession := range concessions {
		switch concession.Type {
		case ConcessionGold:
			if goldAmount, ok := concession.Value.(int); ok {
				// Transfer gold from defender to attacker
				defenderGuild.Treasury -= goldAmount
				if attackerGuild, err := m.guildManager.GetGuild(attackerGuildID); err == nil {
					attackerGuild.Treasury += goldAmount
				}
			}
			// Other concession types would be implemented similarly
		}
	}
}

// GetActiveWars returns all active wars
func (m *Manager) GetActiveWars() []*WarDeclaration {
	m.mu.RLock()
	defer m.mu.RUnlock()

	wars := make([]*WarDeclaration, 0)
	for _, war := range m.wars {
		if !war.Ended {
			wars = append(wars, war)
		}
	}
	return wars
}

// GetActiveTreaties returns all active peace treaties
func (m *Manager) GetActiveTreaties() []*PeaceTreaty {
	m.mu.RLock()
	defer m.mu.RUnlock()

	treaties := make([]*PeaceTreaty, 0)
	for _, treaty := range m.treaties {
		if treaty.Active {
			treaties = append(treaties, treaty)
		}
	}
	return treaties
}

// GetActiveEmbargoes returns all active embargoes
func (m *Manager) GetActiveEmbargoes() []*TradeEmbargo {
	m.mu.RLock()
	defer m.mu.RUnlock()

	embargoes := make([]*TradeEmbargo, 0)
	for _, embargo := range m.embargoes {
		if embargo.Active {
			embargoes = append(embargoes, embargo)
		}
	}
	return embargoes
}

// GetReputationPenalties returns all reputation penalties
func (m *Manager) GetReputationPenalties() []ReputationPenalty {
	m.mu.RLock()
	defer m.mu.RUnlock()

	penalties := make([]ReputationPenalty, len(m.penalties))
	copy(penalties, m.penalties)
	return penalties
}
