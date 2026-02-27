package political_warfare

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"sync"
	"time"

	"github.com/opd-ai/venture/pkg/engine"
	"github.com/opd-ai/venture/pkg/network/federation/guild"
	"github.com/sirupsen/logrus"
)

// managerState holds the serializable state of the Manager for persistence.
// This is used internally by Save/Load methods to marshal all political
// warfare state to/from compressed JSON.
type managerState struct {
	Wars               []*WarDeclaration   `json:"wars"`
	Treaties           []*PeaceTreaty      `json:"treaties"`
	Embargoes          []*TradeEmbargo     `json:"embargoes"`
	AllianceCalls      []*AllianceCall     `json:"alliance_calls"`
	Penalties          []ReputationPenalty `json:"penalties"`
	AppliedConcessions []AppliedConcession `json:"applied_concessions"`
	Seed               int64               `json:"seed"`
}

// Manager coordinates political warfare between guilds
type Manager struct {
	world              *engine.World
	guildManager       *guild.Manager
	wars               map[string]*WarDeclaration // Key: attackerID_defenderID
	treaties           map[string]*PeaceTreaty    // Key: guildID1_guildID2 (sorted)
	embargoes          map[string]*TradeEmbargo   // Key: imposingID_targetID
	allianceCalls      map[string]*AllianceCall   // Key: callingID_targetID
	penalties          []ReputationPenalty
	appliedConcessions []AppliedConcession // Track all applied concessions
	rng                *rand.Rand
	seed               int64
	mu                 sync.RWMutex
}

// NewManager creates a new political warfare manager with deterministic RNG
// using the default seed. For deterministic results tied to world generation,
// use NewManagerWithSeed instead.
func NewManager(world *engine.World, guildManager *guild.Manager) *Manager {
	return NewManagerWithSeed(world, guildManager, DefaultSeed)
}

// NewManagerWithSeed creates a new political warfare manager with a specific seed.
// Use the world seed to ensure political calculations are reproducible
// across game sessions with the same world.
func NewManagerWithSeed(world *engine.World, guildManager *guild.Manager, seed int64) *Manager {
	return &Manager{
		world:              world,
		guildManager:       guildManager,
		wars:               make(map[string]*WarDeclaration),
		treaties:           make(map[string]*PeaceTreaty),
		embargoes:          make(map[string]*TradeEmbargo),
		allianceCalls:      make(map[string]*AllianceCall),
		penalties:          make([]ReputationPenalty, 0),
		appliedConcessions: make([]AppliedConcession, 0),
		rng:                rand.New(rand.NewSource(seed)),
		seed:               seed,
	}
}

// DeclareWar declares war between two guilds with a preparation period
func (m *Manager) DeclareWar(attackerGuildID, defenderGuildID string, preparationPeriod time.Duration) (*WarDeclaration, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Validate guilds are different
	if attackerGuildID == defenderGuildID {
		return nil, fmt.Errorf("a guild cannot declare war on itself")
	}

	// Validate guilds exist
	if _, err := m.guildManager.GetGuild(attackerGuildID); err != nil {
		return nil, fmt.Errorf("attacker guild not found: %w", err)
	}
	if _, err := m.guildManager.GetGuild(defenderGuildID); err != nil {
		return nil, fmt.Errorf("defender guild not found: %w", err)
	}

	// Check for existing war in either direction
	warKey := fmt.Sprintf("%s_%s", attackerGuildID, defenderGuildID)
	reverseWarKey := fmt.Sprintf("%s_%s", defenderGuildID, attackerGuildID)
	if existingWar, exists := m.wars[warKey]; exists && !existingWar.Ended {
		return nil, fmt.Errorf("war already declared between these guilds")
	}
	if existingWar, exists := m.wars[reverseWarKey]; exists && !existingWar.Ended {
		return nil, fmt.Errorf("war already declared between these guilds")
	}

	// Check for active peace treaty
	if m.isUnderPeaceTreaty(attackerGuildID, defenderGuildID) {
		return nil, fmt.Errorf("guilds are under active peace treaty")
	}

	// Create war declaration
	currentTime := now()
	war := &WarDeclaration{
		AttackerGuildID:   attackerGuildID,
		DefenderGuildID:   defenderGuildID,
		DeclaredAt:        currentTime,
		PreparationEnds:   currentTime.Add(preparationPeriod),
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

	// Validate guilds are different
	if guildID1 == guildID2 {
		return nil, fmt.Errorf("a guild cannot sign a peace treaty with itself")
	}

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

	currentTime := now()
	if war, exists := m.wars[warKey]; exists && !war.Ended {
		war.Ended = true
		war.EndedAt = currentTime
		war.Active = false
	}
	if war, exists := m.wars[reverseWarKey]; exists && !war.Ended {
		war.Ended = true
		war.EndedAt = currentTime
		war.Active = false
	}

	// Create peace treaty
	treatyKey := m.makeTreatyKey(guildID1, guildID2)

	// Default cooldown period: 7-14 days
	cooldownPeriod := duration
	if cooldownPeriod == 0 {
		cooldownPeriod = 7 * 24 * time.Hour // Default 7 days
	}

	treaty := &PeaceTreaty{
		GuildID1:     guildID1,
		GuildID2:     guildID2,
		SignedAt:     currentTime,
		ExpiresAt:    currentTime.Add(duration),
		CooldownEnds: currentTime.Add(cooldownPeriod),
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

	// Validate guilds are different
	if imposingGuildID == targetGuildID {
		return nil, fmt.Errorf("a guild cannot impose an embargo on itself")
	}

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
	embargo := &TradeEmbargo{
		ImposingGuildID: imposingGuildID,
		TargetGuildID:   targetGuildID,
		PriceIncrease:   priceIncreasePercent,
		ImposedAt:       now(),
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

	// Validate guilds are different
	if callingGuildID == targetGuildID {
		return nil, fmt.Errorf("a guild cannot call reinforcements against itself")
	}

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
		CalledAt:        now(),
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
			RespondedAt: now(),
			SuccessRate: successRate,
		}

		call.ResponingAllies = append(call.ResponingAllies, response)
	}

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
		war.EndedAt = now()
		war.Active = false
		war.Victor = attackerGuildID
		war.VictoryType = VictoryTypeDiplomatic

		// Apply concessions
		if err := m.applyConcessions(attackerGuildID, defenderGuildID, concessions, defenderGuild); err != nil {
			// Rollback war end state on concession failure
			war.Ended = false
			war.EndedAt = time.Time{}
			war.Active = true
			war.Victor = ""
			war.VictoryType = ""
			return false, fmt.Errorf("diplomatic victory failed: concessions could not be applied: %w", err)
		}
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

	currentTime := now()

	// Activate wars after preparation period
	for _, war := range m.wars {
		if !war.Active && !war.Ended && currentTime.After(war.PreparationEnds) {
			war.Active = true
		}
	}

	// Expire peace treaties
	for _, treaty := range m.treaties {
		if treaty.Active && currentTime.After(treaty.ExpiresAt) {
			treaty.Active = false
		}
	}

	// Expire trade embargoes
	for _, embargo := range m.embargoes {
		if embargo.Active && !embargo.ExpiresAt.IsZero() && currentTime.After(embargo.ExpiresAt) {
			embargo.Active = false
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
		AppliedAt: now(),
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
				totalValue += float64(goldAmount) / GoldValueNormalizer
			}
		case ConcessionTerritory:
			totalValue += TerritoryValueEquivalent
		case ConcessionApology:
			totalValue += ApologyValue
		case ConcessionTribute:
			if items, ok := concession.Value.([]string); ok {
				totalValue += float64(len(items)) * ItemValueEquivalent
			}
		case ConcessionTrade:
			if discount, ok := concession.Value.(float64); ok {
				totalValue += discount * TradeDiscountMultiplier
			}
		}
	}

	return totalValue
}

func (m *Manager) applyConcessions(attackerGuildID, defenderGuildID string, concessions []DiplomaticConcession, defenderGuild *guild.Guild) error {
	currentTime := now()
	var errs []error
	for _, concession := range concessions {
		applied := m.createAppliedConcession(concession, attackerGuildID, defenderGuildID, currentTime)
		if err := m.processConcessionType(concession, &applied, defenderGuild, attackerGuildID, currentTime); err != nil {
			errs = append(errs, fmt.Errorf("failed to apply %s concession: %w", concession.Type, err))
			continue
		}
		m.appliedConcessions = append(m.appliedConcessions, applied)
	}
	if len(errs) > 0 {
		return fmt.Errorf("failed to apply %d concession(s): %v", len(errs), errs)
	}
	return nil
}

// createAppliedConcession initializes a new applied concession record.
func (m *Manager) createAppliedConcession(concession DiplomaticConcession, attackerGuildID, defenderGuildID string, now time.Time) AppliedConcession {
	return AppliedConcession{
		Type:            concession.Type,
		AttackerGuildID: attackerGuildID,
		DefenderGuildID: defenderGuildID,
		AppliedAt:       now,
	}
}

// processConcessionType applies the concession based on its type.
func (m *Manager) processConcessionType(concession DiplomaticConcession, applied *AppliedConcession, defenderGuild *guild.Guild, attackerGuildID string, now time.Time) error {
	switch concession.Type {
	case ConcessionGold:
		return m.applyGoldConcession(concession, applied, defenderGuild, attackerGuildID)
	case ConcessionTerritory:
		m.applyTerritoryConcession(concession, applied)
		return nil
	case ConcessionApology:
		m.applyApologyConcession(concession, applied, attackerGuildID)
		return nil
	case ConcessionTribute:
		m.applyTributeConcession(concession, applied)
		return nil
	case ConcessionTrade:
		m.applyTradeConcession(concession, applied, now)
		return nil
	default:
		return fmt.Errorf("unknown concession type: %s", concession.Type)
	}
}

// applyGoldConcession transfers gold from defender to attacker.
// Returns error if attacker guild is not found, and rolls back defender deduction.
func (m *Manager) applyGoldConcession(concession DiplomaticConcession, applied *AppliedConcession, defenderGuild *guild.Guild, attackerGuildID string) error {
	goldAmount, ok := concession.Value.(int)
	if !ok {
		return nil // Invalid type, skip silently
	}

	// Deduct from defender
	defenderGuild.Treasury -= goldAmount

	// Add to attacker
	attackerGuild, err := m.guildManager.GetGuild(attackerGuildID)
	if err != nil {
		// Rollback defender deduction
		defenderGuild.Treasury += goldAmount
		logrus.WithFields(logrus.Fields{
			"attacker_guild_id": attackerGuildID,
			"defender_guild_id": defenderGuild.ID,
			"gold_amount":       goldAmount,
			"error":             err.Error(),
		}).Error("Failed to add gold to attacker guild during concession, rolled back defender deduction")
		return fmt.Errorf("attacker guild not found: %w", err)
	}

	attackerGuild.Treasury += goldAmount
	applied.GoldAmount = goldAmount
	return nil
}

// applyTerritoryConcession records territory transfer for external processing.
func (m *Manager) applyTerritoryConcession(concession DiplomaticConcession, applied *AppliedConcession) {
	if territoryID, ok := concession.Value.(string); ok && territoryID != "" {
		applied.TerritoryID = territoryID
	}
}

// applyApologyConcession records public apology for broadcast.
func (m *Manager) applyApologyConcession(concession DiplomaticConcession, applied *AppliedConcession, attackerGuildID string) {
	if apologyText, ok := concession.Value.(string); ok && apologyText != "" {
		applied.ApologyText = apologyText
	} else {
		applied.ApologyText = fmt.Sprintf("Guild %s publicly apologizes to guild %s for the conflict.",
			applied.DefenderGuildID, attackerGuildID)
	}
}

// applyTributeConcession records item tribute for transfer.
func (m *Manager) applyTributeConcession(concession DiplomaticConcession, applied *AppliedConcession) {
	if itemIDs, ok := concession.Value.([]string); ok && len(itemIDs) > 0 {
		applied.TributeItemIDs = itemIDs
	}
}

// applyTradeConcession records trade discount for future transactions.
func (m *Manager) applyTradeConcession(concession DiplomaticConcession, applied *AppliedConcession, now time.Time) {
	if discount, ok := concession.Value.(float64); ok && discount > 0 {
		applied.TradeDiscountPct = discount
		applied.TradeDiscountEnds = now.Add(TradeDiscountDuration)
	}
}

// GetAppliedConcessions returns all applied concessions
func (m *Manager) GetAppliedConcessions() []AppliedConcession {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]AppliedConcession, len(m.appliedConcessions))
	copy(result, m.appliedConcessions)
	return result
}

// GetTradeDiscount returns the current trade discount percentage that the
// attacker guild receives when trading with the defender guild.
// Returns 0 if no active discount exists.
func (m *Manager) GetTradeDiscount(attackerGuildID, defenderGuildID string) float64 {
	m.mu.RLock()
	defer m.mu.RUnlock()

	currentTime := now()
	for _, c := range m.appliedConcessions {
		if c.Type == ConcessionTrade &&
			c.AttackerGuildID == attackerGuildID &&
			c.DefenderGuildID == defenderGuildID &&
			(c.TradeDiscountEnds.IsZero() || c.TradeDiscountEnds.After(currentTime)) {
			return c.TradeDiscountPct
		}
	}
	return 0
}

// GetPendingTerritoryTransfers returns territory IDs pending transfer from
// diplomatic victories. External territory system should process these.
func (m *Manager) GetPendingTerritoryTransfers() []AppliedConcession {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []AppliedConcession
	for _, c := range m.appliedConcessions {
		if c.Type == ConcessionTerritory && c.TerritoryID != "" {
			result = append(result, c)
		}
	}
	return result
}

// GetPendingApologies returns public apologies pending broadcast.
// External messaging system should broadcast these to all players.
func (m *Manager) GetPendingApologies() []AppliedConcession {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []AppliedConcession
	for _, c := range m.appliedConcessions {
		if c.Type == ConcessionApology && c.ApologyText != "" {
			result = append(result, c)
		}
	}
	return result
}

// GetPendingTributes returns item tributes pending transfer.
// External inventory system should transfer these items.
func (m *Manager) GetPendingTributes() []AppliedConcession {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []AppliedConcession
	for _, c := range m.appliedConcessions {
		if c.Type == ConcessionTribute && len(c.TributeItemIDs) > 0 {
			result = append(result, c)
		}
	}
	return result
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

// GetActiveAllianceCalls returns all non-completed alliance calls.
func (m *Manager) GetActiveAllianceCalls() []*AllianceCall {
	m.mu.RLock()
	defer m.mu.RUnlock()

	calls := make([]*AllianceCall, 0)
	for _, call := range m.allianceCalls {
		if !call.Completed {
			calls = append(calls, call)
		}
	}
	return calls
}

// GetReputationPenalties returns all reputation penalties
func (m *Manager) GetReputationPenalties() []ReputationPenalty {
	m.mu.RLock()
	defer m.mu.RUnlock()

	penalties := make([]ReputationPenalty, len(m.penalties))
	copy(penalties, m.penalties)
	return penalties
}

// Save serializes the manager state to compressed JSON.
// The returned bytes can be stored to disk and later restored using Load().
// Thread-safe: acquires read lock during serialization.
func (m *Manager) Save() ([]byte, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Convert maps to slices for JSON serialization
	state := managerState{
		Penalties:          m.penalties,
		AppliedConcessions: m.appliedConcessions,
		Seed:               m.seed,
	}

	for _, war := range m.wars {
		state.Wars = append(state.Wars, war)
	}
	for _, treaty := range m.treaties {
		state.Treaties = append(state.Treaties, treaty)
	}
	for _, embargo := range m.embargoes {
		state.Embargoes = append(state.Embargoes, embargo)
	}
	for _, call := range m.allianceCalls {
		state.AllianceCalls = append(state.AllianceCalls, call)
	}

	jsonData, err := json.Marshal(state)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal political warfare state: %w", err)
	}

	var buf bytes.Buffer
	gzipWriter := gzip.NewWriter(&buf)
	if _, err := gzipWriter.Write(jsonData); err != nil {
		return nil, fmt.Errorf("failed to compress political warfare data: %w", err)
	}
	if err := gzipWriter.Close(); err != nil {
		return nil, fmt.Errorf("failed to close gzip writer: %w", err)
	}

	return buf.Bytes(), nil
}

// Load deserializes manager state from compressed JSON.
// Restores all wars, treaties, embargoes, alliance calls, penalties, and concessions.
// The RNG is re-initialized from the saved seed for deterministic continuation.
// Thread-safe: acquires write lock during deserialization.
func (m *Manager) Load(data []byte) error {
	gzipReader, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer gzipReader.Close()

	jsonData, err := io.ReadAll(gzipReader)
	if err != nil {
		return fmt.Errorf("failed to decompress political warfare data: %w", err)
	}

	var state managerState
	if err := json.Unmarshal(jsonData, &state); err != nil {
		return fmt.Errorf("failed to unmarshal political warfare state: %w", err)
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// Restore maps from slices
	m.wars = make(map[string]*WarDeclaration)
	for _, war := range state.Wars {
		key := fmt.Sprintf("%s_%s", war.AttackerGuildID, war.DefenderGuildID)
		m.wars[key] = war
	}

	m.treaties = make(map[string]*PeaceTreaty)
	for _, treaty := range state.Treaties {
		key := m.makeTreatyKey(treaty.GuildID1, treaty.GuildID2)
		m.treaties[key] = treaty
	}

	m.embargoes = make(map[string]*TradeEmbargo)
	for _, embargo := range state.Embargoes {
		key := fmt.Sprintf("%s_%s", embargo.ImposingGuildID, embargo.TargetGuildID)
		m.embargoes[key] = embargo
	}

	m.allianceCalls = make(map[string]*AllianceCall)
	for _, call := range state.AllianceCalls {
		key := fmt.Sprintf("%s_%s", call.CallingGuildID, call.TargetGuildID)
		m.allianceCalls[key] = call
	}

	m.penalties = state.Penalties
	if m.penalties == nil {
		m.penalties = make([]ReputationPenalty, 0)
	}

	m.appliedConcessions = state.AppliedConcessions
	if m.appliedConcessions == nil {
		m.appliedConcessions = make([]AppliedConcession, 0)
	}

	// Restore seed and reinitialize RNG for deterministic continuation
	m.seed = state.Seed
	m.rng = rand.New(rand.NewSource(m.seed))

	return nil
}
