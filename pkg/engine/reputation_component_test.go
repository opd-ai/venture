package engine

import (
	"testing"
	"time"
)

// TestReputationComponent_Type tests the Type method
func TestReputationComponent_Type(t *testing.T) {
	comp := NewReputationComponent()
	if comp.Type() != "reputation" {
		t.Errorf("Expected type 'reputation', got '%s'", comp.Type())
	}
}

// TestNewReputationComponent tests component initialization
func TestNewReputationComponent(t *testing.T) {
	comp := NewReputationComponent()

	if comp == nil {
		t.Fatal("NewReputationComponent returned nil")
	}

	if comp.Factions == nil {
		t.Error("Factions map should be initialized")
	}

	if comp.KarmaDeeds == nil {
		t.Error("KarmaDeeds slice should be initialized")
	}

	if comp.Alignment.LawAxis != 0.0 {
		t.Errorf("Expected neutral law axis (0.0), got %f", comp.Alignment.LawAxis)
	}

	if comp.Alignment.GoodAxis != 0.0 {
		t.Errorf("Expected neutral good axis (0.0), got %f", comp.Alignment.GoodAxis)
	}
}

// TestAlignment_String tests the alignment string representation
func TestAlignment_String(t *testing.T) {
	tests := []struct {
		name     string
		law      float64
		good     float64
		expected string
	}{
		{"True Neutral", 0.0, 0.0, "True Neutral"},
		{"Lawful Good", 0.8, 0.8, "Lawful Good"},
		{"Lawful Neutral", 0.8, 0.0, "Lawful Neutral"},
		{"Lawful Evil", 0.8, -0.8, "Lawful Evil"},
		{"Neutral Good", 0.0, 0.8, "Neutral Good"},
		{"Neutral Evil", 0.0, -0.8, "Neutral Evil"},
		{"Chaotic Good", -0.8, 0.8, "Chaotic Good"},
		{"Chaotic Neutral", -0.8, 0.0, "Chaotic Neutral"},
		{"Chaotic Evil", -0.8, -0.8, "Chaotic Evil"},
		{"Slight Lawful edge", 0.4, 0.1, "Lawful Neutral"},
		{"Slight Chaotic edge", -0.4, 0.1, "Chaotic Neutral"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			alignment := Alignment{
				LawAxis:  tt.law,
				GoodAxis: tt.good,
			}
			result := alignment.String()
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

// TestReputationComponent_GetReputation tests reputation retrieval
func TestReputationComponent_GetReputation(t *testing.T) {
	comp := NewReputationComponent()

	// Unknown faction should return 0
	if rep := comp.GetReputation("UnknownFaction"); rep != 0.0 {
		t.Errorf("Expected 0 for unknown faction, got %f", rep)
	}

	// Set and retrieve reputation
	comp.Factions["TestFaction"] = 50.0
	if rep := comp.GetReputation("TestFaction"); rep != 50.0 {
		t.Errorf("Expected 50.0, got %f", rep)
	}
}

// TestReputationComponent_SetReputation tests reputation setting with clamping
func TestReputationComponent_SetReputation(t *testing.T) {
	tests := []struct {
		name     string
		faction  string
		value    float64
		expected float64
	}{
		{"Normal value", "Faction1", 50.0, 50.0},
		{"Negative value", "Faction2", -30.0, -30.0},
		{"Max clamp", "Faction3", 150.0, 100.0},
		{"Min clamp", "Faction4", -150.0, -100.0},
		{"Zero", "Faction5", 0.0, 0.0},
	}

	comp := NewReputationComponent()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			comp.SetReputation(tt.faction, tt.value)
			result := comp.GetReputation(tt.faction)
			if result != tt.expected {
				t.Errorf("Expected %f, got %f", tt.expected, result)
			}
		})
	}
}

// TestReputationComponent_AdjustReputation tests reputation adjustment
func TestReputationComponent_AdjustReputation(t *testing.T) {
	comp := NewReputationComponent()

	// Start at neutral (0)
	comp.SetReputation("TestFaction", 0.0)

	// Adjust up
	comp.AdjustReputation("TestFaction", 25.0)
	if rep := comp.GetReputation("TestFaction"); rep != 25.0 {
		t.Errorf("Expected 25.0 after +25 adjustment, got %f", rep)
	}

	// Adjust down
	comp.AdjustReputation("TestFaction", -10.0)
	if rep := comp.GetReputation("TestFaction"); rep != 15.0 {
		t.Errorf("Expected 15.0 after -10 adjustment, got %f", rep)
	}

	// Test clamping
	comp.AdjustReputation("TestFaction", 200.0) // Should clamp to 100
	if rep := comp.GetReputation("TestFaction"); rep != 100.0 {
		t.Errorf("Expected 100.0 (clamped), got %f", rep)
	}

	comp.AdjustReputation("TestFaction", -300.0) // Should clamp to -100
	if rep := comp.GetReputation("TestFaction"); rep != -100.0 {
		t.Errorf("Expected -100.0 (clamped), got %f", rep)
	}
}

// TestReputationComponent_GetReputationTier tests tier classification
func TestReputationComponent_GetReputationTier(t *testing.T) {
	tests := []struct {
		reputation float64
		expected   string
	}{
		{100.0, "Revered"},
		{75.0, "Revered"},
		{74.9, "Honored"},
		{50.0, "Honored"},
		{49.9, "Friendly"},
		{25.0, "Friendly"},
		{24.9, "Neutral"},
		{0.0, "Neutral"},
		{-24.9, "Neutral"},
		{-25.0, "Unfriendly"},
		{-49.9, "Unfriendly"},
		{-50.0, "Hostile"},
		{-74.9, "Hostile"},
		{-75.0, "Hated"},
		{-100.0, "Hated"},
	}

	comp := NewReputationComponent()

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			comp.SetReputation("TestFaction", tt.reputation)
			tier := comp.GetReputationTier("TestFaction")
			if tier != tt.expected {
				t.Errorf("At reputation %f, expected tier '%s', got '%s'", tt.reputation, tt.expected, tier)
			}
		})
	}
}

// TestReputationComponent_IsHostile tests hostility check
func TestReputationComponent_IsHostile(t *testing.T) {
	tests := []struct {
		reputation float64
		expected   bool
	}{
		{0.0, false},
		{25.0, false},
		{-25.0, false},
		{-49.9, false},
		{-50.0, true},
		{-75.0, true},
		{-100.0, true},
	}

	comp := NewReputationComponent()

	for _, tt := range tests {
		comp.SetReputation("TestFaction", tt.reputation)
		result := comp.IsHostile("TestFaction")
		if result != tt.expected {
			t.Errorf("At reputation %f, expected IsHostile=%v, got %v", tt.reputation, tt.expected, result)
		}
	}
}

// TestReputationComponent_IsFriendly tests friendliness check
func TestReputationComponent_IsFriendly(t *testing.T) {
	tests := []struct {
		reputation float64
		expected   bool
	}{
		{0.0, false},
		{24.9, false},
		{25.0, true},
		{50.0, true},
		{75.0, true},
		{100.0, true},
		{-25.0, false},
	}

	comp := NewReputationComponent()

	for _, tt := range tests {
		comp.SetReputation("TestFaction", tt.reputation)
		result := comp.IsFriendly("TestFaction")
		if result != tt.expected {
			t.Errorf("At reputation %f, expected IsFriendly=%v, got %v", tt.reputation, tt.expected, result)
		}
	}
}

// TestReputationComponent_AdjustAlignment tests alignment adjustment
func TestReputationComponent_AdjustAlignment(t *testing.T) {
	comp := NewReputationComponent()

	// Test normal adjustment
	comp.AdjustAlignment(0.3, 0.4)
	if comp.Alignment.LawAxis != 0.3 {
		t.Errorf("Expected LawAxis=0.3, got %f", comp.Alignment.LawAxis)
	}
	if comp.Alignment.GoodAxis != 0.4 {
		t.Errorf("Expected GoodAxis=0.4, got %f", comp.Alignment.GoodAxis)
	}

	// Test clamping to max
	comp.AdjustAlignment(1.0, 1.0) // Should clamp both to 1.0
	if comp.Alignment.LawAxis != 1.0 {
		t.Errorf("Expected LawAxis clamped to 1.0, got %f", comp.Alignment.LawAxis)
	}
	if comp.Alignment.GoodAxis != 1.0 {
		t.Errorf("Expected GoodAxis clamped to 1.0, got %f", comp.Alignment.GoodAxis)
	}

	// Reset and test clamping to min
	comp.Alignment.LawAxis = 0.0
	comp.Alignment.GoodAxis = 0.0
	comp.AdjustAlignment(-1.5, -1.5) // Should clamp both to -1.0
	if comp.Alignment.LawAxis != -1.0 {
		t.Errorf("Expected LawAxis clamped to -1.0, got %f", comp.Alignment.LawAxis)
	}
	if comp.Alignment.GoodAxis != -1.0 {
		t.Errorf("Expected GoodAxis clamped to -1.0, got %f", comp.Alignment.GoodAxis)
	}
}

// TestReputationComponent_RecordDeed tests deed recording
func TestReputationComponent_RecordDeed(t *testing.T) {
	comp := NewReputationComponent()

	deed := Deed{
		Description: "Helped a villager",
		FactionImpact: map[string]float64{
			"Village": 10.0,
			"Bandits": -5.0,
		},
		LawImpact:  0.05,
		GoodImpact: 0.1,
		Location:   "Town Square",
	}

	comp.RecordDeed(deed)

	// Check faction impacts were applied
	if rep := comp.GetReputation("Village"); rep != 10.0 {
		t.Errorf("Expected Village reputation=10.0, got %f", rep)
	}
	if rep := comp.GetReputation("Bandits"); rep != -5.0 {
		t.Errorf("Expected Bandits reputation=-5.0, got %f", rep)
	}

	// Check alignment impacts were applied
	if comp.Alignment.LawAxis != 0.05 {
		t.Errorf("Expected LawAxis=0.05, got %f", comp.Alignment.LawAxis)
	}
	if comp.Alignment.GoodAxis != 0.1 {
		t.Errorf("Expected GoodAxis=0.1, got %f", comp.Alignment.GoodAxis)
	}

	// Check deed was added to history
	if len(comp.KarmaDeeds) != 1 {
		t.Fatalf("Expected 1 deed in history, got %d", len(comp.KarmaDeeds))
	}

	recordedDeed := comp.KarmaDeeds[0]
	if recordedDeed.Description != deed.Description {
		t.Errorf("Expected deed description '%s', got '%s'", deed.Description, recordedDeed.Description)
	}

	if recordedDeed.Timestamp.IsZero() {
		t.Error("Deed timestamp should be set")
	}
}

// TestReputationComponent_GetRecentDeeds tests recent deed retrieval
func TestReputationComponent_GetRecentDeeds(t *testing.T) {
	comp := NewReputationComponent()

	// Add 5 deeds
	for i := 0; i < 5; i++ {
		deed := Deed{
			Description: "Action " + string(rune('A'+i)),
			Timestamp:   time.Now().Add(time.Duration(i) * time.Second),
		}
		comp.RecordDeed(deed)
	}

	// Test getting last 3 deeds
	recent := comp.GetRecentDeeds(3)
	if len(recent) != 3 {
		t.Errorf("Expected 3 recent deeds, got %d", len(recent))
	}

	// Should get deeds C, D, E (last 3)
	if recent[0].Description != "Action C" {
		t.Errorf("Expected first recent deed 'Action C', got '%s'", recent[0].Description)
	}
	if recent[2].Description != "Action E" {
		t.Errorf("Expected last recent deed 'Action E', got '%s'", recent[2].Description)
	}

	// Test requesting more than available
	allDeeds := comp.GetRecentDeeds(10)
	if len(allDeeds) != 5 {
		t.Errorf("Expected 5 deeds (all), got %d", len(allDeeds))
	}

	// Test requesting 0
	noneDeeds := comp.GetRecentDeeds(0)
	if len(noneDeeds) != 0 {
		t.Errorf("Expected 0 deeds, got %d", len(noneDeeds))
	}

	// Test requesting negative
	negDeeds := comp.GetRecentDeeds(-5)
	if len(negDeeds) != 0 {
		t.Errorf("Expected 0 deeds for negative count, got %d", len(negDeeds))
	}
}

// TestReputationComponent_MultipleFactions tests managing multiple factions
func TestReputationComponent_MultipleFactions(t *testing.T) {
	comp := NewReputationComponent()

	factions := map[string]float64{
		"Knights":   50.0,
		"Thieves":   -30.0,
		"Merchants": 25.0,
		"Wizards":   0.0,
	}

	for faction, rep := range factions {
		comp.SetReputation(faction, rep)
	}

	// Verify all were set correctly
	for faction, expected := range factions {
		actual := comp.GetReputation(faction)
		if actual != expected {
			t.Errorf("Faction %s: expected reputation %f, got %f", faction, expected, actual)
		}
	}
}
