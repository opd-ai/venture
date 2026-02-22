package trade

import (
	"testing"
	"time"

	"github.com/opd-ai/venture/pkg/engine"
	"github.com/opd-ai/venture/pkg/procgen/item"
)

// TestTimeProvider_DeterministicTimestamps verifies that using MockTimeProvider
// produces deterministic timestamps in trade proposals and records.
func TestTimeProvider_DeterministicTimestamps(t *testing.T) {
	fixedTime := time.Date(2026, 2, 22, 12, 0, 0, 0, time.UTC)
	mockClock := NewMockTimeProvider(fixedTime)

	world := engine.NewWorld()
	ts := NewTradeSystemWithTimeProvider(world, mockClock)

	// Create test entities
	item1 := &item.Item{
		ID:     "test_sword",
		Name:   "Test Sword",
		Type:   item.TypeWeapon,
		Rarity: item.RarityCommon,
		Stats:  item.Stats{Weight: 1.0, Value: 100},
		Tags:   []string{},
	}
	item2 := &item.Item{
		ID:     "test_shield",
		Name:   "Test Shield",
		Type:   item.TypeArmor,
		Rarity: item.RarityCommon,
		Stats:  item.Stats{Weight: 2.0, Value: 150},
		Tags:   []string{},
	}

	proposer := world.CreateEntity()
	proposer.AddComponent(&engine.PositionComponent{X: 0, Y: 0})
	propInv := engine.NewInventoryComponent(100, 1000.0)
	propInv.AddItem(item1)
	proposer.AddComponent(propInv)

	recipient := world.CreateEntity()
	recipient.AddComponent(&engine.PositionComponent{X: 1, Y: 1})
	recInv := engine.NewInventoryComponent(100, 1000.0)
	recInv.AddItem(item2)
	recipient.AddComponent(recInv)

	world.Update(0)

	// Propose trade
	err := ts.ProposeTrade(proposer.ID, recipient.ID, []string{"test_sword"}, []string{"test_shield"})
	if err != nil {
		t.Fatalf("ProposeTrade failed: %v", err)
	}

	// Verify proposal timestamp
	tradeCompRaw, ok := proposer.GetComponent("trade")
	if !ok {
		t.Fatal("trade component not found")
	}
	tradeComp := tradeCompRaw.(*engine.TradeComponent)
	if tradeComp.ActiveTrade == nil {
		t.Fatal("no active trade")
	}

	expectedTimestamp := fixedTime.Unix()
	if tradeComp.ActiveTrade.ProposalTime != expectedTimestamp {
		t.Errorf("ProposalTime = %d, want %d", tradeComp.ActiveTrade.ProposalTime, expectedTimestamp)
	}

	// Accept trade (which records history)
	err = ts.AcceptTrade(recipient.ID)
	if err != nil {
		t.Fatalf("AcceptTrade failed: %v", err)
	}

	// Verify trade record timestamp
	proposerEntity, _ := world.GetEntity(proposer.ID)
	tradeCompRaw, _ = proposerEntity.GetComponent("trade")
	tradeComp = tradeCompRaw.(*engine.TradeComponent)

	if len(tradeComp.TradeHistory) == 0 {
		t.Fatal("no trade history recorded")
	}

	record := tradeComp.TradeHistory[0]
	if record.Timestamp != expectedTimestamp {
		t.Errorf("TradeRecord.Timestamp = %d, want %d", record.Timestamp, expectedTimestamp)
	}
}

// TestTimeProvider_TimeoutDeterminism verifies that trade timeouts work
// correctly with MockTimeProvider.
func TestTimeProvider_TimeoutDeterminism(t *testing.T) {
	startTime := time.Date(2026, 2, 22, 12, 0, 0, 0, time.UTC)
	mockClock := NewMockTimeProvider(startTime)

	world := engine.NewWorld()
	ts := NewTradeSystemWithTimeProvider(world, mockClock)

	// Create test entities
	item1 := &item.Item{
		ID:     "timeout_sword",
		Name:   "Timeout Sword",
		Type:   item.TypeWeapon,
		Rarity: item.RarityCommon,
		Stats:  item.Stats{Weight: 1.0, Value: 100},
		Tags:   []string{},
	}
	item2 := &item.Item{
		ID:     "timeout_shield",
		Name:   "Timeout Shield",
		Type:   item.TypeArmor,
		Rarity: item.RarityCommon,
		Stats:  item.Stats{Weight: 2.0, Value: 150},
		Tags:   []string{},
	}

	proposer := world.CreateEntity()
	proposer.AddComponent(&engine.PositionComponent{X: 0, Y: 0})
	propInv := engine.NewInventoryComponent(100, 1000.0)
	propInv.AddItem(item1)
	proposer.AddComponent(propInv)

	recipient := world.CreateEntity()
	recipient.AddComponent(&engine.PositionComponent{X: 1, Y: 1})
	recInv := engine.NewInventoryComponent(100, 1000.0)
	recInv.AddItem(item2)
	recipient.AddComponent(recInv)

	world.Update(0)

	// Propose trade
	err := ts.ProposeTrade(proposer.ID, recipient.ID, []string{"timeout_sword"}, []string{"timeout_shield"})
	if err != nil {
		t.Fatalf("ProposeTrade failed: %v", err)
	}

	// Trade should be active
	if ts.GetActiveTrade(proposer.ID) == nil {
		t.Error("trade should be active before timeout")
	}

	// Advance time by 25 seconds (less than 30s timeout)
	mockClock.Advance(25 * time.Second)
	ts.Update(0.016) // Process timeout checks

	// Trade should still be active
	if ts.GetActiveTrade(proposer.ID) == nil {
		t.Error("trade should still be active after 25 seconds")
	}

	// Advance time by 10 more seconds (total 35s, exceeds 30s timeout)
	mockClock.Advance(10 * time.Second)
	ts.Update(0.016) // Process timeout checks

	// Trade should now be cancelled due to timeout
	if ts.GetActiveTrade(proposer.ID) != nil {
		t.Error("trade should be cancelled after 35 seconds (exceeds 30s timeout)")
	}
}

// TestMockTimeProvider tests the MockTimeProvider implementation.
func TestMockTimeProvider(t *testing.T) {
	baseTime := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	mock := NewMockTimeProvider(baseTime)

	// Test initial time
	if !mock.Now().Equal(baseTime) {
		t.Errorf("initial time = %v, want %v", mock.Now(), baseTime)
	}

	// Test Advance
	mock.Advance(5 * time.Minute)
	expectedAfterAdvance := baseTime.Add(5 * time.Minute)
	if !mock.Now().Equal(expectedAfterAdvance) {
		t.Errorf("time after Advance = %v, want %v", mock.Now(), expectedAfterAdvance)
	}

	// Test Set
	newTime := time.Date(2026, 6, 15, 10, 30, 0, 0, time.UTC)
	mock.Set(newTime)
	if !mock.Now().Equal(newTime) {
		t.Errorf("time after Set = %v, want %v", mock.Now(), newTime)
	}
}

// TestRealTimeProvider tests that RealTimeProvider returns current time.
func TestRealTimeProvider(t *testing.T) {
	real := RealTimeProvider{}

	before := time.Now()
	got := real.Now()
	after := time.Now()

	if got.Before(before) || got.After(after) {
		t.Errorf("RealTimeProvider.Now() = %v, expected between %v and %v", got, before, after)
	}
}

// TestDefaultTimeProvider tests that DefaultTimeProvider returns RealTimeProvider.
func TestDefaultTimeProvider(t *testing.T) {
	provider := DefaultTimeProvider()

	// Should be a RealTimeProvider
	if _, ok := provider.(RealTimeProvider); !ok {
		t.Errorf("DefaultTimeProvider() returned %T, expected RealTimeProvider", provider)
	}
}

// TestNewTradeSystemWithTimeProvider tests the TimeProvider constructor.
func TestNewTradeSystemWithTimeProvider(t *testing.T) {
	world := engine.NewWorld()
	mockClock := NewMockTimeProvider(time.Now())

	ts := NewTradeSystemWithTimeProvider(world, mockClock)

	if ts == nil {
		t.Fatal("NewTradeSystemWithTimeProvider returned nil")
	}
	if ts.clock != mockClock {
		t.Error("clock field not set correctly")
	}
	if ts.world != world {
		t.Error("world field not set correctly")
	}
}
