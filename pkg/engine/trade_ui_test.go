package engine

import (
	"testing"

	"github.com/opd-ai/venture/pkg/procgen/item"
)

// mockInputComponent is a test component to mark entities as players
type mockInputComponent struct{}

func (m *mockInputComponent) Type() string { return "input" }

// TestTradeUI_New verifies TradeUI initialization.
func TestTradeUI_New(t *testing.T) {
	world := NewWorld()
	tradeSystem := NewTradeSystem(world)

	ui := NewTradeUI(world, tradeSystem, 1920, 1080)

	if ui == nil {
		t.Fatal("NewTradeUI returned nil")
	}

	if ui.visible {
		t.Error("New TradeUI should not be visible")
	}

	if ui.state != TradeUIStateIdle {
		t.Errorf("Expected idle state, got %v", ui.state)
	}

	if ui.gridCols != 6 || ui.gridRows != 2 {
		t.Errorf("Unexpected grid dimensions: %dx%d", ui.gridCols, ui.gridRows)
	}
}

// TestTradeUI_OpenClose verifies open/close behavior.
func TestTradeUI_OpenClose(t *testing.T) {
	world := NewWorld()
	tradeSystem := NewTradeSystem(world)
	ui := NewTradeUI(world, tradeSystem, 1920, 1080)

	// Create player entity
	player := world.CreateEntity()
	player.AddComponent(&PositionComponent{X: 0, Y: 0})
	player.AddComponent(NewInventoryComponent(10, 100.0))
	ui.SetPlayerEntity(player)

	world.Update(0.0) // Commit entities

	// Test open
	ui.Open()
	if !ui.visible {
		t.Error("UI should be visible after Open()")
	}
	if ui.state != TradeUIStateSelectingPartner {
		t.Errorf("Expected SelectingPartner state, got %v", ui.state)
	}

	// Test close
	ui.Close()
	if ui.visible {
		t.Error("UI should not be visible after Close()")
	}
	if ui.state != TradeUIStateIdle {
		t.Errorf("Expected Idle state after close, got %v", ui.state)
	}
}

// TestTradeUI_Toggle verifies toggle behavior.
func TestTradeUI_Toggle(t *testing.T) {
	world := NewWorld()
	tradeSystem := NewTradeSystem(world)
	ui := NewTradeUI(world, tradeSystem, 1920, 1080)

	player := world.CreateEntity()
	player.AddComponent(&PositionComponent{X: 0, Y: 0})
	ui.SetPlayerEntity(player)

	world.Update(0.0)

	// Toggle on
	ui.Toggle()
	if !ui.visible {
		t.Error("Toggle should make UI visible")
	}

	// Toggle off
	ui.Toggle()
	if ui.visible {
		t.Error("Toggle should hide UI")
	}
}

// TestTradeUI_FindNearbyPartners verifies partner detection.
func TestTradeUI_FindNearbyPartners(t *testing.T) {
	world := NewWorld()
	tradeSystem := NewTradeSystem(world)
	ui := NewTradeUI(world, tradeSystem, 1920, 1080)

	// Create player
	player := world.CreateEntity()
	player.AddComponent(&PositionComponent{X: 0, Y: 0})
	player.AddComponent(&mockInputComponent{}) // Mark as player
	ui.SetPlayerEntity(player)

	// Create nearby partner (within 5 tiles)
	partner1 := world.CreateEntity()
	partner1.AddComponent(&PositionComponent{X: 3, Y: 0})
	partner1.AddComponent(&mockInputComponent{}) // Mark as player

	// Create distant player (beyond 5 tiles)
	partner2 := world.CreateEntity()
	partner2.AddComponent(&PositionComponent{X: 10, Y: 0})
	partner2.AddComponent(&mockInputComponent{}) // Mark as player

	// Create nearby NPC (not a player)
	npc := world.CreateEntity()
	npc.AddComponent(&PositionComponent{X: 2, Y: 0})
	// No input component, so not a player

	world.Update(0.0) // Commit entities

	// Find nearby partners
	ui.findNearbyPartners()

	// Should find only partner1 (nearby player)
	if len(ui.nearbyPartners) != 1 {
		t.Errorf("Expected 1 nearby partner, got %d", len(ui.nearbyPartners))
	}

	if len(ui.nearbyPartners) > 0 && ui.nearbyPartners[0].ID != partner1.ID {
		t.Error("Wrong partner found")
	}

	if ui.selectedPartnerIndex != 0 {
		t.Errorf("Expected selectedPartnerIndex=0, got %d", ui.selectedPartnerIndex)
	}
}

// TestTradeUI_SelectPartner verifies partner selection flow.
func TestTradeUI_SelectPartner(t *testing.T) {
	world := NewWorld()
	tradeSystem := NewTradeSystem(world)
	ui := NewTradeUI(world, tradeSystem, 1920, 1080)

	player := world.CreateEntity()
	player.AddComponent(&PositionComponent{X: 0, Y: 0})
	ui.SetPlayerEntity(player)

	partner := world.CreateEntity()
	partner.AddComponent(&PositionComponent{X: 3, Y: 0})

	world.Update(0.0)

	// Select partner
	ui.selectPartner(partner)

	if ui.partnerEntity == nil || ui.partnerEntity.ID != partner.ID {
		t.Error("Partner not set correctly")
	}

	if ui.state != TradeUIStateSelectingItems {
		t.Errorf("Expected SelectingItems state, got %v", ui.state)
	}
}

// TestTradeUI_ProposeTrade verifies trade proposal.
func TestTradeUI_ProposeTrade(t *testing.T) {
	world := NewWorld()
	tradeSystem := NewTradeSystem(world)
	ui := NewTradeUI(world, tradeSystem, 1920, 1080)

	// Create player with inventory
	player := world.CreateEntity()
	player.AddComponent(&PositionComponent{X: 0, Y: 0})
	playerInv := NewInventoryComponent(10, 100.0)
	player.AddComponent(playerInv)
	ui.SetPlayerEntity(player)

	// Add item to player
	item1 := &item.Item{ID: "sword1", Name: "Iron Sword", Rarity: item.RarityCommon}
	playerInv.AddItem(item1)

	// Create partner with inventory
	partner := world.CreateEntity()
	partner.AddComponent(&PositionComponent{X: 3, Y: 0})
	partnerInv := NewInventoryComponent(10, 100.0)
	partner.AddComponent(partnerInv)
	ui.partnerEntity = partner

	// Add item to partner
	item2 := &item.Item{ID: "potion1", Name: "Health Potion", Rarity: item.RarityCommon}
	partnerInv.AddItem(item2)

	world.Update(0.0) // Commit entities

	// Select items to offer and request
	ui.selectedOfferedSlots = []int{0}   // Offer sword
	ui.selectedRequestedSlots = []int{0} // Request potion

	// Propose trade
	ui.proposeTrade()

	// Verify state changed to pending
	if ui.state != TradeUIStatePending {
		t.Errorf("Expected Pending state, got %v", ui.state)
	}

	// Verify trade component created
	comp, ok := player.GetComponent("trade")
	if !ok {
		t.Fatal("Trade component not created")
	}

	tradeComp, ok := comp.(*TradeComponent)
	if !ok || tradeComp.ActiveTrade == nil {
		t.Fatal("Active trade not set")
	}

	if tradeComp.ActiveTrade.Status != "pending" {
		t.Errorf("Expected pending status, got %s", tradeComp.ActiveTrade.Status)
	}
}

// TestTradeUI_AcceptTrade verifies trade acceptance.
func TestTradeUI_AcceptTrade(t *testing.T) {
	world := NewWorld()
	tradeSystem := NewTradeSystem(world)
	ui := NewTradeUI(world, tradeSystem, 1920, 1080)

	// Create player with inventory
	player := world.CreateEntity()
	player.AddComponent(&PositionComponent{X: 0, Y: 0})
	playerInv := NewInventoryComponent(10, 100.0)
	player.AddComponent(playerInv)
	ui.SetPlayerEntity(player)

	item1 := &item.Item{ID: "sword1", Name: "Iron Sword", Rarity: item.RarityCommon}
	playerInv.AddItem(item1)

	// Create partner with inventory
	partner := world.CreateEntity()
	partner.AddComponent(&PositionComponent{X: 3, Y: 0})
	partnerInv := NewInventoryComponent(10, 100.0)
	partner.AddComponent(partnerInv)

	item2 := &item.Item{ID: "potion1", Name: "Health Potion", Rarity: item.RarityCommon}
	partnerInv.AddItem(item2)

	world.Update(0.0)

	// Propose trade from partner to player
	err := tradeSystem.ProposeTrade(partner.ID, player.ID, []string{"potion1"}, []string{"sword1"})
	if err != nil {
		t.Fatalf("ProposeTrade failed: %v", err)
	}

	// Accept trade through UI
	ui.state = TradeUIStateNegotiating
	ui.partnerEntity = partner
	ui.acceptTrade()

	// Verify trade completed
	if ui.state != TradeUIStateCompleted {
		t.Errorf("Expected Completed state, got %v (message: %s)", ui.state, ui.statusMessage)
	}

	// Verify items swapped
	if len(playerInv.Items) != 1 || playerInv.Items[0].ID != "potion1" {
		t.Error("Player should have potion after trade")
	}

	if len(partnerInv.Items) != 1 || partnerInv.Items[0].ID != "sword1" {
		t.Error("Partner should have sword after trade")
	}
}

// TestTradeUI_RejectTrade verifies trade rejection.
func TestTradeUI_RejectTrade(t *testing.T) {
	world := NewWorld()
	tradeSystem := NewTradeSystem(world)
	ui := NewTradeUI(world, tradeSystem, 1920, 1080)

	// Create player
	player := world.CreateEntity()
	player.AddComponent(&PositionComponent{X: 0, Y: 0})
	playerInv := NewInventoryComponent(10, 100.0)
	player.AddComponent(playerInv)
	ui.SetPlayerEntity(player)

	// Create partner
	partner := world.CreateEntity()
	partner.AddComponent(&PositionComponent{X: 3, Y: 0})
	partnerInv := NewInventoryComponent(10, 100.0)
	partner.AddComponent(partnerInv)

	world.Update(0.0)

	// Propose trade
	err := tradeSystem.ProposeTrade(partner.ID, player.ID, []string{}, []string{})
	if err != nil {
		t.Fatalf("ProposeTrade failed: %v", err)
	}

	// Reject trade through UI
	ui.state = TradeUIStateNegotiating
	ui.partnerEntity = partner
	ui.rejectTrade()

	// Verify state changed to cancelled
	if ui.state != TradeUIStateCancelled {
		t.Errorf("Expected Cancelled state, got %v", ui.state)
	}

	// Verify trade cleared
	comp, ok := player.GetComponent("trade")
	if ok {
		tradeComp, ok := comp.(*TradeComponent)
		if ok && tradeComp.ActiveTrade != nil {
			t.Error("Trade should be cleared after rejection")
		}
	}
}

// TestTradeUI_CancelTrade verifies trade cancellation.
func TestTradeUI_CancelTrade(t *testing.T) {
	world := NewWorld()
	tradeSystem := NewTradeSystem(world)
	ui := NewTradeUI(world, tradeSystem, 1920, 1080)

	// Create player
	player := world.CreateEntity()
	player.AddComponent(&PositionComponent{X: 0, Y: 0})
	playerInv := NewInventoryComponent(10, 100.0)
	player.AddComponent(playerInv)
	ui.SetPlayerEntity(player)

	// Create partner
	partner := world.CreateEntity()
	partner.AddComponent(&PositionComponent{X: 3, Y: 0})
	partnerInv := NewInventoryComponent(10, 100.0)
	partner.AddComponent(partnerInv)
	ui.partnerEntity = partner

	world.Update(0.0)

	// Propose trade
	err := tradeSystem.ProposeTrade(player.ID, partner.ID, []string{}, []string{})
	if err != nil {
		t.Fatalf("ProposeTrade failed: %v", err)
	}

	// Cancel trade
	ui.state = TradeUIStatePending
	ui.cancelTrade()

	// Verify state changed to cancelled
	if ui.state != TradeUIStateCancelled {
		t.Errorf("Expected Cancelled state, got %v", ui.state)
	}
}

// TestTradeUI_UpdateTradeState verifies trade state synchronization.
func TestTradeUI_UpdateTradeState(t *testing.T) {
	world := NewWorld()
	tradeSystem := NewTradeSystem(world)
	ui := NewTradeUI(world, tradeSystem, 1920, 1080)

	// Create player
	player := world.CreateEntity()
	player.AddComponent(&PositionComponent{X: 0, Y: 0})
	playerInv := NewInventoryComponent(10, 100.0)
	player.AddComponent(playerInv)
	ui.SetPlayerEntity(player)

	// Create partner
	partner := world.CreateEntity()
	partner.AddComponent(&PositionComponent{X: 3, Y: 0})
	partnerInv := NewInventoryComponent(10, 100.0)
	partner.AddComponent(partnerInv)

	world.Update(0.0)

	// Propose trade
	err := tradeSystem.ProposeTrade(partner.ID, player.ID, []string{}, []string{})
	if err != nil {
		t.Fatalf("ProposeTrade failed: %v", err)
	}

	// Update UI state (should detect pending trade)
	ui.state = TradeUIStateIdle
	ui.updateTradeState()

	// Should transition to pending state
	if ui.state != TradeUIStatePending {
		t.Errorf("Expected Pending state after update, got %v", ui.state)
	}

	// Accept trade (changes status to "accepted")
	err = tradeSystem.AcceptTrade(player.ID)
	if err != nil {
		t.Fatalf("AcceptTrade failed: %v", err)
	}

	// Update UI state (should detect accepted trade)
	ui.updateTradeState()

	// Should transition to negotiating state
	if ui.state != TradeUIStateNegotiating {
		t.Errorf("Expected Negotiating state after accept, got %v", ui.state)
	}
}

// TestTradeUI_StateStrings verifies state string representations.
func TestTradeUI_StateStrings(t *testing.T) {
	tests := []struct {
		state    TradeUIState
		expected string
	}{
		{TradeUIStateIdle, "Idle"},
		{TradeUIStateSelectingPartner, "Selecting Partner"},
		{TradeUIStateSelectingItems, "Selecting Items"},
		{TradeUIStatePending, "Pending"},
		{TradeUIStateNegotiating, "Negotiating"},
		{TradeUIStateCompleted, "Completed"},
		{TradeUIStateCancelled, "Cancelled"},
		{TradeUIState(999), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			got := tt.state.String()
			if got != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, got)
			}
		})
	}
}

// TestTradeUI_Update verifies update behavior.
func TestTradeUI_Update(t *testing.T) {
	world := NewWorld()
	tradeSystem := NewTradeSystem(world)
	ui := NewTradeUI(world, tradeSystem, 1920, 1080)

	player := world.CreateEntity()
	player.AddComponent(&PositionComponent{X: 0, Y: 0})
	ui.SetPlayerEntity(player)

	world.Update(0.0)

	// Update when not visible (should do nothing)
	ui.Update(0.016)

	// Update when visible
	ui.Open()
	ui.setStatusMessage("Test message", nil)
	ui.messageTimeout = 1.0

	// Update for 0.5 seconds (message should still be visible)
	ui.Update(0.5)
	if ui.messageTimeout <= 0 {
		t.Error("Message timeout should not have expired yet")
	}

	// Update for another 0.6 seconds (message should expire)
	ui.Update(0.6)
	if ui.messageTimeout > 0 {
		t.Error("Message timeout should have expired")
	}
}

// TestTradeUI_IsPlayer verifies player detection.
func TestTradeUI_IsPlayer(t *testing.T) {
	world := NewWorld()
	tradeSystem := NewTradeSystem(world)
	ui := NewTradeUI(world, tradeSystem, 1920, 1080)

	// Entity with input component (player)
	player := world.CreateEntity()
	player.AddComponent(&mockInputComponent{})

	// Entity without input component (NPC)
	npc := world.CreateEntity()

	world.Update(0.0)

	if !ui.isPlayer(player) {
		t.Error("Entity with InputComponent should be detected as player")
	}

	if ui.isPlayer(npc) {
		t.Error("Entity without InputComponent should not be detected as player")
	}
}

// TestTradeUI_GetDistance verifies distance calculation.
func TestTradeUI_GetDistance(t *testing.T) {
	world := NewWorld()
	tradeSystem := NewTradeSystem(world)
	ui := NewTradeUI(world, tradeSystem, 1920, 1080)

	e1 := world.CreateEntity()
	e1.AddComponent(&PositionComponent{X: 0, Y: 0})

	e2 := world.CreateEntity()
	e2.AddComponent(&PositionComponent{X: 3, Y: 4})

	world.Update(0.0)

	// Distance should be 3^2 + 4^2 = 25 (squared distance)
	distance := ui.getDistance(e1, e2)
	if distance != 25.0 {
		t.Errorf("Expected distance 25.0, got %f", distance)
	}
}
