package engine

import (
	"math"
	"testing"
)

func TestTargetLockIndicatorComponentType(t *testing.T) {
	comp := NewTargetLockIndicatorComponent()
	if comp.Type() != "target_lock_indicator" {
		t.Errorf("expected type 'target_lock_indicator', got %q", comp.Type())
	}
}

func TestTargetLockIndicatorComponentDefaults(t *testing.T) {
	comp := NewTargetLockIndicatorComponent()
	if comp.Enabled {
		t.Error("expected Enabled=false by default")
	}
	if comp.BaseOpacity != 0.0 {
		t.Errorf("expected BaseOpacity=0.0, got %f", comp.BaseOpacity)
	}
	if comp.ReticleRadius != 12.0 {
		t.Errorf("expected ReticleRadius=12.0, got %f", comp.ReticleRadius)
	}
	if comp.RotationSpeed != 1.0 {
		t.Errorf("expected RotationSpeed=1.0, got %f", comp.RotationSpeed)
	}
}

func TestNewEntityTargetLockIndicatorSystem(t *testing.T) {
	world := NewWorld()
	sys := NewEntityTargetLockIndicatorSystem(world, 42)
	if sys == nil {
		t.Fatal("NewEntityTargetLockIndicatorSystem returned nil")
	}
	if sys.genreID != "fantasy" {
		t.Errorf("expected default genre 'fantasy', got %q", sys.genreID)
	}
	if sys.acquireRange != 160.0 {
		t.Errorf("expected acquireRange=160.0, got %f", sys.acquireRange)
	}
}

func TestEntityTargetLockIndicatorSystemSetGenre(t *testing.T) {
	world := NewWorld()
	sys := NewEntityTargetLockIndicatorSystem(world, 42)

	tests := []struct {
		name    string
		genreID string
	}{
		{"fantasy", "fantasy"},
		{"horror", "horror"},
		{"scifi", "scifi"},
		{"cyberpunk", "cyberpunk"},
		{"postapoc", "postapoc"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sys.SetGenre(tt.genreID)
			if sys.genreID != tt.genreID {
				t.Errorf("expected genre %q, got %q", tt.genreID, sys.genreID)
			}
		})
	}
}

func TestEntityTargetLockIndicatorSystemEmptyEntities(t *testing.T) {
	world := NewWorld()
	sys := NewEntityTargetLockIndicatorSystem(world, 42)
	// Should not panic with empty entity list
	sys.Update([]*Entity{}, 0.016)
	sys.Update(nil, 0.016)
}

func TestEntityTargetLockIndicatorSystemAcquiresTarget(t *testing.T) {
	world := NewWorld()
	sys := NewEntityTargetLockIndicatorSystem(world, 42)
	sys.SetGenre("fantasy")

	// Create player entity with input and position
	player := NewEntity(1)
	player.AddComponent(&PositionComponent{X: 100, Y: 100})
	player.AddComponent(&StubInput{})

	// Create hostile entity nearby
	hostile := NewEntity(2)
	hostile.AddComponent(&PositionComponent{X: 120, Y: 100})
	hostile.AddComponent(&AIComponent{})
	hostile.AddComponent(&HealthComponent{Current: 50, Max: 100})

	entities := []*Entity{player, hostile}

	// Run enough updates to trigger acquisition
	sys.Update(entities, 0.3)

	// Check that hostile got a reticle
	comp, ok := hostile.GetComponent("target_lock_indicator")
	if !ok {
		t.Fatal("expected target_lock_indicator component on hostile")
	}
	tlc, ok := comp.(*TargetLockIndicatorComponent)
	if !ok {
		t.Fatal("component type assertion failed")
	}
	if !tlc.Enabled {
		t.Error("expected reticle to be enabled")
	}
	if tlc.LockedByPlayerID != player.ID {
		t.Errorf("expected LockedByPlayerID=%d, got %d", player.ID, tlc.LockedByPlayerID)
	}
	if tlc.ReticleR == 0 && tlc.ReticleG == 0 && tlc.ReticleB == 0 {
		t.Error("expected non-zero reticle color")
	}
}

func TestEntityTargetLockIndicatorSystemIgnoresDeadHostiles(t *testing.T) {
	world := NewWorld()
	sys := NewEntityTargetLockIndicatorSystem(world, 42)

	player := NewEntity(3)
	player.AddComponent(&PositionComponent{X: 100, Y: 100})
	player.AddComponent(&StubInput{})

	// Dead hostile - should be skipped
	deadHostile := NewEntity(4)
	deadHostile.AddComponent(&PositionComponent{X: 120, Y: 100})
	deadHostile.AddComponent(&AIComponent{})
	deadHostile.AddComponent(&HealthComponent{Current: 0, Max: 100})

	entities := []*Entity{player, deadHostile}
	sys.Update(entities, 0.3)

	_, ok := deadHostile.GetComponent("target_lock_indicator")
	if ok {
		comp, _ := deadHostile.GetComponent("target_lock_indicator")
		if tlc, ok := comp.(*TargetLockIndicatorComponent); ok && tlc.Enabled {
			t.Error("dead hostile should not receive target lock indicator")
		}
	}
}

func TestEntityTargetLockIndicatorSystemPicksClosest(t *testing.T) {
	world := NewWorld()
	sys := NewEntityTargetLockIndicatorSystem(world, 42)

	player := NewEntity(5)
	player.AddComponent(&PositionComponent{X: 100, Y: 100})
	player.AddComponent(&StubInput{})

	// Far hostile
	farHostile := NewEntity(6)
	farHostile.AddComponent(&PositionComponent{X: 200, Y: 100})
	farHostile.AddComponent(&AIComponent{})
	farHostile.AddComponent(&HealthComponent{Current: 50, Max: 100})

	// Close hostile
	closeHostile := NewEntity(7)
	closeHostile.AddComponent(&PositionComponent{X: 110, Y: 100})
	closeHostile.AddComponent(&AIComponent{})
	closeHostile.AddComponent(&HealthComponent{Current: 50, Max: 100})

	entities := []*Entity{player, farHostile, closeHostile}
	sys.Update(entities, 0.3)

	// Close hostile should have the reticle
	comp, ok := closeHostile.GetComponent("target_lock_indicator")
	if !ok {
		t.Fatal("expected target_lock_indicator on close hostile")
	}
	tlc := comp.(*TargetLockIndicatorComponent)
	if !tlc.Enabled {
		t.Error("expected close hostile reticle to be enabled")
	}
}

func TestEntityTargetLockIndicatorSystemOutOfRange(t *testing.T) {
	world := NewWorld()
	sys := NewEntityTargetLockIndicatorSystem(world, 42)

	player := NewEntity(8)
	player.AddComponent(&PositionComponent{X: 100, Y: 100})
	player.AddComponent(&StubInput{})

	// Hostile beyond acquire range (160px)
	farHostile := NewEntity(9)
	farHostile.AddComponent(&PositionComponent{X: 400, Y: 100})
	farHostile.AddComponent(&AIComponent{})
	farHostile.AddComponent(&HealthComponent{Current: 50, Max: 100})

	entities := []*Entity{player, farHostile}
	sys.Update(entities, 0.3)

	_, ok := farHostile.GetComponent("target_lock_indicator")
	if ok {
		comp, _ := farHostile.GetComponent("target_lock_indicator")
		if tlc, ok := comp.(*TargetLockIndicatorComponent); ok && tlc.Enabled {
			t.Error("hostile beyond range should not receive target lock")
		}
	}
}

func TestEntityTargetLockIndicatorSystemAnimation(t *testing.T) {
	world := NewWorld()
	sys := NewEntityTargetLockIndicatorSystem(world, 42)

	player := NewEntity(10)
	player.AddComponent(&PositionComponent{X: 100, Y: 100})
	player.AddComponent(&StubInput{})

	hostile := NewEntity(11)
	hostile.AddComponent(&PositionComponent{X: 120, Y: 100})
	hostile.AddComponent(&AIComponent{})
	hostile.AddComponent(&HealthComponent{Current: 50, Max: 100})

	entities := []*Entity{player, hostile}

	// Initial acquisition
	sys.Update(entities, 0.3)

	comp, _ := hostile.GetComponent("target_lock_indicator")
	tlc := comp.(*TargetLockIndicatorComponent)
	initialAngle := tlc.RotationAngle

	// Run several frames of animation (don't trigger reacquire)
	for i := 0; i < 5; i++ {
		sys.Update(entities, 0.016)
	}

	if tlc.RotationAngle <= initialAngle {
		t.Error("expected rotation angle to increase over time")
	}
	if tlc.CurrentOpacity <= 0 {
		t.Error("expected non-zero current opacity during animation")
	}
}

func TestEntityTargetLockIndicatorSystemGenrePalettes(t *testing.T) {
	tests := []struct {
		genre string
		wantR float64
		wantG float64
		wantB float64
	}{
		{"fantasy", 0.95, 0.80, 0.25},
		{"horror", 0.90, 0.15, 0.10},
		{"scifi", 0.25, 0.90, 0.95},
		{"cyberpunk", 0.95, 0.20, 0.90},
		{"postapoc", 0.80, 0.55, 0.20},
	}

	for _, tt := range tests {
		t.Run(tt.genre, func(t *testing.T) {
			world := NewWorld()
			sys := NewEntityTargetLockIndicatorSystem(world, 42)
			sys.SetGenre(tt.genre)

			player := NewEntity(12)
			player.AddComponent(&PositionComponent{X: 100, Y: 100})
			player.AddComponent(&StubInput{})

			hostile := NewEntity(13)
			hostile.AddComponent(&PositionComponent{X: 120, Y: 100})
			hostile.AddComponent(&AIComponent{})
			hostile.AddComponent(&HealthComponent{Current: 50, Max: 100})

			sys.Update([]*Entity{player, hostile}, 0.3)

			comp, ok := hostile.GetComponent("target_lock_indicator")
			if !ok {
				t.Fatal("expected target_lock_indicator component")
			}
			tlc := comp.(*TargetLockIndicatorComponent)

			if math.Abs(tlc.ReticleR-tt.wantR) > 0.01 {
				t.Errorf("genre %s: expected R=%.2f, got %.2f", tt.genre, tt.wantR, tlc.ReticleR)
			}
			if math.Abs(tlc.ReticleG-tt.wantG) > 0.01 {
				t.Errorf("genre %s: expected G=%.2f, got %.2f", tt.genre, tt.wantG, tlc.ReticleG)
			}
			if math.Abs(tlc.ReticleB-tt.wantB) > 0.01 {
				t.Errorf("genre %s: expected B=%.2f, got %.2f", tt.genre, tt.wantB, tlc.ReticleB)
			}
		})
	}
}

func TestEntityTargetLockIndicatorSystemClearsStaleTargets(t *testing.T) {
	world := NewWorld()
	sys := NewEntityTargetLockIndicatorSystem(world, 42)

	player := NewEntity(14)
	player.AddComponent(&PositionComponent{X: 100, Y: 100})
	player.AddComponent(&StubInput{})

	hostile := NewEntity(15)
	hostile.AddComponent(&PositionComponent{X: 120, Y: 100})
	hostile.AddComponent(&AIComponent{})
	hostile.AddComponent(&HealthComponent{Current: 50, Max: 100})

	entities := []*Entity{player, hostile}
	sys.Update(entities, 0.3)

	// Confirm target acquired
	comp, _ := hostile.GetComponent("target_lock_indicator")
	tlc := comp.(*TargetLockIndicatorComponent)
	if !tlc.Enabled {
		t.Fatal("expected reticle to be enabled initially")
	}

	// Move hostile out of range and reacquire
	hostile.AddComponent(&PositionComponent{X: 500, Y: 500})
	sys.timeSinceAcquire = sys.acquisitionInterval // Force reacquire
	sys.Update(entities, 0.3)

	if tlc.Enabled {
		t.Error("expected reticle to be disabled after target moves out of range")
	}
}

func TestEntityTargetLockIndicatorSystemComputeReticleRadius(t *testing.T) {
	world := NewWorld()
	sys := NewEntityTargetLockIndicatorSystem(world, 42)

	tests := []struct {
		name     string
		collider *ColliderComponent
		wantMin  float64
		wantMax  float64
	}{
		{"no collider", nil, 11.0, 13.0},
		{"small collider", &ColliderComponent{Width: 8, Height: 8}, 8.0, 10.0},
		{"large collider", &ColliderComponent{Width: 40, Height: 40}, 24.0, 32.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entity := NewEntity(16)
			if tt.collider != nil {
				entity.AddComponent(tt.collider)
			}
			radius := sys.computeReticleRadius(entity)
			if radius < tt.wantMin || radius > tt.wantMax {
				t.Errorf("expected radius in [%.1f, %.1f], got %.1f", tt.wantMin, tt.wantMax, radius)
			}
		})
	}
}

func TestEntityTargetLockIndicatorSystemNilEntities(t *testing.T) {
	world := NewWorld()
	sys := NewEntityTargetLockIndicatorSystem(world, 42)

	player := NewEntity(17)
	player.AddComponent(&PositionComponent{X: 100, Y: 100})
	player.AddComponent(&StubInput{})

	// Should not panic with nil entities in the slice
	entities := []*Entity{nil, player, nil}
	sys.Update(entities, 0.3)
}

func TestEntityTargetLockIndicatorSystemUnknownGenreFallback(t *testing.T) {
	world := NewWorld()
	sys := NewEntityTargetLockIndicatorSystem(world, 42)
	sys.SetGenre("unknown_genre")

	palette := sys.getCurrentPalette()
	// Should fall back to fantasy
	if math.Abs(palette.R-0.95) > 0.01 {
		t.Errorf("expected fallback to fantasy palette R=0.95, got %.2f", palette.R)
	}
}

func TestBuildReticlePalettes(t *testing.T) {
	palettes := buildReticlePalettes()
	expected := []string{"fantasy", "horror", "scifi", "cyberpunk", "postapoc"}
	for _, genre := range expected {
		if _, ok := palettes[genre]; !ok {
			t.Errorf("missing palette for genre %q", genre)
		}
	}
	// All palettes should have valid ranges
	for genre, p := range palettes {
		if p.R < 0 || p.R > 1 || p.G < 0 || p.G > 1 || p.B < 0 || p.B > 1 {
			t.Errorf("genre %s: color out of [0,1] range", genre)
		}
		if p.Opacity < 0 || p.Opacity > 1 {
			t.Errorf("genre %s: opacity out of [0,1] range", genre)
		}
		if p.RotationSpeed <= 0 {
			t.Errorf("genre %s: rotation speed must be positive", genre)
		}
		if p.PulseSpeed <= 0 {
			t.Errorf("genre %s: pulse speed must be positive", genre)
		}
	}
}
