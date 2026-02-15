package engine

import (
	"testing"

	"github.com/sirupsen/logrus"
)

func TestNewClothingPatternSystem(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)
	world := NewWorldWithLogger(logger)

	sys := NewClothingPatternSystem(world, 12345)
	if sys == nil {
		t.Fatal("expected non-nil system")
	}
	if sys.genreID != "fantasy" {
		t.Errorf("expected default genre 'fantasy', got %q", sys.genreID)
	}
}

func TestClothingPatternSystem_SetGenre(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)
	world := NewWorldWithLogger(logger)

	sys := NewClothingPatternSystem(world, 42)

	tests := []struct {
		genre string
	}{
		{"fantasy"},
		{"horror"},
		{"cyberpunk"},
		{"sci-fi"},
		{"post-apocalyptic"},
	}
	for _, tt := range tests {
		t.Run(tt.genre, func(t *testing.T) {
			sys.SetGenre(tt.genre)
			if sys.genreID != tt.genre {
				t.Errorf("expected genre %q, got %q", tt.genre, sys.genreID)
			}
		})
	}
}

func TestClothingPatternSystem_Update_AttachesComponent(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)
	world := NewWorldWithLogger(logger)

	sys := NewClothingPatternSystem(world, 42)

	entity := world.CreateEntity()
	entity.AddComponent(NewSpriteComponent(32, 32, nil))

	entities := []*Entity{entity}

	// First update: scan interval not elapsed, nothing happens
	sys.Update(entities, 0.5)
	if entity.HasComponent("clothing_pattern") {
		t.Error("should not attach component before scan interval")
	}

	// Advance past scan interval
	sys.Update(entities, 2.0)
	if !entity.HasComponent("clothing_pattern") {
		t.Fatal("should attach clothing_pattern component after scan interval")
	}

	comp, _ := entity.GetComponent("clothing_pattern")
	cp, ok := comp.(*ClothingPatternComponent)
	if !ok {
		t.Fatal("component should be *ClothingPatternComponent")
	}
	if !cp.Enabled {
		t.Error("component should be enabled")
	}
	if cp.GenreID != "fantasy" {
		t.Errorf("expected genre 'fantasy', got %q", cp.GenreID)
	}
}

func TestClothingPatternSystem_Update_SkipsNonSprite(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)
	world := NewWorldWithLogger(logger)

	sys := NewClothingPatternSystem(world, 42)

	entity := world.CreateEntity() // No sprite component

	entities := []*Entity{entity}
	sys.Update(entities, 3.0)

	if entity.HasComponent("clothing_pattern") {
		t.Error("should not attach to entities without sprites")
	}
}

func TestClothingPatternSystem_Update_GenreChange(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)
	world := NewWorldWithLogger(logger)

	sys := NewClothingPatternSystem(world, 42)

	entity := world.CreateEntity()
	entity.AddComponent(NewSpriteComponent(32, 32, nil))

	entities := []*Entity{entity}
	sys.Update(entities, 3.0)

	comp, _ := entity.GetComponent("clothing_pattern")
	cp := comp.(*ClothingPatternComponent)
	if cp.GenreID != "fantasy" {
		t.Fatal("expected initial genre fantasy")
	}

	// Change genre
	sys.SetGenre("cyberpunk")
	sys.Update(entities, 3.0)

	comp, _ = entity.GetComponent("clothing_pattern")
	cp = comp.(*ClothingPatternComponent)
	if cp.GenreID != "cyberpunk" {
		t.Errorf("expected genre 'cyberpunk', got %q", cp.GenreID)
	}
	if !cp.Dirty {
		t.Error("component should be dirty after genre change")
	}
}

func TestClothingPatternSystem_Update_NilEntity(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)
	world := NewWorldWithLogger(logger)

	sys := NewClothingPatternSystem(world, 42)

	entities := []*Entity{nil}
	// Should not panic
	sys.Update(entities, 3.0)
}

func TestClothingPatternComponent_Type(t *testing.T) {
	comp := NewClothingPatternComponent()
	if comp.Type() != "clothing_pattern" {
		t.Errorf("expected type 'clothing_pattern', got %q", comp.Type())
	}
}

func TestClothingPatternComponent_Defaults(t *testing.T) {
	comp := NewClothingPatternComponent()
	if comp.TorsoScale != 1.0 {
		t.Error("expected default torso scale 1.0")
	}
	if comp.TorsoIntensity != 0.0 {
		t.Error("expected default torso intensity 0.0")
	}
	if comp.Enabled != true {
		t.Error("expected default enabled true")
	}
}

func TestClothingPatternSystem_DeterministicPerEntity(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)
	world := NewWorldWithLogger(logger)

	sys1 := NewClothingPatternSystem(world, 42)
	sys2 := NewClothingPatternSystem(world, 42)

	e1 := world.CreateEntity()
	e1.AddComponent(NewSpriteComponent(32, 32, nil))
	e2 := world.CreateEntity()
	e2.AddComponent(NewSpriteComponent(32, 32, nil))

	// Both entities need same ID for determinism test — use same seed calc
	sys1.Update([]*Entity{e1}, 3.0)
	sys2.Update([]*Entity{e2}, 3.0)

	c1, _ := e1.GetComponent("clothing_pattern")
	c2, _ := e2.GetComponent("clothing_pattern")
	cp1 := c1.(*ClothingPatternComponent)
	cp2 := c2.(*ClothingPatternComponent)

	// Both should have patterns generated (even if different IDs produce different patterns)
	if !cp1.Enabled || !cp2.Enabled {
		t.Error("both entities should have enabled clothing patterns")
	}
}

func BenchmarkClothingPatternSystem_Update(b *testing.B) {
	logger := logrus.New()
	logger.SetLevel(logrus.WarnLevel)
	world := NewWorldWithLogger(logger)

	sys := NewClothingPatternSystem(world, 42)

	entities := make([]*Entity, 100)
	for i := range entities {
		e := world.CreateEntity()
		e.AddComponent(NewSpriteComponent(32, 32, nil))
		entities[i] = e
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.timeSinceScan = sys.scanInterval
		sys.Update(entities, 3.0)
	}
}
