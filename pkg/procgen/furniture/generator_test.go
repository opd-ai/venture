package furniture

import (
	"testing"

	"github.com/opd-ai/venture/pkg/procgen"
)

func TestFurnitureTypes(t *testing.T) {
	tests := []struct {
		furnitureType FurnitureType
		want          string
	}{
		{TypeSeating, "Seating"},
		{TypeStorage, "Storage"},
		{TypeCrafting, "Crafting"},
		{TypeDecoration, "Decoration"},
		{TypeLighting, "Lighting"},
		{TypeBedding, "Bedding"},
		{TypeTable, "Table"},
		{TypeUtility, "Utility"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.furnitureType.String(); got != tt.want {
				t.Errorf("FurnitureType.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMaterialTypes(t *testing.T) {
	tests := []struct {
		material MaterialType
		want     string
	}{
		{MaterialWood, "Wood"},
		{MaterialMetal, "Metal"},
		{MaterialStone, "Stone"},
		{MaterialCrystal, "Crystal"},
		{MaterialFabric, "Fabric"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.material.String(); got != tt.want {
				t.Errorf("MaterialType.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRarityTiers(t *testing.T) {
	tests := []struct {
		rarity         RarityTier
		wantName       string
		wantMultiplier float64
	}{
		{RarityCommon, "Common", 1.0},
		{RarityUncommon, "Uncommon", 1.2},
		{RarityRare, "Rare", 1.5},
		{RarityEpic, "Epic", 2.0},
		{RarityLegendary, "Legendary", 3.0},
	}

	for _, tt := range tests {
		t.Run(tt.wantName, func(t *testing.T) {
			if got := tt.rarity.String(); got != tt.wantName {
				t.Errorf("RarityTier.String() = %v, want %v", got, tt.wantName)
			}
			if got := tt.rarity.DetailMultiplier(); got != tt.wantMultiplier {
				t.Errorf("RarityTier.DetailMultiplier() = %v, want %v", got, tt.wantMultiplier)
			}
		})
	}
}

func TestDirections(t *testing.T) {
	tests := []struct {
		direction Direction
		want      string
		wantSteps int
	}{
		{DirNorth, "North", 0},
		{DirEast, "East", 1},
		{DirSouth, "South", 2},
		{DirWest, "West", 3},
		{DirNorthEast, "NorthEast", 4},
		{DirSouthEast, "SouthEast", 5},
		{DirSouthWest, "SouthWest", 6},
		{DirNorthWest, "NorthWest", 7},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.direction.String(); got != tt.want {
				t.Errorf("Direction.String() = %v, want %v", got, tt.want)
			}
			if got := tt.direction.RotationSteps(); got != tt.wantSteps {
				t.Errorf("Direction.RotationSteps() = %v, want %v", got, tt.wantSteps)
			}
		})
	}
}

func TestGetTemplate(t *testing.T) {
	tests := []struct {
		subType  string
		wantType FurnitureType
	}{
		{"Chair", TypeSeating},
		{"Chest", TypeStorage},
		{"Anvil", TypeCrafting},
		{"Statue", TypeDecoration},
		{"Torch", TypeLighting},
		{"Bed", TypeBedding},
		{"Table", TypeTable},
		{"Fireplace", TypeUtility},
	}

	for _, tt := range tests {
		t.Run(tt.subType, func(t *testing.T) {
			tmpl := GetTemplate(tt.subType)
			if tmpl == nil {
				t.Fatalf("GetTemplate(%q) returned nil", tt.subType)
			}
			if tmpl.Type != tt.wantType {
				t.Errorf("GetTemplate(%q).Type = %v, want %v", tt.subType, tmpl.Type, tt.wantType)
			}
			if tmpl.SubType != tt.subType {
				t.Errorf("GetTemplate(%q).SubType = %v, want %v", tt.subType, tmpl.SubType, tt.subType)
			}
		})
	}
}

func TestGetTemplateInvalid(t *testing.T) {
	tmpl := GetTemplate("InvalidFurniture")
	if tmpl != nil {
		t.Errorf("GetTemplate(invalid) returned non-nil: %+v", tmpl)
	}
}

func TestGetAllSubTypes(t *testing.T) {
	subtypes := GetAllSubTypes()
	if len(subtypes) < 30 {
		t.Errorf("GetAllSubTypes() returned %d subtypes, want at least 30", len(subtypes))
	}

	// Check some expected types
	expected := []string{"Chair", "Chest", "Anvil", "Statue", "Torch", "Bed", "Table", "Fireplace"}
	for _, exp := range expected {
		found := false
		for _, st := range subtypes {
			if st == exp {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("GetAllSubTypes() missing expected type: %s", exp)
		}
	}
}

func TestGetSubTypesByCategory(t *testing.T) {
	tests := []struct {
		category    FurnitureType
		minExpected int
	}{
		{TypeSeating, 5},
		{TypeStorage, 6},
		{TypeCrafting, 5},
		{TypeDecoration, 5},
		{TypeLighting, 4},
		{TypeBedding, 3},
		{TypeTable, 4},
		{TypeUtility, 4},
	}

	for _, tt := range tests {
		t.Run(tt.category.String(), func(t *testing.T) {
			subtypes := GetSubTypesByCategory(tt.category)
			if len(subtypes) < tt.minExpected {
				t.Errorf("GetSubTypesByCategory(%v) = %d types, want at least %d", tt.category, len(subtypes), tt.minExpected)
			}
		})
	}
}

func TestGenerateBasic(t *testing.T) {
	gen := NewGenerator()
	params := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      5,
		GenreID:    "fantasy",
		Custom: map[string]interface{}{
			"SubType": "Chair",
		},
	}

	result, err := gen.Generate(12345, params)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	furniture, ok := result.(*Furniture)
	if !ok {
		t.Fatalf("Generate() returned non-Furniture: %T", result)
	}

	if furniture.SubType != "Chair" {
		t.Errorf("Generate() SubType = %v, want Chair", furniture.SubType)
	}

	if furniture.Width <= 0 || furniture.Height <= 0 || furniture.Depth <= 0 {
		t.Errorf("Generate() invalid dimensions: %.2f×%.2f×%.2f", furniture.Width, furniture.Height, furniture.Depth)
	}

	if furniture.Name == "" {
		t.Error("Generate() Name is empty")
	}

	if furniture.Description == "" {
		t.Error("Generate() Description is empty")
	}
}

func TestGenerateDeterminism(t *testing.T) {
	gen := NewGenerator()
	params := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      10,
		GenreID:    "scifi",
		Custom: map[string]interface{}{
			"SubType": "Chest",
		},
	}

	seed := int64(99999)

	// Generate twice with same seed
	result1, err1 := gen.Generate(seed, params)
	if err1 != nil {
		t.Fatalf("Generate() #1 error = %v", err1)
	}

	result2, err2 := gen.Generate(seed, params)
	if err2 != nil {
		t.Fatalf("Generate() #2 error = %v", err2)
	}

	f1 := result1.(*Furniture)
	f2 := result2.(*Furniture)

	// Check key properties match
	if f1.Material != f2.Material {
		t.Errorf("Determinism failed: Material %v != %v", f1.Material, f2.Material)
	}
	if f1.Rarity != f2.Rarity {
		t.Errorf("Determinism failed: Rarity %v != %v", f1.Rarity, f2.Rarity)
	}
	if f1.Width != f2.Width || f1.Height != f2.Height || f1.Depth != f2.Depth {
		t.Errorf("Determinism failed: Dimensions (%.2f,%.2f,%.2f) != (%.2f,%.2f,%.2f)",
			f1.Width, f1.Height, f1.Depth, f2.Width, f2.Height, f2.Depth)
	}
	if f1.Name != f2.Name {
		t.Errorf("Determinism failed: Name %q != %q", f1.Name, f2.Name)
	}
}

func TestGenerateAllSubTypes(t *testing.T) {
	gen := NewGenerator()
	subtypes := GetAllSubTypes()

	for i, subtype := range subtypes {
		t.Run(subtype, func(t *testing.T) {
			params := procgen.GenerationParams{
				Difficulty: 0.5,
				Depth:      5,
				GenreID:    "fantasy",
				Custom: map[string]interface{}{
					"SubType": subtype,
				},
			}

			result, err := gen.Generate(int64(i)+10000, params)
			if err != nil {
				t.Fatalf("Generate(%s) error = %v", subtype, err)
			}

			furniture := result.(*Furniture)

			// Validate
			if err := gen.Validate(furniture); err != nil {
				t.Errorf("Validate(%s) error = %v", subtype, err)
			}
		})
	}
}

func TestGenerateRarityDistribution(t *testing.T) {
	gen := NewGenerator()

	tests := []struct {
		difficulty float64
		depth      int
		minRare    float64 // Minimum percentage of rare+ items
	}{
		{0.0, 1, 0.0},  // Low difficulty, low depth -> mostly common
		{0.5, 10, 0.2}, // Medium difficulty, medium depth -> some rare
		{1.0, 50, 0.5}, // High difficulty, high depth -> many rare
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			params := procgen.GenerationParams{
				Difficulty: tt.difficulty,
				Depth:      tt.depth,
				GenreID:    "fantasy",
				Custom: map[string]interface{}{
					"SubType": "Chair",
				},
			}

			rareCount := 0
			totalCount := 100

			for i := 0; i < totalCount; i++ {
				result, _ := gen.Generate(int64(i)+20000, params)
				furniture := result.(*Furniture)

				if furniture.Rarity >= RarityRare {
					rareCount++
				}
			}

			rarePct := float64(rareCount) / float64(totalCount)
			if rarePct < tt.minRare {
				t.Errorf("Rare percentage %.2f < minimum %.2f (difficulty=%.1f, depth=%d)",
					rarePct, tt.minRare, tt.difficulty, tt.depth)
			}
		})
	}
}

func TestValidate(t *testing.T) {
	gen := NewGenerator()

	// Valid furniture
	params := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      5,
		GenreID:    "fantasy",
		Custom: map[string]interface{}{
			"SubType": "Table",
		},
	}

	result, _ := gen.Generate(30000, params)
	furniture := result.(*Furniture)

	if err := gen.Validate(furniture); err != nil {
		t.Errorf("Validate() error for valid furniture: %v", err)
	}

	// Invalid dimensions
	badFurniture := *furniture
	badFurniture.Width = -1.0
	if err := gen.Validate(&badFurniture); err == nil {
		t.Error("Validate() should fail for negative width")
	}

	// Invalid detail level
	badFurniture2 := *furniture
	badFurniture2.DetailLevel = 15.0
	if err := gen.Validate(&badFurniture2); err == nil {
		t.Error("Validate() should fail for excessive detail level")
	}

	// Invalid light intensity
	badFurniture3 := *furniture
	badFurniture3.LightIntensity = 2.0
	if err := gen.Validate(&badFurniture3); err == nil {
		t.Error("Validate() should fail for light intensity > 1.0")
	}
}

func BenchmarkGenerate(b *testing.B) {
	gen := NewGenerator()
	params := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      10,
		GenreID:    "fantasy",
		Custom: map[string]interface{}{
			"SubType": "Chair",
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		gen.Generate(int64(i), params)
	}
}

func BenchmarkValidate(b *testing.B) {
	gen := NewGenerator()
	params := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      10,
		GenreID:    "fantasy",
		Custom: map[string]interface{}{
			"SubType": "Chest",
		},
	}

	result, _ := gen.Generate(40000, params)
	furniture := result.(*Furniture)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		gen.Validate(furniture)
	}
}
