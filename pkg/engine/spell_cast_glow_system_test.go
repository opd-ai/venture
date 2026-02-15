package engine

import (
	"math"
	"testing"

	"github.com/opd-ai/venture/pkg/procgen/magic"
)

func TestSpellCastGlowComponentType(t *testing.T) {
	c := NewSpellCastGlowComponent()
	if c.Type() != "spell_cast_glow" {
		t.Errorf("expected type 'spell_cast_glow', got %q", c.Type())
	}
}

func TestSpellCastGlowComponentDefaults(t *testing.T) {
	c := NewSpellCastGlowComponent()
	if c.Active {
		t.Error("expected Active=false by default")
	}
	if c.Intensity != 0 {
		t.Error("expected Intensity=0 by default")
	}
	if c.PulseSpeed < 0.5 {
		t.Error("expected PulseSpeed >= 0.5")
	}
}

func TestNewSpellCastGlowSystem(t *testing.T) {
	world := NewWorld()
	sys := NewSpellCastGlowSystem(world, 42)
	if sys == nil {
		t.Fatal("expected non-nil system")
	}
	if sys.genreID != "fantasy" {
		t.Errorf("expected default genre 'fantasy', got %q", sys.genreID)
	}
}

func TestSpellCastGlowSystemSetGenre(t *testing.T) {
	world := NewWorld()
	sys := NewSpellCastGlowSystem(world, 42)
	sys.SetGenre("horror")
	if sys.genreID != "horror" {
		t.Errorf("expected genre 'horror', got %q", sys.genreID)
	}
}

func TestSpellCastGlowActivatesOnCasting(t *testing.T) {
	world := NewWorld()
	sys := NewSpellCastGlowSystem(world, 42)

	entity := NewEntity(1)
	slots := &SpellSlotComponent{
		Casting:    0,
		CastingBar: 0.5,
	}
	slots.Slots[0] = &magic.Spell{
		Element: magic.ElementFire,
	}
	entity.AddComponent(slots)

	entities := []*Entity{entity}
	sys.Update(entities, 0.016)

	comp, ok := entity.GetComponent("spell_cast_glow")
	if !ok {
		t.Fatal("expected spell_cast_glow component to be created")
	}
	gc := comp.(*SpellCastGlowComponent)
	if !gc.Active {
		t.Error("expected glow to be active during casting")
	}
	if gc.Intensity <= 0 {
		t.Error("expected positive intensity during casting")
	}
}

func TestSpellCastGlowElementColors(t *testing.T) {
	tests := []struct {
		name    string
		element magic.ElementType
		expectR float64
		expectB float64
	}{
		{"fire", magic.ElementFire, 0.9, 0.2},
		{"ice", magic.ElementIce, 0.3, 0.8},
		{"lightning", magic.ElementLightning, 0.8, 0.4},
		{"dark", magic.ElementDark, 0.5, 0.4},
		{"arcane", magic.ElementArcane, 0.55, 0.8},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			world := NewWorld()
			sys := NewSpellCastGlowSystem(world, 12345)

			entity := NewEntity(1)
			slots := &SpellSlotComponent{
				Casting:    0,
				CastingBar: 0.5,
			}
			slots.Slots[0] = &magic.Spell{Element: tt.element}
			entity.AddComponent(slots)

			sys.Update([]*Entity{entity}, 0.016)

			comp, _ := entity.GetComponent("spell_cast_glow")
			gc := comp.(*SpellCastGlowComponent)

			// Check color is in the right ballpark (with genre shift and rng variation)
			if gc.GlowR < tt.expectR-0.3 || gc.GlowR > tt.expectR+0.3 {
				t.Errorf("GlowR=%.2f outside expected range for %s", gc.GlowR, tt.name)
			}
			if gc.GlowB < tt.expectB-0.3 || gc.GlowB > tt.expectB+0.3 {
				t.Errorf("GlowB=%.2f outside expected range for %s", gc.GlowB, tt.name)
			}
		})
	}
}

func TestSpellCastGlowIntensityRampsWithProgress(t *testing.T) {
	world := NewWorld()
	sys := NewSpellCastGlowSystem(world, 42)

	entity := NewEntity(1)
	slots := &SpellSlotComponent{Casting: 0, CastingBar: 0.1}
	slots.Slots[0] = &magic.Spell{Element: magic.ElementArcane}
	entity.AddComponent(slots)

	sys.Update([]*Entity{entity}, 0.016)
	comp, _ := entity.GetComponent("spell_cast_glow")
	gc := comp.(*SpellCastGlowComponent)
	lowIntensity := gc.Intensity

	// Advance casting bar
	slots.CastingBar = 0.9
	sys.Update([]*Entity{entity}, 0.016)
	highIntensity := gc.Intensity

	if highIntensity <= lowIntensity {
		t.Errorf("expected higher intensity at 0.9 (%.3f) vs 0.1 (%.3f)", highIntensity, lowIntensity)
	}
}

func TestSpellCastGlowFadesAfterCastComplete(t *testing.T) {
	world := NewWorld()
	sys := NewSpellCastGlowSystem(world, 42)

	entity := NewEntity(1)
	slots := &SpellSlotComponent{Casting: 0, CastingBar: 0.8}
	slots.Slots[0] = &magic.Spell{Element: magic.ElementFire}
	entity.AddComponent(slots)

	// Activate glow
	sys.Update([]*Entity{entity}, 0.016)
	comp, _ := entity.GetComponent("spell_cast_glow")
	gc := comp.(*SpellCastGlowComponent)
	if !gc.Active {
		t.Fatal("expected active glow")
	}

	// Stop casting
	slots.Casting = -1

	// Several fade frames
	for i := 0; i < 30; i++ {
		sys.Update([]*Entity{entity}, 0.016)
	}

	if gc.Active {
		t.Error("expected glow to deactivate after fade-out")
	}
	if gc.Intensity > 0.01 {
		t.Errorf("expected near-zero intensity after fade, got %.3f", gc.Intensity)
	}
}

func TestSpellCastGlowNoSpellSlots(t *testing.T) {
	world := NewWorld()
	sys := NewSpellCastGlowSystem(world, 42)

	entity := NewEntity(1)
	entity.AddComponent(&PositionComponent{X: 0, Y: 0})

	// Should not panic or create glow component
	sys.Update([]*Entity{entity}, 0.016)

	_, ok := entity.GetComponent("spell_cast_glow")
	if ok {
		t.Error("should not create glow component without spell_slots")
	}
}

func TestSpellCastGlowNotCasting(t *testing.T) {
	world := NewWorld()
	sys := NewSpellCastGlowSystem(world, 42)

	entity := NewEntity(1)
	slots := &SpellSlotComponent{Casting: -1}
	entity.AddComponent(slots)

	sys.Update([]*Entity{entity}, 0.016)

	comp, ok := entity.GetComponent("spell_cast_glow")
	if !ok {
		return // No component created is also fine
	}
	gc := comp.(*SpellCastGlowComponent)
	if gc.Active {
		t.Error("glow should not be active when not casting")
	}
}

func TestSpellCastGlowGenreShifts(t *testing.T) {
	tests := []struct {
		genre         string
		expectBright  float64
		expectSat     float64
	}{
		{"fantasy", 1.0, 1.0},
		{"horror", 0.7, 0.85},
		{"scifi", 1.1, 1.15},
		{"cyberpunk", 1.15, 1.3},
		{"postapoc", 0.8, 0.9},
	}

	for _, tt := range tests {
		t.Run(tt.genre, func(t *testing.T) {
			world := NewWorld()
			sys := NewSpellCastGlowSystem(world, 42)
			sys.SetGenre(tt.genre)

			shift := sys.getGenreShift()
			if math.Abs(shift.Brightness-tt.expectBright) > 0.001 {
				t.Errorf("brightness: got %.2f, want %.2f", shift.Brightness, tt.expectBright)
			}
			if math.Abs(shift.Saturation-tt.expectSat) > 0.001 {
				t.Errorf("saturation: got %.2f, want %.2f", shift.Saturation, tt.expectSat)
			}
		})
	}
}

func TestSpellCastGlowPulseAnimates(t *testing.T) {
	world := NewWorld()
	sys := NewSpellCastGlowSystem(world, 42)

	entity := NewEntity(1)
	slots := &SpellSlotComponent{Casting: 0, CastingBar: 0.5}
	slots.Slots[0] = &magic.Spell{Element: magic.ElementLightning}
	entity.AddComponent(slots)

	sys.Update([]*Entity{entity}, 0.016)
	comp, _ := entity.GetComponent("spell_cast_glow")
	gc := comp.(*SpellCastGlowComponent)

	phase1 := gc.PulsePhase

	sys.Update([]*Entity{entity}, 0.1)
	phase2 := gc.PulsePhase

	if phase2 <= phase1 {
		t.Error("expected pulse phase to advance over time")
	}
}

func TestSpellCastGlowRadiusGrowsWithProgress(t *testing.T) {
	world := NewWorld()
	sys := NewSpellCastGlowSystem(world, 42)

	entity := NewEntity(1)
	slots := &SpellSlotComponent{Casting: 0, CastingBar: 0.1}
	slots.Slots[0] = &magic.Spell{Element: magic.ElementEarth}
	entity.AddComponent(slots)

	sys.Update([]*Entity{entity}, 0.016)
	comp, _ := entity.GetComponent("spell_cast_glow")
	gc := comp.(*SpellCastGlowComponent)
	smallRadius := gc.GlowRadius

	slots.CastingBar = 0.9
	sys.Update([]*Entity{entity}, 0.016)
	largeRadius := gc.GlowRadius

	if largeRadius <= smallRadius {
		t.Errorf("radius should grow: %.2f at 0.1 vs %.2f at 0.9", smallRadius, largeRadius)
	}
}

func TestSpellCastGlowNilSpellHandled(t *testing.T) {
	world := NewWorld()
	sys := NewSpellCastGlowSystem(world, 42)

	entity := NewEntity(1)
	slots := &SpellSlotComponent{Casting: 0, CastingBar: 0.3}
	// Slot 0 is nil — no spell assigned
	entity.AddComponent(slots)

	// Should not panic; falls back to arcane element
	sys.Update([]*Entity{entity}, 0.016)

	comp, ok := entity.GetComponent("spell_cast_glow")
	if !ok {
		t.Fatal("expected component to be created")
	}
	gc := comp.(*SpellCastGlowComponent)
	if !gc.Active {
		t.Error("expected active glow even with nil spell")
	}
}

func TestSpellCastGlowUnknownGenreFallback(t *testing.T) {
	world := NewWorld()
	sys := NewSpellCastGlowSystem(world, 42)
	sys.SetGenre("unknown_genre")

	shift := sys.getGenreShift()
	// Should fall back to fantasy
	if math.Abs(shift.Brightness-1.0) > 0.001 {
		t.Errorf("expected fantasy fallback brightness 1.0, got %.2f", shift.Brightness)
	}
}

func TestClampSpellGlow(t *testing.T) {
	tests := []struct {
		in, want float64
	}{
		{-0.5, 0.0},
		{0.0, 0.0},
		{0.5, 0.5},
		{1.0, 1.0},
		{1.5, 1.0},
	}
	for _, tt := range tests {
		got := clampSpellGlow(tt.in)
		if got != tt.want {
			t.Errorf("clampSpellGlow(%f) = %f, want %f", tt.in, got, tt.want)
		}
	}
}

func BenchmarkSpellCastGlowUpdate(b *testing.B) {
	world := NewWorld()
	sys := NewSpellCastGlowSystem(world, 42)

	entities := make([]*Entity, 100)
	for i := range entities {
		e := NewEntity(uint64(i + 1))
		slots := &SpellSlotComponent{Casting: 0, CastingBar: 0.5}
		slots.Slots[0] = &magic.Spell{Element: magic.ElementFire}
		e.AddComponent(slots)
		e.AddComponent(NewSpellCastGlowComponent())
		entities[i] = e
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.Update(entities, 0.016)
	}
}
