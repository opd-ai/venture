package engine

import (
	"testing"
)

func TestNewStatusEffectAISystem(t *testing.T) {
	world := NewWorld()
	system := NewStatusEffectAISystem(world, 12345)

	if system == nil {
		t.Fatal("NewStatusEffectAISystem returned nil")
	}

	if system.world != world {
		t.Error("System world not set correctly")
	}

	if system.rng == nil {
		t.Error("System RNG not initialized")
	}

	if len(system.controlEffects) == 0 {
		t.Error("Control effects not initialized")
	}

	// Verify standard control effects are registered
	expectedEffects := []string{
		"stun", "stunned", "frozen", "freeze", "fear", "feared",
		"paralyze", "paralyzed", "sleep", "asleep", "charm", "charmed",
	}
	for _, effect := range expectedEffects {
		if !system.controlEffects[effect] {
			t.Errorf("Expected control effect '%s' not registered", effect)
		}
	}
}

func TestStatusEffectAISystem_Update_DisablesStunnedAI(t *testing.T) {
	world := NewWorld()
	system := NewStatusEffectAISystem(world, 12345)

	// Create entity with AI and stun effect
	entity := world.CreateEntity()
	aiComp := NewAIComponent(100, 100)
	aiComp.State = AIStateChase
	entity.AddComponent(aiComp)
	entity.AddComponent(&StatusEffectComponent{
		EffectType: "stun",
		Duration:   5.0,
		Magnitude:  1.0,
	})

	entities := []*Entity{entity}
	system.Update(entities, 1.0/60.0)

	if !aiComp.Disabled {
		t.Error("AI should be disabled when stunned")
	}
	if aiComp.DisabledReason != "stun" {
		t.Errorf("DisabledReason = %s, want 'stun'", aiComp.DisabledReason)
	}
	if aiComp.PreviousState != AIStateChase {
		t.Errorf("PreviousState = %v, want AIStateChase", aiComp.PreviousState)
	}
}

func TestStatusEffectAISystem_Update_DisablesFrozenAI(t *testing.T) {
	world := NewWorld()
	system := NewStatusEffectAISystem(world, 12345)

	entity := world.CreateEntity()
	aiComp := NewAIComponent(100, 100)
	aiComp.State = AIStateAttack
	entity.AddComponent(aiComp)
	entity.AddComponent(&StatusEffectComponent{
		EffectType: "frozen",
		Duration:   3.0,
		Magnitude:  1.0,
	})

	entities := []*Entity{entity}
	system.Update(entities, 1.0/60.0)

	if !aiComp.Disabled {
		t.Error("AI should be disabled when frozen")
	}
	if aiComp.DisabledReason != "frozen" {
		t.Errorf("DisabledReason = %s, want 'frozen'", aiComp.DisabledReason)
	}
}

func TestStatusEffectAISystem_Update_FearForcesFlee(t *testing.T) {
	world := NewWorld()
	system := NewStatusEffectAISystem(world, 12345)

	entity := world.CreateEntity()
	aiComp := NewAIComponent(100, 100)
	aiComp.State = AIStateAttack
	entity.AddComponent(aiComp)
	entity.AddComponent(&StatusEffectComponent{
		EffectType: "fear",
		Duration:   2.0,
		Magnitude:  1.0,
	})

	entities := []*Entity{entity}
	system.Update(entities, 1.0/60.0)

	// Fear should force flee state, not disable
	if aiComp.State != AIStateFlee {
		t.Errorf("State = %v, want AIStateFlee when feared", aiComp.State)
	}
}

func TestStatusEffectAISystem_Update_CharmPreventsAttack(t *testing.T) {
	world := NewWorld()
	system := NewStatusEffectAISystem(world, 12345)

	entity := world.CreateEntity()
	aiComp := NewAIComponent(100, 100)
	aiComp.State = AIStateAttack
	aiComp.Target = world.CreateEntity() // Dummy target
	entity.AddComponent(aiComp)
	entity.AddComponent(&StatusEffectComponent{
		EffectType: "charm",
		Duration:   4.0,
		Magnitude:  1.0,
	})

	entities := []*Entity{entity}
	system.Update(entities, 1.0/60.0)

	if !aiComp.Disabled {
		t.Error("AI should be disabled (for attacks) when charmed")
	}
	if aiComp.State != AIStateIdle {
		t.Errorf("State = %v, want AIStateIdle when charmed from attack", aiComp.State)
	}
	if aiComp.Target != nil {
		t.Error("Target should be cleared when charmed")
	}
}

func TestStatusEffectAISystem_Update_RestoresAfterExpiry(t *testing.T) {
	world := NewWorld()
	system := NewStatusEffectAISystem(world, 12345)

	entity := world.CreateEntity()
	aiComp := NewAIComponent(100, 100)
	aiComp.State = AIStateChase
	aiComp.Disabled = true
	aiComp.DisabledReason = "stun"
	aiComp.PreviousState = AIStateChase
	entity.AddComponent(aiComp)
	// No status effect - simulates effect expired

	entities := []*Entity{entity}
	system.Update(entities, 1.0/60.0)

	if aiComp.Disabled {
		t.Error("AI should be re-enabled after control effect expires")
	}
	if aiComp.DisabledReason != "" {
		t.Errorf("DisabledReason = %s, want empty string", aiComp.DisabledReason)
	}
}

func TestStatusEffectAISystem_Update_IgnoresExpiredEffects(t *testing.T) {
	world := NewWorld()
	system := NewStatusEffectAISystem(world, 12345)

	entity := world.CreateEntity()
	aiComp := NewAIComponent(100, 100)
	aiComp.State = AIStatePatrol
	entity.AddComponent(aiComp)
	entity.AddComponent(&StatusEffectComponent{
		EffectType: "stun",
		Duration:   -1.0, // Expired
		Magnitude:  1.0,
	})

	entities := []*Entity{entity}
	system.Update(entities, 1.0/60.0)

	if aiComp.Disabled {
		t.Error("AI should not be disabled by expired status effect")
	}
}

func TestStatusEffectAISystem_Update_IgnoresNonControlEffects(t *testing.T) {
	world := NewWorld()
	system := NewStatusEffectAISystem(world, 12345)

	entity := world.CreateEntity()
	aiComp := NewAIComponent(100, 100)
	aiComp.State = AIStatePatrol
	entity.AddComponent(aiComp)
	entity.AddComponent(&StatusEffectComponent{
		EffectType: "poison", // Not a control effect
		Duration:   5.0,
		Magnitude:  10.0,
	})

	entities := []*Entity{entity}
	system.Update(entities, 1.0/60.0)

	if aiComp.Disabled {
		t.Error("AI should not be disabled by non-control effects like poison")
	}
}

func TestStatusEffectAISystem_Update_SkipsEntitiesWithoutAI(t *testing.T) {
	world := NewWorld()
	system := NewStatusEffectAISystem(world, 12345)

	// Entity without AI component
	entity := world.CreateEntity()
	entity.AddComponent(&StatusEffectComponent{
		EffectType: "stun",
		Duration:   5.0,
	})

	entities := []*Entity{entity}
	// Should not panic
	system.Update(entities, 1.0/60.0)
}

func TestStatusEffectAISystem_RegisterControlEffect(t *testing.T) {
	world := NewWorld()
	system := NewStatusEffectAISystem(world, 12345)

	// Register custom control effect
	system.RegisterControlEffect("petrify")

	if !system.controlEffects["petrify"] {
		t.Error("Custom control effect 'petrify' not registered")
	}

	// Test it works
	entity := world.CreateEntity()
	aiComp := NewAIComponent(100, 100)
	entity.AddComponent(aiComp)
	entity.AddComponent(&StatusEffectComponent{
		EffectType: "petrify",
		Duration:   5.0,
	})

	entities := []*Entity{entity}
	system.Update(entities, 1.0/60.0)

	if !aiComp.Disabled {
		t.Error("AI should be disabled by custom 'petrify' effect")
	}
}

func TestStatusEffectAISystem_UnregisterControlEffect(t *testing.T) {
	world := NewWorld()
	system := NewStatusEffectAISystem(world, 12345)

	// Unregister stun
	system.UnregisterControlEffect("stun")

	if system.controlEffects["stun"] {
		t.Error("Control effect 'stun' should be unregistered")
	}

	// Verify stun no longer disables AI
	entity := world.CreateEntity()
	aiComp := NewAIComponent(100, 100)
	entity.AddComponent(aiComp)
	entity.AddComponent(&StatusEffectComponent{
		EffectType: "stun",
		Duration:   5.0,
	})

	entities := []*Entity{entity}
	system.Update(entities, 1.0/60.0)

	if aiComp.Disabled {
		t.Error("AI should not be disabled after unregistering 'stun'")
	}
}

func TestStatusEffectAISystem_IsEntityControlled(t *testing.T) {
	world := NewWorld()
	system := NewStatusEffectAISystem(world, 12345)

	entity := world.CreateEntity()
	aiComp := NewAIComponent(100, 100)
	entity.AddComponent(aiComp)
	entity.AddComponent(&StatusEffectComponent{
		EffectType: "stun",
		Duration:   5.0,
	})

	entities := []*Entity{entity}
	system.Update(entities, 1.0/60.0)

	if !system.IsEntityControlled(entity.ID) {
		t.Error("IsEntityControlled should return true for stunned entity")
	}

	// Entity without control effect
	entity2 := world.CreateEntity()
	entity2.AddComponent(NewAIComponent(200, 200))

	entities = append(entities, entity2)
	system.Update(entities, 1.0/60.0)

	if system.IsEntityControlled(entity2.ID) {
		t.Error("IsEntityControlled should return false for uncontrolled entity")
	}
}

func TestStatusEffectAISystem_MultipleControlEffects(t *testing.T) {
	world := NewWorld()
	system := NewStatusEffectAISystem(world, 12345)

	entity := world.CreateEntity()
	aiComp := NewAIComponent(100, 100)
	aiComp.State = AIStateChase
	entity.AddComponent(aiComp)
	// Add multiple control effects
	entity.AddComponent(&StatusEffectComponent{
		EffectType: "stun",
		Duration:   5.0,
	})
	entity.AddComponent(&StatusEffectComponent{
		EffectType: "frozen",
		Duration:   3.0,
	})

	entities := []*Entity{entity}
	system.Update(entities, 1.0/60.0)

	// Should be disabled by first found control effect
	if !aiComp.Disabled {
		t.Error("AI should be disabled with multiple control effects")
	}
}

func BenchmarkStatusEffectAISystem_Update(b *testing.B) {
	world := NewWorld()
	system := NewStatusEffectAISystem(world, 12345)

	// Create 100 entities with AI, 20% have control effects
	entities := make([]*Entity, 100)
	for i := 0; i < 100; i++ {
		entity := world.CreateEntity()
		entity.AddComponent(NewAIComponent(float64(i*10), float64(i*10)))
		if i%5 == 0 { // 20% have stun
			entity.AddComponent(&StatusEffectComponent{
				EffectType: "stun",
				Duration:   5.0,
			})
		}
		entities[i] = entity
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		system.Update(entities, 1.0/60.0)
	}
}

func BenchmarkStatusEffectAISystem_NoEffects(b *testing.B) {
	world := NewWorld()
	system := NewStatusEffectAISystem(world, 12345)

	// Create 100 entities with AI, no control effects
	entities := make([]*Entity, 100)
	for i := 0; i < 100; i++ {
		entity := world.CreateEntity()
		entity.AddComponent(NewAIComponent(float64(i*10), float64(i*10)))
		entities[i] = entity
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		system.Update(entities, 1.0/60.0)
	}
}
