package engine

import (
	"math"
	"math/rand"
	"testing"

	"github.com/opd-ai/venture/pkg/rendering/particles"
)

// TestWeatherFactionResistanceComponent verifies component is pure data.
func TestWeatherFactionResistanceComponent(t *testing.T) {
	comp := &WeatherFactionResistanceComponent{
		FactionID:            "flame_legion",
		DamageModifier:       1.15,
		DefenseModifier:      1.10,
		StatusResistModifier: 1.05,
		Active:               true,
		WeatherType:          "clear",
		AffinityType:         "fire",
	}

	if comp.Type() != "weather_faction_resistance" {
		t.Errorf("expected Type() = 'weather_faction_resistance', got %s", comp.Type())
	}

	// Verify fields are accessible
	if comp.DamageModifier != 1.15 {
		t.Errorf("DamageModifier = %f, want 1.15", comp.DamageModifier)
	}
	if comp.FactionID != "flame_legion" {
		t.Errorf("FactionID = %s, want flame_legion", comp.FactionID)
	}
}

// TestNewWeatherFactionResistanceSystem verifies system creation.
func TestNewWeatherFactionResistanceSystem(t *testing.T) {
	world := &World{}
	sys := NewWeatherFactionResistanceSystem(world, 12345)

	if sys == nil {
		t.Fatal("NewWeatherFactionResistanceSystem returned nil")
	}
	if sys.world != world {
		t.Error("world reference not set")
	}
	if sys.rng == nil {
		t.Error("rng not initialized")
	}
	if len(sys.factionAffinities) == 0 {
		t.Error("factionAffinities not initialized")
	}
	if len(sys.weatherInteractions) == 0 {
		t.Error("weatherInteractions not initialized")
	}
}

// TestFactionAffinityDetection verifies faction ID pattern matching.
func TestFactionAffinityDetection(t *testing.T) {
	world := &World{}
	sys := NewWeatherFactionResistanceSystem(world, 12345)

	tests := []struct {
		factionID     string
		expectedAffin factionAffinityType
	}{
		{"flame_legion", affinityFire},
		{"FIRE_CULT", affinityFire},
		{"ember_knights", affinityFire},
		{"ice_coven", affinityIce},
		{"frost_giants", affinityIce},
		{"frozen_throne", affinityIce},
		{"forest_wardens", affinityNature},
		{"wild_druids", affinityNature},
		{"storm_riders", affinityStorm},
		{"thunder_clan", affinityStorm},
		{"shadow_guild", affinityDark},
		{"void_walkers", affinityDark},
		{"tech_syndicate", affinityTech},
		{"cyber_corp", affinityTech},
		{"generic_faction", affinityNeutral},
		{"merchant_guild", affinityNeutral},
	}

	for _, tt := range tests {
		t.Run(tt.factionID, func(t *testing.T) {
			result := sys.getFactionAffinity(tt.factionID)
			if result != tt.expectedAffin {
				t.Errorf("getFactionAffinity(%s) = %s, want %s", tt.factionID, result, tt.expectedAffin)
			}
		})
	}
}

// TestWeatherInteractions verifies affinity-weather interaction lookups.
func TestWeatherInteractions(t *testing.T) {
	world := &World{}
	sys := NewWeatherFactionResistanceSystem(world, 12345)

	tests := []struct {
		affinity   factionAffinityType
		weather    particles.WeatherType
		expectBuff bool // true if damage multiplier > 1.0
	}{
		// Fire factions
		{affinityFire, particles.WeatherDust, true},  // Fire buffed in dry weather
		{affinityFire, particles.WeatherRain, false}, // Fire weakened in rain
		{affinityFire, particles.WeatherSnow, false}, // Fire weakened in snow

		// Ice factions
		{affinityIce, particles.WeatherSnow, true},  // Ice buffed in snow
		{affinityIce, particles.WeatherDust, false}, // Ice weakened in dry

		// Nature factions
		{affinityNature, particles.WeatherRain, true},       // Nature buffed in rain
		{affinityNature, particles.WeatherSandstorm, false}, // Nature weakened in sandstorm

		// Storm factions
		{affinityStorm, particles.WeatherRain, true},  // Storm buffed in rain
		{affinityStorm, particles.WeatherDust, false}, // Storm weakened in dry

		// Dark factions
		{affinityDark, particles.WeatherFog, true},   // Dark buffed in fog
		{affinityDark, particles.WeatherDust, false}, // Dark weakened in dry/bright

		// Tech factions
		{affinityTech, particles.WeatherDust, true},       // Tech buffed in clear
		{affinityTech, particles.WeatherSandstorm, false}, // Tech weakened in sandstorm
	}

	for _, tt := range tests {
		t.Run(string(tt.affinity)+"_"+tt.weather.String(), func(t *testing.T) {
			interaction := sys.getWeatherInteraction(tt.affinity, tt.weather)
			if tt.expectBuff && interaction.DamageMultiplier <= 1.0 {
				t.Errorf("expected buff (DamageMultiplier > 1.0), got %f", interaction.DamageMultiplier)
			}
			if !tt.expectBuff && interaction.DamageMultiplier >= 1.0 {
				t.Errorf("expected debuff (DamageMultiplier < 1.0), got %f", interaction.DamageMultiplier)
			}
		})
	}
}

// TestIntensityScaling verifies weather intensity affects modifier strength.
func TestIntensityScaling(t *testing.T) {
	world := &World{}
	sys := NewWeatherFactionResistanceSystem(world, 12345)

	tests := []struct {
		intensity     particles.WeatherIntensity
		expectedScale float64
	}{
		{particles.IntensityLight, 0.25},
		{particles.IntensityMedium, 0.50},
		{particles.IntensityHeavy, 0.75},
		{particles.IntensityExtreme, 1.0},
	}

	for _, tt := range tests {
		t.Run(tt.intensity.String(), func(t *testing.T) {
			scale := sys.getIntensityScale(tt.intensity)
			if scale != tt.expectedScale {
				t.Errorf("getIntensityScale(%s) = %f, want %f", tt.intensity.String(), scale, tt.expectedScale)
			}
		})
	}
}

// TestGenreMultipliers verifies genre-based scaling.
func TestGenreMultipliers(t *testing.T) {
	world := &World{}
	sys := NewWeatherFactionResistanceSystem(world, 12345)

	tests := []struct {
		genre    string
		expected float64
	}{
		{"fantasy", 1.0},
		{"scifi", 0.7},
		{"horror", 0.85},
		{"cyberpunk", 0.9},
		{"postapoc", 0.6},
		{"unknown", 1.0}, // Default
	}

	for _, tt := range tests {
		t.Run(tt.genre, func(t *testing.T) {
			sys.SetGenre(tt.genre)
			mult := sys.getGenreMultiplier()
			if mult != tt.expected {
				t.Errorf("getGenreMultiplier() = %f, want %f", mult, tt.expected)
			}
		})
	}
}

// TestBlendModifier verifies modifier blending toward 1.0.
func TestBlendModifier(t *testing.T) {
	world := &World{}
	sys := NewWeatherFactionResistanceSystem(world, 12345)

	tests := []struct {
		baseMod        float64
		genreScale     float64
		intensityScale float64
		expected       float64
	}{
		{1.0, 1.0, 1.0, 1.0},  // No change
		{1.2, 1.0, 1.0, 1.2},  // Full buff at full scales
		{1.2, 1.0, 0.5, 1.1},  // Half intensity = half effect
		{0.8, 1.0, 1.0, 0.8},  // Full debuff at full scales
		{0.8, 1.0, 0.5, 0.9},  // Half intensity = half debuff
		{1.2, 0.5, 1.0, 1.1},  // Genre reduces buff
		{0.8, 0.5, 0.5, 0.95}, // Both scales reduce debuff
	}

	for i, tt := range tests {
		result := sys.blendModifier(tt.baseMod, tt.genreScale, tt.intensityScale)
		if math.Abs(result-tt.expected) > 0.001 {
			t.Errorf("test %d: blendModifier(%f, %f, %f) = %f, want %f",
				i, tt.baseMod, tt.genreScale, tt.intensityScale, result, tt.expected)
		}
	}
}

// TestUpdateWithNoWeather verifies system clears modifiers when no weather active.
func TestUpdateWithNoWeather(t *testing.T) {
	world := &World{}
	sys := NewWeatherFactionResistanceSystem(world, 12345)

	// Create entity with faction
	entity := &Entity{
		ID:         1,
		Components: make(map[string]Component),
	}
	entity.AddComponent(&FactionComponent{FactionID: "flame_legion"})

	// Manually add a resistance component to simulate previous weather
	resistance := &WeatherFactionResistanceComponent{Active: true, DamageModifier: 1.2}
	entity.AddComponent(resistance)
	sys.activeModifiers[entity.ID] = true

	// Update with no weather entity
	entities := []*Entity{entity}
	sys.timeSinceUpdate = sys.updateInterval // Force update
	sys.Update(entities, 0.5)

	// Resistance should be deactivated
	if resistance.Active {
		t.Error("expected resistance to be deactivated when no weather")
	}
	if resistance.DamageModifier != 1.0 {
		t.Errorf("expected DamageModifier reset to 1.0, got %f", resistance.DamageModifier)
	}
}

// TestUpdateAppliesModifiers verifies modifiers are applied correctly.
func TestUpdateAppliesModifiers(t *testing.T) {
	world := &World{}
	sys := NewWeatherFactionResistanceSystem(world, 12345)
	sys.SetGenre("fantasy")

	// Create weather entity
	weatherEntity := &Entity{
		ID:         1,
		Components: make(map[string]Component),
	}
	weatherComp := NewWeatherComponent(particles.WeatherConfig{
		Type:      particles.WeatherRain,
		Intensity: particles.IntensityExtreme,
	})
	weatherComp.Active = true
	weatherEntity.AddComponent(weatherComp)

	// Create fire faction entity
	fireEntity := &Entity{
		ID:         2,
		Components: make(map[string]Component),
	}
	fireEntity.AddComponent(&FactionComponent{FactionID: "flame_legion"})

	// Create ice faction entity
	iceEntity := &Entity{
		ID:         3,
		Components: make(map[string]Component),
	}
	iceEntity.AddComponent(&FactionComponent{FactionID: "frost_giants"})

	entities := []*Entity{weatherEntity, fireEntity, iceEntity}

	// Force update
	sys.timeSinceUpdate = sys.updateInterval
	sys.Update(entities, 0.5)

	// Fire faction should be weakened in rain
	fireDmg := sys.GetDamageModifier(fireEntity)
	if fireDmg >= 1.0 {
		t.Errorf("fire faction should be weakened in rain, got damage modifier %f", fireDmg)
	}

	// Ice faction should be buffed in rain
	iceDmg := sys.GetDamageModifier(iceEntity)
	if iceDmg <= 1.0 {
		t.Errorf("ice faction should be buffed in rain, got damage modifier %f", iceDmg)
	}
}

// TestNeutralFactionNoModifier verifies neutral factions get no modifier.
func TestNeutralFactionNoModifier(t *testing.T) {
	world := &World{}
	sys := NewWeatherFactionResistanceSystem(world, 12345)

	// Create weather entity
	weatherEntity := &Entity{
		ID:         1,
		Components: make(map[string]Component),
	}
	weatherComp := NewWeatherComponent(particles.WeatherConfig{
		Type:      particles.WeatherRain,
		Intensity: particles.IntensityExtreme,
	})
	weatherComp.Active = true
	weatherEntity.AddComponent(weatherComp)

	// Create neutral faction entity
	neutralEntity := &Entity{
		ID:         2,
		Components: make(map[string]Component),
	}
	neutralEntity.AddComponent(&FactionComponent{FactionID: "merchant_guild"})

	entities := []*Entity{weatherEntity, neutralEntity}

	sys.timeSinceUpdate = sys.updateInterval
	sys.Update(entities, 0.5)

	// Neutral faction should have no modifier
	dmgMod := sys.GetDamageModifier(neutralEntity)
	if dmgMod != 1.0 {
		t.Errorf("neutral faction should have damage modifier 1.0, got %f", dmgMod)
	}
}

// TestToLowerSimple verifies lowercase conversion.
func TestToLowerSimple(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"HELLO", "hello"},
		{"Hello", "hello"},
		{"hello", "hello"},
		{"HeLLo123", "hello123"},
		{"", ""},
	}

	for _, tt := range tests {
		result := toLowerSimple(tt.input)
		if result != tt.expected {
			t.Errorf("toLowerSimple(%s) = %s, want %s", tt.input, result, tt.expected)
		}
	}
}

// TestContainsSimple verifies substring check.
func TestContainsSimple(t *testing.T) {
	tests := []struct {
		haystack string
		needle   string
		expected bool
	}{
		{"hello world", "world", true},
		{"hello world", "earth", false},
		{"flame_legion", "flame", true},
		{"frost_giants", "fire", false},
		{"", "", true},
		{"abc", "", true},
		{"", "abc", false},
	}

	for _, tt := range tests {
		result := containsSimple(tt.haystack, tt.needle)
		if result != tt.expected {
			t.Errorf("containsSimple(%s, %s) = %v, want %v", tt.haystack, tt.needle, result, tt.expected)
		}
	}
}

// TestModifierGetters verifies GetXxxModifier functions.
func TestModifierGetters(t *testing.T) {
	world := &World{}
	sys := NewWeatherFactionResistanceSystem(world, 12345)

	entity := &Entity{
		ID:         1,
		Components: make(map[string]Component),
	}

	// No component - should return 1.0
	if sys.GetDamageModifier(entity) != 1.0 {
		t.Error("expected 1.0 with no component")
	}
	if sys.GetDefenseModifier(entity) != 1.0 {
		t.Error("expected 1.0 with no component")
	}
	if sys.GetStatusResistModifier(entity) != 1.0 {
		t.Error("expected 1.0 with no component")
	}

	// Add inactive component
	resistance := &WeatherFactionResistanceComponent{
		Active:               false,
		DamageModifier:       1.5,
		DefenseModifier:      1.4,
		StatusResistModifier: 1.3,
	}
	entity.AddComponent(resistance)

	if sys.GetDamageModifier(entity) != 1.0 {
		t.Error("expected 1.0 with inactive component")
	}

	// Activate component
	resistance.Active = true
	if sys.GetDamageModifier(entity) != 1.5 {
		t.Errorf("expected 1.5, got %f", sys.GetDamageModifier(entity))
	}
	if sys.GetDefenseModifier(entity) != 1.4 {
		t.Errorf("expected 1.4, got %f", sys.GetDefenseModifier(entity))
	}
	if sys.GetStatusResistModifier(entity) != 1.3 {
		t.Errorf("expected 1.3, got %f", sys.GetStatusResistModifier(entity))
	}
}

// BenchmarkFactionAffinityLookup benchmarks affinity detection.
func BenchmarkFactionAffinityLookup(b *testing.B) {
	world := &World{}
	sys := NewWeatherFactionResistanceSystem(world, 12345)
	factionIDs := []string{
		"flame_legion", "frost_giants", "forest_wardens",
		"storm_riders", "shadow_guild", "tech_syndicate",
		"merchant_guild", "generic_faction",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, id := range factionIDs {
			_ = sys.getFactionAffinity(id)
		}
	}
}

// BenchmarkWeatherInteractionLookup benchmarks interaction lookup.
func BenchmarkWeatherInteractionLookup(b *testing.B) {
	world := &World{}
	sys := NewWeatherFactionResistanceSystem(world, 12345)
	affinities := []factionAffinityType{affinityFire, affinityIce, affinityNature, affinityStorm}
	weathers := []particles.WeatherType{particles.WeatherDust, particles.WeatherRain, particles.WeatherSnow}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, a := range affinities {
			for _, w := range weathers {
				_ = sys.getWeatherInteraction(a, w)
			}
		}
	}
}

// BenchmarkUpdate benchmarks full system update.
func BenchmarkUpdate(b *testing.B) {
	world := &World{}
	sys := NewWeatherFactionResistanceSystem(world, 12345)

	// Create weather entity
	weatherEntity := &Entity{ID: 1, Components: make(map[string]Component)}
	weatherComp := NewWeatherComponent(particles.WeatherConfig{
		Type:      particles.WeatherRain,
		Intensity: particles.IntensityMedium,
	})
	weatherComp.Active = true
	weatherEntity.AddComponent(weatherComp)

	// Create faction entities
	rng := rand.New(rand.NewSource(12345))
	factions := []string{"flame_legion", "frost_giants", "forest_wardens", "shadow_guild", "merchant_guild"}
	entities := []*Entity{weatherEntity}
	for i := 0; i < 100; i++ {
		e := &Entity{ID: uint64(i + 2), Components: make(map[string]Component)}
		e.AddComponent(&FactionComponent{FactionID: factions[rng.Intn(len(factions))]})
		entities = append(entities, e)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.timeSinceUpdate = sys.updateInterval
		sys.Update(entities, 0.016)
	}
}
