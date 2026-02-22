package advanced

import (
	"encoding/json"
	"testing"
)

func TestAdvancedClassComponent_Serialize(t *testing.T) {
	tests := []struct {
		name    string
		comp    *AdvancedClassComponent
		wantErr bool
	}{
		{
			name: "full component",
			comp: &AdvancedClassComponent{
				PrimaryClass:   ClassWarrior,
				SecondaryClass: ClassMage,
				PrestigeClass:  PrestigeBattlemage,
				Level:          25,
				RespecCount:    3,
				TalentPoints: TalentAllocation{
					Talents:     map[TalentID]int{"warrior_power_strike": 2, "mage_fireball": 1},
					PointsSpent: 3,
					PointsTotal: 10,
				},
			},
			wantErr: false,
		},
		{
			name: "empty talents",
			comp: &AdvancedClassComponent{
				PrimaryClass: ClassRogue,
				Level:        1,
				TalentPoints: TalentAllocation{
					Talents:     map[TalentID]int{},
					PointsSpent: 0,
					PointsTotal: 5,
				},
			},
			wantErr: false,
		},
		{
			name: "nil talents map",
			comp: &AdvancedClassComponent{
				PrimaryClass: ClassCleric,
				Level:        5,
				TalentPoints: TalentAllocation{
					Talents:     nil,
					PointsSpent: 0,
					PointsTotal: 3,
				},
			},
			wantErr: false,
		},
		{
			name:    "default/zero values",
			comp:    &AdvancedClassComponent{},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.comp.Serialize()
			if (err != nil) != tt.wantErr {
				t.Errorf("Serialize() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && len(got) == 0 {
				t.Error("Serialize() returned empty bytes")
			}
			// Verify it's valid JSON
			var data map[string]interface{}
			if err := json.Unmarshal(got, &data); err != nil {
				t.Errorf("Serialize() produced invalid JSON: %v", err)
			}
		})
	}
}

func TestAdvancedClassComponent_Deserialize(t *testing.T) {
	tests := []struct {
		name    string
		json    string
		wantErr bool
	}{
		{
			name: "valid full data",
			json: `{
				"primary_class": "warrior",
				"secondary_class": "mage",
				"prestige_class": "battlemage",
				"talent_points": {
					"talents": {"warrior_power_strike": 2},
					"points_spent": 2,
					"points_total": 10
				},
				"level": 25,
				"respec_count": 1
			}`,
			wantErr: false,
		},
		{
			name: "minimal data",
			json: `{
				"primary_class": "rogue",
				"level": 1,
				"talent_points": {"points_total": 5}
			}`,
			wantErr: false,
		},
		{
			name:    "empty object",
			json:    `{}`,
			wantErr: false,
		},
		{
			name:    "invalid JSON",
			json:    `{invalid}`,
			wantErr: true,
		},
		{
			name:    "wrong type",
			json:    `{"level": "not_a_number"}`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			comp := &AdvancedClassComponent{}
			err := comp.Deserialize([]byte(tt.json))
			if (err != nil) != tt.wantErr {
				t.Errorf("Deserialize() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAdvancedClassComponent_RoundTrip(t *testing.T) {
	tests := []struct {
		name string
		comp *AdvancedClassComponent
	}{
		{
			name: "full component",
			comp: &AdvancedClassComponent{
				PrimaryClass:   ClassWarrior,
				SecondaryClass: ClassMage,
				PrestigeClass:  PrestigeBattlemage,
				Level:          25,
				RespecCount:    3,
				TalentPoints: TalentAllocation{
					Talents:     map[TalentID]int{"warrior_power_strike": 2, "mage_fireball": 1},
					PointsSpent: 3,
					PointsTotal: 10,
				},
			},
		},
		{
			name: "single class only",
			comp: &AdvancedClassComponent{
				PrimaryClass: ClassRogue,
				Level:        10,
				TalentPoints: TalentAllocation{
					Talents:     map[TalentID]int{"rogue_backstab": 5},
					PointsSpent: 5,
					PointsTotal: 5,
				},
			},
		},
		{
			name: "empty talents",
			comp: &AdvancedClassComponent{
				PrimaryClass: ClassCleric,
				Level:        1,
				TalentPoints: TalentAllocation{
					Talents:     map[TalentID]int{},
					PointsSpent: 0,
					PointsTotal: 0,
				},
			},
		},
		{
			name: "nil talents",
			comp: &AdvancedClassComponent{
				PrimaryClass: ClassBard,
				Level:        5,
				TalentPoints: TalentAllocation{
					Talents:     nil,
					PointsSpent: 0,
					PointsTotal: 2,
				},
			},
		},
		{
			name: "all prestige classes set",
			comp: &AdvancedClassComponent{
				PrimaryClass:   ClassMage,
				SecondaryClass: ClassWarrior,
				PrestigeClass:  PrestigeSpellblade,
				Level:          30,
				RespecCount:    10,
				TalentPoints: TalentAllocation{
					Talents: map[TalentID]int{
						"mage_arcane_mastery":  3,
						"warrior_shield_bash":  2,
						"warrior_intimidation": 1,
					},
					PointsSpent: 6,
					PointsTotal: 15,
				},
			},
		},
		{
			name: "zero values",
			comp: &AdvancedClassComponent{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Serialize
			data, err := tt.comp.Serialize()
			if err != nil {
				t.Fatalf("Serialize() error = %v", err)
			}

			// Deserialize into new component
			restored := &AdvancedClassComponent{}
			if err := restored.Deserialize(data); err != nil {
				t.Fatalf("Deserialize() error = %v", err)
			}

			// Compare fields
			if restored.PrimaryClass != tt.comp.PrimaryClass {
				t.Errorf("PrimaryClass = %v, want %v", restored.PrimaryClass, tt.comp.PrimaryClass)
			}
			if restored.SecondaryClass != tt.comp.SecondaryClass {
				t.Errorf("SecondaryClass = %v, want %v", restored.SecondaryClass, tt.comp.SecondaryClass)
			}
			if restored.PrestigeClass != tt.comp.PrestigeClass {
				t.Errorf("PrestigeClass = %v, want %v", restored.PrestigeClass, tt.comp.PrestigeClass)
			}
			if restored.Level != tt.comp.Level {
				t.Errorf("Level = %v, want %v", restored.Level, tt.comp.Level)
			}
			if restored.RespecCount != tt.comp.RespecCount {
				t.Errorf("RespecCount = %v, want %v", restored.RespecCount, tt.comp.RespecCount)
			}
			if restored.TalentPoints.PointsSpent != tt.comp.TalentPoints.PointsSpent {
				t.Errorf("PointsSpent = %v, want %v", restored.TalentPoints.PointsSpent, tt.comp.TalentPoints.PointsSpent)
			}
			if restored.TalentPoints.PointsTotal != tt.comp.TalentPoints.PointsTotal {
				t.Errorf("PointsTotal = %v, want %v", restored.TalentPoints.PointsTotal, tt.comp.TalentPoints.PointsTotal)
			}

			// Compare talents map
			if tt.comp.TalentPoints.Talents == nil {
				if restored.TalentPoints.Talents != nil {
					t.Errorf("Talents should be nil, got %v", restored.TalentPoints.Talents)
				}
			} else {
				if len(restored.TalentPoints.Talents) != len(tt.comp.TalentPoints.Talents) {
					t.Errorf("Talents length = %v, want %v", len(restored.TalentPoints.Talents), len(tt.comp.TalentPoints.Talents))
				}
				for talentID, rank := range tt.comp.TalentPoints.Talents {
					if restoredRank, ok := restored.TalentPoints.Talents[talentID]; !ok || restoredRank != rank {
						t.Errorf("Talent[%s] = %v, want %v", talentID, restoredRank, rank)
					}
				}
			}
		})
	}
}

func BenchmarkAdvancedClassComponent_Serialize(b *testing.B) {
	comp := &AdvancedClassComponent{
		PrimaryClass:   ClassWarrior,
		SecondaryClass: ClassMage,
		PrestigeClass:  PrestigeBattlemage,
		Level:          25,
		RespecCount:    3,
		TalentPoints: TalentAllocation{
			Talents: map[TalentID]int{
				"warrior_power_strike": 2,
				"mage_fireball":        1,
				"mage_arcane_mastery":  3,
			},
			PointsSpent: 6,
			PointsTotal: 15,
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = comp.Serialize()
	}
}

func BenchmarkAdvancedClassComponent_Deserialize(b *testing.B) {
	comp := &AdvancedClassComponent{
		PrimaryClass:   ClassWarrior,
		SecondaryClass: ClassMage,
		PrestigeClass:  PrestigeBattlemage,
		Level:          25,
		RespecCount:    3,
		TalentPoints: TalentAllocation{
			Talents: map[TalentID]int{
				"warrior_power_strike": 2,
				"mage_fireball":        1,
				"mage_arcane_mastery":  3,
			},
			PointsSpent: 6,
			PointsTotal: 15,
		},
	}

	data, _ := comp.Serialize()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		restored := &AdvancedClassComponent{}
		_ = restored.Deserialize(data)
	}
}
