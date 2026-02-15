package engine

import (
	"testing"
)

func TestNewEntityDropShadowSystem(t *testing.T) {
	world := NewWorld()
	sys := NewEntityDropShadowSystem(world, 12345)
	if sys == nil {
		t.Fatal("NewEntityDropShadowSystem returned nil")
	}
	if sys.genreID != "fantasy" {
		t.Errorf("default genre = %q, want %q", sys.genreID, "fantasy")
	}
}

func TestEntityDropShadowSystem_SetGenre(t *testing.T) {
	world := NewWorld()
	sys := NewEntityDropShadowSystem(world, 12345)

	tests := []struct {
		genre       string
		wantOpScale float64
	}{
		{"horror", 1.4},
		{"cyberpunk", 1.1},
		{"sci-fi", 0.9},
		{"fantasy", 1.0},
		{"postapoc", 1.2},
		{"unknown", 1.0},
	}

	for _, tt := range tests {
		t.Run(tt.genre, func(t *testing.T) {
			sys.SetGenre(tt.genre)
			if sys.preset.OpacityScale != tt.wantOpScale {
				t.Errorf("genre %q: OpacityScale = %v, want %v",
					tt.genre, sys.preset.OpacityScale, tt.wantOpScale)
			}
		})
	}
}

func TestEntityDropShadowSystem_AttachesShadow(t *testing.T) {
	world := NewWorld()
	sys := NewEntityDropShadowSystem(world, 42)

	entity := NewEntity(1)
	entity.AddComponent(&PositionComponent{X: 100, Y: 200})
	entity.AddComponent(&ColliderComponent{Width: 32, Height: 32, Solid: true})

	entities := []*Entity{entity}

	// First call with enough deltaTime triggers full scan
	sys.Update(entities, 2.0)

	comp, _ := entity.GetComponent("drop_shadow")
	shadow, ok := comp.(*DropShadowComponent)
	if !ok || shadow == nil {
		t.Fatal("expected DropShadowComponent to be attached")
	}
	if !shadow.Enabled {
		t.Error("shadow should be enabled")
	}
}

func TestEntityDropShadowSystem_SkipsNoPosition(t *testing.T) {
	world := NewWorld()
	sys := NewEntityDropShadowSystem(world, 42)

	entity := NewEntity(1)
	entity.AddComponent(&ColliderComponent{Width: 32, Height: 32, Solid: true})

	entities := []*Entity{entity}
	sys.Update(entities, 2.0)

	comp, _ := entity.GetComponent("drop_shadow")
	if comp != nil {
		t.Error("entity without position should not get a shadow")
	}
}

func TestEntityDropShadowSystem_SkipsNoColliderOrSprite(t *testing.T) {
	world := NewWorld()
	sys := NewEntityDropShadowSystem(world, 42)

	entity := NewEntity(1)
	entity.AddComponent(&PositionComponent{X: 10, Y: 20})

	entities := []*Entity{entity}
	sys.Update(entities, 2.0)

	comp, _ := entity.GetComponent("drop_shadow")
	if comp != nil {
		t.Error("entity without collider or sprite should not get a shadow")
	}
}

func TestEntityDropShadowSystem_ShadowDimensions(t *testing.T) {
	tests := []struct {
		name       string
		colW, colH float64
		wantW      float64
		wantH      float64
	}{
		{"normal 32x32", 32, 32, 22.4, 8.0},
		{"large 64x64", 64, 64, 44.8, 16.0},
		{"tiny 4x4", 4, 4, 6.0, 3.0}, // clamped minimums
		{"wide 48x16", 48, 16, 33.6, 4.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			world := NewWorld()
			sys := NewEntityDropShadowSystem(world, 99)

			entity := NewEntity(1)
			entity.AddComponent(&PositionComponent{X: 0, Y: 0})
			entity.AddComponent(&ColliderComponent{Width: tt.colW, Height: tt.colH, Solid: true})

			sys.Update([]*Entity{entity}, 2.0)

			comp, _ := entity.GetComponent("drop_shadow")
			shadow := comp.(*DropShadowComponent)

			if !floatClose(shadow.ShadowWidth, tt.wantW, 0.01) {
				t.Errorf("ShadowWidth = %v, want %v", shadow.ShadowWidth, tt.wantW)
			}
			if !floatClose(shadow.ShadowHeight, tt.wantH, 0.01) {
				t.Errorf("ShadowHeight = %v, want %v", shadow.ShadowHeight, tt.wantH)
			}
		})
	}
}

func TestEntityDropShadowSystem_GenreColorPresets(t *testing.T) {
	tests := []struct {
		genre string
		wantR float64
		wantG float64
		wantB float64
	}{
		{"horror", 0.05, 0.0, 0.0},
		{"cyberpunk", 0.0, 0.02, 0.06},
		{"fantasy", 0.02, 0.01, 0.0},
		{"scifi", 0.02, 0.02, 0.04},
		{"postapoc", 0.04, 0.03, 0.01},
	}

	for _, tt := range tests {
		t.Run(tt.genre, func(t *testing.T) {
			world := NewWorld()
			sys := NewEntityDropShadowSystem(world, 7)
			sys.SetGenre(tt.genre)

			entity := NewEntity(1)
			entity.AddComponent(&PositionComponent{X: 0, Y: 0})
			entity.AddComponent(&ColliderComponent{Width: 32, Height: 32, Solid: true})

			sys.Update([]*Entity{entity}, 2.0)

			comp, _ := entity.GetComponent("drop_shadow")
			shadow := comp.(*DropShadowComponent)

			if !floatClose(shadow.ColorR, tt.wantR, 0.001) {
				t.Errorf("ColorR = %v, want %v", shadow.ColorR, tt.wantR)
			}
			if !floatClose(shadow.ColorG, tt.wantG, 0.001) {
				t.Errorf("ColorG = %v, want %v", shadow.ColorG, tt.wantG)
			}
			if !floatClose(shadow.ColorB, tt.wantB, 0.001) {
				t.Errorf("ColorB = %v, want %v", shadow.ColorB, tt.wantB)
			}
		})
	}
}

func TestEntityDropShadowSystem_ThrottledUpdate(t *testing.T) {
	world := NewWorld()
	sys := NewEntityDropShadowSystem(world, 42)

	entity := NewEntity(1)
	entity.AddComponent(&PositionComponent{X: 0, Y: 0})
	entity.AddComponent(&ColliderComponent{Width: 32, Height: 32, Solid: true})

	entities := []*Entity{entity}

	// Small deltaTime should not trigger full scan
	sys.Update(entities, 0.1)
	comp, _ := entity.GetComponent("drop_shadow")
	if comp != nil {
		t.Error("shadow should not be attached during throttled (non-full-scan) update")
	}

	// After accumulating >= 1.0s, full scan should attach shadow
	sys.Update(entities, 1.0)
	comp, _ = entity.GetComponent("drop_shadow")
	if comp == nil {
		t.Error("shadow should be attached after full scan interval elapsed")
	}
}

func TestEntityDropShadowSystem_OpacityClamping(t *testing.T) {
	world := NewWorld()
	sys := NewEntityDropShadowSystem(world, 42)
	sys.SetGenre("horror") // OpacityScale=1.4 → 0.35*1.4=0.49

	entity := NewEntity(1)
	entity.AddComponent(&PositionComponent{X: 0, Y: 0})
	entity.AddComponent(&ColliderComponent{Width: 32, Height: 32, Solid: true})

	sys.Update([]*Entity{entity}, 2.0)

	comp, _ := entity.GetComponent("drop_shadow")
	shadow := comp.(*DropShadowComponent)

	if shadow.Opacity < 0.1 || shadow.Opacity > 0.7 {
		t.Errorf("opacity %v outside clamped range [0.1, 0.7]", shadow.Opacity)
	}
}

func TestDropShadowComponent_Type(t *testing.T) {
	c := NewDropShadowComponent()
	if c.Type() != "drop_shadow" {
		t.Errorf("Type() = %q, want %q", c.Type(), "drop_shadow")
	}
}

func TestDropShadowComponent_Defaults(t *testing.T) {
	c := NewDropShadowComponent()
	if c.ShadowWidth != 20.0 {
		t.Errorf("default ShadowWidth = %v, want 20.0", c.ShadowWidth)
	}
	if c.ShadowHeight != 8.0 {
		t.Errorf("default ShadowHeight = %v, want 8.0", c.ShadowHeight)
	}
	if c.Opacity != 0.35 {
		t.Errorf("default Opacity = %v, want 0.35", c.Opacity)
	}
	if !c.Enabled {
		t.Error("default Enabled should be true")
	}
}

func floatClose(a, b, eps float64) bool {
	d := a - b
	if d < 0 {
		d = -d
	}
	return d < eps
}
