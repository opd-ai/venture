package engine

import (
	"testing"
)

func TestDynamicExpressionSystem_NewCreation(t *testing.T) {
	world := NewWorld()
	sys := NewDynamicExpressionSystem(world, 12345)
	if sys == nil {
		t.Fatal("Expected non-nil system")
	}
	if sys.updateInterval != 0.5 {
		t.Errorf("Expected update interval 0.5, got %f", sys.updateInterval)
	}
}

func TestDynamicExpressionSystem_SkipsEntitiesWithoutComponent(t *testing.T) {
	world := NewWorld()
	sys := NewDynamicExpressionSystem(world, 42)

	entity := NewEntity(1)
	entity.AddComponent(&HealthComponent{Current: 10, Max: 100})

	entities := []*Entity{entity}
	// Should not panic when entity lacks npc_facial_detail
	sys.Update(entities, 1.0)
}

func TestDynamicExpressionSystem_LowHealthScared(t *testing.T) {
	world := NewWorld()
	sys := NewDynamicExpressionSystem(world, 42)

	entity := NewEntity(1)
	faceComp := NewNpcFacialDetailComponent()
	faceComp.ExpressionType = "neutral"
	entity.AddComponent(faceComp)
	entity.AddComponent(&HealthComponent{Current: 10, Max: 100})

	entities := []*Entity{entity}
	sys.Update(entities, 1.0)

	if faceComp.ExpressionType != "scared" {
		t.Errorf("Expected 'scared' for low health, got %q", faceComp.ExpressionType)
	}
}

func TestDynamicExpressionSystem_DeadScared(t *testing.T) {
	world := NewWorld()
	sys := NewDynamicExpressionSystem(world, 42)

	entity := NewEntity(1)
	faceComp := NewNpcFacialDetailComponent()
	entity.AddComponent(faceComp)
	entity.AddComponent(&HealthComponent{Current: 0, Max: 100})

	entities := []*Entity{entity}
	sys.Update(entities, 1.0)

	if faceComp.ExpressionType != "scared" {
		t.Errorf("Expected 'scared' for dead entity, got %q", faceComp.ExpressionType)
	}
}

func TestDynamicExpressionSystem_StatusEffectPoison(t *testing.T) {
	world := NewWorld()
	sys := NewDynamicExpressionSystem(world, 42)

	entity := NewEntity(1)
	faceComp := NewNpcFacialDetailComponent()
	faceComp.ExpressionType = "neutral"
	entity.AddComponent(faceComp)
	entity.AddComponent(&HealthComponent{Current: 80, Max: 100})
	entity.AddComponent(&StatusEffectComponent{EffectType: "poison", Duration: 5.0})

	entities := []*Entity{entity}
	sys.Update(entities, 1.0)

	if faceComp.ExpressionType != "scared" {
		t.Errorf("Expected 'scared' for poisoned entity, got %q", faceComp.ExpressionType)
	}
}

func TestDynamicExpressionSystem_StatusEffectRage(t *testing.T) {
	world := NewWorld()
	sys := NewDynamicExpressionSystem(world, 42)

	entity := NewEntity(1)
	faceComp := NewNpcFacialDetailComponent()
	faceComp.ExpressionType = "neutral"
	entity.AddComponent(faceComp)
	entity.AddComponent(&HealthComponent{Current: 80, Max: 100})
	entity.AddComponent(&StatusEffectComponent{EffectType: "rage", Duration: 5.0})

	entities := []*Entity{entity}
	sys.Update(entities, 1.0)

	if faceComp.ExpressionType != "hostile" {
		t.Errorf("Expected 'hostile' for enraged entity, got %q", faceComp.ExpressionType)
	}
}

func TestDynamicExpressionSystem_AIStateAttack(t *testing.T) {
	world := NewWorld()
	sys := NewDynamicExpressionSystem(world, 42)

	entity := NewEntity(1)
	faceComp := NewNpcFacialDetailComponent()
	faceComp.ExpressionType = "neutral"
	entity.AddComponent(faceComp)
	entity.AddComponent(&HealthComponent{Current: 80, Max: 100})
	entity.AddComponent(&AIComponent{State: AIStateAttack})

	entities := []*Entity{entity}
	sys.Update(entities, 1.0)

	if faceComp.ExpressionType != "hostile" {
		t.Errorf("Expected 'hostile' for attacking entity, got %q", faceComp.ExpressionType)
	}
}

func TestDynamicExpressionSystem_AIStateFlee(t *testing.T) {
	world := NewWorld()
	sys := NewDynamicExpressionSystem(world, 42)

	entity := NewEntity(1)
	faceComp := NewNpcFacialDetailComponent()
	faceComp.ExpressionType = "neutral"
	entity.AddComponent(faceComp)
	entity.AddComponent(&HealthComponent{Current: 80, Max: 100})
	entity.AddComponent(&AIComponent{State: AIStateFlee})

	entities := []*Entity{entity}
	sys.Update(entities, 1.0)

	if faceComp.ExpressionType != "scared" {
		t.Errorf("Expected 'scared' for fleeing entity, got %q", faceComp.ExpressionType)
	}
}

func TestDynamicExpressionSystem_NeutralWhenHealthy(t *testing.T) {
	world := NewWorld()
	sys := NewDynamicExpressionSystem(world, 42)

	entity := NewEntity(1)
	faceComp := NewNpcFacialDetailComponent()
	faceComp.ExpressionType = "hostile" // Start non-neutral
	entity.AddComponent(faceComp)
	entity.AddComponent(&HealthComponent{Current: 90, Max: 100})
	entity.AddComponent(&AIComponent{State: AIStateIdle})

	entities := []*Entity{entity}
	sys.Update(entities, 1.0)

	if faceComp.ExpressionType != "neutral" {
		t.Errorf("Expected 'neutral' for healthy idle entity, got %q", faceComp.ExpressionType)
	}
}

func TestDynamicExpressionSystem_ThrottledUpdate(t *testing.T) {
	world := NewWorld()
	sys := NewDynamicExpressionSystem(world, 42)

	entity := NewEntity(1)
	faceComp := NewNpcFacialDetailComponent()
	faceComp.ExpressionType = "neutral"
	entity.AddComponent(faceComp)
	entity.AddComponent(&HealthComponent{Current: 10, Max: 100})

	entities := []*Entity{entity}

	// First call with small deltaTime should not update (throttled)
	sys.Update(entities, 0.1)
	if faceComp.ExpressionType != "neutral" {
		t.Errorf("Should not update within throttle interval")
	}

	// Call with enough deltaTime to exceed interval
	sys.Update(entities, 0.5)
	if faceComp.ExpressionType != "scared" {
		t.Errorf("Should update after throttle interval, got %q", faceComp.ExpressionType)
	}
}

func TestDynamicExpressionSystem_MarksDirty(t *testing.T) {
	world := NewWorld()
	sys := NewDynamicExpressionSystem(world, 42)

	entity := NewEntity(1)
	faceComp := NewNpcFacialDetailComponent()
	faceComp.ExpressionType = "neutral"
	entity.AddComponent(faceComp)
	entity.AddComponent(&HealthComponent{Current: 10, Max: 100})
	anim := NewAnimationComponent(999)
	anim.Dirty = false
	entity.AddComponent(anim)

	entities := []*Entity{entity}
	sys.Update(entities, 1.0)

	if !anim.Dirty {
		t.Error("Expected animation to be marked dirty after expression change")
	}
}

func TestDynamicExpressionSystem_NoChangeNoDirty(t *testing.T) {
	world := NewWorld()
	sys := NewDynamicExpressionSystem(world, 42)

	entity := NewEntity(1)
	faceComp := NewNpcFacialDetailComponent()
	faceComp.ExpressionType = "neutral"
	entity.AddComponent(faceComp)
	entity.AddComponent(&HealthComponent{Current: 90, Max: 100})
	entity.AddComponent(&AIComponent{State: AIStateIdle})
	anim := NewAnimationComponent(999)
	anim.Dirty = false
	entity.AddComponent(anim)

	entities := []*Entity{entity}
	sys.Update(entities, 1.0)

	// Expression should stay neutral, dirty should not be set
	if anim.Dirty {
		t.Error("Animation should not be marked dirty when expression doesn't change")
	}
}

func TestDynamicExpressionSystem_PriorityOrder(t *testing.T) {
	tests := []struct {
		name     string
		health   float64
		maxHP    float64
		effect   string
		aiState  AIState
		expected string
	}{
		{"dead_overrides_all", 0, 100, "rage", AIStateAttack, "scared"},
		{"low_hp_overrides_status", 20, 100, "heal", AIStateAttack, "scared"},
		{"status_overrides_ai", 80, 100, "poison", AIStateAttack, "scared"},
		{"ai_when_no_status", 80, 100, "", AIStateAttack, "hostile"},
		{"neutral_default", 80, 100, "", AIStateIdle, "neutral"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			world := NewWorld()
			sys := NewDynamicExpressionSystem(world, 42)

			entity := NewEntity(1)
			faceComp := NewNpcFacialDetailComponent()
			faceComp.ExpressionType = "friendly" // Start with non-matching
			entity.AddComponent(faceComp)
			entity.AddComponent(&HealthComponent{Current: tt.health, Max: tt.maxHP})
			if tt.effect != "" {
				entity.AddComponent(&StatusEffectComponent{EffectType: tt.effect, Duration: 5.0})
			}
			entity.AddComponent(&AIComponent{State: tt.aiState})

			entities := []*Entity{entity}
			sys.Update(entities, 1.0)

			if faceComp.ExpressionType != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, faceComp.ExpressionType)
			}
		})
	}
}
