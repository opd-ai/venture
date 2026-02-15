package engine

import (
	"testing"
)

func TestSpriteDepthShadingComponent_Type(t *testing.T) {
	c := NewSpriteDepthShadingComponent()
	if c.Type() != "sprite_depth_shading" {
		t.Errorf("Type() = %s, want sprite_depth_shading", c.Type())
	}
}

func TestNewSpriteDepthShadingComponent_Defaults(t *testing.T) {
	c := NewSpriteDepthShadingComponent()
	if c.LightIntensity <= 0 {
		t.Error("LightIntensity should be positive")
	}
	if c.EdgeDarkening <= 0 {
		t.Error("EdgeDarkening should be positive")
	}
	if !c.Enabled {
		t.Error("should be enabled by default")
	}
	if c.Dirty {
		t.Error("should not be dirty by default")
	}
}

func TestNewSpriteDepthShadingSystem(t *testing.T) {
	world := NewWorld()
	sys := NewSpriteDepthShadingSystem(world, 12345)
	if sys == nil {
		t.Fatal("system should not be nil")
	}
	if sys.genreID != "fantasy" {
		t.Errorf("default genreID = %s, want fantasy", sys.genreID)
	}
}

func TestSpriteDepthShadingSystem_SetGenre(t *testing.T) {
	tests := []struct {
		name    string
		genre   string
		wantR   float64
		checkFn func(spriteDepthGenrePreset) bool
	}{
		{"horror", "horror", 0.85, func(p spriteDepthGenrePreset) bool { return p.EdgeDarkening > 0.35 }},
		{"cyberpunk", "cyberpunk", 0.80, func(p spriteDepthGenrePreset) bool { return p.LightIntensity > 0.35 }},
		{"scifi", "sci-fi", 0.90, func(p spriteDepthGenrePreset) bool { return p.TintB > 0.95 }},
		{"postapoc", "postapoc", 1.0, func(p spriteDepthGenrePreset) bool { return p.DitherStrength > 0.08 }},
		{"fantasy", "fantasy", 1.0, func(p spriteDepthGenrePreset) bool { return p.LightIntensity > 0.30 }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			world := NewWorld()
			sys := NewSpriteDepthShadingSystem(world, 42)
			sys.SetGenre(tt.genre)
			if sys.preset.TintR != tt.wantR {
				t.Errorf("TintR = %f, want %f", sys.preset.TintR, tt.wantR)
			}
			if !tt.checkFn(sys.preset) {
				t.Errorf("genre %s preset check failed", tt.genre)
			}
		})
	}
}

func TestSpriteDepthShadingSystem_Update_AttachesComponent(t *testing.T) {
	world := NewWorld()
	sys := NewSpriteDepthShadingSystem(world, 42)

	entity := world.CreateEntity()
	entity.AddComponent(NewSpriteComponent(32, 32, nil))

	entities := []*Entity{entity}
	// First update won't trigger (scan interval = 2.0)
	sys.Update(entities, 0.5)
	if entity.HasComponent("sprite_depth_shading") {
		t.Error("should not attach on first update before scan interval")
	}

	// Exceed scan interval
	sys.Update(entities, 2.0)
	if !entity.HasComponent("sprite_depth_shading") {
		t.Error("should attach shading component after scan interval")
	}

	comp, _ := entity.GetComponent("sprite_depth_shading")
	shading, ok := comp.(*SpriteDepthShadingComponent)
	if !ok {
		t.Fatal("component should be *SpriteDepthShadingComponent")
	}
	if !shading.Enabled {
		t.Error("shading should be enabled")
	}
}

func TestSpriteDepthShadingSystem_Update_SkipsNonSprite(t *testing.T) {
	world := NewWorld()
	sys := NewSpriteDepthShadingSystem(world, 42)

	entity := world.CreateEntity()
	// No sprite component
	entities := []*Entity{entity}
	sys.Update(entities, 3.0)

	if entity.HasComponent("sprite_depth_shading") {
		t.Error("should not attach shading to entity without sprite")
	}
}

func TestSpriteDepthShadingSystem_Update_GenreChange(t *testing.T) {
	world := NewWorld()
	sys := NewSpriteDepthShadingSystem(world, 42)

	entity := world.CreateEntity()
	entity.AddComponent(NewSpriteComponent(32, 32, nil))

	entities := []*Entity{entity}
	sys.Update(entities, 3.0)

	// Change genre
	sys.SetGenre("horror")
	sys.Update(entities, 3.0)

	comp, _ := entity.GetComponent("sprite_depth_shading")
	shading := comp.(*SpriteDepthShadingComponent)
	if shading.TintR != 0.85 {
		t.Errorf("TintR = %f, want 0.85 for horror", shading.TintR)
	}
	if !shading.Dirty {
		t.Error("shading should be dirty after genre change")
	}
}

func TestSpriteDepthShadingSystem_Update_NilEntity(t *testing.T) {
	world := NewWorld()
	sys := NewSpriteDepthShadingSystem(world, 42)
	// Should not panic
	sys.Update([]*Entity{nil}, 3.0)
}

func TestApproxEqual(t *testing.T) {
	tests := []struct {
		a, b, eps float64
		want      bool
	}{
		{1.0, 1.0, 0.001, true},
		{1.0, 1.0005, 0.001, true},
		{1.0, 1.01, 0.001, false},
		{0.0, 0.0, 0.001, true},
	}
	for _, tt := range tests {
		got := approxEqual(tt.a, tt.b, tt.eps)
		if got != tt.want {
			t.Errorf("approxEqual(%f,%f,%f) = %v, want %v", tt.a, tt.b, tt.eps, got, tt.want)
		}
	}
}

func BenchmarkSpriteDepthShadingSystem_Update(b *testing.B) {
	world := NewWorld()
	sys := NewSpriteDepthShadingSystem(world, 42)

	entities := make([]*Entity, 100)
	for i := range entities {
		e := world.CreateEntity()
		e.AddComponent(NewSpriteComponent(32, 32, nil))
		entities[i] = e
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.timeSinceScan = 3.0 // Force scan every iteration
		sys.Update(entities, 0.016)
	}
}
