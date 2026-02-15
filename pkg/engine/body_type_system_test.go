package engine

import (
	"testing"
)

func TestBodyTypeComponent_Type(t *testing.T) {
	comp := NewBodyTypeComponent()
	if comp.Type() != "body_type" {
		t.Errorf("Type() = %q, want %q", comp.Type(), "body_type")
	}
}

func TestBodyTypeComponent_Defaults(t *testing.T) {
	comp := NewBodyTypeComponent()
	if comp.BodyType != 0 {
		t.Errorf("default BodyType = %d, want 0", comp.BodyType)
	}
	if comp.Assigned {
		t.Error("default Assigned should be false")
	}
	if comp.Dirty {
		t.Error("default Dirty should be false")
	}
}

func TestBodyTypeSystem_Update(t *testing.T) {
	world := NewWorld()
	sys := NewBodyTypeSystem(world, 42)
	sys.SetGenre("fantasy")

	// Create an entity with a sprite component
	entity := NewEntity(12345)
	sprite := &EbitenSprite{}
	entity.AddComponent(sprite)
	world.AddEntity(entity)

	entities := []*Entity{entity}

	// First update should not assign (scan interval = 2.0)
	sys.Update(entities, 0.5)
	_, has := entity.GetComponent("body_type")
	if has {
		t.Error("body type should not be assigned before scan interval")
	}

	// Advance past scan interval
	sys.Update(entities, 2.0)
	comp, has := entity.GetComponent("body_type")
	if !has {
		t.Fatal("body type should be assigned after scan interval")
	}

	btComp, ok := comp.(*BodyTypeComponent)
	if !ok {
		t.Fatal("component should be *BodyTypeComponent")
	}
	if !btComp.Assigned {
		t.Error("BodyTypeComponent.Assigned should be true")
	}
	if btComp.GenreID != "fantasy" {
		t.Errorf("GenreID = %q, want %q", btComp.GenreID, "fantasy")
	}
	if btComp.BodyType < 0 || btComp.BodyType >= 8 {
		t.Errorf("BodyType = %d, out of range [0,8)", btComp.BodyType)
	}
}

func TestBodyTypeSystem_SkipsNonSpriteEntities(t *testing.T) {
	world := NewWorld()
	sys := NewBodyTypeSystem(world, 42)

	entity := NewEntity(1)
	// No sprite component
	entities := []*Entity{entity}

	sys.Update(entities, 3.0)
	_, has := entity.GetComponent("body_type")
	if has {
		t.Error("should not assign body type to entity without sprite")
	}
}

func TestBodyTypeSystem_GenreChange(t *testing.T) {
	world := NewWorld()
	sys := NewBodyTypeSystem(world, 42)
	sys.SetGenre("fantasy")

	entity := NewEntity(99)
	sprite := &EbitenSprite{}
	entity.AddComponent(sprite)

	entities := []*Entity{entity}

	// First scan assigns body type
	sys.Update(entities, 3.0)
	comp, _ := entity.GetComponent("body_type")
	btComp := comp.(*BodyTypeComponent)
	oldType := btComp.BodyType

	// Change genre
	sys.SetGenre("horror")
	sys.Update(entities, 3.0)

	// Should be re-derived and marked dirty
	if btComp.GenreID != "horror" {
		t.Errorf("GenreID = %q, want %q", btComp.GenreID, "horror")
	}
	if !btComp.Dirty {
		t.Error("Dirty should be true after genre change")
	}
	// Body type might be same or different — just verify it's valid
	_ = oldType
	if btComp.BodyType < 0 || btComp.BodyType >= 8 {
		t.Errorf("BodyType = %d, out of range [0,8)", btComp.BodyType)
	}
}

func TestBodyTypeSystem_NilEntity(t *testing.T) {
	world := NewWorld()
	sys := NewBodyTypeSystem(world, 42)

	// Should not panic with nil entity
	entities := []*Entity{nil}
	sys.Update(entities, 3.0)
}

func TestBodyTypeSystem_Deterministic(t *testing.T) {
	world1 := NewWorld()
	sys1 := NewBodyTypeSystem(world1, 42)
	sys1.SetGenre("fantasy")

	world2 := NewWorld()
	sys2 := NewBodyTypeSystem(world2, 42)
	sys2.SetGenre("fantasy")

	entity1 := NewEntity(555)
	entity1.AddComponent(&EbitenSprite{})
	entity2 := NewEntity(555)
	entity2.AddComponent(&EbitenSprite{})

	sys1.Update([]*Entity{entity1}, 3.0)
	sys2.Update([]*Entity{entity2}, 3.0)

	comp1, _ := entity1.GetComponent("body_type")
	comp2, _ := entity2.GetComponent("body_type")
	bt1 := comp1.(*BodyTypeComponent).BodyType
	bt2 := comp2.(*BodyTypeComponent).BodyType

	if bt1 != bt2 {
		t.Errorf("same entity ID should produce same body type: %d != %d", bt1, bt2)
	}
}
