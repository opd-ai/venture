package engine

import (
	"testing"
	"time"
)

func TestEventDecorationSystem_NewEventDecorationSystem(t *testing.T) {
	world := NewWorld()
	system := NewEventDecorationSystem(world)

	if system == nil {
		t.Fatal("Expected non-nil system")
	}
	if system.world != world {
		t.Error("Expected world reference to be set")
	}
}

func TestEventDecorationSystem_Update_NoEvents(t *testing.T) {
	world := NewWorld()
	system := NewEventDecorationSystem(world)

	// Create entity with decoration component
	entity := world.CreateEntity()
	decoComp := NewEventDecorationComponent(12345)
	entity.AddComponent(decoComp)

	// Update with no events should not add decorations
	entities := []*Entity{entity}
	system.Update(entities, 1.0/60.0)

	if decoComp.HasDecorations() {
		t.Error("Should have no decorations without active events")
	}
}

func TestEventDecorationSystem_Update_WithActiveEvent(t *testing.T) {
	world := NewWorld()
	system := NewEventDecorationSystem(world)

	// Create world entity with seasonal events
	worldEntity := world.CreateEntity()
	eventComp := NewSeasonalEventComponent(12345, false)
	eventComp.ActiveEvents = []EventInstance{
		{
			Definition: EventDefinition{
				ID:    "spring_festival",
				Name:  "Spring Festival",
				Theme: EventThemeSpring,
			},
			StartTime: time.Now().Add(-time.Hour),
			EndTime:   time.Now().Add(7 * 24 * time.Hour),
			Phase:     EventPhaseActive,
		},
	}
	worldEntity.AddComponent(eventComp)

	// Create entity with decoration component
	entity := world.CreateEntity()
	decoComp := NewEventDecorationComponent(12345)
	entity.AddComponent(decoComp)

	entities := []*Entity{worldEntity, entity}

	// First update should start applying decorations
	system.Update(entities, 1.0/60.0)

	if decoComp.EventID != "spring_festival" {
		t.Errorf("Expected event ID 'spring_festival', got '%s'", decoComp.EventID)
	}
	if decoComp.ActiveTheme != DecorationThemeSpring {
		t.Errorf("Expected spring theme, got '%s'", decoComp.ActiveTheme)
	}
	if !decoComp.IsTransitioning {
		t.Error("Expected to be transitioning")
	}
	if decoComp.TransitionDirection != 1 {
		t.Errorf("Expected transition direction 1, got %d", decoComp.TransitionDirection)
	}
}

func TestEventDecorationSystem_Update_TransitionProgress(t *testing.T) {
	world := NewWorld()
	system := NewEventDecorationSystem(world)

	// Create world entity with seasonal events
	worldEntity := world.CreateEntity()
	eventComp := NewSeasonalEventComponent(12345, false)
	eventComp.ActiveEvents = []EventInstance{
		{
			Definition: EventDefinition{
				ID:    "summer_solstice",
				Theme: EventThemeSummer,
			},
			Phase: EventPhaseActive,
		},
	}
	worldEntity.AddComponent(eventComp)

	// Create entity with decoration component
	entity := world.CreateEntity()
	decoComp := NewEventDecorationComponent(12345)
	entity.AddComponent(decoComp)

	entities := []*Entity{worldEntity, entity}

	// First update starts transition
	system.Update(entities, 1.0/60.0)
	initialProgress := decoComp.TransitionProgress

	// Second update should advance progress
	system.Update(entities, 1.0/60.0)
	if decoComp.TransitionProgress <= initialProgress {
		t.Error("Transition progress should increase")
	}

	// Many updates should complete transition
	for i := 0; i < 200; i++ {
		system.Update(entities, 1.0/60.0)
	}

	if decoComp.IsTransitioning {
		t.Error("Transition should be complete after many updates")
	}
	if decoComp.TransitionProgress != 1.0 {
		t.Errorf("Expected progress 1.0, got %f", decoComp.TransitionProgress)
	}
}

func TestEventDecorationSystem_Update_RemovesDecorationsWhenEventEnds(t *testing.T) {
	world := NewWorld()
	system := NewEventDecorationSystem(world)

	// Create world entity with seasonal events
	worldEntity := world.CreateEntity()
	eventComp := NewSeasonalEventComponent(12345, false)
	worldEntity.AddComponent(eventComp)

	// Create entity with pre-applied decorations
	entity := world.CreateEntity()
	decoComp := NewEventDecorationComponent(12345)
	decoComp.GenerateDecorations(DecorationThemeWinter, 1.0)
	decoComp.EventID = "winter_celebration"
	decoComp.TransitionProgress = 1.0
	entity.AddComponent(decoComp)

	entities := []*Entity{worldEntity, entity}

	// Update with no active events should start removal
	system.Update(entities, 1.0/60.0)

	if !decoComp.IsTransitioning {
		t.Error("Should be transitioning to remove decorations")
	}
	if decoComp.TransitionDirection != -1 {
		t.Errorf("Expected transition direction -1, got %d", decoComp.TransitionDirection)
	}

	// Many updates should complete removal
	for i := 0; i < 200; i++ {
		system.Update(entities, 1.0/60.0)
	}

	if decoComp.HasDecorations() {
		t.Error("Decorations should be removed after event ends")
	}
}

func TestEventDecorationSystem_CalculateDecorationLevel(t *testing.T) {
	system := NewEventDecorationSystem(nil)

	tests := []struct {
		phase    EventPhase
		expected float64
	}{
		{EventPhaseUpcoming, 0.3},
		{EventPhaseActive, 1.0},
		{EventPhaseEnding, 0.6},
		{EventPhase("unknown"), 0.0},
	}

	for _, tt := range tests {
		t.Run(string(tt.phase), func(t *testing.T) {
			event := &EventInstance{Phase: tt.phase}
			result := system.calculateDecorationLevel(event)
			if result != tt.expected {
				t.Errorf("calculateDecorationLevel(%s) = %f, expected %f", tt.phase, result, tt.expected)
			}
		})
	}
}

func TestEventDecorationSystem_ApplyDecorationsToEntity(t *testing.T) {
	world := NewWorld()
	system := NewEventDecorationSystem(world)

	entity := world.CreateEntity()
	decoComp := NewEventDecorationComponent(12345)
	entity.AddComponent(decoComp)

	err := system.ApplyDecorationsToEntity(entity, DecorationThemeAutumn, 0.8)
	if err != nil {
		t.Fatalf("ApplyDecorationsToEntity failed: %v", err)
	}

	if decoComp.ActiveTheme != DecorationThemeAutumn {
		t.Errorf("Expected autumn theme, got '%s'", decoComp.ActiveTheme)
	}
	if decoComp.TransitionProgress != 1.0 {
		t.Errorf("Expected transition complete, got %f", decoComp.TransitionProgress)
	}
	if decoComp.IsTransitioning {
		t.Error("Should not be transitioning after manual apply")
	}
}

func TestEventDecorationSystem_RemoveDecorationsFromEntity(t *testing.T) {
	world := NewWorld()
	system := NewEventDecorationSystem(world)

	entity := world.CreateEntity()
	decoComp := NewEventDecorationComponent(12345)
	decoComp.GenerateDecorations(DecorationThemeSpring, 1.0)
	decoComp.TransitionProgress = 1.0
	entity.AddComponent(decoComp)

	err := system.RemoveDecorationsFromEntity(entity)
	if err != nil {
		t.Fatalf("RemoveDecorationsFromEntity failed: %v", err)
	}

	if decoComp.HasDecorations() {
		t.Error("Decorations should be removed")
	}
}

func TestEventDecorationSystem_GetDecoratedEntities(t *testing.T) {
	world := NewWorld()
	system := NewEventDecorationSystem(world)

	// Create mix of decorated and non-decorated entities
	decorated1 := world.CreateEntity()
	dc1 := NewEventDecorationComponent(111)
	dc1.GenerateDecorations(DecorationThemeSpring, 1.0)
	decorated1.AddComponent(dc1)

	notDecorated := world.CreateEntity()
	dc2 := NewEventDecorationComponent(222)
	notDecorated.AddComponent(dc2)

	decorated2 := world.CreateEntity()
	dc3 := NewEventDecorationComponent(333)
	dc3.GenerateDecorations(DecorationThemeWinter, 0.5)
	decorated2.AddComponent(dc3)

	noComponent := world.CreateEntity()

	entities := []*Entity{decorated1, notDecorated, decorated2, noComponent}
	result := system.GetDecoratedEntities(entities)

	if len(result) != 2 {
		t.Errorf("Expected 2 decorated entities, got %d", len(result))
	}
}

func TestEventDecorationSystem_GetDecorationStats(t *testing.T) {
	world := NewWorld()
	system := NewEventDecorationSystem(world)

	// Create entities with various decoration states
	e1 := world.CreateEntity()
	dc1 := NewEventDecorationComponent(111)
	dc1.GenerateDecorations(DecorationThemeSpring, 1.0)
	e1.AddComponent(dc1)

	e2 := world.CreateEntity()
	dc2 := NewEventDecorationComponent(222)
	dc2.GenerateDecorations(DecorationThemeSpring, 0.5)
	dc2.IsTransitioning = true
	e2.AddComponent(dc2)

	e3 := world.CreateEntity()
	dc3 := NewEventDecorationComponent(333)
	dc3.GenerateDecorations(DecorationThemeWinter, 0.8)
	e3.AddComponent(dc3)

	entities := []*Entity{e1, e2, e3}
	stats := system.GetDecorationStats(entities)

	if stats.DecoratedEntities != 3 {
		t.Errorf("Expected 3 decorated entities, got %d", stats.DecoratedEntities)
	}
	if stats.TransitioningEntities != 1 {
		t.Errorf("Expected 1 transitioning entity, got %d", stats.TransitioningEntities)
	}
	if stats.ThemeCounts[DecorationThemeSpring] != 2 {
		t.Errorf("Expected 2 spring themed, got %d", stats.ThemeCounts[DecorationThemeSpring])
	}
	if stats.ThemeCounts[DecorationThemeWinter] != 1 {
		t.Errorf("Expected 1 winter themed, got %d", stats.ThemeCounts[DecorationThemeWinter])
	}
	if stats.TotalElements == 0 {
		t.Error("Expected some total elements")
	}
}

func TestEventDecorationSystem_UpdateCityVisuals(t *testing.T) {
	world := NewWorld()
	system := NewEventDecorationSystem(world)

	// Create entity with both components
	entity := world.CreateEntity()

	decoComp := NewEventDecorationComponent(12345)
	decoComp.GenerateDecorations(DecorationThemeWinter, 1.0)
	decoComp.TransitionProgress = 1.0
	entity.AddComponent(decoComp)

	cityVisual := NewCityVisualComponent("test_city")
	cityVisual.DecorationDensity = 0.3
	cityVisual.LightingLevel = 0.5
	entity.AddComponent(cityVisual)

	// Update should boost city visuals
	system.updateCityVisuals(entity, decoComp)

	// Winter theme should boost lighting
	if cityVisual.LightingLevel <= 0.5 {
		t.Error("Lighting should be boosted during winter event")
	}
}

func TestEventDecorationSystem_Update_SwitchEvents(t *testing.T) {
	world := NewWorld()
	system := NewEventDecorationSystem(world)

	// Create world entity with initial event
	worldEntity := world.CreateEntity()
	eventComp := NewSeasonalEventComponent(12345, false)
	eventComp.ActiveEvents = []EventInstance{
		{
			Definition: EventDefinition{ID: "spring_festival", Theme: EventThemeSpring},
			Phase:      EventPhaseActive,
		},
	}
	worldEntity.AddComponent(eventComp)

	// Create entity with decoration
	entity := world.CreateEntity()
	decoComp := NewEventDecorationComponent(12345)
	entity.AddComponent(decoComp)

	entities := []*Entity{worldEntity, entity}

	// Apply spring decorations
	for i := 0; i < 200; i++ {
		system.Update(entities, 1.0/60.0)
	}
	if decoComp.ActiveTheme != DecorationThemeSpring {
		t.Error("Should have spring decorations")
	}

	// Switch to summer event
	eventComp.ActiveEvents = []EventInstance{
		{
			Definition: EventDefinition{ID: "summer_solstice", Theme: EventThemeSummer},
			Phase:      EventPhaseActive,
		},
	}

	// Update should switch to summer
	for i := 0; i < 200; i++ {
		system.Update(entities, 1.0/60.0)
	}
	if decoComp.ActiveTheme != DecorationThemeSummer {
		t.Errorf("Should have summer decorations, got %s", decoComp.ActiveTheme)
	}
	if decoComp.EventID != "summer_solstice" {
		t.Errorf("Should have summer event ID, got %s", decoComp.EventID)
	}
}

func TestEventDecorationSystem_Update_NoDecorationComponent(t *testing.T) {
	world := NewWorld()
	system := NewEventDecorationSystem(world)

	// Create world entity with active event
	worldEntity := world.CreateEntity()
	eventComp := NewSeasonalEventComponent(12345, false)
	eventComp.ActiveEvents = []EventInstance{
		{
			Definition: EventDefinition{ID: "test", Theme: EventThemeSpring},
			Phase:      EventPhaseActive,
		},
	}
	worldEntity.AddComponent(eventComp)

	// Create entity without decoration component
	entity := world.CreateEntity()

	entities := []*Entity{worldEntity, entity}

	// Should not panic
	system.Update(entities, 1.0/60.0)
}

func TestEventDecorationSystem_Update_EmptyEntities(t *testing.T) {
	world := NewWorld()
	system := NewEventDecorationSystem(world)

	// Should not panic with empty slice
	system.Update([]*Entity{}, 1.0/60.0)

	// Should not panic with nil
	system.Update(nil, 1.0/60.0)
}

func BenchmarkEventDecorationSystem_Update(b *testing.B) {
	world := NewWorld()
	system := NewEventDecorationSystem(world)

	// Create world entity with active event
	worldEntity := world.CreateEntity()
	eventComp := NewSeasonalEventComponent(12345, false)
	eventComp.ActiveEvents = []EventInstance{
		{
			Definition: EventDefinition{ID: "test", Theme: EventThemeSummer},
			Phase:      EventPhaseActive,
		},
	}
	worldEntity.AddComponent(eventComp)

	// Create 100 entities with decorations
	entities := make([]*Entity, 101)
	entities[0] = worldEntity
	for i := 1; i <= 100; i++ {
		entity := world.CreateEntity()
		decoComp := NewEventDecorationComponent(int64(i))
		entity.AddComponent(decoComp)
		entities[i] = entity
	}

	// Pre-apply decorations
	for i := 0; i < 200; i++ {
		system.Update(entities, 1.0/60.0)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		system.Update(entities, 1.0/60.0)
	}
}
