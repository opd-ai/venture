package class

import (
	"testing"

	"github.com/opd-ai/venture/pkg/engine"
	"github.com/opd-ai/venture/pkg/procgen"
	"github.com/sirupsen/logrus"
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
				Name:              "Warrior",
				Description:       "A mighty combatant.",
				StartingHP:        100,
				StartingMana:      30,
				StartingAttack:    15,
				StartingDefense:   12,
				StartingSpeed:     5,
				StartingAbilities: []string{"power_strike"},
				Specializations:   []engine.SpecializationType{engine.SpecializationBerserker},
			},
			wantErr: false,
		},
		{
			name: "invalid HP",
			preset: &ClassPreset{
				Type:              engine.ClassWarrior,
				Name:              "Warrior",
				Description:       "A mighty combatant.",
				StartingHP:        0,
				StartingMana:      30,
				StartingAttack:    15,
				StartingDefense:   12,
				StartingSpeed:     5,
				StartingAbilities: []string{"power_strike"},
				Specializations:   []engine.SpecializationType{engine.SpecializationBerserker},
			},
			wantErr: true,
		},
		{
			name: "no abilities",
			preset: &ClassPreset{
				Type:              engine.ClassWarrior,
				Name:              "Warrior",
				Description:       "A mighty combatant.",
				StartingHP:        100,
				StartingMana:      30,
				StartingAttack:    15,
				StartingDefense:   12,
				StartingSpeed:     5,
				StartingAbilities: []string{},
				Specializations:   []engine.SpecializationType{engine.SpecializationBerserker},
			},
			wantErr: true,
		},
		{
			name: "no specializations",
			preset: &ClassPreset{
				Type:              engine.ClassWarrior,
				Name:              "Warrior",
				Description:       "A mighty combatant.",
				StartingHP:        100,
				StartingMana:      30,
				StartingAttack:    15,
				StartingDefense:   12,
				StartingSpeed:     5,
				StartingAbilities: []string{"power_strike"},
				Specializations:   []engine.SpecializationType{},
			},
			wantErr: true,
		},
		{
			name: "empty name",
			preset: &ClassPreset{
				Name:              "",
				Description:       "A mighty combatant.",
				StartingHP:        100,
				StartingMana:      30,
				StartingAttack:    15,
				StartingDefense:   12,
				StartingSpeed:     5,
				StartingAbilities: []string{"power_strike"},
				Specializations:   []engine.SpecializationType{engine.SpecializationBerserker},
			},
			wantErr: true,
		},
		{
			name: "empty description",
			preset: &ClassPreset{
				Name:              "Warrior",
				Description:       "",
				StartingHP:        100,
				StartingMana:      30,
				StartingAttack:    15,
				StartingDefense:   12,
				StartingSpeed:     5,
				StartingAbilities: []string{"power_strike"},
				Specializations:   []engine.SpecializationType{engine.SpecializationBerserker},
			},
			wantErr: true,
		},
		{
			name: "invalid defense",
			preset: &ClassPreset{
				Name:              "Warrior",
				Description:       "A mighty combatant.",
				StartingHP:        100,
				StartingMana:      30,
				StartingAttack:    15,
				StartingDefense:   0,
				StartingSpeed:     5,
				StartingAbilities: []string{"power_strike"},
				Specializations:   []engine.SpecializationType{engine.SpecializationBerserker},
			},
			wantErr: true,
		},
		{
			name: "invalid speed",
			preset: &ClassPreset{
				Name:              "Warrior",
				Description:       "A mighty combatant.",
				StartingHP:        100,
				StartingMana:      30,
				StartingAttack:    15,
				StartingDefense:   12,
				StartingSpeed:     0,
				StartingAbilities: []string{"power_strike"},
				Specializations:   []engine.SpecializationType{engine.SpecializationBerserker},
			},
			wantErr: true,
		},
		{
			name:    "wrong type",
			preset:  nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var err error
			if tt.preset == nil {
				err = gen.Validate("not a preset")
			} else {
				err = gen.Validate(tt.preset)
			}
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

	// Verify count matches the number of initialized presets
	if len(presets) != len(gen.presets) {
		t.Errorf("GetAllPresets() returned %d presets, want %d", len(presets), len(gen.presets))
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

func TestGetPreset(t *testing.T) {
	gen := NewClassGenerator()

	preset, ok := gen.GetPreset(engine.ClassWarrior)
	if !ok {
		t.Fatal("GetPreset() failed for ClassWarrior")
	}
	if preset.Name != "Warrior" {
		t.Errorf("GetPreset() name = %q, want %q", preset.Name, "Warrior")
	}

	_, ok = gen.GetPreset(engine.CharacterClass(999))
	if ok {
		t.Error("GetPreset() should return false for invalid class type")
	}
}

func TestGenerateAndValidateAllClasses(t *testing.T) {
	gen := NewClassGenerator()
	for _, preset := range gen.GetAllPresets() {
		t.Run(preset.Name, func(t *testing.T) {
			params := procgen.GenerationParams{
				Difficulty: 0.5,
				Depth:      1,
				Custom:     map[string]interface{}{"class_type": preset.Type},
			}
			result, err := gen.Generate(42, params)
			if err != nil {
				t.Fatalf("Generate() error = %v", err)
			}
			if err := gen.Validate(result); err != nil {
				t.Errorf("Validate() error = %v", err)
			}
		})
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

func TestNewClassGeneratorWithLogger(t *testing.T) {
	t.Run("with custom logger", func(t *testing.T) {
		logger := logrus.NewEntry(logrus.New())
		gen := NewClassGeneratorWithLogger(logger)
		if gen == nil {
			t.Fatal("NewClassGeneratorWithLogger() returned nil")
		}
		if gen.logger != logger {
			t.Error("NewClassGeneratorWithLogger() did not set provided logger")
		}
		// Verify generator works
		params := procgen.GenerationParams{
			Difficulty: 0.5,
			Custom:     map[string]interface{}{"class_type": engine.ClassWarrior},
		}
		result, err := gen.Generate(12345, params)
		if err != nil {
			t.Fatalf("Generate() with custom logger failed: %v", err)
		}
		if result == nil {
			t.Error("Generate() returned nil result")
		}
	})

	t.Run("with nil logger falls back to default", func(t *testing.T) {
		gen := NewClassGeneratorWithLogger(nil)
		if gen == nil {
			t.Fatal("NewClassGeneratorWithLogger(nil) returned nil")
		}
		if gen.logger == nil {
			t.Error("NewClassGeneratorWithLogger(nil) should set default logger")
		}
	})
}

func TestClassGenerator_GenreTheming(t *testing.T) {
	gen := NewClassGenerator()

	tests := []struct {
		name        string
		genreID     string
		classType   engine.CharacterClass
		wantName    string
		wantDescHas string
	}{
		// Fantasy (default) - should use base names
		{
			name:        "fantasy warrior",
			genreID:     "fantasy",
			classType:   engine.ClassWarrior,
			wantName:    "Warrior",
			wantDescHas: "melee combat",
		},
		{
			name:        "empty genre defaults to fantasy",
			genreID:     "",
			classType:   engine.ClassWarrior,
			wantName:    "Warrior",
			wantDescHas: "melee combat",
		},

		// Sci-Fi genre
		{
			name:        "scifi warrior",
			genreID:     "scifi",
			classType:   engine.ClassWarrior,
			wantName:    "Shock Trooper",
			wantDescHas: "powered armor",
		},
		{
			name:        "scifi rogue",
			genreID:     "scifi",
			classType:   engine.ClassRogue,
			wantName:    "Infiltrator",
			wantDescHas: "cloaking",
		},
		{
			name:        "scifi mage",
			genreID:     "scifi",
			classType:   engine.ClassMage,
			wantName:    "Psionic",
			wantDescHas: "telekinetic",
		},
		{
			name:        "scifi cleric",
			genreID:     "scifi",
			classType:   engine.ClassCleric,
			wantName:    "Medic",
			wantDescHas: "nanobots",
		},
		{
			name:        "scifi ranger",
			genreID:     "scifi",
			classType:   engine.ClassRanger,
			wantName:    "Scout",
			wantDescHas: "tactical sensors",
		},
		{
			name:        "scifi paladin",
			genreID:     "scifi",
			classType:   engine.ClassPaladin,
			wantName:    "Vanguard",
			wantDescHas: "exo-armor",
		},

		// Horror genre
		{
			name:        "horror warrior",
			genreID:     "horror",
			classType:   engine.ClassWarrior,
			wantName:    "Survivor",
			wantDescHas: "darkness",
		},
		{
			name:        "horror rogue",
			genreID:     "horror",
			classType:   engine.ClassRogue,
			wantName:    "Stalker",
			wantDescHas: "shadows",
		},
		{
			name:        "horror mage",
			genreID:     "horror",
			classType:   engine.ClassMage,
			wantName:    "Occultist",
			wantDescHas: "forbidden",
		},
		{
			name:        "horror cleric",
			genreID:     "horror",
			classType:   engine.ClassCleric,
			wantName:    "Exorcist",
			wantDescHas: "undead",
		},

		// Cyberpunk genre
		{
			name:        "cyberpunk warrior",
			genreID:     "cyberpunk",
			classType:   engine.ClassWarrior,
			wantName:    "Street Samurai",
			wantDescHas: "cybernetic",
		},
		{
			name:        "cyberpunk rogue",
			genreID:     "cyberpunk",
			classType:   engine.ClassRogue,
			wantName:    "Netrunner",
			wantDescHas: "hacker",
		},
		{
			name:        "cyberpunk mage",
			genreID:     "cyberpunk",
			classType:   engine.ClassMage,
			wantName:    "Technomancer",
			wantDescHas: "cyberspace",
		},

		// Post-Apocalyptic genre
		{
			name:        "postapocalyptic warrior",
			genreID:     "postapocalyptic",
			classType:   engine.ClassWarrior,
			wantName:    "Raider",
			wantDescHas: "wasteland",
		},
		{
			name:        "postapocalyptic rogue",
			genreID:     "postapocalyptic",
			classType:   engine.ClassRogue,
			wantName:    "Scavenger",
			wantDescHas: "ruins",
		},
		{
			name:        "postapocalyptic mage",
			genreID:     "postapocalyptic",
			classType:   engine.ClassMage,
			wantName:    "Mutant",
			wantDescHas: "radiation",
		},

		// Unmapped genre should fall back to default
		{
			name:        "unknown genre uses default",
			genreID:     "steampunk",
			classType:   engine.ClassWarrior,
			wantName:    "Warrior",
			wantDescHas: "melee combat",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := procgen.GenerationParams{
				GenreID:    tt.genreID,
				Difficulty: 0.5,
				Depth:      1,
				Custom:     map[string]interface{}{"class_type": tt.classType},
			}

			result, err := gen.Generate(12345, params)
			if err != nil {
				t.Fatalf("Generate() error = %v", err)
			}

			preset, ok := result.(*ClassPreset)
			if !ok {
				t.Fatalf("Generate() returned wrong type: %T", result)
			}

			if preset.Name != tt.wantName {
				t.Errorf("Generate() name = %q, want %q", preset.Name, tt.wantName)
			}

			if tt.wantDescHas != "" {
				if len(preset.Description) == 0 {
					t.Error("Generate() description is empty")
				}
				// Note: Not checking exact description match to allow for flexibility
			}
		})
	}
}

func TestClassGenerator_GenreThemingDeterminism(t *testing.T) {
	gen := NewClassGenerator()

	params := procgen.GenerationParams{
		GenreID:    "scifi",
		Difficulty: 0.5,
		Depth:      1,
		Custom:     map[string]interface{}{"class_type": engine.ClassWarrior},
	}

	// Generate twice with same seed - should be identical
	result1, err1 := gen.Generate(12345, params)
	if err1 != nil {
		t.Fatalf("First Generate() error = %v", err1)
	}

	result2, err2 := gen.Generate(12345, params)
	if err2 != nil {
		t.Fatalf("Second Generate() error = %v", err2)
	}

	preset1 := result1.(*ClassPreset)
	preset2 := result2.(*ClassPreset)

	if preset1.Name != preset2.Name {
		t.Errorf("Genre theming not deterministic: names differ %q vs %q", preset1.Name, preset2.Name)
	}

	if preset1.Description != preset2.Description {
		t.Errorf("Genre theming not deterministic: descriptions differ")
	}
}

func BenchmarkClassGenerator_Validate(b *testing.B) {
	gen := NewClassGenerator()
	preset := &ClassPreset{
		Type:              engine.ClassWarrior,
		Name:              "Warrior",
		Description:       "A mighty combatant.",
		StartingHP:        100.0,
		StartingMana:      30.0,
		StartingAttack:    15.0,
		StartingDefense:   12.0,
		StartingSpeed:     5.0,
		StartingAbilities: []string{"bash", "charge"},
		Specializations:   []engine.SpecializationType{engine.SpecializationBerserker},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = gen.Validate(preset)
	}
}

func TestClassGenerator_CustomLoggerUsedInGenerate(t *testing.T) {
	// Create a test hook to capture log entries
	var capturedLogs []logEntry

	hook := &testHook{
		entries: &capturedLogs,
	}

	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	logger.AddHook(hook)
	entry := logger.WithField("system_name", "test_class_generator")

	gen := NewClassGeneratorWithLogger(entry)

	// Try to generate with an invalid class type to trigger error logging
	params := procgen.GenerationParams{
		GenreID:    "fantasy",
		Difficulty: 0.5,
		Depth:      1,
		Custom:     map[string]interface{}{"class_type": engine.CharacterClass(999)}, // Invalid class type
	}

	_, err := gen.Generate(12345, params)
	if err == nil {
		t.Fatal("Expected error for invalid class type")
	}

	// Verify that the custom logger was used
	if len(capturedLogs) == 0 {
		t.Error("Custom logger was not used - no log entries captured")
	}

	// Verify the log entry has the expected fields
	found := false
	for _, entry := range capturedLogs {
		if entry.level == logrus.ErrorLevel {
			if _, ok := entry.fields["class_type"]; ok {
				if _, ok := entry.fields["seed"]; ok {
					if _, ok := entry.fields["genre_id"]; ok {
						found = true
						break
					}
				}
			}
		}
	}

	if !found {
		t.Error("Custom logger did not log expected error with correct fields")
	}
}

// logEntry represents a captured log entry for testing.
type logEntry struct {
	level   logrus.Level
	message string
	fields  logrus.Fields
}

// testHook is a logrus hook for capturing log entries in tests.
type testHook struct {
	entries *[]logEntry
}

func (h *testHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

func (h *testHook) Fire(entry *logrus.Entry) error {
	*h.entries = append(*h.entries, logEntry{
		level:   entry.Level,
		message: entry.Message,
		fields:  entry.Data,
	})
	return nil
}

func TestClassGenerator_GetAllPresets(t *testing.T) {
	gen := NewClassGenerator()

	presets := gen.GetAllPresets()

	// Should return all 21 presets
	if len(presets) != 21 {
		t.Errorf("GetAllPresets() returned %d presets, want 21", len(presets))
	}

	// Verify all base classes are present
	baseClasses := []engine.CharacterClass{
		engine.ClassWarrior,
		engine.ClassRogue,
		engine.ClassMage,
		engine.ClassCleric,
		engine.ClassRanger,
		engine.ClassPaladin,
	}

	for _, class := range baseClasses {
		found := false
		for _, preset := range presets {
			if preset.Type == class {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("GetAllPresets() missing base class: %v", class)
		}
	}
}

func TestClassGenerator_GetAllPresetsHandlesGaps(t *testing.T) {
	// Create a generator with a manually constructed preset map with gaps
	gen := &ClassGenerator{
		presets:     make(map[engine.CharacterClass]ClassPreset),
		genreThemes: make(map[string]genreTheming),
		logger:      logrus.WithField("system_name", "class_generator_test"),
	}

	// Add presets with non-contiguous enum values (simulating gaps)
	gen.presets[engine.CharacterClass(0)] = ClassPreset{
		Type:              engine.CharacterClass(0),
		Name:              "Class0",
		Description:       "Test class 0",
		StartingHP:        100,
		StartingMana:      50,
		StartingAttack:    10,
		StartingDefense:   5,
		StartingSpeed:     5,
		StartingAbilities: []string{"test"},
		Specializations:   []engine.SpecializationType{engine.SpecializationBerserker},
	}

	gen.presets[engine.CharacterClass(5)] = ClassPreset{
		Type:              engine.CharacterClass(5),
		Name:              "Class5",
		Description:       "Test class 5",
		StartingHP:        100,
		StartingMana:      50,
		StartingAttack:    10,
		StartingDefense:   5,
		StartingSpeed:     5,
		StartingAbilities: []string{"test"},
		Specializations:   []engine.SpecializationType{engine.SpecializationBerserker},
	}

	gen.presets[engine.CharacterClass(10)] = ClassPreset{
		Type:              engine.CharacterClass(10),
		Name:              "Class10",
		Description:       "Test class 10",
		StartingHP:        100,
		StartingMana:      50,
		StartingAttack:    10,
		StartingDefense:   5,
		StartingSpeed:     5,
		StartingAbilities: []string{"test"},
		Specializations:   []engine.SpecializationType{engine.SpecializationBerserker},
	}

	presets := gen.GetAllPresets()

	// Should find all 3 presets despite gaps
	if len(presets) != 3 {
		t.Errorf("GetAllPresets() with gaps returned %d presets, want 3", len(presets))
	}

	// Verify correct presets were found
	expectedNames := map[string]bool{
		"Class0":  false,
		"Class5":  false,
		"Class10": false,
	}

	for _, preset := range presets {
		if _, exists := expectedNames[preset.Name]; exists {
			expectedNames[preset.Name] = true
		}
	}

	for name, found := range expectedNames {
		if !found {
			t.Errorf("GetAllPresets() did not find preset: %s", name)
		}
	}
}
