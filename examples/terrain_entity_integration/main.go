package main

import (
	"fmt"
	"log"

	"github.com/opd-ai/venture/pkg/procgen"
	"github.com/opd-ai/venture/pkg/procgen/entity"
	"github.com/opd-ai/venture/pkg/procgen/terrain"
)

// This example demonstrates how to integrate terrain and entity generation
// to create a complete dungeon level with placed enemies.

func main() {
	fmt.Println("=== Venture - Integrated Terrain & Entity Generation Example ===\n")

	// Use a fixed seed for reproducible results
	seed := int64(12345)
	depth := 5 // Dungeon level

	// Step 1: Generate terrain
	fmt.Println("Step 1: Generating terrain...")
	terrainGen := terrain.NewBSPGenerator()
	terrainParams := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      depth,
		Custom: map[string]interface{}{
			"width":  60,
			"height": 40,
		},
	}

	result, err := terrainGen.Generate(seed, terrainParams)
	if err != nil {
		log.Fatalf("Terrain generation failed: %v", err)
	}

	terr := result.(*terrain.Terrain)
	fmt.Printf("✓ Generated terrain: %dx%d with %d rooms\n\n", terr.Width, terr.Height, len(terr.Rooms))

	// Step 2: Generate entities (one per room)
	fmt.Println("Step 2: Generating entities...")
	entityGen := entity.NewEntityGenerator()
	entityParams := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      depth,
		GenreID:    "fantasy",
		Custom: map[string]interface{}{
			"count": len(terr.Rooms), // One entity per room
		},
	}

	// Use different seed for entities to ensure variety
	result, err = entityGen.Generate(seed+1000, entityParams)
	if err != nil {
		log.Fatalf("Entity generation failed: %v", err)
	}

	entities := result.([]*entity.Entity)
	fmt.Printf("✓ Generated %d entities\n\n", len(entities))

	// Step 3: Place entities in rooms
	fmt.Println("Step 3: Placing entities in rooms...")
	placements := placeEntitiesInRooms(terr, entities)

	// Step 4: Display the dungeon with entity placements
	fmt.Println("\nStep 4: Dungeon Overview")
	fmt.Println("========================\n")

	// Show room assignments
	for i, room := range terr.Rooms {
		if i < len(entities) {
			e := entities[i]
			cx, cy := room.Center()
			fmt.Printf("Room %d (%d,%d %dx%d) [center: %d,%d]:\n",
				i+1, room.X, room.Y, room.Width, room.Height, cx, cy)
			fmt.Printf("  Entity: %s (Lv.%d %s)\n",
				e.Name, e.Stats.Level, e.Type.String())
			fmt.Printf("  Stats: HP=%d/%d, DMG=%d, DEF=%d, SPD=%.1f\n",
				e.Stats.Health, e.Stats.MaxHealth, e.Stats.Damage, e.Stats.Defense, e.Stats.Speed)
			fmt.Printf("  Rarity: %s, Hostile: %v, Threat: %d/100\n\n",
				e.Rarity.String(), e.IsHostile(), e.GetThreatLevel())
		}
	}

	// Step 5: Show statistics
	fmt.Println("Dungeon Statistics")
	fmt.Println("==================")
	fmt.Printf("Total Rooms: %d\n", len(terr.Rooms))
	fmt.Printf("Total Entities: %d\n", len(entities))

	// Count entity types
	typeCount := make(map[entity.EntityType]int)
	for _, e := range entities {
		typeCount[e.Type]++
	}
	fmt.Printf("  Monsters: %d\n", typeCount[entity.TypeMonster])
	fmt.Printf("  Bosses: %d\n", typeCount[entity.TypeBoss])
	fmt.Printf("  Minions: %d\n", typeCount[entity.TypeMinion])
	fmt.Printf("  NPCs: %d\n", typeCount[entity.TypeNPC])

	// Calculate total threat
	totalThreat := 0
	for _, e := range entities {
		totalThreat += e.GetThreatLevel()
	}
	avgThreat := 0
	if len(entities) > 0 {
		avgThreat = totalThreat / len(entities)
	}
	fmt.Printf("\nTotal Threat: %d\n", totalThreat)
	fmt.Printf("Average Threat per Entity: %d/100\n", avgThreat)

	// Step 6: Render a small section of the dungeon with entity markers
	fmt.Println("\nDungeon Map (first 30x20):")
	fmt.Println("===========================")
	renderDungeonSection(terr, placements, 30, 20)

	fmt.Println("\nLegend:")
	fmt.Println("  # = Wall")
	fmt.Println("  . = Floor")
	fmt.Println("  : = Corridor")
	fmt.Println("  M = Monster")
	fmt.Println("  B = Boss")
	fmt.Println("  m = Minion")
	fmt.Println("  N = NPC")
}

// EntityPlacement represents an entity placed at a specific location
type EntityPlacement struct {
	Entity *entity.Entity
	X, Y   int
}

// placeEntitiesInRooms assigns entities to room centers
func placeEntitiesInRooms(terr *terrain.Terrain, entities []*entity.Entity) []EntityPlacement {
	placements := make([]EntityPlacement, 0, len(entities))

	for i, room := range terr.Rooms {
		if i >= len(entities) {
			break
		}

		cx, cy := room.Center()
		placements = append(placements, EntityPlacement{
			Entity: entities[i],
			X:      cx,
			Y:      cy,
		})
	}

	return placements
}

// renderDungeonSection renders a portion of the dungeon with entity markers
func renderDungeonSection(terr *terrain.Terrain, placements []EntityPlacement, width, height int) {
	// Create entity lookup map
	entityMap := make(map[int]map[int]*entity.Entity)
	for _, p := range placements {
		if entityMap[p.Y] == nil {
			entityMap[p.Y] = make(map[int]*entity.Entity)
		}
		entityMap[p.Y][p.X] = p.Entity
	}

	// Render
	maxX := min(width, terr.Width)
	maxY := min(height, terr.Height)

	for y := 0; y < maxY; y++ {
		for x := 0; x < maxX; x++ {
			// Check if there's an entity at this position
			if entityMap[y] != nil && entityMap[y][x] != nil {
				e := entityMap[y][x]
				switch e.Type {
				case entity.TypeMonster:
					fmt.Print("M")
				case entity.TypeBoss:
					fmt.Print("B")
				case entity.TypeMinion:
					fmt.Print("m")
				case entity.TypeNPC:
					fmt.Print("N")
				default:
					fmt.Print("?")
				}
			} else {
				// Show terrain
				tile := terr.GetTile(x, y)
				switch tile {
				case terrain.TileWall:
					fmt.Print("#")
				case terrain.TileFloor:
					fmt.Print(".")
				case terrain.TileCorridor:
					fmt.Print(":")
				case terrain.TileDoor:
					fmt.Print("+")
				default:
					fmt.Print("?")
				}
			}
		}
		fmt.Println()
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
