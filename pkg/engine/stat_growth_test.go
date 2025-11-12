package engine

import (
	"math"
	"testing"
)

func TestStatGrowth_CalculateHP(t *testing.T) {
	tests := []struct {
		name     string
		growth   StatGrowth
		baseHP   float64
		level    int
		expected float64
	}{
		{
			name: "level 1 returns base HP",
			growth: StatGrowth{
				HPPerLevel:        10.0,
				HPPercentPerLevel: 0.05,
			},
			baseHP:   100.0,
			level:    1,
			expected: 100.0,
		},
		{
			name: "level 5 with flat bonus",
			growth: StatGrowth{
				HPPerLevel:        10.0,
				HPPercentPerLevel: 0.0,
			},
			baseHP:   100.0,
			level:    5,
			expected: 140.0, // 100 + (10 * 4)
		},
		{
			name: "level 5 with percentage bonus",
			growth: StatGrowth{
				HPPerLevel:        10.0,
				HPPercentPerLevel: 0.05,
			},
			baseHP:   100.0,
			level:    5,
			expected: 168.0, // (100 + 40) * (1 + 0.05 * 4) = 140 * 1.2
		},
		{
			name: "level 0 treated as level 1",
			growth: StatGrowth{
				HPPerLevel:        10.0,
				HPPercentPerLevel: 0.05,
			},
			baseHP:   100.0,
			level:    0,
			expected: 100.0,
		},
		{
			name: "high level scaling (level 20)",
			growth: StatGrowth{
				HPPerLevel:        15.0,
				HPPercentPerLevel: 0.05,
			},
			baseHP:   150.0,
			level:    20,
			expected: 848.25, // (150 + 15*19) * (1 + 0.05*19) = 435 * 1.95 = 848.25 (corrected)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.growth.CalculateHP(tt.baseHP, tt.level)
			if math.Abs(result-tt.expected) > 0.01 {
				t.Errorf("CalculateHP() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestStatGrowth_CalculateAttack(t *testing.T) {
	tests := []struct {
		name       string
		growth     StatGrowth
		baseAttack float64
		level      int
		expected   float64
	}{
		{
			name: "level 1 returns base attack",
			growth: StatGrowth{
				AttackPerLevel:        1.5,
				AttackPercentPerLevel: 0.03,
			},
			baseAttack: 15.0,
			level:      1,
			expected:   15.0,
		},
		{
			name: "level 10 with both bonuses",
			growth: StatGrowth{
				AttackPerLevel:        2.0,
				AttackPercentPerLevel: 0.05,
			},
			baseAttack: 10.0,
			level:      10,
			expected:   40.6, // (10 + 2*9) * (1 + 0.05*9) = 28 * 1.45 = 40.6 (corrected)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.growth.CalculateAttack(tt.baseAttack, tt.level)
			if math.Abs(result-tt.expected) > 0.01 {
				t.Errorf("CalculateAttack() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestStatGrowth_CalculateDefense(t *testing.T) {
	growth := StatGrowth{
		DefensePerLevel: 1.2,
	}

	tests := []struct {
		name        string
		baseDefense float64
		level       int
		expected    float64
	}{
		{"level 1", 12.0, 1, 12.0},
		{"level 5", 12.0, 5, 16.8},   // 12 + 1.2 * 4
		{"level 10", 10.0, 10, 20.8}, // 10 + 1.2 * 9
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := growth.CalculateDefense(tt.baseDefense, tt.level)
			if math.Abs(result-tt.expected) > 0.01 {
				t.Errorf("CalculateDefense() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestStatGrowth_CalculateMagicPower(t *testing.T) {
	growth := StatGrowth{
		MagicPowerPerLevel: 2.0,
	}

	tests := []struct {
		name           string
		baseMagicPower float64
		level          int
		expected       float64
	}{
		{"level 1", 10.0, 1, 10.0},
		{"level 5", 10.0, 5, 18.0},  // 10 + 2 * 4
		{"level 10", 5.0, 10, 23.0}, // 5 + 2 * 9
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := growth.CalculateMagicPower(tt.baseMagicPower, tt.level)
			if math.Abs(result-tt.expected) > 0.01 {
				t.Errorf("CalculateMagicPower() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestStatGrowth_CalculateSpeed(t *testing.T) {
	growth := StatGrowth{
		SpeedPerLevel: 1.5,
	}

	tests := []struct {
		name      string
		baseSpeed float64
		level     int
		expected  float64
	}{
		{"level 1", 100.0, 1, 100.0},
		{"level 5", 100.0, 5, 106.0},   // 100 + 1.5 * 4
		{"level 10", 100.0, 10, 113.5}, // 100 + 1.5 * 9
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := growth.CalculateSpeed(tt.baseSpeed, tt.level)
			if math.Abs(result-tt.expected) > 0.01 {
				t.Errorf("CalculateSpeed() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestGetClassStatGrowth(t *testing.T) {
	tests := []struct {
		name  string
		class CharacterClass
	}{
		{"warrior", ClassWarrior},
		{"rogue", ClassRogue},
		{"mage", ClassMage},
		{"ranger", ClassRanger},
		{"cleric", ClassCleric},
		{"necromancer", ClassNecromancer},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			growth := GetClassStatGrowth(tt.class)

			if growth == nil {
				t.Fatal("GetClassStatGrowth() returned nil")
			}

			// Verify all growth values are non-negative
			if growth.HPPerLevel < 0 {
				t.Errorf("HPPerLevel is negative: %v", growth.HPPerLevel)
			}
			if growth.ManaPerLevel < 0 {
				t.Errorf("ManaPerLevel is negative: %v", growth.ManaPerLevel)
			}
			if growth.AttackPerLevel < 0 {
				t.Errorf("AttackPerLevel is negative: %v", growth.AttackPerLevel)
			}

			// Verify percentage bonuses are reasonable (0-20%)
			if growth.HPPercentPerLevel < 0 || growth.HPPercentPerLevel > 0.2 {
				t.Errorf("HPPercentPerLevel out of range: %v", growth.HPPercentPerLevel)
			}
		})
	}
}

func TestStatGrowth_ClassSpecificScaling(t *testing.T) {
	// Test that each class scales appropriately to its role
	tests := []struct {
		name     string
		class    CharacterClass
		baseHP   float64
		baseMana float64
		level    int
		minHP    float64 // Expected minimum HP at level
		minMana  float64 // Expected minimum Mana at level
	}{
		{
			name:     "warrior has high HP at level 10",
			class:    ClassWarrior,
			baseHP:   150.0,
			baseMana: 50.0,
			level:    10,
			minHP:    250.0, // Warriors should have strong HP growth
			minMana:  50.0,
		},
		{
			name:     "mage has high mana at level 10",
			class:    ClassMage,
			baseHP:   80.0,
			baseMana: 150.0,
			level:    10,
			minHP:    80.0,
			minMana:  250.0, // Mages should have strong mana growth
		},
		{
			name:     "rogue has balanced growth at level 10",
			class:    ClassRogue,
			baseHP:   100.0,
			baseMana: 80.0,
			level:    10,
			minHP:    150.0,
			minMana:  100.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			growth := GetClassStatGrowth(tt.class)

			hp := growth.CalculateHP(tt.baseHP, tt.level)
			mana := growth.CalculateMana(tt.baseMana, tt.level)

			if hp < tt.minHP {
				t.Errorf("HP too low: got %v, want at least %v", hp, tt.minHP)
			}
			if mana < tt.minMana {
				t.Errorf("Mana too low: got %v, want at least %v", mana, tt.minMana)
			}
		})
	}
}

func TestStatGrowth_WarriorVsMage(t *testing.T) {
	// Verify that warrior and mage have distinct scaling patterns
	warriorGrowth := GetClassStatGrowth(ClassWarrior)
	mageGrowth := GetClassStatGrowth(ClassMage)

	level := 10
	baseHP := 100.0
	baseMana := 100.0

	warriorHP := warriorGrowth.CalculateHP(baseHP, level)
	mageHP := mageGrowth.CalculateHP(baseHP, level)

	warriorMana := warriorGrowth.CalculateMana(baseMana, level)
	mageMana := mageGrowth.CalculateMana(baseMana, level)

	// Warriors should have more HP than mages
	if warriorHP <= mageHP {
		t.Errorf("Warrior HP (%v) should be > Mage HP (%v)", warriorHP, mageHP)
	}

	// Mages should have more mana than warriors
	if mageMana <= warriorMana {
		t.Errorf("Mage Mana (%v) should be > Warrior Mana (%v)", mageMana, warriorMana)
	}
}

// Benchmark stat growth calculations
func BenchmarkStatGrowth_CalculateHP(b *testing.B) {
	growth := GetClassStatGrowth(ClassWarrior)
	for i := 0; i < b.N; i++ {
		growth.CalculateHP(100.0, 10)
	}
}

func BenchmarkGetClassStatGrowth(b *testing.B) {
	for i := 0; i < b.N; i++ {
		GetClassStatGrowth(ClassWarrior)
	}
}
