package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/opd-ai/venture/pkg/procgen"
	"github.com/opd-ai/venture/pkg/procgen/building"
)

func main() {
	seed := flag.Int64("seed", 12345, "Random seed")
	genreID := flag.String("genre", "fantasy", "Genre ID (fantasy, scifi, horror, cyberpunk, postapoc)")
	buildingType := flag.Int("type", 0, "Building type (0=House, 1=Workshop, 2=Storage, 3=Tower, 4=Manor)")
	verbose := flag.Bool("verbose", false, "Verbose output")
	flag.Parse()

	gen := building.NewGenerator()
	params := procgen.GenerationParams{
		GenreID: *genreID,
		Custom: map[string]interface{}{
			"buildingType": building.BuildingType(*buildingType),
		},
	}

	result, err := gen.Generate(*seed, params)
	if err != nil {
		log.Fatalf("Generation failed: %v", err)
	}

	b := result.(*building.Building)

	// Display building information
	fmt.Printf("=== Building Generator Test ===\n")
	fmt.Printf("Seed: %d\n", *seed)
	fmt.Printf("Genre: %s\n", *genreID)
	fmt.Printf("Type: %s\n", b.Type)
	fmt.Printf("Style: %s\n", b.Style)
	fmt.Printf("Dimensions: %dx%d tiles\n", b.Width, b.Height)
	fmt.Printf("Rooms: %d\n", len(b.Rooms))
	fmt.Printf("Doors: %d\n", len(b.Doors))
	fmt.Printf("Windows: %d\n", len(b.Windows))
	fmt.Printf("Roof: %s\n", b.RoofType)
	fmt.Printf("Navigable: %v\n\n", b.IsNavigable())

	if *verbose {
		fmt.Println("=== Room Details ===")
		for i, room := range b.Rooms {
			fmt.Printf("%d. %s at (%d,%d) size %dx%d\n",
				i+1, room.Type, room.X, room.Y, room.Width, room.Height)
		}

		fmt.Println("\n=== Door Details ===")
		for i, door := range b.Doors {
			fmt.Printf("%d. %s at (%d,%d)\n", i+1, door.Type, door.X, door.Y)
		}

		fmt.Println("\n=== Window Details ===")
		for i, window := range b.Windows {
			fmt.Printf("%d. %s at (%d,%d)\n", i+1, window.Type, window.X, window.Y)
		}
	}

	// Visual ASCII representation
	fmt.Println("\n=== Floor Plan ===")
	displayFloorPlan(b)

	// Validation
	fmt.Println("\n=== Validation ===")
	if err := gen.Validate(b); err != nil {
		fmt.Printf("❌ Validation failed: %v\n", err)
	} else {
		fmt.Println("✅ Building passes all validation checks")
	}
}

func displayFloorPlan(b *building.Building) {
	// Create grid
	grid := make([][]rune, b.Height)
	for y := 0; y < b.Height; y++ {
		grid[y] = make([]rune, b.Width)
		for x := 0; x < b.Width; x++ {
			grid[y][x] = ' '
		}
	}

	// Draw rooms
	for i, room := range b.Rooms {
		roomChar := getRoomChar(room.Type)
		for y := room.Y; y < room.Y+room.Height; y++ {
			for x := room.X; x < room.X+room.Width; x++ {
				if y < 0 || y >= b.Height || x < 0 || x >= b.Width {
					continue
				}
				grid[y][x] = roomChar
			}
		}
		// Mark room number in center
		centerY := room.Y + room.Height/2
		centerX := room.X + room.Width/2
		if centerY >= 0 && centerY < b.Height && centerX >= 0 && centerX < b.Width {
			grid[centerY][centerX] = rune('0' + (i % 10))
		}
	}

	// Draw doors
	for _, door := range b.Doors {
		if door.Y >= 0 && door.Y < b.Height && door.X >= 0 && door.X < b.Width {
			grid[door.Y][door.X] = 'D'
		}
	}

	// Draw windows
	for _, window := range b.Windows {
		if window.Y >= 0 && window.Y < b.Height && window.X >= 0 && window.X < b.Width {
			grid[window.Y][window.X] = 'W'
		}
	}

	// Print grid
	for y := 0; y < b.Height; y++ {
		for x := 0; x < b.Width; x++ {
			fmt.Printf("%c", grid[y][x])
		}
		fmt.Println()
	}

	// Legend
	fmt.Println("\nLegend:")
	fmt.Println("  E = Entrance")
	fmt.Println("  L = Living")
	fmt.Println("  B = Bedroom")
	fmt.Println("  S = Storage")
	fmt.Println("  K = Kitchen")
	fmt.Println("  w = Workshop")
	fmt.Println("  H = Hallway")
	fmt.Println("  T = Tower")
	fmt.Println("  D = Door")
	fmt.Println("  W = Window")
	fmt.Println("  0-9 = Room number")
}

func getRoomChar(rt building.RoomType) rune {
	switch rt {
	case building.RoomEntrance:
		return 'E'
	case building.RoomLiving:
		return 'L'
	case building.RoomBedroom:
		return 'B'
	case building.RoomStorage:
		return 'S'
	case building.RoomKitchen:
		return 'K'
	case building.RoomWorkshop:
		return 'w'
	case building.RoomHallway:
		return 'H'
	case building.RoomTower:
		return 'T'
	default:
		return '?'
	}
}
