//go:build !android && !ios
// +build !android,!ios

// Package main provides tests for server-side entity spawning.
package main

import (
	"image/color"
	"testing"

	"github.com/opd-ai/venture/pkg/engine"
	"github.com/opd-ai/venture/pkg/procgen"
	"github.com/opd-ai/venture/pkg/procgen/companion"
	"github.com/opd-ai/venture/pkg/procgen/vehicle"
)

// Note: createTestLogger and createTestTerrain are defined in player_management_test.go

// createTestWorld returns a minimal World for testing.
func createTestWorld() *engine.World {
	return engine.NewWorld()
}

// TestCalculateVehicleCount tests vehicle count calculation.
func TestCalculateVehicleCount(t *testing.T) {
	tests := []struct {
		name      string
		roomCount int
		expected  int
	}{
		{"2 rooms", 2, 2},
		{"3 rooms", 3, 2},
		{"6 rooms", 6, 3},
		{"10 rooms", 10, 4},
		{"18 rooms", 18, 5},
		{"22 rooms", 22, 5},     // capped at 5
		{"100 rooms", 100, 5},   // capped at 5
		{"1000 rooms", 1000, 5}, // capped at 5
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculateVehicleCount(tt.roomCount)
			if result != tt.expected {
				t.Errorf("calculateVehicleCount(%d) = %d, expected %d", tt.roomCount, result, tt.expected)
			}
		})
	}
}

// TestCalculateCompanionCount tests companion count calculation.
func TestCalculateCompanionCount(t *testing.T) {
	tests := []struct {
		name      string
		roomCount int
		expected  int
	}{
		{"2 rooms", 2, 1},
		{"3 rooms", 3, 1},
		{"7 rooms", 7, 2},
		{"12 rooms", 12, 3},
		{"17 rooms", 17, 3},     // capped at 3
		{"100 rooms", 100, 3},   // capped at 3
		{"1000 rooms", 1000, 3}, // capped at 3
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculateCompanionCount(tt.roomCount)
			if result != tt.expected {
				t.Errorf("calculateCompanionCount(%d) = %d, expected %d", tt.roomCount, result, tt.expected)
			}
		})
	}
}

// TestMapVehicleType tests vehicle type mapping.
func TestMapVehicleType(t *testing.T) {
	tests := []struct {
		name     string
		input    vehicle.VehicleType
		expected engine.VehicleType
	}{
		{"Mount", vehicle.TypeMount, engine.VehicleMount},
		{"Cart", vehicle.TypeCart, engine.VehicleCart},
		{"Boat", vehicle.TypeBoat, engine.VehicleBoat},
		{"Glider", vehicle.TypeGlider, engine.VehicleGlider},
		{"Mech", vehicle.TypeMech, engine.VehicleMech},
		{"Unknown defaults to Mount", vehicle.VehicleType(999), engine.VehicleMount},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := mapVehicleType(tt.input)
			if result != tt.expected {
				t.Errorf("mapVehicleType(%v) = %v, expected %v", tt.input, result, tt.expected)
			}
		})
	}
}

// TestExtractColor tests color extraction from uint32.
func TestExtractColor(t *testing.T) {
	tests := []struct {
		name     string
		input    uint32
		expected color.RGBA
	}{
		{"Black", 0x000000, color.RGBA{R: 0, G: 0, B: 0, A: 255}},
		{"White", 0xFFFFFF, color.RGBA{R: 255, G: 255, B: 255, A: 255}},
		{"Red", 0xFF0000, color.RGBA{R: 255, G: 0, B: 0, A: 255}},
		{"Green", 0x00FF00, color.RGBA{R: 0, G: 255, B: 0, A: 255}},
		{"Blue", 0x0000FF, color.RGBA{R: 0, G: 0, B: 255, A: 255}},
		{"Yellow", 0xFFFF00, color.RGBA{R: 255, G: 255, B: 0, A: 255}},
		{"Custom", 0x8B4513, color.RGBA{R: 139, G: 69, B: 19, A: 255}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractColor(tt.input)
			if result != tt.expected {
				t.Errorf("extractColor(0x%06X) = %v, expected %v", tt.input, result, tt.expected)
			}
		})
	}
}

// TestGetVehicleSizeForType tests vehicle size lookup.
func TestGetVehicleSizeForType(t *testing.T) {
	tests := []struct {
		name             string
		vehicleType      vehicle.VehicleType
		expectedSize     int
		expectedCollider float64
	}{
		{"Mount", vehicle.TypeMount, 32, 28.0},
		{"Cart", vehicle.TypeCart, 40, 36.0},
		{"Boat", vehicle.TypeBoat, 48, 44.0},
		{"Glider", vehicle.TypeGlider, 36, 32.0},
		{"Mech", vehicle.TypeMech, 44, 40.0},
		{"Unknown", vehicle.VehicleType(999), 32, 28.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			size, collider := getVehicleSizeForType(tt.vehicleType)
			if size != tt.expectedSize {
				t.Errorf("getVehicleSizeForType(%v) size = %d, expected %d", tt.vehicleType, size, tt.expectedSize)
			}
			if collider != tt.expectedCollider {
				t.Errorf("getVehicleSizeForType(%v) collider = %f, expected %f", tt.vehicleType, collider, tt.expectedCollider)
			}
		})
	}
}

// TestGetCompanionSizeForType tests companion size lookup.
func TestGetCompanionSizeForType(t *testing.T) {
	tests := []struct {
		name             string
		companionType    engine.CompanionType
		expectedSize     int
		expectedCollider float64
	}{
		{"Pet", engine.CompanionTypePet, 24, 20.0},
		{"Summon", engine.CompanionTypeSummon, 28, 24.0},
		{"Hireling", engine.CompanionTypeHireling, 28, 24.0},
		{"Elemental", engine.CompanionTypeElemental, 32, 28.0},
		{"Undead", engine.CompanionTypeUndead, 30, 26.0},
		{"Robot", engine.CompanionTypeRobot, 30, 26.0},
		{"Spirit", engine.CompanionTypeSpirit, 26, 22.0},
		{"Insect", engine.CompanionTypeInsect, 22, 18.0},
		{"Unknown", engine.CompanionType(999), 28, 24.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			size, collider := getCompanionSizeForType(tt.companionType)
			if size != tt.expectedSize {
				t.Errorf("getCompanionSizeForType(%v) size = %d, expected %d", tt.companionType, size, tt.expectedSize)
			}
			if collider != tt.expectedCollider {
				t.Errorf("getCompanionSizeForType(%v) collider = %f, expected %f", tt.companionType, collider, tt.expectedCollider)
			}
		})
	}
}

// TestSpawnVehiclesInTerrain_TooFewRooms tests that no vehicles are spawned with insufficient rooms.
func TestSpawnVehiclesInTerrain_TooFewRooms(t *testing.T) {
	tests := []struct {
		name      string
		roomCount int
	}{
		{"0 rooms", 0},
		{"1 room", 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			world := createTestWorld()
			terrainMap := createTestTerrain(tt.roomCount)
			logger := createTestLogger()
			params := procgen.GenerationParams{Difficulty: 0.5}

			spawned, err := spawnVehiclesInTerrain(world, terrainMap, 12345, params, logger)
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if spawned != 0 {
				t.Errorf("Expected 0 vehicles for %d rooms, got %d", tt.roomCount, spawned)
			}
		})
	}
}

// TestSpawnCompanionsInTerrain_TooFewRooms tests that no companions are spawned with insufficient rooms.
func TestSpawnCompanionsInTerrain_TooFewRooms(t *testing.T) {
	tests := []struct {
		name      string
		roomCount int
	}{
		{"0 rooms", 0},
		{"1 room", 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			world := createTestWorld()
			terrainMap := createTestTerrain(tt.roomCount)
			logger := createTestLogger()
			params := procgen.GenerationParams{Difficulty: 0.5}

			spawned, err := spawnCompanionsInTerrain(world, terrainMap, 12345, params, logger)
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if spawned != 0 {
				t.Errorf("Expected 0 companions for %d rooms, got %d", tt.roomCount, spawned)
			}
		})
	}
}

// TestSpawnBookshelvesInTerrain_TooFewRooms tests that no bookshelves are spawned with insufficient rooms.
func TestSpawnBookshelvesInTerrain_TooFewRooms(t *testing.T) {
	tests := []struct {
		name      string
		roomCount int
	}{
		{"0 rooms", 0},
		{"1 room", 1},
		{"2 rooms", 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			world := createTestWorld()
			terrainMap := createTestTerrain(tt.roomCount)
			logger := createTestLogger()
			params := procgen.GenerationParams{Difficulty: 0.5}

			spawned, err := spawnBookshelvesInTerrain(world, terrainMap, 12345, params, logger)
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if spawned != 0 {
				t.Errorf("Expected 0 bookshelves for %d rooms, got %d", tt.roomCount, spawned)
			}
		})
	}
}

// TestSpawnVehiclesInTerrain_Deterministic tests that spawning is deterministic with the same seed.
func TestSpawnVehiclesInTerrain_Deterministic(t *testing.T) {
	world1 := createTestWorld()
	world2 := createTestWorld()
	terrainMap := createTestTerrain(10) // 10 rooms
	logger := createTestLogger()
	params := procgen.GenerationParams{Difficulty: 0.5}
	seed := int64(42)

	// First spawn
	spawned1, err := spawnVehiclesInTerrain(world1, terrainMap, seed, params, logger)
	if err != nil {
		t.Fatalf("First spawn failed: %v", err)
	}

	// Second spawn with same seed should produce same count
	spawned2, err := spawnVehiclesInTerrain(world2, terrainMap, seed, params, logger)
	if err != nil {
		t.Fatalf("Second spawn failed: %v", err)
	}

	if spawned1 != spawned2 {
		t.Errorf("Determinism failure: first spawn = %d, second spawn = %d", spawned1, spawned2)
	}
}

// TestSpawnCompanionsInTerrain_Deterministic tests that companion spawning is deterministic.
func TestSpawnCompanionsInTerrain_Deterministic(t *testing.T) {
	world1 := createTestWorld()
	world2 := createTestWorld()
	terrainMap := createTestTerrain(10) // 10 rooms
	logger := createTestLogger()
	params := procgen.GenerationParams{Difficulty: 0.5}
	seed := int64(42)

	// First spawn
	spawned1, err := spawnCompanionsInTerrain(world1, terrainMap, seed, params, logger)
	if err != nil {
		t.Fatalf("First spawn failed: %v", err)
	}

	// Second spawn with same seed should produce same count
	spawned2, err := spawnCompanionsInTerrain(world2, terrainMap, seed, params, logger)
	if err != nil {
		t.Fatalf("Second spawn failed: %v", err)
	}

	if spawned1 != spawned2 {
		t.Errorf("Determinism failure: first spawn = %d, second spawn = %d", spawned1, spawned2)
	}
}

// TestSpawnBookshelvesInTerrain_Deterministic tests that bookshelf spawning is deterministic.
func TestSpawnBookshelvesInTerrain_Deterministic(t *testing.T) {
	world1 := createTestWorld()
	world2 := createTestWorld()
	terrainMap := createTestTerrain(10) // 10 rooms
	logger := createTestLogger()
	params := procgen.GenerationParams{Difficulty: 0.5}
	seed := int64(42)

	// First spawn
	spawned1, err := spawnBookshelvesInTerrain(world1, terrainMap, seed, params, logger)
	if err != nil {
		t.Fatalf("First spawn failed: %v", err)
	}

	// Second spawn with same seed should produce same count
	spawned2, err := spawnBookshelvesInTerrain(world2, terrainMap, seed, params, logger)
	if err != nil {
		t.Fatalf("Second spawn failed: %v", err)
	}

	if spawned1 != spawned2 {
		t.Errorf("Determinism failure: first spawn = %d, second spawn = %d", spawned1, spawned2)
	}
}

// TestSpawnVehiclesInTerrain_SuccessfulSpawn tests successful vehicle spawning.
func TestSpawnVehiclesInTerrain_SuccessfulSpawn(t *testing.T) {
	world := createTestWorld()
	terrainMap := createTestTerrain(10) // 10 rooms = vehicleCount of 4
	logger := createTestLogger()
	params := procgen.GenerationParams{Difficulty: 0.5}

	spawned, err := spawnVehiclesInTerrain(world, terrainMap, 12345, params, logger)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if spawned <= 0 {
		t.Errorf("Expected positive vehicle count for 10 rooms, got %d", spawned)
	}
}

// TestSpawnCompanionsInTerrain_SuccessfulSpawn tests successful companion spawning.
func TestSpawnCompanionsInTerrain_SuccessfulSpawn(t *testing.T) {
	world := createTestWorld()
	terrainMap := createTestTerrain(10) // 10 rooms
	logger := createTestLogger()
	params := procgen.GenerationParams{Difficulty: 0.5}

	spawned, err := spawnCompanionsInTerrain(world, terrainMap, 12345, params, logger)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if spawned <= 0 {
		t.Errorf("Expected positive companion count for 10 rooms, got %d", spawned)
	}
}

// TestSpawnBookshelvesInTerrain_SuccessfulSpawn tests successful bookshelf spawning.
func TestSpawnBookshelvesInTerrain_SuccessfulSpawn(t *testing.T) {
	world := createTestWorld()
	terrainMap := createTestTerrain(10) // 10 rooms
	logger := createTestLogger()
	params := procgen.GenerationParams{Difficulty: 0.5}

	spawned, err := spawnBookshelvesInTerrain(world, terrainMap, 12345, params, logger)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if spawned <= 0 {
		t.Errorf("Expected positive bookshelf count for 10 rooms, got %d", spawned)
	}
}

// TestCreateCompanionSpawnData tests companion spawn data creation.
func TestCreateCompanionSpawnData(t *testing.T) {
	comp := &companion.Companion{
		Name:     "TestCompanion",
		Type:     engine.CompanionTypePet,
		Level:    5,
		Attack:   10,
		Defense:  8,
		Speed:    1.2,
		HP:       100,
		MaxHP:    100,
		Loyalty:  75.0,
		Commands: []engine.CommandType{engine.CommandStay, engine.CommandAttack},
	}

	spawnData := createCompanionSpawnData(comp)

	if spawnData.Name != "TestCompanion" {
		t.Errorf("Expected name TestCompanion, got %s", spawnData.Name)
	}
	if spawnData.CompanionType != engine.CompanionTypePet {
		t.Errorf("Expected type Pet, got %v", spawnData.CompanionType)
	}
	if spawnData.Level != 5 {
		t.Errorf("Expected level 5, got %d", spawnData.Level)
	}
	if spawnData.Attack != 10 {
		t.Errorf("Expected attack 10, got %f", spawnData.Attack)
	}
	if spawnData.Defense != 8 {
		t.Errorf("Expected defense 8, got %f", spawnData.Defense)
	}
	if spawnData.HP != 100 {
		t.Errorf("Expected HP 100, got %f", spawnData.HP)
	}
	if spawnData.Size != 24 {
		t.Errorf("Expected size 24 for Pet, got %d", spawnData.Size)
	}
}

// TestConvertVehiclesToSpawnData tests vehicle spawn data conversion.
func TestConvertVehiclesToSpawnData(t *testing.T) {
	vehicles := []*vehicle.Vehicle{
		{
			Name:        "TestMount",
			VehicleType: vehicle.TypeMount,
			Color:       0xFF0000, // Red
		},
		{
			Name:        "TestCart",
			VehicleType: vehicle.TypeCart,
			Color:       0x00FF00, // Green
		},
	}

	spawnData := convertVehiclesToSpawnData(vehicles)

	if len(spawnData) != 2 {
		t.Fatalf("Expected 2 spawn data, got %d", len(spawnData))
	}

	if spawnData[0].Name != "TestMount" {
		t.Errorf("Expected first vehicle name TestMount, got %s", spawnData[0].Name)
	}
	if spawnData[0].VehicleType != engine.VehicleMount {
		t.Errorf("Expected first vehicle type Mount, got %v", spawnData[0].VehicleType)
	}
	if spawnData[0].Color.R != 255 {
		t.Errorf("Expected red color R=255, got %d", spawnData[0].Color.R)
	}

	if spawnData[1].Name != "TestCart" {
		t.Errorf("Expected second vehicle name TestCart, got %s", spawnData[1].Name)
	}
	if spawnData[1].VehicleType != engine.VehicleCart {
		t.Errorf("Expected second vehicle type Cart, got %v", spawnData[1].VehicleType)
	}
}

// TestSpawnVehiclesInTerrain_ManyRooms tests spawning with many rooms.
func TestSpawnVehiclesInTerrain_ManyRooms(t *testing.T) {
	world := createTestWorld()
	terrainMap := createTestTerrain(100) // 100 rooms
	logger := createTestLogger()
	params := procgen.GenerationParams{Difficulty: 0.5}

	spawned, err := spawnVehiclesInTerrain(world, terrainMap, 12345, params, logger)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	// With 100 rooms, vehicle count should be capped at 5
	if spawned > 5 {
		t.Errorf("Expected at most 5 vehicles (capped), got %d", spawned)
	}
}

// TestSpawnCompanionsInTerrain_ManyRooms tests spawning with many rooms.
func TestSpawnCompanionsInTerrain_ManyRooms(t *testing.T) {
	world := createTestWorld()
	terrainMap := createTestTerrain(100) // 100 rooms
	logger := createTestLogger()
	params := procgen.GenerationParams{Difficulty: 0.5}

	spawned, err := spawnCompanionsInTerrain(world, terrainMap, 12345, params, logger)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	// With 100 rooms, companion count should be capped at 3
	if spawned > 3 {
		t.Errorf("Expected at most 3 companions (capped), got %d", spawned)
	}
}

// Benchmarks

func BenchmarkCalculateVehicleCount(b *testing.B) {
	for i := 0; i < b.N; i++ {
		calculateVehicleCount(50)
	}
}

func BenchmarkCalculateCompanionCount(b *testing.B) {
	for i := 0; i < b.N; i++ {
		calculateCompanionCount(50)
	}
}

func BenchmarkExtractColor(b *testing.B) {
	for i := 0; i < b.N; i++ {
		extractColor(0xFF5733)
	}
}

func BenchmarkMapVehicleType(b *testing.B) {
	types := []vehicle.VehicleType{vehicle.TypeMount, vehicle.TypeCart, vehicle.TypeBoat, vehicle.TypeGlider, vehicle.TypeMech}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mapVehicleType(types[i%len(types)])
	}
}

func BenchmarkGetVehicleSizeForType(b *testing.B) {
	types := []vehicle.VehicleType{vehicle.TypeMount, vehicle.TypeCart, vehicle.TypeBoat, vehicle.TypeGlider, vehicle.TypeMech}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		getVehicleSizeForType(types[i%len(types)])
	}
}

func BenchmarkGetCompanionSizeForType(b *testing.B) {
	types := []engine.CompanionType{
		engine.CompanionTypePet, engine.CompanionTypeSummon, engine.CompanionTypeHireling,
		engine.CompanionTypeElemental, engine.CompanionTypeUndead, engine.CompanionTypeRobot,
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		getCompanionSizeForType(types[i%len(types)])
	}
}

// TestGenerateWorldFactions tests faction generation and registration.
func TestGenerateWorldFactions(t *testing.T) {
	tests := []struct {
		name      string
		seed      int64
		genreID   string
		expectErr bool
		minCount  int
	}{
		{
			name:      "fantasy factions",
			seed:      12345,
			genreID:   "fantasy",
			expectErr: false,
			minCount:  1,
		},
		{
			name:      "sci-fi factions",
			seed:      67890,
			genreID:   "sci-fi",
			expectErr: false,
			minCount:  1,
		},
		{
			name:      "zero seed",
			seed:      0,
			genreID:   "fantasy",
			expectErr: false,
			minCount:  1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			world := createTestWorld()
			logger := createTestLogger()

			// Add FactionSystem to world for registration
			factionSys := engine.NewFactionSystem(world, logger)
			world.AddSystem(factionSys)

			params := procgen.GenerationParams{
				Difficulty: 0.5,
				Depth:      1,
				GenreID:    tt.genreID,
				Custom:     map[string]interface{}{},
			}

			logEntry := logger.WithField("test", "faction_generation")

			count, err := generateWorldFactions(world, tt.seed, params, logEntry)

			if tt.expectErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.expectErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if !tt.expectErr && count < tt.minCount {
				t.Errorf("expected at least %d factions, got %d", tt.minCount, count)
			}
		})
	}
}

// TestGenerateWorldFactions_Deterministic verifies faction generation is deterministic.
func TestGenerateWorldFactions_Deterministic(t *testing.T) {
	const seed int64 = 42
	const genreID = "fantasy"

	logger := createTestLogger()

	world1 := createTestWorld()
	factionSys1 := engine.NewFactionSystem(world1, logger)
	world1.AddSystem(factionSys1)

	world2 := createTestWorld()
	factionSys2 := engine.NewFactionSystem(world2, logger)
	world2.AddSystem(factionSys2)

	params := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      1,
		GenreID:    genreID,
		Custom:     map[string]interface{}{},
	}

	logEntry := logger.WithField("test", "faction_determinism")

	count1, err1 := generateWorldFactions(world1, seed, params, logEntry)
	count2, err2 := generateWorldFactions(world2, seed, params, logEntry)

	if err1 != nil || err2 != nil {
		t.Fatalf("unexpected errors: err1=%v, err2=%v", err1, err2)
	}

	if count1 != count2 {
		t.Errorf("faction counts differ: %d vs %d (expected deterministic generation)", count1, count2)
	}
}

// Benchmark faction generation performance.
func BenchmarkGenerateWorldFactions(b *testing.B) {
	logger := createTestLogger()
	world := createTestWorld()
	factionSys := engine.NewFactionSystem(world, logger)
	world.AddSystem(factionSys)

	params := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      1,
		GenreID:    "fantasy",
		Custom:     map[string]interface{}{},
	}

	logEntry := logger.WithField("bench", "faction_generation")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = generateWorldFactions(world, int64(i), params, logEntry)
	}
}
