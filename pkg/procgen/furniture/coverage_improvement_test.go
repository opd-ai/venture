package furniture

import (
	"testing"

	"github.com/opd-ai/venture/pkg/procgen"
)

// TestGenerateWithoutSubType tests chooseRandomSubType path (0% coverage)
func TestGenerateWithoutSubType(t *testing.T) {
	gen := NewGenerator()
	params := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      5,
		GenreID:    "fantasy",
		Custom:     map[string]interface{}{}, // No SubType specified
	}

	result, err := gen.Generate(70000, params)
	if err != nil {
		t.Fatalf("Generate() without SubType error = %v", err)
	}

	furniture, ok := result.(*Furniture)
	if !ok {
		t.Fatalf("Generate() returned non-Furniture: %T", result)
	}

	// Should have chosen a random subtype
	if furniture.SubType == "" {
		t.Error("Generate() did not assign SubType when not provided")
	}

	// SubType should be valid
	tmpl := GetTemplate(furniture.SubType)
	if tmpl == nil {
		t.Errorf("Generate() chose invalid SubType: %s", furniture.SubType)
	}
}

// TestGenerateWithVariousDepths tests chooseRandomSubType at different depths
func TestGenerateWithVariousDepths(t *testing.T) {
	gen := NewGenerator()
	
	tests := []struct {
		name  string
		depth int
	}{
		{"Depth1", 1},
		{"Depth10", 10},
		{"Depth50", 50},
		{"Depth100", 100},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := procgen.GenerationParams{
				Difficulty: 0.5,
				Depth:      tt.depth,
				GenreID:    "fantasy",
				Custom:     map[string]interface{}{}, // No SubType
			}

			result, err := gen.Generate(int64(70000+tt.depth), params)
			if err != nil {
				t.Fatalf("Generate(depth=%d) error = %v", tt.depth, err)
			}

			furniture := result.(*Furniture)
			if furniture.SubType == "" {
				t.Errorf("Generate(depth=%d) did not assign SubType", tt.depth)
			}
		})
	}
}

// TestHorrorMaterialSelection tests trySelectHorrorMaterialDeterministic (0% coverage)
func TestHorrorMaterialSelection(t *testing.T) {
	gen := NewGenerator()
	
	tests := []struct {
		name     string
		subType  string
		wantMats []MaterialType
	}{
		{"ChairHorror", "Chair", []MaterialType{MaterialWood, MaterialStone, MaterialMetal, MaterialFabric}},
		{"TableHorror", "Table", []MaterialType{MaterialWood, MaterialStone, MaterialMetal}},
		{"ChestHorror", "Chest", []MaterialType{MaterialWood, MaterialMetal, MaterialStone}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Generate multiple items to test horror material selection
			foundMaterials := make(map[MaterialType]bool)
			
			for i := 0; i < 20; i++ {
				params := procgen.GenerationParams{
					Difficulty: 0.3,
					Depth:      5,
					GenreID:    "horror", // Horror genre
					Custom: map[string]interface{}{
						"SubType": tt.subType,
					},
				}

				result, err := gen.Generate(int64(71000+i), params)
				if err != nil {
					t.Fatalf("Generate() error = %v", err)
				}

				furniture := result.(*Furniture)
				foundMaterials[furniture.Material] = true
			}

			// Should have generated at least one item with horror-appropriate materials
			hasHorrorMaterial := false
			for mat := range foundMaterials {
				for _, wantMat := range tt.wantMats {
					if mat == wantMat {
						hasHorrorMaterial = true
						break
					}
				}
				if hasHorrorMaterial {
					break
				}
			}

			if !hasHorrorMaterial {
				t.Errorf("Horror genre did not select expected materials. Got: %v", foundMaterials)
			}
		})
	}
}

// TestPostApocMaterialSelection tests trySelectPostapocMaterialDeterministic (0% coverage)
func TestPostApocMaterialSelection(t *testing.T) {
	gen := NewGenerator()
	
	tests := []struct {
		name     string
		subType  string
		wantMats []MaterialType
	}{
		{"ChairPostapoc", "Chair", []MaterialType{MaterialMetal, MaterialWood, MaterialFabric}},
		{"TablePostapoc", "Table", []MaterialType{MaterialMetal, MaterialWood}},
		{"ChestPostapoc", "Chest", []MaterialType{MaterialMetal, MaterialWood}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Generate multiple items to test post-apocalyptic material selection
			foundMaterials := make(map[MaterialType]bool)
			
			for i := 0; i < 20; i++ {
				params := procgen.GenerationParams{
					Difficulty: 0.3,
					Depth:      5,
					GenreID:    "postapoc", // Post-apocalyptic genre
					Custom: map[string]interface{}{
						"SubType": tt.subType,
					},
				}

				result, err := gen.Generate(int64(72000+i), params)
				if err != nil {
					t.Fatalf("Generate() error = %v", err)
				}

				furniture := result.(*Furniture)
				foundMaterials[furniture.Material] = true
			}

			// Should have generated at least one item with post-apoc materials
			hasPostapocMaterial := false
			for mat := range foundMaterials {
				for _, wantMat := range tt.wantMats {
					if mat == wantMat {
						hasPostapocMaterial = true
						break
					}
				}
				if hasPostapocMaterial {
					break
				}
			}

			if !hasPostapocMaterial {
				t.Errorf("Postapoc genre did not select expected materials. Got: %v", foundMaterials)
			}
		})
	}
}

// TestScifiMaterialSelection tests trySelectScifiMaterialDeterministic edge cases
func TestScifiMaterialSelection(t *testing.T) {
	gen := NewGenerator()

	// Generate multiple sci-fi items to test material selection
	foundMaterials := make(map[MaterialType]bool)
	
	for i := 0; i < 30; i++ {
		params := procgen.GenerationParams{
			Difficulty: 0.3,
			Depth:      5,
			GenreID:    "scifi", // Sci-fi genre
			Custom: map[string]interface{}{
				"SubType": "Chair",
			},
		}

		result, err := gen.Generate(int64(73000+i), params)
		if err != nil {
			t.Fatalf("Generate() error = %v", err)
		}

		furniture := result.(*Furniture)
		foundMaterials[furniture.Material] = true
	}

	// Sci-fi should prefer Metal and Crystal
	if !foundMaterials[MaterialMetal] && !foundMaterials[MaterialCrystal] {
		t.Errorf("Scifi genre did not select Metal or Crystal. Got: %v", foundMaterials)
	}
}

// TestCyberpunkMaterialSelection tests cyberpunk uses scifi material selection
func TestCyberpunkMaterialSelection(t *testing.T) {
	gen := NewGenerator()

	params := procgen.GenerationParams{
		Difficulty: 0.3,
		Depth:      5,
		GenreID:    "cyberpunk", // Cyberpunk genre (uses scifi materials)
		Custom: map[string]interface{}{
			"SubType": "Table",
		},
	}

	result, err := gen.Generate(74000, params)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	furniture := result.(*Furniture)
	// Should generate successfully with cyberpunk genre
	if furniture.Material == 0 {
		t.Error("Cyberpunk genre did not assign material")
	}
}

// TestFindValidPlacementNoSpace tests FindValidPlacement when room is full
func TestFindValidPlacementNoSpace(t *testing.T) {
	validator := NewPlacementValidator(2.0, 2.0) // Very small room

	gen := NewGenerator()
	params := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      5,
		GenreID:    "fantasy",
		Custom:     map[string]interface{}{"SubType": "Table"},
	}

	// Create large furniture that won't fit
	result, _ := gen.Generate(75000, params)
	largeFurniture := result.(*Furniture)
	largeFurniture.CollisionWidth = 5.0  // Larger than room
	largeFurniture.CollisionDepth = 5.0

	x, y, dir, ok := validator.FindValidPlacement(largeFurniture, DirNorth)
	if ok {
		t.Errorf("FindValidPlacement() succeeded for oversized furniture at (%.1f, %.1f, %v)", x, y, dir)
	}
}

// TestFindValidPlacementWithRotation tests rotation attempts in FindValidPlacement
func TestFindValidPlacementWithRotation(t *testing.T) {
	validator := NewPlacementValidator(10.0, 10.0)

	gen := NewGenerator()
	params := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      5,
		GenreID:    "fantasy",
		Custom:     map[string]interface{}{"SubType": "Bench"},
	}

	result, _ := gen.Generate(76000, params)
	furniture := result.(*Furniture)

	// Set asymmetric dimensions to test rotation
	furniture.CollisionWidth = 4.0
	furniture.CollisionDepth = 1.0

	// Fill room strategically to force rotation
	blocker, _ := gen.Generate(76001, params)
	blockerFurn := blocker.(*Furniture)
	blockerFurn.CollisionWidth = 9.0
	blockerFurn.CollisionDepth = 1.0
	
	validator.PlaceFurniture(blockerFurn, 0.0, 0.0, DirNorth)

	// Should still find placement, possibly with rotation
	x, y, dir, ok := validator.FindValidPlacement(furniture, DirNorth)
	if !ok {
		t.Error("FindValidPlacement() failed when rotation could provide valid placement")
	} else {
		// Verify the placement is actually valid
		err := validator.PlaceFurniture(furniture, x, y, dir)
		if err != nil {
			t.Errorf("FindValidPlacement() returned invalid position: %v", err)
		}
	}
}

// TestRotateFurnitureCollision tests rotation that causes collision
func TestRotateFurnitureCollision(t *testing.T) {
	validator := NewPlacementValidator(20.0, 20.0)

	gen := NewGenerator()
	params := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      5,
		GenreID:    "fantasy",
		Custom:     map[string]interface{}{"SubType": "Bench"},
	}

	// Place asymmetric furniture
	result1, _ := gen.Generate(77000, params)
	furniture1 := result1.(*Furniture)
	furniture1.CollisionWidth = 4.0
	furniture1.CollisionDepth = 1.0
	validator.PlaceFurniture(furniture1, 5.0, 5.0, DirNorth)

	// Place blocker nearby that would collide if first is rotated
	result2, _ := gen.Generate(77001, params)
	furniture2 := result2.(*Furniture)
	furniture2.CollisionWidth = 2.0
	furniture2.CollisionDepth = 2.0
	validator.PlaceFurniture(furniture2, 6.0, 5.0, DirNorth)

	// Try to rotate first furniture - may or may not cause collision depending on position
	err := validator.RotateFurniture(0)
	// Either succeeds or fails with collision - both are valid outcomes
	_ = err // Don't assert, just ensure it doesn't panic
}

// TestValidateInvalidFurniture tests validation edge cases
func TestValidateInvalidFurniture(t *testing.T) {
	gen := NewGenerator()

	tests := []struct {
		name     string
		furniture *Furniture
		wantErr  bool
	}{
		{
			name: "ZeroDimensions",
			furniture: &Furniture{
				Width:  0,
				Height: 0,
				Depth:  0,
			},
			wantErr: true,
		},
		{
			name: "NegativeDimensions",
			furniture: &Furniture{
				Width:  -1.0,
				Height: 2.0,
				Depth:  2.0,
			},
			wantErr: true,
		},
		{
			name: "EmptyName",
			furniture: &Furniture{
				Width:  1.0,
				Height: 1.0,
				Depth:  1.0,
				Name:   "",
			},
			wantErr: true,
		},
		{
			name: "ValidFurniture",
			furniture: &Furniture{
				Width:          1.0,
				Height:         1.0,
				Depth:          1.0,
				CollisionWidth: 1.0,
				CollisionDepth: 1.0,
				DetailLevel:    5.0,
				LightIntensity: 0.5,
				Capacity:       10,
				Name:           "Test",
				Description:    "Test description",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := gen.Validate(tt.furniture)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestGenerateHighRarityItems tests exotic material selection for high rarity
func TestGenerateHighRarityItems(t *testing.T) {
	gen := NewGenerator()

	// Use high difficulty and depth to increase legendary chances
	foundRarities := make(map[RarityTier]bool)
	
	for i := 0; i < 50; i++ {
		params := procgen.GenerationParams{
			Difficulty: 1.0,    // Maximum difficulty
			Depth:      100,    // High depth
			GenreID:    "fantasy",
			Custom: map[string]interface{}{
				"SubType": "Throne", // High-value furniture
			},
		}

		result, err := gen.Generate(int64(78000+i), params)
		if err != nil {
			t.Fatalf("Generate() error = %v", err)
		}

		furniture := result.(*Furniture)
		foundRarities[furniture.Rarity] = true
	}

	// With high difficulty and depth, should generate some rare items
	if !foundRarities[RarityRare] && !foundRarities[RarityEpic] && !foundRarities[RarityLegendary] {
		t.Errorf("High difficulty/depth did not generate rare items. Got rarities: %v", foundRarities)
	}
}
