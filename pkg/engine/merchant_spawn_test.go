package engine

import (
	"testing"

	"github.com/opd-ai/venture/pkg/procgen"
	procgenEntity "github.com/opd-ai/venture/pkg/procgen/entity"
	"github.com/opd-ai/venture/pkg/procgen/item"
	"github.com/opd-ai/venture/pkg/procgen/terrain"
)

func TestSpawnMerchantFromData(t *testing.T) {
	tests := []struct {
		name              string
		setupMerchant     func() *procgenEntity.MerchantData
		x, y              float64
		expectNil         bool
		expectedCompCount int
	}{
		{
			name: "valid merchant data",
			setupMerchant: func() *procgenEntity.MerchantData {
				return &procgenEntity.MerchantData{
					Entity: &procgenEntity.Entity{
						Name: "Test Merchant",
						Seed: 12345,
						Stats: procgenEntity.Stats{
							Health:    100,
							MaxHealth: 100,
						},
					},
					MerchantType: procgenEntity.MerchantFixed,
					Inventory: []*item.Item{
						{Name: "Sword", Stats: item.Stats{Value: 50}},
						{Name: "Potion", Stats: item.Stats{Value: 10}},
					},
					PriceMultiplier:   1.5,
					BuyBackPercentage: 0.5,
				}
			},
			x:                 100.0,
			y:                 200.0,
			expectNil:         false,
			expectedCompCount: 8, // position, velocity, health, team, sprite, animation, collider, merchant, dialog
		},
		{
			name: "nil merchant data",
			setupMerchant: func() *procgenEntity.MerchantData {
				return nil
			},
			x:         100.0,
			y:         200.0,
			expectNil: true,
		},
		{
			name: "nil entity in merchant data",
			setupMerchant: func() *procgenEntity.MerchantData {
				return &procgenEntity.MerchantData{
					Entity:          nil,
					MerchantType:    procgenEntity.MerchantFixed,
					Inventory:       []*item.Item{},
					PriceMultiplier: 1.5,
				}
			},
			x:         100.0,
			y:         200.0,
			expectNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			world := NewWorld()
			merchantData := tt.setupMerchant()

			result := SpawnMerchantFromData(world, merchantData, tt.x, tt.y)

			if tt.expectNil {
				if result != nil {
					t.Errorf("expected nil result, got entity %d", result.ID)
				}
				return
			}

			if result == nil {
				t.Fatal("expected merchant entity, got nil")
			}

			// Verify position
			posComp, ok := result.GetComponent("position")
			if !ok {
				t.Fatal("merchant missing position component")
			}
			pos := posComp.(*PositionComponent)
			if pos.X != tt.x || pos.Y != tt.y {
				t.Errorf("position = (%.0f, %.0f), want (%.0f, %.0f)", pos.X, pos.Y, tt.x, tt.y)
			}

			// Verify merchant component
			merchComp, ok := result.GetComponent("merchant")
			if !ok {
				t.Fatal("merchant missing merchant component")
			}
			merchant := merchComp.(*MerchantComponent)

			if merchant.MerchantName != merchantData.Entity.Name {
				t.Errorf("merchant name = %s, want %s", merchant.MerchantName, merchantData.Entity.Name)
			}

			if len(merchant.Inventory) != len(merchantData.Inventory) {
				t.Errorf("inventory size = %d, want %d", len(merchant.Inventory), len(merchantData.Inventory))
			}

			// Verify dialog component
			if !result.HasComponent("dialog") {
				t.Error("merchant missing dialog component")
			}

			// Verify collider
			if !result.HasComponent("collider") {
				t.Error("merchant missing collider component")
			}

			// Verify animation
			if !result.HasComponent("animation") {
				t.Error("merchant missing animation component")
			}
		})
	}
}

func TestSpawnMerchantsInTerrain(t *testing.T) {
	// Create a simple test terrain
	testTerrain := &terrain.Terrain{
		Width:  20,
		Height: 15,
		Tiles:  make([][]terrain.TileType, 15),
		Rooms:  make([]*terrain.Room, 0),
	}

	// Initialize tiles as walkable
	for y := 0; y < 15; y++ {
		testTerrain.Tiles[y] = make([]terrain.TileType, 20)
		for x := 0; x < 20; x++ {
			testTerrain.Tiles[y][x] = terrain.TileFloor
		}
	}

	tests := []struct {
		name          string
		terrain       *terrain.Terrain
		worldSeed     int64
		merchantCount int
		expectedMin   int // Minimum merchants spawned
		expectedMax   int // Maximum merchants spawned
	}{
		{
			name:          "spawn 2 merchants",
			terrain:       testTerrain,
			worldSeed:     12345,
			merchantCount: 2,
			expectedMin:   0, // Some may fail due to spawn point validation
			expectedMax:   2,
		},
		{
			name:          "spawn 0 merchants",
			terrain:       testTerrain,
			worldSeed:     12345,
			merchantCount: 0,
			expectedMin:   0,
			expectedMax:   0,
		},
		{
			name:          "spawn 5 merchants",
			terrain:       testTerrain,
			worldSeed:     54321,
			merchantCount: 5,
			expectedMin:   0,
			expectedMax:   5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			world := NewWorld()
			params := procgen.GenerationParams{
				Difficulty: 0.5,
				Depth:      1,
				GenreID:    "fantasy",
			}

			count, err := SpawnMerchantsInTerrain(world, tt.terrain, tt.worldSeed, params, tt.merchantCount)
			if err != nil {
				t.Fatalf("SpawnMerchantsInTerrain failed: %v", err)
			}

			if count < tt.expectedMin || count > tt.expectedMax {
				t.Errorf("spawned %d merchants, want between %d and %d", count, tt.expectedMin, tt.expectedMax)
			}

			// Verify all spawned entities have merchant component
			merchantCount := 0
			for _, entity := range world.GetEntities() {
				if entity.HasComponent("merchant") {
					merchantCount++
				}
			}

			if merchantCount != count {
				t.Errorf("found %d merchants in world, but function returned %d", merchantCount, count)
			}
		})
	}
}

func TestGetNearbyMerchants(t *testing.T) {
	world := NewWorld()

	// Create test merchants at different positions
	merchant1 := world.CreateEntity()
	merchant1.AddComponent(&PositionComponent{X: 100, Y: 100})
	merchant1.AddComponent(NewMerchantComponent(10, MerchantFixed, 1.5))

	merchant2 := world.CreateEntity()
	merchant2.AddComponent(&PositionComponent{X: 150, Y: 100})
	merchant2.AddComponent(NewMerchantComponent(10, MerchantFixed, 1.5))

	merchant3 := world.CreateEntity()
	merchant3.AddComponent(&PositionComponent{X: 300, Y: 100})
	merchant3.AddComponent(NewMerchantComponent(10, MerchantFixed, 1.5))

	// Create non-merchant entity
	notMerchant := world.CreateEntity()
	notMerchant.AddComponent(&PositionComponent{X: 110, Y: 100})

	// Process initial additions
	world.Update(0)

	tests := []struct {
		name          string
		x, y          float64
		radius        float64
		expectedCount int
	}{
		{
			name:          "nearby merchants within 60 pixels",
			x:             100,
			y:             100,
			radius:        60,
			expectedCount: 2, // merchant1 (distance 0) and merchant2 (distance 50)
		},
		{
			name:          "all merchants within 300 pixels",
			x:             100,
			y:             100,
			radius:        300,
			expectedCount: 3, // all merchants
		},
		{
			name:          "no merchants within 30 pixels",
			x:             100,
			y:             100,
			radius:        30,
			expectedCount: 1, // only merchant1 at exact position
		},
		{
			name:          "far from all merchants",
			x:             1000,
			y:             1000,
			radius:        50,
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			merchants := GetNearbyMerchants(world, tt.x, tt.y, tt.radius)

			if len(merchants) != tt.expectedCount {
				t.Errorf("found %d nearby merchants, want %d", len(merchants), tt.expectedCount)
			}

			// Verify all returned entities are actually merchants
			for _, m := range merchants {
				if !m.HasComponent("merchant") {
					t.Error("GetNearbyMerchants returned non-merchant entity")
				}
			}
		})
	}
}

func TestFindClosestMerchant(t *testing.T) {
	world := NewWorld()

	// Create test merchants at different positions
	merchant1 := world.CreateEntity()
	merchant1.AddComponent(&PositionComponent{X: 100, Y: 100})
	merchant1.AddComponent(NewMerchantComponent(10, MerchantFixed, 1.5))

	merchant2 := world.CreateEntity()
	merchant2.AddComponent(&PositionComponent{X: 150, Y: 100})
	merchant2.AddComponent(NewMerchantComponent(10, MerchantFixed, 1.5))

	merchant3 := world.CreateEntity()
	merchant3.AddComponent(&PositionComponent{X: 300, Y: 100})
	merchant3.AddComponent(NewMerchantComponent(10, MerchantFixed, 1.5))

	world.Update(0)

	tests := []struct {
		name           string
		x, y           float64
		radius         float64
		expectMerchant bool
		expectedID     uint64 // ID of expected closest merchant (0 if none)
	}{
		{
			name:           "closest to merchant1",
			x:              100,
			y:              100,
			radius:         300,
			expectMerchant: true,
			expectedID:     merchant1.ID,
		},
		{
			name:           "closest to merchant2",
			x:              150,
			y:              100,
			radius:         300,
			expectMerchant: true,
			expectedID:     merchant2.ID,
		},
		{
			name:           "no merchants in range",
			x:              1000,
			y:              1000,
			radius:         50,
			expectMerchant: false,
			expectedID:     0,
		},
		{
			name:           "only merchant1 in range",
			x:              100,
			y:              100,
			radius:         40,
			expectMerchant: true,
			expectedID:     merchant1.ID,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			merchant, dist := FindClosestMerchant(world, tt.x, tt.y, tt.radius)

			if tt.expectMerchant {
				if merchant == nil {
					t.Fatal("expected to find closest merchant, got nil")
				}
				if merchant.ID != tt.expectedID {
					t.Errorf("found merchant %d, want %d", merchant.ID, tt.expectedID)
				}
				if dist < 0 {
					t.Errorf("distance = %.1f, want >= 0", dist)
				}
			} else {
				if merchant != nil {
					t.Errorf("expected no merchant, found merchant %d", merchant.ID)
				}
				if dist != -1 {
					t.Errorf("distance = %.1f, want -1 for no merchant", dist)
				}
			}
		})
	}
}

func TestGetMerchantInteractionPrompt(t *testing.T) {
	tests := []struct {
		name           string
		setupMerchant  func(*World) *Entity
		expectedPrompt string
	}{
		{
			name: "merchant with name",
			setupMerchant: func(w *World) *Entity {
				m := w.CreateEntity()
				merchComp := NewMerchantComponent(10, MerchantFixed, 1.5)
				merchComp.MerchantName = "Aldric the Trader"
				m.AddComponent(merchComp)
				return m
			},
			expectedPrompt: "Press S to talk to Aldric the Trader",
		},
		{
			name: "merchant with default name",
			setupMerchant: func(w *World) *Entity {
				m := w.CreateEntity()
				m.AddComponent(NewMerchantComponent(10, MerchantFixed, 1.5))
				return m
			},
			expectedPrompt: "Press S to talk to Merchant",
		},
		{
			name: "nil merchant",
			setupMerchant: func(w *World) *Entity {
				return nil
			},
			expectedPrompt: "",
		},
		{
			name: "entity without merchant component",
			setupMerchant: func(w *World) *Entity {
				e := w.CreateEntity()
				e.AddComponent(&PositionComponent{X: 0, Y: 0})
				return e
			},
			expectedPrompt: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			world := NewWorld()
			merchant := tt.setupMerchant(world)

			prompt := GetMerchantInteractionPrompt(merchant)

			if prompt != tt.expectedPrompt {
				t.Errorf("prompt = %q, want %q", prompt, tt.expectedPrompt)
			}
		})
	}
}
