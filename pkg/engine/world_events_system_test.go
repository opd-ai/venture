package engine

import (
	"testing"

	"github.com/opd-ai/venture/pkg/integration/world_events"
)

// Stub components for testing
type stubGuildComponent struct{}

func (s *stubGuildComponent) Type() string { return "guild" }

type stubMerchantComponent struct{}

func (s *stubMerchantComponent) Type() string { return "merchant" }

type stubWeatherComponent struct{}

func (s *stubWeatherComponent) Type() string { return "weather" }

type stubPosComponent struct{ X, Y float64 }

func (s *stubPosComponent) Type() string { return "position" }

type stubHealthComponent struct{ Current, Max float64 }

func (s *stubHealthComponent) Type() string { return "health" }

type stubInventoryComponent struct{}

func (s *stubInventoryComponent) Type() string { return "inventory" }

func TestNewWorldEventsSystem(t *testing.T) {
	world := NewWorld()
	seed := int64(12345)

	system := NewWorldEventsSystem(world, seed)

	if system == nil {
		t.Fatal("expected system to be created")
	}
	if system.world != world {
		t.Error("system world not set correctly")
	}
	if system.eventManager == nil {
		t.Error("event manager not initialized")
	}
	if system.updateInterval != 30.0 {
		t.Errorf("expected update interval 30.0, got %f", system.updateInterval)
	}
}

func TestWorldEventsSystemUpdate(t *testing.T) {
	world := NewWorld()
	seed := int64(12345)
	system := NewWorldEventsSystem(world, seed)

	// Update with less than interval
	system.Update([]*Entity{}, 10.0)
	if system.updateTimer != 10.0 {
		t.Errorf("expected timer 10.0, got %f", system.updateTimer)
	}

	// Update past interval — timer should subtract interval, not reset to 0
	system.Update([]*Entity{}, 25.0)
	if system.updateTimer >= 30.0 {
		t.Error("timer should have been reduced after passing interval")
	}
	// With 10+25=35 accumulated and interval=30, remainder should be 5.0
	expected := 35.0 - 30.0
	if system.updateTimer != expected {
		t.Errorf("expected timer %f (carried over remainder), got %f", expected, system.updateTimer)
	}
}

func TestWorldEventsSystemTriggerEvent(t *testing.T) {
	world := NewWorld()
	seed := int64(12345)
	system := NewWorldEventsSystem(world, seed)

	params := world_events.TriggerParams{
		TriggerType: world_events.TriggerGuildWar,
		Severity:    world_events.SeverityMajor,
		Location:    "test_location",
		ServerID:    "test_server",
		GuildID:     "guild_1",
	}

	event, err := system.TriggerEvent(world_events.TriggerGuildWar, params)
	if err != nil {
		t.Fatalf("failed to trigger event: %v", err)
	}
	if event == nil {
		t.Fatal("expected event to be created")
	}
	if event.Type != world_events.EventGuildWarfare {
		t.Errorf("expected guild warfare event, got %s", event.Type)
	}
}

func TestWorldEventsSystemTriggerEventInvalid(t *testing.T) {
	world := NewWorld()
	seed := int64(12345)
	system := NewWorldEventsSystem(world, seed)

	// Invalid params (missing required fields)
	params := world_events.TriggerParams{
		TriggerType: world_events.TriggerGuildWar,
		Severity:    world_events.SeverityMajor,
	}

	_, err := system.TriggerEvent(world_events.TriggerGuildWar, params)
	if err == nil {
		t.Error("expected error for invalid params")
	}
}

func TestWorldEventsSystemGetActiveEvents(t *testing.T) {
	world := NewWorld()
	seed := int64(12345)
	system := NewWorldEventsSystem(world, seed)

	// Initially no events
	events := system.GetActiveEvents()
	if len(events) != 0 {
		t.Errorf("expected 0 active events, got %d", len(events))
	}

	// Trigger an event
	params := world_events.TriggerParams{
		TriggerType: world_events.TriggerGuildWar,
		Severity:    world_events.SeverityMajor,
		Location:    "test_location",
		ServerID:    "test_server",
		GuildID:     "guild_1",
	}

	_, err := system.TriggerEvent(world_events.TriggerGuildWar, params)
	if err != nil {
		t.Fatalf("failed to trigger event: %v", err)
	}

	// Events may not be immediately active (they start with a delay)
	// so we just verify no crash
	events = system.GetActiveEvents()
	if events == nil {
		t.Error("expected events slice, got nil")
	}
}

func TestWorldEventsSystemCheckGuildWarfareTriggers(t *testing.T) {
	world := NewWorld()
	seed := int64(12345)
	system := NewWorldEventsSystem(world, seed)

	// With no guilds, should not crash
	system.checkGuildWarfareTriggers()

	// With one guild, should not trigger (need >1)
	guild1 := world.CreateEntity()
	guild1.AddComponent(&stubGuildComponent{})
	system.checkGuildWarfareTriggers()

	// With two guilds, should trigger warfare event
	guild2 := world.CreateEntity()
	guild2.AddComponent(&stubGuildComponent{})
	system.checkGuildWarfareTriggers()
}

func TestWorldEventsSystemCheckEconomicTriggers(t *testing.T) {
	world := NewWorld()
	seed := int64(12345)
	system := NewWorldEventsSystem(world, seed)

	// Add merchants
	for i := 0; i < 10; i++ {
		merchant := world.CreateEntity()
		merchant.AddComponent(&stubMerchantComponent{})
	}

	// Check triggers (should not error)
	system.checkEconomicTriggers()

	// Verify no crash
}

func TestWorldEventsSystemCheckWeatherTriggers(t *testing.T) {
	world := NewWorld()
	seed := int64(12345)
	system := NewWorldEventsSystem(world, seed)

	// Add weather entity
	weather := world.CreateEntity()
	weather.AddComponent(&stubWeatherComponent{})

	// Check triggers (should not error)
	system.checkWeatherTriggers()

	// Verify no crash
}

func TestWorldEventsSystemApplyEventImpact(t *testing.T) {
	world := NewWorld()
	seed := int64(12345)
	system := NewWorldEventsSystem(world, seed)

	// Create entity with position and health
	entity := world.CreateEntity()
	entity.AddComponent(&stubPosComponent{X: 0, Y: 0})
	entity.AddComponent(&stubHealthComponent{Current: 100, Max: 100})

	// Apply damage impact
	impact := world_events.Impact{
		Type:     world_events.ImpactSpawnRate,
		Target:   "test_target",
		Modifier: 10.0,
	}

	// Should not crash
	system.applyEventImpact(impact)
}

func TestWorldEventsSystemApplyAreaDamage(t *testing.T) {
	world := NewWorld()
	seed := int64(12345)
	system := NewWorldEventsSystem(world, seed)

	// Create entity with position and health
	entity := world.CreateEntity()
	entity.AddComponent(&stubPosComponent{X: 0, Y: 0})
	entity.AddComponent(&stubHealthComponent{Current: 100, Max: 100})

	impact := world_events.Impact{
		Type:     world_events.ImpactSpawnRate,
		Modifier: 50.0,
	}

	// Should not crash
	system.applyAreaDamage(impact)
}

func TestWorldEventsSystemModifyResources(t *testing.T) {
	world := NewWorld()
	seed := int64(12345)
	system := NewWorldEventsSystem(world, seed)

	// Create merchant with inventory
	merchant := world.CreateEntity()
	merchant.AddComponent(&stubMerchantComponent{})
	merchant.AddComponent(&stubInventoryComponent{})

	impact := world_events.Impact{
		Type:     world_events.ImpactPriceChange,
		Modifier: 1.5,
	}

	// Should not crash
	system.modifyResources(impact)
}

func TestWorldEventsSystemUpdateFactionRelations(t *testing.T) {
	world := NewWorld()
	seed := int64(12345)
	system := NewWorldEventsSystem(world, seed)

	impact := world_events.Impact{
		Type:     world_events.ImpactNPCReputation,
		Target:   "faction_1",
		Modifier: -0.3,
	}

	// Should not crash
	system.updateFactionRelations(impact)
}

func TestWorldEventsSystemUpdateWeather(t *testing.T) {
	world := NewWorld()
	seed := int64(12345)
	system := NewWorldEventsSystem(world, seed)

	// Create weather entity
	weather := world.CreateEntity()
	weather.AddComponent(&stubWeatherComponent{})

	impact := world_events.Impact{
		Type:     world_events.ImpactWeather,
		Modifier: 0.9,
	}

	// Should not crash
	system.updateWeather(impact)
}

func TestWorldEventsSystemGetEventChain(t *testing.T) {
	world := NewWorld()
	seed := int64(12345)
	system := NewWorldEventsSystem(world, seed)

	// Trigger event
	params := world_events.TriggerParams{
		TriggerType: world_events.TriggerPlayerChoice,
		Severity:    world_events.SeverityMajor,
		Location:    "test_location",
		ServerID:    "test_server",
		ChoiceID:    "choice_1",
	}

	event, err := system.TriggerEvent(world_events.TriggerPlayerChoice, params)
	if err != nil {
		t.Fatalf("failed to trigger event: %v", err)
	}

	// Get event chain
	chain := system.GetEventChain(event.ID)

	if len(chain) == 0 {
		t.Error("expected chain to have at least the event ID")
	}
}
