// Package engine provides the WeatherFactionResistanceSystem which bridges
// weather conditions with faction-based combat modifiers. Different factions
// have affinities or vulnerabilities to specific weather types, creating
// tactical depth: fighting a fire cult in the rain weakens them, but engaging
// ice mages during a blizzard strengthens them.
package engine

import (
	"math/rand"

	"github.com/opd-ai/venture/pkg/rendering/particles"
	"github.com/sirupsen/logrus"
)

// WeatherFactionResistanceComponent stores weather-based modifiers for
// entities belonging to factions with elemental affinities.
// Pure data — all logic lives in WeatherFactionResistanceSystem.
type WeatherFactionResistanceComponent struct {
	// FactionID identifies the entity's faction for affinity lookup.
	FactionID string

	// DamageModifier multiplies outgoing damage (1.0 = no change).
	// Values > 1.0 = buffed, < 1.0 = weakened.
	DamageModifier float64

	// DefenseModifier multiplies incoming damage reduction (1.0 = no change).
	// Values > 1.0 = more resistant, < 1.0 = more vulnerable.
	DefenseModifier float64

	// StatusResistModifier affects resistance to status effects (1.0 = no change).
	StatusResistModifier float64

	// Active indicates whether a weather-faction interaction is in effect.
	Active bool

	// WeatherType records the weather type currently affecting this entity.
	WeatherType string

	// AffinityType describes the faction's elemental affinity (fire, ice, etc.).
	AffinityType string
}

// Type returns the component type identifier.
func (c *WeatherFactionResistanceComponent) Type() string {
	return "weather_faction_resistance"
}

// factionAffinityType categorizes faction elemental affinities.
type factionAffinityType string

const (
	affinityFire    factionAffinityType = "fire"
	affinityIce     factionAffinityType = "ice"
	affinityNature  factionAffinityType = "nature"
	affinityStorm   factionAffinityType = "storm"
	affinityDark    factionAffinityType = "dark"
	affinityTech    factionAffinityType = "tech"
	affinityNeutral factionAffinityType = "neutral"
)

// weatherFactionInteraction defines how a faction affinity interacts with weather.
type weatherFactionInteraction struct {
	// DamageMultiplier applied to outgoing damage in this weather.
	DamageMultiplier float64
	// DefenseMultiplier applied to defense in this weather.
	DefenseMultiplier float64
	// StatusResistMultiplier applied to status resist in this weather.
	StatusResistMultiplier float64
}

// WeatherFactionResistanceSystem modifies entity combat stats based on the
// interaction between their faction's elemental affinity and current weather.
//
// Affinity-Weather Interactions:
//   - Fire factions: Buffed in heat/clear, weakened in rain/snow
//   - Ice factions: Buffed in snow/rain, weakened in heat/clear
//   - Nature factions: Buffed in rain, weakened in sandstorm/heat
//   - Storm factions: Buffed in rain/storm, weakened in clear/snow
//   - Dark factions: Buffed in fog/overcast, weakened in clear
//   - Tech factions: Weakened in sandstorm (sensor interference)
//
// Genre Modifiers:
//   - Fantasy: Full affinity effects (magic is strong)
//   - Scifi: Reduced natural affinities, tech affinity boosted
//   - Horror: Dark affinity boosted, others reduced
//   - Cyberpunk: Tech and storm affinities boosted
//   - PostApoc: All affinities reduced (harsh world)
type WeatherFactionResistanceSystem struct {
	world  *World
	rng    *rand.Rand
	logger *logrus.Entry

	genreID string

	// updateInterval controls how often weather checks occur (seconds).
	updateInterval  float64
	timeSinceUpdate float64

	// factionAffinities maps faction ID patterns to their elemental affinity.
	// Uses substring matching: "flame" in faction ID -> fire affinity.
	factionAffinities map[string]factionAffinityType

	// weatherInteractions maps (affinity, weather) -> interaction modifiers.
	weatherInteractions map[factionAffinityType]map[particles.WeatherType]weatherFactionInteraction

	// genreMultipliers scale affinity effects by genre.
	genreMultipliers map[string]float64

	// activeModifiers tracks entities with current weather modifiers.
	activeModifiers map[uint64]bool
}

// NewWeatherFactionResistanceSystem creates a new weather-faction resistance system.
func NewWeatherFactionResistanceSystem(world *World, seed int64) *WeatherFactionResistanceSystem {
	var logEntry *logrus.Entry
	if world != nil && world.logger != nil {
		logEntry = world.logger.Logger.WithField("system_name", "weather_faction_resistance")
	}

	sys := &WeatherFactionResistanceSystem{
		world:           world,
		rng:             rand.New(rand.NewSource(seed)),
		logger:          logEntry,
		updateInterval:  0.5, // Update twice per second
		genreID:         "fantasy",
		activeModifiers: make(map[uint64]bool, 64),
	}

	sys.initializeFactionAffinities()
	sys.initializeWeatherInteractions()
	sys.initializeGenreMultipliers()

	if logEntry != nil {
		logEntry.Debug("weather faction resistance system created")
	}

	return sys
}

// initializeFactionAffinities sets up faction ID pattern -> affinity mappings.
func (s *WeatherFactionResistanceSystem) initializeFactionAffinities() {
	s.factionAffinities = map[string]factionAffinityType{
		// Fire-related factions
		"flame":   affinityFire,
		"fire":    affinityFire,
		"ember":   affinityFire,
		"ash":     affinityFire,
		"inferno": affinityFire,
		"pyro":    affinityFire,
		"sun":     affinityFire,
		"blaze":   affinityFire,

		// Ice-related factions
		"ice":      affinityIce,
		"frost":    affinityIce,
		"snow":     affinityIce,
		"winter":   affinityIce,
		"glacier":  affinityIce,
		"frozen":   affinityIce,
		"cold":     affinityIce,
		"cryomage": affinityIce,

		// Nature-related factions
		"forest":  affinityNature,
		"grove":   affinityNature,
		"nature":  affinityNature,
		"druid":   affinityNature,
		"wild":    affinityNature,
		"verdant": affinityNature,
		"leaf":    affinityNature,
		"moss":    affinityNature,

		// Storm-related factions
		"storm":     affinityStorm,
		"thunder":   affinityStorm,
		"lightning": affinityStorm,
		"tempest":   affinityStorm,
		"hurricane": affinityStorm,
		"wind":      affinityStorm,

		// Dark-related factions
		"shadow":   affinityDark,
		"dark":     affinityDark,
		"night":    affinityDark,
		"void":     affinityDark,
		"umbra":    affinityDark,
		"eclipse":  affinityDark,
		"necro":    affinityDark,
		"undead":   affinityDark,
		"spectral": affinityDark,

		// Tech-related factions
		"tech":   affinityTech,
		"cyber":  affinityTech,
		"mech":   affinityTech,
		"droid":  affinityTech,
		"robot":  affinityTech,
		"synth":  affinityTech,
		"corp":   affinityTech,
		"chrome": affinityTech,
	}
}

// initializeWeatherInteractions sets up the affinity-weather interaction matrix.
func (s *WeatherFactionResistanceSystem) initializeWeatherInteractions() {
	s.weatherInteractions = make(map[factionAffinityType]map[particles.WeatherType]weatherFactionInteraction)

	// Fire affinity interactions
	s.weatherInteractions[affinityFire] = map[particles.WeatherType]weatherFactionInteraction{
		particles.WeatherDust:      {DamageMultiplier: 1.15, DefenseMultiplier: 1.10, StatusResistMultiplier: 1.10}, // Dry = good for fire
		particles.WeatherRain:      {DamageMultiplier: 0.80, DefenseMultiplier: 0.85, StatusResistMultiplier: 0.90}, // Rain weakens fire
		particles.WeatherSnow:      {DamageMultiplier: 0.75, DefenseMultiplier: 0.80, StatusResistMultiplier: 0.85}, // Cold weakens fire
		particles.WeatherFog:       {DamageMultiplier: 0.95, DefenseMultiplier: 1.00, StatusResistMultiplier: 1.00}, // Moisture weakens
		particles.WeatherSandstorm: {DamageMultiplier: 1.05, DefenseMultiplier: 1.00, StatusResistMultiplier: 1.00}, // Dry sand = neutral+
		particles.WeatherAsh:       {DamageMultiplier: 1.20, DefenseMultiplier: 1.10, StatusResistMultiplier: 1.05}, // Fire loves ash
	}

	// Ice affinity interactions
	s.weatherInteractions[affinityIce] = map[particles.WeatherType]weatherFactionInteraction{
		particles.WeatherDust:      {DamageMultiplier: 0.85, DefenseMultiplier: 0.90, StatusResistMultiplier: 0.90}, // Dry = bad for ice
		particles.WeatherRain:      {DamageMultiplier: 1.05, DefenseMultiplier: 1.05, StatusResistMultiplier: 1.05}, // Moisture helps
		particles.WeatherSnow:      {DamageMultiplier: 1.20, DefenseMultiplier: 1.15, StatusResistMultiplier: 1.15}, // Snow = home turf
		particles.WeatherFog:       {DamageMultiplier: 1.05, DefenseMultiplier: 1.05, StatusResistMultiplier: 1.00}, // Cold fog is good
		particles.WeatherSandstorm: {DamageMultiplier: 0.90, DefenseMultiplier: 0.95, StatusResistMultiplier: 0.95}, // Hot sand = bad
	}

	// Nature affinity interactions
	s.weatherInteractions[affinityNature] = map[particles.WeatherType]weatherFactionInteraction{
		particles.WeatherDust:      {DamageMultiplier: 0.90, DefenseMultiplier: 0.95, StatusResistMultiplier: 0.95}, // Dry = slight penalty
		particles.WeatherRain:      {DamageMultiplier: 1.15, DefenseMultiplier: 1.10, StatusResistMultiplier: 1.10}, // Rain nourishes nature
		particles.WeatherSnow:      {DamageMultiplier: 0.90, DefenseMultiplier: 0.95, StatusResistMultiplier: 0.95}, // Cold slows growth
		particles.WeatherFog:       {DamageMultiplier: 1.05, DefenseMultiplier: 1.05, StatusResistMultiplier: 1.05}, // Forest fog is good
		particles.WeatherSandstorm: {DamageMultiplier: 0.75, DefenseMultiplier: 0.80, StatusResistMultiplier: 0.80}, // Sand kills plants
	}

	// Storm affinity interactions
	s.weatherInteractions[affinityStorm] = map[particles.WeatherType]weatherFactionInteraction{
		particles.WeatherDust:      {DamageMultiplier: 0.90, DefenseMultiplier: 0.90, StatusResistMultiplier: 0.95}, // Clear skies = no storms
		particles.WeatherRain:      {DamageMultiplier: 1.20, DefenseMultiplier: 1.15, StatusResistMultiplier: 1.10}, // Storm power!
		particles.WeatherSnow:      {DamageMultiplier: 0.95, DefenseMultiplier: 0.95, StatusResistMultiplier: 1.00}, // Cold storms = ok
		particles.WeatherFog:       {DamageMultiplier: 1.10, DefenseMultiplier: 1.05, StatusResistMultiplier: 1.05}, // Pre-storm conditions
		particles.WeatherSandstorm: {DamageMultiplier: 1.10, DefenseMultiplier: 1.05, StatusResistMultiplier: 1.00}, // Sandstorms have wind
	}

	// Dark affinity interactions
	s.weatherInteractions[affinityDark] = map[particles.WeatherType]weatherFactionInteraction{
		particles.WeatherDust:      {DamageMultiplier: 0.85, DefenseMultiplier: 0.90, StatusResistMultiplier: 0.90}, // Clear = bright = bad
		particles.WeatherRain:      {DamageMultiplier: 1.00, DefenseMultiplier: 1.00, StatusResistMultiplier: 1.00}, // Overcast is neutral
		particles.WeatherSnow:      {DamageMultiplier: 0.95, DefenseMultiplier: 0.95, StatusResistMultiplier: 0.95}, // Snow is bright
		particles.WeatherFog:       {DamageMultiplier: 1.20, DefenseMultiplier: 1.15, StatusResistMultiplier: 1.15}, // Fog obscures = power
		particles.WeatherSmog:      {DamageMultiplier: 1.15, DefenseMultiplier: 1.10, StatusResistMultiplier: 1.10}, // Smog = darkness
		particles.WeatherSandstorm: {DamageMultiplier: 1.10, DefenseMultiplier: 1.05, StatusResistMultiplier: 1.05}, // Sand obscures vision
	}

	// Tech affinity interactions (mainly affected by sensor-disrupting weather)
	s.weatherInteractions[affinityTech] = map[particles.WeatherType]weatherFactionInteraction{
		particles.WeatherDust:      {DamageMultiplier: 1.10, DefenseMultiplier: 1.05, StatusResistMultiplier: 1.05}, // Clear = sensors work
		particles.WeatherRain:      {DamageMultiplier: 0.95, DefenseMultiplier: 0.95, StatusResistMultiplier: 0.95}, // Water interference
		particles.WeatherSnow:      {DamageMultiplier: 0.90, DefenseMultiplier: 0.90, StatusResistMultiplier: 0.90}, // Cold slows circuits
		particles.WeatherFog:       {DamageMultiplier: 0.85, DefenseMultiplier: 0.90, StatusResistMultiplier: 0.95}, // Visual sensors fail
		particles.WeatherSandstorm: {DamageMultiplier: 0.70, DefenseMultiplier: 0.75, StatusResistMultiplier: 0.80}, // Sand destroys tech
		particles.WeatherRadiation: {DamageMultiplier: 0.80, DefenseMultiplier: 0.85, StatusResistMultiplier: 0.90}, // EMI interference
	}

	// Neutral affinity - no weather interactions
	s.weatherInteractions[affinityNeutral] = map[particles.WeatherType]weatherFactionInteraction{
		particles.WeatherDust:      {DamageMultiplier: 1.0, DefenseMultiplier: 1.0, StatusResistMultiplier: 1.0},
		particles.WeatherRain:      {DamageMultiplier: 1.0, DefenseMultiplier: 1.0, StatusResistMultiplier: 1.0},
		particles.WeatherSnow:      {DamageMultiplier: 1.0, DefenseMultiplier: 1.0, StatusResistMultiplier: 1.0},
		particles.WeatherFog:       {DamageMultiplier: 1.0, DefenseMultiplier: 1.0, StatusResistMultiplier: 1.0},
		particles.WeatherSandstorm: {DamageMultiplier: 1.0, DefenseMultiplier: 1.0, StatusResistMultiplier: 1.0},
	}
}

// initializeGenreMultipliers sets up genre-based scaling for affinity effects.
func (s *WeatherFactionResistanceSystem) initializeGenreMultipliers() {
	s.genreMultipliers = map[string]float64{
		"fantasy":   1.0,  // Full magical affinities
		"scifi":     0.7,  // Reduced natural, tech still works
		"horror":    0.85, // Dark affinity boosted via separate check
		"cyberpunk": 0.9,  // Tech and storm work well
		"postapoc":  0.6,  // Harsh world reduces all affinities
	}
}

// SetGenre sets the genre for genre-aware affinity scaling.
func (s *WeatherFactionResistanceSystem) SetGenre(genreID string) {
	s.genreID = genreID
	if s.logger != nil {
		s.logger.WithField("genre", genreID).Debug("genre set for weather faction resistance")
	}
}

// Update processes all entities and applies weather-faction modifiers.
func (s *WeatherFactionResistanceSystem) Update(entities []*Entity, deltaTime float64) {
	s.timeSinceUpdate += deltaTime
	if s.timeSinceUpdate < s.updateInterval {
		return
	}
	s.timeSinceUpdate = 0

	// Find active weather
	weatherType, intensity, found := s.findActiveWeather(entities)
	if !found {
		s.clearAllModifiers(entities)
		return
	}

	// Process each entity with a faction
	for _, entity := range entities {
		s.processEntity(entity, weatherType, intensity)
	}
}

// findActiveWeather scans for active WeatherComponent and returns its type.
func (s *WeatherFactionResistanceSystem) findActiveWeather(entities []*Entity) (particles.WeatherType, particles.WeatherIntensity, bool) {
	for _, entity := range entities {
		comp, ok := entity.GetComponent("weather")
		if !ok {
			continue
		}
		weather, ok := comp.(*WeatherComponent)
		if !ok || !weather.Active {
			continue
		}
		return weather.Config.Type, weather.Config.Intensity, true
	}
	return 0, 0, false
}

// processEntity checks if entity has faction affinity and applies weather modifiers.
func (s *WeatherFactionResistanceSystem) processEntity(entity *Entity, weatherType particles.WeatherType, intensity particles.WeatherIntensity) {
	// Get faction component
	factionComp, ok := entity.GetComponent("faction")
	if !ok {
		s.removeModifierIfPresent(entity)
		return
	}
	faction, ok := factionComp.(*FactionComponent)
	if !ok {
		s.removeModifierIfPresent(entity)
		return
	}

	// Determine faction affinity
	affinity := s.getFactionAffinity(faction.FactionID)
	if affinity == affinityNeutral {
		s.removeModifierIfPresent(entity)
		return
	}

	// Get weather interaction for this affinity
	interaction := s.getWeatherInteraction(affinity, weatherType)
	if interaction.DamageMultiplier == 1.0 && interaction.DefenseMultiplier == 1.0 {
		s.removeModifierIfPresent(entity)
		return
	}

	// Apply genre scaling
	genreScale := s.getGenreMultiplier()

	// Scale effects by intensity (light=25%, medium=50%, heavy=75%, extreme=100%)
	intensityScale := s.getIntensityScale(intensity)

	// Calculate final modifiers (blend toward 1.0 based on intensity and genre)
	dmgMod := s.blendModifier(interaction.DamageMultiplier, genreScale, intensityScale)
	defMod := s.blendModifier(interaction.DefenseMultiplier, genreScale, intensityScale)
	statusMod := s.blendModifier(interaction.StatusResistMultiplier, genreScale, intensityScale)

	// Apply or update resistance component
	s.applyResistanceComponent(entity, faction.FactionID, affinity, weatherType, dmgMod, defMod, statusMod)
}

// getFactionAffinity determines elemental affinity from faction ID via pattern matching.
func (s *WeatherFactionResistanceSystem) getFactionAffinity(factionID string) factionAffinityType {
	lowerID := toLowerSimple(factionID)
	for pattern, affinity := range s.factionAffinities {
		if containsSimple(lowerID, pattern) {
			return affinity
		}
	}
	return affinityNeutral
}

// getWeatherInteraction returns the interaction modifiers for affinity+weather combo.
func (s *WeatherFactionResistanceSystem) getWeatherInteraction(affinity factionAffinityType, weather particles.WeatherType) weatherFactionInteraction {
	affinityMap, ok := s.weatherInteractions[affinity]
	if !ok {
		return weatherFactionInteraction{DamageMultiplier: 1.0, DefenseMultiplier: 1.0, StatusResistMultiplier: 1.0}
	}
	interaction, ok := affinityMap[weather]
	if !ok {
		return weatherFactionInteraction{DamageMultiplier: 1.0, DefenseMultiplier: 1.0, StatusResistMultiplier: 1.0}
	}
	return interaction
}

// getGenreMultiplier returns the genre-based scaling factor.
func (s *WeatherFactionResistanceSystem) getGenreMultiplier() float64 {
	mult, ok := s.genreMultipliers[s.genreID]
	if !ok {
		return 1.0
	}
	return mult
}

// getIntensityScale converts weather intensity to a 0.25-1.0 scale factor.
func (s *WeatherFactionResistanceSystem) getIntensityScale(intensity particles.WeatherIntensity) float64 {
	switch intensity {
	case particles.IntensityLight:
		return 0.25
	case particles.IntensityMedium:
		return 0.50
	case particles.IntensityHeavy:
		return 0.75
	case particles.IntensityExtreme:
		return 1.0
	default:
		return 0.50
	}
}

// blendModifier blends a modifier toward 1.0 based on genre and intensity scaling.
func (s *WeatherFactionResistanceSystem) blendModifier(baseMod, genreScale, intensityScale float64) float64 {
	// Calculate deviation from 1.0
	deviation := baseMod - 1.0
	// Scale deviation by genre and intensity
	scaledDeviation := deviation * genreScale * intensityScale
	// Return 1.0 + scaled deviation
	return 1.0 + scaledDeviation
}

// applyResistanceComponent adds or updates the resistance component on entity.
func (s *WeatherFactionResistanceSystem) applyResistanceComponent(
	entity *Entity,
	factionID string,
	affinity factionAffinityType,
	weatherType particles.WeatherType,
	dmgMod, defMod, statusMod float64,
) {
	comp, exists := entity.GetComponent("weather_faction_resistance")

	var resistance *WeatherFactionResistanceComponent
	if exists {
		resistance = comp.(*WeatherFactionResistanceComponent)
	} else {
		resistance = &WeatherFactionResistanceComponent{}
		entity.AddComponent(resistance)
	}

	resistance.FactionID = factionID
	resistance.AffinityType = string(affinity)
	resistance.WeatherType = weatherType.String()
	resistance.DamageModifier = dmgMod
	resistance.DefenseModifier = defMod
	resistance.StatusResistModifier = statusMod
	resistance.Active = true

	s.activeModifiers[entity.ID] = true

	if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
		s.logger.WithFields(logrus.Fields{
			"entity_id":   entity.ID,
			"faction":     factionID,
			"affinity":    affinity,
			"weather":     weatherType.String(),
			"damage_mod":  dmgMod,
			"defense_mod": defMod,
			"status_mod":  statusMod,
		}).Debug("applied weather faction resistance")
	}
}

// removeModifierIfPresent removes the resistance component if entity had one.
func (s *WeatherFactionResistanceSystem) removeModifierIfPresent(entity *Entity) {
	if _, tracked := s.activeModifiers[entity.ID]; !tracked {
		return
	}
	comp, exists := entity.GetComponent("weather_faction_resistance")
	if !exists {
		delete(s.activeModifiers, entity.ID)
		return
	}
	resistance := comp.(*WeatherFactionResistanceComponent)
	resistance.Active = false
	resistance.DamageModifier = 1.0
	resistance.DefenseModifier = 1.0
	resistance.StatusResistModifier = 1.0
	delete(s.activeModifiers, entity.ID)
}

// clearAllModifiers removes resistance from all tracked entities.
func (s *WeatherFactionResistanceSystem) clearAllModifiers(entities []*Entity) {
	for _, entity := range entities {
		s.removeModifierIfPresent(entity)
	}
}

// toLowerSimple converts string to lowercase without importing strings package.
func toLowerSimple(s string) string {
	result := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			result[i] = c + 32
		} else {
			result[i] = c
		}
	}
	return string(result)
}

// containsSimple checks if haystack contains needle without importing strings.
func containsSimple(haystack, needle string) bool {
	if len(needle) == 0 {
		return true
	}
	if len(needle) > len(haystack) {
		return false
	}
	for i := 0; i <= len(haystack)-len(needle); i++ {
		match := true
		for j := 0; j < len(needle); j++ {
			if haystack[i+j] != needle[j] {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}

// GetDamageModifier returns the weather-faction damage modifier for an entity.
// Returns 1.0 if no modifier is active.
func (s *WeatherFactionResistanceSystem) GetDamageModifier(entity *Entity) float64 {
	comp, exists := entity.GetComponent("weather_faction_resistance")
	if !exists {
		return 1.0
	}
	resistance, ok := comp.(*WeatherFactionResistanceComponent)
	if !ok || !resistance.Active {
		return 1.0
	}
	return resistance.DamageModifier
}

// GetDefenseModifier returns the weather-faction defense modifier for an entity.
// Returns 1.0 if no modifier is active.
func (s *WeatherFactionResistanceSystem) GetDefenseModifier(entity *Entity) float64 {
	comp, exists := entity.GetComponent("weather_faction_resistance")
	if !exists {
		return 1.0
	}
	resistance, ok := comp.(*WeatherFactionResistanceComponent)
	if !ok || !resistance.Active {
		return 1.0
	}
	return resistance.DefenseModifier
}

// GetStatusResistModifier returns the weather-faction status resist modifier.
// Returns 1.0 if no modifier is active.
func (s *WeatherFactionResistanceSystem) GetStatusResistModifier(entity *Entity) float64 {
	comp, exists := entity.GetComponent("weather_faction_resistance")
	if !exists {
		return 1.0
	}
	resistance, ok := comp.(*WeatherFactionResistanceComponent)
	if !ok || !resistance.Active {
		return 1.0
	}
	return resistance.StatusResistModifier
}
