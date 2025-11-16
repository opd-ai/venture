package federation

import (
	"fmt"
	"testing"
	"time"

	"github.com/opd-ai/venture/pkg/engine"
)

// TestTradeIntegrationCreation tests creating a trade integration system.
func TestTradeIntegrationCreation(t *testing.T) {
	market := NewFederatedMarket()
	world := engine.NewWorld()
	politics := engine.NewPoliticsSystem(world)

	ti := NewTradeIntegration(market, politics)

	if ti == nil {
		t.Fatal("expected non-nil trade integration")
	}

	if ti.market != market {
		t.Error("market not set correctly")
	}

	if ti.politicsSystem != politics {
		t.Error("politics system not set correctly")
	}

	if ti.maxTradesPerWindow != 10 {
		t.Errorf("expected default max trades 10, got %d", ti.maxTradesPerWindow)
	}

	if ti.windowDuration != 60*time.Second {
		t.Errorf("expected default window 60s, got %v", ti.windowDuration)
	}
}

// TestRateLimitEnforcement tests trade rate limiting.
func TestRateLimitEnforcement(t *testing.T) {
	market := NewFederatedMarket()
	world := engine.NewWorld()
	politics := engine.NewPoliticsSystem(world)
	ti := NewTradeIntegration(market, politics)

	playerID := "player1"
	serverID := "server1"

	// Should allow up to 10 trades
	for i := 0; i < 10; i++ {
		err := ti.ValidateTrade(playerID, serverID, 1, 100.0)
		if err != nil {
			t.Errorf("trade %d should succeed: %v", i+1, err)
		}
		ti.RecordTrade(playerID)
	}

	// 11th trade should fail
	err := ti.ValidateTrade(playerID, serverID, 1, 100.0)
	if err == nil {
		t.Error("expected rate limit error on 11th trade")
	}
}

// TestRateLimitWindowReset tests rate limit window reset.
func TestRateLimitWindowReset(t *testing.T) {
	market := NewFederatedMarket()
	world := engine.NewWorld()
	politics := engine.NewPoliticsSystem(world)
	ti := NewTradeIntegration(market, politics)

	// Set short window for testing
	ti.SetWindowDuration(100 * time.Millisecond)

	playerID := "player1"
	serverID := "server1"

	// Max out trades
	for i := 0; i < 10; i++ {
		ti.RecordTrade(playerID)
	}

	// Should fail
	err := ti.ValidateTrade(playerID, serverID, 1, 100.0)
	if err == nil {
		t.Error("expected rate limit error")
	}

	// Wait for window reset
	time.Sleep(150 * time.Millisecond)

	// Should succeed after reset
	err = ti.ValidateTrade(playerID, serverID, 1, 100.0)
	if err != nil {
		t.Errorf("trade should succeed after window reset: %v", err)
	}
}

// TestTradesRemaining tests getting remaining trade count.
func TestTradesRemaining(t *testing.T) {
	market := NewFederatedMarket()
	world := engine.NewWorld()
	politics := engine.NewPoliticsSystem(world)
	ti := NewTradeIntegration(market, politics)

	playerID := "player1"

	// Initially should have 10 remaining
	remaining := ti.GetTradesRemaining(playerID)
	if remaining != 10 {
		t.Errorf("expected 10 remaining, got %d", remaining)
	}

	// After 3 trades
	for i := 0; i < 3; i++ {
		ti.RecordTrade(playerID)
	}

	remaining = ti.GetTradesRemaining(playerID)
	if remaining != 7 {
		t.Errorf("expected 7 remaining, got %d", remaining)
	}
}

// TestReputationLimits tests trade limits based on server reputation.
func TestReputationLimits(t *testing.T) {
	market := NewFederatedMarket()
	world := engine.NewWorld()
	politics := engine.NewPoliticsSystem(world)
	ti := NewTradeIntegration(market, politics)

	playerID := "player1"
	serverID := "low_rep_server"

	tests := []struct {
		name       string
		reputation float64
		itemCount  int
		value      float64
		shouldPass bool
	}{
		{"blocked_low_rep", 0.1, 1, 50.0, false},        // Requires approval
		{"restricted_ok", 0.3, 1, 400.0, true},          // Within limit
		{"restricted_fail_value", 0.3, 1, 600.0, false}, // Exceeds value limit
		{"restricted_fail_count", 0.3, 6, 100.0, false}, // Exceeds item limit
		{"limited_ok", 0.5, 5, 1500.0, true},            // Within limit
		{"limited_fail", 0.5, 15, 500.0, false},         // Exceeds item limit
		{"trusted_ok", 0.7, 15, 8000.0, true},           // Within limit
		{"trusted_fail", 0.7, 25, 5000.0, false},        // Exceeds item limit
		{"verified_ok", 0.9, 40, 50000.0, true},         // Within limit
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset trade count
			ti.mu.Lock()
			ti.tradeCounts = make(map[string]int)
			ti.serverReputation[serverID] = tt.reputation
			ti.mu.Unlock()

			err := ti.ValidateTrade(playerID, serverID, tt.itemCount, tt.value)

			if tt.shouldPass && err != nil {
				t.Errorf("expected trade to pass but got error: %v", err)
			}
			if !tt.shouldPass && err == nil {
				t.Errorf("expected trade to fail but it passed")
			}
		})
	}
}

// TestServerReputationUpdate tests updating server reputation.
func TestServerReputationUpdate(t *testing.T) {
	market := NewFederatedMarket()
	world := engine.NewWorld()
	politics := engine.NewPoliticsSystem(world)
	ti := NewTradeIntegration(market, politics)

	serverID := "server1"

	// Initial should be neutral (0.5)
	rep := ti.GetServerReputation(serverID)
	if rep != 0.5 {
		t.Errorf("expected initial reputation 0.5, got %.2f", rep)
	}

	// Increase reputation
	ti.UpdateServerReputation(serverID, 0.1)
	rep = ti.GetServerReputation(serverID)
	if rep != 0.6 {
		t.Errorf("expected reputation 0.6, got %.2f", rep)
	}

	// Decrease reputation
	ti.UpdateServerReputation(serverID, -0.2)
	rep = ti.GetServerReputation(serverID)
	if rep < 0.39 || rep > 0.41 {
		t.Errorf("expected reputation ~0.4, got %.2f", rep)
	}

	// Test clamping at 0.0
	ti.UpdateServerReputation(serverID, -1.0)
	rep = ti.GetServerReputation(serverID)
	if rep != 0.0 {
		t.Errorf("expected reputation clamped to 0.0, got %.2f", rep)
	}

	// Test clamping at 1.0
	ti.UpdateServerReputation(serverID, 2.0)
	rep = ti.GetServerReputation(serverID)
	if rep != 1.0 {
		t.Errorf("expected reputation clamped to 1.0, got %.2f", rep)
	}
}

// TestAIMerchantBaseline tests AI merchants maintaining supply/demand.
func TestAIMerchantBaseline(t *testing.T) {
	market := NewFederatedMarket()
	world := engine.NewWorld()
	politics := engine.NewPoliticsSystem(world)
	ti := NewTradeIntegration(market, politics)

	itemID := "sword"
	baseSupply := 100
	baseDemand := 50

	// Register item in market
	market.RegisterItem(itemID, "server1", 100.0)

	// Add AI merchant
	ti.AddAIMerchant(itemID, baseSupply, baseDemand, 1*time.Second)

	// Initially supply/demand should be 0
	if market.GetSupply(itemID) != 0 {
		t.Errorf("expected initial supply 0, got %d", market.GetSupply(itemID))
	}

	// Update AI merchants
	ti.UpdateAIMerchants()

	// Should now have baseline supply/demand
	if market.GetSupply(itemID) != baseSupply {
		t.Errorf("expected supply %d, got %d", baseSupply, market.GetSupply(itemID))
	}

	if market.GetDemand(itemID) != baseDemand {
		t.Errorf("expected demand %d, got %d", baseDemand, market.GetDemand(itemID))
	}
}

// TestAIMerchantMaintenance tests AI merchants maintaining baseline after changes.
func TestAIMerchantMaintenance(t *testing.T) {
	market := NewFederatedMarket()
	world := engine.NewWorld()
	politics := engine.NewPoliticsSystem(world)
	ti := NewTradeIntegration(market, politics)

	itemID := "potion"
	baseSupply := 200
	baseDemand := 100

	// Register item
	market.RegisterItem(itemID, "server1", 50.0)

	// Add AI merchant with very short interval
	ti.AddAIMerchant(itemID, baseSupply, baseDemand, 1*time.Millisecond)

	// Initial update
	ti.UpdateAIMerchants()

	// Simulate player purchases (reduce supply)
	market.UpdateSupply(itemID, -50)

	// Wait for update interval
	time.Sleep(10 * time.Millisecond)

	// AI merchant should replenish to baseline
	ti.UpdateAIMerchants()

	supply := market.GetSupply(itemID)
	if supply < baseSupply {
		t.Errorf("expected AI merchant to maintain baseline supply %d, got %d", baseSupply, supply)
	}
}

// TestPoliticsIntegration tests price calculation with political modifiers.
func TestPoliticsIntegration(t *testing.T) {
	market := NewFederatedMarket()
	world := engine.NewWorld()
	politics := engine.NewPoliticsSystem(world)

	// Set up server faction
	faction := &engine.ServerFaction{
		ServerID:    "server_a",
		FactionName: "Test Faction",
	}
	politics.SetServerFaction(faction)

	ti := NewTradeIntegration(market, politics)

	itemID := "weapon"
	basePrice := 100.0

	// Register item
	market.RegisterItem(itemID, "server_a", basePrice)

	// Create alliance with server B (20% discount via 0.8x multiplier)
	_, err := politics.CreateAlliance("server_b", 86400)
	if err != nil {
		t.Fatalf("failed to create alliance: %v", err)
	}

	// Get price with politics
	price, err := ti.GetPriceWithPolitics(itemID, "server_b")
	if err != nil {
		t.Fatalf("failed to get price: %v", err)
	}

	expectedPrice := basePrice * 0.8 // 20% discount
	if price != expectedPrice {
		t.Errorf("expected price %.2f with alliance discount, got %.2f", expectedPrice, price)
	}

	// Declare war on server C (50% markup via 1.5x multiplier)
	_, err = politics.DeclareWar("server_c", 86400)
	if err != nil {
		t.Fatalf("failed to declare war: %v", err)
	}

	price, err = ti.GetPriceWithPolitics(itemID, "server_c")
	if err != nil {
		t.Fatalf("failed to get price: %v", err)
	}

	expectedPrice = basePrice * 1.5 // 50% markup
	if price != expectedPrice {
		t.Errorf("expected price %.2f with war markup, got %.2f", expectedPrice, price)
	}
}

// TestReputationDecay tests slow reputation decay toward neutral.
func TestReputationDecay(t *testing.T) {
	market := NewFederatedMarket()
	world := engine.NewWorld()
	politics := engine.NewPoliticsSystem(world)
	ti := NewTradeIntegration(market, politics)

	serverID := "server1"

	// Set high reputation
	ti.UpdateServerReputation(serverID, 0.5) // Start at 1.0 (0.5 + 0.5)

	// Simulate 1 hour of decay (3600 seconds)
	// Decay rate: 0.01 per hour, so should move 0.01 toward 0.5
	deltaTime := 3600.0
	ti.decayReputationTowardNeutral(deltaTime)

	rep := ti.GetServerReputation(serverID)
	expectedRep := 0.99 // 1.0 - 0.01

	// Allow small floating point error
	if rep < expectedRep-0.01 || rep > expectedRep+0.01 {
		t.Errorf("expected reputation ~%.2f after 1 hour decay, got %.2f", expectedRep, rep)
	}
}

// TestConfigurableRateLimits tests setting custom rate limits.
func TestConfigurableRateLimits(t *testing.T) {
	market := NewFederatedMarket()
	world := engine.NewWorld()
	politics := engine.NewPoliticsSystem(world)
	ti := NewTradeIntegration(market, politics)

	// Set custom limits
	ti.SetMaxTradesPerWindow(5)
	ti.SetWindowDuration(30 * time.Second)

	playerID := "player1"
	serverID := "server1"

	// Should allow up to 5 trades
	for i := 0; i < 5; i++ {
		err := ti.ValidateTrade(playerID, serverID, 1, 100.0)
		if err != nil {
			t.Errorf("trade %d should succeed: %v", i+1, err)
		}
		ti.RecordTrade(playerID)
	}

	// 6th trade should fail
	err := ti.ValidateTrade(playerID, serverID, 1, 100.0)
	if err == nil {
		t.Error("expected rate limit error on 6th trade with custom limit")
	}
}

// TestIntegrationGetStats tests retrieving integration statistics.
func TestIntegrationGetStats(t *testing.T) {
	market := NewFederatedMarket()
	world := engine.NewWorld()
	politics := engine.NewPoliticsSystem(world)
	ti := NewTradeIntegration(market, politics)

	// Add some state
	ti.RecordTrade("player1")
	ti.RecordTrade("player2")
	ti.AddAIMerchant("item1", 100, 50, 5*time.Minute)
	ti.AddAIMerchant("item2", 200, 100, 5*time.Minute)
	ti.UpdateServerReputation("server1", 0.3)
	ti.UpdateServerReputation("server2", 0.7)

	stats := ti.GetStats()

	if stats.ActivePlayers != 2 {
		t.Errorf("expected 2 active players, got %d", stats.ActivePlayers)
	}

	if stats.TotalAIMerchants != 2 {
		t.Errorf("expected 2 AI merchants, got %d", stats.TotalAIMerchants)
	}

	expectedAvgRep := (0.8 + 1.2) / 2 // 0.5+0.3 and 0.5+0.7
	if stats.AverageReputation < expectedAvgRep-0.1 || stats.AverageReputation > expectedAvgRep+0.1 {
		t.Errorf("expected average reputation ~%.2f, got %.2f", expectedAvgRep, stats.AverageReputation)
	}

	if stats.WindowTimeRemaining <= 0 {
		t.Error("expected positive window time remaining")
	}
}

// TestUpdateMethod tests the periodic update method.
func TestUpdateMethod(t *testing.T) {
	market := NewFederatedMarket()
	world := engine.NewWorld()
	politics := engine.NewPoliticsSystem(world)
	ti := NewTradeIntegration(market, politics)

	// Set short update interval for testing
	ti.systemUpdateInterval = 100 * time.Millisecond
	ti.lastAIUpdate = time.Now().Add(-200 * time.Millisecond) // Trigger update

	itemID := "test_item"
	market.RegisterItem(itemID, "server1", 100.0)
	ti.AddAIMerchant(itemID, 100, 50, 1*time.Millisecond)

	// Initial supply should be 0
	if market.GetSupply(itemID) != 0 {
		t.Errorf("expected initial supply 0, got %d", market.GetSupply(itemID))
	}

	// Call update (should trigger AI merchant update)
	ti.Update(0.1) // 0.1 seconds

	// Supply should now be at baseline
	if market.GetSupply(itemID) != 100 {
		t.Errorf("expected supply 100 after update, got %d", market.GetSupply(itemID))
	}
}

// BenchmarkValidateTrade benchmarks trade validation.
func BenchmarkValidateTrade(b *testing.B) {
	market := NewFederatedMarket()
	world := engine.NewWorld()
	politics := engine.NewPoliticsSystem(world)
	ti := NewTradeIntegration(market, politics)

	playerID := "player1"
	serverID := "server1"

	ti.UpdateServerReputation(serverID, 0.3) // Set known reputation

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ti.ValidateTrade(playerID, serverID, 5, 500.0)
	}
}

// BenchmarkUpdateAIMerchants benchmarks AI merchant updates.
func BenchmarkUpdateAIMerchants(b *testing.B) {
	market := NewFederatedMarket()
	world := engine.NewWorld()
	politics := engine.NewPoliticsSystem(world)
	ti := NewTradeIntegration(market, politics)

	// Add multiple AI merchants
	for i := 0; i < 10; i++ {
		itemID := fmt.Sprintf("item%d", i)
		market.RegisterItem(itemID, "server1", 100.0)
		ti.AddAIMerchant(itemID, 100, 50, 1*time.Nanosecond) // Very short interval
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ti.UpdateAIMerchants()
	}
}

// BenchmarkGetPriceWithPolitics benchmarks price calculation with politics.
func BenchmarkGetPriceWithPolitics(b *testing.B) {
	market := NewFederatedMarket()
	world := engine.NewWorld()
	politics := engine.NewPoliticsSystem(world)

	faction := &engine.ServerFaction{
		ServerID:    "server_a",
		FactionName: "Test Faction",
	}
	politics.SetServerFaction(faction)
	politics.CreateAlliance("server_b", 86400)

	ti := NewTradeIntegration(market, politics)

	itemID := "weapon"
	market.RegisterItem(itemID, "server_a", 100.0)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ti.GetPriceWithPolitics(itemID, "server_b")
	}
}
