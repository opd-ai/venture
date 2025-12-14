package engine

import (
	"testing"
	"time"
)

func TestNewCityEvolutionTriggersComponent(t *testing.T) {
	c := NewCityEvolutionTriggersComponent("city_test")

	if c.CityID != "city_test" {
		t.Errorf("CityID = %v, want city_test", c.CityID)
	}
	if c.PendingTriggers == nil {
		t.Error("PendingTriggers should not be nil")
	}
	if len(c.PendingTriggers) != 0 {
		t.Errorf("PendingTriggers len = %d, want 0", len(c.PendingTriggers))
	}
	if !c.ProcessingEnabled {
		t.Error("ProcessingEnabled should be true by default")
	}
	if c.MaxRecentTriggers != 50 {
		t.Errorf("MaxRecentTriggers = %d, want 50", c.MaxRecentTriggers)
	}
}

func TestCityEvolutionTriggersComponent_Type(t *testing.T) {
	c := NewCityEvolutionTriggersComponent("test")
	if got := c.Type(); got != "city_evolution_triggers" {
		t.Errorf("Type() = %v, want city_evolution_triggers", got)
	}
}

func TestCityEvolutionTriggersComponent_QueueAndPop(t *testing.T) {
	c := NewCityEvolutionTriggersComponent("test")

	// Queue some triggers
	trigger1 := EvolutionTrigger{
		TriggerType:    EvolutionTradeComplete,
		Magnitude:      0.5,
		SourceEntityID: "player_1",
	}
	trigger2 := EvolutionTrigger{
		TriggerType:    EvolutionQuestComplete,
		Magnitude:      1.0,
		SourceEntityID: "player_2",
	}

	c.QueueTrigger(trigger1)
	c.QueueTrigger(trigger2)

	if !c.HasPendingTriggers() {
		t.Error("HasPendingTriggers() should be true after queueing")
	}
	if c.GetPendingCount() != 2 {
		t.Errorf("GetPendingCount() = %d, want 2", c.GetPendingCount())
	}

	// Pop first trigger
	popped := c.PopTrigger()
	if popped == nil {
		t.Fatal("PopTrigger() returned nil")
	}
	if popped.TriggerType != EvolutionTradeComplete {
		t.Errorf("First pop type = %v, want %v", popped.TriggerType, EvolutionTradeComplete)
	}

	// Pop second trigger
	popped = c.PopTrigger()
	if popped == nil {
		t.Fatal("PopTrigger() returned nil")
	}
	if popped.TriggerType != EvolutionQuestComplete {
		t.Errorf("Second pop type = %v, want %v", popped.TriggerType, EvolutionQuestComplete)
	}

	// Pop empty queue
	popped = c.PopTrigger()
	if popped != nil {
		t.Error("PopTrigger() should return nil when empty")
	}

	if c.HasPendingTriggers() {
		t.Error("HasPendingTriggers() should be false after popping all")
	}
}

func TestCityEvolutionTriggersComponent_QueueSetsTimestamp(t *testing.T) {
	c := NewCityEvolutionTriggersComponent("test")

	before := time.Now().Add(-time.Millisecond)
	trigger := EvolutionTrigger{
		TriggerType: EvolutionTradeComplete,
		Magnitude:   1.0,
	}
	c.QueueTrigger(trigger)
	after := time.Now().Add(time.Millisecond)

	popped := c.PopTrigger()
	if popped.Timestamp.Before(before) || popped.Timestamp.After(after) {
		t.Error("QueueTrigger should set timestamp when zero")
	}
}

func TestCityEvolutionTriggersComponent_RecordProcessed(t *testing.T) {
	c := NewCityEvolutionTriggersComponent("test")

	for i := 0; i < 60; i++ {
		trigger := EvolutionTrigger{
			TriggerType: EvolutionTradeComplete,
			Magnitude:   float64(i) / 60.0,
		}
		c.RecordProcessed(trigger)
	}

	// Should be trimmed to max
	if len(c.RecentTriggers) != c.MaxRecentTriggers {
		t.Errorf("RecentTriggers len = %d, want %d", len(c.RecentTriggers), c.MaxRecentTriggers)
	}

	// First should be the 11th trigger (after trimming first 10)
	if c.RecentTriggers[0].Magnitude < 0.1 {
		t.Error("Old triggers should be trimmed from history")
	}
}

func TestCityEvolutionTriggersComponent_GetRecentTriggersByType(t *testing.T) {
	c := NewCityEvolutionTriggersComponent("test")

	c.RecordProcessed(EvolutionTrigger{TriggerType: EvolutionTradeComplete, Magnitude: 1.0})
	c.RecordProcessed(EvolutionTrigger{TriggerType: EvolutionQuestComplete, Magnitude: 1.0})
	c.RecordProcessed(EvolutionTrigger{TriggerType: EvolutionTradeComplete, Magnitude: 0.5})
	c.RecordProcessed(EvolutionTrigger{TriggerType: EvolutionRaidAttack, Magnitude: 1.0})

	trades := c.GetRecentTriggersByType(EvolutionTradeComplete)
	if len(trades) != 2 {
		t.Errorf("GetRecentTriggersByType(trade) len = %d, want 2", len(trades))
	}

	raids := c.GetRecentTriggersByType(EvolutionRaidAttack)
	if len(raids) != 1 {
		t.Errorf("GetRecentTriggersByType(raid) len = %d, want 1", len(raids))
	}

	guards := c.GetRecentTriggersByType(EvolutionGuardHired)
	if len(guards) != 0 {
		t.Errorf("GetRecentTriggersByType(guard) len = %d, want 0", len(guards))
	}
}

func TestCityEvolutionTriggersComponent_ClearPending(t *testing.T) {
	c := NewCityEvolutionTriggersComponent("test")

	c.QueueTrigger(EvolutionTrigger{TriggerType: EvolutionTradeComplete, Magnitude: 1.0})
	c.QueueTrigger(EvolutionTrigger{TriggerType: EvolutionQuestComplete, Magnitude: 1.0})

	if c.GetPendingCount() != 2 {
		t.Error("Should have 2 pending triggers")
	}

	c.ClearPending()

	if c.GetPendingCount() != 0 {
		t.Errorf("GetPendingCount() = %d after clear, want 0", c.GetPendingCount())
	}
	if c.HasPendingTriggers() {
		t.Error("HasPendingTriggers() should be false after clear")
	}
}

func TestCityEvolutionTriggersComponent_Serialization(t *testing.T) {
	original := NewCityEvolutionTriggersComponent("city_serialize")
	original.QueueTrigger(EvolutionTrigger{
		TriggerType:    EvolutionBuildingConstructed,
		Magnitude:      0.8,
		SourceEntityID: "builder_1",
		Timestamp:      time.Now(),
	})
	original.RecordProcessed(EvolutionTrigger{
		TriggerType: EvolutionTradeComplete,
		Magnitude:   0.5,
	})
	original.ProcessingEnabled = false

	// Serialize
	data, err := original.Serialize()
	if err != nil {
		t.Fatalf("Serialize() error = %v", err)
	}

	// Deserialize
	restored := &CityEvolutionTriggersComponent{}
	err = restored.Deserialize(data)
	if err != nil {
		t.Fatalf("Deserialize() error = %v", err)
	}

	if restored.CityID != original.CityID {
		t.Errorf("CityID = %v, want %v", restored.CityID, original.CityID)
	}
	if restored.ProcessingEnabled != original.ProcessingEnabled {
		t.Errorf("ProcessingEnabled = %v, want %v", restored.ProcessingEnabled, original.ProcessingEnabled)
	}
	if len(restored.PendingTriggers) != len(original.PendingTriggers) {
		t.Errorf("PendingTriggers len = %d, want %d", len(restored.PendingTriggers), len(original.PendingTriggers))
	}
	if len(restored.RecentTriggers) != len(original.RecentTriggers) {
		t.Errorf("RecentTriggers len = %d, want %d", len(restored.RecentTriggers), len(original.RecentTriggers))
	}
}

func TestCityEvolutionTriggersComponent_Deserialize_InvalidData(t *testing.T) {
	c := &CityEvolutionTriggersComponent{}
	err := c.Deserialize([]byte("not json"))
	if err == nil {
		t.Error("Deserialize() expected error for invalid JSON")
	}
}

func TestGetTriggerImpact(t *testing.T) {
	tests := []struct {
		name        string
		triggerType EvolutionTriggerType
		magnitude   float64
		checkField  string
		wantNonZero bool
	}{
		{"trade prosperity", EvolutionTradeComplete, 1.0, "prosperity", true},
		{"trade resources", EvolutionTradeComplete, 1.0, "resources", true},
		{"raid damage", EvolutionRaidAttack, 1.0, "prosperity", true},
		{"raid population loss", EvolutionRaidAttack, 1.0, "population", true},
		{"building infrastructure", EvolutionBuildingConstructed, 1.0, "infrastructure", true},
		{"guard defense", EvolutionGuardHired, 1.0, "defense", true},
		{"half magnitude", EvolutionTradeComplete, 0.5, "prosperity", true},
		{"zero magnitude", EvolutionTradeComplete, 0.0, "prosperity", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			impact := GetTriggerImpact(tt.triggerType, tt.magnitude)

			var value float64
			switch tt.checkField {
			case "prosperity":
				value = impact.ProsperityDelta
			case "infrastructure":
				value = impact.InfrastructureDelta
			case "defense":
				value = impact.DefenseDelta
			case "population":
				value = float64(impact.PopulationDelta)
			case "resources":
				value = impact.ResourceDelta
			}

			if tt.wantNonZero && value == 0 {
				t.Errorf("%s should be non-zero for %v", tt.checkField, tt.triggerType)
			}
			if !tt.wantNonZero && value != 0 {
				t.Errorf("%s should be zero for magnitude 0", tt.checkField)
			}
		})
	}
}

func TestGetTriggerImpact_UnknownType(t *testing.T) {
	impact := GetTriggerImpact(EvolutionTriggerType("unknown"), 1.0)

	if impact.ProsperityDelta != 0 ||
		impact.InfrastructureDelta != 0 ||
		impact.DefenseDelta != 0 ||
		impact.PopulationDelta != 0 ||
		impact.ResourceDelta != 0 {
		t.Error("Unknown trigger type should return zero impact")
	}
}

func TestGetTriggerImpact_AllTriggerTypes(t *testing.T) {
	triggerTypes := []EvolutionTriggerType{
		EvolutionTradeComplete,
		EvolutionQuestComplete,
		EvolutionRaidAttack,
		EvolutionRaidDefended,
		EvolutionBuildingConstructed,
		EvolutionBuildingDestroyed,
		EvolutionPopulationArrived,
		EvolutionPopulationDeparted,
		EvolutionResourceDonation,
		EvolutionResourceShortage,
		EvolutionGuardHired,
		EvolutionGuardLost,
	}

	for _, tt := range triggerTypes {
		t.Run(string(tt), func(t *testing.T) {
			impact := GetTriggerImpact(tt, 1.0)

			// Each trigger should have at least one non-zero impact
			hasImpact := impact.ProsperityDelta != 0 ||
				impact.InfrastructureDelta != 0 ||
				impact.DefenseDelta != 0 ||
				impact.PopulationDelta != 0 ||
				impact.ResourceDelta != 0

			if !hasImpact {
				t.Errorf("Trigger type %v should have at least one impact", tt)
			}
		})
	}
}
