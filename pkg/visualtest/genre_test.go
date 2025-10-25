package visualtest

import (
	"image"
	"image/color"
	"testing"
)

// TestGenreValidator tests basic genre validation.
func TestGenreValidator(t *testing.T) {
	validator := NewGenreValidator(0.3) // 30% distinctness required

	// Add distinct genres (different colors)
	validator.AddGenreSnapshot("fantasy", &Snapshot{
		GenreID:      "fantasy",
		SpriteImage:  CreateTestImage(28, 28, color.RGBA{255, 0, 0, 255}), // Red
		TileImage:    CreateTestImage(32, 32, color.RGBA{200, 50, 50, 255}),
		PaletteImage: CreateTestImage(16, 16, color.RGBA{150, 0, 0, 255}),
	})

	validator.AddGenreSnapshot("scifi", &Snapshot{
		GenreID:      "scifi",
		SpriteImage:  CreateTestImage(28, 28, color.RGBA{0, 0, 255, 255}), // Blue
		TileImage:    CreateTestImage(32, 32, color.RGBA{50, 50, 200, 255}),
		PaletteImage: CreateTestImage(16, 16, color.RGBA{0, 0, 150, 255}),
	})

	result := validator.Validate()

	if !result.Passed {
		t.Errorf("Validation failed with distinct genres: %+v", result.Issues)
	}

	if result.Summary.TotalGenres != 2 {
		t.Errorf("TotalGenres = %d, want 2", result.Summary.TotalGenres)
	}

	if result.Summary.TotalComparisons != 1 {
		t.Errorf("TotalComparisons = %d, want 1", result.Summary.TotalComparisons)
	}
}

// TestGenreValidator_SimilarGenres tests detection of similar genres.
func TestGenreValidator_SimilarGenres(t *testing.T) {
	validator := NewGenreValidator(0.3) // 30% distinctness required

	// Add similar genres (same colors)
	validator.AddGenreSnapshot("genre1", &Snapshot{
		GenreID:      "genre1",
		SpriteImage:  CreateTestImage(28, 28, color.RGBA{128, 128, 128, 255}),
		TileImage:    CreateTestImage(32, 32, color.RGBA{128, 128, 128, 255}),
		PaletteImage: CreateTestImage(16, 16, color.RGBA{128, 128, 128, 255}),
	})

	validator.AddGenreSnapshot("genre2", &Snapshot{
		GenreID:      "genre2",
		SpriteImage:  CreateTestImage(28, 28, color.RGBA{128, 128, 128, 255}),
		TileImage:    CreateTestImage(32, 32, color.RGBA{128, 128, 128, 255}),
		PaletteImage: CreateTestImage(16, 16, color.RGBA{128, 128, 128, 255}),
	})

	result := validator.Validate()

	if result.Passed {
		t.Error("Validation passed with identical genres, expected failure")
	}

	if len(result.Issues) == 0 {
		t.Error("Expected issues with similar genres, got none")
	}

	if result.Summary.FailedComparisons == 0 {
		t.Error("Expected failed comparisons, got 0")
	}
}

// TestGenreValidator_MultipleGenres tests validation with 5 genres.
func TestGenreValidator_MultipleGenres(t *testing.T) {
	validator := NewGenreValidator(0.3)

	genres := map[string]color.RGBA{
		"fantasy":         {255, 0, 0, 255},   // Red
		"scifi":           {0, 0, 255, 255},   // Blue
		"horror":          {100, 0, 100, 255}, // Purple
		"cyberpunk":       {255, 0, 255, 255}, // Magenta
		"postapocalyptic": {150, 75, 0, 255},  // Brown
	}

	for genreID, col := range genres {
		validator.AddGenreSnapshot(genreID, &Snapshot{
			GenreID:      genreID,
			SpriteImage:  CreateTestImage(28, 28, col),
			TileImage:    CreateTestImage(32, 32, col),
			PaletteImage: CreateTestImage(16, 16, col),
		})
	}

	result := validator.Validate()

	if result.Summary.TotalGenres != 5 {
		t.Errorf("TotalGenres = %d, want 5", result.Summary.TotalGenres)
	}

	// 5 genres = C(5,2) = 10 comparisons
	expectedComparisons := 10
	if result.Summary.TotalComparisons != expectedComparisons {
		t.Errorf("TotalComparisons = %d, want %d", result.Summary.TotalComparisons, expectedComparisons)
	}

	if len(result.Comparisons) != expectedComparisons {
		t.Errorf("Comparisons count = %d, want %d", len(result.Comparisons), expectedComparisons)
	}
}

// TestColorSimilarity tests color similarity calculation.
func TestColorSimilarity(t *testing.T) {
	tests := []struct {
		name          string
		c1            color.RGBA
		c2            color.RGBA
		minSimilarity float64
		maxSimilarity float64
	}{
		{
			name:          "identical colors",
			c1:            color.RGBA{255, 0, 0, 255},
			c2:            color.RGBA{255, 0, 0, 255},
			minSimilarity: 1.0,
			maxSimilarity: 1.0,
		},
		{
			name:          "opposite colors",
			c1:            color.RGBA{255, 255, 255, 255},
			c2:            color.RGBA{0, 0, 0, 255},
			minSimilarity: 0.0,
			maxSimilarity: 0.1,
		},
		{
			name:          "similar reds",
			c1:            color.RGBA{255, 0, 0, 255},
			c2:            color.RGBA{250, 0, 0, 255},
			minSimilarity: 0.95,
			maxSimilarity: 1.0,
		},
		{
			name:          "red vs blue",
			c1:            color.RGBA{255, 0, 0, 255},
			c2:            color.RGBA{0, 0, 255, 255},
			minSimilarity: 0.0,
			maxSimilarity: 0.5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sim := colorSimilarity(tt.c1, tt.c2)

			if sim < tt.minSimilarity || sim > tt.maxSimilarity {
				t.Errorf("Similarity = %f, want range [%f, %f]",
					sim, tt.minSimilarity, tt.maxSimilarity)
			}
		})
	}
}

// TestExtractDominantColors tests color extraction.
func TestExtractDominantColors(t *testing.T) {
	img := CreateTestImage(16, 16, color.RGBA{128, 64, 32, 255})

	colors := extractDominantColors(img, 4)

	if len(colors) == 0 {
		t.Error("No colors extracted")
	}

	if len(colors) > 4 {
		t.Errorf("Extracted %d colors, want at most 4", len(colors))
	}

	// Check that extracted colors match the solid fill
	for _, col := range colors {
		if col.R != 128 || col.G != 64 || col.B != 32 {
			t.Errorf("Extracted color (%d,%d,%d), want (128,64,32)", col.R, col.G, col.B)
		}
	}
}

// TestExtractDominantColors_Nil tests nil image handling.
func TestExtractDominantColors_Nil(t *testing.T) {
	colors := extractDominantColors(nil, 4)

	if colors != nil {
		t.Error("Expected nil for nil image, got colors")
	}
}

// TestCalculatePaletteSimilarity tests palette comparison.
func TestCalculatePaletteSimilarity(t *testing.T) {
	tests := []struct {
		name          string
		palette1      *image.RGBA
		palette2      *image.RGBA
		minSimilarity float64
		maxSimilarity float64
	}{
		{
			name:          "identical palettes",
			palette1:      CreateTestImage(16, 16, color.RGBA{255, 0, 0, 255}),
			palette2:      CreateTestImage(16, 16, color.RGBA{255, 0, 0, 255}),
			minSimilarity: 0.95,
			maxSimilarity: 1.0,
		},
		{
			name:          "different palettes",
			palette1:      CreateTestImage(16, 16, color.RGBA{255, 0, 0, 255}),
			palette2:      CreateTestImage(16, 16, color.RGBA{0, 0, 255, 255}),
			minSimilarity: 0.0,
			maxSimilarity: 0.5,
		},
		{
			name:          "nil palettes",
			palette1:      nil,
			palette2:      nil,
			minSimilarity: 1.0,
			maxSimilarity: 1.0,
		},
		{
			name:          "one nil palette",
			palette1:      CreateTestImage(16, 16, color.RGBA{255, 0, 0, 255}),
			palette2:      nil,
			minSimilarity: 0.0,
			maxSimilarity: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sim := calculatePaletteSimilarity(tt.palette1, tt.palette2)

			if sim < tt.minSimilarity || sim > tt.maxSimilarity {
				t.Errorf("Palette similarity = %f, want range [%f, %f]",
					sim, tt.minSimilarity, tt.maxSimilarity)
			}
		})
	}
}

// TestValidateGenreSet tests the convenience function.
func TestValidateGenreSet(t *testing.T) {
	genreSnapshots := map[string]*Snapshot{
		"fantasy": {
			GenreID:      "fantasy",
			SpriteImage:  CreateTestImage(28, 28, color.RGBA{255, 0, 0, 255}),
			TileImage:    CreateTestImage(32, 32, color.RGBA{200, 50, 50, 255}),
			PaletteImage: CreateTestImage(16, 16, color.RGBA{150, 0, 0, 255}),
		},
		"scifi": {
			GenreID:      "scifi",
			SpriteImage:  CreateTestImage(28, 28, color.RGBA{0, 0, 255, 255}),
			TileImage:    CreateTestImage(32, 32, color.RGBA{50, 50, 200, 255}),
			PaletteImage: CreateTestImage(16, 16, color.RGBA{0, 0, 150, 255}),
		},
	}

	result := ValidateGenreSet(genreSnapshots, 0.3)

	if !result.Passed {
		t.Errorf("Validation failed: %+v", result.Issues)
	}

	if result.Summary.TotalGenres != 2 {
		t.Errorf("TotalGenres = %d, want 2", result.Summary.TotalGenres)
	}
}

// TestGenreComparison_Metrics tests comparison metric calculation.
func TestGenreComparison_Metrics(t *testing.T) {
	validator := NewGenreValidator(0.3)

	validator.AddGenreSnapshot("fantasy", &Snapshot{
		GenreID:      "fantasy",
		SpriteImage:  CreateTestImage(28, 28, color.RGBA{255, 0, 0, 255}),
		TileImage:    CreateTestImage(32, 32, color.RGBA{255, 0, 0, 255}),
		PaletteImage: CreateTestImage(16, 16, color.RGBA{255, 0, 0, 255}),
	})

	validator.AddGenreSnapshot("scifi", &Snapshot{
		GenreID:      "scifi",
		SpriteImage:  CreateTestImage(28, 28, color.RGBA{0, 0, 255, 255}),
		TileImage:    CreateTestImage(32, 32, color.RGBA{0, 0, 255, 255}),
		PaletteImage: CreateTestImage(16, 16, color.RGBA{0, 0, 255, 255}),
	})

	result := validator.Validate()

	if len(result.Comparisons) != 1 {
		t.Fatalf("Expected 1 comparison, got %d", len(result.Comparisons))
	}

	comp := result.Comparisons[0]

	// Verify all metrics are in valid range
	if comp.SpriteSimilarity < 0.0 || comp.SpriteSimilarity > 1.0 {
		t.Errorf("SpriteSimilarity = %f, want [0.0, 1.0]", comp.SpriteSimilarity)
	}
	if comp.TileSimilarity < 0.0 || comp.TileSimilarity > 1.0 {
		t.Errorf("TileSimilarity = %f, want [0.0, 1.0]", comp.TileSimilarity)
	}
	if comp.PaletteSimilarity < 0.0 || comp.PaletteSimilarity > 1.0 {
		t.Errorf("PaletteSimilarity = %f, want [0.0, 1.0]", comp.PaletteSimilarity)
	}
	if comp.OverallSimilarity < 0.0 || comp.OverallSimilarity > 1.0 {
		t.Errorf("OverallSimilarity = %f, want [0.0, 1.0]", comp.OverallSimilarity)
	}
}

// TestGenreValidationSummary tests summary statistics.
func TestGenreValidationSummary(t *testing.T) {
	validator := NewGenreValidator(0.3)

	// Add 3 distinct genres
	validator.AddGenreSnapshot("a", &Snapshot{
		GenreID:     "a",
		SpriteImage: CreateTestImage(10, 10, color.RGBA{255, 0, 0, 255}),
	})
	validator.AddGenreSnapshot("b", &Snapshot{
		GenreID:     "b",
		SpriteImage: CreateTestImage(10, 10, color.RGBA{0, 255, 0, 255}),
	})
	validator.AddGenreSnapshot("c", &Snapshot{
		GenreID:     "c",
		SpriteImage: CreateTestImage(10, 10, color.RGBA{0, 0, 255, 255}),
	})

	result := validator.Validate()

	summary := result.Summary

	// 3 genres = 3 comparisons
	if summary.TotalComparisons != 3 {
		t.Errorf("TotalComparisons = %d, want 3", summary.TotalComparisons)
	}

	// Check min/max/avg are in valid ranges
	if summary.MinSimilarity < 0.0 || summary.MinSimilarity > 1.0 {
		t.Errorf("MinSimilarity = %f, want [0.0, 1.0]", summary.MinSimilarity)
	}
	if summary.MaxSimilarity < 0.0 || summary.MaxSimilarity > 1.0 {
		t.Errorf("MaxSimilarity = %f, want [0.0, 1.0]", summary.MaxSimilarity)
	}
	if summary.AvgSimilarity < 0.0 || summary.AvgSimilarity > 1.0 {
		t.Errorf("AvgSimilarity = %f, want [0.0, 1.0]", summary.AvgSimilarity)
	}

	// Min should be <= Max
	if summary.MinSimilarity > summary.MaxSimilarity {
		t.Errorf("MinSimilarity (%f) > MaxSimilarity (%f)", summary.MinSimilarity, summary.MaxSimilarity)
	}
}

// BenchmarkGenreValidation benchmarks genre validation performance.
func BenchmarkGenreValidation(b *testing.B) {
	validator := NewGenreValidator(0.3)

	// Add 5 genres
	genres := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapocalyptic"}
	colors := []color.RGBA{
		{255, 0, 0, 255},
		{0, 0, 255, 255},
		{100, 0, 100, 255},
		{255, 0, 255, 255},
		{150, 75, 0, 255},
	}

	for i, genreID := range genres {
		validator.AddGenreSnapshot(genreID, &Snapshot{
			GenreID:      genreID,
			SpriteImage:  CreateTestImage(28, 28, colors[i]),
			TileImage:    CreateTestImage(32, 32, colors[i]),
			PaletteImage: CreateTestImage(16, 16, colors[i]),
		})
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = validator.Validate()
	}
}

// BenchmarkColorSimilarity benchmarks color similarity calculation.
func BenchmarkColorSimilarity(b *testing.B) {
	c1 := color.RGBA{255, 128, 64, 255}
	c2 := color.RGBA{250, 120, 60, 255}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = colorSimilarity(c1, c2)
	}
}
