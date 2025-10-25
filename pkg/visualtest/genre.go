package visualtest

import (
	"fmt"
	"image"
	"image/color"
	"math"
)

// GenreValidator validates that genres maintain distinct visual characteristics.
type GenreValidator struct {
	snapshots map[string]*Snapshot // genreID -> snapshot
	threshold float64              // Minimum distinctness threshold (0.0-1.0)
}

// NewGenreValidator creates a new genre validator.
func NewGenreValidator(distinctnessThreshold float64) *GenreValidator {
	return &GenreValidator{
		snapshots: make(map[string]*Snapshot),
		threshold: distinctnessThreshold,
	}
}

// AddGenreSnapshot adds a snapshot for a specific genre.
func (gv *GenreValidator) AddGenreSnapshot(genreID string, snapshot *Snapshot) {
	gv.snapshots[genreID] = snapshot
}

// GenreValidationResult contains validation results for genre distinctness.
type GenreValidationResult struct {
	Passed      bool                   `json:"passed"`
	Issues      []GenreIssue           `json:"issues,omitempty"`
	Comparisons []GenreComparison      `json:"comparisons"`
	Summary     GenreValidationSummary `json:"summary"`
}

// GenreIssue describes a problem with genre distinctness.
type GenreIssue struct {
	GenreA      string  `json:"genre_a"`
	GenreB      string  `json:"genre_b"`
	Similarity  float64 `json:"similarity"`
	Description string  `json:"description"`
	Severity    string  `json:"severity"` // "critical", "warning"
}

// GenreComparison contains similarity metrics between two genres.
type GenreComparison struct {
	GenreA               string  `json:"genre_a"`
	GenreB               string  `json:"genre_b"`
	SpriteSimilarity     float64 `json:"sprite_similarity"`
	TileSimilarity       float64 `json:"tile_similarity"`
	PaletteSimilarity    float64 `json:"palette_similarity"`
	OverallSimilarity    float64 `json:"overall_similarity"`
	SufficientlyDistinct bool    `json:"sufficiently_distinct"`
}

// GenreValidationSummary provides aggregate validation metrics.
type GenreValidationSummary struct {
	TotalGenres       int     `json:"total_genres"`
	TotalComparisons  int     `json:"total_comparisons"`
	PassedComparisons int     `json:"passed_comparisons"`
	FailedComparisons int     `json:"failed_comparisons"`
	AvgSimilarity     float64 `json:"avg_similarity"`
	MinSimilarity     float64 `json:"min_similarity"`
	MaxSimilarity     float64 `json:"max_similarity"`
}

// Validate performs genre distinctness validation.
func (gv *GenreValidator) Validate() GenreValidationResult {
	result := GenreValidationResult{
		Passed:      true,
		Issues:      []GenreIssue{},
		Comparisons: []GenreComparison{},
	}

	genreList := make([]string, 0, len(gv.snapshots))
	for genreID := range gv.snapshots {
		genreList = append(genreList, genreID)
	}

	result.Summary.TotalGenres = len(genreList)

	var totalSimilarity float64
	minSim := 1.0
	maxSim := 0.0

	// Compare each pair of genres
	for i := 0; i < len(genreList); i++ {
		for j := i + 1; j < len(genreList); j++ {
			genreA := genreList[i]
			genreB := genreList[j]

			comparison := gv.compareGenres(genreA, genreB)
			result.Comparisons = append(result.Comparisons, comparison)
			result.Summary.TotalComparisons++

			totalSimilarity += comparison.OverallSimilarity

			if comparison.OverallSimilarity < minSim {
				minSim = comparison.OverallSimilarity
			}
			if comparison.OverallSimilarity > maxSim {
				maxSim = comparison.OverallSimilarity
			}

			if !comparison.SufficientlyDistinct {
				result.Passed = false
				result.Summary.FailedComparisons++

				// Add issue
				severity := "warning"
				if comparison.OverallSimilarity > 0.85 {
					severity = "critical"
				}

				result.Issues = append(result.Issues, GenreIssue{
					GenreA:     genreA,
					GenreB:     genreB,
					Similarity: comparison.OverallSimilarity,
					Description: fmt.Sprintf("Genres '%s' and '%s' are too similar (%.1f%% similarity, threshold: %.1f%%)",
						genreA, genreB, comparison.OverallSimilarity*100, (1.0-gv.threshold)*100),
					Severity: severity,
				})
			} else {
				result.Summary.PassedComparisons++
			}
		}
	}

	// Calculate summary statistics
	if result.Summary.TotalComparisons > 0 {
		result.Summary.AvgSimilarity = totalSimilarity / float64(result.Summary.TotalComparisons)
		result.Summary.MinSimilarity = minSim
		result.Summary.MaxSimilarity = maxSim
	}

	return result
}

// compareGenres compares two genres for distinctness.
func (gv *GenreValidator) compareGenres(genreA, genreB string) GenreComparison {
	snapA := gv.snapshots[genreA]
	snapB := gv.snapshots[genreB]

	spriteSim := calculateSimilarity(snapA.SpriteImage, snapB.SpriteImage)
	tileSim := calculateSimilarity(snapA.TileImage, snapB.TileImage)
	paletteSim := calculatePaletteSimilarity(snapA.PaletteImage, snapB.PaletteImage)

	overallSim := (spriteSim + tileSim + paletteSim) / 3.0

	// Genres should be distinct (similarity < threshold means distinct)
	distinct := overallSim < (1.0 - gv.threshold)

	return GenreComparison{
		GenreA:               genreA,
		GenreB:               genreB,
		SpriteSimilarity:     spriteSim,
		TileSimilarity:       tileSim,
		PaletteSimilarity:    paletteSim,
		OverallSimilarity:    overallSim,
		SufficientlyDistinct: distinct,
	}
}

// calculatePaletteSimilarity computes color palette similarity.
// This uses a specialized algorithm for palette comparison.
func calculatePaletteSimilarity(palette1, palette2 *image.RGBA) float64 {
	if palette1 == nil || palette2 == nil {
		if palette1 == nil && palette2 == nil {
			return 1.0
		}
		return 0.0
	}

	// Extract dominant colors from each palette
	colors1 := extractDominantColors(palette1, 8)
	colors2 := extractDominantColors(palette2, 8)

	if len(colors1) == 0 || len(colors2) == 0 {
		return 0.0
	}

	// Calculate color space similarity
	totalSimilarity := 0.0
	matchCount := 0

	for _, c1 := range colors1 {
		bestMatch := 0.0
		for _, c2 := range colors2 {
			sim := colorSimilarity(c1, c2)
			if sim > bestMatch {
				bestMatch = sim
			}
		}
		totalSimilarity += bestMatch
		matchCount++
	}

	if matchCount == 0 {
		return 0.0
	}

	return totalSimilarity / float64(matchCount)
}

// extractDominantColors extracts the N most dominant colors from an image.
func extractDominantColors(img *image.RGBA, count int) []color.RGBA {
	if img == nil {
		return nil
	}

	// Simple sampling: extract colors from grid points
	bounds := img.Bounds()
	colors := make([]color.RGBA, 0, count)

	stepX := bounds.Dx() / count
	stepY := bounds.Dy() / count
	if stepX < 1 {
		stepX = 1
	}
	if stepY < 1 {
		stepY = 1
	}

	for y := bounds.Min.Y; y < bounds.Max.Y && len(colors) < count; y += stepY {
		for x := bounds.Min.X; x < bounds.Max.X && len(colors) < count; x += stepX {
			r, g, b, a := img.At(x, y).RGBA()
			colors = append(colors, color.RGBA{
				R: uint8(r >> 8),
				G: uint8(g >> 8),
				B: uint8(b >> 8),
				A: uint8(a >> 8),
			})
		}
	}

	return colors
}

// colorSimilarity computes similarity between two colors (0.0-1.0).
func colorSimilarity(c1, c2 color.RGBA) float64 {
	// Convert to LAB color space for perceptual similarity
	// Simplified: use weighted Euclidean distance in RGB
	dr := float64(c1.R) - float64(c2.R)
	dg := float64(c1.G) - float64(c2.G)
	db := float64(c1.B) - float64(c2.B)

	// Weighted by human perception (green > red > blue)
	distance := math.Sqrt(2*dr*dr + 4*dg*dg + 3*db*db)

	// Normalize (max distance ~ 764 for weighted RGB)
	maxDistance := math.Sqrt(2*255*255 + 4*255*255 + 3*255*255)
	similarity := 1.0 - (distance / maxDistance)

	return similarity
}

// ValidateGenreSet validates a complete set of genre snapshots.
func ValidateGenreSet(genreSnapshots map[string]*Snapshot, distinctnessThreshold float64) GenreValidationResult {
	validator := NewGenreValidator(distinctnessThreshold)

	for genreID, snapshot := range genreSnapshots {
		validator.AddGenreSnapshot(genreID, snapshot)
	}

	return validator.Validate()
}
