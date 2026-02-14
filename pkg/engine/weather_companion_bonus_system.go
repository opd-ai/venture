// Package engine provides the WeatherCompanionBonusSystem which bridges
// weather conditions with companion combat statistics based on companion type.
// This creates tactical depth where weather affects companion effectiveness
// based on their nature (e.g., water elementals gain bonuses in rain).
package engine

import (
	"math/rand"

	"github.com/opd-ai/venture/pkg/rendering/particles"
	"github.com/sirupsen/logrus"
)

// WeatherCompanionBonusSystem modifies companion stats based on weather conditions
// and companion type. Different companion types benefit from different weather:
//   - Elemental companions: bonuses in matching element weather (water in rain, etc.)
//   - Spirit companions: bonuses in fog/mystical weather
//   - Robot companions: penalties in rain/wet weather, bonuses in clear weather
//   - Undead companions: bonuses in fog/ash/dark weather
//   - Insect companions: penalties in heavy rain, bonuses in mild conditions
//
// Genre-specific modifiers adjust bonus strength.
type WeatherCompanionBonusSystem struct {
	world  *World
	rng    *rand.Rand
	logger *logrus.Entry
	genre  string

	// Update throttling
	updateInterval float64
	timeSinceCheck float64

	// Cache for weather bonuses to avoid recalculating each frame
	bonusCache map[uint64]*WeatherCompanionBonusComponent

	// Cache for last known weather state
	lastWeatherType      particles.WeatherType
	lastWeatherIntensity particles.WeatherIntensity
	lastWeatherActive    bool

	// Genre multipliers for bonus scaling
	genreMultipliers map[string]float64

	// Weather type to companion type bonus mappings
	weatherBonuses map[particles.WeatherType]map[CompanionType]weatherCompanionBonus
}

// weatherCompanionBonus defines stat bonuses for a weather/companion combination
type weatherCompanionBonus struct {
	attackMult  float64 // Multiplier (1.0 = no change, 1.2 = +20%)
	defenseMult float64
	speedMult   float64
}

// WeatherCompanionBonusComponent stores weather-based combat modifiers for companions.
// This is a transient component recalculated when weather changes.
type WeatherCompanionBonusComponent struct {
	// AttackBonus multiplier (1.0 = no bonus, 1.2 = +20%)
	AttackBonus float64

	// DefenseBonus multiplier (1.0 = no bonus, 0.8 = -20%)
	DefenseBonus float64

	// SpeedBonus multiplier (1.0 = no bonus, 1.15 = +15%)
	SpeedBonus float64

	// WeatherType is the current weather providing the bonus
	WeatherType string

	// CompanionTypeName is the companion type receiving the bonus
	CompanionTypeName string
}

// Type returns the component type identifier.
func (c *WeatherCompanionBonusComponent) Type() string {
	return "weather_companion_bonus"
}

// NewWeatherCompanionBonusSystem creates a new weather companion bonus system.
func NewWeatherCompanionBonusSystem(world *World, seed int64) *WeatherCompanionBonusSystem {
	var logEntry *logrus.Entry
	if world != nil && world.logger != nil {
		logEntry = world.logger.Logger.WithField("system_name", "weather_companion_bonus")
	}

	if logEntry != nil && logEntry.Logger.GetLevel() >= logrus.DebugLevel {
		logEntry.Debug("WeatherCompanionBonusSystem created")
	}

	s := &WeatherCompanionBonusSystem{
		world:             world,
		rng:               rand.New(rand.NewSource(seed)),
		logger:            logEntry,
		updateInterval:    0.5, // Check twice per second
		bonusCache:        make(map[uint64]*WeatherCompanionBonusComponent, 32),
		lastWeatherType:   particles.WeatherRain,
		lastWeatherActive: false,
		genreMultipliers: map[string]float64{
			"fantasy":   1.0, // Standard elemental effects
			"scifi":     0.7, // Less magical weather effects
			"horror":    1.3, // Heightened atmospheric influence
			"cyberpunk": 0.5, // Urban environments diminish natural effects
			"postapoc":  1.1, // Harsh weather has stronger effects
		},
	}

	s.initWeatherBonuses()
	return s
}

// initWeatherBonuses initializes the weather-to-companion bonus mappings.
func (s *WeatherCompanionBonusSystem) initWeatherBonuses() {
	s.weatherBonuses = make(map[particles.WeatherType]map[CompanionType]weatherCompanionBonus)

	// Rain bonuses - water elementals thrive, robots struggle
	s.weatherBonuses[particles.WeatherRain] = map[CompanionType]weatherCompanionBonus{
		CompanionTypeElemental: {attackMult: 1.25, defenseMult: 1.20, speedMult: 1.10}, // Water element boost
		CompanionTypeSpirit:    {attackMult: 1.10, defenseMult: 1.05, speedMult: 1.00}, // Spirits neutral+
		CompanionTypeUndead:    {attackMult: 1.05, defenseMult: 1.10, speedMult: 0.95}, // Undead slightly empowered
		CompanionTypeRobot:     {attackMult: 0.85, defenseMult: 0.90, speedMult: 0.90}, // Robots hampered by water
		CompanionTypeInsect:    {attackMult: 0.90, defenseMult: 0.85, speedMult: 0.80}, // Insects struggle in rain
		CompanionTypePet:       {attackMult: 0.95, defenseMult: 1.00, speedMult: 0.90}, // Pets slightly slower
	}

	// Snow bonuses - ice affinity, movement penalties
	s.weatherBonuses[particles.WeatherSnow] = map[CompanionType]weatherCompanionBonus{
		CompanionTypeElemental: {attackMult: 1.20, defenseMult: 1.25, speedMult: 0.95}, // Ice element boost
		CompanionTypeSpirit:    {attackMult: 1.15, defenseMult: 1.10, speedMult: 1.00}, // Spirits empowered
		CompanionTypeUndead:    {attackMult: 1.10, defenseMult: 1.15, speedMult: 0.90}, // Undead preserved by cold
		CompanionTypeRobot:     {attackMult: 0.90, defenseMult: 0.95, speedMult: 0.85}, // Cold affects circuits
		CompanionTypeInsect:    {attackMult: 0.75, defenseMult: 0.70, speedMult: 0.60}, // Insects torpid in cold
		CompanionTypePet:       {attackMult: 0.90, defenseMult: 0.95, speedMult: 0.85}, // Pets slowed by cold
	}

	// Fog bonuses - spirits thrive, visibility penalties
	s.weatherBonuses[particles.WeatherFog] = map[CompanionType]weatherCompanionBonus{
		CompanionTypeSpirit:    {attackMult: 1.30, defenseMult: 1.25, speedMult: 1.15}, // Spirits thrive in mist
		CompanionTypeUndead:    {attackMult: 1.20, defenseMult: 1.15, speedMult: 1.10}, // Undead empowered
		CompanionTypeElemental: {attackMult: 1.05, defenseMult: 1.00, speedMult: 1.00}, // Neutral
		CompanionTypePet:       {attackMult: 0.90, defenseMult: 0.90, speedMult: 0.90}, // Pets disoriented
		CompanionTypeRobot:     {attackMult: 0.85, defenseMult: 0.90, speedMult: 0.95}, // Sensors impaired
		CompanionTypeInsect:    {attackMult: 0.95, defenseMult: 1.00, speedMult: 0.90}, // Slightly impaired
	}

	// Dust/Sandstorm bonuses - harsh conditions, cover
	s.weatherBonuses[particles.WeatherDust] = map[CompanionType]weatherCompanionBonus{
		CompanionTypeRobot:     {attackMult: 0.80, defenseMult: 0.85, speedMult: 0.80}, // Grit damages robots
		CompanionTypeInsect:    {attackMult: 1.15, defenseMult: 1.20, speedMult: 1.10}, // Some insects thrive
		CompanionTypeElemental: {attackMult: 1.10, defenseMult: 1.05, speedMult: 1.00}, // Earth affinity
		CompanionTypePet:       {attackMult: 0.85, defenseMult: 0.90, speedMult: 0.85}, // Pets struggle
		CompanionTypeSpirit:    {attackMult: 0.95, defenseMult: 0.95, speedMult: 1.00}, // Neutral
		CompanionTypeUndead:    {attackMult: 1.00, defenseMult: 1.05, speedMult: 0.95}, // Slightly protected
	}

	// Ash bonuses - dark/fire affinity
	s.weatherBonuses[particles.WeatherAsh] = map[CompanionType]weatherCompanionBonus{
		CompanionTypeUndead:    {attackMult: 1.25, defenseMult: 1.20, speedMult: 1.05}, // Death/destruction affinity
		CompanionTypeElemental: {attackMult: 1.15, defenseMult: 1.10, speedMult: 1.00}, // Fire element boost
		CompanionTypeSpirit:    {attackMult: 1.10, defenseMult: 1.05, speedMult: 1.00}, // Dark spirits empowered
		CompanionTypeRobot:     {attackMult: 0.80, defenseMult: 0.85, speedMult: 0.85}, // Ash clogs mechanisms
		CompanionTypePet:       {attackMult: 0.75, defenseMult: 0.80, speedMult: 0.80}, // Pets suffer
		CompanionTypeInsect:    {attackMult: 0.85, defenseMult: 0.85, speedMult: 0.80}, // Most insects struggle
	}

	// Neon Rain (Cyberpunk) - tech boost
	s.weatherBonuses[particles.WeatherNeonRain] = map[CompanionType]weatherCompanionBonus{
		CompanionTypeRobot:     {attackMult: 1.20, defenseMult: 1.15, speedMult: 1.10}, // Tech synergy
		CompanionTypeElemental: {attackMult: 1.15, defenseMult: 1.10, speedMult: 1.05}, // Energy affinity
		CompanionTypeSpirit:    {attackMult: 0.90, defenseMult: 0.95, speedMult: 1.00}, // Unnatural environment
		CompanionTypeUndead:    {attackMult: 0.95, defenseMult: 1.00, speedMult: 1.00}, // Neutral
		CompanionTypePet:       {attackMult: 0.90, defenseMult: 0.90, speedMult: 0.95}, // Disoriented
		CompanionTypeInsect:    {attackMult: 0.85, defenseMult: 0.90, speedMult: 0.90}, // Confused by lights
	}

	// Smog - oppressive conditions
	s.weatherBonuses[particles.WeatherSmog] = map[CompanionType]weatherCompanionBonus{
		CompanionTypeRobot:     {attackMult: 0.90, defenseMult: 0.95, speedMult: 0.95}, // Filters strained
		CompanionTypeUndead:    {attackMult: 1.15, defenseMult: 1.10, speedMult: 1.00}, // Thrives in decay
		CompanionTypeSpirit:    {attackMult: 1.05, defenseMult: 1.00, speedMult: 1.00}, // Neutral
		CompanionTypeElemental: {attackMult: 0.95, defenseMult: 0.95, speedMult: 0.95}, // Polluted elements
		CompanionTypePet:       {attackMult: 0.80, defenseMult: 0.85, speedMult: 0.85}, // Breathing difficulties
		CompanionTypeInsect:    {attackMult: 0.90, defenseMult: 0.90, speedMult: 0.90}, // Affected
	}

	// Radiation - mutation/chaos
	s.weatherBonuses[particles.WeatherRadiation] = map[CompanionType]weatherCompanionBonus{
		CompanionTypeUndead:    {attackMult: 1.30, defenseMult: 1.25, speedMult: 1.05}, // Radiation empowers undead
		CompanionTypeRobot:     {attackMult: 0.75, defenseMult: 0.80, speedMult: 0.85}, // Electronics damaged
		CompanionTypeElemental: {attackMult: 1.20, defenseMult: 1.15, speedMult: 1.00}, // Energy absorption
		CompanionTypeSpirit:    {attackMult: 1.10, defenseMult: 1.05, speedMult: 1.05}, // Chaotic empowerment
		CompanionTypePet:       {attackMult: 0.70, defenseMult: 0.75, speedMult: 0.80}, // Severely affected
		CompanionTypeInsect:    {attackMult: 1.10, defenseMult: 1.05, speedMult: 1.00}, // Some insects resistant
	}

	// Blood Rain (Horror) - dark magic
	s.weatherBonuses[particles.WeatherBloodRain] = map[CompanionType]weatherCompanionBonus{
		CompanionTypeUndead:    {attackMult: 1.40, defenseMult: 1.30, speedMult: 1.15}, // Dark ritual empowers
		CompanionTypeSpirit:    {attackMult: 1.25, defenseMult: 1.20, speedMult: 1.10}, // Dark spirits surge
		CompanionTypeElemental: {attackMult: 1.10, defenseMult: 1.05, speedMult: 1.00}, // Blood element
		CompanionTypePet:       {attackMult: 0.75, defenseMult: 0.80, speedMult: 0.80}, // Terrified
		CompanionTypeRobot:     {attackMult: 0.85, defenseMult: 0.90, speedMult: 0.90}, // Corrosive
		CompanionTypeInsect:    {attackMult: 1.00, defenseMult: 1.00, speedMult: 0.95}, // Neutral
	}

	// Sandstorm - harsh desert conditions
	s.weatherBonuses[particles.WeatherSandstorm] = map[CompanionType]weatherCompanionBonus{
		CompanionTypeInsect:    {attackMult: 1.20, defenseMult: 1.25, speedMult: 1.10}, // Desert insects thrive
		CompanionTypeElemental: {attackMult: 1.15, defenseMult: 1.10, speedMult: 0.95}, // Earth element
		CompanionTypeUndead:    {attackMult: 1.05, defenseMult: 1.10, speedMult: 0.90}, // Sand preserves
		CompanionTypeRobot:     {attackMult: 0.70, defenseMult: 0.75, speedMult: 0.75}, // Sand destroys joints
		CompanionTypePet:       {attackMult: 0.75, defenseMult: 0.80, speedMult: 0.75}, // Struggling
		CompanionTypeSpirit:    {attackMult: 0.90, defenseMult: 0.95, speedMult: 0.95}, // Slightly impaired
	}
}

// SetGenre sets the genre for genre-specific weather modifiers.
func (s *WeatherCompanionBonusSystem) SetGenre(genreID string) {
	s.genre = genreID
	if s.logger != nil {
		s.logger.WithField("genre", genreID).Debug("genre set")
	}
}

// Update processes all companion entities and applies weather bonuses.
func (s *WeatherCompanionBonusSystem) Update(entities []*Entity, deltaTime float64) {
	s.timeSinceCheck += deltaTime

	if s.timeSinceCheck < s.updateInterval {
		return
	}
	s.timeSinceCheck = 0

	// Find active weather
	weatherType, intensity, active := s.findActiveWeather(entities)

	// If weather state changed, recalculate all companion bonuses
	if active != s.lastWeatherActive ||
		(active && (weatherType != s.lastWeatherType || intensity != s.lastWeatherIntensity)) {
		s.updateAllCompanionBonuses(entities, weatherType, intensity, active)
		s.lastWeatherActive = active
		s.lastWeatherType = weatherType
		s.lastWeatherIntensity = intensity
	}
}

// findActiveWeather locates active weather from entities.
func (s *WeatherCompanionBonusSystem) findActiveWeather(entities []*Entity) (particles.WeatherType, particles.WeatherIntensity, bool) {
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

	return particles.WeatherRain, particles.IntensityLight, false
}

// updateAllCompanionBonuses recalculates bonuses for all companions when weather changes.
func (s *WeatherCompanionBonusSystem) updateAllCompanionBonuses(entities []*Entity, weatherType particles.WeatherType, intensity particles.WeatherIntensity, active bool) {
	for _, entity := range entities {
		// Only process companions
		compComp, hasCompanion := entity.GetComponent("companion")
		if !hasCompanion {
			continue
		}

		companion, ok := compComp.(*CompanionComponent)
		if !ok {
			continue
		}

		if active {
			s.applyWeatherBonus(entity, companion, weatherType, intensity)
		} else {
			s.removeWeatherBonus(entity)
		}
	}
}

// applyWeatherBonus applies weather-based bonuses to a companion entity.
func (s *WeatherCompanionBonusSystem) applyWeatherBonus(entity *Entity, companion *CompanionComponent, weatherType particles.WeatherType, intensity particles.WeatherIntensity) {
	bonus := s.calculateWeatherBonus(companion.CompanionType, weatherType, intensity)
	if bonus == nil {
		s.removeWeatherBonus(entity)
		return
	}

	// Reverse old bonus before applying new
	if existingComp, ok := entity.GetComponent("weather_companion_bonus"); ok {
		if existingBonus, ok := existingComp.(*WeatherCompanionBonusComponent); ok {
			s.reverseStatBonus(entity, existingBonus)
		}
	}

	// Apply new bonus to companion stats
	s.applyBonusToStats(entity, bonus)

	// Update or add component
	if existing, hasExisting := entity.GetComponent("weather_companion_bonus"); hasExisting {
		if existingBonus, ok := existing.(*WeatherCompanionBonusComponent); ok {
			*existingBonus = *bonus
		}
	} else {
		entity.AddComponent(bonus)
	}

	s.bonusCache[entity.ID] = bonus
	s.logBonusApplication(entity, bonus, companion.CompanionType, weatherType)
}

// removeWeatherBonus removes weather bonus from a companion entity.
func (s *WeatherCompanionBonusSystem) removeWeatherBonus(entity *Entity) {
	if existingComp, ok := entity.GetComponent("weather_companion_bonus"); ok {
		if bonus, ok := existingComp.(*WeatherCompanionBonusComponent); ok {
			s.reverseStatBonus(entity, bonus)
		}
		entity.RemoveComponent("weather_companion_bonus")
	}
	delete(s.bonusCache, entity.ID)
}

// calculateWeatherBonus computes bonuses for a weather/companion type combination.
func (s *WeatherCompanionBonusSystem) calculateWeatherBonus(compType CompanionType, weatherType particles.WeatherType, intensity particles.WeatherIntensity) *WeatherCompanionBonusComponent {
	weatherMap, hasWeather := s.weatherBonuses[weatherType]
	if !hasWeather {
		return nil
	}

	bonusData, hasBonus := weatherMap[compType]
	if !hasBonus {
		return nil
	}

	// Apply genre multiplier
	genreMult := s.genreMultipliers[s.genre]
	if genreMult == 0 {
		genreMult = 1.0
	}

	// Apply intensity modifier
	intensityMod := s.getIntensityModifier(intensity)

	// Scale bonus away from 1.0 by genre and intensity multipliers
	attackBonus := 1.0 + (bonusData.attackMult-1.0)*genreMult*intensityMod
	defenseBonus := 1.0 + (bonusData.defenseMult-1.0)*genreMult*intensityMod
	speedBonus := 1.0 + (bonusData.speedMult-1.0)*genreMult*intensityMod

	return &WeatherCompanionBonusComponent{
		AttackBonus:       attackBonus,
		DefenseBonus:      defenseBonus,
		SpeedBonus:        speedBonus,
		WeatherType:       weatherType.String(),
		CompanionTypeName: s.companionTypeName(compType),
	}
}

// getIntensityModifier returns a multiplier based on weather intensity.
func (s *WeatherCompanionBonusSystem) getIntensityModifier(intensity particles.WeatherIntensity) float64 {
	switch intensity {
	case particles.IntensityLight:
		return 0.5 // Half effect
	case particles.IntensityMedium:
		return 1.0 // Full effect
	case particles.IntensityHeavy:
		return 1.3 // 30% stronger
	case particles.IntensityExtreme:
		return 1.6 // 60% stronger
	default:
		return 1.0
	}
}

// applyBonusToStats applies the weather bonus to companion stats.
func (s *WeatherCompanionBonusSystem) applyBonusToStats(entity *Entity, bonus *WeatherCompanionBonusComponent) {
	statsComp, ok := entity.GetComponent("companionstats")
	if !ok {
		return
	}

	stats, ok := statsComp.(*CompanionStatsComponent)
	if !ok {
		return
	}

	stats.Attack *= bonus.AttackBonus
	stats.Defense *= bonus.DefenseBonus
	stats.Speed *= bonus.SpeedBonus
}

// reverseStatBonus reverses previously applied bonus from companion stats.
func (s *WeatherCompanionBonusSystem) reverseStatBonus(entity *Entity, bonus *WeatherCompanionBonusComponent) {
	statsComp, ok := entity.GetComponent("companionstats")
	if !ok {
		return
	}

	stats, ok := statsComp.(*CompanionStatsComponent)
	if !ok {
		return
	}

	if bonus.AttackBonus != 0 {
		stats.Attack /= bonus.AttackBonus
	}
	if bonus.DefenseBonus != 0 {
		stats.Defense /= bonus.DefenseBonus
	}
	if bonus.SpeedBonus != 0 {
		stats.Speed /= bonus.SpeedBonus
	}
}

// companionTypeName returns a string name for the companion type.
func (s *WeatherCompanionBonusSystem) companionTypeName(compType CompanionType) string {
	switch compType {
	case CompanionTypePet:
		return "Pet"
	case CompanionTypeSummon:
		return "Summon"
	case CompanionTypeHireling:
		return "Hireling"
	case CompanionTypeElemental:
		return "Elemental"
	case CompanionTypeUndead:
		return "Undead"
	case CompanionTypeRobot:
		return "Robot"
	case CompanionTypeSpirit:
		return "Spirit"
	case CompanionTypeInsect:
		return "Insect"
	default:
		return "Unknown"
	}
}

// logBonusApplication logs when a weather bonus is applied.
func (s *WeatherCompanionBonusSystem) logBonusApplication(entity *Entity, bonus *WeatherCompanionBonusComponent, compType CompanionType, weatherType particles.WeatherType) {
	if s.logger == nil || s.logger.Logger.GetLevel() < logrus.DebugLevel {
		return
	}

	s.logger.WithFields(logrus.Fields{
		"entity_id":      entity.ID,
		"weather_type":   weatherType.String(),
		"companion_type": s.companionTypeName(compType),
		"attack_bonus":   bonus.AttackBonus,
		"defense_bonus":  bonus.DefenseBonus,
		"speed_bonus":    bonus.SpeedBonus,
		"genre":          s.genre,
	}).Debug("weather companion bonus applied")
}

// HasActiveBonus returns whether a companion has an active weather bonus.
func (s *WeatherCompanionBonusSystem) HasActiveBonus(companionID uint64) bool {
	_, exists := s.bonusCache[companionID]
	return exists
}

// GetBonusCount returns the number of companions with active weather bonuses.
func (s *WeatherCompanionBonusSystem) GetBonusCount() int {
	return len(s.bonusCache)
}

// IsWeatherActive returns whether weather is currently affecting companions.
func (s *WeatherCompanionBonusSystem) IsWeatherActive() bool {
	return s.lastWeatherActive
}

// GetActiveWeatherType returns the current weather type affecting companions.
func (s *WeatherCompanionBonusSystem) GetActiveWeatherType() particles.WeatherType {
	return s.lastWeatherType
}
