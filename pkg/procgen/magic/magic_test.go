package magic

import (
	"testing"

	"github.com/opd-ai/venture/pkg/procgen"
)

func TestSpellGenerator_Generate(t *testing.T) {
	gen := NewSpellGenerator()

	tests := []struct {
		name    string
		seed    int64
		params  procgen.GenerationParams
		count   int
		wantErr bool
	}{
		{
			name: "fantasy spells default",
			seed: 12345,
			params: procgen.GenerationParams{
				Difficulty: 0.5,
				Depth:      5,
				GenreID:    "fantasy",
				Custom: map[string]interface{}{
					"count": 10,
				},
			},
			count:   10,
			wantErr: false,
		},
		{
			name: "scifi spells",
			seed: 54321,
			params: procgen.GenerationParams{
				Difficulty: 0.7,
				Depth:      10,
				GenreID:    "scifi",
				Custom: map[string]interface{}{
					"count": 15,
				},
			},
			count:   15,
			wantErr: false,
		},
		{
			name: "high depth progression",
			seed: 99999,
			params: procgen.GenerationParams{
				Difficulty: 0.9,
				Depth:      25,
				GenreID:    "fantasy",
				Custom: map[string]interface{}{
					"count": 20,
				},
			},
			count:   20,
			wantErr: false,
		},
		{
			name: "negative depth",
			seed: 11111,
			params: procgen.GenerationParams{
				Difficulty: 0.5,
				Depth:      -1,
				GenreID:    "fantasy",
				Custom: map[string]interface{}{
					"count": 5,
				},
			},
			wantErr: true,
		},
		{
			name: "invalid difficulty",
			seed: 22222,
			params: procgen.GenerationParams{
				Difficulty: 1.5,
				Depth:      5,
				GenreID:    "fantasy",
				Custom: map[string]interface{}{
					"count": 5,
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := gen.Generate(tt.seed, tt.params)

			if (err != nil) != tt.wantErr {
				t.Errorf("Generate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return
			}

			spells, ok := result.([]*Spell)
			if !ok {
				t.Errorf("Generate() result is not []*Spell")
				return
			}

			if len(spells) != tt.count {
				t.Errorf("Generate() produced %d spells, want %d", len(spells), tt.count)
			}

			// Validate all spells
			if err := gen.Validate(spells); err != nil {
				t.Errorf("Validate() failed: %v", err)
			}
		})
	}
}

func TestSpellGenerator_Determinism(t *testing.T) {
	gen := NewSpellGenerator()
	seed := int64(42)
	params := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      5,
		GenreID:    "fantasy",
		Custom: map[string]interface{}{
			"count": 10,
		},
	}

	// Generate twice with same seed
	result1, err1 := gen.Generate(seed, params)
	if err1 != nil {
		t.Fatalf("First generation failed: %v", err1)
	}

	result2, err2 := gen.Generate(seed, params)
	if err2 != nil {
		t.Fatalf("Second generation failed: %v", err2)
	}

	spells1 := result1.([]*Spell)
	spells2 := result2.([]*Spell)

	if len(spells1) != len(spells2) {
		t.Fatalf("Different number of spells generated: %d vs %d", len(spells1), len(spells2))
	}

	// Check that spells are identical
	for i := range spells1 {
		s1, s2 := spells1[i], spells2[i]

		if s1.Name != s2.Name {
			t.Errorf("Spell %d name mismatch: %s vs %s", i, s1.Name, s2.Name)
		}
		if s1.Type != s2.Type {
			t.Errorf("Spell %d type mismatch: %v vs %v", i, s1.Type, s2.Type)
		}
		if s1.Element != s2.Element {
			t.Errorf("Spell %d element mismatch: %v vs %v", i, s1.Element, s2.Element)
		}
		if s1.Rarity != s2.Rarity {
			t.Errorf("Spell %d rarity mismatch: %v vs %v", i, s1.Rarity, s2.Rarity)
		}
		if s1.Stats.Damage != s2.Stats.Damage {
			t.Errorf("Spell %d damage mismatch: %d vs %d", i, s1.Stats.Damage, s2.Stats.Damage)
		}
		if s1.Stats.ManaCost != s2.Stats.ManaCost {
			t.Errorf("Spell %d mana cost mismatch: %d vs %d", i, s1.Stats.ManaCost, s2.Stats.ManaCost)
		}
	}
}

func TestSpellGenerator_DepthScaling(t *testing.T) {
	gen := NewSpellGenerator()
	seed := int64(12345)

	depths := []int{1, 5, 10, 20, 30}
	var prevAvgDamage int

	for _, depth := range depths {
		params := procgen.GenerationParams{
			Difficulty: 0.5,
			Depth:      depth,
			GenreID:    "fantasy",
			Custom: map[string]interface{}{
				"count": 20,
			},
		}

		result, err := gen.Generate(seed, params)
		if err != nil {
			t.Fatalf("Generate() failed for depth %d: %v", depth, err)
		}

		spells := result.([]*Spell)
		totalDamage := 0
		damageCount := 0

		for _, spell := range spells {
			if spell.Stats.Damage > 0 {
				totalDamage += spell.Stats.Damage
				damageCount++
			}
		}

		if damageCount > 0 {
			avgDamage := totalDamage / damageCount
			t.Logf("Depth %d: Average damage = %d", depth, avgDamage)

			// Verify damage increases with depth
			if depth > 1 && avgDamage <= prevAvgDamage {
				t.Errorf("Damage did not increase with depth: %d <= %d at depth %d", avgDamage, prevAvgDamage, depth)
			}
			prevAvgDamage = avgDamage
		}
	}
}

func TestSpellGenerator_RarityDistribution(t *testing.T) {
	gen := NewSpellGenerator()

	tests := []struct {
		name       string
		depth      int
		difficulty float64
		wantCommon bool
		wantRare   bool
	}{
		{
			name:       "low depth should have mostly common",
			depth:      1,
			difficulty: 0.1,
			wantCommon: true,
			wantRare:   false,
		},
		{
			name:       "high depth should have some rare",
			depth:      20,
			difficulty: 0.8,
			wantCommon: true,
			wantRare:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := procgen.GenerationParams{
				Difficulty: tt.difficulty,
				Depth:      tt.depth,
				GenreID:    "fantasy",
				Custom: map[string]interface{}{
					"count": 100,
				},
			}

			result, err := gen.Generate(12345, params)
			if err != nil {
				t.Fatalf("Generate() failed: %v", err)
			}

			spells := result.([]*Spell)

			// Count rarities
			rarityCounts := make(map[Rarity]int)
			for _, spell := range spells {
				rarityCounts[spell.Rarity]++
			}

			hasCommon := rarityCounts[RarityCommon] > 0
			hasRare := rarityCounts[RarityRare] > 0 || rarityCounts[RarityEpic] > 0 || rarityCounts[RarityLegendary] > 0

			if tt.wantCommon && !hasCommon {
				t.Errorf("Expected common spells but found none")
			}
			if tt.wantRare && !hasRare {
				t.Errorf("Expected rare spells but found none")
			}

			t.Logf("Rarity distribution: Common=%d, Uncommon=%d, Rare=%d, Epic=%d, Legendary=%d",
				rarityCounts[RarityCommon], rarityCounts[RarityUncommon],
				rarityCounts[RarityRare], rarityCounts[RarityEpic], rarityCounts[RarityLegendary])
		})
	}
}

func TestSpellGenerator_GenreDifferences(t *testing.T) {
	gen := NewSpellGenerator()
	seed := int64(12345)

	params := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      5,
		Custom: map[string]interface{}{
			"count": 20,
		},
	}

	// Generate fantasy spells
	params.GenreID = "fantasy"
	fantasyResult, err := gen.Generate(seed, params)
	if err != nil {
		t.Fatalf("Fantasy generation failed: %v", err)
	}
	fantasySpells := fantasyResult.([]*Spell)

	// Generate sci-fi spells
	params.GenreID = "scifi"
	scifiResult, err := gen.Generate(seed, params)
	if err != nil {
		t.Fatalf("Sci-fi generation failed: %v", err)
	}
	scifiSpells := scifiResult.([]*Spell)

	// Both should generate spells
	if len(fantasySpells) == 0 {
		t.Error("No fantasy spells generated")
	}
	if len(scifiSpells) == 0 {
		t.Error("No sci-fi spells generated")
	}

	// Spells should be different (names will differ due to different templates)
	if fantasySpells[0].Name == scifiSpells[0].Name {
		t.Log("Warning: Fantasy and sci-fi spells have same name (possible but unlikely)")
	}
}

func TestSpell_IsOffensive(t *testing.T) {
	tests := []struct {
		name      string
		spellType SpellType
		want      bool
	}{
		{"offensive spell", TypeOffensive, true},
		{"debuff spell", TypeDebuff, true},
		{"healing spell", TypeHealing, false},
		{"buff spell", TypeBuff, false},
		{"defensive spell", TypeDefensive, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			spell := &Spell{Type: tt.spellType}
			if got := spell.IsOffensive(); got != tt.want {
				t.Errorf("IsOffensive() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSpell_IsSupport(t *testing.T) {
	tests := []struct {
		name      string
		spellType SpellType
		want      bool
	}{
		{"healing spell", TypeHealing, true},
		{"buff spell", TypeBuff, true},
		{"defensive spell", TypeDefensive, true},
		{"offensive spell", TypeOffensive, false},
		{"debuff spell", TypeDebuff, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			spell := &Spell{Type: tt.spellType}
			if got := spell.IsSupport(); got != tt.want {
				t.Errorf("IsSupport() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSpell_GetPowerLevel(t *testing.T) {
	tests := []struct {
		name     string
		spell    *Spell
		minPower int
		maxPower int
	}{
		{
			name: "weak common spell",
			spell: &Spell{
				Rarity: RarityCommon,
				Stats: Stats{
					Damage:   10,
					ManaCost: 10,
				},
			},
			minPower: 10,
			maxPower: 100,
		},
		{
			name: "powerful legendary spell",
			spell: &Spell{
				Rarity: RarityLegendary,
				Stats: Stats{
					Damage:   100,
					ManaCost: 50,
					Duration: 30,
					AreaSize: 10,
				},
			},
			minPower: 80,
			maxPower: 100,
		},
		{
			name: "zero power spell",
			spell: &Spell{
				Rarity: RarityCommon,
				Stats: Stats{
					ManaCost: 10,
				},
			},
			minPower: 0,
			maxPower: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			power := tt.spell.GetPowerLevel()
			if power < tt.minPower || power > tt.maxPower {
				t.Errorf("GetPowerLevel() = %d, want between %d and %d", power, tt.minPower, tt.maxPower)
			}
			t.Logf("%s: power level = %d", tt.name, power)
		})
	}
}

func TestSpellGenerator_Validate(t *testing.T) {
	gen := NewSpellGenerator()

	tests := []struct {
		name    string
		spells  []*Spell
		wantErr bool
	}{
		{
			name: "valid spells",
			spells: []*Spell{
				{
					Name:    "Fire Bolt",
					Type:    TypeOffensive,
					Element: ElementFire,
					Rarity:  RarityCommon,
					Target:  TargetSingle,
					Stats: Stats{
						Damage:        20,
						ManaCost:      10,
						Cooldown:      2.0,
						CastTime:      0.5,
						Range:         15.0,
						RequiredLevel: 1,
					},
				},
			},
			wantErr: false,
		},
		{
			name:    "empty spell list",
			spells:  []*Spell{},
			wantErr: true,
		},
		{
			name: "nil spell in list",
			spells: []*Spell{
				nil,
			},
			wantErr: true,
		},
		{
			name: "empty name",
			spells: []*Spell{
				{
					Name:    "",
					Type:    TypeOffensive,
					Element: ElementFire,
					Rarity:  RarityCommon,
					Target:  TargetSingle,
					Stats: Stats{
						Damage:        20,
						RequiredLevel: 1,
					},
				},
			},
			wantErr: true,
		},
		{
			name: "offensive spell with no damage",
			spells: []*Spell{
				{
					Name:    "Fire Bolt",
					Type:    TypeOffensive,
					Element: ElementFire,
					Rarity:  RarityCommon,
					Target:  TargetSingle,
					Stats: Stats{
						Damage:        0,
						RequiredLevel: 1,
					},
				},
			},
			wantErr: true,
		},
		{
			name: "healing spell with no healing",
			spells: []*Spell{
				{
					Name:    "Heal",
					Type:    TypeHealing,
					Element: ElementLight,
					Rarity:  RarityCommon,
					Target:  TargetSingle,
					Stats: Stats{
						Healing:       0,
						RequiredLevel: 1,
					},
				},
			},
			wantErr: true,
		},
		{
			name: "negative mana cost",
			spells: []*Spell{
				{
					Name:    "Fire Bolt",
					Type:    TypeOffensive,
					Element: ElementFire,
					Rarity:  RarityCommon,
					Target:  TargetSingle,
					Stats: Stats{
						Damage:        20,
						ManaCost:      -10,
						RequiredLevel: 1,
					},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := gen.Validate(tt.spells)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSpellGenerator_ValidateWrongType(t *testing.T) {
	gen := NewSpellGenerator()

	err := gen.Validate("not a spell slice")
	if err == nil {
		t.Error("Validate() should return error for wrong type")
	}
}

func TestSpellType_String(t *testing.T) {
	tests := []struct {
		spellType SpellType
		want      string
	}{
		{TypeOffensive, "offensive"},
		{TypeDefensive, "defensive"},
		{TypeHealing, "healing"},
		{TypeBuff, "buff"},
		{TypeDebuff, "debuff"},
		{TypeUtility, "utility"},
		{TypeSummon, "summon"},
		{SpellType(99), "unknown"},
	}

	for _, tt := range tests {
		if got := tt.spellType.String(); got != tt.want {
			t.Errorf("SpellType(%d).String() = %v, want %v", tt.spellType, got, tt.want)
		}
	}
}

func TestElementType_String(t *testing.T) {
	tests := []struct {
		element ElementType
		want    string
	}{
		{ElementFire, "fire"},
		{ElementIce, "ice"},
		{ElementLightning, "lightning"},
		{ElementEarth, "earth"},
		{ElementWind, "wind"},
		{ElementLight, "light"},
		{ElementDark, "dark"},
		{ElementArcane, "arcane"},
		{ElementNone, "none"},
		{ElementType(99), "unknown"},
	}

	for _, tt := range tests {
		if got := tt.element.String(); got != tt.want {
			t.Errorf("ElementType(%d).String() = %v, want %v", tt.element, got, tt.want)
		}
	}
}

func TestRarity_String(t *testing.T) {
	tests := []struct {
		rarity Rarity
		want   string
	}{
		{RarityCommon, "common"},
		{RarityUncommon, "uncommon"},
		{RarityRare, "rare"},
		{RarityEpic, "epic"},
		{RarityLegendary, "legendary"},
		{Rarity(99), "unknown"},
	}

	for _, tt := range tests {
		if got := tt.rarity.String(); got != tt.want {
			t.Errorf("Rarity(%d).String() = %v, want %v", tt.rarity, got, tt.want)
		}
	}
}

func TestTargetType_String(t *testing.T) {
	tests := []struct {
		target TargetType
		want   string
	}{
		{TargetSelf, "self"},
		{TargetSingle, "single"},
		{TargetArea, "area"},
		{TargetCone, "cone"},
		{TargetLine, "line"},
		{TargetAllAllies, "all_allies"},
		{TargetAllEnemies, "all_enemies"},
		{TargetType(99), "unknown"},
	}

	for _, tt := range tests {
		if got := tt.target.String(); got != tt.want {
			t.Errorf("TargetType(%d).String() = %v, want %v", tt.target, got, tt.want)
		}
	}
}

// Benchmarks
func BenchmarkSpellGenerator_Generate(b *testing.B) {
	gen := NewSpellGenerator()
	params := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      10,
		GenreID:    "fantasy",
		Custom: map[string]interface{}{
			"count": 20,
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := gen.Generate(int64(i), params)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkSpellGenerator_Validate(b *testing.B) {
	gen := NewSpellGenerator()
	params := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      10,
		GenreID:    "fantasy",
		Custom: map[string]interface{}{
			"count": 100,
		},
	}

	result, err := gen.Generate(12345, params)
	if err != nil {
		b.Fatal(err)
	}
	spells := result.([]*Spell)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = gen.Validate(spells)
	}
}
