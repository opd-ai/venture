package combat

import (
	"math"
	"testing"
)

// MockEntityProvider implements EntityStatsProvider for testing.
type MockEntityProvider struct {
	stats        map[uint64]*Stats
	attackDamage map[uint64]Damage
}

func NewMockEntityProvider() *MockEntityProvider {
	return &MockEntityProvider{
		stats:        make(map[uint64]*Stats),
		attackDamage: make(map[uint64]Damage),
	}
}

func (m *MockEntityProvider) GetStats(entityID uint64) *Stats {
	return m.stats[entityID]
}

func (m *MockEntityProvider) GetAttackDamage(entityID uint64) Damage {
	if d, ok := m.attackDamage[entityID]; ok {
		return d
	}
	// Default attack if not configured
	return Damage{Amount: 10, Type: DamagePhysical}
}

func (m *MockEntityProvider) SetStats(entityID uint64, stats *Stats) {
	m.stats[entityID] = stats
}

func (m *MockEntityProvider) SetAttackDamage(entityID uint64, damage Damage) {
	m.attackDamage[entityID] = damage
}

// TestNewDefaultCombatResolver verifies resolver creation.
func TestNewDefaultCombatResolver(t *testing.T) {
	provider := NewMockEntityProvider()
	resolver := NewDefaultCombatResolver(provider)

	if resolver == nil {
		t.Fatal("Expected non-nil resolver")
	}

	if resolver.MinDamageMultiplier != 0.1 {
		t.Errorf("Expected MinDamageMultiplier 0.1, got %f", resolver.MinDamageMultiplier)
	}

	if resolver.EntityLookup != provider {
		t.Error("EntityLookup not set correctly")
	}
}

// TestCalculateDamage_NilStats verifies damage passes through with nil stats.
func TestCalculateDamage_NilStats(t *testing.T) {
	resolver := NewDefaultCombatResolver(nil)
	damage := Damage{Amount: 100, Type: DamagePhysical}

	result := resolver.CalculateDamage(damage, nil)

	if result != 100 {
		t.Errorf("Expected 100 damage with nil stats, got %f", result)
	}
}

// TestCalculateDamage_ZeroOrNegativeDamage verifies edge cases.
func TestCalculateDamage_ZeroOrNegativeDamage(t *testing.T) {
	resolver := NewDefaultCombatResolver(nil)
	stats := NewStats()

	tests := []struct {
		name   string
		amount float64
		want   float64
	}{
		{"zero damage", 0, 0},
		{"negative damage", -10, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			damage := Damage{Amount: tt.amount, Type: DamagePhysical}
			result := resolver.CalculateDamage(damage, stats)
			if result != tt.want {
				t.Errorf("Expected %f, got %f", tt.want, result)
			}
		})
	}
}

// TestCalculateDamage_DefenseReduction verifies defense mitigation.
func TestCalculateDamage_DefenseReduction(t *testing.T) {
	resolver := NewDefaultCombatResolver(nil)

	tests := []struct {
		name     string
		amount   float64
		defense  float64
		wantLess float64 // result should be less than this
		wantMore float64 // result should be more than this
	}{
		{"no defense", 100, 0, 101, 99},
		{"moderate defense", 100, 50, 75, 65}, // 100 * (100/150) ≈ 66.67
		{"high defense", 100, 100, 55, 45},    // 100 * (100/200) = 50
		{"extreme defense", 100, 500, 20, 15}, // 100 * (100/600) ≈ 16.67
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stats := NewStats()
			stats.Defense = tt.defense
			damage := Damage{Amount: tt.amount, Type: DamagePhysical}

			result := resolver.CalculateDamage(damage, stats)

			if result >= tt.wantLess {
				t.Errorf("Expected result < %f, got %f", tt.wantLess, result)
			}
			if result <= tt.wantMore {
				t.Errorf("Expected result > %f, got %f", tt.wantMore, result)
			}
		})
	}
}

// TestCalculateDamage_MagicDefense verifies magical damage uses MagicDefense.
func TestCalculateDamage_MagicDefense(t *testing.T) {
	resolver := NewDefaultCombatResolver(nil)

	magicTypes := []DamageType{DamageMagical, DamageFire, DamageIce, DamageLightning, DamagePoison}

	for _, damageType := range magicTypes {
		t.Run(damageType.String(), func(t *testing.T) {
			stats := NewStats()
			stats.Defense = 0        // High physical defense
			stats.MagicDefense = 100 // Should use this

			damage := Damage{Amount: 100, Type: damageType}
			result := resolver.CalculateDamage(damage, stats)

			// With 100 magic defense: 100 * (100/200) = 50
			if result > 55 || result < 45 {
				t.Errorf("Expected ~50 damage with 100 magic defense, got %f", result)
			}
		})
	}
}

// TestCalculateDamage_Resistances verifies resistance application.
func TestCalculateDamage_Resistances(t *testing.T) {
	resolver := NewDefaultCombatResolver(nil)

	tests := []struct {
		name       string
		resistance float64
		wantLess   float64
		wantMore   float64
	}{
		{"no resistance", 0.0, 101, 99},
		{"50% resistance", 0.5, 55, 45},
		{"90% resistance", 0.9, 15, 9}, // minimum damage kicks in
		{"immunity", 1.0, 15, 9},       // minimum damage
		{"weakness", -0.25, 130, 120},
		{"extreme weakness capped", -1.0, 155, 145}, // capped at -0.5
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stats := NewStats()
			stats.Resistances[DamageFire] = tt.resistance

			damage := Damage{Amount: 100, Type: DamageFire}
			result := resolver.CalculateDamage(damage, stats)

			if result >= tt.wantLess {
				t.Errorf("Expected result < %f, got %f", tt.wantLess, result)
			}
			if result <= tt.wantMore {
				t.Errorf("Expected result > %f, got %f", tt.wantMore, result)
			}
		})
	}
}

// TestCalculateDamage_MinimumDamage verifies minimum damage floor.
func TestCalculateDamage_MinimumDamage(t *testing.T) {
	resolver := NewDefaultCombatResolver(nil)
	resolver.MinDamageMultiplier = 0.1

	stats := NewStats()
	stats.Defense = 1000                     // Extreme defense
	stats.Resistances[DamagePhysical] = 0.99 // Near immunity

	damage := Damage{Amount: 100, Type: DamagePhysical}
	result := resolver.CalculateDamage(damage, stats)

	// Minimum should be 100 * 0.1 = 10
	if result < 10 {
		t.Errorf("Expected minimum damage >= 10, got %f", result)
	}
}

// TestCalculateDamage_CombinedMitigation verifies defense + resistance combo.
func TestCalculateDamage_CombinedMitigation(t *testing.T) {
	resolver := NewDefaultCombatResolver(nil)

	stats := NewStats()
	stats.Defense = 50                      // 100 * (100/150) ≈ 66.67
	stats.Resistances[DamagePhysical] = 0.3 // 66.67 * 0.7 ≈ 46.67

	damage := Damage{Amount: 100, Type: DamagePhysical}
	result := resolver.CalculateDamage(damage, stats)

	// Expected: ~46.67
	if result > 50 || result < 43 {
		t.Errorf("Expected ~46.67 damage, got %f", result)
	}
}

// TestResolveCombat_NilLookup verifies behavior without EntityLookup.
func TestResolveCombat_NilLookup(t *testing.T) {
	resolver := &DefaultCombatResolver{}

	result := resolver.ResolveCombat(1, 2)

	if result != nil {
		t.Errorf("Expected nil result with nil EntityLookup, got %v", result)
	}
}

// TestResolveCombat_MissingEntities verifies behavior with missing entities.
func TestResolveCombat_MissingEntities(t *testing.T) {
	provider := NewMockEntityProvider()
	resolver := NewDefaultCombatResolver(provider)

	tests := []struct {
		name           string
		attackerExists bool
		defenderExists bool
	}{
		{"attacker missing", false, true},
		{"defender missing", true, false},
		{"both missing", false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.attackerExists {
				provider.SetStats(1, NewStats())
			} else {
				delete(provider.stats, 1)
			}
			if tt.defenderExists {
				provider.SetStats(2, NewStats())
			} else {
				delete(provider.stats, 2)
			}

			result := resolver.ResolveCombat(1, 2)

			if result != nil {
				t.Errorf("Expected nil result with missing entity, got %v", result)
			}
		})
	}
}

// TestResolveCombat_BasicCombat verifies complete combat resolution.
func TestResolveCombat_BasicCombat(t *testing.T) {
	provider := NewMockEntityProvider()
	resolver := NewDefaultCombatResolver(provider)

	// Setup attacker
	attackerStats := NewStats()
	attackerStats.Attack = 50
	provider.SetStats(1, attackerStats)
	provider.SetAttackDamage(1, Damage{Amount: 50, Type: DamagePhysical})

	// Setup defender
	defenderStats := NewStats()
	defenderStats.Defense = 25
	provider.SetStats(2, defenderStats)

	result := resolver.ResolveCombat(1, 2)

	if len(result) != 1 {
		t.Fatalf("Expected 1 damage event, got %d", len(result))
	}

	damage := result[0]
	if damage.SourceID != 1 {
		t.Errorf("Expected SourceID 1, got %d", damage.SourceID)
	}
	if damage.TargetID != 2 {
		t.Errorf("Expected TargetID 2, got %d", damage.TargetID)
	}
	if damage.Type != DamagePhysical {
		t.Errorf("Expected Physical damage, got %v", damage.Type)
	}

	// 50 * (100/125) = 40
	if damage.Amount > 45 || damage.Amount < 35 {
		t.Errorf("Expected ~40 damage, got %f", damage.Amount)
	}
}

// TestResolveCombat_MagicalAttack verifies magical combat resolution.
func TestResolveCombat_MagicalAttack(t *testing.T) {
	provider := NewMockEntityProvider()
	resolver := NewDefaultCombatResolver(provider)

	// Setup attacker with fire damage
	attackerStats := NewStats()
	attackerStats.MagicPower = 100
	provider.SetStats(1, attackerStats)
	provider.SetAttackDamage(1, Damage{Amount: 80, Type: DamageFire})

	// Setup defender with fire resistance
	defenderStats := NewStats()
	defenderStats.MagicDefense = 50
	defenderStats.Resistances[DamageFire] = 0.25
	provider.SetStats(2, defenderStats)

	result := resolver.ResolveCombat(1, 2)

	if len(result) != 1 {
		t.Fatalf("Expected 1 damage event, got %d", len(result))
	}

	damage := result[0]
	if damage.Type != DamageFire {
		t.Errorf("Expected Fire damage, got %v", damage.Type)
	}

	// 80 * (100/150) * 0.75 ≈ 40
	if damage.Amount > 45 || damage.Amount < 35 {
		t.Errorf("Expected ~40 damage, got %f", damage.Amount)
	}
}

// TestResolveCombat_IDsCorrect verifies entity IDs are set correctly.
func TestResolveCombat_IDsCorrect(t *testing.T) {
	provider := NewMockEntityProvider()
	resolver := NewDefaultCombatResolver(provider)

	provider.SetStats(100, NewStats())
	provider.SetStats(200, NewStats())

	result := resolver.ResolveCombat(100, 200)

	if len(result) != 1 {
		t.Fatalf("Expected 1 damage event, got %d", len(result))
	}

	if result[0].SourceID != 100 {
		t.Errorf("Expected SourceID 100, got %d", result[0].SourceID)
	}
	if result[0].TargetID != 200 {
		t.Errorf("Expected TargetID 200, got %d", result[0].TargetID)
	}
}

// TestDefaultCombatResolver_Interface verifies interface implementation.
func TestDefaultCombatResolver_Interface(t *testing.T) {
	var _ CombatResolver = (*DefaultCombatResolver)(nil)
}

// TestCalculateDamage_PrecisionEdgeCases verifies floating point precision.
func TestCalculateDamage_PrecisionEdgeCases(t *testing.T) {
	resolver := NewDefaultCombatResolver(nil)

	tests := []struct {
		name   string
		amount float64
	}{
		{"very small", 0.001},
		{"very large", 1000000},
		{"fractional", 33.333},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stats := NewStats()
			damage := Damage{Amount: tt.amount, Type: DamagePhysical}

			result := resolver.CalculateDamage(damage, stats)

			if math.IsNaN(result) || math.IsInf(result, 0) {
				t.Errorf("Result should not be NaN or Inf, got %f", result)
			}
		})
	}
}

// Benchmark for CalculateDamage
func BenchmarkCalculateDamage(b *testing.B) {
	resolver := NewDefaultCombatResolver(nil)
	stats := NewStats()
	stats.Defense = 50
	stats.Resistances[DamageFire] = 0.25
	damage := Damage{Amount: 100, Type: DamageFire}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resolver.CalculateDamage(damage, stats)
	}
}

// Benchmark for ResolveCombat
func BenchmarkResolveCombat(b *testing.B) {
	provider := NewMockEntityProvider()
	resolver := NewDefaultCombatResolver(provider)

	provider.SetStats(1, NewStats())
	provider.SetStats(2, NewStats())

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resolver.ResolveCombat(1, 2)
	}
}
