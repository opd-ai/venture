// Package magic provides procedural magic and spell generation.
// This file implements spell generators for offensive, defensive, utility,
// and summoning spells with procedural effects.
package magic

import (
	"fmt"
	"math/rand"

	"github.com/opd-ai/venture/pkg/procgen"
	"github.com/sirupsen/logrus"
)

// SpellGenerator implements the Generator interface for procedural spell creation.
type SpellGenerator struct {
	logger *logrus.Entry
}

// NewSpellGenerator creates a new spell generator.
func NewSpellGenerator() *SpellGenerator {
	return NewSpellGeneratorWithLogger(nil)
}

// NewSpellGeneratorWithLogger creates a new spell generator with a logger.
func NewSpellGeneratorWithLogger(logger *logrus.Logger) *SpellGenerator {
	var logEntry *logrus.Entry
	if logger != nil {
		logEntry = logger.WithField("generator", "spell")
		logEntry.Debug("spell generator initialized")
	}
	return &SpellGenerator{
		logger: logEntry,
	}
}

// Generate creates spells based on the seed and parameters.
// Returns []*Spell or error.
func (g *SpellGenerator) Generate(seed int64, params procgen.GenerationParams) (interface{}, error) {
	g.logDebug("starting spell generation", logrus.Fields{
		"seed":    seed,
		"genreID": params.GenreID,
		"depth":   params.Depth,
	})

	if err := g.validateParams(params); err != nil {
		return nil, err
	}

	count := g.getSpellCount(params)
	rng := rand.New(rand.NewSource(seed))

	templates, err := g.getTemplatesForGenre(params.GenreID)
	if err != nil {
		return nil, err
	}

	spells := g.generateSpells(rng, templates, params, seed, count)

	g.logInfo("spell generation complete", logrus.Fields{
		"count":   len(spells),
		"seed":    seed,
		"genreID": params.GenreID,
	})

	return spells, nil
}

// validateParams validates generation parameters.
func (g *SpellGenerator) validateParams(params procgen.GenerationParams) error {
	if params.Depth < 0 {
		g.logWarn("invalid depth parameter", logrus.Fields{"depth": params.Depth})
		return fmt.Errorf("depth must be non-negative")
	}
	if params.Difficulty < 0 || params.Difficulty > 1 {
		g.logWarn("invalid difficulty parameter", logrus.Fields{"difficulty": params.Difficulty})
		return fmt.Errorf("difficulty must be between 0 and 1")
	}
	return nil
}

// getSpellCount extracts the spell count from custom parameters.
func (g *SpellGenerator) getSpellCount(params procgen.GenerationParams) int {
	count := 10 // default
	if c, ok := params.Custom["count"].(int); ok {
		count = c
	}
	return count
}

// getTemplatesForGenre returns spell templates for the specified genre.
func (g *SpellGenerator) getTemplatesForGenre(genreID string) ([]SpellTemplate, error) {
	var templates []SpellTemplate
	switch genreID {
	case "scifi":
		templates = append(templates, GetSciFiOffensiveTemplates()...)
		templates = append(templates, GetSciFiSupportTemplates()...)
	case "fantasy":
		fallthrough
	default:
		templates = append(templates, GetFantasyOffensiveTemplates()...)
		templates = append(templates, GetFantasySupportTemplates()...)
	}

	if len(templates) == 0 {
		g.logError("no templates available", logrus.Fields{"genreID": genreID})
		return nil, fmt.Errorf("no templates available for genre: %s", genreID)
	}

	return templates, nil
}

// generateSpells generates the specified count of spells from templates.
func (g *SpellGenerator) generateSpells(rng *rand.Rand, templates []SpellTemplate, params procgen.GenerationParams, seed int64, count int) []*Spell {
	spells := make([]*Spell, 0, count)
	for i := 0; i < count; i++ {
		template := templates[rng.Intn(len(templates))]
		spell := g.generateFromTemplate(rng, template, params)
		spell.Seed = seed + int64(i)
		spells = append(spells, spell)
	}
	return spells
}

// generateFromTemplate creates a single spell from a template.
func (g *SpellGenerator) generateFromTemplate(rng *rand.Rand, template SpellTemplate, params procgen.GenerationParams) *Spell {
	spell := &Spell{
		Type:    template.BaseType,
		Element: template.BaseElement,
		Target:  template.BaseTarget,
		Tags:    make([]string, len(template.Tags)),
	}

	// Copy tags
	copy(spell.Tags, template.Tags)

	// Determine rarity based on depth and difficulty
	spell.Rarity = g.determineRarity(rng, params.Depth, params.Difficulty)

	// Generate name
	prefix := template.NamePrefixes[rng.Intn(len(template.NamePrefixes))]
	suffix := template.NameSuffixes[rng.Intn(len(template.NameSuffixes))]
	spell.Name = fmt.Sprintf("%s %s", prefix, suffix)

	// Add rarity prefix for higher rarities
	if spell.Rarity >= RarityRare {
		rarityPrefixes := []string{"Greater", "Superior", "Ultimate", "Ancient", "Legendary"}
		spell.Name = fmt.Sprintf("%s %s", rarityPrefixes[spell.Rarity-RarityRare], spell.Name)
	}

	// Generate stats with scaling
	depthScale := 1.0 + float64(params.Depth)*0.1
	difficultyScale := 0.8 + params.Difficulty*0.4
	rarityScale := 1.0 + float64(spell.Rarity)*0.25

	spell.Stats = g.generateStats(rng, template, depthScale, difficultyScale, rarityScale)
	spell.Stats.RequiredLevel = 1 + params.Depth + int(spell.Rarity)*2

	// Generate description
	spell.Description = g.generateDescription(spell)

	return spell
}

// generateStats generates spell statistics from template ranges.
func (g *SpellGenerator) generateStats(rng *rand.Rand, template SpellTemplate, depthScale, difficultyScale, rarityScale float64) Stats {
	stats := Stats{}

	// Damage
	if template.DamageRange[1] > 0 {
		baseMin := float64(template.DamageRange[0])
		baseMax := float64(template.DamageRange[1])
		damage := baseMin + rng.Float64()*(baseMax-baseMin)
		stats.Damage = int(damage * depthScale * difficultyScale * rarityScale)
	}

	// Healing
	if template.HealingRange[1] > 0 {
		baseMin := float64(template.HealingRange[0])
		baseMax := float64(template.HealingRange[1])
		healing := baseMin + rng.Float64()*(baseMax-baseMin)
		stats.Healing = int(healing * depthScale * rarityScale)
	}

	// Mana cost
	if template.ManaCostRange[1] > 0 {
		baseMin := float64(template.ManaCostRange[0])
		baseMax := float64(template.ManaCostRange[1])
		manaCost := baseMin + rng.Float64()*(baseMax-baseMin)
		// Higher rarity costs more mana
		stats.ManaCost = int(manaCost * rarityScale)
	}

	// Cooldown
	if template.CooldownRange[1] > 0 {
		stats.Cooldown = template.CooldownRange[0] +
			rng.Float64()*(template.CooldownRange[1]-template.CooldownRange[0])
		// Higher rarity has shorter cooldown
		stats.Cooldown = stats.Cooldown / rarityScale
	}

	// Cast time
	if template.CastTimeRange[1] > 0 {
		stats.CastTime = template.CastTimeRange[0] +
			rng.Float64()*(template.CastTimeRange[1]-template.CastTimeRange[0])
		// Higher rarity has faster cast time
		stats.CastTime = stats.CastTime / (1.0 + float64(rarityScale)*0.1)
	}

	// Range
	if template.RangeRange[1] > 0 {
		stats.Range = template.RangeRange[0] +
			rng.Float64()*(template.RangeRange[1]-template.RangeRange[0])
		// Higher rarity has better range
		stats.Range = stats.Range * (1.0 + float64(rarityScale)*0.1)
	}

	// Area size
	if template.AreaSizeRange[1] > 0 {
		stats.AreaSize = template.AreaSizeRange[0] +
			rng.Float64()*(template.AreaSizeRange[1]-template.AreaSizeRange[0])
		// Higher rarity has larger area
		stats.AreaSize = stats.AreaSize * (1.0 + float64(rarityScale)*0.15)
	}

	// Duration
	if template.DurationRange[1] > 0 {
		stats.Duration = template.DurationRange[0] +
			rng.Float64()*(template.DurationRange[1]-template.DurationRange[0])
		// Higher rarity has longer duration
		stats.Duration = stats.Duration * (1.0 + float64(rarityScale)*0.2)
	}

	return stats
}

// determineRarity calculates spell rarity based on depth and difficulty.
func (g *SpellGenerator) determineRarity(rng *rand.Rand, depth int, difficulty float64) Rarity {
	// Base chance influenced by depth
	roll := rng.Float64()

	// Depth increases chance of higher rarity
	depthBonus := float64(depth) * 0.02
	difficultyBonus := difficulty * 0.1

	roll += depthBonus + difficultyBonus

	// Determine rarity thresholds
	switch {
	case roll < 0.50:
		return RarityCommon
	case roll < 0.75:
		return RarityUncommon
	case roll < 0.90:
		return RarityRare
	case roll < 0.97:
		return RarityEpic
	default:
		return RarityLegendary
	}
}

// generateDescription creates flavor text for the spell.
func (g *SpellGenerator) generateDescription(spell *Spell) string {
	// Build description based on spell type and element
	var action string
	switch spell.Type {
	case TypeOffensive:
		action = "Unleashes"
	case TypeDefensive:
		action = "Creates"
	case TypeHealing:
		action = "Channels"
	case TypeBuff:
		action = "Grants"
	case TypeDebuff:
		action = "Inflicts"
	case TypeUtility:
		action = "Manifests"
	case TypeSummon:
		action = "Summons"
	}

	var elementDesc string
	switch spell.Element {
	case ElementFire:
		elementDesc = "searing flames"
	case ElementIce:
		elementDesc = "freezing cold"
	case ElementLightning:
		elementDesc = "crackling lightning"
	case ElementEarth:
		elementDesc = "crushing stone"
	case ElementWind:
		elementDesc = "howling winds"
	case ElementLight:
		elementDesc = "radiant light"
	case ElementDark:
		elementDesc = "shadowy darkness"
	case ElementArcane:
		elementDesc = "pure magical energy"
	case ElementNone:
		elementDesc = "raw power"
	}

	var targetDesc string
	switch spell.Target {
	case TargetSelf:
		targetDesc = "upon the caster"
	case TargetSingle:
		targetDesc = "at a target"
	case TargetArea:
		targetDesc = "in an area"
	case TargetCone:
		targetDesc = "in a cone"
	case TargetLine:
		targetDesc = "in a line"
	case TargetAllAllies:
		targetDesc = "upon all allies"
	case TargetAllEnemies:
		targetDesc = "upon all enemies"
	}

	return fmt.Sprintf("%s %s %s.", action, elementDesc, targetDesc)
}

// Validate checks if the generated spells are valid.
func (g *SpellGenerator) Validate(result interface{}) error {
	spells, ok := result.([]*Spell)
	if !ok {
		return fmt.Errorf("result is not []*Spell")
	}

	if len(spells) == 0 {
		return fmt.Errorf("no spells generated")
	}

	for i, spell := range spells {
		if spell == nil {
			return fmt.Errorf("spell %d is nil", i)
		}

		// Validate name
		if spell.Name == "" {
			return fmt.Errorf("spell %d has empty name", i)
		}

		// Validate type is valid
		if spell.Type < TypeOffensive || spell.Type > TypeSummon {
			return fmt.Errorf("spell %d has invalid type: %d", i, spell.Type)
		}

		// Validate element is valid
		if spell.Element < ElementNone || spell.Element > ElementArcane {
			return fmt.Errorf("spell %d has invalid element: %d", i, spell.Element)
		}

		// Validate rarity is valid
		if spell.Rarity < RarityCommon || spell.Rarity > RarityLegendary {
			return fmt.Errorf("spell %d has invalid rarity: %d", i, spell.Rarity)
		}

		// Validate target is valid
		if spell.Target < TargetSelf || spell.Target > TargetAllEnemies {
			return fmt.Errorf("spell %d has invalid target: %d", i, spell.Target)
		}

		// Validate stats make sense
		if spell.Stats.ManaCost < 0 {
			return fmt.Errorf("spell %d has negative mana cost", i)
		}
		if spell.Stats.Cooldown < 0 {
			return fmt.Errorf("spell %d has negative cooldown", i)
		}
		if spell.Stats.CastTime < 0 {
			return fmt.Errorf("spell %d has negative cast time", i)
		}
		if spell.Stats.Range < 0 {
			return fmt.Errorf("spell %d has negative range", i)
		}
		if spell.Stats.RequiredLevel < 1 {
			return fmt.Errorf("spell %d has invalid required level: %d", i, spell.Stats.RequiredLevel)
		}

		// Type-specific validation
		if spell.IsOffensive() && spell.Stats.Damage <= 0 {
			return fmt.Errorf("offensive spell %d has no damage", i)
		}
		if spell.Type == TypeHealing && spell.Stats.Healing <= 0 {
			return fmt.Errorf("healing spell %d has no healing", i)
		}
	}

	return nil
}

// logDebug logs a debug message if logger and level are configured.
func (g *SpellGenerator) logDebug(msg string, fields logrus.Fields) {
if g.logger != nil && g.logger.Logger.GetLevel() >= logrus.DebugLevel {
g.logger.WithFields(fields).Debug(msg)
}
}

// logInfo logs an info message if logger is configured.
func (g *SpellGenerator) logInfo(msg string, fields logrus.Fields) {
if g.logger != nil {
g.logger.WithFields(fields).Info(msg)
}
}

// logWarn logs a warning message if logger is configured.
func (g *SpellGenerator) logWarn(msg string, fields logrus.Fields) {
if g.logger != nil {
g.logger.WithFields(fields).Warn(msg)
}
}

// logError logs an error message if logger is configured.
func (g *SpellGenerator) logError(msg string, fields logrus.Fields) {
if g.logger != nil {
g.logger.WithFields(fields).Error(msg)
}
}
