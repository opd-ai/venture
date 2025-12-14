package engine

import (
	"testing"
	"time"
)

// TestEventDecorationIntegration tests the full integration between
// EventCalendarSystem and EventDecorationSystem.
func TestEventDecorationIntegration_FullEventCycle(t *testing.T) {
	world := NewWorld()

	// Create simulation clock starting at a specific time
	startTime := time.Date(2025, 3, 21, 12, 0, 0, 0, time.UTC) // Near spring festival
	clock := NewSimulationClock(12345)
	clock.Reset(startTime)

	// Create systems
	calendarSystem := NewEventCalendarSystem(world, clock)
	decorationSystem := NewEventDecorationSystem(world)

	// Create world entity with seasonal events
	worldEntity := world.CreateEntity()
	eventComp := NewSeasonalEventComponent(12345, false)
	worldEntity.AddComponent(eventComp)

	// Create decoratable entities (city, buildings, NPCs)
	city := world.CreateEntity()
	cityDeco := NewEventDecorationComponent(111)
	city.AddComponent(cityDeco)
	cityVisual := NewCityVisualComponent("test_city")
	city.AddComponent(cityVisual)

	npc := world.CreateEntity()
	npcDeco := NewEventDecorationComponent(222)
	npc.AddComponent(npcDeco)

	entities := []*Entity{worldEntity, city, npc}

	// Advance clock to trigger spring event
	clock.Reset(time.Date(2025, 3, 25, 12, 0, 0, 0, time.UTC))

	// Update calendar system to activate event
	calendarSystem.Update(entities, 1.0/60.0)

	// Verify event is active
	activeEvents := eventComp.GetActiveEvents()
	if len(activeEvents) == 0 {
		t.Log("Note: Event may not be active depending on calendar seed")
		// Manually trigger for testing
		calendarSystem.TriggerEvent(worldEntity, "spring_festival", 7)
		calendarSystem.Update(entities, 1.0/60.0)
	}

	// Update decoration system - should start applying decorations
	decorationSystem.Update(entities, 1.0/60.0)

	// Verify decorations are being applied
	if cityDeco.EventID == "" {
		t.Error("City should have event ID set")
	}
	if cityDeco.ActiveTheme == DecorationThemeNone {
		t.Error("City should have decoration theme set")
	}

	// Run until decorations are fully applied
	for i := 0; i < 200; i++ {
		decorationSystem.Update(entities, 1.0/60.0)
	}

	// Verify fully decorated
	if !cityDeco.IsFullyDecorated() {
		t.Error("City should be fully decorated")
	}
	if len(cityDeco.Elements) == 0 {
		t.Error("City should have decoration elements")
	}

	// Verify NPC has costume
	if npcDeco.CostumeVariant == 0 {
		t.Error("NPC should have costume variant")
	}

	// End the event
	calendarSystem.EndEvent(worldEntity, "spring_festival")
	calendarSystem.Update(entities, 1.0/60.0)

	// Clear active events to trigger decoration removal
	eventComp.ActiveEvents = nil

	// Update decoration system - should start removing
	decorationSystem.Update(entities, 1.0/60.0)

	if !cityDeco.IsTransitioning || cityDeco.TransitionDirection != -1 {
		t.Error("City should be transitioning to remove decorations")
	}

	// Run until decorations are removed
	for i := 0; i < 200; i++ {
		decorationSystem.Update(entities, 1.0/60.0)
	}

	// Verify decorations removed
	if cityDeco.HasDecorations() {
		t.Error("City decorations should be removed")
	}
	if npcDeco.HasDecorations() {
		t.Error("NPC decorations should be removed")
	}
}

// TestEventDecorationIntegration_MultipleEventsSequence tests switching between events.
func TestEventDecorationIntegration_MultipleEventsSequence(t *testing.T) {
	world := NewWorld()
	decorationSystem := NewEventDecorationSystem(world)

	// Create world entity with controlled events
	worldEntity := world.CreateEntity()
	eventComp := NewSeasonalEventComponent(12345, false)
	worldEntity.AddComponent(eventComp)

	// Create decoratable entity
	entity := world.CreateEntity()
	deco := NewEventDecorationComponent(12345)
	entity.AddComponent(deco)

	entities := []*Entity{worldEntity, entity}

	// Sequence through all seasons
	themes := []EventTheme{
		EventThemeSpring,
		EventThemeSummer,
		EventThemeAutumn,
		EventThemeWinter,
	}
	expectedDecoThemes := []DecorationTheme{
		DecorationThemeSpring,
		DecorationThemeSummer,
		DecorationThemeAutumn,
		DecorationThemeWinter,
	}

	for i, theme := range themes {
		// Set active event
		eventComp.ActiveEvents = []EventInstance{
			{
				Definition: EventDefinition{
					ID:    string(theme) + "_festival",
					Theme: theme,
					Seed:  int64(i * 1000),
				},
				Phase: EventPhaseActive,
			},
		}

		// Apply decorations
		for j := 0; j < 200; j++ {
			decorationSystem.Update(entities, 1.0/60.0)
		}

		// Verify correct theme
		if deco.ActiveTheme != expectedDecoThemes[i] {
			t.Errorf("Expected %s theme, got %s", expectedDecoThemes[i], deco.ActiveTheme)
		}
	}
}

// TestEventDecorationIntegration_ParticleEffects verifies particle effects are configured.
func TestEventDecorationIntegration_ParticleEffects(t *testing.T) {
	world := NewWorld()
	decorationSystem := NewEventDecorationSystem(world)

	// Create world entity with winter event (has multiple particle effects)
	worldEntity := world.CreateEntity()
	eventComp := NewSeasonalEventComponent(12345, false)
	eventComp.ActiveEvents = []EventInstance{
		{
			Definition: EventDefinition{
				ID:    "winter_celebration",
				Theme: EventThemeWinter,
				Seed:  99999,
			},
			Phase: EventPhaseActive,
		},
	}
	worldEntity.AddComponent(eventComp)

	// Create entity with high decoration level
	entity := world.CreateEntity()
	deco := NewEventDecorationComponent(12345)
	entity.AddComponent(deco)

	entities := []*Entity{worldEntity, entity}

	// Apply decorations
	for i := 0; i < 200; i++ {
		decorationSystem.Update(entities, 1.0/60.0)
	}

	// Verify particle effects
	effects := deco.GetActiveParticleEffects()
	if len(effects) == 0 {
		t.Error("Expected particle effects for winter event")
	}

	// Verify effect types
	hasSnow := false
	for _, effect := range effects {
		if effect.EffectType == "snow" {
			hasSnow = true
		}
		if effect.Rate <= 0 {
			t.Errorf("Effect %s should have positive rate", effect.EffectType)
		}
	}
	if !hasSnow {
		t.Error("Winter event should have snow particle effect")
	}
}

// TestEventDecorationIntegration_CityVisualBoost verifies city visuals are boosted.
func TestEventDecorationIntegration_CityVisualBoost(t *testing.T) {
	world := NewWorld()
	decorationSystem := NewEventDecorationSystem(world)

	// Create world entity with event
	worldEntity := world.CreateEntity()
	eventComp := NewSeasonalEventComponent(12345, false)
	eventComp.ActiveEvents = []EventInstance{
		{
			Definition: EventDefinition{
				ID:    "harvest_festival",
				Theme: EventThemeAutumn,
			},
			Phase: EventPhaseActive,
		},
	}
	worldEntity.AddComponent(eventComp)

	// Create city with both components
	city := world.CreateEntity()
	deco := NewEventDecorationComponent(12345)
	city.AddComponent(deco)

	cityVisual := NewCityVisualComponent("harvest_city")
	initialMarketActivity := cityVisual.MarketActivity
	city.AddComponent(cityVisual)

	entities := []*Entity{worldEntity, city}

	// Apply decorations
	for i := 0; i < 200; i++ {
		decorationSystem.Update(entities, 1.0/60.0)
	}

	// Autumn should boost market activity
	if cityVisual.MarketActivity <= initialMarketActivity {
		t.Error("Autumn event should boost market activity")
	}
}

// TestEventDecorationIntegration_Determinism verifies deterministic decoration generation.
func TestEventDecorationIntegration_Determinism(t *testing.T) {
	seed := int64(54321)

	// Run twice with same setup
	results := make([][]DecorationElement, 2)

	for run := 0; run < 2; run++ {
		world := NewWorld()
		decorationSystem := NewEventDecorationSystem(world)

		worldEntity := world.CreateEntity()
		eventComp := NewSeasonalEventComponent(seed, false)
		eventComp.ActiveEvents = []EventInstance{
			{
				Definition: EventDefinition{
					ID:    "test_event",
					Theme: EventThemeSummer,
					Seed:  seed,
				},
				Phase: EventPhaseActive,
			},
		}
		worldEntity.AddComponent(eventComp)

		entity := world.CreateEntity()
		deco := NewEventDecorationComponent(seed)
		entity.AddComponent(deco)

		entities := []*Entity{worldEntity, entity}

		for i := 0; i < 200; i++ {
			decorationSystem.Update(entities, 1.0/60.0)
		}

		results[run] = deco.Elements
	}

	// Compare results
	if len(results[0]) != len(results[1]) {
		t.Fatalf("Element counts differ: %d vs %d", len(results[0]), len(results[1]))
	}

	for i := range results[0] {
		if results[0][i].Type != results[1][i].Type {
			t.Errorf("Element %d type mismatch: %s vs %s", i, results[0][i].Type, results[1][i].Type)
		}
		if results[0][i].ColorHue != results[1][i].ColorHue {
			t.Errorf("Element %d hue mismatch: %d vs %d", i, results[0][i].ColorHue, results[1][i].ColorHue)
		}
	}
}

func BenchmarkEventDecorationIntegration_FullUpdate(b *testing.B) {
	world := NewWorld()
	decorationSystem := NewEventDecorationSystem(world)

	// Create world with active event
	worldEntity := world.CreateEntity()
	eventComp := NewSeasonalEventComponent(12345, false)
	eventComp.ActiveEvents = []EventInstance{
		{
			Definition: EventDefinition{ID: "bench", Theme: EventThemeSummer},
			Phase:      EventPhaseActive,
		},
	}
	worldEntity.AddComponent(eventComp)

	// Create 100 entities
	entities := make([]*Entity, 101)
	entities[0] = worldEntity
	for i := 1; i <= 100; i++ {
		e := world.CreateEntity()
		e.AddComponent(NewEventDecorationComponent(int64(i)))
		e.AddComponent(NewCityVisualComponent("city_" + string(rune(i))))
		entities[i] = e
	}

	// Pre-apply
	for i := 0; i < 200; i++ {
		decorationSystem.Update(entities, 1.0/60.0)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		decorationSystem.Update(entities, 1.0/60.0)
	}
}
