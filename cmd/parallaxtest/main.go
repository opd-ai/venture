// Package main provides a CLI tool for testing parallax depth tile generation.
// This tool generates and exports tiles with parallax depth effects for visual verification.
package main

import (
	"flag"
	"fmt"
	"image"
	"image/png"
	"log"
	"math"
	"os"
	"path/filepath"

	"github.com/opd-ai/venture/pkg/rendering/tiles"
)

func main() {
	// Command-line flags
	var (
		outputDir    = flag.String("output", "./parallax_output", "Output directory for generated images")
		tileType     = flag.String("type", "floor", "Tile type (floor, wall, door, corridor, water, lava, trap, stairs)")
		genreID      = flag.String("genre", "fantasy", "Genre ID (fantasy, scifi, horror, cyberpunk, postapocalyptic)")
		seed         = flag.Int64("seed", 12345, "Random seed for generation")
		width        = flag.Int("width", 32, "Tile width in pixels")
		height       = flag.Int("height", 32, "Tile height in pixels")
		cameraX      = flag.Float64("camera-x", 10.0, "Camera X position for parallax")
		cameraY      = flag.Float64("camera-y", 5.0, "Camera Y position for parallax")
		aoIntensity  = flag.Float64("ao", 0.5, "Ambient occlusion intensity (0.0-1.0)")
		shadowHeight = flag.Float64("shadow", 0.3, "Shadow height (0.0-1.0)")
	)
	flag.Parse()

	// Parse tile type
	tileTypeMap := map[string]tiles.TileType{
		"floor":    tiles.TileFloor,
		"wall":     tiles.TileWall,
		"door":     tiles.TileDoor,
		"corridor": tiles.TileCorridor,
		"water":    tiles.TileWater,
		"lava":     tiles.TileLava,
		"trap":     tiles.TileTrap,
		"stairs":   tiles.TileStairs,
	}

	tType, ok := tileTypeMap[*tileType]
	if !ok {
		log.Fatalf("Invalid tile type: %s (valid types: floor, wall, door, corridor, water, lava, trap, stairs)", *tileType)
	}

	// Create output directory
	if err := os.MkdirAll(*outputDir, 0755); err != nil {
		log.Fatalf("Failed to create output directory: %v", err)
	}

	// Create generator
	gen := tiles.NewGenerator()

	// Base configuration
	config := tiles.Config{
		Type:    tType,
		Width:   *width,
		Height:  *height,
		GenreID: *genreID,
		Seed:    *seed,
		Variant: 0.5,
	}

	fmt.Printf("Generating parallax depth tiles...\n")
	fmt.Printf("  Type: %s\n", *tileType)
	fmt.Printf("  Genre: %s\n", *genreID)
	fmt.Printf("  Seed: %d\n", *seed)
	fmt.Printf("  Size: %dx%d\n", *width, *height)
	fmt.Printf("  Camera: (%.1f, %.1f)\n", *cameraX, *cameraY)
	fmt.Printf("  AO Intensity: %.2f\n", *aoIntensity)
	fmt.Printf("  Shadow Height: %.2f\n", *shadowHeight)
	fmt.Println()

	// 1. Generate base tile (no parallax)
	fmt.Println("1. Generating base tile (no parallax)...")
	baseTile, err := gen.Generate(config)
	if err != nil {
		log.Fatalf("Failed to generate base tile: %v", err)
	}
	saveImage(filepath.Join(*outputDir, "01_base_tile.png"), baseTile)

	// 2. Generate background layer
	fmt.Println("2. Generating background layer...")
	bgConfig := tiles.ParallaxConfig{
		BaseConfig:    config,
		Layer:         tiles.LayerBackground,
		CameraX:       *cameraX,
		CameraY:       *cameraY,
		ParallaxDepth: 0.3,
		AOIntensity:   *aoIntensity * 0.6,
		ShadowHeight:  *shadowHeight * 0.3,
		ShadowAngle:   math.Pi / 4,
	}
	bgTile, err := gen.GenerateWithParallax(bgConfig)
	if err != nil {
		log.Fatalf("Failed to generate background layer: %v", err)
	}
	saveImage(filepath.Join(*outputDir, "02_background_layer.png"), bgTile)

	// 3. Generate base layer with full effects
	fmt.Println("3. Generating base layer with AO and shadows...")
	baseLayerConfig := tiles.ParallaxConfig{
		BaseConfig:    config,
		Layer:         tiles.LayerBase,
		CameraX:       *cameraX,
		CameraY:       *cameraY,
		ParallaxDepth: 1.0,
		AOIntensity:   *aoIntensity,
		ShadowHeight:  *shadowHeight,
		ShadowAngle:   math.Pi / 4,
	}
	baseLayer, err := gen.GenerateWithParallax(baseLayerConfig)
	if err != nil {
		log.Fatalf("Failed to generate base layer: %v", err)
	}
	saveImage(filepath.Join(*outputDir, "03_base_layer_with_effects.png"), baseLayer)

	// 4. Generate foreground layer
	fmt.Println("4. Generating foreground layer...")
	fgConfig := tiles.ParallaxConfig{
		BaseConfig:    config,
		Layer:         tiles.LayerForeground,
		CameraX:       *cameraX,
		CameraY:       *cameraY,
		ParallaxDepth: 1.4,
		AOIntensity:   *aoIntensity * 0.4,
		ShadowHeight:  *shadowHeight * 1.5,
		ShadowAngle:   math.Pi / 4,
	}
	fgTile, err := gen.GenerateWithParallax(fgConfig)
	if err != nil {
		log.Fatalf("Failed to generate foreground layer: %v", err)
	}
	saveImage(filepath.Join(*outputDir, "04_foreground_layer.png"), fgTile)

	// 5. Generate all layers at once
	fmt.Println("5. Generating all three layers...")
	bg, base, fg, err := gen.GenerateLayeredTile(config, *cameraX, *cameraY)
	if err != nil {
		log.Fatalf("Failed to generate layered tile: %v", err)
	}
	saveImage(filepath.Join(*outputDir, "05_layered_background.png"), bg)
	saveImage(filepath.Join(*outputDir, "06_layered_base.png"), base)
	saveImage(filepath.Join(*outputDir, "07_layered_foreground.png"), fg)

	// 6. Composite layers
	fmt.Println("6. Compositing all layers...")
	composite := tiles.CompositeLayers(bg, base, fg)
	saveImage(filepath.Join(*outputDir, "08_composite.png"), composite)

	// 7. Generate AO map
	fmt.Println("7. Generating ambient occlusion map...")
	aoMap, err := gen.GenerateAOMap(config, *aoIntensity)
	if err != nil {
		log.Fatalf("Failed to generate AO map: %v", err)
	}
	saveImage(filepath.Join(*outputDir, "09_ao_map.png"), aoMap)

	// 8. Generate comparison: no parallax vs with parallax
	fmt.Println("8. Generating comparison images...")
	noParallaxConfig := tiles.ParallaxConfig{
		BaseConfig:    config,
		Layer:         tiles.LayerBase,
		CameraX:       0,
		CameraY:       0,
		ParallaxDepth: 1.0,
		AOIntensity:   0,
		ShadowHeight:  0,
		ShadowAngle:   0,
	}
	noEffects, err := gen.GenerateWithParallax(noParallaxConfig)
	if err != nil {
		log.Fatalf("Failed to generate no-effects tile: %v", err)
	}
	saveImage(filepath.Join(*outputDir, "10_no_effects.png"), noEffects)

	withEffectsConfig := tiles.ParallaxConfig{
		BaseConfig:    config,
		Layer:         tiles.LayerBase,
		CameraX:       *cameraX,
		CameraY:       *cameraY,
		ParallaxDepth: 1.0,
		AOIntensity:   *aoIntensity,
		ShadowHeight:  *shadowHeight,
		ShadowAngle:   math.Pi / 4,
	}
	withEffects, err := gen.GenerateWithParallax(withEffectsConfig)
	if err != nil {
		log.Fatalf("Failed to generate with-effects tile: %v", err)
	}
	saveImage(filepath.Join(*outputDir, "11_with_effects.png"), withEffects)

	fmt.Println()
	fmt.Printf("✓ Successfully generated all parallax test images in: %s\n", *outputDir)
	fmt.Println()
	fmt.Println("Generated files:")
	fmt.Println("  01_base_tile.png                - Original tile without any effects")
	fmt.Println("  02_background_layer.png         - Background layer (darkened, slow parallax)")
	fmt.Println("  03_base_layer_with_effects.png  - Base layer with AO and shadows")
	fmt.Println("  04_foreground_layer.png         - Foreground layer (brightened, fast parallax)")
	fmt.Println("  05_layered_background.png       - Background from layered generation")
	fmt.Println("  06_layered_base.png             - Base from layered generation")
	fmt.Println("  07_layered_foreground.png       - Foreground from layered generation")
	fmt.Println("  08_composite.png                - All layers composited together")
	fmt.Println("  09_ao_map.png                   - Ambient occlusion map (grayscale)")
	fmt.Println("  10_no_effects.png               - Comparison: no effects applied")
	fmt.Println("  11_with_effects.png             - Comparison: full effects applied")
}

// saveImage saves an image to a PNG file.
func saveImage(filename string, img *image.RGBA) {
	f, err := os.Create(filename)
	if err != nil {
		log.Fatalf("Failed to create file %s: %v", filename, err)
	}
	defer f.Close()

	if err := png.Encode(f, img); err != nil {
		log.Fatalf("Failed to encode PNG %s: %v", filename, err)
	}

	fmt.Printf("  ✓ Saved: %s\n", filename)
}
