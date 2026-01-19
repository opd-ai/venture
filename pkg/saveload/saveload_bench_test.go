//go:build !js
// +build !js

// Package saveload provides benchmark tests for save/load operations.
// Target: <1s for full world save, sub-100ms for typical saves.
package saveload

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// createBenchmarkSave creates a save with specified complexity level.
func createBenchmarkSave(complexity string) *GameSave {
	save := NewGameSave()
	save.Timestamp = time.Now()

	// Player state
	save.PlayerState.EntityID = 1
	save.PlayerState.X = 100.0
	save.PlayerState.Y = 200.0
	save.PlayerState.CurrentHealth = 80.0
	save.PlayerState.MaxHealth = 100.0
	save.PlayerState.Level = 10
	save.PlayerState.Experience = 5000
	save.PlayerState.Attack = 25.0
	save.PlayerState.Defense = 15.0
	save.PlayerState.MagicPower = 20.0
	save.PlayerState.Speed = 12.0
	save.PlayerState.Gold = 1500
	save.PlayerState.CurrentMana = 50
	save.PlayerState.MaxMana = 100

	// World state
	save.WorldState.Seed = 12345
	save.WorldState.GenreID = "fantasy"
	save.WorldState.Width = 200
	save.WorldState.Height = 200
	save.WorldState.GameTime = 3600.0
	save.WorldState.Difficulty = 0.5
	save.WorldState.Depth = 5

	switch complexity {
	case "minimal":
		// Just basic data
		return save

	case "small":
		// Add a few items and spells
		save.PlayerState.Items = createItems(10)
		save.PlayerState.Spells = createSpells(5)
		save.WorldState.ModifiedEntities = createModifiedEntities(20)
		return save

	case "medium":
		// Add moderate amount of data
		save.PlayerState.Items = createItems(50)
		save.PlayerState.Spells = createSpells(20)
		save.PlayerState.OwnedPlots = createHousingPlots(2)
		save.PlayerState.OwnedVehicles = createVehicles(3)
		save.PlayerState.ActiveCompanions = createCompanions(2)
		save.WorldState.ModifiedEntities = createModifiedEntities(100)
		save.WorldState.FogOfWar = createFogOfWar(100, 100)
		return save

	case "large":
		// Full complexity save
		save.PlayerState.Items = createItems(200)
		save.PlayerState.Spells = createSpells(50)
		save.PlayerState.OwnedPlots = createHousingPlots(5)
		save.PlayerState.OwnedVehicles = createVehicles(10)
		save.PlayerState.ActiveCompanions = createCompanions(5)
		save.PlayerState.TrustScores = createTrustScores(50)
		save.PlayerState.ReputationScores = createReputationScores(20)
		save.WorldState.ModifiedEntities = createModifiedEntities(500)
		save.WorldState.FogOfWar = createFogOfWar(200, 200)
		save.WorldState.GuildHalls = createGuildHalls(10)
		save.WorldState.TerritoryControl = createTerritoryControl(50)
		return save

	default:
		return save
	}
}

func createItems(count int) []ItemData {
	items := make([]ItemData, count)
	for i := 0; i < count; i++ {
		items[i] = ItemData{
			ID:          "item_" + string(rune('a'+i%26)),
			Name:        "Test Item",
			Type:        "weapon",
			Rarity:      "rare",
			Seed:        int64(i * 1000),
			Damage:      10 + i,
			Value:       100 + i*10,
			Weight:      1.5,
			Description: "A test item for benchmarking save/load operations.",
		}
	}
	return items
}

func createSpells(count int) []SpellData {
	spells := make([]SpellData, count)
	for i := 0; i < count; i++ {
		spells[i] = SpellData{
			Name:        "Spell " + string(rune('A'+i%26)),
			Type:        "offensive",
			Element:     "fire",
			Target:      "single",
			Rarity:      "uncommon",
			Seed:        int64(i * 2000),
			Damage:      20 + i*5,
			ManaCost:    10 + i,
			Cooldown:    2.0,
			CastTime:    0.5,
			Range:       10.0,
			Description: "A test spell for benchmarking.",
		}
	}
	return spells
}

func createModifiedEntities(count int) []ModifiedEntity {
	entities := make([]ModifiedEntity, count)
	for i := 0; i < count; i++ {
		entities[i] = ModifiedEntity{
			EntityID: uint64(1000 + i),
			X:        float64(i * 10),
			Y:        float64(i * 5),
			Health:   100.0 - float64(i%50),
			IsAlive:  i%10 != 0,
			IsPicked: i%5 == 0,
		}
	}
	return entities
}

func createHousingPlots(count int) []HousingPlotData {
	plots := make([]HousingPlotData, count)
	for i := 0; i < count; i++ {
		plots[i] = HousingPlotData{
			PlotID: "plot_" + string(rune('a'+i)),
			X:      float64(i * 100),
			Y:      float64(i * 100),
			Width:  64.0,
			Height: 64.0,
			Tier:   i + 1,
			Furniture: []FurnitureData{
				{FurnitureID: "furn_1", Type: "table", X: 10, Y: 10},
				{FurnitureID: "furn_2", Type: "chair", X: 20, Y: 10},
			},
		}
	}
	return plots
}

func createVehicles(count int) []VehicleData {
	vehicles := make([]VehicleData, count)
	for i := 0; i < count; i++ {
		vehicles[i] = VehicleData{
			VehicleID:     "vehicle_" + string(rune('a'+i)),
			Type:          "horse",
			Name:          "Mount " + string(rune('A'+i)),
			Health:        100.0,
			MaxHealth:     100.0,
			Speed:         20.0,
			Durability:    100,
			MaxDurability: 100,
			Seed:          int64(i * 3000),
		}
	}
	return vehicles
}

func createCompanions(count int) []CompanionData {
	companions := make([]CompanionData, count)
	for i := 0; i < count; i++ {
		companions[i] = CompanionData{
			EntityID:   uint64(5000 + i),
			Name:       "Companion " + string(rune('A'+i)),
			Type:       "warrior",
			Level:      5 + i,
			Experience: 1000 * (i + 1),
			Health:     80.0,
			MaxHealth:  100.0,
			Attack:     15.0,
			Defense:    10.0,
			Loyalty:    0.8,
			Seed:       int64(i * 4000),
		}
	}
	return companions
}

func createFogOfWar(width, height int) [][]bool {
	fog := make([][]bool, height)
	for y := 0; y < height; y++ {
		fog[y] = make([]bool, width)
		for x := 0; x < width; x++ {
			fog[y][x] = (x+y)%3 == 0 // Some explored, some not
		}
	}
	return fog
}

func createGuildHalls(count int) []GuildHallData {
	halls := make([]GuildHallData, count)
	for i := 0; i < count; i++ {
		halls[i] = GuildHallData{
			GuildID:  "guild_" + string(rune('a'+i)),
			X:        float64(i * 200),
			Y:        float64(i * 200),
			Width:    128.0,
			Height:   128.0,
			Tier:     i + 1,
			PlacedAt: time.Now(),
			Rooms: []RoomData{
				{RoomID: "room_1", Type: "main", X: 0, Y: 0, Width: 64, Height: 64},
				{RoomID: "room_2", Type: "storage", X: 64, Y: 0, Width: 32, Height: 32},
			},
		}
	}
	return halls
}

func createTrustScores(count int) map[string]float64 {
	scores := make(map[string]float64, count)
	for i := 0; i < count; i++ {
		scores["player_"+string(rune('a'+i%26))+string(rune('0'+i/26))] = 0.5 + float64(i%50)/100.0
	}
	return scores
}

func createReputationScores(count int) map[string]int {
	scores := make(map[string]int, count)
	categories := []string{"trade", "combat", "exploration", "crafting", "social"}
	for i := 0; i < count; i++ {
		scores[categories[i%len(categories)]+"_"+string(rune('0'+i))] = i * 100
	}
	return scores
}

func createTerritoryControl(count int) map[string]string {
	control := make(map[string]string, count)
	for i := 0; i < count; i++ {
		control["zone_"+string(rune('a'+i%26))+string(rune('0'+i/26))] = "guild_" + string(rune('a'+i%5))
	}
	return control
}

// BenchmarkSaveGameMinimal benchmarks saving minimal game state.
func BenchmarkSaveGameMinimal(b *testing.B) {
	tmpDir := b.TempDir()
	manager, err := NewSaveManager(tmpDir)
	if err != nil {
		b.Fatal(err)
	}

	save := createBenchmarkSave("minimal")

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		err := manager.SaveGame("bench_minimal", save)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkSaveGameSmall benchmarks saving small game state.
func BenchmarkSaveGameSmall(b *testing.B) {
	tmpDir := b.TempDir()
	manager, err := NewSaveManager(tmpDir)
	if err != nil {
		b.Fatal(err)
	}

	save := createBenchmarkSave("small")

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		err := manager.SaveGame("bench_small", save)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkSaveGameMedium benchmarks saving medium complexity game state.
func BenchmarkSaveGameMedium(b *testing.B) {
	tmpDir := b.TempDir()
	manager, err := NewSaveManager(tmpDir)
	if err != nil {
		b.Fatal(err)
	}

	save := createBenchmarkSave("medium")

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		err := manager.SaveGame("bench_medium", save)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkSaveGameLarge benchmarks saving large complexity game state.
func BenchmarkSaveGameLarge(b *testing.B) {
	tmpDir := b.TempDir()
	manager, err := NewSaveManager(tmpDir)
	if err != nil {
		b.Fatal(err)
	}

	save := createBenchmarkSave("large")

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		err := manager.SaveGame("bench_large", save)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkLoadGameMinimal benchmarks loading minimal game state.
func BenchmarkLoadGameMinimal(b *testing.B) {
	tmpDir := b.TempDir()
	manager, err := NewSaveManager(tmpDir)
	if err != nil {
		b.Fatal(err)
	}

	save := createBenchmarkSave("minimal")
	if err := manager.SaveGame("bench_minimal", save); err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := manager.LoadGame("bench_minimal")
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkLoadGameSmall benchmarks loading small game state.
func BenchmarkLoadGameSmall(b *testing.B) {
	tmpDir := b.TempDir()
	manager, err := NewSaveManager(tmpDir)
	if err != nil {
		b.Fatal(err)
	}

	save := createBenchmarkSave("small")
	if err := manager.SaveGame("bench_small", save); err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := manager.LoadGame("bench_small")
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkLoadGameMedium benchmarks loading medium complexity game state.
func BenchmarkLoadGameMedium(b *testing.B) {
	tmpDir := b.TempDir()
	manager, err := NewSaveManager(tmpDir)
	if err != nil {
		b.Fatal(err)
	}

	save := createBenchmarkSave("medium")
	if err := manager.SaveGame("bench_medium", save); err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := manager.LoadGame("bench_medium")
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkLoadGameLarge benchmarks loading large complexity game state.
func BenchmarkLoadGameLarge(b *testing.B) {
	tmpDir := b.TempDir()
	manager, err := NewSaveManager(tmpDir)
	if err != nil {
		b.Fatal(err)
	}

	save := createBenchmarkSave("large")
	if err := manager.SaveGame("bench_large", save); err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := manager.LoadGame("bench_large")
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkSaveLoadRoundTrip benchmarks full save/load cycle.
func BenchmarkSaveLoadRoundTrip(b *testing.B) {
	tmpDir := b.TempDir()
	manager, err := NewSaveManager(tmpDir)
	if err != nil {
		b.Fatal(err)
	}

	save := createBenchmarkSave("medium")

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		name := "bench_roundtrip"
		if err := manager.SaveGame(name, save); err != nil {
			b.Fatal(err)
		}
		_, err := manager.LoadGame(name)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkListSaves benchmarks listing save files.
func BenchmarkListSaves(b *testing.B) {
	tmpDir := b.TempDir()
	manager, err := NewSaveManager(tmpDir)
	if err != nil {
		b.Fatal(err)
	}

	// Create 20 save files
	for i := 0; i < 20; i++ {
		save := createBenchmarkSave("small")
		name := "save_" + string(rune('a'+i))
		if err := manager.SaveGame(name, save); err != nil {
			b.Fatal(err)
		}
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := manager.ListSaves()
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkGetSaveMetadata benchmarks reading save metadata.
func BenchmarkGetSaveMetadata(b *testing.B) {
	tmpDir := b.TempDir()
	manager, err := NewSaveManager(tmpDir)
	if err != nil {
		b.Fatal(err)
	}

	save := createBenchmarkSave("medium")
	if err := manager.SaveGame("bench_metadata", save); err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := manager.GetSaveMetadata("bench_metadata")
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkSaveExists benchmarks save existence check.
func BenchmarkSaveExists(b *testing.B) {
	tmpDir := b.TempDir()
	manager, err := NewSaveManager(tmpDir)
	if err != nil {
		b.Fatal(err)
	}

	save := createBenchmarkSave("minimal")
	if err := manager.SaveGame("exists_test", save); err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = manager.SaveExists("exists_test")
	}
}

// BenchmarkDeleteSave benchmarks save deletion.
func BenchmarkDeleteSave(b *testing.B) {
	tmpDir := b.TempDir()
	manager, err := NewSaveManager(tmpDir)
	if err != nil {
		b.Fatal(err)
	}

	save := createBenchmarkSave("minimal")

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		b.StopTimer()
		name := "delete_test"
		if err := manager.SaveGame(name, save); err != nil {
			b.Fatal(err)
		}
		b.StartTimer()

		if err := manager.DeleteSave(name); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkSaveGameParallel benchmarks parallel save operations.
func BenchmarkSaveGameParallel(b *testing.B) {
	tmpDir := b.TempDir()
	manager, err := NewSaveManager(tmpDir)
	if err != nil {
		b.Fatal(err)
	}

	save := createBenchmarkSave("small")

	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			name := "parallel_" + string(rune('a'+i%26)) + string(rune('0'+i/26%10))
			if err := manager.SaveGame(name, save); err != nil {
				b.Fatal(err)
			}
			i++
		}
	})
}

// BenchmarkSaveFileSizeEstimate reports approximate save file sizes.
func BenchmarkSaveFileSizeEstimate(b *testing.B) {
	tmpDir := b.TempDir()
	manager, err := NewSaveManager(tmpDir)
	if err != nil {
		b.Fatal(err)
	}

	complexities := []string{"minimal", "small", "medium", "large"}
	for _, complexity := range complexities {
		b.Run(complexity, func(b *testing.B) {
			save := createBenchmarkSave(complexity)
			name := "size_test_" + complexity
			if err := manager.SaveGame(name, save); err != nil {
				b.Fatal(err)
			}

			// Report file size
			info, err := os.Stat(filepath.Join(tmpDir, name+".sav"))
			if err != nil {
				b.Fatal(err)
			}
			b.ReportMetric(float64(info.Size()), "bytes")

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				if err := manager.SaveGame(name, save); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
