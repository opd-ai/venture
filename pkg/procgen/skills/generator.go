// Package skills provides procedural skill tree generation.
// This file implements skill tree generators with prerequisites,
// progression paths, and balanced skill nodes.
package skills

import (
	"fmt"
	"math/rand"
	"strings"

	"github.com/opd-ai/venture/pkg/procgen"
	"github.com/sirupsen/logrus"
)

// SkillTreeGenerator implements the Generator interface for procedural skill tree creation.
type SkillTreeGenerator struct {
	logger *logrus.Entry
}

// NewSkillTreeGenerator creates a new skill tree generator.
func NewSkillTreeGenerator() *SkillTreeGenerator {
	return NewSkillTreeGeneratorWithLogger(nil)
}

// NewSkillTreeGeneratorWithLogger creates a new skill tree generator with a logger.
func NewSkillTreeGeneratorWithLogger(logger *logrus.Logger) *SkillTreeGenerator {
	var logEntry *logrus.Entry
	if logger != nil {
		logEntry = logger.WithFields(logrus.Fields{
			"generator": "skill_tree",
		})
	}
	return &SkillTreeGenerator{
		logger: logEntry,
	}
}

// Generate creates skill trees based on the seed and parameters.
// Returns []*SkillTree or error.
func (g *SkillTreeGenerator) Generate(seed int64, params procgen.GenerationParams) (interface{}, error) {
	if g.logger != nil && g.logger.Logger.GetLevel() >= logrus.DebugLevel {
		g.logger.WithFields(logrus.Fields{
			"seed":       seed,
			"genreID":    params.GenreID,
			"depth":      params.Depth,
			"difficulty": params.Difficulty,
		}).Debug("starting skill tree generation")
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
	count := 3 // default: generate 3 trees
	if c, ok := params.Custom["count"].(int); ok {
		count = c
	}

	// Create deterministic RNG
	rng := rand.New(rand.NewSource(seed))

	// Get templates based on genre
	var templates []SkillTreeTemplate
	switch params.GenreID {
	case "scifi":
		templates = GetSciFiTreeTemplates()
	case "fantasy":
		fallthrough
	default:
		templates = GetFantasyTreeTemplates()
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
		}).Debug("generating skill trees")
	}

	// Generate skill trees
	trees := make([]*SkillTree, 0, count)
	for i := 0; i < count; i++ {
		// Select template
		template := templates[i%len(templates)]

		// Generate tree from template
		tree := g.generateTree(rng, template, params, seed+int64(i))
		trees = append(trees, tree)

		if g.logger != nil && g.logger.Logger.GetLevel() >= logrus.DebugLevel {
			g.logger.WithFields(logrus.Fields{
				"treeIndex":  i,
				"treeName":   tree.Name,
				"nodeCount":  len(tree.Nodes),
				"maxPoints":  tree.MaxPoints,
			}).Debug("skill tree generated")
		}
	}

	if g.logger != nil {
		g.logger.WithFields(logrus.Fields{
			"treeCount": len(trees),
			"seed":      seed,
		}).Info("skill tree generation complete")
	}

	return trees, nil
}

// generateTree creates a complete skill tree from a template.
func (g *SkillTreeGenerator) generateTree(rng *rand.Rand, template SkillTreeTemplate, params procgen.GenerationParams, treeSeed int64) *SkillTree {
	tree := &SkillTree{
		ID:          fmt.Sprintf("%s_%d", template.Name, treeSeed),
		Name:        template.Name,
		Description: template.Description,
		Category:    template.Category,
		Genre:       params.GenreID,
		MaxPoints:   50 + params.Depth*5, // Scale points with depth
		Seed:        treeSeed,
		Nodes:       make([]*SkillNode, 0),
		RootNodes:   make([]*SkillNode, 0),
	}

	// Generate skills for each tier
	skillsByTier := make(map[int][]*SkillNode)
	skillID := 0

	for tier := 0; tier <= 6; tier++ {
		// Determine number of skills in this tier
		tierSkillCount := g.getTierSkillCount(tier, params.Depth)

		for i := 0; i < tierSkillCount; i++ {
			// Select appropriate template for this tier
			skillTemplate := g.selectSkillTemplate(rng, template.SkillTemplates, tier)
			if skillTemplate == nil {
				continue
			}

			// Generate skill
			skill := g.generateSkill(rng, *skillTemplate, tier, skillID, treeSeed, params)
			skillID++

			// Create node
			node := &SkillNode{
				Skill:    skill,
				Children: make([]*SkillNode, 0),
				Position: Position{
					X: i,
					Y: tier,
				},
			}

			skillsByTier[tier] = append(skillsByTier[tier], node)
			tree.Nodes = append(tree.Nodes, node)

			// Root nodes (tier 0)
			if tier == 0 {
				tree.RootNodes = append(tree.RootNodes, node)
			}
		}
	}

	// Connect nodes (establish prerequisites)
	g.connectNodes(rng, skillsByTier, tree)

	return tree
}

// generateSkill creates a single skill from a template.
func (g *SkillTreeGenerator) generateSkill(rng *rand.Rand, template SkillTemplate, tier, id int, treeSeed int64, params procgen.GenerationParams) *Skill {
	skill := &Skill{
		ID:       fmt.Sprintf("skill_%d_%d", treeSeed, id),
		Type:     template.BaseType,
		Category: template.BaseCategory,
		Tier:     g.getTierEnum(tier),
		Tags:     make([]string, len(template.Tags)),
		Seed:     treeSeed + int64(id)*100,
	}

	// Copy tags
	copy(skill.Tags, template.Tags)

	// Generate name
	prefix := template.NamePrefixes[rng.Intn(len(template.NamePrefixes))]
	suffix := template.NameSuffixes[rng.Intn(len(template.NameSuffixes))]
	skill.Name = fmt.Sprintf("%s %s", prefix, suffix)

	// Determine max level based on tier and type
	minLevel := template.MaxLevelRange[0]
	maxLevel := template.MaxLevelRange[1]
	skill.MaxLevel = minLevel + rng.Intn(maxLevel-minLevel+1)

	// Ultimate skills have fixed max level
	if skill.Type == TypeUltimate {
		skill.MaxLevel = 1
	}

	// Set requirements based on tier
	skill.Requirements = Requirements{
		PlayerLevel:       1 + tier*5 + params.Depth,
		SkillPoints:       1 + tier/2,
		PrerequisiteIDs:   []string{}, // Set later when connecting nodes
		AttributeMinimums: make(map[string]int),
	}

	// Generate effects
	skill.Effects = g.generateEffects(rng, template, tier, params)

	// Generate description
	skill.Description = g.generateDescription(skill, template.DescriptionFormat)

	return skill
}

// generateEffects creates skill effects from template.
func (g *SkillTreeGenerator) generateEffects(rng *rand.Rand, template SkillTemplate, tier int, params procgen.GenerationParams) []Effect {
	effects := make([]Effect, 0)

	// Tier scaling
	tierScale := 1.0 + float64(tier)*0.3
	depthScale := 1.0 + float64(params.Depth)*0.05

	// Generate 1-3 effects per skill
	numEffects := 1 + rng.Intn(3)
	usedTypes := make(map[string]bool)

	for i := 0; i < numEffects && i < len(template.EffectTypes); i++ {
		// Select random effect type
		effectType := template.EffectTypes[rng.Intn(len(template.EffectTypes))]

		// Avoid duplicates
		if usedTypes[effectType] {
			continue
		}
		usedTypes[effectType] = true

		// Get value range for this effect type
		valueRange, ok := template.ValueRanges[effectType]
		if !ok {
			continue
		}

		// Generate value
		min := valueRange[0]
		max := valueRange[1]
		value := min + rng.Float64()*(max-min)
		value = value * tierScale * depthScale

		// Determine if percentage
		isPercent := strings.HasSuffix(effectType, "_percent") ||
			strings.Contains(effectType, "bonus") ||
			strings.Contains(effectType, "reduction")

		effect := Effect{
			Type:        effectType,
			Value:       value,
			IsPercent:   isPercent,
			Description: g.formatEffectDescription(effectType, value, isPercent),
		}

		effects = append(effects, effect)
	}

	return effects
}

// formatEffectDescription creates a human-readable effect description.
func (g *SkillTreeGenerator) formatEffectDescription(effectType string, value float64, isPercent bool) string {
	format := "%.1f"
	if isPercent {
		format = "%.0f%%"
		value = value * 100 // Convert to percentage
	}

	// Clean up effect type for display
	displayType := strings.ReplaceAll(effectType, "_", " ")
	displayType = strings.Title(displayType)

	if value >= 0 {
		return fmt.Sprintf("+"+format+" %s", value, displayType)
	}
	return fmt.Sprintf(format+" %s", value, displayType)
}

// generateDescription creates a skill description.
func (g *SkillTreeGenerator) generateDescription(skill *Skill, format string) string {
	if format != "" {
		// Use template format if provided
		return fmt.Sprintf(format, skill.Name)
	}

	// Generate description from effects
	parts := make([]string, 0, len(skill.Effects))
	for _, effect := range skill.Effects {
		parts = append(parts, effect.Description)
	}

	typeDesc := ""
	switch skill.Type {
	case TypePassive:
		typeDesc = "Passive: "
	case TypeActive:
		typeDesc = "Active: "
	case TypeUltimate:
		typeDesc = "Ultimate: "
	case TypeSynergy:
		typeDesc = "Synergy: "
	}

	return typeDesc + strings.Join(parts, ", ")
}

// getTierSkillCount determines how many skills should be in a tier.
func (g *SkillTreeGenerator) getTierSkillCount(tier, depth int) int {
	// Pyramid structure: more skills in lower tiers
	switch tier {
	case 0:
		return 3 // Base skills
	case 1, 2:
		return 4 + depth/5 // Intermediate
	case 3, 4:
		return 3 + depth/5 // Advanced
	case 5:
		return 2 // Master
	case 6:
		return 1 // Ultimate
	default:
		return 0
	}
}

// getTierEnum converts numeric tier to Tier enum.
func (g *SkillTreeGenerator) getTierEnum(tier int) Tier {
	switch {
	case tier == 0:
		return TierBasic
	case tier <= 2:
		return TierIntermediate
	case tier <= 4:
		return TierAdvanced
	default:
		return TierMaster
	}
}

// selectSkillTemplate selects an appropriate template for a tier.
func (g *SkillTreeGenerator) selectSkillTemplate(rng *rand.Rand, templates []SkillTemplate, tier int) *SkillTemplate {
	// Filter templates appropriate for this tier
	suitable := make([]SkillTemplate, 0)
	for _, t := range templates {
		if tier >= t.TierRange[0] && tier <= t.TierRange[1] {
			suitable = append(suitable, t)
		}
	}

	if len(suitable) == 0 {
		return nil
	}

	template := suitable[rng.Intn(len(suitable))]
	return &template
}

// connectNodes establishes prerequisite relationships between skills.
func (g *SkillTreeGenerator) connectNodes(rng *rand.Rand, skillsByTier map[int][]*SkillNode, tree *SkillTree) {
	// Connect each tier to previous tier
	for tier := 1; tier <= 6; tier++ {
		currentTier := skillsByTier[tier]
		previousTier := skillsByTier[tier-1]

		if len(previousTier) == 0 {
			continue
		}

		for _, node := range currentTier {
			// Each skill requires 1-2 skills from previous tier
			numPrereqs := 1
			if tier >= 3 && rng.Float64() < 0.3 {
				numPrereqs = 2
			}

			// Select random prerequisites
			prereqs := make(map[int]bool)
			attempts := 0
			for len(prereqs) < numPrereqs && attempts < 10 {
				prereqIdx := rng.Intn(len(previousTier))
				prereqs[prereqIdx] = true
				attempts++
			}

			// Establish connections
			for prereqIdx := range prereqs {
				prereqNode := previousTier[prereqIdx]
				prereqNode.Children = append(prereqNode.Children, node)
				node.Skill.Requirements.PrerequisiteIDs = append(
					node.Skill.Requirements.PrerequisiteIDs,
					prereqNode.Skill.ID,
				)
			}
		}
	}
}

// Validate checks if generated skill trees are valid.
func (g *SkillTreeGenerator) Validate(result interface{}) error {
	trees, ok := result.([]*SkillTree)
	if !ok {
		return fmt.Errorf("expected []*SkillTree, got %T", result)
	}

	if len(trees) == 0 {
		return fmt.Errorf("no skill trees generated")
	}

	for i, tree := range trees {
		if tree == nil {
			return fmt.Errorf("tree %d is nil", i)
		}

		if tree.Name == "" {
			return fmt.Errorf("tree %d has empty name", i)
		}

		if len(tree.Nodes) == 0 {
			return fmt.Errorf("tree %d has no skills", i)
		}

		if len(tree.RootNodes) == 0 {
			return fmt.Errorf("tree %d has no root nodes", i)
		}

		// Validate each skill
		for j, node := range tree.Nodes {
			if node == nil || node.Skill == nil {
				return fmt.Errorf("tree %d node %d is nil", i, j)
			}

			skill := node.Skill
			if skill.Name == "" {
				return fmt.Errorf("tree %d skill %d has empty name", i, j)
			}

			if skill.MaxLevel < 1 {
				return fmt.Errorf("tree %d skill %d has invalid max level: %d", i, j, skill.MaxLevel)
			}

			if len(skill.Effects) == 0 {
				return fmt.Errorf("tree %d skill %d has no effects", i, j)
			}
		}

		// Validate prerequisites exist
		for _, node := range tree.Nodes {
			for _, prereqID := range node.Skill.Requirements.PrerequisiteIDs {
				if tree.GetSkillByID(prereqID) == nil {
					return fmt.Errorf("prerequisite %s not found in tree %s", prereqID, tree.ID)
				}
			}
		}
	}

	return nil
}
