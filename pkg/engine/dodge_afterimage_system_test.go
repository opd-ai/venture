package engine

import (
	"testing"
)

func TestAfterimageComponentType(t *testing.T) {
	c := NewAfterimageComponent()
	if c.Type() != "afterimage" {
		t.Errorf("expected type afterimage, got %s", c.Type())
	}
}

func TestAfterimageComponentDefaults(t *testing.T) {
	c := NewAfterimageComponent()
	if c.MaxGhost != 5 {
		t.Errorf("expected MaxGhost 5, got %d", c.MaxGhost)
	}
	if c.Decay != 3.0 {
		t.Errorf("expected Decay 3.0, got %f", c.Decay)
	}
	if c.Interval != 0.06 {
		t.Errorf("expected Interval 0.06, got %f", c.Interval)
	}
	if len(c.Ghosts) != 0 {
		t.Errorf("expected 0 ghosts, got %d", len(c.Ghosts))
	}
	if c.SpeedThresholdSq != 14400.0 {
		t.Errorf("expected SpeedThresholdSq 14400, got %f", c.SpeedThresholdSq)
	}
}

func TestNewDodgeAfterimageSystem(t *testing.T) {
	world := NewWorld()
	sys := NewDodgeAfterimageSystem(world, 42)
	if sys == nil {
		t.Fatal("NewDodgeAfterimageSystem returned nil")
	}
	if sys.genreID != "fantasy" {
		t.Errorf("expected default genre fantasy, got %s", sys.genreID)
	}
	if sys.world != world {
		t.Error("world not set")
	}
}

func TestDodgeAfterimageSystem_SetGenre(t *testing.T) {
	tests := []struct {
		name    string
		genre   string
		wantR   float64
		wantG   float64
		wantB   float64
	}{
		{"fantasy", "fantasy", 1.0, 0.85, 0.4},
		{"horror", "horror", 0.4, 0.1, 0.1},
		{"cyberpunk", "cyberpunk", 0.2, 1.0, 0.9},
		{"scifi", "sci-fi", 0.3, 0.6, 1.0},
		{"postapoc", "postapoc", 0.8, 0.6, 0.3},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			world := NewWorld()
			sys := NewDodgeAfterimageSystem(world, 42)
			sys.SetGenre(tt.genre)
			if sys.preset.R != tt.wantR {
				t.Errorf("R = %f, want %f", sys.preset.R, tt.wantR)
			}
			if sys.preset.G != tt.wantG {
				t.Errorf("G = %f, want %f", sys.preset.G, tt.wantG)
			}
			if sys.preset.B != tt.wantB {
				t.Errorf("B = %f, want %f", sys.preset.B, tt.wantB)
			}
		})
	}
}

func TestDodgeAfterimageSystem_SkipsNoDeltaTime(t *testing.T) {
	world := NewWorld()
	sys := NewDodgeAfterimageSystem(world, 42)
	entity := NewEntity(1)
	entity.AddComponent(&PositionComponent{X: 10, Y: 20})
	entity.AddComponent(&VelocityComponent{VX: 200, VY: 0})
	// Zero delta should be a no-op
	sys.Update([]*Entity{entity}, 0)
	_, has := entity.GetComponent("afterimage")
	if has {
		t.Error("should not attach afterimage on zero deltaTime")
	}
}

func TestDodgeAfterimageSystem_SkipsLargeDeltaTime(t *testing.T) {
	world := NewWorld()
	sys := NewDodgeAfterimageSystem(world, 42)
	entity := NewEntity(1)
	entity.AddComponent(&PositionComponent{X: 10, Y: 20})
	entity.AddComponent(&VelocityComponent{VX: 200, VY: 0})
	// Large delta (lag spike) should be skipped
	sys.Update([]*Entity{entity}, 1.0)
	_, has := entity.GetComponent("afterimage")
	if has {
		t.Error("should not process on large deltaTime")
	}
}

func TestDodgeAfterimageSystem_SkipsNoPosition(t *testing.T) {
	world := NewWorld()
	sys := NewDodgeAfterimageSystem(world, 42)
	entity := NewEntity(1)
	entity.AddComponent(&VelocityComponent{VX: 200, VY: 0})
	sys.timeSinceScan = 2.0 // Force past scan interval
	sys.Update([]*Entity{entity}, 0.016)
	_, has := entity.GetComponent("afterimage")
	if has {
		t.Error("should not attach afterimage without position")
	}
}

func TestDodgeAfterimageSystem_SkipsNoVelocity(t *testing.T) {
	world := NewWorld()
	sys := NewDodgeAfterimageSystem(world, 42)
	entity := NewEntity(1)
	entity.AddComponent(&PositionComponent{X: 10, Y: 20})
	sys.timeSinceScan = 2.0
	sys.Update([]*Entity{entity}, 0.016)
	_, has := entity.GetComponent("afterimage")
	if has {
		t.Error("should not attach afterimage without velocity")
	}
}

func TestDodgeAfterimageSystem_SkipsSlowEntity(t *testing.T) {
	world := NewWorld()
	sys := NewDodgeAfterimageSystem(world, 42)
	entity := NewEntity(1)
	entity.AddComponent(&PositionComponent{X: 10, Y: 20})
	entity.AddComponent(&VelocityComponent{VX: 10, VY: 0}) // Too slow
	sys.timeSinceScan = 2.0
	sys.Update([]*Entity{entity}, 0.016)
	_, has := entity.GetComponent("afterimage")
	if has {
		t.Error("should not attach afterimage for slow entity")
	}
}

func TestDodgeAfterimageSystem_AttachesOnFastMovement(t *testing.T) {
	world := NewWorld()
	sys := NewDodgeAfterimageSystem(world, 42)
	entity := NewEntity(1)
	entity.AddComponent(&PositionComponent{X: 10, Y: 20})
	entity.AddComponent(&VelocityComponent{VX: 200, VY: 0}) // Fast enough
	sys.timeSinceScan = 2.0 // Force past scan interval
	sys.Update([]*Entity{entity}, 0.016)
	comp, has := entity.GetComponent("afterimage")
	if !has {
		t.Fatal("expected afterimage component to be attached")
	}
	ai, ok := comp.(*AfterimageComponent)
	if !ok {
		t.Fatal("expected *AfterimageComponent")
	}
	if ai.TintR != sys.preset.R || ai.TintG != sys.preset.G || ai.TintB != sys.preset.B {
		t.Error("tint colors not set from genre preset")
	}
}

func TestDodgeAfterimageSystem_SpawnsGhosts(t *testing.T) {
	world := NewWorld()
	sys := NewDodgeAfterimageSystem(world, 42)
	entity := NewEntity(1)
	entity.AddComponent(&PositionComponent{X: 50, Y: 60})
	entity.AddComponent(&VelocityComponent{VX: 200, VY: 0})
	ai := NewAfterimageComponent()
	ai.TimeSinceSpawn = 1.0 // Ready to spawn
	entity.AddComponent(ai)

	sys.Update([]*Entity{entity}, 0.016)

	if len(ai.Ghosts) != 1 {
		t.Fatalf("expected 1 ghost, got %d", len(ai.Ghosts))
	}
	if ai.Ghosts[0].X != 50 || ai.Ghosts[0].Y != 60 {
		t.Errorf("ghost position = (%f,%f), want (50,60)", ai.Ghosts[0].X, ai.Ghosts[0].Y)
	}
	if ai.Ghosts[0].Opacity != 0.6 {
		t.Errorf("ghost opacity = %f, want 0.6", ai.Ghosts[0].Opacity)
	}
}

func TestDodgeAfterimageSystem_DecaysGhosts(t *testing.T) {
	world := NewWorld()
	sys := NewDodgeAfterimageSystem(world, 42)
	entity := NewEntity(1)
	entity.AddComponent(&PositionComponent{X: 50, Y: 60})
	entity.AddComponent(&VelocityComponent{VX: 10, VY: 0}) // Slow — no new spawns
	ai := NewAfterimageComponent()
	ai.Ghosts = append(ai.Ghosts, AfterimageGhost{X: 40, Y: 50, Opacity: 0.5})
	entity.AddComponent(ai)

	// With Decay=3.0 and dt=0.1, opacity drops by 0.3 → 0.5-0.3=0.2
	sys.Update([]*Entity{entity}, 0.1)

	if len(ai.Ghosts) != 1 {
		t.Fatalf("expected 1 ghost remaining, got %d", len(ai.Ghosts))
	}
	expected := 0.5 - 3.0*0.1
	if ai.Ghosts[0].Opacity < expected-0.01 || ai.Ghosts[0].Opacity > expected+0.01 {
		t.Errorf("ghost opacity = %f, want ~%f", ai.Ghosts[0].Opacity, expected)
	}
}

func TestDodgeAfterimageSystem_RemovesFullyDecayedGhosts(t *testing.T) {
	world := NewWorld()
	sys := NewDodgeAfterimageSystem(world, 42)
	entity := NewEntity(1)
	entity.AddComponent(&PositionComponent{X: 50, Y: 60})
	entity.AddComponent(&VelocityComponent{VX: 10, VY: 0})
	ai := NewAfterimageComponent()
	ai.Ghosts = append(ai.Ghosts, AfterimageGhost{X: 40, Y: 50, Opacity: 0.02})
	entity.AddComponent(ai)

	// Decay of 3.0 * 0.1 = 0.3, opacity goes to -0.28, should be removed
	sys.Update([]*Entity{entity}, 0.1)

	if len(ai.Ghosts) != 0 {
		t.Errorf("expected 0 ghosts after full decay, got %d", len(ai.Ghosts))
	}
}

func TestDodgeAfterimageSystem_MaxGhostEviction(t *testing.T) {
	world := NewWorld()
	sys := NewDodgeAfterimageSystem(world, 42)
	entity := NewEntity(1)
	entity.AddComponent(&PositionComponent{X: 100, Y: 200})
	entity.AddComponent(&VelocityComponent{VX: 200, VY: 0})
	ai := NewAfterimageComponent()
	ai.MaxGhost = 3
	ai.TimeSinceSpawn = 1.0
	// Pre-fill with max ghosts
	ai.Ghosts = []AfterimageGhost{
		{X: 10, Y: 10, Opacity: 0.5},
		{X: 20, Y: 20, Opacity: 0.5},
		{X: 30, Y: 30, Opacity: 0.5},
	}
	entity.AddComponent(ai)

	sys.Update([]*Entity{entity}, 0.016)

	if len(ai.Ghosts) != 3 {
		t.Fatalf("expected 3 ghosts (max), got %d", len(ai.Ghosts))
	}
	// Newest ghost should be at index 2 with position (100,200)
	newest := ai.Ghosts[2]
	if newest.X != 100 || newest.Y != 200 {
		t.Errorf("newest ghost = (%f,%f), want (100,200)", newest.X, newest.Y)
	}
	// Oldest ghost (10,10) should have been evicted; index 0 should now be (20,20)
	if ai.Ghosts[0].X != 20 || ai.Ghosts[0].Y != 20 {
		t.Errorf("after eviction, ghost[0] = (%f,%f), want (20,20)", ai.Ghosts[0].X, ai.Ghosts[0].Y)
	}
}

func TestDodgeAfterimageSystem_SpawnIntervalThrottling(t *testing.T) {
	world := NewWorld()
	sys := NewDodgeAfterimageSystem(world, 42)
	entity := NewEntity(1)
	entity.AddComponent(&PositionComponent{X: 50, Y: 60})
	entity.AddComponent(&VelocityComponent{VX: 200, VY: 0})
	ai := NewAfterimageComponent()
	ai.TimeSinceSpawn = 0.0 // Just spawned
	entity.AddComponent(ai)

	// dt=0.01 < interval=0.06, should not spawn
	sys.Update([]*Entity{entity}, 0.01)
	if len(ai.Ghosts) != 0 {
		t.Errorf("expected 0 ghosts during cooldown, got %d", len(ai.Ghosts))
	}
}

func TestDodgeAfterimageSystem_GenrePresetsDistinct(t *testing.T) {
	world := NewWorld()
	sys := NewDodgeAfterimageSystem(world, 42)

	genres := []string{"fantasy", "horror", "cyberpunk", "sci-fi", "postapoc"}
	seen := make(map[float64]bool)
	for _, g := range genres {
		sys.SetGenre(g)
		key := sys.preset.R*1000 + sys.preset.G*100 + sys.preset.B*10
		if seen[key] {
			t.Errorf("duplicate preset colors for genre %s", g)
		}
		seen[key] = true
	}
}

func BenchmarkDodgeAfterimageSystem(b *testing.B) {
	world := NewWorld()
	sys := NewDodgeAfterimageSystem(world, 42)

	entities := make([]*Entity, 200)
	for i := range entities {
		e := NewEntity(uint64(i + 1))
		e.AddComponent(&PositionComponent{X: float64(i * 10), Y: float64(i * 5)})
		e.AddComponent(&VelocityComponent{VX: 200, VY: 100})
		ai := NewAfterimageComponent()
		ai.TimeSinceSpawn = 1.0
		ai.Ghosts = []AfterimageGhost{
			{X: 0, Y: 0, Opacity: 0.5},
			{X: 1, Y: 1, Opacity: 0.4},
			{X: 2, Y: 2, Opacity: 0.3},
		}
		e.AddComponent(ai)
		entities[i] = e
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.Update(entities, 0.016)
	}
}
