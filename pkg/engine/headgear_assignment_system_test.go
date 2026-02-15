package engine

import (
	"testing"
)

func TestHeadgearComponentType(t *testing.T) {
	hg := &HeadgearComponent{
		HeadgearType: 3,
		Genre:        "fantasy",
		Role:         "mage",
	}
	if hg.Type() != "headgear" {
		t.Errorf("HeadgearComponent.Type() = %q, want %q", hg.Type(), "headgear")
	}
}

func TestNewHeadgearAssignmentSystem(t *testing.T) {
	world := NewWorld()
	sys := NewHeadgearAssignmentSystem(world, 42)
	if sys == nil {
		t.Fatal("NewHeadgearAssignmentSystem returned nil")
	}
	if sys.genreID != "fantasy" {
		t.Errorf("default genre = %q, want %q", sys.genreID, "fantasy")
	}
}

func TestHeadgearAssignmentSetGenre(t *testing.T) {
	world := NewWorld()
	sys := NewHeadgearAssignmentSystem(world, 42)
	sys.SetGenre("horror")
	if sys.genreID != "horror" {
		t.Errorf("genre after SetGenre = %q, want %q", sys.genreID, "horror")
	}
}

func TestHeadgearAssignmentUpdateSkipsNonSprite(t *testing.T) {
	world := NewWorld()
	sys := NewHeadgearAssignmentSystem(world, 42)

	entity := NewEntity(1)
	sys.Update([]*Entity{entity}, 3.0) // Force scan by exceeding interval

	_, has := entity.GetComponent("headgear")
	if has {
		t.Error("headgear should not be assigned to entity without sprite")
	}
}

func TestHeadgearAssignmentAssignsToHumanoidWithSprite(t *testing.T) {
	world := NewWorld()
	sys := NewHeadgearAssignmentSystem(world, 42)

	entity := NewEntity(100)
	entity.AddComponent(NewSpriteComponent(32, 32, nil))
	entity.AddComponent(&HealthComponent{Current: 100, Max: 100})

	sys.Update([]*Entity{entity}, 3.0) // Force scan

	comp, has := entity.GetComponent("headgear")
	if !has {
		t.Fatal("headgear not assigned to humanoid entity with sprite and health")
	}
	hg, ok := comp.(*HeadgearComponent)
	if !ok {
		t.Fatal("component is not HeadgearComponent")
	}
	if hg.Genre != "fantasy" {
		t.Errorf("genre = %q, want %q", hg.Genre, "fantasy")
	}
}

func TestHeadgearAssignmentDeterministic(t *testing.T) {
	for i := 0; i < 5; i++ {
		world := NewWorld()
		sys := NewHeadgearAssignmentSystem(world, 42)

		entity := NewEntity(100)
		entity.AddComponent(NewSpriteComponent(32, 32, nil))
		entity.AddComponent(&HealthComponent{Current: 100, Max: 100})
		sys.Update([]*Entity{entity}, 3.0)

		comp, _ := entity.GetComponent("headgear")
		hg := comp.(*HeadgearComponent)
		if i > 0 {
			// All iterations should produce same result
			world2 := NewWorld()
			sys2 := NewHeadgearAssignmentSystem(world2, 42)
			entity2 := NewEntity(100)
			entity2.AddComponent(NewSpriteComponent(32, 32, nil))
			entity2.AddComponent(&HealthComponent{Current: 100, Max: 100})
			sys2.Update([]*Entity{entity2}, 3.0)
			comp2, _ := entity2.GetComponent("headgear")
			hg2 := comp2.(*HeadgearComponent)
			if hg.HeadgearType != hg2.HeadgearType {
				t.Fatalf("non-deterministic: %d vs %d", hg.HeadgearType, hg2.HeadgearType)
			}
		}
	}
}

func TestHeadgearAssignmentRespectsGenreChange(t *testing.T) {
	world := NewWorld()
	sys := NewHeadgearAssignmentSystem(world, 42)

	entity := NewEntity(100)
	entity.AddComponent(NewSpriteComponent(32, 32, nil))
	entity.AddComponent(&HealthComponent{Current: 100, Max: 100})

	sys.Update([]*Entity{entity}, 3.0)
	comp, _ := entity.GetComponent("headgear")
	hg := comp.(*HeadgearComponent)
	initialGenre := hg.Genre

	// Change genre and update again
	sys.SetGenre("cyberpunk")
	sys.timeSinceScan = 0
	sys.Update([]*Entity{entity}, 3.0)

	comp2, _ := entity.GetComponent("headgear")
	hg2 := comp2.(*HeadgearComponent)
	if hg2.Genre == initialGenre {
		t.Error("genre should have changed after SetGenre + Update")
	}
	if hg2.Genre != "cyberpunk" {
		t.Errorf("genre = %q, want cyberpunk", hg2.Genre)
	}
}

func TestHeadgearAssignmentSkipsScanBeforeInterval(t *testing.T) {
	world := NewWorld()
	sys := NewHeadgearAssignmentSystem(world, 42)

	entity := NewEntity(100)
	entity.AddComponent(NewSpriteComponent(32, 32, nil))
	entity.AddComponent(&HealthComponent{Current: 100, Max: 100})

	// Small delta should not trigger scan
	sys.Update([]*Entity{entity}, 0.5)
	_, has := entity.GetComponent("headgear")
	if has {
		t.Error("headgear should not be assigned before scan interval")
	}
}
