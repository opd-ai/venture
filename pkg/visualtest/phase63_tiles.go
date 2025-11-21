// Package visualtest provides Phase 63.1 tile rendering regression tests.
package visualtest

import (
	"fmt"
	"image"
	"image/color"

	"github.com/opd-ai/venture/pkg/procgen/terrain"
)

// TileTestCase represents a tile rendering test scenario.
type TileTestCase struct {
	TileType terrain.TileType
	GenreID  string
	Seed     int64
	Size     int // Pixel size (typically 32x32)
}

// GenerateAllTileTests creates all 40 tile test cases for Phase 63.1.
// Covers 8 terrain types × 5 genres = 40 tiles.
func GenerateAllTileTests() []TileTestCase {
	tileTypes := []terrain.TileType{
		terrain.TileFloor,
		terrain.TileWall,
		terrain.TileDoor,
		terrain.TileStairsDown,
		terrain.TileStairsUp,
		terrain.TileWaterDeep,
		terrain.TileLavaFlow,
		terrain.TileTree,
	}

	genres := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapoc"}

	var tests []TileTestCase
	seed := int64(54321)

	for _, tileType := range tileTypes {
		for _, genre := range genres {
			tests = append(tests, TileTestCase{
				TileType: tileType,
				GenreID:  genre,
				Seed:     seed,
				Size:     32,
			})
			seed++
		}
	}

	return tests
}

// RenderTileTestCase renders a single tile test case to an image.
func RenderTileTestCase(tc TileTestCase) (*image.RGBA, error) {
	// Create tile image
	img := image.NewRGBA(image.Rect(0, 0, tc.Size, tc.Size))

	// Get genre-appropriate color for tile type
	tileColor := GetTileColor(tc.TileType, tc.GenreID)

	// Simple tile rendering (actual implementation would use rendering/tiles package)
	for y := 0; y < tc.Size; y++ {
		for x := 0; x < tc.Size; x++ {
			// Create pattern based on tile type and seed
			pattern := getDeterministicPattern(x, y, tc.Seed, tc.TileType)

			// Apply pattern to base color
			r := clampUint8(int(tileColor.R) + pattern)
			g := clampUint8(int(tileColor.G) + pattern)
			b := clampUint8(int(tileColor.B) + pattern)

			img.Set(x, y, color.RGBA{r, g, b, 255})
		}
	}

	return img, nil
}

// GetTileColor returns genre-appropriate color for tile type.
func GetTileColor(tileType terrain.TileType, genreID string) color.RGBA {
	genreColors := map[string]map[terrain.TileType]color.RGBA{
		"fantasy": {
			terrain.TileFloor:      {120, 100, 80, 255},
			terrain.TileWall:       {80, 80, 80, 255},
			terrain.TileDoor:       {139, 69, 19, 255},
			terrain.TileStairsDown: {100, 90, 70, 255},
			terrain.TileStairsUp:   {110, 100, 80, 255},
			terrain.TileWaterDeep:  {50, 100, 200, 255},
			terrain.TileLavaFlow:   {255, 100, 0, 255},
			terrain.TileTree:       {60, 120, 60, 255},
		},
		"scifi": {
			terrain.TileFloor:      {140, 140, 150, 255},
			terrain.TileWall:       {100, 100, 120, 255},
			terrain.TileDoor:       {0, 200, 255, 255},
			terrain.TileStairsDown: {120, 120, 140, 255},
			terrain.TileStairsUp:   {130, 130, 150, 255},
			terrain.TileWaterDeep:  {0, 150, 200, 255},
			terrain.TileLavaFlow:   {255, 150, 0, 255},
			terrain.TileTree:       {80, 140, 80, 255},
		},
		"horror": {
			terrain.TileFloor:      {60, 50, 50, 255},
			terrain.TileWall:       {40, 40, 40, 255},
			terrain.TileDoor:       {80, 60, 60, 255},
			terrain.TileStairsDown: {50, 45, 45, 255},
			terrain.TileStairsUp:   {55, 50, 50, 255},
			terrain.TileWaterDeep:  {30, 50, 80, 255},
			terrain.TileLavaFlow:   {150, 50, 0, 255},
			terrain.TileTree:       {40, 60, 40, 255},
		},
		"cyberpunk": {
			terrain.TileFloor:      {80, 80, 100, 255},
			terrain.TileWall:       {60, 60, 80, 255},
			terrain.TileDoor:       {255, 0, 128, 255},
			terrain.TileStairsDown: {70, 70, 90, 255},
			terrain.TileStairsUp:   {75, 75, 95, 255},
			terrain.TileWaterDeep:  {0, 200, 200, 255},
			terrain.TileLavaFlow:   {255, 0, 255, 255},
			terrain.TileTree:       {50, 100, 50, 255},
		},
		"postapoc": {
			terrain.TileFloor:      {100, 90, 70, 255},
			terrain.TileWall:       {70, 65, 55, 255},
			terrain.TileDoor:       {110, 100, 80, 255},
			terrain.TileStairsDown: {90, 80, 65, 255},
			terrain.TileStairsUp:   {95, 85, 70, 255},
			terrain.TileWaterDeep:  {60, 80, 100, 255},
			terrain.TileLavaFlow:   {200, 80, 20, 255},
			terrain.TileTree:       {70, 90, 60, 255},
		},
	}

	if genrePalette, ok := genreColors[genreID]; ok {
		if color, ok := genrePalette[tileType]; ok {
			return color
		}
	}

	return color.RGBA{100, 100, 100, 255}
}

// getDeterministicPattern creates a deterministic pattern value based on position and seed.
func getDeterministicPattern(x, y int, seed int64, tileType terrain.TileType) int {
	// Simple deterministic pattern using seed and position
	value := int(seed) + x*7 + y*13 + int(tileType)*31
	return (value % 31) - 15 // Range: -15 to +15
}

func clampUint8(value int) uint8 {
	if value < 0 {
		return 0
	}
	if value > 255 {
		return 255
	}
	return uint8(value)
}

// VerifyTileRegression runs all tile regression tests and returns results.
func VerifyTileRegression(baseline, current map[string]*image.RGBA) *ComparisonResult {
	var differences []Difference
	totalSimilarity := 0.0
	testCount := 0

	for testID, baselineImg := range baseline {
		currentImg, ok := current[testID]
		if !ok {
			differences = append(differences, Difference{
				Type:        "tile",
				Description: fmt.Sprintf("Missing tile test: %s", testID),
				Severity:    "critical",
				Similarity:  0.0,
			})
			continue
		}

		similarity := compareImages(baselineImg, currentImg)
		totalSimilarity += similarity
		testCount++

		// Tiles should have very high similarity (>0.95)
		threshold := 0.95

		if similarity < threshold {
			severity := "critical"
			if similarity >= threshold-0.05 {
				severity = "major"
			} else if similarity >= threshold-0.1 {
				severity = "minor"
			}

			differences = append(differences, Difference{
				Type:        "tile",
				Description: fmt.Sprintf("Tile regression for %s (%.2f%% similar, threshold %.2f%%)", testID, similarity*100, threshold*100),
				Severity:    severity,
				Similarity:  similarity,
			})
		}
	}

	avgSimilarity := 0.0
	if testCount > 0 {
		avgSimilarity = totalSimilarity / float64(testCount)
	}

	return &ComparisonResult{
		Passed:      len(differences) == 0,
		Differences: differences,
		Metrics: Metrics{
			TileSimilarity:    avgSimilarity,
			OverallSimilarity: avgSimilarity,
		},
	}
}

// ParticleTestCase represents a particle effect test scenario.
type ParticleTestCase struct {
	EffectType string // "weather" or "combat"
	Name       string // Specific effect name
	GenreID    string
	Seed       int64
	Size       int // Image size
}

// GenerateAllParticleTests creates all 20 particle test cases for Phase 63.1.
// 10 weather types + 10 combat effects = 20 tests.
func GenerateAllParticleTests() []ParticleTestCase {
	weatherTypes := []string{
		"rain", "snow", "fog", "dust", "ash",
		"neonrain", "smog", "radiation", "sandstorm", "bloodrain",
	}

	combatEffects := []string{
		"explosion", "fireball", "lightning", "frost", "poison",
		"blood", "smoke", "sparks", "magic", "energy",
	}

	var tests []ParticleTestCase
	seed := int64(99999)

	// Weather particle tests
	for _, weather := range weatherTypes {
		tests = append(tests, ParticleTestCase{
			EffectType: "weather",
			Name:       weather,
			GenreID:    "fantasy", // Weather is genre-agnostic, use fantasy as default
			Seed:       seed,
			Size:       128,
		})
		seed++
	}

	// Combat particle tests
	for _, combat := range combatEffects {
		tests = append(tests, ParticleTestCase{
			EffectType: "combat",
			Name:       combat,
			GenreID:    "fantasy",
			Seed:       seed,
			Size:       128,
		})
		seed++
	}

	return tests
}

// RenderParticleTestCase renders a particle effect test case.
func RenderParticleTestCase(tc ParticleTestCase) (*image.RGBA, error) {
	img := image.NewRGBA(image.Rect(0, 0, tc.Size, tc.Size))

	// Fill with transparent background
	transparent := color.RGBA{0, 0, 0, 0}
	for y := 0; y < tc.Size; y++ {
		for x := 0; x < tc.Size; x++ {
			img.Set(x, y, transparent)
		}
	}

	// Get effect color
	effectColor := GetParticleColor(tc.Name, tc.EffectType)

	// Render simple particle pattern (actual implementation uses particle system)
	// This creates deterministic particle distribution based on seed
	particleCount := 50
	for i := 0; i < particleCount; i++ {
		px := (int(tc.Seed) + i*17) % tc.Size
		py := (int(tc.Seed) + i*23) % tc.Size

		// Draw particle with falloff
		for dy := -2; dy <= 2; dy++ {
			for dx := -2; dx <= 2; dx++ {
				x := px + dx
				y := py + dy
				if x >= 0 && x < tc.Size && y >= 0 && y < tc.Size {
					dist := dx*dx + dy*dy
					if dist <= 4 {
						alpha := uint8((5 - dist) * 50)
						img.Set(x, y, color.RGBA{
							effectColor.R,
							effectColor.G,
							effectColor.B,
							alpha,
						})
					}
				}
			}
		}
	}

	return img, nil
}

// GetParticleColor returns color for particle effect.
func GetParticleColor(name, effectType string) color.RGBA {
	weatherColors := map[string]color.RGBA{
		"rain":      {100, 150, 255, 200},
		"snow":      {240, 240, 255, 220},
		"fog":       {200, 200, 200, 150},
		"dust":      {150, 140, 120, 180},
		"ash":       {100, 100, 100, 160},
		"neonrain":  {255, 0, 255, 200},
		"smog":      {120, 120, 100, 170},
		"radiation": {100, 255, 100, 190},
		"sandstorm": {200, 180, 120, 180},
		"bloodrain": {150, 0, 0, 200},
	}

	combatColors := map[string]color.RGBA{
		"explosion": {255, 150, 0, 255},
		"fireball":  {255, 100, 0, 240},
		"lightning": {200, 200, 255, 255},
		"frost":     {150, 200, 255, 220},
		"poison":    {100, 200, 50, 200},
		"blood":     {180, 0, 0, 230},
		"smoke":     {100, 100, 100, 180},
		"sparks":    {255, 200, 100, 255},
		"magic":     {200, 100, 255, 240},
		"energy":    {100, 255, 255, 250},
	}

	if effectType == "weather" {
		if color, ok := weatherColors[name]; ok {
			return color
		}
	} else if effectType == "combat" {
		if color, ok := combatColors[name]; ok {
			return color
		}
	}

	return color.RGBA{128, 128, 128, 200}
}

// VerifyParticleRegression runs particle regression tests.
func VerifyParticleRegression(baseline, current map[string]*image.RGBA) *ComparisonResult {
	var differences []Difference
	totalSimilarity := 0.0
	testCount := 0

	for testID, baselineImg := range baseline {
		currentImg, ok := current[testID]
		if !ok {
			differences = append(differences, Difference{
				Type:        "particle",
				Description: fmt.Sprintf("Missing particle test: %s", testID),
				Severity:    "critical",
				Similarity:  0.0,
			})
			continue
		}

		similarity := compareImages(baselineImg, currentImg)
		totalSimilarity += similarity
		testCount++

		// Particles can vary slightly due to transparency (lower threshold)
		threshold := 0.90

		if similarity < threshold {
			severity := "major"
			if similarity < threshold-0.1 {
				severity = "critical"
			}

			differences = append(differences, Difference{
				Type:        "particle",
				Description: fmt.Sprintf("Particle regression for %s (%.2f%% similar, threshold %.2f%%)", testID, similarity*100, threshold*100),
				Severity:    severity,
				Similarity:  similarity,
			})
		}
	}

	avgSimilarity := 0.0
	if testCount > 0 {
		avgSimilarity = totalSimilarity / float64(testCount)
	}

	return &ComparisonResult{
		Passed:      len(differences) == 0,
		Differences: differences,
		Metrics: Metrics{
			OverallSimilarity: avgSimilarity,
		},
	}
}
