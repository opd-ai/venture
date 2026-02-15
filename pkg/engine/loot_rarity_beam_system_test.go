package engine

import (
	"testing"

	"github.com/opd-ai/venture/pkg/procgen/item"
)

func TestNewLootRarityBeamSystem(t *testing.T) {
	world := NewWorld()
	sys := NewLootRarityBeamSystem(world, 42)

	if sys == nil {
		t.Fatal("expected non-nil system")
	}
	if sys.world != world {
		t.Error("world not set")
	}
	if sys.seed != 42 {
		t.Error("seed not set")
	}
	if len(sys.beamConfigs) != 4 {
		t.Errorf("expected 4 beam configs, got %d", len(sys.beamConfigs))
	}
	if sys.genreMultiplier != 1.0 {
		t.Errorf("expected default genre multiplier 1.0, got %f", sys.genreMultiplier)
	}
}

func TestLootRarityBeamSystem_SetGenre(t *testing.T) {
	tests := []struct {
		name       string
		genre      string
		wantMult   float64
	}{
		{"fantasy", "fantasy", 1.0},
		{"horror", "horror", 1.5},
		{"cyberpunk", "cyberpunk", 0.7},
		{"sci-fi", "sci-fi", 0.85},
		{"postapoc", "postapoc", 1.3},
		{"unknown", "steampunk", 1.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sys := NewLootRarityBeamSystem(NewWorld(), 100)
			sys.SetGenre(tt.genre)
			if sys.genreMultiplier != tt.wantMult {
				t.Errorf("genre %q: want multiplier %f, got %f", tt.genre, tt.wantMult, sys.genreMultiplier)
			}
			if sys.genreID != tt.genre {
				t.Errorf("genre not set: want %q, got %q", tt.genre, sys.genreID)
			}
		})
	}
}

func TestLootRarityBeamSystem_UpdateSkipsNoParticleSystem(t *testing.T) {
	world := NewWorld()
	sys := NewLootRarityBeamSystem(world, 1)
	// No particle system set - should not panic
	entity := NewEntity(1)
	entity.AddComponent(&ItemComponent{Item: &item.Item{Rarity: item.RarityLegendary}})
	entity.AddComponent(&PositionComponent{X: 10, Y: 20})
	sys.Update([]*Entity{entity}, 0.016)
	// No panic = pass
}

func TestLootRarityBeamSystem_UpdateSkipsCommonItems(t *testing.T) {
	world := NewWorld()
	sys := NewLootRarityBeamSystem(world, 1)
	ps := NewParticleSystem()
	sys.SetParticleSystem(ps)

	entity := NewEntity(1)
	entity.AddComponent(&ItemComponent{Item: &item.Item{Rarity: item.RarityCommon}})
	entity.AddComponent(&PositionComponent{X: 10, Y: 20})

	sys.Update([]*Entity{entity}, 0.016)

	// Common items should not generate cooldowns
	if _, exists := sys.cooldowns[entity.ID]; exists {
		t.Error("common items should not have cooldown entries")
	}
}

func TestLootRarityBeamSystem_UpdateSkipsPlayerEntities(t *testing.T) {
	world := NewWorld()
	sys := NewLootRarityBeamSystem(world, 1)
	ps := NewParticleSystem()
	sys.SetParticleSystem(ps)

	entity := NewEntity(1)
	entity.AddComponent(&ItemComponent{Item: &item.Item{Rarity: item.RarityEpic}})
	entity.AddComponent(&PositionComponent{X: 10, Y: 20})
	entity.AddComponent(&EbitenInput{})

	sys.Update([]*Entity{entity}, 0.016)

	// Entities with input (players) should be skipped
	if _, exists := sys.cooldowns[entity.ID]; exists {
		t.Error("player entities should not have cooldown entries")
	}
}

func TestLootRarityBeamSystem_UpdateSkipsEquippedEntities(t *testing.T) {
	world := NewWorld()
	sys := NewLootRarityBeamSystem(world, 1)
	ps := NewParticleSystem()
	sys.SetParticleSystem(ps)

	entity := NewEntity(1)
	entity.AddComponent(&ItemComponent{Item: &item.Item{Rarity: item.RarityRare}})
	entity.AddComponent(&PositionComponent{X: 10, Y: 20})
	entity.AddComponent(NewEquipmentComponent())

	sys.Update([]*Entity{entity}, 0.016)

	if _, exists := sys.cooldowns[entity.ID]; exists {
		t.Error("equipped entities should not have cooldown entries")
	}
}

func TestLootRarityBeamSystem_GetItemRarity(t *testing.T) {
	tests := []struct {
		name       string
		item       *item.Item
		wantRarity item.Rarity
		wantOK     bool
	}{
		{"legendary", &item.Item{Rarity: item.RarityLegendary}, item.RarityLegendary, true},
		{"epic", &item.Item{Rarity: item.RarityEpic}, item.RarityEpic, true},
		{"rare", &item.Item{Rarity: item.RarityRare}, item.RarityRare, true},
		{"uncommon", &item.Item{Rarity: item.RarityUncommon}, item.RarityUncommon, true},
		{"common", &item.Item{Rarity: item.RarityCommon}, item.RarityCommon, true},
		{"nil item", nil, item.RarityCommon, false},
	}

	sys := NewLootRarityBeamSystem(NewWorld(), 1)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entity := NewEntity(1)
			if tt.item != nil {
				entity.AddComponent(&ItemComponent{Item: tt.item})
			} else {
				entity.AddComponent(&ItemComponent{Item: nil})
			}

			rarity, ok := sys.getItemRarity(entity)
			if ok != tt.wantOK {
				t.Errorf("want ok=%v, got %v", tt.wantOK, ok)
			}
			if rarity != tt.wantRarity {
				t.Errorf("want rarity %v, got %v", tt.wantRarity, rarity)
			}
		})
	}
}

func TestLootRarityBeamSystem_BeamConfigScaling(t *testing.T) {
	sys := NewLootRarityBeamSystem(NewWorld(), 1)

	tests := []struct {
		rarity         item.Rarity
		wantColor      string
		wantMinCount   int
		wantMinHeight  float64
	}{
		{item.RarityUncommon, "green", 3, 20.0},
		{item.RarityRare, "blue", 5, 28.0},
		{item.RarityEpic, "purple", 8, 36.0},
		{item.RarityLegendary, "gold", 12, 44.0},
	}

	prevCount := 0
	prevHeight := 0.0
	for _, tt := range tests {
		t.Run(tt.rarity.String(), func(t *testing.T) {
			cfg, ok := sys.beamConfigs[tt.rarity]
			if !ok {
				t.Fatalf("no config for rarity %v", tt.rarity)
			}
			if cfg.ColorHint != tt.wantColor {
				t.Errorf("want color %q, got %q", tt.wantColor, cfg.ColorHint)
			}
			if cfg.ParticleCount < tt.wantMinCount {
				t.Errorf("want min count %d, got %d", tt.wantMinCount, cfg.ParticleCount)
			}
			if cfg.BeamHeight < tt.wantMinHeight {
				t.Errorf("want min height %f, got %f", tt.wantMinHeight, cfg.BeamHeight)
			}
			// Verify progressive scaling
			if cfg.ParticleCount <= prevCount {
				t.Error("particle count should increase with rarity")
			}
			if cfg.BeamHeight <= prevHeight {
				t.Error("beam height should increase with rarity")
			}
			prevCount = cfg.ParticleCount
			prevHeight = cfg.BeamHeight
		})
	}
}

func TestLootRarityBeamSystem_CooldownRespected(t *testing.T) {
	world := NewWorld()
	sys := NewLootRarityBeamSystem(world, 1)
	ps := NewParticleSystem()
	sys.SetParticleSystem(ps)

	entity := NewEntity(1)
	entity.AddComponent(&ItemComponent{Item: &item.Item{Rarity: item.RarityRare}})
	entity.AddComponent(&PositionComponent{X: 10, Y: 20})

	// First update should spawn and set cooldown
	sys.Update([]*Entity{entity}, 0.016)
	cd1, exists := sys.cooldowns[entity.ID]
	if !exists {
		t.Fatal("expected cooldown to be set after first update")
	}
	if cd1 <= 0 {
		t.Error("cooldown should be positive after spawn")
	}

	// Second update with small dt should not reset cooldown
	sys.Update([]*Entity{entity}, 0.1)
	cd2 := sys.cooldowns[entity.ID]
	if cd2 >= cd1 {
		t.Error("cooldown should decrease over time")
	}
}

func TestGenreEmitMultiplier(t *testing.T) {
	tests := []struct {
		genre string
		want  float64
	}{
		{"fantasy", 1.0},
		{"horror", 1.5},
		{"cyberpunk", 0.7},
		{"sci-fi", 0.85},
		{"postapoc", 1.3},
		{"", 1.0},
		{"unknown", 1.0},
	}

	for _, tt := range tests {
		t.Run(tt.genre, func(t *testing.T) {
			got := genreEmitMultiplier(tt.genre)
			if got != tt.want {
				t.Errorf("genreEmitMultiplier(%q) = %f, want %f", tt.genre, got, tt.want)
			}
		})
	}
}

func BenchmarkLootRarityBeamSystem_Update(b *testing.B) {
	world := NewWorld()
	sys := NewLootRarityBeamSystem(world, 42)
	ps := NewParticleSystem()
	sys.SetParticleSystem(ps)
	sys.SetGenre("fantasy")

	entities := make([]*Entity, 200)
	rarities := []item.Rarity{
		item.RarityCommon, item.RarityCommon, item.RarityCommon,
		item.RarityUncommon, item.RarityRare, item.RarityEpic,
	}
	for i := range entities {
		e := NewEntity(uint64(i + 1))
		e.AddComponent(&ItemComponent{Item: &item.Item{Rarity: rarities[i%len(rarities)]}})
		e.AddComponent(&PositionComponent{X: float64(i * 32), Y: float64(i * 16)})
		entities[i] = e
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.Update(entities, 0.016)
	}
}
