// Package particles provides weather particle effects.
package particles

import (
	"testing"
)

// TestWeatherType_String tests weather type string conversion.
func TestWeatherType_String(t *testing.T) {
	tests := []struct {
		name     string
		weather  WeatherType
		expected string
	}{
		{"rain", WeatherRain, "Rain"},
		{"snow", WeatherSnow, "Snow"},
		{"fog", WeatherFog, "Fog"},
		{"dust", WeatherDust, "Dust"},
		{"ash", WeatherAsh, "Ash"},
		{"neon_rain", WeatherNeonRain, "NeonRain"},
		{"smog", WeatherSmog, "Smog"},
		{"radiation", WeatherRadiation, "Radiation"},
		{"sandstorm", WeatherSandstorm, "Sandstorm"},
		{"blood_rain", WeatherBloodRain, "BloodRain"},
		{"unknown", WeatherType(99), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.weather.String(); got != tt.expected {
				t.Errorf("WeatherType.String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// TestWeatherIntensity_String tests intensity string conversion.
func TestWeatherIntensity_String(t *testing.T) {
	tests := []struct {
		name      string
		intensity WeatherIntensity
		expected  string
	}{
		{"light", IntensityLight, "Light"},
		{"medium", IntensityMedium, "Medium"},
		{"heavy", IntensityHeavy, "Heavy"},
		{"extreme", IntensityExtreme, "Extreme"},
		{"unknown", WeatherIntensity(99), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.intensity.String(); got != tt.expected {
				t.Errorf("WeatherIntensity.String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// TestDefaultWeatherConfig tests default configuration.
func TestDefaultWeatherConfig(t *testing.T) {
	config := DefaultWeatherConfig()

	if config.Type != WeatherRain {
		t.Errorf("DefaultWeatherConfig Type = %v, want %v", config.Type, WeatherRain)
	}
	if config.Width <= 0 {
		t.Errorf("DefaultWeatherConfig Width = %v, want positive", config.Width)
	}
	if config.Height <= 0 {
		t.Errorf("DefaultWeatherConfig Height = %v, want positive", config.Height)
	}
	if config.GenreID == "" {
		t.Error("DefaultWeatherConfig GenreID is empty")
	}
	if config.Custom == nil {
		t.Error("DefaultWeatherConfig Custom map is nil")
	}
}

// TestWeatherConfig_Validate tests configuration validation.
func TestWeatherConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  WeatherConfig
		wantErr bool
	}{
		{
			name: "valid config",
			config: WeatherConfig{
				Type:      WeatherRain,
				Intensity: IntensityMedium,
				Width:     800,
				Height:    600,
				GenreID:   "fantasy",
			},
			wantErr: false,
		},
		{
			name: "zero width",
			config: WeatherConfig{
				Type:      WeatherRain,
				Intensity: IntensityMedium,
				Width:     0,
				Height:    600,
				GenreID:   "fantasy",
			},
			wantErr: true,
		},
		{
			name: "zero height",
			config: WeatherConfig{
				Type:      WeatherRain,
				Intensity: IntensityMedium,
				Width:     800,
				Height:    0,
				GenreID:   "fantasy",
			},
			wantErr: true,
		},
		{
			name: "empty genre",
			config: WeatherConfig{
				Type:      WeatherRain,
				Intensity: IntensityMedium,
				Width:     800,
				Height:    600,
				GenreID:   "",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("WeatherConfig.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestWeatherConfig_GetParticleCount tests particle count calculation.
func TestWeatherConfig_GetParticleCount(t *testing.T) {
	tests := []struct {
		name      string
		config    WeatherConfig
		wantRange [2]int // min, max expected range
	}{
		{
			name: "light intensity",
			config: WeatherConfig{
				Type:      WeatherRain,
				Intensity: IntensityLight,
				Width:     1000,
				Height:    1000,
			},
			wantRange: [2]int{1000, 3000},
		},
		{
			name: "medium intensity",
			config: WeatherConfig{
				Type:      WeatherRain,
				Intensity: IntensityMedium,
				Width:     1000,
				Height:    1000,
			},
			wantRange: [2]int{3000, 7000},
		},
		{
			name: "heavy intensity",
			config: WeatherConfig{
				Type:      WeatherRain,
				Intensity: IntensityHeavy,
				Width:     1000,
				Height:    1000,
			},
			wantRange: [2]int{7000, 15000},
		},
		{
			name: "small area",
			config: WeatherConfig{
				Type:      WeatherRain,
				Intensity: IntensityMedium,
				Width:     100,
				Height:    100,
			},
			wantRange: [2]int{10, 100},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			count := tt.config.GetParticleCount()
			if count < tt.wantRange[0] || count > tt.wantRange[1] {
				t.Errorf("GetParticleCount() = %v, want range %v-%v", count, tt.wantRange[0], tt.wantRange[1])
			}
		})
	}
}

// TestGenerateWeather tests weather system generation.
func TestGenerateWeather(t *testing.T) {
	config := WeatherConfig{
		Type:      WeatherRain,
		Intensity: IntensityMedium,
		Width:     800,
		Height:    600,
		GenreID:   "fantasy",
		Seed:      12345,
	}

	ws, err := GenerateWeather(config)
	if err != nil {
		t.Fatalf("GenerateWeather failed: %v", err)
	}

	if ws == nil {
		t.Fatal("GenerateWeather returned nil")
	}

	if len(ws.Particles) == 0 {
		t.Error("GenerateWeather created no particles")
	}

	// Verify particles are within bounds
	for i, p := range ws.Particles {
		if p.X < 0 || p.X > float64(config.Width) {
			t.Errorf("Particle %d X position out of bounds: %v", i, p.X)
		}
		if p.Y < 0 || p.Y > float64(config.Height) {
			t.Errorf("Particle %d Y position out of bounds: %v", i, p.Y)
		}
		if p.Size <= 0 {
			t.Errorf("Particle %d has invalid size: %v", i, p.Size)
		}
	}
}

// TestGenerateWeather_AllTypes tests generation of all weather types.
func TestGenerateWeather_AllTypes(t *testing.T) {
	weatherTypes := []WeatherType{
		WeatherRain,
		WeatherSnow,
		WeatherFog,
		WeatherDust,
		WeatherAsh,
		WeatherNeonRain,
		WeatherSmog,
		WeatherRadiation,
		WeatherSandstorm,
		WeatherBloodRain,
	}

	for _, weatherType := range weatherTypes {
		t.Run(weatherType.String(), func(t *testing.T) {
			config := WeatherConfig{
				Type:      weatherType,
				Intensity: IntensityMedium,
				Width:     800,
				Height:    600,
				GenreID:   "fantasy",
				Seed:      12345,
			}

			ws, err := GenerateWeather(config)
			if err != nil {
				t.Errorf("GenerateWeather(%v) failed: %v", weatherType, err)
				return
			}

			if ws == nil {
				t.Errorf("GenerateWeather(%v) returned nil", weatherType)
				return
			}

			if len(ws.Particles) == 0 {
				t.Errorf("GenerateWeather(%v) created no particles", weatherType)
			}

			// Verify weather effects are initialized
			if ws.Effects == nil {
				t.Errorf("GenerateWeather(%v) has nil Effects", weatherType)
			}
		})
	}
}

// TestGenerateWeather_InvalidConfig tests error handling.
func TestGenerateWeather_InvalidConfig(t *testing.T) {
	config := WeatherConfig{
		Type:      WeatherRain,
		Intensity: IntensityMedium,
		Width:     0, // Invalid
		Height:    600,
		GenreID:   "fantasy",
		Seed:      12345,
	}

	ws, err := GenerateWeather(config)
	if err == nil {
		t.Error("GenerateWeather accepted invalid config")
	}
	if ws != nil {
		t.Error("GenerateWeather returned non-nil with error")
	}
}

// TestGenerateWeather_Determinism tests deterministic generation.
func TestGenerateWeather_Determinism(t *testing.T) {
	config := WeatherConfig{
		Type:      WeatherRain,
		Intensity: IntensityMedium,
		Width:     800,
		Height:    600,
		GenreID:   "fantasy",
		Seed:      12345,
	}

	ws1, err1 := GenerateWeather(config)
	if err1 != nil {
		t.Fatalf("First generation failed: %v", err1)
	}

	ws2, err2 := GenerateWeather(config)
	if err2 != nil {
		t.Fatalf("Second generation failed: %v", err2)
	}

	if len(ws1.Particles) != len(ws2.Particles) {
		t.Errorf("Particle counts differ: %d vs %d", len(ws1.Particles), len(ws2.Particles))
	}

	// Check that particles match
	for i := 0; i < min(len(ws1.Particles), len(ws2.Particles)); i++ {
		p1 := ws1.Particles[i]
		p2 := ws2.Particles[i]

		if p1.X != p2.X || p1.Y != p2.Y {
			t.Errorf("Particle %d position differs: (%v,%v) vs (%v,%v)", i, p1.X, p1.Y, p2.X, p2.Y)
		}
		if p1.VX != p2.VX || p1.VY != p2.VY {
			t.Errorf("Particle %d velocity differs", i)
		}
	}
}

// TestWeatherSystem_Update tests system updates.
func TestWeatherSystem_Update(t *testing.T) {
	config := WeatherConfig{
		Type:      WeatherRain,
		Intensity: IntensityLight,
		Width:     800,
		Height:    600,
		GenreID:   "fantasy",
		Seed:      12345,
	}

	ws, err := GenerateWeather(config)
	if err != nil {
		t.Fatalf("GenerateWeather failed: %v", err)
	}

	// Store initial positions
	initialPositions := make([]struct{ X, Y float64 }, len(ws.Particles))
	for i, p := range ws.Particles {
		initialPositions[i].X = p.X
		initialPositions[i].Y = p.Y
	}

	// Update system
	deltaTime := 0.016 // ~60 FPS
	ws.Update(deltaTime)

	// Check that particles moved
	movedCount := 0
	for i, p := range ws.Particles {
		if p.X != initialPositions[i].X || p.Y != initialPositions[i].Y {
			movedCount++
		}
	}

	if movedCount == 0 {
		t.Error("No particles moved after update")
	}

	// Check elapsed time
	if ws.ElapsedTime != deltaTime {
		t.Errorf("ElapsedTime = %v, want %v", ws.ElapsedTime, deltaTime)
	}
}

// TestWeatherSystem_Update_Wrapping tests particle wrapping.
func TestWeatherSystem_Update_Wrapping(t *testing.T) {
	config := WeatherConfig{
		Type:      WeatherRain,
		Intensity: IntensityLight,
		Width:     100,
		Height:    100,
		GenreID:   "fantasy",
		Seed:      12345,
	}

	ws, err := GenerateWeather(config)
	if err != nil {
		t.Fatalf("GenerateWeather failed: %v", err)
	}

	// Force a particle to go out of bounds
	if len(ws.Particles) > 0 {
		ws.Particles[0].Y = float64(config.Height) + 10
	}

	// Update
	ws.Update(0.016)

	// Check that particle wrapped
	if len(ws.Particles) > 0 {
		if ws.Particles[0].Y < 0 || ws.Particles[0].Y > float64(config.Height) {
			t.Errorf("Particle Y not wrapped: %v", ws.Particles[0].Y)
		}
	}
}

// TestGetGenreWeather tests genre-specific weather.
func TestGetGenreWeather(t *testing.T) {
	tests := []struct {
		name            string
		genreID         string
		expectedWeather []WeatherType
	}{
		{
			name:            "fantasy",
			genreID:         "fantasy",
			expectedWeather: []WeatherType{WeatherRain, WeatherSnow, WeatherFog},
		},
		{
			name:            "scifi",
			genreID:         "scifi",
			expectedWeather: []WeatherType{WeatherRain, WeatherDust, WeatherFog},
		},
		{
			name:            "horror",
			genreID:         "horror",
			expectedWeather: []WeatherType{WeatherFog, WeatherBloodRain, WeatherAsh},
		},
		{
			name:            "cyberpunk",
			genreID:         "cyberpunk",
			expectedWeather: []WeatherType{WeatherNeonRain, WeatherSmog, WeatherFog},
		},
		{
			name:            "postapoc",
			genreID:         "postapoc",
			expectedWeather: []WeatherType{WeatherSandstorm, WeatherAsh, WeatherRadiation},
		},
		{
			name:            "unknown",
			genreID:         "unknown",
			expectedWeather: []WeatherType{WeatherRain, WeatherSnow, WeatherFog},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			weather := GetGenreWeather(tt.genreID)

			if len(weather) != len(tt.expectedWeather) {
				t.Errorf("GetGenreWeather(%v) returned %d types, want %d",
					tt.genreID, len(weather), len(tt.expectedWeather))
				return
			}

			for i, w := range weather {
				if w != tt.expectedWeather[i] {
					t.Errorf("GetGenreWeather(%v)[%d] = %v, want %v",
						tt.genreID, i, w, tt.expectedWeather[i])
				}
			}
		})
	}
}

// TestWeatherSystem_MultipleUpdates tests multiple update cycles.
func TestWeatherSystem_MultipleUpdates(t *testing.T) {
	config := WeatherConfig{
		Type:      WeatherRain,
		Intensity: IntensityMedium,
		Width:     800,
		Height:    600,
		GenreID:   "fantasy",
		Seed:      12345,
	}

	ws, err := GenerateWeather(config)
	if err != nil {
		t.Fatalf("GenerateWeather failed: %v", err)
	}

	// Update multiple times
	deltaTime := 0.016
	for i := 0; i < 60; i++ { // Simulate 1 second at 60 FPS
		ws.Update(deltaTime)
	}

	// Check that system is still valid
	if len(ws.Particles) == 0 {
		t.Error("All particles disappeared after updates")
	}

	// Check elapsed time
	expectedTime := deltaTime * 60
	if ws.ElapsedTime < expectedTime*0.99 || ws.ElapsedTime > expectedTime*1.01 {
		t.Errorf("ElapsedTime = %v, want ~%v", ws.ElapsedTime, expectedTime)
	}

	// Verify particles are still within bounds
	for i, p := range ws.Particles {
		if p.X < 0 || p.X > float64(config.Width) {
			t.Errorf("Particle %d X out of bounds after updates: %v", i, p.X)
		}
		if p.Y < 0 || p.Y > float64(config.Height) {
			t.Errorf("Particle %d Y out of bounds after updates: %v", i, p.Y)
		}
	}
}

// TestWeatherSystem_Wind tests wind effects.
func TestWeatherSystem_Wind(t *testing.T) {
	config := WeatherConfig{
		Type:      WeatherRain,
		Intensity: IntensityLight,
		Width:     800,
		Height:    600,
		GenreID:   "fantasy",
		Seed:      12345,
		WindX:     50.0, // Strong wind
		WindY:     0.0,
	}

	ws, err := GenerateWeather(config)
	if err != nil {
		t.Fatalf("GenerateWeather failed: %v", err)
	}

	if len(ws.Particles) == 0 {
		t.Skip("No particles to test")
	}

	// Store initial X position
	initialX := ws.Particles[0].X

	// Update
	ws.Update(0.1) // Larger delta for more obvious effect

	// Check that particle moved in wind direction
	finalX := ws.Particles[0].X
	if finalX <= initialX {
		t.Error("Particle did not drift with wind")
	}
}

// BenchmarkGenerateWeather benchmarks weather generation.
func BenchmarkGenerateWeather(b *testing.B) {
	config := WeatherConfig{
		Type:      WeatherRain,
		Intensity: IntensityMedium,
		Width:     800,
		Height:    600,
		GenreID:   "fantasy",
		Seed:      12345,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = GenerateWeather(config)
	}
}

// BenchmarkWeatherSystem_Update benchmarks system updates.
func BenchmarkWeatherSystem_Update(b *testing.B) {
	config := WeatherConfig{
		Type:      WeatherRain,
		Intensity: IntensityMedium,
		Width:     800,
		Height:    600,
		GenreID:   "fantasy",
		Seed:      12345,
	}

	ws, err := GenerateWeather(config)
	if err != nil {
		b.Fatalf("GenerateWeather failed: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ws.Update(0.016)
	}
}

// BenchmarkGenerateWeather_AllTypes benchmarks all weather types.
func BenchmarkGenerateWeather_AllTypes(b *testing.B) {
	weatherTypes := []WeatherType{
		WeatherRain, WeatherSnow, WeatherFog, WeatherDust,
		WeatherAsh, WeatherNeonRain, WeatherSmog, WeatherRadiation,
		WeatherSandstorm, WeatherBloodRain,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, wt := range weatherTypes {
			config := WeatherConfig{
				Type:      wt,
				Intensity: IntensityMedium,
				Width:     800,
				Height:    600,
				GenreID:   "fantasy",
				Seed:      12345,
			}
			_, _ = GenerateWeather(config)
		}
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// TestWeatherEffect_NewWeatherEffect tests weather effect initialization.
func TestWeatherEffect_NewWeatherEffect(t *testing.T) {
	effect := NewWeatherEffect()

	if effect == nil {
		t.Fatal("NewWeatherEffect returned nil")
	}

	if effect.Puddles == nil {
		t.Error("Puddles map is nil")
	}

	if effect.SnowLevel == nil {
		t.Error("SnowLevel map is nil")
	}

	if effect.VisibilityModifier != 1.0 {
		t.Errorf("VisibilityModifier = %v, want 1.0", effect.VisibilityModifier)
	}
}

// TestWeatherSystem_VisibilityModifier tests visibility reduction.
func TestWeatherSystem_VisibilityModifier(t *testing.T) {
	tests := []struct {
		name              string
		weatherType       WeatherType
		intensity         WeatherIntensity
		minVisibility     float64
		maxVisibility     float64
	}{
		{"fog_light", WeatherFog, IntensityLight, 0.84, 0.86},
		{"fog_heavy", WeatherFog, IntensityHeavy, 0.49, 0.51},
		{"sandstorm_medium", WeatherSandstorm, IntensityMedium, 0.54, 0.56},
		{"sandstorm_extreme", WeatherSandstorm, IntensityExtreme, 0.19, 0.21},
		{"rain_medium", WeatherRain, IntensityMedium, 0.95, 1.05}, // No visibility impact
		{"blood_rain_heavy", WeatherBloodRain, IntensityHeavy, 0.84, 0.86},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := WeatherConfig{
				Type:      tt.weatherType,
				Intensity: tt.intensity,
				Width:     800,
				Height:    600,
				GenreID:   "fantasy",
				Seed:      12345,
			}

			ws, err := GenerateWeather(config)
			if err != nil {
				t.Fatalf("GenerateWeather failed: %v", err)
			}

			visibility := ws.GetVisibilityModifier()
			if visibility < tt.minVisibility || visibility > tt.maxVisibility {
				t.Errorf("Visibility = %v, want range [%v, %v]",
					visibility, tt.minVisibility, tt.maxVisibility)
			}
		})
	}
}

// TestWeatherSystem_RainAccumulation tests puddle formation.
func TestWeatherSystem_RainAccumulation(t *testing.T) {
	config := WeatherConfig{
		Type:      WeatherRain,
		Intensity: IntensityHeavy,
		Width:     800,
		Height:    600,
		GenreID:   "fantasy",
		Seed:      12345,
	}

	ws, err := GenerateWeather(config)
	if err != nil {
		t.Fatalf("GenerateWeather failed: %v", err)
	}

	// Simulate many updates to cause accumulation
	for i := 0; i < 1000; i++ {
		ws.Update(0.016)
	}

	// Check that some puddles have formed
	puddleCount := len(ws.Effects.Puddles)
	if puddleCount == 0 {
		t.Error("No puddles formed after 1000 updates")
	}

	// Verify puddle levels are within valid range
	for key, level := range ws.Effects.Puddles {
		if level < 0 || level > 1.0 {
			t.Errorf("Puddle %s has invalid level: %v", key, level)
		}
	}
}

// TestWeatherSystem_SnowAccumulation tests snow settling.
func TestWeatherSystem_SnowAccumulation(t *testing.T) {
	config := WeatherConfig{
		Type:      WeatherSnow,
		Intensity: IntensityHeavy,
		Width:     800,
		Height:    600,
		GenreID:   "fantasy",
		Seed:      12345,
	}

	ws, err := GenerateWeather(config)
	if err != nil {
		t.Fatalf("GenerateWeather failed: %v", err)
	}

	// Simulate many updates to cause accumulation
	for i := 0; i < 2000; i++ {
		ws.Update(0.016)
	}

	// Check that some snow has accumulated
	snowCount := len(ws.Effects.SnowLevel)
	if snowCount == 0 {
		t.Error("No snow accumulated after 2000 updates")
	}

	// Verify snow levels are within valid range
	for key, level := range ws.Effects.SnowLevel {
		if level < 0 || level > 1.0 {
			t.Errorf("Snow %s has invalid level: %v", key, level)
		}
	}

	// Check wind drift is set for snow
	if ws.Effects.WindDriftX == 0 && ws.Config.WindX != 0 {
		t.Error("Wind drift not set for snow")
	}
}

// TestWeatherSystem_GetPuddleLevel tests puddle level retrieval.
func TestWeatherSystem_GetPuddleLevel(t *testing.T) {
	config := WeatherConfig{
		Type:      WeatherRain,
		Intensity: IntensityMedium,
		Width:     800,
		Height:    600,
		GenreID:   "fantasy",
		Seed:      12345,
	}

	ws, err := GenerateWeather(config)
	if err != nil {
		t.Fatalf("GenerateWeather failed: %v", err)
	}

	// Manually set a puddle level
	ws.Effects.Puddles["5,10"] = 0.75

	level := ws.GetPuddleLevel(5, 10)
	if level != 0.75 {
		t.Errorf("GetPuddleLevel(5, 10) = %v, want 0.75", level)
	}

	// Non-existent puddle should return 0
	level = ws.GetPuddleLevel(100, 100)
	if level != 0.0 {
		t.Errorf("GetPuddleLevel(100, 100) = %v, want 0.0", level)
	}
}

// TestWeatherSystem_GetSnowLevel tests snow level retrieval.
func TestWeatherSystem_GetSnowLevel(t *testing.T) {
	config := WeatherConfig{
		Type:      WeatherSnow,
		Intensity: IntensityMedium,
		Width:     800,
		Height:    600,
		GenreID:   "fantasy",
		Seed:      12345,
	}

	ws, err := GenerateWeather(config)
	if err != nil {
		t.Fatalf("GenerateWeather failed: %v", err)
	}

	// Manually set a snow level
	ws.Effects.SnowLevel["3,7"] = 0.60

	level := ws.GetSnowLevel(3, 7)
	if level != 0.60 {
		t.Errorf("GetSnowLevel(3, 7) = %v, want 0.60", level)
	}

	// Non-existent snow should return 0
	level = ws.GetSnowLevel(200, 200)
	if level != 0.0 {
		t.Errorf("GetSnowLevel(200, 200) = %v, want 0.0", level)
	}
}

// TestGetGenreWeather_UpdatedGenres tests updated genre weather assignments.
func TestGetGenreWeather_UpdatedGenres(t *testing.T) {
	tests := []struct {
		name            string
		genreID         string
		expectedWeather []WeatherType
	}{
		{
			name:            "horror_includes_blood_rain",
			genreID:         "horror",
			expectedWeather: []WeatherType{WeatherFog, WeatherBloodRain, WeatherAsh},
		},
		{
			name:            "postapoc_includes_sandstorm",
			genreID:         "postapoc",
			expectedWeather: []WeatherType{WeatherSandstorm, WeatherAsh, WeatherRadiation},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			weather := GetGenreWeather(tt.genreID)

			if len(weather) != len(tt.expectedWeather) {
				t.Errorf("GetGenreWeather(%v) returned %d types, want %d",
					tt.genreID, len(weather), len(tt.expectedWeather))
				return
			}

			for i, w := range weather {
				if w != tt.expectedWeather[i] {
					t.Errorf("GetGenreWeather(%v)[%d] = %v, want %v",
						tt.genreID, i, w, tt.expectedWeather[i])
				}
			}
		})
	}
}

// TestWeatherSystem_HighParticleCount tests performance with 500-1000 particles.
func TestWeatherSystem_HighParticleCount(t *testing.T) {
	config := WeatherConfig{
		Type:      WeatherRain,
		Intensity: IntensityExtreme,
		Width:     1000,
		Height:    1000,
		GenreID:   "fantasy",
		Seed:      12345,
	}

	ws, err := GenerateWeather(config)
	if err != nil {
		t.Fatalf("GenerateWeather failed: %v", err)
	}

	particleCount := len(ws.Particles)
	if particleCount < 500 {
		t.Errorf("Particle count = %d, want >= 500", particleCount)
	}

	// Verify system can handle updates with many particles
	for i := 0; i < 60; i++ {
		ws.Update(0.016)
	}

	// System should still have particles
	if len(ws.Particles) == 0 {
		t.Error("All particles disappeared")
	}
}

// BenchmarkWeatherSystem_Update_HighCount benchmarks with 1000 particles.
func BenchmarkWeatherSystem_Update_HighCount(b *testing.B) {
	config := WeatherConfig{
		Type:      WeatherRain,
		Intensity: IntensityExtreme,
		Width:     1000,
		Height:    1000,
		GenreID:   "fantasy",
		Seed:      12345,
	}

	ws, err := GenerateWeather(config)
	if err != nil {
		b.Fatalf("GenerateWeather failed: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ws.Update(0.016)
	}
}
