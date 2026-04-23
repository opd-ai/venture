package legendary

import (
	"encoding/json"
	"fmt"
	"io"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/opd-ai/venture/pkg/procgen"
	"github.com/opd-ai/venture/pkg/world/raids"
)

// QuestManager manages legendary quests, player progress, and cross-server validation.
type QuestManager struct {
	mu               sync.RWMutex
	activeQuests     map[string]*LegendaryQuest  // questID -> quest
	playerProgress   map[string]*ProgressTracker // playerID -> tracker
	generator        *LegendaryQuestGenerator
	raidManager      *raids.Manager // Integration with Phase 59.1
	serverValidation *ServerValidator
	rewardCatalog    *RewardCatalog
	timeProvider     TimeProvider // TimeProvider for deterministic timestamps
}

// ServerValidator validates cross-server quest step completion.
type ServerValidator struct {
	mu               sync.RWMutex
	visitedServers   map[string]map[string]bool // playerID -> serverID -> visited
	federatedServers []string                   // List of known federated servers
}

// RewardCatalog tracks unique legendary items and ensures one-time rewards.
type RewardCatalog struct {
	mu             sync.RWMutex
	claimedRewards map[string]map[string]bool // playerID -> rewardID -> claimed
	rewardPool     []*LegendaryReward
	titlePool      []string
}

// LegendaryReward represents a unique one-time legendary reward.
type LegendaryReward struct {
	ID          string
	Name        string
	Description string
	ItemID      string // Links to item system
	Stats       map[string]int
	Rarity      float64 // 3.0 for legendary tier
	Unique      bool    // True for one-time rewards
}

// NewQuestManager creates a new legendary quest manager with default time provider.
func NewQuestManager(raidMgr *raids.Manager) *QuestManager {
	return NewQuestManagerWithTimeProvider(raidMgr, DefaultTimeProvider())
}

// NewQuestManagerWithTimeProvider creates a new legendary quest manager with a custom time provider.
// This enables deterministic timestamps for testing and reproducible state.
func NewQuestManagerWithTimeProvider(raidMgr *raids.Manager, tp TimeProvider) *QuestManager {
	return &QuestManager{
		activeQuests:     make(map[string]*LegendaryQuest),
		playerProgress:   make(map[string]*ProgressTracker),
		generator:        NewLegendaryQuestGenerator(),
		raidManager:      raidMgr,
		serverValidation: NewServerValidator(),
		rewardCatalog:    NewRewardCatalog(),
		timeProvider:     tp,
	}
}

// NewServerValidator creates a new server validator.
func NewServerValidator() *ServerValidator {
	return &ServerValidator{
		visitedServers:   make(map[string]map[string]bool),
		federatedServers: make([]string, 0),
	}
}

// NewRewardCatalog creates a new reward catalog with default legendary items.
func NewRewardCatalog() *RewardCatalog {
	return &RewardCatalog{
		claimedRewards: make(map[string]map[string]bool),
		rewardPool:     generateLegendaryRewards(),
		titlePool:      generateLegendaryTitles(),
	}
}

// GenerateQuest creates a new legendary quest for a player.
func (qm *QuestManager) GenerateQuest(playerID string, seed int64, params procgen.GenerationParams) (*LegendaryQuest, error) {
	result, err := qm.generator.Generate(seed, params)
	if err != nil {
		return nil, fmt.Errorf("failed to generate quest: %w", err)
	}

	quest := result.(*LegendaryQuest)

	// Validate quest meets requirements
	if err := qm.generator.Validate(quest); err != nil {
		return nil, fmt.Errorf("quest validation failed: %w", err)
	}

	qm.mu.Lock()
	qm.activeQuests[quest.ID] = quest
	qm.mu.Unlock()

	// Initialize progress tracker for player
	qm.getOrCreateTracker(playerID)

	return quest, nil
}

// UpdatePhaseProgress updates a player's progress on a specific quest phase.
func (qm *QuestManager) UpdatePhaseProgress(playerID, questID string, phaseIndex int, progress float64) error {
	qm.mu.RLock()
	quest, exists := qm.activeQuests[questID]
	qm.mu.RUnlock()

	if !exists {
		log.WithFields(log.Fields{
			"playerID": playerID,
			"questID":  questID,
		}).Warn("quest not found during phase progress update")
		return fmt.Errorf("quest not found: %s", questID)
	}

	if phaseIndex < 0 || phaseIndex >= len(quest.Phases) {
		log.WithFields(log.Fields{
			"playerID":   playerID,
			"questID":    questID,
			"phaseIndex": phaseIndex,
			"maxPhases":  len(quest.Phases),
		}).Warn("invalid phase index during progress update")
		return fmt.Errorf("invalid phase index: %d", phaseIndex)
	}

	tracker := qm.getOrCreateTracker(playerID)
	tracker.UpdatePhase(questID, playerID, phaseIndex, progress)

	// Mark phase as completed if progress reaches 1.0
	if progress >= 1.0 {
		qm.mu.Lock()
		quest.Phases[phaseIndex].Completed = true
		quest.Phases[phaseIndex].CompletedAt = qm.timeProvider.Now()
		qm.mu.Unlock()
	}

	return nil
}

// ValidateServerVisit validates that a player visited a required server.
func (qm *QuestManager) ValidateServerVisit(playerID, questID, serverID string) error {
	qm.mu.RLock()
	quest, exists := qm.activeQuests[questID]
	qm.mu.RUnlock()

	if !exists {
		log.WithFields(log.Fields{
			"playerID": playerID,
			"questID":  questID,
			"serverID": serverID,
		}).Warn("quest not found during server visit validation")
		return fmt.Errorf("quest not found: %s", questID)
	}

	// Check if server is required by quest
	found := false
	for _, phase := range quest.Phases {
		if phase.Type == PhaseTravel && phase.Requirements != nil {
			for _, reqServer := range phase.Requirements.ServersToVisit {
				if reqServer == serverID {
					found = true
					break
				}
			}
		}
	}

	if !found {
		log.WithFields(log.Fields{
			"playerID": playerID,
			"questID":  questID,
			"serverID": serverID,
		}).Debug("server not required by quest")
		return fmt.Errorf("server %s not required by quest", serverID)
	}

	// Validate server visit
	if err := qm.serverValidation.RecordVisit(playerID, serverID); err != nil {
		log.WithFields(log.Fields{
			"playerID": playerID,
			"questID":  questID,
			"serverID": serverID,
			"error":    err.Error(),
		}).Error("failed to record server visit")
		return fmt.Errorf("failed to record server visit: %w", err)
	}

	// Update quest progress
	tracker := qm.getOrCreateTracker(playerID)
	tracker.AddServerVisited(questID, playerID, serverID)

	return nil
}

// ValidateRaidCompletion validates that a player completed a required raid.
func (qm *QuestManager) ValidateRaidCompletion(playerID, questID, raidID string, tier raids.RaidTier) error {
	qm.mu.RLock()
	quest, exists := qm.activeQuests[questID]
	qm.mu.RUnlock()

	if !exists {
		log.WithFields(log.Fields{
			"playerID": playerID,
			"questID":  questID,
			"raidID":   raidID,
			"raidTier": tier.String(),
		}).Warn("quest not found during raid completion validation")
		return fmt.Errorf("quest not found: %s", questID)
	}

	// Check if raid is required by quest
	found := false
	for _, phase := range quest.Phases {
		if phase.Type == PhaseRaid && phase.Requirements != nil {
			for _, raidReq := range phase.Requirements.RaidEncounters {
				if raidReq.Tier == tier {
					found = true
					break
				}
			}
		}
	}

	if !found {
		log.WithFields(log.Fields{
			"playerID": playerID,
			"questID":  questID,
			"raidID":   raidID,
			"raidTier": tier.String(),
		}).Debug("raid tier not required by quest")
		return fmt.Errorf("raid tier %s not required by quest", tier.String())
	}

	// Update quest progress
	tracker := qm.getOrCreateTracker(playerID)
	tracker.AddRaidCompleted(questID, playerID, raidID)

	return nil
}

// ValidateCraftingCompletion validates that a player crafted a required item.
func (qm *QuestManager) ValidateCraftingCompletion(playerID, questID, itemID string, stationQuality int) error {
	qm.mu.RLock()
	quest, exists := qm.activeQuests[questID]
	qm.mu.RUnlock()

	if !exists {
		log.WithFields(log.Fields{
			"playerID":       playerID,
			"questID":        questID,
			"itemID":         itemID,
			"stationQuality": stationQuality,
		}).Warn("quest not found during crafting completion validation")
		return fmt.Errorf("quest not found: %s", questID)
	}

	// Check if crafting is required by quest
	found := false
	minQuality := 0
	for _, phase := range quest.Phases {
		if phase.Type == PhaseCraft && phase.Requirements != nil {
			for _, craftReq := range phase.Requirements.CraftItems {
				// Convert station quality string to int (Basic=1, Standard=2, Advanced=3, Master=4)
				qualityMap := map[string]int{
					"Basic":    1,
					"Standard": 2,
					"Advanced": 3,
					"Master":   4,
				}
				if reqQuality, ok := qualityMap[craftReq.StationQuality]; ok && reqQuality > minQuality {
					minQuality = reqQuality
				}
				found = true
			}
		}
	}

	if !found {
		log.WithFields(log.Fields{
			"playerID": playerID,
			"questID":  questID,
			"itemID":   itemID,
		}).Debug("crafting not required by quest")
		return fmt.Errorf("crafting not required by quest")
	}

	if stationQuality < minQuality {
		log.WithFields(log.Fields{
			"playerID":       playerID,
			"questID":        questID,
			"itemID":         itemID,
			"stationQuality": stationQuality,
			"requiredMin":    minQuality,
		}).Info("insufficient crafting station quality")
		return fmt.Errorf("insufficient crafting station quality: need %d, got %d", minQuality, stationQuality)
	}

	// Update quest progress
	tracker := qm.getOrCreateTracker(playerID)
	tracker.AddMaterial(questID, playerID, itemID, 1)

	return nil
}

// CompleteQuest marks a quest as complete and grants rewards.
func (qm *QuestManager) CompleteQuest(playerID, questID string) (*QuestRewards, error) {
	qm.mu.RLock()
	quest, exists := qm.activeQuests[questID]
	qm.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("quest not found: %s", questID)
	}

	// Verify all phases are complete
	if !quest.IsComplete() {
		return nil, fmt.Errorf("quest not complete")
	}

	// Grant rewards
	rewards, err := qm.grantRewards(playerID, quest)
	if err != nil {
		return nil, fmt.Errorf("failed to grant rewards: %w", err)
	}

	// Mark quest as complete
	tracker := qm.getOrCreateTracker(playerID)
	tracker.CompleteQuest(questID, playerID)

	return rewards, nil
}

// GetPlayerProgress returns a player's progress on a specific quest.
func (qm *QuestManager) GetPlayerProgress(playerID, questID string) *PlayerProgress {
	tracker := qm.getOrCreateTracker(playerID)
	return tracker.GetProgress(questID, playerID)
}

// GetActiveQuests returns all active legendary quests.
func (qm *QuestManager) GetActiveQuests() []*LegendaryQuest {
	qm.mu.RLock()
	defer qm.mu.RUnlock()

	quests := make([]*LegendaryQuest, 0, len(qm.activeQuests))
	for _, quest := range qm.activeQuests {
		quests = append(quests, quest)
	}
	return quests
}

// Save serializes the quest manager state.
func (qm *QuestManager) Save(w io.Writer) error {
	qm.mu.RLock()
	defer qm.mu.RUnlock()

	data := struct {
		ActiveQuests   map[string]*LegendaryQuest
		PlayerProgress map[string]*ProgressTracker
		ClaimedRewards map[string]map[string]bool
	}{
		ActiveQuests:   qm.activeQuests,
		PlayerProgress: qm.playerProgress,
		ClaimedRewards: qm.rewardCatalog.claimedRewards,
	}

	encoder := json.NewEncoder(w)
	return encoder.Encode(data)
}

// Load deserializes the quest manager state.
func (qm *QuestManager) Load(r io.Reader) error {
	var data struct {
		ActiveQuests   map[string]*LegendaryQuest
		PlayerProgress map[string]*ProgressTracker
		ClaimedRewards map[string]map[string]bool
	}

	decoder := json.NewDecoder(r)
	if err := decoder.Decode(&data); err != nil {
		return err
	}

	qm.mu.Lock()
	qm.activeQuests = data.ActiveQuests
	qm.playerProgress = data.PlayerProgress
	qm.rewardCatalog.claimedRewards = data.ClaimedRewards
	qm.mu.Unlock()

	return nil
}

// RecordVisit records a player's visit to a server.
func (sv *ServerValidator) RecordVisit(playerID, serverID string) error {
	sv.mu.Lock()
	defer sv.mu.Unlock()

	if sv.visitedServers[playerID] == nil {
		sv.visitedServers[playerID] = make(map[string]bool)
	}

	sv.visitedServers[playerID][serverID] = true
	return nil
}

// GetVisitedServers returns all servers a player has visited.
func (sv *ServerValidator) GetVisitedServers(playerID string) []string {
	sv.mu.RLock()
	defer sv.mu.RUnlock()

	servers := make([]string, 0)
	for serverID := range sv.visitedServers[playerID] {
		servers = append(servers, serverID)
	}
	return servers
}

// RegisterFederatedServer adds a server to the federated server list.
func (sv *ServerValidator) RegisterFederatedServer(serverID string) {
	sv.mu.Lock()
	defer sv.mu.Unlock()

	sv.federatedServers = append(sv.federatedServers, serverID)
}

// ClaimReward marks a reward as claimed by a player.
func (rc *RewardCatalog) ClaimReward(playerID, rewardID string) error {
	rc.mu.Lock()
	defer rc.mu.Unlock()

	if rc.claimedRewards[playerID] == nil {
		rc.claimedRewards[playerID] = make(map[string]bool)
	}

	if rc.claimedRewards[playerID][rewardID] {
		return fmt.Errorf("reward already claimed")
	}

	rc.claimedRewards[playerID][rewardID] = true
	return nil
}

// IsRewardClaimed checks if a player has claimed a specific reward.
func (rc *RewardCatalog) IsRewardClaimed(playerID, rewardID string) bool {
	rc.mu.RLock()
	defer rc.mu.RUnlock()

	return rc.claimedRewards[playerID][rewardID]
}

// GetAvailableRewards returns rewards that haven't been claimed by a player.
func (rc *RewardCatalog) GetAvailableRewards(playerID string) []*LegendaryReward {
	rc.mu.RLock()
	defer rc.mu.RUnlock()

	available := make([]*LegendaryReward, 0)
	for _, reward := range rc.rewardPool {
		if !reward.Unique || !rc.claimedRewards[playerID][reward.ID] {
			available = append(available, reward)
		}
	}
	return available
}

// getOrCreateTracker gets or creates a progress tracker for a player.
func (qm *QuestManager) getOrCreateTracker(playerID string) *ProgressTracker {
	qm.mu.Lock()
	defer qm.mu.Unlock()

	if qm.playerProgress[playerID] == nil {
		qm.playerProgress[playerID] = NewProgressTrackerWithTimeProvider(qm.timeProvider)
	}
	return qm.playerProgress[playerID]
}

// grantRewards grants quest rewards to a player.
func (qm *QuestManager) grantRewards(playerID string, quest *LegendaryQuest) (*QuestRewards, error) {
	rewards := &QuestRewards{
		Items:        make([]string, 0),
		Titles:       make([]string, 0),
		Gold:         quest.Rewards.Gold,
		SkippedItems: make([]string, 0),
	}

	// Grant legendary items
	for _, item := range quest.Rewards.Items {
		itemKey := fmt.Sprintf("item_%s", item.Name)
		if err := qm.rewardCatalog.ClaimReward(playerID, itemKey); err != nil {
			// Already claimed, log and record
			log.WithFields(log.Fields{
				"playerID": playerID,
				"questID":  quest.ID,
				"itemKey":  itemKey,
				"reason":   err.Error(),
			}).Info("reward already claimed, skipping")
			rewards.SkippedItems = append(rewards.SkippedItems, item.Name)
			continue
		}
		rewards.Items = append(rewards.Items, item.Name)
	}

	// Grant titles
	for _, title := range quest.Rewards.Titles {
		titleID := fmt.Sprintf("title_%s", title)
		if err := qm.rewardCatalog.ClaimReward(playerID, titleID); err != nil {
			log.WithFields(log.Fields{
				"playerID": playerID,
				"questID":  quest.ID,
				"titleID":  titleID,
				"reason":   err.Error(),
			}).Info("title already claimed, skipping")
			continue
		}
		rewards.Titles = append(rewards.Titles, title)
	}

	return rewards, nil
}

// QuestRewards represents rewards granted to a player.
type QuestRewards struct {
	Items        []string
	Titles       []string
	Gold         int
	SkippedItems []string // Items that were already claimed
}

// generateLegendaryRewards creates the pool of legendary rewards.
func generateLegendaryRewards() []*LegendaryReward {
	rewards := make([]*LegendaryReward, 0, 50)

	// Generate 50 unique legendary items
	for i := 1; i <= 50; i++ {
		reward := &LegendaryReward{
			ID:          fmt.Sprintf("legendary_item_%d", i),
			Name:        fmt.Sprintf("Legendary Item %d", i),
			Description: "A unique legendary item of immense power",
			ItemID:      fmt.Sprintf("item_%d", i),
			Stats: map[string]int{
				"damage":  100 + i*5,
				"defense": 50 + i*3,
				"health":  1000 + i*50,
			},
			Rarity: 3.0, // Legendary tier
			Unique: true,
		}
		rewards = append(rewards, reward)
	}

	return rewards
}

// generateLegendaryTitles creates the pool of legendary titles.
func generateLegendaryTitles() []string {
	return []string{
		"Legendary Hero",
		"Savior of Worlds",
		"Slayer of Legends",
		"Master of Fate",
		"Champion of the Realm",
		"Eternal Wanderer",
		"Destroyer of Darkness",
		"Bringer of Light",
		"Keeper of Secrets",
		"Lord of Legends",
		"Mythic Traveler",
		"Cosmic Champion",
		"Dimensional Guardian",
		"Eternal Protector",
		"Legendary Craftsman",
		"Master of Mysteries",
		"Vanquisher of Evil",
		"Hero of Heroes",
		"Legendary Explorer",
		"Ultimate Survivor",
	}
}

// QuestStatistics tracks quest completion statistics.
type QuestStatistics struct {
	TotalQuests     int
	CompletedQuests int
	ActiveQuests    int
	AverageHours    float64
	CompletionRate  float64
	LastUpdated     time.Time
}

// GetStatistics returns quest statistics.
// GetStatistics returns statistics about all legendary quests.
func (qm *QuestManager) GetStatistics() *QuestStatistics {
	qm.mu.RLock()
	defer qm.mu.RUnlock()

	stats := &QuestStatistics{
		TotalQuests:  len(qm.activeQuests),
		ActiveQuests: len(qm.activeQuests),
		LastUpdated:  qm.timeProvider.Now(),
	}

	qm.calculateCompletionRate(stats)
	return stats
}

// calculateCompletionRate calculates the quest completion rate from player progress.
func (qm *QuestManager) calculateCompletionRate(stats *QuestStatistics) {
	totalPlayers := len(qm.playerProgress)
	if totalPlayers == 0 {
		return
	}

	completedCount := qm.countCompletedQuests()
	stats.CompletedQuests = completedCount
	stats.CompletionRate = float64(completedCount) / float64(totalPlayers*len(qm.activeQuests))
}

// countCompletedQuests counts the total number of completed quests across all players.
func (qm *QuestManager) countCompletedQuests() int {
	completedCount := 0
	for _, tracker := range qm.playerProgress {
		completedCount += qm.countTrackerCompletedQuests(tracker)
	}
	return completedCount
}

// countTrackerCompletedQuests counts completed quests for a single progress tracker.
func (qm *QuestManager) countTrackerCompletedQuests(tracker *ProgressTracker) int {
	count := 0
	if len(tracker.Progress) == 0 {
		return 0
	}

	for questID, playerMap := range tracker.Progress {
		count += qm.countQuestCompletions(questID, playerMap)
	}
	return count
}

// countQuestCompletions counts completions for a specific quest.
func (qm *QuestManager) countQuestCompletions(questID string, playerMap map[string]*PlayerProgress) int {
	if questID == "" {
		return 0
	}

	count := 0
	for _, progress := range playerMap {
		if qm.isQuestCompleted(questID, progress) {
			count++
		}
	}
	return count
}

// isQuestCompleted checks if a quest is marked as completed.
func (qm *QuestManager) isQuestCompleted(questID string, progress *PlayerProgress) bool {
	quest, exists := qm.activeQuests[questID]
	return exists && quest.IsComplete() && progress.IsCompleted
}
