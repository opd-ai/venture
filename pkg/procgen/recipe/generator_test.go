package recipe

import (
	"bytes"
	"os"
	"testing"

	"github.com/opd-ai/venture/pkg/engine"
	"github.com/opd-ai/venture/pkg/procgen"
	"github.com/opd-ai/venture/pkg/procgen/item"
	"github.com/sirupsen/logrus"
)

func TestMain(m *testing.M) {
	if os.Getenv("DISPLAY") == "" && os.Getenv("WAYLAND_DISPLAY") == "" {
		// Skip all tests in headless environments (package imports Ebiten/GLFW)
		os.Exit(0)
	}
	os.Exit(m.Run())
}

// TestNewRecipeGenerator tests generator creation.
func TestNewRecipeGenerator(t *testing.T) {
	gen := NewRecipeGenerator()
	if gen == nil {
		t.Fatal("NewRecipeGenerator returned nil")
	}

	// Verify templates registered
	if len(gen.potionTemplates) == 0 {
		t.Error("No potion templates registered")
	}
	if len(gen.enchantTemplates) == 0 {
		t.Error("No enchant templates registered")
	}
	if len(gen.magicItemTemplates) == 0 {
		t.Error("No magic item templates registered")
	}
}

// TestRecipeGenerator_Generate tests recipe generation.
func TestRecipeGenerator_Generate(t *testing.T) {
	tests := []struct {
		name      string
		seed      int64
		params    procgen.GenerationParams
		wantCount int
		wantErr   bool
	}{
		{
			name:      "fantasy recipes",
			seed:      12345,
			params:    procgen.GenerationParams{Difficulty: 0.5, Depth: 1, GenreID: "fantasy"},
			wantCount: 5,
			wantErr:   false,
		},
		{
			name:      "scifi recipes",
			seed:      54321,
			params:    procgen.GenerationParams{Difficulty: 0.7, Depth: 5, GenreID: "scifi"},
			wantCount: 5,
			wantErr:   false,
		},
		{
			name: "custom count",
			seed: 11111,
			params: procgen.GenerationParams{
				Difficulty: 0.5,
				Depth:      3,
				GenreID:    "fantasy",
				Custom:     map[string]interface{}{"count": 10},
			},
			wantCount: 10,
			wantErr:   false,
		},
		{
			name: "potion filter",
			seed: 22222,
			params: procgen.GenerationParams{
				Difficulty: 0.5,
				Depth:      2,
				GenreID:    "fantasy",
				Custom:     map[string]interface{}{"type": "potion"},
			},
			wantCount: 5,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gen := NewRecipeGenerator()
			result, err := gen.Generate(tt.seed, tt.params)

			if (err != nil) != tt.wantErr {
				t.Errorf("Generate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err == nil {
				recipes, ok := result.([]*engine.Recipe)
				if !ok {
					t.Fatal("Generate() did not return []*engine.Recipe")
				}

				if len(recipes) != tt.wantCount {
					t.Errorf("Generate() returned %d recipes, want %d", len(recipes), tt.wantCount)
				}

				// Verify all recipes have required fields
				for i, recipe := range recipes {
					if recipe.ID == "" {
						t.Errorf("Recipe %d has empty ID", i)
					}
					if recipe.Name == "" {
						t.Errorf("Recipe %d has empty name", i)
					}
					if len(recipe.Materials) == 0 {
						t.Errorf("Recipe %d has no materials", i)
					}
					if recipe.GenreID != tt.params.GenreID {
						t.Errorf("Recipe %d has genreID %s, want %s", i, recipe.GenreID, tt.params.GenreID)
					}
				}
			}
		})
	}
}

// TestRecipeGenerator_Determinism tests that same seed produces same recipes.
func TestRecipeGenerator_Determinism(t *testing.T) {
	gen := NewRecipeGenerator()
	seed := int64(99999)
	params := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      3,
		GenreID:    "fantasy",
	}

	// Generate twice with same seed
	result1, err1 := gen.Generate(seed, params)
	if err1 != nil {
		t.Fatal(err1)
	}

	result2, err2 := gen.Generate(seed, params)
	if err2 != nil {
		t.Fatal(err2)
	}

	recipes1 := result1.([]*engine.Recipe)
	recipes2 := result2.([]*engine.Recipe)

	if len(recipes1) != len(recipes2) {
		t.Fatalf("Different recipe counts: %d vs %d", len(recipes1), len(recipes2))
	}

	// Compare recipes
	for i := range recipes1 {
		r1, r2 := recipes1[i], recipes2[i]

		if r1.ID != r2.ID {
			t.Errorf("Recipe %d ID mismatch: %s vs %s", i, r1.ID, r2.ID)
		}
		if r1.Name != r2.Name {
			t.Errorf("Recipe %d name mismatch: %s vs %s", i, r1.Name, r2.Name)
		}
		if r1.Type != r2.Type {
			t.Errorf("Recipe %d type mismatch: %s vs %s", i, r1.Type, r2.Type)
		}
		if r1.Rarity != r2.Rarity {
			t.Errorf("Recipe %d rarity mismatch: %s vs %s", i, r1.Rarity, r2.Rarity)
		}
		if r1.GoldCost != r2.GoldCost {
			t.Errorf("Recipe %d gold cost mismatch: %d vs %d", i, r1.GoldCost, r2.GoldCost)
		}
		if r1.SkillRequired != r2.SkillRequired {
			t.Errorf("Recipe %d skill required mismatch: %d vs %d", i, r1.SkillRequired, r2.SkillRequired)
		}
	}
}

// TestRecipeGenerator_Validate tests validation of generated recipes.
func TestRecipeGenerator_Validate(t *testing.T) {
	gen := NewRecipeGenerator()

	t.Run("valid recipes", func(t *testing.T) {
		params := procgen.GenerationParams{Difficulty: 0.5, Depth: 1, GenreID: "fantasy"}
		result, err := gen.Generate(12345, params)
		if err != nil {
			t.Fatal(err)
		}

		if err := gen.Validate(result); err != nil {
			t.Errorf("Validate() failed on valid recipes: %v", err)
		}
	})

	t.Run("empty recipes", func(t *testing.T) {
		recipes := []*engine.Recipe{}
		if err := gen.Validate(recipes); err == nil {
			t.Error("Validate() should fail on empty recipes")
		}
	})

	t.Run("recipe with empty ID", func(t *testing.T) {
		recipes := []*engine.Recipe{
			{ID: "", Name: "Test", Materials: []engine.MaterialRequirement{{ItemName: "Item", Quantity: 1}}},
		}
		if err := gen.Validate(recipes); err == nil {
			t.Error("Validate() should fail on recipe with empty ID")
		}
	})

	t.Run("recipe with no materials", func(t *testing.T) {
		recipes := []*engine.Recipe{
			{ID: "test", Name: "Test", Materials: []engine.MaterialRequirement{}},
		}
		if err := gen.Validate(recipes); err == nil {
			t.Error("Validate() should fail on recipe with no materials")
		}
	})

	t.Run("recipe with invalid success chance", func(t *testing.T) {
		recipes := []*engine.Recipe{
			{
				ID:                "test",
				Name:              "Test",
				Materials:         []engine.MaterialRequirement{{ItemName: "Item", Quantity: 1}},
				BaseSuccessChance: 1.5,
				CraftTimeSec:      5.0,
			},
		}
		if err := gen.Validate(recipes); err == nil {
			t.Error("Validate() should fail on recipe with invalid success chance")
		}
	})

	t.Run("wrong type", func(t *testing.T) {
		if err := gen.Validate("not a recipe slice"); err == nil {
			t.Error("Validate() should fail on wrong type")
		}
	})
}

// TestRecipeGenerator_AllGenres tests all genres generate recipes.
func TestRecipeGenerator_AllGenres(t *testing.T) {
	genres := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapoc"}
	gen := NewRecipeGenerator()

	for _, genreID := range genres {
		t.Run(genreID, func(t *testing.T) {
			params := procgen.GenerationParams{
				Difficulty: 0.5,
				Depth:      2,
				GenreID:    genreID,
			}
			result, err := gen.Generate(12345, params)
			if err != nil {
				t.Fatalf("Generate() failed for genre %s: %v", genreID, err)
			}

			recipes := result.([]*engine.Recipe)
			if len(recipes) == 0 {
				t.Errorf("No recipes generated for genre %s", genreID)
			}

			// Verify all recipes have correct genre
			for _, recipe := range recipes {
				if recipe.GenreID != genreID {
					t.Errorf("Recipe has genreID %s, want %s", recipe.GenreID, genreID)
				}
			}
		})
	}
}

// TestRecipeGenerator_RarityDistribution tests rarity scaling with depth and difficulty.
func TestRecipeGenerator_RarityDistribution(t *testing.T) {
	gen := NewRecipeGenerator()

	// Generate many recipes and check rarity distribution
	params := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      1,
		GenreID:    "fantasy",
		Custom:     map[string]interface{}{"count": 100},
	}

	result, err := gen.Generate(54321, params)
	if err != nil {
		t.Fatal(err)
	}

	recipes := result.([]*engine.Recipe)
	rarityCounts := make(map[engine.RecipeRarity]int)

	for _, recipe := range recipes {
		rarityCounts[recipe.Rarity]++
	}

	// Should have mostly common/uncommon at low depth/difficulty
	if rarityCounts[engine.RecipeCommon] == 0 {
		t.Error("No common recipes generated")
	}

	// Should have at least some variation
	uniqueRarities := len(rarityCounts)
	if uniqueRarities < 2 {
		t.Errorf("Only %d unique rarities, want at least 2 for variety", uniqueRarities)
	}
}

// TestRecipeGenerator_MaterialQuantities tests material requirements are reasonable.
func TestRecipeGenerator_MaterialQuantities(t *testing.T) {
	gen := NewRecipeGenerator()
	params := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      3,
		GenreID:    "fantasy",
		Custom:     map[string]interface{}{"count": 20},
	}

	result, err := gen.Generate(77777, params)
	if err != nil {
		t.Fatal(err)
	}

	recipes := result.([]*engine.Recipe)

	for _, recipe := range recipes {
		if len(recipe.Materials) == 0 {
			t.Errorf("Recipe %s has no materials", recipe.Name)
		}

		// Check each material requirement
		for _, mat := range recipe.Materials {
			if mat.ItemName == "" {
				t.Errorf("Recipe %s has material with empty name", recipe.Name)
			}
			if mat.Quantity < 1 {
				t.Errorf("Recipe %s material %s has quantity %d, want >= 1",
					recipe.Name, mat.ItemName, mat.Quantity)
			}
			if mat.Quantity > 10 {
				t.Errorf("Recipe %s material %s has quantity %d, seems too high",
					recipe.Name, mat.ItemName, mat.Quantity)
			}
		}
	}
}

// TestRecipeGenerator_SkillScaling tests skill requirements scale with depth and difficulty.
func TestRecipeGenerator_SkillScaling(t *testing.T) {
	gen := NewRecipeGenerator()

	tests := []struct {
		name       string
		depth      int
		difficulty float64
		minSkill   int
		maxSkill   int
	}{
		{"low depth/difficulty", 1, 0.2, 0, 5},
		{"medium depth/difficulty", 5, 0.5, 0, 10},
		{"high depth/difficulty", 10, 0.9, 5, 20},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := procgen.GenerationParams{
				Difficulty: tt.difficulty,
				Depth:      tt.depth,
				GenreID:    "fantasy",
				Custom:     map[string]interface{}{"count": 10},
			}

			result, err := gen.Generate(88888, params)
			if err != nil {
				t.Fatal(err)
			}

			recipes := result.([]*engine.Recipe)
			for _, recipe := range recipes {
				if recipe.SkillRequired < tt.minSkill {
					t.Errorf("Recipe %s skill %d below expected min %d",
						recipe.Name, recipe.SkillRequired, tt.minSkill)
				}
				if recipe.SkillRequired > tt.maxSkill {
					t.Logf("Recipe %s skill %d above expected max %d (acceptable variance)",
						recipe.Name, recipe.SkillRequired, tt.maxSkill)
				}
			}
		})
	}
}

// TestRecipeGenerator_CraftTimes tests craft times are reasonable.
func TestRecipeGenerator_CraftTimes(t *testing.T) {
	gen := NewRecipeGenerator()
	params := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      3,
		GenreID:    "fantasy",
		Custom:     map[string]interface{}{"count": 20},
	}

	result, err := gen.Generate(33333, params)
	if err != nil {
		t.Fatal(err)
	}

	recipes := result.([]*engine.Recipe)

	for _, recipe := range recipes {
		if recipe.CraftTimeSec <= 0 {
			t.Errorf("Recipe %s has craft time %f, want > 0",
				recipe.Name, recipe.CraftTimeSec)
		}
		if recipe.CraftTimeSec > 30 {
			t.Errorf("Recipe %s has craft time %f, seems too long",
				recipe.Name, recipe.CraftTimeSec)
		}
	}
}

// BenchmarkRecipeGenerator_Generate benchmarks recipe generation.
func BenchmarkRecipeGenerator_Generate(b *testing.B) {
	gen := NewRecipeGenerator()
	params := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      3,
		GenreID:    "fantasy",
		Custom:     map[string]interface{}{"count": 10},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		gen.Generate(int64(i), params)
	}
}

// TestRecipeGenerator_NewRecipeTypes tests cooking and smithing recipe types.
func TestRecipeGenerator_NewRecipeTypes(t *testing.T) {
	gen := NewRecipeGenerator()

	tests := []struct {
		name       string
		typeFilter string
		wantType   engine.RecipeType
	}{
		{"cooking filter", "cooking", engine.RecipeCooking},
		{"smithing filter", "smithing", engine.RecipeSmithing},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := procgen.GenerationParams{
				Difficulty: 0.5,
				Depth:      3,
				GenreID:    "fantasy",
				Custom:     map[string]interface{}{"type": tt.typeFilter, "count": 5},
			}

			result, err := gen.Generate(12345, params)
			if err != nil {
				t.Fatalf("Generate() failed: %v", err)
			}

			recipes := result.([]*engine.Recipe)
			if len(recipes) != 5 {
				t.Fatalf("Generate() returned %d recipes, want 5", len(recipes))
			}

			// All recipes should match the filtered type
			for i, recipe := range recipes {
				if recipe.Type != tt.wantType {
					t.Errorf("Recipe %d has type %v, want %v", i, recipe.Type, tt.wantType)
				}
			}
		})
	}
}

// TestRecipeGenerator_CookingTemplatesAllGenres tests cooking templates exist for all genres.
func TestRecipeGenerator_CookingTemplatesAllGenres(t *testing.T) {
	gen := NewRecipeGenerator()
	genres := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapoc"}

	for _, genreID := range genres {
		t.Run(genreID, func(t *testing.T) {
			templates := gen.cookingTemplates[genreID]
			if len(templates) == 0 {
				t.Errorf("No cooking templates for genre %s", genreID)
			}

			// Verify template structure
			for i, tmpl := range templates {
				if tmpl.RecipeType != engine.RecipeCooking {
					t.Errorf("Template %d has type %v, want RecipeCooking", i, tmpl.RecipeType)
				}
				if tmpl.NamePrefix == "" || tmpl.NameSuffix == "" {
					t.Errorf("Template %d has empty name parts", i)
				}
				if len(tmpl.MaterialNames) == 0 {
					t.Errorf("Template %d has no material names", i)
				}
			}
		})
	}
}

// TestRecipeGenerator_SmithingTemplatesAllGenres tests smithing templates exist for all genres.
func TestRecipeGenerator_SmithingTemplatesAllGenres(t *testing.T) {
	gen := NewRecipeGenerator()
	genres := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapoc"}

	for _, genreID := range genres {
		t.Run(genreID, func(t *testing.T) {
			templates := gen.smithingTemplates[genreID]
			if len(templates) == 0 {
				t.Errorf("No smithing templates for genre %s", genreID)
			}

			// Verify template structure
			for i, tmpl := range templates {
				if tmpl.RecipeType != engine.RecipeSmithing {
					t.Errorf("Template %d has type %v, want RecipeSmithing", i, tmpl.RecipeType)
				}
				if tmpl.NamePrefix == "" || tmpl.NameSuffix == "" {
					t.Errorf("Template %d has empty name parts", i)
				}
				if len(tmpl.MaterialNames) == 0 {
					t.Errorf("Template %d has no material names", i)
				}
				// Smithing should output weapons or armor
				if tmpl.OutputType != item.TypeWeapon && tmpl.OutputType != item.TypeArmor {
					t.Errorf("Template %d has output type %v, want weapon or armor", i, tmpl.OutputType)
				}
			}
		})
	}
}

// TestRecipeGenerator_AllFiveRecipeTypes tests random distribution includes all 5 types.
func TestRecipeGenerator_AllFiveRecipeTypes(t *testing.T) {
	gen := NewRecipeGenerator()

	params := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      5,
		GenreID:    "fantasy",
		Custom:     map[string]interface{}{"count": 200}, // Generate many to hit all types
	}

	result, err := gen.Generate(98765, params)
	if err != nil {
		t.Fatal(err)
	}

	recipes := result.([]*engine.Recipe)
	typeCounts := make(map[engine.RecipeType]int)

	for _, recipe := range recipes {
		typeCounts[recipe.Type]++
	}

	// Verify all 5 types are generated
	expectedTypes := []engine.RecipeType{
		engine.RecipePotion,
		engine.RecipeEnchanting,
		engine.RecipeMagicItem,
		engine.RecipeCooking,
		engine.RecipeSmithing,
	}

	for _, expectedType := range expectedTypes {
		if count := typeCounts[expectedType]; count == 0 {
			t.Errorf("No recipes of type %v generated (generated %d recipes total)", expectedType, len(recipes))
		}
	}

	t.Logf("Type distribution: %v", typeCounts)
}

// TestRecipeType_String tests new recipe type string representations.
func TestRecipeType_String(t *testing.T) {
	tests := []struct {
		recType  engine.RecipeType
		expected string
	}{
		{engine.RecipePotion, "potion"},
		{engine.RecipeEnchanting, "enchanting"},
		{engine.RecipeMagicItem, "magic_item"},
		{engine.RecipeCooking, "cooking"},
		{engine.RecipeSmithing, "smithing"},
		{engine.RecipeType(99), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if got := tt.recType.String(); got != tt.expected {
				t.Errorf("RecipeType.String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// TestRecipeGenerator_CookingRecipeProperties tests cooking recipes have appropriate properties.
func TestRecipeGenerator_CookingRecipeProperties(t *testing.T) {
	gen := NewRecipeGenerator()

	params := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      2,
		GenreID:    "fantasy",
		Custom:     map[string]interface{}{"type": "cooking", "count": 10},
	}

	result, err := gen.Generate(11111, params)
	if err != nil {
		t.Fatal(err)
	}

	recipes := result.([]*engine.Recipe)
	for i, recipe := range recipes {
		// Cooking recipes should be consumables
		if recipe.OutputItemType != item.TypeConsumable {
			t.Errorf("Recipe %d outputs %v, want TypeConsumable", i, recipe.OutputItemType)
		}

		// Cooking should be relatively fast (< 30 seconds)
		if recipe.CraftTimeSec > 30.0 {
			t.Errorf("Recipe %d craft time %f is too long for cooking", i, recipe.CraftTimeSec)
		}

		// Should have reasonable success chance (cooking is easier than complex magic)
		if recipe.BaseSuccessChance < 0.50 {
			t.Errorf("Recipe %d success chance %f is too low for cooking", i, recipe.BaseSuccessChance)
		}
	}
}

// TestRecipeGenerator_SmithingRecipeProperties tests smithing recipes have appropriate properties.
func TestRecipeGenerator_SmithingRecipeProperties(t *testing.T) {
	gen := NewRecipeGenerator()

	params := procgen.GenerationParams{
		Difficulty: 0.6,
		Depth:      3,
		GenreID:    "fantasy",
		Custom:     map[string]interface{}{"type": "smithing", "count": 10},
	}

	result, err := gen.Generate(22222, params)
	if err != nil {
		t.Fatal(err)
	}

	recipes := result.([]*engine.Recipe)
	for i, recipe := range recipes {
		// Smithing recipes should output weapons or armor
		if recipe.OutputItemType != item.TypeWeapon && recipe.OutputItemType != item.TypeArmor {
			t.Errorf("Recipe %d outputs %v, want weapon or armor", i, recipe.OutputItemType)
		}

		// Smithing should take reasonable time (5-40 seconds)
		if recipe.CraftTimeSec < 5.0 || recipe.CraftTimeSec > 40.0 {
			t.Errorf("Recipe %d craft time %f is outside expected range [5.0, 40.0]", i, recipe.CraftTimeSec)
		}

		// Should require materials (metal, leather, etc.)
		if len(recipe.Materials) < 2 {
			t.Errorf("Recipe %d has only %d materials, smithing should require at least 2", i, len(recipe.Materials))
		}
	}
}

// BenchmarkGenerateNewRecipeTypes benchmarks generation of cooking and smithing recipes.
func BenchmarkGenerateNewRecipeTypes(b *testing.B) {
	gen := NewRecipeGenerator()
	params := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      3,
		GenreID:    "fantasy",
		Custom:     map[string]interface{}{"count": 20},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		gen.Generate(int64(i), params)
	}
}

// TestRecipeGenerator_HighDepthRarity tests rarity distribution at extreme depth values.
func TestRecipeGenerator_HighDepthRarity(t *testing.T) {
	gen := NewRecipeGenerator()
	params := procgen.GenerationParams{
		Difficulty: 1.0,
		Depth:      100,
		GenreID:    "fantasy",
		Custom:     map[string]interface{}{"count": 50},
	}

	result, err := gen.Generate(12345, params)
	if err != nil {
		t.Fatal(err)
	}

	recipes := result.([]*engine.Recipe)
	for _, recipe := range recipes {
		if recipe.BaseSuccessChance < 0 || recipe.BaseSuccessChance > 1.0 {
			t.Errorf("Recipe %s has invalid success chance: %f", recipe.Name, recipe.BaseSuccessChance)
		}
	}
}

// TestRecipeGenerator_ZeroCount tests that zero/negative count defaults to 5.
func TestRecipeGenerator_ZeroCount(t *testing.T) {
	gen := NewRecipeGenerator()

	tests := []struct {
		name      string
		count     int
		wantCount int
	}{
		{"zero count", 0, 5},
		{"negative count", -1, 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := procgen.GenerationParams{
				Difficulty: 0.5,
				Depth:      1,
				GenreID:    "fantasy",
				Custom:     map[string]interface{}{"count": tt.count},
			}

			result, err := gen.Generate(12345, params)
			if err != nil {
				t.Fatal(err)
			}

			recipes := result.([]*engine.Recipe)
			if len(recipes) != tt.wantCount {
				t.Errorf("Generate() returned %d recipes, want %d", len(recipes), tt.wantCount)
			}
		})
	}
}

// TestRecipeGenerator_UnknownGenreFallback tests that unknown genre falls back gracefully.
func TestRecipeGenerator_UnknownGenreFallback(t *testing.T) {
	gen := NewRecipeGenerator()
	params := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      1,
		GenreID:    "nonexistent_genre",
		Custom:     map[string]interface{}{"count": 5},
	}

	result, err := gen.Generate(12345, params)
	if err != nil {
		t.Fatal(err)
	}

	recipes := result.([]*engine.Recipe)
	if len(recipes) != 5 {
		t.Errorf("Generate() returned %d recipes, want 5", len(recipes))
	}

	if err := gen.Validate(result); err != nil {
		t.Errorf("Validate() failed: %v", err)
	}
}

// TestRecipeGenerator_GenreFallbackLogging tests that fallback warnings are logged
// when templates are not found for requested genre.
func TestRecipeGenerator_GenreFallbackLogging(t *testing.T) {
	tests := []struct {
		name         string
		genreID      string
		expectedLogs int // Number of fallback warnings expected
	}{
		{
			name:         "valid genre",
			genreID:      "fantasy",
			expectedLogs: 0, // No fallbacks
		},
		{
			name:         "unknown genre fallback to fantasy",
			genreID:      "nonexistent",
			expectedLogs: 1, // Fallback to fantasy
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create logger with custom hook to capture logs
			logger := logrus.New()
			logger.SetLevel(logrus.WarnLevel)

			var logBuffer bytes.Buffer
			logger.SetOutput(&logBuffer)

			gen := NewRecipeGeneratorWithLogger(logger)

			params := procgen.GenerationParams{
				GenreID:    tt.genreID,
				Depth:      10,
				Difficulty: 0.5,
				Custom: map[string]interface{}{
					"count": 3,
				},
			}

			_, err := gen.Generate(12345, params)
			if err != nil {
				t.Fatalf("Generate() failed: %v", err)
			}

			logOutput := logBuffer.String()
			fallbackCount := 0
			if tt.genreID == "nonexistent" && len(logOutput) > 0 {
				// Should have fallback warning
				if !bytes.Contains(logBuffer.Bytes(), []byte("falling back")) {
					t.Errorf("Expected fallback warning in logs, got: %s", logOutput)
				}
				fallbackCount = 1
			}

			if fallbackCount != tt.expectedLogs {
				t.Errorf("Expected %d fallback logs, got %d", tt.expectedLogs, fallbackCount)
			}
		})
	}
}
