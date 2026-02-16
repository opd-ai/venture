package engine

import (
	"testing"
)

func TestNewBackAccessorySystem(t *testing.T) {
	world := NewWorld()
	sys := NewBackAccessorySystem(world, 42)
	if sys == nil {
		t.Fatal("NewBackAccessorySystem returned nil")
	}
	if sys.genreID != "fantasy" {
		t.Errorf("default genre should be fantasy, got %q", sys.genreID)
	}
}

func TestBackAccessorySystem_SetGenre(t *testing.T) {
	world := NewWorld()
	sys := NewBackAccessorySystem(world, 42)
	sys.SetGenre("horror")
	if sys.genreID != "horror" {
		t.Errorf("genre should be horror, got %q", sys.genreID)
	}
}

func TestBackAccessoryComponent_Type(t *testing.T) {
	comp := &BackAccessoryComponent{}
	if comp.Type() != "back_accessory" {
		t.Errorf("expected type 'back_accessory', got %q", comp.Type())
	}
}

func TestBackAccessorySystem_AssignsToHumanoid(t *testing.T) {
	world := NewWorld()
	sys := NewBackAccessorySystem(world, 42)

	entity := NewEntity(1)
	entity.AddComponent(NewSpriteComponent(32, 32, nil))
	entity.AddComponent(&HealthComponent{Current: 100, Max: 100})
	entity.AddComponent(&AIComponent{})

	// Should not have component before update
	if entity.HasComponent("back_accessory") {
		t.Fatal("should not have back_accessory before update")
	}

	// Run update past scan interval
	sys.Update([]*Entity{entity}, 3.0)

	// Should now have the component
	if !entity.HasComponent("back_accessory") {
		t.Fatal("should have back_accessory after update")
	}

	comp, _ := entity.GetComponent("back_accessory")
	ba, ok := comp.(*BackAccessoryComponent)
	if !ok {
		t.Fatal("component should be *BackAccessoryComponent")
	}
	if ba.Genre != "fantasy" {
		t.Errorf("genre should be fantasy, got %q", ba.Genre)
	}
	if ba.Role == "" {
		t.Error("role should not be empty")
	}
}

func TestBackAccessorySystem_SkipsNonSprite(t *testing.T) {
	world := NewWorld()
	sys := NewBackAccessorySystem(world, 42)

	entity := NewEntity(2)
	// No sprite component
	entity.AddComponent(&HealthComponent{Current: 100, Max: 100})

	sys.Update([]*Entity{entity}, 3.0)

	if entity.HasComponent("back_accessory") {
		t.Fatal("should not assign back_accessory to entity without sprite")
	}
}

func TestBackAccessorySystem_SkipsExisting(t *testing.T) {
	world := NewWorld()
	sys := NewBackAccessorySystem(world, 42)

	entity := NewEntity(3)
	entity.AddComponent(NewSpriteComponent(32, 32, nil))
	entity.AddComponent(&HealthComponent{Current: 100, Max: 100})
	entity.AddComponent(&AIComponent{})

	// Pre-assign
	existing := &BackAccessoryComponent{AccessoryType: 3, Genre: "fantasy", Role: "warrior"}
	entity.AddComponent(existing)

	sys.Update([]*Entity{entity}, 3.0)

	comp, _ := entity.GetComponent("back_accessory")
	ba := comp.(*BackAccessoryComponent)
	// Should be unchanged
	if ba.AccessoryType != 3 || ba.Genre != "fantasy" {
		t.Error("should not modify existing component with same genre")
	}
}

func TestBackAccessorySystem_ReassignsOnGenreChange(t *testing.T) {
	world := NewWorld()
	sys := NewBackAccessorySystem(world, 42)

	entity := NewEntity(4)
	entity.AddComponent(NewSpriteComponent(32, 32, nil))
	entity.AddComponent(&HealthComponent{Current: 100, Max: 100})
	entity.AddComponent(&AIComponent{})

	// Pre-assign with old genre
	existing := &BackAccessoryComponent{AccessoryType: 1, Genre: "fantasy", Role: "warrior"}
	entity.AddComponent(existing)

	// Change genre
	sys.SetGenre("horror")
	sys.Update([]*Entity{entity}, 3.0)

	comp, _ := entity.GetComponent("back_accessory")
	ba := comp.(*BackAccessoryComponent)
	if ba.Genre != "horror" {
		t.Errorf("genre should be horror after reassignment, got %q", ba.Genre)
	}
}

func TestBackAccessorySystem_ScanInterval(t *testing.T) {
	world := NewWorld()
	sys := NewBackAccessorySystem(world, 42)

	entity := NewEntity(5)
	entity.AddComponent(NewSpriteComponent(32, 32, nil))
	entity.AddComponent(&HealthComponent{Current: 100, Max: 100})
	entity.AddComponent(&AIComponent{})

	// Update with delta less than scan interval
	sys.Update([]*Entity{entity}, 0.5)
	if entity.HasComponent("back_accessory") {
		t.Fatal("should not assign before scan interval elapses")
	}

	// Accumulate past interval
	sys.Update([]*Entity{entity}, 2.0)
	if !entity.HasComponent("back_accessory") {
		t.Fatal("should assign after scan interval elapses")
	}
}

func TestBackAccessorySystem_NilEntity(t *testing.T) {
	world := NewWorld()
	sys := NewBackAccessorySystem(world, 42)

	// Should not panic on nil entity
	sys.Update([]*Entity{nil}, 3.0)
}

func TestBackAccessorySystem_PlayerEntity(t *testing.T) {
	world := NewWorld()
	sys := NewBackAccessorySystem(world, 42)

	entity := NewEntity(6)
	entity.AddComponent(NewSpriteComponent(32, 32, nil))
	entity.AddComponent(&StubInput{})

	sys.Update([]*Entity{entity}, 3.0)

	if !entity.HasComponent("back_accessory") {
		t.Fatal("player entity should receive back_accessory")
	}
}
