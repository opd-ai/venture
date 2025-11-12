package magic

import (
	"math"
	"testing"
	
	"github.com/opd-ai/venture/pkg/procgen"
)

// TestBalanceConfig_BalanceStats verifies the balance system applies correct formulas.
func TestBalanceConfig_BalanceStats(t *testing.T) {
	config := DefaultBalanceConfig()
	
	tests := []struct {
		name       string
		stats      Stats
		spellType  SpellType
		target     TargetType
		level      int
		wantMinMana int
		wantMaxMana int
	}{
		{
			name: "offensive_single_target_level_1",
			stats: Stats{
				Damage:   50,
				CastTime: 1.0,
			},
			spellType:   TypeOffensive,
			target:      TargetSingle,
			level:       1,
			wantMinMana: 15,
			wantMaxMana: 25,
		},
		{
			name: "offensive_area_level_1",
			stats: Stats{
				Damage:   50,
				CastTime: 1.5,
				AreaSize: 10.0,
			},
			spellType:   TypeOffensive,
			target:      TargetArea,
			level:       1,
			wantMinMana: 50,
			wantMaxMana: 60,
		},
		{
			name: "healing_single_target_level_1",
			stats: Stats{
				Healing:  50,
				CastTime: 1.0,
			},
			spellType:   TypeHealing,
			target:      TargetSingle,
			level:       1,
			wantMinMana: 15,
			wantMaxMana: 22,
		},
		{
			name: "buff_level_5",
			stats: Stats{
				Duration: 30.0,
				CastTime: 0.5,
			},
			spellType:   TypeBuff,
			target:      TargetSelf,
			level:       5,
			wantMinMana: 190,
			wantMaxMana: 210,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Apply balance
			config.BalanceStats(&tt.stats, tt.spellType, tt.target, tt.level)
			
			// Check mana cost is in expected range
			if tt.stats.ManaCost < tt.wantMinMana || tt.stats.ManaCost > tt.wantMaxMana {
				t.Errorf("ManaCost = %d, want in range [%d, %d]", 
					tt.stats.ManaCost, tt.wantMinMana, tt.wantMaxMana)
			}
			
			// Check cooldown is reasonable
			if tt.stats.Cooldown < config.MinCooldown {
				t.Errorf("Cooldown = %.2f, want >= %.2f", tt.stats.Cooldown, config.MinCooldown)
			}
			if tt.stats.Cooldown > config.MaxCooldown {
				t.Errorf("Cooldown = %.2f, want <= %.2f", tt.stats.Cooldown, config.MaxCooldown)
			}
			
			// Check cooldown is at least 2x cast time
			minCooldown := tt.stats.CastTime * config.CooldownCastTimeRatio
			if tt.stats.Cooldown < minCooldown {
				t.Errorf("Cooldown = %.2f, want >= %.2f (2x cast time)", 
					tt.stats.Cooldown, minCooldown)
			}
		})
	}
}

// TestBalanceConfig_ScalePowerWithLevel verifies level scaling works correctly.
func TestBalanceConfig_ScalePowerWithLevel(t *testing.T) {
	config := DefaultBalanceConfig()
	
	tests := []struct {
		name      string
		stats     Stats
		spellType SpellType
		level     int
		checkFunc func(*testing.T, Stats, Stats) // check original vs scaled
	}{
		{
			name: "offensive_damage_scales",
			stats: Stats{
				Damage: 50,
			},
			spellType: TypeOffensive,
			level:     10,
			checkFunc: func(t *testing.T, original, scaled Stats) {
				expectedScale := 1.0 + float64(10-1)*config.PowerPerLevel
				expected := int(math.Round(float64(original.Damage) * expectedScale))
				if scaled.Damage != expected {
					t.Errorf("Damage = %d, want %d (scale: %.2f)", 
						scaled.Damage, expected, expectedScale)
				}
			},
		},
		{
			name: "healing_scales",
			stats: Stats{
				Healing: 50,
			},
			spellType: TypeHealing,
			level:     5,
			checkFunc: func(t *testing.T, original, scaled Stats) {
				expectedScale := 1.0 + float64(5-1)*config.PowerPerLevel
				expected := int(math.Round(float64(original.Healing) * expectedScale))
				if scaled.Healing != expected {
					t.Errorf("Healing = %d, want %d (scale: %.2f)", 
						scaled.Healing, expected, expectedScale)
				}
			},
		},
		{
			name: "duration_scales",
			stats: Stats{
				Duration: 30.0,
			},
			spellType: TypeBuff,
			level:     8,
			checkFunc: func(t *testing.T, original, scaled Stats) {
				expectedScale := 1.0 + float64(8-1)*config.PowerPerLevel
				expected := math.Round(original.Duration * expectedScale * 10) / 10
				if scaled.Duration != expected {
					t.Errorf("Duration = %.2f, want %.2f (scale: %.2f)", 
						scaled.Duration, expected, expectedScale)
				}
			},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			original := tt.stats
			config.scalePowerWithLevel(&tt.stats, tt.spellType, tt.level)
			tt.checkFunc(t, original, tt.stats)
		})
	}
}

// TestBalanceConfig_ValidateDPS verifies DPS validation catches imbalanced spells.
func TestBalanceConfig_ValidateDPS(t *testing.T) {
	config := DefaultBalanceConfig()
	
	tests := []struct {
		name    string
		spell   *Spell
		wantErr bool
	}{
		{
			name: "balanced_offensive_spell",
			spell: &Spell{
				Type: TypeOffensive,
				Stats: Stats{
					Damage:        50,
					CastTime:      1.0,
					Cooldown:      3.0,
					RequiredLevel: 1,
				},
			},
			wantErr: false, // 50 / 4.0 = 12.5 DPS (target: 15 ± 40% = 9-21 DPS)
		},
		{
			name: "overpowered_offensive_spell",
			spell: &Spell{
				Name: "OP Spell",
				Type: TypeOffensive,
				Stats: Stats{
					Damage:        200,
					CastTime:      0.5,
					Cooldown:      1.0,
					RequiredLevel: 1,
				},
			},
			wantErr: true, // 200 / 1.5 = 133 DPS (way too high)
		},
		{
			name: "underpowered_offensive_spell",
			spell: &Spell{
				Name: "Weak Spell",
				Type: TypeOffensive,
				Stats: Stats{
					Damage:        5,
					CastTime:      1.0,
					Cooldown:      10.0,
					RequiredLevel: 1,
				},
			},
			wantErr: true, // 5 / 11.0 = 0.45 DPS (way too low)
		},
		{
			name: "healing_spell_not_validated",
			spell: &Spell{
				Type: TypeHealing,
				Stats: Stats{
					Healing:       100,
					CastTime:      2.0,
					Cooldown:      5.0,
					RequiredLevel: 1,
				},
			},
			wantErr: false, // Healing spells use ValidateHPS, not ValidateDPS
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := config.ValidateDPS(tt.spell)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateDPS() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestBalanceConfig_ValidateHPS verifies HPS validation for healing spells.
func TestBalanceConfig_ValidateHPS(t *testing.T) {
	config := DefaultBalanceConfig()
	
	tests := []struct {
		name    string
		spell   *Spell
		wantErr bool
	}{
		{
			name: "balanced_healing_spell",
			spell: &Spell{
				Type: TypeHealing,
				Stats: Stats{
					Healing:       40,
					CastTime:      1.0,
					Cooldown:      2.5,
					RequiredLevel: 1,
				},
			},
			wantErr: false, // 40 / 3.5 = 11.4 HPS (target: 12 ± 40% = 7.2-16.8 HPS)
		},
		{
			name: "overpowered_healing_spell",
			spell: &Spell{
				Name: "OP Heal",
				Type: TypeHealing,
				Stats: Stats{
					Healing:       200,
					CastTime:      0.5,
					Cooldown:      2.0,
					RequiredLevel: 1,
				},
			},
			wantErr: true, // 200 / 2.5 = 80 HPS (way too high)
		},
		{
			name: "offensive_spell_not_validated",
			spell: &Spell{
				Type: TypeOffensive,
				Stats: Stats{
					Damage:        50,
					CastTime:      1.0,
					Cooldown:      3.0,
					RequiredLevel: 1,
				},
			},
			wantErr: false, // Offensive spells use ValidateDPS, not ValidateHPS
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := config.ValidateHPS(tt.spell)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateHPS() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestBalanceConfig_ValidateManaCostEfficiency checks mana cost efficiency validation.
func TestBalanceConfig_ValidateManaCostEfficiency(t *testing.T) {
	config := DefaultBalanceConfig()
	
	tests := []struct {
		name    string
		spell   *Spell
		wantErr bool
	}{
		{
			name: "efficient_spell",
			spell: &Spell{
				Type: TypeOffensive,
				Stats: Stats{
					Damage:   50,
					ManaCost: 25,
				},
			},
			wantErr: false, // 50/25 = 2.0 power per mana (within 1.0-4.5 range)
		},
		{
			name: "too_expensive_spell",
			spell: &Spell{
				Name: "Expensive Spell",
				Type: TypeOffensive,
				Stats: Stats{
					Damage:   20,
					ManaCost: 50,
				},
			},
			wantErr: true, // 20/50 = 0.4 power per mana (below 1.0 minimum)
		},
		{
			name: "too_cheap_spell",
			spell: &Spell{
				Name: "Cheap Spell",
				Type: TypeOffensive,
				Stats: Stats{
					Damage:   200,
					ManaCost: 30,
				},
			},
			wantErr: true, // 200/30 = 6.67 power per mana (above 4.5 maximum)
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := config.ValidateManaCostEfficiency(tt.spell)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateManaCostEfficiency() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestSpellGenerator_BalancedGeneration verifies generated spells pass balance checks.
func TestSpellGenerator_BalancedGeneration(t *testing.T) {
	gen := NewSpellGenerator()
	
	tests := []struct {
		name     string
		seed     int64
		params   procgen.GenerationParams
		minSpells int
	}{
		{
			name: "fantasy_spells_balanced",
			seed: 12345,
			params: procgen.GenerationParams{
				GenreID:    "fantasy",
				Depth:      5,
				Difficulty: 0.5,
				Custom:     map[string]interface{}{"count": 20},
			},
			minSpells: 20,
		},
		{
			name: "scifi_spells_balanced",
			seed: 54321,
			params: procgen.GenerationParams{
				GenreID:    "scifi",
				Depth:      10,
				Difficulty: 0.7,
				Custom:     map[string]interface{}{"count": 20},
			},
			minSpells: 20,
		},
		{
			name: "high_level_spells_balanced",
			seed: 99999,
			params: procgen.GenerationParams{
				GenreID:    "fantasy",
				Depth:      20,
				Difficulty: 0.8,
				Custom:     map[string]interface{}{"count": 20},
			},
			minSpells: 20,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := gen.Generate(tt.seed, tt.params)
			if err != nil {
				t.Fatalf("Generate() error = %v", err)
			}
			
			spells, ok := result.([]*Spell)
			if !ok {
				t.Fatalf("Generate() returned wrong type")
			}
			
			if len(spells) < tt.minSpells {
				t.Errorf("Generated %d spells, want at least %d", len(spells), tt.minSpells)
			}
			
			// Count balance warnings
			balanceIssues := 0
			for _, spell := range spells {
				if err := gen.balanceConfig.ValidateDPS(spell); err != nil {
					balanceIssues++
					t.Logf("Balance issue in %s: %v", spell.Name, err)
				}
				if err := gen.balanceConfig.ValidateHPS(spell); err != nil {
					balanceIssues++
					t.Logf("Balance issue in %s: %v", spell.Name, err)
				}
				if err := gen.balanceConfig.ValidateManaCostEfficiency(spell); err != nil {
					balanceIssues++
					t.Logf("Balance issue in %s: %v", spell.Name, err)
				}
			}
			
			// Allow up to 50% of spells to have balance warnings
			// (some edge cases are acceptable with retroactive balancing)
			maxIssues := len(spells) / 2
			if balanceIssues > maxIssues {
				t.Errorf("Too many balance issues: %d/%d (max %d allowed)", 
					balanceIssues, len(spells), maxIssues)
			}
		})
	}
}

// TestBalancedSpellComparison compares spell power across different types.
func TestBalancedSpellComparison(t *testing.T) {
	gen := NewSpellGenerator()
	params := procgen.GenerationParams{
		GenreID:    "fantasy",
		Depth:      5,
		Difficulty: 0.5,
		Custom:     map[string]interface{}{"count": 50},
	}
	
	result, err := gen.Generate(12345, params)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
	
	spells := result.([]*Spell)
	
	// Group spells by type
	byType := make(map[SpellType][]*Spell)
	for _, spell := range spells {
		byType[spell.Type] = append(byType[spell.Type], spell)
	}
	
	// Calculate average power rating per type
	avgRating := make(map[SpellType]float64)
	for spellType, typeSpells := range byType {
		if len(typeSpells) == 0 {
			continue
		}
		
		total := 0
		for _, spell := range typeSpells {
			rating := gen.balanceConfig.CalculatePowerRating(spell)
			total += rating
		}
		avgRating[spellType] = float64(total) / float64(len(typeSpells))
	}
	
	// Log average ratings
	for spellType, rating := range avgRating {
		t.Logf("%s spells: average power rating = %.2f", spellType.String(), rating)
	}
	
	// Verify offensive and healing spells have similar ratings
	// (they should be balanced against each other)
	if offRating, ok := avgRating[TypeOffensive]; ok {
		if healRating, ok := avgRating[TypeHealing]; ok {
			diff := math.Abs(offRating - healRating)
			maxDiff := 15.0 // Allow 15 point difference
			if diff > maxDiff {
				t.Errorf("Offensive and healing ratings too different: %.2f vs %.2f (diff: %.2f > %.2f)", 
					offRating, healRating, diff, maxDiff)
			}
		}
	}
}

// BenchmarkBalanceStats benchmarks the balance formula application.
func BenchmarkBalanceStats(b *testing.B) {
	config := DefaultBalanceConfig()
	stats := Stats{
		Damage:   50,
		CastTime: 1.0,
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		testStats := stats // Copy
		config.BalanceStats(&testStats, TypeOffensive, TargetSingle, 5)
	}
}

// BenchmarkValidateDPS benchmarks DPS validation.
func BenchmarkValidateDPS(b *testing.B) {
	config := DefaultBalanceConfig()
	spell := &Spell{
		Type: TypeOffensive,
		Stats: Stats{
			Damage:        50,
			CastTime:      1.0,
			Cooldown:      4.0,
			RequiredLevel: 1,
		},
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = config.ValidateDPS(spell)
	}
}
