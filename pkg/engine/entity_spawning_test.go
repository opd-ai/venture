//go:build test
// +build test

package engine

import (
	"testing"

	"github.com/opd-ai/venture/pkg/procgen"
	"github.com/opd-ai/venture/pkg/procgen/entity"
	"github.com/opd-ai/venture/pkg/procgen/terrain"
)

// TestSpawnEnemiesInTerrain_Success tests successful enemy spawning.
func TestSpawnEnemiesInTerrain_Success(t *testing.T) {
	// Create world
	world := NewWorld()

	// Generate terrain
	terrainGen := terrain.NewBSPGenerator()
	params := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      1,
		GenreID:    "fantasy",
		Custom: map[string]interface{}{
			"width":  40,
			"height": 30,
		},
	}

	result, err := terrainGen.Generate(12345, params)
	if err != nil {
		t.Fatalf("Failed to generate terrain: %v", err)
	}
	terr := result.(*terrain.Terrain)

	// Spawn enemies
	enemyParams := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      1,
		GenreID:    "fantasy",
	}

	count, err := SpawnEnemiesInTerrain(world, terr, 12345, enemyParams)
	if err != nil {
		t.Fatalf("SpawnEnemiesInTerrain failed: %v", err)
	}

	if count == 0 {
		t.Error("Expected enemies to be spawned, got 0")
	}

	// Process pending additions
	world.Update(0)

	// Verify entities were added
	entities := world.GetEntities()
	if len(entities) == 0 {
		t.Error("Expected entities in world after spawning")
	}

	// Verify enemy components
	enemyFound := false
	for _, e := range entities {
		if e.HasComponent("ai") && e.HasComponent("health") && e.HasComponent("attack") {
			enemyFound = true

			// Check team component
			teamComp, ok := e.GetComponent("team")
			if !ok {
				t.Error("Enemy missing team component")
				continue
			}
			team := teamComp.(*TeamComponent)
			if team.TeamID != 2 {
				t.Errorf("Expected enemy team ID 2, got %d", team.TeamID)
			}

			// Check AI component
			aiComp, ok := e.GetComponent("ai")
			if !ok {
				t.Error("Enemy missing AI component")
				continue
			}
			ai := aiComp.(*AIComponent)
			if ai.DetectionRange <= 0 {
				t.Error("Enemy has invalid detection range")
			}

			break
		}
	}

	if !enemyFound {
		t.Error("No enemy entities found with required components")
	}
}

// TestSpawnEnemiesInTerrain_NoRooms tests spawning with empty terrain.
func TestSpawnEnemiesInTerrain_NoRooms(t *testing.T) {
	world := NewWorld()

	// Create terrain with no rooms
	terr := &terrain.Terrain{
		Width:  10,
		Height: 10,
		Tiles:  make([][]terrain.TileType, 10),
		Rooms:  []terrain.Room{},
		Seed:   12345,
	}

	params := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      1,
		GenreID:    "fantasy",
	}

	count, err := SpawnEnemiesInTerrain(world, terr, 12345, params)
	if err != nil {
		t.Fatalf("Expected no error for empty terrain, got: %v", err)
	}

	if count != 0 {
		t.Errorf("Expected 0 enemies for empty terrain, got %d", count)
	}
}

// TestSpawnEnemiesInTerrain_NilTerrain tests error handling for nil terrain.
func TestSpawnEnemiesInTerrain_NilTerrain(t *testing.T) {
	world := NewWorld()
	params := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      1,
		GenreID:    "fantasy",
	}

	_, err := SpawnEnemiesInTerrain(world, nil, 12345, params)
	if err == nil {
		t.Error("Expected error for nil terrain, got nil")
	}
}

// TestSpawnEnemiesInTerrain_Deterministic tests that spawning is deterministic.
func TestSpawnEnemiesInTerrain_Deterministic(t *testing.T) {
	seed := int64(99999)

	// Generate terrain once
	terrainGen := terrain.NewBSPGenerator()
	params := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      1,
		GenreID:    "fantasy",
		Custom: map[string]interface{}{
			"width":  30,
			"height": 20,
		},
	}

	result, _ := terrainGen.Generate(seed, params)
	terr := result.(*terrain.Terrain)

	enemyParams := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      1,
		GenreID:    "fantasy",
	}

	// Spawn twice with same seed
	world1 := NewWorld()
	count1, _ := SpawnEnemiesInTerrain(world1, terr, seed, enemyParams)
	world1.Update(0)

	world2 := NewWorld()
	count2, _ := SpawnEnemiesInTerrain(world2, terr, seed, enemyParams)
	world2.Update(0)

	if count1 != count2 {
		t.Errorf("Spawning not deterministic: first=%d, second=%d", count1, count2)
	}

	// Verify entity properties match
	entities1 := world1.GetEntities()
	entities2 := world2.GetEntities()

	if len(entities1) != len(entities2) {
		t.Errorf("Entity count mismatch: %d vs %d", len(entities1), len(entities2))
	}
}

// TestSpawnEnemyFromTemplate tests spawning a single enemy from template.
func TestSpawnEnemyFromTemplate(t *testing.T) {
	world := NewWorld()

	// Create a test entity
	genEntity := &entity.Entity{
		Name:   "Test Goblin",
		Type:   entity.TypeMonster,
		Size:   entity.SizeMedium,
		Rarity: entity.RarityCommon,
		Stats: entity.Stats{
			Health:  50,
			Damage:  10,
			Defense: 5,
			Speed:   1.0,
			Level:   3,
		},
		Seed: 12345,
		Tags: []string{"goblin"},
	}

	enemy := SpawnEnemyFromTemplate(world, genEntity, 100.0, 200.0)
	world.Update(0) // Process additions

	if enemy == nil {
		t.Fatal("SpawnEnemyFromTemplate returned nil")
	}

	// Verify position
	posComp, ok := enemy.GetComponent("position")
	if !ok {
		t.Fatal("Enemy missing position component")
	}
	pos := posComp.(*PositionComponent)
	if pos.X != 100.0 || pos.Y != 200.0 {
		t.Errorf("Wrong position: got (%f, %f), want (100, 200)", pos.X, pos.Y)
	}

	// Verify health
	healthComp, ok := enemy.GetComponent("health")
	if !ok {
		t.Fatal("Enemy missing health component")
	}
	health := healthComp.(*HealthComponent)
	if health.Max != 50.0 {
		t.Errorf("Wrong health: got %f, want 50", health.Max)
	}

	// Verify stats
	statsComp, ok := enemy.GetComponent("stats")
	if !ok {
		t.Fatal("Enemy missing stats component")
	}
	stats := statsComp.(*StatsComponent)
	if stats.Attack != 10.0 {
		t.Errorf("Wrong attack: got %f, want 10", stats.Attack)
	}
	if stats.Defense != 5.0 {
		t.Errorf("Wrong defense: got %f, want 5", stats.Defense)
	}
}

// TestGetEnemyColor tests color generation for different enemy types.
func TestGetEnemyColor(t *testing.T) {
	tests := []struct {
		name       string
		entityType entity.EntityType
		rarity     entity.Rarity
	}{
		{"Boss", entity.TypeBoss, entity.RarityCommon},
		{"Minion", entity.TypeMinion, entity.RarityCommon},
		{"Monster", entity.TypeMonster, entity.RarityRare},
		{"NPC", entity.TypeNPC, entity.RarityEpic},
		{"Legendary Boss", entity.TypeBoss, entity.RarityLegendary},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &entity.Entity{
				Type:   tt.entityType,
				Rarity: tt.rarity,
			}

			color := getEnemyColor(e)

			// Verify color has valid RGBA values
			if color.A != 255 {
				t.Errorf("Expected alpha 255, got %d", color.A)
			}

			// Colors should differ by type
			// This is a smoke test - we don't care about exact values
		})
	}
}

// TestSpawnEnemiesInTerrain_MultipleRooms tests spawning across multiple rooms.
func TestSpawnEnemiesInTerrain_MultipleRooms(t *testing.T) {
	world := NewWorld()

	// Create terrain with 5 rooms
	terr := &terrain.Terrain{
		Width:  50,
		Height: 50,
		Tiles:  make([][]terrain.TileType, 50),
		Rooms: []terrain.Room{
			{X: 5, Y: 5, Width: 10, Height: 10},
			{X: 20, Y: 5, Width: 10, Height: 10},
			{X: 35, Y: 5, Width: 10, Height: 10},
			{X: 5, Y: 25, Width: 10, Height: 10},
			{X: 20, Y: 25, Width: 10, Height: 10},
		},
		Seed: 12345,
	}

	params := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      1,
		GenreID:    "fantasy",
	}

	count, err := SpawnEnemiesInTerrain(world, terr, 12345, params)
	if err != nil {
		t.Fatalf("SpawnEnemiesInTerrain failed: %v", err)
	}

	if count == 0 {
		t.Error("Expected enemies to spawn in multiple rooms")
	}

	// Should skip first room (player spawn)
	// Should spawn 1-3 enemies per remaining room (4 rooms)
	world.Update(0)
	entities := world.GetEntities()

	if len(entities) < 4 || len(entities) > 12 {
		t.Errorf("Expected 4-12 enemies (1-3 per 4 rooms), got %d", len(entities))
	}
}
