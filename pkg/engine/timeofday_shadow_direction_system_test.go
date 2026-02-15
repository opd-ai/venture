package engine

import (
	"testing"

	"github.com/opd-ai/venture/pkg/rendering/palette"
)

func TestNewTimeOfDayShadowDirectionSystem(t *testing.T) {
	world := NewWorld()
	sys := NewTimeOfDayShadowDirectionSystem(world, 12345)

	if sys == nil {
		t.Fatal("expected non-nil system")
	}
	if sys.genreID != "fantasy" {
		t.Errorf("default genre = %q, want %q", sys.genreID, "fantasy")
	}
	if sys.updateInterval != 0.5 {
		t.Errorf("updateInterval = %f, want 0.5", sys.updateInterval)
	}
}

func TestTimeOfDayShadowDirectionSystem_SetGenre(t *testing.T) {
	tests := []struct {
		name        string
		genre       string
		wantStretch float64
		wantNightOp float64
	}{
		{"fantasy", "fantasy", 1.0, 0.15},
		{"horror", "horror", 1.4, 0.30},
		{"cyberpunk", "cyberpunk", 1.1, 0.22},
		{"scifi", "scifi", 1.0, 0.12},
		{"postapoc", "postapoc", 1.2, 0.10},
		{"unknown", "unknown", 1.0, 0.15},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			world := NewWorld()
			sys := NewTimeOfDayShadowDirectionSystem(world, 42)
			sys.SetGenre(tt.genre)

			if sys.preset.StretchScale != tt.wantStretch {
				t.Errorf("StretchScale = %f, want %f", sys.preset.StretchScale, tt.wantStretch)
			}
			if sys.preset.NightOpacity != tt.wantNightOp {
				t.Errorf("NightOpacity = %f, want %f", sys.preset.NightOpacity, tt.wantNightOp)
			}
		})
	}
}

func TestTimeOfDayShadowDirectionSystem_SetLightingSystem(t *testing.T) {
	world := NewWorld()
	sys := NewTimeOfDayShadowDirectionSystem(world, 42)
	ls := NewTimeOfDayLightingSystem(world, 99)
	sys.SetLightingSystem(ls)

	if sys.lightingSystem != ls {
		t.Error("expected lighting system to be set")
	}
}

func TestTimeOfDayShadowDirectionSystem_UpdateNoLighting(t *testing.T) {
	world := NewWorld()
	sys := NewTimeOfDayShadowDirectionSystem(world, 42)

	entity := NewEntity(1)
	entity.AddComponent(&PositionComponent{X: 10, Y: 10})
	shadow := NewDropShadowComponent()
	entity.AddComponent(shadow)

	// Should not panic without lighting system
	sys.Update([]*Entity{entity}, 1.0)

	// Shadow should remain unchanged
	if shadow.OffsetX != 0 {
		t.Errorf("OffsetX = %f, want 0 (no lighting system)", shadow.OffsetX)
	}
}

func TestTimeOfDayShadowDirectionSystem_UpdateThrottle(t *testing.T) {
	world := NewWorld()
	sys := NewTimeOfDayShadowDirectionSystem(world, 42)
	ls := NewTimeOfDayLightingSystem(world, 99)
	ls.ForceTimeOfDay(palette.TimeOfDayDawn)
	sys.SetLightingSystem(ls)

	entity := NewEntity(1)
	entity.AddComponent(&PositionComponent{X: 10, Y: 10})
	entity.AddComponent(&ColliderComponent{Width: 32, Height: 32})
	shadow := NewDropShadowComponent()
	entity.AddComponent(shadow)

	// First update with small dt should not trigger (below interval)
	sys.Update([]*Entity{entity}, 0.1)
	if shadow.OffsetX != 0 {
		t.Errorf("OffsetX = %f, want 0 (throttled)", shadow.OffsetX)
	}

	// Update with enough dt should trigger
	sys.Update([]*Entity{entity}, 0.5)
	if shadow.OffsetX >= 0 {
		t.Errorf("OffsetX = %f, want negative (dawn = shadows west)", shadow.OffsetX)
	}
}

func TestTimeOfDayShadowDirectionSystem_ShadowDirections(t *testing.T) {
	tests := []struct {
		name      string
		timeOfDay palette.TimeOfDay
		wantXSign int // -1 = negative, 0 = zero, 1 = positive
		wantOpLow float64
		wantOpHi  float64
	}{
		{"dawn_shadows_west", palette.TimeOfDayDawn, -1, 0.15, 0.40},
		{"day_shadows_centered", palette.TimeOfDayDay, 0, 0.30, 0.55},
		{"dusk_shadows_east", palette.TimeOfDayDusk, 1, 0.15, 0.40},
		{"night_faint_centered", palette.TimeOfDayNight, 0, 0.05, 0.25},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			world := NewWorld()
			sys := NewTimeOfDayShadowDirectionSystem(world, 42)
			ls := NewTimeOfDayLightingSystem(world, 99)
			ls.ForceTimeOfDay(tt.timeOfDay)
			sys.SetLightingSystem(ls)

			entity := NewEntity(1)
			entity.AddComponent(&PositionComponent{X: 100, Y: 100})
			entity.AddComponent(&ColliderComponent{Width: 32, Height: 32})
			shadow := NewDropShadowComponent()
			entity.AddComponent(shadow)

			sys.Update([]*Entity{entity}, 1.0)

			// Check X direction
			switch tt.wantXSign {
			case -1:
				if shadow.OffsetX >= 0 {
					t.Errorf("OffsetX = %f, want negative", shadow.OffsetX)
				}
			case 0:
				if shadow.OffsetX != 0 {
					t.Errorf("OffsetX = %f, want 0", shadow.OffsetX)
				}
			case 1:
				if shadow.OffsetX <= 0 {
					t.Errorf("OffsetX = %f, want positive", shadow.OffsetX)
				}
			}

			// Check opacity range
			if shadow.Opacity < tt.wantOpLow || shadow.Opacity > tt.wantOpHi {
				t.Errorf("Opacity = %f, want [%f, %f]", shadow.Opacity, tt.wantOpLow, tt.wantOpHi)
			}
		})
	}
}

func TestTimeOfDayShadowDirectionSystem_DawnDuskStretch(t *testing.T) {
	world := NewWorld()
	sys := NewTimeOfDayShadowDirectionSystem(world, 42)
	ls := NewTimeOfDayLightingSystem(world, 99)
	sys.SetLightingSystem(ls)

	entity := NewEntity(1)
	entity.AddComponent(&PositionComponent{X: 50, Y: 50})
	entity.AddComponent(&ColliderComponent{Width: 32, Height: 32})
	shadow := NewDropShadowComponent()
	entity.AddComponent(shadow)

	// Day: baseline dimensions
	ls.ForceTimeOfDay(palette.TimeOfDayDay)
	sys.timeSinceCheck = sys.updateInterval // Force update
	sys.Update([]*Entity{entity}, 1.0)
	dayWidth := shadow.ShadowWidth
	dayHeight := shadow.ShadowHeight

	// Dawn: shadows should be stretched (taller)
	ls.ForceTimeOfDay(palette.TimeOfDayDawn)
	sys.timeSinceCheck = sys.updateInterval
	sys.Update([]*Entity{entity}, 1.0)
	dawnHeight := shadow.ShadowHeight

	if dawnHeight <= dayHeight {
		t.Errorf("dawn ShadowHeight (%f) should be > day ShadowHeight (%f)", dawnHeight, dayHeight)
	}
	_ = dayWidth // width may also change, but height stretch is primary indicator
}

func TestTimeOfDayShadowDirectionSystem_GenreHorrorLonger(t *testing.T) {
	world := NewWorld()

	sysFantasy := NewTimeOfDayShadowDirectionSystem(world, 42)
	sysHorror := NewTimeOfDayShadowDirectionSystem(world, 42)
	sysHorror.SetGenre("horror")

	ls := NewTimeOfDayLightingSystem(world, 99)
	ls.ForceTimeOfDay(palette.TimeOfDayDawn)
	sysFantasy.SetLightingSystem(ls)
	sysHorror.SetLightingSystem(ls)

	mkEntity := func() (*Entity, *DropShadowComponent) {
		e := NewEntity(1)
		e.AddComponent(&PositionComponent{X: 50, Y: 50})
		e.AddComponent(&ColliderComponent{Width: 32, Height: 32})
		s := NewDropShadowComponent()
		e.AddComponent(s)
		return e, s
	}

	e1, s1 := mkEntity()
	sysFantasy.Update([]*Entity{e1}, 1.0)
	fantasyOffsetX := s1.OffsetX

	e2, s2 := mkEntity()
	sysHorror.Update([]*Entity{e2}, 1.0)
	horrorOffsetX := s2.OffsetX

	// Horror shadows should stretch further (more negative at dawn)
	if horrorOffsetX >= fantasyOffsetX {
		t.Errorf("horror OffsetX (%f) should be < fantasy OffsetX (%f) at dawn", horrorOffsetX, fantasyOffsetX)
	}
}

func TestTimeOfDayShadowDirectionSystem_LargerEntityScaling(t *testing.T) {
	world := NewWorld()
	sys := NewTimeOfDayShadowDirectionSystem(world, 42)
	ls := NewTimeOfDayLightingSystem(world, 99)
	ls.ForceTimeOfDay(palette.TimeOfDayDusk)
	sys.SetLightingSystem(ls)

	entity := NewEntity(1)
	entity.AddComponent(&PositionComponent{X: 50, Y: 50})
	entity.AddComponent(&ColliderComponent{Width: 48, Height: 48})
	shadow := NewDropShadowComponent()
	entity.AddComponent(shadow)

	sys.Update([]*Entity{entity}, 1.0)

	// Shadow should use collider dimensions for scaling
	if shadow.OffsetX <= 0 {
		t.Errorf("OffsetX = %f, want positive (dusk = shadows east)", shadow.OffsetX)
	}
	// Offset should scale with entity size (48/32 = 1.5x)
	expectedMinOffset := 5.0
	if shadow.OffsetX < expectedMinOffset {
		t.Errorf("OffsetX = %f, want > %f for 48px entity", shadow.OffsetX, expectedMinOffset)
	}
}

func TestTimeOfDayShadowDirectionSystem_DisabledShadowSkipped(t *testing.T) {
	world := NewWorld()
	sys := NewTimeOfDayShadowDirectionSystem(world, 42)
	ls := NewTimeOfDayLightingSystem(world, 99)
	ls.ForceTimeOfDay(palette.TimeOfDayDawn)
	sys.SetLightingSystem(ls)

	entity := NewEntity(1)
	entity.AddComponent(&PositionComponent{X: 50, Y: 50})
	entity.AddComponent(&ColliderComponent{Width: 32, Height: 32})
	shadow := NewDropShadowComponent()
	shadow.Enabled = false
	shadow.OffsetX = 99.0 // Sentinel value
	entity.AddComponent(shadow)

	sys.Update([]*Entity{entity}, 1.0)

	if shadow.OffsetX != 99.0 {
		t.Errorf("OffsetX changed to %f, want 99.0 (disabled shadow should be skipped)", shadow.OffsetX)
	}
}

func TestComputeShadowParams_AllTimes(t *testing.T) {
	world := NewWorld()
	sys := NewTimeOfDayShadowDirectionSystem(world, 42)

	tests := []struct {
		name string
		tod  palette.TimeOfDay
	}{
		{"dawn", palette.TimeOfDayDawn},
		{"day", palette.TimeOfDayDay},
		{"dusk", palette.TimeOfDayDusk},
		{"night", palette.TimeOfDayNight},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			offsetX, offsetYScale, opMult, wScale, hScale := sys.computeShadowParams(tt.tod)

			// All scales should be positive
			if offsetYScale <= 0 {
				t.Errorf("offsetYScale = %f, want > 0", offsetYScale)
			}
			if opMult <= 0 {
				t.Errorf("opacityMult = %f, want > 0", opMult)
			}
			if wScale <= 0 {
				t.Errorf("widthScale = %f, want > 0", wScale)
			}
			if hScale <= 0 {
				t.Errorf("heightScale = %f, want > 0", hScale)
			}
			_ = offsetX // sign varies by time
		})
	}
}

func BenchmarkTimeOfDayShadowDirectionSystem(b *testing.B) {
	world := NewWorld()
	sys := NewTimeOfDayShadowDirectionSystem(world, 42)
	ls := NewTimeOfDayLightingSystem(world, 99)
	ls.ForceTimeOfDay(palette.TimeOfDayDawn)
	sys.SetLightingSystem(ls)

	entities := make([]*Entity, 200)
	for i := range entities {
		e := NewEntity(uint64(i + 1))
		e.AddComponent(&PositionComponent{X: float64(i * 10), Y: float64(i * 10)})
		e.AddComponent(&ColliderComponent{Width: 32, Height: 32})
		shadow := NewDropShadowComponent()
		e.AddComponent(shadow)
		entities[i] = e
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.timeSinceCheck = sys.updateInterval // Force update each iteration
		sys.Update(entities, 0.016)
	}
}
