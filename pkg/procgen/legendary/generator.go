package legendary

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/opd-ai/venture/pkg/procgen"
	"github.com/opd-ai/venture/pkg/procgen/entity"
	"github.com/opd-ai/venture/pkg/procgen/genre"
	"github.com/opd-ai/venture/pkg/world/raids"
)

// Generator generates procedural legendary quests.
type Generator struct {
	baseSeed int64
	registry *genre.Registry
}

// NewGenerator creates a new legendary quest generator.
func NewGenerator() *Generator {
	return &Generator{
		baseSeed: time.Now().UnixNano(),
		registry: genre.DefaultRegistry(),
	}
}

// Generate implements the procgen.Generator interface.
func (g *Generator) Generate(seed int64, params procgen.GenerationParams) (interface{}, error) {
	if err := g.validateParams(params); err != nil {
		return nil, err
	}

	rng := rand.New(rand.NewSource(seed))
	genreID, err := g.getValidatedGenreID(params.GenreID)
	if err != nil {
		return nil, err
	}

	playerLevel, serversVisited := g.extractCustomParams(params.Custom)
	quest := g.createBaseQuest(seed, rng, genreID, playerLevel, serversVisited)

	g.populateQuestPhases(quest, rng, genreID, serversVisited, params)
	g.populateQuestRewards(quest, rng, genreID)

	return quest, nil
}

func (g *Generator) validateParams(params procgen.GenerationParams) error {
	if params.Difficulty < 0.0 || params.Difficulty > 1.0 {
		return fmt.Errorf("difficulty must be 0.0-1.0, got %.2f", params.Difficulty)
	}
	if params.Depth < 1 {
		return fmt.Errorf("depth must be >= 1, got %d", params.Depth)
	}
	return nil
}

func (g *Generator) getValidatedGenreID(genreID string) (string, error) {
	if genreID == "" {
		genreID = "fantasy"
	}
	if _, err := g.registry.Get(genreID); err != nil {
		return "", fmt.Errorf("invalid genre: %s", genreID)
	}
	return genreID, nil
}

func (g *Generator) extractCustomParams(custom map[string]interface{}) (int, int) {
	playerLevel := 50
	if lvl, ok := custom["player_level"]; ok {
		if l, ok := lvl.(int); ok {
			playerLevel = l
		}
	}

	serversVisited := 3
	if sv, ok := custom["servers_visited"]; ok {
		if s, ok := sv.(int); ok {
			serversVisited = s
			if serversVisited < 3 {
				serversVisited = 3
			}
			if serversVisited > 5 {
				serversVisited = 5
			}
		}
	}

	return playerLevel, serversVisited
}

func (g *Generator) createBaseQuest(seed int64, rng *rand.Rand, genreID string, playerLevel, serversVisited int) *LegendaryQuest {
	return &LegendaryQuest{
		ID:              fmt.Sprintf("legendary_%d", seed),
		Name:            g.generateQuestName(rng, genreID),
		Description:     g.generateDescription(rng, genreID),
		Lore:            g.generateLore(rng, genreID),
		MinLevel:        playerLevel,
		EstimatedHours:  10 + rng.Intn(11),
		ServersRequired: serversVisited,
		RaidsRequired:   make([]string, 0),
		Phases:          make([]*QuestPhase, 0),
		Rewards:         make([]*LegendaryReward, 0),
		CreatedAt:       time.Now(),
	}
}

func (g *Generator) populateQuestPhases(quest *LegendaryQuest, rng *rand.Rand, genreID string, serversVisited int, params procgen.GenerationParams) {
	numPhases := 5 + rng.Intn(6)
	for i := 0; i < numPhases; i++ {
		phase := g.generatePhase(rng, i, numPhases, genreID, serversVisited, params)
		quest.Phases = append(quest.Phases, phase)

		if phase.Type == PhaseRaid && phase.RaidID != "" {
			g.trackRaidRequirement(quest, phase.RaidID)
		}
	}
}

func (g *Generator) trackRaidRequirement(quest *LegendaryQuest, raidID string) {
	for _, r := range quest.RaidsRequired {
		if r == raidID {
			return
		}
	}
	quest.RaidsRequired = append(quest.RaidsRequired, raidID)
}

func (g *Generator) populateQuestRewards(quest *LegendaryQuest, rng *rand.Rand, genreID string) {
	numRewards := 1 + rng.Intn(3)
	for i := 0; i < numRewards; i++ {
		reward := g.generateReward(rng, i, genreID, quest.Name)
		quest.Rewards = append(quest.Rewards, reward)
	}
}

// Validate implements the procgen.Generator interface.
func (g *Generator) Validate(result interface{}) error {
	quest, ok := result.(*LegendaryQuest)
	if !ok {
		return fmt.Errorf("expected *LegendaryQuest, got %T", result)
	}

	// Validate basic fields
	if quest.ID == "" {
		return fmt.Errorf("quest ID is empty")
	}
	if quest.Name == "" {
		return fmt.Errorf("quest name is empty")
	}
	if quest.MinLevel < 1 {
		return fmt.Errorf("min level must be >= 1, got %d", quest.MinLevel)
	}
	if quest.EstimatedHours < 10 || quest.EstimatedHours > 20 {
		return fmt.Errorf("estimated hours must be 10-20, got %d", quest.EstimatedHours)
	}
	if quest.ServersRequired < 3 || quest.ServersRequired > 5 {
		return fmt.Errorf("servers required must be 3-5, got %d", quest.ServersRequired)
	}

	// Validate phases
	if len(quest.Phases) < 5 || len(quest.Phases) > 10 {
		return fmt.Errorf("quest must have 5-10 phases, got %d", len(quest.Phases))
	}

	// Validate last phase is final type
	if quest.Phases[len(quest.Phases)-1].Type != PhaseFinal {
		return fmt.Errorf("last phase must be PhaseFinal, got %s", quest.Phases[len(quest.Phases)-1].Type)
	}

	// Validate rewards
	if len(quest.Rewards) < 1 || len(quest.Rewards) > 3 {
		return fmt.Errorf("quest must have 1-3 rewards, got %d", len(quest.Rewards))
	}

	// Validate at least one reward is unique
	hasUnique := false
	for _, r := range quest.Rewards {
		if r.IsUnique {
			hasUnique = true
			break
		}
	}
	if !hasUnique {
		return fmt.Errorf("quest must have at least one unique reward")
	}

	return nil
}

// generateQuestName generates a quest name based on genre.
func (g *Generator) generateQuestName(rng *rand.Rand, genreID string) string {
	prefixes := map[string][]string{
		"fantasy":         {"The Legend of", "The Chronicles of", "The Saga of", "The Tale of", "The Quest for"},
		"scifi":           {"Mission", "Operation", "Project", "Initiative", "Protocol"},
		"horror":          {"The Curse of", "The Horror of", "The Nightmare of", "The Terror of", "The Haunting of"},
		"cyberpunk":       {"The Network", "The System", "The Matrix", "The Grid", "The Web"},
		"postapocalyptic": {"The Wasteland", "The Ruins of", "The Ashes of", "The Remnants of", "The Last"},
	}

	suffixes := map[string][]string{
		"fantasy":         {"the Eternal", "the Forgotten", "the Ancient", "the Lost", "the Legendary"},
		"scifi":           {"Omega", "Alpha", "Zero", "Prime", "Genesis"},
		"horror":          {"the Damned", "the Cursed", "the Forsaken", "the Haunted", "the Dead"},
		"cyberpunk":       {"Override", "Breach", "Hack", "Exploit", "Intrusion"},
		"postapocalyptic": {"Survivor", "Haven", "Sanctuary", "Frontier", "Horizon"},
	}

	prefix := prefixes[genreID][rng.Intn(len(prefixes[genreID]))]
	suffix := suffixes[genreID][rng.Intn(len(suffixes[genreID]))]

	return fmt.Sprintf("%s %s", prefix, suffix)
}

// generateDescription generates a quest description.
func (g *Generator) generateDescription(rng *rand.Rand, genreID string) string {
	templates := map[string][]string{
		"fantasy": {
			"Seek out the ancient artifact that can restore balance to the realm.",
			"Uncover the truth behind the legendary hero's disappearance.",
			"Unite the scattered forces to stand against an ancient evil.",
		},
		"scifi": {
			"Investigate the anomaly threatening the fabric of space-time.",
			"Recover critical data from across the galactic network.",
			"Establish contact with the lost colony beyond known space.",
		},
		"horror": {
			"Survive the nightmare that haunts the cursed lands.",
			"Unravel the mystery of the disappearances.",
			"Confront the entity that feeds on fear itself.",
		},
		"cyberpunk": {
			"Expose the corporation's darkest secrets.",
			"Hack into the most secure systems in the megacity.",
			"Unite the underground resistance against oppression.",
		},
		"postapocalyptic": {
			"Find the legendary sanctuary that offers hope.",
			"Recover the technology that could rebuild civilization.",
			"Lead the survivors to a new beginning.",
		},
	}

	descriptions := templates[genreID]
	return descriptions[rng.Intn(len(descriptions))]
}

// generateLore generates quest lore text.
func (g *Generator) generateLore(rng *rand.Rand, genreID string) string {
	templates := map[string][]string{
		"fantasy": {
			"Long ago, when the world was young, a great power was sealed away. Now, as darkness rises once more, only those brave enough to seek the truth can prevent catastrophe.",
			"The ancient texts speak of a chosen one who would arise in the darkest hour. Whether prophecy or chance, the time has come to discover your destiny.",
		},
		"scifi": {
			"According to historical records, an event occurred that changed the course of humanity. The data was classified, the truth buried. Until now.",
			"Beyond the explored sectors lies a mystery that could revolutionize our understanding of the universe. But some secrets come at a price.",
		},
		"horror": {
			"They say the dead don't rest easy in these lands. The tales passed down through generations speak of a curse that cannot be broken. But every curse has an origin.",
			"Madness spreads like a plague, consuming all who venture too close. Yet within the darkness lies the key to understanding what lurks beyond mortal comprehension.",
		},
		"cyberpunk": {
			"In a world where information is power, someone discovered something that could tear down the entire system. They were silenced, but their legacy remains.",
			"The megacorporations rule with an iron fist, but cracks are forming in their perfect facade. You've stumbled upon evidence that could change everything.",
		},
		"postapocalyptic": {
			"Before the fall, humanity reached heights unimaginable. Now, scattered survivors search the ruins for any trace of what was lost.",
			"Legends speak of a place untouched by the catastrophe, where the old world still thrives. Most dismiss it as wishful thinking. You're about to find out the truth.",
		},
	}

	lore := templates[genreID]
	return lore[rng.Intn(len(lore))]
}

// generatePhase generates a single quest phase.
func (g *Generator) generatePhase(rng *rand.Rand, index, total int, genreID string, serversRequired int, params procgen.GenerationParams) *QuestPhase {
	phase := &QuestPhase{
		ID: fmt.Sprintf("phase_%d", index),
	}

	// Last phase is always final
	if index == total-1 {
		phase.Type = PhaseFinal
		phase.Name = "The Final Challenge"
		phase.Description = "Face the ultimate test to claim your legendary reward."
		phase.XPReward = 10000 + rng.Intn(5000)
		phase.GoldReward = 5000 + rng.Intn(5000)
		return phase
	}

	// Distribute phase types
	phaseTypes := []PhaseType{PhaseExploration, PhaseCombat, PhaseCrafting, PhaseCollection, PhaseRaid}
	phase.Type = phaseTypes[rng.Intn(len(phaseTypes))]

	// Generate phase-specific content
	switch phase.Type {
	case PhaseExploration:
		serverIndex := rng.Intn(serversRequired)
		phase.Name = fmt.Sprintf("Journey to Server %d", serverIndex+1)
		phase.Description = "Travel to a distant server and discover its secrets."
		phase.ServerID = fmt.Sprintf("server_%d", serverIndex)
		phase.LocationX = rng.Intn(1000)
		phase.LocationY = rng.Intn(1000)
		phase.XPReward = 1000 + rng.Intn(500)
		phase.GoldReward = 500 + rng.Intn(500)

	case PhaseCombat:
		phase.Name = "Defeat the Champion"
		phase.Description = "Defeat a powerful enemy and prove your worth."
		phase.BossName = fmt.Sprintf("Champion of %s", genreID)
		phase.EntityType = entity.TypeBoss
		phase.KillCount = 1
		phase.XPReward = 2000 + rng.Intn(1000)
		phase.GoldReward = 1000 + rng.Intn(1000)

	case PhaseCrafting:
		phase.Name = "Forge the Artifact"
		phase.Description = "Create a powerful item using master crafting techniques."
		phase.RecipeID = fmt.Sprintf("legendary_recipe_%d", index)
		phase.ItemName = fmt.Sprintf("Legendary %s", genreID)
		phase.Quantity = 1
		phase.StationTier = 4 // Master tier
		phase.XPReward = 1500 + rng.Intn(500)
		phase.GoldReward = 0 // Crafting costs gold instead

	case PhaseCollection:
		numMaterials := 3 + rng.Intn(3) // 3-5 materials
		phase.Name = "Gather the Components"
		phase.Description = "Collect rare materials from across the servers."
		phase.MaterialIDs = make([]string, numMaterials)
		phase.Quantities = make([]int, numMaterials)
		for i := 0; i < numMaterials; i++ {
			phase.MaterialIDs[i] = fmt.Sprintf("material_%d_%d", index, i)
			phase.Quantities[i] = 5 + rng.Intn(16) // 5-20
		}
		phase.XPReward = 1000 + rng.Intn(500)
		phase.GoldReward = 500 + rng.Intn(500)

	case PhaseRaid:
		tiers := []raids.RaidTier{raids.TierNormal, raids.TierHeroic, raids.TierMythic}
		tier := tiers[rng.Intn(len(tiers))]
		phase.Name = fmt.Sprintf("Clear the %s Raid", tier.String())
		phase.Description = "Defeat the raid bosses to prove your strength."
		phase.RaidID = fmt.Sprintf("raid_%d", index)
		phase.RaidTier = tier
		phase.BossIndex = rng.Intn(5) // 0-4, raids have 3-5 bosses
		phase.XPReward = 3000 + rng.Intn(2000)
		phase.GoldReward = 2000 + rng.Intn(2000)
	}

	return phase
}

// generateReward generates a legendary reward.
func (g *Generator) generateReward(rng *rand.Rand, index int, genreID, questName string) *LegendaryReward {
	rewardTypes := []RewardType{RewardItem, RewardTitle, RewardMount, RewardCompanion, RewardAchievement, RewardAccountBonus}
	rewardType := rewardTypes[rng.Intn(len(rewardTypes))]

	reward := &LegendaryReward{
		ID:       fmt.Sprintf("reward_%d", index),
		Type:     rewardType,
		IsUnique: true, // All legendary rewards are unique
	}

	switch rewardType {
	case RewardItem:
		reward.Name = fmt.Sprintf("Legendary %s Weapon", genreID)
		reward.Description = "A powerful artifact of immense power."
		reward.ItemID = fmt.Sprintf("legendary_item_%d", index)

	case RewardTitle:
		reward.Name = "Hero of Legend"
		reward.Description = fmt.Sprintf("Earned by completing: %s", questName)
		reward.Title = fmt.Sprintf("<%s>", reward.Name)

	case RewardMount:
		reward.Name = fmt.Sprintf("Legendary %s Mount", genreID)
		reward.Description = "A rare and majestic companion for your journey."
		reward.MountID = fmt.Sprintf("legendary_mount_%d", index)

	case RewardCompanion:
		reward.Name = fmt.Sprintf("Legendary %s Companion", genreID)
		reward.Description = "A powerful ally to fight by your side."
		reward.CompanionID = fmt.Sprintf("legendary_companion_%d", index)

	case RewardAchievement:
		reward.Name = "Legendary Achiever"
		reward.Description = "A permanent mark of your legendary accomplishment."
		reward.AchievementID = fmt.Sprintf("legendary_achievement_%d", index)

	case RewardAccountBonus:
		reward.Name = "Legendary Mastery"
		reward.Description = "All characters on this account gain a permanent bonus."
		reward.AccountBonusID = fmt.Sprintf("legendary_bonus_%d", index)
		reward.BonusPercent = 5.0 + float64(rng.Intn(6)) // 5-10%
	}

	return reward
}
