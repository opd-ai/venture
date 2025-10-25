// Package environment provides procedural generation of environmental objects.
package environment

import (
	"testing"
)

// TestObjectType_String tests ObjectType string conversion.
func TestObjectType_String(t *testing.T) {
	tests := []struct {
		name     string
		objType  ObjectType
		expected string
	}{
		{"furniture", ObjectFurniture, "Furniture"},
		{"decoration", ObjectDecoration, "Decoration"},
		{"obstacle", ObjectObstacle, "Obstacle"},
		{"hazard", ObjectHazard, "Hazard"},
		{"unknown", ObjectType(99), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.objType.String(); got != tt.expected {
				t.Errorf("ObjectType.String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// TestSubType_String tests SubType string conversion.
func TestSubType_String(t *testing.T) {
	tests := []struct {
		name     string
		subType  SubType
		expected string
	}{
		// Furniture
		{"table", SubTypeTable, "Table"},
		{"chair", SubTypeChair, "Chair"},
		{"chest", SubTypeChest, "Chest"},

		// Decorations
		{"plant", SubTypePlant, "Plant"},
		{"statue", SubTypeStatue, "Statue"},
		{"torch", SubTypeTorch, "Torch"},

		// Obstacles
		{"barrel", SubTypeBarrel, "Barrel"},
		{"pillar", SubTypePillar, "Pillar"},
		{"boulder", SubTypeBoulder, "Boulder"},

		// Hazards
		{"spikes", SubTypeSpikes, "Spikes"},
		{"fire_pit", SubTypeFirePit, "FirePit"},
		{"acid_pool", SubTypeAcidPool, "AcidPool"},

		{"unknown", SubType(999), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.subType.String(); got != tt.expected {
				t.Errorf("SubType.String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// TestSubType_GetObjectType tests subtype to object type conversion.
func TestSubType_GetObjectType(t *testing.T) {
	tests := []struct {
		name     string
		subType  SubType
		expected ObjectType
	}{
		{"table_is_furniture", SubTypeTable, ObjectFurniture},
		{"chair_is_furniture", SubTypeChair, ObjectFurniture},
		{"plant_is_decoration", SubTypePlant, ObjectDecoration},
		{"statue_is_decoration", SubTypeStatue, ObjectDecoration},
		{"barrel_is_obstacle", SubTypeBarrel, ObjectObstacle},
		{"pillar_is_obstacle", SubTypePillar, ObjectObstacle},
		{"spikes_is_hazard", SubTypeSpikes, ObjectHazard},
		{"fire_pit_is_hazard", SubTypeFirePit, ObjectHazard},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.subType.GetObjectType(); got != tt.expected {
				t.Errorf("SubType.GetObjectType() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// TestGetProperties tests default property retrieval.
func TestGetProperties(t *testing.T) {
	tests := []struct {
		name               string
		subType            SubType
		expectCollidable   bool
		expectInteractable bool
		expectHarmful      bool
		expectDamage       int
	}{
		// Furniture: collidable, interactable, not harmful
		{"table", SubTypeTable, true, true, false, 0},
		{"chest", SubTypeChest, true, true, false, 0},

		// Decorations: varies
		{"plant", SubTypePlant, false, false, false, 0},
		{"torch", SubTypeTorch, false, true, false, 0},
		{"statue", SubTypeStatue, false, false, false, 0},

		// Obstacles: collidable, not interactable, not harmful
		{"barrel", SubTypeBarrel, true, false, false, 0},
		{"pillar", SubTypePillar, true, false, false, 0},

		// Hazards: harmful, varies on collidable
		{"spikes", SubTypeSpikes, true, false, true, 10},
		{"fire_pit", SubTypeFirePit, true, false, true, 10},
		{"poison_gas", SubTypePoisonGas, false, false, true, 5},
		{"electric_field", SubTypeElectricField, false, false, true, 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			collidable, interactable, harmful, damage := GetProperties(tt.subType)
			if collidable != tt.expectCollidable {
				t.Errorf("GetProperties(%v) collidable = %v, want %v", tt.subType, collidable, tt.expectCollidable)
			}
			if interactable != tt.expectInteractable {
				t.Errorf("GetProperties(%v) interactable = %v, want %v", tt.subType, interactable, tt.expectInteractable)
			}
			if harmful != tt.expectHarmful {
				t.Errorf("GetProperties(%v) harmful = %v, want %v", tt.subType, harmful, tt.expectHarmful)
			}
			if damage != tt.expectDamage {
				t.Errorf("GetProperties(%v) damage = %v, want %v", tt.subType, damage, tt.expectDamage)
			}
		})
	}
}

// TestDefaultConfig tests default configuration.
func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	if config.SubType != SubTypeTable {
		t.Errorf("DefaultConfig SubType = %v, want %v", config.SubType, SubTypeTable)
	}
	if config.Width != 32 {
		t.Errorf("DefaultConfig Width = %v, want 32", config.Width)
	}
	if config.Height != 32 {
		t.Errorf("DefaultConfig Height = %v, want 32", config.Height)
	}
	if config.GenreID != "fantasy" {
		t.Errorf("DefaultConfig GenreID = %v, want fantasy", config.GenreID)
	}
	if config.Custom == nil {
		t.Error("DefaultConfig Custom map is nil")
	}
}

// TestConfig_Validate tests configuration validation.
func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: Config{
				SubType: SubTypeTable,
				Width:   32,
				Height:  32,
				GenreID: "fantasy",
			},
			wantErr: false,
		},
		{
			name: "zero width",
			config: Config{
				SubType: SubTypeTable,
				Width:   0,
				Height:  32,
				GenreID: "fantasy",
			},
			wantErr: true,
		},
		{
			name: "negative width",
			config: Config{
				SubType: SubTypeTable,
				Width:   -10,
				Height:  32,
				GenreID: "fantasy",
			},
			wantErr: true,
		},
		{
			name: "zero height",
			config: Config{
				SubType: SubTypeTable,
				Width:   32,
				Height:  0,
				GenreID: "fantasy",
			},
			wantErr: true,
		},
		{
			name: "empty genre",
			config: Config{
				SubType: SubTypeTable,
				Width:   32,
				Height:  32,
				GenreID: "",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Config.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestGenerator_Generate tests object generation.
func TestGenerator_Generate(t *testing.T) {
	gen := NewGenerator()

	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{
			name: "generate table",
			config: Config{
				SubType: SubTypeTable,
				Width:   32,
				Height:  32,
				GenreID: "fantasy",
				Seed:    12345,
			},
			wantErr: false,
		},
		{
			name: "generate chest",
			config: Config{
				SubType: SubTypeChest,
				Width:   32,
				Height:  32,
				GenreID: "fantasy",
				Seed:    12345,
			},
			wantErr: false,
		},
		{
			name: "generate plant",
			config: Config{
				SubType: SubTypePlant,
				Width:   32,
				Height:  32,
				GenreID: "scifi",
				Seed:    12345,
			},
			wantErr: false,
		},
		{
			name: "generate spikes",
			config: Config{
				SubType: SubTypeSpikes,
				Width:   32,
				Height:  32,
				GenreID: "horror",
				Seed:    12345,
			},
			wantErr: false,
		},
		{
			name: "invalid config",
			config: Config{
				SubType: SubTypeTable,
				Width:   0,
				Height:  32,
				GenreID: "fantasy",
				Seed:    12345,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			obj, err := gen.Generate(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("Generator.Generate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if obj == nil {
					t.Error("Generator.Generate() returned nil object")
					return
				}
				if obj.Sprite == nil {
					t.Error("Generated object has nil sprite")
				}
				if obj.Width != tt.config.Width {
					t.Errorf("Generated object Width = %v, want %v", obj.Width, tt.config.Width)
				}
				if obj.Height != tt.config.Height {
					t.Errorf("Generated object Height = %v, want %v", obj.Height, tt.config.Height)
				}
				if obj.GenreID != tt.config.GenreID {
					t.Errorf("Generated object GenreID = %v, want %v", obj.GenreID, tt.config.GenreID)
				}
				if obj.Name == "" {
					t.Error("Generated object has empty name")
				}
			}
		})
	}
}

// TestGenerator_GenerateAllSubTypes tests generation of all subtypes.
func TestGenerator_GenerateAllSubTypes(t *testing.T) {
	gen := NewGenerator()

	subtypes := []SubType{
		// Furniture
		SubTypeTable, SubTypeChair, SubTypeBed, SubTypeShelf, SubTypeChest,
		SubTypeDesk, SubTypeBench, SubTypeCabinet,
		// Decorations
		SubTypePlant, SubTypeStatue, SubTypePainting, SubTypeBanner,
		SubTypeTorch, SubTypeCandlestick, SubTypeVase, SubTypeTapestry,
		SubTypeCrystal, SubTypeBook,
		// Obstacles
		SubTypeBarrel, SubTypeCrate, SubTypeRubble, SubTypePillar,
		SubTypeBoulder, SubTypeDebris, SubTypeWreckage, SubTypeColumn,
		// Hazards
		SubTypeSpikes, SubTypeFirePit, SubTypeAcidPool, SubTypeBearTrap,
		SubTypePoisonGas, SubTypeLavaPit, SubTypeElectricField, SubTypeIceField,
	}

	for _, subType := range subtypes {
		t.Run(subType.String(), func(t *testing.T) {
			config := Config{
				SubType: subType,
				Width:   32,
				Height:  32,
				GenreID: "fantasy",
				Seed:    12345,
			}

			obj, err := gen.Generate(config)
			if err != nil {
				t.Errorf("Generator.Generate(%v) error = %v", subType, err)
				return
			}
			if obj == nil {
				t.Errorf("Generator.Generate(%v) returned nil", subType)
				return
			}
			if obj.Sprite == nil {
				t.Errorf("Generator.Generate(%v) returned nil sprite", subType)
			}
			if obj.SubType != subType {
				t.Errorf("Generated object SubType = %v, want %v", obj.SubType, subType)
			}
		})
	}
}

// TestGenerator_GenerateDeterminism tests deterministic generation.
func TestGenerator_GenerateDeterminism(t *testing.T) {
	gen := NewGenerator()

	config := Config{
		SubType: SubTypeTable,
		Width:   32,
		Height:  32,
		GenreID: "fantasy",
		Seed:    12345,
	}

	// Generate twice with same seed
	obj1, err1 := gen.Generate(config)
	if err1 != nil {
		t.Fatalf("First generation failed: %v", err1)
	}

	obj2, err2 := gen.Generate(config)
	if err2 != nil {
		t.Fatalf("Second generation failed: %v", err2)
	}

	// Compare properties
	if obj1.Type != obj2.Type {
		t.Error("Objects have different types")
	}
	if obj1.SubType != obj2.SubType {
		t.Error("Objects have different subtypes")
	}
	if obj1.Width != obj2.Width {
		t.Error("Objects have different widths")
	}
	if obj1.Height != obj2.Height {
		t.Error("Objects have different heights")
	}
	if obj1.Name != obj2.Name {
		t.Error("Objects have different names")
	}

	// Compare sprites pixel by pixel
	if obj1.Sprite.Bounds() != obj2.Sprite.Bounds() {
		t.Fatal("Sprites have different bounds")
	}

	bounds := obj1.Sprite.Bounds()
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			c1 := obj1.Sprite.At(x, y)
			c2 := obj2.Sprite.At(x, y)
			if c1 != c2 {
				t.Errorf("Sprites differ at (%d, %d)", x, y)
				return
			}
		}
	}
}

// TestGenerator_GenerateGenres tests generation across different genres.
func TestGenerator_GenerateGenres(t *testing.T) {
	gen := NewGenerator()

	genres := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapoc"}

	for _, genre := range genres {
		t.Run(genre, func(t *testing.T) {
			config := Config{
				SubType: SubTypeChest,
				Width:   32,
				Height:  32,
				GenreID: genre,
				Seed:    12345,
			}

			obj, err := gen.Generate(config)
			if err != nil {
				t.Errorf("Generation failed for genre %s: %v", genre, err)
				return
			}
			if obj == nil {
				t.Errorf("Generated nil object for genre %s", genre)
				return
			}
			if obj.GenreID != genre {
				t.Errorf("Object GenreID = %v, want %v", obj.GenreID, genre)
			}
			// Genre should affect the name prefix
			if obj.Name == "" {
				t.Error("Object has empty name")
			}
		})
	}
}

// TestGenerator_GenerateSizes tests generation with different sizes.
func TestGenerator_GenerateSizes(t *testing.T) {
	gen := NewGenerator()

	sizes := []struct {
		width  int
		height int
	}{
		{16, 16},
		{32, 32},
		{64, 64},
		{32, 64},
		{64, 32},
	}

	for _, size := range sizes {
		t.Run("", func(t *testing.T) {
			config := Config{
				SubType: SubTypeTable,
				Width:   size.width,
				Height:  size.height,
				GenreID: "fantasy",
				Seed:    12345,
			}

			obj, err := gen.Generate(config)
			if err != nil {
				t.Errorf("Generation failed for size %dx%d: %v", size.width, size.height, err)
				return
			}
			if obj.Width != size.width {
				t.Errorf("Object Width = %v, want %v", obj.Width, size.width)
			}
			if obj.Height != size.height {
				t.Errorf("Object Height = %v, want %v", obj.Height, size.height)
			}
			if obj.Sprite.Bounds().Dx() != size.width {
				t.Errorf("Sprite width = %v, want %v", obj.Sprite.Bounds().Dx(), size.width)
			}
			if obj.Sprite.Bounds().Dy() != size.height {
				t.Errorf("Sprite height = %v, want %v", obj.Sprite.Bounds().Dy(), size.height)
			}
		})
	}
}

// TestGenerator_GenerateProperties tests that generated objects have correct properties.
func TestGenerator_GenerateProperties(t *testing.T) {
	gen := NewGenerator()

	tests := []struct {
		subType            SubType
		expectCollidable   bool
		expectInteractable bool
		expectHarmful      bool
	}{
		{SubTypeTable, true, true, false},
		{SubTypeChest, true, true, false},
		{SubTypePlant, false, false, false},
		{SubTypeTorch, false, true, false},
		{SubTypeBarrel, true, false, false},
		{SubTypePillar, true, false, false},
		{SubTypeSpikes, true, false, true},
		{SubTypeFirePit, true, false, true},
		{SubTypePoisonGas, false, false, true},
	}

	for _, tt := range tests {
		t.Run(tt.subType.String(), func(t *testing.T) {
			config := Config{
				SubType: tt.subType,
				Width:   32,
				Height:  32,
				GenreID: "fantasy",
				Seed:    12345,
			}

			obj, err := gen.Generate(config)
			if err != nil {
				t.Fatalf("Generation failed: %v", err)
			}

			if obj.Collidable != tt.expectCollidable {
				t.Errorf("Object Collidable = %v, want %v", obj.Collidable, tt.expectCollidable)
			}
			if obj.Interactable != tt.expectInteractable {
				t.Errorf("Object Interactable = %v, want %v", obj.Interactable, tt.expectInteractable)
			}
			if obj.Harmful != tt.expectHarmful {
				t.Errorf("Object Harmful = %v, want %v", obj.Harmful, tt.expectHarmful)
			}
		})
	}
}

// BenchmarkGenerator_Generate benchmarks object generation.
func BenchmarkGenerator_Generate(b *testing.B) {
	gen := NewGenerator()
	config := Config{
		SubType: SubTypeTable,
		Width:   32,
		Height:  32,
		GenreID: "fantasy",
		Seed:    12345,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = gen.Generate(config)
	}
}

// BenchmarkGenerator_GenerateAllTypes benchmarks generation of all object types.
func BenchmarkGenerator_GenerateAllTypes(b *testing.B) {
	gen := NewGenerator()

	subtypes := []SubType{
		SubTypeTable, SubTypeChest, SubTypePlant, SubTypeStatue,
		SubTypeBarrel, SubTypePillar, SubTypeSpikes, SubTypeFirePit,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for j, subType := range subtypes {
			config := Config{
				SubType: subType,
				Width:   32,
				Height:  32,
				GenreID: "fantasy",
				Seed:    int64(12345 + j),
			}
			_, _ = gen.Generate(config)
		}
	}
}
