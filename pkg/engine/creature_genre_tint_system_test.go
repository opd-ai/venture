package engine

import (
	"testing"
)

func TestCreatureGenreTintComponent_Type(t *testing.T) {
	comp := NewCreatureGenreTintComponent()
	if comp.Type() != "creature_genre_tint" {
		t.Errorf("expected type creature_genre_tint, got %s", comp.Type())
	}
}

func TestCreatureGenreTintComponent_Defaults(t *testing.T) {
	comp := NewCreatureGenreTintComponent()
	if comp.TintR != 1.0 || comp.TintG != 1.0 || comp.TintB != 1.0 {
		t.Errorf("expected neutral tint (1,1,1), got (%f,%f,%f)", comp.TintR, comp.TintG, comp.TintB)
	}
}

func TestCreatureGenreTintSystem_SetGenre(t *testing.T) {
	world := NewWorld()
	sys := NewCreatureGenreTintSystem(world, 42)
	sys.SetGenre("horror")
	if sys.genreID != "horror" {
		t.Errorf("expected genre horror, got %s", sys.genreID)
	}
}

func TestCreatureGenreTintSystem_Update_GenrePresets(t *testing.T) {
	tests := []struct {
		name    string
		genre   string
		wantR   float64
		wantG   float64
		wantB   float64
		nearTol float64
	}{
		{"fantasy_warm", "fantasy", 1.00, 0.97, 0.88, 0.02},
		{"horror_cold", "horror", 0.82, 0.75, 0.80, 0.02},
		{"scifi_blue", "scifi", 0.88, 0.92, 1.00, 0.02},
		{"cyberpunk_neon", "cyberpunk", 1.00, 0.85, 0.95, 0.02},
		{"postapoc_dusty", "postapoc", 0.95, 0.88, 0.75, 0.02},
		{"unknown_neutral", "steampunk", 1.00, 1.00, 1.00, 0.02},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			world := NewWorld()
			sys := NewCreatureGenreTintSystem(world, 42)
			sys.SetGenre(tt.genre)

			creature := world.CreateEntity()
			creature.AddComponent(&PositionComponent{X: 50, Y: 50})
			creature.AddComponent(&EbitenSprite{Width: 32, Height: 32, Visible: true})
			creature.AddComponent(NewAIComponent(50, 50))

			entities := world.GetEntities()
			// Force update by exceeding interval
			sys.Update(entities, 2.0)

			comp, ok := creature.GetComponent("creature_genre_tint")
			if !ok {
				t.Fatal("creature_genre_tint component not created")
			}
			tint := comp.(*CreatureGenreTintComponent)

			if !nearEqual(tint.TintR, tt.wantR, tt.nearTol) {
				t.Errorf("TintR: got %f, want ~%f", tint.TintR, tt.wantR)
			}
			if !nearEqual(tint.TintG, tt.wantG, tt.nearTol) {
				t.Errorf("TintG: got %f, want ~%f", tint.TintG, tt.wantG)
			}
			if !nearEqual(tint.TintB, tt.wantB, tt.nearTol) {
				t.Errorf("TintB: got %f, want ~%f", tint.TintB, tt.wantB)
			}
		})
	}
}

func TestCreatureGenreTintSystem_SkipsPlayers(t *testing.T) {
	world := NewWorld()
	sys := NewCreatureGenreTintSystem(world, 42)
	sys.SetGenre("horror")

	player := world.CreateEntity()
	player.AddComponent(&PositionComponent{X: 50, Y: 50})
	player.AddComponent(&EbitenSprite{Width: 32, Height: 32, Visible: true})
	player.AddComponent(NewAIComponent(50, 50))
	player.AddComponent(NewStubInput())

	entities := world.GetEntities()
	sys.Update(entities, 2.0)

	_, ok := player.GetComponent("creature_genre_tint")
	if ok {
		t.Error("player should not have creature_genre_tint component")
	}
}

func TestCreatureGenreTintSystem_SkipsNonAIEntities(t *testing.T) {
	world := NewWorld()
	sys := NewCreatureGenreTintSystem(world, 42)
	sys.SetGenre("horror")

	wall := world.CreateEntity()
	wall.AddComponent(&PositionComponent{X: 10, Y: 10})
	wall.AddComponent(&EbitenSprite{Width: 32, Height: 32, Visible: true})

	entities := world.GetEntities()
	sys.Update(entities, 2.0)

	_, ok := wall.GetComponent("creature_genre_tint")
	if ok {
		t.Error("non-AI entity should not have creature_genre_tint component")
	}
}

func TestCreatureGenreTintSystem_BossModifier(t *testing.T) {
	world := NewWorld()
	sys := NewCreatureGenreTintSystem(world, 42)
	sys.SetGenre("fantasy")

	boss := world.CreateEntity()
	boss.AddComponent(&PositionComponent{X: 50, Y: 50})
	boss.AddComponent(&EbitenSprite{Width: 64, Height: 64, Visible: true})
	boss.AddComponent(NewAIComponent(50, 50))
	boss.AddComponent(&FactionComponent{FactionID: "boss_faction"})

	entities := world.GetEntities()
	sys.Update(entities, 2.0)

	comp, ok := boss.GetComponent("creature_genre_tint")
	if !ok {
		t.Fatal("creature_genre_tint component not created for boss")
	}
	tint := comp.(*CreatureGenreTintComponent)

	// Boss modifier is 0.85 × genre preset
	expectedR := clampCreatureTint(1.00 * 0.85)
	if !nearEqual(tint.TintR, expectedR, 0.02) {
		t.Errorf("Boss TintR: got %f, want ~%f", tint.TintR, expectedR)
	}
}

func TestCreatureGenreTintSystem_Throttled(t *testing.T) {
	world := NewWorld()
	sys := NewCreatureGenreTintSystem(world, 42)
	sys.SetGenre("horror")

	creature := world.CreateEntity()
	creature.AddComponent(&PositionComponent{X: 50, Y: 50})
	creature.AddComponent(&EbitenSprite{Width: 32, Height: 32, Visible: true})
	creature.AddComponent(NewAIComponent(50, 50))

	entities := world.GetEntities()

	// First update with enough time triggers application
	sys.Update(entities, 2.0)
	_, ok := creature.GetComponent("creature_genre_tint")
	if !ok {
		t.Fatal("expected tint after first sufficient interval")
	}

	// Second update within interval should not reprocess (genre unchanged)
	sys.SetGenre("scifi")
	sys.Update(entities, 0.1) // Not enough time
	comp, _ := creature.GetComponent("creature_genre_tint")
	tint := comp.(*CreatureGenreTintComponent)
	// Should still be horror tint since interval not met
	if nearEqual(tint.TintR, 0.88, 0.02) {
		t.Error("tint should not have changed to scifi yet (throttled)")
	}
}

func TestClampCreatureTint(t *testing.T) {
	tests := []struct {
		name string
		in   float64
		want float64
	}{
		{"normal", 0.85, 0.85},
		{"below_min", 0.3, 0.5},
		{"above_max", 1.5, 1.1},
		{"exact_min", 0.5, 0.5},
		{"exact_max", 1.1, 1.1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := clampCreatureTint(tt.in)
			if got != tt.want {
				t.Errorf("clampCreatureTint(%f) = %f, want %f", tt.in, got, tt.want)
			}
		})
	}
}

func TestBuildCreatureGenrePresets(t *testing.T) {
	presets := buildCreatureGenrePresets()
	expected := []string{"fantasy", "horror", "scifi", "cyberpunk", "postapoc"}
	for _, genre := range expected {
		if _, ok := presets[genre]; !ok {
			t.Errorf("missing preset for genre %s", genre)
		}
	}
}

// nearEqual checks if two floats are within tolerance.
func nearEqual(a, b, tol float64) bool {
	diff := a - b
	if diff < 0 {
		diff = -diff
	}
	return diff <= tol
}
