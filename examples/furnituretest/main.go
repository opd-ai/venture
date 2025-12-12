package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/opd-ai/venture/pkg/procgen"
	"github.com/opd-ai/venture/pkg/procgen/furniture"
)

func main() {
	// Command-line flags
	subType := flag.String("type", "", "Furniture subtype (e.g., Chair, Table, Chest)")
	genre := flag.String("genre", "fantasy", "Genre ID (fantasy, scifi, horror, cyberpunk, postapoc)")
	seed := flag.Int64("seed", 12345, "Random seed")
	depth := flag.Int("depth", 5, "Depth level (affects rarity)")
	difficulty := flag.Float64("difficulty", 0.5, "Difficulty (0.0-1.0, affects rarity)")
	listTypes := flag.Bool("list", false, "List all furniture types")
	listCategory := flag.String("category", "", "List furniture types in category")
	count := flag.Int("count", 1, "Number of furniture items to generate")
	testPlacement := flag.Bool("placement", false, "Test placement validation")
	roomWidth := flag.Float64("room-width", 16.0, "Room width for placement testing")
	roomHeight := flag.Float64("room-height", 12.0, "Room height for placement testing")

	flag.Parse()

	// List all types
	if *listTypes {
		printAllTypes()
		return
	}

	// List category
	if *listCategory != "" {
		printCategory(*listCategory)
		return
	}

	// Test placement
	if *testPlacement {
		testPlacementValidation(*subType, *genre, *seed, *depth, *difficulty, *roomWidth, *roomHeight, *count)
		return
	}

	// Generate furniture
	gen := furniture.NewGenerator()

	for i := 0; i < *count; i++ {
		params := procgen.GenerationParams{
			Difficulty: *difficulty,
			Depth:      *depth,
			GenreID:    *genre,
			Custom:     make(map[string]interface{}),
		}

		if *subType != "" {
			params.Custom["SubType"] = *subType
		}

		currentSeed := *seed + int64(i)

		result, err := gen.Generate(currentSeed, params)
		if err != nil {
			log.Fatalf("Generation error: %v", err)
		}

		item := result.(*furniture.Furniture)

		// Validate
		if err := gen.Validate(item); err != nil {
			log.Fatalf("Validation error: %v", err)
		}

		// Print furniture details
		printFurniture(item, currentSeed)

		if i < *count-1 {
			fmt.Println("---")
		}
	}
}

func printAllTypes() {
	fmt.Println("Available Furniture Types:")
	fmt.Println()

	categories := []furniture.FurnitureType{
		furniture.TypeSeating,
		furniture.TypeStorage,
		furniture.TypeCrafting,
		furniture.TypeDecoration,
		furniture.TypeLighting,
		furniture.TypeBedding,
		furniture.TypeTable,
		furniture.TypeUtility,
	}

	for _, category := range categories {
		subtypes := furniture.GetSubTypesByCategory(category)
		fmt.Printf("%s (%d types):\n", category.String(), len(subtypes))
		for _, st := range subtypes {
			fmt.Printf("  - %s\n", st)
		}
		fmt.Println()
	}

	total := len(furniture.GetAllSubTypes())
	fmt.Printf("Total: %d furniture types\n", total)
}

func printCategory(categoryName string) {
	// Map category name to type
	categoryMap := map[string]furniture.FurnitureType{
		"seating":    furniture.TypeSeating,
		"storage":    furniture.TypeStorage,
		"crafting":   furniture.TypeCrafting,
		"decoration": furniture.TypeDecoration,
		"lighting":   furniture.TypeLighting,
		"bedding":    furniture.TypeBedding,
		"table":      furniture.TypeTable,
		"utility":    furniture.TypeUtility,
	}

	category, ok := categoryMap[categoryName]
	if !ok {
		log.Fatalf("Unknown category: %s (valid: seating, storage, crafting, decoration, lighting, bedding, table, utility)", categoryName)
	}

	subtypes := furniture.GetSubTypesByCategory(category)
	fmt.Printf("%s Furniture Types:\n", category.String())
	for _, st := range subtypes {
		fmt.Printf("  %s\n", st)
	}
	fmt.Printf("Total: %d types\n", len(subtypes))
}

func printFurniture(item *furniture.Furniture, seed int64) {
	fmt.Printf("Seed: %d\n", seed)
	fmt.Printf("Name: %s\n", item.Name)
	fmt.Printf("Type: %s / %s\n", item.Type.String(), item.SubType)
	fmt.Printf("Material: %s\n", item.Material.String())
	fmt.Printf("Rarity: %s (detail multiplier: %.1fx)\n", item.Rarity.String(), item.Rarity.DetailMultiplier())
	fmt.Printf("Genre: %s\n", item.GenreID)
	fmt.Printf("Dimensions: %.2f × %.2f × %.2f tiles (W×H×D)\n", item.Width, item.Height, item.Depth)
	fmt.Printf("Collision: %.2f × %.2f tiles\n", item.CollisionWidth, item.CollisionDepth)
	fmt.Printf("Walkable: %v\n", item.Walkable)
	fmt.Printf("Functional: %v\n", item.Functional)

	if item.Capacity > 0 {
		fmt.Printf("Storage Capacity: %d slots\n", item.Capacity)
	}

	if item.LightIntensity > 0 {
		fmt.Printf("Light Intensity: %.2f\n", item.LightIntensity)
	}

	if item.CraftingType != "" {
		fmt.Printf("Crafting Type: %s\n", item.CraftingType)
	}

	fmt.Printf("Primary Color: RGBA(%d, %d, %d, %d)\n", item.PrimaryColor.R, item.PrimaryColor.G, item.PrimaryColor.B, item.PrimaryColor.A)
	fmt.Printf("Secondary Color: RGBA(%d, %d, %d, %d)\n", item.SecondaryColor.R, item.SecondaryColor.G, item.SecondaryColor.B, item.SecondaryColor.A)
	fmt.Printf("Detail Level: %.2f\n", item.DetailLevel)
	fmt.Printf("Description: %s\n", item.Description)
}

func testPlacementValidation(subType, genre string, seed int64, depth int, difficulty, roomWidth, roomHeight float64, count int) {
	fmt.Printf("Testing Placement Validation\n")
	fmt.Printf("Room Size: %.1f × %.1f tiles\n", roomWidth, roomHeight)
	fmt.Println()

	validator := furniture.NewPlacementValidator(roomWidth, roomHeight)
	gen := furniture.NewGenerator()

	placedCount := 0
	failedCount := 0

	for i := 0; i < count; i++ {
		params := procgen.GenerationParams{
			Difficulty: difficulty,
			Depth:      depth,
			GenreID:    genre,
			Custom:     make(map[string]interface{}),
		}

		if subType != "" {
			params.Custom["SubType"] = subType
		}

		currentSeed := seed + int64(i)
		result, err := gen.Generate(currentSeed, params)
		if err != nil {
			log.Printf("Generation error: %v", err)
			continue
		}

		item := result.(*furniture.Furniture)

		// Try to find valid placement
		x, y, dir, ok := validator.FindValidPlacement(item, furniture.DirNorth)
		if !ok {
			fmt.Printf("  ✗ Failed to place %s (%.1f×%.1f)\n", item.Name, item.CollisionWidth, item.CollisionDepth)
			failedCount++
			continue
		}

		// Place furniture
		if err := validator.PlaceFurniture(item, x, y, dir); err != nil {
			fmt.Printf("  ✗ Placement error for %s: %v\n", item.Name, err)
			failedCount++
			continue
		}

		placedCount++
		fmt.Printf("  ✓ Placed %s at (%.1f, %.1f) facing %s\n", item.Name, x, y, dir.String())
	}

	fmt.Println()
	fmt.Printf("Results:\n")
	fmt.Printf("  Placed: %d/%d\n", placedCount, count)
	fmt.Printf("  Failed: %d/%d\n", failedCount, count)
	fmt.Printf("  Room Occupancy: %.1f%%\n", validator.GetOccupancy())
}
