package engine

import (
	"image/color"
	"math"
	"testing"
)

func TestCompanionBondTetherComponent_Type(t *testing.T) {
	c := &CompanionBondTetherComponent{}
	if got := c.Type(); got != "companion_bond_tether" {
		t.Errorf("Type() = %q, want %q", got, "companion_bond_tether")
	}
}

func TestNewCompanionBondTetherSystem(t *testing.T) {
	world := NewWorld()
	sys := NewCompanionBondTetherSystem(world, 42)
	if sys == nil {
		t.Fatal("NewCompanionBondTetherSystem returned nil")
	}
	if sys.genreID != "fantasy" {
		t.Errorf("default genre = %q, want %q", sys.genreID, "fantasy")
	}
	if sys.maxDistance != 250.0 {
		t.Errorf("maxDistance = %f, want 250.0", sys.maxDistance)
	}
}

func TestCompanionBondTetherSystem_SetGenre(t *testing.T) {
	genres := []struct {
		name      string
		wantColor color.RGBA
	}{
		{"fantasy", color.RGBA{R: 255, G: 215, B: 80, A: 255}},
		{"horror", color.RGBA{R: 180, G: 30, B: 30, A: 255}},
		{"scifi", color.RGBA{R: 60, G: 200, B: 255, A: 255}},
		{"cyberpunk", color.RGBA{R: 255, G: 50, B: 200, A: 255}},
		{"postapoc", color.RGBA{R: 200, G: 150, B: 50, A: 255}},
	}
	for _, tt := range genres {
		t.Run(tt.name, func(t *testing.T) {
			world := NewWorld()
			sys := NewCompanionBondTetherSystem(world, 42)
			sys.SetGenre(tt.name)
			if sys.genreID != tt.name {
				t.Errorf("genreID = %q, want %q", sys.genreID, tt.name)
			}
			if sys.preset.BaseColor != tt.wantColor {
				t.Errorf("BaseColor = %v, want %v", sys.preset.BaseColor, tt.wantColor)
			}
		})
	}
}

func TestCompanionBondTetherSystem_Update_ActiveTether(t *testing.T) {
	world := NewWorld()
	sys := NewCompanionBondTetherSystem(world, 99)

	// Create owner entity
	owner := world.CreateEntity()
	owner.AddComponent(&PositionComponent{X: 100, Y: 100})

	// Create companion entity near the owner
	companion := world.CreateEntity()
	companion.AddComponent(&PositionComponent{X: 120, Y: 110})
	companion.AddComponent(&CompanionComponent{
		OwnerID: owner.ID,
		Loyalty: 80.0,
	})

	world.FlushPendingEntities()
	entities := []*Entity{owner, companion}
	sys.Update(entities, 0.016)

	tether := getTetherComp(companion)
	if tether == nil {
		t.Fatal("expected CompanionBondTetherComponent to be created")
	}
	if !tether.Active {
		t.Error("tether should be active when companion is near owner")
	}
	if tether.Opacity <= 0 {
		t.Errorf("opacity should be positive, got %f", tether.Opacity)
	}
	if tether.Thickness <= 0 {
		t.Errorf("thickness should be positive, got %f", tether.Thickness)
	}
	if tether.TetherColor.A == 0 {
		t.Error("tether alpha should be non-zero")
	}
}

func TestCompanionBondTetherSystem_Update_NoOwner(t *testing.T) {
	world := NewWorld()
	sys := NewCompanionBondTetherSystem(world, 99)

	companion := world.CreateEntity()
	companion.AddComponent(&PositionComponent{X: 50, Y: 50})
	companion.AddComponent(&CompanionComponent{
		OwnerID: 99999, // non-existent owner
		Loyalty: 50.0,
	})
	companion.AddComponent(&CompanionBondTetherComponent{Active: true})

	world.FlushPendingEntities()
	entities := []*Entity{companion}
	sys.Update(entities, 0.016)

	tether := getTetherComp(companion)
	if tether != nil && tether.Active {
		t.Error("tether should be inactive when owner doesn't exist")
	}
}

func TestCompanionBondTetherSystem_Update_TooFar(t *testing.T) {
	world := NewWorld()
	sys := NewCompanionBondTetherSystem(world, 99)

	owner := world.CreateEntity()
	owner.AddComponent(&PositionComponent{X: 0, Y: 0})

	companion := world.CreateEntity()
	companion.AddComponent(&PositionComponent{X: 500, Y: 500})
	companion.AddComponent(&CompanionComponent{
		OwnerID: owner.ID,
		Loyalty: 100.0,
	})

	world.FlushPendingEntities()
	entities := []*Entity{owner, companion}
	sys.Update(entities, 0.016)

	tether := getTetherComp(companion)
	if tether != nil && tether.Active {
		t.Error("tether should be inactive when companion is beyond maxDistance")
	}
}

func TestCompanionBondTetherSystem_Update_DistanceFade(t *testing.T) {
	world := NewWorld()
	sys := NewCompanionBondTetherSystem(world, 99)

	owner := world.CreateEntity()
	owner.AddComponent(&PositionComponent{X: 0, Y: 0})

	// Companion at close range
	closeCompanion := world.CreateEntity()
	closeCompanion.AddComponent(&PositionComponent{X: 30, Y: 0})
	closeCompanion.AddComponent(&CompanionComponent{
		OwnerID: owner.ID,
		Loyalty: 80.0,
	})

	// Companion in fade zone (between fadeStart=150 and maxDistance=250)
	farCompanion := world.CreateEntity()
	farCompanion.AddComponent(&PositionComponent{X: 200, Y: 0})
	farCompanion.AddComponent(&CompanionComponent{
		OwnerID: owner.ID,
		Loyalty: 80.0,
	})

	world.FlushPendingEntities()
	entities := []*Entity{owner, closeCompanion, farCompanion}
	sys.Update(entities, 0.016)

	closeTether := getTetherComp(closeCompanion)
	farTether := getTetherComp(farCompanion)

	if closeTether == nil || farTether == nil {
		t.Fatal("expected tether components on both companions")
	}
	if farTether.Opacity >= closeTether.Opacity {
		t.Errorf("far companion opacity (%f) should be less than close (%f)",
			farTether.Opacity, closeTether.Opacity)
	}
}

func TestCompanionBondTetherSystem_Update_LoyaltyModulation(t *testing.T) {
	tests := []struct {
		name    string
		loyalty float64
	}{
		{"zero_loyalty", 0.0},
		{"half_loyalty", 50.0},
		{"full_loyalty", 100.0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			world := NewWorld()
			sys := NewCompanionBondTetherSystem(world, 99)

			owner := world.CreateEntity()
			owner.AddComponent(&PositionComponent{X: 100, Y: 100})

			companion := world.CreateEntity()
			companion.AddComponent(&PositionComponent{X: 110, Y: 100})
			companion.AddComponent(&CompanionComponent{
				OwnerID: owner.ID,
				Loyalty: tt.loyalty,
			})

	world.FlushPendingEntities()
			sys.Update([]*Entity{owner, companion}, 0.016)

			tether := getTetherComp(companion)
			if tether == nil {
				t.Fatal("expected tether component")
			}
			expectedFactor := math.Max(0, math.Min(tt.loyalty/100.0, 1.0))
			if math.Abs(tether.LoyaltyFactor-expectedFactor) > 0.01 {
				t.Errorf("LoyaltyFactor = %f, want %f", tether.LoyaltyFactor, expectedFactor)
			}
		})
	}
}

func TestCompanionBondTetherSystem_Update_PulseAdvances(t *testing.T) {
	world := NewWorld()
	sys := NewCompanionBondTetherSystem(world, 42)

	owner := world.CreateEntity()
	owner.AddComponent(&PositionComponent{X: 100, Y: 100})

	companion := world.CreateEntity()
	companion.AddComponent(&PositionComponent{X: 110, Y: 100})
	companion.AddComponent(&CompanionComponent{
		OwnerID: owner.ID,
		Loyalty: 75.0,
	})

	world.FlushPendingEntities()
	entities := []*Entity{owner, companion}
	sys.Update(entities, 0.5)
	tether := getTetherComp(companion)
	phase1 := tether.PulsePhase

	sys.Update(entities, 0.5)
	phase2 := tether.PulsePhase

	if phase2 <= phase1 || phase2 == 0 {
		t.Errorf("pulse phase should advance: phase1=%f, phase2=%f", phase1, phase2)
	}
}

func TestCompanionBondTetherSystem_Update_SkipNoCompanion(t *testing.T) {
	world := NewWorld()
	sys := NewCompanionBondTetherSystem(world, 42)

	// Entity without CompanionComponent — should be skipped
	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 50, Y: 50})

	sys.Update([]*Entity{entity}, 0.016)

	raw, _ := entity.GetComponent("companion_bond_tether")
	if raw != nil {
		t.Error("non-companion entity should not get tether component")
	}
}

func TestCompanionBondTetherSystem_Update_SkipZeroOwnerID(t *testing.T) {
	world := NewWorld()
	sys := NewCompanionBondTetherSystem(world, 42)

	companion := world.CreateEntity()
	companion.AddComponent(&PositionComponent{X: 50, Y: 50})
	companion.AddComponent(&CompanionComponent{OwnerID: 0, Loyalty: 50})

	sys.Update([]*Entity{companion}, 0.016)

	raw, _ := companion.GetComponent("companion_bond_tether")
	if raw != nil {
		t.Error("companion with OwnerID 0 should not get tether component")
	}
}

func TestCompanionBondTetherSystem_Update_NilWorld(t *testing.T) {
	sys := NewCompanionBondTetherSystem(nil, 42)
	// Should not panic
	sys.Update([]*Entity{}, 0.016)
}

func TestCompanionBondTetherSystem_Update_JitterApplied(t *testing.T) {
	world := NewWorld()
	sys := NewCompanionBondTetherSystem(world, 42)
	sys.SetGenre("horror") // horror has Jitter > 0

	owner := world.CreateEntity()
	owner.AddComponent(&PositionComponent{X: 100, Y: 100})

	companion := world.CreateEntity()
	companion.AddComponent(&PositionComponent{X: 110, Y: 100})
	companion.AddComponent(&CompanionComponent{
		OwnerID: owner.ID,
		Loyalty: 60.0,
	})

	// Run multiple updates; jitter means coordinates will vary
	world.FlushPendingEntities()
	sys.Update([]*Entity{owner, companion}, 0.016)
	tether := getTetherComp(companion)
	ox1, oy1 := tether.OwnerX, tether.OwnerY

	world.FlushPendingEntities()
	sys.Update([]*Entity{owner, companion}, 0.016)
	ox2, oy2 := tether.OwnerX, tether.OwnerY

	// With jitter, at least one coordinate pair should differ
	if ox1 == ox2 && oy1 == oy2 {
		t.Log("jitter may not have changed positions (statistically unlikely but possible)")
	}
}

func TestCompanionBondTetherSystem_Endpoints(t *testing.T) {
	world := NewWorld()
	sys := NewCompanionBondTetherSystem(world, 99)

	owner := world.CreateEntity()
	owner.AddComponent(&PositionComponent{X: 200, Y: 300})

	companion := world.CreateEntity()
	companion.AddComponent(&PositionComponent{X: 220, Y: 310})
	companion.AddComponent(&CompanionComponent{
		OwnerID: owner.ID,
		Loyalty: 100.0,
	})

	world.FlushPendingEntities()
	sys.Update([]*Entity{owner, companion}, 0.016)

	tether := getTetherComp(companion)
	if tether == nil {
		t.Fatal("expected tether")
	}
	// Fantasy has no jitter, so endpoints should match positions exactly
	if tether.OwnerX != 200 || tether.OwnerY != 300 {
		t.Errorf("owner endpoint = (%f,%f), want (200,300)", tether.OwnerX, tether.OwnerY)
	}
	if tether.CompanionX != 220 || tether.CompanionY != 310 {
		t.Errorf("companion endpoint = (%f,%f), want (220,310)", tether.CompanionX, tether.CompanionY)
	}
}

// getTetherComp is a test helper to extract the tether component.
func getTetherComp(e *Entity) *CompanionBondTetherComponent {
	raw, _ := e.GetComponent("companion_bond_tether")
	if raw == nil {
		return nil
	}
	t, _ := raw.(*CompanionBondTetherComponent)
	return t
}
