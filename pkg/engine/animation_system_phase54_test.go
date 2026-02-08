// Package engine provides Phase 5.4 tests for genre palette integration.
package engine

import (
	"testing"

	"github.com/opd-ai/venture/pkg/rendering/palette"
	"github.com/opd-ai/venture/pkg/rendering/sprites"
)

// TestAnimationSystem_SetPaletteOptions tests palette options configuration.
func TestAnimationSystem_SetPaletteOptions(t *testing.T) {
	spriteGen := sprites.NewGenerator()
	sys := NewAnimationSystem(spriteGen)

	// Test setting palette options
	opts := &palette.GenerationOptions{
		Harmony:   palette.HarmonyTriadic,
		Mood:      palette.MoodVibrant,
		Rarity:    palette.RarityEpic,
		MinColors: 16,
	}

	sys.SetPaletteOptions(opts)

	// Verify cache was cleared (implementation detail)
	sys.cacheMutex.RLock()
	cacheSize := len(sys.frameCache)
	sys.cacheMutex.RUnlock()

	if cacheSize != 0 {
		t.Errorf("Expected cache to be cleared after setting palette options, got %d entries", cacheSize)
	}

	// Verify options were set
	if sys.paletteOptions == nil {
		t.Fatal("Expected palette options to be set, got nil")
	}

	if sys.paletteOptions.Harmony != palette.HarmonyTriadic {
		t.Errorf("Expected harmony %v, got %v", palette.HarmonyTriadic, sys.paletteOptions.Harmony)
	}

	if sys.paletteOptions.Mood != palette.MoodVibrant {
		t.Errorf("Expected mood %v, got %v", palette.MoodVibrant, sys.paletteOptions.Mood)
	}

	if sys.paletteOptions.Rarity != palette.RarityEpic {
		t.Errorf("Expected rarity %v, got %v", palette.RarityEpic, sys.paletteOptions.Rarity)
	}

	if sys.paletteOptions.MinColors != 16 {
		t.Errorf("Expected min colors 16, got %d", sys.paletteOptions.MinColors)
	}
}

// TestAnimationSystem_SetPaletteOptions_Nil tests clearing palette options.
func TestAnimationSystem_SetPaletteOptions_Nil(t *testing.T) {
	spriteGen := sprites.NewGenerator()
	sys := NewAnimationSystem(spriteGen)

	// Set options then clear
	opts := &palette.GenerationOptions{
		Harmony: palette.HarmonyComplementary,
		Mood:    palette.MoodNormal,
		Rarity:  palette.RarityCommon,
	}
	sys.SetPaletteOptions(opts)
	sys.SetPaletteOptions(nil)

	// Verify options were cleared
	if sys.paletteOptions != nil {
		t.Error("Expected palette options to be nil after clearing")
	}
}

// TestAnimationSystem_BuildSpriteConfig_WithPaletteOptions tests sprite config includes palette options.
func TestAnimationSystem_BuildSpriteConfig_WithPaletteOptions(t *testing.T) {
	spriteGen := sprites.NewGenerator()
	sys := NewAnimationSystem(spriteGen)

	// Set custom palette options
	opts := &palette.GenerationOptions{
		Harmony:   palette.HarmonyAnalogous,
		Mood:      palette.MoodDark,
		Rarity:    palette.RarityRare,
		MinColors: 12,
	}
	sys.SetPaletteOptions(opts)

	// Create test entity with components
	entity := &Entity{ID: 12345}
	sprite := &EbitenSprite{Width: 32, Height: 32}
	anim := &AnimationComponent{Seed: 12345}

	// Build sprite config
	config := sys.buildSpriteConfig(entity, sprite, anim)

	// Verify palette options are included in config
	if config.PaletteOptions == nil {
		t.Fatal("Expected palette options in sprite config, got nil")
	}

	if config.PaletteOptions.Harmony != palette.HarmonyAnalogous {
		t.Errorf("Expected harmony %v, got %v", palette.HarmonyAnalogous, config.PaletteOptions.Harmony)
	}

	if config.PaletteOptions.Mood != palette.MoodDark {
		t.Errorf("Expected mood %v, got %v", palette.MoodDark, config.PaletteOptions.Mood)
	}

	if config.PaletteOptions.Rarity != palette.RarityRare {
		t.Errorf("Expected rarity %v, got %v", palette.RarityRare, config.PaletteOptions.Rarity)
	}
}

// TestAnimationSystem_BuildSpriteConfig_DefaultPalette tests default palette without options.
func TestAnimationSystem_BuildSpriteConfig_DefaultPalette(t *testing.T) {
	spriteGen := sprites.NewGenerator()
	sys := NewAnimationSystem(spriteGen)

	// Don't set palette options (use defaults)

	// Create test entity
	entity := &Entity{ID: 54321}
	sprite := &EbitenSprite{Width: 32, Height: 32}
	anim := &AnimationComponent{Seed: 54321}

	// Build sprite config
	config := sys.buildSpriteConfig(entity, sprite, anim)

	// Verify palette options are nil (will use defaults in sprite generator)
	if config.PaletteOptions != nil {
		t.Error("Expected nil palette options for default behavior")
	}
}

// TestPaletteOptions_AllHarmonyTypes tests all harmony type options.
func TestPaletteOptions_AllHarmonyTypes(t *testing.T) {
	tests := []struct {
		name    string
		harmony palette.HarmonyType
	}{
		{"Complementary", palette.HarmonyComplementary},
		{"Analogous", palette.HarmonyAnalogous},
		{"Triadic", palette.HarmonyTriadic},
		{"Tetradic", palette.HarmonyTetradic},
		{"SplitComplementary", palette.HarmonySplitComplementary},
		{"Monochromatic", palette.HarmonyMonochromatic},
	}

	spriteGen := sprites.NewGenerator()
	sys := NewAnimationSystem(spriteGen)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := &palette.GenerationOptions{
				Harmony:   tt.harmony,
				Mood:      palette.MoodNormal,
				Rarity:    palette.RarityCommon,
				MinColors: 12,
			}

			sys.SetPaletteOptions(opts)

			if sys.paletteOptions.Harmony != tt.harmony {
				t.Errorf("Expected harmony %v, got %v", tt.harmony, sys.paletteOptions.Harmony)
			}
		})
	}
}

// TestPaletteOptions_AllMoodTypes tests all mood type options.
func TestPaletteOptions_AllMoodTypes(t *testing.T) {
	moods := []palette.MoodType{
		palette.MoodNormal,
		palette.MoodBright,
		palette.MoodDark,
		palette.MoodSaturated,
		palette.MoodMuted,
		palette.MoodVibrant,
		palette.MoodPastel,
		palette.MoodTense,
		palette.MoodCalm,
		palette.MoodVictorious,
		palette.MoodMelancholic,
		palette.MoodEnergetic,
		palette.MoodMystical,
		palette.MoodOminous,
		palette.MoodSerene,
		palette.MoodAggressive,
		palette.MoodPlayful,
		palette.MoodSomber,
		palette.MoodEthereal,
		palette.MoodDangerous,
		palette.MoodPeaceful,
		palette.MoodChaotic,
		palette.MoodRegal,
		palette.MoodDesolate,
	}

	spriteGen := sprites.NewGenerator()
	sys := NewAnimationSystem(spriteGen)

	for _, mood := range moods {
		opts := &palette.GenerationOptions{
			Harmony:   palette.HarmonyComplementary,
			Mood:      mood,
			Rarity:    palette.RarityCommon,
			MinColors: 12,
		}

		sys.SetPaletteOptions(opts)

		if sys.paletteOptions.Mood != mood {
			t.Errorf("Expected mood %v, got %v", mood, sys.paletteOptions.Mood)
		}
	}
}

// TestPaletteOptions_AllRarityTypes tests all rarity type options.
func TestPaletteOptions_AllRarityTypes(t *testing.T) {
	rarities := []struct {
		name   string
		rarity palette.Rarity
	}{
		{"Common", palette.RarityCommon},
		{"Uncommon", palette.RarityUncommon},
		{"Rare", palette.RarityRare},
		{"Epic", palette.RarityEpic},
		{"Legendary", palette.RarityLegendary},
	}

	spriteGen := sprites.NewGenerator()
	sys := NewAnimationSystem(spriteGen)

	for _, tt := range rarities {
		t.Run(tt.name, func(t *testing.T) {
			opts := &palette.GenerationOptions{
				Harmony:   palette.HarmonyComplementary,
				Mood:      palette.MoodNormal,
				Rarity:    tt.rarity,
				MinColors: 12,
			}

			sys.SetPaletteOptions(opts)

			if sys.paletteOptions.Rarity != tt.rarity {
				t.Errorf("Expected rarity %v, got %v", tt.rarity, sys.paletteOptions.Rarity)
			}
		})
	}
}

// TestPaletteOptions_CacheClear tests that cache is cleared when options change.
func TestPaletteOptions_CacheClear(t *testing.T) {
	spriteGen := sprites.NewGenerator()
	sys := NewAnimationSystem(spriteGen)

	// Populate cache with dummy data
	sys.cacheMutex.Lock()
	sys.frameCache[uint64(12345)] = nil
	sys.cacheKeys = []uint64{uint64(12345)}
	initialSize := len(sys.frameCache)
	sys.cacheMutex.Unlock()

	if initialSize != 1 {
		t.Fatalf("Expected initial cache size 1, got %d", initialSize)
	}

	// Change palette options
	opts := &palette.GenerationOptions{
		Harmony:   palette.HarmonyTriadic,
		Mood:      palette.MoodVibrant,
		Rarity:    palette.RarityEpic,
		MinColors: 12,
	}
	sys.SetPaletteOptions(opts)

	// Verify cache was cleared
	sys.cacheMutex.RLock()
	cacheSize := len(sys.frameCache)
	sys.cacheMutex.RUnlock()

	if cacheSize != 0 {
		t.Errorf("Expected cache to be cleared, got %d entries", cacheSize)
	}
}
