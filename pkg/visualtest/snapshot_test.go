package visualtest

import (
	"image"
	"image/color"
	"os"
	"path/filepath"
	"testing"
)

// TestHashImage tests image hashing.
func TestHashImage(t *testing.T) {
	tests := []struct {
		name     string
		img1     *image.RGBA
		img2     *image.RGBA
		wantSame bool
	}{
		{
			name:     "identical images",
			img1:     CreateTestImage(10, 10, color.RGBA{255, 0, 0, 255}),
			img2:     CreateTestImage(10, 10, color.RGBA{255, 0, 0, 255}),
			wantSame: true,
		},
		{
			name:     "different colors",
			img1:     CreateTestImage(10, 10, color.RGBA{255, 0, 0, 255}),
			img2:     CreateTestImage(10, 10, color.RGBA{0, 255, 0, 255}),
			wantSame: false,
		},
		{
			name:     "different sizes",
			img1:     CreateTestImage(10, 10, color.RGBA{255, 0, 0, 255}),
			img2:     CreateTestImage(20, 20, color.RGBA{255, 0, 0, 255}),
			wantSame: false,
		},
		{
			name:     "both nil",
			img1:     nil,
			img2:     nil,
			wantSame: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash1 := hashImage(tt.img1)
			hash2 := hashImage(tt.img2)

			same := (hash1 == hash2)
			if same != tt.wantSame {
				t.Errorf("Hash comparison = %v, want %v (hash1=%s, hash2=%s)",
					same, tt.wantSame, hash1[:16], hash2[:16])
			}
		})
	}
}

// TestCalculateSimilarity tests perceptual similarity calculation.
func TestCalculateSimilarity(t *testing.T) {
	tests := []struct {
		name            string
		img1            *image.RGBA
		img2            *image.RGBA
		minSimilarity   float64
		maxSimilarity   float64
		expectIdentical bool
		expectDifferent bool
	}{
		{
			name:            "identical images",
			img1:            CreateTestImage(10, 10, color.RGBA{255, 0, 0, 255}),
			img2:            CreateTestImage(10, 10, color.RGBA{255, 0, 0, 255}),
			expectIdentical: true,
			minSimilarity:   1.0,
			maxSimilarity:   1.0,
		},
		{
			name:            "completely different",
			img1:            CreateTestImage(10, 10, color.RGBA{255, 255, 255, 255}),
			img2:            CreateTestImage(10, 10, color.RGBA{0, 0, 0, 255}),
			expectDifferent: true,
			minSimilarity:   0.0,
			maxSimilarity:   0.5,
		},
		{
			name:          "slightly different",
			img1:          CreateTestImage(10, 10, color.RGBA{255, 0, 0, 255}),
			img2:          CreateTestImage(10, 10, color.RGBA{254, 0, 0, 255}),
			minSimilarity: 0.99,
			maxSimilarity: 1.0,
		},
		{
			name:            "different sizes",
			img1:            CreateTestImage(10, 10, color.RGBA{255, 0, 0, 255}),
			img2:            CreateTestImage(20, 20, color.RGBA{255, 0, 0, 255}),
			expectDifferent: true,
			minSimilarity:   0.0,
			maxSimilarity:   0.0,
		},
		{
			name:            "nil images",
			img1:            nil,
			img2:            nil,
			expectIdentical: true,
			minSimilarity:   1.0,
			maxSimilarity:   1.0,
		},
		{
			name:            "one nil",
			img1:            CreateTestImage(10, 10, color.RGBA{255, 0, 0, 255}),
			img2:            nil,
			expectDifferent: true,
			minSimilarity:   0.0,
			maxSimilarity:   0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			similarity := calculateSimilarity(tt.img1, tt.img2)

			if similarity < tt.minSimilarity || similarity > tt.maxSimilarity {
				t.Errorf("Similarity = %f, want range [%f, %f]",
					similarity, tt.minSimilarity, tt.maxSimilarity)
			}

			if tt.expectIdentical && similarity != 1.0 {
				t.Errorf("Expected identical (1.0), got %f", similarity)
			}

			if tt.expectDifferent && similarity >= 0.95 {
				t.Errorf("Expected different (<0.95), got %f", similarity)
			}
		})
	}
}

// TestCompare tests snapshot comparison.
func TestCompare(t *testing.T) {
	options := DefaultOptions()

	tests := []struct {
		name       string
		baseline   *Snapshot
		current    *Snapshot
		wantPassed bool
		wantDiffs  int
	}{
		{
			name: "identical snapshots",
			baseline: &Snapshot{
				Seed:         12345,
				GenreID:      "fantasy",
				SpriteImage:  CreateTestImage(28, 28, color.RGBA{255, 0, 0, 255}),
				TileImage:    CreateTestImage(32, 32, color.RGBA{0, 255, 0, 255}),
				PaletteImage: CreateTestImage(16, 16, color.RGBA{0, 0, 255, 255}),
			},
			current: &Snapshot{
				Seed:         12345,
				GenreID:      "fantasy",
				SpriteImage:  CreateTestImage(28, 28, color.RGBA{255, 0, 0, 255}),
				TileImage:    CreateTestImage(32, 32, color.RGBA{0, 255, 0, 255}),
				PaletteImage: CreateTestImage(16, 16, color.RGBA{0, 0, 255, 255}),
			},
			wantPassed: true,
			wantDiffs:  0,
		},
		{
			name: "sprite regression",
			baseline: &Snapshot{
				Seed:        12345,
				GenreID:     "fantasy",
				SpriteImage: CreateTestImage(28, 28, color.RGBA{255, 0, 0, 255}),
			},
			current: &Snapshot{
				Seed:        12345,
				GenreID:     "fantasy",
				SpriteImage: CreateTestImage(28, 28, color.RGBA{0, 255, 0, 255}),
			},
			wantPassed: false,
			wantDiffs:  1,
		},
		{
			name: "multiple regressions",
			baseline: &Snapshot{
				Seed:        12345,
				GenreID:     "fantasy",
				SpriteImage: CreateTestImage(28, 28, color.RGBA{255, 0, 0, 255}),
				TileImage:   CreateTestImage(32, 32, color.RGBA{0, 255, 0, 255}),
			},
			current: &Snapshot{
				Seed:        12345,
				GenreID:     "fantasy",
				SpriteImage: CreateTestImage(28, 28, color.RGBA{0, 0, 255, 255}),
				TileImage:   CreateTestImage(32, 32, color.RGBA{255, 255, 0, 255}),
			},
			wantPassed: false,
			wantDiffs:  2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Update hashes
			tt.baseline.SpriteHash = hashImage(tt.baseline.SpriteImage)
			tt.baseline.TileHash = hashImage(tt.baseline.TileImage)
			tt.baseline.PaletteHash = hashImage(tt.baseline.PaletteImage)
			tt.current.SpriteHash = hashImage(tt.current.SpriteImage)
			tt.current.TileHash = hashImage(tt.current.TileImage)
			tt.current.PaletteHash = hashImage(tt.current.PaletteImage)

			result := Compare(tt.baseline, tt.current, options)

			if result.Passed != tt.wantPassed {
				t.Errorf("Passed = %v, want %v", result.Passed, tt.wantPassed)
			}

			if len(result.Differences) != tt.wantDiffs {
				t.Errorf("Differences count = %d, want %d", len(result.Differences), tt.wantDiffs)
			}

			// Verify metrics are in valid range
			if result.Metrics.OverallSimilarity < 0.0 || result.Metrics.OverallSimilarity > 1.0 {
				t.Errorf("OverallSimilarity = %f, want [0.0, 1.0]", result.Metrics.OverallSimilarity)
			}
		})
	}
}

// TestSaveAndLoadSnapshot tests snapshot persistence.
func TestSaveAndLoadSnapshot(t *testing.T) {
	// Create temporary directory
	tmpDir := t.TempDir()

	options := SnapshotOptions{
		SaveImages:          true,
		OutputDir:           tmpDir,
		SimilarityThreshold: 0.99,
	}

	// Create test snapshot
	original := &Snapshot{
		Seed:         12345,
		GenreID:      "fantasy",
		SpriteImage:  CreateTestImage(28, 28, color.RGBA{255, 0, 0, 255}),
		TileImage:    CreateTestImage(32, 32, color.RGBA{0, 255, 0, 255}),
		PaletteImage: CreateTestImage(16, 16, color.RGBA{0, 0, 255, 255}),
	}
	original.SpriteHash = hashImage(original.SpriteImage)
	original.TileHash = hashImage(original.TileImage)
	original.PaletteHash = hashImage(original.PaletteImage)

	// Save snapshot
	err := SaveSnapshot(original, options)
	if err != nil {
		t.Fatalf("SaveSnapshot failed: %v", err)
	}

	// Verify files exist
	spriteFile := filepath.Join(tmpDir, "sprite_fantasy_12345.png")
	if _, err := os.Stat(spriteFile); os.IsNotExist(err) {
		t.Error("Sprite image file not created")
	}

	// Load snapshot
	loaded, err := LoadSnapshot("fantasy", 12345, options)
	if err != nil {
		t.Fatalf("LoadSnapshot failed: %v", err)
	}

	// Verify loaded data matches original
	if loaded.SpriteHash != original.SpriteHash {
		t.Errorf("Loaded sprite hash doesn't match original")
	}
	if loaded.TileHash != original.TileHash {
		t.Errorf("Loaded tile hash doesn't match original")
	}
	if loaded.PaletteHash != original.PaletteHash {
		t.Errorf("Loaded palette hash doesn't match original")
	}
}

// TestGetSeverity tests severity categorization.
func TestGetSeverity(t *testing.T) {
	tests := []struct {
		similarity float64
		want       string
	}{
		{1.0, "minor"},
		{0.99, "minor"},
		{0.95, "minor"},
		{0.94, "major"},
		{0.90, "major"},
		{0.85, "major"},
		{0.84, "critical"},
		{0.50, "critical"},
		{0.0, "critical"},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			got := getSeverity(tt.similarity)
			if got != tt.want {
				t.Errorf("getSeverity(%f) = %s, want %s", tt.similarity, got, tt.want)
			}
		})
	}
}

// TestDefaultOptions tests default option values.
func TestDefaultOptions(t *testing.T) {
	opts := DefaultOptions()

	if opts.SaveImages {
		t.Error("Expected SaveImages = false by default")
	}

	if opts.SimilarityThreshold != 0.99 {
		t.Errorf("SimilarityThreshold = %f, want 0.99", opts.SimilarityThreshold)
	}

	if opts.OutputDir != "testdata/visual_snapshots" {
		t.Errorf("OutputDir = %s, want testdata/visual_snapshots", opts.OutputDir)
	}
}

// TestCreateTestImage tests test image creation.
func TestCreateTestImage(t *testing.T) {
	img := CreateTestImage(100, 50, color.RGBA{128, 64, 32, 255})

	if img == nil {
		t.Fatal("CreateTestImage returned nil")
	}

	bounds := img.Bounds()
	if bounds.Dx() != 100 {
		t.Errorf("Width = %d, want 100", bounds.Dx())
	}
	if bounds.Dy() != 50 {
		t.Errorf("Height = %d, want 50", bounds.Dy())
	}

	// Check a sample pixel
	r, g, b, a := img.At(50, 25).RGBA()
	if r>>8 != 128 || g>>8 != 64 || b>>8 != 32 || a>>8 != 255 {
		t.Errorf("Pixel color = (%d, %d, %d, %d), want (128, 64, 32, 255)",
			r>>8, g>>8, b>>8, a>>8)
	}
}

// TestSnapshotHashConsistency tests that hashing is deterministic.
func TestSnapshotHashConsistency(t *testing.T) {
	img := CreateTestImage(50, 50, color.RGBA{100, 150, 200, 255})

	// Hash same image multiple times
	hash1 := hashImage(img)
	hash2 := hashImage(img)
	hash3 := hashImage(img)

	if hash1 != hash2 || hash2 != hash3 {
		t.Errorf("Hash inconsistency: %s, %s, %s", hash1[:16], hash2[:16], hash3[:16])
	}
}

// TestCompareMetrics tests that similarity metrics are calculated correctly.
func TestCompareMetrics(t *testing.T) {
	options := DefaultOptions()

	baseline := &Snapshot{
		SpriteImage:  CreateTestImage(10, 10, color.RGBA{255, 0, 0, 255}),
		TileImage:    CreateTestImage(10, 10, color.RGBA{0, 255, 0, 255}),
		PaletteImage: CreateTestImage(10, 10, color.RGBA{0, 0, 255, 255}),
	}
	baseline.SpriteHash = hashImage(baseline.SpriteImage)
	baseline.TileHash = hashImage(baseline.TileImage)
	baseline.PaletteHash = hashImage(baseline.PaletteImage)

	current := &Snapshot{
		SpriteImage:  CreateTestImage(10, 10, color.RGBA{255, 0, 0, 255}), // Identical
		TileImage:    CreateTestImage(10, 10, color.RGBA{0, 255, 0, 255}), // Identical
		PaletteImage: CreateTestImage(10, 10, color.RGBA{0, 0, 255, 255}), // Identical
	}
	current.SpriteHash = hashImage(current.SpriteImage)
	current.TileHash = hashImage(current.TileImage)
	current.PaletteHash = hashImage(current.PaletteImage)

	result := Compare(baseline, current, options)

	// All should be 1.0 (identical)
	if result.Metrics.SpriteSimilarity != 1.0 {
		t.Errorf("SpriteSimilarity = %f, want 1.0", result.Metrics.SpriteSimilarity)
	}
	if result.Metrics.TileSimilarity != 1.0 {
		t.Errorf("TileSimilarity = %f, want 1.0", result.Metrics.TileSimilarity)
	}
	if result.Metrics.PaletteSimilarity != 1.0 {
		t.Errorf("PaletteSimilarity = %f, want 1.0", result.Metrics.PaletteSimilarity)
	}
	if result.Metrics.OverallSimilarity != 1.0 {
		t.Errorf("OverallSimilarity = %f, want 1.0", result.Metrics.OverallSimilarity)
	}
}

// BenchmarkHashImage benchmarks image hashing performance.
func BenchmarkHashImage(b *testing.B) {
	img := CreateTestImage(100, 100, color.RGBA{128, 128, 128, 255})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = hashImage(img)
	}
}

// BenchmarkCalculateSimilarity benchmarks similarity calculation.
func BenchmarkCalculateSimilarity(b *testing.B) {
	img1 := CreateTestImage(100, 100, color.RGBA{128, 128, 128, 255})
	img2 := CreateTestImage(100, 100, color.RGBA{130, 130, 130, 255})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = calculateSimilarity(img1, img2)
	}
}

// BenchmarkCompare benchmarks full snapshot comparison.
func BenchmarkCompare(b *testing.B) {
	options := DefaultOptions()

	baseline := &Snapshot{
		SpriteImage:  CreateTestImage(28, 28, color.RGBA{255, 0, 0, 255}),
		TileImage:    CreateTestImage(32, 32, color.RGBA{0, 255, 0, 255}),
		PaletteImage: CreateTestImage(16, 16, color.RGBA{0, 0, 255, 255}),
	}
	baseline.SpriteHash = hashImage(baseline.SpriteImage)
	baseline.TileHash = hashImage(baseline.TileImage)
	baseline.PaletteHash = hashImage(baseline.PaletteImage)

	current := &Snapshot{
		SpriteImage:  CreateTestImage(28, 28, color.RGBA{255, 0, 0, 255}),
		TileImage:    CreateTestImage(32, 32, color.RGBA{0, 255, 0, 255}),
		PaletteImage: CreateTestImage(16, 16, color.RGBA{0, 0, 255, 255}),
	}
	current.SpriteHash = hashImage(current.SpriteImage)
	current.TileHash = hashImage(current.TileImage)
	current.PaletteHash = hashImage(current.PaletteImage)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Compare(baseline, current, options)
	}
}
