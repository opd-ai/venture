// Package main provides a CLI tool for testing UI visual hierarchy features.
package main

import (
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"

	"github.com/opd-ai/venture/pkg/rendering/ui"
)

func main() {
	// Command line flags
	genreID := flag.String("genre", "fantasy", "Genre ID (fantasy, scifi, horror, cyberpunk, postapoc)")
	seed := flag.Int64("seed", 12345, "Random seed for deterministic generation")
	outputDir := flag.String("output", ".", "Output directory for test images")
	testType := flag.String("test", "all", "Test type (all, hierarchy, transitions, separators, groups, borders)")
	width := flag.Int("width", 800, "Image width")
	height := flag.Int("height", 600, "Image height")

	flag.Parse()

	// Create output directory if it doesn't exist
	if err := os.MkdirAll(*outputDir, 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create output directory: %v\n", err)
		os.Exit(1)
	}

	// Create generator
	gen := ui.NewGenerator()

	fmt.Printf("UI Visual Hierarchy Test Tool\n")
	fmt.Printf("Genre: %s, Seed: %d\n", *genreID, *seed)
	fmt.Printf("Output directory: %s\n\n", *outputDir)

	// Run requested tests
	switch *testType {
	case "all":
		testHierarchyLevels(gen, *genreID, *seed, *outputDir)
		testTransitions(gen, *genreID, *seed, *outputDir)
		testSeparators(gen, *genreID, *seed, *outputDir)
		testGroupContainers(gen, *genreID, *seed, *outputDir)
		testBorderStyles(gen, *genreID, *seed, *outputDir)
	case "hierarchy":
		testHierarchyLevels(gen, *genreID, *seed, *outputDir)
	case "transitions":
		testTransitions(gen, *genreID, *seed, *outputDir)
	case "separators":
		testSeparators(gen, *genreID, *seed, *outputDir)
	case "groups":
		testGroupContainers(gen, *genreID, *seed, *outputDir)
	case "borders":
		testBorderStyles(gen, *genreID, *seed, *outputDir)
	case "showcase":
		testShowcase(gen, *genreID, *seed, *outputDir, *width, *height)
	default:
		fmt.Fprintf(os.Stderr, "Unknown test type: %s\n", *testType)
		os.Exit(1)
	}

	fmt.Printf("\nTest complete! Images saved to %s\n", *outputDir)
}

func testHierarchyLevels(gen *ui.Generator, genreID string, seed int64, outputDir string) {
	fmt.Println("Testing hierarchy levels...")

	levels := []ui.HierarchyLevel{
		ui.HierarchyPrimary,
		ui.HierarchySecondary,
		ui.HierarchyTertiary,
		ui.HierarchyQuaternary,
	}

	for _, level := range levels {
		config := ui.Config{
			Type:           ui.ElementButton,
			Width:          200,
			Height:         60,
			GenreID:        genreID,
			Seed:           seed,
			Text:           fmt.Sprintf("%s Button", level.String()),
			HierarchyLevel: level,
		}

		img, err := gen.Generate(config)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to generate %s button: %v\n", level.String(), err)
			continue
		}

		filename := filepath.Join(outputDir, fmt.Sprintf("hierarchy_%s.png", level.String()))
		if err := savePNG(img, filename); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to save %s: %v\n", filename, err)
		} else {
			fmt.Printf("  ✓ Saved %s\n", filename)
		}
	}
}

func testTransitions(gen *ui.Generator, genreID string, seed int64, outputDir string) {
	fmt.Println("Testing transitions...")

	transitions := []struct {
		name   string
		tType  ui.TransitionType
		easing ui.EasingFunction
	}{
		{"fade_linear", ui.TransitionFade, ui.EaseLinear},
		{"fade_inout", ui.TransitionFade, ui.EaseInOutQuad},
		{"slide_left", ui.TransitionSlideLeft, ui.EaseOutQuad},
		{"slide_right", ui.TransitionSlideRight, ui.EaseInQuad},
		{"slide_up", ui.TransitionSlideUp, ui.EaseLinear},
		{"slide_down", ui.TransitionSlideDown, ui.EaseInOutCubic},
		{"zoom", ui.TransitionZoom, ui.EaseOutCubic},
	}

	for _, trans := range transitions {
		// Generate transitions at different progress levels
		for _, progress := range []float64{0.0, 0.25, 0.5, 0.75, 1.0} {
			transConfig := ui.TransitionConfig{
				Type:     trans.tType,
				Duration: 300,
				Easing:   trans.easing,
				Progress: progress,
			}

			config := ui.Config{
				Type:       ui.ElementPanel,
				Width:      150,
				Height:     100,
				GenreID:    genreID,
				Seed:       seed,
				Transition: &transConfig,
			}

			img, err := gen.Generate(config)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to generate %s transition: %v\n", trans.name, err)
				continue
			}

			filename := filepath.Join(outputDir, fmt.Sprintf("transition_%s_%.0f.png", trans.name, progress*100))
			if err := savePNG(img, filename); err != nil {
				fmt.Fprintf(os.Stderr, "Failed to save %s: %v\n", filename, err)
			} else {
				fmt.Printf("  ✓ Saved %s (progress: %.0f%%)\n", filename, progress*100)
			}
		}
	}
}

func testSeparators(gen *ui.Generator, genreID string, seed int64, outputDir string) {
	fmt.Println("Testing separators...")

	styles := []ui.SeparatorStyle{
		ui.SeparatorLine,
		ui.SeparatorDashed,
		ui.SeparatorDotted,
		ui.SeparatorGradient,
		ui.SeparatorOrnamental,
	}

	for _, style := range styles {
		img := gen.GenerateSeparator(400, 20, style, genreID, seed)

		filename := filepath.Join(outputDir, fmt.Sprintf("separator_%s.png", style.String()))
		if err := savePNG(img, filename); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to save %s: %v\n", filename, err)
		} else {
			fmt.Printf("  ✓ Saved %s\n", filename)
		}
	}
}

func testGroupContainers(gen *ui.Generator, genreID string, seed int64, outputDir string) {
	fmt.Println("Testing group containers...")

	configs := []struct {
		name           string
		level          ui.HierarchyLevel
		showBorder     bool
		showBackground bool
	}{
		{"primary_full", ui.HierarchyPrimary, true, true},
		{"secondary_border", ui.HierarchySecondary, true, false},
		{"tertiary_bg", ui.HierarchyTertiary, false, true},
		{"quaternary_full", ui.HierarchyQuaternary, true, true},
	}

	for _, cfg := range configs {
		groupCfg := ui.GroupConfig{
			Width:          300,
			Height:         200,
			Padding:        10,
			Level:          cfg.level,
			GenreID:        genreID,
			Seed:           seed,
			ShowBorder:     cfg.showBorder,
			ShowBackground: cfg.showBackground,
		}

		img := gen.GenerateGroupContainer(groupCfg)

		filename := filepath.Join(outputDir, fmt.Sprintf("group_%s.png", cfg.name))
		if err := savePNG(img, filename); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to save %s: %v\n", filename, err)
		} else {
			fmt.Printf("  ✓ Saved %s\n", filename)
		}
	}
}

func testBorderStyles(gen *ui.Generator, genreID string, seed int64, outputDir string) {
	fmt.Println("Testing border styles...")

	styles := []ui.BorderStyle{
		ui.BorderSolid,
		ui.BorderDouble,
		ui.BorderOrnate,
		ui.BorderGlow,
		ui.BorderDashed,
		ui.BorderDotted,
		ui.BorderEmbossed,
		ui.BorderEngraved,
	}

	for _, style := range styles {
		config := ui.Config{
			Type:    ui.ElementFrame,
			Width:   150,
			Height:  100,
			GenreID: genreID,
			Seed:    seed,
			Custom:  map[string]interface{}{"borderStyle": style},
		}

		img, err := gen.Generate(config)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to generate %s border: %v\n", style.String(), err)
			continue
		}

		filename := filepath.Join(outputDir, fmt.Sprintf("border_%s.png", style.String()))
		if err := savePNG(img, filename); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to save %s: %v\n", filename, err)
		} else {
			fmt.Printf("  ✓ Saved %s\n", filename)
		}
	}
}

func testShowcase(gen *ui.Generator, genreID string, seed int64, outputDir string, width, height int) {
	fmt.Println("Generating showcase image...")

	// Create a composite image showcasing all features
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	// Fill background
	bgColor := color.RGBA{R: 20, G: 20, B: 30, A: 255}
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, bgColor)
		}
	}

	// Add title with primary hierarchy
	titleConfig := ui.Config{
		Type:           ui.ElementButton,
		Width:          400,
		Height:         60,
		GenreID:        genreID,
		Seed:           seed,
		HierarchyLevel: ui.HierarchyPrimary,
	}
	titleImg, _ := gen.Generate(titleConfig)
	if titleImg != nil {
		drawImage(img, titleImg, 200, 20)
	}

	// Add separator
	sep1 := gen.GenerateSeparator(width-40, 10, ui.SeparatorOrnamental, genreID, seed)
	drawImage(img, sep1, 20, 100)

	// Add buttons with different hierarchy levels at different positions
	y := 130
	for _, level := range []ui.HierarchyLevel{ui.HierarchyPrimary, ui.HierarchySecondary, ui.HierarchyTertiary} {
		btnConfig := ui.Config{
			Type:           ui.ElementButton,
			Width:          200,
			Height:         50,
			GenreID:        genreID,
			Seed:           seed + int64(level),
			HierarchyLevel: level,
		}
		btnImg, _ := gen.Generate(btnConfig)
		if btnImg != nil {
			drawImage(img, btnImg, 50, y)
		}
		y += 70
	}

	// Add group container
	groupCfg := ui.GroupConfig{
		Width:          350,
		Height:         200,
		Padding:        10,
		Level:          ui.HierarchySecondary,
		GenreID:        genreID,
		Seed:           seed + 100,
		ShowBorder:     true,
		ShowBackground: true,
	}
	groupImg := gen.GenerateGroupContainer(groupCfg)
	drawImage(img, groupImg, 420, 130)

	// Save showcase
	filename := filepath.Join(outputDir, fmt.Sprintf("showcase_%s.png", genreID))
	if err := savePNG(img, filename); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to save showcase: %v\n", err)
	} else {
		fmt.Printf("  ✓ Saved %s\n", filename)
	}
}

func savePNG(img *image.RGBA, filename string) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	return png.Encode(f, img)
}

func drawImage(dst, src *image.RGBA, x, y int) {
	bounds := src.Bounds()
	for dy := 0; dy < bounds.Dy(); dy++ {
		for dx := 0; dx < bounds.Dx(); dx++ {
			px := x + dx
			py := y + dy
			if px >= 0 && px < dst.Bounds().Dx() && py >= 0 && py < dst.Bounds().Dy() {
				srcColor := src.RGBAAt(dx, dy)
				// Simple alpha blending
				if srcColor.A > 0 {
					dst.Set(px, py, srcColor)
				}
			}
		}
	}
}
