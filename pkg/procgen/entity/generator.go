// Package entity provides procedural entity generation.
// This file implements entity generators for monsters, NPCs, and bosses
// with stats, abilities, and AI behaviors.
package entity

import (
	"fmt"
	"math/rand"

	"github.com/opd-ai/venture/pkg/procgen"
	"github.com/sirupsen/logrus"
)

// EntityGenerator generates procedural entities (monsters, NPCs).
type EntityGenerator struct {
	templates map[string][]EntityTemplate
	logger    *logrus.Entry
}

// NewEntityGenerator creates a new entity generator.
func NewEntityGenerator() *EntityGenerator {
	return NewEntityGeneratorWithLogger(nil)
}

// NewEntityGeneratorWithLogger creates a new entity generator with a logger.
func NewEntityGeneratorWithLogger(logger *logrus.Logger) *EntityGenerator {
	var logEntry *logrus.Entry
	if logger != nil {
		logEntry = logger.WithField("generator", "entity")
	}

	gen := &EntityGenerator{
		templates: make(map[string][]EntityTemplate),
		logger:    logEntry,
	}

	// Register genre templates
	gen.templates["fantasy"] = GetFantasyTemplates()
	gen.templates["scifi"] = GetSciFiTemplates()
	gen.templates["horror"] = GetHorrorTemplates()       // GAP-005 REPAIR
	gen.templates["cyberpunk"] = GetCyberpunkTemplates() // GAP-005 REPAIR
	gen.templates["postapoc"] = GetPostApocTemplates()   // GAP-005 REPAIR
	gen.templates[""] = GetFantasyTemplates()            // default

	if logEntry != nil {
		logEntry.Debug("entity generator initialized")
	}

	return gen
}

// Generate creates entities based on the seed and parameters.
func (g *EntityGenerator) Generate(seed int64, params procgen.GenerationParams) (interface{}, error) {
	g.logDebug("starting entity generation", logrus.Fields{
		"seed":    seed,
		"genreID": params.GenreID,
		"depth":   params.Depth,
	})

	count := 10
	if params.Custom != nil {
		if c, ok := params.Custom["count"].(int); ok {
			count = c
		}
	}

	templates := g.templates[params.GenreID]
	if templates == nil {
		templates = g.templates[""]
	}

	rng := rand.New(rand.NewSource(seed))

	entities := make([]*Entity, count)
	for i := 0; i < count; i++ {
		entitySeed := seed + int64(i)*1000
		entities[i] = g.generateSingleEntity(entitySeed, params, templates, rng)
	}

	g.logInfo("entity generation complete", logrus.Fields{
		"count":   len(entities),
		"seed":    seed,
		"genreID": params.GenreID,
	})

	return entities, nil
}

// generateSingleEntity creates one entity.
func (g *EntityGenerator) generateSingleEntity(seed int64, params procgen.GenerationParams, templates []EntityTemplate, rng *rand.Rand) *Entity {
	// Select template based on depth and rng
	templateIndex := rng.Intn(len(templates))

	// Increase boss chance at higher depths
	if params.Depth > 10 && rng.Float64() < 0.1 {
		// Find boss template
		for i, t := range templates {
			if t.BaseType == TypeBoss {
				templateIndex = i
				break
			}
		}
	}

	template := templates[templateIndex]

	// Generate entity
	entity := &Entity{
		Type: template.BaseType,
		Size: template.BaseSize,
		Seed: seed,
		Tags: make([]string, len(template.Tags)),
	}
	copy(entity.Tags, template.Tags)

	// Generate name
	entity.Name = g.generateName(template, rng)

	// Determine rarity
	entity.Rarity = g.determineRarity(template.BaseType, params.Depth, rng)

	// Calculate level based on depth and difficulty
	entity.Stats.Level = g.calculateLevel(params.Depth, params.Difficulty, rng)

	// Generate stats
	entity.Stats = g.generateStats(template, entity.Stats.Level, entity.Rarity, rng)

	return entity
}

// generateName creates a procedural name for the entity.
func (g *EntityGenerator) generateName(template EntityTemplate, rng *rand.Rand) string {
	prefix := template.NamePrefixes[rng.Intn(len(template.NamePrefixes))]

	// Sometimes add suffix
	if len(template.NameSuffixes) > 0 && rng.Float64() < 0.7 {
		suffix := template.NameSuffixes[rng.Intn(len(template.NameSuffixes))]
		return fmt.Sprintf("%s %s", prefix, suffix)
	}

	return prefix
}

// determineRarity calculates the rarity based on type and depth.
func (g *EntityGenerator) determineRarity(entityType EntityType, depth int, rng *rand.Rand) Rarity {
	// Bosses are always at least rare
	if entityType == TypeBoss {
		rarityRoll := rng.Float64()
		if rarityRoll < 0.5 {
			return RarityRare
		} else if rarityRoll < 0.85 {
			return RarityEpic
		}
		return RarityLegendary
	}

	// Minions are mostly common
	if entityType == TypeMinion {
		if rng.Float64() < 0.9 {
			return RarityCommon
		}
		return RarityUncommon
	}

	// Regular monsters and NPCs - rarity increases with depth
	rarityRoll := rng.Float64()
	depthBonus := float64(depth) * 0.02 // 2% per depth level

	if rarityRoll < 0.6-depthBonus {
		return RarityCommon
	} else if rarityRoll < 0.85-depthBonus {
		return RarityUncommon
	} else if rarityRoll < 0.95-depthBonus {
		return RarityRare
	} else if rarityRoll < 0.99-depthBonus {
		return RarityEpic
	}
	return RarityLegendary
}

// calculateLevel determines the entity level.
func (g *EntityGenerator) calculateLevel(depth int, difficulty float64, rng *rand.Rand) int {
	// Base level from depth
	baseLevel := depth
	if baseLevel < 1 {
		baseLevel = 1
	}

	// Add some variance (+/- 20%)
	variance := int(float64(baseLevel) * 0.2)
	if variance < 1 {
		variance = 1
	}
	level := baseLevel + rng.Intn(variance*2+1) - variance

	// Apply difficulty modifier
	level = int(float64(level) * (0.5 + difficulty))

	if level < 1 {
		level = 1
	}

	return level
}

// generateStats creates entity stats based on template and modifiers.
func (g *EntityGenerator) generateStats(template EntityTemplate, level int, rarity Rarity, rng *rand.Rand) Stats {
	stats := Stats{
		Level: level,
	}

	// Generate base stats from template ranges
	healthRange := template.HealthRange[1] - template.HealthRange[0]
	stats.MaxHealth = template.HealthRange[0] + rng.Intn(healthRange+1)
	stats.Health = stats.MaxHealth

	damageRange := template.DamageRange[1] - template.DamageRange[0]
	stats.Damage = template.DamageRange[0] + rng.Intn(damageRange+1)

	defenseRange := template.DefenseRange[1] - template.DefenseRange[0]
	stats.Defense = template.DefenseRange[0] + rng.Intn(defenseRange+1)

	speedRange := template.SpeedRange[1] - template.SpeedRange[0]
	stats.Speed = template.SpeedRange[0] + rng.Float64()*speedRange

	// Apply level scaling
	levelMultiplier := 1.0 + float64(level-1)*0.15
	stats.MaxHealth = int(float64(stats.MaxHealth) * levelMultiplier)
	stats.Health = stats.MaxHealth
	stats.Damage = int(float64(stats.Damage) * levelMultiplier)
	stats.Defense = int(float64(stats.Defense) * levelMultiplier)

	// Apply rarity bonus
	rarityMultiplier := g.getRarityMultiplier(rarity)
	stats.MaxHealth = int(float64(stats.MaxHealth) * rarityMultiplier)
	stats.Health = stats.MaxHealth
	stats.Damage = int(float64(stats.Damage) * rarityMultiplier)
	stats.Defense = int(float64(stats.Defense) * rarityMultiplier)
	stats.Speed *= rarityMultiplier

	return stats
}

// getRarityMultiplier returns the stat multiplier for a rarity level.
func (g *EntityGenerator) getRarityMultiplier(rarity Rarity) float64 {
	switch rarity {
	case RarityCommon:
		return 1.0
	case RarityUncommon:
		return 1.2
	case RarityRare:
		return 1.5
	case RarityEpic:
		return 2.0
	case RarityLegendary:
		return 3.0
	default:
		return 1.0
	}
}

// Validate checks if the generated entities are valid.
func (g *EntityGenerator) Validate(result interface{}) error {
	entities, ok := result.([]*Entity)
	if !ok {
		return fmt.Errorf("result is not []*Entity")
	}

	if len(entities) == 0 {
		return fmt.Errorf("no entities generated")
	}

	// Validate each entity
	for i, entity := range entities {
		if entity.Name == "" {
			return fmt.Errorf("entity %d has empty name", i)
		}

		if entity.Stats.MaxHealth <= 0 {
			return fmt.Errorf("entity %d (%s) has invalid max health: %d", i, entity.Name, entity.Stats.MaxHealth)
		}

		if entity.Stats.Level <= 0 {
			return fmt.Errorf("entity %d (%s) has invalid level: %d", i, entity.Name, entity.Stats.Level)
		}

		if entity.Stats.Speed <= 0 {
			return fmt.Errorf("entity %d (%s) has invalid speed: %f", i, entity.Name, entity.Stats.Speed)
		}
	}

	return nil
}

// logDebug logs a debug message if logger and level are configured.
func (g *EntityGenerator) logDebug(msg string, fields logrus.Fields) {
	if g.logger != nil && g.logger.Logger.GetLevel() >= logrus.DebugLevel {
		g.logger.WithFields(fields).Debug(msg)
	}
}

// logInfo logs an info message if logger is configured.
func (g *EntityGenerator) logInfo(msg string, fields logrus.Fields) {
	if g.logger != nil {
		g.logger.WithFields(fields).Info(msg)
	}
}
