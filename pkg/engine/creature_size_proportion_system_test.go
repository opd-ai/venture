package engine

import (
	"testing"
)

func TestCreatureSizeProportionComponent_Type(t *testing.T) {
	comp := NewCreatureSizeProportionComponent()
	if comp.Type() != "creature_size_proportion" {
		t.Errorf("expected type creature_size_proportion, got %s", comp.Type())
	}
}

func TestCreatureSizeProportionComponent_Defaults(t *testing.T) {
	comp := NewCreatureSizeProportionComponent()
	if comp.HeadRatio != 0.19 {
		t.Errorf("expected HeadRatio 0.19, got %f", comp.HeadRatio)
	}
	if comp.TorsoRatio != 0.41 {
		t.Errorf("expected TorsoRatio 0.41, got %f", comp.TorsoRatio)
	}
	if comp.LegRatio != 0.40 {
		t.Errorf("expected LegRatio 0.40, got %f", comp.LegRatio)
	}
	if comp.WidthScale != 1.0 {
		t.Errorf("expected WidthScale 1.0, got %f", comp.WidthScale)
	}
	if comp.SizeTier != "medium" {
		t.Errorf("expected SizeTier medium, got %s", comp.SizeTier)
	}
}

func TestCreatureSizeProportionSystem_SetGenre(t *testing.T) {
	world := NewWorld()
	sys := NewCreatureSizeProportionSystem(world, 42)
	sys.SetGenre("horror")
	if sys.genreID != "horror" {
		t.Errorf("expected genre horror, got %s", sys.genreID)
	}
}

func TestCreatureSizeProportionSystem_InferSizeTier(t *testing.T) {
	tests := []struct {
		name     string
		width    float64
		height   float64
		wantTier string
	}{
		{"tiny_16", 16, 16, "tiny"},
		{"small_24", 24, 24, "small"},
		{"medium_32", 32, 32, "medium"},
		{"large_48", 48, 48, "large"},
		{"huge_64", 64, 64, "huge"},
		{"borderline_18", 18, 18, "tiny"},
		{"borderline_26", 26, 26, "small"},
		{"borderline_40", 40, 40, "medium"},
		{"borderline_56", 56, 56, "large"},
		{"asymmetric", 20, 48, "large"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			world := NewWorld()
			sys := NewCreatureSizeProportionSystem(world, 42)

			entity := world.CreateEntity()
			entity.AddComponent(&EbitenSprite{Width: tt.width, Height: tt.height, Visible: true})
			entity.AddComponent(NewAIComponent(0, 0))

			tier := sys.inferSizeTier(entity)
			if tier != tt.wantTier {
				t.Errorf("inferSizeTier(%v×%v) = %s, want %s", tt.width, tt.height, tier, tt.wantTier)
			}
		})
	}
}

func TestCreatureSizeProportionSystem_Update_SizeTiers(t *testing.T) {
	tests := []struct {
		name       string
		spriteSize float64
		wantTier   string
		wantHead   float64
		wantWidth  float64
		headTol    float64
		widthTol   float64
	}{
		{"tiny_big_head", 16, "tiny", 0.30, 0.85, 0.02, 0.02},
		{"small_moderate", 24, "small", 0.25, 0.90, 0.02, 0.02},
		{"medium_standard", 32, "medium", 0.19, 1.00, 0.02, 0.02},
		{"large_stocky", 48, "large", 0.15, 1.15, 0.02, 0.02},
		{"huge_massive", 64, "huge", 0.12, 1.30, 0.02, 0.02},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			world := NewWorld()
			sys := NewCreatureSizeProportionSystem(world, 42)
			sys.SetGenre("fantasy") // neutral genre

			creature := world.CreateEntity()
			creature.AddComponent(&PositionComponent{X: 50, Y: 50})
			creature.AddComponent(&EbitenSprite{Width: tt.spriteSize, Height: tt.spriteSize, Visible: true})
			creature.AddComponent(NewAIComponent(50, 50))

			entities := world.GetEntities()
			sys.Update(entities, 2.0)

			comp, ok := creature.GetComponent("creature_size_proportion")
			if !ok {
				t.Fatal("creature_size_proportion component not created")
			}
			prop := comp.(*CreatureSizeProportionComponent)

			if prop.SizeTier != tt.wantTier {
				t.Errorf("SizeTier: got %s, want %s", prop.SizeTier, tt.wantTier)
			}
			if !nearEqual(prop.HeadRatio, tt.wantHead, tt.headTol) {
				t.Errorf("HeadRatio: got %f, want ~%f", prop.HeadRatio, tt.wantHead)
			}
			if !nearEqual(prop.WidthScale, tt.wantWidth, tt.widthTol) {
				t.Errorf("WidthScale: got %f, want ~%f", prop.WidthScale, tt.wantWidth)
			}

			// Verify proportions sum to 1.0
			sum := prop.HeadRatio + prop.TorsoRatio + prop.LegRatio
			if !nearEqual(sum, 1.0, 0.01) {
				t.Errorf("proportions sum = %f, want ~1.0", sum)
			}
		})
	}
}

func TestCreatureSizeProportionSystem_GenreModifiers(t *testing.T) {
	tests := []struct {
		name         string
		genre        string
		expectNarrow bool // horror creatures are narrower
		expectWide   bool // cyberpunk creatures are wider
	}{
		{"fantasy_neutral", "fantasy", false, false},
		{"horror_narrow", "horror", true, false},
		{"scifi_wider", "scifi", false, true},
		{"cyberpunk_wider", "cyberpunk", false, true},
		{"unknown_neutral", "steampunk", false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			world := NewWorld()
			sys := NewCreatureSizeProportionSystem(world, 42)
			sys.SetGenre(tt.genre)

			creature := world.CreateEntity()
			creature.AddComponent(&EbitenSprite{Width: 32, Height: 32, Visible: true})
			creature.AddComponent(NewAIComponent(50, 50))

			entities := world.GetEntities()
			sys.Update(entities, 2.0)

			comp, ok := creature.GetComponent("creature_size_proportion")
			if !ok {
				t.Fatal("creature_size_proportion component not created")
			}
			prop := comp.(*CreatureSizeProportionComponent)

			if tt.expectNarrow && prop.WidthScale >= 1.0 {
				t.Errorf("expected narrow width (<1.0) for %s, got %f", tt.genre, prop.WidthScale)
			}
			if tt.expectWide && prop.WidthScale <= 1.0 {
				t.Errorf("expected wide width (>1.0) for %s, got %f", tt.genre, prop.WidthScale)
			}
		})
	}
}

func TestCreatureSizeProportionSystem_SkipsPlayers(t *testing.T) {
	world := NewWorld()
	sys := NewCreatureSizeProportionSystem(world, 42)

	player := world.CreateEntity()
	player.AddComponent(&EbitenSprite{Width: 32, Height: 32, Visible: true})
	player.AddComponent(NewAIComponent(0, 0))
	player.AddComponent(NewStubInput())

	entities := world.GetEntities()
	sys.Update(entities, 2.0)

	_, ok := player.GetComponent("creature_size_proportion")
	if ok {
		t.Error("player should not receive creature_size_proportion component")
	}
}

func TestCreatureSizeProportionSystem_NewEntitiesPickedUp(t *testing.T) {
	world := NewWorld()
	sys := NewCreatureSizeProportionSystem(world, 42)
	sys.SetGenre("fantasy")

	// First update initializes
	creature1 := world.CreateEntity()
	creature1.AddComponent(&EbitenSprite{Width: 32, Height: 32, Visible: true})
	creature1.AddComponent(NewAIComponent(0, 0))

	entities := world.GetEntities()
	sys.Update(entities, 2.0)

	// Add new creature after initialization
	creature2 := world.CreateEntity()
	creature2.AddComponent(&EbitenSprite{Width: 48, Height: 48, Visible: true})
	creature2.AddComponent(NewAIComponent(0, 0))

	entities = world.GetEntities()
	sys.Update(entities, 2.0) // should assign to new entity

	comp, ok := creature2.GetComponent("creature_size_proportion")
	if !ok {
		t.Fatal("new creature should receive creature_size_proportion component")
	}
	prop := comp.(*CreatureSizeProportionComponent)
	if prop.SizeTier != "large" {
		t.Errorf("expected large tier for 48px creature, got %s", prop.SizeTier)
	}
}

func TestCreatureSizeProportionSystem_Throttled(t *testing.T) {
	world := NewWorld()
	sys := NewCreatureSizeProportionSystem(world, 42)
	sys.SetGenre("fantasy")

	creature := world.CreateEntity()
	creature.AddComponent(&EbitenSprite{Width: 32, Height: 32, Visible: true})
	creature.AddComponent(NewAIComponent(0, 0))

	entities := world.GetEntities()

	// First update with enough delta
	sys.Update(entities, 2.0)
	_, ok := creature.GetComponent("creature_size_proportion")
	if !ok {
		t.Fatal("expected component after first update")
	}

	// Quick successive update should be throttled
	sys.timeSinceCheck = 0
	sys.Update(entities, 0.1)
	// Should not error; throttle just skips
}

func TestClampProportion(t *testing.T) {
	tests := []struct {
		name string
		in   float64
		want float64
	}{
		{"below_min", 0.01, 0.05},
		{"above_max", 0.80, 0.60},
		{"in_range", 0.30, 0.30},
		{"at_min", 0.05, 0.05},
		{"at_max", 0.60, 0.60},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := clampProportion(tt.in)
			if got != tt.want {
				t.Errorf("clampProportion(%f) = %f, want %f", tt.in, got, tt.want)
			}
		})
	}
}

func TestClampWidthScale(t *testing.T) {
	tests := []struct {
		name string
		in   float64
		want float64
	}{
		{"below_min", 0.5, 0.7},
		{"above_max", 2.0, 1.5},
		{"in_range", 1.1, 1.1},
		{"at_min", 0.7, 0.7},
		{"at_max", 1.5, 1.5},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := clampWidthScale(tt.in)
			if got != tt.want {
				t.Errorf("clampWidthScale(%f) = %f, want %f", tt.in, got, tt.want)
			}
		})
	}
}

func TestCreatureSizeProportionSystem_NilWorld(t *testing.T) {
	sys := NewCreatureSizeProportionSystem(nil, 42)
	if sys == nil {
		t.Fatal("expected non-nil system with nil world")
	}
	sys.SetGenre("fantasy")
	sys.Update(nil, 2.0) // should not panic
}

func BenchmarkCreatureSizeProportionSystem(b *testing.B) {
	world := NewWorld()
	sys := NewCreatureSizeProportionSystem(world, 42)
	sys.SetGenre("fantasy")

	for i := 0; i < 200; i++ {
		e := world.CreateEntity()
		e.AddComponent(&EbitenSprite{Width: 32, Height: 32, Visible: true})
		e.AddComponent(NewAIComponent(0, 0))
	}

	entities := world.GetEntities()
	sys.Update(entities, 2.0) // prime

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.timeSinceCheck = 2.0
		sys.initialized = false
		sys.Update(entities, 2.0)
	}
}
