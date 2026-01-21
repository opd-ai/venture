// Package postprocess provides post-processing effects for rendered scenes.
package postprocess

import (
	"testing"
)

func TestFantasyPreset(t *testing.T) {
	preset := FantasyPreset()

	if preset.Name != "Fantasy" {
		t.Errorf("FantasyPreset.Name = %q, want %q", preset.Name, "Fantasy")
	}

	if preset.Config.ColorGrading.Temperature <= 0 {
		t.Error("Fantasy preset should have warm temperature (positive)")
	}

	if preset.Config.ColorGrading.Saturation <= 1.0 {
		t.Error("Fantasy preset should have increased saturation")
	}

	if !preset.Config.Vignette.Enabled {
		t.Error("Fantasy preset should have vignette enabled")
	}

	if preset.Config.ChromaticAberration.Enabled {
		t.Error("Fantasy preset should not have chromatic aberration")
	}
}

func TestSciFiPreset(t *testing.T) {
	preset := SciFiPreset()

	if preset.Name != "Sci-Fi" {
		t.Errorf("SciFiPreset.Name = %q, want %q", preset.Name, "Sci-Fi")
	}

	if preset.Config.ColorGrading.Temperature >= 0 {
		t.Error("Sci-Fi preset should have cool temperature (negative)")
	}

	if preset.Config.ColorGrading.Contrast <= 1.0 {
		t.Error("Sci-Fi preset should have increased contrast")
	}

	if !preset.Config.ChromaticAberration.Enabled {
		t.Error("Sci-Fi preset should have chromatic aberration enabled")
	}
}

func TestHorrorPreset(t *testing.T) {
	preset := HorrorPreset()

	if preset.Name != "Horror" {
		t.Errorf("HorrorPreset.Name = %q, want %q", preset.Name, "Horror")
	}

	if preset.Config.ColorGrading.Saturation >= 1.0 {
		t.Error("Horror preset should have reduced saturation")
	}

	if preset.Config.ColorGrading.Brightness >= 0 {
		t.Error("Horror preset should have reduced brightness")
	}

	if !preset.Config.Vignette.Enabled {
		t.Error("Horror preset should have vignette enabled")
	}

	if preset.Config.Vignette.Intensity <= 0.5 {
		t.Error("Horror preset should have strong vignette (>0.5)")
	}
}

func TestCyberpunkPreset(t *testing.T) {
	preset := CyberpunkPreset()

	if preset.Name != "Cyberpunk" {
		t.Errorf("CyberpunkPreset.Name = %q, want %q", preset.Name, "Cyberpunk")
	}

	if preset.Config.ColorGrading.Saturation <= 1.0 {
		t.Error("Cyberpunk preset should have high saturation")
	}

	if preset.Config.ColorGrading.Contrast <= 1.0 {
		t.Error("Cyberpunk preset should have high contrast")
	}

	if !preset.Config.ChromaticAberration.Enabled {
		t.Error("Cyberpunk preset should have chromatic aberration enabled")
	}
}

func TestPostApocalypticPreset(t *testing.T) {
	preset := PostApocalypticPreset()

	if preset.Name != "Post-Apocalyptic" {
		t.Errorf("PostApocalypticPreset.Name = %q, want %q", preset.Name, "Post-Apocalyptic")
	}

	if preset.Config.ColorGrading.Saturation >= 1.0 {
		t.Error("Post-Apocalyptic preset should have reduced saturation")
	}

	if preset.Config.ColorGrading.Temperature <= 0 {
		t.Error("Post-Apocalyptic preset should have warm/dusty temperature")
	}

	if !preset.Config.Vignette.Enabled {
		t.Error("Post-Apocalyptic preset should have vignette enabled")
	}
}

func TestNeutralPreset(t *testing.T) {
	preset := NeutralPreset()

	if preset.Name != "Neutral" {
		t.Errorf("NeutralPreset.Name = %q, want %q", preset.Name, "Neutral")
	}

	defaultConfig := DefaultConfig()

	if preset.Config.ColorGrading.Saturation != defaultConfig.ColorGrading.Saturation {
		t.Error("Neutral preset should use default saturation")
	}

	if preset.Config.ColorGrading.Contrast != defaultConfig.ColorGrading.Contrast {
		t.Error("Neutral preset should use default contrast")
	}
}

func TestCinematicPreset(t *testing.T) {
	preset := CinematicPreset()

	if preset.Name != "Cinematic" {
		t.Errorf("CinematicPreset.Name = %q, want %q", preset.Name, "Cinematic")
	}

	if !preset.Config.DepthBlur.Enabled {
		t.Error("Cinematic preset should have depth blur enabled")
	}

	if !preset.Config.Vignette.Enabled {
		t.Error("Cinematic preset should have strong vignette enabled")
	}

	if preset.Config.Vignette.Intensity <= 0.5 {
		t.Error("Cinematic preset should have strong vignette (>0.5)")
	}
}

func TestGetPresetByGenre(t *testing.T) {
	tests := []struct {
		name     string
		genreID  string
		wantName string
	}{
		{"fantasy", "fantasy", "Fantasy"},
		{"sci-fi", "scifi", "Sci-Fi"},
		{"sci-fi-hyphen", "sci-fi", "Sci-Fi"},
		{"horror", "horror", "Horror"},
		{"cyberpunk", "cyberpunk", "Cyberpunk"},
		{"post-apocalyptic", "postapoc", "Post-Apocalyptic"},
		{"post-apocalyptic-full", "post-apocalyptic", "Post-Apocalyptic"},
		{"cinematic", "cinematic", "Cinematic"},
		{"neutral", "neutral", "Neutral"},
		{"unknown", "unknown", "Neutral"},
		{"empty", "", "Neutral"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			preset := GetPresetByGenre(tt.genreID)
			if preset.Name != tt.wantName {
				t.Errorf("GetPresetByGenre(%q).Name = %q, want %q",
					tt.genreID, preset.Name, tt.wantName)
			}
		})
	}
}

func TestAllPresets(t *testing.T) {
	presets := AllPresets()

	expectedCount := 7 // Fantasy, SciFi, Horror, Cyberpunk, PostApoc, Neutral, Cinematic
	if len(presets) != expectedCount {
		t.Errorf("AllPresets() returned %d presets, want %d", len(presets), expectedCount)
	}

	// Check that all presets have names
	for i, preset := range presets {
		if preset.Name == "" {
			t.Errorf("Preset %d has empty name", i)
		}
		if preset.Description == "" {
			t.Errorf("Preset %d (%s) has empty description", i, preset.Name)
		}
	}
}

func TestPresetConsistency(t *testing.T) {
	// Test that GetPresetByGenre returns the same as direct calls
	tests := []struct {
		genreID string
		direct  Preset
	}{
		{"fantasy", FantasyPreset()},
		{"scifi", SciFiPreset()},
		{"horror", HorrorPreset()},
		{"cyberpunk", CyberpunkPreset()},
		{"postapoc", PostApocalypticPreset()},
	}

	for _, tt := range tests {
		t.Run(tt.genreID, func(t *testing.T) {
			byGenre := GetPresetByGenre(tt.genreID)

			if byGenre.Name != tt.direct.Name {
				t.Errorf("GetPresetByGenre(%q).Name = %q, want %q",
					tt.genreID, byGenre.Name, tt.direct.Name)
			}

			if byGenre.Config.ColorGrading.Saturation != tt.direct.Config.ColorGrading.Saturation {
				t.Errorf("GetPresetByGenre(%q) saturation mismatch", tt.genreID)
			}

			if byGenre.Config.Vignette.Intensity != tt.direct.Config.Vignette.Intensity {
				t.Errorf("GetPresetByGenre(%q) vignette intensity mismatch", tt.genreID)
			}
		})
	}
}

// Benchmark preset retrieval
func BenchmarkGetPresetByGenre(b *testing.B) {
	genreIDs := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapoc"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		genreID := genreIDs[i%len(genreIDs)]
		_ = GetPresetByGenre(genreID)
	}
}

func BenchmarkAllPresets(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = AllPresets()
	}
}
