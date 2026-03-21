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
	g.logGenerationStart(seed, params)

	if err := g.validateParams(params); err != nil {
		return nil, err
	}

	count := g.extractQuestCount(params)
	rng := rand.New(rand.NewSource(seed))
	templates, err := g.selectTemplates(params.GenreID)
	if err != nil {
		return nil, err
	}

	g.logTemplateSelection(count, len(templates))
	quests := g.generateQuests(rng, templates, params, seed, count)
	g.logGenerationComplete(len(quests), seed)

	return quests, nil
}

// validateParams validates generation parameters.
func (g *QuestGenerator) validateParams(params procgen.GenerationParams) error {
	if params.Depth < 0 {
		err := fmt.Errorf("depth must be non-negative")
		if g.logger != nil {
			g.logger.WithError(err).WithField("depth", params.Depth).Error("invalid depth parameter")
		}
		return err
	}
	if params.Difficulty < 0 || params.Difficulty > 1 {
		err := fmt.Errorf("difficulty must be between 0 and 1")
		if g.logger != nil {
			g.logger.WithError(err).WithField("difficulty", params.Difficulty).Error("invalid difficulty parameter")
		}
		return err
	}
	return nil
}

// extractQuestCount extracts the quest count from custom parameters.
func (g *QuestGenerator) extractQuestCount(params procgen.GenerationParams) int {
	count := 5
	if c, ok := params.Custom["count"].(int); ok {
		count = c
	}
	return count
}

// selectTemplates selects quest templates based on genre.
func (g *QuestGenerator) selectTemplates(genreID string) ([]QuestTemplate, error) {
	var templates []QuestTemplate
	switch genreID {
	case "scifi":
		templates = append(templates, GetSciFiKillTemplates()...)
		templates = append(templates, GetSciFiCollectTemplates()...)
		templates = append(templates, GetSciFiBossTemplates()...)
		templates = append(templates, GetSciFiExploreTemplates()...)
	case "horror":
		templates = append(templates, GetHorrorKillTemplates()...)
		templates = append(templates, GetHorrorCollectTemplates()...)
		templates = append(templates, GetHorrorBossTemplates()...)
		templates = append(templates, GetHorrorExploreTemplates()...)
	case "cyberpunk":
		templates = append(templates, GetCyberpunkKillTemplates()...)
		templates = append(templates, GetCyberpunkCollectTemplates()...)
		templates = append(templates, GetCyberpunkBossTemplates()...)
		templates = append(templates, GetCyberpunkExploreTemplates()...)
	case "postapoc":
		templates = append(templates, GetPostApocKillTemplates()...)
		templates = append(templates, GetPostApocCollectTemplates()...)
		templates = append(templates, GetPostApocBossTemplates()...)
		templates = append(templates, GetPostApocExploreTemplates()...)
	case "fantasy":
		fallthrough
	default:
		templates = append(templates, GetFantasyKillTemplates()...)
		templates = append(templates, GetFantasyCollectTemplates()...)
		templates = append(templates, GetFantasyBossTemplates()...)
		templates = append(templates, GetFantasyExploreTemplates()...)
	}

	if len(templates) == 0 {
		err := fmt.Errorf("no templates available for genre: %s", genreID)
		if g.logger != nil {
			g.logger.WithError(err).WithField("genreID", genreID).Error("template selection failed")
		}
		return nil, err
	}
	return templates, nil
}

// generateQuests generates multiple quests from templates.
func (g *QuestGenerator) generateQuests(rng *rand.Rand, templates []QuestTemplate, params procgen.GenerationParams, seed int64, count int) []*Quest {
	quests := make([]*Quest, count)
	for i := 0; i < count; i++ {
		template := templates[rng.Intn(len(templates))]
		quest := g.generateFromTemplate(rng, template, params, i)
		quest.Seed = seed + int64(i)
		quests[i] = quest
		g.logQuestGenerated(i, quest)
	}
	return quests
}

// logGenerationStart logs the start of quest generation.
func (g *QuestGenerator) logGenerationStart(seed int64, params procgen.GenerationParams) {
	if g.logger != nil && g.logger.Logger.GetLevel() >= logrus.DebugLevel {
		g.logger.WithFields(logrus.Fields{
			"seed":       seed,
			"genreID":    params.GenreID,
			"depth":      params.Depth,
			"difficulty": params.Difficulty,
		}).Debug("starting quest generation")
	}
}

// logTemplateSelection logs template selection details.
func (g *QuestGenerator) logTemplateSelection(count, templateCount int) {
	if g.logger != nil && g.logger.Logger.GetLevel() >= logrus.DebugLevel {
		g.logger.WithFields(logrus.Fields{
			"count":         count,
			"templateCount": templateCount,
		}).Debug("generating quests")
	}
}

// logQuestGenerated logs details of a generated quest.
func (g *QuestGenerator) logQuestGenerated(index int, quest *Quest) {
	if g.logger != nil && g.logger.Logger.GetLevel() >= logrus.DebugLevel {
		g.logger.WithFields(logrus.Fields{
			"questIndex": index,
			"questName":  quest.Name,
			"questType":  quest.Type,
			"difficulty": quest.Difficulty,
		}).Debug("quest generated")
	}
}

// logGenerationComplete logs completion of quest generation.
func (g *QuestGenerator) logGenerationComplete(questCount int, seed int64) {
	if g.logger != nil {
		g.logger.WithFields(logrus.Fields{
			"questCount": questCount,
			"seed":       seed,
		}).Info("quest generation complete")
	}
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

	// Generate objectives (3-5 objectives per quest for quality requirements)
	numObjectives := 3 + rng.Intn(3) // 3-5 objectives
	quest.Objectives = make([]Objective, numObjectives)
	for i := 0; i < numObjectives; i++ {
		targetType := template.TargetTypes[rng.Intn(len(template.TargetTypes))]
		quest.Objectives[i] = g.generateObjective(rng, template, params, targetType)
	}

	// Generate description (use first objective for description)
	firstObjective := quest.Objectives[0]
	quest.Description = g.generateQuestDescription(rng, template, params, firstObjective.Target, firstObjective.Required)

	// Generate rewards
	depthScale := 1.0 + float64(params.Depth)*0.15
	g.generateRewards(rng, quest, template, depthScale)

	// Optional properties (use first objective's target for location)
	quest.RequiredLevel = 1 + params.Depth
	g.setOptionalProperties(rng, quest, template, quest.Objectives[0].Target)

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
		switch {
		case params.GenreID == "scifi" && descIdx == 2:
			return fmt.Sprintf(descTemplate, required, targetType)
		case params.GenreID == "cyberpunk" && descIdx == 1:
			return fmt.Sprintf(descTemplate, required, targetType)
		default:
			return fmt.Sprintf(descTemplate, targetType, required)
		}
	case TypeCollect:
		switch {
		case params.GenreID == "fantasy" && descIdx == 2:
			return fmt.Sprintf(descTemplate, targetType, required)
		case params.GenreID == "scifi" && descIdx == 1:
			return fmt.Sprintf(descTemplate, targetType, required)
		case params.GenreID == "horror" && descIdx == 1:
			return fmt.Sprintf(descTemplate, targetType, required)
		case params.GenreID == "cyberpunk" && descIdx == 1:
			return fmt.Sprintf(descTemplate, targetType, required)
		default:
			return fmt.Sprintf(descTemplate, required, targetType)
		}
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
		if err := validateSingleQuest(quest, i); err != nil {
			return err
		}
	}

	return nil
}

// validateSingleQuest validates a single quest's structure and data.
func validateSingleQuest(quest *Quest, index int) error {
	if quest == nil {
		return fmt.Errorf("quest %d is nil", index)
	}

	if err := validateQuestBasics(quest, index); err != nil {
		return err
	}

	if err := validateQuestObjectives(quest, index); err != nil {
		return err
	}

	return validateQuestRewards(quest, index)
}

// validateQuestBasics validates basic quest properties.
func validateQuestBasics(quest *Quest, index int) error {
	if quest.Name == "" {
		return fmt.Errorf("quest %d has empty name", index)
	}

	if quest.Description == "" {
		return fmt.Errorf("quest %d has empty description", index)
	}

	if quest.RequiredLevel < 0 {
		return fmt.Errorf("quest %d has negative required level", index)
	}

	return nil
}

// validateQuestObjectives validates quest objectives.
func validateQuestObjectives(quest *Quest, index int) error {
	if len(quest.Objectives) == 0 {
		return fmt.Errorf("quest %d has no objectives", index)
	}

	for j, obj := range quest.Objectives {
		if obj.Description == "" {
			return fmt.Errorf("quest %d objective %d has empty description", index, j)
		}
		if obj.Required <= 0 {
			return fmt.Errorf("quest %d objective %d has invalid required amount: %d", index, j, obj.Required)
		}
	}

	return nil
}

// validateQuestRewards validates quest rewards.
func validateQuestRewards(quest *Quest, index int) error {
	if quest.Reward.XP <= 0 {
		return fmt.Errorf("quest %d has no XP reward", index)
	}
	if quest.Reward.Gold < 0 {
		return fmt.Errorf("quest %d has negative gold reward", index)
	}
	return nil
}
