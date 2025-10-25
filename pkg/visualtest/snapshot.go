// Package visualtest provides visual regression testing for procedurally generated content.package visualtest

// This package enables establishing visual baselines and detecting unintended changes
// in sprite generation, color palettes, tile rendering, and animation sequences.
//
// Visual regression testing ensures that:
// - Procedural generation remains deterministic (same seed = same output)
// - Code changes don't introduce unintended visual artifacts
// - Genre-specific styles remain consistent and distinct
// - Performance characteristics don't degrade
//
// Usage:
//
//	// Create a snapshot of current visuals
//	snapshot := visualtest.CaptureSnapshot(seed, genreID)
//
//	// Compare against baseline
//	result := visualtest.Compare(baseline, snapshot)
//	if !result.Passed {
//	    for _, diff := range result.Differences {
//	        log.Printf("Visual regression: %s", diff.Description)
//	    }
//	}
//
// The package uses perceptual hashing to detect visual differences while allowing
// for minor rendering variations (anti-aliasing, floating-point precision, etc.).
package visualtest

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
)

// Snapshot represents captured visual output at a specific point in time.
type Snapshot struct {
	// Generation parameters
	Seed    int64  `json:"seed"`
	GenreID string `json:"genre_id"`

	// Visual hashes (for quick comparison)
	SpriteHash  string `json:"sprite_hash"`
	TileHash    string `json:"tile_hash"`
	PaletteHash string `json:"palette_hash"`

	// Metadata
	Timestamp string `json:"timestamp"`
	Version   string `json:"version"` // Code version/commit

	// Images (optional, for detailed comparison)
	SpriteImage  *image.RGBA `json:"-"`
	TileImage    *image.RGBA `json:"-"`
	PaletteImage *image.RGBA `json:"-"`
}

// ComparisonResult contains the results of comparing two snapshots.
type ComparisonResult struct {
	Passed      bool         `json:"passed"`
	Differences []Difference `json:"differences,omitempty"`
	Metrics     Metrics      `json:"metrics"`
}

// Difference describes a visual regression detected between snapshots.
type Difference struct {
	Type        string  `json:"type"`        // "sprite", "tile", "palette", "animation"
	Description string  `json:"description"` // Human-readable description
	Severity    string  `json:"severity"`    // "critical", "major", "minor"
	Similarity  float64 `json:"similarity"`  // 0.0 (completely different) to 1.0 (identical)
}

// Metrics contains quantitative comparison metrics.
type Metrics struct {
	SpriteSimilarity  float64 `json:"sprite_similarity"`
	TileSimilarity    float64 `json:"tile_similarity"`
	PaletteSimilarity float64 `json:"palette_similarity"`
	OverallSimilarity float64 `json:"overall_similarity"`
}

// SnapshotOptions configures snapshot capture behavior.
type SnapshotOptions struct {
	// SaveImages controls whether to save full images (for detailed comparison)
	SaveImages bool

	// OutputDir specifies where to save snapshot images
	OutputDir string

	// Threshold for considering snapshots identical (0.0-1.0)
	SimilarityThreshold float64
}

// DefaultOptions returns default snapshot options.
func DefaultOptions() SnapshotOptions {
	return SnapshotOptions{
		SaveImages:          false, // Don't save images by default (use hashes)
		OutputDir:           "testdata/visual_snapshots",
		SimilarityThreshold: 0.99, // 99% similarity required
	}
}

// hashImage computes a SHA-256 hash of an image.
func hashImage(img *image.RGBA) string {
	if img == nil {
		return ""
	}

	hasher := sha256.New()
	bounds := img.Bounds()

	// Hash pixel data
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, a := img.At(x, y).RGBA()
			// Use 8-bit values for consistency
			hasher.Write([]byte{byte(r >> 8), byte(g >> 8), byte(b >> 8), byte(a >> 8)})
		}
	}

	return hex.EncodeToString(hasher.Sum(nil))
}

// calculateSimilarity computes perceptual similarity between two images.
// Returns a value from 0.0 (completely different) to 1.0 (identical).
func calculateSimilarity(img1, img2 *image.RGBA) float64 {
	if img1 == nil || img2 == nil {
		if img1 == nil && img2 == nil {
			return 1.0 // Both nil = identical
		}
		return 0.0 // One nil = completely different
	}

	bounds1 := img1.Bounds()
	bounds2 := img2.Bounds()

	// Different sizes = different images
	if bounds1.Dx() != bounds2.Dx() || bounds1.Dy() != bounds2.Dy() {
		return 0.0
	}

	var totalDiff float64
	var pixelCount int

	for y := 0; y < bounds1.Dy(); y++ {
		for x := 0; x < bounds1.Dx(); x++ {
			r1, g1, b1, a1 := img1.At(bounds1.Min.X+x, bounds1.Min.Y+y).RGBA()
			r2, g2, b2, a2 := img2.At(bounds2.Min.X+x, bounds2.Min.Y+y).RGBA()

			// Calculate color difference (Euclidean distance in RGBA space)
			dr := float64(r1) - float64(r2)
			dg := float64(g1) - float64(g2)
			db := float64(b1) - float64(b2)
			da := float64(a1) - float64(a2)

			diff := (dr*dr + dg*dg + db*db + da*da) / (65535.0 * 65535.0 * 4.0)
			totalDiff += diff
			pixelCount++
		}
	}

	if pixelCount == 0 {
		return 1.0
	}

	// Convert average difference to similarity
	avgDiff := totalDiff / float64(pixelCount)
	similarity := 1.0 - avgDiff

	return similarity
}

// SaveSnapshot saves a snapshot to disk (images and metadata).
func SaveSnapshot(snapshot *Snapshot, options SnapshotOptions) error {
	if !options.SaveImages {
		return nil // Nothing to save if images not enabled
	}

	// Create output directory
	if err := os.MkdirAll(options.OutputDir, 0o755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Save sprite image
	if snapshot.SpriteImage != nil {
		path := filepath.Join(options.OutputDir, fmt.Sprintf("sprite_%s_%d.png", snapshot.GenreID, snapshot.Seed))
		if err := saveImage(snapshot.SpriteImage, path); err != nil {
			return fmt.Errorf("failed to save sprite image: %w", err)
		}
	}

	// Save tile image
	if snapshot.TileImage != nil {
		path := filepath.Join(options.OutputDir, fmt.Sprintf("tile_%s_%d.png", snapshot.GenreID, snapshot.Seed))
		if err := saveImage(snapshot.TileImage, path); err != nil {
			return fmt.Errorf("failed to save tile image: %w", err)
		}
	}

	// Save palette image
	if snapshot.PaletteImage != nil {
		path := filepath.Join(options.OutputDir, fmt.Sprintf("palette_%s_%d.png", snapshot.GenreID, snapshot.Seed))
		if err := saveImage(snapshot.PaletteImage, path); err != nil {
			return fmt.Errorf("failed to save palette image: %w", err)
		}
	}

	return nil
}

// saveImage saves an image to a PNG file.
func saveImage(img *image.RGBA, path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	return png.Encode(file, img)
}

// LoadSnapshot loads a snapshot from disk.
func LoadSnapshot(genreID string, seed int64, options SnapshotOptions) (*Snapshot, error) {
	snapshot := &Snapshot{
		Seed:    seed,
		GenreID: genreID,
	}

	// Load sprite image
	spritePath := filepath.Join(options.OutputDir, fmt.Sprintf("sprite_%s_%d.png", genreID, seed))
	spriteImg, err := loadImage(spritePath)
	if err == nil {
		snapshot.SpriteImage = spriteImg
		snapshot.SpriteHash = hashImage(spriteImg)
	}

	// Load tile image
	tilePath := filepath.Join(options.OutputDir, fmt.Sprintf("tile_%s_%d.png", genreID, seed))
	tileImg, err := loadImage(tilePath)
	if err == nil {
		snapshot.TileImage = tileImg
		snapshot.TileHash = hashImage(tileImg)
	}

	// Load palette image
	palettePath := filepath.Join(options.OutputDir, fmt.Sprintf("palette_%s_%d.png", genreID, seed))
	paletteImg, err := loadImage(palettePath)
	if err == nil {
		snapshot.PaletteImage = paletteImg
		snapshot.PaletteHash = hashImage(paletteImg)
	}

	return snapshot, nil
}

// loadImage loads an image from a PNG file.
func loadImage(path string) (*image.RGBA, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	img, err := png.Decode(file)
	if err != nil {
		return nil, err
	}

	// Convert to RGBA
	bounds := img.Bounds()
	rgba := image.NewRGBA(bounds)
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			rgba.Set(x, y, img.At(x, y))
		}
	}

	return rgba, nil
}

// Compare compares two snapshots and returns detailed results.
func Compare(baseline, current *Snapshot, options SnapshotOptions) ComparisonResult {
	result := ComparisonResult{
		Passed:      true,
		Differences: []Difference{},
	}

	// Compare sprite hashes first (fast path)
	if baseline.SpriteHash != current.SpriteHash {
		// Hashes differ, compute detailed similarity
		similarity := calculateSimilarity(baseline.SpriteImage, current.SpriteImage)
		result.Metrics.SpriteSimilarity = similarity

		if similarity < options.SimilarityThreshold {
			result.Passed = false
			result.Differences = append(result.Differences, Difference{
				Type:        "sprite",
				Description: fmt.Sprintf("Sprite visual regression detected (%.2f%% similar)", similarity*100),
				Severity:    getSeverity(similarity),
				Similarity:  similarity,
			})
		}
	} else {
		result.Metrics.SpriteSimilarity = 1.0
	}

	// Compare tile hashes
	if baseline.TileHash != current.TileHash {
		similarity := calculateSimilarity(baseline.TileImage, current.TileImage)
		result.Metrics.TileSimilarity = similarity

		if similarity < options.SimilarityThreshold {
			result.Passed = false
			result.Differences = append(result.Differences, Difference{
				Type:        "tile",
				Description: fmt.Sprintf("Tile visual regression detected (%.2f%% similar)", similarity*100),
				Severity:    getSeverity(similarity),
				Similarity:  similarity,
			})
		}
	} else {
		result.Metrics.TileSimilarity = 1.0
	}

	// Compare palette hashes
	if baseline.PaletteHash != current.PaletteHash {
		similarity := calculateSimilarity(baseline.PaletteImage, current.PaletteImage)
		result.Metrics.PaletteSimilarity = similarity

		if similarity < options.SimilarityThreshold {
			result.Passed = false
			result.Differences = append(result.Differences, Difference{
				Type:        "palette",
				Description: fmt.Sprintf("Palette visual regression detected (%.2f%% similar)", similarity*100),
				Severity:    getSeverity(similarity),
				Similarity:  similarity,
			})
		}
	} else {
		result.Metrics.PaletteSimilarity = 1.0
	}

	// Calculate overall similarity
	result.Metrics.OverallSimilarity = (result.Metrics.SpriteSimilarity +
		result.Metrics.TileSimilarity +
		result.Metrics.PaletteSimilarity) / 3.0

	return result
}

// getSeverity categorizes visual differences by severity.
func getSeverity(similarity float64) string {
	if similarity >= 0.95 {
		return "minor" // >95% similar
	} else if similarity >= 0.85 {
		return "major" // 85-95% similar
	}
	return "critical" // <85% similar
}

// CreateTestImage creates a simple test image for testing.
func CreateTestImage(width, height int, fillColor color.Color) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, fillColor)
		}
	}
	return img
}
