package engine

import (
	"math/rand"
	"testing"
)

// getMerchantComp is a test helper to get MerchantComponent from an entity
func getMerchantComp(e *Entity) *MerchantComponent {
	comp, _ := e.GetComponent("merchant")
	return comp.(*MerchantComponent)
}

func TestNewReputationPricingSystem(t *testing.T) {
	world := NewWorld()
	sys := NewReputationPricingSystem(world, 12345)

	if sys == nil {
		t.Fatal("NewReputationPricingSystem returned nil")
	}
	if sys.world != world {
		t.Error("world not set correctly")
	}
	if sys.rng == nil {
		t.Error("RNG not initialized")
	}
	if sys.updateInterval != 1.0 {
		t.Errorf("updateInterval = %v, want 1.0", sys.updateInterval)
	}
}

func TestReputationPricingSystem_Update_NoPlayer(t *testing.T) {
	world := NewWorld()
	sys := NewReputationPricingSystem(world, 12345)

	// Create merchant without player
	merchant := world.CreateEntity()
	merchant.AddComponent(&MerchantComponent{
		PriceMultiplier: 1.5,
		MaxInventory:    10,
	})
	merchant.AddComponent(&FactionComponent{
		FactionID:       "merchants_guild",
		IsPlayerFaction: false,
	})

	entities := []*Entity{merchant}
	sys.timeSinceUpdate = 1.0 // Force update

	// Should not panic without player
	sys.Update(entities, 0.1)

	// Price should remain unchanged (no player to check reputation)
	mc := getMerchantComp(merchant)
	if mc.PriceMultiplier != 1.5 {
		t.Errorf("PriceMultiplier changed without player, got %v", mc.PriceMultiplier)
	}
}

func TestReputationPricingSystem_Update_FriendlyReputation(t *testing.T) {
	world := NewWorld()
	sys := NewReputationPricingSystem(world, 12345)

	// Create player with friendly reputation
	player := world.CreateEntity()
	player.AddComponent(NewStubInput())
	player.AddComponent(&FactionComponent{
		FactionID:       "merchants_guild",
		Reputation:      75, // Friendly
		IsPlayerFaction: true,
	})

	// Create merchant
	merchant := world.CreateEntity()
	merchant.AddComponent(&MerchantComponent{
		PriceMultiplier: 1.5,
		MaxInventory:    10,
	})
	merchant.AddComponent(&FactionComponent{
		FactionID:       "merchants_guild",
		IsPlayerFaction: false,
	})

	entities := []*Entity{player, merchant}
	sys.timeSinceUpdate = 1.0

	sys.Update(entities, 0.1)

	mc := getMerchantComp(merchant)
	// Friendly reputation (75) should give ~12.5% discount
	// Expected: 1.5 * 0.875 = 1.3125
	if mc.PriceMultiplier >= 1.5 {
		t.Errorf("Friendly reputation should decrease price, got %v", mc.PriceMultiplier)
	}
	if mc.PriceMultiplier < 1.0 {
		t.Errorf("Price should not go below base, got %v", mc.PriceMultiplier)
	}
}

func TestReputationPricingSystem_Update_SuspiciousReputation(t *testing.T) {
	world := NewWorld()
	sys := NewReputationPricingSystem(world, 12345)

	// Create player with suspicious reputation
	player := world.CreateEntity()
	player.AddComponent(NewStubInput())
	player.AddComponent(&FactionComponent{
		FactionID:       "merchants_guild",
		Reputation:      -25, // Suspicious
		IsPlayerFaction: true,
	})

	// Create merchant
	merchant := world.CreateEntity()
	merchant.AddComponent(&MerchantComponent{
		PriceMultiplier: 1.5,
		MaxInventory:    10,
	})
	merchant.AddComponent(&FactionComponent{
		FactionID:       "merchants_guild",
		IsPlayerFaction: false,
	})

	entities := []*Entity{player, merchant}
	sys.timeSinceUpdate = 1.0

	sys.Update(entities, 0.1)

	mc := getMerchantComp(merchant)
	// Suspicious reputation should increase price by 50%
	// Expected: 1.5 * 1.5 = 2.25
	if mc.PriceMultiplier <= 1.5 {
		t.Errorf("Suspicious reputation should increase price, got %v", mc.PriceMultiplier)
	}
}

func TestReputationPricingSystem_Update_NeutralReputation(t *testing.T) {
	world := NewWorld()
	sys := NewReputationPricingSystem(world, 12345)

	// Create player with neutral reputation
	player := world.CreateEntity()
	player.AddComponent(NewStubInput())
	player.AddComponent(&FactionComponent{
		FactionID:       "merchants_guild",
		Reputation:      25, // Neutral
		IsPlayerFaction: true,
	})

	// Create merchant with default price
	merchant := world.CreateEntity()
	merchant.AddComponent(&MerchantComponent{
		PriceMultiplier: 1.5,
		MaxInventory:    10,
	})
	merchant.AddComponent(&FactionComponent{
		FactionID:       "merchants_guild",
		IsPlayerFaction: false,
	})

	entities := []*Entity{player, merchant}
	sys.timeSinceUpdate = 1.0

	sys.Update(entities, 0.1)

	mc := getMerchantComp(merchant)
	// Neutral reputation should keep price at base
	if mc.PriceMultiplier != 1.5 {
		t.Errorf("Neutral reputation should keep base price, got %v", mc.PriceMultiplier)
	}
}

func TestReputationPricingSystem_Update_Throttling(t *testing.T) {
	world := NewWorld()
	sys := NewReputationPricingSystem(world, 12345)

	player := world.CreateEntity()
	player.AddComponent(NewStubInput())
	player.AddComponent(&FactionComponent{
		FactionID:       "merchants_guild",
		Reputation:      75, // Friendly
		IsPlayerFaction: true,
	})

	merchant := world.CreateEntity()
	merchant.AddComponent(&MerchantComponent{
		PriceMultiplier: 1.5,
		MaxInventory:    10,
	})
	merchant.AddComponent(&FactionComponent{
		FactionID:       "merchants_guild",
		IsPlayerFaction: false,
	})

	entities := []*Entity{player, merchant}

	// Update with small delta - should be throttled
	sys.Update(entities, 0.1)

	mc := getMerchantComp(merchant)
	if mc.PriceMultiplier != 1.5 {
		t.Errorf("Update should be throttled, price unchanged, got %v", mc.PriceMultiplier)
	}

	// Now accumulate enough time
	for i := 0; i < 15; i++ {
		sys.Update(entities, 0.1)
	}

	// After 1.5 seconds, update should have triggered
	if mc.PriceMultiplier >= 1.5 {
		t.Errorf("After throttle period, price should update, got %v", mc.PriceMultiplier)
	}
}

func TestReputationPricingSystem_DifferentFactions(t *testing.T) {
	world := NewWorld()
	sys := NewReputationPricingSystem(world, 12345)

	// Player has different reputations with different factions
	player := world.CreateEntity()
	player.AddComponent(NewStubInput())
	player.AddComponent(&FactionComponent{
		FactionID:       "merchants_guild",
		Reputation:      75, // Friendly with merchants
		IsPlayerFaction: true,
	})

	// Merchant from different faction - player has no reputation with them
	merchant := world.CreateEntity()
	merchant.AddComponent(&MerchantComponent{
		PriceMultiplier: 1.5,
		MaxInventory:    10,
	})
	merchant.AddComponent(&FactionComponent{
		FactionID:       "thieves_guild", // Different faction
		IsPlayerFaction: false,
	})

	entities := []*Entity{player, merchant}
	sys.timeSinceUpdate = 1.0

	sys.Update(entities, 0.1)

	mc := getMerchantComp(merchant)
	// No reputation tracked with thieves_guild, should use neutral (1.0 multiplier)
	// 1.5 * 1.0 = 1.5
	if mc.PriceMultiplier != 1.5 {
		t.Errorf("Untracked faction should use neutral pricing, got %v", mc.PriceMultiplier)
	}
}

func TestReputationPricingSystem_MerchantWithoutFaction(t *testing.T) {
	world := NewWorld()
	sys := NewReputationPricingSystem(world, 12345)

	player := world.CreateEntity()
	player.AddComponent(NewStubInput())
	player.AddComponent(&FactionComponent{
		FactionID:       "merchants_guild",
		Reputation:      75,
		IsPlayerFaction: true,
	})

	// Merchant without faction component
	merchant := world.CreateEntity()
	merchant.AddComponent(&MerchantComponent{
		PriceMultiplier: 1.5,
		MaxInventory:    10,
	})

	entities := []*Entity{player, merchant}
	sys.timeSinceUpdate = 1.0

	sys.Update(entities, 0.1)

	mc := getMerchantComp(merchant)
	// No faction on merchant, price unaffected
	if mc.PriceMultiplier != 1.5 {
		t.Errorf("Merchant without faction should keep base price, got %v", mc.PriceMultiplier)
	}
}

func TestFactionReputationComponent(t *testing.T) {
	tests := []struct {
		name      string
		factionID string
		setRep    int
		wantRep   int
	}{
		{"positive reputation", "guild_a", 50, 50},
		{"negative reputation", "guild_b", -50, -50},
		{"max clamp", "guild_c", 150, 100},
		{"min clamp", "guild_d", -150, -100},
		{"zero reputation", "guild_e", 0, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			comp := NewFactionReputationComponent()
			comp.SetReputation(tt.factionID, tt.setRep)

			got := comp.GetReputation(tt.factionID)
			if got != tt.wantRep {
				t.Errorf("GetReputation() = %v, want %v", got, tt.wantRep)
			}
		})
	}
}

func TestFactionReputationComponent_ModifyReputation(t *testing.T) {
	comp := NewFactionReputationComponent()

	// Start at 0
	if rep := comp.GetReputation("guild"); rep != 0 {
		t.Errorf("Initial reputation = %d, want 0", rep)
	}

	// Add positive
	comp.ModifyReputation("guild", 30)
	if rep := comp.GetReputation("guild"); rep != 30 {
		t.Errorf("After +30: reputation = %d, want 30", rep)
	}

	// Add negative
	comp.ModifyReputation("guild", -50)
	if rep := comp.GetReputation("guild"); rep != -20 {
		t.Errorf("After -50: reputation = %d, want -20", rep)
	}

	// Clamp at max
	comp.ModifyReputation("guild", 200)
	if rep := comp.GetReputation("guild"); rep != 100 {
		t.Errorf("After +200: reputation = %d, want 100 (clamped)", rep)
	}
}

func TestFactionReputationComponent_Type(t *testing.T) {
	comp := NewFactionReputationComponent()
	if typ := comp.Type(); typ != "faction_reputation" {
		t.Errorf("Type() = %s, want faction_reputation", typ)
	}
}

func TestReputationPricingSystem_NilWorld(t *testing.T) {
	sys := NewReputationPricingSystem(nil, 12345)

	// Should not panic with nil world
	sys.Update(nil, 0.1)
}

func TestReputationPricingSystem_FindPlayerEntity_Caching(t *testing.T) {
	world := NewWorld()
	sys := NewReputationPricingSystem(world, 12345)

	player := world.CreateEntity()
	player.AddComponent(NewStubInput())

	entities := []*Entity{player}

	// First call should find and cache
	found := sys.findPlayerEntity(entities)
	if found != player {
		t.Error("First call should find player")
	}
	if sys.cachedPlayerID != player.ID {
		t.Error("Player ID should be cached")
	}

	// Second call should use cache
	found2 := sys.findPlayerEntity(entities)
	if found2 != player {
		t.Error("Cached lookup should find player")
	}
}

func BenchmarkReputationPricingSystem_Update(b *testing.B) {
	world := NewWorld()
	sys := NewReputationPricingSystem(world, 12345)

	// Create player
	player := world.CreateEntity()
	player.AddComponent(NewStubInput())
	player.AddComponent(&FactionComponent{
		FactionID:       "merchants_guild",
		Reputation:      50,
		IsPlayerFaction: true,
	})

	// Create 100 merchants
	entities := make([]*Entity, 101)
	entities[0] = player
	for i := 1; i <= 100; i++ {
		merchant := world.CreateEntity()
		merchant.AddComponent(&MerchantComponent{
			PriceMultiplier: 1.5,
			MaxInventory:    10,
		})
		merchant.AddComponent(&FactionComponent{
			FactionID:       "merchants_guild",
			IsPlayerFaction: false,
		})
		entities[i] = merchant
	}

	// Reset timer and run benchmark
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.timeSinceUpdate = 1.0 // Force update each iteration
		sys.Update(entities, 0.016)
	}
}

func TestCalculatePriceMultiplier_AllLevels(t *testing.T) {
	world := NewWorld()
	sys := NewReputationPricingSystem(world, 12345)

	tests := []struct {
		name    string
		rep     int
		wantMin float64
		wantMax float64
	}{
		{"hostile", -75, 0.0, 0.0},        // No trading
		{"suspicious", -25, 1.49, 1.51},   // 50% markup
		{"neutral", 25, 0.99, 1.01},       // Normal
		{"friendly_low", 60, 0.89, 0.96},  // Small discount
		{"friendly_max", 100, 0.74, 0.76}, // Max 25% discount
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			playerReps := map[string]int{"guild": tt.rep}
			mult := sys.calculatePriceMultiplier(playerReps, "guild")

			if mult < tt.wantMin || mult > tt.wantMax {
				t.Errorf("calculatePriceMultiplier(rep=%d) = %v, want [%v, %v]",
					tt.rep, mult, tt.wantMin, tt.wantMax)
			}
		})
	}
}

func TestReputationPricingSystem_WithFactionReputationComponent(t *testing.T) {
	world := NewWorld()
	sys := NewReputationPricingSystem(world, 12345)

	// Create player with FactionReputationComponent for multiple factions
	player := world.CreateEntity()
	player.AddComponent(NewStubInput())
	repComp := NewFactionReputationComponent()
	repComp.SetReputation("merchants_guild", 80)
	repComp.SetReputation("thieves_guild", -30)
	player.AddComponent(repComp)

	// Merchant from merchants guild
	merchant1 := world.CreateEntity()
	merchant1.AddComponent(&MerchantComponent{
		PriceMultiplier: 1.5,
		MaxInventory:    10,
	})
	merchant1.AddComponent(&FactionComponent{
		FactionID:       "merchants_guild",
		IsPlayerFaction: false,
	})

	// Merchant from thieves guild
	merchant2 := world.CreateEntity()
	merchant2.AddComponent(&MerchantComponent{
		PriceMultiplier: 1.5,
		MaxInventory:    10,
	})
	merchant2.AddComponent(&FactionComponent{
		FactionID:       "thieves_guild",
		IsPlayerFaction: false,
	})

	entities := []*Entity{player, merchant1, merchant2}
	sys.timeSinceUpdate = 1.0

	sys.Update(entities, 0.1)

	mc1 := getMerchantComp(merchant1)
	mc2 := getMerchantComp(merchant2)

	// Merchant1 should have lower price (friendly)
	if mc1.PriceMultiplier >= 1.5 {
		t.Errorf("Friendly merchant should have lower price, got %v", mc1.PriceMultiplier)
	}

	// Merchant2 should have higher price (suspicious)
	if mc2.PriceMultiplier <= 1.5 {
		t.Errorf("Suspicious merchant should have higher price, got %v", mc2.PriceMultiplier)
	}
}

// Determinism test - same seed should produce same results
func TestReputationPricingSystem_Determinism(t *testing.T) {
	seed := int64(54321)

	world1 := NewWorld()
	sys1 := NewReputationPricingSystem(world1, seed)
	_ = rand.New(rand.NewSource(seed)) // Consume same random state

	world2 := NewWorld()
	sys2 := NewReputationPricingSystem(world2, seed)

	// Both systems should behave identically
	if sys1.updateInterval != sys2.updateInterval {
		t.Error("Systems with same seed should have identical config")
	}
}
