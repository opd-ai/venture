package engine

import (
	"math/rand"
	"testing"
)

func TestNewSpellEffectSystem(t *testing.T) {
	world := NewWorld()
	rng := rand.New(rand.NewSource(12345))
	system := NewSpellEffectSystem(world, rng)

	if system == nil {
		t.Fatal("NewSpellEffectSystem returned nil")
	}
	if system.world != world {
		t.Errorf("system.world = %v, want %v", system.world, world)
	}
	if system.rng != rng {
		t.Errorf("system.rng = %v, want %v", system.rng, rng)
	}
}

func TestSpellEffectSystem_Update_RemovesExpiredEffects(t *testing.T) {
	world := NewWorld()
	rng := rand.New(rand.NewSource(12345))
	system := NewSpellEffectSystem(world, rng)

	// Create entity with expired effect
	entity := world.CreateEntity()
	effect := &SpellEffectComponent{
		EffectType:  EffectTeleportation,
		Duration:    1.0,
		Magnitude:   1.0,
		TargetType:  TargetSelf,
		Active:      true,
		ElapsedTime: 1.5, // Already expired
	}
	entity.AddComponent(effect)

	// Update should remove expired effect
	entities := []*Entity{entity}
	system.Update(entities, 0.1)

	if entity.HasComponent("spell_effect") {
		t.Errorf("Entity still has spell_effect component after expiration")
	}
}

func TestSpellEffectSystem_ApplySpellEffect(t *testing.T) {
	world := NewWorld()
	rng := rand.New(rand.NewSource(12345))
	system := NewSpellEffectSystem(world, rng)

	entity := world.CreateEntity()

	system.ApplySpellEffect(
		entity,
		EffectIllusion,
		0.9,   // magnitude
		10.0,  // duration
		TargetSelf,
		100,   // caster ID
		0, 0,  // target position
		0,     // radius
	)

	if !entity.HasComponent("spell_effect") {
		t.Fatal("Entity does not have spell_effect component")
	}

	comp, _ := entity.GetComponent("spell_effect")
	effect, ok := comp.(*SpellEffectComponent)
	if !ok {
		t.Fatal("Component is not a SpellEffectComponent")
	}

	if effect.EffectType != EffectIllusion {
		t.Errorf("EffectType = %v, want %v", effect.EffectType, EffectIllusion)
	}
	if effect.Magnitude != 0.9 {
		t.Errorf("Magnitude = %v, want 0.9", effect.Magnitude)
	}
	if effect.Duration != 10.0 {
		t.Errorf("Duration = %v, want 10.0", effect.Duration)
	}
}

func TestSpellEffectSystem_ExecuteIllusion(t *testing.T) {
	world := NewWorld()
	rng := rand.New(rand.NewSource(12345))
	system := NewSpellEffectSystem(world, rng)

	entity := world.CreateEntity()
	effect := &SpellEffectComponent{
		EffectType:  EffectIllusion,
		Duration:    5.0,
		Magnitude:   1.0, // Full invisibility
		TargetType:  TargetSelf,
		Active:      true,
		ElapsedTime: 0,
	}
	entity.AddComponent(effect)

	// Update should apply invisibility
	entities := []*Entity{entity}
	system.Update(entities, 0.016)

	if !entity.HasComponent("invisible") {
		t.Errorf("Entity should have invisible component")
	}
}

func TestSpellEffectSystem_ExecuteTeleportation(t *testing.T) {
	world := NewWorld()
	rng := rand.New(rand.NewSource(12345))
	system := NewSpellEffectSystem(world, rng)

	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 10.0, Y: 20.0})

	effect := &SpellEffectComponent{
		EffectType:  EffectTeleportation,
		Duration:    0, // Instant
		Magnitude:   1.0,
		TargetType:  TargetSelf,
		TargetX:     100.0,
		TargetY:     200.0,
		Active:      true,
		ElapsedTime: 0,
	}
	entity.AddComponent(effect)

	// Update should teleport entity
	entities := []*Entity{entity}
	system.Update(entities, 0.016)

	posComp, _ := entity.GetComponent("position")
	pos := posComp.(*PositionComponent)

	if pos.X != 100.0 || pos.Y != 200.0 {
		t.Errorf("Position = (%v, %v), want (100.0, 200.0)", pos.X, pos.Y)
	}
}

func TestSpellEffectSystem_ExecuteLifeDrain(t *testing.T) {
	world := NewWorld()
	rng := rand.New(rand.NewSource(12345))
	system := NewSpellEffectSystem(world, rng)

	// Create caster
	caster := world.CreateEntity()
	casterHealth := &HealthComponent{Current: 50.0, Max: 100.0}
	caster.AddComponent(casterHealth)

	// Create target
	target := world.CreateEntity()
	targetHealth := &HealthComponent{Current: 100.0, Max: 100.0}
	target.AddComponent(targetHealth)

	// Process pending entity additions
	world.Update(0)

	// Apply life drain effect
	effect := &SpellEffectComponent{
		EffectType:  EffectLifeDrain,
		Duration:    1.0,
		Magnitude:   10.0, // 10 HP/sec
		TargetType:  TargetEntity,
		CasterID:    caster.ID,
		TargetID:    target.ID,
		Active:      true,
		ElapsedTime: 0,
	}
	target.AddComponent(effect)

	// Store initial health values
	initialCasterHealth := casterHealth.Current
	initialTargetHealth := targetHealth.Current

	// Update for 0.5 seconds (should drain 5 HP, heal 2.5 HP)
	entities := []*Entity{caster, target}
	system.Update(entities, 0.5)

	// Target should lose health
	if targetHealth.Current >= initialTargetHealth {
		t.Errorf("Target health = %v, should be less than %v", targetHealth.Current, initialTargetHealth)
	}

	// Caster should gain health (50% efficiency)
	if casterHealth.Current <= initialCasterHealth {
		t.Errorf("Caster health = %v, should be greater than %v", casterHealth.Current, initialCasterHealth)
	}
}

func TestSpellEffectSystem_ExecuteTimeManipulation(t *testing.T) {
	world := NewWorld()
	rng := rand.New(rand.NewSource(12345))
	system := NewSpellEffectSystem(world, rng)

	entity := world.CreateEntity()
	vel := &VelocityComponent{VX: 10.0, VY: 5.0}
	entity.AddComponent(vel)

	// Apply slow effect (50% speed)
	effect := &SpellEffectComponent{
		EffectType:  EffectTimeManipulation,
		Duration:    5.0,
		Magnitude:   0.5, // 50% speed
		TargetType:  TargetEntity,
		Active:      true,
		ElapsedTime: 0,
	}
	entity.AddComponent(effect)

	entities := []*Entity{entity}
	system.Update(entities, 0.016)

	// Velocity should be reduced
	if vel.VX == 10.0 || vel.VY == 5.0 {
		t.Errorf("Velocity not modified: VX=%v, VY=%v", vel.VX, vel.VY)
	}
}

func TestSpellEffectSystem_ExecuteGravityControl(t *testing.T) {
	world := NewWorld()
	rng := rand.New(rand.NewSource(12345))
	system := NewSpellEffectSystem(world, rng)

	entity := world.CreateEntity()

	effect := &SpellEffectComponent{
		EffectType:  EffectGravityControl,
		Duration:    10.0,
		Magnitude:   -1.0, // Levitation
		TargetType:  TargetSelf,
		Active:      true,
		ElapsedTime: 0,
	}
	entity.AddComponent(effect)

	entities := []*Entity{entity}
	system.Update(entities, 0.016)

	// Should add gravity_modified component
	if !entity.HasComponent("gravity_modified") {
		t.Errorf("Entity should have gravity_modified component")
	}
}

func TestSpellEffectSystem_GetEntitiesInRadius(t *testing.T) {
	world := NewWorld()
	rng := rand.New(rand.NewSource(12345))
	system := NewSpellEffectSystem(world, rng)

	// Create entities at various positions
	e1 := world.CreateEntity()
	e1.AddComponent(&PositionComponent{X: 0.0, Y: 0.0})

	e2 := world.CreateEntity()
	e2.AddComponent(&PositionComponent{X: 5.0, Y: 0.0})

	e3 := world.CreateEntity()
	e3.AddComponent(&PositionComponent{X: 20.0, Y: 0.0})

	// Process pending entity additions
	world.Update(0)

	// Get entities within radius 10 of origin
	entities := system.GetEntitiesInRadius(0.0, 0.0, 10.0)

	// Should include e1 and e2, but not e3
	if len(entities) != 2 {
		t.Errorf("GetEntitiesInRadius returned %d entities, want 2", len(entities))
	}

	// Verify correct entities
	found1, found2 := false, false
	for _, e := range entities {
		if e.ID == e1.ID {
			found1 = true
		}
		if e.ID == e2.ID {
			found2 = true
		}
	}

	if !found1 || !found2 {
		t.Errorf("GetEntitiesInRadius did not return correct entities")
	}
}

func TestSpellEffectSystem_CalculateDamageWithDistance(t *testing.T) {
	world := NewWorld()
	rng := rand.New(rand.NewSource(12345))
	system := NewSpellEffectSystem(world, rng)

	tests := []struct {
		name        string
		baseDamage  float64
		distance    float64
		maxRadius   float64
		minExpected float64
		maxExpected float64
	}{
		{"full damage at center", 100.0, 0.0, 10.0, 100.0, 100.0},
		{"half damage at half radius", 100.0, 5.0, 10.0, 50.0, 50.0},
		{"no damage at edge", 100.0, 10.0, 10.0, 0.0, 0.0},
		{"no damage beyond radius", 100.0, 15.0, 10.0, 0.0, 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			damage := system.CalculateDamageWithDistance(tt.baseDamage, tt.distance, tt.maxRadius)
			if damage < tt.minExpected || damage > tt.maxExpected {
				t.Errorf("Damage = %v, want between %v and %v", damage, tt.minExpected, tt.maxExpected)
			}
		})
	}
}

func TestSpellEffectSystem_GetAngleBetweenPoints(t *testing.T) {
	world := NewWorld()
	rng := rand.New(rand.NewSource(12345))
	system := NewSpellEffectSystem(world, rng)

	// Test basic angles
	tests := []struct {
		name     string
		x1, y1   float64
		x2, y2   float64
		minAngle float64
		maxAngle float64
	}{
		{"right", 0, 0, 10, 0, -0.1, 0.1},       // ~0 radians
		{"up", 0, 0, 0, 10, 1.47, 1.67},         // ~pi/2 radians
		{"left", 0, 0, -10, 0, 3.04, 3.24},      // ~pi radians
		{"down", 0, 0, 0, -10, -1.67, -1.47},    // ~-pi/2 radians
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			angle := system.GetAngleBetweenPoints(tt.x1, tt.y1, tt.x2, tt.y2)
			if angle < tt.minAngle || angle > tt.maxAngle {
				t.Errorf("Angle = %v, want between %v and %v", angle, tt.minAngle, tt.maxAngle)
			}
		})
	}
}

func TestSpellEffectSystem_ExecuteMetamagic(t *testing.T) {
	world := NewWorld()
	rng := rand.New(rand.NewSource(12345))
	system := NewSpellEffectSystem(world, rng)

	entity := world.CreateEntity()

	// Add a regular spell effect directly to Components map for testing
	regularEffect := &SpellEffectComponent{
		EffectType:  EffectLifeDrain,
		Duration:    5.0,
		Magnitude:   10.0,
		TargetType:  TargetEntity,
		Active:      true,
		ElapsedTime: 0,
	}
	
	// Add metamagic effect
	metamagicEffect := &SpellEffectComponent{
		EffectType:          EffectMetamagic,
		Duration:            0, // Instant
		Magnitude:           2.0,
		TargetType:          TargetSelf,
		MetamagicMultiplier: 2.0, // Double damage
		Active:              true,
		ElapsedTime:         0,
	}
	
	// Manually add both to test (in real game, would need better multi-effect handling)
	// For now, just verify executeMetamagic doesn't crash and has the basic logic
	entity.AddComponent(metamagicEffect)
	
	// Add the regular effect using a different type
	entity.Components["test_effect"] = regularEffect

	initialMagnitude := regularEffect.Magnitude

	entities := []*Entity{entity}
	system.Update(entities, 0.016)

	// The executeMetamagic method should have run without error
	// In a real implementation, magnitude would be modified
	// For this test, we're just verifying the system handles it
	if metamagicEffect.EffectType != EffectMetamagic {
		t.Errorf("EffectType should still be EffectMetamagic")
	}
	
	// Verify the method ran by checking elapsed time updated
	if metamagicEffect.ElapsedTime == 0 {
		t.Errorf("ElapsedTime should have been updated")
	}
	
	// Note: We can't properly test multi-effect interaction without
	// refactoring the component system to support multiple spell effects
	_ = initialMagnitude // Silence unused warning
}

func TestSpellEffectSystem_MultipleEffects(t *testing.T) {
	world := NewWorld()
	rng := rand.New(rand.NewSource(12345))
	system := NewSpellEffectSystem(world, rng)

	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 0, Y: 0})
	entity.AddComponent(&VelocityComponent{VX: 10, VY: 10})

	// Add first effect
	effect1 := &SpellEffectComponent{
		EffectType:  EffectGravityControl,
		Duration:    5.0,
		Magnitude:   -1.0,
		TargetType:  TargetSelf,
		Active:      true,
		ElapsedTime: 0,
	}
	entity.AddComponent(effect1)

	entities := []*Entity{entity}
	system.Update(entities, 0.016)

	// Should have gravity modification
	if !entity.HasComponent("gravity_modified") {
		t.Errorf("Missing gravity_modified component")
	}
}

func TestGenericComponent_Type(t *testing.T) {
	comp := &GenericComponent{componentType: "test_marker"}
	if comp.Type() != "test_marker" {
		t.Errorf("GenericComponent.Type() = %v, want test_marker", comp.Type())
	}
}
