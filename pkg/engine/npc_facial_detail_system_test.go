package engine

import (
	"testing"
)

func TestNpcFacialDetailComponent_Type(t *testing.T) {
	comp := NewNpcFacialDetailComponent()
	if comp.Type() != "npc_facial_detail" {
		t.Errorf("expected type 'npc_facial_detail', got %q", comp.Type())
	}
}

func TestNpcFacialDetailComponent_Defaults(t *testing.T) {
	comp := NewNpcFacialDetailComponent()
	if comp.EyeSize != 2.0 {
		t.Errorf("expected default EyeSize 2.0, got %f", comp.EyeSize)
	}
	if comp.MouthSize != 1.0 {
		t.Errorf("expected default MouthSize 1.0, got %f", comp.MouthSize)
	}
	if comp.ExpressionType != "neutral" {
		t.Errorf("expected default expression 'neutral', got %q", comp.ExpressionType)
	}
	if comp.HeadShapeTag != "circle" {
		t.Errorf("expected default head shape 'circle', got %q", comp.HeadShapeTag)
	}
}

func TestNpcFacialDetailSystem_Creation(t *testing.T) {
	world := NewWorld()
	sys := NewNpcFacialDetailSystem(world, 42)
	if sys == nil {
		t.Fatal("expected non-nil system")
	}
	if sys.genreID != "fantasy" {
		t.Errorf("expected default genre 'fantasy', got %q", sys.genreID)
	}
}

func TestNpcFacialDetailSystem_SetGenre(t *testing.T) {
	world := NewWorld()
	sys := NewNpcFacialDetailSystem(world, 42)
	sys.SetGenre("horror")
	if sys.genreID != "horror" {
		t.Errorf("expected genre 'horror', got %q", sys.genreID)
	}
}

func TestNpcFacialDetailSystem_GenrePalettes(t *testing.T) {
	tests := []struct {
		name      string
		genre     string
		wantEyeR  float64
		wantShape string
	}{
		{"fantasy warm amber", "fantasy", 0.85, "circle"},
		{"horror pale grey", "horror", 0.70, "skull"},
		{"scifi cyan", "scifi", 0.30, "angular"},
		{"cyberpunk neon green", "cyberpunk", 0.20, "geometric"},
		{"postapoc ochre", "postapoc", 0.75, "rugged"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			world := NewWorld()
			sys := NewNpcFacialDetailSystem(world, 12345)
			sys.SetGenre(tt.genre)

			entity := NewEntity(0)
			entity.AddComponent(NewAIComponent(0, 0))
			entity.AddComponent(&StubSprite{Visible: true})

			sys.Update([]*Entity{entity}, 2.0) // exceed updateInterval

			comp, ok := entity.GetComponent("npc_facial_detail")
			if !ok {
				t.Fatal("expected npc_facial_detail component")
			}
			fc := comp.(*NpcFacialDetailComponent)

			// Eye color should be within ±5% of palette value
			diff := fc.EyeR - tt.wantEyeR
			if diff < -0.06 || diff > 0.06 {
				t.Errorf("genre %s: EyeR = %f, want ~%f", tt.genre, fc.EyeR, tt.wantEyeR)
			}

			// Head shape should be from genre palette
			palette := sys.palettes[tt.genre]
			found := false
			for _, s := range palette.HeadShapes {
				if fc.HeadShapeTag == s {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("genre %s: HeadShapeTag %q not in palette shapes %v",
					tt.genre, fc.HeadShapeTag, palette.HeadShapes)
			}
		})
	}
}

func TestNpcFacialDetailSystem_SkipsPlayer(t *testing.T) {
	world := NewWorld()
	sys := NewNpcFacialDetailSystem(world, 42)

	player := NewEntity(0)
	player.AddComponent(NewAIComponent(0, 0))
	player.AddComponent(&StubSprite{Visible: true})
	player.AddComponent(NewStubInput())

	sys.Update([]*Entity{player}, 2.0)

	_, ok := player.GetComponent("npc_facial_detail")
	if ok {
		t.Error("player entity should not get npc_facial_detail component")
	}
}

func TestNpcFacialDetailSystem_SkipsNoAI(t *testing.T) {
	world := NewWorld()
	sys := NewNpcFacialDetailSystem(world, 42)

	entity := NewEntity(0)
	entity.AddComponent(&StubSprite{Visible: true})

	sys.Update([]*Entity{entity}, 2.0)

	_, ok := entity.GetComponent("npc_facial_detail")
	if ok {
		t.Error("entity without AI should not get npc_facial_detail component")
	}
}

func TestNpcFacialDetailSystem_SkipsNoSprite(t *testing.T) {
	world := NewWorld()
	sys := NewNpcFacialDetailSystem(world, 42)

	entity := NewEntity(0)
	entity.AddComponent(NewAIComponent(0, 0))

	sys.Update([]*Entity{entity}, 2.0)

	_, ok := entity.GetComponent("npc_facial_detail")
	if ok {
		t.Error("entity without sprite should not get npc_facial_detail component")
	}
}

func TestNpcFacialDetailSystem_ThrottledUpdate(t *testing.T) {
	world := NewWorld()
	sys := NewNpcFacialDetailSystem(world, 42)

	entity := NewEntity(0)
	entity.AddComponent(NewAIComponent(0, 0))
	entity.AddComponent(&StubSprite{Visible: true})

	// First call: deltaTime < interval, should not apply
	sys.Update([]*Entity{entity}, 0.5)
	_, ok := entity.GetComponent("npc_facial_detail")
	if ok {
		t.Error("should not apply before update interval")
	}

	// Second call: accumulate past interval
	sys.Update([]*Entity{entity}, 0.6)
	_, ok = entity.GetComponent("npc_facial_detail")
	if !ok {
		t.Error("should apply after update interval elapsed")
	}
}

func TestNpcFacialDetailSystem_NoRecomputeWithoutGenreChange(t *testing.T) {
	world := NewWorld()
	sys := NewNpcFacialDetailSystem(world, 42)
	sys.SetGenre("horror")

	entity := NewEntity(0)
	entity.AddComponent(NewAIComponent(0, 0))
	entity.AddComponent(&StubSprite{Visible: true})

	// First update initializes
	sys.Update([]*Entity{entity}, 2.0)
	comp, _ := entity.GetComponent("npc_facial_detail")
	fc := comp.(*NpcFacialDetailComponent)
	origEyeR := fc.EyeR

	// Reset rng to different state to detect recomputation
	sys.rng.Float64()
	sys.rng.Float64()

	// Second update with same genre should not recompute
	sys.Update([]*Entity{entity}, 2.0)
	comp2, _ := entity.GetComponent("npc_facial_detail")
	fc2 := comp2.(*NpcFacialDetailComponent)
	if fc2.EyeR != origEyeR {
		t.Error("should not recompute when genre hasn't changed")
	}
}

func TestNpcFacialDetailSystem_ExpressionFromFaction(t *testing.T) {
	tests := []struct {
		name      string
		factionID string
		wantExpr  string
	}{
		{"boss hostile", "boss_faction", "hostile"},
		{"neutral friendly", "neutral_faction", "friendly"},
		{"merchant friendly", "merchant_faction", "friendly"},
		{"horror scared", "horror_faction", "scared"},
		{"default neutral", "random_faction", "neutral"},
		{"no faction", "", "neutral"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			world := NewWorld()
			sys := NewNpcFacialDetailSystem(world, 42)

			entity := NewEntity(0)
			entity.AddComponent(NewAIComponent(0, 0))
			entity.AddComponent(&StubSprite{Visible: true})
			if tt.factionID != "" {
				entity.AddComponent(&FactionComponent{FactionID: tt.factionID})
			}

			sys.Update([]*Entity{entity}, 2.0)

			comp, ok := entity.GetComponent("npc_facial_detail")
			if !ok {
				t.Fatal("expected npc_facial_detail component")
			}
			fc := comp.(*NpcFacialDetailComponent)
			if fc.ExpressionType != tt.wantExpr {
				t.Errorf("faction %q: expression = %q, want %q",
					tt.factionID, fc.ExpressionType, tt.wantExpr)
			}
		})
	}
}

func TestNpcFacialDetailSystem_EyeSizeFromCreatureSize(t *testing.T) {
	tests := []struct {
		name       string
		widthScale float64
		wantEye    float64
		wantMouth  float64
	}{
		{"large creature", 1.3, 3.0, 2.0},
		{"small creature", 0.7, 1.0, 1.0},
		{"medium creature", 1.0, 2.0, 1.0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			world := NewWorld()
			sys := NewNpcFacialDetailSystem(world, 42)

			entity := NewEntity(0)
			entity.AddComponent(NewAIComponent(0, 0))
			entity.AddComponent(&StubSprite{Visible: true})
			entity.AddComponent(&CreatureSizeProportionComponent{
				WidthScale: tt.widthScale,
				SizeTier:   "medium",
			})

			sys.Update([]*Entity{entity}, 2.0)

			comp, _ := entity.GetComponent("npc_facial_detail")
			fc := comp.(*NpcFacialDetailComponent)
			if fc.EyeSize != tt.wantEye {
				t.Errorf("WidthScale %f: EyeSize = %f, want %f",
					tt.widthScale, fc.EyeSize, tt.wantEye)
			}
			if fc.MouthSize != tt.wantMouth {
				t.Errorf("WidthScale %f: MouthSize = %f, want %f",
					tt.widthScale, fc.MouthSize, tt.wantMouth)
			}
		})
	}
}

func TestNpcFacialDetailSystem_UnknownGenreFallback(t *testing.T) {
	world := NewWorld()
	sys := NewNpcFacialDetailSystem(world, 42)
	sys.SetGenre("unknown_genre")

	entity := NewEntity(0)
	entity.AddComponent(NewAIComponent(0, 0))
	entity.AddComponent(&StubSprite{Visible: true})

	sys.Update([]*Entity{entity}, 2.0)

	comp, ok := entity.GetComponent("npc_facial_detail")
	if !ok {
		t.Fatal("expected npc_facial_detail component even for unknown genre")
	}
	fc := comp.(*NpcFacialDetailComponent)
	// Should fall back to fantasy palette
	diff := fc.EyeR - 0.85
	if diff < -0.06 || diff > 0.06 {
		t.Errorf("unknown genre should fall back to fantasy, EyeR = %f", fc.EyeR)
	}
}

func TestClampFacialColor(t *testing.T) {
	tests := []struct {
		name string
		in   float64
		want float64
	}{
		{"normal", 0.5, 0.5},
		{"below zero", -0.1, 0.0},
		{"above one", 1.5, 1.0},
		{"zero", 0.0, 0.0},
		{"one", 1.0, 1.0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := clampFacialColor(tt.in)
			if got != tt.want {
				t.Errorf("clampFacialColor(%f) = %f, want %f", tt.in, got, tt.want)
			}
		})
	}
}

func TestNpcFacialDetailSystem_DeterministicWithSameSeed(t *testing.T) {
	for _, genre := range []string{"fantasy", "horror", "scifi", "cyberpunk", "postapoc"} {
		t.Run(genre, func(t *testing.T) {
			world1 := NewWorld()
			sys1 := NewNpcFacialDetailSystem(world1, 99999)
			sys1.SetGenre(genre)

			world2 := NewWorld()
			sys2 := NewNpcFacialDetailSystem(world2, 99999)
			sys2.SetGenre(genre)

			e1 := NewEntity(0)
			e1.AddComponent(NewAIComponent(0, 0))
			e1.AddComponent(&StubSprite{Visible: true})

			e2 := NewEntity(0)
			e2.AddComponent(NewAIComponent(0, 0))
			e2.AddComponent(&StubSprite{Visible: true})

			sys1.Update([]*Entity{e1}, 2.0)
			sys2.Update([]*Entity{e2}, 2.0)

			c1, _ := e1.GetComponent("npc_facial_detail")
			c2, _ := e2.GetComponent("npc_facial_detail")
			fc1 := c1.(*NpcFacialDetailComponent)
			fc2 := c2.(*NpcFacialDetailComponent)

			if fc1.EyeR != fc2.EyeR || fc1.EyeG != fc2.EyeG || fc1.EyeB != fc2.EyeB {
				t.Errorf("same seed should produce same eye colors: %v vs %v",
					[3]float64{fc1.EyeR, fc1.EyeG, fc1.EyeB},
					[3]float64{fc2.EyeR, fc2.EyeG, fc2.EyeB})
			}
			if fc1.HeadShapeTag != fc2.HeadShapeTag {
				t.Errorf("same seed should produce same head shape: %q vs %q",
					fc1.HeadShapeTag, fc2.HeadShapeTag)
			}
		})
	}
}

func TestNpcFacialDetailSystem_NilWorld(t *testing.T) {
	sys := NewNpcFacialDetailSystem(nil, 42)
	if sys == nil {
		t.Fatal("should handle nil world")
	}
	// Should not panic on update
	entity := NewEntity(0)
	entity.AddComponent(NewAIComponent(0, 0))
	entity.AddComponent(&StubSprite{Visible: true})
	sys.Update([]*Entity{entity}, 2.0)
}

func BenchmarkNpcFacialDetailSystem_Update(b *testing.B) {
	world := NewWorld()
	sys := NewNpcFacialDetailSystem(world, 42)
	sys.SetGenre("fantasy")

	entities := make([]*Entity, 200)
	for i := range entities {
		e := NewEntity(0)
		e.AddComponent(NewAIComponent(0, 0))
		e.AddComponent(&StubSprite{Visible: true})
		entities[i] = e
	}

	// Force initial apply
	sys.Update(entities, 2.0)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// After initialization, throttled updates are near-zero cost
		sys.timeSinceCheck = 0
		sys.Update(entities, 0.5)
	}
}
