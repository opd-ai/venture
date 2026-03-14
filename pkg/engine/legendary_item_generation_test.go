package engine

import (
	"testing"

	"github.com/opd-ai/venture/pkg/procgen/item"
	"github.com/opd-ai/venture/pkg/world/raids"
)

func TestLegendaryQuestSystem_GenerateLegendaryItem(t *testing.T) {
	world := NewWorld()
	seed := int64(12345)
	raidManager := raids.NewManager(seed, "fantasy")
	system := NewLegendaryQuestSystem(world, seed, raidManager)

	// Test item generation
	itemID := "legendary_item_test_1"
	genreID := "fantasy"

	generatedItem := system.generateLegendaryItem(itemID, genreID)

	if generatedItem == nil {
		t.Fatal("generateLegendaryItem returned nil")
	}

	// Verify legendary rarity
	if generatedItem.Rarity != item.RarityLegendary {
		t.Errorf("Expected RarityLegendary, got %v", generatedItem.Rarity)
	}

	// Verify item has a name
	if generatedItem.Name == "" {
		t.Error("Generated item has empty name")
	}

	// Verify item has a description
	if generatedItem.Description == "" {
		t.Error("Generated item has empty description")
	}

	// Verify item has stats
	if generatedItem.Stats.Damage == 0 && generatedItem.Stats.Defense == 0 &&
		generatedItem.Stats.Value == 0 {
		t.Error("Generated item has no stats")
	}

	t.Logf("Generated legendary item: %s", generatedItem.Name)
	t.Logf("  Description: %s", generatedItem.Description)
	t.Logf("  Stats: Damage=%d, Defense=%d, Value=%d, Level=%d",
		generatedItem.Stats.Damage, generatedItem.Stats.Defense,
		generatedItem.Stats.Value, generatedItem.Stats.RequiredLevel)
}

func TestLegendaryQuestSystem_GenerateLegendaryItem_Determinism(t *testing.T) {
	world := NewWorld()
	seed := int64(42)
	raidManager := raids.NewManager(seed, "fantasy")
	system := NewLegendaryQuestSystem(world, seed, raidManager)

	itemID := "legendary_item_determinism_test"
	genreID := "scifi"

	// Generate same item twice
	item1 := system.generateLegendaryItem(itemID, genreID)
	item2 := system.generateLegendaryItem(itemID, genreID)

	if item1 == nil || item2 == nil {
		t.Fatal("generateLegendaryItem returned nil")
	}

	// Verify determinism: same itemID should produce same item
	if item1.Name != item2.Name {
		t.Errorf("Determinism failed: names differ (%s vs %s)", item1.Name, item2.Name)
	}

	if item1.Stats.Damage != item2.Stats.Damage {
		t.Errorf("Determinism failed: damage differs (%d vs %d)", item1.Stats.Damage, item2.Stats.Damage)
	}

	if item1.Stats.Defense != item2.Stats.Defense {
		t.Errorf("Determinism failed: defense differs (%d vs %d)", item1.Stats.Defense, item2.Stats.Defense)
	}
}

func TestLegendaryQuestSystem_GenerateLegendaryItem_DifferentGenres(t *testing.T) {
	world := NewWorld()
	seed := int64(99999)
	raidManager := raids.NewManager(seed, "fantasy")
	system := NewLegendaryQuestSystem(world, seed, raidManager)

	genres := []string{"fantasy", "scifi", "horror", "cyberpunk", "post-apocalyptic"}
	itemID := "legendary_item_genre_test"

	for _, genre := range genres {
		t.Run(genre, func(t *testing.T) {
			generatedItem := system.generateLegendaryItem(itemID, genre)

			if generatedItem == nil {
				t.Fatalf("generateLegendaryItem returned nil for genre %s", genre)
			}

			if generatedItem.Rarity != item.RarityLegendary {
				t.Errorf("Expected RarityLegendary for %s, got %v", genre, generatedItem.Rarity)
			}

			if generatedItem.Name == "" {
				t.Errorf("Generated item for %s has empty name", genre)
			}

			t.Logf("%s legendary: %s", genre, generatedItem.Name)
		})
	}
}

func TestLegendaryQuestSystem_CreateRewardItemByID(t *testing.T) {
	world := NewWorld()
	seed := int64(54321)
	raidManager := raids.NewManager(seed, "fantasy")
	system := NewLegendaryQuestSystem(world, seed, raidManager)

	// Create player entity with position
	player := world.CreateEntity()
	player.AddComponent(&PositionComponent{X: 100, Y: 200})

	// Add legendary quest component with genre
	questComp := &LegendaryQuestComponent{
		QuestID:         "test_quest",
		GenreID:         "fantasy",
		CurrentPhase:    0,
		PhasesCompleted: []bool{false, false, false},
	}
	player.AddComponent(questComp)

	// Create reward item
	itemID := "legendary_reward_test"
	system.createRewardItemByID(player, itemID)

	// Flush pending entities so the item appears in GetEntities()
	world.FlushPendingEntities()

	// Find the spawned item entity (should be the second entity after player)
	entities := world.GetEntities()
	if len(entities) < 2 {
		t.Fatalf("Expected at least 2 entities (player + spawned item), got %d", len(entities))
	}

	// Check the item entity
	itemEntity := entities[1]

	// Verify position component
	if !itemEntity.HasComponent("position") {
		t.Error("Item entity missing position component")
	}

	// Verify legendary item component
	if !itemEntity.HasComponent("legendary_item") {
		t.Fatal("Item entity missing legendary_item component")
	}

	comp, _ := itemEntity.GetComponent("legendary_item")
	legendaryComp := comp.(*LegendaryItemComponent)

	// Verify component fields are not placeholder values
	if legendaryComp.Name == "Legendary Item" {
		t.Error("Item still has placeholder name, expected procedurally generated name")
	}

	if legendaryComp.Description == "A legendary reward" {
		t.Error("Item still has placeholder description, expected procedurally generated description")
	}

	if len(legendaryComp.Stats) == 0 {
		t.Error("Item has no stats, expected procedurally generated stats")
	}

	if legendaryComp.Rarity != 3.0 {
		t.Errorf("Expected rarity 3.0 (legendary), got %.1f", legendaryComp.Rarity)
	}

	if !legendaryComp.Unique {
		t.Error("Legendary item should be unique")
	}

	t.Logf("Spawned legendary item: %s", legendaryComp.Name)
	t.Logf("  Description: %s", legendaryComp.Description)
	t.Logf("  Stats count: %d", len(legendaryComp.Stats))
}

func TestLegendaryQuestSystem_CreateRewardItemByID_NoPosition(t *testing.T) {
	world := NewWorld()
	seed := int64(11111)
	raidManager := raids.NewManager(seed, "fantasy")
	system := NewLegendaryQuestSystem(world, seed, raidManager)

	// Create player entity WITHOUT position
	player := world.CreateEntity()

	initialEntityCount := len(world.GetEntities())

	// Attempt to create reward item (should fail gracefully)
	itemID := "legendary_reward_no_pos"
	system.createRewardItemByID(player, itemID)

	// Verify no item was spawned
	finalEntityCount := len(world.GetEntities())
	if finalEntityCount != initialEntityCount {
		t.Error("Item should not spawn when player has no position component")
	}
}

func TestConvertItemStatsToIntMap(t *testing.T) {
	stats := item.Stats{
		Damage:        150,
		Defense:       100,
		AttackSpeed:   2.5,
		Value:         5000,
		Weight:        10.5,
		RequiredLevel: 50,
		Durability:    100,
	}

	result := convertItemStatsToIntMap(stats)

	tests := []struct {
		key      string
		expected int
	}{
		{"damage", 150},
		{"defense", 100},
		{"attack_speed", 25}, // 2.5 * 10
		{"value", 5000},
		{"weight", 105}, // 10.5 * 10
		{"required_level", 50},
		{"durability", 100},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			if val, ok := result[tt.key]; !ok {
				t.Errorf("Missing key %s in result map", tt.key)
			} else if val != tt.expected {
				t.Errorf("Expected %s=%d, got %d", tt.key, tt.expected, val)
			}
		})
	}
}

func TestHashString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantZero bool
	}{
		{"non-empty string", "legendary_item_123", false},
		{"empty string", "", true},
		{"single char", "a", false},
		{"unicode", "日本語", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := hashString(tt.input)

			if tt.wantZero && result != 0 {
				t.Errorf("Expected hash of empty string to be 0, got %d", result)
			}

			if !tt.wantZero && result == 0 {
				t.Errorf("Expected non-zero hash for %q, got 0", tt.input)
			}
		})
	}

	// Test determinism
	hash1 := hashString("test_determinism")
	hash2 := hashString("test_determinism")

	if hash1 != hash2 {
		t.Errorf("Hash function not deterministic: %d vs %d", hash1, hash2)
	}

	// Test uniqueness (different strings should have different hashes, usually)
	hash3 := hashString("different_string")
	if hash1 == hash3 {
		t.Logf("Warning: hash collision detected (rare but possible)")
	}
}

func TestLegendaryQuestSystem_ItemGeneratorInitialized(t *testing.T) {
	world := NewWorld()
	seed := int64(77777)
	raidManager := raids.NewManager(seed, "fantasy")
	system := NewLegendaryQuestSystem(world, seed, raidManager)

	if system.itemGen == nil {
		t.Fatal("ItemGenerator should be initialized in NewLegendaryQuestSystem")
	}
}

func TestLegendaryQuestComponent_WithGenreID(t *testing.T) {
	comp := &LegendaryQuestComponent{
		QuestID:         "test_quest_genre",
		GenreID:         "scifi",
		CurrentPhase:    2,
		PhasesCompleted: []bool{true, true, false},
	}

	if comp.GenreID != "scifi" {
		t.Errorf("Expected GenreID=scifi, got %s", comp.GenreID)
	}

	if comp.Type() != "legendary_quest" {
		t.Errorf("Expected Type()=legendary_quest, got %s", comp.Type())
	}
}

func BenchmarkGenerateLegendaryItem(b *testing.B) {
	world := NewWorld()
	seed := int64(88888)
	raidManager := raids.NewManager(seed, "fantasy")
	system := NewLegendaryQuestSystem(world, seed, raidManager)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		itemID := "bench_item"
		_ = system.generateLegendaryItem(itemID, "fantasy")
	}
}

func BenchmarkCreateRewardItemByID(b *testing.B) {
	world := NewWorld()
	seed := int64(99999)
	raidManager := raids.NewManager(seed, "fantasy")
	system := NewLegendaryQuestSystem(world, seed, raidManager)

	player := world.CreateEntity()
	player.AddComponent(&PositionComponent{X: 0, Y: 0})
	player.AddComponent(&LegendaryQuestComponent{
		QuestID: "bench_quest",
		GenreID: "fantasy",
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		system.createRewardItemByID(player, "bench_reward")
	}
}
