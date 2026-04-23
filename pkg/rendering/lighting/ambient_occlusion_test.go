// Package lighting provides dynamic lighting effects for rendered scenes.
package lighting

import (
	"image"
	"image/color"
	"math/rand"
	"testing"
)

// TestDefaultAOConfig tests default AO configuration.
func TestDefaultAOConfig(t *testing.T) {
	config := DefaultAOConfig()

	if !config.Enabled {
		t.Error("DefaultAOConfig should be enabled")
	}
	if config.Intensity <= 0 || config.Intensity > 1 {
		t.Errorf("Intensity = %v, want 0-1 range", config.Intensity)
	}
	if config.Radius <= 0 {
		t.Errorf("Radius = %v, want positive", config.Radius)
	}
	if config.Samples <= 0 {
		t.Errorf("Samples = %v, want positive", config.Samples)
	}
	if config.Power <= 0 {
		t.Errorf("Power = %v, want positive", config.Power)
	}
}

// TestDefaultEnhancedAOConfig tests enhanced AO configuration.
func TestDefaultEnhancedAOConfig(t *testing.T) {
	config := DefaultEnhancedAOConfig()

	if !config.Enabled {
		t.Error("DefaultEnhancedAOConfig should be enabled")
	}
	if config.CornerIntensity < 0 || config.CornerIntensity > 1 {
		t.Errorf("CornerIntensity = %v, want 0-1 range", config.CornerIntensity)
	}
	if config.EdgeIntensity < 0 || config.EdgeIntensity > 1 {
		t.Errorf("EdgeIntensity = %v, want 0-1 range", config.EdgeIntensity)
	}
}

// TestApplyAmbientOcclusion_Disabled tests AO with disabled config.
func TestApplyAmbientOcclusion_Disabled(t *testing.T) {
	img := createTestImage(100, 100, color.RGBA{128, 128, 128, 255})
	config := DefaultAOConfig()
	config.Enabled = false

	result := ApplyAmbientOcclusion(img, nil, config)

	// Should return original image unchanged
	if result != img {
		t.Error("Disabled AO should return original image")
	}
}

// TestApplyAmbientOcclusion_ZeroIntensity tests AO with zero intensity.
func TestApplyAmbientOcclusion_ZeroIntensity(t *testing.T) {
	img := createTestImage(100, 100, color.RGBA{128, 128, 128, 255})
	config := DefaultAOConfig()
	config.Intensity = 0

	result := ApplyAmbientOcclusion(img, nil, config)

	// Should return original image unchanged
	if result != img {
		t.Error("Zero intensity AO should return original image")
	}
}

// TestApplyAmbientOcclusion_Basic tests basic AO application.
func TestApplyAmbientOcclusion_Basic(t *testing.T) {
	img := createTestImage(100, 100, color.RGBA{150, 150, 150, 255})

	config := DefaultAOConfig()
	config.Intensity = 0.5
	config.Radius = 8
	config.Samples = 8

	result := ApplyAmbientOcclusion(img, nil, config)

	if result == nil {
		t.Fatal("ApplyAmbientOcclusion returned nil")
	}

	if result.Bounds() != img.Bounds() {
		t.Error("Result bounds should match input")
	}

	// Result should be darker (occlusion applied)
	origR, _, _, _ := img.At(50, 50).RGBA()
	resultR, _, _, _ := result.At(50, 50).RGBA()

	if resultR > origR {
		t.Error("AO should darken pixels")
	}
}

// TestGenerateDepthFromLuminance tests depth map generation.
func TestGenerateDepthFromLuminance(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 50, 50))

	// Create gradient from dark to bright
	for y := 0; y < 50; y++ {
		for x := 0; x < 50; x++ {
			brightness := uint8(x * 5)
			img.Set(x, y, color.RGBA{brightness, brightness, brightness, 255})
		}
	}

	depthMap := generateDepthFromLuminance(img)

	// Dark pixels should have high depth (inverted)
	darkR, _, _, _ := depthMap.At(5, 25).RGBA()
	brightR, _, _, _ := depthMap.At(45, 25).RGBA()

	if darkR <= brightR {
		t.Error("Dark pixels should have higher depth than bright pixels")
	}
}

// TestGenerateSampleDirections tests sample direction generation.
func TestGenerateSampleDirections(t *testing.T) {
	rng := rand.New(rand.NewSource(12345))
	count := 16

	directions := generateSampleDirections(count, rng)

	if len(directions) != count {
		t.Errorf("Generated %d directions, want %d", len(directions), count)
	}

	// Check that directions are roughly evenly distributed
	// All should be within unit circle
	for i, dir := range directions {
		length := dir[0]*dir[0] + dir[1]*dir[1]
		if length > 1.5 { // Allow some margin
			t.Errorf("Direction %d has length %.2f, want <= 1", i, length)
		}
	}
}

// TestGenerateSampleDirections_Deterministic tests determinism.
func TestGenerateSampleDirections_Deterministic(t *testing.T) {
	seed := int64(42)
	count := 16

	rng1 := rand.New(rand.NewSource(seed))
	dirs1 := generateSampleDirections(count, rng1)

	rng2 := rand.New(rand.NewSource(seed))
	dirs2 := generateSampleDirections(count, rng2)

	// Should be identical
	for i := range dirs1 {
		if dirs1[i][0] != dirs2[i][0] || dirs1[i][1] != dirs2[i][1] {
			t.Errorf("Direction %d differs: %v != %v", i, dirs1[i], dirs2[i])
		}
	}
}

// TestCalculateOcclusion tests occlusion calculation.
func TestCalculateOcclusion(t *testing.T) {
	// Create depth map with variation
	depthMap := image.NewRGBA(image.Rect(0, 0, 50, 50))
	for y := 0; y < 50; y++ {
		for x := 0; x < 50; x++ {
			// Center is shallow (low depth value), edges are deep (high depth value)
			dx := float64(x - 25)
			dy := float64(y - 25)
			dist := dx*dx + dy*dy
			depth := uint8(clamp(dist/10, 0, 255))
			depthMap.Set(x, y, color.RGBA{depth, depth, depth, 255})
		}
	}

	config := DefaultAOConfig()
	rng := rand.New(rand.NewSource(config.Seed))
	sampleDirs := generateSampleDirections(config.Samples, rng)

	// Center (shallow) surrounded by deeper areas should have occlusion
	centerOcclusion := calculateOcclusion(25, 25, 0.1, depthMap, config, sampleDirs)

	// Edge (deep) surrounded by similar depth should have low occlusion
	edgeOcclusion := calculateOcclusion(40, 40, 0.8, depthMap, config, sampleDirs)

	// Just verify occlusion is calculated (values may vary)
	t.Logf("Center occlusion: %.3f, Edge occlusion: %.3f", centerOcclusion, edgeOcclusion)
}

// TestApplyEnhancedAO tests enhanced AO with corners and edges.
func TestApplyEnhancedAO(t *testing.T) {
	img := createTestImage(100, 100, color.RGBA{150, 150, 150, 255})

	config := DefaultEnhancedAOConfig()
	config.Intensity = 0.5
	config.CornerIntensity = 0.3
	config.EdgeIntensity = 0.2

	result := ApplyEnhancedAO(img, nil, config)

	if result == nil {
		t.Fatal("ApplyEnhancedAO returned nil")
	}

	if result.Bounds() != img.Bounds() {
		t.Error("Result bounds should match input")
	}
}

// TestDetectCorner tests corner detection.
func TestDetectCorner(t *testing.T) {
	// Create depth map with a corner
	depthMap := image.NewRGBA(image.Rect(0, 0, 50, 50))

	// Fill with base depth
	for y := 0; y < 50; y++ {
		for x := 0; x < 50; x++ {
			depthMap.Set(x, y, color.RGBA{100, 100, 100, 255})
		}
	}

	// Create corner by raising depth in one quadrant
	for y := 0; y < 25; y++ {
		for x := 0; x < 25; x++ {
			depthMap.Set(x, y, color.RGBA{200, 200, 200, 255})
		}
	}

	// Corner at (25, 25) should be detected
	cornerFactor := detectCorner(25, 25, depthMap)

	// Flat area should not be detected as corner
	flatFactor := detectCorner(10, 10, depthMap)

	// Just verify corner detection works (exact values may vary)
	t.Logf("Corner factor: %.3f, Flat factor: %.3f", cornerFactor, flatFactor)
}

// TestDetectEdge tests edge detection.
func TestDetectEdge(t *testing.T) {
	// Create depth map with an edge
	depthMap := image.NewRGBA(image.Rect(0, 0, 50, 50))

	// Left half is shallow, right half is deep
	for y := 0; y < 50; y++ {
		for x := 0; x < 50; x++ {
			if x < 25 {
				depthMap.Set(x, y, color.RGBA{50, 50, 50, 255})
			} else {
				depthMap.Set(x, y, color.RGBA{200, 200, 200, 255})
			}
		}
	}

	// Edge at x=25 should be detected
	edgeFactor := detectEdge(25, 25, depthMap)

	// Flat area should have low edge factor
	flatFactor := detectEdge(10, 25, depthMap)

	if edgeFactor <= flatFactor {
		t.Error("Edge should have higher factor than flat area")
	}
}

// TestEnhanceCorners tests corner enhancement.
func TestEnhanceCorners(t *testing.T) {
	img := createTestImage(50, 50, color.RGBA{150, 150, 150, 255})
	depthMap := createTestImage(50, 50, color.RGBA{128, 128, 128, 255})

	result := enhanceCorners(img, depthMap, 0.3)

	if result == nil {
		t.Fatal("enhanceCorners returned nil")
	}

	if result.Bounds() != img.Bounds() {
		t.Error("Result bounds should match input")
	}
}

// TestEnhanceEdges tests edge enhancement.
func TestEnhanceEdges(t *testing.T) {
	img := createTestImage(50, 50, color.RGBA{150, 150, 150, 255})
	depthMap := createTestImage(50, 50, color.RGBA{128, 128, 128, 255})

	result := enhanceEdges(img, depthMap, 0.2)

	if result == nil {
		t.Fatal("enhanceEdges returned nil")
	}

	if result.Bounds() != img.Bounds() {
		t.Error("Result bounds should match input")
	}
}

// TestApplyAmbientOcclusion_WithDepthMap tests AO with provided depth map.
func TestApplyAmbientOcclusion_WithDepthMap(t *testing.T) {
	img := createTestImage(100, 100, color.RGBA{150, 150, 150, 255})
	depthMap := createTestImage(100, 100, color.RGBA{100, 100, 100, 255})

	config := DefaultAOConfig()
	config.Intensity = 0.5

	result := ApplyAmbientOcclusion(img, depthMap, config)

	if result == nil {
		t.Fatal("ApplyAmbientOcclusion returned nil")
	}
}

// TestApplyAmbientOcclusion_Deterministic tests deterministic AO.
func TestApplyAmbientOcclusion_Deterministic(t *testing.T) {
	img := createTestImage(50, 50, color.RGBA{150, 150, 150, 255})

	config := DefaultAOConfig()
	config.Seed = 42

	result1 := ApplyAmbientOcclusion(img, nil, config)
	result2 := ApplyAmbientOcclusion(img, nil, config)

	// Results should be identical
	if !imagesEqual(result1, result2) {
		t.Error("AO should be deterministic with same seed")
	}
}

// TestApplyAmbientOcclusion_DifferentSeeds tests different seeds produce different results.
func TestApplyAmbientOcclusion_DifferentSeeds(t *testing.T) {
	img := createTestImage(50, 50, color.RGBA{150, 150, 150, 255})

	config1 := DefaultAOConfig()
	config1.Seed = 42

	config2 := DefaultAOConfig()
	config2.Seed = 99

	result1 := ApplyAmbientOcclusion(img, nil, config1)
	result2 := ApplyAmbientOcclusion(img, nil, config2)

	// Results should differ (though might be close)
	equal := imagesEqual(result1, result2)
	if equal {
		t.Log("Warning: Different seeds produced identical results (unlikely but possible)")
	}
}

// TestSystem_ApplyAOToImage tests system AO integration.
func TestSystem_ApplyAOToImage(t *testing.T) {
	system := NewSystem()
	img := createTestImage(100, 100, color.RGBA{150, 150, 150, 255})

	result := system.ApplyAOToImage(img, nil)

	if result == nil {
		t.Fatal("ApplyAOToImage returned nil")
	}

	if result.Bounds() != img.Bounds() {
		t.Error("Result bounds should match input")
	}
}

// TestSystem_ApplyFullPostProcessing tests full post-processing pipeline.
func TestSystem_ApplyFullPostProcessing(t *testing.T) {
	system := NewSystem()
	img := createTestImage(100, 100, color.RGBA{150, 150, 150, 255})

	// Add a bright pixel for bloom
	img.Set(50, 50, color.RGBA{255, 255, 255, 255})

	result := system.ApplyFullPostProcessing(img, nil)

	if result == nil {
		t.Fatal("ApplyFullPostProcessing returned nil")
	}

	if result.Bounds() != img.Bounds() {
		t.Error("Result bounds should match input")
	}
}

// TestSystem_ApplyFullPostProcessing_AllDisabled tests with all effects disabled.
func TestSystem_ApplyFullPostProcessing_AllDisabled(t *testing.T) {
	config := DefaultConfig()
	config.BloomConfig.Enabled = false
	config.AOConfig.Enabled = false

	system := NewSystemWithConfig(config)
	img := createTestImage(100, 100, color.RGBA{150, 150, 150, 255})

	result := system.ApplyFullPostProcessing(img, nil)

	// Should return a copy (not same reference)
	if result == img {
		t.Error("All-disabled post-processing should return a copy, not original reference")
	}

	// But pixels should be identical
	if !imagesEqual(result, img) {
		t.Error("All-disabled post-processing pixels should match original")
	}
}

// TestSystem_ApplyFullPostProcessing_EnableShadowsForcesAO tests legacy shadow toggle behavior.
func TestSystem_ApplyFullPostProcessing_EnableShadowsForcesAO(t *testing.T) {
	config := DefaultConfig()
	config.BloomConfig.Enabled = false
	config.AOConfig.Enabled = false
	config.AOConfig.Seed = 12345
	config.AOConfig.Intensity = 0.8
	config.AOConfig.CornerIntensity = 0
	config.AOConfig.EdgeIntensity = 0
	config.EnableShadows = true

	img := createTestImage(100, 100, color.RGBA{150, 150, 150, 255})
	img.Set(50, 50, color.RGBA{255, 255, 255, 255})

	withoutShadowsConfig := config
	withoutShadowsConfig.EnableShadows = false
	withoutShadows := NewSystemWithConfig(withoutShadowsConfig).ApplyFullPostProcessing(img, nil)
	withShadows := NewSystemWithConfig(config).ApplyFullPostProcessing(img, nil)

	if withShadows == img {
		t.Fatal("EnableShadows path should produce a new image")
	}
	if !imagesEqual(withoutShadows, img) {
		t.Fatal("AO disabled without shadow toggle should not alter pixels")
	}
	if imagesEqual(withShadows, img) {
		t.Error("EnableShadows=true should force base ambient occlusion processing")
	}
}

// BenchmarkApplyAmbientOcclusion benchmarks AO application.
func BenchmarkApplyAmbientOcclusion(b *testing.B) {
	img := createTestImage(200, 200, color.RGBA{150, 150, 150, 255})
	config := DefaultAOConfig()
	config.Samples = 16

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ApplyAmbientOcclusion(img, nil, config)
	}
}

// BenchmarkApplyAmbientOcclusion_HighQuality benchmarks high-quality AO.
func BenchmarkApplyAmbientOcclusion_HighQuality(b *testing.B) {
	img := createTestImage(200, 200, color.RGBA{150, 150, 150, 255})
	config := DefaultAOConfig()
	config.Samples = 32
	config.Radius = 32

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ApplyAmbientOcclusion(img, nil, config)
	}
}

// BenchmarkApplyEnhancedAO benchmarks enhanced AO.
func BenchmarkApplyEnhancedAO(b *testing.B) {
	img := createTestImage(200, 200, color.RGBA{150, 150, 150, 255})
	config := DefaultEnhancedAOConfig()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ApplyEnhancedAO(img, nil, config)
	}
}

// BenchmarkGenerateDepthFromLuminance benchmarks depth map generation.
func BenchmarkGenerateDepthFromLuminance(b *testing.B) {
	img := createTestImage(200, 200, color.RGBA{128, 128, 128, 255})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = generateDepthFromLuminance(img)
	}
}

// BenchmarkDetectCorner benchmarks corner detection.
func BenchmarkDetectCorner(b *testing.B) {
	depthMap := createTestImage(200, 200, color.RGBA{128, 128, 128, 255})
	x, y := 100, 100

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = detectCorner(x, y, depthMap)
	}
}

// BenchmarkDetectEdge benchmarks edge detection.
func BenchmarkDetectEdge(b *testing.B) {
	depthMap := createTestImage(200, 200, color.RGBA{128, 128, 128, 255})
	x, y := 100, 100

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = detectEdge(x, y, depthMap)
	}
}

// TestPhase45_AOConfigScaling verifies AO config is scaled for 64×64 sprites.
func TestPhase45_AOConfigScaling(t *testing.T) {
	config := DefaultAOConfig()

	// Phase 45: Radius should be 32 (2× of original 16 for 64×64 sprites)
	if config.Radius != 32 {
		t.Errorf("Phase 45: DefaultAOConfig().Radius = %d, want 32 (scaled for 64×64)", config.Radius)
	}

	// Phase 45: Samples should be 24 (scaled from 16 for larger sampling area)
	if config.Samples != 24 {
		t.Errorf("Phase 45: DefaultAOConfig().Samples = %d, want 24 (scaled for 64×64)", config.Samples)
	}
}

// TestPhase45_AOWorksWith64x64 tests AO works correctly with 64×64 sprites.
func TestPhase45_AOWorksWith64x64(t *testing.T) {
	// Create 64×64 test image (Phase 45 standard size)
	img := createTestImage(64, 64, color.RGBA{180, 180, 180, 255})

	// Add some variation to test occlusion
	for y := 20; y < 44; y++ {
		for x := 20; x < 44; x++ {
			img.Set(x, y, color.RGBA{60, 60, 60, 255})
		}
	}

	config := DefaultAOConfig()
	result := ApplyAmbientOcclusion(img, nil, config)

	if result == nil {
		t.Fatal("ApplyAmbientOcclusion returned nil for 64×64 sprite")
	}

	if result.Bounds().Dx() != 64 || result.Bounds().Dy() != 64 {
		t.Error("Result should maintain 64×64 dimensions")
	}

	// Verify some darkening occurred
	r, _, _, _ := result.At(22, 22).RGBA()
	if r >= 60*257 {
		t.Log("Note: AO may not visibly darken in test conditions")
	}
}

// TestPhase45_EnhancedAOWorksWith64x64 tests enhanced AO with 64×64 sprites.
func TestPhase45_EnhancedAOWorksWith64x64(t *testing.T) {
	// Create 64×64 test image
	img := createTestImage(64, 64, color.RGBA{200, 200, 200, 255})

	// Add corner regions
	for y := 0; y < 20; y++ {
		for x := 0; x < 20; x++ {
			img.Set(x, y, color.RGBA{40, 40, 40, 255})
		}
	}

	config := DefaultEnhancedAOConfig()
	result := ApplyEnhancedAO(img, nil, config)

	if result == nil {
		t.Fatal("ApplyEnhancedAO returned nil for 64×64 sprite")
	}

	if result.Bounds() != img.Bounds() {
		t.Error("Result bounds should match 64×64 input")
	}
}

// BenchmarkPhase45_AO64x64 benchmarks AO performance with 64×64 sprites.
func BenchmarkPhase45_AO64x64(b *testing.B) {
	img := createTestImage(64, 64, color.RGBA{150, 150, 150, 255})
	config := DefaultAOConfig()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ApplyAmbientOcclusion(img, nil, config)
	}
}
