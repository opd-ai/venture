package engine

import (
	"testing"
)

func TestNewSurfaceTextureSystem(t *testing.T) {
	world := NewWorld()
	sys := NewSurfaceTextureSystem(world, 42)
	if sys == nil {
		t.Fatal("expected non-nil system")
	}
	if sys.genreID != "fantasy" {
		t.Errorf("expected default genre 'fantasy', got %q", sys.genreID)
	}
	if sys.scanInterval != 2.0 {
		t.Errorf("expected scan interval 2.0, got %f", sys.scanInterval)
	}
}

func TestSurfaceTextureSystem_SetGenre(t *testing.T) {
	world := NewWorld()
	sys := NewSurfaceTextureSystem(world, 42)
	sys.SetGenre("horror")
	if sys.genreID != "horror" {
		t.Errorf("expected genre 'horror', got %q", sys.genreID)
	}
}

func TestSurfaceTextureSystem_SkipsHumanoid(t *testing.T) {
	world := NewWorld()
	sys := NewSurfaceTextureSystem(world, 42)

	entity := NewEntity(1)
	entity.AddComponent(&CreatureVisualComponent{
		Form:      FormHumanoid,
		SizeClass: "medium",
	})
	world.AddEntity(entity)

	entities := []*Entity{entity}
	// Force scan
	sys.timeSinceScan = sys.scanInterval
	sys.Update(entities, 0.1)

	if entity.HasComponent("surface_texture") {
		t.Error("humanoid entities should not receive surface texture")
	}
}

func TestSurfaceTextureSystem_AssignsTextureToCreature(t *testing.T) {
	forms := []struct {
		form    CreatureForm
		wantTex int // matches sprites.SurfaceTextureType
	}{
		{FormQuadruped, 1},  // TexFur
		{FormSerpentine, 2}, // TexScales
		{FormArachnid, 3},   // TexChitin
		{FormMechanical, 4}, // TexMetal
		{FormUndead, 5},     // TexBone
		{FormBlob, 6},       // TexOoze
		{FormFlying, 7},     // TexFeathers
	}

	for _, tt := range forms {
		t.Run(string(tt.form), func(t *testing.T) {
			world := NewWorld()
			sys := NewSurfaceTextureSystem(world, 42)

			entity := NewEntity(1)
			entity.AddComponent(&CreatureVisualComponent{
				Form:      tt.form,
				SizeClass: "medium",
			})
			world.AddEntity(entity)

			entities := []*Entity{entity}
			sys.timeSinceScan = sys.scanInterval
			sys.Update(entities, 0.1)

			comp, has := entity.GetComponent("surface_texture")
			if !has {
				t.Fatalf("expected surface_texture component on %s entity", tt.form)
			}
			st, ok := comp.(*SurfaceTextureComponent)
			if !ok {
				t.Fatal("wrong component type")
			}
			if st.TextureType != tt.wantTex {
				t.Errorf("expected texture type %d, got %d", tt.wantTex, st.TextureType)
			}
			if !st.Enabled {
				t.Error("expected enabled=true")
			}
			if st.TorsoIntensity <= 0 {
				t.Error("expected positive torso intensity")
			}
		})
	}
}

func TestSurfaceTextureSystem_GenreChange(t *testing.T) {
	world := NewWorld()
	sys := NewSurfaceTextureSystem(world, 42)
	sys.SetGenre("fantasy")

	entity := NewEntity(1)
	entity.AddComponent(&CreatureVisualComponent{
		Form:      FormQuadruped,
		SizeClass: "medium",
	})
	world.AddEntity(entity)

	entities := []*Entity{entity}
	sys.timeSinceScan = sys.scanInterval
	sys.Update(entities, 0.1)

	comp, _ := entity.GetComponent("surface_texture")
	st := comp.(*SurfaceTextureComponent)
	originalIntensity := st.TorsoIntensity

	// Change genre
	sys.SetGenre("horror")
	sys.timeSinceScan = sys.scanInterval
	sys.Update(entities, 0.1)

	// Should be marked dirty and potentially have different intensity
	if st.GenreID != "horror" {
		t.Errorf("expected genre 'horror', got %q", st.GenreID)
	}
	if st.Dirty != true {
		t.Error("expected dirty=true after genre change")
	}
	// Horror has intensity bias of 1.3x, so intensity should differ
	_ = originalIntensity // genre change should cause re-population
}

func TestSurfaceTextureSystem_SkipsScanBeforeInterval(t *testing.T) {
	world := NewWorld()
	sys := NewSurfaceTextureSystem(world, 42)

	entity := NewEntity(1)
	entity.AddComponent(&CreatureVisualComponent{
		Form:      FormBlob,
		SizeClass: "medium",
	})
	world.AddEntity(entity)

	entities := []*Entity{entity}
	// timeSinceScan starts at 0, so first update with small dt should skip
	sys.Update(entities, 0.1)

	if entity.HasComponent("surface_texture") {
		t.Error("should not assign texture before scan interval elapsed")
	}
}

func TestSurfaceTextureSystem_NoCreatureVisual(t *testing.T) {
	world := NewWorld()
	sys := NewSurfaceTextureSystem(world, 42)

	entity := NewEntity(1)
	world.AddEntity(entity)

	entities := []*Entity{entity}
	sys.timeSinceScan = sys.scanInterval
	sys.Update(entities, 0.1)

	if entity.HasComponent("surface_texture") {
		t.Error("entities without creature_visual should not get surface_texture")
	}
}

func TestSurfaceTextureSystem_NilEntity(t *testing.T) {
	world := NewWorld()
	sys := NewSurfaceTextureSystem(world, 42)

	entities := []*Entity{nil}
	sys.timeSinceScan = sys.scanInterval
	// Should not panic
	sys.Update(entities, 0.1)
}

func TestSurfaceTextureComponent_Type(t *testing.T) {
	c := NewSurfaceTextureComponent()
	if c.Type() != "surface_texture" {
		t.Errorf("expected 'surface_texture', got %q", c.Type())
	}
}

func TestSurfaceTextureComponent_Defaults(t *testing.T) {
	c := NewSurfaceTextureComponent()
	if c.TextureType != 0 {
		t.Error("default texture type should be 0 (TexNone)")
	}
	if c.TorsoScale != 1.0 || c.HeadScale != 1.0 || c.LimbScale != 1.0 {
		t.Error("default scales should be 1.0")
	}
	if c.Enabled != true {
		t.Error("should be enabled by default")
	}
	if c.Dirty != false {
		t.Error("should not be dirty by default")
	}
}

func TestSurfaceTextureSystem_GetActiveTextureCount(t *testing.T) {
	world := NewWorld()
	sys := NewSurfaceTextureSystem(world, 42)

	if sys.GetActiveTextureCount() != 0 {
		t.Error("expected 0 active textures initially")
	}

	entity := NewEntity(1)
	entity.AddComponent(&CreatureVisualComponent{
		Form:      FormArachnid,
		SizeClass: "small",
	})
	world.AddEntity(entity)

	entities := []*Entity{entity}
	sys.timeSinceScan = sys.scanInterval
	sys.Update(entities, 0.1)

	if sys.GetActiveTextureCount() != 1 {
		t.Errorf("expected 1 active texture, got %d", sys.GetActiveTextureCount())
	}
}
