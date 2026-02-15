package engine

import (
	"testing"
)

func TestGleamSweepComponentType(t *testing.T) {
	c := NewGleamSweepComponent()
	if c.Type() != "gleam_sweep" {
		t.Errorf("expected type 'gleam_sweep', got %q", c.Type())
	}
}

func TestGleamSweepComponentDefaults(t *testing.T) {
	c := NewGleamSweepComponent()
	if c.SweepPosition != 0.0 {
		t.Errorf("expected SweepPosition 0, got %f", c.SweepPosition)
	}
	if c.Enabled {
		t.Error("expected Enabled=false by default")
	}
	if c.Active {
		t.Error("expected Active=false by default")
	}
	if c.SweepWidth <= 0 {
		t.Errorf("expected positive SweepWidth, got %f", c.SweepWidth)
	}
}

func TestEquipmentGleamSweepSystemCreation(t *testing.T) {
	world := NewWorld()
	sys := NewEquipmentGleamSweepSystem(world, 42)
	if sys == nil {
		t.Fatal("expected non-nil system")
	}
	if sys.genreID != "fantasy" {
		t.Errorf("expected default genre 'fantasy', got %q", sys.genreID)
	}
}

func TestEquipmentGleamSweepSystemSetGenre(t *testing.T) {
	tests := []struct {
		name    string
		genreID string
	}{
		{"fantasy", "fantasy"},
		{"horror", "horror"},
		{"cyberpunk", "cyberpunk"},
		{"scifi", "scifi"},
		{"postapoc", "postapoc"},
		{"unknown", "unknown_genre"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			world := NewWorld()
			sys := NewEquipmentGleamSweepSystem(world, 42)
			sys.SetGenre(tt.genreID)
			if sys.genreID != tt.genreID {
				t.Errorf("expected genre %q, got %q", tt.genreID, sys.genreID)
			}
		})
	}
}

func TestEquipmentGleamSweepSystemSkipsWithoutSheen(t *testing.T) {
	world := NewWorld()
	sys := NewEquipmentGleamSweepSystem(world, 42)

	entity := NewEntity(1)
	entities := []*Entity{entity}

	// Full scan (first update always scans)
	sys.Update(entities, 1.1)

	// Should not attach gleam component without material_sheen
	if entity.HasComponent("gleam_sweep") {
		t.Error("should not attach gleam_sweep without material_sheen")
	}
}

func TestEquipmentGleamSweepSystemAttachesWithSheen(t *testing.T) {
	world := NewWorld()
	sys := NewEquipmentGleamSweepSystem(world, 42)

	entity := NewEntity(1)
	sheen := NewMaterialSheenComponent()
	sheen.Enabled = true
	sheen.SheenIntensity = 0.8
	sheen.Reflectivity = 0.7
	sheen.DominantMaterial = "Metal"
	entity.AddComponent(sheen)

	entities := []*Entity{entity}

	// Force full scan
	sys.Update(entities, 1.1)

	comp, ok := entity.GetComponent("gleam_sweep")
	if !ok {
		t.Fatal("expected gleam_sweep component to be attached")
	}
	gleam := comp.(*GleamSweepComponent)
	if !gleam.Enabled {
		t.Error("expected gleam to be enabled for high-sheen metal")
	}
	if gleam.Intensity <= 0 {
		t.Errorf("expected positive intensity, got %f", gleam.Intensity)
	}
	if gleam.MaterialHint != "Metal" {
		t.Errorf("expected MaterialHint 'Metal', got %q", gleam.MaterialHint)
	}
}

func TestEquipmentGleamSweepSystemDisabledForLowSheen(t *testing.T) {
	world := NewWorld()
	sys := NewEquipmentGleamSweepSystem(world, 42)

	entity := NewEntity(1)
	sheen := NewMaterialSheenComponent()
	sheen.Enabled = true
	sheen.SheenIntensity = 0.01
	sheen.Reflectivity = 0.01
	sheen.DominantMaterial = "Cloth"
	entity.AddComponent(sheen)

	entities := []*Entity{entity}
	sys.Update(entities, 1.1)

	comp, ok := entity.GetComponent("gleam_sweep")
	if !ok {
		t.Fatal("expected gleam_sweep component to be attached")
	}
	gleam := comp.(*GleamSweepComponent)
	if gleam.Enabled {
		t.Error("expected gleam disabled for very low sheen")
	}
}

func TestEquipmentGleamSweepSystemSweepAnimation(t *testing.T) {
	world := NewWorld()
	sys := NewEquipmentGleamSweepSystem(world, 42)

	entity := NewEntity(1)
	sheen := NewMaterialSheenComponent()
	sheen.Enabled = true
	sheen.SheenIntensity = 0.9
	sheen.Reflectivity = 0.9
	sheen.DominantMaterial = "Metal"
	entity.AddComponent(sheen)

	entities := []*Entity{entity}

	// Initial scan attaches component
	sys.Update(entities, 1.1)

	comp, _ := entity.GetComponent("gleam_sweep")
	gleam := comp.(*GleamSweepComponent)

	// Force sweep to start by zeroing cooldown
	gleam.Active = true
	gleam.SweepPosition = 0.0

	initialPos := gleam.SweepPosition

	// Advance a few frames
	for i := 0; i < 10; i++ {
		sys.Update(entities, 0.016)
	}

	if gleam.SweepPosition <= initialPos {
		t.Error("expected sweep position to advance")
	}
}

func TestEquipmentGleamSweepSystemCooldownCycle(t *testing.T) {
	world := NewWorld()
	sys := NewEquipmentGleamSweepSystem(world, 42)

	entity := NewEntity(1)
	sheen := NewMaterialSheenComponent()
	sheen.Enabled = true
	sheen.SheenIntensity = 0.9
	sheen.Reflectivity = 0.9
	sheen.DominantMaterial = "Crystal"
	entity.AddComponent(sheen)

	entities := []*Entity{entity}
	sys.Update(entities, 1.1)

	comp, _ := entity.GetComponent("gleam_sweep")
	gleam := comp.(*GleamSweepComponent)

	// Force sweep completion
	gleam.Active = true
	gleam.SweepPosition = 2.0 // Beyond 1.0 + width
	sys.Update(entities, 0.016)

	if gleam.Active {
		t.Error("expected sweep to end after passing edge")
	}
	if gleam.CooldownRemaining <= 0 {
		t.Error("expected positive cooldown after sweep completion")
	}

	// Tick down the cooldown
	gleam.CooldownRemaining = 0.01
	sys.Update(entities, 0.02)

	if !gleam.Active {
		t.Error("expected sweep to restart after cooldown expires")
	}
}

func TestEquipmentGleamSweepSystemGenreIntensity(t *testing.T) {
	tests := []struct {
		genre         string
		expectHigher  bool // Relative to horror (dim)
	}{
		{"cyberpunk", true},
		{"horror", false},
	}

	for _, tt := range tests {
		t.Run(tt.genre, func(t *testing.T) {
			world := NewWorld()
			sys := NewEquipmentGleamSweepSystem(world, 42)
			sys.SetGenre(tt.genre)

			entity := NewEntity(1)
			sheen := NewMaterialSheenComponent()
			sheen.Enabled = true
			sheen.SheenIntensity = 0.7
			sheen.Reflectivity = 0.7
			sheen.DominantMaterial = "Metal"
			entity.AddComponent(sheen)

			gleam := NewGleamSweepComponent()
			entity.AddComponent(gleam)

			sys.configureGleam(entity, gleam)

			if tt.expectHigher && gleam.Intensity <= 0.3 {
				t.Errorf("expected higher intensity for %s, got %f", tt.genre, gleam.Intensity)
			}
			if !tt.expectHigher && gleam.Intensity > 0.5 {
				t.Errorf("expected lower intensity for %s, got %f", tt.genre, gleam.Intensity)
			}
		})
	}
}

func TestEquipmentGleamSweepSystemMaterialProfiles(t *testing.T) {
	tests := []struct {
		material string
		faster   bool // Relative to wood baseline
	}{
		{"Metal", true},
		{"Energy", true},
		{"Crystal", false},
		{"Cloth", false},
	}

	for _, tt := range tests {
		t.Run(tt.material, func(t *testing.T) {
			world := NewWorld()
			sys := NewEquipmentGleamSweepSystem(world, 42)

			entity := NewEntity(1)
			sheen := NewMaterialSheenComponent()
			sheen.Enabled = true
			sheen.SheenIntensity = 0.8
			sheen.Reflectivity = 0.8
			sheen.DominantMaterial = tt.material
			entity.AddComponent(sheen)

			gleam := NewGleamSweepComponent()
			entity.AddComponent(gleam)

			sys.configureGleam(entity, gleam)

			woodSpeed := sys.baseSweepSpeed * 0.9 // Wood mult
			if tt.faster && gleam.SweepSpeed <= woodSpeed {
				t.Errorf("expected %s to be faster than wood, got %f vs %f", tt.material, gleam.SweepSpeed, woodSpeed)
			}
			if !tt.faster && gleam.SweepSpeed > woodSpeed {
				t.Errorf("expected %s to be slower than wood, got %f vs %f", tt.material, gleam.SweepSpeed, woodSpeed)
			}
		})
	}
}

func TestClampGleamFloat(t *testing.T) {
	tests := []struct {
		name     string
		v        float64
		min, max float64
		want     float64
	}{
		{"within range", 0.5, 0.0, 1.0, 0.5},
		{"below min", -0.5, 0.0, 1.0, 0.0},
		{"above max", 1.5, 0.0, 1.0, 1.0},
		{"at min", 0.0, 0.0, 1.0, 0.0},
		{"at max", 1.0, 0.0, 1.0, 1.0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := clampGleamFloat(tt.v, tt.min, tt.max)
			if got != tt.want {
				t.Errorf("clampGleamFloat(%f, %f, %f) = %f, want %f", tt.v, tt.min, tt.max, got, tt.want)
			}
		})
	}
}
