package legendary

import (
	"fmt"
	"math/rand"

	"github.com/opd-ai/venture/pkg/procgen"
	"github.com/opd-ai/venture/pkg/world/raids"
)

// LegendaryQuestGenerator generates multi-phase legendary quests.
type LegendaryQuestGenerator struct {
	templates []*QuestTemplate
}

// NewLegendaryQuestGenerator creates a new legendary quest generator.
func NewLegendaryQuestGenerator() *LegendaryQuestGenerator {
	return &LegendaryQuestGenerator{
		templates: defaultQuestTemplates(),
	}
}

// Generate creates a new legendary quest.
// Implements procgen.Generator interface.
func (g *LegendaryQuestGenerator) Generate(seed int64, params procgen.GenerationParams) (interface{}, error) {
	rng := rand.New(rand.NewSource(seed))

	// Select template based on difficulty
	templateIndex := int(params.Difficulty * float64(len(g.templates)))
	if templateIndex < 0 {
		templateIndex = 0
	}
	if templateIndex >= len(g.templates) {
		templateIndex = len(g.templates) - 1
	}
	template := g.templates[templateIndex]

	// Generate quest ID
	questID := fmt.Sprintf("legendary_quest_%d", seed)

	// Generate quest name
	questName := g.generateQuestName(rng, template, params.GenreID)

	// Generate quest description
	description := g.generateDescription(rng, questName, params.GenreID)

	// Determine number of phases
	numPhases := template.MinPhases + rng.Intn(template.MaxPhases-template.MinPhases+1)

	// Ensure quest meets requirements
	if numPhases < 5 {
		numPhases = 5 // Minimum per spec
	}
	if numPhases > 10 {
		numPhases = 10 // Maximum per spec
	}

	// Generate phases
	phases := make([]*QuestPhase, numPhases)
	usedRaid := false
	usedCrafting := false
	usedTravel := false

	for i := 0; i < numPhases; i++ {
		phaseType := template.PhaseTypes[rng.Intn(len(template.PhaseTypes))]

		// Ensure requirements are met
		if i == numPhases-2 && template.RequiresRaid && !usedRaid {
			phaseType = PhaseRaid
		}
		if i == numPhases-1 && template.RequiresCrafting && !usedCrafting {
			phaseType = PhaseCraft
		}

		phase := g.generatePhase(rng, i+1, phaseType, params, template)
		phases[i] = phase

		if phaseType == PhaseRaid {
			usedRaid = true
		}
		if phaseType == PhaseCraft {
			usedCrafting = true
		}
		if phaseType == PhaseTravel {
			usedTravel = true
		}
	}

	// Ensure at least one travel phase for cross-server requirement
	if !usedTravel {
		// Replace a random phase with travel
		replaceIndex := 1 + rng.Intn(numPhases-2) // Not first or last
		phases[replaceIndex] = g.generatePhase(rng, replaceIndex+1, PhaseTravel, params, template)
	}

	// Generate rewards
	rewards := g.generateRewards(rng, params, template)

	// Calculate estimated hours
	estimatedHours := template.EstimatedHoursMin + rng.Float64()*(template.EstimatedHoursMax-template.EstimatedHoursMin)

	// Create quest
	quest := &LegendaryQuest{
		ID:             questID,
		Name:           questName,
		Description:    description,
		Phases:         phases,
		Rewards:        rewards,
		RequiredLevel:  20 + int(params.Depth*5), // Scale with depth
		Seed:           seed,
		EstimatedHours: estimatedHours,
	}

	return quest, nil
}

// Validate ensures the generated quest meets quality standards.
// Implements procgen.Generator interface.
func (g *LegendaryQuestGenerator) Validate(result interface{}) error {
	quest, ok := result.(*LegendaryQuest)
	if !ok {
		return fmt.Errorf("result is not a *LegendaryQuest")
	}

	// Check phase count
	if len(quest.Phases) < 5 || len(quest.Phases) > 10 {
		return fmt.Errorf("quest must have 5-10 phases, has %d", len(quest.Phases))
	}

	// Check for travel phase (cross-server requirement)
	hasTravel := false
	minServers := 0
	for _, phase := range quest.Phases {
		if phase.Type == PhaseTravel && phase.Requirements != nil {
			hasTravel = true
			if phase.Requirements.MinServers >= 3 {
				minServers = phase.Requirements.MinServers
			}
		}
	}
	if !hasTravel || minServers < 3 {
		return fmt.Errorf("quest must require visiting at least 3 servers")
	}

	// Check for rewards
	if quest.Rewards == nil || len(quest.Rewards.Items) == 0 {
		return fmt.Errorf("quest must have legendary item rewards")
	}

	// Check estimated time
	if quest.EstimatedHours < 10.0 || quest.EstimatedHours > 20.0 {
		return fmt.Errorf("quest estimated hours must be 10-20, got %.1f", quest.EstimatedHours)
	}

	return nil
}

// generatePhase creates a single quest phase.
func (g *LegendaryQuestGenerator) generatePhase(rng *rand.Rand, phaseNum int, phaseType PhaseType, params procgen.GenerationParams, template *QuestTemplate) *QuestPhase {
	requirements := NewPhaseRequirements()

	phaseName := g.generatePhaseName(rng, phaseNum, phaseType, params.GenreID)
	description := g.generatePhaseDescription(rng, phaseType, params.GenreID)

	switch phaseType {
	case PhaseKill:
		g.addKillRequirements(rng, requirements, params)
	case PhaseCollect:
		g.addCollectRequirements(rng, requirements, params)
	case PhaseCraft:
		g.addCraftRequirements(rng, requirements, params)
	case PhaseRaid:
		g.addRaidRequirements(rng, requirements, params)
	case PhaseTravel:
		g.addTravelRequirements(rng, requirements, params, template)
	case PhaseExplore:
		g.addExploreRequirements(rng, requirements, params)
	case PhaseTalk:
		g.addDialogueRequirements(rng, requirements, params)
	case PhaseChallenge:
		g.addChallengeRequirements(rng, requirements, params)
	}

	return &QuestPhase{
		PhaseNumber:  phaseNum,
		Name:         phaseName,
		Description:  description,
		Type:         phaseType,
		Requirements: requirements,
		Completed:    false,
	}
}

// addKillRequirements adds enemy kill requirements.
func (g *LegendaryQuestGenerator) addKillRequirements(rng *rand.Rand, req *PhaseRequirements, params procgen.GenerationParams) {
	req.KillTargets = make(map[string]int)

	// Add 2-4 different enemy types
	numTypes := 2 + rng.Intn(3)
	enemyTypes := []string{"dragon", "demon", "undead_lord", "elemental_titan", "void_creature"}

	for i := 0; i < numTypes; i++ {
		enemyType := enemyTypes[rng.Intn(len(enemyTypes))]
		count := 10 + rng.Intn(41) // 10-50
		req.KillTargets[enemyType] = count
	}

	// Add boss kills
	numBosses := 1 + rng.Intn(3) // 1-3 bosses
	req.KillBosses = make([]string, numBosses)
	for i := 0; i < numBosses; i++ {
		req.KillBosses[i] = fmt.Sprintf("legendary_boss_%d_%d", params.Depth, i)
	}
}

// addCollectRequirements adds item collection requirements.
func (g *LegendaryQuestGenerator) addCollectRequirements(rng *rand.Rand, req *PhaseRequirements, params procgen.GenerationParams) {
	req.CollectItems = make(map[string]int)

	// Add 2-5 rare items to collect
	numItems := 2 + rng.Intn(4)
	itemTypes := []string{"ancient_artifact", "elemental_essence", "divine_fragment", "void_shard", "dragon_scale"}

	for i := 0; i < numItems; i++ {
		itemType := itemTypes[rng.Intn(len(itemTypes))]
		count := 5 + rng.Intn(16) // 5-20
		req.CollectItems[itemType] = count
	}
}

// addCraftRequirements adds crafting requirements.
func (g *LegendaryQuestGenerator) addCraftRequirements(rng *rand.Rand, req *PhaseRequirements, params procgen.GenerationParams) {
	// 1-3 legendary items to craft
	numItems := 1 + rng.Intn(3)
	req.CraftItems = make([]CraftRequirement, numItems)

	itemTypes := []string{"weapon", "armor", "accessory", "consumable"}
	qualities := []string{"Advanced", "Master"}

	for i := 0; i < numItems; i++ {
		req.CraftItems[i] = CraftRequirement{
			ItemType:       itemTypes[rng.Intn(len(itemTypes))],
			ItemName:       fmt.Sprintf("legendary_%s_%d", itemTypes[i%len(itemTypes)], i),
			Quantity:       1,
			StationQuality: qualities[rng.Intn(len(qualities))],
			Completed:      false,
		}
	}
}

// addRaidRequirements adds raid encounter requirements.
func (g *LegendaryQuestGenerator) addRaidRequirements(rng *rand.Rand, req *PhaseRequirements, params procgen.GenerationParams) {
	// 1-2 raid encounters
	numRaids := 1 + rng.Intn(2)
	req.RaidEncounters = make([]*RaidRequirement, numRaids)

	tiers := []raids.RaidTier{raids.TierHeroic, raids.TierMythic, raids.TierLegendary}

	for i := 0; i < numRaids; i++ {
		tier := tiers[rng.Intn(len(tiers))]

		raidReq := &RaidRequirement{
			RaidID:       fmt.Sprintf("raid_%d_%d", params.Depth, i),
			RaidName:     fmt.Sprintf("Legendary Raid %d", i+1),
			Tier:         tier,
			BossesToKill: []string{},         // Empty means all bosses
			MinPartySize: 5 + rng.Intn(6),    // 5-10 players
			MaxDeaths:    3 + rng.Intn(8),    // 3-10 deaths allowed
			TimeLimit:    60 + rng.Intn(121), // 60-180 minutes
		}

		// Sometimes require specific bosses
		if rng.Float64() < 0.5 {
			numBosses := 1 + rng.Intn(3)
			raidReq.BossesToKill = make([]string, numBosses)
			for j := 0; j < numBosses; j++ {
				raidReq.BossesToKill[j] = fmt.Sprintf("boss_%d", j+1)
			}
		}

		req.RaidEncounters[i] = raidReq
	}
}

// addTravelRequirements adds cross-server travel requirements.
func (g *LegendaryQuestGenerator) addTravelRequirements(rng *rand.Rand, req *PhaseRequirements, params procgen.GenerationParams, template *QuestTemplate) {
	// Minimum 3 servers per spec, up to 5
	minServers := template.MinServers
	if minServers < 3 {
		minServers = 3
	}
	maxServers := template.MaxServers
	if maxServers < minServers {
		maxServers = 5
	}

	numServers := minServers + rng.Intn(maxServers-minServers+1)
	req.MinServers = numServers
	req.ServersToVisit = make([]string, numServers)

	for i := 0; i < numServers; i++ {
		req.ServersToVisit[i] = fmt.Sprintf("server_%d", i+1)
	}

	// Add specific locations to discover on each server
	numLocations := 2 + rng.Intn(4) // 2-5 locations
	req.LocationsToDiscover = make([]string, numLocations)

	locationTypes := []string{"ancient_temple", "forgotten_ruins", "void_portal", "dragon_lair", "sacred_grove"}
	for i := 0; i < numLocations; i++ {
		locationType := locationTypes[rng.Intn(len(locationTypes))]
		req.LocationsToDiscover[i] = fmt.Sprintf("%s_%d", locationType, i+1)
	}
}

// addExploreRequirements adds exploration requirements.
func (g *LegendaryQuestGenerator) addExploreRequirements(rng *rand.Rand, req *PhaseRequirements, params procgen.GenerationParams) {
	// Add locations to discover
	numLocations := 3 + rng.Intn(5) // 3-7 locations
	req.LocationsToDiscover = make([]string, numLocations)

	for i := 0; i < numLocations; i++ {
		req.LocationsToDiscover[i] = fmt.Sprintf("location_%d", i+1)
	}
}

// addDialogueRequirements adds NPC dialogue requirements.
func (g *LegendaryQuestGenerator) addDialogueRequirements(rng *rand.Rand, req *PhaseRequirements, params procgen.GenerationParams) {
	// 3-5 NPCs to talk to
	numNPCs := 3 + rng.Intn(3)
	req.NPCsToTalk = make([]string, numNPCs)

	npcTypes := []string{"ancient_sage", "dragon_oracle", "void_prophet", "elemental_guardian", "legendary_hero"}
	for i := 0; i < numNPCs; i++ {
		npcType := npcTypes[rng.Intn(len(npcTypes))]
		req.NPCsToTalk[i] = fmt.Sprintf("%s_%d", npcType, i+1)
	}
}

// addChallengeRequirements adds special challenge requirements.
func (g *LegendaryQuestGenerator) addChallengeRequirements(rng *rand.Rand, req *PhaseRequirements, params procgen.GenerationParams) {
	// 1-2 special challenges
	numChallenges := 1 + rng.Intn(2)
	req.Challenges = make([]*Challenge, numChallenges)

	challengeTypes := []ChallengeType{ChallengeSurvival, ChallengePuzzle, ChallengeCombat, ChallengeSpeed, ChallengePerfection}

	for i := 0; i < numChallenges; i++ {
		challengeType := challengeTypes[rng.Intn(len(challengeTypes))]

		req.Challenges[i] = &Challenge{
			ID:          fmt.Sprintf("challenge_%d", i+1),
			Name:        fmt.Sprintf("%s Challenge", challengeType.String()),
			Description: g.generateChallengeDescription(rng, challengeType),
			Type:        challengeType,
			Difficulty:  0.7 + rng.Float64()*0.3, // 0.7-1.0
			Completed:   false,
		}
	}
}

// generateQuestName creates a legendary quest name.
func (g *LegendaryQuestGenerator) generateQuestName(rng *rand.Rand, template *QuestTemplate, genreID string) string {
	prefixes := []string{"The", "An", "A"}
	adjectives := []string{"Ancient", "Legendary", "Eternal", "Forgotten", "Lost", "Sacred", "Cursed", "Divine"}
	nouns := []string{"Legacy", "Prophecy", "Destiny", "Trial", "Journey", "Chronicle", "Saga", "Testament"}

	prefix := prefixes[rng.Intn(len(prefixes))]
	adjective := adjectives[rng.Intn(len(adjectives))]
	noun := nouns[rng.Intn(len(nouns))]

	return fmt.Sprintf("%s %s %s", prefix, adjective, noun)
}

// generateDescription creates quest lore text.
func (g *LegendaryQuestGenerator) generateDescription(rng *rand.Rand, questName, genreID string) string {
	templates := []string{
		"A legendary quest that will test your courage, skill, and determination across multiple worlds.",
		"An epic journey that spans servers and challenges, seeking to uncover ancient secrets.",
		"The ultimate test for heroes willing to face impossible odds and forge their own legend.",
		"A multi-phase odyssey that will require mastery of combat, crafting, and exploration.",
	}

	return templates[rng.Intn(len(templates))]
}

// generatePhaseName creates a phase title.
func (g *LegendaryQuestGenerator) generatePhaseName(rng *rand.Rand, phaseNum int, phaseType PhaseType, genreID string) string {
	return fmt.Sprintf("Phase %d: %s", phaseNum, phaseType.String())
}

// generatePhaseDescription creates phase objective text.
func (g *LegendaryQuestGenerator) generatePhaseDescription(rng *rand.Rand, phaseType PhaseType, genreID string) string {
	templates := map[PhaseType][]string{
		PhaseKill:      {"Defeat legendary enemies to prove your combat prowess.", "Eliminate powerful foes that threaten the realm."},
		PhaseCollect:   {"Gather rare artifacts scattered across the world.", "Collect ancient relics of immense power."},
		PhaseCraft:     {"Forge legendary equipment using master crafting stations.", "Create items of unmatched quality and power."},
		PhaseRaid:      {"Clear challenging raid encounters with your allies.", "Overcome epic boss battles in legendary dungeons."},
		PhaseTravel:    {"Journey across multiple servers to complete your quest.", "Explore distant realms in search of ancient knowledge."},
		PhaseExplore:   {"Discover hidden locations and uncover their secrets.", "Find forgotten places that hold the key to your destiny."},
		PhaseTalk:      {"Seek wisdom from legendary NPCs across the world.", "Consult with ancient beings who hold vital knowledge."},
		PhaseChallenge: {"Complete special trials that test your abilities.", "Overcome unique challenges designed for true heroes."},
	}

	options := templates[phaseType]
	return options[rng.Intn(len(options))]
}

// generateChallengeDescription creates challenge-specific text.
func (g *LegendaryQuestGenerator) generateChallengeDescription(rng *rand.Rand, challengeType ChallengeType) string {
	templates := map[ChallengeType][]string{
		ChallengeSurvival:   {"Survive against overwhelming odds for the required duration."},
		ChallengePuzzle:     {"Solve the ancient puzzle to unlock the path forward."},
		ChallengeCombat:     {"Defeat enemies with specific constraints on your abilities."},
		ChallengeSpeed:      {"Complete the objective within the strict time limit."},
		ChallengePerfection: {"Achieve victory without making any mistakes."},
	}

	options := templates[challengeType]
	return options[rng.Intn(len(options))]
}

// generateRewards creates legendary quest rewards.
func (g *LegendaryQuestGenerator) generateRewards(rng *rand.Rand, params procgen.GenerationParams, template *QuestTemplate) *LegendaryRewards {
	// Generate 1-3 legendary items
	numItems := 1 + rng.Intn(3)
	items := make([]LegendaryItem, numItems)

	for i := 0; i < numItems; i++ {
		// Use existing legendary item generator
		items[i] = LegendaryItem{
			Name:   fmt.Sprintf("Legendary Reward %d", i+1),
			Rarity: RarityLegendary,
		}
	}

	// Generate titles
	titles := []string{
		"Legendary Hero",
		"World Walker",
		"Realm Conqueror",
		"Eternal Champion",
	}
	numTitles := 1 + rng.Intn(2)
	rewardTitles := make([]string, numTitles)
	for i := 0; i < numTitles; i++ {
		rewardTitles[i] = titles[rng.Intn(len(titles))]
	}

	// Generate gold reward (100k-500k)
	gold := 100000 + rng.Intn(400001)

	// Generate XP reward
	experience := 50000 + rng.Intn(50001)

	// Prestige levels (1-3)
	prestigeLevels := 1 + rng.Intn(3)

	// Achievements
	achievements := []string{
		"Completed Legendary Quest",
		"World Traveler",
		"Raid Master",
		"Master Craftsman",
	}

	// Cosmetics
	cosmetics := []string{
		"Legendary Aura",
		"Epic Mount",
		"Unique Visual Effect",
	}
	numCosmetics := 1 + rng.Intn(3)
	rewardCosmetics := make([]string, numCosmetics)
	for i := 0; i < numCosmetics; i++ {
		rewardCosmetics[i] = cosmetics[rng.Intn(len(cosmetics))]
	}

	return &LegendaryRewards{
		Items:          items,
		Titles:         rewardTitles,
		Gold:           gold,
		Experience:     experience,
		PrestigeLevels: prestigeLevels,
		Achievements:   achievements,
		Cosmetics:      rewardCosmetics,
	}
}

// defaultQuestTemplates returns the standard legendary quest templates.
func defaultQuestTemplates() []*QuestTemplate {
	return []*QuestTemplate{
		{
			NamePattern:       "epic",
			MinPhases:         5,
			MaxPhases:         7,
			PhaseTypes:        []PhaseType{PhaseKill, PhaseCollect, PhaseCraft, PhaseTravel, PhaseExplore},
			MinServers:        3,
			MaxServers:        4,
			RequiresRaid:      false,
			RequiresCrafting:  true,
			EstimatedHoursMin: 10.0,
			EstimatedHoursMax: 15.0,
			RewardTier:        1,
		},
		{
			NamePattern:       "mythic",
			MinPhases:         7,
			MaxPhases:         9,
			PhaseTypes:        []PhaseType{PhaseKill, PhaseCollect, PhaseCraft, PhaseRaid, PhaseTravel, PhaseChallenge},
			MinServers:        4,
			MaxServers:        5,
			RequiresRaid:      true,
			RequiresCrafting:  true,
			EstimatedHoursMin: 15.0,
			EstimatedHoursMax: 20.0,
			RewardTier:        2,
		},
		{
			NamePattern:       "legendary",
			MinPhases:         8,
			MaxPhases:         10,
			PhaseTypes:        []PhaseType{PhaseKill, PhaseCollect, PhaseCraft, PhaseRaid, PhaseTravel, PhaseTalk, PhaseChallenge},
			MinServers:        5,
			MaxServers:        5,
			RequiresRaid:      true,
			RequiresCrafting:  true,
			EstimatedHoursMin: 15.0,
			EstimatedHoursMax: 20.0,
			RewardTier:        3,
		},
	}
}
