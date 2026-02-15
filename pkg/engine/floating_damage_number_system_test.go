package engine

import (
	"testing"
)

func TestFloatingDamageNumberComponentType(t *testing.T) {
	comp := NewFloatingDamageNumberComponent()
	if comp.Type() != "floating_damage_number" {
		t.Errorf("expected type 'floating_damage_number', got '%s'", comp.Type())
	}
}

func TestFloatingDamageNumberComponentDefaults(t *testing.T) {
	comp := NewFloatingDamageNumberComponent()
	if comp.MaxCount != 8 {
		t.Errorf("expected MaxCount 8, got %d", comp.MaxCount)
	}
	if len(comp.Numbers) != 0 {
		t.Errorf("expected empty Numbers, got %d", len(comp.Numbers))
	}
}

func TestFloatingDamageNumberSystemCreation(t *testing.T) {
	world := NewWorld()
	sys := NewFloatingDamageNumberSystem(world, 42)
	if sys == nil {
		t.Fatal("expected non-nil system")
	}
	if sys.genreID != "fantasy" {
		t.Errorf("expected default genre 'fantasy', got '%s'", sys.genreID)
	}
}

func TestFloatingDamageNumberSystemSetGenre(t *testing.T) {
	tests := []struct {
		name    string
		genreID string
	}{
		{"fantasy", "fantasy"},
		{"horror", "horror"},
		{"cyberpunk", "cyberpunk"},
		{"scifi", "scifi"},
		{"postapoc", "postapoc"},
	}

	world := NewWorld()
	sys := NewFloatingDamageNumberSystem(world, 42)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sys.SetGenre(tt.genreID)
			if sys.genreID != tt.genreID {
				t.Errorf("expected genre '%s', got '%s'", tt.genreID, sys.genreID)
			}
			if sys.preset.Duration <= 0 {
				t.Error("expected positive duration")
			}
			if sys.preset.RiseSpeed <= 0 {
				t.Error("expected positive rise speed")
			}
		})
	}
}

func TestFloatingDamageNumberSystemDamageSpawn(t *testing.T) {
	world := NewWorld()
	sys := NewFloatingDamageNumberSystem(world, 42)

	entity := NewEntity(1)
	healthComp := &HealthComponent{Current: 100, Max: 100}
	entity.AddComponent(healthComp)

	entities := []*Entity{entity}

	// First update: records initial health, no numbers spawned
	sys.Update(entities, 0.016)
	fdnComp, _ := entity.GetComponent("floating_damage_number")
	if fdnComp != nil {
		comp := fdnComp.(*FloatingDamageNumberComponent)
		if len(comp.Numbers) != 0 {
			t.Errorf("expected 0 numbers on first update, got %d", len(comp.Numbers))
		}
	}

	// Simulate damage
	healthComp.Current = 85
	sys.Update(entities, 0.016)

	fdnComp, _ = entity.GetComponent("floating_damage_number")
	if fdnComp == nil {
		t.Fatal("expected FloatingDamageNumberComponent after damage")
	}
	comp := fdnComp.(*FloatingDamageNumberComponent)
	if len(comp.Numbers) != 1 {
		t.Fatalf("expected 1 floating number, got %d", len(comp.Numbers))
	}
	if comp.Numbers[0].Amount != 15 {
		t.Errorf("expected amount 15, got %f", comp.Numbers[0].Amount)
	}
	if comp.Numbers[0].IsHeal {
		t.Error("expected damage, not heal")
	}
}

func TestFloatingDamageNumberSystemHealSpawn(t *testing.T) {
	world := NewWorld()
	sys := NewFloatingDamageNumberSystem(world, 42)

	entity := NewEntity(1)
	healthComp := &HealthComponent{Current: 50, Max: 100}
	entity.AddComponent(healthComp)

	entities := []*Entity{entity}

	// Record initial health
	sys.Update(entities, 0.016)

	// Simulate healing
	healthComp.Current = 70
	sys.Update(entities, 0.016)

	fdnComp, _ := entity.GetComponent("floating_damage_number")
	if fdnComp == nil {
		t.Fatal("expected FloatingDamageNumberComponent after healing")
	}
	comp := fdnComp.(*FloatingDamageNumberComponent)
	if len(comp.Numbers) != 1 {
		t.Fatalf("expected 1 floating number, got %d", len(comp.Numbers))
	}
	if !comp.Numbers[0].IsHeal {
		t.Error("expected heal, not damage")
	}
	if comp.Numbers[0].Amount != 20 {
		t.Errorf("expected amount 20, got %f", comp.Numbers[0].Amount)
	}
}

func TestFloatingDamageNumberSystemCriticalHit(t *testing.T) {
	world := NewWorld()
	sys := NewFloatingDamageNumberSystem(world, 42)

	entity := NewEntity(1)
	healthComp := &HealthComponent{Current: 100, Max: 100}
	entity.AddComponent(healthComp)

	entities := []*Entity{entity}
	sys.Update(entities, 0.016)

	// Damage >= 20% of max health = crit
	healthComp.Current = 75
	sys.Update(entities, 0.016)

	fdnCompRaw, _ := entity.GetComponent("floating_damage_number")
	comp := fdnCompRaw.(*FloatingDamageNumberComponent)
	if !comp.Numbers[0].IsCrit {
		t.Error("expected critical hit for 25% max health damage")
	}
	if comp.Numbers[0].Scale <= 1.0 {
		t.Errorf("expected crit scale > 1.0, got %f", comp.Numbers[0].Scale)
	}
}

func TestFloatingDamageNumberSystemNumberFadeout(t *testing.T) {
	world := NewWorld()
	sys := NewFloatingDamageNumberSystem(world, 42)

	entity := NewEntity(1)
	healthComp := &HealthComponent{Current: 100, Max: 100}
	entity.AddComponent(healthComp)

	entities := []*Entity{entity}
	sys.Update(entities, 0.016)

	// Spawn a number
	healthComp.Current = 90
	sys.Update(entities, 0.016)

	fdnCompRaw, _ := entity.GetComponent("floating_damage_number")
	comp := fdnCompRaw.(*FloatingDamageNumberComponent)
	if len(comp.Numbers) != 1 {
		t.Fatalf("expected 1 number, got %d", len(comp.Numbers))
	}

	// Advance past duration - number should expire
	sys.Update(entities, sys.preset.Duration+0.1)

	if len(comp.Numbers) != 0 {
		t.Errorf("expected 0 numbers after expiry, got %d", len(comp.Numbers))
	}
}

func TestFloatingDamageNumberSystemRises(t *testing.T) {
	world := NewWorld()
	sys := NewFloatingDamageNumberSystem(world, 42)

	entity := NewEntity(1)
	healthComp := &HealthComponent{Current: 100, Max: 100}
	entity.AddComponent(healthComp)

	entities := []*Entity{entity}
	sys.Update(entities, 0.016)

	healthComp.Current = 90
	sys.Update(entities, 0.016)

	fdnCompRaw, _ := entity.GetComponent("floating_damage_number")
	comp := fdnCompRaw.(*FloatingDamageNumberComponent)
	initialY := comp.Numbers[0].OffsetY

	// Advance time - number should rise
	sys.Update(entities, 0.2)
	if comp.Numbers[0].OffsetY >= initialY {
		t.Error("expected number to rise (more negative OffsetY)")
	}
}

func TestFloatingDamageNumberSystemMaxCount(t *testing.T) {
	world := NewWorld()
	sys := NewFloatingDamageNumberSystem(world, 42)

	entity := NewEntity(1)
	healthComp := &HealthComponent{Current: 100, Max: 100}
	entity.AddComponent(healthComp)

	entities := []*Entity{entity}
	sys.Update(entities, 0.016)

	// Spawn more than MaxCount numbers
	for i := 0; i < 10; i++ {
		healthComp.Current -= 1
		sys.Update(entities, 0.001) // small delta so numbers don't expire
	}

	fdnCompRaw, _ := entity.GetComponent("floating_damage_number")
	comp := fdnCompRaw.(*FloatingDamageNumberComponent)
	if len(comp.Numbers) > comp.MaxCount {
		t.Errorf("expected at most %d numbers, got %d", comp.MaxCount, len(comp.Numbers))
	}
}

func TestFloatingDamageNumberSystemNoChangeNoSpawn(t *testing.T) {
	world := NewWorld()
	sys := NewFloatingDamageNumberSystem(world, 42)

	entity := NewEntity(1)
	healthComp := &HealthComponent{Current: 100, Max: 100}
	entity.AddComponent(healthComp)

	entities := []*Entity{entity}

	// Multiple updates with no health change
	for i := 0; i < 5; i++ {
		sys.Update(entities, 0.016)
	}

	fdnComp, _ := entity.GetComponent("floating_damage_number")
	if fdnComp != nil {
		comp := fdnComp.(*FloatingDamageNumberComponent)
		if len(comp.Numbers) != 0 {
			t.Errorf("expected 0 numbers with no health change, got %d", len(comp.Numbers))
		}
	}
}

func TestFloatingDamageNumberSystemStaleCleanup(t *testing.T) {
	world := NewWorld()
	sys := NewFloatingDamageNumberSystem(world, 42)
	sys.cleanupInterval = 0.1

	entity := NewEntity(1)
	healthComp := &HealthComponent{Current: 100, Max: 100}
	entity.AddComponent(healthComp)

	entities := []*Entity{entity}
	sys.Update(entities, 0.016)

	if len(sys.prevHealth) != 1 {
		t.Fatalf("expected 1 tracked entity, got %d", len(sys.prevHealth))
	}

	// Update with empty entity list + enough time for cleanup
	sys.Update([]*Entity{}, 0.2)

	if len(sys.prevHealth) != 0 {
		t.Errorf("expected 0 tracked entities after cleanup, got %d", len(sys.prevHealth))
	}
}

func TestFloatingDamageNumberSystemGenreColors(t *testing.T) {
	world := NewWorld()
	sys := NewFloatingDamageNumberSystem(world, 42)

	genres := []string{"fantasy", "horror", "cyberpunk", "scifi", "postapoc"}
	for _, genre := range genres {
		t.Run(genre, func(t *testing.T) {
			sys.SetGenre(genre)

			entity := NewEntity(uint64(100))
			healthComp := &HealthComponent{Current: 100, Max: 100}
			entity.AddComponent(healthComp)

			entities := []*Entity{entity}

			// Reset tracking
			sys.prevHealth = make(map[uint64]float64, 128)
			sys.Update(entities, 0.016)

			healthComp.Current = 90
			sys.Update(entities, 0.016)

			fdnCompRaw, _ := entity.GetComponent("floating_damage_number")
	comp := fdnCompRaw.(*FloatingDamageNumberComponent)
			if len(comp.Numbers) == 0 {
				t.Fatal("expected at least 1 number")
			}
			if comp.Numbers[0].Color.A == 0 {
				t.Error("expected non-zero alpha on damage number color")
			}
		})
	}
}

func TestFloatingDamageNumberSystemEntityWithoutHealth(t *testing.T) {
	world := NewWorld()
	sys := NewFloatingDamageNumberSystem(world, 42)

	entity := NewEntity(1)
	// No health component
	entities := []*Entity{entity}
	sys.Update(entities, 0.016) // Should not panic
}

func BenchmarkFloatingDamageNumberSystem(b *testing.B) {
	world := NewWorld()
	sys := NewFloatingDamageNumberSystem(world, 42)

	entities := make([]*Entity, 200)
	for i := range entities {
		e := NewEntity(uint64(i + 1))
		healthComp := &HealthComponent{Current: 100, Max: 100}
		e.AddComponent(healthComp)
		e.AddComponent(NewFloatingDamageNumberComponent())
		entities[i] = e
	}

	// Prime the system
	sys.Update(entities, 0.016)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.Update(entities, 0.016)
	}
}
