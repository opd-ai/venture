// Package engine provides tests for CreatureEyePatternSystem.
package engine

import (
	"math/rand"
	"testing"
)

func TestNewCreatureEyePatternSystem(t *testing.T) {
	world := NewWorld()
	sys := NewCreatureEyePatternSystem(world, 12345)

	if sys == nil {
		t.Fatal("Expected non-nil system")
	}
	if sys.world != world {
		t.Error("Expected world to be set")
	}
	if sys.rng == nil {
		t.Error("Expected rng to be initialized")
	}
	if len(sys.patterns) == 0 {
		t.Error("Expected patterns to be populated")
	}
	if len(sys.genreColors) == 0 {
		t.Error("Expected genreColors to be populated")
	}
}

func TestCreatureEyePatternComponent(t *testing.T) {
	comp := NewCreatureEyePatternComponent()

	if comp.Type() != "creature_eye_pattern" {
		t.Errorf("Expected type 'creature_eye_pattern', got %s", comp.Type())
	}
	if comp.EyeCount != 2 {
		t.Errorf("Expected default EyeCount 2, got %d", comp.EyeCount)
	}
	if len(comp.EyePositions) != 4 {
		t.Errorf("Expected 4 position values (2 eyes), got %d", len(comp.EyePositions))
	}
	if !comp.Dirty {
		t.Error("Expected Dirty to be true initially")
	}
}

func TestCreatureEyePatternSystemSetGenre(t *testing.T) {
	world := NewWorld()
	sys := NewCreatureEyePatternSystem(world, 12345)

	sys.SetGenre("horror")
	if sys.genreID != "horror" {
		t.Errorf("Expected genreID 'horror', got %s", sys.genreID)
	}

	sys.SetGenre("scifi")
	if sys.genreID != "scifi" {
		t.Errorf("Expected genreID 'scifi', got %s", sys.genreID)
	}
}

func TestCreatureEyePatternSystemUpdate(t *testing.T) {
	world := NewWorld()
	sys := NewCreatureEyePatternSystem(world, 12345)
	sys.SetGenre("fantasy")

	// Create test entities with different creature forms
	tests := []struct {
		name         string
		form         CreatureForm
		wantPattern  string
		wantEyeCount int
	}{
		{"spider", FormArachnid, "arachnid_8", 8},
		{"snake", FormSerpentine, "serpent_slit", 2},
		{"beetle", FormInsect, "insect_compound", 2},
		{"dragon", FormFlying, "flying_raptor", 2},
		{"slime", FormBlob, "blob_nucleus", 1},
		{"robot", FormMechanical, "mechanical_sensors", 3},
		{"skeleton", FormUndead, "undead_sockets", 2},
		{"wolf", FormQuadruped, "quadruped_2", 2},
	}

	entities := make([]*Entity, len(tests))
	for i, tt := range tests {
		entity := world.CreateEntity()
		entity.AddComponent(&CreatureVisualComponent{
			Form:      tt.form,
			SizeClass: "medium",
		})
		entity.AddComponent(&AnimationComponent{
			SpriteID: "test_sprite",
		})
		entities[i] = entity
	}

	// Run update with enough delta time to trigger
	sys.Update(entities, 1.0)

	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entity := entities[i]
			comp, ok := entity.GetComponent("creature_eye_pattern")
			if !ok {
				t.Fatalf("Expected creature_eye_pattern component")
			}
			eyeComp, ok := comp.(*CreatureEyePatternComponent)
			if !ok {
				t.Fatalf("Expected *CreatureEyePatternComponent")
			}

			if eyeComp.EyePattern != tt.wantPattern {
				t.Errorf("Expected pattern %s, got %s", tt.wantPattern, eyeComp.EyePattern)
			}

			// Multi-limbed has variable eye count, so skip exact check
			if tt.form != FormMultiLimbed && eyeComp.EyeCount != tt.wantEyeCount {
				t.Errorf("Expected eye count %d, got %d", tt.wantEyeCount, eyeComp.EyeCount)
			}
		})
	}
}

func TestCreatureEyePatternSystemSkipsHumanoids(t *testing.T) {
	world := NewWorld()
	sys := NewCreatureEyePatternSystem(world, 12345)

	entity := world.CreateEntity()
	entity.AddComponent(&CreatureVisualComponent{
		Form:      FormHumanoid,
		SizeClass: "medium",
	})

	entities := []*Entity{entity}
	sys.Update(entities, 1.0)

	// Humanoids should NOT get eye pattern component
	if entity.HasComponent("creature_eye_pattern") {
		t.Error("Humanoids should not receive creature_eye_pattern component")
	}
}

func TestCreatureEyePatternSystemGenreColors(t *testing.T) {
	world := NewWorld()

	genres := []string{"fantasy", "horror", "scifi", "cyberpunk", "postapoc"}

	for _, genre := range genres {
		t.Run(genre, func(t *testing.T) {
			sys := NewCreatureEyePatternSystem(world, 12345)
			sys.SetGenre(genre)

			entity := world.CreateEntity()
			entity.AddComponent(&CreatureVisualComponent{
				Form:      FormQuadruped,
				SizeClass: "medium",
			})
			entity.AddComponent(&AnimationComponent{})

			sys.Update([]*Entity{entity}, 1.0)

			comp, ok := entity.GetComponent("creature_eye_pattern")
			if !ok {
				t.Fatal("Expected creature_eye_pattern component")
			}
			eyeComp := comp.(*CreatureEyePatternComponent)

			// Check that colors are within valid range
			if eyeComp.EyeR < 0 || eyeComp.EyeR > 1 {
				t.Errorf("EyeR out of range: %f", eyeComp.EyeR)
			}
			if eyeComp.EyeG < 0 || eyeComp.EyeG > 1 {
				t.Errorf("EyeG out of range: %f", eyeComp.EyeG)
			}
			if eyeComp.EyeB < 0 || eyeComp.EyeB > 1 {
				t.Errorf("EyeB out of range: %f", eyeComp.EyeB)
			}
		})
	}
}

func TestCreatureEyePatternSystemMultiLimbedRandomization(t *testing.T) {
	world := NewWorld()

	// Run multiple times to verify randomization
	eyeCounts := make(map[int]int)

	for i := 0; i < 50; i++ {
		sys := NewCreatureEyePatternSystem(world, int64(i*1000))
		sys.SetGenre("horror")

		entity := world.CreateEntity()
		entity.AddComponent(&CreatureVisualComponent{
			Form:      FormMultiLimbed,
			SizeClass: "large",
		})
		entity.AddComponent(&AnimationComponent{})

		sys.Update([]*Entity{entity}, 1.0)

		comp, _ := entity.GetComponent("creature_eye_pattern")
		eyeComp := comp.(*CreatureEyePatternComponent)

		eyeCounts[eyeComp.EyeCount]++

		// Verify asymmetric flag is set
		if !eyeComp.Asymmetric {
			t.Error("Multi-limbed should have Asymmetric=true")
		}

		// Verify eye count is in valid range (3-6)
		if eyeComp.EyeCount < 3 || eyeComp.EyeCount > 6 {
			t.Errorf("Multi-limbed eye count should be 3-6, got %d", eyeComp.EyeCount)
		}
	}

	// Verify we got some variety in eye counts
	if len(eyeCounts) < 2 {
		t.Errorf("Expected variety in eye counts, only got %v", eyeCounts)
	}
}

func TestCreatureEyePatternSystemArachnidEyes(t *testing.T) {
	world := NewWorld()
	sys := NewCreatureEyePatternSystem(world, 12345)

	entity := world.CreateEntity()
	entity.AddComponent(&CreatureVisualComponent{
		Form:      FormArachnid,
		SizeClass: "medium",
	})
	entity.AddComponent(&AnimationComponent{})

	sys.Update([]*Entity{entity}, 1.0)

	comp, _ := entity.GetComponent("creature_eye_pattern")
	eyeComp := comp.(*CreatureEyePatternComponent)

	// Arachnids should have 8 eyes
	if eyeComp.EyeCount != 8 {
		t.Errorf("Arachnid should have 8 eyes, got %d", eyeComp.EyeCount)
	}

	// Should have 16 position values (8 eyes * 2 coordinates)
	if len(eyeComp.EyePositions) != 16 {
		t.Errorf("Expected 16 position values, got %d", len(eyeComp.EyePositions))
	}

	// Should have 8 size values
	if len(eyeComp.EyeSizes) != 8 {
		t.Errorf("Expected 8 size values, got %d", len(eyeComp.EyeSizes))
	}

	// Spider eyes should have no pupils
	if eyeComp.PupilStyle != "none" {
		t.Errorf("Spider eyes should have no pupils, got %s", eyeComp.PupilStyle)
	}
}

func TestCreatureEyePatternSystemSerpentSlitPupils(t *testing.T) {
	world := NewWorld()
	sys := NewCreatureEyePatternSystem(world, 12345)

	entity := world.CreateEntity()
	entity.AddComponent(&CreatureVisualComponent{
		Form:      FormSerpentine,
		SizeClass: "medium",
	})
	entity.AddComponent(&AnimationComponent{})

	sys.Update([]*Entity{entity}, 1.0)

	comp, _ := entity.GetComponent("creature_eye_pattern")
	eyeComp := comp.(*CreatureEyePatternComponent)

	// Serpents should have vertical slit pupils
	if eyeComp.PupilStyle != "slit_vertical" {
		t.Errorf("Serpent eyes should have slit_vertical pupils, got %s", eyeComp.PupilStyle)
	}

	// Should have yellowish eyes
	if eyeComp.EyeR < 0.5 {
		t.Errorf("Serpent eyes should be warm-colored (yellow/gold), R=%f", eyeComp.EyeR)
	}
}

func TestCreatureEyePatternSystemMechanicalGlow(t *testing.T) {
	world := NewWorld()
	sys := NewCreatureEyePatternSystem(world, 12345)

	entity := world.CreateEntity()
	entity.AddComponent(&CreatureVisualComponent{
		Form:      FormMechanical,
		SizeClass: "medium",
	})
	entity.AddComponent(&AnimationComponent{})

	sys.Update([]*Entity{entity}, 1.0)

	comp, _ := entity.GetComponent("creature_eye_pattern")
	eyeComp := comp.(*CreatureEyePatternComponent)

	// Mechanical entities should have sensor glow
	if eyeComp.GlowIntensity < 0.3 {
		t.Errorf("Mechanical eyes should have glow, got intensity %f", eyeComp.GlowIntensity)
	}
}

func TestClampEyePattern(t *testing.T) {
	tests := []struct {
		v, min, max, want float64
	}{
		{0.5, 0.0, 1.0, 0.5},
		{-0.5, 0.0, 1.0, 0.0},
		{1.5, 0.0, 1.0, 1.0},
		{0.3, 0.2, 0.8, 0.3},
		{0.1, 0.2, 0.8, 0.2},
		{0.9, 0.2, 0.8, 0.8},
	}

	for _, tt := range tests {
		got := clampEyePattern(tt.v, tt.min, tt.max)
		if got != tt.want {
			t.Errorf("clampEyePattern(%f, %f, %f) = %f, want %f", tt.v, tt.min, tt.max, got, tt.want)
		}
	}
}

func BenchmarkCreatureEyePatternSystemUpdate(b *testing.B) {
	world := NewWorld()
	sys := NewCreatureEyePatternSystem(world, 12345)
	sys.SetGenre("fantasy")

	// Create 100 creatures of various types
	forms := []CreatureForm{
		FormQuadruped, FormArachnid, FormSerpentine, FormInsect,
		FormFlying, FormBlob, FormMechanical, FormUndead, FormMultiLimbed,
	}

	entities := make([]*Entity, 100)
	rng := rand.New(rand.NewSource(12345))
	for i := range entities {
		entity := world.CreateEntity()
		entity.AddComponent(&CreatureVisualComponent{
			Form:      forms[rng.Intn(len(forms))],
			SizeClass: "medium",
		})
		entity.AddComponent(&AnimationComponent{})
		entities[i] = entity
	}

	// Initial population
	sys.Update(entities, 1.0)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.Update(entities, 0.016)
	}
}
