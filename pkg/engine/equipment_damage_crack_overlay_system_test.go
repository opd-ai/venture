package engine

import (
	"testing"
)

func TestEquipmentCrackOverlayComponentType(t *testing.T) {
	comp := NewEquipmentCrackOverlayComponent()
	if comp.Type() != "equipment_crack_overlay" {
		t.Errorf("expected type 'equipment_crack_overlay', got %q", comp.Type())
	}
}

func TestEquipmentCrackOverlayComponentDefaults(t *testing.T) {
	comp := NewEquipmentCrackOverlayComponent()
	if comp.Enabled {
		t.Error("expected Enabled=false by default")
	}
	if comp.Intensity != 0.0 {
		t.Errorf("expected Intensity=0.0, got %f", comp.Intensity)
	}
	if len(comp.Segments) != 0 {
		t.Errorf("expected no segments, got %d", len(comp.Segments))
	}
}

func TestEquipmentDamageCrackOverlaySystemCreation(t *testing.T) {
	world := NewWorld()
	sys := NewEquipmentDamageCrackOverlaySystem(world, 42)
	if sys == nil {
		t.Fatal("expected non-nil system")
	}
	if sys.genreID != "fantasy" {
		t.Errorf("expected default genre 'fantasy', got %q", sys.genreID)
	}
}

func TestEquipmentDamageCrackOverlaySystemSetGenre(t *testing.T) {
	genres := []string{"fantasy", "horror", "cyberpunk", "sci-fi", "scifi", "post-apocalyptic", "postapoc"}
	world := NewWorld()
	sys := NewEquipmentDamageCrackOverlaySystem(world, 42)
	for _, g := range genres {
		sys.SetGenre(g)
		if sys.genreID != g {
			t.Errorf("expected genre %q, got %q", g, sys.genreID)
		}
	}
}

func TestEquipmentDamageCrackOverlayNoCracksWhenPristine(t *testing.T) {
	world := NewWorld()
	sys := NewEquipmentDamageCrackOverlaySystem(world, 42)

	entity := NewEntity(1)
	tint := NewEquipmentWearTintComponent()
	tint.Enabled = true
	tint.CrackDensity = 0.0
	tint.EdgeRoughness = 0.0
	entity.AddComponent(tint)

	overlay := NewEquipmentCrackOverlayComponent()
	entity.AddComponent(overlay)

	sys.Update([]*Entity{entity}, 1.0)

	if overlay.Enabled {
		t.Error("expected overlay disabled when crack density is 0")
	}
	if len(overlay.Segments) != 0 {
		t.Errorf("expected no segments, got %d", len(overlay.Segments))
	}
}

func TestEquipmentDamageCrackOverlayGeneratesCracks(t *testing.T) {
	world := NewWorld()
	sys := NewEquipmentDamageCrackOverlaySystem(world, 42)

	entity := NewEntity(1)
	tint := NewEquipmentWearTintComponent()
	tint.Enabled = true
	tint.CrackDensity = 0.5
	tint.EdgeRoughness = 0.3
	entity.AddComponent(tint)

	overlay := NewEquipmentCrackOverlayComponent()
	entity.AddComponent(overlay)

	sys.Update([]*Entity{entity}, 1.0)

	if !overlay.Enabled {
		t.Error("expected overlay enabled with crack density 0.5")
	}
	if len(overlay.Segments) == 0 {
		t.Error("expected segments to be generated")
	}
	if overlay.Intensity < 0.4 || overlay.Intensity > 0.6 {
		t.Errorf("expected intensity near 0.5, got %f", overlay.Intensity)
	}
	if overlay.TreeCount < 1 {
		t.Error("expected at least 1 crack tree")
	}
}

func TestEquipmentDamageCrackOverlayMaxDensity(t *testing.T) {
	world := NewWorld()
	sys := NewEquipmentDamageCrackOverlaySystem(world, 42)

	entity := NewEntity(1)
	tint := NewEquipmentWearTintComponent()
	tint.Enabled = true
	tint.CrackDensity = 1.0
	tint.EdgeRoughness = 1.0
	entity.AddComponent(tint)

	overlay := NewEquipmentCrackOverlayComponent()
	entity.AddComponent(overlay)

	sys.Update([]*Entity{entity}, 1.0)

	if !overlay.Enabled {
		t.Error("expected overlay enabled at max density")
	}
	if overlay.TreeCount != 5 {
		t.Errorf("expected 5 crack trees at max density, got %d", overlay.TreeCount)
	}
	// At max density we expect many segments
	if len(overlay.Segments) < 10 {
		t.Errorf("expected at least 10 segments at max density, got %d", len(overlay.Segments))
	}
}

func TestEquipmentDamageCrackOverlayCaching(t *testing.T) {
	world := NewWorld()
	sys := NewEquipmentDamageCrackOverlaySystem(world, 42)

	entity := NewEntity(1)
	tint := NewEquipmentWearTintComponent()
	tint.Enabled = true
	tint.CrackDensity = 0.5
	tint.EdgeRoughness = 0.3
	entity.AddComponent(tint)

	overlay := NewEquipmentCrackOverlayComponent()
	entity.AddComponent(overlay)

	// First update generates cracks
	sys.Update([]*Entity{entity}, 1.0)
	firstCount := len(overlay.Segments)
	firstKey := overlay.CacheKey

	// Second update with same density should use cache
	sys.Update([]*Entity{entity}, 1.0)
	if overlay.CacheKey != firstKey {
		t.Error("expected cache key to remain same for unchanged density")
	}
	if len(overlay.Segments) != firstCount {
		t.Error("expected segment count to remain same for cached pattern")
	}
}

func TestEquipmentDamageCrackOverlayGenreColors(t *testing.T) {
	tests := []struct {
		genre string
		wantR float64
		wantG float64
		wantB float64
	}{
		{"fantasy", 0.2, 0.2, 0.2},
		{"horror", 0.35, 0.15, 0.1},
		{"cyberpunk", 0.15, 0.25, 0.3},
		{"sci-fi", 0.3, 0.35, 0.45},
		{"postapoc", 0.4, 0.25, 0.1},
	}

	for _, tt := range tests {
		t.Run(tt.genre, func(t *testing.T) {
			world := NewWorld()
			sys := NewEquipmentDamageCrackOverlaySystem(world, 99)
			sys.SetGenre(tt.genre)

			entity := NewEntity(1)
			tint := NewEquipmentWearTintComponent()
			tint.Enabled = true
			tint.CrackDensity = 0.7
			tint.EdgeRoughness = 0.5
			entity.AddComponent(tint)

			overlay := NewEquipmentCrackOverlayComponent()
			entity.AddComponent(overlay)

			sys.Update([]*Entity{entity}, 1.0)

			if overlay.ColorR != tt.wantR || overlay.ColorG != tt.wantG || overlay.ColorB != tt.wantB {
				t.Errorf("genre %s: want color (%.2f,%.2f,%.2f), got (%.2f,%.2f,%.2f)",
					tt.genre, tt.wantR, tt.wantG, tt.wantB,
					overlay.ColorR, overlay.ColorG, overlay.ColorB)
			}
		})
	}
}

func TestEquipmentDamageCrackOverlayNoTintComponent(t *testing.T) {
	world := NewWorld()
	sys := NewEquipmentDamageCrackOverlaySystem(world, 42)

	entity := NewEntity(1)
	// No tint component, no equipment - should skip
	sys.Update([]*Entity{entity}, 1.0)

	_, has := entity.GetComponent("equipment_crack_overlay")
	if has {
		t.Error("should not attach overlay without wear tint component")
	}
}

func TestEquipmentDamageCrackOverlayDisabledTint(t *testing.T) {
	world := NewWorld()
	sys := NewEquipmentDamageCrackOverlaySystem(world, 42)

	entity := NewEntity(1)
	tint := NewEquipmentWearTintComponent()
	tint.Enabled = false
	tint.CrackDensity = 0.5
	entity.AddComponent(tint)

	overlay := NewEquipmentCrackOverlayComponent()
	entity.AddComponent(overlay)

	sys.Update([]*Entity{entity}, 1.0)

	if overlay.Enabled {
		t.Error("expected overlay disabled when tint is disabled")
	}
}

func TestEquipmentDamageCrackOverlaySegmentBounds(t *testing.T) {
	world := NewWorld()
	sys := NewEquipmentDamageCrackOverlaySystem(world, 42)

	entity := NewEntity(1)
	tint := NewEquipmentWearTintComponent()
	tint.Enabled = true
	tint.CrackDensity = 0.8
	tint.EdgeRoughness = 0.6
	entity.AddComponent(tint)

	overlay := NewEquipmentCrackOverlayComponent()
	entity.AddComponent(overlay)

	sys.Update([]*Entity{entity}, 1.0)

	for i, seg := range overlay.Segments {
		if seg.X1 < 0 || seg.X1 > 1 || seg.Y1 < 0 || seg.Y1 > 1 {
			t.Errorf("segment %d start out of bounds: (%.2f, %.2f)", i, seg.X1, seg.Y1)
		}
		if seg.X2 < 0 || seg.X2 > 1 || seg.Y2 < 0 || seg.Y2 > 1 {
			t.Errorf("segment %d end out of bounds: (%.2f, %.2f)", i, seg.X2, seg.Y2)
		}
		if seg.Width <= 0 {
			t.Errorf("segment %d has non-positive width: %.2f", i, seg.Width)
		}
		if seg.Depth < 0 || seg.Depth > 1 {
			t.Errorf("segment %d depth out of range: %.2f", i, seg.Depth)
		}
	}
}

func TestEquipmentDamageCrackOverlayThrottling(t *testing.T) {
	world := NewWorld()
	sys := NewEquipmentDamageCrackOverlaySystem(world, 42)

	entity := NewEntity(1)
	tint := NewEquipmentWearTintComponent()
	tint.Enabled = true
	tint.CrackDensity = 0.5
	tint.EdgeRoughness = 0.3
	entity.AddComponent(tint)

	// Small delta - should not trigger full scan for new entities
	sys.Update([]*Entity{entity}, 0.1)
	_, has := entity.GetComponent("equipment_crack_overlay")
	if has {
		t.Error("should not attach overlay on small delta (throttled)")
	}

	// Accumulate to 1.0
	sys.Update([]*Entity{entity}, 0.9)
	_, has = entity.GetComponent("equipment_crack_overlay")
	if !has {
		t.Error("should attach overlay after 1.0s accumulation")
	}
}

func BenchmarkEquipmentDamageCrackOverlay(b *testing.B) {
	world := NewWorld()
	sys := NewEquipmentDamageCrackOverlaySystem(world, 42)

	entities := make([]*Entity, 100)
	for i := range entities {
		e := NewEntity(1)
		tint := NewEquipmentWearTintComponent()
		tint.Enabled = true
		tint.CrackDensity = 0.5
		tint.EdgeRoughness = 0.3
		e.AddComponent(tint)
		overlay := NewEquipmentCrackOverlayComponent()
		e.AddComponent(overlay)
		entities[i] = e
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Reset cache keys to force regeneration
		for _, e := range entities {
			comp, _ := e.GetComponent("equipment_crack_overlay")
			comp.(*EquipmentCrackOverlayComponent).CacheKey = 0
		}
		sys.timeSinceCheck = 1.0 // Force full scan
		sys.Update(entities, 1.0)
	}
}
