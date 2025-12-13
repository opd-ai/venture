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
	logger        *logrus.Entry
	balanceConfig BalanceConfig
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
		logger:        logEntry,
		balanceConfig: DefaultBalanceConfig(),
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
	if err := procgen.ValidateDepth(params.Depth); err != nil {
		g.logWarn("invalid depth parameter", logrus.Fields{"depth": params.Depth})
		return err
	}
	if err := procgen.ValidateDifficulty(params.Difficulty); err != nil {
		g.logWarn("invalid difficulty parameter", logrus.Fields{"difficulty": params.Difficulty})
		return err
	}
	return nil
}

// getSpellCount extracts the spell count from custom parameters.
func (g *SpellGenerator) getSpellCount(params procgen.GenerationParams) int {
	g.logDebug("determining spell count", logrus.Fields{
		"custom_params": params.Custom,
	})
	count := 10 // default
	if c, ok := params.Custom["count"].(int); ok {
		count = c
		g.logDebug("using custom spell count", logrus.Fields{"count": count})
	} else {
		g.logDebug("using default spell count", logrus.Fields{"count": count})
	}
	return count
}

// getTemplatesForGenre returns spell templates for the specified genre.
func (g *SpellGenerator) getTemplatesForGenre(genreID string) ([]SpellTemplate, error) {
	g.logDebug("retrieving templates for genre", logrus.Fields{"genre_id": genreID})
	var templates []SpellTemplate
	switch genreID {
	case "scifi":
		g.logDebug("loading scifi templates", logrus.Fields{"genre_id": genreID})
		templates = append(templates, GetSciFiOffensiveTemplates()...)
		templates = append(templates, GetSciFiSupportTemplates()...)
		templates = append(templates, GetSciFiAdvancedTemplates()...)
		templates = append(templates, GetAdvancedOffensiveTemplates()...)
		templates = append(templates, GetAdvancedUtilityTemplates()...)
		templates = append(templates, GetAdvancedSupportTemplates()...)
	case "horror":
		g.logDebug("loading horror templates", logrus.Fields{"genre_id": genreID})
		templates = append(templates, GetFantasyOffensiveTemplates()...)
		templates = append(templates, GetFantasySupportTemplates()...)
		templates = append(templates, GetHorrorAdvancedTemplates()...)
		templates = append(templates, GetAdvancedOffensiveTemplates()...)
		templates = append(templates, GetAdvancedUtilityTemplates()...)
		templates = append(templates, GetAdvancedSupportTemplates()...)
	case "fantasy":
		fallthrough
	default:
		g.logDebug("loading fantasy templates (default)", logrus.Fields{"genre_id": genreID})
		templates = append(templates, GetFantasyOffensiveTemplates()...)
		templates = append(templates, GetFantasySupportTemplates()...)
		templates = append(templates, GetAdvancedOffensiveTemplates()...)
		templates = append(templates, GetAdvancedUtilityTemplates()...)
		templates = append(templates, GetAdvancedSupportTemplates()...)
	}

	if len(templates) == 0 {
		g.logError("no templates available", logrus.Fields{"genre_id": genreID})
		return nil, fmt.Errorf("no templates available for genre: %s", genreID)
	}

	g.logDebug("templates loaded successfully", logrus.Fields{
		"genre_id":       genreID,
		"template_count": len(templates),
	})
	return templates, nil
}

// generateSpells generates the specified count of spells from templates.
func (g *SpellGenerator) generateSpells(rng *rand.Rand, templates []SpellTemplate, params procgen.GenerationParams, seed int64, count int) []*Spell {
	g.logDebug("generating spells from templates", logrus.Fields{
		"count":          count,
		"template_count": len(templates),
		"seed":           seed,
	})
	spells := make([]*Spell, count)
	for i := 0; i < count; i++ {
		template := templates[rng.Intn(len(templates))]
		spell := g.generateFromTemplate(rng, template, params)
		spell.Seed = seed + int64(i)
		spells[i] = spell
		g.logDebug("generated spell from template", logrus.Fields{
			"spell_index": i,
			"spell_name":  spell.Name,
			"spell_type":  spell.Type,
			"rarity":      spell.Rarity,
			"seed":        spell.Seed,
		})
	}
	g.logDebug("spell generation loop complete", logrus.Fields{"spell_count": len(spells)})
	return spells
}

// generateFromTemplate creates a single spell from a template.
func (g *SpellGenerator) generateFromTemplate(rng *rand.Rand, template SpellTemplate, params procgen.GenerationParams) *Spell {
	g.logDebug("generating spell from template", logrus.Fields{
		"template_type":    template.BaseType,
		"template_element": template.BaseElement,
		"depth":            params.Depth,
		"difficulty":       params.Difficulty,
	})
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
	g.logDebug("determined spell rarity", logrus.Fields{"rarity": spell.Rarity})

	// Generate name
	prefix := template.NamePrefixes[rng.Intn(len(template.NamePrefixes))]
	suffix := template.NameSuffixes[rng.Intn(len(template.NameSuffixes))]
	spell.Name = fmt.Sprintf("%s %s", prefix, suffix)

	// Add rarity prefix for higher rarities
	if spell.Rarity >= RarityRare {
		rarityPrefixes := []string{"Greater", "Superior", "Ultimate", "Ancient", "Legendary"}
		spell.Name = fmt.Sprintf("%s %s", rarityPrefixes[spell.Rarity-RarityRare], spell.Name)
	}
	g.logDebug("generated spell name", logrus.Fields{"spell_name": spell.Name})

	// Generate stats with scaling
	depthScale := 1.0 + float64(params.Depth)*0.1
	difficultyScale := 0.8 + params.Difficulty*0.4
	rarityScale := 1.0 + float64(spell.Rarity)*0.25

	g.logDebug("calculated scaling factors", logrus.Fields{
		"depth_scale":      depthScale,
		"difficulty_scale": difficultyScale,
		"rarity_scale":     rarityScale,
	})

	spell.Stats = g.generateStats(rng, template, depthScale, difficultyScale, rarityScale)
	spell.Stats.RequiredLevel = 1 + params.Depth + int(spell.Rarity)*2

	// Apply balance formulas to ensure consistent power levels
	g.balanceConfig.BalanceStats(&spell.Stats, spell.Type, spell.Target, spell.Stats.RequiredLevel)

	// Generate description
	spell.Description = g.generateDescription(spell)

	g.logDebug("spell generation complete", logrus.Fields{
		"spell_name":     spell.Name,
		"required_level": spell.Stats.RequiredLevel,
		"mana_cost":      spell.Stats.ManaCost,
		"damage":         spell.Stats.Damage,
	})
	return spell
}

// generateStats generates spell statistics from template ranges.
func (g *SpellGenerator) generateStats(rng *rand.Rand, template SpellTemplate, depthScale, difficultyScale, rarityScale float64) Stats {
	g.logDebug("generating spell stats", logrus.Fields{
		"depth_scale":      depthScale,
		"difficulty_scale": difficultyScale,
		"rarity_scale":     rarityScale,
	})
	stats := Stats{}

	// Damage
	if template.DamageRange[1] > 0 {
		baseMin := float64(template.DamageRange[0])
		baseMax := float64(template.DamageRange[1])
		damage := baseMin + rng.Float64()*(baseMax-baseMin)
		stats.Damage = int(damage * depthScale * difficultyScale * rarityScale)
		g.logDebug("generated damage stat", logrus.Fields{
			"base_range": template.DamageRange,
			"damage":     stats.Damage,
		})
	}

	// Healing
	if template.HealingRange[1] > 0 {
		baseMin := float64(template.HealingRange[0])
		baseMax := float64(template.HealingRange[1])
		healing := baseMin + rng.Float64()*(baseMax-baseMin)
		stats.Healing = int(healing * depthScale * rarityScale)
		g.logDebug("generated healing stat", logrus.Fields{
			"base_range": template.HealingRange,
			"healing":    stats.Healing,
		})
	}

	// Mana cost
	if template.ManaCostRange[1] > 0 {
		baseMin := float64(template.ManaCostRange[0])
		baseMax := float64(template.ManaCostRange[1])
		manaCost := baseMin + rng.Float64()*(baseMax-baseMin)
		// Higher rarity costs more mana
		stats.ManaCost = int(manaCost * rarityScale)
		g.logDebug("generated mana cost", logrus.Fields{
			"base_range": template.ManaCostRange,
			"mana_cost":  stats.ManaCost,
		})
	}

	// Cooldown
	if template.CooldownRange[1] > 0 {
		stats.Cooldown = template.CooldownRange[0] +
			rng.Float64()*(template.CooldownRange[1]-template.CooldownRange[0])
		// Higher rarity has shorter cooldown
		stats.Cooldown = stats.Cooldown / rarityScale
		g.logDebug("generated cooldown", logrus.Fields{
			"cooldown": stats.Cooldown,
		})
	}

	// Cast time
	if template.CastTimeRange[1] > 0 {
		stats.CastTime = template.CastTimeRange[0] +
			rng.Float64()*(template.CastTimeRange[1]-template.CastTimeRange[0])
		// Higher rarity has faster cast time
		stats.CastTime = stats.CastTime / (1.0 + float64(rarityScale)*0.1)
		g.logDebug("generated cast time", logrus.Fields{
			"cast_time": stats.CastTime,
		})
	}

	// Range
	if template.RangeRange[1] > 0 {
		stats.Range = template.RangeRange[0] +
			rng.Float64()*(template.RangeRange[1]-template.RangeRange[0])
		// Higher rarity has better range
		stats.Range = stats.Range * (1.0 + float64(rarityScale)*0.1)
		g.logDebug("generated range", logrus.Fields{
			"range": stats.Range,
		})
	}

	// Area size
	if template.AreaSizeRange[1] > 0 {
		stats.AreaSize = template.AreaSizeRange[0] +
			rng.Float64()*(template.AreaSizeRange[1]-template.AreaSizeRange[0])
		// Higher rarity has larger area
		stats.AreaSize = stats.AreaSize * (1.0 + float64(rarityScale)*0.15)
		g.logDebug("generated area size", logrus.Fields{
			"area_size": stats.AreaSize,
		})
	}

	// Duration
	if template.DurationRange[1] > 0 {
		stats.Duration = template.DurationRange[0] +
			rng.Float64()*(template.DurationRange[1]-template.DurationRange[0])
		// Higher rarity has longer duration
		stats.Duration = stats.Duration * (1.0 + float64(rarityScale)*0.2)
		g.logDebug("generated duration", logrus.Fields{
			"duration": stats.Duration,
		})
	}

	g.logDebug("spell stats generation complete", logrus.Fields{
		"damage":    stats.Damage,
		"healing":   stats.Healing,
		"mana_cost": stats.ManaCost,
		"cooldown":  stats.Cooldown,
	})
	return stats
}

// determineRarity calculates spell rarity based on depth and difficulty.
func (g *SpellGenerator) determineRarity(rng *rand.Rand, depth int, difficulty float64) Rarity {
	g.logDebug("determining rarity", logrus.Fields{
		"depth":      depth,
		"difficulty": difficulty,
	})
	// Base chance influenced by depth
	roll := rng.Float64()

	// Depth increases chance of higher rarity
	depthBonus := float64(depth) * 0.02
	difficultyBonus := difficulty * 0.1

	roll += depthBonus + difficultyBonus

	g.logDebug("rarity roll calculated", logrus.Fields{
		"base_roll":        roll - depthBonus - difficultyBonus,
		"depth_bonus":      depthBonus,
		"difficulty_bonus": difficultyBonus,
		"final_roll":       roll,
	})

	// Determine rarity thresholds
	var rarity Rarity
	switch {
	case roll < 0.50:
		rarity = RarityCommon
	case roll < 0.75:
		rarity = RarityUncommon
	case roll < 0.90:
		rarity = RarityRare
	case roll < 0.97:
		rarity = RarityEpic
	default:
		rarity = RarityLegendary
	}

	g.logDebug("rarity determined", logrus.Fields{"rarity": rarity})
	return rarity
}

// generateDescription creates flavor text for the spell.
func (g *SpellGenerator) generateDescription(spell *Spell) string {
	g.logDebug("generating spell description", logrus.Fields{
		"spell_type":    spell.Type,
		"spell_element": spell.Element,
		"spell_target":  spell.Target,
	})
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

	description := fmt.Sprintf("%s %s %s.", action, elementDesc, targetDesc)
	g.logDebug("spell description generated", logrus.Fields{"description": description})
	return description
}

// Validate checks if the generated spells are valid.
func (g *SpellGenerator) Validate(result interface{}) error {
	g.logDebug("validating spell generation result", logrus.Fields{})
	spells, ok := result.([]*Spell)
	if !ok {
		g.logError("validation failed: invalid result type", logrus.Fields{
			"expected_type": "[]*Spell",
		})
		return fmt.Errorf("result is not []*Spell")
	}

	if len(spells) == 0 {
		g.logError("validation failed: no spells generated", logrus.Fields{})
		return fmt.Errorf("no spells generated")
	}

	g.logDebug("validating individual spells", logrus.Fields{"spell_count": len(spells)})
	for i, spell := range spells {
		if err := g.validateSpell(i, spell); err != nil {
			g.logError("spell validation failed", logrus.Fields{
				"spell_index": i,
				"error":       err.Error(),
			})
			return err
		}
	}

	g.logInfo("spell validation complete", logrus.Fields{
		"validated_count": len(spells),
	})
	return nil
}

// validateSpell validates a single spell's properties and stats.
func (g *SpellGenerator) validateSpell(index int, spell *Spell) error {
	if spell == nil {
		g.logError("spell is nil", logrus.Fields{"spell_index": index})
		return fmt.Errorf("spell %d is nil", index)
	}

	g.logDebug("validating spell", logrus.Fields{
		"spell_index": index,
		"spell_name":  spell.Name,
	})

	if err := g.validateSpellBasicProperties(index, spell); err != nil {
		return err
	}

	if err := g.validateSpellEnumFields(index, spell); err != nil {
		return err
	}

	if err := g.validateSpellStats(index, spell); err != nil {
		return err
	}

	if err := g.validateSpellTypeSpecific(index, spell); err != nil {
		return err
	}

	g.validateSpellBalance(spell)
	g.logDebug("spell validation passed", logrus.Fields{
		"spell_index": index,
		"spell_name":  spell.Name,
	})
	return nil
}

// validateSpellBasicProperties validates spell name and basic properties.
func (g *SpellGenerator) validateSpellBasicProperties(index int, spell *Spell) error {
	g.logDebug("validating basic properties", logrus.Fields{
		"spell_index": index,
		"spell_name":  spell.Name,
	})
	if spell.Name == "" {
		g.logError("spell has empty name", logrus.Fields{"spell_index": index})
		return fmt.Errorf("spell %d has empty name", index)
	}
	return nil
}

// validateSpellEnumFields validates spell enum fields are within valid ranges.
func (g *SpellGenerator) validateSpellEnumFields(index int, spell *Spell) error {
	g.logDebug("validating enum fields", logrus.Fields{
		"spell_index": index,
		"type":        spell.Type,
		"element":     spell.Element,
		"rarity":      spell.Rarity,
		"target":      spell.Target,
	})
	if spell.Type < TypeOffensive || spell.Type > TypeSummon {
		g.logError("invalid spell type", logrus.Fields{
			"spell_index": index,
			"type":        spell.Type,
		})
		return fmt.Errorf("spell %d has invalid type: %d", index, spell.Type)
	}

	if spell.Element < ElementNone || spell.Element > ElementArcane {
		g.logError("invalid spell element", logrus.Fields{
			"spell_index": index,
			"element":     spell.Element,
		})
		return fmt.Errorf("spell %d has invalid element: %d", index, spell.Element)
	}

	if spell.Rarity < RarityCommon || spell.Rarity > RarityLegendary {
		g.logError("invalid spell rarity", logrus.Fields{
			"spell_index": index,
			"rarity":      spell.Rarity,
		})
		return fmt.Errorf("spell %d has invalid rarity: %d", index, spell.Rarity)
	}

	if spell.Target < TargetSelf || spell.Target > TargetAllEnemies {
		g.logError("invalid spell target", logrus.Fields{
			"spell_index": index,
			"target":      spell.Target,
		})
		return fmt.Errorf("spell %d has invalid target: %d", index, spell.Target)
	}

	return nil
}

// validateSpellStats validates spell stats are non-negative and within valid ranges.
func (g *SpellGenerator) validateSpellStats(index int, spell *Spell) error {
	g.logDebug("validating spell stats", logrus.Fields{
		"spell_index":    index,
		"mana_cost":      spell.Stats.ManaCost,
		"cooldown":       spell.Stats.Cooldown,
		"cast_time":      spell.Stats.CastTime,
		"range":          spell.Stats.Range,
		"required_level": spell.Stats.RequiredLevel,
	})
	if spell.Stats.ManaCost < 0 {
		g.logError("negative mana cost", logrus.Fields{
			"spell_index": index,
			"mana_cost":   spell.Stats.ManaCost,
		})
		return fmt.Errorf("spell %d has negative mana cost", index)
	}
	if spell.Stats.Cooldown < 0 {
		g.logError("negative cooldown", logrus.Fields{
			"spell_index": index,
			"cooldown":    spell.Stats.Cooldown,
		})
		return fmt.Errorf("spell %d has negative cooldown", index)
	}
	if spell.Stats.CastTime < 0 {
		g.logError("negative cast time", logrus.Fields{
			"spell_index": index,
			"cast_time":   spell.Stats.CastTime,
		})
		return fmt.Errorf("spell %d has negative cast time", index)
	}
	if spell.Stats.Range < 0 {
		g.logError("negative range", logrus.Fields{
			"spell_index": index,
			"range":       spell.Stats.Range,
		})
		return fmt.Errorf("spell %d has negative range", index)
	}
	if spell.Stats.RequiredLevel < 1 {
		g.logError("invalid required level", logrus.Fields{
			"spell_index":    index,
			"required_level": spell.Stats.RequiredLevel,
		})
		return fmt.Errorf("spell %d has invalid required level: %d", index, spell.Stats.RequiredLevel)
	}
	return nil
}

// validateSpellTypeSpecific validates type-specific spell requirements.
func (g *SpellGenerator) validateSpellTypeSpecific(index int, spell *Spell) error {
	g.logDebug("validating type-specific requirements", logrus.Fields{
		"spell_index": index,
		"spell_type":  spell.Type,
		"damage":      spell.Stats.Damage,
		"healing":     spell.Stats.Healing,
	})
	if spell.IsOffensive() && spell.Stats.Damage <= 0 {
		g.logError("offensive spell has no damage", logrus.Fields{
			"spell_index": index,
			"spell_name":  spell.Name,
			"damage":      spell.Stats.Damage,
		})
		return fmt.Errorf("offensive spell %d has no damage", index)
	}
	if spell.Type == TypeHealing && spell.Stats.Healing <= 0 {
		g.logError("healing spell has no healing", logrus.Fields{
			"spell_index": index,
			"spell_name":  spell.Name,
			"healing":     spell.Stats.Healing,
		})
		return fmt.Errorf("healing spell %d has no healing", index)
	}
	return nil
}

// validateSpellBalance validates spell balance metrics and logs warnings.
func (g *SpellGenerator) validateSpellBalance(spell *Spell) {
	if err := g.balanceConfig.ValidateDPS(spell); err != nil {
		g.logWarn("spell balance warning", logrus.Fields{"spell": spell.Name, "error": err.Error()})
	}
	if err := g.balanceConfig.ValidateHPS(spell); err != nil {
		g.logWarn("spell balance warning", logrus.Fields{"spell": spell.Name, "error": err.Error()})
	}
	if err := g.balanceConfig.ValidateManaCostEfficiency(spell); err != nil {
		g.logWarn("spell balance warning", logrus.Fields{"spell": spell.Name, "error": err.Error()})
	}
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
