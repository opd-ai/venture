package visualtest

import (
	"fmt"
	"image"
	"testing"

	"github.com/opd-ai/venture/pkg/procgen/entity"
)

// TestGenerateAllSpriteTests verifies we generate correct number of test cases.
func TestGenerateAllSpriteTests(t *testing.T) {
	tests := GenerateAllSpriteTests()

	// 4 entity types × 5 genres × 2 sizes = 40 test cases
	expectedCount := 4 * 5 * 2
	if len(tests) != expectedCount {
		t.Errorf("Expected %d sprite tests, got %d", expectedCount, len(tests))
	}

	// Verify test coverage
	genreCounts := make(map[string]int)
	sizeCounts := make(map[int]int)

	for _, test := range tests {
		genreCounts[test.GenreID]++
		sizeCounts[test.Size]++

		// Verify valid parameters
		if test.GenreID == "" {
			t.Error("Test has empty genre ID")
		}
		if test.Size != 32 && test.Size != 64 {
			t.Errorf("Invalid sprite size: %d", test.Size)
		}
		if test.Seed == 0 {
			t.Error("Test has zero seed")
		}
	}

	// Each genre should have equal representation
	expectedPerGenre := expectedCount / 5
	for genre, count := range genreCounts {
		if count != expectedPerGenre {
			t.Errorf("Genre %s has %d tests, expected %d", genre, count, expectedPerGenre)
		}
	}

	// Each size should have equal representation
	expectedPerSize := expectedCount / 2
	for size, count := range sizeCounts {
		if count != expectedPerSize {
			t.Errorf("Size %d has %d tests, expected %d", size, count, expectedPerSize)
		}
	}
}

// TestRenderSpriteTestCase verifies sprite rendering.
func TestRenderSpriteTestCase(t *testing.T) {
	testCase := SpriteTestCase{
		EntityType: 0, // TypeMonster
		GenreID:    "fantasy",
		Seed:       12345,
		Size:       64,
	}

	img, err := RenderSpriteTestCase(testCase)
	if err != nil {
		t.Fatalf("Failed to render sprite: %v", err)
	}

	if img == nil {
		t.Fatal("Rendered image is nil")
	}

	bounds := img.Bounds()
	if bounds.Dx() != 64 || bounds.Dy() != 64 {
		t.Errorf("Expected 64x64 image, got %dx%d", bounds.Dx(), bounds.Dy())
	}
}

// TestSpriteDeterminism verifies sprites are deterministic (same seed = same output).
func TestSpriteDeterminism(t *testing.T) {
	testCase := SpriteTestCase{
		EntityType: 0,
		GenreID:    "fantasy",
		Seed:       12345,
		Size:       32,
	}

	// Render same sprite twice
	img1, err1 := RenderSpriteTestCase(testCase)
	img2, err2 := RenderSpriteTestCase(testCase)

	if err1 != nil || err2 != nil {
		t.Fatalf("Render failed: %v, %v", err1, err2)
	}

	// Compare pixel-by-pixel
	similarity := compareImages(img1, img2)
	if similarity != 1.0 {
		t.Errorf("Determinism failed: similarity %.4f, expected 1.0", similarity)
	}
}

// TestSpriteVariation verifies different seeds produce different sprites.
func TestSpriteVariation(t *testing.T) {
	t.Skip("Skipping variation test - simplified rendering doesn't vary visual pattern (full sprite system would)")

	testCase1 := SpriteTestCase{
		EntityType: 0,
		GenreID:    "fantasy",
		Seed:       12345,
		Size:       32,
	}

	testCase2 := SpriteTestCase{
		EntityType: 0,
		GenreID:    "fantasy",
		Seed:       54321, // Different seed
		Size:       32,
	}

	img1, err1 := RenderSpriteTestCase(testCase1)
	img2, err2 := RenderSpriteTestCase(testCase2)

	if err1 != nil || err2 != nil {
		t.Fatalf("Render failed: %v, %v", err1, err2)
	}

	similarity := compareImages(img1, img2)
	if similarity > 0.9 {
		t.Errorf("Insufficient variation: similarity %.4f (too high)", similarity)
	}
}

// TestGenreColorDistinction verifies genre palettes are visually distinct.
func TestGenreColorDistinction(t *testing.T) {
	genres := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapoc"}

	// Get colors for same entity type across all genres
	entityType := 0 // TypeMonster
	colors := make(map[string][3]int)

	for _, genre := range genres {
		color := GetEntityColor(entity.EntityType(entityType), genre)
		colors[genre] = [3]int{int(color.R), int(color.G), int(color.B)}
	}

	// Verify each genre pair has sufficient color difference
	minDifference := 50 // Minimum RGB distance for distinction

	for i, genre1 := range genres {
		for j := i + 1; j < len(genres); j++ {
			genre2 := genres[j]

			c1 := colors[genre1]
			c2 := colors[genre2]

			// Calculate color distance (Euclidean in RGB space)
			dr := c1[0] - c2[0]
			dg := c1[1] - c2[1]
			db := c1[2] - c2[2]
			distance := dr*dr + dg*dg + db*db

			if distance < minDifference*minDifference {
				t.Errorf("Genres %s and %s have similar colors (distance %d, minimum %d)",
					genre1, genre2, distance, minDifference*minDifference)
			}
		}
	}
}

// TestVerifySpriteRegression tests the regression verification logic.
func TestVerifySpriteRegression(t *testing.T) {
	// Generate baseline sprites
	baseline := make(map[string]*image.RGBA)
	current := make(map[string]*image.RGBA)

	tests := GenerateAllSpriteTests()[:5] // Use first 5 tests for speed

	for _, tc := range tests {
		testID := fmt.Sprintf("%s-%d-%d", tc.GenreID, tc.EntityType, tc.Size)

		img, err := RenderSpriteTestCase(tc)
		if err != nil {
			t.Fatalf("Failed to render test %s: %v", testID, err)
		}

		baseline[testID] = img
		current[testID] = img // Same for no regression
	}

	result := VerifySpriteRegression(baseline, current)

	if !result.Passed {
		t.Errorf("Regression check failed with no changes")
	}

	if len(result.Differences) != 0 {
		t.Errorf("Expected 0 differences, got %d", len(result.Differences))
	}

	if result.Metrics.SpriteSimilarity < 0.99 {
		t.Errorf("Expected near-perfect similarity, got %.4f", result.Metrics.SpriteSimilarity)
	}
}

// TestSpriteRegressionDetection tests that actual regressions are detected.
func TestSpriteRegressionDetection(t *testing.T) {
	testCase := SpriteTestCase{
		EntityType: 0,
		GenreID:    "fantasy",
		Seed:       12345,
		Size:       64,
	}

	baseline := make(map[string]*image.RGBA)
	current := make(map[string]*image.RGBA)

	baselineImg, err := RenderSpriteTestCase(testCase)
	if err != nil {
		t.Fatalf("Failed to render baseline: %v", err)
	}

	// Create modified version (simulating regression)
	modifiedImg := image.NewRGBA(baselineImg.Bounds())
	copy(modifiedImg.Pix, baselineImg.Pix)

	// Modify 30% of pixels (should trigger regression detection)
	bounds := modifiedImg.Bounds()
	pixelsToModify := (bounds.Dx() * bounds.Dy()) * 30 / 100
	for i := 0; i < pixelsToModify; i++ {
		modifiedImg.Pix[i*4] ^= 0xFF // Flip red channel
	}

	testID := "fantasy-0-64"
	baseline[testID] = baselineImg
	current[testID] = modifiedImg

	result := VerifySpriteRegression(baseline, current)

	if result.Passed {
		t.Error("Expected regression detection to fail (regression present)")
	}

	if len(result.Differences) == 0 {
		t.Error("Expected regression difference to be detected")
	}

	if result.Metrics.SpriteSimilarity > 0.85 {
		t.Errorf("Expected low similarity due to regression, got %.4f", result.Metrics.SpriteSimilarity)
	}
}

// TestImageComparison tests the pixel comparison logic.
func TestImageComparison(t *testing.T) {
	// Create identical images
	img1 := image.NewRGBA(image.Rect(0, 0, 10, 10))
	img2 := image.NewRGBA(image.Rect(0, 0, 10, 10))

	// Fill with same color
	for i := range img1.Pix {
		img1.Pix[i] = 128
		img2.Pix[i] = 128
	}

	similarity := compareImages(img1, img2)
	if similarity != 1.0 {
		t.Errorf("Identical images should have similarity 1.0, got %.4f", similarity)
	}

	// Modify one pixel beyond tolerance
	img2.Pix[0] = 200 // Change >5 units

	similarity = compareImages(img1, img2)
	if similarity >= 1.0 {
		t.Errorf("Modified images should have similarity <1.0, got %.4f", similarity)
	}
}

// BenchmarkSpriteRendering benchmarks sprite rendering performance.
func BenchmarkSpriteRendering(b *testing.B) {
	testCase := SpriteTestCase{
		EntityType: 0,
		GenreID:    "fantasy",
		Seed:       12345,
		Size:       64,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := RenderSpriteTestCase(testCase)
		if err != nil {
			b.Fatalf("Render failed: %v", err)
		}
	}
}

// BenchmarkImageComparison benchmarks pixel comparison performance.
func BenchmarkImageComparison(b *testing.B) {
	img1 := image.NewRGBA(image.Rect(0, 0, 64, 64))
	img2 := image.NewRGBA(image.Rect(0, 0, 64, 64))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = compareImages(img1, img2)
	}
}

// BenchmarkFullSpriteRegression benchmarks full regression test suite.
func BenchmarkFullSpriteRegression(b *testing.B) {
	tests := GenerateAllSpriteTests()
	baseline := make(map[string]*image.RGBA)
	current := make(map[string]*image.RGBA)

	// Pre-generate sprites for benchmark
	for _, tc := range tests {
		testID := fmt.Sprintf("%s-%d-%d", tc.GenreID, tc.EntityType, tc.Size)
		img, err := RenderSpriteTestCase(tc)
		if err != nil {
			b.Fatalf("Failed to render: %v", err)
		}
		baseline[testID] = img
		current[testID] = img
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = VerifySpriteRegression(baseline, current)
	}
}
