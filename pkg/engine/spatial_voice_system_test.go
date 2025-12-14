package engine

import (
	"testing"
)

func TestNewSpatialVoiceSystem(t *testing.T) {
	world := NewWorld()
	system := NewSpatialVoiceSystem(world)

	if system == nil {
		t.Fatal("NewSpatialVoiceSystem returned nil")
	}
	if system.world != world {
		t.Error("world not set correctly")
	}
	if system.listenerEntityID != 0 {
		t.Errorf("expected listenerEntityID 0, got %d", system.listenerEntityID)
	}
}

func TestSpatialVoiceSystem_SetListener(t *testing.T) {
	world := NewWorld()
	system := NewSpatialVoiceSystem(world)

	entity := world.CreateEntity()
	system.SetListener(entity)

	if system.listenerEntityID != entity.ID {
		t.Errorf("expected listenerEntityID %d, got %d", entity.ID, system.listenerEntityID)
	}

	// Set nil listener
	system.SetListener(nil)
	if system.listenerEntityID != 0 {
		t.Errorf("expected listenerEntityID 0 after nil, got %d", system.listenerEntityID)
	}
}

func TestSpatialVoiceSystem_Update(t *testing.T) {
	world := NewWorld()
	system := NewSpatialVoiceSystem(world)

	// Create listener at origin
	listener := world.CreateEntity()
	listener.AddComponent(&PositionComponent{X: 0, Y: 0})
	system.SetListener(listener)

	// Create voice source at (100, 0)
	source := world.CreateEntity()
	source.AddComponent(&PositionComponent{X: 100, Y: 0})
	spatial := NewSpatialVoiceComponent()
	spatial.SetRange(0, 200)
	spatial.SetFalloffCurve(VoiceFalloffLinear)
	source.AddComponent(spatial)

	entities := []*Entity{listener, source}

	// Run update
	system.Update(entities, 0.016)

	// Check spatial calculations
	if spatial.CurrentDistance != 100.0 {
		t.Errorf("expected CurrentDistance 100.0, got %f", spatial.CurrentDistance)
	}
	if !spatial.IsAudible {
		t.Error("expected IsAudible to be true")
	}
	// Linear falloff: 100/200 = 50% distance, so volume = 0.5
	if spatial.CurrentVolume != 0.5 {
		t.Errorf("expected CurrentVolume 0.5, got %f", spatial.CurrentVolume)
	}
}

func TestSpatialVoiceSystem_UpdateNoListener(t *testing.T) {
	world := NewWorld()
	system := NewSpatialVoiceSystem(world)

	// No listener set
	source := world.CreateEntity()
	source.AddComponent(&PositionComponent{X: 100, Y: 0})
	spatial := NewSpatialVoiceComponent()
	source.AddComponent(spatial)

	entities := []*Entity{source}

	// Should not panic and not update
	system.Update(entities, 0.016)

	// Spatial should not be updated
	if spatial.CurrentDistance != 0 {
		t.Errorf("expected CurrentDistance 0 without listener, got %f", spatial.CurrentDistance)
	}
}

func TestSpatialVoiceSystem_GetVolumeForEntity(t *testing.T) {
	world := NewWorld()
	system := NewSpatialVoiceSystem(world)

	// Entity without spatial component
	entity1 := world.CreateEntity()
	vol := system.GetVolumeForEntity(entity1)
	if vol != 1.0 {
		t.Errorf("expected volume 1.0 without spatial component, got %f", vol)
	}

	// Entity with spatial component
	entity2 := world.CreateEntity()
	spatial := NewSpatialVoiceComponent()
	spatial.CurrentVolume = 0.75
	entity2.AddComponent(spatial)

	vol = system.GetVolumeForEntity(entity2)
	if vol != 0.75 {
		t.Errorf("expected volume 0.75, got %f", vol)
	}

	// Nil entity
	vol = system.GetVolumeForEntity(nil)
	if vol != 0 {
		t.Errorf("expected volume 0 for nil entity, got %f", vol)
	}
}

func TestSpatialVoiceSystem_GetPanForEntity(t *testing.T) {
	world := NewWorld()
	system := NewSpatialVoiceSystem(world)

	// Entity without spatial component
	entity1 := world.CreateEntity()
	pan := system.GetPanForEntity(entity1)
	if pan != 0 {
		t.Errorf("expected pan 0 without spatial component, got %f", pan)
	}

	// Entity with spatial component
	entity2 := world.CreateEntity()
	spatial := NewSpatialVoiceComponent()
	spatial.CurrentPan = 0.5
	entity2.AddComponent(spatial)

	pan = system.GetPanForEntity(entity2)
	if pan != 0.5 {
		t.Errorf("expected pan 0.5, got %f", pan)
	}

	// Nil entity
	pan = system.GetPanForEntity(nil)
	if pan != 0 {
		t.Errorf("expected pan 0 for nil entity, got %f", pan)
	}
}

func TestSpatialVoiceSystem_GetDistanceToEntity(t *testing.T) {
	world := NewWorld()
	system := NewSpatialVoiceSystem(world)

	// Entity with spatial component
	entity := world.CreateEntity()
	spatial := NewSpatialVoiceComponent()
	spatial.CurrentDistance = 150.0
	entity.AddComponent(spatial)

	dist := system.GetDistanceToEntity(entity)
	if dist != 150.0 {
		t.Errorf("expected distance 150.0, got %f", dist)
	}

	// Entity without spatial component
	entity2 := world.CreateEntity()
	dist = system.GetDistanceToEntity(entity2)
	if dist != 0 {
		t.Errorf("expected distance 0 without component, got %f", dist)
	}

	// Nil entity
	dist = system.GetDistanceToEntity(nil)
	if dist != 0 {
		t.Errorf("expected distance 0 for nil entity, got %f", dist)
	}
}

func TestSpatialVoiceSystem_IsEntityAudible(t *testing.T) {
	world := NewWorld()
	system := NewSpatialVoiceSystem(world)

	// Entity without spatial component (default audible)
	entity1 := world.CreateEntity()
	if !system.IsEntityAudible(entity1) {
		t.Error("expected entity without spatial component to be audible")
	}

	// Entity with spatial component, within range
	entity2 := world.CreateEntity()
	spatial := NewSpatialVoiceComponent()
	spatial.IsAudible = true
	entity2.AddComponent(spatial)

	if !system.IsEntityAudible(entity2) {
		t.Error("expected entity within range to be audible")
	}

	// Entity out of range
	spatial.IsAudible = false
	spatial.SetEnabled(true)
	if system.IsEntityAudible(entity2) {
		t.Error("expected entity out of range to not be audible")
	}

	// Nil entity
	if system.IsEntityAudible(nil) {
		t.Error("expected nil entity to not be audible")
	}
}

func TestSpatialVoiceSystem_SetEntityRange(t *testing.T) {
	world := NewWorld()
	system := NewSpatialVoiceSystem(world)

	entity := world.CreateEntity()
	spatial := NewSpatialVoiceComponent()
	entity.AddComponent(spatial)

	err := system.SetEntityRange(entity, 100, 500)
	if err != nil {
		t.Fatalf("SetEntityRange failed: %v", err)
	}

	if spatial.MinRange != 100 {
		t.Errorf("expected MinRange 100, got %f", spatial.MinRange)
	}
	if spatial.MaxRange != 500 {
		t.Errorf("expected MaxRange 500, got %f", spatial.MaxRange)
	}

	// Nil entity
	err = system.SetEntityRange(nil, 100, 500)
	if err != ErrNilEntity {
		t.Errorf("expected ErrNilEntity, got %v", err)
	}

	// No component
	entity2 := world.CreateEntity()
	err = system.SetEntityRange(entity2, 100, 500)
	if err != ErrNoSpatialComponent {
		t.Errorf("expected ErrNoSpatialComponent, got %v", err)
	}
}

func TestSpatialVoiceSystem_SetEntityFalloff(t *testing.T) {
	world := NewWorld()
	system := NewSpatialVoiceSystem(world)

	entity := world.CreateEntity()
	spatial := NewSpatialVoiceComponent()
	entity.AddComponent(spatial)

	err := system.SetEntityFalloff(entity, VoiceFalloffExponential)
	if err != nil {
		t.Fatalf("SetEntityFalloff failed: %v", err)
	}

	if spatial.FalloffCurve != VoiceFalloffExponential {
		t.Errorf("expected FalloffCurve exponential, got %s", spatial.FalloffCurve)
	}

	// Nil entity
	err = system.SetEntityFalloff(nil, VoiceFalloffLinear)
	if err != ErrNilEntity {
		t.Errorf("expected ErrNilEntity, got %v", err)
	}

	// No component
	entity2 := world.CreateEntity()
	err = system.SetEntityFalloff(entity2, VoiceFalloffLinear)
	if err != ErrNoSpatialComponent {
		t.Errorf("expected ErrNoSpatialComponent, got %v", err)
	}
}

func TestSpatialVoiceSystem_EnableSpatialAudio(t *testing.T) {
	world := NewWorld()
	system := NewSpatialVoiceSystem(world)

	entity := world.CreateEntity()
	spatial := NewSpatialVoiceComponent()
	entity.AddComponent(spatial)

	err := system.EnableSpatialAudio(entity, false)
	if err != nil {
		t.Fatalf("EnableSpatialAudio failed: %v", err)
	}

	if spatial.Enabled {
		t.Error("expected Enabled to be false")
	}

	err = system.EnableSpatialAudio(entity, true)
	if err != nil {
		t.Fatalf("EnableSpatialAudio failed: %v", err)
	}

	if !spatial.Enabled {
		t.Error("expected Enabled to be true")
	}

	// Nil entity
	err = system.EnableSpatialAudio(nil, true)
	if err != ErrNilEntity {
		t.Errorf("expected ErrNilEntity, got %v", err)
	}

	// No component
	entity2 := world.CreateEntity()
	err = system.EnableSpatialAudio(entity2, true)
	if err != ErrNoSpatialComponent {
		t.Errorf("expected ErrNoSpatialComponent, got %v", err)
	}
}

func TestSpatialVoiceSystem_GetAudibleEntities(t *testing.T) {
	world := NewWorld()
	system := NewSpatialVoiceSystem(world)

	// Create listener
	listener := world.CreateEntity()
	listener.AddComponent(&PositionComponent{X: 0, Y: 0})
	system.SetListener(listener)

	// Create entities with different ranges
	entity1 := world.CreateEntity()
	entity1.AddComponent(&PositionComponent{X: 100, Y: 0})
	spatial1 := NewSpatialVoiceComponent()
	spatial1.SetRange(0, 200)
	spatial1.IsAudible = true
	entity1.AddComponent(spatial1)

	entity2 := world.CreateEntity()
	entity2.AddComponent(&PositionComponent{X: 300, Y: 0})
	spatial2 := NewSpatialVoiceComponent()
	spatial2.SetRange(0, 100)
	spatial2.IsAudible = false // Out of range
	entity2.AddComponent(spatial2)

	entity3 := world.CreateEntity()
	entity3.AddComponent(&PositionComponent{X: 50, Y: 0})
	spatial3 := NewSpatialVoiceComponent()
	spatial3.IsAudible = true
	entity3.AddComponent(spatial3)

	entities := []*Entity{listener, entity1, entity2, entity3}

	audible := system.GetAudibleEntities(entities)

	// Should include entity1 and entity3, but not listener or entity2
	if len(audible) != 2 {
		t.Errorf("expected 2 audible entities, got %d", len(audible))
	}
}

func TestSpatialVoiceSystem_SkipsListener(t *testing.T) {
	world := NewWorld()
	system := NewSpatialVoiceSystem(world)

	// Create listener with spatial component
	listener := world.CreateEntity()
	listener.AddComponent(&PositionComponent{X: 0, Y: 0})
	listenerSpatial := NewSpatialVoiceComponent()
	listener.AddComponent(listenerSpatial)
	system.SetListener(listener)

	entities := []*Entity{listener}

	// Update should skip the listener
	system.Update(entities, 0.016)

	// Listener's spatial component should not be updated
	if listenerSpatial.CurrentDistance != 0 {
		t.Error("expected listener's spatial to not be updated")
	}
}

func TestSpatialVoiceSystem_EntityWithoutPosition(t *testing.T) {
	world := NewWorld()
	system := NewSpatialVoiceSystem(world)

	// Create listener
	listener := world.CreateEntity()
	listener.AddComponent(&PositionComponent{X: 0, Y: 0})
	system.SetListener(listener)

	// Create entity without position component
	entity := world.CreateEntity()
	spatial := NewSpatialVoiceComponent()
	entity.AddComponent(spatial)

	entities := []*Entity{listener, entity}

	// Update should not panic
	system.Update(entities, 0.016)

	// Spatial should not be updated
	if spatial.CurrentDistance != 0 {
		t.Error("expected spatial to not be updated without position")
	}
}
