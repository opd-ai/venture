package class

import (
	"testing"

	"github.com/opd-ai/venture/pkg/engine"
	"github.com/opd-ai/venture/pkg/procgen"
)

func TestClassGenerator_Generate(t *testing.T) {
	gen := NewClassGenerator()

	tests := []struct {
		name      string
		seed      int64
		params    procgen.GenerationParams
		classType engine.CharacterClass
		wantErr   bool
	}{
		{
			name: "warrior generation",
			seed: 12345,
			params: procgen.GenerationParams{
				Difficulty: 0.5,
				Depth:      1,
				Custom:     map[string]interface{}{"class_type": engine.ClassWarrior},
			},
			classType: engine.ClassWarrior,
			wantErr:   false,
		},
		{
			name: "rogue generation",
			seed: 54321,
			params: procgen.GenerationParams{
				Difficulty: 0.7,
				Depth:      5,
				Custom:     map[string]interface{}{"class_type": engine.ClassRogue},
			},
			classType: engine.ClassRogue,
			wantErr:   false,
		},
		{
			name: "mage generation",
			seed: 99999,
			params: procgen.GenerationParams{
				Difficulty: 0.3,
				Depth:      10,
				Custom:     map[string]interface{}{"class_type": engine.ClassMage},
			},
			classType: engine.ClassMage,
			wantErr:   false,
		},
		{
			name: "random class generation",
			seed: 11111,
			params: procgen.GenerationParams{
				Difficulty: 0.5,
				Depth:      1,
				Custom:     map[string]interface{}{},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := gen.Generate(tt.seed, tt.params)
			if (err != nil) != tt.wantErr {
				t.Errorf("Generate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err != nil {
				return
			}

			preset, ok := result.(*ClassPreset)
			if !ok {
				t.Errorf("Generate() returned wrong type")
				return
			}

			// Validate class type if specified
			if tt.classType != 0 && preset.Type != tt.classType {
				t.Errorf("Generate() class type = %v, want %v", preset.Type, tt.classType)
			}

			// Validate generated data
			if preset.StartingHP <= 0 {
				t.Errorf("Generate() invalid HP: %f", preset.StartingHP)
			}

			if preset.StartingMana < 0 {
				t.Errorf("Generate() invalid mana: %f", preset.StartingMana)
			}

			if len(preset.StartingAbilities) == 0 {
				t.Errorf("Generate() no starting abilities")
			}

			if len(preset.Specializations) == 0 {
				t.Errorf("Generate() no specializations")
			}
		})
	}
}

func TestClassGenerator_Validate(t *testing.T) {
	gen := NewClassGenerator()

	tests := []struct {
		name    string
		preset  *ClassPreset
		wantErr bool
	}{
		{
			name: "valid warrior",
			preset: &ClassPreset{
				Type:              engine.ClassWarrior,
				StartingHP:        100,
				StartingMana:      30,
				StartingAttack:    15,
				StartingAbilities: []string{"power_strike"},
				Specializations:   []engine.SpecializationType{engine.SpecializationBerserker},
			},
			wantErr: false,
		},
		{
			name: "invalid HP",
			preset: &ClassPreset{
				Type:              engine.ClassWarrior,
				StartingHP:        0,
				StartingMana:      30,
				StartingAttack:    15,
				StartingAbilities: []string{"power_strike"},
				Specializations:   []engine.SpecializationType{engine.SpecializationBerserker},
			},
			wantErr: true,
		},
		{
			name: "no abilities",
			preset: &ClassPreset{
				Type:              engine.ClassWarrior,
				StartingHP:        100,
				StartingMana:      30,
				StartingAttack:    15,
				StartingAbilities: []string{},
				Specializations:   []engine.SpecializationType{engine.SpecializationBerserker},
			},
			wantErr: true,
		},
		{
			name: "no specializations",
			preset: &ClassPreset{
				Type:              engine.ClassWarrior,
				StartingHP:        100,
				StartingMana:      30,
				StartingAttack:    15,
				StartingAbilities: []string{"power_strike"},
				Specializations:   []engine.SpecializationType{},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := gen.Validate(tt.preset)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestClassGenerator_Determinism(t *testing.T) {
	gen := NewClassGenerator()
	seed := int64(42)
	params := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      5,
		Custom:     map[string]interface{}{"class_type": engine.ClassMage},
	}

	// Generate twice with same seed
	result1, err1 := gen.Generate(seed, params)
	result2, err2 := gen.Generate(seed, params)

	if err1 != nil || err2 != nil {
		t.Fatalf("Generate() errors: %v, %v", err1, err2)
	}

	preset1 := result1.(*ClassPreset)
	preset2 := result2.(*ClassPreset)

	// Check determinism
	if preset1.StartingHP != preset2.StartingHP {
		t.Errorf("Non-deterministic HP: %f != %f", preset1.StartingHP, preset2.StartingHP)
	}

	if preset1.StartingMana != preset2.StartingMana {
		t.Errorf("Non-deterministic mana: %f != %f", preset1.StartingMana, preset2.StartingMana)
	}

	if preset1.StartingAttack != preset2.StartingAttack {
		t.Errorf("Non-deterministic attack: %f != %f", preset1.StartingAttack, preset2.StartingAttack)
	}
}

func TestGetAllPresets(t *testing.T) {
	gen := NewClassGenerator()
	presets := gen.GetAllPresets()

	if len(presets) != 21 {
		t.Errorf("GetAllPresets() returned %d presets, want 21 (6 base + 15 hybrid)", len(presets))
	}

	// Check all class types are present
	types := make(map[engine.CharacterClass]bool)
	for _, preset := range presets {
		types[preset.Type] = true
	}

	expectedTypes := []engine.CharacterClass{
		engine.ClassWarrior,
		engine.ClassRogue,
		engine.ClassMage,
		engine.ClassRanger,
		engine.ClassCleric,
		engine.ClassNecromancer,
	}

	for _, expectedType := range expectedTypes {
		if !types[expectedType] {
			t.Errorf("GetAllPresets() missing class type: %v", expectedType)
		}
	}
}

func BenchmarkClassGenerator_Generate(b *testing.B) {
	gen := NewClassGenerator()
	params := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      5,
		Custom:     map[string]interface{}{"class_type": engine.ClassWarrior},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = gen.Generate(int64(i), params)
	}
}
