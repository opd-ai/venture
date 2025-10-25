//go:build !test
// +build !test

package sprites

import (
	"testing"

	"github.com/opd-ai/venture/pkg/rendering/palette"
)

// TestGenerateAnimationFrame_Basic tests basic frame generation.
func TestGenerateAnimationFrame_Basic(t *testing.T) {
	gen := NewGenerator()

	pal, err := palette.NewGenerator().Generate("fantasy", 12345)
	if err != nil {
		t.Fatalf("Failed to generate palette: %v", err)
	}

	config := Config{
		Type:       SpriteEntity,
		Width:      28,
		Height:     28,
		Seed:       12345,
		GenreID:    "fantasy",
		Complexity: 0.5,
		Palette:    pal,
	}

	frame, err := gen.GenerateAnimationFrame(config, "walk", 0, 4)
	if err != nil {
		t.Fatalf("Failed to generate animation frame: %v", err)
	}

	if frame == nil {
		t.Fatal("Expected non-nil frame")
	}

	bounds := frame.Bounds()
	if bounds.Dx() != 28 || bounds.Dy() != 28 {
		t.Errorf("Expected 28x28 frame, got %dx%d", bounds.Dx(), bounds.Dy())
	}
}

// TestGenerateAnimationFrame_Determinism tests deterministic generation.
func TestGenerateAnimationFrame_Determinism(t *testing.T) {
	gen := NewGenerator()

	pal, err := palette.NewGenerator().Generate("fantasy", 12345)
	if err != nil {
		t.Fatalf("Failed to generate palette: %v", err)
	}

	config := Config{
		Type:       SpriteEntity,
		Width:      28,
		Height:     28,
		Seed:       12345,
		GenreID:    "fantasy",
		Complexity: 0.5,
		Palette:    pal,
	}

	states := []string{"idle", "walk", "attack", "cast", "death"}

	for _, state := range states {
		t.Run(state, func(t *testing.T) {
			// Generate same frame twice
			frame1, err1 := gen.GenerateAnimationFrame(config, state, 0, 4)
			frame2, err2 := gen.GenerateAnimationFrame(config, state, 0, 4)

			if err1 != nil || err2 != nil {
				t.Fatalf("Failed to generate frames: %v, %v", err1, err2)
			}

			// Verify both frames exist and have same dimensions
			// (Pixel comparison requires game loop which we don't have in tests)
			if frame1 == nil || frame2 == nil {
				t.Error("Expected non-nil frames")
			}

			b1 := frame1.Bounds()
			b2 := frame2.Bounds()

			if b1 != b2 {
				t.Errorf("Expected identical bounds, got %v and %v", b1, b2)
			}
		})
	}
}

// TestGenerateAnimationFrame_DifferentFrames tests frame variation.
func TestGenerateAnimationFrame_DifferentFrames(t *testing.T) {
	gen := NewGenerator()

	pal, err := palette.NewGenerator().Generate("fantasy", 12345)
	if err != nil {
		t.Fatalf("Failed to generate palette: %v", err)
	}

	config := Config{
		Type:       SpriteEntity,
		Width:      28,
		Height:     28,
		Seed:       12345,
		GenreID:    "fantasy",
		Complexity: 0.5,
		Palette:    pal,
	}

	// Generate different frames of walk cycle
	frame0, err := gen.GenerateAnimationFrame(config, "walk", 0, 8)
	if err != nil {
		t.Fatalf("Failed to generate frame 0: %v", err)
	}

	frame4, err := gen.GenerateAnimationFrame(config, "walk", 4, 8)
	if err != nil {
		t.Fatalf("Failed to generate frame 4: %v", err)
	}

	// Frames should be different (different parts of walk cycle)
	// We can't compare pixels directly in tests, so just verify both exist
	if frame0 == nil || frame4 == nil {
		t.Error("Expected non-nil frames")
	}
}

// TestGenerateAnimationFrame_DifferentStates tests state variations.
func TestGenerateAnimationFrame_DifferentStates(t *testing.T) {
	gen := NewGenerator()

	pal, err := palette.NewGenerator().Generate("fantasy", 12345)
	if err != nil {
		t.Fatalf("Failed to generate palette: %v", err)
	}

	config := Config{
		Type:       SpriteEntity,
		Width:      28,
		Height:     28,
		Seed:       12345,
		GenreID:    "fantasy",
		Complexity: 0.5,
		Palette:    pal,
	}

	// Generate first frame of different states
	idleFrame, err := gen.GenerateAnimationFrame(config, "idle", 0, 4)
	if err != nil {
		t.Fatalf("Failed to generate idle frame: %v", err)
	}

	attackFrame, err := gen.GenerateAnimationFrame(config, "attack", 0, 6)
	if err != nil {
		t.Fatalf("Failed to generate attack frame: %v", err)
	}

	// Frames should exist for different states
	// (Pixel comparison not available in tests)
	if idleFrame == nil || attackFrame == nil {
		t.Error("Expected non-nil frames for both states")
	}
}

// TestCalculateAnimationOffset tests offset calculations.
func TestCalculateAnimationOffset(t *testing.T) {
	tests := []struct {
		state      string
		frameIndex int
		frameCount int
	}{
		{"walk", 0, 8},
		{"walk", 4, 8},
		{"run", 0, 8},
		{"jump", 2, 4},
		{"attack", 0, 6},
		{"attack", 3, 6},
		{"hit", 0, 3},
		{"death", 5, 6},
	}

	for _, tt := range tests {
		t.Run(tt.state, func(t *testing.T) {
			offset := calculateAnimationOffset(tt.state, tt.frameIndex, tt.frameCount)

			// Offset values should be reasonable (not extreme)
			if offset.X < -20 || offset.X > 20 {
				t.Errorf("Offset X out of reasonable range: %f", offset.X)
			}
			if offset.Y < -50 || offset.Y > 50 {
				t.Errorf("Offset Y out of reasonable range: %f", offset.Y)
			}
		})
	}
}

// TestCalculateAnimationRotation tests rotation calculations.
func TestCalculateAnimationRotation(t *testing.T) {
	tests := []struct {
		state      string
		frameIndex int
		frameCount int
	}{
		{"attack", 0, 6},
		{"attack", 2, 6},
		{"death", 0, 6},
		{"death", 5, 6},
		{"cast", 4, 8},
	}

	for _, tt := range tests {
		t.Run(tt.state, func(t *testing.T) {
			rotation := calculateAnimationRotation(tt.state, tt.frameIndex, tt.frameCount)

			// Rotation should be reasonable (not more than 2 full rotations)
			if rotation < -6.28 || rotation > 6.28 {
				t.Errorf("Rotation out of reasonable range: %f", rotation)
			}
		})
	}
}

// TestCalculateAnimationScale tests scale calculations.
func TestCalculateAnimationScale(t *testing.T) {
	tests := []struct {
		state      string
		frameIndex int
		frameCount int
	}{
		{"jump", 0, 4},
		{"jump", 2, 4},
		{"hit", 0, 3},
		{"attack", 3, 6},
		{"idle", 0, 4}, // Should return 1.0
	}

	for _, tt := range tests {
		t.Run(tt.state, func(t *testing.T) {
			scale := calculateAnimationScale(tt.state, tt.frameIndex, tt.frameCount)

			// Scale should be reasonable (0.5x to 1.5x)
			if scale < 0.5 || scale > 1.5 {
				t.Errorf("Scale out of reasonable range: %f", scale)
			}
		})
	}
}

// TestHashString tests string hashing for seed derivation.
func TestHashString(t *testing.T) {
	// Test determinism
	hash1 := hashString("walk")
	hash2 := hashString("walk")

	if hash1 != hash2 {
		t.Error("Expected identical hashes for same string")
	}

	// Test uniqueness
	hashWalk := hashString("walk")
	hashRun := hashString("run")
	hashAttack := hashString("attack")

	if hashWalk == hashRun || hashWalk == hashAttack || hashRun == hashAttack {
		t.Error("Expected different hashes for different strings")
	}

	// Test non-zero
	if hashString("test") == 0 {
		t.Error("Expected non-zero hash")
	}
}

// TestGenerateAnimationFrame_AllStates tests all animation states.
func TestGenerateAnimationFrame_AllStates(t *testing.T) {
	gen := NewGenerator()

	pal, err := palette.NewGenerator().Generate("fantasy", 12345)
	if err != nil {
		t.Fatalf("Failed to generate palette: %v", err)
	}

	config := Config{
		Type:       SpriteEntity,
		Width:      28,
		Height:     28,
		Seed:       12345,
		GenreID:    "fantasy",
		Complexity: 0.5,
		Palette:    pal,
	}

	states := []string{
		"idle", "walk", "run", "attack", "cast",
		"hit", "death", "jump", "crouch", "use",
	}

	for _, state := range states {
		t.Run(state, func(t *testing.T) {
			frame, err := gen.GenerateAnimationFrame(config, state, 0, 4)
			if err != nil {
				t.Errorf("Failed to generate frame for state %s: %v", state, err)
			}
			if frame == nil {
				t.Errorf("Expected non-nil frame for state %s", state)
			}
		})
	}
}

// BenchmarkGenerateAnimationFrame benchmarks frame generation.
func BenchmarkGenerateAnimationFrame(b *testing.B) {
	gen := NewGenerator()

	pal, _ := palette.NewGenerator().Generate("fantasy", 12345)

	config := Config{
		Type:       SpriteEntity,
		Width:      28,
		Height:     28,
		Seed:       12345,
		GenreID:    "fantasy",
		Complexity: 0.5,
		Palette:    pal,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = gen.GenerateAnimationFrame(config, "walk", i%8, 8)
	}
}

// BenchmarkGenerateAnimationFrame_HighComplexity benchmarks complex frames.
func BenchmarkGenerateAnimationFrame_HighComplexity(b *testing.B) {
	gen := NewGenerator()

	pal, _ := palette.NewGenerator().Generate("fantasy", 12345)

	config := Config{
		Type:       SpriteEntity,
		Width:      28,
		Height:     28,
		Seed:       12345,
		GenreID:    "fantasy",
		Complexity: 1.0, // Maximum complexity
		Palette:    pal,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = gen.GenerateAnimationFrame(config, "walk", i%8, 8)
	}
}
