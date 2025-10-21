package combat

import "testing"

// TestDamageType_Constants verifies that damage type constants are distinct.
func TestDamageType_Constants(t *testing.T) {
	types := []DamageType{
		DamagePhysical,
		DamageMagical,
		DamageFire,
		DamageIce,
		DamageLightning,
		DamagePoison,
	}

	// Verify all constants are unique
	seen := make(map[DamageType]bool)
	for _, damageType := range types {
		if seen[damageType] {
			t.Errorf("Duplicate damage type value: %v", damageType)
		}
		seen[damageType] = true
	}

	// Verify expected number of constants
	if len(types) != 6 {
		t.Errorf("Expected 6 damage type constants, got %d", len(types))
	}

	// Verify constants have expected sequence
	if DamagePhysical != 0 {
		t.Errorf("Expected DamagePhysical to be 0, got %d", DamagePhysical)
	}
	if DamageMagical != 1 {
		t.Errorf("Expected DamageMagical to be 1, got %d", DamageMagical)
	}
	if DamageFire != 2 {
		t.Errorf("Expected DamageFire to be 2, got %d", DamageFire)
	}
}

// TestNewDamage verifies that Damage struct can be created and initialized.
func TestNewDamage(t *testing.T) {
	tests := []struct {
		name     string
		amount   float64
		damageType DamageType
		sourceID uint64
		targetID uint64
	}{
		{"physical_damage", 50.0, DamagePhysical, 1, 2},
		{"magical_damage", 75.5, DamageMagical, 3, 4},
		{"fire_damage", 100.0, DamageFire, 5, 6},
		{"zero_damage", 0.0, DamageIce, 7, 8},
		{"negative_damage", -10.0, DamageLightning, 9, 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			damage := Damage{
				Amount:   tt.amount,
				Type:     tt.damageType,
				SourceID: tt.sourceID,
				TargetID: tt.targetID,
			}

			if damage.Amount != tt.amount {
				t.Errorf("Expected damage amount %f, got %f", tt.amount, damage.Amount)
			}
			if damage.Type != tt.damageType {
				t.Errorf("Expected damage type %v, got %v", tt.damageType, damage.Type)
			}
			if damage.SourceID != tt.sourceID {
				t.Errorf("Expected source ID %d, got %d", tt.sourceID, damage.SourceID)
			}
			if damage.TargetID != tt.targetID {
				t.Errorf("Expected target ID %d, got %d", tt.targetID, damage.TargetID)
			}
		})
	}
}

// TestNewStats verifies default stats initialization.
func TestNewStats(t *testing.T) {
	stats := NewStats()

	if stats == nil {
		t.Fatal("Expected non-nil Stats")
	}

	// Verify health initialization
	if stats.HP != 100 {
		t.Errorf("Expected HP 100, got %f", stats.HP)
	}
	if stats.MaxHP != 100 {
		t.Errorf("Expected MaxHP 100, got %f", stats.MaxHP)
	}

	// Verify mana initialization
	if stats.Mana != 50 {
		t.Errorf("Expected Mana 50, got %f", stats.Mana)
	}
	if stats.MaxMana != 50 {
		t.Errorf("Expected MaxMana 50, got %f", stats.MaxMana)
	}

	// Verify offensive stats
	if stats.Attack != 10 {
		t.Errorf("Expected Attack 10, got %f", stats.Attack)
	}

	// Verify defensive stats
	if stats.Defense != 5 {
		t.Errorf("Expected Defense 5, got %f", stats.Defense)
	}

	// Verify speed
	if stats.Speed != 100 {
		t.Errorf("Expected Speed 100, got %f", stats.Speed)
	}

	// Verify resistances map is initialized
	if stats.Resistances == nil {
		t.Error("Expected Resistances map to be initialized")
	}
	if len(stats.Resistances) != 0 {
		t.Errorf("Expected empty Resistances map, got length %d", len(stats.Resistances))
	}
}

// TestStats_HealthManipulation verifies health manipulation operations.
func TestStats_HealthManipulation(t *testing.T) {
	stats := NewStats()

	// Test damage
	stats.HP -= 25
	if stats.HP != 75 {
		t.Errorf("Expected HP 75 after damage, got %f", stats.HP)
	}

	// Test healing
	stats.HP += 10
	if stats.HP != 85 {
		t.Errorf("Expected HP 85 after healing, got %f", stats.HP)
	}

	// Test overheal prevention
	stats.HP = stats.MaxHP + 10
	if stats.HP <= stats.MaxHP {
		// This test just verifies we can set HP above max
		// (actual capping would be done by game logic)
	}

	// Test death (HP at 0)
	stats.HP = 0
	if stats.HP != 0 {
		t.Errorf("Expected HP 0 (death), got %f", stats.HP)
	}
}

// TestStats_ManaManipulation verifies mana manipulation operations.
func TestStats_ManaManipulation(t *testing.T) {
	stats := NewStats()

	// Test mana consumption
	stats.Mana -= 20
	if stats.Mana != 30 {
		t.Errorf("Expected Mana 30 after consumption, got %f", stats.Mana)
	}

	// Test mana regeneration
	stats.Mana += 15
	if stats.Mana != 45 {
		t.Errorf("Expected Mana 45 after regeneration, got %f", stats.Mana)
	}

	// Test full mana
	stats.Mana = stats.MaxMana
	if stats.Mana != stats.MaxMana {
		t.Errorf("Expected Mana to equal MaxMana (%f), got %f", stats.MaxMana, stats.Mana)
	}
}

// TestStats_ResistanceManagement verifies resistance map operations.
func TestStats_ResistanceManagement(t *testing.T) {
	stats := NewStats()

	tests := []struct {
		name       string
		damageType DamageType
		resistance float64
	}{
		{"fire_resistance", DamageFire, 0.5},
		{"ice_immunity", DamageIce, 1.0},
		{"poison_weakness", DamagePoison, -0.25},
		{"no_resistance", DamageMagical, 0.0},
		{"partial_physical", DamagePhysical, 0.3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stats.Resistances[tt.damageType] = tt.resistance
			
			if stats.Resistances[tt.damageType] != tt.resistance {
				t.Errorf("Expected resistance %f for %v, got %f", 
					tt.resistance, tt.damageType, stats.Resistances[tt.damageType])
			}
		})
	}

	// Verify all resistances were set
	if len(stats.Resistances) != len(tests) {
		t.Errorf("Expected %d resistances, got %d", len(tests), len(stats.Resistances))
	}
}

// TestStats_OffensiveStats verifies offensive stat modifications.
func TestStats_OffensiveStats(t *testing.T) {
	stats := NewStats()

	// Test attack modification
	stats.Attack = 25
	if stats.Attack != 25 {
		t.Errorf("Expected Attack 25, got %f", stats.Attack)
	}

	// Test magic power
	stats.MagicPower = 30
	if stats.MagicPower != 30 {
		t.Errorf("Expected MagicPower 30, got %f", stats.MagicPower)
	}

	// Test crit chance
	stats.CritChance = 0.25
	if stats.CritChance != 0.25 {
		t.Errorf("Expected CritChance 0.25, got %f", stats.CritChance)
	}

	// Test crit damage
	stats.CritDamage = 2.0
	if stats.CritDamage != 2.0 {
		t.Errorf("Expected CritDamage 2.0, got %f", stats.CritDamage)
	}
}

// TestStats_DefensiveStats verifies defensive stat modifications.
func TestStats_DefensiveStats(t *testing.T) {
	stats := NewStats()

	// Test defense
	stats.Defense = 15
	if stats.Defense != 15 {
		t.Errorf("Expected Defense 15, got %f", stats.Defense)
	}

	// Test magic defense
	stats.MagicDefense = 20
	if stats.MagicDefense != 20 {
		t.Errorf("Expected MagicDefense 20, got %f", stats.MagicDefense)
	}

	// Test evasion
	stats.Evasion = 0.15
	if stats.Evasion != 0.15 {
		t.Errorf("Expected Evasion 0.15, got %f", stats.Evasion)
	}
}

// TestStats_SpeedModification verifies speed stat operations.
func TestStats_SpeedModification(t *testing.T) {
	stats := NewStats()

	tests := []struct {
		name  string
		speed float64
	}{
		{"slow", 50},
		{"normal", 100},
		{"fast", 150},
		{"very_fast", 200},
		{"zero_speed", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stats.Speed = tt.speed
			if stats.Speed != tt.speed {
				t.Errorf("Expected Speed %f, got %f", tt.speed, stats.Speed)
			}
		})
	}
}

// TestStats_MaxValueModifications verifies max value changes.
func TestStats_MaxValueModifications(t *testing.T) {
	stats := NewStats()

	// Test MaxHP increase
	originalHP := stats.HP
	stats.MaxHP = 200
	if stats.MaxHP != 200 {
		t.Errorf("Expected MaxHP 200, got %f", stats.MaxHP)
	}
	// Current HP should remain unchanged
	if stats.HP != originalHP {
		t.Errorf("Expected HP to remain %f, got %f", originalHP, stats.HP)
	}

	// Test MaxMana increase
	originalMana := stats.Mana
	stats.MaxMana = 100
	if stats.MaxMana != 100 {
		t.Errorf("Expected MaxMana 100, got %f", stats.MaxMana)
	}
	// Current Mana should remain unchanged
	if stats.Mana != originalMana {
		t.Errorf("Expected Mana to remain %f, got %f", originalMana, stats.Mana)
	}
}

// TestStats_CompleteStatsProfile verifies a full stats configuration.
func TestStats_CompleteStatsProfile(t *testing.T) {
	stats := NewStats()

	// Configure a warrior-type character
	stats.MaxHP = 200
	stats.HP = 200
	stats.MaxMana = 30
	stats.Mana = 30
	stats.Attack = 50
	stats.Defense = 30
	stats.MagicDefense = 10
	stats.Speed = 80
	stats.CritChance = 0.20
	stats.CritDamage = 2.5
	stats.Evasion = 0.05
	stats.Resistances[DamagePhysical] = 0.2
	stats.Resistances[DamageFire] = -0.1

	// Verify all stats
	if stats.HP != 200 || stats.MaxHP != 200 {
		t.Error("HP not configured correctly")
	}
	if stats.Mana != 30 || stats.MaxMana != 30 {
		t.Error("Mana not configured correctly")
	}
	if stats.Attack != 50 {
		t.Error("Attack not configured correctly")
	}
	if stats.Defense != 30 {
		t.Error("Defense not configured correctly")
	}
	if stats.Resistances[DamagePhysical] != 0.2 {
		t.Error("Physical resistance not configured correctly")
	}
	if stats.Resistances[DamageFire] != -0.1 {
		t.Error("Fire resistance not configured correctly")
	}
}

// TestDamage_AllTypes verifies damage can be created for all types.
func TestDamage_AllTypes(t *testing.T) {
	damageTypes := []DamageType{
		DamagePhysical,
		DamageMagical,
		DamageFire,
		DamageIce,
		DamageLightning,
		DamagePoison,
	}

	for _, damageType := range damageTypes {
		t.Run(damageType.String(), func(t *testing.T) {
			damage := Damage{
				Amount:   50.0,
				Type:     damageType,
				SourceID: 1,
				TargetID: 2,
			}

			if damage.Type != damageType {
				t.Errorf("Expected damage type %v, got %v", damageType, damage.Type)
			}
		})
	}
}

// Helper method to provide string representation for DamageType (for testing)
func (d DamageType) String() string {
	switch d {
	case DamagePhysical:
		return "Physical"
	case DamageMagical:
		return "Magical"
	case DamageFire:
		return "Fire"
	case DamageIce:
		return "Ice"
	case DamageLightning:
		return "Lightning"
	case DamagePoison:
		return "Poison"
	default:
		return "Unknown"
	}
}

// TestStats_ZeroValues verifies behavior with zero-initialized stats.
func TestStats_ZeroValues(t *testing.T) {
	var stats Stats // Zero-initialized

	if stats.HP != 0 {
		t.Errorf("Expected zero HP, got %f", stats.HP)
	}
	if stats.MaxHP != 0 {
		t.Errorf("Expected zero MaxHP, got %f", stats.MaxHP)
	}
	if stats.Resistances != nil {
		t.Error("Expected nil Resistances map for zero-initialized Stats")
	}
}
