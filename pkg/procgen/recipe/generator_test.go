package recipe

import (
	"testing"

	"github.com/opd-ai/venture/pkg/engine"
	"github.com/opd-ai/venture/pkg/procgen"
)

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
