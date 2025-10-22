package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/opd-ai/venture/pkg/procgen/genre"
)

var (
	primaryID   = flag.String("primary", "fantasy", "Primary genre ID")
	secondaryID = flag.String("secondary", "scifi", "Secondary genre ID")
	weight      = flag.Float64("weight", 0.5, "Blend weight (0.0=all primary, 1.0=all secondary)")
	seed        = flag.Int64("seed", 12345, "Random seed for deterministic blending")
	preset      = flag.String("preset", "", "Use a preset blend (e.g., 'sci-fi-horror', 'dark-fantasy')")
	listPresets = flag.Bool("list-presets", false, "List all available preset blends")
	listGenres  = flag.Bool("list-genres", false, "List all available base genres")
	verbose     = flag.Bool("verbose", false, "Show detailed blend information")
)

func main() {
	flag.Parse()

	// List presets
	if *listPresets {
		showPresets()
		return
	}

	// List genres
	if *listGenres {
		showGenres()
		return
	}

	// Create blender
	registry := genre.DefaultRegistry()
	blender := genre.NewGenreBlender(registry)

	var blended *genre.BlendedGenre
	var err error

	// Blend genres
	if *preset != "" {
		fmt.Printf("Creating preset blend: %s (seed: %d)\n\n", *preset, *seed)
		blended, err = blender.CreatePresetBlend(*preset, *seed)
	} else {
		fmt.Printf("Blending %s + %s (weight: %.2f, seed: %d)\n\n",
			*primaryID, *secondaryID, *weight, *seed)
		blended, err = blender.Blend(*primaryID, *secondaryID, *weight, *seed)
	}

	if err != nil {
		log.Fatalf("Blend failed: %v", err)
	}

	// Display results
	showBlendedGenre(blended, *verbose)
}

func showGenres() {
	registry := genre.DefaultRegistry()
	fmt.Println("=== Available Base Genres ===\n")

	for _, g := range registry.All() {
		fmt.Printf("ID: %s\n", g.ID)
		fmt.Printf("Name: %s\n", g.Name)
		fmt.Printf("Description: %s\n", g.Description)
		fmt.Printf("Themes: %v\n", g.Themes)
		fmt.Println()
	}
}

func showPresets() {
	presets := genre.PresetBlends()
	fmt.Println("=== Available Preset Blends ===\n")

	for name, config := range presets {
		fmt.Printf("Name: %s\n", name)
		fmt.Printf("  Primary: %s\n", config.Primary)
		fmt.Printf("  Secondary: %s\n", config.Secondary)
		fmt.Printf("  Weight: %.2f\n", config.Weight)
		fmt.Println()
	}

	fmt.Println("Usage: genreblend -preset=<name> -seed=<seed>")
}

func showBlendedGenre(blended *genre.BlendedGenre, verbose bool) {
	fmt.Println("=== Blended Genre ===")
	fmt.Println()
	fmt.Printf("ID: %s\n", blended.ID)
	fmt.Printf("Name: %s\n", blended.Name)
	fmt.Printf("Description: %s\n", blended.Description)
	fmt.Println()

	fmt.Println("Themes:")
	for i, theme := range blended.Themes {
		fmt.Printf("  %d. %s\n", i+1, theme)
	}
	fmt.Println()

	fmt.Println("Color Palette:")
	fmt.Printf("  Primary:   %s %s\n", blended.PrimaryColor, colorBar(blended.PrimaryColor))
	fmt.Printf("  Secondary: %s %s\n", blended.SecondaryColor, colorBar(blended.SecondaryColor))
	fmt.Printf("  Accent:    %s %s\n", blended.AccentColor, colorBar(blended.AccentColor))
	fmt.Println()

	fmt.Println("Naming Prefixes:")
	fmt.Printf("  Entity:   %s\n", blended.EntityPrefix)
	fmt.Printf("  Item:     %s\n", blended.ItemPrefix)
	fmt.Printf("  Location: %s\n", blended.LocationPrefix)
	fmt.Println()

	if verbose {
		primary, secondary := blended.GetBaseGenres()
		fmt.Println("=== Base Genres ===")
		fmt.Println()
		fmt.Printf("Primary Genre: %s\n", primary.Name)
		fmt.Printf("  Themes: %v\n", primary.Themes)
		fmt.Printf("  Colors: %s, %s, %s\n",
			primary.PrimaryColor, primary.SecondaryColor, primary.AccentColor)
		fmt.Println()
		fmt.Printf("Secondary Genre: %s\n", secondary.Name)
		fmt.Printf("  Themes: %v\n", secondary.Themes)
		fmt.Printf("  Colors: %s, %s, %s\n",
			secondary.PrimaryColor, secondary.SecondaryColor, secondary.AccentColor)
		fmt.Println()
		fmt.Printf("Blend Weight: %.2f\n", blended.BlendWeight)
		if blended.BlendWeight < 0.33 {
			fmt.Printf("  (Primarily %s with %s elements)\n", primary.Name, secondary.Name)
		} else if blended.BlendWeight > 0.67 {
			fmt.Printf("  (Primarily %s with %s elements)\n", secondary.Name, primary.Name)
		} else {
			fmt.Printf("  (Equal blend of %s and %s)\n", primary.Name, secondary.Name)
		}
		fmt.Println()
	}

	fmt.Println("=== Example Content ===")
	fmt.Println()
	fmt.Printf("Entity Name: %s Destroyer\n", blended.EntityPrefix)
	fmt.Printf("Item Name: %s Blade\n", blended.ItemPrefix)
	fmt.Printf("Location Name: %s Fortress\n", blended.LocationPrefix)
	fmt.Println()
}

func colorBar(hexColor string) string {
	// Simple ASCII representation
	color := strings.ToLower(hexColor)
	if strings.Contains(color, "ff0000") || strings.Contains(color, "8b0000") {
		return "█████ (Red)"
	} else if strings.Contains(color, "00ff00") || strings.Contains(color, "008000") {
		return "█████ (Green)"
	} else if strings.Contains(color, "0000ff") || strings.Contains(color, "000080") {
		return "█████ (Blue)"
	} else if strings.Contains(color, "ffff00") || strings.Contains(color, "daa520") {
		return "█████ (Yellow/Gold)"
	} else if strings.Contains(color, "ff00ff") || strings.Contains(color, "ff1493") {
		return "█████ (Magenta/Pink)"
	} else if strings.Contains(color, "00ffff") || strings.Contains(color, "00ced1") {
		return "█████ (Cyan)"
	} else if strings.Contains(color, "ffffff") {
		return "█████ (White)"
	} else if strings.Contains(color, "000000") {
		return "█████ (Black)"
	}
	return "█████"
}

func init() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Genre Blender - Create hybrid genres for Venture\n\n")
		fmt.Fprintf(os.Stderr, "Usage:\n")
		fmt.Fprintf(os.Stderr, "  %s [options]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  # Blend fantasy and sci-fi equally\n")
		fmt.Fprintf(os.Stderr, "  %s -primary=fantasy -secondary=scifi -weight=0.5\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  # Create dark fantasy (30%% horror)\n")
		fmt.Fprintf(os.Stderr, "  %s -primary=fantasy -secondary=horror -weight=0.3\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  # Use a preset blend\n")
		fmt.Fprintf(os.Stderr, "  %s -preset=sci-fi-horror\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  # List all available presets\n")
		fmt.Fprintf(os.Stderr, "  %s -list-presets\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  # List all base genres\n")
		fmt.Fprintf(os.Stderr, "  %s -list-genres\n\n", os.Args[0])
	}
}
