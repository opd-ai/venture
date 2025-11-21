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
	entityType, genre, size, detailed, verbose := parseFlags()
	template := sprites.SelectTemplate64(*entityType, *genre, *size, *detailed)

	displayTemplateHeader(*entityType, *genre, *size, *detailed, template.Name)
	displayBodyParts(template, *size, *verbose)
	displaySummaryStatistics(template, *entityType, *size)
	displayPhaseFeatures(*size, *detailed, *entityType)
	displayUsageExamples()

	os.Exit(0)
}

// parseFlags parses command-line flags and returns their values.
func parseFlags() (*string, *string, *int, *bool, *bool) {
	entityType := flag.String("entity", "humanoid", "Entity type (humanoid, quadruped, blob, mechanical)")
	genre := flag.String("genre", "fantasy", "Genre (fantasy, scifi, horror, cyberpunk, postapoc)")
	size := flag.Int("size", 64, "Sprite size in pixels (32, 48, 64, etc.)")
	detailed := flag.Bool("detailed", false, "Use detailed variant with facial features (humanoid only)")
	verbose := flag.Bool("verbose", false, "Verbose output showing template details")
	flag.Parse()
	return entityType, genre, size, detailed, verbose
}

// displayTemplateHeader prints the template information header.
func displayTemplateHeader(entityType, genre string, size int, detailed bool, templateName string) {
	fmt.Printf("Phase 45 Enhanced Sprite Template\n")
	fmt.Printf("==================================\n\n")
	fmt.Printf("Entity Type:    %s\n", entityType)
	fmt.Printf("Genre:          %s\n", genre)
	fmt.Printf("Sprite Size:    %dx%d\n", size, size)
	fmt.Printf("Detailed Mode:  %v\n", detailed)
	fmt.Printf("Template Name:  %s\n\n", templateName)
}

// displayBodyParts shows the body part layout and specifications.
func displayBodyParts(template sprites.AnatomicalTemplate, size int, verbose bool) {
	fmt.Printf("Body Part Layout:\n")
	fmt.Printf("-----------------\n")

	sortedParts := template.GetSortedParts()
	for _, item := range sortedParts {
		printBodyPartInfo(item.Part.String(), item.Spec, size, verbose)
	}
}

// printBodyPartInfo prints information for a single body part.
func printBodyPartInfo(part string, spec sprites.PartSpec, size int, verbose bool) {
	fmt.Printf("\n%s (Z-Index: %d)\n", part, spec.ZIndex)

	if spec.PreferredPixelSize != nil {
		width := spec.PreferredPixelSize.Width
		height := spec.PreferredPixelSize.Height
		percentage := float64(height) / float64(size) * 100
		fmt.Printf("  Dimensions:  %dx%d pixels (%.1f%% of sprite height)\n", width, height, percentage)
	} else {
		width := spec.RelativeWidth * float64(size)
		height := spec.RelativeHeight * float64(size)
		fmt.Printf("  Dimensions:  %.1fx%.1f pixels (relative)\n", width, height)
	}

	x := spec.RelativeX * float64(size)
	y := spec.RelativeY * float64(size)
	fmt.Printf("  Position:    (%.1f, %.1f)\n", x, y)

	if spec.ColorRole != "" {
		fmt.Printf("  Color Role:  %s\n", spec.ColorRole)
	}
	if spec.Opacity < 1.0 {
		fmt.Printf("  Opacity:     %.2f\n", spec.Opacity)
	}
	if spec.Rotation != 0 {
		fmt.Printf("  Rotation:    %.0f°\n", spec.Rotation)
	}
	if verbose && len(spec.ShapeTypes) > 0 {
		fmt.Printf("  Shapes:      %d types available\n", len(spec.ShapeTypes))
	}
}

// displaySummaryStatistics shows summary statistics for the template.
func displaySummaryStatistics(template sprites.AnatomicalTemplate, entityType string, size int) {
	fmt.Printf("\nSummary Statistics:\n")
	fmt.Printf("-------------------\n")
	sortedParts := template.GetSortedParts()
	fmt.Printf("Total Body Parts: %d\n", len(template.BodyPartLayout))
	fmt.Printf("Layers (Z-Index): %d\n", len(sortedParts))

	if isHumanoidEntity(entityType) {
		displayHumanoidProportion(sortedParts, size)
	}
}

// isHumanoidEntity checks if entity type is humanoid-based.
func isHumanoidEntity(entityType string) bool {
	return entityType == "humanoid" || entityType == "player" || entityType == "npc"
}

// displayHumanoidProportion calculates and displays body proportion for humanoid entities.
func displayHumanoidProportion(sortedParts []struct {
	Part sprites.BodyPart
	Spec sprites.PartSpec
}, size int,
) {
	var totalHeight float64
	for _, item := range sortedParts {
		if item.Spec.PreferredPixelSize != nil {
			totalHeight += float64(item.Spec.PreferredPixelSize.Height)
		}
	}
	if totalHeight > 0 {
		proportion := totalHeight / float64(size) * 100
		fmt.Printf("Body Proportion:  %.1f%% of sprite height\n", proportion)
	}
}

// displayPhaseFeatures shows Phase 45 specific features.
func displayPhaseFeatures(size int, detailed bool, entityType string) {
	fmt.Printf("\nPhase 45 Features:\n")
	fmt.Printf("------------------\n")
	if size >= 64 {
		fmt.Printf("✓ Enhanced 64x64 template selected\n")
		fmt.Printf("✓ Improved proportions: head 12%%, torso 40%%, legs 48%%\n")
		fmt.Printf("✓ Target silhouette score: 0.85+ (vs 0.75 for 32x32)\n")
		if detailed && (entityType == "humanoid" || entityType == "player") {
			fmt.Printf("✓ Facial features included (eyes, mouth)\n")
		}
	} else if size >= 48 {
		fmt.Printf("✓ Enhanced template selected (scaled for medium size)\n")
	} else {
		fmt.Printf("✓ Standard template selected (optimal for small sprites)\n")
	}
}

// displayUsageExamples prints usage examples.
func displayUsageExamples() {
	fmt.Printf("\nUsage Examples:\n")
	fmt.Printf("---------------\n")
	fmt.Printf("  ./sprite64test -entity=humanoid -genre=fantasy -size=64 -detailed\n")
	fmt.Printf("  ./sprite64test -entity=quadruped -genre=scifi -size=64\n")
	fmt.Printf("  ./sprite64test -entity=blob -genre=horror -size=64\n")
	fmt.Printf("  ./sprite64test -entity=mechanical -genre=cyberpunk -size=64\n")
	fmt.Printf("  ./sprite64test -entity=humanoid -genre=postapoc -size=32  # Fallback to standard\n")
	fmt.Printf("\nFor verbose output with shape details, add -verbose flag\n")
}
