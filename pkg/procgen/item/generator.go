// Package item provides procedural item generation.
// This file implements item generators for weapons, armor, consumables,
// and accessories with procedural stats and effects.
package item

import (
	"fmt"
	"math/rand"

	"github.com/opd-ai/venture/pkg/procgen"
	"github.com/sirupsen/logrus"
)

// ItemGenerator generates procedural items (weapons, armor, consumables).
type ItemGenerator struct {
	weaponTemplates     map[string][]ItemTemplate
	armorTemplates      map[string][]ItemTemplate
	consumableTemplates map[string][]ItemTemplate
	logger              *logrus.Entry
}

// NewItemGenerator creates a new item generator.
func NewItemGenerator() *ItemGenerator {
	return NewItemGeneratorWithLogger(nil)
}

// NewItemGeneratorWithLogger creates a new item generator with a logger.
func NewItemGeneratorWithLogger(logger *logrus.Logger) *ItemGenerator {
	var logEntry *logrus.Entry
	if logger != nil {
		logEntry = logger.WithField("generator", "item")
	}

	gen := &ItemGenerator{
		weaponTemplates:     make(map[string][]ItemTemplate),
		armorTemplates:      make(map[string][]ItemTemplate),
		consumableTemplates: make(map[string][]ItemTemplate),
		logger:              logEntry,
	}

	// Register fantasy genre templates
	gen.weaponTemplates["fantasy"] = GetFantasyWeaponTemplates()
	gen.armorTemplates["fantasy"] = GetFantasyArmorTemplates()
	gen.consumableTemplates["fantasy"] = GetFantasyConsumableTemplates()

	// Register sci-fi genre templates
	gen.weaponTemplates["scifi"] = GetSciFiWeaponTemplates()
	gen.armorTemplates["scifi"] = GetSciFiArmorTemplates()
	gen.consumableTemplates["scifi"] = GetSciFiConsumableTemplates()

	// Register horror genre templates
	gen.weaponTemplates["horror"] = GetHorrorWeaponTemplates()
	gen.armorTemplates["horror"] = GetHorrorArmorTemplates()
	gen.consumableTemplates["horror"] = GetHorrorConsumableTemplates()

	// Register cyberpunk genre templates
	gen.weaponTemplates["cyberpunk"] = GetCyberpunkWeaponTemplates()
	gen.armorTemplates["cyberpunk"] = GetCyberpunkArmorTemplates()
	gen.consumableTemplates["cyberpunk"] = GetCyberpunkConsumableTemplates()

	// Register post-apocalyptic genre templates
	gen.weaponTemplates["postapoc"] = GetPostApocWeaponTemplates()
	gen.armorTemplates["postapoc"] = GetPostApocArmorTemplates()
	gen.consumableTemplates["postapoc"] = GetPostApocConsumableTemplates()

	// Default templates
	gen.weaponTemplates[""] = GetFantasyWeaponTemplates()
	gen.armorTemplates[""] = GetFantasyArmorTemplates()
	gen.consumableTemplates[""] = GetFantasyConsumableTemplates()

	if logEntry != nil {
		logEntry.Debug("item generator initialized")
	}

	return gen
}

// Generate creates items based on the seed and parameters.
func (g *ItemGenerator) Generate(seed int64, params procgen.GenerationParams) (interface{}, error) {
	g.logDebug("starting item generation", logrus.Fields{
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

	itemTypeFilter := g.getItemTypeFilter(params)
	rng := rand.New(rand.NewSource(seed))

	items := make([]*Item, count)
	for i := 0; i < count; i++ {
		items[i] = g.generateSingleItem(seed, params, itemTypeFilter, i, rng)
	}

	typeFilter := "all"
	if itemTypeFilter != nil {
		typeFilter = itemTypeFilter.String()
	}

	g.logInfo("item generation complete", logrus.Fields{
		"count":      len(items),
		"seed":       seed,
		"genreID":    params.GenreID,
		"typeFilter": typeFilter,
	})

	return items, nil
}

// getItemTypeFilter extracts item type filter from custom parameters.
func (g *ItemGenerator) getItemTypeFilter(params procgen.GenerationParams) *ItemType {
	if params.Custom == nil {
		return nil
	}

	typeStr, ok := params.Custom["type"].(string)
	if !ok {
		return nil
	}

	var itemType ItemType
	switch typeStr {
	case "weapon":
		itemType = TypeWeapon
	case "armor":
		itemType = TypeArmor
	case "consumable":
		itemType = TypeConsumable
	default:
		return nil
	}
	return &itemType
}

// generateSingleItem creates one item using the provided random number generator.
// The rng parameter should be initialized with the desired seed for deterministic generation.
// baseSeed is used for the Item.Seed field to enable reproducibility.
func (g *ItemGenerator) generateSingleItem(baseSeed int64, params procgen.GenerationParams, itemTypeFilter *ItemType, index int, rng *rand.Rand) *Item {
	// Determine item type
	itemType := g.determineItemType(itemTypeFilter, rng)

	// Get templates based on type and genre
	var templates []ItemTemplate
	switch itemType {
	case TypeWeapon:
		templates = g.getWeaponTemplates(params.GenreID)
	case TypeArmor:
		templates = g.getArmorTemplates(params.GenreID)
	case TypeConsumable:
		templates = g.getConsumableTemplates(params.GenreID)
	case TypeAccessory:
		// For now, accessories use armor templates
		templates = g.getArmorTemplates(params.GenreID)
	}

	if len(templates) == 0 {
		// Fallback to fantasy weapons
		templates = g.weaponTemplates["fantasy"]
	}

	// Select template
	templateIndex := rng.Intn(len(templates))
	template := templates[templateIndex]

	// Generate item
	item := &Item{
		ID:             fmt.Sprintf("item_%d_%d", params.Depth, index),
		Type:           itemType,
		WeaponType:     template.WeaponType,
		ArmorType:      template.ArmorType,
		ConsumableType: template.ConsumableType,
		Seed:           baseSeed + int64(index)*1000, // Unique seed per item
		Tags:           make([]string, len(template.Tags)),
	}
	copy(item.Tags, template.Tags)

	// Copy class restrictions from template
	if len(template.ClassRestrictions) > 0 {
		item.ClassRestrictions = make([]string, len(template.ClassRestrictions))
		copy(item.ClassRestrictions, template.ClassRestrictions)
	}

	// Determine rarity based on depth
	item.Rarity = g.determineRarity(params.Depth, rng)

	// Generate name
	item.Name = g.generateName(template, item.Rarity, rng)

	// Generate stats
	item.Stats = g.generateStats(template, params.Depth, item.Rarity, params.Difficulty, rng)

	// Generate description
	item.Description = g.generateDescription(item, template, rng)

	// Generate spell effect for scrolls
	if item.ConsumableType == ConsumableScroll && len(template.SpellEffectIDs) > 0 {
		spellIndex := rng.Intn(len(template.SpellEffectIDs))
		item.SpellEffectID = template.SpellEffectIDs[spellIndex]
		if spellIndex < len(template.SpellDurations) {
			item.SpellDuration = template.SpellDurations[spellIndex]
		}
		if spellIndex < len(template.SpellTargetTypes) {
			item.SpellTargetType = template.SpellTargetTypes[spellIndex]
		}
		if spellIndex < len(template.SpellRadii) {
			item.SpellRadius = template.SpellRadii[spellIndex]
		}
	}

	return item
}

// determineItemType selects what type of item to generate.
func (g *ItemGenerator) determineItemType(filter *ItemType, rng *rand.Rand) ItemType {
	if filter != nil {
		return *filter
	}

	// Random distribution: 40% weapon, 35% armor, 20% consumable, 5% accessory
	roll := rng.Float64()
	switch {
	case roll < 0.40:
		return TypeWeapon
	case roll < 0.75:
		return TypeArmor
	case roll < 0.95:
		return TypeConsumable
	default:
		return TypeAccessory
	}
}

// getWeaponTemplates retrieves weapon templates for a genre.
func (g *ItemGenerator) getWeaponTemplates(genreID string) []ItemTemplate {
	templates := g.weaponTemplates[genreID]
	if templates == nil {
		if g.logger != nil && genreID != "" {
			g.logger.WithField("genreID", genreID).Warn("unknown genre, using default weapon templates")
		}
		templates = g.weaponTemplates[""] // fallback to default
	}
	return templates
}

// getArmorTemplates retrieves armor templates for a genre.
func (g *ItemGenerator) getArmorTemplates(genreID string) []ItemTemplate {
	templates := g.armorTemplates[genreID]
	if templates == nil {
		if g.logger != nil && genreID != "" {
			g.logger.WithField("genreID", genreID).Warn("unknown genre, using default armor templates")
		}
		templates = g.armorTemplates[""] // fallback to default
	}
	return templates
}

// getConsumableTemplates retrieves consumable templates for a genre.
func (g *ItemGenerator) getConsumableTemplates(genreID string) []ItemTemplate {
	templates := g.consumableTemplates[genreID]
	if templates == nil {
		if g.logger != nil && genreID != "" {
			g.logger.WithField("genreID", genreID).Warn("unknown genre, using default consumable templates")
		}
		templates = g.consumableTemplates[""] // fallback to default
	}
	return templates
}

// determineRarity calculates item rarity based on depth and random chance.
func (g *ItemGenerator) determineRarity(depth int, rng *rand.Rand) Rarity {
	// Base probabilities
	roll := rng.Float64()

	// Increase rare drops at higher depths
	depthBonus := float64(depth) * 0.01
	if depthBonus > 0.3 {
		depthBonus = 0.3
	}

	// Adjust thresholds based on depth
	commonThreshold := 0.50 - depthBonus
	uncommonThreshold := 0.80 - depthBonus
	rareThreshold := 0.93 - depthBonus
	epicThreshold := 0.98 - depthBonus

	switch {
	case roll < commonThreshold:
		return RarityCommon
	case roll < uncommonThreshold:
		return RarityUncommon
	case roll < rareThreshold:
		return RarityRare
	case roll < epicThreshold:
		return RarityEpic
	default:
		return RarityLegendary
	}
}

// generateName creates a name for the item.
func (g *ItemGenerator) generateName(template ItemTemplate, rarity Rarity, rng *rand.Rand) string {
	prefix := template.NamePrefixes[rng.Intn(len(template.NamePrefixes))]
	suffix := template.NameSuffixes[rng.Intn(len(template.NameSuffixes))]

	// Add rarity prefix for rare+ items
	rarityPrefix := ""
	switch rarity {
	case RarityEpic:
		rarityPrefixes := []string{"Masterwork", "Superior", "Exquisite", "Prime"}
		rarityPrefix = rarityPrefixes[rng.Intn(len(rarityPrefixes))] + " "
	case RarityLegendary:
		rarityPrefixes := []string{"Legendary", "Mythic", "Ancient", "Divine"}
		rarityPrefix = rarityPrefixes[rng.Intn(len(rarityPrefixes))] + " "
	}

	return rarityPrefix + prefix + " " + suffix
}

// generateStats creates stats for the item.
func (g *ItemGenerator) generateStats(template ItemTemplate, depth int, rarity Rarity, difficulty float64, rng *rand.Rand) Stats {
	stats := Stats{RequiredLevel: 1 + depth/2}

	g.generateCombatStats(&stats, template, depth, rarity, difficulty, rng)
	g.generateItemProperties(&stats, template, rarity, rng)

	if template.IsProjectile {
		g.generateProjectileStats(&stats, template, rarity, rng)
	}

	return stats
}

// generateCombatStats generates damage, defense, and attack speed stats.
func (g *ItemGenerator) generateCombatStats(stats *Stats, template ItemTemplate, depth int, rarity Rarity, difficulty float64, rng *rand.Rand) {
	if template.DamageRange[1] > 0 {
		baseDamage := template.DamageRange[0] + rng.Intn(template.DamageRange[1]-template.DamageRange[0]+1)
		stats.Damage = g.scaleStatByFactors(baseDamage, depth, rarity, difficulty)
	}

	if template.DefenseRange[1] > 0 {
		baseDefense := template.DefenseRange[0] + rng.Intn(template.DefenseRange[1]-template.DefenseRange[0]+1)
		stats.Defense = g.scaleStatByFactors(baseDefense, depth, rarity, difficulty)
	}

	if template.AttackSpeedRange[1] > 0 {
		stats.AttackSpeed = template.AttackSpeedRange[0] + rng.Float64()*(template.AttackSpeedRange[1]-template.AttackSpeedRange[0])
		stats.AttackSpeed += float64(rarity) * 0.05
	}
}

// generateItemProperties generates value, weight, and durability.
func (g *ItemGenerator) generateItemProperties(stats *Stats, template ItemTemplate, rarity Rarity, rng *rand.Rand) {
	baseValue := template.ValueRange[0] + rng.Intn(template.ValueRange[1]-template.ValueRange[0]+1)
	stats.Value = g.scaleStatByFactors(baseValue, 0, rarity, 1.0)
	stats.Weight = template.WeightRange[0] + rng.Float64()*(template.WeightRange[1]-template.WeightRange[0])

	if template.DurabilityRange[1] > 0 {
		stats.DurabilityMax = template.DurabilityRange[0] + rng.Intn(template.DurabilityRange[1]-template.DurabilityRange[0]+1)
		stats.DurabilityMax += int(rarity) * 20
		stats.Durability = stats.DurabilityMax
	}
}

// generateProjectileStats generates projectile-specific stats and properties.
func (g *ItemGenerator) generateProjectileStats(stats *Stats, template ItemTemplate, rarity Rarity, rng *rand.Rand) {
	stats.IsProjectile = true
	stats.ProjectileType = template.ProjectileType
	stats.ProjectileSpeed = template.ProjectileSpeedRange[0] + rng.Float64()*(template.ProjectileSpeedRange[1]-template.ProjectileSpeedRange[0])
	stats.ProjectileSpeed += float64(rarity) * 20.0
	stats.ProjectileLifetime = template.ProjectileLifetime

	g.generateProjectileSpecialProperties(stats, template, rarity, rng)
}

// generateProjectileSpecialProperties generates pierce, bounce, and explosive properties.
func (g *ItemGenerator) generateProjectileSpecialProperties(stats *Stats, template ItemTemplate, rarity Rarity, rng *rand.Rand) {
	rarityMultiplier := g.getRarityChanceMultiplier(rarity)

	if rng.Float64() < template.PierceChance*rarityMultiplier {
		if template.PierceRange[1] > template.PierceRange[0] {
			stats.Pierce = template.PierceRange[0] + rng.Intn(template.PierceRange[1]-template.PierceRange[0]+1)
		} else {
			stats.Pierce = template.PierceRange[0]
		}
		stats.Pierce += int(rarity) / 2
	}

	if rng.Float64() < template.BounceChance*rarityMultiplier {
		if template.BounceRange[1] > template.BounceRange[0] {
			stats.Bounce = template.BounceRange[0] + rng.Intn(template.BounceRange[1]-template.BounceRange[0]+1)
		} else {
			stats.Bounce = template.BounceRange[0]
		}
	}

	if rng.Float64() < template.ExplosiveChance*rarityMultiplier {
		stats.Explosive = true
		stats.ExplosionRadius = template.ExplosionRadiusRange[0] + rng.Float64()*(template.ExplosionRadiusRange[1]-template.ExplosionRadiusRange[0])
		stats.ExplosionRadius += float64(rarity) * 10.0
	}
}

// scaleStatByFactors applies scaling based on depth, rarity, and difficulty.
func (g *ItemGenerator) scaleStatByFactors(baseStat, depth int, rarity Rarity, difficulty float64) int {
	// Depth scaling: +10% per depth level
	depthMultiplier := 1.0 + (float64(depth) * 0.1)

	// Rarity scaling
	rarityMultiplier := 1.0
	switch rarity {
	case RarityUncommon:
		rarityMultiplier = 1.2
	case RarityRare:
		rarityMultiplier = 1.5
	case RarityEpic:
		rarityMultiplier = 2.0
	case RarityLegendary:
		rarityMultiplier = 3.0
	}

	// Difficulty scaling
	difficultyMultiplier := 0.8 + (difficulty * 0.4) // Range: 0.8 to 1.2

	result := float64(baseStat) * depthMultiplier * rarityMultiplier * difficultyMultiplier
	return int(result)
}

// getRarityChanceMultiplier returns a multiplier for special property chances based on rarity.
// Higher rarity items have increased chance of special projectile properties.
func (g *ItemGenerator) getRarityChanceMultiplier(rarity Rarity) float64 {
	switch rarity {
	case RarityCommon:
		return 1.0
	case RarityUncommon:
		return 1.5
	case RarityRare:
		return 2.0
	case RarityEpic:
		return 2.5
	case RarityLegendary:
		return 3.0
	default:
		return 1.0
	}
}

// generateDescription creates flavor text for the item using deterministic RNG.
func (g *ItemGenerator) generateDescription(item *Item, template ItemTemplate, rng *rand.Rand) string {
	descriptions := make([]string, 0)

	// Add type-specific descriptions
	switch item.Type {
	case TypeWeapon:
		descriptions = append(descriptions,
			"A finely crafted weapon.",
			"This weapon has seen many battles.",
			"The blade gleams with deadly intent.",
			"A reliable tool for any warrior.",
		)
	case TypeArmor:
		descriptions = append(descriptions,
			"Sturdy protection for the weary traveler.",
			"This armor has saved many lives.",
			"Well-crafted and dependable.",
			"A solid piece of defensive equipment.",
		)
	case TypeConsumable:
		descriptions = append(descriptions,
			"This should prove useful in a pinch.",
			"A valuable resource for any adventurer.",
			"Use wisely, supplies are limited.",
		)
	}

	// Add rarity-specific text
	switch item.Rarity {
	case RarityEpic:
		descriptions = append(descriptions,
			"An exceptional piece of craftsmanship.",
			"The work of a true master.",
		)
	case RarityLegendary:
		descriptions = append(descriptions,
			"Legends speak of this item's power.",
			"Few have wielded such a treasure.",
		)
	}

	// Return a random description using the seeded RNG
	if len(descriptions) > 0 {
		return descriptions[rng.Intn(len(descriptions))]
	}
	return "A mysterious item."
}

// Validate checks if the generated items are valid.
func (g *ItemGenerator) Validate(result interface{}) error {
	items, ok := result.([]*Item)
	if !ok {
		return fmt.Errorf("invalid result type: expected []*Item")
	}

	for i, item := range items {
		if item == nil {
			return fmt.Errorf("item %d is nil", i)
		}
		if item.Name == "" {
			return fmt.Errorf("item %d has empty name", i)
		}
		if item.Type == TypeWeapon && item.Stats.Damage <= 0 {
			return fmt.Errorf("weapon %d (%s) has invalid damage: %d", i, item.Name, item.Stats.Damage)
		}
		if item.Type == TypeArmor && item.Stats.Defense <= 0 {
			return fmt.Errorf("armor %d (%s) has invalid defense: %d", i, item.Name, item.Stats.Defense)
		}
		if item.Stats.Value < 0 {
			return fmt.Errorf("item %d (%s) has negative value", i, item.Name)
		}
	}

	return nil
}

// logDebug logs a debug message if logger and level are configured.
func (g *ItemGenerator) logDebug(msg string, fields logrus.Fields) {
	if g.logger != nil && g.logger.Logger.GetLevel() >= logrus.DebugLevel {
		g.logger.WithFields(fields).Debug(msg)
	}
}

// logInfo logs an info message if logger is configured.
func (g *ItemGenerator) logInfo(msg string, fields logrus.Fields) {
	if g.logger != nil {
		g.logger.WithFields(fields).Info(msg)
	}
}
