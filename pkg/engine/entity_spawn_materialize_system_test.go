package engine

import (
	"math"
	"testing"
)

func TestSpawnMaterializeComponentType(t *testing.T) {
	comp := &SpawnMaterializeComponent{}
	if comp.Type() != "spawn_materialize" {
		t.Errorf("expected type 'spawn_materialize', got %q", comp.Type())
	}
}

func TestNewEntitySpawnMaterializeSystem(t *testing.T) {
	world := NewWorld()
	sys := NewEntitySpawnMaterializeSystem(world, 42)
	if sys == nil {
		t.Fatal("expected non-nil system")
	}
	if sys.genreID != "fantasy" {
		t.Errorf("expected default genre 'fantasy', got %q", sys.genreID)
	}
}

func TestEntitySpawnMaterializeSetGenre(t *testing.T) {
	world := NewWorld()
	sys := NewEntitySpawnMaterializeSystem(world, 42)
	sys.SetGenre("horror")
	if sys.genreID != "horror" {
		t.Errorf("expected genre 'horror', got %q", sys.genreID)
	}
}

func TestEntitySpawnMaterializeGenrePresets(t *testing.T) {
	tests := []struct {
		name  string
		genre string
		wantR float64
		wantG float64
		wantB float64
	}{
		{"fantasy golden", "fantasy", 0.95, 0.80, 0.30},
		{"horror red", "horror", 0.70, 0.10, 0.10},
		{"scifi cyan", "scifi", 0.20, 0.85, 0.95},
		{"cyberpunk magenta", "cyberpunk", 0.90, 0.20, 0.80},
		{"postapoc brown", "postapoc", 0.65, 0.50, 0.30},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			world := NewWorld()
			sys := NewEntitySpawnMaterializeSystem(world, 12345)
			sys.SetGenre(tt.genre)

			entity := world.CreateEntity()
			entity.AddComponent(&StubSprite{Visible: true})

			sys.Update([]*Entity{entity}, 0.5)

			comp, ok := entity.GetComponent("spawn_materialize")
			if !ok {
				t.Fatal("expected spawn_materialize component")
			}
			mc := comp.(*SpawnMaterializeComponent)

			// Color should be close to genre preset (±0.03 variation)
			if math.Abs(mc.ParticleR-tt.wantR) > 0.04 {
				t.Errorf("ParticleR=%.2f, want ~%.2f", mc.ParticleR, tt.wantR)
			}
			if math.Abs(mc.ParticleG-tt.wantG) > 0.04 {
				t.Errorf("ParticleG=%.2f, want ~%.2f", mc.ParticleG, tt.wantG)
			}
			if math.Abs(mc.ParticleB-tt.wantB) > 0.04 {
				t.Errorf("ParticleB=%.2f, want ~%.2f", mc.ParticleB, tt.wantB)
			}
		})
	}
}

func TestEntitySpawnMaterializeOpacityFadeIn(t *testing.T) {
	world := NewWorld()
	sys := NewEntitySpawnMaterializeSystem(world, 42)
	sys.SetGenre("fantasy")

	entity := world.CreateEntity()
	entity.AddComponent(&StubSprite{Visible: true})

	// First update: detect + start materialization
	sys.Update([]*Entity{entity}, 0.0)

	comp, ok := entity.GetComponent("spawn_materialize")
	if !ok {
		t.Fatal("expected spawn_materialize component")
	}
	mc := comp.(*SpawnMaterializeComponent)

	if mc.Opacity != 0.0 {
		t.Errorf("expected initial opacity 0.0, got %f", mc.Opacity)
	}

	// Animate halfway through
	halfDuration := mc.Duration / 2
	sys.Update([]*Entity{entity}, halfDuration)

	if mc.Opacity <= 0.0 || mc.Opacity >= 1.0 {
		t.Errorf("expected opacity between 0 and 1 at halfway, got %f", mc.Opacity)
	}

	// Animate past completion
	sys.Update([]*Entity{entity}, mc.Duration)

	if mc.Opacity != 1.0 {
		t.Errorf("expected opacity 1.0 after completion, got %f", mc.Opacity)
	}
	if !mc.Complete {
		t.Error("expected Complete=true after full duration")
	}
}

func TestEntitySpawnMaterializeNoSpriteSkipped(t *testing.T) {
	world := NewWorld()
	sys := NewEntitySpawnMaterializeSystem(world, 42)
	sys.SetGenre("fantasy")

	entity := world.CreateEntity()
	// No sprite component - should not get materialization

	sys.Update([]*Entity{entity}, 0.5)

	_, ok := entity.GetComponent("spawn_materialize")
	if ok {
		t.Error("entity without sprite should not get spawn_materialize")
	}
}

func TestEntitySpawnMaterializeExistingComponentSkipped(t *testing.T) {
	world := NewWorld()
	sys := NewEntitySpawnMaterializeSystem(world, 42)
	sys.SetGenre("scifi")

	entity := world.CreateEntity()
	entity.AddComponent(&StubSprite{Visible: true})

	// Pre-attach component (simulating deserialized entity)
	existing := &SpawnMaterializeComponent{
		Elapsed:  0.2,
		Duration: 1.0,
		Opacity:  0.5,
	}
	entity.AddComponent(existing)

	sys.Update([]*Entity{entity}, 0.5)

	comp, _ := entity.GetComponent("spawn_materialize")
	mc := comp.(*SpawnMaterializeComponent)

	// Should be the same pre-existing component (opacity should advance, not reset)
	if mc.Duration != 1.0 {
		t.Errorf("expected preserved duration 1.0, got %f", mc.Duration)
	}
}

func TestEntitySpawnMaterializeUnknownGenreFallback(t *testing.T) {
	world := NewWorld()
	sys := NewEntitySpawnMaterializeSystem(world, 42)
	sys.SetGenre("unknown_genre")

	entity := world.CreateEntity()
	entity.AddComponent(&StubSprite{Visible: true})

	sys.Update([]*Entity{entity}, 0.5)

	comp, ok := entity.GetComponent("spawn_materialize")
	if !ok {
		t.Fatal("expected component")
	}
	mc := comp.(*SpawnMaterializeComponent)

	// Should fall back to fantasy preset
	if math.Abs(mc.ParticleR-0.95) > 0.04 {
		t.Errorf("expected fantasy fallback R~0.95, got %.2f", mc.ParticleR)
	}
}

func TestEntitySpawnMaterializeEaseOutCurve(t *testing.T) {
	// Verify the ease-out curve: opacity should be >0.5 at t=0.5
	// because ease-out front-loads the animation
	world := NewWorld()
	sys := NewEntitySpawnMaterializeSystem(world, 99)
	sys.SetGenre("fantasy")

	entity := world.CreateEntity()
	entity.AddComponent(&StubSprite{Visible: true})

	sys.Update([]*Entity{entity}, 0.0) // detect

	comp, _ := entity.GetComponent("spawn_materialize")
	mc := comp.(*SpawnMaterializeComponent)

	// Move to exactly 50% of duration
	sys.Update([]*Entity{entity}, mc.Duration*0.5)

	// Ease-out at t=0.5: 1 - (1-0.5)^2 = 1 - 0.25 = 0.75
	expected := 0.75
	if math.Abs(mc.Opacity-expected) > 0.1 {
		t.Errorf("at t=0.5, opacity=%f, expected ~%f (ease-out)", mc.Opacity, expected)
	}
}

func TestEntitySpawnMaterializeParticleProgress(t *testing.T) {
	world := NewWorld()
	sys := NewEntitySpawnMaterializeSystem(world, 42)
	sys.SetGenre("scifi") // 10 particles

	entity := world.CreateEntity()
	entity.AddComponent(&StubSprite{Visible: true})

	sys.Update([]*Entity{entity}, 0.0) // detect

	comp, _ := entity.GetComponent("spawn_materialize")
	mc := comp.(*SpawnMaterializeComponent)

	// Should have 0 particles emitted initially
	if mc.ParticlesEmitted != 0 {
		t.Errorf("expected 0 particles emitted initially, got %d", mc.ParticlesEmitted)
	}

	// Animate past completion
	sys.Update([]*Entity{entity}, mc.Duration+0.5)

	if mc.ParticlesEmitted != mc.ParticleCount {
		t.Errorf("expected %d particles emitted after completion, got %d", mc.ParticleCount, mc.ParticlesEmitted)
	}
}

func TestEntitySpawnMaterializeScanThrottling(t *testing.T) {
	world := NewWorld()
	sys := NewEntitySpawnMaterializeSystem(world, 42)
	sys.SetGenre("fantasy")

	entity1 := world.CreateEntity()
	entity1.AddComponent(&StubSprite{Visible: true})

	// First update detects entity1
	sys.Update([]*Entity{entity1}, 0.5)
	_, ok1 := entity1.GetComponent("spawn_materialize")
	if !ok1 {
		t.Fatal("entity1 should have spawn_materialize")
	}

	// Add entity2 but call with small delta (below scanInterval=0.25)
	entity2 := world.CreateEntity()
	entity2.AddComponent(&StubSprite{Visible: true})

	sys.Update([]*Entity{entity1, entity2}, 0.01)
	_, ok2 := entity2.GetComponent("spawn_materialize")
	if ok2 {
		t.Error("entity2 should not be detected yet due to throttling")
	}

	// Call with enough delta to trigger scan
	sys.Update([]*Entity{entity1, entity2}, 0.3)
	_, ok2 = entity2.GetComponent("spawn_materialize")
	if !ok2 {
		t.Error("entity2 should be detected after scan interval elapsed")
	}
}

func TestClampMaterialize(t *testing.T) {
	tests := []struct {
		input float64
		want  float64
	}{
		{-0.5, 0.0},
		{0.0, 0.0},
		{0.5, 0.5},
		{1.0, 1.0},
		{1.5, 1.0},
	}
	for _, tt := range tests {
		got := clampMaterialize(tt.input)
		if got != tt.want {
			t.Errorf("clampMaterialize(%f) = %f, want %f", tt.input, got, tt.want)
		}
	}
}

func BenchmarkEntitySpawnMaterializeUpdate(b *testing.B) {
	world := NewWorld()
	sys := NewEntitySpawnMaterializeSystem(world, 42)
	sys.SetGenre("fantasy")

	entities := make([]*Entity, 200)
	for i := range entities {
		e := world.CreateEntity()
		e.AddComponent(&StubSprite{Visible: true})
		entities[i] = e
	}

	// Initialize materialization components
	sys.Update(entities, 0.5)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.Update(entities, 0.016)
	}
}
