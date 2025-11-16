// Package main demonstrates Phase 45 enhanced 64x64 sprite generation.
//
// This tool showcases the improved sprite templates with better anatomical
// proportions (head 12%, torso 40%, legs 48%) compared to the 32x32 sprites.
//
// Usage:
//
//	go run ./cmd/sprite64test -entity=humanoid -genre=fantasy -size=64 -detailed
//	go run ./cmd/sprite64test -entity=quadruped -genre=scifi -size=64
//	go run ./cmd/sprite64test -entity=blob -genre=horror -size=64
//	go run ./cmd/sprite64test -entity=mechanical -genre=cyberpunk -size=64
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/opd-ai/venture/pkg/rendering/sprites"
)

func main() {
	// Command-line flags
	entityType := flag.String("entity", "humanoid", "Entity type (humanoid, quadruped, blob, mechanical)")
	genre := flag.String("genre", "fantasy", "Genre (fantasy, scifi, horror, cyberpunk, postapoc)")
	size := flag.Int("size", 64, "Sprite size in pixels (32, 48, 64, etc.)")
	detailed := flag.Bool("detailed", false, "Use detailed variant with facial features (humanoid only)")
	verbose := flag.Bool("verbose", false, "Verbose output showing template details")
	flag.Parse()

	// Select appropriate template using Phase 45 SelectTemplate64
	template := sprites.SelectTemplate64(*entityType, *genre, *size, *detailed)

	// Display template information
	fmt.Printf("Phase 45 Enhanced Sprite Template\n")
	fmt.Printf("==================================\n\n")
	fmt.Printf("Entity Type:    %s\n", *entityType)
	fmt.Printf("Genre:          %s\n", *genre)
	fmt.Printf("Sprite Size:    %dx%d\n", *size, *size)
	fmt.Printf("Detailed Mode:  %v\n", *detailed)
	fmt.Printf("Template Name:  %s\n\n", template.Name)

	// Display body parts and their specifications
	fmt.Printf("Body Part Layout:\n")
	fmt.Printf("-----------------\n")

	sortedParts := template.GetSortedParts()
	for _, item := range sortedParts {
		part := item.Part
		spec := item.Spec

		fmt.Printf("\n%s (Z-Index: %d)\n", part, spec.ZIndex)

		// Show pixel dimensions if available
		if spec.PreferredPixelSize != nil {
			width := spec.PreferredPixelSize.Width
			height := spec.PreferredPixelSize.Height
			percentage := float64(height) / float64(*size) * 100
			fmt.Printf("  Dimensions:  %dx%d pixels (%.1f%% of sprite height)\n", width, height, percentage)
		} else {
			width := spec.RelativeWidth * float64(*size)
			height := spec.RelativeHeight * float64(*size)
			fmt.Printf("  Dimensions:  %.1fx%.1f pixels (relative)\n", width, height)
		}

		// Show position
		x := spec.RelativeX * float64(*size)
		y := spec.RelativeY * float64(*size)
		fmt.Printf("  Position:    (%.1f, %.1f)\n", x, y)

		// Show color role and opacity
		if spec.ColorRole != "" {
			fmt.Printf("  Color Role:  %s\n", spec.ColorRole)
		}
		if spec.Opacity < 1.0 {
			fmt.Printf("  Opacity:     %.2f\n", spec.Opacity)
		}

		// Show rotation if non-zero
		if spec.Rotation != 0 {
			fmt.Printf("  Rotation:    %.0f°\n", spec.Rotation)
		}

		// Show shape types in verbose mode
		if *verbose && len(spec.ShapeTypes) > 0 {
			fmt.Printf("  Shapes:      %d types available\n", len(spec.ShapeTypes))
		}
	}

	// Display summary statistics
	fmt.Printf("\nSummary Statistics:\n")
	fmt.Printf("-------------------\n")
	fmt.Printf("Total Body Parts: %d\n", len(template.BodyPartLayout))
	fmt.Printf("Layers (Z-Index): %d\n", len(sortedParts))

	// Calculate total height proportion for humanoid templates
	if *entityType == "humanoid" || *entityType == "player" || *entityType == "npc" {
		var totalHeight float64
		for _, item := range sortedParts {
			spec := item.Spec
			if spec.PreferredPixelSize != nil {
				totalHeight += float64(spec.PreferredPixelSize.Height)
			}
		}
		if totalHeight > 0 {
			proportion := totalHeight / float64(*size) * 100
			fmt.Printf("Body Proportion:  %.1f%% of sprite height\n", proportion)
		}
	}

	fmt.Printf("\nPhase 45 Features:\n")
	fmt.Printf("------------------\n")
	if *size >= 64 {
		fmt.Printf("✓ Enhanced 64x64 template selected\n")
		fmt.Printf("✓ Improved proportions: head 12%%, torso 40%%, legs 48%%\n")
		fmt.Printf("✓ Target silhouette score: 0.85+ (vs 0.75 for 32x32)\n")
		if *detailed && (*entityType == "humanoid" || *entityType == "player") {
			fmt.Printf("✓ Facial features included (eyes, mouth)\n")
		}
	} else if *size >= 48 {
		fmt.Printf("✓ Enhanced template selected (scaled for medium size)\n")
	} else {
		fmt.Printf("✓ Standard template selected (optimal for small sprites)\n")
	}

	fmt.Printf("\nUsage Examples:\n")
	fmt.Printf("---------------\n")
	fmt.Printf("  ./sprite64test -entity=humanoid -genre=fantasy -size=64 -detailed\n")
	fmt.Printf("  ./sprite64test -entity=quadruped -genre=scifi -size=64\n")
	fmt.Printf("  ./sprite64test -entity=blob -genre=horror -size=64\n")
	fmt.Printf("  ./sprite64test -entity=mechanical -genre=cyberpunk -size=64\n")
	fmt.Printf("  ./sprite64test -entity=humanoid -genre=postapoc -size=32  # Fallback to standard\n")

	fmt.Printf("\nFor verbose output with shape details, add -verbose flag\n")

	os.Exit(0)
}
