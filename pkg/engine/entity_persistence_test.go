package engine

import (
	"testing"
)

// TestEntityLifecycleTracker tests the lifecycle tracking functionality
func TestEntityLifecycleTracker(t *testing.T) {
	tracker := NewEntityLifecycleTracker()

	// Test spawning
	tracker.MarkSpawned(1)
	tracker.MarkSpawned(2)

	// Test modification
	tracker.MarkModified(1)
	if !tracker.IsModified(1) {
		t.Error("Entity 1 should be marked as modified")
	}
	if tracker.IsModified(2) {
		t.Error("Entity 2 should not be marked as modified")
	}

	// Test killing
	tracker.MarkKilled(1)
	if !tracker.IsKilled(1) {
		t.Error("Entity 1 should be marked as killed")
	}
	if tracker.IsModified(1) {
		t.Error("Entity 1 should no longer be in modified list after being killed")
	}

	// Test GetModifiedEntities
	tracker.MarkModified(2)
	tracker.MarkModified(3)
	modifiedEntities := tracker.GetModifiedEntities()
	if len(modifiedEntities) != 2 { // 2 and 3
		t.Errorf("Expected 2 modified entities, got %d", len(modifiedEntities))
	}

	// Test ClearModified
	tracker.ClearModified()
	if tracker.IsModified(2) {
		t.Error("Entity 2 should not be modified after clearing")
	}
}

// TestGetRespawnRule tests respawn rule determination
func TestGetRespawnRule(t *testing.T) {
	tests := []struct {
		typeName    string
		wantRespawn RespawnRule
	}{
		{"Monster", RespawnAlways},
		{"NPC", RespawnNever},
		{"Merchant", RespawnNever},
		{"Companion", RespawnNever},
		{"Boss", RespawnConditional},
		{"Item", RespawnNever},
		{"Weapon", RespawnNever},
		{"Armor", RespawnNever},
		{"Consumable", RespawnNever},
		{"Unknown", RespawnNever}, // Default
	}

	for _, tt := range tests {
		t.Run(tt.typeName, func(t *testing.T) {
			got := GetRespawnRule(tt.typeName)
			if got != tt.wantRespawn {
				t.Errorf("GetRespawnRule(%s) = %v, want %v", tt.typeName, got, tt.wantRespawn)
			}
		})
	}
}

// TestSerializeDeserializeEntity tests entity serialization round-trip
func TestSerializeDeserializeEntity(t *testing.T) {
	world := NewWorld()

	// Create test entity with various components
	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 100.0, Y: 200.0})
	entity.AddComponent(&VelocityComponent{VX: 10.0, VY: 20.0})
	entity.AddComponent(&HealthComponent{Current: 75.0, Max: 100.0})
	entity.AddComponent(&ColliderComponent{
		Width:     32.0,
		Height:    64.0,
		Solid:     true,
		IsTrigger: false,
		Layer:     1,
		OffsetX:   -16.0,
		OffsetY:   -32.0,
	})

	// Serialize
	state, err := SerializeEntity(entity)
	if err != nil {
		t.Fatalf("Failed to serialize entity: %v", err)
	}

	// Verify state fields
	if state.ID != entity.ID {
		t.Errorf("State ID = %d, want %d", state.ID, entity.ID)
	}
	if len(state.Components) != 4 {
		t.Errorf("State has %d components, want 4", len(state.Components))
	}

	// Create new world and deserialize
	world2 := NewWorld()
	deserialized, err := DeserializeEntity(state, world2)
	if err != nil {
		t.Fatalf("Failed to deserialize entity: %v", err)
	}

	// Verify deserialized entity
	if deserialized.ID != entity.ID {
		t.Errorf("Deserialized ID = %d, want %d", deserialized.ID, entity.ID)
	}

	// Verify position component
	posComp, ok := deserialized.GetComponent("position")
	if !ok {
		t.Fatal("Missing position component")
	}
	pos := posComp.(*PositionComponent)
	if pos.X != 100.0 || pos.Y != 200.0 {
		t.Errorf("Position = (%f, %f), want (100, 200)", pos.X, pos.Y)
	}

	// Verify velocity component
	velComp, ok := deserialized.GetComponent("velocity")
	if !ok {
		t.Fatal("Missing velocity component")
	}
	vel := velComp.(*VelocityComponent)
	if vel.VX != 10.0 || vel.VY != 20.0 {
		t.Errorf("Velocity = (%f, %f), want (10, 20)", vel.VX, vel.VY)
	}

	// Verify health component
	healthComp, ok := deserialized.GetComponent("health")
	if !ok {
		t.Fatal("Missing health component")
	}
	health := healthComp.(*HealthComponent)
	if health.Current != 75.0 || health.Max != 100.0 {
		t.Errorf("Health = (%f/%f), want (75/100)", health.Current, health.Max)
	}

	// Verify collider component
	colliderComp, ok := deserialized.GetComponent("collider")
	if !ok {
		t.Fatal("Missing collider component")
	}
	collider := colliderComp.(*ColliderComponent)
	if collider.Width != 32.0 || collider.Height != 64.0 {
		t.Errorf("Collider size = (%f, %f), want (32, 64)", collider.Width, collider.Height)
	}
	if !collider.Solid {
		t.Error("Collider should be solid")
	}
	if collider.IsTrigger {
		t.Error("Collider should not be trigger")
	}
	if collider.Layer != 1 {
		t.Errorf("Collider layer = %d, want 1", collider.Layer)
	}
}

// TestSerializeNilEntity tests error handling for nil entity
func TestSerializeNilEntity(t *testing.T) {
	_, err := SerializeEntity(nil)
	if err == nil {
		t.Error("Expected error when serializing nil entity")
	}
}

// TestDeserializeNilState tests error handling for nil state
func TestDeserializeNilState(t *testing.T) {
	world := NewWorld()
	_, err := DeserializeEntity(nil, world)
	if err == nil {
		t.Error("Expected error when deserializing nil state")
	}
}

// TestDeserializeNilWorld tests error handling for nil world
func TestDeserializeNilWorld(t *testing.T) {
	state := &EntityState{
		ID:         1,
		TypeName:   "Monster",
		Components: make(map[string][]byte),
	}
	_, err := DeserializeEntity(state, nil)
	if err == nil {
		t.Error("Expected error when deserializing with nil world")
	}
}

// TestGetEntityTypeName tests entity type name extraction
func TestGetEntityTypeName(t *testing.T) {
	world := NewWorld()

	tests := []struct {
		name         string
		setupFunc    func() *Entity
		wantTypeName string
	}{
		{
			name: "Monster with hostile AI",
			setupFunc: func() *Entity {
				e := world.CreateEntity()
				e.AddComponent(&AIComponent{State: AIStateAttack})
				return e
			},
			wantTypeName: "Monster",
		},
		{
			name: "NPC with friendly AI",
			setupFunc: func() *Entity {
				e := world.CreateEntity()
				e.AddComponent(&AIComponent{State: AIStateIdle})
				return e
			},
			wantTypeName: "NPC",
		},
		{
			name: "Companion",
			setupFunc: func() *Entity {
				e := world.CreateEntity()
				e.AddComponent(&CompanionComponent{})
				return e
			},
			wantTypeName: "Companion",
		},
		{
			name: "Generic entity",
			setupFunc: func() *Entity {
				e := world.CreateEntity()
				e.AddComponent(&PositionComponent{})
				return e
			},
			wantTypeName: "Entity",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entity := tt.setupFunc()
			typeName := getEntityTypeName(entity)
			if typeName != tt.wantTypeName {
				t.Errorf("getEntityTypeName() = %s, want %s", typeName, tt.wantTypeName)
			}
		})
	}
}

// TestComponentSerializationRoundTrip tests serialization for individual components
func TestComponentSerializationRoundTrip(t *testing.T) {
	tests := []struct {
		name      string
		component ComponentSerializer
		validate  func(t *testing.T, original, deserialized ComponentSerializer)
	}{
		{
			name:      "PositionComponent",
			component: &PositionComponent{X: 123.456, Y: 789.012},
			validate: func(t *testing.T, original, deserialized ComponentSerializer) {
				orig := original.(*PositionComponent)
				deser := deserialized.(*PositionComponent)
				if orig.X != deser.X || orig.Y != deser.Y {
					t.Errorf("Position mismatch: got (%f, %f), want (%f, %f)", deser.X, deser.Y, orig.X, orig.Y)
				}
			},
		},
		{
			name:      "VelocityComponent",
			component: &VelocityComponent{VX: -45.0, VY: 90.0},
			validate: func(t *testing.T, original, deserialized ComponentSerializer) {
				orig := original.(*VelocityComponent)
				deser := deserialized.(*VelocityComponent)
				if orig.VX != deser.VX || orig.VY != deser.VY {
					t.Errorf("Velocity mismatch: got (%f, %f), want (%f, %f)", deser.VX, deser.VY, orig.VX, orig.VY)
				}
			},
		},
		{
			name:      "HealthComponent",
			component: &HealthComponent{Current: 50.5, Max: 100.0},
			validate: func(t *testing.T, original, deserialized ComponentSerializer) {
				orig := original.(*HealthComponent)
				deser := deserialized.(*HealthComponent)
				if orig.Current != deser.Current || orig.Max != deser.Max {
					t.Errorf("Health mismatch: got (%f/%f), want (%f/%f)", deser.Current, deser.Max, orig.Current, orig.Max)
				}
			},
		},
		{
			name: "ColliderComponent",
			component: &ColliderComponent{
				Width: 64.0, Height: 48.0,
				Solid: true, IsTrigger: false, Layer: 2,
				OffsetX: -32.0, OffsetY: -24.0,
			},
			validate: func(t *testing.T, original, deserialized ComponentSerializer) {
				orig := original.(*ColliderComponent)
				deser := deserialized.(*ColliderComponent)
				if orig.Width != deser.Width || orig.Height != deser.Height {
					t.Errorf("Collider size mismatch")
				}
				if orig.Solid != deser.Solid || orig.IsTrigger != deser.IsTrigger || orig.Layer != deser.Layer {
					t.Errorf("Collider flags mismatch")
				}
				if orig.OffsetX != deser.OffsetX || orig.OffsetY != deser.OffsetY {
					t.Errorf("Collider offset mismatch")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Serialize
			data, err := tt.component.Serialize()
			if err != nil {
				t.Fatalf("Failed to serialize: %v", err)
			}

			// Create new component of same type
			var deserialized ComponentSerializer
			switch tt.component.(type) {
			case *PositionComponent:
				deserialized = &PositionComponent{}
			case *VelocityComponent:
				deserialized = &VelocityComponent{}
			case *HealthComponent:
				deserialized = &HealthComponent{}
			case *ColliderComponent:
				deserialized = &ColliderComponent{}
			}

			// Deserialize
			err = deserialized.Deserialize(data)
			if err != nil {
				t.Fatalf("Failed to deserialize: %v", err)
			}

			// Validate
			tt.validate(t, tt.component, deserialized)
		})
	}
}

// TestInvalidComponentData tests error handling for invalid serialization data
func TestInvalidComponentData(t *testing.T) {
	tests := []struct {
		name      string
		component ComponentSerializer
		dataLen   int
	}{
		{"PositionComponent with short data", &PositionComponent{}, 8},  // Should be 16
		{"VelocityComponent with short data", &VelocityComponent{}, 8},  // Should be 16
		{"HealthComponent with short data", &HealthComponent{}, 8},      // Should be 16
		{"ColliderComponent with short data", &ColliderComponent{}, 16}, // Should be 38
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := make([]byte, tt.dataLen)
			err := tt.component.Deserialize(data)
			if err == nil {
				t.Error("Expected error when deserializing invalid data")
			}
			if err != ErrInvalidComponentData {
				t.Errorf("Expected ErrInvalidComponentData, got %v", err)
			}
		})
	}
}

// BenchmarkSerializeEntity benchmarks entity serialization
func BenchmarkSerializeEntity(b *testing.B) {
	world := NewWorld()
	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 100.0, Y: 200.0})
	entity.AddComponent(&VelocityComponent{VX: 10.0, VY: 20.0})
	entity.AddComponent(&HealthComponent{Current: 75.0, Max: 100.0})
	entity.AddComponent(&ColliderComponent{Width: 32.0, Height: 64.0, Solid: true})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := SerializeEntity(entity)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkDeserializeEntity benchmarks entity deserialization
func BenchmarkDeserializeEntity(b *testing.B) {
	world := NewWorld()
	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 100.0, Y: 200.0})
	entity.AddComponent(&VelocityComponent{VX: 10.0, VY: 20.0})
	entity.AddComponent(&HealthComponent{Current: 75.0, Max: 100.0})
	entity.AddComponent(&ColliderComponent{Width: 32.0, Height: 64.0, Solid: true})

	state, _ := SerializeEntity(entity)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		world2 := NewWorld()
		_, err := DeserializeEntity(state, world2)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkComponentSerialization benchmarks individual component serialization
func BenchmarkComponentSerialization(b *testing.B) {
	component := &PositionComponent{X: 123.456, Y: 789.012}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := component.Serialize()
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkComponentDeserialization benchmarks individual component deserialization
func BenchmarkComponentDeserialization(b *testing.B) {
	component := &PositionComponent{X: 123.456, Y: 789.012}
	data, _ := component.Serialize()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		dest := &PositionComponent{}
		err := dest.Deserialize(data)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// TestCreateComponentByType tests component creation by type
func TestCreateComponentByType(t *testing.T) {
	tests := []struct {
		componentType string
		wantErr       bool
	}{
		{"position", false},
		{"velocity", false},
		{"health", false},
		{"stats", false},
		{"base_stats", false},
		{"ai", false},
		{"inventory", false},
		{"experience", false},
		{"collider", false},
		{"animation", false},
		{"companion", false},
		{"vehicle", false},
		{"mount", false},
		{"unknown_type", true}, // Should error
	}

	for _, tt := range tests {
		t.Run(tt.componentType, func(t *testing.T) {
			component, err := createComponentByType(tt.componentType)
			if (err != nil) != tt.wantErr {
				t.Errorf("createComponentByType(%s) error = %v, wantErr %v", tt.componentType, err, tt.wantErr)
				return
			}
			if !tt.wantErr && component == nil {
				t.Errorf("createComponentByType(%s) returned nil component", tt.componentType)
			}
		})
	}
}
