package engine

import (
	"math"
	"testing"
)

func TestCreatureEyeGlowComponentType(t *testing.T) {
	comp := NewCreatureEyeGlowComponent()
	if comp.Type() != "creature_eye_glow" {
		t.Errorf("expected type 'creature_eye_glow', got %q", comp.Type())
	}
}

func TestCreatureEyeGlowComponentDefaults(t *testing.T) {
	comp := NewCreatureEyeGlowComponent()
	if comp.Enabled {
		t.Error("expected Enabled=false by default")
	}
	if comp.BaseIntensity != 0 {
		t.Errorf("expected BaseIntensity=0, got %f", comp.BaseIntensity)
	}
	if comp.GlowRadius != 1.0 {
		t.Errorf("expected GlowRadius=1.0, got %f", comp.GlowRadius)
	}
}

func TestNewCreatureEyeGlowSystem(t *testing.T) {
	world := NewWorld()
	sys := NewCreatureEyeGlowSystem(world, 42)
	if sys == nil {
		t.Fatal("expected non-nil system")
	}
	if sys.genreID != "fantasy" {
		t.Errorf("expected default genre 'fantasy', got %q", sys.genreID)
	}
}

func TestCreatureEyeGlowSystemSetGenre(t *testing.T) {
	world := NewWorld()
	sys := NewCreatureEyeGlowSystem(world, 42)
	sys.SetGenre("horror")
	if sys.genreID != "horror" {
		t.Errorf("expected genre 'horror', got %q", sys.genreID)
	}
}

func TestCreatureEyeGlowGenreColors(t *testing.T) {
	tests := []struct {
		name  string
		genre string
		wantR float64
		wantG float64
		wantB float64
	}{
		{"fantasy golden", "fantasy", 0.95, 0.75, 0.20},
		{"horror red", "horror", 0.95, 0.15, 0.10},
		{"scifi cyan", "scifi", 0.20, 0.85, 0.95},
		{"cyberpunk magenta", "cyberpunk", 0.90, 0.20, 0.85},
		{"postapoc green", "postapoc", 0.40, 0.85, 0.20},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			world := NewWorld()
			sys := NewCreatureEyeGlowSystem(world, 12345)
			sys.SetGenre(tt.genre)

			entity := world.CreateEntity()
			entity.AddComponent(&AIComponent{DetectionRange: 250})
			entity.AddComponent(&StubSprite{Visible: true})
			entity.AddComponent(&HealthComponent{Current: 200, Max: 200})
			entity.AddComponent(&FactionComponent{FactionID: "monster_faction", Reputation: -60})

			entities := []*Entity{entity}
			sys.Update(entities, 2.0) // Exceed genre check interval

			comp, ok := entity.GetComponent("creature_eye_glow")
			if !ok {
				t.Fatal("expected creature_eye_glow component")
			}
			gc := comp.(*CreatureEyeGlowComponent)

			if !gc.Enabled {
				t.Error("expected glow to be enabled for hostile creature")
			}

			// Color should be close to genre palette (±0.04 variation)
			if math.Abs(gc.GlowR-tt.wantR) > 0.05 {
				t.Errorf("GlowR=%.2f, want ~%.2f", gc.GlowR, tt.wantR)
			}
			if math.Abs(gc.GlowG-tt.wantG) > 0.05 {
				t.Errorf("GlowG=%.2f, want ~%.2f", gc.GlowG, tt.wantG)
			}
			if math.Abs(gc.GlowB-tt.wantB) > 0.05 {
				t.Errorf("GlowB=%.2f, want ~%.2f", gc.GlowB, tt.wantB)
			}
		})
	}
}

func TestCreatureEyeGlowBossHigherIntensity(t *testing.T) {
	world := NewWorld()

	// Boss entity
	bossEntity := world.CreateEntity()
	bossEntity.AddComponent(&AIComponent{DetectionRange: 100})
	bossEntity.AddComponent(&StubSprite{Visible: true})
	bossEntity.AddComponent(&HealthComponent{Current: 500, Max: 500})
	bossEntity.AddComponent(&FactionComponent{FactionID: "boss_faction", Reputation: -100})

	// Regular enemy
	regularEntity := world.CreateEntity()
	regularEntity.AddComponent(&AIComponent{DetectionRange: 100})
	regularEntity.AddComponent(&StubSprite{Visible: true})
	regularEntity.AddComponent(&HealthComponent{Current: 80, Max: 80})
	regularEntity.AddComponent(&FactionComponent{FactionID: "enemy_faction", Reputation: -60})

	sys := NewCreatureEyeGlowSystem(world, 42)
	sys.SetGenre("fantasy")
	entities := []*Entity{bossEntity, regularEntity}
	sys.Update(entities, 2.0)

	bossComp, _ := bossEntity.GetComponent("creature_eye_glow")
	bossGlow := bossComp.(*CreatureEyeGlowComponent)

	regComp, _ := regularEntity.GetComponent("creature_eye_glow")
	regGlow := regComp.(*CreatureEyeGlowComponent)

	if !bossGlow.Enabled || !regGlow.Enabled {
		t.Fatal("expected both glows enabled")
	}

	if bossGlow.BaseIntensity <= regGlow.BaseIntensity {
		t.Errorf("boss intensity %.2f should exceed regular %.2f", bossGlow.BaseIntensity, regGlow.BaseIntensity)
	}

	if bossGlow.GlowRadius <= regGlow.GlowRadius {
		t.Errorf("boss glow radius %.2f should exceed regular %.2f", bossGlow.GlowRadius, regGlow.GlowRadius)
	}
}

func TestCreatureEyeGlowNeutralNoGlow(t *testing.T) {
	world := NewWorld()
	entity := world.CreateEntity()
	entity.AddComponent(&AIComponent{DetectionRange: 100})
	entity.AddComponent(&StubSprite{Visible: true})
	entity.AddComponent(&HealthComponent{Current: 100, Max: 100})
	entity.AddComponent(&FactionComponent{FactionID: "neutral_faction", Reputation: 50})

	sys := NewCreatureEyeGlowSystem(world, 42)
	sys.SetGenre("fantasy")
	sys.Update([]*Entity{entity}, 2.0)

	comp, ok := entity.GetComponent("creature_eye_glow")
	if !ok {
		t.Fatal("expected component to be created")
	}
	gc := comp.(*CreatureEyeGlowComponent)
	if gc.Enabled {
		t.Error("neutral creatures should not have eye glow")
	}
}

func TestCreatureEyeGlowPlayerExcluded(t *testing.T) {
	world := NewWorld()
	entity := world.CreateEntity()
	entity.AddComponent(&AIComponent{DetectionRange: 100})
	entity.AddComponent(&StubSprite{Visible: true})
	entity.AddComponent(NewStubInput())
	entity.AddComponent(&HealthComponent{Current: 200, Max: 200})

	sys := NewCreatureEyeGlowSystem(world, 42)
	sys.SetGenre("horror")
	sys.Update([]*Entity{entity}, 2.0)

	_, ok := entity.GetComponent("creature_eye_glow")
	if ok {
		t.Error("player entities should not receive eye glow")
	}
}

func TestCreatureEyeGlowPulseAnimation(t *testing.T) {
	world := NewWorld()
	entity := world.CreateEntity()
	entity.AddComponent(&AIComponent{DetectionRange: 300})
	entity.AddComponent(&StubSprite{Visible: true})
	entity.AddComponent(&HealthComponent{Current: 300, Max: 300})
	entity.AddComponent(&FactionComponent{FactionID: "boss_faction", Reputation: -100})

	sys := NewCreatureEyeGlowSystem(world, 42)
	sys.SetGenre("horror")
	sys.Update([]*Entity{entity}, 2.0) // Initialize glow

	comp, _ := entity.GetComponent("creature_eye_glow")
	gc := comp.(*CreatureEyeGlowComponent)
	initialIntensity := gc.CurrentIntensity

	// Run several pulse updates
	for i := 0; i < 30; i++ {
		sys.Update([]*Entity{entity}, 0.016) // ~60fps frames
	}

	// Intensity should have changed due to pulse
	if gc.CurrentIntensity == initialIntensity && gc.PulseAmplitude > 0 {
		t.Error("expected pulse to modulate intensity over time")
	}
}

func TestCreatureEyeGlowThrottling(t *testing.T) {
	world := NewWorld()
	sys := NewCreatureEyeGlowSystem(world, 42)
	sys.SetGenre("fantasy")

	entity := world.CreateEntity()
	entity.AddComponent(&AIComponent{DetectionRange: 200})
	entity.AddComponent(&StubSprite{Visible: true})
	entity.AddComponent(&HealthComponent{Current: 100, Max: 100})
	entity.AddComponent(&FactionComponent{FactionID: "enemy_faction", Reputation: -60})

	entities := []*Entity{entity}

	// First update with enough time triggers assignment
	sys.Update(entities, 2.0)
	_, ok := entity.GetComponent("creature_eye_glow")
	if !ok {
		t.Fatal("expected component after first update")
	}

	// Change genre and call with small delta - should NOT reassign yet
	sys.SetGenre("horror")
	comp, _ := entity.GetComponent("creature_eye_glow")
	gc := comp.(*CreatureEyeGlowComponent)
	oldR := gc.GlowR

	sys.Update(entities, 0.1) // Below 1.0s threshold
	if gc.GlowR != oldR {
		t.Error("expected throttled update to not change glow color")
	}
}

func TestClampEyeGlow(t *testing.T) {
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
		got := clampEyeGlow(tt.input)
		if got != tt.want {
			t.Errorf("clampEyeGlow(%f) = %f, want %f", tt.input, got, tt.want)
		}
	}
}

func TestCreatureEyeGlowUnknownGenreFallback(t *testing.T) {
	world := NewWorld()
	sys := NewCreatureEyeGlowSystem(world, 42)
	sys.SetGenre("unknown_genre")

	entity := world.CreateEntity()
	entity.AddComponent(&AIComponent{DetectionRange: 300})
	entity.AddComponent(&StubSprite{Visible: true})
	entity.AddComponent(&HealthComponent{Current: 200, Max: 200})
	entity.AddComponent(&FactionComponent{FactionID: "boss_faction", Reputation: -100})

	sys.Update([]*Entity{entity}, 2.0)

	comp, ok := entity.GetComponent("creature_eye_glow")
	if !ok {
		t.Fatal("expected component")
	}
	gc := comp.(*CreatureEyeGlowComponent)

	// Should fall back to fantasy palette
	if math.Abs(gc.GlowR-0.95) > 0.05 {
		t.Errorf("expected fantasy fallback R~0.95, got %.2f", gc.GlowR)
	}
}

func BenchmarkCreatureEyeGlowUpdate(b *testing.B) {
	world := NewWorld()
	sys := NewCreatureEyeGlowSystem(world, 42)
	sys.SetGenre("fantasy")

	entities := make([]*Entity, 200)
	for i := range entities {
		e := world.CreateEntity()
		e.AddComponent(&AIComponent{DetectionRange: 200})
		e.AddComponent(&StubSprite{Visible: true})
		e.AddComponent(&HealthComponent{Current: 100, Max: float64(50 + i*2)})
		e.AddComponent(&FactionComponent{FactionID: "enemy_faction", Reputation: -60})
		entities[i] = e
	}

	// Initialize glow components
	sys.Update(entities, 2.0)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.Update(entities, 0.016)
	}
}
