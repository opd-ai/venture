// Package main provides a CLI tool for testing lighting, bloom, and ambient occlusion effects.
// This tool generates test images demonstrating the Phase 17.1 lighting enhancements.
//
// Usage:
//
//	go run ./cmd/lightingtest
//	go run ./cmd/lightingtest -output /tmp/lighting
//	go run ./cmd/lightingtest -bloom-only
//	go run ./cmd/lightingtest -ao-only
//	go run ./cmd/lightingtest -seed 42
package main

import (
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"log"
	"math"
	"os"
	"path/filepath"

	"github.com/opd-ai/venture/pkg/rendering/lighting"
)

var (
	outputDir  = flag.String("output", ".", "Output directory for test images")
	bloomOnly  = flag.Bool("bloom-only", false, "Generate only bloom test images")
	aoOnly     = flag.Bool("ao-only", false, "Generate only AO test images")
	seedFlag   = flag.Int64("seed", 12345, "Random seed for deterministic generation")
	width      = flag.Int("width", 400, "Image width")
	height     = flag.Int("height", 400, "Image height")
	allEffects = flag.Bool("all", false, "Generate all effect variations")
)

func main() {
	flag.Parse()

	log.Printf("Lighting Test Tool - Phase 17.1")
	log.Printf("Output: %s", *outputDir)
	log.Printf("Size: %dx%d", *width, *height)
	log.Printf("Seed: %d", *seedFlag)

	// Create output directory
	if err := os.MkdirAll(*outputDir, 0755); err != nil {
		log.Fatalf("Failed to create output directory: %v", err)
	}

	if !*aoOnly {
		log.Println("\n=== Bloom/Glow Tests ===")
		generateBloomTests()
	}

	if !*bloomOnly {
		log.Println("\n=== Ambient Occlusion Tests ===")
		generateAOTests()
	}

	if *allEffects {
		log.Println("\n=== Combined Effects Tests ===")
		generateCombinedTests()
	}

	log.Println("\n✅ All test images generated successfully")
}

// generateBloomTests creates bloom effect test images.
func generateBloomTests() {
	// Test 1: Single bright point
	log.Println("Generating bloom test 1: Single bright point")
	img1 := createBaseImage(*width, *height, color.RGBA{30, 30, 40, 255})
	drawCircle(img1, *width/2, *height/2, 20, color.RGBA{255, 220, 180, 255})

	config1 := lighting.DefaultBloomConfig()
	config1.Threshold = 0.7
	config1.Intensity = 1.5
	config1.Radius = 15

	bloom1 := lighting.ApplyBloom(img1, config1)
	saveImage(bloom1, "01_bloom_single_point.png")

	// Test 2: Multiple bright sources
	log.Println("Generating bloom test 2: Multiple bright sources")
	img2 := createBaseImage(*width, *height, color.RGBA{20, 25, 35, 255})

	positions := [][2]int{
		{*width / 4, *height / 4},
		{*width * 3 / 4, *height / 4},
		{*width / 4, *height * 3 / 4},
		{*width * 3 / 4, *height * 3 / 4},
	}

	colors := []color.RGBA{
		{255, 200, 100, 255}, // Warm
		{100, 200, 255, 255}, // Cool
		{255, 100, 200, 255}, // Magenta
		{200, 255, 100, 255}, // Yellow-green
	}

	for i, pos := range positions {
		drawCircle(img2, pos[0], pos[1], 15, colors[i])
	}

	config2 := lighting.DefaultBloomConfig()
	config2.Threshold = 0.75
	config2.Intensity = 1.2
	config2.Radius = 12

	bloom2 := lighting.ApplyBloom(img2, config2)
	saveImage(bloom2, "02_bloom_multiple_sources.png")

	// Test 3: Intensity comparison
	log.Println("Generating bloom test 3: Intensity variations")
	img3 := createBaseImage(*width, *height, color.RGBA{25, 25, 35, 255})
	drawCircle(img3, *width/2, *height/2, 25, color.RGBA{255, 240, 200, 255})

	intensities := []float64{0.5, 1.0, 1.5, 2.0}
	for _, intensity := range intensities {
		config := lighting.DefaultBloomConfig()
		config.Intensity = intensity

		result := lighting.ApplyBloom(img3, config)
		filename := fmt.Sprintf("03_bloom_intensity_%.1f.png", intensity)
		saveImage(result, filename)
	}

	// Test 4: Radius comparison
	log.Println("Generating bloom test 4: Radius variations")
	img4 := createBaseImage(*width, *height, color.RGBA{25, 25, 35, 255})
	drawCircle(img4, *width/2, *height/2, 25, color.RGBA{255, 240, 200, 255})

	radii := []int{5, 10, 15, 20}
	for _, radius := range radii {
		config := lighting.DefaultBloomConfig()
		config.Radius = radius

		result := lighting.ApplyBloom(img4, config)
		filename := fmt.Sprintf("04_bloom_radius_%d.png", radius)
		saveImage(result, filename)
	}
}

// generateAOTests creates ambient occlusion test images.
func generateAOTests() {
	// Test 1: Corner occlusion
	log.Println("Generating AO test 1: Corner detection")
	img1 := createBaseImage(*width, *height, color.RGBA{150, 150, 150, 255})

	// Create depth map with corners
	depthMap1 := createCornerDepthMap(*width, *height)

	config1 := lighting.DefaultEnhancedAOConfig()
	config1.Intensity = 0.6
	config1.CornerIntensity = 0.4
	config1.Seed = *seedFlag

	ao1 := lighting.ApplyEnhancedAO(img1, depthMap1, config1)
	saveImage(ao1, "05_ao_corner_detection.png")

	// Test 2: Edge occlusion
	log.Println("Generating AO test 2: Edge detection")
	img2 := createBaseImage(*width, *height, color.RGBA{150, 150, 150, 255})

	depthMap2 := createEdgeDepthMap(*width, *height)

	config2 := lighting.DefaultEnhancedAOConfig()
	config2.Intensity = 0.5
	config2.EdgeIntensity = 0.3
	config2.Seed = *seedFlag

	ao2 := lighting.ApplyEnhancedAO(img2, depthMap2, config2)
	saveImage(ao2, "06_ao_edge_detection.png")

	// Test 3: Sample count comparison
	log.Println("Generating AO test 3: Sample count variations")
	img3 := createBaseImage(*width, *height, color.RGBA{150, 150, 150, 255})
	depthMap3 := createComplexDepthMap(*width, *height)

	samples := []int{4, 8, 16, 32}
	for _, sampleCount := range samples {
		config := lighting.DefaultEnhancedAOConfig()
		config.Samples = sampleCount
		config.Seed = *seedFlag

		result := lighting.ApplyEnhancedAO(img3, depthMap3, config)
		filename := fmt.Sprintf("07_ao_samples_%d.png", sampleCount)
		saveImage(result, filename)
	}

	// Test 4: Intensity comparison
	log.Println("Generating AO test 4: Intensity variations")
	img4 := createBaseImage(*width, *height, color.RGBA{150, 150, 150, 255})
	depthMap4 := createComplexDepthMap(*width, *height)

	intensities := []float64{0.3, 0.5, 0.7, 0.9}
	for _, intensity := range intensities {
		config := lighting.DefaultEnhancedAOConfig()
		config.Intensity = intensity
		config.Seed = *seedFlag

		result := lighting.ApplyEnhancedAO(img4, depthMap4, config)
		filename := fmt.Sprintf("08_ao_intensity_%.1f.png", intensity)
		saveImage(result, filename)
	}
}

// generateCombinedTests creates images with both bloom and AO.
func generateCombinedTests() {
	log.Println("Generating combined test: Full lighting pipeline")

	// Create base image with some features
	img := createBaseImage(*width, *height, color.RGBA{40, 40, 50, 255})

	// Add bright light sources
	drawCircle(img, *width/3, *height/3, 20, color.RGBA{255, 220, 180, 255})
	drawCircle(img, *width*2/3, *height*2/3, 15, color.RGBA{180, 220, 255, 255})

	// Create depth map
	depthMap := createComplexDepthMap(*width, *height)

	// Apply AO first
	aoConfig := lighting.DefaultEnhancedAOConfig()
	aoConfig.Intensity = 0.5
	aoConfig.Seed = *seedFlag
	withAO := lighting.ApplyEnhancedAO(img, depthMap, aoConfig)

	// Then apply bloom
	bloomConfig := lighting.DefaultBloomConfig()
	bloomConfig.Intensity = 1.3
	bloomConfig.Radius = 12
	final := lighting.ApplyBloom(withAO, bloomConfig)

	saveImage(final, "09_combined_ao_bloom.png")

	// Save individual stages for comparison
	saveImage(img, "09a_base.png")
	saveImage(withAO, "09b_ao_only.png")
	saveImage(lighting.ApplyBloom(img, bloomConfig), "09c_bloom_only.png")
}

// Helper functions

func createBaseImage(width, height int, c color.Color) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, c)
		}
	}
	return img
}

func drawCircle(img *image.RGBA, cx, cy, radius int, c color.Color) {
	bounds := img.Bounds()
	for y := cy - radius; y <= cy+radius; y++ {
		for x := cx - radius; x <= cx+radius; x++ {
			if x < bounds.Min.X || x >= bounds.Max.X || y < bounds.Min.Y || y >= bounds.Max.Y {
				continue
			}

			dx := float64(x - cx)
			dy := float64(y - cy)
			dist := math.Sqrt(dx*dx + dy*dy)

			if dist <= float64(radius) {
				// Smooth edge with anti-aliasing
				alpha := 1.0
				if dist > float64(radius)-1 {
					alpha = float64(radius) - dist
				}

				r, g, b, a := c.RGBA()
				existing := img.At(x, y)
				er, eg, eb, ea := existing.RGBA()

				// Alpha blend
				finalR := uint8((float64(r)/257*alpha + float64(er)/257*(1-alpha)))
				finalG := uint8((float64(g)/257*alpha + float64(eg)/257*(1-alpha)))
				finalB := uint8((float64(b)/257*alpha + float64(eb)/257*(1-alpha)))
				finalA := uint8(math.Max(float64(a)/257, float64(ea)/257))

				img.Set(x, y, color.RGBA{R: finalR, G: finalG, B: finalB, A: finalA})
			}
		}
	}
}

func createCornerDepthMap(width, height int) *image.RGBA {
	depthMap := image.NewRGBA(image.Rect(0, 0, width, height))

	// Create depth variation with corners
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			// Create four quadrants with different depths
			depth := uint8(100)

			if x < width/2 && y < height/2 {
				depth = 50 // Top-left shallow
			} else if x >= width/2 && y < height/2 {
				depth = 150 // Top-right deep
			} else if x < width/2 && y >= height/2 {
				depth = 150 // Bottom-left deep
			} else {
				depth = 200 // Bottom-right very deep
			}

			depthMap.Set(x, y, color.RGBA{depth, depth, depth, 255})
		}
	}

	return depthMap
}

func createEdgeDepthMap(width, height int) *image.RGBA {
	depthMap := image.NewRGBA(image.Rect(0, 0, width, height))

	// Create horizontal edge in middle
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			depth := uint8(80)
			if y > height/2 {
				depth = 180
			}

			// Smooth transition
			if y > height/2-5 && y < height/2+5 {
				t := float64(y-(height/2-5)) / 10.0
				depth = uint8(80 + t*100)
			}

			depthMap.Set(x, y, color.RGBA{depth, depth, depth, 255})
		}
	}

	return depthMap
}

func createComplexDepthMap(width, height int) *image.RGBA {
	depthMap := image.NewRGBA(image.Rect(0, 0, width, height))

	cx, cy := width/2, height/2

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			// Radial gradient with noise
			dx := float64(x - cx)
			dy := float64(y - cy)
			dist := math.Sqrt(dx*dx + dy*dy)

			maxDist := math.Sqrt(float64(cx*cx + cy*cy))
			depth := dist / maxDist * 200

			// Add some variation
			noise := math.Sin(float64(x)*0.1) * math.Cos(float64(y)*0.1) * 20

			finalDepth := uint8(math.Max(0, math.Min(255, depth+noise)))
			depthMap.Set(x, y, color.RGBA{finalDepth, finalDepth, finalDepth, 255})
		}
	}

	return depthMap
}

func saveImage(img *image.RGBA, filename string) {
	path := filepath.Join(*outputDir, filename)

	f, err := os.Create(path)
	if err != nil {
		log.Printf("Failed to create %s: %v", filename, err)
		return
	}
	defer f.Close()

	if err := png.Encode(f, img); err != nil {
		log.Printf("Failed to encode %s: %v", filename, err)
		return
	}

	log.Printf("  ✓ Saved %s", filename)
}
