// Package engine provides WeatherEquipmentDurabilitySystem which applies
// equipment durability damage based on active weather conditions.
// This bridges WeatherComponent with equipment degradation for environmental wear.
package engine

import (
	"math/rand"

	"github.com/opd-ai/venture/pkg/rendering/particles"
	"github.com/sirupsen/logrus"
)

// WeatherEquipmentDurabilitySystem degrades equipment durability when entities
// are exposed to damaging weather conditions. This connects weather effects with
// the equipment system to enable environmental wear.
//
// Weather-to-durability damage mappings (damage per second at Heavy intensity):
//   - Rain/NeonRain → 0.3 (weapons rust slowly)
//   - BloodRain → 0.6 (corrosive blood damages all equipment)
//   - Snow → 0.2 (cold contracts metal, loosens fittings)
//   - Sandstorm → 0.8 (abrasive sand scours surfaces)
//   - Radiation → 1.0 (radiation degrades materials)
//   - Smog → 0.4 (acidic smog corrodes metal)
//   - Ash → 0.25 (volcanic ash abrades surfaces)
//
// Genre-specific multipliers adjust damage rates:
//   - Fantasy: 1.0 (baseline)
//   - Scifi: 0.5 (advanced materials resist environmental damage)
//   - Horror: 1.3 (cursed environments accelerate decay)
//   - Cyberpunk: 0.7 (nano-coatings provide protection)
//   - PostApoc: 1.4 (harsh wasteland accelerates wear)
type WeatherEquipmentDurabilitySystem struct {
	world   *World
	rng     *rand.Rand
	logger  *logrus.Entry
	genreID string

	// updateInterval controls how often we apply damage (in seconds)
	updateInterval float64
	timeSinceCheck float64

	// Genre-specific damage multipliers
	genreMultipliers map[string]float64

	// Weather-specific base damage rates (damage per second at Heavy intensity)
	weatherDamageRates map[particles.WeatherType]float64

	// Slot-specific damage multipliers per weather type
	slotMultipliers map[particles.WeatherType]map[EquipmentSlot]float64

	// Intensity multipliers
	intensityMultipliers map[particles.WeatherIntensity]float64

	// Cache last known weather state to detect changes
	lastWeatherType   particles.WeatherType
	lastIntensity     particles.WeatherIntensity
	lastWeatherActive bool
}

// NewWeatherEquipmentDurabilitySystem creates a new weather equipment durability system.
func NewWeatherEquipmentDurabilitySystem(world *World, seed int64) *WeatherEquipmentDurabilitySystem {
	logger := logrus.WithFields(logrus.Fields{
		"system_name": "weather_equipment_durability",
		"seed":        seed,
	})
	logger.Debug("Creating weather equipment durability system")

	s := &WeatherEquipmentDurabilitySystem{
		world:          world,
		rng:            rand.New(rand.NewSource(seed)),
		logger:         logger,
		genreID:        "fantasy",
		updateInterval: 1.0, // Apply damage once per second
		genreMultipliers: map[string]float64{
			"fantasy":   1.0,
			"scifi":     0.5,
			"horror":    1.3,
			"cyberpunk": 0.7,
			"postapoc":  1.4,
		},
		weatherDamageRates: map[particles.WeatherType]float64{
			particles.WeatherRain:      0.3,
			particles.WeatherNeonRain:  0.3,
			particles.WeatherBloodRain: 0.6,
			particles.WeatherSnow:      0.2,
			particles.WeatherSandstorm: 0.8,
			particles.WeatherRadiation: 1.0,
			particles.WeatherSmog:      0.4,
			particles.WeatherAsh:       0.25,
			particles.WeatherDust:      0.15,
			particles.WeatherFog:       0.0, // Fog doesn't damage equipment
		},
		slotMultipliers: make(map[particles.WeatherType]map[EquipmentSlot]float64),
		intensityMultipliers: map[particles.WeatherIntensity]float64{
			particles.IntensityLight:   0.25,
			particles.IntensityMedium:  0.5,
			particles.IntensityHeavy:   1.0,
			particles.IntensityExtreme: 1.5,
		},
	}

	// Set up slot-specific multipliers
	// Rain damages weapons more (rust)
	s.slotMultipliers[particles.WeatherRain] = map[EquipmentSlot]float64{
		SlotMainHand: 2.0,
		SlotOffHand:  2.0,
		SlotChest:    0.8, // Armor somewhat protected
	}
	s.slotMultipliers[particles.WeatherNeonRain] = s.slotMultipliers[particles.WeatherRain]

	// Blood rain damages everything equally (corrosive)
	s.slotMultipliers[particles.WeatherBloodRain] = map[EquipmentSlot]float64{
		SlotMainHand: 1.2,
		SlotOffHand:  1.2,
		SlotHead:     1.3,
		SlotChest:    1.0,
		SlotGloves:   1.5, // Exposed hands
	}

	// Snow affects joints and exposed parts
	s.slotMultipliers[particles.WeatherSnow] = map[EquipmentSlot]float64{
		SlotBoots:  1.5, // Feet in snow
		SlotGloves: 1.3,
		SlotLegs:   1.2,
	}

	// Sandstorm scours everything
	s.slotMultipliers[particles.WeatherSandstorm] = map[EquipmentSlot]float64{
		SlotHead:   1.5, // Face protection wears
		SlotChest:  1.3,
		SlotGloves: 1.2,
	}

	// Radiation damages all equipment uniformly
	s.slotMultipliers[particles.WeatherRadiation] = map[EquipmentSlot]float64{}

	return s
}

// SetGenre sets the genre for genre-specific damage modifiers.
func (s *WeatherEquipmentDurabilitySystem) SetGenre(genreID string) {
	s.genreID = genreID
	if s.logger != nil {
		s.logger.WithField("genre", genreID).Debug("Genre set for weather equipment durability")
	}
}

// Update processes all entities and applies equipment durability damage
// based on active weather conditions.
func (s *WeatherEquipmentDurabilitySystem) Update(entities []*Entity, deltaTime float64) {
	s.timeSinceCheck += deltaTime
	if s.timeSinceCheck < s.updateInterval {
		return
	}
	s.timeSinceCheck = 0

	// Find active weather
	weatherType, intensity, active := s.findActiveWeather(entities)

	if !active {
		s.lastWeatherActive = false
		return
	}

	// Check if weather has damage associated
	baseDamage, hasDamage := s.weatherDamageRates[weatherType]
	if !hasDamage || baseDamage <= 0 {
		return
	}

	// Cache weather state
	s.lastWeatherType = weatherType
	s.lastIntensity = intensity
	s.lastWeatherActive = true

	// Process entities with equipment
	for _, entity := range entities {
		s.processEntity(entity, weatherType, intensity, baseDamage)
	}
}

// findActiveWeather searches for active weather in the world.
func (s *WeatherEquipmentDurabilitySystem) findActiveWeather(entities []*Entity) (particles.WeatherType, particles.WeatherIntensity, bool) {
	for _, entity := range entities {
		if !entity.HasComponent("weather") {
			continue
		}

		comp, _ := entity.GetComponent("weather")
		weatherComp, ok := comp.(*WeatherComponent)
		if !ok || weatherComp == nil || !weatherComp.Active {
			continue
		}

		return weatherComp.Config.Type, weatherComp.Config.Intensity, true
	}

	return particles.WeatherRain, particles.IntensityLight, false
}

// processEntity applies weather-based durability damage to a single entity.
func (s *WeatherEquipmentDurabilitySystem) processEntity(entity *Entity, weatherType particles.WeatherType, intensity particles.WeatherIntensity, baseDamage float64) {
	// Get equipment component
	equipComp := s.getEquipmentComponent(entity)
	if equipComp == nil {
		return
	}

	// Skip entities in shelter (would need shelter component)
	// For now, process all entities with equipment

	// Apply genre multiplier
	genreMult := s.getGenreMultiplier()

	// Apply intensity multiplier
	intensityMult := s.getIntensityMultiplier(intensity)

	// Calculate effective damage per second
	effectiveDamage := baseDamage * genreMult * intensityMult * s.updateInterval

	// Apply damage to all equipped items
	visualDirty := false
	for slot, item := range equipComp.Slots {
		if item == nil {
			continue
		}

		// Get slot-specific multiplier
		slotMult := s.getSlotMultiplier(weatherType, slot)
		damage := effectiveDamage * slotMult

		// Apply damage to item durability
		if item.Stats.DurabilityMax > 0 {
			oldDurability := item.Stats.Durability
			item.Stats.Durability -= int(damage)
			if item.Stats.Durability < 0 {
				item.Stats.Durability = 0
			}

			// Check if damage state changed (for visual update)
			if s.damageStateChanged(oldDurability, item.Stats.Durability, item.Stats.DurabilityMax) {
				visualDirty = true
				if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
					s.logger.WithFields(logrus.Fields{
						"entity_id":      entity.ID,
						"slot":           slot.String(),
						"item_id":        item.ID,
						"old_durability": oldDurability,
						"new_durability": item.Stats.Durability,
						"weather":        weatherType.String(),
						"intensity":      intensity.String(),
					}).Debug("equipment durability degraded by weather")
				}
			}
		}
	}

	// Mark equipment visual component dirty if needed
	if visualDirty {
		s.markVisualDirty(entity)
		equipComp.StatsDirty = true
	}
}

// getEquipmentComponent retrieves the equipment component from an entity.
func (s *WeatherEquipmentDurabilitySystem) getEquipmentComponent(entity *Entity) *EquipmentComponent {
	comp, ok := entity.GetComponent("equipment")
	if !ok || comp == nil {
		return nil
	}
	equipComp, ok := comp.(*EquipmentComponent)
	if !ok {
		return nil
	}
	return equipComp
}

// getGenreMultiplier returns the damage multiplier for the current genre.
func (s *WeatherEquipmentDurabilitySystem) getGenreMultiplier() float64 {
	if mult, ok := s.genreMultipliers[s.genreID]; ok {
		return mult
	}
	return 1.0
}

// getIntensityMultiplier returns the damage multiplier for weather intensity.
func (s *WeatherEquipmentDurabilitySystem) getIntensityMultiplier(intensity particles.WeatherIntensity) float64 {
	if mult, ok := s.intensityMultipliers[intensity]; ok {
		return mult
	}
	return 1.0
}

// getSlotMultiplier returns the damage multiplier for a specific slot in weather.
func (s *WeatherEquipmentDurabilitySystem) getSlotMultiplier(weatherType particles.WeatherType, slot EquipmentSlot) float64 {
	if slotMults, ok := s.slotMultipliers[weatherType]; ok {
		if mult, ok := slotMults[slot]; ok {
			return mult
		}
	}
	return 1.0
}

// damageStateChanged checks if the durability crossed a damage state threshold.
// Thresholds: Pristine (100-76%), Worn (75-51%), Damaged (50-26%), Broken (25-0%)
func (s *WeatherEquipmentDurabilitySystem) damageStateChanged(oldDur, newDur, maxDur int) bool {
	if maxDur == 0 {
		return false
	}

	oldPct := float64(oldDur) / float64(maxDur)
	newPct := float64(newDur) / float64(maxDur)

	// Check threshold crossings
	thresholds := []float64{0.75, 0.50, 0.25}
	for _, t := range thresholds {
		if oldPct > t && newPct <= t {
			return true
		}
	}
	return false
}

// markVisualDirty marks the equipment visual component as dirty for regeneration.
func (s *WeatherEquipmentDurabilitySystem) markVisualDirty(entity *Entity) {
	comp, ok := entity.GetComponent("equipment_visual")
	if !ok || comp == nil {
		return
	}
	visualComp, ok := comp.(*EquipmentVisualComponent)
	if !ok {
		return
	}
	visualComp.MarkDirty()
}
