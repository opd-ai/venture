package quest

import (
	"fmt"
	"math/rand"

	"github.com/opd-ai/venture/pkg/procgen"
)

// QuestGenerator implements the Generator interface for procedural quest creation.
type QuestGenerator struct{}

// NewQuestGenerator creates a new quest generator.
func NewQuestGenerator() *QuestGenerator {
	return &QuestGenerator{}
}

// Generate creates quests based on the seed and parameters.
// Returns []*Quest or error.
func (g *QuestGenerator) Generate(seed int64, params procgen.GenerationParams) (interface{}, error) {
	// Validate parameters
	if params.Depth < 0 {
		return nil, fmt.Errorf("depth must be non-negative")
	}
	if params.Difficulty < 0 || params.Difficulty > 1 {
		return nil, fmt.Errorf("difficulty must be between 0 and 1")
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
		return nil, fmt.Errorf("no templates available for genre: %s", params.GenreID)
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

	// Copy tags
	copy(quest.Tags, template.Tags)

	// Generate ID
	quest.ID = fmt.Sprintf("quest_%d_%d", params.Depth, index)

	// Determine difficulty based on depth and difficulty parameter
	quest.Difficulty = g.determineDifficulty(rng, params.Depth, params.Difficulty)

	// Generate name
	prefix := template.NamePrefixes[rng.Intn(len(template.NamePrefixes))]
	suffix := template.NameSuffixes[rng.Intn(len(template.NameSuffixes))]
	quest.Name = fmt.Sprintf("%s %s", prefix, suffix)

	// Select target type
	targetType := template.TargetTypes[rng.Intn(len(template.TargetTypes))]

	// Generate objectives with scaling
	depthScale := 1.0 + float64(params.Depth)*0.15
	difficultyScale := 0.7 + params.Difficulty*0.6

	// Calculate required amount
	minRequired := int(float64(template.RequiredRange[0]) * difficultyScale)
	maxRequired := int(float64(template.RequiredRange[1]) * difficultyScale * depthScale)
	if minRequired < 1 {
		minRequired = 1
	}
	if maxRequired < minRequired {
		maxRequired = minRequired
	}

	required := minRequired
	if maxRequired > minRequired {
		required = minRequired + rng.Intn(maxRequired-minRequired+1)
	}

	// Create objective
	objective := Objective{
		Target:   targetType,
		Required: required,
		Current:  0,
	}

	// Generate objective description
	switch template.BaseType {
	case TypeKill:
		objective.Description = fmt.Sprintf("Defeat %d %s", required, targetType)
	case TypeCollect:
		objective.Description = fmt.Sprintf("Collect %d %s", required, targetType)
	case TypeBoss:
		objective.Description = fmt.Sprintf("Defeat %s", targetType)
	case TypeExplore:
		objective.Description = fmt.Sprintf("Discover %s", targetType)
	case TypeEscort:
		objective.Description = fmt.Sprintf("Escort %s safely", targetType)
	case TypeTalk:
		objective.Description = fmt.Sprintf("Speak with %s", targetType)
	}

	quest.Objectives = []Objective{objective}

	// Generate description from template
	descIdx := rng.Intn(len(template.DescTemplates))
	descTemplate := template.DescTemplates[descIdx]

	// Generate description based on quest type
	switch template.BaseType {
	case TypeKill:
		// Fantasy kill templates: "%s have been..." (target, count)
		// Sci-fi kill template 2: "Destroy %d %s..." (count, target)
		if params.GenreID == "scifi" && descIdx == 2 {
			quest.Description = fmt.Sprintf(descTemplate, required, targetType)
		} else {
			quest.Description = fmt.Sprintf(descTemplate, targetType, required)
		}
	case TypeCollect:
		// Collect templates vary by genre and index
		// Fantasy template 2: "Ancient %s are scattered... Collect %d of them." (target, count)
		// Sci-fi template 1: "Scanning systems detected %s nearby. Collect %d units." (target, count)
		// Others: "%d %s" (count, target)
		if (params.GenreID == "fantasy" && descIdx == 2) || (params.GenreID == "scifi" && descIdx == 1) {
			quest.Description = fmt.Sprintf(descTemplate, targetType, required)
		} else {
			quest.Description = fmt.Sprintf(descTemplate, required, targetType)
		}
	case TypeBoss, TypeExplore, TypeEscort, TypeTalk:
		// Single target name only
		quest.Description = fmt.Sprintf(descTemplate, targetType)
	default:
		quest.Description = fmt.Sprintf(descTemplate, targetType, required)
	}

	// Calculate rewards with scaling
	rarityMultiplier := 1.0 + float64(quest.Difficulty)*0.3

	minXP := int(float64(template.XPRewardRange[0]) * depthScale * rarityMultiplier)
	maxXP := int(float64(template.XPRewardRange[1]) * depthScale * rarityMultiplier)
	quest.Reward.XP = minXP
	if maxXP > minXP {
		quest.Reward.XP = minXP + rng.Intn(maxXP-minXP+1)
	}

	minGold := int(float64(template.GoldRewardRange[0]) * depthScale * rarityMultiplier)
	maxGold := int(float64(template.GoldRewardRange[1]) * depthScale * rarityMultiplier)
	quest.Reward.Gold = minGold
	if maxGold > minGold {
		quest.Reward.Gold = minGold + rng.Intn(maxGold-minGold+1)
	}

	// Item rewards
	if rng.Float64() < template.ItemRewardChance {
		numItems := 1 + rng.Intn(2) // 1-2 items
		quest.Reward.Items = make([]string, numItems)
		for i := 0; i < numItems; i++ {
			quest.Reward.Items[i] = fmt.Sprintf("item_%s_%d", quest.Difficulty.String(), i)
		}
	}

	// Skill point rewards
	if rng.Float64() < template.SkillPointChance {
		quest.Reward.SkillPoints = 1 + rng.Intn(2) // 1-2 skill points
	}

	// Set required level based on depth
	quest.RequiredLevel = 1 + params.Depth

	// Set location (optional)
	if template.BaseType == TypeExplore || template.BaseType == TypeBoss {
		quest.Location = targetType
	}

	// Set quest giver NPC (optional)
	if template.BaseType != TypeExplore {
		giverNames := []string{"Elder", "Captain", "Merchant", "Wizard", "Guard", "Scout", "Leader"}
		quest.GiverNPC = giverNames[rng.Intn(len(giverNames))]
	}

	return quest
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
