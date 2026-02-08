// Package engine provides the NG+ Reward component for ECS.
// This file implements NGPlusRewardComponent for tracking NG+ exclusive unlocks,
// achievements, items, titles, and challenges.
//
// Phase 114: NG+ Exclusive Content
package engine

import (
	"encoding/json"
	"math/rand"
	"sync"
)

// NGPlusRewardComponent tracks NG+ exclusive content unlocks.
// This includes special achievements, items, titles, and challenges
// that are only available in New Game Plus mode.
type NGPlusRewardComponent struct {
	mu sync.RWMutex

	// ExclusiveAchievements lists NG+ achievements earned
	ExclusiveAchievements []string `json:"exclusive_achievements"`

	// ExclusiveItems lists NG+ exclusive item IDs acquired
	ExclusiveItems []string `json:"exclusive_items"`

	// TitlesUnlocked lists NG+ exclusive titles earned
	TitlesUnlocked []string `json:"titles_unlocked"`

	// ChallengesCompleted tracks challenge ID -> completion status
	ChallengesCompleted map[string]bool `json:"challenges_completed"`

	// ChallengesActive tracks currently active challenge IDs
	ChallengesActive []string `json:"challenges_active"`

	// HighestTierReached is the highest NG+ tier for tiered rewards (1-10)
	HighestTierReached int `json:"highest_tier_reached"`

	// CurrentTitle is the currently equipped title
	CurrentTitle string `json:"current_title"`

	// TimeAttackBestTimes tracks best times for time attack challenges (ms)
	TimeAttackBestTimes map[string]int64 `json:"time_attack_best_times"`

	// NoDeathRunProgress tracks no-death run progress per cycle
	NoDeathRunProgress map[int]NoDeathRunData `json:"no_death_run_progress"`

	// NPCDialogVariationsUnlocked tracks unlocked NG+ dialog variations
	NPCDialogVariationsUnlocked []string `json:"npc_dialog_variations_unlocked"`
}

// NoDeathRunData tracks progress for a no-death challenge run.
type NoDeathRunData struct {
	CycleNumber    int    `json:"cycle_number"`
	BossesDefeated int    `json:"bosses_defeated"`
	AreasCleared   int    `json:"areas_cleared"`
	IsActive       bool   `json:"is_active"`
	WasCompleted   bool   `json:"was_completed"`
	FailedAt       string `json:"failed_at,omitempty"`
}

// Type returns the component type identifier.
func (n *NGPlusRewardComponent) Type() string {
	return "ngplus_reward"
}

// NewNGPlusRewardComponent creates a new NG+ reward component with defaults.
func NewNGPlusRewardComponent() *NGPlusRewardComponent {
	return &NGPlusRewardComponent{
		ExclusiveAchievements:       []string{},
		ExclusiveItems:              []string{},
		TitlesUnlocked:              []string{},
		ChallengesCompleted:         make(map[string]bool),
		ChallengesActive:            []string{},
		HighestTierReached:          0,
		CurrentTitle:                "",
		TimeAttackBestTimes:         make(map[string]int64),
		NoDeathRunProgress:          make(map[int]NoDeathRunData),
		NPCDialogVariationsUnlocked: []string{},
	}
}

// HasAchievement checks if a specific NG+ achievement has been earned.
func (n *NGPlusRewardComponent) HasAchievement(achievementID string) bool {
	n.mu.RLock()
	defer n.mu.RUnlock()
	for _, a := range n.ExclusiveAchievements {
		if a == achievementID {
			return true
		}
	}
	return false
}

// UnlockAchievement adds an achievement if not already earned.
// Returns true if the achievement was newly unlocked.
func (n *NGPlusRewardComponent) UnlockAchievement(achievementID string) bool {
	n.mu.Lock()
	defer n.mu.Unlock()
	for _, a := range n.ExclusiveAchievements {
		if a == achievementID {
			return false
		}
	}
	n.ExclusiveAchievements = append(n.ExclusiveAchievements, achievementID)
	return true
}

// GetAchievements returns a copy of all earned achievements.
func (n *NGPlusRewardComponent) GetAchievements() []string {
	n.mu.RLock()
	defer n.mu.RUnlock()
	result := make([]string, len(n.ExclusiveAchievements))
	copy(result, n.ExclusiveAchievements)
	return result
}

// HasItem checks if a specific NG+ item has been acquired.
func (n *NGPlusRewardComponent) HasItem(itemID string) bool {
	n.mu.RLock()
	defer n.mu.RUnlock()
	for _, item := range n.ExclusiveItems {
		if item == itemID {
			return true
		}
	}
	return false
}

// AddItem adds an exclusive item if not already owned.
// Returns true if the item was newly added.
func (n *NGPlusRewardComponent) AddItem(itemID string) bool {
	n.mu.Lock()
	defer n.mu.Unlock()
	for _, item := range n.ExclusiveItems {
		if item == itemID {
			return false
		}
	}
	n.ExclusiveItems = append(n.ExclusiveItems, itemID)
	return true
}

// GetItems returns a copy of all acquired exclusive items.
func (n *NGPlusRewardComponent) GetItems() []string {
	n.mu.RLock()
	defer n.mu.RUnlock()
	result := make([]string, len(n.ExclusiveItems))
	copy(result, n.ExclusiveItems)
	return result
}

// HasTitle checks if a specific title has been unlocked.
func (n *NGPlusRewardComponent) HasTitle(titleID string) bool {
	n.mu.RLock()
	defer n.mu.RUnlock()
	for _, t := range n.TitlesUnlocked {
		if t == titleID {
			return true
		}
	}
	return false
}

// UnlockTitle adds a title if not already unlocked.
// Returns true if the title was newly unlocked.
func (n *NGPlusRewardComponent) UnlockTitle(titleID string) bool {
	n.mu.Lock()
	defer n.mu.Unlock()
	for _, t := range n.TitlesUnlocked {
		if t == titleID {
			return false
		}
	}
	n.TitlesUnlocked = append(n.TitlesUnlocked, titleID)
	return true
}

// GetTitles returns a copy of all unlocked titles.
func (n *NGPlusRewardComponent) GetTitles() []string {
	n.mu.RLock()
	defer n.mu.RUnlock()
	result := make([]string, len(n.TitlesUnlocked))
	copy(result, n.TitlesUnlocked)
	return result
}

// SetCurrentTitle sets the active display title.
func (n *NGPlusRewardComponent) SetCurrentTitle(titleID string) {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.CurrentTitle = titleID
}

// GetCurrentTitle returns the currently equipped title.
func (n *NGPlusRewardComponent) GetCurrentTitle() string {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return n.CurrentTitle
}

// IsChallengeCompleted checks if a challenge has been completed.
func (n *NGPlusRewardComponent) IsChallengeCompleted(challengeID string) bool {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return n.ChallengesCompleted[challengeID]
}

// CompleteChallenge marks a challenge as completed.
// Returns true if the challenge was newly completed.
func (n *NGPlusRewardComponent) CompleteChallenge(challengeID string) bool {
	n.mu.Lock()
	defer n.mu.Unlock()
	if n.ChallengesCompleted[challengeID] {
		return false
	}
	n.ChallengesCompleted[challengeID] = true
	// Remove from active
	for i, c := range n.ChallengesActive {
		if c == challengeID {
			n.ChallengesActive = append(n.ChallengesActive[:i], n.ChallengesActive[i+1:]...)
			break
		}
	}
	return true
}

// ActivateChallenge adds a challenge to the active list.
func (n *NGPlusRewardComponent) ActivateChallenge(challengeID string) {
	n.mu.Lock()
	defer n.mu.Unlock()
	// Don't add if already completed or active
	if n.ChallengesCompleted[challengeID] {
		return
	}
	for _, c := range n.ChallengesActive {
		if c == challengeID {
			return
		}
	}
	n.ChallengesActive = append(n.ChallengesActive, challengeID)
}

// GetActiveChallenges returns a copy of active challenge IDs.
func (n *NGPlusRewardComponent) GetActiveChallenges() []string {
	n.mu.RLock()
	defer n.mu.RUnlock()
	result := make([]string, len(n.ChallengesActive))
	copy(result, n.ChallengesActive)
	return result
}

// GetCompletedChallenges returns all completed challenge IDs.
func (n *NGPlusRewardComponent) GetCompletedChallenges() []string {
	n.mu.RLock()
	defer n.mu.RUnlock()
	result := []string{}
	for id, completed := range n.ChallengesCompleted {
		if completed {
			result = append(result, id)
		}
	}
	return result
}

// UpdateHighestTier updates the highest NG+ tier if the new value is higher.
func (n *NGPlusRewardComponent) UpdateHighestTier(tier int) {
	n.mu.Lock()
	defer n.mu.Unlock()
	if tier > n.HighestTierReached {
		n.HighestTierReached = tier
	}
}

// GetHighestTierReached returns the highest NG+ tier achieved.
func (n *NGPlusRewardComponent) GetHighestTierReached() int {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return n.HighestTierReached
}

// RecordTimeAttack records a time attack completion time.
// Only updates if the new time is better (lower).
func (n *NGPlusRewardComponent) RecordTimeAttack(challengeID string, timeMs int64) bool {
	n.mu.Lock()
	defer n.mu.Unlock()
	existing, ok := n.TimeAttackBestTimes[challengeID]
	if !ok || timeMs < existing {
		n.TimeAttackBestTimes[challengeID] = timeMs
		return true
	}
	return false
}

// GetTimeAttackBest returns the best time for a time attack challenge.
// Returns 0 if not attempted.
func (n *NGPlusRewardComponent) GetTimeAttackBest(challengeID string) int64 {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return n.TimeAttackBestTimes[challengeID]
}

// StartNoDeathRun starts a no-death challenge run for the current cycle.
func (n *NGPlusRewardComponent) StartNoDeathRun(cycle int) {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.NoDeathRunProgress[cycle] = NoDeathRunData{
		CycleNumber:    cycle,
		BossesDefeated: 0,
		AreasCleared:   0,
		IsActive:       true,
		WasCompleted:   false,
	}
}

// UpdateNoDeathRun updates progress on an active no-death run.
func (n *NGPlusRewardComponent) UpdateNoDeathRun(cycle, bossesDefeated, areasCleared int) {
	n.mu.Lock()
	defer n.mu.Unlock()
	if data, ok := n.NoDeathRunProgress[cycle]; ok && data.IsActive {
		data.BossesDefeated = bossesDefeated
		data.AreasCleared = areasCleared
		n.NoDeathRunProgress[cycle] = data
	}
}

// FailNoDeathRun marks a no-death run as failed.
func (n *NGPlusRewardComponent) FailNoDeathRun(cycle int, failLocation string) {
	n.mu.Lock()
	defer n.mu.Unlock()
	if data, ok := n.NoDeathRunProgress[cycle]; ok {
		data.IsActive = false
		data.WasCompleted = false
		data.FailedAt = failLocation
		n.NoDeathRunProgress[cycle] = data
	}
}

// CompleteNoDeathRun marks a no-death run as successfully completed.
func (n *NGPlusRewardComponent) CompleteNoDeathRun(cycle int) {
	n.mu.Lock()
	defer n.mu.Unlock()
	if data, ok := n.NoDeathRunProgress[cycle]; ok {
		data.IsActive = false
		data.WasCompleted = true
		n.NoDeathRunProgress[cycle] = data
	}
}

// GetNoDeathRunData returns the no-death run data for a cycle.
func (n *NGPlusRewardComponent) GetNoDeathRunData(cycle int) (NoDeathRunData, bool) {
	n.mu.RLock()
	defer n.mu.RUnlock()
	data, ok := n.NoDeathRunProgress[cycle]
	return data, ok
}

// HasNoDeathCompletion returns true if any no-death run was completed.
func (n *NGPlusRewardComponent) HasNoDeathCompletion() bool {
	n.mu.RLock()
	defer n.mu.RUnlock()
	for _, data := range n.NoDeathRunProgress {
		if data.WasCompleted {
			return true
		}
	}
	return false
}

// UnlockNPCDialogVariation adds an NPC dialog variation as unlocked.
func (n *NGPlusRewardComponent) UnlockNPCDialogVariation(variationID string) bool {
	n.mu.Lock()
	defer n.mu.Unlock()
	for _, v := range n.NPCDialogVariationsUnlocked {
		if v == variationID {
			return false
		}
	}
	n.NPCDialogVariationsUnlocked = append(n.NPCDialogVariationsUnlocked, variationID)
	return true
}

// HasNPCDialogVariation checks if a dialog variation is unlocked.
func (n *NGPlusRewardComponent) HasNPCDialogVariation(variationID string) bool {
	n.mu.RLock()
	defer n.mu.RUnlock()
	for _, v := range n.NPCDialogVariationsUnlocked {
		if v == variationID {
			return true
		}
	}
	return false
}

// GetNPCDialogVariations returns all unlocked NPC dialog variations.
func (n *NGPlusRewardComponent) GetNPCDialogVariations() []string {
	n.mu.RLock()
	defer n.mu.RUnlock()
	result := make([]string, len(n.NPCDialogVariationsUnlocked))
	copy(result, n.NPCDialogVariationsUnlocked)
	return result
}

// Serialize converts the component to JSON for persistence.
func (n *NGPlusRewardComponent) Serialize() ([]byte, error) {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return json.Marshal(n)
}

// Deserialize restores the component from JSON data.
func (n *NGPlusRewardComponent) Deserialize(data []byte) error {
	n.mu.Lock()
	defer n.mu.Unlock()

	temp, err := unmarshalRewardData(data)
	if err != nil {
		return err
	}

	copyRewardFields(n, &temp)
	initializeNilFields(n)

	return nil
}

// unmarshalRewardData deserializes JSON data into a temporary component.
func unmarshalRewardData(data []byte) (NGPlusRewardComponent, error) {
	var temp NGPlusRewardComponent
	err := json.Unmarshal(data, &temp)
	return temp, err
}

// copyRewardFields copies all fields from source to destination.
func copyRewardFields(dst, src *NGPlusRewardComponent) {
	dst.ExclusiveAchievements = src.ExclusiveAchievements
	dst.ExclusiveItems = src.ExclusiveItems
	dst.TitlesUnlocked = src.TitlesUnlocked
	dst.ChallengesCompleted = src.ChallengesCompleted
	dst.ChallengesActive = src.ChallengesActive
	dst.HighestTierReached = src.HighestTierReached
	dst.CurrentTitle = src.CurrentTitle
	dst.TimeAttackBestTimes = src.TimeAttackBestTimes
	dst.NoDeathRunProgress = src.NoDeathRunProgress
	dst.NPCDialogVariationsUnlocked = src.NPCDialogVariationsUnlocked
}

// initializeNilFields ensures all maps and slices are non-nil.
func initializeNilFields(n *NGPlusRewardComponent) {
	if n.ExclusiveAchievements == nil {
		n.ExclusiveAchievements = []string{}
	}
	if n.ExclusiveItems == nil {
		n.ExclusiveItems = []string{}
	}
	if n.TitlesUnlocked == nil {
		n.TitlesUnlocked = []string{}
	}
	if n.ChallengesCompleted == nil {
		n.ChallengesCompleted = make(map[string]bool)
	}
	if n.ChallengesActive == nil {
		n.ChallengesActive = []string{}
	}
	if n.TimeAttackBestTimes == nil {
		n.TimeAttackBestTimes = make(map[string]int64)
	}
	if n.NoDeathRunProgress == nil {
		n.NoDeathRunProgress = make(map[int]NoDeathRunData)
	}
	if n.NPCDialogVariationsUnlocked == nil {
		n.NPCDialogVariationsUnlocked = []string{}
	}
}

// NGPlusAchievementDefinition defines an NG+ exclusive achievement.
type NGPlusAchievementDefinition struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	MinCycle    int    `json:"min_cycle"`
	IconID      string `json:"icon_id"`
}

// GetNGPlusAchievements returns all 10 NG+ exclusive achievement definitions.
func GetNGPlusAchievements() []NGPlusAchievementDefinition {
	return []NGPlusAchievementDefinition{
		{ID: "ngp_first_cycle", Name: "Reborn", Description: "Complete the game and enter NG+", MinCycle: 1, IconID: "icon_ngp_reborn"},
		{ID: "ngp_double", Name: "Twice-Fallen", Description: "Complete NG+1", MinCycle: 2, IconID: "icon_ngp_twice"},
		{ID: "ngp_triple", Name: "Thrice-Tempered", Description: "Complete NG+2", MinCycle: 3, IconID: "icon_ngp_thrice"},
		{ID: "ngp_veteran", Name: "Cycle Veteran", Description: "Reach NG+5", MinCycle: 5, IconID: "icon_ngp_veteran"},
		{ID: "ngp_master", Name: "Eternal Challenger", Description: "Reach NG+10", MinCycle: 10, IconID: "icon_ngp_master"},
		{ID: "ngp_speedrun", Name: "Swift Rebirth", Description: "Complete an NG+ cycle in under 2 hours", MinCycle: 1, IconID: "icon_ngp_speed"},
		{ID: "ngp_nodeaths", Name: "Deathless Legend", Description: "Complete an NG+ cycle with no deaths", MinCycle: 1, IconID: "icon_ngp_nodeaths"},
		{ID: "ngp_allbosses", Name: "Boss Slayer Eternal", Description: "Defeat all bosses in NG+5 or higher", MinCycle: 5, IconID: "icon_ngp_allbosses"},
		{ID: "ngp_collector", Name: "Eternal Collector", Description: "Collect all NG+ exclusive items", MinCycle: 1, IconID: "icon_ngp_collector"},
		{ID: "ngp_legend", Name: "Legendary Returner", Description: "Reach NG+99", MinCycle: 99, IconID: "icon_ngp_legend"},
	}
}

// NGPlusLegendaryItemDefinition defines an NG+ exclusive legendary item.
type NGPlusLegendaryItemDefinition struct {
	ID            string  `json:"id"`
	Name          string  `json:"name"`
	Description   string  `json:"description"`
	MinCycle      int     `json:"min_cycle"`
	BaseDamage    float64 `json:"base_damage"`
	BaseDefense   float64 `json:"base_defense"`
	SpecialEffect string  `json:"special_effect"`
}

// GetNGPlusLegendaryItems returns all 10 NG+ exclusive legendary items.
func GetNGPlusLegendaryItems() []NGPlusLegendaryItemDefinition {
	return []NGPlusLegendaryItemDefinition{
		{ID: "ngp_sword_cycle1", Name: "Blade of Rebirth", Description: "A sword forged from the essence of a completed cycle", MinCycle: 1, BaseDamage: 50, SpecialEffect: "+5% XP gain"},
		{ID: "ngp_shield_cycle2", Name: "Aegis of Persistence", Description: "A shield that grows stronger with each cycle", MinCycle: 2, BaseDefense: 40, SpecialEffect: "+10% block chance"},
		{ID: "ngp_ring_cycle3", Name: "Ring of Eternal Return", Description: "A ring that binds the wearer to fate", MinCycle: 3, SpecialEffect: "+15% rare drop chance"},
		{ID: "ngp_armor_cycle4", Name: "Vestments of the Returner", Description: "Armor worn by those who defy endings", MinCycle: 4, BaseDefense: 60, SpecialEffect: "+5% damage resistance"},
		{ID: "ngp_weapon_cycle5", Name: "Terminus, Blade of Cycles", Description: "A legendary blade that marks the end of eras", MinCycle: 5, BaseDamage: 100, SpecialEffect: "+10% critical chance"},
		{ID: "ngp_helm_cycle6", Name: "Crown of the Reborn", Description: "A crown for those who have conquered death", MinCycle: 6, BaseDefense: 35, SpecialEffect: "+20% experience from bosses"},
		{ID: "ngp_boots_cycle7", Name: "Striders of Infinity", Description: "Boots that have walked countless paths", MinCycle: 7, SpecialEffect: "+15% movement speed"},
		{ID: "ngp_amulet_cycle8", Name: "Amulet of Transcendence", Description: "An amulet that transcends mortal limits", MinCycle: 8, SpecialEffect: "+10% all stats"},
		{ID: "ngp_staff_cycle9", Name: "Staff of Eternal Wisdom", Description: "A staff imbued with knowledge of all cycles", MinCycle: 9, BaseDamage: 80, SpecialEffect: "+25% spell damage"},
		{ID: "ngp_artifact_cycle10", Name: "Heart of the Infinite", Description: "The ultimate artifact for true legends", MinCycle: 10, BaseDamage: 50, BaseDefense: 50, SpecialEffect: "+5% to all stats per NG+ cycle"},
	}
}

// NGPlusTitleDefinition defines an NG+ exclusive cosmetic title.
type NGPlusTitleDefinition struct {
	ID       string `json:"id"`
	Display  string `json:"display"`
	MinCycle int    `json:"min_cycle"`
}

// GetNGPlusTitles returns all NG+ exclusive cosmetic titles.
func GetNGPlusTitles() []NGPlusTitleDefinition {
	return []NGPlusTitleDefinition{
		{ID: "title_reborn", Display: "Reborn", MinCycle: 1},
		{ID: "title_twice_fallen", Display: "Twice-Fallen", MinCycle: 2},
		{ID: "title_thrice_tempered", Display: "Thrice-Tempered", MinCycle: 3},
		{ID: "title_cycle_walker", Display: "Cycle Walker", MinCycle: 4},
		{ID: "title_eternal_challenger", Display: "Eternal Challenger", MinCycle: 5},
		{ID: "title_deathless", Display: "Deathless", MinCycle: 1},     // Requires no-death completion
		{ID: "title_speedrunner", Display: "Speedrunner", MinCycle: 1}, // Requires time attack
		{ID: "title_legendary", Display: "Legendary", MinCycle: 10},
		{ID: "title_infinite", Display: "The Infinite", MinCycle: 25},
		{ID: "title_eternal_legend", Display: "Eternal Legend", MinCycle: 50},
	}
}

// NGPlusChallengeDefinition defines an NG+ exclusive challenge.
type NGPlusChallengeDefinition struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Description  string `json:"description"`
	MinCycle     int    `json:"min_cycle"`
	ChallengeTyp string `json:"type"` // "time_attack", "no_death", "boss_rush", "collect"
	TargetValue  int64  `json:"target_value"`
}

// GetNGPlusChallenges returns all NG+ exclusive challenge definitions.
func GetNGPlusChallenges() []NGPlusChallengeDefinition {
	return []NGPlusChallengeDefinition{
		{ID: "challenge_time_2h", Name: "Swift Cycle", Description: "Complete the cycle in under 2 hours", MinCycle: 1, ChallengeTyp: "time_attack", TargetValue: 7200000},
		{ID: "challenge_time_1h", Name: "Lightning Cycle", Description: "Complete the cycle in under 1 hour", MinCycle: 3, ChallengeTyp: "time_attack", TargetValue: 3600000},
		{ID: "challenge_nodeaths", Name: "Deathless Run", Description: "Complete the cycle without dying", MinCycle: 1, ChallengeTyp: "no_death", TargetValue: 0},
		{ID: "challenge_bossrush", Name: "Boss Rush", Description: "Defeat all bosses in sequence without resting", MinCycle: 5, ChallengeTyp: "boss_rush", TargetValue: 10},
		{ID: "challenge_collect_all", Name: "Collector's Edition", Description: "Find all NG+ exclusive items", MinCycle: 1, ChallengeTyp: "collect", TargetValue: 10},
	}
}

// GenerateDeterministicLegendaryItem returns the legendary item ID for a cycle using seed.
// Same seed and cycle always returns the same item.
func GenerateDeterministicLegendaryItem(seed int64, cycle int) string {
	if cycle <= 0 {
		return ""
	}

	items := GetNGPlusLegendaryItems()
	eligibleItems := []NGPlusLegendaryItemDefinition{}

	for _, item := range items {
		if item.MinCycle <= cycle {
			eligibleItems = append(eligibleItems, item)
		}
	}

	if len(eligibleItems) == 0 {
		return ""
	}

	// Deterministic selection using seed and cycle
	rng := rand.New(rand.NewSource(seed + int64(cycle)*1000))
	idx := rng.Intn(len(eligibleItems))
	return eligibleItems[idx].ID
}

// GetTierForCycle returns the reward tier (1-10) for a given NG+ cycle.
func GetTierForCycle(cycle int) int {
	if cycle <= 0 {
		return 0
	}
	if cycle >= 10 {
		return 10
	}
	return cycle
}
