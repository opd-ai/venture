package engine

import (
	"testing"
	"time"
)

func TestNewPoliticsSystem(t *testing.T) {
	world := NewWorld()
	ps := NewPoliticsSystem(world)

	if ps.world != world {
		t.Error("PoliticsSystem.world should be set")
	}
	if ps.activeEvents == nil {
		t.Error("activeEvents should be initialized")
	}
	if ps.eventHistory == nil {
		t.Error("eventHistory should be initialized")
	}
}

func TestPoliticsSystem_SetServerFaction(t *testing.T) {
	world := NewWorld()
	ps := NewPoliticsSystem(world)

	faction := NewServerFaction("server1", "TestFaction", Alignment{LawAxis: 0.5, GoodAxis: 0.5})
	ps.SetServerFaction(faction)

	if ps.GetServerFaction() != faction {
		t.Error("SetServerFaction did not set faction correctly")
	}
}

func TestPoliticsSystem_CreateAlliance(t *testing.T) {
	world := NewWorld()
	ps := NewPoliticsSystem(world)

	// Test without faction (should error)
	_, err := ps.CreateAlliance("server2", 86400)
	if err == nil {
		t.Error("CreateAlliance should error when server has no faction")
	}

	// Set faction and create alliance
	faction := NewServerFaction("server1", "Faction1", Alignment{})
	ps.SetServerFaction(faction)

	event, err := ps.CreateAlliance("server2", 86400)
	if err != nil {
		t.Errorf("CreateAlliance failed: %v", err)
	}
	if event.Type != EventTypeAlliance {
		t.Errorf("Event type = %v, want %v", event.Type, EventTypeAlliance)
	}
	if !faction.IsAlly("server2") {
		t.Error("server2 should be an ally after alliance creation")
	}

	// Check effects
	if mult, exists := event.GetEffect("trade_price_multiplier"); !exists || mult != 0.8 {
		t.Errorf("trade_price_multiplier = %v, want 0.8", mult)
	}

	// Check event was added
	activeEvents := ps.GetActiveEvents()
	if len(activeEvents) != 1 {
		t.Errorf("activeEvents length = %d, want 1", len(activeEvents))
	}
}

func TestPoliticsSystem_DeclareWar(t *testing.T) {
	world := NewWorld()
	ps := NewPoliticsSystem(world)
	faction := NewServerFaction("server1", "Faction1", Alignment{})
	ps.SetServerFaction(faction)

	event, err := ps.DeclareWar("server3", 86400)
	if err != nil {
		t.Errorf("DeclareWar failed: %v", err)
	}
	if event.Type != EventTypeWar {
		t.Errorf("Event type = %v, want %v", event.Type, EventTypeWar)
	}
	if !faction.IsEnemy("server3") {
		t.Error("server3 should be an enemy after war declaration")
	}

	// Check effects
	if mult, exists := event.GetEffect("trade_price_multiplier"); !exists || mult != 1.5 {
		t.Errorf("trade_price_multiplier = %v, want 1.5", mult)
	}
	if borders, exists := event.GetEffect("contested_borders"); !exists || borders != true {
		t.Errorf("contested_borders = %v, want true", borders)
	}
}

func TestPoliticsSystem_SignTreaty(t *testing.T) {
	world := NewWorld()
	ps := NewPoliticsSystem(world)
	faction := NewServerFaction("server1", "Faction1", Alignment{})
	ps.SetServerFaction(faction)

	// Add server as enemy first
	faction.AddEnemy("server4")

	event, err := ps.SignTreaty("server4", 86400)
	if err != nil {
		t.Errorf("SignTreaty failed: %v", err)
	}
	if event.Type != EventTypeTreaty {
		t.Errorf("Event type = %v, want %v", event.Type, EventTypeTreaty)
	}
	if faction.IsEnemy("server4") {
		t.Error("server4 should not be an enemy after treaty")
	}

	// Check effects
	if mult, exists := event.GetEffect("trade_price_multiplier"); !exists || mult != 1.0 {
		t.Errorf("trade_price_multiplier = %v, want 1.0", mult)
	}
}

func TestPoliticsSystem_ImposeEmbargo(t *testing.T) {
	world := NewWorld()
	ps := NewPoliticsSystem(world)
	faction := NewServerFaction("server1", "Faction1", Alignment{})
	ps.SetServerFaction(faction)

	event, err := ps.ImposeEmbargo("server5", 86400)
	if err != nil {
		t.Errorf("ImposeEmbargo failed: %v", err)
	}
	if event.Type != EventTypeEmbargo {
		t.Errorf("Event type = %v, want %v", event.Type, EventTypeEmbargo)
	}

	// Check effects
	if blocked, exists := event.GetEffect("trade_blocked"); !exists || blocked != true {
		t.Errorf("trade_blocked = %v, want true", blocked)
	}
}

func TestPoliticsSystem_EstablishTradePact(t *testing.T) {
	world := NewWorld()
	ps := NewPoliticsSystem(world)
	faction := NewServerFaction("server1", "Faction1", Alignment{})
	ps.SetServerFaction(faction)

	event, err := ps.EstablishTradePact("server6", 86400)
	if err != nil {
		t.Errorf("EstablishTradePact failed: %v", err)
	}
	if event.Type != EventTypeTradePact {
		t.Errorf("Event type = %v, want %v", event.Type, EventTypeTradePact)
	}

	// Check effects
	if mult, exists := event.GetEffect("trade_price_multiplier"); !exists || mult != 0.9 {
		t.Errorf("trade_price_multiplier = %v, want 0.9", mult)
	}
}

func TestPoliticsSystem_Update(t *testing.T) {
	world := NewWorld()
	ps := NewPoliticsSystem(world)
	faction := NewServerFaction("server1", "Faction1", Alignment{})
	ps.SetServerFaction(faction)

	// Create an entity with PoliticsComponent
	entity := world.CreateEntity()
	comp := PoliticsComponent{}
	entity.AddComponent(comp)
	world.Update(0.0) // Commit entity

	// Create an active event
	ps.CreateAlliance("server2", 86400)

	// Update should propagate events to entity
	ps.Update(0.016)

	// Check entity component was updated
	compInterface, ok := entity.GetComponent("politics")
	if !ok {
		t.Fatal("Entity should have PoliticsComponent")
	}
	updatedComp, ok := compInterface.(PoliticsComponent)
	if !ok {
		t.Fatal("Component should be PoliticsComponent type")
	}
	if updatedComp.Faction != faction {
		t.Error("Component should have server faction")
	}
	if len(updatedComp.Events) != 1 {
		t.Errorf("Component should have 1 event, got %d", len(updatedComp.Events))
	}
}

func TestPoliticsSystem_ExpiredEvents(t *testing.T) {
	world := NewWorld()
	ps := NewPoliticsSystem(world)
	faction := NewServerFaction("server1", "Faction1", Alignment{})
	ps.SetServerFaction(faction)

	// Create event with very short duration
	event := NewPoliticalEvent(EventTypeAlliance, []string{"server1", "server2"}, 1)
	event.StartTime = time.Now().Unix() - 10 // Started 10 seconds ago
	ps.activeEvents = append(ps.activeEvents, *event)

	// Update should move expired event to history
	ps.Update(0.016)

	activeEvents := ps.GetActiveEvents()
	history := ps.GetEventHistory()

	if len(activeEvents) != 0 {
		t.Errorf("Should have 0 active events, got %d", len(activeEvents))
	}
	if len(history) != 1 {
		t.Errorf("Should have 1 historical event, got %d", len(history))
	}
}

func TestPoliticsSystem_GetTradeMultiplier(t *testing.T) {
	world := NewWorld()
	ps := NewPoliticsSystem(world)
	faction := NewServerFaction("server1", "Faction1", Alignment{})
	ps.SetServerFaction(faction)

	// No events: default multiplier
	if mult := ps.GetTradeMultiplier("server2"); mult != 1.0 {
		t.Errorf("Default trade multiplier = %v, want 1.0", mult)
	}

	// Alliance: 20% discount
	ps.CreateAlliance("server2", 86400)
	if mult := ps.GetTradeMultiplier("server2"); mult != 0.8 {
		t.Errorf("Alliance trade multiplier = %v, want 0.8", mult)
	}

	// War: 50% markup
	ps.DeclareWar("server3", 86400)
	if mult := ps.GetTradeMultiplier("server3"); mult != 1.5 {
		t.Errorf("War trade multiplier = %v, want 1.5", mult)
	}

	// Trade pact: 10% discount
	ps.EstablishTradePact("server4", 86400)
	if mult := ps.GetTradeMultiplier("server4"); mult != 0.9 {
		t.Errorf("Trade pact multiplier = %v, want 0.9", mult)
	}

	// No relationship: default
	if mult := ps.GetTradeMultiplier("server99"); mult != 1.0 {
		t.Errorf("Unknown server multiplier = %v, want 1.0", mult)
	}
}

func TestPoliticsSystem_IsTravelAllowed(t *testing.T) {
	world := NewWorld()
	ps := NewPoliticsSystem(world)
	faction := NewServerFaction("server1", "Faction1", Alignment{})
	ps.SetServerFaction(faction)

	// Default: travel allowed
	if !ps.IsTravelAllowed("server2") {
		t.Error("Travel should be allowed by default")
	}

	// Embargo: travel still allowed (only blocks trade)
	ps.ImposeEmbargo("server2", 86400)
	if !ps.IsTravelAllowed("server2") {
		t.Error("Travel should be allowed even with embargo")
	}

	// War: travel allowed but may be dangerous
	ps.DeclareWar("server3", 86400)
	if !ps.IsTravelAllowed("server3") {
		t.Error("Travel should be allowed even during war")
	}
}

func TestPoliticsSystem_IsTradeBlocked(t *testing.T) {
	world := NewWorld()
	ps := NewPoliticsSystem(world)
	faction := NewServerFaction("server1", "Faction1", Alignment{})
	ps.SetServerFaction(faction)

	// Default: trade not blocked
	if ps.IsTradeBlocked("server2") {
		t.Error("Trade should not be blocked by default")
	}

	// Embargo: trade blocked
	ps.ImposeEmbargo("server2", 86400)
	if !ps.IsTradeBlocked("server2") {
		t.Error("Trade should be blocked with embargo")
	}

	// War: trade not blocked (just expensive)
	ps.DeclareWar("server3", 86400)
	if ps.IsTradeBlocked("server3") {
		t.Error("Trade should not be blocked during war (only more expensive)")
	}

	// Alliance: trade not blocked
	ps.CreateAlliance("server4", 86400)
	if ps.IsTradeBlocked("server4") {
		t.Error("Trade should not be blocked with alliance")
	}
}

func TestPoliticsSystem_Concurrency(t *testing.T) {
	world := NewWorld()
	ps := NewPoliticsSystem(world)
	faction := NewServerFaction("server1", "Faction1", Alignment{})
	ps.SetServerFaction(faction)

	// Test concurrent access
	done := make(chan bool)

	go func() {
		for i := 0; i < 10; i++ {
			ps.GetActiveEvents()
		}
		done <- true
	}()

	go func() {
		for i := 0; i < 10; i++ {
			ps.GetTradeMultiplier("server2")
		}
		done <- true
	}()

	go func() {
		for i := 0; i < 5; i++ {
			ps.CreateAlliance("server2", 86400)
		}
		done <- true
	}()

	// Wait for all goroutines
	<-done
	<-done
	<-done
}
