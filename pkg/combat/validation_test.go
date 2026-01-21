// validation_test.go contains tests for the Validate methods on Damage and Stats.
package combat

import (
	"errors"
	"testing"
)

// TestDamage_Validate verifies Damage validation logic.
func TestDamage_Validate(t *testing.T) {
	tests := []struct {
		name    string
		damage  Damage
		wantErr error
	}{
		{
			name:    "valid_physical_damage",
			damage:  Damage{Amount: 50.0, Type: DamagePhysical, SourceID: 1, TargetID: 2},
			wantErr: nil,
		},
		{
			name:    "valid_zero_damage",
			damage:  Damage{Amount: 0.0, Type: DamageMagical, SourceID: 1, TargetID: 2},
			wantErr: nil,
		},
		{
			name:    "valid_fire_damage",
			damage:  Damage{Amount: 100.0, Type: DamageFire, SourceID: 1, TargetID: 2},
			wantErr: nil,
		},
		{
			name:    "valid_ice_damage",
			damage:  Damage{Amount: 25.5, Type: DamageIce, SourceID: 1, TargetID: 2},
			wantErr: nil,
		},
		{
			name:    "valid_lightning_damage",
			damage:  Damage{Amount: 75.0, Type: DamageLightning, SourceID: 1, TargetID: 2},
			wantErr: nil,
		},
		{
			name:    "valid_poison_damage",
			damage:  Damage{Amount: 10.0, Type: DamagePoison, SourceID: 1, TargetID: 2},
			wantErr: nil,
		},
		{
			name:    "negative_damage",
			damage:  Damage{Amount: -10.0, Type: DamagePhysical, SourceID: 1, TargetID: 2},
			wantErr: ErrNegativeDamage,
		},
		{
			name:    "invalid_damage_type_negative",
			damage:  Damage{Amount: 50.0, Type: DamageType(-1), SourceID: 1, TargetID: 2},
			wantErr: ErrInvalidDamageType,
		},
		{
			name:    "invalid_damage_type_too_high",
			damage:  Damage{Amount: 50.0, Type: DamageType(100), SourceID: 1, TargetID: 2},
			wantErr: ErrInvalidDamageType,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.damage.Validate()

			if tt.wantErr == nil {
				if err != nil {
					t.Errorf("Validate() error = %v, wantErr nil", err)
				}
			} else {
				if err == nil {
					t.Errorf("Validate() error = nil, wantErr %v", tt.wantErr)
				} else if !errors.Is(err, tt.wantErr) {
					t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
				}
			}
		})
	}
}

// TestStats_Validate verifies Stats validation logic.
func TestStats_Validate(t *testing.T) {
	tests := []struct {
		name    string
		setup   func() *Stats
		wantErr error
	}{
		{
			name: "valid_default_stats",
			setup: func() *Stats {
				return NewStats()
			},
			wantErr: nil,
		},
		{
			name: "valid_custom_stats",
			setup: func() *Stats {
				s := NewStats()
				s.HP = 150
				s.MaxHP = 200
				s.Mana = 80
				s.MaxMana = 100
				s.Attack = 50
				s.Defense = 30
				s.MagicPower = 40
				s.MagicDefense = 25
				s.Speed = 120
				s.CritChance = 0.25
				s.CritDamage = 2.0
				s.Evasion = 0.15
				s.Resistances[DamageFire] = 0.5
				return s
			},
			wantErr: nil,
		},
		{
			name: "valid_zero_mana_entity",
			setup: func() *Stats {
				s := NewStats()
				s.Mana = 0
				s.MaxMana = 0
				return s
			},
			wantErr: nil,
		},
		{
			name: "valid_resistance_weakness",
			setup: func() *Stats {
				s := NewStats()
				s.Resistances[DamageFire] = -0.5 // Fire weakness
				return s
			},
			wantErr: nil,
		},
		{
			name: "valid_resistance_immunity",
			setup: func() *Stats {
				s := NewStats()
				s.Resistances[DamagePoison] = 1.0 // Poison immunity
				return s
			},
			wantErr: nil,
		},
		{
			name: "valid_crit_damage_zero",
			setup: func() *Stats {
				s := NewStats()
				s.CritDamage = 0 // Not set
				return s
			},
			wantErr: nil,
		},
		{
			name: "negative_hp",
			setup: func() *Stats {
				s := NewStats()
				s.HP = -10
				return s
			},
			wantErr: ErrNegativeHP,
		},
		{
			name: "zero_maxhp",
			setup: func() *Stats {
				s := NewStats()
				s.MaxHP = 0
				return s
			},
			wantErr: ErrNegativeMaxHP,
		},
		{
			name: "negative_maxhp",
			setup: func() *Stats {
				s := NewStats()
				s.MaxHP = -50
				return s
			},
			wantErr: ErrNegativeMaxHP,
		},
		{
			name: "hp_exceeds_maxhp",
			setup: func() *Stats {
				s := NewStats()
				s.HP = 150
				s.MaxHP = 100
				return s
			},
			wantErr: ErrHPExceedsMax,
		},
		{
			name: "negative_mana",
			setup: func() *Stats {
				s := NewStats()
				s.Mana = -5
				return s
			},
			wantErr: ErrNegativeMana,
		},
		{
			name: "negative_maxmana",
			setup: func() *Stats {
				s := NewStats()
				s.MaxMana = -10
				return s
			},
			wantErr: ErrNegativeMaxMana,
		},
		{
			name: "mana_exceeds_maxmana",
			setup: func() *Stats {
				s := NewStats()
				s.Mana = 100
				s.MaxMana = 50
				return s
			},
			wantErr: ErrManaExceedsMax,
		},
		{
			name: "negative_attack",
			setup: func() *Stats {
				s := NewStats()
				s.Attack = -5
				return s
			},
			wantErr: ErrNegativeAttack,
		},
		{
			name: "negative_magic_power",
			setup: func() *Stats {
				s := NewStats()
				s.MagicPower = -10
				return s
			},
			wantErr: errors.New("MagicPower cannot be negative"),
		},
		{
			name: "negative_defense",
			setup: func() *Stats {
				s := NewStats()
				s.Defense = -5
				return s
			},
			wantErr: ErrNegativeDefense,
		},
		{
			name: "negative_magic_defense",
			setup: func() *Stats {
				s := NewStats()
				s.MagicDefense = -10
				return s
			},
			wantErr: errors.New("MagicDefense cannot be negative"),
		},
		{
			name: "negative_speed",
			setup: func() *Stats {
				s := NewStats()
				s.Speed = -20
				return s
			},
			wantErr: ErrNegativeSpeed,
		},
		{
			name: "crit_chance_negative",
			setup: func() *Stats {
				s := NewStats()
				s.CritChance = -0.1
				return s
			},
			wantErr: ErrInvalidCritChance,
		},
		{
			name: "crit_chance_over_one",
			setup: func() *Stats {
				s := NewStats()
				s.CritChance = 1.5
				return s
			},
			wantErr: ErrInvalidCritChance,
		},
		{
			name: "crit_damage_below_one",
			setup: func() *Stats {
				s := NewStats()
				s.CritDamage = 0.5 // Below 1.0 but not zero
				return s
			},
			wantErr: ErrInvalidCritDamage,
		},
		{
			name: "evasion_negative",
			setup: func() *Stats {
				s := NewStats()
				s.Evasion = -0.1
				return s
			},
			wantErr: ErrInvalidEvasion,
		},
		{
			name: "evasion_over_one",
			setup: func() *Stats {
				s := NewStats()
				s.Evasion = 1.2
				return s
			},
			wantErr: ErrInvalidEvasion,
		},
		{
			name: "resistance_too_low",
			setup: func() *Stats {
				s := NewStats()
				s.Resistances[DamageFire] = -1.5
				return s
			},
			wantErr: ErrInvalidResistance,
		},
		{
			name: "resistance_too_high",
			setup: func() *Stats {
				s := NewStats()
				s.Resistances[DamageIce] = 1.5
				return s
			},
			wantErr: ErrInvalidResistance,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stats := tt.setup()
			err := stats.Validate()

			if tt.wantErr == nil {
				if err != nil {
					t.Errorf("Validate() error = %v, wantErr nil", err)
				}
			} else {
				if err == nil {
					t.Errorf("Validate() error = nil, wantErr containing %q", tt.wantErr.Error())
				}
			}
		})
	}
}

// TestStats_IsDead verifies the IsDead helper method.
func TestStats_IsDead(t *testing.T) {
	tests := []struct {
		name   string
		hp     float64
		isDead bool
	}{
		{"zero_hp", 0, true},
		{"negative_hp", -10, true},
		{"one_hp", 1, false},
		{"full_hp", 100, false},
		{"partial_hp", 50, false},
		{"barely_alive", 0.001, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stats := NewStats()
			stats.HP = tt.hp

			if got := stats.IsDead(); got != tt.isDead {
				t.Errorf("IsDead() = %v, want %v", got, tt.isDead)
			}
		})
	}
}

// TestStats_GetResistance verifies the GetResistance helper method.
func TestStats_GetResistance(t *testing.T) {
	tests := []struct {
		name       string
		setup      func() *Stats
		damageType DamageType
		want       float64
	}{
		{
			name: "no_resistance_set",
			setup: func() *Stats {
				return NewStats()
			},
			damageType: DamageFire,
			want:       0.0,
		},
		{
			name: "resistance_set",
			setup: func() *Stats {
				s := NewStats()
				s.Resistances[DamageFire] = 0.5
				return s
			},
			damageType: DamageFire,
			want:       0.5,
		},
		{
			name: "different_resistance",
			setup: func() *Stats {
				s := NewStats()
				s.Resistances[DamageFire] = 0.5
				return s
			},
			damageType: DamageIce,
			want:       0.0,
		},
		{
			name: "nil_resistances_map",
			setup: func() *Stats {
				return &Stats{HP: 100, MaxHP: 100} // No Resistances map
			},
			damageType: DamageFire,
			want:       0.0,
		},
		{
			name: "weakness_resistance",
			setup: func() *Stats {
				s := NewStats()
				s.Resistances[DamageFire] = -0.5
				return s
			},
			damageType: DamageFire,
			want:       -0.5,
		},
		{
			name: "immunity",
			setup: func() *Stats {
				s := NewStats()
				s.Resistances[DamagePoison] = 1.0
				return s
			},
			damageType: DamagePoison,
			want:       1.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stats := tt.setup()
			if got := stats.GetResistance(tt.damageType); got != tt.want {
				t.Errorf("GetResistance() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestStats_ApplyDamage verifies the ApplyDamage helper method.
func TestStats_ApplyDamage(t *testing.T) {
	tests := []struct {
		name       string
		startHP    float64
		damage     float64
		wantHP     float64
		wantActual float64
	}{
		{
			name:       "normal_damage",
			startHP:    100,
			damage:     30,
			wantHP:     70,
			wantActual: 30,
		},
		{
			name:       "lethal_damage",
			startHP:    50,
			damage:     60,
			wantHP:     0,
			wantActual: 50,
		},
		{
			name:       "overkill_damage",
			startHP:    25,
			damage:     100,
			wantHP:     0,
			wantActual: 25,
		},
		{
			name:       "zero_damage",
			startHP:    100,
			damage:     0,
			wantHP:     100,
			wantActual: 0,
		},
		{
			name:       "negative_damage_ignored",
			startHP:    100,
			damage:     -10,
			wantHP:     100,
			wantActual: 0,
		},
		{
			name:       "exact_lethal",
			startHP:    50,
			damage:     50,
			wantHP:     0,
			wantActual: 50,
		},
		{
			name:       "damage_to_already_dead",
			startHP:    0,
			damage:     50,
			wantHP:     0,
			wantActual: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stats := NewStats()
			stats.HP = tt.startHP

			actual := stats.ApplyDamage(tt.damage)

			if stats.HP != tt.wantHP {
				t.Errorf("ApplyDamage() HP = %v, want %v", stats.HP, tt.wantHP)
			}
			if actual != tt.wantActual {
				t.Errorf("ApplyDamage() returned = %v, want %v", actual, tt.wantActual)
			}
		})
	}
}

// TestStats_ApplyHealing verifies the ApplyHealing helper method.
func TestStats_ApplyHealing(t *testing.T) {
	tests := []struct {
		name       string
		startHP    float64
		maxHP      float64
		healing    float64
		wantHP     float64
		wantActual float64
	}{
		{
			name:       "normal_healing",
			startHP:    50,
			maxHP:      100,
			healing:    30,
			wantHP:     80,
			wantActual: 30,
		},
		{
			name:       "overheal_capped",
			startHP:    80,
			maxHP:      100,
			healing:    50,
			wantHP:     100,
			wantActual: 20,
		},
		{
			name:       "healing_at_max",
			startHP:    100,
			maxHP:      100,
			healing:    50,
			wantHP:     100,
			wantActual: 0,
		},
		{
			name:       "zero_healing",
			startHP:    50,
			maxHP:      100,
			healing:    0,
			wantHP:     50,
			wantActual: 0,
		},
		{
			name:       "negative_healing_ignored",
			startHP:    50,
			maxHP:      100,
			healing:    -20,
			wantHP:     50,
			wantActual: 0,
		},
		{
			name:       "exact_to_max",
			startHP:    70,
			maxHP:      100,
			healing:    30,
			wantHP:     100,
			wantActual: 30,
		},
		{
			name:       "healing_from_near_death",
			startHP:    1,
			maxHP:      100,
			healing:    50,
			wantHP:     51,
			wantActual: 50,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stats := NewStats()
			stats.HP = tt.startHP
			stats.MaxHP = tt.maxHP

			actual := stats.ApplyHealing(tt.healing)

			if stats.HP != tt.wantHP {
				t.Errorf("ApplyHealing() HP = %v, want %v", stats.HP, tt.wantHP)
			}
			if actual != tt.wantActual {
				t.Errorf("ApplyHealing() returned = %v, want %v", actual, tt.wantActual)
			}
		})
	}
}

// TestValidationErrors verifies that error variables are distinct and usable.
func TestValidationErrors(t *testing.T) {
	// Verify all error variables are non-nil
	errorVars := []error{
		ErrNegativeDamage,
		ErrInvalidDamageType,
		ErrNegativeHP,
		ErrNegativeMaxHP,
		ErrHPExceedsMax,
		ErrNegativeMana,
		ErrNegativeMaxMana,
		ErrManaExceedsMax,
		ErrNegativeAttack,
		ErrNegativeDefense,
		ErrNegativeSpeed,
		ErrInvalidCritChance,
		ErrInvalidCritDamage,
		ErrInvalidEvasion,
		ErrInvalidResistance,
	}

	for i, err := range errorVars {
		if err == nil {
			t.Errorf("Error variable at index %d is nil", i)
		}
	}

	// Verify error messages are non-empty
	for _, err := range errorVars {
		if err.Error() == "" {
			t.Errorf("Error %v has empty message", err)
		}
	}
}
