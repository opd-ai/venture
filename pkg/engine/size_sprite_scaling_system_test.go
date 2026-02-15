package engine

import (
	"testing"
)

func TestSizeSpriteScalingSystem_SkipsEntitiesWithoutCreatureVisual(t *testing.T) {
	world := NewWorld()
	sys := NewSizeSpriteScalingSystem(world)

	e := NewEntity(1)
	e.AddComponent(&AnimationComponent{Seed: 1})
	entities := []*Entity{e}

	sys.Update(entities, 0.016)

	anim := e.GetAnimation()
	if anim != nil && anim.Dirty {
		t.Error("entity without creature_visual should not be marked dirty")
	}
}

func TestSizeSpriteScalingSystem_SkipsMediumSize(t *testing.T) {
	world := NewWorld()
	sys := NewSizeSpriteScalingSystem(world)

	e := NewEntity(1)
	e.AddComponent(&AnimationComponent{Seed: 1})
	e.AddComponent(&CreatureVisualComponent{
		Form:      FormQuadruped,
		SizeClass: "medium",
	})
	entities := []*Entity{e}

	sys.Update(entities, 0.016)

	anim := e.GetAnimation()
	if anim.Dirty {
		t.Error("medium-sized entity should not be marked dirty")
	}
}

func TestSizeSpriteScalingSystem_MarksLargeEntityDirty(t *testing.T) {
	world := NewWorld()
	sys := NewSizeSpriteScalingSystem(world)

	e := NewEntity(1)
	anim := &AnimationComponent{Seed: 1}
	e.AddComponent(anim)
	e.AddComponent(&CreatureVisualComponent{
		Form:      FormFlying,
		SizeClass: "huge",
	})
	entities := []*Entity{e}

	sys.Update(entities, 0.016)

	if !anim.Dirty {
		t.Error("huge entity should be marked dirty for sprite regeneration")
	}
}

func TestSizeSpriteScalingSystem_SetsMarkerComponent(t *testing.T) {
	world := NewWorld()
	sys := NewSizeSpriteScalingSystem(world)

	e := NewEntity(1)
	e.AddComponent(&AnimationComponent{Seed: 1})
	e.AddComponent(&CreatureVisualComponent{
		Form:      FormArachnid,
		SizeClass: "tiny",
	})
	entities := []*Entity{e}

	sys.Update(entities, 0.016)

	markerComp, ok := e.GetComponent("size_sprite_marker")
	if !ok {
		t.Fatal("marker component should be set after propagation")
	}
	marker := markerComp.(*SizeSpriteMarkerComponent)
	if marker.PropagatedSizeClass != "tiny" {
		t.Errorf("marker should record 'tiny', got %q", marker.PropagatedSizeClass)
	}
}

func TestSizeSpriteScalingSystem_SkipsAlreadyPropagated(t *testing.T) {
	world := NewWorld()
	sys := NewSizeSpriteScalingSystem(world)

	e := NewEntity(1)
	anim := &AnimationComponent{Seed: 1}
	e.AddComponent(anim)
	e.AddComponent(&CreatureVisualComponent{
		Form:      FormQuadruped,
		SizeClass: "large",
	})
	entities := []*Entity{e}

	// First update — should mark dirty
	sys.Update(entities, 0.016)
	if !anim.Dirty {
		t.Fatal("first update should mark dirty")
	}

	// Reset dirty flag
	anim.Dirty = false

	// Second update — should NOT mark dirty (already propagated)
	sys.Update(entities, 0.016)
	if anim.Dirty {
		t.Error("second update should skip already-propagated entity")
	}
}

func TestSizeSpriteScalingSystem_RePropagatOnSizeChange(t *testing.T) {
	world := NewWorld()
	sys := NewSizeSpriteScalingSystem(world)

	e := NewEntity(1)
	anim := &AnimationComponent{Seed: 1}
	e.AddComponent(anim)
	cv := &CreatureVisualComponent{
		Form:      FormBlob,
		SizeClass: "small",
	}
	e.AddComponent(cv)
	entities := []*Entity{e}

	// First propagation
	sys.Update(entities, 0.016)
	anim.Dirty = false

	// Change size class (e.g., growth spell)
	cv.SizeClass = "large"

	// Should re-propagate
	sys.Update(entities, 0.016)
	if !anim.Dirty {
		t.Error("size change should trigger re-propagation")
	}

	markerComp, _ := e.GetComponent("size_sprite_marker")
	marker := markerComp.(*SizeSpriteMarkerComponent)
	if marker.PropagatedSizeClass != "large" {
		t.Errorf("marker should update to 'large', got %q", marker.PropagatedSizeClass)
	}
}

func TestSizeSpriteScalingSystem_SkipsEntitiesWithoutAnimation(t *testing.T) {
	world := NewWorld()
	sys := NewSizeSpriteScalingSystem(world)

	e := NewEntity(1)
	e.AddComponent(&CreatureVisualComponent{
		Form:      FormSerpentine,
		SizeClass: "huge",
	})
	entities := []*Entity{e}

	// Should not panic
	sys.Update(entities, 0.016)

	_, ok := e.GetComponent("size_sprite_marker")
	if ok {
		t.Error("entity without animation should not get marker")
	}
}

func TestSizeSpriteMarkerComponent_Type(t *testing.T) {
	c := &SizeSpriteMarkerComponent{PropagatedSizeClass: "large"}
	if c.Type() != "size_sprite_marker" {
		t.Errorf("Type() = %q, want 'size_sprite_marker'", c.Type())
	}
}

func TestSizeSpriteScalingSystem_NilWorld(t *testing.T) {
	sys := NewSizeSpriteScalingSystem(nil)

	e := NewEntity(1)
	e.AddComponent(&AnimationComponent{Seed: 1})
	e.AddComponent(&CreatureVisualComponent{
		Form:      FormFlying,
		SizeClass: "large",
	})

	// Should not panic with nil world/logger
	sys.Update([]*Entity{e}, 0.016)
}

func TestSizeSpriteScalingSystem_MultipleSizes(t *testing.T) {
	world := NewWorld()
	sys := NewSizeSpriteScalingSystem(world)

	sizes := []string{"tiny", "small", "large", "huge"}
	entities := make([]*Entity, len(sizes))

	for i, size := range sizes {
		e := NewEntity(1)
		e.AddComponent(&AnimationComponent{Seed: int64(i)})
		e.AddComponent(&CreatureVisualComponent{
			Form:      FormHumanoid,
			SizeClass: size,
		})
		entities[i] = e
	}

	sys.Update(entities, 0.016)

	for i, e := range entities {
		anim := e.GetAnimation()
		if !anim.Dirty {
			t.Errorf("entity %d (size=%s) should be marked dirty", i, sizes[i])
		}
		markerComp, ok := e.GetComponent("size_sprite_marker")
		if !ok {
			t.Errorf("entity %d should have marker", i)
			continue
		}
		marker := markerComp.(*SizeSpriteMarkerComponent)
		if marker.PropagatedSizeClass != sizes[i] {
			t.Errorf("entity %d marker = %q, want %q", i, marker.PropagatedSizeClass, sizes[i])
		}
	}
}
