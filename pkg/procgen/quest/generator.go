// Package quest provides procedural quest generation.
// This file implements quest generators for main story, side quests,
// and dynamic objectives with rewards.
package quest

import (
	"fmt"
	"math/rand"

	"github.com/opd-ai/venture/pkg/procgen"
	"github.com/sirupsen/logrus"
)

// QuestGenerator implements the Generator interface for procedural quest creation.
type QuestGenerator struct {
	logger *logrus.Entry
}

// NewQuestGenerator creates a new quest generator.
func NewQuestGenerator() *QuestGenerator {
	return NewQuestGeneratorWithLogger(nil)
}

// NewQuestGeneratorWithLogger creates a new quest generator with a logger.
func NewQuestGeneratorWithLogger(logger *logrus.Logger) *QuestGenerator {
	var logEntry *logrus.Entry
	if logger != nil {
		logEntry = logger.WithFields(logrus.Fields{
			"generator": "quest",
		})
	}
	return &QuestGenerator{
		logger: logEntry,
	}
}

// Generate creates quests based on the seed and parameters.
// Returns []*Quest or error.
func (g *QuestGenerator) Generate(seed int64, params procgen.GenerationParams) (interface{}, error) {
	if g.logger != nil && g.logger.Logger.GetLevel() >= logrus.DebugLevel {
		g.logger.WithFields(logrus.Fields{
			"seed":       seed,
			"genreID":    params.GenreID,
			"depth":      params.Depth,
			"difficulty": params.Difficulty,
		}).Debug("starting quest generation")
	}

	// Validate parameters
	if params.Depth < 0 {
		err := fmt.Errorf("depth must be non-negative")
		if g.logger != nil {
			g.logger.WithError(err).WithField("depth", params.Depth).Error("invalid depth parameter")
		}
		return nil, err
	}
	if params.Difficulty < 0 || params.Difficulty > 1 {
		err := fmt.Errorf("difficulty must be between 0 and 1")
		if g.logger != nil {
			g.logger.WithError(err).WithField("difficulty", params.Difficulty).Error("invalid difficulty parameter")
		}
		return nil, err
	}

	// Extract custom parameters
	count := 5 // default
	if c, ok := params.Custom["count"].(int); ok {
		count = c
	}

	// Create deterministic RNG
	rng := rand.New(rand.NewSource(seed))

	// Get templates based on genre
	var templates []QuestTemplate
	switch params.GenreID {
	case "scifi":
		templates = append(templates, GetSciFiKillTemplates()...)
		templates = append(templates, GetSciFiCollectTemplates()...)
		templates = append(templates, GetSciFiBossTemplates()...)
	case "fantasy":
		fallthrough
	default:
		templates = append(templates, GetFantasyKillTemplates()...)
		templates = append(templates, GetFantasyCollectTemplates()...)
		templates = append(templates, GetFantasyBossTemplates()...)
		templates = append(templates, GetFantasyExploreTemplates()...)
	}

	if len(templates) == 0 {
		err := fmt.Errorf("no templates available for genre: %s", params.GenreID)
		if g.logger != nil {
			g.logger.WithError(err).WithField("genreID", params.GenreID).Error("template selection failed")
		}
		return nil, err
	}

	if g.logger != nil && g.logger.Logger.GetLevel() >= logrus.DebugLevel {
		g.logger.WithFields(logrus.Fields{
			"count":         count,
			"templateCount": len(templates),
		}).Debug("generating quests")
	}

	// Generate quests
	quests := make([]*Quest, 0, count)
	for i := 0; i < count; i++ {
		// Select random template
		template := templates[rng.Intn(len(templates))]

		// Generate quest from template
		quest := g.generateFromTemplate(rng, template, params, i)
		quest.Seed = seed + int64(i)

		quests = append(quests, quest)

		if g.logger != nil && g.logger.Logger.GetLevel() >= logrus.DebugLevel {
			g.logger.WithFields(logrus.Fields{
				"questIndex": i,
				"questName":  quest.Name,
				"questType":  quest.Type,
				"difficulty": quest.Difficulty,
			}).Debug("quest generated")
		}
	}

	if g.logger != nil {
		g.logger.WithFields(logrus.Fields{
			"questCount": len(quests),
			"seed":       seed,
		}).Info("quest generation complete")
	}

	return quests, nil
}

// generateFromTemplate creates a single quest from a template.
func (g *QuestGenerator) generateFromTemplate(rng *rand.Rand, template QuestTemplate, params procgen.GenerationParams, index int) *Quest {
	quest := &Quest{
		Type:   template.BaseType,
		Status: StatusNotStarted,
		Tags:   make([]string, len(template.Tags)),
	}
	copy(quest.Tags, template.Tags)

	// Basic quest properties
	quest.ID = fmt.Sprintf("quest_%d_%d", params.Depth, index)
	quest.Difficulty = g.determineDifficulty(rng, params.Depth, params.Difficulty)
	quest.Name = g.generateQuestName(rng, template)

	// Generate objectives
	targetType := template.TargetTypes[rng.Intn(len(template.TargetTypes))]
	objective := g.generateObjective(rng, template, params, targetType)
	quest.Objectives = []Objective{objective}

	// Generate description
	quest.Description = g.generateQuestDescription(rng, template, params, targetType, objective.Required)

	// Generate rewards
	depthScale := 1.0 + float64(params.Depth)*0.15
	g.generateRewards(rng, quest, template, depthScale)

	// Optional properties
	quest.RequiredLevel = 1 + params.Depth
	g.setOptionalProperties(rng, quest, template, targetType)

	return quest
}

// generateQuestName creates a quest name from template prefixes and suffixes.
func (g *QuestGenerator) generateQuestName(rng *rand.Rand, template QuestTemplate) string {
	prefix := template.NamePrefixes[rng.Intn(len(template.NamePrefixes))]
	suffix := template.NameSuffixes[rng.Intn(len(template.NameSuffixes))]
	return fmt.Sprintf("%s %s", prefix, suffix)
}

// generateObjective creates a quest objective with scaling based on parameters.
func (g *QuestGenerator) generateObjective(rng *rand.Rand, template QuestTemplate, params procgen.GenerationParams, targetType string) Objective {
	depthScale := 1.0 + float64(params.Depth)*0.15
	difficultyScale := 0.7 + params.Difficulty*0.6

	required := g.calculateRequiredAmount(rng, template.RequiredRange, difficultyScale, depthScale)

	objective := Objective{
		Target:      targetType,
		Required:    required,
		Current:     0,
		Description: g.generateObjectiveDescription(template.BaseType, targetType, required),
	}

	return objective
}

// calculateRequiredAmount computes the scaled required count for quest objectives.
func (g *QuestGenerator) calculateRequiredAmount(rng *rand.Rand, reqRange [2]int, difficultyScale, depthScale float64) int {
	minRequired := int(float64(reqRange[0]) * difficultyScale)
	maxRequired := int(float64(reqRange[1]) * difficultyScale * depthScale)

	if minRequired < 1 {
		minRequired = 1
	}
	if maxRequired < minRequired {
		maxRequired = minRequired
	}

	if maxRequired > minRequired {
		return minRequired + rng.Intn(maxRequired-minRequired+1)
	}
	return minRequired
}

// generateObjectiveDescription creates a description for a quest objective.
func (g *QuestGenerator) generateObjectiveDescription(questType QuestType, targetType string, required int) string {
	switch questType {
	case TypeKill:
		return fmt.Sprintf("Defeat %d %s", required, targetType)
	case TypeCollect:
		return fmt.Sprintf("Collect %d %s", required, targetType)
	case TypeBoss:
		return fmt.Sprintf("Defeat %s", targetType)
	case TypeExplore:
		return fmt.Sprintf("Discover %s", targetType)
	case TypeEscort:
		return fmt.Sprintf("Escort %s safely", targetType)
	case TypeTalk:
		return fmt.Sprintf("Speak with %s", targetType)
	default:
		return fmt.Sprintf("Complete objective with %s", targetType)
	}
}

// generateQuestDescription creates the quest description with genre-aware formatting.
func (g *QuestGenerator) generateQuestDescription(rng *rand.Rand, template QuestTemplate, params procgen.GenerationParams, targetType string, required int) string {
	descIdx := rng.Intn(len(template.DescTemplates))
	descTemplate := template.DescTemplates[descIdx]

	switch template.BaseType {
	case TypeKill:
		if params.GenreID == "scifi" && descIdx == 2 {
			return fmt.Sprintf(descTemplate, required, targetType)
		}
		return fmt.Sprintf(descTemplate, targetType, required)
	case TypeCollect:
		if (params.GenreID == "fantasy" && descIdx == 2) || (params.GenreID == "scifi" && descIdx == 1) {
			return fmt.Sprintf(descTemplate, targetType, required)
		}
		return fmt.Sprintf(descTemplate, required, targetType)
	case TypeBoss, TypeExplore, TypeEscort, TypeTalk:
		return fmt.Sprintf(descTemplate, targetType)
	default:
		return fmt.Sprintf(descTemplate, targetType, required)
	}
}

// generateRewards calculates and assigns quest rewards based on scaling factors.
func (g *QuestGenerator) generateRewards(rng *rand.Rand, quest *Quest, template QuestTemplate, depthScale float64) {
	rarityMultiplier := 1.0 + float64(quest.Difficulty)*0.3

	// XP rewards
	minXP := int(float64(template.XPRewardRange[0]) * depthScale * rarityMultiplier)
	maxXP := int(float64(template.XPRewardRange[1]) * depthScale * rarityMultiplier)
	quest.Reward.XP = g.randomInRange(rng, minXP, maxXP)

	// Gold rewards
	minGold := int(float64(template.GoldRewardRange[0]) * depthScale * rarityMultiplier)
	maxGold := int(float64(template.GoldRewardRange[1]) * depthScale * rarityMultiplier)
	quest.Reward.Gold = g.randomInRange(rng, minGold, maxGold)

	// Item rewards
	if rng.Float64() < template.ItemRewardChance {
		numItems := 1 + rng.Intn(2)
		quest.Reward.Items = make([]string, numItems)
		for i := 0; i < numItems; i++ {
			quest.Reward.Items[i] = fmt.Sprintf("item_%s_%d", quest.Difficulty.String(), i)
		}
	}

	// Skill point rewards
	if rng.Float64() < template.SkillPointChance {
		quest.Reward.SkillPoints = 1 + rng.Intn(2)
	}
}

// randomInRange returns a random value between min and max inclusive.
func (g *QuestGenerator) randomInRange(rng *rand.Rand, min, max int) int {
	if max > min {
		return min + rng.Intn(max-min+1)
	}
	return min
}

// setOptionalProperties sets location and quest giver based on quest type.
func (g *QuestGenerator) setOptionalProperties(rng *rand.Rand, quest *Quest, template QuestTemplate, targetType string) {
	if template.BaseType == TypeExplore || template.BaseType == TypeBoss {
		quest.Location = targetType
	}

	if template.BaseType != TypeExplore {
		giverNames := []string{"Elder", "Captain", "Merchant", "Wizard", "Guard", "Scout", "Leader"}
		quest.GiverNPC = giverNames[rng.Intn(len(giverNames))]
	}
}

// determineDifficulty calculates quest difficulty based on depth and parameters.
func (g *QuestGenerator) determineDifficulty(rng *rand.Rand, depth int, difficulty float64) Difficulty {
	// Base difficulty on depth
	baseLevel := depth / 3

	// Add difficulty parameter influence
	baseLevel += int(difficulty * 2)

	// Add random variance (-1 to +1)
	variance := rng.Intn(3) - 1
	level := baseLevel + variance

	// Clamp to valid range
	if level < 0 {
		return DifficultyTrivial
	}
	if level > int(DifficultyLegendary) {
		return DifficultyLegendary
	}

	return Difficulty(level)
}

// Validate checks if the generated quests are valid.
func (g *QuestGenerator) Validate(result interface{}) error {
	quests, ok := result.([]*Quest)
	if !ok {
		return fmt.Errorf("expected []*Quest, got %T", result)
	}

	if len(quests) == 0 {
		return fmt.Errorf("no quests generated")
	}

	for i, quest := range quests {
		if quest == nil {
			return fmt.Errorf("quest %d is nil", i)
		}

		if quest.Name == "" {
			return fmt.Errorf("quest %d has empty name", i)
		}

		if quest.Description == "" {
			return fmt.Errorf("quest %d has empty description", i)
		}

		if len(quest.Objectives) == 0 {
			return fmt.Errorf("quest %d has no objectives", i)
		}

		for j, obj := range quest.Objectives {
			if obj.Description == "" {
				return fmt.Errorf("quest %d objective %d has empty description", i, j)
			}
			if obj.Required <= 0 {
				return fmt.Errorf("quest %d objective %d has invalid required amount: %d", i, j, obj.Required)
			}
		}

		if quest.Reward.XP <= 0 {
			return fmt.Errorf("quest %d has no XP reward", i)
		}

		if quest.RequiredLevel < 0 {
			return fmt.Errorf("quest %d has negative required level", i)
		}
	}

	return nil
}
