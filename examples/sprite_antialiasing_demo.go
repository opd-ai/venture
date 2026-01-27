// Example program demonstrating sprite anti-aliasing.
// This program shows how to generate sprites with different anti-aliasing quality levels.
//
// Usage:
//   go run examples/sprite_antialiasing_demo.go
//
// The program generates sprites at different AA quality levels to demonstrate visual differences.
package main

import (
	"fmt"
	"log"

	"github.com/opd-ai/venture/pkg/rendering/shapes"
	"github.com/opd-ai/venture/pkg/rendering/sprites"
)

func main() {
	fmt.Println("Sprite Anti-Aliasing Demo")
	fmt.Println("==========================")
	fmt.Println()

	// Create sprite generator
	gen := sprites.NewGenerator()

	// Quality levels to demonstrate
	qualityLevels := []struct {
		name      string
		antiAlias shapes.AntiAliasQuality
		desc      string
	}{
		{"Off", shapes.AntiAliasOff, "Hard edges (fastest)"},
		{"Low", shapes.AntiAliasLow, "2x2 super-sampling (good balance)"},
		{"Medium", shapes.AntiAliasMedium, "4x4 super-sampling (high quality)"},
		{"High", shapes.AntiAliasHigh, "8x8 super-sampling (maximum quality)"},
	}

	fmt.Println("Generating sprites with different anti-aliasing levels:")
	fmt.Println()

	// Generate sprites at each quality level
	for _, level := range qualityLevels {
		fmt.Printf("- %s Quality (%s)\n", level.name, level.desc)

		config := sprites.Config{
			Type:       sprites.SpriteEntity,
			Width:      64,
			Height:     64,
			Seed:       12345,
			GenreID:    "fantasy",
			Complexity: 0.7,
			AntiAlias:  level.antiAlias,
		}

		img, err := gen.Generate(config)
		if err != nil {
			log.Printf("  ERROR: Failed to generate sprite: %v", err)
			continue
		}

		bounds := img.Bounds()
		fmt.Printf("  ✓ Generated %dx%d sprite\n", bounds.Dx(), bounds.Dy())
	}

	fmt.Println()
	fmt.Println("Visual Quality Comparison:")
	fmt.Println("- AntiAliasOff:    Jagged edges, pixel-perfect crisp (best performance)")
	fmt.Println("- AntiAliasLow:    Slightly smoothed edges (recommended default)")
	fmt.Println("- AntiAliasMedium: Smooth edges (good for close-up views)")
	fmt.Println("- AntiAliasHigh:   Very smooth edges (best for large sprites)")
	fmt.Println()

	// Demonstrate complex entity with anti-aliasing
	fmt.Println("Generating complex entity with high-quality AA:")
	complexConfig := sprites.Config{
		Type:       sprites.SpriteEntity,
		Width:      128,
		Height:     128,
		Seed:       54321,
		GenreID:    "sci-fi",
		Complexity: 0.9,
		Custom: map[string]interface{}{
			"entityType": "warrior",
			"facing":     "down",
			"hasWeapon":  true,
		},
		AntiAlias: shapes.AntiAliasHigh,
	}

	complexImg, err := gen.Generate(complexConfig)
	if err != nil {
		log.Fatalf("Failed to generate complex entity: %v", err)
	}

	bounds := complexImg.Bounds()
	fmt.Printf("✓ Generated complex warrior sprite: %dx%d with high-quality AA\n", bounds.Dx(), bounds.Dy())
	fmt.Println()

	// Demonstrate all sprite types with anti-aliasing
	fmt.Println("All sprite types support anti-aliasing:")
	spriteTypes := []sprites.SpriteType{
		sprites.SpriteEntity,
		sprites.SpriteItem,
		sprites.SpriteTile,
		sprites.SpriteParticle,
		sprites.SpriteUI,
	}

	for _, spriteType := range spriteTypes {
		config := sprites.Config{
			Type:       spriteType,
			Width:      64,
			Height:     64,
			Seed:       99999,
			GenreID:    "cyberpunk",
			Complexity: 0.5,
			AntiAlias:  shapes.AntiAliasMedium,
		}

		img, err := gen.Generate(config)
		if err != nil {
			log.Printf("  ERROR: %s generation failed: %v", spriteType, err)
			continue
		}

		bounds := img.Bounds()
		fmt.Printf("  ✓ %s sprite (%dx%d) with medium AA\n", spriteType, bounds.Dx(), bounds.Dy())
	}

	fmt.Println()
	fmt.Println("Performance Considerations:")
	fmt.Println("- AntiAliasOff:    1x cost (baseline)")
	fmt.Println("- AntiAliasLow:    ~2x cost (4 samples per pixel)")
	fmt.Println("- AntiAliasMedium: ~4x cost (16 samples per pixel)")
	fmt.Println("- AntiAliasHigh:   ~8x cost (64 samples per pixel)")
	fmt.Println()
	fmt.Println("Recommendation: Use AntiAliasLow as default for 60 FPS gameplay")
	fmt.Println("                Use AntiAliasMedium for high-quality screenshots")
	fmt.Println("                Use AntiAliasHigh for marketing materials only")
	fmt.Println()
	fmt.Println("Demo complete!")
}
