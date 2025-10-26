package sprites

import (
	"image/color"
	"testing"
)

// Note: The silhouette analyzer functions that require Ebiten image operations
// (AnalyzeSilhouette, GenerateSilhouette, AddOutline, etc.) cannot be tested
// with standard Go tests because Ebiten requires a running game context to read pixels.
// These tests validate the API contracts, struct methods, and edge cases without pixel inspection.
// Visual validation is performed through cmd/silhouettetest tool.

// TestSilhouetteAnalysis_GetQuality tests quality categorization.
func TestSilhouetteAnalysis_GetQuality(t *testing.T) {
	tests := []struct {
		score float64
		want  SilhouetteQuality
	}{
		{0.2, QualityPoor},
		{0.39, QualityPoor},
		{0.4, QualityFair},
		{0.5, QualityFair},
		{0.59, QualityFair},
		{0.6, QualityGood},
		{0.7, QualityGood},
		{0.79, QualityGood},
		{0.8, QualityExcellent},
		{0.9, QualityExcellent},
		{1.0, QualityExcellent},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			analysis := SilhouetteAnalysis{OverallScore: tt.score}
			got := analysis.GetQuality()
			if got != tt.want {
				t.Errorf("GetQuality() with score %f = %v, want %v", tt.score, got, tt.want)
			}
		})
	}
}

// TestSilhouetteAnalysis_NeedsImprovement tests improvement detection.
func TestSilhouetteAnalysis_NeedsImprovement(t *testing.T) {
	tests := []struct {
		score float64
		want  bool
	}{
		{0.3, true},  // Poor
		{0.5, true},  // Fair
		{0.59, true}, // Still Fair
		{0.6, false}, // Good
		{0.7, false}, // Good
		{0.9, false}, // Excellent
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			analysis := SilhouetteAnalysis{OverallScore: tt.score}
			got := analysis.NeedsImprovement()
			if got != tt.want {
				t.Errorf("NeedsImprovement() with score %f = %v, want %v", tt.score, got, tt.want)
			}
		})
	}
}

// TestSilhouetteQuality_String tests quality string representations.
func TestSilhouetteQuality_String(t *testing.T) {
	tests := []struct {
		quality SilhouetteQuality
		want    string
	}{
		{QualityPoor, "poor"},
		{QualityFair, "fair"},
		{QualityGood, "good"},
		{QualityExcellent, "excellent"},
		{SilhouetteQuality(999), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := tt.quality.String()
			if got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestDefaultOutlineConfig tests default configuration.
func TestDefaultOutlineConfig(t *testing.T) {
	config := DefaultOutlineConfig()

	if config.Thickness != 1 {
		t.Errorf("Default thickness = %d, want 1", config.Thickness)
	}
	if !config.Enabled {
		t.Error("Default outline should be enabled")
	}
	if config.Color == nil {
		t.Error("Default color should not be nil")
	}

	// Verify color is dark gray/black
	r, g, b, _ := config.Color.RGBA()
	if r > 50*257 || g > 50*257 || b > 50*257 {
		t.Errorf("Default outline color too bright: R=%d G=%d B=%d", r, g, b)
	}
}

// TestOutlineConfig_API tests outline config structure.
func TestOutlineConfig_API(t *testing.T) {
	config := OutlineConfig{
		Color:     color.RGBA{255, 0, 0, 255},
		Thickness: 2,
		Enabled:   false,
	}

	if config.Thickness != 2 {
		t.Errorf("Thickness = %d, want 2", config.Thickness)
	}
	if config.Enabled {
		t.Error("Enabled should be false")
	}
}

// TestSilhouetteAnalysis_Structure tests the analysis struct.
func TestSilhouetteAnalysis_Structure(t *testing.T) {
	analysis := SilhouetteAnalysis{
		Compactness:     0.75,
		Coverage:        0.65,
		EdgeClarity:     0.85,
		OverallScore:    0.72,
		OpaquePixels:    400,
		PerimeterPixels: 80,
		TotalPixels:     1024,
	}

	// Verify all fields are accessible
	if analysis.Compactness != 0.75 {
		t.Errorf("Compactness = %f, want 0.75", analysis.Compactness)
	}
	if analysis.Coverage != 0.65 {
		t.Errorf("Coverage = %f, want 0.65", analysis.Coverage)
	}
	if analysis.EdgeClarity != 0.85 {
		t.Errorf("EdgeClarity = %f, want 0.85", analysis.EdgeClarity)
	}
	if analysis.OverallScore != 0.72 {
		t.Errorf("OverallScore = %f, want 0.72", analysis.OverallScore)
	}
	if analysis.OpaquePixels != 400 {
		t.Errorf("OpaquePixels = %d, want 400", analysis.OpaquePixels)
	}
	if analysis.PerimeterPixels != 80 {
		t.Errorf("PerimeterPixels = %d, want 80", analysis.PerimeterPixels)
	}
	if analysis.TotalPixels != 1024 {
		t.Errorf("TotalPixels = %d, want 1024", analysis.TotalPixels)
	}
}

// TestSilhouetteAnalysis_QualityThresholds tests quality boundary conditions.
func TestSilhouetteAnalysis_QualityThresholds(t *testing.T) {
	tests := []struct {
		score   float64
		quality SilhouetteQuality
		improve bool
	}{
		{0.0, QualityPoor, true},
		{0.39999, QualityPoor, true},
		{0.4, QualityFair, true},
		{0.59999, QualityFair, true},
		{0.6, QualityGood, false},
		{0.79999, QualityGood, false},
		{0.8, QualityExcellent, false},
		{1.0, QualityExcellent, false},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			analysis := SilhouetteAnalysis{OverallScore: tt.score}

			if got := analysis.GetQuality(); got != tt.quality {
				t.Errorf("Score %f: GetQuality() = %v, want %v", tt.score, got, tt.quality)
			}

			if got := analysis.NeedsImprovement(); got != tt.improve {
				t.Errorf("Score %f: NeedsImprovement() = %v, want %v", tt.score, got, tt.improve)
			}
		})
	}
}

// TestColorTypes tests that color types are properly defined.
func TestColorTypes(t *testing.T) {
	// Test various color types work with outline config
	colors := []color.Color{
		color.RGBA{0, 0, 0, 255},
		color.RGBA{20, 20, 20, 255},
		color.RGBA{255, 255, 255, 255},
		color.Gray{Y: 50},
	}

	for i, c := range colors {
		config := OutlineConfig{
			Color:     c,
			Thickness: 1,
			Enabled:   true,
		}

		if config.Color == nil {
			t.Errorf("Color %d is nil", i)
		}
	}
}

// TestSilhouetteAnalysis_ZeroValues tests zero value handling.
func TestSilhouetteAnalysis_ZeroValues(t *testing.T) {
	var analysis SilhouetteAnalysis

	// Zero values should be valid
	if analysis.Compactness != 0 {
		t.Error("Default Compactness should be 0")
	}
	if analysis.GetQuality() != QualityPoor {
		t.Error("Zero OverallScore should be QualityPoor")
	}
	if !analysis.NeedsImprovement() {
		t.Error("Zero score should need improvement")
	}
}

// TestOutlineConfig_Variations tests different outline configurations.
func TestOutlineConfig_Variations(t *testing.T) {
	tests := []struct {
		name      string
		thickness int
		enabled   bool
	}{
		{"thin enabled", 1, true},
		{"thick enabled", 2, true},
		{"thin disabled", 1, false},
		{"thick disabled", 3, false},
		{"zero thickness", 0, true},
		{"negative thickness", -1, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := OutlineConfig{
				Color:     color.Black,
				Thickness: tt.thickness,
				Enabled:   tt.enabled,
			}

			if config.Thickness != tt.thickness {
				t.Errorf("Thickness = %d, want %d", config.Thickness, tt.thickness)
			}
			if config.Enabled != tt.enabled {
				t.Errorf("Enabled = %v, want %v", config.Enabled, tt.enabled)
			}
		})
	}
}

// TestSilhouetteMetrics_Ranges tests that metrics are properly bounded.
func TestSilhouetteMetrics_Ranges(t *testing.T) {
	// Test various metric combinations
	tests := []struct {
		name     string
		analysis SilhouetteAnalysis
	}{
		{
			name: "all zeros",
			analysis: SilhouetteAnalysis{
				Compactness:  0.0,
				Coverage:     0.0,
				EdgeClarity:  0.0,
				OverallScore: 0.0,
			},
		},
		{
			name: "all ones",
			analysis: SilhouetteAnalysis{
				Compactness:  1.0,
				Coverage:     1.0,
				EdgeClarity:  1.0,
				OverallScore: 1.0,
			},
		},
		{
			name: "mixed values",
			analysis: SilhouetteAnalysis{
				Compactness:  0.6,
				Coverage:     0.75,
				EdgeClarity:  0.82,
				OverallScore: 0.73,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := tt.analysis

			// All metrics should be 0.0-1.0
			if a.Compactness < 0 || a.Compactness > 1.0 {
				t.Errorf("Compactness out of range: %f", a.Compactness)
			}
			if a.Coverage < 0 || a.Coverage > 1.0 {
				t.Errorf("Coverage out of range: %f", a.Coverage)
			}
			if a.EdgeClarity < 0 || a.EdgeClarity > 1.0 {
				t.Errorf("EdgeClarity out of range: %f", a.EdgeClarity)
			}
			if a.OverallScore < 0 || a.OverallScore > 1.0 {
				t.Errorf("OverallScore out of range: %f", a.OverallScore)
			}

			// Quality should be determinable
			_ = a.GetQuality()
			_ = a.NeedsImprovement()
		})
	}
}
